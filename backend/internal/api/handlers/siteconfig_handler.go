package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/system/siteconfig"
)

// siteConfigAPIHandler 负责 /site-config/* 相关操作。
type siteConfigAPIHandler struct {
	siteConfigSvc siteconfig.Service
	audit         audit.Recorder
	logger        *zap.Logger
}

func newSiteConfigAPIHandler(siteConfigSvc siteconfig.Service, auditRecorder audit.Recorder, logger *zap.Logger) *siteConfigAPIHandler {
	return &siteConfigAPIHandler{siteConfigSvc: siteConfigSvc, audit: auditRecorder, logger: logger}
}
