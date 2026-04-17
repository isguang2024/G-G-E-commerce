package handlers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/collaborationworkspace"
	"github.com/maben/backend/internal/modules/system/featurepackage"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/collaborationworkspaceboundary"
)

// cwAPIHandler 负责协作工作区（collaboration workspace）域的所有操作：
// 成员、角色、功能包绑定、边界控制等。
// 内含 CW 域所有文件共用的 helper 方法（resolveCWMember 等）。
type cwAPIHandler struct {
	cwSvc            collaborationworkspace.CollaborationWorkspaceService
	featurePkgSvc    featurepackage.Service
	boundarySvc      collaborationworkspaceboundary.Service
	cwMemberRepo     user.CollaborationWorkspaceMemberRepository
	roleRepo         user.RoleRepository
	userRoleRepo     user.UserRoleRepository
	featurePkgRepo   user.FeaturePackageRepository
	cwFeaturePkgRepo user.CollaborationWorkspaceFeaturePackageRepository
	keyRepo          user.PermissionKeyRepository
	db               *gorm.DB
	logger           *zap.Logger
}

func newCWAPIHandler(
	cwSvc collaborationworkspace.CollaborationWorkspaceService,
	featurePkgSvc featurepackage.Service,
	boundarySvc collaborationworkspaceboundary.Service,
	cwMemberRepo user.CollaborationWorkspaceMemberRepository,
	roleRepo user.RoleRepository,
	userRoleRepo user.UserRoleRepository,
	featurePkgRepo user.FeaturePackageRepository,
	cwFeaturePkgRepo user.CollaborationWorkspaceFeaturePackageRepository,
	keyRepo user.PermissionKeyRepository,
	db *gorm.DB,
	logger *zap.Logger,
) *cwAPIHandler {
	return &cwAPIHandler{
		cwSvc:            cwSvc,
		featurePkgSvc:    featurePkgSvc,
		boundarySvc:      boundarySvc,
		cwMemberRepo:     cwMemberRepo,
		roleRepo:         roleRepo,
		userRoleRepo:     userRoleRepo,
		featurePkgRepo:   featurePkgRepo,
		cwFeaturePkgRepo: cwFeaturePkgRepo,
		keyRepo:          keyRepo,
		db:               db,
		logger:           logger,
	}
}

