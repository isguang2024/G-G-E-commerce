// cw_member.go — ogen handler implementations for CW member role management.
package handlers

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/user"
)

// ─── GetCurrentCollaborationWorkspaceMemberRoles ──────────────────────────────

func (h *APIHandler) GetCurrentCollaborationWorkspaceMemberRoles(ctx context.Context, params gen.GetCurrentCollaborationWorkspaceMemberRolesParams) (*gen.CollaborationWorkspaceMemberRolesResponse, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationWorkspaceMemberRolesResponse{
			RoleIds: []uuid.UUID{},
			Roles:   []gen.UserRoleRef{},
		}, nil
	}
	targetMember, err := h.cwMemberRepo.GetByUserAndCollaborationWorkspace(params.UserId, member.CollaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errText("成员不存在")
		}
		return nil, err
	}
	roleIDs, err := h.cwGetWorkspaceAwareTeamRoleIDs(params.UserId, member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("get cw member role ids failed", zap.Error(err))
		return nil, err
	}
	roles, err := h.roleRepo.GetByIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	_ = targetMember // used for binding meta if needed
	return &gen.CollaborationWorkspaceMemberRolesResponse{
		RoleIds: roleIDs,
		Roles:   roleRefsFromModels(roles),
	}, nil
}

// ─── SetCurrentCollaborationWorkspaceMemberRoles ──────────────────────────────

func (h *APIHandler) SetCurrentCollaborationWorkspaceMemberRoles(ctx context.Context, req *gen.UUIDListRequest, params gen.SetCurrentCollaborationWorkspaceMemberRolesParams) (*gen.MutationResult, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return nil, err
	}
	targetMember, err := h.cwMemberRepo.GetByUserAndCollaborationWorkspace(params.UserId, member.CollaborationWorkspaceID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errText("成员不存在")
		}
		return nil, err
	}
	roleIDs := uuidIDsFromRequest(req)

	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(member.CollaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	allowedTeamRoleIDs := make(map[uuid.UUID]user.Role)
	protectedRoleID := uuid.Nil
	for _, role := range allRoles {
		allowedTeamRoleIDs[role.ID] = role
		if role.Code == targetMember.RoleCode {
			protectedRoleID = role.ID
		}
	}
	filteredIDs := make([]uuid.UUID, 0, len(roleIDs)+1)
	seen := make(map[uuid.UUID]struct{}, len(roleIDs)+1)
	for _, roleID := range roleIDs {
		if _, ok := allowedTeamRoleIDs[roleID]; !ok {
			continue
		}
		if _, dup := seen[roleID]; dup {
			continue
		}
		seen[roleID] = struct{}{}
		filteredIDs = append(filteredIDs, roleID)
	}
	if protectedRoleID != uuid.Nil {
		if _, exists := seen[protectedRoleID]; !exists {
			filteredIDs = append(filteredIDs, protectedRoleID)
		}
	}

	if err := h.userRoleRepo.SetUserRoles(params.UserId, filteredIDs, &member.CollaborationWorkspaceID); err != nil {
		h.logger.Error("set cw member roles failed", zap.Error(err))
		return nil, err
	}
	if err := h.cwSyncWorkspaceRoleBindings(member.CollaborationWorkspaceID, params.UserId, filteredIDs); err != nil {
		h.logger.Error("sync cw role bindings failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

