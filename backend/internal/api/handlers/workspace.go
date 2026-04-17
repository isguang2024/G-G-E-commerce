// Package handlers contains ogen Handler implementations for the v5
// OpenAPI-first API. Each handler is the single entry point for one
// generated operation interface; legacy Gin handlers are removed as
// each domain migrates over.
package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/logpolicy"
	"github.com/maben/backend/internal/modules/observability/telemetry"
	"github.com/maben/backend/internal/modules/system/apiendpoint"
	"github.com/maben/backend/internal/modules/system/app"
	"github.com/maben/backend/internal/modules/system/auth"
	"github.com/maben/backend/internal/modules/system/collaborationworkspace"
	"github.com/maben/backend/internal/modules/system/featurepackage"
	"github.com/maben/backend/internal/modules/system/menu"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/navigation"
	"github.com/maben/backend/internal/modules/system/page"
	"github.com/maben/backend/internal/modules/system/permission"
	"github.com/maben/backend/internal/modules/system/register"
	"github.com/maben/backend/internal/modules/system/role"
	"github.com/maben/backend/internal/modules/system/siteconfig"
	"github.com/maben/backend/internal/modules/system/social"
	"github.com/maben/backend/internal/modules/system/space"
	systemmod "github.com/maben/backend/internal/modules/system/system"
	"github.com/maben/backend/internal/modules/system/upload"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/modules/system/workspace"
	"github.com/maben/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/maben/backend/internal/pkg/permission/evaluator"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
	"github.com/maben/backend/internal/pkg/platformaccess"
	"github.com/maben/backend/internal/pkg/platformroleaccess"
)

// ctxKey is the request-scoped key carrying the authenticated account id
// from the Gin layer into the ogen handler. The router middleware seeds it
// before handing the request to the generated server.
type ctxKey string

const (
	CtxUserID            ctxKey = "user_id"
	CtxAuthWorkspaceID   ctxKey = "auth_workspace_id"
	CtxAuthWorkspaceType ctxKey = "auth_workspace_type"
	CtxCollaborationID   ctxKey = "collaboration_id"
	CtxAuthTime          ctxKey = "auth_time"
)

// handlerBase 把 gen.UnimplementedHandler 下沉一层，让 APIHandler 嵌入的
// 域专属 sub-handler（如 *dictionaryAPIHandler，位于深度 1）能覆盖
// base 里的 stub（位于深度 2）。这样既保持 ogen 新增 op 不强制立刻实现
// 的便利，又允许按域拆分而不触发方法歧义。
type handlerBase struct {
	gen.UnimplementedHandler
}

// APIHandler implements gen.Handler。
//
// 域拆分完成后 APIHandler 只保留：
//   - 工作空间域方法（GetWorkspace / ListMyWorkspaces / GetCurrentWorkspace /
//     SwitchWorkspace / ExplainPermissions）
//   - 构造函数 NewAPIHandler
//   - 各域 sub-handler 的匿名嵌入
//
// 每个域 sub-handler 通过 *{domain}APIHandler 匿名嵌入接管自己的操作；
// 未迁移或新增的操作由 handlerBase.UnimplementedHandler 兜底。
type APIHandler struct {
	handlerBase
	// 工作空间域保留字段
	logger    *zap.Logger
	service   workspace.Service
	evaluator evaluator.Evaluator
	// ── 域 sub-handler 嵌入 ────────────────────────────────────────────────
	*navigationAPIHandler
	*dictionaryAPIHandler
	*authAPIHandler
	*userAPIHandler
	*roleAPIHandler
	*menuAPIHandler
	*pageAPIHandler
	*permissionAPIHandler
	*featurePackageAPIHandler
	*cwAPIHandler
	*extrasAPIHandler
	*phase4ExtrasAPIHandler
	*systemAPIHandler
	*apiEndpointAPIHandler
	*mediaAPIHandler
	*storageAdminAPIHandler
	*siteConfigAPIHandler
	*logPolicyAPIHandler
	*observabilityAPIHandler
	*telemetryAPIHandler
	*messageAPIHandler
}

