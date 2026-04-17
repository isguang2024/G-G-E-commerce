package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/featurepackage"
	"github.com/maben/backend/internal/modules/system/menu"
	"github.com/maben/backend/internal/modules/system/page"
	"github.com/maben/backend/internal/modules/system/permission"
)

// extrasAPIHandler 负责跨域聚合查询类操作（extras.go）。
type extrasAPIHandler struct {
	featurePkgSvc featurepackage.Service
	menuSvc       menu.MenuService
	pageSvc       page.Service
	permSvc       permission.PermissionService
	logger        *zap.Logger
}

func newExtrasAPIHandler(
	featurePkgSvc featurepackage.Service,
	menuSvc menu.MenuService,
	pageSvc page.Service,
	permSvc permission.PermissionService,
	logger *zap.Logger,
) *extrasAPIHandler {
	return &extrasAPIHandler{
		featurePkgSvc: featurePkgSvc,
		menuSvc:       menuSvc,
		pageSvc:       pageSvc,
		permSvc:       permSvc,
		logger:        logger,
	}
}

