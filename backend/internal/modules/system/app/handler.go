package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
)

type Handler struct {
	logger  *zap.Logger
	service Service
}

func NewHandler(logger *zap.Logger, service Service) *Handler {
	return &Handler{logger: logger, service: service}
}

func (h *Handler) List(c *gin.Context) {
	items, err := h.service.ListApps()
	if err != nil {
		h.logger.Error("List apps failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取应用列表失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) GetCurrent(c *gin.Context) {
	item, err := h.service.GetCurrent(spacepkg.RequestHost(c), RequestAppKey(c))
	if err != nil {
		h.logger.Error("Get current app failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取当前应用失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) SaveApp(c *gin.Context) {
	var req SaveAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	saved, err := h.service.SaveApp(&req)
	if err != nil {
		h.logger.Error("Save app failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(saved))
}

func (h *Handler) ListHostBindings(c *gin.Context) {
	items, err := h.service.ListHostBindings(RequestAppKey(c))
	if err != nil {
		h.logger.Error("List app host bindings failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取应用 Host 绑定失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) SaveHostBinding(c *gin.Context) {
	var req SaveHostBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	saved, err := h.service.SaveHostBinding(&req)
	if err != nil {
		h.logger.Error("Save app host binding failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(saved))
}
