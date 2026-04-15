// auth.go: ogen Handler implementations for the /auth/* OpenAPI surface.
// First slice of the auth domain migration: login (public) + me (token).
// The legacy gin handlers in internal/modules/system/auth/handler.go remain
// behind /api/v1 until every operation here is mounted via the ogen bridge.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/apperr"
	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/register"
)

// CtxClientIP carries the originating client IP from the gin bridge into
// ogen handlers (used by Login for last-login bookkeeping).
const (
	CtxClientIP    ctxKey = "client_ip"
	CtxRequestHost ctxKey = "request_host"
	CtxRequestPath ctxKey = "request_path"
	CtxUserAgent   ctxKey = "user_agent"
)

func (h *APIHandler) Login(ctx context.Context, req *gen.LoginRequest) (gen.LoginRes, error) {
	if req == nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return nil, &apperr.ParamError{Msg: "用户名和密码必填"}
	}

	// prompt 语义处理（OIDC-like）:
	// prompt=none 意味着调用方期望静默登录，不应到达 Login 端点；直接返回 login_required。
	// prompt=login 意味着强制重认证，记录日志后正常继续。
	if req.Prompt.Set {
		prompt := strings.TrimSpace(req.Prompt.Value)
		if prompt == "none" {
			h.logger.Warn("login endpoint received prompt=none; caller should use /auth/callback/silent instead")
			return (*gen.LoginUnauthorized)(&gen.Error{
				Code:    apperr.CodeLoginRequired,
				Message: "login_required: silent authentication not available at login endpoint",
			}), nil
		}
		if prompt == "login" {
			h.logger.Info("login with prompt=login (force re-authentication)")
		}
	}

	ip := clientIPFromCtx(ctx)
	username := strings.TrimSpace(req.Username)
	resp, err := h.authSvc.Login(req.Username, req.Password, ip)
	if err != nil {
		h.logger.Debug("login failed", zap.Error(err))
		// 登录失败也要审计：用户名 + 失败原因（password 不会被序列化，字段已在 redact 名单）。
		h.audit.Record(ctx, audit.Event{
			Action:       "system.auth.login",
			ResourceType: "user",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata:     map[string]any{"username": username, "prompt": optString(req.Prompt)},
		})
		return nil, err
	}
	// 登录成功：actor_id 在 Event.After 里自带，供后续离线分析（ctx 里此时还没有 actor_id）。
	var userIDStr string
	if userMap, ok := resp.User.(map[string]interface{}); ok {
		if id, ok := userMap["id"].(string); ok {
			userIDStr = id
		}
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.auth.login",
		ResourceType: "user",
		ResourceID:   userIDStr,
		Outcome:      audit.OutcomeSuccess,
		Metadata:     map[string]any{"username": username},
	})
	out := &gen.LoginResponse{
		AccessToken:  gen.NewOptNilString(resp.AccessToken),
		RefreshToken: gen.NewOptNilString(resp.RefreshToken),
		ExpiresIn:    gen.NewOptNilInt(resp.ExpiresIn),
	}
	if userMap, ok := resp.User.(map[string]interface{}); ok {
		out.User = gen.NewOptNilLoginResponseUser(toJxRawMap(userMap))
		if centralizedLoginRequested(req) && h.centralizedAuthSvc != nil {
			userID, parseErr := uuidFromMapValue(userMap["id"])
			if parseErr != nil {
				return nil, &apperr.ParamError{Msg: "登录态用户标识无效"}
			}
			callback, callbackErr := h.centralizedAuthSvc.CreateCallback(ctx, auth.CreateAuthCallbackInput{
				UserID:             userID,
				TargetAppKey:       optString(req.TargetAppKey),
				RedirectURI:        optString(req.RedirectURI),
				TargetPath:         optString(req.TargetPath),
				NavigationSpaceKey: optString(req.NavigationSpaceKey),
				State:              optString(req.State),
				Nonce:              optString(req.Nonce),
				RequestHost:        requestHostFromCtx(ctx),
			})
			if callbackErr != nil {
				return nil, &apperr.ParamError{Msg: callbackErr.Error()}
			}
			out.AccessToken = gen.OptNilString{}
			out.RefreshToken = gen.OptNilString{}
			out.ExpiresIn = gen.OptNilInt{}
			out.User = gen.OptNilLoginResponseUser{}
			out.Callback = gen.OptAuthCallbackPayload{Value: gen.AuthCallbackPayload{
				Mode:                gen.AuthCallbackPayloadModeTokenExchange,
				Code:                callback.Code,
				State:               callback.State,
				TargetAppKey:        callback.TargetAppKey,
				RedirectURI:         callback.RedirectURI,
				RedirectTo:          callback.RedirectTo,
				TargetPath:          gen.NewOptString(callback.TargetPath),
				NavigationSpaceKey:  gen.NewOptString(callback.NavigationSpaceKey),
				AuthProtocolVersion: gen.NewOptString(callback.AuthProtocolVersion),
			}, Set: true}
		}
	}
	return out, nil
}

