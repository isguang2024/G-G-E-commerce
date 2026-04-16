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
	"github.com/maben/backend/internal/pkg/permissionseed"
)

// ErrPermissionDenied is now defined in apperr to avoid an import cycle.
// Re-exported here for callers that already reference middleware.ErrPermissionDenied.
var ErrPermissionDenied = apperr.ErrPermissionDenied

// OpenAPIPermission builds an ogen middleware that auto-enforces
// x-permission-key on every operation by calling evaluator.Can.
//
// 权限键解析主路径：operation_id → endpoint_code → permission_key。
// endpoint_code 由 StableID("openapi-api-endpoint", METHOD+" "+path) 派生，
// 与 api_endpoints.code / api_endpoint_permission_bindings.endpoint_code 对齐，
// 这样即便 operation_id 改名，只要 METHOD+path 不变就不影响权限命中。
// operation_id 直查 permission_key 是 deprecated fallback，仅在主路径漏时触发
// 并记一条 warn 日志，便于治理。
//
// auditRecorder 不可为 nil；关闭审计时显式传 audit.Noop{}。拒绝事件以统一的
// "system.permission.denied" action 写入 audit_logs，运维可按此 action
// 一条 SQL 查出所有越权尝试。
func OpenAPIPermission(eval evaluator.Evaluator, lookup *permissionseed.PermissionLookup, logger *zap.Logger, auditRecorder audit.Recorder) apigen.Middleware {
	if auditRecorder == nil {
		auditRecorder = audit.Noop{}
	}
	if lookup == nil {
		lookup = &permissionseed.PermissionLookup{}
	}
	return func(req ogenmw.Request, next ogenmw.Next) (ogenmw.Response, error) {
		key, endpointCode, viaFallback := resolvePermissionKey(lookup, req.OperationID)
		if key == "" {
			// Public op (x-access-mode: public) or unmapped operation —
			// pass through without invoking the evaluator.
			logger.Debug("openapi perm: pass-through",
				zap.String("op", req.OperationID),
				zap.String("endpoint_code", endpointCode))
			return next(req)
		}
		if viaFallback {
			// 主路径（endpoint_code）未命中但 operation_id 命中 —— 通常意味着
			// api_endpoint_permission_bindings 与 seed 不同步。记 warn 方便
			// 被运维监控捕捉；但本次请求仍然允许放行，降级不失控。
			logger.Warn("openapi perm: deprecated operationID fallback hit; check bindings vs seed",
				zap.String("op", req.OperationID),
				zap.String("endpoint_code", endpointCode),
				zap.String("key", key))
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
				zap.String("op", req.OperationID),
				zap.String("endpoint_code", endpointCode),
				zap.Error(err))
			return ogenmw.Response{}, err
		}
		if !allowed {
			logger.Info("openapi perm: denied",
				zap.String("op", req.OperationID),
				zap.String("endpoint_code", endpointCode),
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
					"endpoint_code":  endpointCode,
					"workspace_id":   workspaceID.String(),
				},
			})
			return ogenmw.Response{}, ErrPermissionDenied
		}
		return next(req)
	}
}

// resolvePermissionKey 主路径走 endpoint_code；operation_id 直查只做 deprecated
// fallback。返回：permission_key、命中的 endpoint_code（便于日志/审计）、
// 是否走了 fallback（便于 middleware 记 warn）。
func resolvePermissionKey(lookup *permissionseed.PermissionLookup, operationID string) (key, endpointCode string, viaFallback bool) {
	endpointCode = lookup.OperationToEndpointCode[operationID]
	if endpointCode != "" {
		if k := lookup.ByEndpointCode[endpointCode]; k != "" {
			return k, endpointCode, false
		}
	}
	if k := lookup.ByOperationIDDeprecated[operationID]; k != "" {
		return k, endpointCode, true
	}
	return "", endpointCode, false
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

