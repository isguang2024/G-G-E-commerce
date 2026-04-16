// cw_roles.go — ogen handler implementations for CW role CRUD.
package handlers

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/api/gen"
	"github.com/maben/backend/internal/modules/system/user"
)

// ─── ListCurrentCollaborationWorkspaceRoles ───────────────────────────────────

func (h *APIHandler) ListCurrentCollaborationWorkspaceRoles(ctx context.Context) (*gen.CollaborationWorkspaceRoleList, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return &gen.CollaborationWorkspaceRoleList{Records: []gen.CollaborationWorkspaceRoleItem{}, Total: 0}, nil
	}
	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(member.CollaborationWorkspaceID)
	if err != nil {
		h.logger.Error("list cw roles failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceRoleList{Records: collaborationWorkspaceRoleItemsFromModels(allRoles), Total: int64(len(allRoles))}, nil
}

// ─── CreateCurrentCollaborationWorkspaceRole ──────────────────────────────────

func (h *APIHandler) CreateCurrentCollaborationWorkspaceRole(ctx context.Context, req *gen.CollaborationWorkspaceRoleSaveRequest) (*gen.MutationResult, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, errText("请求体不能为空")
	}
	code := strings.TrimSpace(req.Code)
	if code == "" || strings.TrimSpace(req.Name) == "" {
		return nil, errText("角色编码和角色名称不能为空")
	}
	existingRoles, err := h.roleRepo.FindByCode(code)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingRoles {
		if existing.CollaborationWorkspaceID != nil && *existing.CollaborationWorkspaceID == member.CollaborationWorkspaceID {
			return nil, errText("角色编码已存在")
		}
	}
	role := &user.Role{
		CollaborationWorkspaceID: &member.CollaborationWorkspaceID,
		Code:                     code,
		Name:                     strings.TrimSpace(req.Name),
		Description:              optString(req.Description),
		SortOrder:                optInt(req.SortOrder, 0),
		Status:                   cwNormalizeStatus(optString(req.Status)),
	}
	if err := h.roleRepo.Create(role); err != nil {
		h.logger.Error("create cw role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── ListCurrentCollaborationWorkspaceBoundaryRoles ───────────────────────────

func (h *APIHandler) ListCurrentCollaborationWorkspaceBoundaryRoles(ctx context.Context) (*gen.CollaborationWorkspaceRoleList, error) {
	return h.ListCurrentCollaborationWorkspaceRoles(ctx)
}

// ─── CreateCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) CreateCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, req *gen.CollaborationWorkspaceRoleSaveRequest) (*gen.MutationResult, error) {
	return h.CreateCurrentCollaborationWorkspaceRole(ctx, req)
}

// ─── UpdateCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) UpdateCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, req *gen.CollaborationWorkspaceRoleSaveRequest, params gen.UpdateCurrentCollaborationWorkspaceBoundaryRoleParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"name":        strings.TrimSpace(cwDefaultString(req.Name, role.Name)),
		"description": strings.TrimSpace(cwDefaultString(optString(req.Description), role.Description)),
		"sort_order":  optInt(req.SortOrder, int(role.SortOrder)),
		"status":      cwNormalizeStatus(cwDefaultString(optString(req.Status), role.Status)),
	}
	if code := strings.TrimSpace(req.Code); code != "" && code != role.Code {
		existingRoles, findErr := h.roleRepo.FindByCode(code)
		if findErr != nil {
			return nil, findErr
		}
		for _, existing := range existingRoles {
			if existing.ID == role.ID {
				continue
			}
			if existing.CollaborationWorkspaceID != nil && *existing.CollaborationWorkspaceID == member.CollaborationWorkspaceID {
				return nil, errText("角色编码已存在")
			}
		}
		updates["code"] = code
	}
	if err := h.roleRepo.UpdateWithMap(role.ID, updates); err != nil {
		h.logger.Error("update cw role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── DeleteCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) DeleteCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, params gen.DeleteCurrentCollaborationWorkspaceBoundaryRoleParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ? AND collaboration_workspace_id = ?", role.ID, member.CollaborationWorkspaceID).
			Delete(&user.UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleFeaturePackage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleHiddenMenu{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDisabledAction{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", role.ID).Delete(&user.RoleDataPermission{}).Error; err != nil {
			return err
		}
		return tx.Delete(&user.Role{}, role.ID).Error
	}); err != nil {
		h.logger.Error("delete cw role failed", zap.Error(err))
		return nil, err
	}
	return ok(), nil
}

// ─── ListCollaborationWorkspaceRoles ──────────────────────────────────────────

func (h *APIHandler) ListCollaborationWorkspaceRoles(ctx context.Context, params gen.ListCollaborationWorkspaceRolesParams) (*gen.CollaborationWorkspaceRoleList, error) {
	allRoles, err := h.roleRepo.ListCollaborationWorkspaceRoles(params.ID)
	if err != nil {
		h.logger.Error("list cw roles failed", zap.Error(err))
		return nil, err
	}
	return &gen.CollaborationWorkspaceRoleList{Records: collaborationWorkspaceRoleItemsFromModels(allRoles), Total: int64(len(allRoles))}, nil
}

