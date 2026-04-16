// Package middleware: openapiperm.go wires the ogen request pipeline into
// the permission evaluator. For every operation it looks up the
// x-permission-key (carried into runtime via gen-permissions seed) and
// invokes evaluator.Can. If a workspace UUID can be derived from the
// operation parameters (path "id" or query/path "workspace_id"), the check
// is workspace-scoped; otherwise the middleware passes through with a
// debug log and falls back to the evaluator account-union path.
package middleware

import (
	"context"

	"github.com/google/uuid"
	ogenmw "github.com/ogen-go/ogen/middleware"
	"go.uber.org/zap"

	apigen "github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/apperr"
	"github.com/maben/backend/internal/api/handlers"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/pkg/permission/evaluator"
)

// ErrPermissionDenied is now defined in apperr to avoid an import cycle.
// Re-exported here for callers that already reference middleware.ErrPermissionDenied.
var ErrPermissionDenied = apperr.ErrPermissionDenied

// OpenAPIPermission builds an ogen middleware that auto-enforces
// x-permission-key on every operation by calling evaluator.Can.
//
// auditRecorder 不可为 nil；关闭审计时显式传 audit.Noop{}。拒绝事件以统一的
// "system.permission.denied" action 写入 audit_logs，运维可按此 action
// 一条 SQL 查出所有越权尝试。
func OpenAPIPermission(eval evaluator.Evaluator, lookup map[string]string, logger *zap.Logger, auditRecorder audit.Recorder) apigen.Middleware {
	if auditRecorder == nil {
		auditRecorder = audit.Noop{}
	}
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
			// 把"越权尝试"单独记一条审计，便于按 permission_key / actor 聚合。
			auditRecorder.Record(req.Context, audit.Event{
				Action:       "system.permission.denied",
				ResourceType: "operation",
				ResourceID:   req.OperationID,
				Outcome:      audit.OutcomeDenied,
				Metadata: map[string]any{
					"permission_key": key,
					"workspace_id":   workspaceID.String(),
				},
			})
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

