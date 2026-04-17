package handlers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/auth"
	"github.com/maben/backend/internal/modules/system/register"
	"github.com/maben/backend/internal/modules/system/social"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/modules/system/workspace"
	"github.com/maben/backend/internal/pkg/permission/evaluator"
)

// authAPIHandler 负责 /auth/* 相关操作。
//
// 注意：auth.go 中的 login helper 需要 workspace.Service（ensurePersonalWorkspace）
// 和 evaluator（登录后 resolve 权限），因此一并注入。
type authAPIHandler struct {
	authSvc            auth.AuthService
	centralizedAuthSvc auth.CentralizedAuthService
	userRepo           user.UserRepository
	socialSvc          social.Service
	registerSvc        *register.Service
	registerResolver   *register.Resolver
	evaluator          evaluator.Evaluator
	service            workspace.Service
	audit              audit.Recorder
	db                 *gorm.DB
	logger             *zap.Logger
}

func newAuthAPIHandler(
	authSvc auth.AuthService,
	centralizedAuthSvc auth.CentralizedAuthService,
	userRepo user.UserRepository,
	socialSvc social.Service,
	registerSvc *register.Service,
	registerResolver *register.Resolver,
	eval evaluator.Evaluator, // 参数名保留 eval，赋给字段 evaluator
	service workspace.Service,
	auditRecorder audit.Recorder,
	db *gorm.DB,
	logger *zap.Logger,
) *authAPIHandler {
	return &authAPIHandler{
		authSvc:            authSvc,
		centralizedAuthSvc: centralizedAuthSvc,
		userRepo:           userRepo,
		socialSvc:          socialSvc,
		registerSvc:        registerSvc,
		registerResolver:   registerResolver,
		evaluator:          eval,
		service:            service,
		audit:              auditRecorder,
		db:                 db,
		logger:             logger,
	}
}

