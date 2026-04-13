package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/register"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

const (
	socialTenantID       = "default"
	oauthStateTTL        = 5 * time.Minute
	socialTokenTTL       = 10 * time.Minute
	socialTokenIssuer    = "gge-social-oauth"
	socialIntentLogin    = "login"
	socialIntentRegister = "register"
	socialIntentConflict = "conflict"
)

type Service interface {
	BuildAuthorizeURL(ctx context.Context, providerKey string, input AuthorizeInput) (string, error)
	HandleCallback(ctx context.Context, providerKey string, input CallbackInput) (*CallbackResult, error)
	ExchangeSocialToken(ctx context.Context, token string) (*ExchangeResult, error)
	BindBySocialToken(ctx context.Context, tx *gorm.DB, token string, userID uuid.UUID) error
}

type AuthorizeInput struct {
	RequestHost  string
	RequestPath  string
	LoginPageKey string
	PageScene    string
	TargetAppKey string
}

type CallbackInput struct {
	Code  string
	State string
}

type CallbackResult struct {
	RedirectPath string
	SocialToken  string
}

type ExchangeResult struct {
	Intent        string
	ProviderKey   string
	ProviderName  string
	ProviderUID   string
	ProviderUser  string
	Email         string
	AvatarURL     string
	MatchedUserID string
	NeedRegister  bool
	LoginResponse *dto.LoginResponse
}

type socialTokenClaims struct {
	Intent        string `json:"intent"`
	ProviderKey   string `json:"provider_key"`
	ProviderName  string `json:"provider_name"`
	ProviderUID   string `json:"provider_uid"`
	ProviderUser  string `json:"provider_user"`
	Email         string `json:"email"`
	AvatarURL     string `json:"avatar_url"`
	UserID        string `json:"user_id"`
	MatchedUserID string `json:"matched_user_id"`
	jwt.RegisteredClaims
}

type service struct {
	db       *gorm.DB
	repo     *Repository
	authSvc  auth.AuthService
	userRepo user.UserRepository
	resolver *register.Resolver
	secret   string
	logger   *zap.Logger
	client   *http.Client
}

