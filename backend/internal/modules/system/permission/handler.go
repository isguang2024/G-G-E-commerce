package permission

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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
	scopeID := ""
	scopeCode := ""
	scopeName := ""
	if action.Scope.ID != (uuid.UUID{}) {
		scopeID = action.Scope.ID.String()
		scopeCode = action.Scope.Code
		scopeName = action.Scope.Name
	}
	return gin.H{
		"id":                      action.ID.String(),
		"resource_code":           action.ResourceCode,
		"action_code":             action.ActionCode,
		"module_code":             action.ModuleCode,
		"permission_key":          action.ResourceCode + ":" + action.ActionCode,
		"category":                action.Category,
		"source":                  action.Source,
		"feature_kind":            action.FeatureKind,
		"name":                    action.Name,
		"description":             action.Description,
		"scope_id":                scopeID,
		"scope_code":              scopeCode,
		"scope_name":              scopeName,
		"data_permission_code":    action.Scope.DataPermissionCode,
		"data_permission_name":    action.Scope.DataPermissionName,
		"scope":                   scopeCode,
		"requires_tenant_context": action.RequiresTenantContext,
		"status":                  action.Status,
		"sort_order":              action.SortOrder,
		"created_at":              action.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":              action.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
