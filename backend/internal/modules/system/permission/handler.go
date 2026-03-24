package permission

import (
	"net/http"

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
	var req dto.PermissionActionListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.permissionService.List(&req)
	if err != nil {
		h.logger.Error("List permission actions failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, action := range list {
		records = append(records, actionToMap(&action))
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *PermissionHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	action, err := h.permissionService.Get(id)
	if err != nil {
		if err == ErrPermissionActionNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(actionToMap(action)))
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
		if err == ErrPermissionActionNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("List permission action endpoints failed", zap.Error(err))
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
	var req dto.PermissionActionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	action, err := h.permissionService.Create(&req)
	if err != nil {
		if err == ErrPermissionActionExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能权限编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create permission action failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建功能权限失败: "+err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"id": action.ID.String()}))
}

func (h *PermissionHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的功能权限ID")
		c.JSON(status, resp)
		return
	}
	var req dto.PermissionActionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.permissionService.Update(id, &req); err != nil {
		if err == ErrPermissionActionNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		if err == ErrPermissionActionExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "功能权限编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update permission action failed", zap.Error(err))
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
		if err == ErrPermissionActionNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "功能权限不存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete permission action failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除功能权限失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func actionToMap(action *user.PermissionAction) gin.H {
	permissionKey := canonicalPermissionKey(action.PermissionKey)
	mapping := permissionkey.FromKey(permissionKey)
	contextType := action.ContextType
	if contextType == "" {
		contextType = mapping.ContextType
	}
	name := action.Name
	description := action.Description
	if description == "" && mapping.Description != "" {
		description = mapping.Description
	}
	moduleCode := action.ModuleCode
	if action.ModuleGroup != nil && action.ModuleGroup.Code != "" {
		moduleCode = action.ModuleGroup.Code
	}
	featureKind := action.FeatureKind
	if action.FeatureGroup != nil && action.FeatureGroup.Code != "" {
		featureKind = action.FeatureGroup.Code
	}
	return gin.H{
		"id":               action.ID.String(),
		"permission_key":   permissionKey,
		"module_code":      moduleCode,
		"module_group_id":  stringifyUUIDPointer(action.ModuleGroupID),
		"feature_group_id": stringifyUUIDPointer(action.FeatureGroupID),
		"module_group":     permissionGroupToMap(action.ModuleGroup),
		"feature_group":    permissionGroupToMap(action.FeatureGroup),
		"context_type":     contextType,
		"feature_kind":     featureKind,
		"name":             name,
		"description":      description,
		"status":           action.Status,
		"sort_order":       action.SortOrder,
		"is_builtin":       action.IsBuiltin,
		"created_at":       action.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":       action.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
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
		"method":        endpoint.Method,
		"path":          endpoint.Path,
		"spec":          endpoint.Method + " " + endpoint.Path,
		"module":        endpoint.Module,
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