// NewAPIHandler 构建统一的 v5 API handler。
//
// auditRecorder / telemetryIngester 由上层 router 注入；传入 Noop 可关闭相应能力。
// 不允许 nil —— handler 内部直接调用，避免每个操作都判空。
func NewAPIHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger, eval evaluator.Evaluator, apiEndpointSvc apiendpoint.Service, auditRecorder audit.Recorder, telemetryIngester telemetry.Ingester) *APIHandler {
	if auditRecorder == nil {
		auditRecorder = audit.Noop{}
	}
	if telemetryIngester == nil {
		telemetryIngester = telemetry.Noop{}
	}
	var policyRepo logpolicy.Repository
	var policyEngine logpolicy.Engine
	if db != nil {
		repo := logpolicy.NewRepository(db)
		policyRepo = repo
		policyEngine = logpolicy.NewEngine(repo, logger)
		if err := policyEngine.Refresh(context.Background()); err != nil {
			logger.Warn("logpolicy.initial_refresh_failed", zap.Error(err))
		}
	}
	// ── repos ──────────────────────────────────────────────────────────────
	userRepo := user.NewUserRepository(db)
	roleRepo := user.NewRoleRepository(db)
	menuRepo := user.NewMenuRepository(db)
	cwRepo := user.NewCollaborationWorkspaceRepository(db)
	cwMemberRepo := user.NewCollaborationWorkspaceMemberRepository(db)
	userRoleRepo := user.NewUserRoleRepository(db)
	apiEndpointRepo := user.NewAPIEndpointRepository(db)
	apiEndpointBindingRepo := user.NewAPIEndpointPermissionBindingRepository(db)
	featurePkgRepo := user.NewFeaturePackageRepository(db)
	featurePkgBundleRepo := user.NewFeaturePackageBundleRepository(db)
	featurePkgKeyRepo := user.NewFeaturePackageKeyRepository(db)
	featurePkgMenuRepo := user.NewFeaturePackageMenuRepository(db)
	cwFeaturePkgRepo := user.NewCollaborationWorkspaceFeaturePackageRepository(db)
	rolePackageRepo := user.NewRoleFeaturePackageRepository(db)
	permGroupRepo := user.NewPermissionGroupRepository(db)
	keyRepo := user.NewPermissionKeyRepository(db)
	roleHiddenMenuRepo := user.NewRoleHiddenMenuRepository(db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(db)
	roleDataRepo := user.NewRoleDataPermissionRepository(db)
	cwBlockedActionRepo := user.NewCollaborationWorkspaceBlockedActionRepository(db)
	userActionRepo := user.NewUserActionPermissionRepository(db)

	// ── infrastructure services ────────────────────────────────────────────
	boundarySvc := collaborationworkspaceboundary.NewService(db)
	personalAccess := platformaccess.NewService(db)
	roleSnapshot := platformroleaccess.NewService(db)
	refresher := permissionrefresh.NewService(db, boundarySvc, personalAccess, roleSnapshot)

	// ── domain services ────────────────────────────────────────────────────
	userSvc := user.NewUserService(db, userRepo, roleRepo, refresher, logger)

	roleSvc := role.NewRoleService(
		db,
		roleRepo,
		rolePackageRepo,
		featurePkgRepo,
		featurePkgKeyRepo,
		featurePkgMenuRepo,
		featurePkgBundleRepo,
		roleHiddenMenuRepo,
		roleDisabledActionRepo,
		roleDataRepo,
		keyRepo,
		roleSnapshot,
		refresher,
		logger,
	)

	appSvc := app.NewService(db)

	menuSvc := menu.NewMenuService(db, menuRepo, refresher, logger)

	pageSvc := page.NewService(db, menuRepo)

	spaceSvc := space.NewService(db, refresher, logger)

	navSvc := navigation.NewService(db, appSvc, menuSvc, pageSvc, spaceSvc)

	featurePkgSvc := featurepackage.NewService(
		db,
		featurePkgRepo,
		featurePkgBundleRepo,
		featurePkgKeyRepo,
		featurePkgMenuRepo,
		cwFeaturePkgRepo,
		rolePackageRepo,
		keyRepo,
		menuRepo,
		cwRepo,
		boundarySvc,
		refresher,
	)

	permSvc := permission.NewPermissionService(
		db,
		permGroupRepo,
		keyRepo,
		apiEndpointRepo,
		apiEndpointBindingRepo,
		featurePkgKeyRepo,
		cwFeaturePkgRepo,
		roleDisabledActionRepo,
		cwBlockedActionRepo,
		userActionRepo,
		boundarySvc,
		refresher,
	)

	cwSvc := collaborationworkspace.NewCollaborationWorkspaceService(
		db,
		cwRepo,
		cwMemberRepo,
		userRepo,
		roleRepo,
		userRoleRepo,
		refresher,
		logger,
	)

	workspaceSvc := workspace.NewService(db, logger)

	authSvc := auth.NewAuthService(userRepo, &cfg.JWT, logger)
	centralizedAuthSvc := auth.NewCentralizedAuthService(db, authSvc, userRepo)

	registerResolver := register.NewResolver(register.NewRepository(db))
	registerSvc := register.NewService(db, registerResolver, authSvc, workspaceSvc, logger)
	socialSvc := social.NewService(db, authSvc, userRepo, registerResolver, cfg.JWT.Secret, logger)

	uploadSvc := upload.NewService(upload.NewRepository(db).WithAuditRecorder(auditRecorder), cfg, logger)

	siteConfigSvc := siteconfig.NewService(siteconfig.NewRepository(db), cfg, logger)

	systemFacade := systemmod.NewFacade(db, logger, nil)

	h := &APIHandler{
		logger:    logger,
		service:   workspaceSvc,
		evaluator: eval,
		// ── sub-handlers ──────────────────────────────────────────────────
		navigationAPIHandler: newNavigationAPIHandler(navSvc, logger),
		authAPIHandler: newAuthAPIHandler(
			authSvc,
			centralizedAuthSvc,
			userRepo,
			socialSvc,
			registerSvc,
			registerResolver,
			eval,
			workspaceSvc,
			auditRecorder,
			db,
			logger,
		),
		userAPIHandler: newUserAPIHandler(
			userSvc,
			userRepo,
			boundarySvc,
			menuSvc,
			featurePkgRepo,
			personalAccess,
			refresher,
			auditRecorder,
			db,
			logger,
		),
		roleAPIHandler:           newRoleAPIHandler(roleSvc, logger),
		menuAPIHandler:           newMenuAPIHandler(menuSvc, auditRecorder, logger),
		pageAPIHandler:           newPageAPIHandler(pageSvc, auditRecorder, logger),
		permissionAPIHandler:     newPermissionAPIHandler(permSvc, logger),
		featurePackageAPIHandler: newFeaturePackageAPIHandler(featurePkgSvc, logger),
		cwAPIHandler: newCWAPIHandler(
			cwSvc,
			featurePkgSvc,
			boundarySvc,
			cwMemberRepo,
			roleRepo,
			userRoleRepo,
			featurePkgRepo,
			cwFeaturePkgRepo,
			keyRepo,
			db,
			logger,
		),
		extrasAPIHandler: newExtrasAPIHandler(
			featurePkgSvc,
			menuSvc,
			pageSvc,
			permSvc,
			logger,
		),
		phase4ExtrasAPIHandler: newPhase4ExtrasAPIHandler(
			systemmod.NewFastEnterService(db),
			userSvc,
			systemmod.NewViewPagesService(logger, nil),
			cwMemberRepo,
			refresher,
			logger,
		),
		systemAPIHandler:      newSystemAPIHandler(appSvc, spaceSvc, auditRecorder, db, logger),
		apiEndpointAPIHandler: newAPIEndpointAPIHandler(apiEndpointSvc, logger),
		mediaAPIHandler:       newMediaAPIHandler(uploadSvc, eval, logger),
		storageAdminAPIHandler: newStorageAdminAPIHandler(uploadSvc, logger),
		siteConfigAPIHandler:  newSiteConfigAPIHandler(siteConfigSvc, auditRecorder, logger),
		logPolicyAPIHandler:   newLogPolicyAPIHandler(policyRepo, policyEngine, auditRecorder, logger),
		observabilityAPIHandler: newObservabilityAPIHandler(
			db,
			auditRecorder,
			telemetryIngester,
			logger,
		),
		telemetryAPIHandler: newTelemetryAPIHandler(telemetryIngester, logger),
		messageAPIHandler:   newMessageAPIHandler(systemFacade, auditRecorder, logger),
	}
	h.dictionaryAPIHandler = newDictionaryAPIHandler(db, logger)
	if cache := siteConfigSvc.Cache(); cache != nil {
		cache.Start(context.Background())
	}
	return h
}

func (h *APIHandler) GetWorkspace(ctx context.Context, params gen.GetWorkspaceParams) (gen.GetWorkspaceRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.GetWorkspaceForbidden{Code: 401, Message: "未认证"}, nil
	}

	if _, err := h.service.GetMember(params.ID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.GetWorkspaceForbidden{Code: 403, Message: "无权访问该工作空间"}, nil
		}
		h.logger.Error("workspace member lookup failed", zap.Error(err))
		return nil, err
	}

	ws, err := h.service.GetByID(params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.GetWorkspaceNotFound{Code: 404, Message: "工作空间不存在"}, nil
		}
		h.logger.Error("workspace lookup failed", zap.Error(err))
		return nil, err
	}

	return mapWorkspaceToSummary(ws), nil
}

