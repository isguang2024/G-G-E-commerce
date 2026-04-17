package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/upload"
	"github.com/maben/backend/internal/pkg/permission/evaluator"
)

// mediaAPIHandler 负责 /media/* 相关操作。
type mediaAPIHandler struct {
	uploadSvc upload.Service
	evaluator evaluator.Evaluator
	logger    *zap.Logger
}

func newMediaAPIHandler(uploadSvc upload.Service, eval evaluator.Evaluator, logger *zap.Logger) *mediaAPIHandler {
	return &mediaAPIHandler{uploadSvc: uploadSvc, evaluator: eval, logger: logger}
}