func NewService(db *gorm.DB, authSvc auth.AuthService, userRepo user.UserRepository, resolver *register.Resolver, secret string, logger *zap.Logger) Service {
	return &service{
		db:       db,
		repo:     NewRepository(db),
		authSvc:  authSvc,
		userRepo: userRepo,
		resolver: resolver,
		secret:   strings.TrimSpace(secret),
		logger:   logger,
		client:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *service) BuildAuthorizeURL(ctx context.Context, providerKey string, input AuthorizeInput) (string, error) {
	provider, err := s.repo.FindProviderByKey(ctx, socialTenantID, providerKey)
	if err != nil {
		return "", fmt.Errorf("加载社交提供方失败: %w", err)
	}
	if !provider.Enabled {
		return "", errors.New("社交登录未启用")
	}

	clientID := firstNonEmpty(strings.TrimSpace(provider.ClientID), strings.TrimSpace(os.Getenv("GG_SOCIAL_GITHUB_CLIENT_ID")))
	if clientID == "" {
		return "", errors.New("社交登录未配置 client_id")
	}

	callbackURI := s.resolveCallbackURI(provider, input.RequestHost)
	rawState := strings.ReplaceAll(uuid.NewString(), "-", "")
	nonce := strings.ReplaceAll(uuid.NewString(), "-", "")
	scene := normalizeScene(input.PageScene)

	stateRecord := &systemmodels.SocialOAuthState{
		TenantID:     socialTenantID,
		ProviderKey:  provider.ProviderKey,
		State:        rawState,
		LoginPageKey: strings.TrimSpace(input.LoginPageKey),
		PageScene:    scene,
		TargetAppKey: strings.TrimSpace(input.TargetAppKey),
		RequestPath:  strings.TrimSpace(input.RequestPath),
		RedirectURI:  callbackURI,
		Nonce:        nonce,
		Meta: systemmodels.MetaJSON{
			"request_host": strings.TrimSpace(input.RequestHost),
		},
		ExpiresAt: time.Now().Add(oauthStateTTL),
	}
	if err := s.db.WithContext(ctx).Create(stateRecord).Error; err != nil {
		return "", fmt.Errorf("保存 OAuth state 失败: %w", err)
	}

	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", callbackURI)
	q.Set("response_type", "code")
	q.Set("state", rawState)
	if scope := strings.TrimSpace(provider.Scope); scope != "" {
		q.Set("scope", scope)
	}

	authURL := strings.TrimSpace(provider.AuthURL)
	if authURL == "" {
		return "", errors.New("社交登录未配置 auth_url")
	}
	u, parseErr := url.Parse(authURL)
	if parseErr != nil || u == nil {
		return "", errors.New("社交登录 auth_url 非法")
	}
	ex := u.Query()
	for k, values := range q {
		for _, value := range values {
			ex.Set(k, value)
		}
	}
	u.RawQuery = ex.Encode()
	return u.String(), nil
}

func (s *service) HandleCallback(ctx context.Context, providerKey string, input CallbackInput) (*CallbackResult, error) {
	code := strings.TrimSpace(input.Code)
	state := strings.TrimSpace(input.State)
	if code == "" || state == "" {
		return nil, errors.New("缺少 OAuth 回调参数")
	}

	provider, err := s.repo.FindProviderByKey(ctx, socialTenantID, providerKey)
	if err != nil {
		return nil, fmt.Errorf("加载社交提供方失败: %w", err)
	}

	var stateRecord systemmodels.SocialOAuthState
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND provider_key = ? AND state = ? AND deleted_at IS NULL", socialTenantID, provider.ProviderKey, state).
			First(&stateRecord).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("oauth_state_not_found")
			}
			return err
		}
		if stateRecord.UsedAt != nil {
			return errors.New("oauth_state_used")
		}
		if time.Now().After(stateRecord.ExpiresAt) {
			return errors.New("oauth_state_expired")
		}
		now := time.Now()
		return tx.Model(&stateRecord).Updates(map[string]interface{}{"used_at": now, "updated_at": now}).Error
	}); err != nil {
		return nil, err
	}

	profile, err := s.fetchGitHubProfile(ctx, provider, code, stateRecord.RedirectURI)
	if err != nil {
		return nil, err
	}

	redirectPath := "/account/auth/social-callback"
	targetPath := defaultScenePath(stateRecord.PageScene)
	if rp := strings.TrimSpace(stateRecord.RequestPath); rp != "" && strings.HasPrefix(rp, "/") {
		targetPath = rp
	}

	claims := socialTokenClaims{
		ProviderKey:  provider.ProviderKey,
		ProviderName: provider.ProviderName,
		ProviderUID:  profile.ProviderUID,
		ProviderUser: profile.Username,
		Email:        profile.Email,
		AvatarURL:    profile.AvatarURL,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    socialTokenIssuer,
			Subject:   profile.ProviderUID,
			ID:        strings.ReplaceAll(uuid.NewString(), "-", ""),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(socialTokenTTL)),
		},
	}

	if linked, findErr := s.repo.FindByProviderUID(ctx, socialTenantID, provider.ProviderKey, profile.ProviderUID); findErr == nil && linked != nil {
		claims.Intent = socialIntentLogin
		claims.UserID = linked.UserID.String()
		_ = s.repo.UpdateLastLogin(ctx, socialTenantID, linked.ID)
	} else {
		claims.Intent = socialIntentRegister
		if email := strings.TrimSpace(profile.Email); email != "" {
			if existing, userErr := s.userRepo.GetByEmail(email); userErr == nil && existing != nil {
				claims.Intent = socialIntentConflict
				claims.MatchedUserID = existing.ID.String()
			}
		}
	}

	token, err := s.signSocialToken(claims)
	if err != nil {
		return nil, err
	}

	return &CallbackResult{RedirectPath: appendQuery(redirectPath, map[string]string{
		"social_token":   token,
		"provider":       provider.ProviderKey,
		"target_path":    targetPath,
		"login_page_key": strings.TrimSpace(stateRecord.LoginPageKey),
	}), SocialToken: token}, nil
}

