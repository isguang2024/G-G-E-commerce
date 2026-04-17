package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/observability/audit"
	"github.com/maben/backend/internal/modules/observability/logpolicy"
)

// logPolicyAPIHandler 负责日志策略控制面操作。
type logPolicyAPIHandler struct {
	policyRepo   logpolicy.Repository
	policyEngine logpolicy.Engine
	audit        audit.Recorder
	logger       *zap.Logger
}

func newLogPolicyAPIHandler(
	policyRepo logpolicy.Repository,
	policyEngine logpolicy.Engine,
	auditRecorder audit.Recorder,
	logger *zap.Logger,
) *logPolicyAPIHandler {
	return &logPolicyAPIHandler{
		policyRepo:   policyRepo,
		policyEngine: policyEngine,
		audit:        auditRecorder,
		logger:       logger,
	}
}

