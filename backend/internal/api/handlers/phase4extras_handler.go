package handlers

import (
	"go.uber.org/zap"

	systemmod "github.com/maben/backend/internal/modules/system/system"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
)


// phase4ExtrasAPIHandler 负责 Phase 4 补充操作（fast-enter、view-pages、用户权限刷新）。
type phase4ExtrasAPIHandler struct {
	fastEnterSvc systemmod.FastEnterService
	userSvc      user.UserService
	viewPagesSvc systemmod.ViewPagesService
	cwMemberRepo user.CollaborationWorkspaceMemberRepository
	refresher    permissionrefresh.Service
	logger       *zap.Logger
}

func newPhase4ExtrasAPIHandler(
	fastEnterSvc systemmod.FastEnterService,
	userSvc user.UserService,
	viewPagesSvc systemmod.ViewPagesService,
	cwMemberRepo user.CollaborationWorkspaceMemberRepository,
	refresher permissionrefresh.Service,
	logger *zap.Logger,
) *phase4ExtrasAPIHandler {
	return &phase4ExtrasAPIHandler{
		fastEnterSvc: fastEnterSvc,
		userSvc:      userSvc,
		viewPagesSvc: viewPagesSvc,
		cwMemberRepo: cwMemberRepo,
		refresher:    refresher,
		logger:       logger,
	}
}

