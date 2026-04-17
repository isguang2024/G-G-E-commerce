package middleware

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

type appSecurityProfile struct {
	AppKey           string
	AuthMode         string
	FrontendEntryURL string
	BackendEntryURL  string
	HealthCheckURL   string
	Capabilities     models.MetaJSON
}

type appSecurityProfileCacheItem struct {
	profile   appSecurityProfile
	expiresAt time.Time
}

var appSecurityProfileCache sync.Map

const appSecurityProfileTTL = 60 * time.Second

// DynamicAppSecurity 按当前请求命中的 app_key 动态注入 CORS / CSP / shared_cookie 策略头。
func DynamicAppSecurity(db *gorm.DB, logger *zap.Logger, env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		appKey := strings.TrimSpace(c.GetString("app_key"))
		requestHost := strings.TrimSpace(c.GetString("request_host"))
		origin := strings.TrimSpace(c.GetHeader("Origin"))
		requestScheme := "http"
		if c.Request != nil && c.Request.TLS != nil {
			requestScheme = "https"
		}

		profile := loadAppSecurityProfile(db, appKey)
		allowedOrigins, hasConfiguredAllowlist := resolveAllowedOrigins(profile, requestHost, requestScheme)
		corsAllowed := origin == "" || allowOrigin(origin, requestHost, allowedOrigins, hasConfiguredAllowlist, env)

		allowHeaders := "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, X-API-Key, X-Auth-Workspace-Id"
		c.Writer.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Vary", "Origin")
		if origin != "" && corsAllowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		cspValue := resolveCSP(profile, allowedOrigins)
		if cspValue != "" {
			c.Writer.Header().Set("Content-Security-Policy", cspValue)
		}
		if appKey != "" {
			c.Writer.Header().Set("X-App-Key", appKey)
		}

		authMode, cookieScopeMode, cookieDomain := resolveCookiePolicy(profile, c)
		if authMode == "shared_cookie" {
			c.Writer.Header().Set("X-Auth-Cookie-Mode", "shared_cookie")
			c.Writer.Header().Set("X-Auth-Cookie-SameSite", "None")
			c.Writer.Header().Set("X-Auth-Cookie-Secure", "true")
			c.Writer.Header().Set("X-Auth-Cookie-Scope-Mode", cookieScopeMode)
			if cookieDomain != "" {
				c.Writer.Header().Set("X-Auth-Cookie-Domain", cookieDomain)
			}
		}
		c.Set("app_auth_mode", authMode)

		if c.Request != nil && c.Request.Method == http.MethodOptions {
			if !corsAllowed {
				if logger != nil {
					logger.Warn("CORS preflight rejected by dynamic allowlist",
						zap.String("app_key", appKey),
						zap.String("origin", origin),
						zap.String("request_host", requestHost))
				}
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "origin not allowed"})
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func loadAppSecurityProfile(db *gorm.DB, appKey string) appSecurityProfile {
	key := strings.TrimSpace(appKey)
	if key == "" || db == nil {
		return appSecurityProfile{}
	}
	if cached, ok := appSecurityProfileCache.Load(key); ok {
		item := cached.(appSecurityProfileCacheItem)
		if time.Now().Before(item.expiresAt) {
			return item.profile
		}
		appSecurityProfileCache.Delete(key)
	}

	var app models.App
	if err := db.Select("app_key", "auth_mode", "frontend_entry_url", "backend_entry_url", "health_check_url", "capabilities").
		Where("app_key = ? AND deleted_at IS NULL", key).
		Take(&app).Error; err != nil {
		return appSecurityProfile{AppKey: key}
	}
	profile := appSecurityProfile{
		AppKey:           strings.TrimSpace(app.AppKey),
		AuthMode:         strings.TrimSpace(app.AuthMode),
		FrontendEntryURL: strings.TrimSpace(app.FrontendEntryURL),
		BackendEntryURL:  strings.TrimSpace(app.BackendEntryURL),
		HealthCheckURL:   strings.TrimSpace(app.HealthCheckURL),
		Capabilities:     app.Capabilities,
	}
	appSecurityProfileCache.Store(key, appSecurityProfileCacheItem{
		profile:   profile,
		expiresAt: time.Now().Add(appSecurityProfileTTL),
	})
	return profile
}

func resolveAllowedOrigins(profile appSecurityProfile, requestHost, requestScheme string) ([]string, bool) {
	origins := make([]string, 0)
	if profile.Capabilities != nil {
		if raw, ok := profile.Capabilities["cors_origins"]; ok {
			if arr, ok := raw.([]interface{}); ok {
				for _, item := range arr {
					text := strings.TrimSpace(toString(item))
					if text != "" {
						origins = append(origins, text)
					}
				}
			}
			if arr, ok := raw.([]string); ok {
				for _, item := range arr {
					text := strings.TrimSpace(item)
					if text != "" {
						origins = append(origins, text)
					}
				}
			}
		}
	}
	if origin := originFromURL(profile.FrontendEntryURL); origin != "" {
		origins = append(origins, origin)
	}
	if requestHost = strings.TrimSpace(requestHost); requestHost != "" {
		origins = append(origins, requestScheme+"://"+requestHost)
	}
	uniq := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))
	for _, item := range origins {
		val := strings.TrimSpace(item)
		if val == "" {
			continue
		}
		if _, ok := seen[val]; ok {
			continue
		}
		seen[val] = struct{}{}
		uniq = append(uniq, val)
	}
	return uniq, len(uniq) > 0
}

