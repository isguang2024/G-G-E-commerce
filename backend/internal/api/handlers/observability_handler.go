package handlers

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/telemetry"
)

// observabilityAPIHandler 负责可观测性查询类操作（审计日志、遥测查询等）。
type observabilityAPIHandler struct {
	db        *gorm.DB
	audit     audit.Recorder
	telemetry telemetry.Ingester
	logger    *zap.Logger
}

func newObservabilityAPIHandler(
	db *gorm.DB,
	auditRecorder audit.Recorder,
	telemetryIngester telemetry.Ingester,
	logger *zap.Logger,
) *observabilityAPIHandler {
	return &observabilityAPIHandler{
		db:        db,
		audit:     auditRecorder,
		telemetry: telemetryIngester,
		logger:    logger,
	}
}
