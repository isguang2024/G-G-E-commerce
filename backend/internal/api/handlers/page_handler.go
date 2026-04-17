package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/page"
)

// pageAPIHandler 负责 /pages/* 相关操作。
type pageAPIHandler struct {
	pageSvc page.Service
	audit   audit.Recorder
	logger  *zap.Logger
}

func newPageAPIHandler(pageSvc page.Service, auditRecorder audit.Recorder, logger *zap.Logger) *pageAPIHandler {
	return &pageAPIHandler{pageSvc: pageSvc, audit: auditRecorder, logger: logger}
}