func (h *APIHandler) Register(ctx context.Context, req *gen.RegisterRequest) (gen.RegisterRes, error) {
	if req == nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return nil, &apperr.ParamError{Msg: "用户名和密码必填"}
	}
	if h.registerSvc == nil {
		return nil, errors.New("register service not configured")
	}
	result, err := h.registerSvc.Register(ctx, register.RegisterInput{
		Username:                 req.Username,
		Password:                 req.Password,
		ConfirmPassword:          optString(req.ConfirmPassword),
		Email:                    optString(req.Email),
		Nickname:                 optString(req.Nickname),
		CaptchaToken:             optString(req.CaptchaToken),
		InvitationCode:           optString(req.InvitationCode),
		AgreementVersion:         optString(req.AgreementVersion),
		Host:                     stringFromCtx(ctx, CtxRequestHost),
		Path:                     stringFromCtx(ctx, CtxRequestPath),
		IP:                       clientIPFromCtx(ctx),
		UserAgent:                stringFromCtx(ctx, CtxUserAgent),
		SourceAppKey:             optString(req.SourceAppKey),
		SourceNavigationSpaceKey: optString(req.SourceNavigationSpaceKey),
		SourceHomePath:           optString(req.SourceHomePath),
	})
	if err != nil {
		h.logger.Debug("register failed", zap.Error(err))
		h.audit.Record(ctx, audit.Event{
			Action:       "system.auth.register",
			ResourceType: "user",
			Outcome:      audit.OutcomeError,
			ErrorCode:    errorCodeOf(err),
			Metadata: map[string]any{
				"username":      strings.TrimSpace(req.Username),
				"register_mode": optString(req.SourceAppKey),
			},
		})
		if errors.Is(err, register.ErrPublicRegisterDisabled) {
			return nil, &apperr.ParamError{Msg: "公开注册未开启"}
		}
		return nil, err
	}
	if h.socialSvc != nil && req.SocialToken.Set && result.User != nil {
		if bindErr := h.socialSvc.BindBySocialToken(ctx, nil, strings.TrimSpace(req.SocialToken.Value), result.User.ID); bindErr != nil {
			return nil, &apperr.ParamError{Msg: bindErr.Error()}
		}
	}
	// 成功：record resource_id = 新建用户 ID，便于按 actor 反查注册事件。
	var newUserID string
	if result.User != nil {
		newUserID = result.User.ID.String()
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.auth.register",
		ResourceType: "user",
		ResourceID:   newUserID,
		Outcome:      audit.OutcomeSuccess,
		Metadata: map[string]any{
			"username":      strings.TrimSpace(req.Username),
			"register_mode": optString(req.SourceAppKey),
		},
	})
	out := &gen.LoginResponse{}
	if result.Login != nil {
		out.AccessToken = gen.NewOptNilString(result.Login.AccessToken)
		out.RefreshToken = gen.NewOptNilString(result.Login.RefreshToken)
		out.ExpiresIn = gen.NewOptNilInt(result.Login.ExpiresIn)
		if userMap, ok := result.Login.User.(map[string]interface{}); ok {
			out.User = gen.NewOptNilLoginResponseUser(toJxRawMap(userMap))
		}
	}
	if result.Pending {
		out.Pending = gen.NewOptNilBool(true)
	}
	if result.Landing != nil {
		out.Landing = gen.NewOptNilLoginResponseLanding(gen.LoginResponseLanding{
			URL:                gen.NewOptString(result.Landing.URL),
			AppKey:             gen.NewOptString(result.Landing.AppKey),
			NavigationSpaceKey: gen.NewOptString(result.Landing.NavigationSpaceKey),
			HomePath:           gen.NewOptString(result.Landing.HomePath),
		})
	}
	return out, nil
}