func allowOrigin(origin, requestHost string, allowedOrigins []string, hasConfiguredAllowlist bool, env string) bool {
	if strings.EqualFold(strings.TrimSpace(env), "development") && !hasConfiguredAllowlist {
		return true
	}
	for _, item := range allowedOrigins {
		if item == "*" || strings.EqualFold(item, origin) {
			return true
		}
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if strings.EqualFold(parsed.Host, strings.TrimSpace(requestHost)) {
		return true
	}
	return false
}

func resolveCSP(profile appSecurityProfile, allowedOrigins []string) string {
	if profile.Capabilities != nil {
		if raw, ok := profile.Capabilities["csp"]; ok {
			if value := strings.TrimSpace(toString(raw)); value != "" {
				return value
			}
		}
	}
	connectSrc := []string{"'self'"}
	for _, item := range allowedOrigins {
		if strings.TrimSpace(item) != "" {
			connectSrc = append(connectSrc, item)
		}
	}
	return "default-src 'self'; img-src 'self' data: https:; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'; connect-src " + strings.Join(connectSrc, " ") + "; frame-ancestors 'self'"
}

func resolveCookiePolicy(profile appSecurityProfile, c *gin.Context) (string, string, string) {
	authMode := strings.TrimSpace(profile.AuthMode)
	cookieScopeMode := "inherit"
	cookieDomain := ""

	if rawBinding, ok := c.Get("app_entry_binding"); ok && rawBinding != nil {
		if binding, ok := rawBinding.(*models.AppHostBinding); ok && binding != nil {
			metaAuthMode := strings.TrimSpace(toString(binding.Meta["auth_mode"]))
			if metaAuthMode == "" {
				metaAuthMode = strings.TrimSpace(toString(binding.Meta["authMode"]))
			}
			if metaAuthMode != "" {
				authMode = metaAuthMode
			}
			cookieScopeMode = strings.TrimSpace(toString(binding.Meta["cookie_scope_mode"]))
			if cookieScopeMode == "" {
				cookieScopeMode = strings.TrimSpace(toString(binding.Meta["cookieScopeMode"]))
			}
			if cookieScopeMode == "" {
				cookieScopeMode = "inherit"
			}
			cookieDomain = strings.TrimSpace(toString(binding.Meta["cookie_domain"]))
			if cookieDomain == "" {
				cookieDomain = strings.TrimSpace(toString(binding.Meta["cookieDomain"]))
			}
		}
	}

	if authMode == "" {
		authMode = "inherit_host"
	}
	return authMode, cookieScopeMode, cookieDomain
}

func originFromURL(rawURL string) string {
	target := strings.TrimSpace(rawURL)
	if target == "" {
		return ""
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}

func toString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}
