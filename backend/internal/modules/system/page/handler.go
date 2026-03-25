package page

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
)

type Handler struct {
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) List(c *gin.Context) {
	var req ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	list, total, err := h.service.List(&req)
	if err != nil {
		h.logger.Error("List pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取页面列表失败")
		c.JSON(status, resp)
		return
	}
	current, size := normalizePageAndSize(&req)
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": list,
		"total":   total,
		"current": current,
		"size":    size,
	}))
}

func (h *Handler) ListRuntime(c *gin.Context) {
	list, err := h.service.ListRuntime()
	if err != nil {
		h.logger.Error("List runtime pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取运行时页面注册表失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": list,
		"total":   len(list),
	}))
}

func (h *Handler) ListUnregistered(c *gin.Context) {
	items, err := h.service.ListUnregistered()
	if err != nil {
		h.logger.Error("List unregistered pages failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取未注册页面失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) Sync(c *gin.Context) {
	result, err := h.service.Sync()
	if err != nil {
		h.respondServiceError(c, err, "同步页面注册表失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(result))
}

func (h *Handler) ListMenuOptions(c *gin.Context) {
	items, err := h.service.ListMenuOptions()
	if err != nil {
		h.logger.Error("List page menu options failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取上级菜单候选失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	item, getErr := h.service.Get(id)
	if getErr != nil {
		h.respondServiceError(c, getErr, "获取页面详情失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) PreviewBreadcrumb(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	items, previewErr := h.service.PreviewBreadcrumb(id)
	if previewErr != nil {
		h.respondServiceError(c, previewErr, "预览页面面包屑失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"items": items,
		"total": len(items),
	}))
}

func (h *Handler) Create(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item, createErr := h.service.Create(&req)
	if createErr != nil {
		h.respondServiceError(c, createErr, "创建页面失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	item, updateErr := h.service.Update(id, &req)
	if updateErr != nil {
		h.respondServiceError(c, updateErr, "更新页面失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的页面ID")
		c.JSON(status, resp)
		return
	}
	if deleteErr := h.service.Delete(id); deleteErr != nil {
		h.respondServiceError(c, deleteErr, "删除页面失败")
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(nil))
}

func (h *Handler) respondServiceError(c *gin.Context, err error, fallback string) {
	switch err {
	case ErrPageNotFound:
		status, resp := errcode.ResponseWithMsg(errcode.ErrNotFound, "页面不存在")
		c.JSON(status, resp)
	case ErrPageKeyExists:
		status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "页面标识已存在")
		c.JSON(status, resp)
	case ErrRouteNameExists:
		status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "路由名称已存在")
		c.JSON(status, resp)
	case ErrParentMenuInvalid:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, "无效的上级菜单")
		c.JSON(status, resp)
	case ErrParentPageInvalid:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, "无效的上级页面")
		c.JSON(status, resp)
	case ErrDisplayGroupInvalid:
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidParent, "无效的普通分组")
		c.JSON(status, resp)
	case ErrPageHasChildren:
		status, resp := errcode.ResponseWithMsg(errcode.ErrConflict, "当前页面下仍有子页面或分组，不能直接删除")
		c.JSON(status, resp)
	default:
		if strings.Contains(err.Error(), ErrPageValidation.Error()) {
			status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, strings.TrimPrefix(err.Error(), ErrPageValidation.Error()+": "))
			c.JSON(status, resp)
			return
		}
		h.logger.Error(fallback, zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, fallback)
		c.JSON(status, resp)
	}
}
