package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/service"
)

// ScopeHandler 作用域管理处理器
type ScopeHandler struct {
	scopeService service.ScopeService
	logger       *zap.Logger
}

// NewScopeHandler 创建作用域管理处理器
func NewScopeHandler(scopeService service.ScopeService, logger *zap.Logger) *ScopeHandler {
	return &ScopeHandler{scopeService: scopeService, logger: logger}
}

// List 作用域列表（分页）
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
			"scopeId":    s.ID.String(),
			"scopeCode":  s.Code,
			"scopeName":  s.Name,
			"description": s.Description,
			"sortOrder":  s.SortOrder,
			"createTime": s.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": records,
		"total":   total,
		"current": req.Current,
		"size":    req.Size,
	}))
}

// GetAll 获取所有作用域（用于下拉选择）
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
			"scopeId":   s.ID.String(),
			"scopeCode": s.Code,
			"scopeName": s.Name,
		})
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"records": records}))
}

// Get 作用域详情
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
		if err == service.ErrScopeNotFound {
			status, resp := errcode.Response(errcode.ErrNotFound)
			c.JSON(status, resp)
			return
		}
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取作用域失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"scopeId":    scope.ID.String(),
		"scopeCode":   scope.Code,
		"scopeName":   scope.Name,
		"description": scope.Description,
		"sortOrder":   scope.SortOrder,
		"createTime":  scope.CreatedAt.Format("2006-01-02 15:04:05"),
	}))
}

// Create 创建作用域
func (h *ScopeHandler) Create(c *gin.Context) {
	var req dto.ScopeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	scope, err := h.scopeService.Create(&req)
	if err != nil {
		if err == service.ErrScopeCodeExists {
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

// Update 更新作用域
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
		if err == service.ErrScopeNotFound {
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

// Delete 删除作用域
func (h *ScopeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的作用域ID")
		c.JSON(status, resp)
		return
	}
	if err := h.scopeService.Delete(id); err != nil {
		if err == service.ErrScopeNotFound {
			status, resp := errcode.Response(errcode.ErrNotFound)
			c.JSON(status, resp)
			return
		}
		if err == service.ErrScopeInUse {
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "默认作用域无法删除")
			c.JSON(status, resp)
			return
		}
		// 检查是否是 ErrScopeInUseWithRoles 错误
		if errWithRoles, ok := err.(*service.ErrScopeInUseWithRoles); ok {
			roleNames := make([]string, 0, len(errWithRoles.Roles))
			for _, r := range errWithRoles.Roles {
				roleNames = append(roleNames, r.Name)
			}
			status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "作用域正在使用中，无法删除")
			// 在响应中添加角色列表信息
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
