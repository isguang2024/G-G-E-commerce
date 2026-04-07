package permissionseed

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
	usermodel "github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func EnsureDefaultAPIEndpointCategories(db *gorm.DB) error {
	for _, seed := range DefaultAPIEndpointCategories() {
		var item usermodel.APIEndpointCategory
		result := db.Where("code = ?", seed.Code).First(&item)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				item = usermodel.APIEndpointCategory{
					ID:        seed.ID,
					Code:      seed.Code,
					Name:      seed.Name,
					NameEn:    seed.NameEn,
					SortOrder: seed.SortOrder,
					Status:    seed.Status,
				}
				if err := db.Create(&item).Error; err != nil {
					return err
				}
				continue
			}
			return result.Error
		}
		updates := map[string]interface{}{
			"name":       seed.Name,
			"name_en":    seed.NameEn,
			"sort_order": seed.SortOrder,
			"status":     seed.Status,
			"updated_at": time.Now(),
		}
		if err := db.Model(&item).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsureDefaultPermissionGroups(db *gorm.DB) error {
	for _, seed := range DefaultPermissionGroups() {
		item := usermodel.PermissionGroup{
			ID:          seed.ID,
			GroupType:   seed.GroupType,
			Code:        seed.Code,
			Name:        seed.Name,
			NameEn:      seed.NameEn,
			Description: seed.Description,
			Status:      seed.Status,
			SortOrder:   seed.SortOrder,
			IsBuiltin:   seed.IsBuiltin,
		}
		var existing usermodel.PermissionGroup
		result := db.Where("group_type = ? AND code = ?", item.GroupType, item.Code).First(&existing)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := db.Create(&item).Error; err != nil {
					return err
				}
				continue
			}
			return result.Error
		}
		updates := map[string]interface{}{
			"name":        item.Name,
			"name_en":     item.NameEn,
			"description": item.Description,
			"status":      item.Status,
			"sort_order":  item.SortOrder,
			"is_builtin":  item.IsBuiltin,
			"updated_at":  time.Now(),
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsureDefaultPermissionKeys(db *gorm.DB) error {
	groupIDs := make(map[string]uuid.UUID, len(DefaultPermissionGroups()))
	for _, seed := range DefaultPermissionGroups() {
		groupIDs[seed.GroupType+":"+seed.Code] = seed.ID
	}
	for _, actionSeed := range DefaultPermissionKeys() {
		moduleGroupID := groupIDs["module:"+actionSeed.ModuleGroupCode]
		featureGroupID := groupIDs["feature:"+actionSeed.FeatureGroupCode]
		actionData := usermodel.PermissionKey{
			ID:             actionSeed.ID,
			Code:           StableID("permission-action-code", actionSeed.Key).String(),
			PermissionKey:  actionSeed.Key,
			ModuleCode:     actionSeed.ModuleCode,
			ModuleGroupID:  &moduleGroupID,
			FeatureGroupID: &featureGroupID,
			ContextType:    actionSeed.ContextType,
			FeatureKind:    actionSeed.FeatureKind,
			Name:           actionSeed.Name,
			Description:    actionSeed.Description,
			Status:         actionSeed.Status,
			SortOrder:      actionSeed.SortOrder,
			IsBuiltin:      actionSeed.IsBuiltin,
		}
		var action usermodel.PermissionKey
		result := db.Where("permission_key = ?", actionData.PermissionKey).First(&action)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := db.Create(&actionData).Error; err != nil {
					return err
				}
				continue
			}
			return result.Error
		}
		updates := map[string]interface{}{
			"code":             actionData.Code,
			"module_code":      actionData.ModuleCode,
			"module_group_id":  actionData.ModuleGroupID,
			"feature_group_id": actionData.FeatureGroupID,
			"context_type":     actionData.ContextType,
			"feature_kind":     actionData.FeatureKind,
			"name":             actionData.Name,
			"description":      actionData.Description,
			"status":           actionData.Status,
			"sort_order":       actionData.SortOrder,
			"is_builtin":       actionData.IsBuiltin,
			"updated_at":       time.Now(),
		}
		if err := db.Model(&action).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsureDefaultFeaturePackages(db *gorm.DB) error {
	for _, seed := range DefaultFeaturePackages() {
		item := usermodel.FeaturePackage{
			ID:            seed.ID,
			AppKey:        systemmodels.DefaultAppKey,
			AppKeys:       seed.AppKeys,
			PackageKey:    seed.PackageKey,
			PackageType:   seed.PackageType,
			Name:          seed.Name,
			Description:   seed.Description,
			WorkspaceScope: seed.WorkspaceScope,
			ContextType:   seed.ContextType,
			IsBuiltin:     seed.IsBuiltin,
			Status:        seed.Status,
			SortOrder:     seed.SortOrder,
		}

		var existing usermodel.FeaturePackage
		result := db.Where("package_key = ?", item.PackageKey).First(&existing)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				if err := db.Create(&item).Error; err != nil {
					return err
				}
				existing = item
			} else {
				return result.Error
			}
		}

		actionIDs := make([]uuid.UUID, 0, len(seed.PermissionKeys))
		for _, permissionKey := range seed.PermissionKeys {
			var action usermodel.PermissionKey
			if err := db.Where("permission_key = ?", permissionKey).First(&action).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			actionIDs = append(actionIDs, action.ID)
		}
		seenActionIDs := make(map[uuid.UUID]struct{}, len(actionIDs))
		for _, actionID := range actionIDs {
			if _, ok := seenActionIDs[actionID]; ok {
				continue
			}
			seenActionIDs[actionID] = struct{}{}
			record := usermodel.FeaturePackageKey{PackageID: existing.ID, ActionID: actionID}
			if err := db.Where("package_id = ? AND action_id = ?", existing.ID, actionID).FirstOrCreate(&record).Error; err != nil {
				return err
			}
		}

		menuIDs := make([]uuid.UUID, 0, len(seed.MenuNames))
		for _, menuName := range seed.MenuNames {
			var menu systemmodels.MenuDefinition
			if err := db.Where("app_key = ? AND name = ?", item.AppKey, menuName).First(&menu).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return err
			}
			menuIDs = append(menuIDs, menu.ID)
		}
		seenMenuIDs := make(map[uuid.UUID]struct{}, len(menuIDs))
		for _, menuID := range menuIDs {
			if _, ok := seenMenuIDs[menuID]; ok {
				continue
			}
			seenMenuIDs[menuID] = struct{}{}
			record := usermodel.FeaturePackageMenu{PackageID: existing.ID, MenuID: menuID}
			if err := db.Where("package_id = ? AND menu_id = ?", existing.ID, menuID).FirstOrCreate(&record).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func EnsureDefaultFeaturePackageBundles(db *gorm.DB) error {
	seeds := DefaultFeaturePackageBundles()
	if len(seeds) == 0 {
		return nil
	}

	var packages []usermodel.FeaturePackage
	packageKeys := make([]string, 0, len(seeds)*2)
	for _, seed := range seeds {
		packageKeys = append(packageKeys, seed.ParentPackageKey, seed.ChildPackageKey)
	}
	packageKeys = uniqueStrings(packageKeys)
	if err := db.Where("package_key IN ?", packageKeys).Find(&packages).Error; err != nil {
		return err
	}
	packageByKey := make(map[string]usermodel.FeaturePackage, len(packages))
	for _, item := range packages {
		packageByKey[item.PackageKey] = item
	}

	for _, seed := range seeds {
		parentPkg, ok := packageByKey[seed.ParentPackageKey]
		if !ok {
			continue
		}
		childPkg, ok := packageByKey[seed.ChildPackageKey]
		if !ok || childPkg.ID == parentPkg.ID {
			continue
		}
		record := usermodel.FeaturePackageBundle{
			PackageID:      parentPkg.ID,
			ChildPackageID: childPkg.ID,
		}
		if err := db.Where("package_id = ? AND child_package_id = ?", parentPkg.ID, childPkg.ID).FirstOrCreate(&record).Error; err != nil {
			return err
		}
	}
	return nil
}

func EnsureDefaultRoleFeaturePackages(db *gorm.DB) error {
	defaultBindings := DefaultRolePackageBindings()
	roleCodes := DefaultRoleCodes()
	packageKeys := DefaultFeaturePackageKeys()
	for _, binding := range defaultBindings {
		roleCodes = append(roleCodes, binding.RoleCode)
		packageKeys = append(packageKeys, binding.PackageKey)
	}
	roleCodes = uniqueStrings(roleCodes)
	packageKeys = uniqueStrings(packageKeys)

	var roles []usermodel.Role
	if err := db.Where("collaboration_workspace_id IS NULL AND code IN ?", roleCodes).Find(&roles).Error; err != nil {
		return err
	}
	roleByCode := make(map[string]usermodel.Role, len(roles))
	for _, role := range roles {
		roleByCode[role.Code] = role
	}

	var packages []usermodel.FeaturePackage
	if err := db.Where("package_key IN ?", packageKeys).Find(&packages).Error; err != nil {
		return err
	}
	packageByKey := make(map[string]usermodel.FeaturePackage, len(packages))
	for _, item := range packages {
		packageByKey[item.PackageKey] = item
	}

	assignments := map[string][]string{}
	for _, binding := range defaultBindings {
		assignments[binding.RoleCode] = append(assignments[binding.RoleCode], binding.PackageKey)
	}
	for roleCode, keys := range assignments {
		role, ok := roleByCode[roleCode]
		if !ok {
			continue
		}
		for _, packageKey := range keys {
			pkg, ok := packageByKey[packageKey]
			if !ok {
				continue
			}
			var seedID uuid.UUID
			for _, binding := range defaultBindings {
				if binding.RoleCode == roleCode && binding.PackageKey == packageKey {
					seedID = binding.ID
					break
				}
			}
			var record usermodel.RoleFeaturePackage
			err := db.Where("role_id = ? AND package_id = ?", role.ID, pkg.ID).First(&record).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				record = usermodel.RoleFeaturePackage{
					ID:        seedID,
					RoleID:    role.ID,
					PackageID: pkg.ID,
					Enabled:   true,
				}
				if err := db.Create(&record).Error; err != nil {
					return err
				}
			} else {
				updates := map[string]interface{}{
					"enabled": true,
				}
				if seedID != uuid.Nil && record.ID == uuid.Nil {
					updates["id"] = seedID
				}
				if err := db.Model(&record).Updates(updates).Error; err != nil {
					return err
				}
			}
		}
	}

	adminRole, ok := roleByCode["admin"]
	if ok {
		adminPackageKeys := []string{"platform_admin.system_manage", "platform_admin.menu_manage", "platform_admin.api_manage"}
		adminPackageIDs := make([]uuid.UUID, 0, len(adminPackageKeys))
		for _, packageKey := range adminPackageKeys {
			if pkg, exists := packageByKey[packageKey]; exists {
				adminPackageIDs = append(adminPackageIDs, pkg.ID)
			}
		}
		if len(adminPackageIDs) > 0 {
			if err := db.Where("role_id = ? AND package_id IN ?", adminRole.ID, adminPackageIDs).Delete(&usermodel.RoleFeaturePackage{}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
