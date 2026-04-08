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
	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
)

// CtxClientIP carries the originating client IP from the gin bridge into
// ogen handlers (used by Login for last-login bookkeeping).
const CtxClientIP ctxKey = "client_ip"

func (h *APIHandler) Login(ctx context.Context, req *gen.LoginRequest) (gen.LoginRes, error) {
	if req == nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return &gen.LoginBadRequest{Code: 400, Message: "用户名和密码必填"}, nil
	}
	ip := clientIPFromCtx(ctx)
	resp, err := h.authSvc.Login(req.Username, req.Password, ip)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrUserInactive) {
			return &gen.LoginUnauthorized{Code: 401, Message: err.Error()}, nil
		}
		h.logger.Error("login failed", zap.Error(err))
		return nil, err
	}
	out := &gen.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}
	if userMap, ok := resp.User.(map[string]interface{}); ok {
		out.User = gen.NewOptNilLoginResponseUser(toJxRawMap(userMap))
	}
	return out, nil
}

func (h *APIHandler) Register(ctx context.Context, req *gen.RegisterRequest) (gen.RegisterRes, error) {
	if req == nil || strings.TrimSpace(req.Username) == "" || req.Password == "" {
		return &gen.Error{Code: 400, Message:"用户名和密码必填"}, nil
	}
	dtoReq := &dto.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    optString(req.Email),
		Nickname: optString(req.Nickname),
	}
	resp, err := h.authSvc.Register(dtoReq)
	if err != nil {
		if errors.Is(err, auth.ErrUserExists) || errors.Is(err, auth.ErrEmailExists) {
			return &gen.Error{Code: 400, Message:err.Error()}, nil
		}
		h.logger.Error("register failed", zap.Error(err))
		return nil, err
	}
	out := &gen.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}
	if userMap, ok := resp.User.(map[string]interface{}); ok {
		out.User = gen.NewOptNilLoginResponseUser(toJxRawMap(userMap))
	}
	return out, nil
}

func (h *APIHandler) RefreshToken(ctx context.Context, req *gen.RefreshTokenRequest) (gen.RefreshTokenRes, error) {
	if req == nil || strings.TrimSpace(req.RefreshToken) == "" {
		return &gen.Error{Code: 401, Message: "缺少 refresh_token"}, nil
	}
	resp, err := h.authSvc.RefreshToken(req.RefreshToken)
	if err != nil {
		return &gen.Error{Code: 401, Message: "无效或过期的 refresh_token"}, nil
	}
	return &gen.TokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
	}, nil
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
		return &gen.Error{Code: 401, Message: "未认证"}, nil
	}
	u, err := h.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Warn("auth.me user not found", zap.String("user_id", userID.String()))
			return &gen.Error{Code: 401, Message: "登录状态已失效，请重新登录"}, nil
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

var _ = uuid.Nil // reserved for future auth handler additions
