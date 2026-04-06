package navigation

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
)

type Handler struct {
	logger   *zap.Logger
	compiler Compiler
}

func NewHandler(logger *zap.Logger, compiler Compiler) *Handler {
	return &Handler{logger: logger, compiler: compiler}
}

func (h *Handler) GetNavigation(c *gin.Context) {
	manifest, err := h.compiler.Compile(
		apppkg.CurrentAppKey(c),
		spacepkg.RequestHost(c),
		spacepkg.RequestSpaceKey(c),
		currentUserID(c),
		currentCollaborationWorkspaceID(c),
	)
	if err != nil {
		h.logger.Error("Compile runtime navigation failed", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取运行时导航清单失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(manifest))
}

func currentUserID(c *gin.Context) *uuid.UUID {
	raw, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	value, ok := raw.(string)
	if !ok {
		return nil
	}
	parsedID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return nil
	}
	return &parsedID
}

func currentCollaborationWorkspaceID(c *gin.Context) *uuid.UUID {
	if value := strings.TrimSpace(c.GetString("collaboration_workspace_id")); value != "" {
		if parsedID, err := uuid.Parse(value); err == nil {
			return &parsedID
		}
	}
	if headerValue := strings.TrimSpace(c.GetHeader("X-Collaboration-Workspace-Id")); headerValue != "" {
		if parsedID, err := uuid.Parse(headerValue); err == nil {
			return &parsedID
		}
	}
	return nil
}
