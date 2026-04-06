package space

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacerolebinding"
)

const (
	DefaultMenuSpaceKey        = models.DefaultMenuSpaceKey
	requestSpaceKeyQuery       = "space_key"
	requestSpaceKeyHeader      = "X-Space-Key"
	requestSpaceAltHeader      = "X-Menu-Space"
	requestForwardedHostHeader = "X-Forwarded-Host"
	requestRealHostHeader      = "X-Real-Host"
	spaceAccessModeAll         = "all"
	spaceAccessModePlatform    = "platform_admin"
	spaceAccessModeTeam        = "team_admin"
	spaceAccessModeRoleCodes   = "role_codes"
	spaceModeSingle            = "single"
	spaceModeMulti             = "multi"
)

type SpaceAccessProfile struct {
	Mode             string   `json:"mode"`
	AllowedRoleCodes []string `json:"allowed_role_codes"`
}

func NormalizeSpaceKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return DefaultMenuSpaceKey
	}
	return target
}

func NormalizeSpaceMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case spaceModeMulti:
		return spaceModeMulti
	default:
		return spaceModeSingle
	}
}

func CurrentSpaceMode(db *gorm.DB, appKey string) string {
	if db == nil {
		return spaceModeSingle
	}
	var app models.App
	err := db.Where("app_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey)).First(&app).Error
	if err != nil {
		return spaceModeSingle
	}
	return NormalizeSpaceMode(app.SpaceMode)
}

func SaveCurrentSpaceMode(db *gorm.DB, appKey, mode string) (string, error) {
	normalized := NormalizeSpaceMode(mode)
	if db == nil {
		return normalized, nil
	}
	return normalized, db.Model(&models.App{}).
		Where("app_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey)).
		Updates(map[string]interface{}{"space_mode": normalized, "status": "normal"}).Error
}

func IsSingleSpaceMode(db *gorm.DB, appKey string) bool {
	return CurrentSpaceMode(db, appKey) == spaceModeSingle
}

func NormalizeHost(value string) string {
	target := strings.TrimSpace(strings.ToLower(value))
	if target == "" {
		return ""
	}
	if idx := strings.Index(target, ","); idx >= 0 {
		target = strings.TrimSpace(target[:idx])
	}
	if idx := strings.Index(target, "://"); idx >= 0 {
		if parsed, err := url.Parse(target); err == nil && parsed.Host != "" {
			target = parsed.Host
		}
	}
	if host, _, err := net.SplitHostPort(target); err == nil {
		target = host
	}
	target = strings.TrimSuffix(target, ".")
	return target
}

func RequestHost(c *gin.Context) string {
	if c == nil {
		return ""
	}
	for _, header := range []string{requestForwardedHostHeader, requestRealHostHeader} {
		if value := NormalizeHost(c.GetHeader(header)); value != "" {
			return value
		}
	}
	if c.Request == nil {
		return ""
	}
	return NormalizeHost(c.Request.Host)
}

func RequestSpaceKey(c *gin.Context) string {
	if c == nil {
		return ""
	}
	if value := strings.TrimSpace(c.Query(requestSpaceKeyQuery)); value != "" {
		return NormalizeSpaceKey(value)
	}
	for _, header := range []string{requestSpaceKeyHeader, requestSpaceAltHeader} {
		if value := strings.TrimSpace(c.GetHeader(header)); value != "" {
			return NormalizeSpaceKey(value)
		}
	}
	return ""
}

func ResolveSpaceKey(db *gorm.DB, c *gin.Context) string {
	if db == nil {
		return DefaultMenuSpaceKey
	}
	appKey := currentContextAppKey(c)
	userID := currentContextUserID(c)
	tenantID := currentContextCollaborationWorkspaceID(c)
	key, _, err := ResolveCurrentSpaceKey(db, appKey, RequestHost(c), RequestSpaceKey(c), userID, tenantID)
	if err == nil && strings.TrimSpace(key) != "" {
		return key
	}
	return DefaultMenuSpaceKey
}

func ResolveCurrentSpaceKey(db *gorm.DB, appKey, host, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (string, string, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	defaultSpaceKey, err := loadAppDefaultSpaceKey(db, normalizedAppKey)
	if err != nil {
		return DefaultMenuSpaceKey, "", err
	}
	if defaultSpaceKey == "" {
		defaultSpaceKey = DefaultMenuSpaceKey
	}

	if IsSingleSpaceMode(db, normalizedAppKey) {
		explicit := NormalizeSpaceKey(requestedSpaceKey)
		if explicit != "" && explicit != defaultSpaceKey {
			ok, existsErr := spaceExists(db, normalizedAppKey, explicit)
			if existsErr != nil {
				return defaultSpaceKey, "", existsErr
			}
			if ok {
				allowed, accessErr := CanAccessSpace(db, userID, tenantID, explicit)
				if accessErr != nil {
					return defaultSpaceKey, "", accessErr
				}
				if allowed {
					return explicit, "single_mode_explicit", nil
				}
			}
		}
		return defaultSpaceKey, "single_mode_default", nil
	}

	if explicit := NormalizeSpaceKey(requestedSpaceKey); explicit != "" {
		ok, existsErr := spaceExists(db, normalizedAppKey, explicit)
		if existsErr != nil {
			return defaultSpaceKey, "", existsErr
		}
		if ok {
			allowed, accessErr := CanAccessSpace(db, userID, tenantID, explicit)
			if accessErr != nil {
				return defaultSpaceKey, "", accessErr
			}
			if allowed {
				return explicit, "explicit", nil
			}
		}
	}

	resolvedByHost, source, hostErr := ResolveSpaceKeyByHost(db, normalizedAppKey, host)
	if hostErr != nil {
		return defaultSpaceKey, "", hostErr
	}
	if strings.TrimSpace(resolvedByHost) != "" {
		allowed, accessErr := CanAccessSpace(db, userID, tenantID, resolvedByHost)
		if accessErr != nil {
			return defaultSpaceKey, "", accessErr
		}
		if allowed {
			return resolvedByHost, source, nil
		}
	}

	if allowed, accessErr := CanAccessSpace(db, userID, tenantID, defaultSpaceKey); accessErr == nil && allowed {
		return defaultSpaceKey, "app_default", nil
	}

	return defaultSpaceKey, "fallback_default", nil
}

func currentContextUserID(c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &id
}

func currentContextCollaborationWorkspaceID(c *gin.Context) *uuid.UUID {
	if c == nil {
		return nil
	}
	raw, ok := c.Get("collaboration_workspace_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok || strings.TrimSpace(value) == "" {
		return nil
	}
	id, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &id
}

func ExtractSpaceAccessProfile(meta models.MetaJSON) SpaceAccessProfile {
	mode := strings.TrimSpace(toMetaString(meta, "access_mode", "accessMode"))
	switch mode {
	case spaceAccessModePlatform, spaceAccessModeTeam, spaceAccessModeRoleCodes:
	default:
		mode = spaceAccessModeAll
	}
	return SpaceAccessProfile{
		Mode:             mode,
		AllowedRoleCodes: normalizeMetaStringList(meta["allowed_role_codes"], meta["allowedRoleCodes"]),
	}
}

func CanAccessSpace(db *gorm.DB, userID *uuid.UUID, tenantID *uuid.UUID, spaceKey string) (bool, error) {
	if db == nil {
		return NormalizeSpaceKey(spaceKey) == DefaultMenuSpaceKey, nil
	}
	var record models.MenuSpace
	if err := db.Select("space_key", "status", "meta").
		Where("space_key = ? AND deleted_at IS NULL", NormalizeSpaceKey(spaceKey)).
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if record.Status == "disabled" {
		return false, nil
	}
	profile := ExtractSpaceAccessProfile(record.Meta)
	switch profile.Mode {
	case spaceAccessModeAll:
		return true, nil
	case spaceAccessModePlatform:
		if userID == nil {
			return false, nil
		}
		return hasPlatformRoleCode(db, *userID, []string{"admin"})
	case spaceAccessModeTeam:
		if userID == nil {
			return false, nil
		}
		return hasTeamAdminRole(db, *userID, tenantID)
	case spaceAccessModeRoleCodes:
		if userID == nil || len(profile.AllowedRoleCodes) == 0 {
			return false, nil
		}
		return hasAnyRoleCode(db, *userID, tenantID, profile.AllowedRoleCodes)
	default:
		return true, nil
	}
}

func hasPlatformRoleCode(db *gorm.DB, userID uuid.UUID, roleCodes []string) (bool, error) {
	matched, err := workspacerolebinding.HasPersonalRoleCodesByUserID(db, userID, roleCodes, true)
	if err != nil || matched {
		return matched, err
	}
	return hasLegacyPlatformRoleCode(db, userID, roleCodes)
}

func hasLegacyPlatformRoleCode(db *gorm.DB, userID uuid.UUID, roleCodes []string) (bool, error) {
	var count int64
	err := db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.collaboration_workspace_id IS NULL").
		Where("roles.collaboration_workspace_id IS NULL").
		Where("roles.code IN ?", roleCodes).
		Where("roles.status = ?", "normal").
		Where("roles.deleted_at IS NULL").
		Count(&count).Error
	return count > 0, err
}

func hasTeamAdminRole(db *gorm.DB, userID uuid.UUID, tenantID *uuid.UUID) (bool, error) {
	query := db.Table("collaboration_workspace_members").
		Where("user_id = ?", userID).
		Where("status = ?", "active").
		Where("role_code = ?", "team_admin")
	if tenantID != nil {
		query = query.Where("collaboration_workspace_id = ?", *tenantID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func hasAnyRoleCode(db *gorm.DB, userID uuid.UUID, tenantID *uuid.UUID, roleCodes []string) (bool, error) {
	normalized := normalizeStringSlice(roleCodes)
	if len(normalized) == 0 {
		return false, nil
	}
	platformMatched, err := hasPlatformRoleCode(db, userID, normalized)
	if err != nil || platformMatched {
		return platformMatched, err
	}
	if tenantID != nil {
		workspaceMatched, err := workspacerolebinding.HasTeamRoleCodesByTenantAndUser(db, *tenantID, userID, normalized, true)
		if err != nil {
			return false, err
		}
		if workspaceMatched {
			return true, nil
		}
		var attachedRoleCount int64
		if err := db.Table("user_roles").
			Joins("JOIN roles ON roles.id = user_roles.role_id").
			Where("user_roles.user_id = ?", userID).
			Where("user_roles.collaboration_workspace_id = ?", *tenantID).
			Where("roles.code IN ?", normalized).
			Where("roles.status = ?", "normal").
			Where("roles.deleted_at IS NULL").
			Count(&attachedRoleCount).Error; err != nil {
			return false, err
		}
		if attachedRoleCount > 0 {
			return true, nil
		}
		var teamCount int64
		if err := db.Table("collaboration_workspace_members").
			Where("user_id = ?", userID).
			Where("collaboration_workspace_id = ?", *tenantID).
			Where("status = ?", "active").
			Where("role_code IN ?", normalized).
			Count(&teamCount).Error; err != nil {
			return false, err
		}
		if teamCount > 0 {
			return true, nil
		}
	}
	return false, nil
}

func normalizeStringSlice(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, raw := range values {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func toMetaString(meta models.MetaJSON, keys ...string) string {
	for _, key := range keys {
		if value, ok := meta[key]; ok {
			if text, ok := value.(string); ok {
				return text
			}
		}
	}
	return ""
}

func normalizeMetaStringList(values ...interface{}) []string {
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, raw := range values {
		switch typed := raw.(type) {
		case []string:
			for _, item := range typed {
				value := strings.TrimSpace(item)
				if value == "" {
					continue
				}
				if _, ok := seen[value]; ok {
					continue
				}
				seen[value] = struct{}{}
				result = append(result, value)
			}
		case []interface{}:
			for _, item := range typed {
				text, ok := item.(string)
				if !ok {
					continue
				}
				value := strings.TrimSpace(text)
				if value == "" {
					continue
				}
				if _, ok := seen[value]; ok {
					continue
				}
				seen[value] = struct{}{}
				result = append(result, value)
			}
		}
	}
	return result
}

func ResolveSpaceKeyByHost(db *gorm.DB, appKey, host string) (string, string, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	if IsSingleSpaceMode(db, normalizedAppKey) {
		return DefaultMenuSpaceKey, "single_mode_default", nil
	}
	if db == nil {
		return DefaultMenuSpaceKey, "fallback_default", nil
	}
	normalizedHost := NormalizeHost(host)
	if normalizedHost != "" {
		var appBinding models.AppHostBinding
		if err := db.Where("host = ? AND status = ? AND deleted_at IS NULL", normalizedHost, "normal").First(&appBinding).Error; err == nil {
			key := NormalizeSpaceKey(appBinding.DefaultSpaceKey)
			if normalizeAppKey(appBinding.AppKey) == normalizedAppKey {
				ok, err := spaceExists(db, normalizedAppKey, key)
				if err == nil && ok {
					return key, "app_host_binding", nil
				}
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultMenuSpaceKey, "", err
		}

		var binding models.MenuSpaceHostBinding
		if err := db.Where("host = ? AND status = ?", normalizedHost, "normal").First(&binding).Error; err == nil {
			key := NormalizeSpaceKey(binding.SpaceKey)
			ok, err := spaceExists(db, normalizedAppKey, key)
			if err == nil && ok {
				return key, "legacy_space_host_binding", nil
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return DefaultMenuSpaceKey, "", err
		}
	}
	defaultSpaceKey, err := loadAppDefaultSpaceKey(db, normalizedAppKey)
	if err != nil {
		return DefaultMenuSpaceKey, "", err
	}
	if defaultSpaceKey == "" {
		defaultSpaceKey = DefaultMenuSpaceKey
	}
	ok, err := spaceExists(db, normalizedAppKey, defaultSpaceKey)
	if err != nil {
		return defaultSpaceKey, "", err
	}
	if ok {
		return defaultSpaceKey, "app_default", nil
	}
	return defaultSpaceKey, "fallback_default", nil
}

func spaceExists(db *gorm.DB, appKey, spaceKey string) (bool, error) {
	var count int64
	if err := db.Model(&models.MenuSpace{}).
		Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey), NormalizeSpaceKey(spaceKey)).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func loadAppDefaultSpaceKey(db *gorm.DB, appKey string) (string, error) {
	if db == nil {
		return DefaultMenuSpaceKey, nil
	}
	var app models.App
	err := db.Select("default_space_key").Where("app_key = ? AND deleted_at IS NULL", normalizeAppKey(appKey)).First(&app).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return DefaultMenuSpaceKey, nil
	}
	if err != nil {
		return DefaultMenuSpaceKey, err
	}
	return NormalizeSpaceKey(app.DefaultSpaceKey), nil
}

func normalizeAppKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return models.DefaultAppKey
	}
	return target
}

func currentContextAppKey(c *gin.Context) string {
	if c == nil {
		return models.DefaultAppKey
	}
	return normalizeAppKey(c.GetString("app_key"))
}
