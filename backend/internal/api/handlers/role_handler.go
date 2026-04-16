package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/role"
)

// roleAPIHandler 负责 /roles/* 相关操作。
type roleAPIHandler struct {
	roleSvc role.RoleService
	logger  *zap.Logger
}

func newRoleAPIHandler(roleSvc role.RoleService, logger *zap.Logger) *roleAPIHandler {
	return &roleAPIHandler{roleSvc: roleSvc, logger: logger}
}
