// Package handlers contains ogen Handler implementations for the v5
// OpenAPI-first API. Each handler is the single entry point for one
// generated operation interface; legacy Gin handlers are removed as
// each domain migrates over.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/workspace"
)

// ctxKey is the request-scoped key carrying the authenticated account id
// from the Gin layer into the ogen handler. The router middleware seeds it
// before handing the request to the generated server.
type ctxKey string

const CtxUserID ctxKey = "user_id"

// WorkspaceHandler implements gen.Handler. It deliberately embeds
// gen.UnimplementedHandler so future operations compile without forcing us
// to stub every method while migrating one domain at a time.
type WorkspaceHandler struct {
	gen.UnimplementedHandler
	logger  *zap.Logger
	service workspace.Service
}

func NewWorkspaceHandler(db *gorm.DB, logger *zap.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		logger:  logger,
		service: workspace.NewService(db, logger),
	}
}

func (h *WorkspaceHandler) GetWorkspace(ctx context.Context, params gen.GetWorkspaceParams) (gen.GetWorkspaceRes, error) {
	userID, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.GetWorkspaceForbidden{Code: 401, Message: "未认证"}, nil
	}

	if _, err := h.service.GetMember(params.ID, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.GetWorkspaceForbidden{Code: 403, Message: "无权访问该工作空间"}, nil
		}
		h.logger.Error("workspace member lookup failed", zap.Error(err))
		return nil, err
	}

	ws, err := h.service.GetByID(params.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &gen.GetWorkspaceNotFound{Code: 404, Message: "工作空间不存在"}, nil
		}
		h.logger.Error("workspace lookup failed", zap.Error(err))
		return nil, err
	}

	return mapWorkspaceToSummary(ws), nil
}

func userIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	raw, ok := ctx.Value(CtxUserID).(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func mapWorkspaceToSummary(ws *models.Workspace) *gen.WorkspaceSummary {
	out := &gen.WorkspaceSummary{
		ID:            ws.ID,
		WorkspaceType: gen.WorkspaceSummaryWorkspaceType(ws.WorkspaceType),
		Name:          ws.Name,
		Code:          ws.Code,
		Status:        ws.Status,
	}
	if ws.OwnerUserID != nil {
		out.OwnerUserID = gen.NewOptNilUUID(*ws.OwnerUserID)
	}
	if ws.CollaborationWorkspaceID != nil {
		out.CollaborationWorkspaceID = gen.NewOptNilUUID(*ws.CollaborationWorkspaceID)
	}
	return out
}