func (h *APIHandler) ExchangeSocialToken(ctx context.Context, req *gen.SocialTokenExchangeRequest) (gen.ExchangeSocialTokenRes, error) {
	if req == nil || h.socialSvc == nil {
		return nil, &apperr.ParamError{Msg: "social exchange 未启用"}
	}
	result, err := h.socialSvc.ExchangeSocialToken(ctx, strings.TrimSpace(req.SocialToken))
	if err != nil {
		return nil, &apperr.ParamError{Msg: err.Error()}
	}
	out := &gen.SocialTokenExchangeResponse{
		Intent:      gen.SocialTokenExchangeResponseIntent(result.Intent),
		ProviderKey: result.ProviderKey,
		ProviderUID: result.ProviderUID,
	}
	if strings.TrimSpace(result.ProviderName) != "" {
		out.ProviderName = gen.NewOptString(result.ProviderName)
	}
	if strings.TrimSpace(result.ProviderUser) != "" {
		out.ProviderUser = gen.NewOptString(result.ProviderUser)
	}
	if strings.TrimSpace(result.Email) != "" {
		out.Email = gen.NewOptString(result.Email)
	}
	if strings.TrimSpace(result.AvatarURL) != "" {
		out.AvatarURL = gen.NewOptString(result.AvatarURL)
	}
	if strings.TrimSpace(result.MatchedUserID) != "" {
		out.MatchedUserID = gen.NewOptString(result.MatchedUserID)
	}
	out.NeedRegister = gen.NewOptBool(result.NeedRegister)
	if result.LoginResponse != nil {
		out.AccessToken = gen.NewOptString(result.LoginResponse.AccessToken)
		out.RefreshToken = gen.NewOptString(result.LoginResponse.RefreshToken)
		out.ExpiresIn = gen.NewOptInt(result.LoginResponse.ExpiresIn)
		if userMap, ok := result.LoginResponse.User.(map[string]interface{}); ok {
			out.User = gen.NewOptSocialTokenExchangeResponseUser(gen.SocialTokenExchangeResponseUser(toJxRawMap(userMap)))
		}
	}
	return out, nil
}

// GetRegisterContext: 由 host + path 命中入口并合并策略，返回前端注册页所需的
// 有效上下文。公开接口，未命中任何 entry 时退回 default entry。
func (h *APIHandler) GetRegisterContext(ctx context.Context, params gen.GetRegisterContextParams) (gen.GetRegisterContextRes, error) {
	host := ""
	if params.Host.Set {
		host = strings.TrimSpace(params.Host.Value)
	}
	path := ""
	if params.Path.Set {
		path = strings.TrimSpace(params.Path.Value)
	}
	if path == "" {
		path = "/account/auth/register"
	}
	if h.registerResolver == nil {
		return nil, errors.New("register resolver not configured")
	}
	eff, err := h.registerResolver.Resolve(ctx, host, path)
	if err != nil {
		return nil, err
	}
	out := &gen.RegisterContext{
		EntryCode:           eff.EntryCode,
		EntryAppKey:         eff.EntryAppKey,
		LoginPageKey:        eff.LoginPageKey,
		IsSystemReserved:    eff.IsSystemReserved,
		AllowPublicRegister: eff.AllowPublicRegister,
		RequireInvite:       eff.RequireInvite,
		RequireEmailVerify:  eff.RequireEmailVerify,
		RequireCaptcha:      eff.RequireCaptcha,
		AutoLogin:           eff.AutoLogin,
		RoleCodes:           eff.RoleCodes,
		FeaturePackageKeys:  eff.FeaturePackageKeys,
	}
	if eff.EntryName != "" {
		out.EntryName = gen.NewOptString(eff.EntryName)
	}
	if eff.RegisterSource != "" {
		out.RegisterSource = gen.NewOptString(eff.RegisterSource)
	}
	if eff.TargetURL != "" {
		out.TargetURL = gen.NewOptString(eff.TargetURL)
	}
	if eff.TargetAppKey != "" {
		out.TargetAppKey = gen.NewOptString(eff.TargetAppKey)
	}
	if eff.TargetNavigationSpaceKey != "" {
		out.TargetNavigationSpaceKey = gen.NewOptString(eff.TargetNavigationSpaceKey)
	}
	if eff.TargetHomePath != "" {
		out.TargetHomePath = gen.NewOptString(eff.TargetHomePath)
	}
	if eff.AgreementVersion != "" {
		out.AgreementVersion = gen.NewOptString(eff.AgreementVersion)
	}
	if eff.CaptchaProvider != "" && eff.CaptchaProvider != "none" {
		out.CaptchaProvider = gen.NewOptString(eff.CaptchaProvider)
		out.CaptchaSiteKey = gen.NewOptString(eff.CaptchaSiteKey)
	}
	return out, nil
}