func (h *APIHandler) ListMyWorkspaces(ctx context.Context) (gen.ListMyWorkspacesRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.Error{Code: 401, Message: "未认证"}, nil
	}
	items, err := h.service.ListByUserID(userID)
	if err != nil {
		h.logger.Error("list workspaces failed", zap.Error(err))
		return nil, err
	}
	records := make([]gen.WorkspaceSummary, 0, len(items))
	for _, item := range items {
		records = append(records, summaryFromService(item))
	}
	return &gen.WorkspaceList{Records: records, Total: len(records)}, nil
}

func (h *APIHandler) GetCurrentWorkspace(ctx context.Context) (gen.GetCurrentWorkspaceRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.Error{Code: 401, Message: "未认证"}, nil
	}
	// 优先返回 gin middleware 已校验过 member 关系的 auth_workspace_id；
	// 没有再回落 personal workspace，避免 /workspaces/switch 后刷新页面
	// 拿回 personal 与权限实际生效空间错位。
	if raw := strings.TrimSpace(stringFromCtx(ctx, CtxAuthWorkspaceID)); raw != "" {
		if authWsID, parseErr := uuid.Parse(raw); parseErr == nil {
			if _, err := h.service.GetMember(authWsID, userID); err == nil {
				if ws, err := h.service.GetByID(authWsID); err == nil {
					return mapWorkspaceToSummary(ws), nil
				}
			}
		}
	}
	ws, err := h.service.EnsurePersonalWorkspaceForUser(userID)
	if err != nil {
		h.logger.Error("get current workspace failed", zap.Error(err))
		return nil, err
	}
	return mapWorkspaceToSummary(ws), nil
}