func (s *service) ExchangeSocialToken(ctx context.Context, token string) (*ExchangeResult, error) {
	claims, err := s.parseSocialToken(token)
	if err != nil {
		return nil, err
	}
	out := &ExchangeResult{
		Intent:        claims.Intent,
		ProviderKey:   claims.ProviderKey,
		ProviderName:  claims.ProviderName,
		ProviderUID:   claims.ProviderUID,
		ProviderUser:  claims.ProviderUser,
		Email:         claims.Email,
		AvatarURL:     claims.AvatarURL,
		MatchedUserID: claims.MatchedUserID,
		NeedRegister:  claims.Intent != socialIntentLogin,
	}
	if claims.Intent == socialIntentLogin {
		userID, parseErr := uuid.Parse(strings.TrimSpace(claims.UserID))
		if parseErr != nil || userID == uuid.Nil {
			return nil, errors.New("social_token 用户信息无效")
		}
		u, userErr := s.userRepo.GetByID(userID)
		if userErr != nil {
			return nil, fmt.Errorf("加载社交绑定用户失败: %w", userErr)
		}
		loginResp, loginErr := s.authSvc.BuildLoginResponse(u)
		if loginErr != nil {
			return nil, fmt.Errorf("生成登录态失败: %w", loginErr)
		}
		out.LoginResponse = loginResp
		out.NeedRegister = false
	}
	return out, nil
}

func (s *service) BindBySocialToken(ctx context.Context, tx *gorm.DB, token string, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return errors.New("用户标识为空")
	}
	claims, err := s.parseSocialToken(token)
	if err != nil {
		return err
	}
	if claims.Intent != socialIntentRegister && claims.Intent != socialIntentConflict {
		return errors.New("当前 social_token 不可用于绑定")
	}
	db := tx
	if db == nil {
		db = s.db
	}
	repo := NewRepository(db)
	account := &systemmodels.UserSocialAccount{
		TenantID:         socialTenantID,
		UserID:           userID,
		ProviderKey:      claims.ProviderKey,
		ProviderUID:      claims.ProviderUID,
		ProviderUsername: claims.ProviderUser,
		ProviderEmail:    claims.Email,
		AvatarURL:        claims.AvatarURL,
		Profile: systemmodels.MetaJSON{
			"source": "social_token",
		},
		LinkedAt: time.Now(),
	}
	return repo.CreateInTx(db, account)
}

type gitHubProfile struct {
	ProviderUID string
	Username    string
	Email       string
	AvatarURL   string
}

