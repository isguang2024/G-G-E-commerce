package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/upload"
)

// storageAdminAPIHandler 负责 /storage-admin/* 相关操作。
type storageAdminAPIHandler struct {
	uploadSvc upload.Service
	logger    *zap.Logger
}

func newStorageAdminAPIHandler(uploadSvc upload.Service, logger *zap.Logger) *storageAdminAPIHandler {
	return &storageAdminAPIHandler{uploadSvc: uploadSvc, logger: logger}
}

