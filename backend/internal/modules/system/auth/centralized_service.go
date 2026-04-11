package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

const (
	defaultAuthTenantID         = "default"
	authCallbackCodeTTL         = 5 * time.Minute
	authCallbackStatusPending   = "pending"
	authCallbackStatusExchanged = "exchanged"
	authProtocolVersionDefault  = "callback-v1"
)

type CentralizedAuthService interface {
	CreateCallback(ctx context.Context, input CreateAuthCallbackInput) (*CreateAuthCallbackResult, error)
	ExchangeCallback(ctx context.Context, input ExchangeAuthCallbackInput) (*ExchangeAuthCallbackResult, error)
}

type CreateAuthCallbackInput struct {
	UserID             uuid.UUID
	TargetAppKey       string
	RedirectURI        string
	TargetPath         string
	NavigationSpaceKey string
	State              string
	Nonce              string
	RequestHost        string
}

type CreateAuthCallbackResult struct {
	Code                string
	State               string
	TargetAppKey        string
	RedirectURI         string
	RedirectTo          string
	TargetPath          string
	NavigationSpaceKey  string
	AuthProtocolVersion string
}

type ExchangeAuthCallbackInput struct {
	Code         string
	State        string
	Nonce        string
	TargetAppKey string
	RedirectURI  string
}

type ExchangeAuthCallbackResult struct {
	LoginResponse       *LoginServiceResponse
	AppKey              string
	NavigationSpaceKey  string
	HomePath            string
	AuthProtocolVersion string
}

type LoginServiceResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         map[string]interface{}
}

type centralizedAuthService struct {
	db       *gorm.DB
	authSvc  AuthService
	userRepo user.UserRepository
}

func NewCentralizedAuthService(db *gorm.DB, authSvc AuthService, userRepo user.UserRepository) CentralizedAuthService {
	return &centralizedAuthService{db: db, authSvc: authSvc, userRepo: userRepo}
}

func (s *centralizedAuthService) CreateCallback(ctx context.Context, input CreateAuthCallbackInput) (*CreateAuthCallbackResult, error) {
	if input.UserID == uuid.Nil {
		return nil, errors.New("缺少用户标识")
	}
	targetAppKey := strings.TrimSpace(input.TargetAppKey)
	redirectURI := strings.TrimSpace(input.RedirectURI)
	state := strings.TrimSpace(input.State)
	nonce := strings.TrimSpace(input.Nonce)
	if targetAppKey == "" || redirectURI == "" || state == "" || nonce == "" {
		return nil, errors.New("缺少 centralized_login 必填参数")
	}
	parsedRedirect, err := s.validateRedirectURI(targetAppKey, redirectURI)
	if err != nil {
		return nil, err
	}
	targetPath := normalizeInternalPath(input.TargetPath)
	callbackCode := strings.ReplaceAll(uuid.NewString(), "-", "")
	record := &models.AuthCallbackCode{
		TenantID:           defaultAuthTenantID,
		Code:               callbackCode,
		UserID:             input.UserID,
		TargetAppKey:       targetAppKey,
		RedirectURI:        parsedRedirect.String(),
		TargetPath:         targetPath,
		NavigationSpaceKey: strings.TrimSpace(input.NavigationSpaceKey),
		State:              state,
		Nonce:              nonce,
		RequestHost:        normalizeHost(input.RequestHost),
		Status:             authCallbackStatusPending,
		ExpiresAt:          time.Now().Add(authCallbackCodeTTL),
	}
	if err := s.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, fmt.Errorf("创建 callback code 失败: %w", err)
	}
	redirectTo := appendQuery(parsedRedirect, map[string]string{
		"code":                  callbackCode,
		"state":                 state,
		"target_app_key":        targetAppKey,
		"redirect_uri":          parsedRedirect.String(),
		"auth_protocol_version": authProtocolVersionDefault,
	})
	if targetPath != "" {
		redirectTo = appendQueryString(redirectTo, "target_path", targetPath)
	}
	if strings.TrimSpace(input.NavigationSpaceKey) != "" {
		redirectTo = appendQueryString(redirectTo, "navigation_space_key", strings.TrimSpace(input.NavigationSpaceKey))
	}
	return &CreateAuthCallbackResult{
		Code:                callbackCode,
		State:               state,
		TargetAppKey:        targetAppKey,
		RedirectURI:         parsedRedirect.String(),
		RedirectTo:          redirectTo,
		TargetPath:          targetPath,
		NavigationSpaceKey:  strings.TrimSpace(input.NavigationSpaceKey),
		AuthProtocolVersion: authProtocolVersionDefault,
	}, nil
}

