// cwcurrent.go — Phase 4 ogen handlers for current/my collaboration workspace
// basic operations (list, get, members CRUD). Hooks into cwSvc + cwMemberRepo.
package handlers

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

// resolveCurrentCwID returns the current collaboration workspace ID for the
// authenticated user, preferring the JWT-carried value and falling back to the
// member repo lookup.
func (h *APIHandler) resolveCurrentCwID(ctx context.Context) (uuid.UUID, error) {
	if raw := stringFromCtx(ctx, CtxCollaborationWorkspaceID); raw != "" {
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

func (h *APIHandler) ListMyCollaborationWorkspaces(ctx context.Context) (*gen.AnyListResponse, error) {
	uid, ok := userIDFromContext(ctx)
	if !ok {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
	}
	items, err := h.cwMemberRepo.GetCollaborationWorkspacesByUserID(uid)
	if err != nil {
		h.logger.Error("list my collaboration workspaces failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(items), Total: len(items)}, nil
}

func (h *APIHandler) GetCurrentCollaborationWorkspace(ctx context.Context) (gen.AnyObject, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return gen.AnyObject{}, nil
	}
	cw, err := h.cwSvc.Get(cwID)
	if err != nil {
		h.logger.Error("get current collaboration workspace failed", zap.Error(err))
		return nil, err
	}
	return marshalAnyObject(cw), nil
}

func (h *APIHandler) ListCurrentCollaborationWorkspaceMembers(ctx context.Context) (*gen.AnyListResponse, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return &gen.AnyListResponse{Records: []gen.AnyObject{}, Total: 0}, nil
	}
	members, err := h.cwSvc.ListMembers(cwID, &user.MemberSearchParams{})
	if err != nil {
		h.logger.Error("list current cw members failed", zap.Error(err))
		return nil, err
	}
	return &gen.AnyListResponse{Records: marshalList(members), Total: len(members)}, nil
}

func (h *APIHandler) AddCurrentCollaborationWorkspaceMember(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return nil, err
	}
	var body dto.CollaborationWorkspaceAddMemberRequest
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
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

func (h *APIHandler) RemoveCurrentCollaborationWorkspaceMember(ctx context.Context, params gen.RemoveCurrentCollaborationWorkspaceMemberParams) (*gen.MutationResult, error) {
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

func (h *APIHandler) UpdateCurrentCollaborationWorkspaceMemberRole(ctx context.Context, req gen.AnyObject, params gen.UpdateCurrentCollaborationWorkspaceMemberRoleParams) (*gen.MutationResult, error) {
	cwID, err := h.resolveCurrentCwID(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		RoleCode string `json:"role_code"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	if err := h.cwSvc.UpdateMemberRole(cwID, params.UserId, body.RoleCode); err != nil {
		h.logger.Error("update current cw member role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}
