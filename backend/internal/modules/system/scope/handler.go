package scope

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
)

type ScopeHandler struct {
	scopeService ScopeService
	logger       *zap.Logger
}

func NewScopeHandler(scopeService ScopeService, logger *zap.Logger) *ScopeHandler {
	return &ScopeHandler{
		scopeService: scopeService,
		logger:       logger,
	}
}

func (h *ScopeHandler) List(c *gin.Context) {
	var req dto.ScopeListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	list, total, err := h.scopeService.List(&req)
	if err != nil {
		h.logger.Error("Scope list failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取作用域列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, s := range list {
		records = append(records, gin.H{
			"scopeId":            s.ID.String(),
			"scopeCode":          s.Code,
			"scopeName":          s.Name,
			"description":        s.Description,
			"isSystem":           s.IsSystem,
			"contextKind":        s.ContextKind,
			"dataPermissionCode": s.DataPermissionCode,
			"dataPermissionName": s.DataPermissionName,
			"sortOrder":          s.SortOrder,
			"createTime":         s.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

func (h *ScopeHandler) GetAll(c *gin.Context) {
	list, err := h.scopeService.GetAll()
	if err != nil {
		h.logger.Error("Get all scopes failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取作用域列表失败")
		c.JSON(status, resp)
		return
	}
	records := make([]gin.H, 0, len(list))
	for _, s := range list {
		records = append(records, gin.H{
			"scopeId":            s.ID.String(),
			"scopeCode":          s.Code,
			"scopeName":          s.Name,
			"isSystem":           s.IsSystem,
			"contextKind":        s.ContextKind,
			"dataPermissionCode": s.DataPermissionCode,
			"dataPermissionName": s.DataPermissionName,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"records": records}))
}

func (h *ScopeHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的作用域ID")
		c.JSON(status, resp)
		return
	}
	scope, err := h.scopeService.Get(id)
	if err != nil {
		if err == ErrScopeNotFound {
			status, resp := errcode.Response(errcode.ErrNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取作用域失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"scopeId":            scope.ID.String(),
		"scopeCode":          scope.Code,
		"scopeName":          scope.Name,
		"description":        scope.Description,
		"isSystem":           scope.IsSystem,
		"contextKind":        scope.ContextKind,
		"dataPermissionCode": scope.DataPermissionCode,
		"dataPermissionName": scope.DataPermissionName,
		"sortOrder":          scope.SortOrder,
		"createTime":         scope.CreatedAt.Format("2006-01-02 15:04:05"),
	}))
}

func (h *ScopeHandler) Create(c *gin.Context) {
	var req dto.ScopeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	scope, err := h.scopeService.Create(&req)
	if err != nil {
		if err == ErrScopeCodeExists {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "作用域编码已存在")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Create scope failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建作用域失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"scopeId": scope.ID.String()}))
}

func (h *ScopeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的作用域ID")
		c.JSON(status, resp)
		return
	}
	var req dto.ScopeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	if err := h.scopeService.Update(id, &req); err != nil {
		if err == ErrScopeNotFound {
			status, resp := errcode.Response(errcode.ErrNotFound)
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Update scope failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "更新作用域失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *ScopeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的作用域ID")
		c.JSON(status, resp)
		return
	}
	if err := h.scopeService.Delete(id); err != nil {
		if err == ErrScopeNotFound {
			status, resp := errcode.Response(errcode.ErrNotFound)
			c.JSON(status, resp)
			return
		}
		if err == ErrScopeInUse {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "系统内置作用域无法删除")
			c.JSON(status, resp)
			return
		}
		if errWithRoles, ok := err.(*ErrScopeInUseWithRoles); ok {
			roleNames := make([]string, 0, len(errWithRoles.Roles))
			for _, r := range errWithRoles.Roles {
				roleNames = append(roleNames, r.Name)
			}
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "作用域正在使用中，无法删除")
			if resp.Data == nil {
				resp.Data = gin.H{}
			}
			if data, ok := resp.Data.(gin.H); ok {
				data["roles"] = roleNames
				data["roleCount"] = len(roleNames)
			}
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Delete scope failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "删除作用域失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}
