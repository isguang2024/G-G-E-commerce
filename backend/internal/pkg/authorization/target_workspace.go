package authorization

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func (s *Service) RequirePersonalWorkspaceTargetWorkspace(authCtx *AuthorizationContext, workspaceID uuid.UUID) (*models.Workspace, error) {
	if err := ensureWorkspaceContext(authCtx); err != nil {
		return nil, err
	}
	if authCtx.AuthWorkspaceType != models.WorkspaceTypePersonal {
		return nil, ErrWorkspaceTypeForbidden
	}
	if workspaceID == uuid.Nil {
		return nil, ErrTargetWorkspaceRequired
	}

	var workspace models.Workspace
	if err := s.db.
		Where("id = ? AND deleted_at IS NULL", workspaceID).
		First(&workspace).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetWorkspaceForbidden
		}
		return nil, err
	}
	if workspace.WorkspaceType != models.WorkspaceTypeCollaboration || workspace.Status != models.WorkspaceStatusActive {
		return nil, ErrTargetWorkspaceForbidden
	}

	var membershipCount int64
	if err := s.db.Model(&models.WorkspaceMember{}).
		Where("workspace_id = ? AND user_id = ? AND status = ? AND deleted_at IS NULL", workspace.ID, authCtx.UserID, models.WorkspaceStatusActive).
		Count(&membershipCount).Error; err != nil {
		return nil, err
	}
	if membershipCount == 0 {
		return nil, ErrTargetWorkspaceForbidden
	}
	return &workspace, nil
}

func (s *Service) RequirePersonalWorkspaceTargetWorkspaces(authCtx *AuthorizationContext, collaborationWorkspaceIDs []uuid.UUID) (map[uuid.UUID]*models.Workspace, error) {
	if err := ensureWorkspaceContext(authCtx); err != nil {
		return nil, err
	}
	if authCtx.AuthWorkspaceType != models.WorkspaceTypePersonal {
		return nil, ErrWorkspaceTypeForbidden
	}
	uniqueCollaborationWorkspaceIDs := make([]uuid.UUID, 0, len(collaborationWorkspaceIDs))
	seen := make(map[uuid.UUID]struct{}, len(collaborationWorkspaceIDs))
	for _, collaborationWorkspaceID := range collaborationWorkspaceIDs {
		if collaborationWorkspaceID == uuid.Nil {
			return nil, ErrTargetWorkspaceRequired
		}
		if _, exists := seen[collaborationWorkspaceID]; exists {
			continue
		}
		seen[collaborationWorkspaceID] = struct{}{}
		uniqueCollaborationWorkspaceIDs = append(uniqueCollaborationWorkspaceIDs, collaborationWorkspaceID)
	}
	if len(uniqueCollaborationWorkspaceIDs) == 0 {
		return map[uuid.UUID]*models.Workspace{}, nil
	}

	workspaces := make([]models.Workspace, 0, len(uniqueCollaborationWorkspaceIDs))
	if err := s.db.
		Where("workspace_type = ? AND status = ? AND collaboration_workspace_id IN ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, models.WorkspaceStatusActive, uniqueCollaborationWorkspaceIDs).
		Find(&workspaces).Error; err != nil {
		return nil, err
	}
	if len(workspaces) != len(uniqueCollaborationWorkspaceIDs) {
		return nil, ErrTargetWorkspaceForbidden
	}

	result := make(map[uuid.UUID]*models.Workspace, len(workspaces))
	workspaceIDs := make([]uuid.UUID, 0, len(workspaces))
	for idx := range workspaces {
		workspace := &workspaces[idx]
		if workspace.CollaborationWorkspaceID == nil || *workspace.CollaborationWorkspaceID == uuid.Nil {
			return nil, ErrTargetWorkspaceForbidden
		}
		result[*workspace.CollaborationWorkspaceID] = workspace
		workspaceIDs = append(workspaceIDs, workspace.ID)
	}
	if len(result) != len(uniqueCollaborationWorkspaceIDs) {
		return nil, ErrTargetWorkspaceForbidden
	}

	var membershipCount int64
	if err := s.db.Model(&models.WorkspaceMember{}).
		Where("workspace_id IN ? AND user_id = ? AND status = ? AND deleted_at IS NULL", workspaceIDs, authCtx.UserID, models.WorkspaceStatusActive).
		Count(&membershipCount).Error; err != nil {
		return nil, err
	}
	if membershipCount != int64(len(workspaceIDs)) {
		return nil, ErrTargetWorkspaceForbidden
	}
	return result, nil
}