func (h *APIHandler) SwitchWorkspace(ctx context.Context, req *gen.WorkspaceSwitchRequest) (gen.SwitchWorkspaceRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.SwitchWorkspaceForbidden{Code: 401, Message: "未认证"}, nil
	}
	if req == nil || req.WorkspaceID == uuid.Nil {
		return &gen.SwitchWorkspaceBadRequest{Code: 400, Message: "无效的工作空间ID"}, nil
	}
	if _, err := h.service.GetMember(req.WorkspaceID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.SwitchWorkspaceForbidden{Code: 403, Message: "无权切换到该工作空间"}, nil
		}
		h.logger.Error("workspace member lookup failed", zap.Error(err))
		return nil, err
	}
	ws, err := h.service.GetByID(req.WorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.SwitchWorkspaceForbidden{Code: 403, Message: "工作空间不存在"}, nil
		}
		h.logger.Error("workspace lookup failed", zap.Error(err))
		return nil, err
	}
	out := &gen.WorkspaceSwitchResponse{
		AuthWorkspaceID:   ws.ID,
		AuthWorkspaceType: gen.WorkspaceSwitchResponseAuthWorkspaceType(ws.WorkspaceType),
		Workspace:         *mapWorkspaceToSummary(ws),
	}
	if ws.CollaborationWorkspaceID != nil {
		out.WorkspaceID = gen.NewOptNilUUID(*ws.CollaborationWorkspaceID)
	}
	return out, nil
}

