package workspacefeaturebinding

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
	appctx "github.com/maben/backend/internal/pkg/appctx"
	"github.com/maben/backend/internal/pkg/workspacerolebinding"
)

func ListPersonalPackageIDsByUserID(db *gorm.DB, userID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if db == nil || userID == uuid.Nil {
		return []uuid.UUID{}, nil
	}
	workspace, err := workspacerolebinding.GetPersonalWorkspaceByUserID(db, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}
	return listPackageIDsByWorkspaceID(db, workspace.ID, appKey)
}

func ListCollaborationWorkspacePackageIDsByCollaborationWorkspaceID(db *gorm.DB, collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if db == nil || collaborationWorkspaceID == uuid.Nil {
		return []uuid.UUID{}, nil
	}
	workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(db, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []uuid.UUID{}, nil
		}
		return nil, err
	}
	return listPackageIDsByWorkspaceID(db, workspace.ID, appKey)
}

func ReplacePersonalPackageBindings(tx *gorm.DB, userID uuid.UUID, appKey string, packageIDs []uuid.UUID) error {
	if tx == nil || userID == uuid.Nil {
		return nil
	}
	workspace, err := workspacerolebinding.EnsurePersonalWorkspace(tx, userID)
	if err != nil {
		return err
	}
	return replacePackageBindingsByWorkspaceID(tx, workspace.ID, appKey, packageIDs)
}

func ReplaceCollaborationWorkspacePackageBindings(tx *gorm.DB, collaborationWorkspaceID uuid.UUID, appKey string, packageIDs []uuid.UUID) error {
	if tx == nil || collaborationWorkspaceID == uuid.Nil {
		return nil
	}
	workspace, err := workspacerolebinding.GetCollaborationWorkspaceByCollaborationWorkspaceID(tx, collaborationWorkspaceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return replacePackageBindingsByWorkspaceID(tx, workspace.ID, appKey, packageIDs)
}

func ListPlatformUserIDsByPackageIDs(db *gorm.DB, packageIDs []uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if db == nil || len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var userIDs []uuid.UUID
	query := db.Model(&models.WorkspaceFeaturePackage{}).
		Select("workspaces.owner_user_id").
		Joins("JOIN workspaces ON workspaces.id = workspace_feature_packages.workspace_id").
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.package_id IN ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", dedupeUUIDs(packageIDs), true).
		Where("workspaces.workspace_type = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypePersonal).
		Where("workspaces.owner_user_id IS NOT NULL")
	if strings.TrimSpace(appKey) != "" {
		query = query.Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", appctx.NormalizeAppKey(appKey))
	} else {
		query = query.Where("feature_packages.deleted_at IS NULL")
	}
	err := query.
		Distinct("workspaces.owner_user_id").
		Pluck("workspaces.owner_user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(userIDs), nil
}

func ListCollaborationWorkspaceIDsByPackageIDs(db *gorm.DB, packageIDs []uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if db == nil || len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	type row struct {
		CollaborationWorkspaceID *uuid.UUID `gorm:"column:collaboration_workspace_id"`
	}
	rows := make([]row, 0)
	query := db.Model(&models.WorkspaceFeaturePackage{}).
		Select("workspaces.collaboration_workspace_id").
		Joins("JOIN workspaces ON workspaces.id = workspace_feature_packages.workspace_id").
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.package_id IN ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", dedupeUUIDs(packageIDs), true).
		Where("workspaces.workspace_type = ? AND workspaces.deleted_at IS NULL", models.WorkspaceTypeCollaboration)
	if strings.TrimSpace(appKey) != "" {
		query = query.Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", appctx.NormalizeAppKey(appKey))
	} else {
		query = query.Where("feature_packages.deleted_at IS NULL")
	}
	err := query.Find(&rows).Error
	if err != nil {
		return nil, err
	}
	collaborationWorkspaceIDs := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, item := range rows {
		if item.CollaborationWorkspaceID == nil || *item.CollaborationWorkspaceID == uuid.Nil {
			continue
		}
		if _, ok := seen[*item.CollaborationWorkspaceID]; ok {
			continue
		}
		seen[*item.CollaborationWorkspaceID] = struct{}{}
		collaborationWorkspaceIDs = append(collaborationWorkspaceIDs, *item.CollaborationWorkspaceID)
	}
	return collaborationWorkspaceIDs, nil
}

func ListPackageIDsByCollaborationWorkspaceID(db *gorm.DB, collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	return ListCollaborationWorkspacePackageIDsByCollaborationWorkspaceID(db, collaborationWorkspaceID, appKey)
}

func listPackageIDsByWorkspaceID(db *gorm.DB, workspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	if db == nil || workspaceID == uuid.Nil {
		return []uuid.UUID{}, nil
	}
	var packageIDs []uuid.UUID
	query := db.Model(&models.WorkspaceFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
		Where("workspace_feature_packages.workspace_id = ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", workspaceID, true).
		Where("feature_packages.deleted_at IS NULL")
	if strings.TrimSpace(appKey) != "" {
		query = query.Where("feature_packages.app_key = ?", appctx.NormalizeAppKey(appKey))
	}
	err := query.Distinct("workspace_feature_packages.package_id").
		Pluck("workspace_feature_packages.package_id", &packageIDs).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(packageIDs), nil
}

func replacePackageBindingsByWorkspaceID(tx *gorm.DB, workspaceID uuid.UUID, appKey string, packageIDs []uuid.UUID) error {
	scopedIDs, err := filterPackageIDs(tx, appKey, packageIDs)
	if err != nil {
		return err
	}
	existingIDs, err := listPackageIDsByWorkspaceID(tx, workspaceID, appKey)
	if err != nil {
		return err
	}
	return tx.Transaction(func(inner *gorm.DB) error {
		if len(existingIDs) > 0 {
			if err := inner.Where("workspace_id = ? AND package_id IN ?", workspaceID, existingIDs).Delete(&models.WorkspaceFeaturePackage{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		items := make([]models.WorkspaceFeaturePackage, 0, len(scopedIDs))
		for _, packageID := range scopedIDs {
			items = append(items, models.WorkspaceFeaturePackage{
				WorkspaceID: workspaceID,
				PackageID:   packageID,
				Enabled:     true,
			})
		}
		return inner.Create(&items).Error
	})
}

func filterPackageIDs(db *gorm.DB, appKey string, packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	if db == nil || len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var ids []uuid.UUID
	err := db.Model(&models.FeaturePackage{}).
		Where("app_key = ? AND id IN ? AND deleted_at IS NULL", appctx.NormalizeAppKey(appKey), dedupeUUIDs(packageIDs)).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, err
	}
	return dedupeUUIDs(ids), nil
}

func dedupeUUIDs(items []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
