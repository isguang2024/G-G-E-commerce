package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/observability/telemetry"
)

// telemetryAPIHandler 负责 /telemetry/* 相关操作。
type telemetryAPIHandler struct {
	telemetry telemetry.Ingester
	logger    *zap.Logger
}

func newTelemetryAPIHandler(telemetryIngester telemetry.Ingester, logger *zap.Logger) *telemetryAPIHandler {
	return &telemetryAPIHandler{telemetry: telemetryIngester, logger: logger}
}