func (s *centralizedAuthService) ExchangeCallback(ctx context.Context, input ExchangeAuthCallbackInput) (*ExchangeAuthCallbackResult, error) {
	code := strings.TrimSpace(input.Code)
	state := strings.TrimSpace(input.State)
	nonce := strings.TrimSpace(input.Nonce)
	targetAppKey := strings.TrimSpace(input.TargetAppKey)
	redirectURI := strings.TrimSpace(input.RedirectURI)
	if code == "" || state == "" || nonce == "" || targetAppKey == "" || redirectURI == "" {
		return nil, errors.New("缺少 callback exchange 必填参数")
	}
	parsedRedirect, err := s.validateRedirectURI(targetAppKey, redirectURI)
	if err != nil {
		return nil, err
	}

	var record models.AuthCallbackCode
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND code = ? AND deleted_at IS NULL", defaultAuthTenantID, code).
			First(&record).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("callback code 不存在或已失效")
			}
			return err
		}
		if record.Status != authCallbackStatusPending || record.UsedAt != nil {
			return errors.New("callback code 已被消费")
		}
		if time.Now().After(record.ExpiresAt) {
			return errors.New("callback code 已过期")
		}
		if record.TargetAppKey != targetAppKey ||
			record.RedirectURI != parsedRedirect.String() ||
			record.State != state ||
			record.Nonce != nonce {
			return errors.New("callback 参数校验失败")
		}
		now := time.Now()
		return tx.Model(&record).Updates(map[string]interface{}{
			"status":     authCallbackStatusExchanged,
			"used_at":    now,
			"updated_at": now,
		}).Error
	})
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetByID(record.UserID)
	if err != nil {
		return nil, fmt.Errorf("加载 callback 用户失败: %w", err)
	}
	loginResp, err := s.authSvc.BuildLoginResponse(u)
	if err != nil {
		return nil, err
	}
	landingAppKey, landingSpaceKey, landingHomePath, err := s.resolveLanding(record.TargetAppKey, record.TargetPath, record.NavigationSpaceKey)
	if err != nil {
		return nil, err
	}
	userMap, _ := loginResp.User.(map[string]interface{})
	return &ExchangeAuthCallbackResult{
		LoginResponse: &LoginServiceResponse{
			AccessToken:  loginResp.AccessToken,
			RefreshToken: loginResp.RefreshToken,
			ExpiresIn:    loginResp.ExpiresIn,
			User:         userMap,
		},
		AppKey:              landingAppKey,
		NavigationSpaceKey:  landingSpaceKey,
		HomePath:            landingHomePath,
		AuthProtocolVersion: authProtocolVersionDefault,
	}, nil
}

func (s *centralizedAuthService) resolveLanding(targetAppKey, targetPath, navigationSpaceKey string) (string, string, string, error) {
	var app models.App
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", targetAppKey).First(&app).Error; err != nil {
		return "", "", "", fmt.Errorf("目标 APP 不存在: %w", err)
	}
	homePath := normalizeInternalPath(targetPath)
	if homePath == "" {
		homePath = normalizeInternalPath(app.FrontendEntryURL)
	}
	if homePath == "" {
		homePath = "/"
	}
	spaceKey := strings.TrimSpace(navigationSpaceKey)
	if spaceKey == "" {
		spaceKey = strings.TrimSpace(app.DefaultSpaceKey)
	}
	return app.AppKey, spaceKey, homePath, nil
}

func (s *centralizedAuthService) validateRedirectURI(targetAppKey, redirectURI string) (*url.URL, error) {
	parsed, err := url.Parse(redirectURI)
	if err != nil || parsed == nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("redirect_uri 必须是绝对地址")
	}
	if !strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https") {
		return nil, errors.New("redirect_uri 协议仅允许 http/https")
	}
	if normalizeInternalPath(parsed.Path) == "" {
		return nil, errors.New("redirect_uri path 非法")
	}

	var bindings []models.AppHostBinding
	if err := s.db.Where("app_key = ? AND status = ? AND deleted_at IS NULL", targetAppKey, "normal").Find(&bindings).Error; err != nil {
		return nil, fmt.Errorf("加载 APP host binding 失败: %w", err)
	}
	targetHost := normalizeHost(parsed.Host)
	targetHostname := normalizeHostname(parsed.Host)
	allowed := false
	for _, binding := range bindings {
		if hostMatchesBinding(targetHost, targetHostname, binding.Host) {
			allowed = true
			break
		}
		if hostMatchesBinding(targetHost, targetHostname, metaString(binding.Meta, "callback_host", "callbackHost")) {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, errors.New("redirect_uri 不在已登记的 APP/callback host 白名单内")
	}
	return parsed, nil
}

func appendQuery(parsed *url.URL, values map[string]string) string {
	next := *parsed
	query := next.Query()
	for key, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}
		query.Set(key, value)
	}
	next.RawQuery = query.Encode()
	return next.String()
}

func appendQueryString(rawURL string, key string, value string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func metaString(meta models.MetaJSON, keys ...string) string {
	for _, key := range keys {
		if value, ok := meta[key]; ok {
			if text, ok := value.(string); ok {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func normalizeInternalPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if !strings.HasPrefix(trimmed, "/") || strings.HasPrefix(trimmed, "//") {
		return ""
	}
	return trimmed
}

func normalizeHost(host string) string {
	trimmed := strings.TrimSpace(strings.ToLower(host))
	trimmed = strings.TrimPrefix(trimmed, "http://")
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.Trim(trimmed, "/")
	if slash := strings.Index(trimmed, "/"); slash >= 0 {
		trimmed = trimmed[:slash]
	}
	return trimmed
}

func normalizeHostname(host string) string {
	normalized := normalizeHost(host)
	if normalized == "" {
		return ""
	}
	if strings.HasPrefix(normalized, "[") && strings.Contains(normalized, "]") {
		return strings.Trim(strings.SplitN(normalized, "]", 2)[0], "[]")
	}
	if parsedPortHost, _, err := net.SplitHostPort(normalized); err == nil {
		return strings.Trim(strings.ToLower(parsedPortHost), "[]")
	}
	if parsed, err := url.Parse("http://" + normalized); err == nil {
		return strings.Trim(strings.ToLower(parsed.Hostname()), "[]")
	}
	return strings.Trim(strings.ToLower(normalized), "[]")
}

func hostMatchesBinding(targetHost, targetHostname, bindingHost string) bool {
	normalizedBindingHost := normalizeHost(bindingHost)
	if normalizedBindingHost == "" {
		return false
	}
	if normalizedBindingHost == targetHost {
		return true
	}
	return normalizeHostname(normalizedBindingHost) == targetHostname
}