func (h *APIHandler) GetLoginPageContext(
	ctx context.Context,
	params gen.GetLoginPageContextParams,
) (gen.GetLoginPageContextRes, error) {
	host := ""
	if params.Host.Set {
		host = strings.TrimSpace(params.Host.Value)
	}
	path := strings.TrimSpace(optString(params.Path))
	if path == "" {
		path = "/account/auth/login"
	}
	targetAppKey := strings.TrimSpace(optString(params.TargetAppKey))
	loginPageKey := strings.TrimSpace(optString(params.LoginPageKey))
	pageScene := ""
	if params.PageScene.Set {
		pageScene = strings.TrimSpace(string(params.PageScene.Value))
	}
	if h.registerResolver == nil {
		return nil, errors.New("register resolver not configured")
	}
	info, err := h.registerResolver.ResolveLoginPageContext(ctx, register.ResolveLoginPageContextInput{
		Host:         host,
		Path:         path,
		TargetAppKey: targetAppKey,
		LoginPageKey: loginPageKey,
		PageScene:    pageScene,
	})
	if err != nil {
		return nil, err
	}
	socialCapability := h.resolveSocialCapability(ctx, info.PageScene, host, path)
	if info.TemplateConfig == nil {
		info.TemplateConfig = map[string]interface{}{}
	}
	social := map[string]interface{}{}
	if existing, ok := info.TemplateConfig["social"].(map[string]interface{}); ok && existing != nil {
		for k, v := range existing {
			social[k] = v
		}
	}
	social["capability"] = socialCapability
	info.TemplateConfig["social"] = social
	out := &gen.LoginPageContext{
		AppKey:         info.AppKey,
		LoginPageKey:   info.LoginPageKey,
		LoginUIMode:    info.LoginUiMode,
		SSOMode:        info.SsoMode,
		ResolvedBy:     info.ResolvedBy,
		PageScene:      info.PageScene,
		TargetAppKey:   info.TargetAppKey,
		RegisterPath:   info.RegisterPath,
		RegisterAppKey: info.RegisterAppKey,
		EntryCode:      gen.NewOptString(info.EntryCode),
		EntryName:      gen.NewOptString(info.EntryName),
	}
	if info.TemplateName != "" {
		out.TemplateName = gen.NewOptString(info.TemplateName)
	}
	if len(info.TemplateConfig) > 0 {
		raw := make(gen.LoginPageContextTemplateConfig, len(info.TemplateConfig))
		for k, v := range info.TemplateConfig {
			b, jsonErr := json.Marshal(v)
			if jsonErr == nil {
				raw[k] = jx.Raw(b)
			}
		}
		out.TemplateConfig = gen.NewOptLoginPageContextTemplateConfig(raw)
	}
	return out, nil
}

