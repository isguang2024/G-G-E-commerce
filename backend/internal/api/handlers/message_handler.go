package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/observability/audit"
	systemmod "github.com/maben/backend/internal/modules/system/system"
)

// messageAPIHandler 负责消息收件箱及分发操作。
type messageAPIHandler struct {
	systemFacade *systemmod.Facade
	audit        audit.Recorder
	logger       *zap.Logger
}

func newMessageAPIHandler(systemFacade *systemmod.Facade, auditRecorder audit.Recorder, logger *zap.Logger) *messageAPIHandler {
	return &messageAPIHandler{systemFacade: systemFacade, audit: auditRecorder, logger: logger}
}