func (s *service) fetchGitHubProfile(ctx context.Context, provider *systemmodels.SocialAuthProvider, code, redirectURI string) (*gitHubProfile, error) {
	clientID := firstNonEmpty(strings.TrimSpace(provider.ClientID), strings.TrimSpace(os.Getenv("GG_SOCIAL_GITHUB_CLIENT_ID")))
	clientSecret := firstNonEmpty(strings.TrimSpace(provider.ClientSecret), strings.TrimSpace(os.Getenv("GG_SOCIAL_GITHUB_CLIENT_SECRET")))
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("GitHub OAuth 未配置 client_id/client_secret")
	}

	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", redirectURI)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, provider.TokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub token 失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub token 响应异常: %s", strings.TrimSpace(string(body)))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
		Desc        string `json:"error_description"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析 GitHub token 响应失败: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("GitHub token 获取失败: %s %s", tokenResp.Error, tokenResp.Desc)
	}

	userReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, provider.UserInfoURL, nil)
	userReq.Header.Set("Accept", "application/vnd.github+json")
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	userResp, err := s.client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("请求 GitHub 用户信息失败: %w", err)
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)
	if userResp.StatusCode >= 300 {
		return nil, fmt.Errorf("GitHub 用户信息响应异常: %s", strings.TrimSpace(string(userBody)))
	}

	var profileResp struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(userBody, &profileResp); err != nil {
		return nil, fmt.Errorf("解析 GitHub 用户信息失败: %w", err)
	}

	email := strings.TrimSpace(profileResp.Email)
	if email == "" {
		emailReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
		emailReq.Header.Set("Accept", "application/vnd.github+json")
		emailReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
		emailResp, emailErr := s.client.Do(emailReq)
		if emailErr == nil {
			defer emailResp.Body.Close()
			payload, _ := io.ReadAll(emailResp.Body)
			if emailResp.StatusCode < 300 {
				var emails []struct {
					Email    string `json:"email"`
					Primary  bool   `json:"primary"`
					Verified bool   `json:"verified"`
				}
				if json.Unmarshal(payload, &emails) == nil {
					for _, item := range emails {
						if item.Primary && item.Verified && strings.TrimSpace(item.Email) != "" {
							email = strings.TrimSpace(item.Email)
							break
						}
					}
					if email == "" {
						for _, item := range emails {
							if item.Verified && strings.TrimSpace(item.Email) != "" {
								email = strings.TrimSpace(item.Email)
								break
							}
						}
					}
				}
			}
		}
	}

	if profileResp.ID == 0 {
		return nil, errors.New("GitHub 用户标识为空")
	}
	return &gitHubProfile{
		ProviderUID: fmt.Sprintf("%d", profileResp.ID),
		Username:    strings.TrimSpace(profileResp.Login),
		Email:       email,
		AvatarURL:   strings.TrimSpace(profileResp.AvatarURL),
	}, nil
}

func (s *service) resolveCallbackURI(provider *systemmodels.SocialAuthProvider, requestHost string) string {
	if redirect := strings.TrimSpace(provider.RedirectURI); redirect != "" {
		return redirect
	}
	host := strings.TrimSpace(requestHost)
	if host == "" {
		host = "127.0.0.1:8080"
	}
	scheme := "https"
	if strings.HasPrefix(host, "127.0.0.1") || strings.HasPrefix(host, "localhost") || strings.Contains(host, ":") {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/api/v1/auth/oauth/%s/callback", scheme, host, provider.ProviderKey)
}

func (s *service) signSocialToken(claims socialTokenClaims) (string, error) {
	if s.secret == "" {
		return "", errors.New("jwt secret 未配置")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

func (s *service) parseSocialToken(token string) (*socialTokenClaims, error) {
	if s.secret == "" {
		return nil, errors.New("jwt secret 未配置")
	}
	parsed, err := jwt.ParseWithClaims(token, &socialTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token")
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, errors.New("social_token 无效或已过期")
	}
	claims, ok := parsed.Claims.(*socialTokenClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("social_token 无效")
	}
	if strings.TrimSpace(claims.ProviderKey) == "" || strings.TrimSpace(claims.ProviderUID) == "" {
		return nil, errors.New("social_token 缺少关键信息")
	}
	return claims, nil
}

func defaultScenePath(scene string) string {
	switch normalizeScene(scene) {
	case "register":
		return "/account/auth/register"
	case "forget_password":
		return "/account/auth/forget-password"
	default:
		return "/account/auth/login"
	}
}

func normalizeScene(scene string) string {
	switch strings.TrimSpace(scene) {
	case "register":
		return "register"
	case "forget_password", "forgetPassword":
		return "forget_password"
	default:
		return "login"
	}
}

func appendQuery(rawURL string, query map[string]string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed == nil {
		return rawURL
	}
	values := parsed.Query()
	for key, value := range query {
		if strings.TrimSpace(value) != "" {
			values.Set(key, value)
		}
	}
	parsed.RawQuery = values.Encode()
	return parsed.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
