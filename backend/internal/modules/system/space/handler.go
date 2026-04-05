package space

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
)

type Handler struct {
	logger  *zap.Logger
	service Service
}

type spaceModeSaveRequest struct {
	Mode string `json:"mode"`
}

func NewHandler(logger *zap.Logger, service Service) *Handler {
	return &Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) GetCurrent(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	current, err := h.service.GetCurrent(appKey, RequestHost(c), RequestSpaceKey(c), currentContextUserID(c), currentContextTenantID(c))
	if err != nil {
		h.logger.Error("Get current menu space failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取当前空间失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(current))
}

func (h *Handler) List(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	items, err := h.service.ListSpaces(appKey)
	if err != nil {
		h.logger.Error("List menu spaces failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取空间列表失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) GetMode(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	mode, err := h.service.GetMode()
	if err != nil {
		h.logger.Error("Get menu space mode failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取菜单空间模式失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"mode": mode}))
}

func (h *Handler) SaveMode(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	var req spaceModeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	mode, err := h.service.SaveMode(req.Mode)
	if err != nil {
		h.logger.Error("Save menu space mode failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存菜单空间模式失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{"mode": mode}))
}

func (h *Handler) ListHostBindings(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	items, err := h.service.ListHostBindings(appKey)
	if err != nil {
		h.logger.Error("List menu space host bindings failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取 Host 绑定失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) SaveSpace(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	var req SaveSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	item, err := h.service.SaveSpace(appKey, &req)
	if err != nil {
		h.logger.Error("Save menu space failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) SaveHostBinding(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	var req SaveHostBindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	item, err := h.service.SaveHostBinding(appKey, &req)
	if err != nil {
		h.logger.Error("Save menu space host binding failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}

func (h *Handler) InitializeFromDefault(c *gin.Context) {
	if h.service == nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "空间服务未初始化")
		c.JSON(status, resp)
		return
	}
	spaceKey := c.Param("spaceKey")
	force := strings.EqualFold(strings.TrimSpace(c.Query("force")), "true")
	appKey, err := appctx.RequireRequestAppKey(c)
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "app_key is required")
		c.JSON(status, resp)
		return
	}
	var actorUserID *uuid.UUID
	if value := strings.TrimSpace(c.GetString("user_id")); value != "" {
		if parsedID, parseErr := uuid.Parse(value); parseErr == nil {
			actorUserID = &parsedID
		}
	}
	item, err := h.service.InitializeFromDefault(appKey, spaceKey, force, actorUserID)
	if err != nil {
		h.logger.Error("Initialize menu space from default failed", zap.String("space_key", spaceKey), zap.Bool("force", force), zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, err.Error())
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(item))
}