func (h *APIHandler) resolveSocialCapability(ctx context.Context, pageScene, host, path string) map[string]interface{} {
	capability := map[string]interface{}{
		"allow":     true,
		"reason":    "",
		"providers": []string{},
	}
	var providers []struct {
		ProviderKey string `gorm:"column:provider_key"`
	}
	if err := h.db.WithContext(ctx).
		Model(&models.SocialAuthProvider{}).
		Select("provider_key").
		Where("tenant_id = ? AND enabled = ? AND deleted_at IS NULL", "default", true).
		Find(&providers).Error; err != nil {
		capability["allow"] = false
		capability["reason"] = "provider_query_failed"
		return capability
	}
	keys := make([]string, 0, len(providers))
	for _, item := range providers {
		if strings.TrimSpace(item.ProviderKey) != "" {
			keys = append(keys, strings.TrimSpace(item.ProviderKey))
		}
	}
	capability["providers"] = keys
	if len(keys) == 0 {
		capability["allow"] = false
		capability["reason"] = "no_enabled_provider"
		return capability
	}
	if strings.TrimSpace(pageScene) == "register" && h.registerResolver != nil {
		eff, err := h.registerResolver.Resolve(ctx, host, path)
		if err == nil && eff != nil && !eff.AllowPublicRegister {
			capability["allow"] = false
			capability["reason"] = "public_register_disabled"
		}
	}
	return capability
}

func (h *APIHandler) RefreshToken(ctx context.Context, req *gen.RefreshTokenRequest) (gen.RefreshTokenRes, error) {
	if req == nil || strings.TrimSpace(req.RefreshToken) == "" {
		return nil, &apperr.ParamError{Msg: "缺少 refresh_token"}
	}
	resp, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &gen.TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
}

func (h *APIHandler) Logout(ctx context.Context) (gen.LogoutRes, error) {
	// Logout 目前是无状态操作（token 失效由前端清理），但仍落审计，方便追踪用户主动退出。
	var actorID string
	if id, ok := userIDFromContext(ctx); ok {
		actorID = id.String()
	}
	h.audit.Record(ctx, audit.Event{
		Action:       "system.auth.logout",
		ResourceType: "user",
		ResourceID:   actorID,
		Outcome:      audit.OutcomeSuccess,
	})
	return ok(), nil
}

func (h *APIHandler) ExchangeAuthCallback(ctx context.Context, req *gen.AuthCallbackExchangeRequest) (gen.ExchangeAuthCallbackRes, error) {
	if req == nil || h.centralizedAuthSvc == nil {
		return nil, &apperr.ParamError{Msg: "callback exchange 未启用"}
	}
	result, err := h.centralizedAuthSvc.ExchangeCallback(ctx, auth.ExchangeAuthCallbackInput{
		Code:         req.Code,
		State:        req.State,
		Nonce:        req.Nonce,
		TargetAppKey: req.TargetAppKey,
		RedirectURI:  req.RedirectURI,
	})
	if err != nil {
		return nil, &apperr.ParamError{Msg: err.Error()}
	}
	out := &gen.LoginResponse{
		AccessToken:  gen.NewOptNilString(result.LoginResponse.AccessToken),
		RefreshToken: gen.NewOptNilString(result.LoginResponse.RefreshToken),
		ExpiresIn:    gen.NewOptNilInt(result.LoginResponse.ExpiresIn),
		Landing: gen.NewOptNilLoginResponseLanding(gen.LoginResponseLanding{
			AppKey:             gen.NewOptString(result.AppKey),
			NavigationSpaceKey: gen.NewOptString(result.NavigationSpaceKey),
			HomePath:           gen.NewOptString(result.HomePath),
		}),
	}
	if len(result.LoginResponse.User) > 0 {
		out.User = gen.NewOptNilLoginResponseUser(toJxRawMap(result.LoginResponse.User))
	}
	return out, nil
}

func toJxRawMap(m map[string]interface{}) gen.LoginResponseUser {
	raw := make(gen.LoginResponseUser, len(m))
	for k, v := range m {
		b, err := json.Marshal(v)
		if err != nil {
			continue
		}
		raw[k] = jx.Raw(b)
	}
	return raw
}

