package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/permission"
)

// permissionAPIHandler 负责 /permission-* 相关操作。
type permissionAPIHandler struct {
	permSvc permission.PermissionService
	logger  *zap.Logger
}

func newPermissionAPIHandler(permSvc permission.PermissionService, logger *zap.Logger) *permissionAPIHandler {
	return &permissionAPIHandler{permSvc: permSvc, logger: logger}
}

