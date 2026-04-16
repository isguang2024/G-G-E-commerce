package handlers

import (
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/featurepackage"
)

// featurePackageAPIHandler 负责 /feature-packages/* 相关操作。
type featurePackageAPIHandler struct {
	featurePkgSvc featurepackage.Service
	logger        *zap.Logger
}

func newFeaturePackageAPIHandler(featurePkgSvc featurepackage.Service, logger *zap.Logger) *featurePackageAPIHandler {
	return &featurePackageAPIHandler{featurePkgSvc: featurePkgSvc, logger: logger}
}
