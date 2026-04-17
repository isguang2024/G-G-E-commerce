package handlers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/menu"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/collaborationworkspaceboundary"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
	"github.com/maben/backend/internal/pkg/platformaccess"
)

// userAPIHandler 负责 /users/* 及用户子路由操作。
type userAPIHandler struct {
	userSvc        user.UserService
	userRepo       user.UserRepository
	boundarySvc    collaborationworkspaceboundary.Service
	menuSvc        menu.MenuService
	featurePkgRepo user.FeaturePackageRepository
	personalAccess platformaccess.Service
	refresher      permissionrefresh.Service
	audit          audit.Recorder
	db             *gorm.DB
	logger         *zap.Logger
}

func newUserAPIHandler(
	userSvc user.UserService,
	userRepo user.UserRepository,
	boundarySvc collaborationworkspaceboundary.Service,
	menuSvc menu.MenuService,
	featurePkgRepo user.FeaturePackageRepository,
	personalAccess platformaccess.Service,
	refresher permissionrefresh.Service,
	auditRecorder audit.Recorder,
	db *gorm.DB,
	logger *zap.Logger,
) *userAPIHandler {
	return &userAPIHandler{
		userSvc:        userSvc,
		userRepo:       userRepo,
		boundarySvc:    boundarySvc,
		menuSvc:        menuSvc,
		featurePkgRepo: featurePkgRepo,
		personalAccess: personalAccess,
		refresher:      refresher,
		audit:          auditRecorder,
		db:             db,
		logger:         logger,
	}
}