func (h *APIHandler) GetAuthMe(ctx context.Context) (gen.GetAuthMeRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, &apperr.UnauthError{Msg: "未认证"}
	}
	u, err := h.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("auth.me user not found", zap.String("user_id", userID.String()))
			return nil, &apperr.UnauthError{Msg: "登录状态已失效，请重新登录"}
		}
		h.logger.Error("auth.me lookup failed", zap.Error(err))
		return nil, err
	}

	// Resolve current auth workspace context. Falls back to the personal
	// workspace when the JWT did not pin one (mirrors legacy GetUserInfo).
	var authWorkspaceID uuid.UUID
	authWorkspaceType := strings.TrimSpace(stringFromCtx(ctx, CtxAuthWorkspaceType))
	if raw := strings.TrimSpace(stringFromCtx(ctx, CtxAuthWorkspaceID)); raw != "" {
		if parsed, err := uuid.Parse(raw); err == nil {
			authWorkspaceID = parsed
		}
	}
	var collaborationWorkspaceID uuid.UUID
	if raw := strings.TrimSpace(stringFromCtx(ctx, CtxCollaborationWorkspaceID)); raw != "" {
		if parsed, err := uuid.Parse(raw); err == nil {
			collaborationWorkspaceID = parsed
		}
	}
	if authWorkspaceID == uuid.Nil && h.service != nil {
		if ws, err := h.service.EnsurePersonalWorkspaceForUser(userID); err == nil && ws != nil {
			authWorkspaceID = ws.ID
			authWorkspaceType = ws.WorkspaceType
			if collaborationWorkspaceID == uuid.Nil && ws.CollaborationWorkspaceID != nil {
				collaborationWorkspaceID = *ws.CollaborationWorkspaceID
			}
		}
	}

	// Resolve effective permission keys for the current workspace via the
	// evaluator (super_admin shortcut + workspace ∩ role intersection).
	actions := make([]string, 0)
	if h.evaluator != nil {
		resolved, err := h.evaluator.Resolve(ctx, userID, authWorkspaceID)
		if err != nil {
			h.logger.Warn("auth.me resolve actions failed", zap.Error(err))
		} else {
			for k := range resolved.Keys {
				actions = append(actions, k)
			}
		}
	}

	roles := make([]gen.AuthMeRole, 0, len(u.Roles))
	for _, r := range u.Roles {
		role := gen.AuthMeRole{ID: r.ID, Code: r.Code, Name: r.Name}
		if r.Description != "" {
			role.Description = gen.NewOptNilString(r.Description)
		}
		roles = append(roles, role)
	}

	out := &gen.AuthMe{
		ID:           u.ID,
		Username:     u.Username,
		IsSuperAdmin: u.IsSuperAdmin,
		Actions:      actions,
		Roles:        roles,
	}
	if u.Nickname != "" {
		out.Nickname = gen.NewOptNilString(u.Nickname)
	}
	if u.Email != "" {
		out.Email = gen.NewOptNilString(u.Email)
	}
	if u.AvatarURL != "" {
		out.Avatar = gen.NewOptNilString(u.AvatarURL)
		out.AvatarURL = gen.NewOptNilString(u.AvatarURL)
	}
	if u.Phone != "" {
		out.Phone = gen.NewOptNilString(u.Phone)
	}
	if u.Status != "" {
		out.Status = gen.NewOptNilString(u.Status)
	}
	if !u.CreatedAt.IsZero() {
		out.CreatedAt = gen.NewOptNilString(u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	}
	if authWorkspaceID != uuid.Nil {
		out.CurrentAuthWorkspaceID = gen.NewOptNilUUID(authWorkspaceID)
	}
	if authWorkspaceType != "" {
		out.CurrentAuthWorkspaceType = gen.NewOptNilString(authWorkspaceType)
	}
	if collaborationWorkspaceID != uuid.Nil {
		out.CollaborationWorkspaceID = gen.NewOptNilUUID(collaborationWorkspaceID)
		out.CurrentCollaborationWorkspaceID = gen.NewOptNilUUID(collaborationWorkspaceID)
	}
	return out, nil
}

func stringFromCtx(ctx context.Context, key ctxKey) string {
	v, _ := ctx.Value(key).(string)
	return v
}

func requestHostFromCtx(ctx context.Context) string {
	return strings.TrimSpace(stringFromCtx(ctx, CtxRequestHost))
}

func requestPathFromCtx(ctx context.Context) string {
	return strings.TrimSpace(stringFromCtx(ctx, CtxRequestPath))
}

func collaborationWorkspaceIDFromContext(ctx context.Context) (*uuid.UUID, bool) {
	raw := strings.TrimSpace(stringFromCtx(ctx, CtxCollaborationWorkspaceID))
	if raw == "" {
		return nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return nil, false
	}
	return &id, true
}

func clientIPFromCtx(ctx context.Context) string {
	raw, _ := ctx.Value(CtxClientIP).(string)
	if raw == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		return host
	}
	return raw
}