func (h *APIHandler) ExplainPermissions(ctx context.Context, params gen.ExplainPermissionsParams) (gen.ExplainPermissionsRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.Error{Code: 401, Message: "未认证"}, nil
	}
	if h.evaluator == nil {
		return &gen.Error{Code: 500, Message: "evaluator 未初始化"}, nil
	}
	exp, err := h.evaluator.Explain(ctx, userID, params.WorkspaceID)
	if err != nil {
		h.logger.Error("explain permissions failed", zap.Error(err))
		return nil, err
	}
	keys := make([]string, 0, len(exp.Resolved.Keys))
	for k := range exp.Resolved.Keys {
		keys = append(keys, k)
	}
	out := &gen.PermissionExplanation{
		AccountID:   exp.Resolved.AccountID,
		WorkspaceID: exp.Resolved.WorkspaceID,
		Keys:        keys,
	}
	if len(exp.FeaturePackageKeys) > 0 {
		fps := gen.PermissionExplanationFeaturePackageSources{}
		for k, ids := range exp.FeaturePackageKeys {
			fps[k] = ids
		}
		out.SetFeaturePackageSources(gen.NewOptPermissionExplanationFeaturePackageSources(fps))
	}
	if len(exp.RoleKeys) > 0 {
		rs := gen.PermissionExplanationRoleSources{}
		for k, ids := range exp.RoleKeys {
			rs[k] = ids
		}
		out.SetRoleSources(gen.NewOptPermissionExplanationRoleSources(rs))
	}
	return out, nil
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(CtxUserID).(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func mapWorkspaceToSummary(ws *models.Workspace) *gen.WorkspaceSummary {
	out := &gen.WorkspaceSummary{
		ID:            ws.ID,
		WorkspaceType: gen.WorkspaceSummaryWorkspaceType(ws.WorkspaceType),
		Name:          ws.Name,
		Code:          ws.Code,
		Status:        ws.Status,
	}
	if ws.OwnerUserID != nil {
		out.OwnerUserID = gen.NewOptNilUUID(*ws.OwnerUserID)
	}
	if ws.CollaborationWorkspaceID != nil {
		out.WorkspaceID = gen.NewOptNilUUID(*ws.CollaborationWorkspaceID)
	}
	return out
}

func summaryFromService(item workspace.Summary) gen.WorkspaceSummary {
	out := gen.WorkspaceSummary{
		ID:            item.ID,
		WorkspaceType: gen.WorkspaceSummaryWorkspaceType(item.WorkspaceType),
		Name:          item.Name,
		Code:          item.Code,
		Status:        item.Status,
	}
	if item.OwnerUserID != nil {
		out.OwnerUserID = gen.NewOptNilUUID(*item.OwnerUserID)
	}
	if item.CollaborationWorkspaceID != nil {
		out.WorkspaceID = gen.NewOptNilUUID(*item.CollaborationWorkspaceID)
	}
	return out
}

