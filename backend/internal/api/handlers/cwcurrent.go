// cwcurrent.go — Phase 4 ogen handlers for current/my collaboration workspace
// basic operations (list, get, members CRUD). Hooks into cwSvc + cwMemberRepo.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/api/dto"
	"github.com/maben/backend/internal/modules/system/user"
)

// resolveCurrentCwID returns the current collaboration workspace ID for the
// authenticated user, preferring the JWT-carried value and falling back to the
// member repo lookup.
func (h *cwAPIHandler) resolveCurrentCwID(ctx context.Context) (uuid.UUID, error) {
	if raw := stringFromCtx(ctx, CtxCollaborationID); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			return id, nil
		}
	}
	uid, ok := userIDFromContext(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthenticated")
	}
	m, err := h.cwMemberRepo.GetByUserID(uid)
	if err != nil {
		return uuid.Nil, err
	}
	return m.CollaborationWorkspaceID, nil
}

func (h *cwAPIHandler) ListMyCollaborations(ctx context.Context) (*gen.CollaborationList, error) {
	uid, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.CollaborationList{Records: []gen.CollaborationItem{}, Total: 0}, nil
	}
	items, err := h.cwMemberRepo.GetCollaborationWorkspacesByUserID(uid)
	if err != nil {
		h.logger.Error("list my collaboration workspaces failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationList{
		Records: collaborationWorkspaceItemsFromModels(items),
		Total:   len(items),
	}, nil
}

func (h *cwAPIHandler) GetCurrentCollaboration(ctx context.Context) (*gen.CollaborationItem, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return &gen.CollaborationItem{}, nil
	}
	cw, err := h.cwSvc.Get(cwID)
	if err != nil {
		h.logger.Error("get current collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	item := collaborationWorkspaceItemFromModel(*cw)
	return &item, nil
}

func (h *cwAPIHandler) ListCurrentCollaborationMembers(ctx context.Context) (*gen.CollaborationMemberList, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return &gen.CollaborationMemberList{
			Records: []gen.CollaborationMemberItem{},
			Total:   0,
		}, nil
	}
	members, err := h.cwSvc.ListMembers(cwID, &user.MemberSearchParams{})
	if err != nil {
		h.logger.Error("list current cw members failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationMemberList{
		Records: collaborationWorkspaceMemberItemsFromModels(members),
		Total:   len(members),
	}, nil
}

func (h *cwAPIHandler) AddCurrentCollaborationMember(ctx context.Context, req *gen.CollaborationMemberAddRequest) (*gen.MutationResult, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("request body required")
	}
	body := dto.CollaborationWorkspaceAddMemberRequest{
		UserID:   req.UserID.String(),
		RoleCode: optString(req.RoleCode),
	}
	var invitedBy *uuid.UUID
	if uid, ok := userIDFromContext(ctx); ok {
		invitedBy = &uid
	}
	if err := h.cwSvc.AddMember(cwID, &body, invitedBy); err != nil {
		h.logger.Error("add current cw member failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *cwAPIHandler) RemoveCurrentCollaborationMember(ctx context.Context, params gen.RemoveCurrentCollaborationMemberParams) (*gen.MutationResult, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.cwSvc.RemoveMember(cwID, params.UserId); err != nil {
		h.logger.Error("remove current cw member failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

func (h *cwAPIHandler) UpdateCurrentCollaborationMemberRole(ctx context.Context, req *gen.CollaborationMemberRoleRequest, params gen.UpdateCurrentCollaborationMemberRoleParams) (*gen.MutationResult, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errors.New("request body required")
	}
	if err := h.cwSvc.UpdateMemberRole(cwID, params.UserId, req.RoleCode); err != nil {
		return nil, err
	}
	return ok(), nil
}


