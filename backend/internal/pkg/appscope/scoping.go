package appscope

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/workspacefeaturebinding"
)

func Normalize(appKey string) string {
	return appctx.NormalizeAppKey(appKey)
}

func FilterPackageIDs(db *gorm.DB, appKey string, packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var ids []uuid.UUID
	err := db.Model(&models.FeaturePackage{}).
		Where("app_key = ? AND id IN ? AND deleted_at IS NULL", Normalize(appKey), dedupeUUIDs(packageIDs)).
		Pluck("id", &ids).Error
	return dedupeUUIDs(ids), err
}

func FilterMenuIDs(db *gorm.DB, appKey string, menuIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(menuIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var ids []uuid.UUID
	err := db.Model(&models.MenuDefinition{}).
		Where("app_key = ? AND id IN ? AND deleted_at IS NULL", Normalize(appKey), dedupeUUIDs(menuIDs)).
		Pluck("id", &ids).Error
	return dedupeUUIDs(ids), err
}

func PackageIDsByRole(db *gorm.DB, roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := db.Model(&models.RoleFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = role_feature_packages.package_id").
		Where("role_feature_packages.role_id = ? AND role_feature_packages.enabled = ?", roleID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", Normalize(appKey)).
		Pluck("role_feature_packages.package_id", &packageIDs).Error
	return dedupeUUIDs(packageIDs), err
}

func PackageIDsByUser(db *gorm.DB, userID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	workspacePackageIDs, err := workspacefeaturebinding.ListPersonalPackageIDsByUserID(db, userID, appKey)
	if err != nil {
		return nil, err
	}
	if len(workspacePackageIDs) > 0 {
		return dedupeUUIDs(workspacePackageIDs), nil
	}
	var packageIDs []uuid.UUID
	err = db.Model(&models.UserFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = user_feature_packages.package_id").
		Where("user_feature_packages.user_id = ? AND user_feature_packages.enabled = ?", userID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", Normalize(appKey)).
		Pluck("user_feature_packages.package_id", &packageIDs).Error
	return dedupeUUIDs(packageIDs), err
}

func PackageIDsByTeam(db *gorm.DB, teamID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	workspacePackageIDs, err := workspacefeaturebinding.ListTeamPackageIDsByCollaborationWorkspaceID(db, teamID, appKey)
	if err != nil {
		return nil, err
	}
	if len(workspacePackageIDs) > 0 {
		return dedupeUUIDs(workspacePackageIDs), nil
	}
	var packageIDs []uuid.UUID
	err = db.Model(&models.CollaborationWorkspaceFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = collaboration_workspace_feature_packages.package_id").
		Where("collaboration_workspace_feature_packages.collaboration_workspace_id = ? AND collaboration_workspace_feature_packages.enabled = ?", teamID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", Normalize(appKey)).
		Pluck("collaboration_workspace_feature_packages.package_id", &packageIDs).Error
	return dedupeUUIDs(packageIDs), err
}

func HiddenMenuIDsByRole(db *gorm.DB, roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := db.Model(&models.RoleHiddenMenu{}).
		Joins("JOIN menu_definitions ON menu_definitions.id = role_hidden_menus.menu_id").
		Where("role_hidden_menus.role_id = ?", roleID).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", Normalize(appKey)).
		Pluck("role_hidden_menus.menu_id", &menuIDs).Error
	return dedupeUUIDs(menuIDs), err
}

func HiddenMenuIDsByUser(db *gorm.DB, userID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := db.Model(&models.UserHiddenMenu{}).
		Joins("JOIN menu_definitions ON menu_definitions.id = user_hidden_menus.menu_id").
		Where("user_hidden_menus.user_id = ?", userID).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", Normalize(appKey)).
		Pluck("user_hidden_menus.menu_id", &menuIDs).Error
	return dedupeUUIDs(menuIDs), err
}

func BlockedMenuIDsByTeam(db *gorm.DB, teamID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := db.Model(&models.CollaborationWorkspaceBlockedMenu{}).
		Joins("JOIN menu_definitions ON menu_definitions.id = collaboration_workspace_blocked_menus.menu_id").
		Where("collaboration_workspace_blocked_menus.collaboration_workspace_id = ?", teamID).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", Normalize(appKey)).
		Pluck("collaboration_workspace_blocked_menus.menu_id", &menuIDs).Error
	return dedupeUUIDs(menuIDs), err
}

func ReplaceRolePackagesInApp(db *gorm.DB, roleID uuid.UUID, appKey string, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	scopedIDs, err := FilterPackageIDs(db, appKey, packageIDs)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		existingIDs, err := PackageIDsByRole(tx, roleID, appKey)
		if err != nil {
			return err
		}
		if len(existingIDs) > 0 {
			if err := tx.Where("role_id = ? AND package_id IN ?", roleID, existingIDs).Delete(&models.RoleFeaturePackage{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		now := time.Now()
		items := make([]models.RoleFeaturePackage, 0, len(scopedIDs))
		for _, packageID := range scopedIDs {
			items = append(items, models.RoleFeaturePackage{
				RoleID:    roleID,
				PackageID: packageID,
				Enabled:   true,
				GrantedBy: grantedBy,
				GrantedAt: &now,
			})
		}
		return tx.Create(&items).Error
	})
}

func ReplaceUserPackagesInApp(db *gorm.DB, userID uuid.UUID, appKey string, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	scopedIDs, err := FilterPackageIDs(db, appKey, packageIDs)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := workspacefeaturebinding.ReplacePersonalPackageBindings(tx, userID, appKey, scopedIDs); err != nil {
			return err
		}
		existingIDs, err := legacyPackageIDsByUser(tx, userID, appKey)
		if err != nil {
			return err
		}
		if len(existingIDs) > 0 {
			if err := tx.Where("user_id = ? AND package_id IN ?", userID, existingIDs).Delete(&models.UserFeaturePackage{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		now := time.Now()
		items := make([]models.UserFeaturePackage, 0, len(scopedIDs))
		for _, packageID := range scopedIDs {
			items = append(items, models.UserFeaturePackage{
				UserID:    userID,
				PackageID: packageID,
				Enabled:   true,
				GrantedBy: grantedBy,
				GrantedAt: &now,
			})
		}
		return tx.Create(&items).Error
	})
}

func ReplaceTeamPackagesInApp(db *gorm.DB, teamID uuid.UUID, appKey string, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	scopedIDs, err := FilterPackageIDs(db, appKey, packageIDs)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		existingIDs, err := legacyPackageIDsByTeam(tx, teamID, appKey)
		if err != nil {
			return err
		}
		if len(existingIDs) > 0 {
			if err := tx.Where("collaboration_workspace_id = ? AND package_id IN ?", teamID, existingIDs).Delete(&models.CollaborationWorkspaceFeaturePackage{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		now := time.Now()
		items := make([]models.CollaborationWorkspaceFeaturePackage, 0, len(scopedIDs))
		for _, packageID := range scopedIDs {
			items = append(items, models.CollaborationWorkspaceFeaturePackage{
				CollaborationWorkspaceID: teamID,
				PackageID:                packageID,
				Enabled:                  true,
				GrantedBy:                grantedBy,
				GrantedAt:                &now,
			})
		}
		return tx.Create(&items).Error
	})
}

func ReplaceRoleHiddenMenusInApp(db *gorm.DB, roleID uuid.UUID, appKey string, hiddenMenuIDs []uuid.UUID) error {
	scopedIDs, err := FilterMenuIDs(db, appKey, hiddenMenuIDs)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		existingIDs, err := HiddenMenuIDsByRole(tx, roleID, appKey)
		if err != nil {
			return err
		}
		if len(existingIDs) > 0 {
			if err := tx.Where("role_id = ? AND menu_id IN ?", roleID, existingIDs).Delete(&models.RoleHiddenMenu{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		items := make([]models.RoleHiddenMenu, 0, len(scopedIDs))
		for _, menuID := range scopedIDs {
			items = append(items, models.RoleHiddenMenu{RoleID: roleID, MenuID: menuID})
		}
		return tx.Create(&items).Error
	})
}

func ReplaceUserHiddenMenusInApp(db *gorm.DB, userID uuid.UUID, appKey string, hiddenMenuIDs []uuid.UUID) error {
	scopedIDs, err := FilterMenuIDs(db, appKey, hiddenMenuIDs)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		existingIDs, err := HiddenMenuIDsByUser(tx, userID, appKey)
		if err != nil {
			return err
		}
		if len(existingIDs) > 0 {
			if err := tx.Where("user_id = ? AND menu_id IN ?", userID, existingIDs).Delete(&models.UserHiddenMenu{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		items := make([]models.UserHiddenMenu, 0, len(scopedIDs))
		for _, menuID := range scopedIDs {
			items = append(items, models.UserHiddenMenu{UserID: userID, MenuID: menuID})
		}
		return tx.Create(&items).Error
	})
}

func ReplaceCollaborationWorkspaceBlockedMenusInApp(db *gorm.DB, teamID uuid.UUID, appKey string, blockedMenuIDs []uuid.UUID) error {
	scopedIDs, err := FilterMenuIDs(db, appKey, blockedMenuIDs)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		existingIDs, err := BlockedMenuIDsByTeam(tx, teamID, appKey)
		if err != nil {
			return err
		}
		if len(existingIDs) > 0 {
			if err := tx.Where("collaboration_workspace_id = ? AND menu_id IN ?", teamID, existingIDs).Delete(&models.CollaborationWorkspaceBlockedMenu{}).Error; err != nil {
				return err
			}
		}
		if len(scopedIDs) == 0 {
			return nil
		}
		items := make([]models.CollaborationWorkspaceBlockedMenu, 0, len(scopedIDs))
		for _, menuID := range scopedIDs {
			items = append(items, models.CollaborationWorkspaceBlockedMenu{CollaborationWorkspaceID: teamID, MenuID: menuID})
		}
		return tx.Create(&items).Error
	})
}

func ReplaceRoleDisabledActionsInScope(db *gorm.DB, roleID uuid.UUID, scopedActionIDs []uuid.UUID, disabledActionIDs []uuid.UUID) error {
	return replaceRoleActionsInScope(db, roleID, scopedActionIDs, disabledActionIDs)
}

func ReplaceCollaborationWorkspaceBlockedActionsInScope(db *gorm.DB, teamID uuid.UUID, scopedActionIDs []uuid.UUID, blockedActionIDs []uuid.UUID) error {
	return replaceTeamActionsInScope(db, teamID, scopedActionIDs, blockedActionIDs)
}

func replaceRoleActionsInScope(db *gorm.DB, roleID uuid.UUID, scopedActionIDs []uuid.UUID, targetActionIDs []uuid.UUID) error {
	scoped := dedupeUUIDs(scopedActionIDs)
	target := intersectUUIDs(dedupeUUIDs(targetActionIDs), scoped)
	return db.Transaction(func(tx *gorm.DB) error {
		if len(scoped) > 0 {
			if err := tx.Where("role_id = ? AND action_id IN ?", roleID, scoped).Delete(&models.RoleDisabledAction{}).Error; err != nil {
				return err
			}
		}
		if len(target) == 0 {
			return nil
		}
		items := make([]models.RoleDisabledAction, 0, len(target))
		for _, actionID := range target {
			items = append(items, models.RoleDisabledAction{RoleID: roleID, ActionID: actionID})
		}
		return tx.Create(&items).Error
	})
}

func replaceTeamActionsInScope(db *gorm.DB, teamID uuid.UUID, scopedActionIDs []uuid.UUID, targetActionIDs []uuid.UUID) error {
	scoped := dedupeUUIDs(scopedActionIDs)
	target := intersectUUIDs(dedupeUUIDs(targetActionIDs), scoped)
	return db.Transaction(func(tx *gorm.DB) error {
		if len(scoped) > 0 {
			if err := tx.Where("collaboration_workspace_id = ? AND action_id IN ?", teamID, scoped).Delete(&models.CollaborationWorkspaceBlockedAction{}).Error; err != nil {
				return err
			}
		}
		if len(target) == 0 {
			return nil
		}
		items := make([]models.CollaborationWorkspaceBlockedAction, 0, len(target))
		for _, actionID := range target {
			items = append(items, models.CollaborationWorkspaceBlockedAction{CollaborationWorkspaceID: teamID, ActionID: actionID})
		}
		return tx.Create(&items).Error
	})
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

func legacyPackageIDsByUser(db *gorm.DB, userID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := db.Model(&models.UserFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = user_feature_packages.package_id").
		Where("user_feature_packages.user_id = ? AND user_feature_packages.enabled = ?", userID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", Normalize(appKey)).
		Pluck("user_feature_packages.package_id", &packageIDs).Error
	return dedupeUUIDs(packageIDs), err
}

func legacyPackageIDsByTeam(db *gorm.DB, teamID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := db.Model(&models.CollaborationWorkspaceFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = collaboration_workspace_feature_packages.package_id").
		Where("collaboration_workspace_feature_packages.collaboration_workspace_id = ? AND collaboration_workspace_feature_packages.enabled = ?", teamID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", Normalize(appKey)).
		Pluck("collaboration_workspace_feature_packages.package_id", &packageIDs).Error
	return dedupeUUIDs(packageIDs), err
}

func intersectUUIDs(left []uuid.UUID, right []uuid.UUID) []uuid.UUID {
	if len(left) == 0 || len(right) == 0 {
		return []uuid.UUID{}
	}
	rightSet := make(map[uuid.UUID]struct{}, len(right))
	for _, item := range right {
		rightSet[item] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(left))
	for _, item := range left {
		if _, ok := rightSet[item]; ok {
			result = append(result, item)
		}
	}
	return result
}
