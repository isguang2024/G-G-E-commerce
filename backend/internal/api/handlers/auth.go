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
	"strings"

	"github.com/go-faster/jx"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/apperr"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
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
	ip := clientIPFromCtx(ctx)
	resp, err := h.authSvc.Login(req.Username, req.Password, ip)
	if err != nil {
		h.logger.Debug("login failed", zap.Error(err))
		return nil, err
	}
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
		Username:         req.Username,
		Password:         req.Password,
		ConfirmPassword:  optString(req.ConfirmPassword),
		Email:            optString(req.Email),
		Nickname:         optString(req.Nickname),
		CaptchaToken:     optString(req.CaptchaToken),
		InvitationCode:   optString(req.InvitationCode),
		AgreementVersion: optString(req.AgreementVersion),
		Host:             stringFromCtx(ctx, CtxRequestHost),
		Path:             stringFromCtx(ctx, CtxRequestPath),
		IP:               clientIPFromCtx(ctx),
		UserAgent:        stringFromCtx(ctx, CtxUserAgent),
	})
	if err != nil {
		h.logger.Debug("register failed", zap.Error(err))
		if errors.Is(err, register.ErrPublicRegisterDisabled) {
			return nil, &apperr.ParamError{Msg: "公开注册未开启"}
		}
		return nil, err
	}
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
			AppKey:             gen.NewOptString(result.Landing.AppKey),
			NavigationSpaceKey: gen.NewOptString(result.Landing.NavigationSpaceKey),
			HomePath:           gen.NewOptString(result.Landing.HomePath),
		})
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
		EntryCode:                eff.EntryCode,
		EntryAppKey:              eff.EntryAppKey,
		PolicyCode:               eff.PolicyCode,
		TargetAppKey:             eff.TargetAppKey,
		TargetNavigationSpaceKey: eff.TargetNavigationSpaceKey,
		TargetHomePath:           eff.TargetHomePath,
		AllowPublicRegister:      eff.AllowPublicRegister,
		RequireInvite:            eff.RequireInvite,
		RequireEmailVerify:       eff.RequireEmailVerify,
		RequireCaptcha:           eff.RequireCaptcha,
		AutoLogin:                eff.AutoLogin,
	}
	if eff.EntryName != "" {
		out.EntryName = gen.NewOptString(eff.EntryName)
	}
	if eff.RegisterSource != "" {
		out.RegisterSource = gen.NewOptString(eff.RegisterSource)
	}
	if eff.DefaultWorkspaceType != "" {
		out.DefaultWorkspaceType = gen.NewOptString(eff.DefaultWorkspaceType)
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

var _ = uuid.Nil // reserved for future auth handler additions
