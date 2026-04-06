package workspace

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type Handler struct {
	logger  *zap.Logger
	service Service
}

type switchRequest struct {
	WorkspaceID string `json:"workspace_id" binding:"required"`
}

func NewHandler(logger *zap.Logger, service Service) *Handler {
	return &Handler{logger: logger, service: service}
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	items, err := h.service.ListByUserID(userID)
	if err != nil {
		h.logger.Error("List workspaces failed", zap.Error(err), zap.String("user_id", userID.String()))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取工作空间列表失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"records": items,
		"total":   len(items),
	}))
}

func (h *Handler) GetCurrent(c *gin.Context) {
	userID, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	workspace, err := h.resolveCurrentWorkspace(c, userID)
	if err != nil {
		h.logger.Error("Get current workspace failed", zap.Error(err), zap.String("user_id", userID.String()))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取当前工作空间失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(workspaceToSummary(workspace)))
}

func (h *Handler) Get(c *gin.Context) {
	userID, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	workspaceID, err := uuid.Parse(strings.TrimSpace(c.Param("id")))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的工作空间ID")
		c.JSON(status, resp)
		return
	}

	workspace, err := h.getAccessibleWorkspace(userID, workspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "无权访问该工作空间")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Get workspace failed", zap.Error(err), zap.String("user_id", userID.String()), zap.String("workspace_id", workspaceID.String()))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取工作空间失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(workspaceToSummary(workspace)))
}

func (h *Handler) Switch(c *gin.Context) {
	userID, err := h.mustUserID(c)
	if err != nil {
		status, resp := errcode.Response(errcode.ErrUnauthorized)
		c.JSON(status, resp)
		return
	}

	var req switchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	workspaceID, err := uuid.Parse(strings.TrimSpace(req.WorkspaceID))
	if err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrInvalidID, "无效的工作空间ID")
		c.JSON(status, resp)
		return
	}

	workspace, err := h.getAccessibleWorkspace(userID, workspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			status, resp := errcode.ResponseWithMsg(errcode.ErrForbidden, "无权切换到该工作空间")
			c.JSON(status, resp)
			return
		}
		h.logger.Error("Switch workspace failed", zap.Error(err), zap.String("user_id", userID.String()), zap.String("workspace_id", workspaceID.String()))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "切换工作空间失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"auth_workspace_id":   workspace.ID.String(),
		"auth_workspace_type": workspace.WorkspaceType,
		"current_collaboration_workspace_id": func() string {
			if workspace.WorkspaceType == models.WorkspaceTypeCollaboration {
				return uuidPtrToString(workspace.CollaborationWorkspaceID)
			}
			return ""
		}(),
		"workspace":                  workspaceToSummary(workspace),
		"collaboration_workspace_id": uuidPtrToString(workspace.CollaborationWorkspaceID),
	}))
}

func (h *Handler) mustUserID(c *gin.Context) (uuid.UUID, error) {
	value := strings.TrimSpace(c.GetString("user_id"))
	if value == "" {
		return uuid.Nil, gorm.ErrRecordNotFound
	}
	return uuid.Parse(value)
}

func (h *Handler) resolveCurrentWorkspace(c *gin.Context, userID uuid.UUID) (*models.Workspace, error) {
	authWorkspaceID := strings.TrimSpace(c.GetString("auth_workspace_id"))
	if authWorkspaceID != "" {
		if parsedID, err := uuid.Parse(authWorkspaceID); err == nil {
			return h.getAccessibleWorkspace(userID, parsedID)
		}
	}
	return h.service.EnsurePersonalWorkspaceForUser(userID)
}

func (h *Handler) getAccessibleWorkspace(userID, workspaceID uuid.UUID) (*models.Workspace, error) {
	if _, err := h.service.GetMember(workspaceID, userID); err != nil {
		return nil, err
	}
	return h.service.GetByID(workspaceID)
}

func uuidPtrToString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}
