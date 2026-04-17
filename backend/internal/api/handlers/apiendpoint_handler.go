package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/apiendpoint"
)

// apiEndpointAPIHandler 负责 /api-endpoints/* 相关操作。
type apiEndpointAPIHandler struct {
	apiEndpointSvc apiendpoint.Service
	logger         *zap.Logger
}

func newAPIEndpointAPIHandler(apiEndpointSvc apiendpoint.Service, logger *zap.Logger) *apiEndpointAPIHandler {
	return &apiEndpointAPIHandler{apiEndpointSvc: apiEndpointSvc, logger: logger}
}

