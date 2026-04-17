package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/menu"
)

// menuAPIHandler 负责 /menus/* 相关操作。
type menuAPIHandler struct {
	menuSvc menu.MenuService
	audit   audit.Recorder
	logger  *zap.Logger
}

func newMenuAPIHandler(menuSvc menu.MenuService, auditRecorder audit.Recorder, logger *zap.Logger) *menuAPIHandler {
	return &menuAPIHandler{menuSvc: menuSvc, audit: auditRecorder, logger: logger}
}

