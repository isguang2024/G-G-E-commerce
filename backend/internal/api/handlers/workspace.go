// Package handlers contains ogen Handler implementations for the v5
// OpenAPI-first API. Each handler is the single entry point for one
// generated operation interface; legacy Gin handlers are removed as
// each domain migrates over.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/config"
	"github.com/gg-ecommerce/backend/internal/modules/system/app"
	"github.com/gg-ecommerce/backend/internal/modules/system/auth"
	"github.com/gg-ecommerce/backend/internal/modules/system/collaborationworkspace"
	"github.com/gg-ecommerce/backend/internal/modules/system/featurepackage"
	"github.com/gg-ecommerce/backend/internal/modules/system/menu"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/navigation"
	"github.com/gg-ecommerce/backend/internal/modules/system/page"
	"github.com/gg-ecommerce/backend/internal/modules/system/permission"
	"github.com/gg-ecommerce/backend/internal/modules/system/role"
	"github.com/gg-ecommerce/backend/internal/modules/system/space"
	systemmod "github.com/gg-ecommerce/backend/internal/modules/system/system"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/modules/system/workspace"
	"github.com/gg-ecommerce/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/gg-ecommerce/backend/internal/pkg/permission/evaluator"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
)

// ctxKey is the request-scoped key carrying the authenticated account id
// from the Gin layer into the ogen handler. The router middleware seeds it
// before handing the request to the generated server.
type ctxKey string

const (
	CtxUserID                   ctxKey = "user_id"
	CtxAuthWorkspaceID          ctxKey = "auth_workspace_id"
	CtxAuthWorkspaceType        ctxKey = "auth_workspace_type"
	CtxCollaborationWorkspaceID ctxKey = "collaboration_workspace_id"
)

// APIHandler implements gen.Handler. It deliberately embeds
// gen.UnimplementedHandler so future operations compile without forcing us
// to stub every method while migrating one domain at a time.
type APIHandler struct {
	gen.UnimplementedHandler
	db            *gorm.DB
	logger        *zap.Logger
	service       workspace.Service
	evaluator   evaluator.Evaluator
	authSvc     auth.AuthService
	userRepo    user.UserRepository
	userSvc     user.UserService
	roleSvc     role.RoleService
	navSvc      navigation.Compiler
	menuSvc     menu.MenuService
	pageSvc     page.Service
	featurePkgSvc  featurepackage.Service
	featurePkgRepo user.FeaturePackageRepository
	permSvc        permission.PermissionService
	cwSvc       collaborationworkspace.CollaborationWorkspaceService
	appSvc      app.Service
	spaceSvc    space.Service
	boundarySvc   collaborationworkspaceboundary.Service
	personalAccess platformaccess.Service
	cwMemberRepo  user.CollaborationWorkspaceMemberRepository
	systemFacade *systemmod.Facade
	refresher    permissionrefresh.Service
}

func NewAPIHandler(db *gorm.DB, cfg *config.Config, logger *zap.Logger, eval evaluator.Evaluator) *APIHandler {
	// ── repos ──────────────────────────────────────────────────────────────
	userRepo              := user.NewUserRepository(db)
	roleRepo              := user.NewRoleRepository(db)
	menuRepo              := user.NewMenuRepository(db)
	cwRepo                := user.NewCollaborationWorkspaceRepository(db)
	cwMemberRepo          := user.NewCollaborationWorkspaceMemberRepository(db)
	userRoleRepo          := user.NewUserRoleRepository(db)
	apiEndpointRepo       := user.NewAPIEndpointRepository(db)
	apiEndpointBindingRepo := user.NewAPIEndpointPermissionBindingRepository(db)
	featurePkgRepo        := user.NewFeaturePackageRepository(db)
	featurePkgBundleRepo  := user.NewFeaturePackageBundleRepository(db)
	featurePkgKeyRepo     := user.NewFeaturePackageKeyRepository(db)
	featurePkgMenuRepo    := user.NewFeaturePackageMenuRepository(db)
	cwFeaturePkgRepo      := user.NewCollaborationWorkspaceFeaturePackageRepository(db)
	rolePackageRepo       := user.NewRoleFeaturePackageRepository(db)
	permGroupRepo         := user.NewPermissionGroupRepository(db)
	keyRepo               := user.NewPermissionKeyRepository(db)
	roleHiddenMenuRepo    := user.NewRoleHiddenMenuRepository(db)
	roleDisabledActionRepo := user.NewRoleDisabledActionRepository(db)
	roleDataRepo          := user.NewRoleDataPermissionRepository(db)
	cwBlockedActionRepo   := user.NewCollaborationWorkspaceBlockedActionRepository(db)
	userActionRepo        := user.NewUserActionPermissionRepository(db)

	// ── infrastructure services ────────────────────────────────────────────
	boundarySvc    := collaborationworkspaceboundary.NewService(db)
	personalAccess := platformaccess.NewService(db)
	roleSnapshot   := platformroleaccess.NewService(db)
	refresher      := permissionrefresh.NewService(db, boundarySvc, personalAccess, roleSnapshot)

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

	return &APIHandler{
		db:            db,
		logger:        logger,
		service:       workspace.NewService(db, logger),
		evaluator:     eval,
		authSvc:       auth.NewAuthService(userRepo, &cfg.JWT, logger),
		userRepo:      userRepo,
		userSvc:       userSvc,
		roleSvc:       roleSvc,
		navSvc:        navSvc,
		menuSvc:       menuSvc,
		pageSvc:       pageSvc,
		featurePkgSvc:  featurePkgSvc,
		featurePkgRepo: featurePkgRepo,
		permSvc:        permSvc,
		cwSvc:         cwSvc,
		appSvc:        appSvc,
		spaceSvc:      spaceSvc,
		boundarySvc:    boundarySvc,
		personalAccess: personalAccess,
		cwMemberRepo:   cwMemberRepo,
		systemFacade:  systemmod.NewFacade(db, logger, nil),
		refresher:     refresher,
	}
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
	// Phase 4 follow-up: respect the auth_workspace_id carried in the JWT
	// claims. For now we fall back to the user's personal workspace, which
	// matches the legacy behaviour for the common case.
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
		out.CollaborationWorkspaceID = gen.NewOptNilUUID(*ws.CollaborationWorkspaceID)
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
		out.CollaborationWorkspaceID = gen.NewOptNilUUID(*ws.CollaborationWorkspaceID)
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
		out.CollaborationWorkspaceID = gen.NewOptNilUUID(*item.CollaborationWorkspaceID)
	}
	return out
}
