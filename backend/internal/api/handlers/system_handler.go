package handlers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/app"
	"github.com/maben/backend/internal/modules/system/space"
)

// systemAPIHandler 负责系统级操作（应用、空间、注册表等）。
// system.go + system_register.go 均使用此 sub-handler。
type systemAPIHandler struct {
	appSvc   app.Service
	spaceSvc space.Service
	audit    audit.Recorder
	db       *gorm.DB
	logger   *zap.Logger
}

func newSystemAPIHandler(
	appSvc app.Service,
	spaceSvc space.Service,
	auditRecorder audit.Recorder,
	db *gorm.DB,
	logger *zap.Logger,
) *systemAPIHandler {
	return &systemAPIHandler{
		appSvc:   appSvc,
		spaceSvc: spaceSvc,
		audit:    auditRecorder,
		db:       db,
		logger:   logger,
	}
}