func centralizedLoginRequested(req *gen.LoginRequest) bool {
	return strings.TrimSpace(optString(req.TargetAppKey)) != "" ||
		strings.TrimSpace(optString(req.RedirectURI)) != "" ||
		strings.TrimSpace(optString(req.State)) != "" ||
		strings.TrimSpace(optString(req.Nonce)) != ""
}

func uuidFromMapValue(value interface{}) (uuid.UUID, error) {
	raw, ok := value.(string)
	if !ok {
		return uuid.Nil, errors.New("invalid uuid value")
	}
	return uuid.Parse(strings.TrimSpace(raw))
}

func (h *APIHandler) SilentSSOCallback(ctx context.Context, req *gen.SilentSSORequest) (gen.SilentSSOCallbackRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return nil, &apperr.UnauthError{Msg: "未认证"}
	}

	// 检查 max_age: 若请求指定了 max_age，则验证 auth_time 是否满足要求
	if req.MaxAge.Set {
		maxAge := int64(req.MaxAge.Value)
		authTime, _ := ctx.Value(CtxAuthTime).(int64)
		if authTime == 0 {
			// 旧 token 无 auth_time，视为需要重认证
			return &gen.Error{
				Code:    apperr.CodeLoginRequired,
				Message: "login_required",
			}, nil
		}
		now := time.Now().Unix()
		if now-authTime > maxAge {
			return &gen.Error{
				Code:    apperr.CodeLoginRequired,
				Message: "login_required",
			}, nil
		}
	}

	// 验证用户状态
	u, err := h.userRepo.GetByID(userID)
	if err != nil {
		return nil, &apperr.UnauthError{Msg: "用户不存在或已禁用"}
	}
	if u.Status == "disabled" {
		return nil, &apperr.UnauthError{Msg: "账号已被禁用"}
	}

	// 签发 callback code
	if h.centralizedAuthSvc == nil {
		return nil, errors.New("centralized auth service not configured")
	}
	callback, err := h.centralizedAuthSvc.CreateCallback(ctx, auth.CreateAuthCallbackInput{
		UserID:             userID,
		TargetAppKey:       req.TargetAppKey,
		RedirectURI:        req.RedirectURI,
		TargetPath:         optString(req.TargetPath),
		NavigationSpaceKey: optString(req.NavigationSpaceKey),
		State:              req.State,
		Nonce:              req.Nonce,
		RequestHost:        requestHostFromCtx(ctx),
	})
	if err != nil {
		h.logger.Warn("silent sso callback creation failed", zap.Error(err))
		return &gen.Error{
			Code:    apperr.CodeLoginRequired,
			Message: err.Error(),
		}, nil
	}

	out := &gen.LoginResponse{
		Callback: gen.OptAuthCallbackPayload{Value: gen.AuthCallbackPayload{
			Mode:                gen.AuthCallbackPayloadModeTokenExchange,
			Code:                callback.Code,
			State:               callback.State,
			TargetAppKey:        callback.TargetAppKey,
			RedirectURI:         callback.RedirectURI,
			RedirectTo:          callback.RedirectTo,
			TargetPath:          gen.NewOptString(callback.TargetPath),
			NavigationSpaceKey:  gen.NewOptString(callback.NavigationSpaceKey),
			AuthProtocolVersion: gen.NewOptString(callback.AuthProtocolVersion),
		}, Set: true},
	}
	return out, nil
}

var _ = uuid.Nil // reserved for future auth handler additions

// errorCodeOf 把 handler 返回的 error 翻译成审计可读的业务码字符串。
// 约定：成功路径传 ""；错误路径统一走 apperr.Map，避免 audit 行与响应
// 携带的 code 不一致。未命中映射（即非业务错误）时返回 CodeInternal。
func errorCodeOf(err error) string {
	if err == nil {
		return ""
	}
	_, body := apperr.Map(err)
	if body == nil || body.Code == 0 {
		return strconvItoa(apperr.CodeInternal)
	}
	return strconvItoa(body.Code)
}

// strconvItoa 封装 strconv.Itoa —— 独立出来方便将来切换成 strconv.FormatInt 之类。
func strconvItoa(n int) string {
	return strconv.Itoa(n)
}
