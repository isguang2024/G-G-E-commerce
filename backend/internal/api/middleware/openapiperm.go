// Package middleware: openapiperm.go wires the ogen request pipeline into
// the permission evaluator. For every operation it looks up the
// x-permission-key (carried into runtime via gen-permissions seed) and
// invokes evaluator.Can. If a workspace UUID can be derived from the
// operation parameters (path "id" or query/path "workspace_id"), the check
// is workspace-scoped; otherwise the middleware passes through with a
// debug log so that account-level operations don't get blocked while the
// evaluator's account-only path is still TODO.
package middleware

import (
	"context"
	"errors"

	"github.com/google/uuid"
	ogenmw "github.com/ogen-go/ogen/middleware"
	"go.uber.org/zap"

	apigen "github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/handlers"
	"github.com/gg-ecommerce/backend/internal/pkg/permission/evaluator"
)

// ErrPermissionDenied is returned when the evaluator rejects an operation.
// The router maps it to a 403 JSON response via gen.WithErrorHandler so we
// don't have to teach ogen about per-operation error variants.
var ErrPermissionDenied = errors.New("openapi permission denied")

// OpenAPIPermission builds an ogen middleware that auto-enforces
// x-permission-key on every operation by calling evaluator.Can.
func OpenAPIPermission(eval evaluator.Evaluator, lookup map[string]string, logger *zap.Logger) apigen.Middleware {
	return func(req ogenmw.Request, next ogenmw.Next) (ogenmw.Response, error) {
		key, ok := lookup[req.OperationID]
		if !ok || key == "" {
			// Public op (x-access-mode: public) or unmapped operation —
			// pass through without invoking the evaluator.
			logger.Debug("openapi perm: pass-through", zap.String("op", req.OperationID))
			return next(req)
		}

		userID, ok := userIDFromCtx(req.Context)
		if !ok {
			return next(req)
		}

		// Workspace 取值三段优先级：
		// 1. 上游 gin middleware 注入的 auth_workspace_id（已校验 member 关系，最权威）
		// 2. operation path/query 中的 workspace_id / id（多见于资源级 op）
		// 3. uuid.Nil — 落到 evaluator 的账号级 union 路径（仅用于 account-only ops）
		workspaceID, _ := authWorkspaceIDFromCtx(req.Context)
		if workspaceID == uuid.Nil {
			workspaceID, _ = workspaceIDFromParams(req.Params)
		}

		allowed, err := eval.Can(req.Context, userID, workspaceID, key)
		if err != nil {
			logger.Error("openapi perm: evaluator.Can failed",
				zap.String("op", req.OperationID), zap.Error(err))
			return ogenmw.Response{}, err
		}
		if !allowed {
			logger.Info("openapi perm: denied",
				zap.String("op", req.OperationID),
				zap.String("key", key),
				zap.String("user", userID.String()),
				zap.String("workspace", workspaceID.String()))
			return ogenmw.Response{}, ErrPermissionDenied
		}
		return next(req)
	}
}

func userIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(handlers.CtxUserID).(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// authWorkspaceIDFromCtx 读取 gin middleware 已校验 member 关系后写入的 auth_workspace_id。
func authWorkspaceIDFromCtx(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(handlers.CtxAuthWorkspaceID).(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func workspaceIDFromParams(params ogenmw.Parameters) (uuid.UUID, bool) {
	for k, v := range params {
		if k.Name != "workspace_id" && k.Name != "id" {
			continue
		}
		if id, ok := v.(uuid.UUID); ok && id != uuid.Nil {
			return id, true
		}
	}
	return uuid.Nil, false
}
