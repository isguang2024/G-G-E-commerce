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

func optString(o gen.OptString) string {
	if !o.Set {
		return ""
	}
	return o.Value
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
		h.logger.Error("auth.me lookup failed", zap.Error(err))
		return nil, err
	}
	out := &gen.AuthMe{
		ID:       u.ID,
		Username: u.Username,
	}
	if u.Nickname != "" {
		out.Nickname = gen.NewOptNilString(u.Nickname)
	}
	if u.Email != "" {
		out.Email = gen.NewOptNilString(u.Email)
	}
	if u.AvatarURL != "" {
		out.Avatar = gen.NewOptNilString(u.AvatarURL)
	}
	return out, nil
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
