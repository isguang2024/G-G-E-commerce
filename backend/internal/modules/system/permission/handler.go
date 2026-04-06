package permission

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionkey"
)

type PermissionHandler struct {
	permissionService PermissionService
	logger            *zap.Logger
}

func NewPermissionHandler(permissionService PermissionService, logger *zap.Logger) *PermissionHandler {
	return &PermissionHandler{permissionService: permissionService, logger: logger}
}

func (h *PermissionHandler) List(c *gin.Context) {
	var req dto.PermissionKeyListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, summary, err := h.permissionService.List(&req)
	if err != nil {
		h.logger.Error("List permission keys failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, item := range list {
		records = append(records, permissionKeyToMap(&item.PermissionKey, item.Audit))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records":       records,
		"total":         total,
		"current":       req.Current,
		"size":          req.Size,
		"audit_summary": permissionAuditSummaryToMap(summary),
	}))
}

func (h *PermissionHandler) ListOptions(c *gin.Context) {
	var req dto.PermissionKeyListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, err := h.permissionService.ListOptions(&req)
	if err != nil {
		h.logger.Error("List permission key options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限候选失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, item := range list {
		action := item
		records = append(records, permissionKeyToMap(&action, PermissionAuditProfile{}))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *PermissionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	item, err := h.permissionService.Get(id)
	if err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(permissionKeyToMap(item, PermissionAuditProfile{})))
}

func (h *PermissionHandler) ListGroups(c *gin.Context) {
	var req dto.PermissionGroupListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.permissionService.ListGroups(&req)
	if err != nil {
		h.logger.Error("List permission groups failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限分组失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, item := range list {
		records = append(records, permissionGroupToMap(&item))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *PermissionHandler) ListEndpoints(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	list, err := h.permissionService.ListEndpoints(id)
	if err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("List permission key endpoints failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限关联接口失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, endpoint := range list {
		records = append(records, endpointToMap(&endpoint))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *PermissionHandler) GetConsumerDetails(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	result, err := h.permissionService.GetConsumerDetails(id)
	if err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get permission key consumer details failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限消费明细失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *PermissionHandler) GetImpactPreview(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	result, err := h.permissionService.GetImpactPreview(id)
	if err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get permission impact preview failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取影响预览失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *PermissionHandler) BatchUpdate(c *gin.Context) {
	var req PermissionBatchUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	operatorID := parseCurrentUserID(c)
	result, err := h.permissionService.BatchUpdate(&req, operatorID, strings.TrimSpace(c.GetHeader("X-Request-ID")))
	if err != nil {
		h.logger.Error("Batch update permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "批量更新功能权限失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *PermissionHandler) SaveBatchTemplate(c *gin.Context) {
	var req PermissionBatchTemplateSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	operatorID := parseCurrentUserID(c)
	item, err := h.permissionService.SaveBatchTemplate(&req, operatorID)
	if err != nil {
		h.logger.Error("Save permission batch template failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存批量模板失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"id":          item.ID.String(),
		"name":        item.Name,
		"description": item.Description,
		"payload":     item.Payload,
		"updated_at":  item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}))
}

func (h *PermissionHandler) ListBatchTemplates(c *gin.Context) {
	items, err := h.permissionService.ListBatchTemplates()
	if err != nil {
		h.logger.Error("List permission batch templates failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取批量模板失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(items))
	for _, item := range items {
		records = append(records, gin.H{
			"id":          item.ID.String(),
			"name":        item.Name,
			"description": item.Description,
			"payload":     item.Payload,
			"created_by":  stringifyUUIDPointer(item.CreatedBy),
			"created_at":  item.CreatedAt.Format("2006-01-02 15:04:05"),
			"updated_at":  item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   len(records),
	}))
}

func (h *PermissionHandler) ListRiskAudits(c *gin.Context) {
	current := parsePositiveInt(c.Query("current"), 1)
	size := parsePositiveInt(c.Query("size"), 20)
	objectID := strings.TrimSpace(c.Query("object_id"))
	items, total, err := h.permissionService.ListRiskAudits(objectID, current, size)
	if err != nil {
		h.logger.Error("List permission risk audits failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取最近变更失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(items))
	for _, item := range items {
		records = append(records, gin.H{
			"id":             item.ID.String(),
			"operator_id":    stringifyUUIDPointer(item.OperatorID),
			"object_type":    item.ObjectType,
			"object_id":      item.ObjectID,
			"operation_type": item.OperationType,
			"before_summary": item.BeforeSummary,
			"after_summary":  item.AfterSummary,
			"impact_summary": item.ImpactSummary,
			"request_id":     item.RequestID,
			"created_at":     item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

func (h *PermissionHandler) CleanupUnused(c *gin.Context) {
	result, err := h.permissionService.CleanupUnused()
	if err != nil {
		h.logger.Error("Cleanup unused permission keys failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "清理未消费功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"deleted_count": result.DeletedCount,
		"deleted_keys":  result.DeletedKeys,
	}))
}

func (h *PermissionHandler) AddEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	var req dto.PermissionKeyEndpointBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	endpointCode := strings.TrimSpace(req.EndpointCode)
	if endpointCode == "" {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "无效的接口编码")
		c.JSON(status, resp)
		return
	}
	if err := h.permissionService.AddEndpoint(id, endpointCode); err != nil {
		switch err {
		case ErrPermissionKeyNotFound:
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		case ErrAPIEndpointNotFound:
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "接口不存在")
			c.JSON(status, resp)
			return
		default:
			h.logger.Error("Add permission key endpoint failed", zap.Error(err))
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "新增关联接口失败: "+err.Error())
			c.JSON(status, resp)
			return
		}
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *PermissionHandler) RemoveEndpoint(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	endpointCode := strings.TrimSpace(c.Param("endpointCode"))
	if endpointCode == "" {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "无效的接口编码")
		c.JSON(status, resp)
		return
	}
	if err := h.permissionService.RemoveEndpoint(id, endpointCode); err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Remove permission key endpoint failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除关联接口失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *PermissionHandler) CreateGroup(c *gin.Context) {
	var req dto.PermissionGroupSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item, err := h.permissionService.CreateGroup(&req)
	if err != nil {
		if err == ErrPermissionGroupExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能权限分组编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create permission group failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建功能权限分组失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": item.ID.String()}))
}

func (h *PermissionHandler) UpdateGroup(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限分组ID")
		c.JSON(status, resp)
		return
	}
	var req dto.PermissionGroupSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.permissionService.UpdateGroup(id, &req); err != nil {
		if err == ErrPermissionGroupNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限分组不存在")
			c.JSON(status, resp)
			return
		}
		if err == ErrPermissionGroupExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能权限分组编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update permission group failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新功能权限分组失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *PermissionHandler) Create(c *gin.Context) {
	var req dto.PermissionKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item, err := h.permissionService.Create(&req)
	if err != nil {
		if err == ErrPermissionKeyExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能权限编码已存在")
			c.JSON(status, resp)
			return
		}
		if errors.Is(err, ErrPermissionContextInvalid) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create permission key failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建功能权限失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": item.ID.String()}))
}

func (h *PermissionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	var req dto.PermissionKeyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.permissionService.Update(id, &req); err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		if err == ErrPermissionKeyExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能权限编码已存在")
			c.JSON(status, resp)
			return
		}
		if errors.Is(err, ErrPermissionContextInvalid) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, err.Error())
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update permission key failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新功能权限失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *PermissionHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	if err := h.permissionService.Delete(id); err != nil {
		if err == ErrPermissionKeyNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete permission key failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func permissionKeyToMap(item *user.PermissionKey, audit PermissionAuditProfile) gin.H {
	permissionKey := canonicalPermissionKey(item.PermissionKey)
	mapping := permissionkey.FromKey(permissionKey)
	contextType := item.ContextType
	if contextType == "" {
		contextType = mapping.ContextType
	}
	name := item.Name
	description := item.Description
	if description == "" && mapping.Description != "" {
		description = mapping.Description
	}
	moduleCode := item.ModuleCode
	if item.ModuleGroup != nil && item.ModuleGroup.Code != "" {
		moduleCode = item.ModuleGroup.Code
	}
	featureKind := item.FeatureKind
	if item.FeatureGroup != nil && item.FeatureGroup.Code != "" {
		featureKind = item.FeatureGroup.Code
	}
	effectiveStatus := resolvePermissionKeyStatus(item)
	return gin.H{
		"id":                      item.ID.String(),
		"permission_key":          permissionKey,
		"app_key":                 item.AppKey,
		"module_code":             moduleCode,
		"module_group_id":         stringifyUUIDPointer(item.ModuleGroupID),
		"feature_group_id":        stringifyUUIDPointer(item.FeatureGroupID),
		"module_group":            permissionGroupToMap(item.ModuleGroup),
		"feature_group":           permissionGroupToMap(item.FeatureGroup),
		"context_type":            contextType,
		"feature_kind":            featureKind,
		"data_policy":             item.DataPolicy,
		"allowed_workspace_types": item.AllowedWorkspaceTypes,
		"name":                    name,
		"description":             description,
		"status":                  effectiveStatus,
		"self_status":             item.Status,
		"sort_order":              item.SortOrder,
		"is_builtin":              item.IsBuiltin,
		"api_count":               audit.APICount,
		"page_count":              audit.PageCount,
		"package_count":           audit.PackageCount,
		"consumer_types":          audit.ConsumerTypes,
		"usage_pattern":           audit.UsagePattern,
		"usage_note":              audit.UsageNote,
		"duplicate_pattern":       audit.DuplicatePattern,
		"duplicate_group":         audit.DuplicateGroup,
		"duplicate_keys":          audit.DuplicateKeys,
		"duplicate_note":          audit.DuplicateNote,
		"created_at":              item.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":              item.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func permissionAuditSummaryToMap(summary PermissionAuditSummary) gin.H {
	return gin.H{
		"total_count":                summary.TotalCount,
		"unused_count":               summary.UnusedCount,
		"api_only_count":             summary.APIOnlyCount,
		"page_only_count":            summary.PageOnlyCount,
		"package_only_count":         summary.PackageOnlyCount,
		"multi_consumer_count":       summary.MultiConsumerCount,
		"cross_context_mirror_count": summary.CrossContextMirrorCount,
		"suspected_duplicate_count":  summary.SuspectedDuplicateCount,
	}
}

func resolvePermissionKeyStatus(item *user.PermissionKey) string {
	if item == nil {
		return "normal"
	}
	if item.Status == "suspended" {
		return "suspended"
	}
	if item.ModuleGroup != nil && item.ModuleGroup.Status == "suspended" {
		return "suspended"
	}
	if item.FeatureGroup != nil && item.FeatureGroup.Status == "suspended" {
		return "suspended"
	}
	return "normal"
}

func permissionGroupToMap(group *user.PermissionGroup) gin.H {
	if group == nil {
		return nil
	}
	return gin.H{
		"id":          group.ID.String(),
		"group_type":  group.GroupType,
		"code":        group.Code,
		"name":        group.Name,
		"name_en":     group.NameEn,
		"description": group.Description,
		"status":      group.Status,
		"sort_order":  group.SortOrder,
		"is_builtin":  group.IsBuiltin,
	}
}

func endpointToMap(endpoint *user.APIEndpoint) gin.H {
	authMode := "jwt"
	switch {
	case endpoint.Path == "/health":
		authMode = "public"
	case endpoint.Path == "/api/v1/auth/login" || endpoint.Path == "/api/v1/auth/register" || endpoint.Path == "/api/v1/auth/refresh":
		authMode = "public"
	case len(endpoint.Path) >= len("/open/v1/") && endpoint.Path[:len("/open/v1/")] == "/open/v1/":
		authMode = "api_key"
	default:
		authMode = "permission"
	}
	return gin.H{
		"id":            endpoint.ID.String(),
		"code":          endpoint.Code,
		"method":        endpoint.Method,
		"path":          endpoint.Path,
		"spec":          endpoint.Method + " " + endpoint.Path,
		"feature_kind":  endpoint.FeatureKind,
		"handler":       endpoint.Handler,
		"summary":       endpoint.Summary,
		"auth_mode":     authMode,
		"category_id":   stringifyUUIDPointer(endpoint.CategoryID),
		"context_scope": endpoint.ContextScope,
		"source":        endpoint.Source,
		"status":        endpoint.Status,
		"created_at":    endpoint.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":    endpoint.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func stringifyUUIDPointer(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}

func parseCurrentUserID(c *gin.Context) *uuid.UUID {
	value, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	userIDStr, ok := value.(string)
	if !ok {
		return nil
	}
	userID, err := uuid.Parse(strings.TrimSpace(userIDStr))
	if err != nil {
		return nil
	}
	return &userID
}

func parsePositiveInt(value string, fallback int) int {
	target := strings.TrimSpace(value)
	if target == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(target, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
