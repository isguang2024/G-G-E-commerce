package collaborationworkspaceboundary

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/maben/backend/internal/modules/system/models"
	appctx "github.com/maben/backend/internal/pkg/appctx"
)

type Snapshot struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	DerivedIDs         []uuid.UUID
	DerivedMap         map[uuid.UUID][]uuid.UUID
	BlockedIDs         []uuid.UUID
	EffectiveIDs       []uuid.UUID
}

type MenuSnapshot struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	DerivedIDs         []uuid.UUID
	DerivedMap         map[uuid.UUID][]uuid.UUID
	BlockedIDs         []uuid.UUID
	EffectiveIDs       []uuid.UUID
}

type RoleSnapshot struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	AvailableActionIDs []uuid.UUID
	DisabledActionIDs  []uuid.UUID
	ActionIDs          []uuid.UUID
	ActionSourceMap    map[uuid.UUID][]uuid.UUID
	AvailableMenuIDs   []uuid.UUID
	HiddenMenuIDs      []uuid.UUID
	MenuIDs            []uuid.UUID
	MenuSourceMap      map[uuid.UUID][]uuid.UUID
	Inherited          bool
}

type Service interface {
	GetSnapshot(collaborationWorkspaceID uuid.UUID, appKey ...string) (*Snapshot, error)
	GetMenuSnapshot(collaborationWorkspaceID uuid.UUID, appKey ...string) (*MenuSnapshot, error)
	GetEffectiveActionSet(collaborationWorkspaceID uuid.UUID, appKey ...string) (map[uuid.UUID]bool, error)
	GetRoleSnapshot(collaborationWorkspaceID, roleID uuid.UUID, inheritAll bool, appKey ...string) (*RoleSnapshot, error)
	RefreshSnapshot(collaborationWorkspaceID uuid.UUID, appKey ...string) (*Snapshot, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{
		db: db,
	}
}

func (s *service) GetSnapshot(collaborationWorkspaceID uuid.UUID, appKey ...string) (*Snapshot, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	snapshot, err := s.loadActionSnapshot(collaborationWorkspaceID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return s.RefreshSnapshot(collaborationWorkspaceID, resolvedAppKey)
}

func (s *service) calculateActionSnapshot(collaborationWorkspaceID uuid.UUID, appKey string) (*Snapshot, error) {
	directPackageIDs, err := s.getPackageIDsByCollaborationWorkspaceID(collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	packageIDs, expandedPackageIDs, err := s.resolvePackageSet(directPackageIDs, "collaboration", appKey)
	if err != nil {
		return nil, err
	}
	derivedIDs, derivedMap, err := s.getDerivedActionIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	blockedIDs, err := s.getBlockedActionIDsByCollaborationWorkspaceID(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	blockedIDs = intersectIDs(derivedIDs, blockedIDs)
	effectiveIDs := subtractIDs(derivedIDs, blockedIDs)
	return &Snapshot{
		PackageIDs:         packageIDs,
		ExpandedPackageIDs: expandedPackageIDs,
		DerivedIDs:         derivedIDs,
		DerivedMap:         derivedMap,
		BlockedIDs:         blockedIDs,
		EffectiveIDs:       effectiveIDs,
	}, nil
}

func (s *service) GetEffectiveActionSet(collaborationWorkspaceID uuid.UUID, appKey ...string) (map[uuid.UUID]bool, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	snapshot, err := s.GetSnapshot(collaborationWorkspaceID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool, len(snapshot.EffectiveIDs))
	for _, actionID := range snapshot.EffectiveIDs {
		result[actionID] = true
	}
	return result, nil
}

func (s *service) GetMenuSnapshot(collaborationWorkspaceID uuid.UUID, appKey ...string) (*MenuSnapshot, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	snapshot, err := s.loadMenuSnapshot(collaborationWorkspaceID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	if _, err := s.RefreshSnapshot(collaborationWorkspaceID, resolvedAppKey); err != nil {
		return nil, err
	}
	snapshot, err = s.loadMenuSnapshot(collaborationWorkspaceID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return emptyMenuSnapshot(), nil
}

func (s *service) calculateMenuSnapshot(collaborationWorkspaceID uuid.UUID, appKey string) (*MenuSnapshot, error) {
	directPackageIDs, err := s.getPackageIDsByCollaborationWorkspaceID(collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	packageIDs, expandedPackageIDs, err := s.resolvePackageSet(directPackageIDs, "collaboration", appKey)
	if err != nil {
		return nil, err
	}
	derivedIDs, derivedMap, err := s.getMenuIDsByPackageIDs(expandedPackageIDs, appKey)
	if err != nil {
		return nil, err
	}
	blockedIDs, err := s.getBlockedMenuIDsByCollaborationWorkspaceID(collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	effectiveIDs := subtractIDs(derivedIDs, blockedIDs)
	return &MenuSnapshot{
		PackageIDs:         packageIDs,
		ExpandedPackageIDs: expandedPackageIDs,
		DerivedIDs:         derivedIDs,
		DerivedMap:         derivedMap,
		BlockedIDs:         blockedIDs,
		EffectiveIDs:       effectiveIDs,
	}, nil
}

func (s *service) GetRoleSnapshot(collaborationWorkspaceID, roleID uuid.UUID, inheritAll bool, appKey ...string) (*RoleSnapshot, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	snapshot, err := s.loadRoleSnapshot(collaborationWorkspaceID, roleID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	snapshot, err = s.calculateRoleSnapshot(collaborationWorkspaceID, roleID, inheritAll, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if err := s.saveRoleSnapshot(collaborationWorkspaceID, roleID, resolvedAppKey, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *service) calculateRoleSnapshot(collaborationWorkspaceID, roleID uuid.UUID, inheritAll bool, appKey string) (*RoleSnapshot, error) {
	directCollaborationWorkspacePackageIDs, err := s.getPackageIDsByCollaborationWorkspaceID(collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	collaborationWorkspaceFeaturePackageIDs, expandedCollaborationWorkspacePackageIDs, err := s.resolvePackageSet(directCollaborationWorkspacePackageIDs, "collaboration", appKey)
	if err != nil {
		return nil, err
	}
	effectivePackageIDs := collaborationWorkspaceFeaturePackageIDs
	effectiveExpandedPackageIDs := expandedCollaborationWorkspacePackageIDs
	inherited := inheritAll
	if !inheritAll {
		directRolePackageIDs, directRoleErr := s.getPackageIDsByRoleID(roleID, appKey)
		if directRoleErr != nil {
			return nil, directRoleErr
		}
		rolePackageIDs, expandedRolePackageIDs, roleErr := s.resolvePackageSet(directRolePackageIDs, "collaboration", appKey)
		if roleErr != nil {
			return nil, roleErr
		}
		if len(rolePackageIDs) > 0 {
			effectivePackageIDs = intersectIDs(collaborationWorkspaceFeaturePackageIDs, rolePackageIDs)
			effectiveExpandedPackageIDs = intersectIDs(expandedCollaborationWorkspacePackageIDs, expandedRolePackageIDs)
			inherited = false
		} else {
			inherited = true
		}
	}
	availableActionIDs, actionSourceMap, err := s.getDerivedActionIDs(effectiveExpandedPackageIDs)
	if err != nil {
		return nil, err
	}
	collaborationWorkspaceBlockedActionIDs, err := s.getBlockedActionIDsByCollaborationWorkspaceID(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	availableActionIDs = subtractIDs(availableActionIDs, collaborationWorkspaceBlockedActionIDs)
	actionSourceMap = filterSourceMap(actionSourceMap, availableActionIDs)

	roleDisabledActionIDs, err := s.getDisabledActionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	disabledActionIDs := intersectIDs(availableActionIDs, roleDisabledActionIDs)
	effectiveActionIDs := subtractIDs(availableActionIDs, disabledActionIDs)

	availableMenuIDs, menuSourceMap, err := s.getMenuIDsByPackageIDs(effectiveExpandedPackageIDs, appKey)
	if err != nil {
		return nil, err
	}
	collaborationWorkspaceBlockedMenuIDs, err := s.getBlockedMenuIDsByCollaborationWorkspaceID(collaborationWorkspaceID, appKey)
	if err != nil {
		return nil, err
	}
	availableMenuIDs = subtractIDs(availableMenuIDs, collaborationWorkspaceBlockedMenuIDs)
	menuSourceMap = filterSourceMap(menuSourceMap, availableMenuIDs)

	roleHiddenMenuIDs, err := s.getHiddenMenuIDsByRoleID(roleID, appKey)
	if err != nil {
		return nil, err
	}
	hiddenMenuIDs := intersectIDs(availableMenuIDs, roleHiddenMenuIDs)
	effectiveMenuIDs := subtractIDs(availableMenuIDs, hiddenMenuIDs)

	return &RoleSnapshot{
		PackageIDs:         effectivePackageIDs,
		ExpandedPackageIDs: effectiveExpandedPackageIDs,
		AvailableActionIDs: availableActionIDs,
		DisabledActionIDs:  disabledActionIDs,
		ActionIDs:          effectiveActionIDs,
		ActionSourceMap:    actionSourceMap,
		AvailableMenuIDs:   availableMenuIDs,
		HiddenMenuIDs:      hiddenMenuIDs,
		MenuIDs:            effectiveMenuIDs,
		MenuSourceMap:      menuSourceMap,
		Inherited:          inherited,
	}, nil
}

func (s *service) RefreshSnapshot(collaborationWorkspaceID uuid.UUID, appKey ...string) (*Snapshot, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	actionSnapshot, err := s.calculateActionSnapshot(collaborationWorkspaceID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	menuSnapshot, err := s.calculateMenuSnapshot(collaborationWorkspaceID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if err := s.saveSnapshot(collaborationWorkspaceID, resolvedAppKey, actionSnapshot, menuSnapshot); err != nil {
		return nil, err
	}
	if err := s.refreshRoleSnapshots(collaborationWorkspaceID, resolvedAppKey); err != nil {
		return nil, err
	}
	return actionSnapshot, nil
}

func (s *service) loadActionSnapshot(collaborationWorkspaceID uuid.UUID, appKey string) (*Snapshot, error) {
	var record models.CollaborationWorkspaceAccessSnapshot
	if err := s.db.Where("app_key = ? AND collaboration_workspace_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &Snapshot{
		PackageIDs:         uuidStringsToIDs(record.PackageIDs),
		ExpandedPackageIDs: uuidStringsToIDs(record.ExpandedPackageIDs),
		DerivedIDs:         uuidStringsToIDs(record.DerivedActionIDs),
		DerivedMap:         sourceMapStringsToUUIDs(record.DerivedActionMap),
		BlockedIDs:         uuidStringsToIDs(record.BlockedActionIDs),
		EffectiveIDs:       uuidStringsToIDs(record.EffectiveActionIDs),
	}, nil
}

func (s *service) loadMenuSnapshot(collaborationWorkspaceID uuid.UUID, appKey string) (*MenuSnapshot, error) {
	var record models.CollaborationWorkspaceAccessSnapshot
	if err := s.db.Where("app_key = ? AND collaboration_workspace_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &MenuSnapshot{
		PackageIDs:         uuidStringsToIDs(record.PackageIDs),
		ExpandedPackageIDs: uuidStringsToIDs(record.ExpandedPackageIDs),
		DerivedIDs:         uuidStringsToIDs(record.DerivedMenuIDs),
		DerivedMap:         sourceMapStringsToUUIDs(record.DerivedMenuMap),
		BlockedIDs:         uuidStringsToIDs(record.BlockedMenuIDs),
		EffectiveIDs:       uuidStringsToIDs(record.EffectiveMenuIDs),
	}, nil
}

func (s *service) saveSnapshot(collaborationWorkspaceID uuid.UUID, appKey string, actionSnapshot *Snapshot, menuSnapshot *MenuSnapshot) error {
	record := models.CollaborationWorkspaceAccessSnapshot{
		AppKey:                   appctx.NormalizeAppKey(appKey),
		CollaborationWorkspaceID: collaborationWorkspaceID,
		PackageIDs:               idsToUUIDStrings(actionSnapshot.PackageIDs),
		ExpandedPackageIDs:       idsToUUIDStrings(actionSnapshot.ExpandedPackageIDs),
		DerivedActionIDs:         idsToUUIDStrings(actionSnapshot.DerivedIDs),
		DerivedActionMap:         sourceMapUUIDsToStrings(actionSnapshot.DerivedMap),
		BlockedActionIDs:         idsToUUIDStrings(actionSnapshot.BlockedIDs),
		EffectiveActionIDs:       idsToUUIDStrings(actionSnapshot.EffectiveIDs),
		DerivedMenuIDs:           idsToUUIDStrings(menuSnapshot.DerivedIDs),
		DerivedMenuMap:           sourceMapUUIDsToStrings(menuSnapshot.DerivedMap),
		BlockedMenuIDs:           idsToUUIDStrings(menuSnapshot.BlockedIDs),
		EffectiveMenuIDs:         idsToUUIDStrings(menuSnapshot.EffectiveIDs),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app_key"}, {Name: "collaboration_workspace_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) loadRoleSnapshot(collaborationWorkspaceID, roleID uuid.UUID, appKey string) (*RoleSnapshot, error) {
	var record models.CollaborationWorkspaceRoleAccessSnapshot
	if err := s.db.Where("app_key = ? AND collaboration_workspace_id = ? AND role_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID, roleID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &RoleSnapshot{
		PackageIDs:         uuidStringsToIDs(record.PackageIDs),
		ExpandedPackageIDs: uuidStringsToIDs(record.ExpandedPackageIDs),
		AvailableActionIDs: uuidStringsToIDs(record.AvailableActionIDs),
		DisabledActionIDs:  uuidStringsToIDs(record.DisabledActionIDs),
		ActionIDs:          uuidStringsToIDs(record.ActionIDs),
		ActionSourceMap:    sourceMapStringsToUUIDs(record.ActionSourceMap),
		AvailableMenuIDs:   uuidStringsToIDs(record.AvailableMenuIDs),
		HiddenMenuIDs:      uuidStringsToIDs(record.HiddenMenuIDs),
		MenuIDs:            uuidStringsToIDs(record.MenuIDs),
		MenuSourceMap:      sourceMapStringsToUUIDs(record.MenuSourceMap),
		Inherited:          record.Inherited,
	}, nil
}

func (s *service) saveRoleSnapshot(collaborationWorkspaceID, roleID uuid.UUID, appKey string, snapshot *RoleSnapshot) error {
	record := models.CollaborationWorkspaceRoleAccessSnapshot{
		AppKey:                   appctx.NormalizeAppKey(appKey),
		CollaborationWorkspaceID: collaborationWorkspaceID,
		RoleID:                   roleID,
		PackageIDs:               idsToUUIDStrings(snapshot.PackageIDs),
		ExpandedPackageIDs:       idsToUUIDStrings(snapshot.ExpandedPackageIDs),
		AvailableActionIDs:       idsToUUIDStrings(snapshot.AvailableActionIDs),
		DisabledActionIDs:        idsToUUIDStrings(snapshot.DisabledActionIDs),
		ActionIDs:                idsToUUIDStrings(snapshot.ActionIDs),
		ActionSourceMap:          sourceMapUUIDsToStrings(snapshot.ActionSourceMap),
		AvailableMenuIDs:         idsToUUIDStrings(snapshot.AvailableMenuIDs),
		HiddenMenuIDs:            idsToUUIDStrings(snapshot.HiddenMenuIDs),
		MenuIDs:                  idsToUUIDStrings(snapshot.MenuIDs),
		MenuSourceMap:            sourceMapUUIDsToStrings(snapshot.MenuSourceMap),
		Inherited:                snapshot.Inherited,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app_key"}, {Name: "collaboration_workspace_id"}, {Name: "role_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) refreshRoleSnapshots(collaborationWorkspaceID uuid.UUID, appKey string) error {
	roleIDs, err := s.getRelevantRoleIDs(collaborationWorkspaceID)
	if err != nil {
		return err
	}
	inheritMap, err := s.getInheritedRoleMap(roleIDs)
	if err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		inheritAll := inheritMap[roleID]
		snapshot, snapshotErr := s.calculateRoleSnapshot(collaborationWorkspaceID, roleID, inheritAll, appKey)
		if snapshotErr != nil {
			return snapshotErr
		}
		if err := s.saveRoleSnapshot(collaborationWorkspaceID, roleID, appKey, snapshot); err != nil {
			return err
		}
	}
	if len(roleIDs) == 0 {
		return s.db.Where("app_key = ? AND collaboration_workspace_id = ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID).Delete(&models.CollaborationWorkspaceRoleAccessSnapshot{}).Error
	}
	return s.db.Where("app_key = ? AND collaboration_workspace_id = ? AND role_id NOT IN ?", appctx.NormalizeAppKey(appKey), collaborationWorkspaceID, roleIDs).Delete(&models.CollaborationWorkspaceRoleAccessSnapshot{}).Error
}

func (s *service) getDerivedActionIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	type packageActionRow struct {
		PackageID uuid.UUID
		ActionID  uuid.UUID
	}
	var rows []packageActionRow
	if err := s.db.Model(&models.FeaturePackageKey{}).
		Select("package_id", "action_id").
		Where("package_id IN ?", packageIDs).
		Order("package_id ASC").
		Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	result := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	derivedMap := make(map[uuid.UUID][]uuid.UUID)
	for _, row := range rows {
		derivedMap[row.ActionID] = appendDerivedPackageID(derivedMap[row.ActionID], row.PackageID)
		if _, ok := seen[row.ActionID]; ok {
			continue
		}
		seen[row.ActionID] = struct{}{}
		result = append(result, row.ActionID)
	}
	return result, derivedMap, nil
}

func (s *service) getPackageIDsByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var workspace models.Workspace
	if err := s.db.
		Where("workspace_type = ? AND collaboration_workspace_id = ? AND deleted_at IS NULL", models.WorkspaceTypeCollaboration, collaborationWorkspaceID).
		First(&workspace).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err == nil {
		var workspacePackageIDs []uuid.UUID
		if queryErr := s.db.Model(&models.WorkspaceFeaturePackage{}).
			Joins("JOIN feature_packages ON feature_packages.id = workspace_feature_packages.package_id").
			Where("workspace_feature_packages.workspace_id = ? AND workspace_feature_packages.enabled = ? AND workspace_feature_packages.deleted_at IS NULL", workspace.ID, true).
			Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
			Distinct("workspace_feature_packages.package_id").
			Pluck("package_id", &workspacePackageIDs).Error; queryErr != nil {
			return nil, queryErr
		}
		if len(workspacePackageIDs) > 0 {
			return workspacePackageIDs, nil
		}
	}

	var packageIDs []uuid.UUID
	err := s.db.Model(&models.CollaborationWorkspaceFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = collaboration_workspace_feature_packages.package_id").
		Where("collaboration_workspace_id = ? AND enabled = ?", collaborationWorkspaceID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Distinct("collaboration_workspace_feature_packages.package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) resolvePackageSet(ids []uuid.UUID, workspaceContext string, appKey string) ([]uuid.UUID, []uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, []uuid.UUID{}, nil
	}
	packages, err := s.getPackagesByIDs(ids, appKey)
	if err != nil {
		return nil, nil, err
	}
	packageMap := make(map[uuid.UUID]models.FeaturePackage, len(packages))
	for _, item := range packages {
		packageMap[item.ID] = item
	}
	bundleChildrenMap, err := s.getBundleChildrenMap()
	if err != nil {
		return nil, nil, err
	}
	validDirect := make([]uuid.UUID, 0, len(ids))
	seenDirect := make(map[uuid.UUID]struct{}, len(ids))
	expanded := make([]uuid.UUID, 0, len(ids))
	seenExpanded := make(map[uuid.UUID]struct{}, len(ids))
	visited := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		pkg, ok := packageMap[id]
		if !ok || !workspaceContextAllowsPackage(workspaceContext, pkg.ContextType) {
			continue
		}
		if _, ok := seenDirect[id]; !ok {
			seenDirect[id] = struct{}{}
			validDirect = append(validDirect, id)
		}
		if err := s.expandPackageID(id, workspaceContext, packageMap, bundleChildrenMap, visited, seenExpanded, &expanded); err != nil {
			return nil, nil, err
		}
	}
	return validDirect, expanded, nil
}

func (s *service) expandPackageID(packageID uuid.UUID, workspaceContext string, packageMap map[uuid.UUID]models.FeaturePackage, bundleChildrenMap map[uuid.UUID][]uuid.UUID, visited map[uuid.UUID]struct{}, seenExpanded map[uuid.UUID]struct{}, expanded *[]uuid.UUID) error {
	if _, ok := visited[packageID]; ok {
		return nil
	}
	visited[packageID] = struct{}{}
	pkg, ok := packageMap[packageID]
	if !ok || !workspaceContextAllowsPackage(workspaceContext, pkg.ContextType) {
		return nil
	}
	if pkg.PackageType == "bundle" {
		for _, childID := range bundleChildrenMap[packageID] {
			if err := s.expandPackageID(childID, workspaceContext, packageMap, bundleChildrenMap, visited, seenExpanded, expanded); err != nil {
				return err
			}
		}
		return nil
	}
	if _, ok := seenExpanded[packageID]; ok {
		return nil
	}
	seenExpanded[packageID] = struct{}{}
	*expanded = append(*expanded, packageID)
	return nil
}

func (s *service) getMenuIDsByPackageIDs(packageIDs []uuid.UUID, appKey string) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	type packageMenuRow struct {
		PackageID uuid.UUID
		MenuID    uuid.UUID
	}
	var rows []packageMenuRow
	if err := s.db.Model(&models.FeaturePackageMenu{}).
		Select("feature_package_menus.package_id", "feature_package_menus.menu_id").
		Joins("JOIN menu_definitions ON menu_definitions.id = feature_package_menus.menu_id").
		Where("package_id IN ?", packageIDs).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Order("package_id ASC").
		Find(&rows).Error; err != nil {
		return nil, nil, err
	}
	menuIDs := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	sourceMap := make(map[uuid.UUID][]uuid.UUID)
	for _, row := range rows {
		sourceMap[row.MenuID] = appendDerivedPackageID(sourceMap[row.MenuID], row.PackageID)
		if _, ok := seen[row.MenuID]; ok {
			continue
		}
		seen[row.MenuID] = struct{}{}
		menuIDs = append(menuIDs, row.MenuID)
	}
	return menuIDs, sourceMap, nil
}

func (s *service) getPackageIDsByRoleID(roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = role_feature_packages.package_id").
		Where("role_id = ? AND enabled = ?", roleID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Distinct("role_feature_packages.package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) getRelevantRoleIDs(collaborationWorkspaceID uuid.UUID) ([]uuid.UUID, error) {
	roleIDs := make([]uuid.UUID, 0)
	var directRoleIDs []uuid.UUID
	if err := s.db.Model(&models.Role{}).
		Where("collaboration_workspace_id = ? AND status = ?", collaborationWorkspaceID, "normal").
		Distinct("id").
		Pluck("id", &directRoleIDs).Error; err != nil {
		return nil, err
	}
	roleIDs = append(roleIDs, directRoleIDs...)
	var assignedRoleIDs []uuid.UUID
	if err := s.db.Model(&models.UserRole{}).
		Where("collaboration_workspace_id = ?", collaborationWorkspaceID).
		Distinct("role_id").
		Pluck("role_id", &assignedRoleIDs).Error; err != nil {
		return nil, err
	}
	identityRoleIDs, err := s.getCollaborationWorkspaceIdentityRoleIDs(collaborationWorkspaceID)
	if err != nil {
		return nil, err
	}
	return mergeActionIDs(roleIDs, assignedRoleIDs, identityRoleIDs), nil
}

func (s *service) getCollaborationWorkspaceIdentityRoleIDs(collaborationWorkspaceID uuid.UUID) ([]uuid.UUID, error) {
	var roleCodes []string
	if err := s.db.Model(&models.CollaborationWorkspaceMember{}).
		Where("collaboration_workspace_id = ? AND status = ?", collaborationWorkspaceID, "active").
		Distinct("role_code").
		Pluck("role_code", &roleCodes).Error; err != nil {
		return nil, err
	}
	if len(roleCodes) == 0 {
		return []uuid.UUID{}, nil
	}

	var roleIDs []uuid.UUID
	err := s.db.Model(&models.Role{}).
		Where("collaboration_workspace_id IS NULL AND status = ? AND code IN ?", "normal", roleCodes).
		Distinct("id").
		Pluck("id", &roleIDs).Error
	return roleIDs, err
}

func (s *service) getInheritedRoleMap(roleIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, nil
	}
	type roleCollaborationWorkspaceRow struct {
		ID                       uuid.UUID
		CollaborationWorkspaceID *uuid.UUID
	}
	var rows []roleCollaborationWorkspaceRow
	if err := s.db.Model(&models.Role{}).
		Select("id", "collaboration_workspace_id").
		Where("id IN ?", roleIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row.CollaborationWorkspaceID == nil || *row.CollaborationWorkspaceID == uuid.Nil
	}
	return result, nil
}

func (s *service) getPackagesByIDs(seedIDs []uuid.UUID, appKey string) ([]models.FeaturePackage, error) {
	if len(seedIDs) == 0 {
		return []models.FeaturePackage{}, nil
	}
	var bundleRows []models.FeaturePackageBundle
	if err := s.db.Model(&models.FeaturePackageBundle{}).
		Select("package_id", "child_package_id").
		Find(&bundleRows).Error; err != nil {
		return nil, err
	}
	queue := append([]uuid.UUID{}, seedIDs...)
	seen := make(map[uuid.UUID]struct{}, len(seedIDs))
	for _, id := range seedIDs {
		seen[id] = struct{}{}
	}
	for index := 0; index < len(queue); index++ {
		current := queue[index]
		for _, row := range bundleRows {
			if row.PackageID != current {
				continue
			}
			if _, ok := seen[row.ChildPackageID]; ok {
				continue
			}
			seen[row.ChildPackageID] = struct{}{}
			queue = append(queue, row.ChildPackageID)
		}
	}
	var items []models.FeaturePackage
	if err := s.db.Where("app_key = ? AND id IN ? AND status = ?", appctx.NormalizeAppKey(appKey), queue, "normal").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *service) getBundleChildrenMap() (map[uuid.UUID][]uuid.UUID, error) {
	var rows []models.FeaturePackageBundle
	if err := s.db.Model(&models.FeaturePackageBundle{}).
		Select("package_id", "child_package_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]uuid.UUID)
	for _, row := range rows {
		result[row.PackageID] = append(result[row.PackageID], row.ChildPackageID)
	}
	return result, nil
}

func emptyActionSnapshot() *Snapshot {
	return &Snapshot{
		PackageIDs:         []uuid.UUID{},
		ExpandedPackageIDs: []uuid.UUID{},
		DerivedIDs:         []uuid.UUID{},
		DerivedMap:         map[uuid.UUID][]uuid.UUID{},
		BlockedIDs:         []uuid.UUID{},
		EffectiveIDs:       []uuid.UUID{},
	}
}

func emptyMenuSnapshot() *MenuSnapshot {
	return &MenuSnapshot{
		PackageIDs:         []uuid.UUID{},
		ExpandedPackageIDs: []uuid.UUID{},
		DerivedIDs:         []uuid.UUID{},
		DerivedMap:         map[uuid.UUID][]uuid.UUID{},
		BlockedIDs:         []uuid.UUID{},
		EffectiveIDs:       []uuid.UUID{},
	}
}

func emptyRoleSnapshot(inherited bool) *RoleSnapshot {
	return &RoleSnapshot{
		PackageIDs:         []uuid.UUID{},
		ExpandedPackageIDs: []uuid.UUID{},
		AvailableActionIDs: []uuid.UUID{},
		DisabledActionIDs:  []uuid.UUID{},
		ActionIDs:          []uuid.UUID{},
		ActionSourceMap:    map[uuid.UUID][]uuid.UUID{},
		AvailableMenuIDs:   []uuid.UUID{},
		HiddenMenuIDs:      []uuid.UUID{},
		MenuIDs:            []uuid.UUID{},
		MenuSourceMap:      map[uuid.UUID][]uuid.UUID{},
		Inherited:          inherited,
	}
}

func (s *service) getBlockedActionIDsByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.CollaborationWorkspaceBlockedAction{}).
		Where("collaboration_workspace_id = ?", collaborationWorkspaceID).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) getBlockedMenuIDsByCollaborationWorkspaceID(collaborationWorkspaceID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.CollaborationWorkspaceBlockedMenu{}).
		Joins("JOIN menu_definitions ON menu_definitions.id = collaboration_workspace_blocked_menus.menu_id").
		Where("collaboration_workspace_id = ?", collaborationWorkspaceID).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Distinct("collaboration_workspace_blocked_menus.menu_id").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (s *service) getDisabledActionIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.RoleDisabledAction{}).
		Where("role_id = ?", roleID).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) getHiddenMenuIDsByRoleID(roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.RoleHiddenMenu{}).
		Joins("JOIN menu_definitions ON menu_definitions.id = role_hidden_menus.menu_id").
		Where("role_id = ?", roleID).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Distinct("role_hidden_menus.menu_id").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func mergeActionIDs(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, actionID := range group {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			result = append(result, actionID)
		}
	}
	return result
}

func subtractIDs(source []uuid.UUID, blocked []uuid.UUID) []uuid.UUID {
	if len(source) == 0 {
		return []uuid.UUID{}
	}
	if len(blocked) == 0 {
		return append([]uuid.UUID{}, source...)
	}
	blockedSet := make(map[uuid.UUID]struct{}, len(blocked))
	for _, id := range blocked {
		blockedSet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(source))
	seen := make(map[uuid.UUID]struct{}, len(source))
	for _, id := range source {
		if _, ok := blockedSet[id]; ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func filterSourceMap(sourceMap map[uuid.UUID][]uuid.UUID, allowedIDs []uuid.UUID) map[uuid.UUID][]uuid.UUID {
	if len(sourceMap) == 0 {
		return map[uuid.UUID][]uuid.UUID{}
	}
	allowedSet := make(map[uuid.UUID]struct{}, len(allowedIDs))
	for _, id := range allowedIDs {
		allowedSet[id] = struct{}{}
	}
	result := make(map[uuid.UUID][]uuid.UUID, len(allowedSet))
	for id, packageIDs := range sourceMap {
		if _, ok := allowedSet[id]; !ok {
			continue
		}
		result[id] = append([]uuid.UUID{}, packageIDs...)
	}
	return result
}

func appendDerivedPackageID(current []uuid.UUID, packageID uuid.UUID) []uuid.UUID {
	for _, item := range current {
		if item == packageID {
			return current
		}
	}
	return append(current, packageID)
}

func intersectIDs(primary, boundary []uuid.UUID) []uuid.UUID {
	if len(primary) == 0 || len(boundary) == 0 {
		return []uuid.UUID{}
	}
	boundarySet := make(map[uuid.UUID]struct{}, len(boundary))
	for _, id := range boundary {
		boundarySet[id] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(primary))
	seen := make(map[uuid.UUID]struct{}, len(primary))
	for _, id := range primary {
		if _, ok := boundarySet[id]; !ok {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func workspaceContextAllowsPackage(workspaceContext, packageContext string) bool {
	if packageContext == "" {
		return true
	}
	switch workspaceContext {
	case "collaboration":
		return packageContext == "collaboration" || packageContext == "common"
	case "personal":
		return packageContext == "personal" || packageContext == "common"
	default:
		return false
	}
}

func idsToUUIDStrings(items []uuid.UUID) []string {
	if len(items) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item.String())
	}
	return result
}

func uuidStringsToIDs(items []string) []uuid.UUID {
	if len(items) == 0 {
		return []uuid.UUID{}
	}
	result := make([]uuid.UUID, 0, len(items))
	seen := make(map[uuid.UUID]struct{}, len(items))
	for _, item := range items {
		id, err := uuid.Parse(item)
		if err != nil || id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func sourceMapUUIDsToStrings(sourceMap map[uuid.UUID][]uuid.UUID) map[string][]string {
	if len(sourceMap) == 0 {
		return map[string][]string{}
	}
	result := make(map[string][]string, len(sourceMap))
	for id, packageIDs := range sourceMap {
		if id == uuid.Nil {
			continue
		}
		result[id.String()] = idsToUUIDStrings(packageIDs)
	}
	return result
}

func sourceMapStringsToUUIDs(sourceMap map[string][]string) map[uuid.UUID][]uuid.UUID {
	if len(sourceMap) == 0 {
		return map[uuid.UUID][]uuid.UUID{}
	}
	result := make(map[uuid.UUID][]uuid.UUID, len(sourceMap))
	for idText, packageIDs := range sourceMap {
		id, err := uuid.Parse(idText)
		if err != nil || id == uuid.Nil {
			continue
		}
		result[id] = uuidStringsToIDs(packageIDs)
	}
	return result
}

func resolveAppKey(appKey ...string) string {
	if len(appKey) == 0 {
		return appctx.NormalizeAppKey("")
	}
	return appctx.NormalizeAppKey(appKey[0])
}

