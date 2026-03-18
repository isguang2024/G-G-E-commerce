package apiendpoint

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c *gin.Context) {
	var req struct {
		Current               int    `form:"current"`
		Size                  int    `form:"size"`
		Method                string `form:"method"`
		Path                  string `form:"path"`
		Module                string `form:"module"`
		FeatureKind           string `form:"feature_kind"`
		ResourceCode          string `form:"resource_code"`
		ActionCode            string `form:"action_code"`
		ScopeCode             string `form:"scope_code"`
		Status                string `form:"status"`
		RequiresTenantContext *bool  `form:"requires_tenant_context"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.service.List(&ListRequest{
		Current:               req.Current,
		Size:                  req.Size,
		Method:                req.Method,
		Path:                  req.Path,
		Module:                req.Module,
		FeatureKind:           req.FeatureKind,
		ResourceCode:          req.ResourceCode,
		ActionCode:            req.ActionCode,
		ScopeCode:             req.ScopeCode,
		Status:                req.Status,
		RequiresTenantContext: req.RequiresTenantContext,
	})
	if err != nil {
		h.logger.Error("List api endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取 API 列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, endpoint := range list {
		records = append(records, endpointToMap(&endpoint))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": maxInt(req.Current, 1),
		"size":    maxInt(req.Size, 20),
	}))
}

func (h *Handler) Sync(c *gin.Context) {
	if err := h.service.Sync(); err != nil {
		h.logger.Error("Sync api endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "同步 API 注册表失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func endpointToMap(endpoint *user.APIEndpoint) gin.H {
	scopeID := ""
	scopeCode := ""
	scopeName := ""
	if endpoint.Scope.ID != (uuid.UUID{}) {
		scopeID = endpoint.Scope.ID.String()
		scopeCode = endpoint.Scope.Code
		scopeName = endpoint.Scope.Name
	}
	return gin.H{
		"id":                      endpoint.ID.String(),
		"method":                  endpoint.Method,
		"path":                    endpoint.Path,
		"module":                  endpoint.Module,
		"feature_kind":            endpoint.FeatureKind,
		"handler":                 endpoint.Handler,
		"summary":                 endpoint.Summary,
		"resource_code":           endpoint.ResourceCode,
		"action_code":             endpoint.ActionCode,
		"scope_id":                scopeID,
		"scope_code":              scopeCode,
		"scope_name":              scopeName,
		"scope_context_kind":      endpoint.Scope.ContextKind,
		"data_permission_code":    endpoint.Scope.DataPermissionCode,
		"data_permission_name":    endpoint.Scope.DataPermissionName,
		"requires_tenant_context": endpoint.RequiresTenantContext,
		"status":                  endpoint.Status,
		"created_at":              endpoint.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":              endpoint.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func maxInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
