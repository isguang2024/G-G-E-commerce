// cw_roles.go — ogen handler implementations for CW role CRUD.
package handlers

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
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

func (h *APIHandler) CreateCurrentCollaborationWorkspaceRole(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	member, err := h.resolveCWMember(ctx)
	if err != nil {
		return nil, err
	}
	var body struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
		Priority    int    `json:"priority"`
		Status      string `json:"status"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	code := strings.TrimSpace(body.Code)
	if code == "" || strings.TrimSpace(body.Name) == "" {
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
		Name:                     strings.TrimSpace(body.Name),
		Description:              strings.TrimSpace(body.Description),
		SortOrder:                body.SortOrder,
		Priority:                 body.Priority,
		Status:                   cwNormalizeStatus(body.Status),
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

func (h *APIHandler) CreateCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, req gen.AnyObject) (*gen.MutationResult, error) {
	return h.CreateCurrentCollaborationWorkspaceRole(ctx, req)
}

// ─── UpdateCurrentCollaborationWorkspaceBoundaryRole ─────────────────────────

func (h *APIHandler) UpdateCurrentCollaborationWorkspaceBoundaryRole(ctx context.Context, req gen.AnyObject, params gen.UpdateCurrentCollaborationWorkspaceBoundaryRoleParams) (*gen.MutationResult, error) {
	member, role, err := h.resolveCWRoleEditable(ctx, params.RoleId)
	if err != nil {
		return nil, err
	}
	var body struct {
		Code        string `json:"code"`
		Name        string `json:"name"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
		Priority    int    `json:"priority"`
		Status      string `json:"status"`
	}
	if err := unmarshalAnyObject(req, &body); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{
		"name":        strings.TrimSpace(cwDefaultString(body.Name, role.Name)),
		"description": strings.TrimSpace(cwDefaultString(body.Description, role.Description)),
		"sort_order":  body.SortOrder,
		"priority":    body.Priority,
		"status":      cwNormalizeStatus(cwDefaultString(body.Status, role.Status)),
	}
	if code := strings.TrimSpace(body.Code); code != "" && code != role.Code {
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
