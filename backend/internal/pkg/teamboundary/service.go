package teamboundary

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
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
	GetSnapshot(teamID uuid.UUID) (*Snapshot, error)
	GetMenuSnapshot(teamID uuid.UUID) (*MenuSnapshot, error)
	GetEffectiveActionSet(teamID uuid.UUID) (map[uuid.UUID]bool, error)
	GetRoleSnapshot(teamID, roleID uuid.UUID, inheritAll bool) (*RoleSnapshot, error)
	RefreshSnapshot(teamID uuid.UUID) (*Snapshot, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{
		db: db,
	}
}

func (s *service) GetSnapshot(teamID uuid.UUID) (*Snapshot, error) {
	snapshot, err := s.loadActionSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return s.RefreshSnapshot(teamID)
}

func (s *service) calculateActionSnapshot(teamID uuid.UUID) (*Snapshot, error) {
	directPackageIDs, err := s.getPackageIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	packageIDs, expandedPackageIDs, err := s.resolvePackageSet(directPackageIDs, "team")
	if err != nil {
		return nil, err
	}
	derivedIDs, derivedMap, err := s.getDerivedActionIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	blockedIDs, err := s.getBlockedActionIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
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

func (s *service) GetEffectiveActionSet(teamID uuid.UUID) (map[uuid.UUID]bool, error) {
	snapshot, err := s.GetSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]bool, len(snapshot.EffectiveIDs))
	for _, actionID := range snapshot.EffectiveIDs {
		result[actionID] = true
	}
	return result, nil
}

func (s *service) GetMenuSnapshot(teamID uuid.UUID) (*MenuSnapshot, error) {
	snapshot, err := s.loadMenuSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	if _, err := s.RefreshSnapshot(teamID); err != nil {
		return nil, err
	}
	snapshot, err = s.loadMenuSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return emptyMenuSnapshot(), nil
}

func (s *service) calculateMenuSnapshot(teamID uuid.UUID) (*MenuSnapshot, error) {
	directPackageIDs, err := s.getPackageIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	packageIDs, expandedPackageIDs, err := s.resolvePackageSet(directPackageIDs, "team")
	if err != nil {
		return nil, err
	}
	derivedIDs, derivedMap, err := s.getMenuIDsByPackageIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	blockedIDs, err := s.getBlockedMenuIDsByTeamID(teamID)
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

func (s *service) GetRoleSnapshot(teamID, roleID uuid.UUID, inheritAll bool) (*RoleSnapshot, error) {
	snapshot, err := s.loadRoleSnapshot(teamID, roleID)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	snapshot, err = s.calculateRoleSnapshot(teamID, roleID, inheritAll)
	if err != nil {
		return nil, err
	}
	if err := s.saveRoleSnapshot(teamID, roleID, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *service) calculateRoleSnapshot(teamID, roleID uuid.UUID, inheritAll bool) (*RoleSnapshot, error) {
	directTeamPackageIDs, err := s.getPackageIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	teamPackageIDs, expandedTeamPackageIDs, err := s.resolvePackageSet(directTeamPackageIDs, "team")
	if err != nil {
		return nil, err
	}
	effectivePackageIDs := teamPackageIDs
	effectiveExpandedPackageIDs := expandedTeamPackageIDs
	inherited := inheritAll
	if !inheritAll {
		directRolePackageIDs, directRoleErr := s.getPackageIDsByRoleID(roleID)
		if directRoleErr != nil {
			return nil, directRoleErr
		}
		rolePackageIDs, expandedRolePackageIDs, roleErr := s.resolvePackageSet(directRolePackageIDs, "team")
		if roleErr != nil {
			return nil, roleErr
		}
		if len(rolePackageIDs) > 0 {
			effectivePackageIDs = intersectIDs(teamPackageIDs, rolePackageIDs)
			effectiveExpandedPackageIDs = intersectIDs(expandedTeamPackageIDs, expandedRolePackageIDs)
			inherited = false
		} else {
			inherited = true
		}
	}
	availableActionIDs, actionSourceMap, err := s.getDerivedActionIDs(effectiveExpandedPackageIDs)
	if err != nil {
		return nil, err
	}
	teamBlockedActionIDs, err := s.getBlockedActionIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	availableActionIDs = subtractIDs(availableActionIDs, teamBlockedActionIDs)
	actionSourceMap = filterSourceMap(actionSourceMap, availableActionIDs)

	roleDisabledActionIDs, err := s.getDisabledActionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	disabledActionIDs := intersectIDs(availableActionIDs, roleDisabledActionIDs)
	effectiveActionIDs := subtractIDs(availableActionIDs, disabledActionIDs)

	availableMenuIDs, menuSourceMap, err := s.getMenuIDsByPackageIDs(effectiveExpandedPackageIDs)
	if err != nil {
		return nil, err
	}
	teamBlockedMenuIDs, err := s.getBlockedMenuIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	availableMenuIDs = subtractIDs(availableMenuIDs, teamBlockedMenuIDs)
	menuSourceMap = filterSourceMap(menuSourceMap, availableMenuIDs)

	roleHiddenMenuIDs, err := s.getHiddenMenuIDsByRoleID(roleID)
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

func (s *service) RefreshSnapshot(teamID uuid.UUID) (*Snapshot, error) {
	actionSnapshot, err := s.calculateActionSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	menuSnapshot, err := s.calculateMenuSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	if err := s.saveSnapshot(teamID, actionSnapshot, menuSnapshot); err != nil {
		return nil, err
	}
	if err := s.refreshRoleSnapshots(teamID); err != nil {
		return nil, err
	}
	return actionSnapshot, nil
}

func (s *service) loadActionSnapshot(teamID uuid.UUID) (*Snapshot, error) {
	var record models.TeamAccessSnapshot
	if err := s.db.Where("team_id = ?", teamID).First(&record).Error; err != nil {
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

func (s *service) loadMenuSnapshot(teamID uuid.UUID) (*MenuSnapshot, error) {
	var record models.TeamAccessSnapshot
	if err := s.db.Where("team_id = ?", teamID).First(&record).Error; err != nil {
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

func (s *service) saveSnapshot(teamID uuid.UUID, actionSnapshot *Snapshot, menuSnapshot *MenuSnapshot) error {
	record := models.TeamAccessSnapshot{
		TeamID:             teamID,
		PackageIDs:         idsToUUIDStrings(actionSnapshot.PackageIDs),
		ExpandedPackageIDs: idsToUUIDStrings(actionSnapshot.ExpandedPackageIDs),
		DerivedActionIDs:   idsToUUIDStrings(actionSnapshot.DerivedIDs),
		DerivedActionMap:   sourceMapUUIDsToStrings(actionSnapshot.DerivedMap),
		BlockedActionIDs:   idsToUUIDStrings(actionSnapshot.BlockedIDs),
		EffectiveActionIDs: idsToUUIDStrings(actionSnapshot.EffectiveIDs),
		DerivedMenuIDs:     idsToUUIDStrings(menuSnapshot.DerivedIDs),
		DerivedMenuMap:     sourceMapUUIDsToStrings(menuSnapshot.DerivedMap),
		BlockedMenuIDs:     idsToUUIDStrings(menuSnapshot.BlockedIDs),
		EffectiveMenuIDs:   idsToUUIDStrings(menuSnapshot.EffectiveIDs),
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) loadRoleSnapshot(teamID, roleID uuid.UUID) (*RoleSnapshot, error) {
	var record models.TeamRoleAccessSnapshot
	if err := s.db.Where("team_id = ? AND role_id = ?", teamID, roleID).First(&record).Error; err != nil {
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

func (s *service) saveRoleSnapshot(teamID, roleID uuid.UUID, snapshot *RoleSnapshot) error {
	record := models.TeamRoleAccessSnapshot{
		TeamID:             teamID,
		RoleID:             roleID,
		PackageIDs:         idsToUUIDStrings(snapshot.PackageIDs),
		ExpandedPackageIDs: idsToUUIDStrings(snapshot.ExpandedPackageIDs),
		AvailableActionIDs: idsToUUIDStrings(snapshot.AvailableActionIDs),
		DisabledActionIDs:  idsToUUIDStrings(snapshot.DisabledActionIDs),
		ActionIDs:          idsToUUIDStrings(snapshot.ActionIDs),
		ActionSourceMap:    sourceMapUUIDsToStrings(snapshot.ActionSourceMap),
		AvailableMenuIDs:   idsToUUIDStrings(snapshot.AvailableMenuIDs),
		HiddenMenuIDs:      idsToUUIDStrings(snapshot.HiddenMenuIDs),
		MenuIDs:            idsToUUIDStrings(snapshot.MenuIDs),
		MenuSourceMap:      sourceMapUUIDsToStrings(snapshot.MenuSourceMap),
		Inherited:          snapshot.Inherited,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "role_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) refreshRoleSnapshots(teamID uuid.UUID) error {
	roleIDs, err := s.getRelevantRoleIDs(teamID)
	if err != nil {
		return err
	}
	inheritMap, err := s.getInheritedRoleMap(roleIDs)
	if err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		inheritAll := inheritMap[roleID]
		snapshot, snapshotErr := s.calculateRoleSnapshot(teamID, roleID, inheritAll)
		if snapshotErr != nil {
			return snapshotErr
		}
		if err := s.saveRoleSnapshot(teamID, roleID, snapshot); err != nil {
			return err
		}
	}
	if len(roleIDs) == 0 {
		return s.db.Where("team_id = ?", teamID).Delete(&models.TeamRoleAccessSnapshot{}).Error
	}
	return s.db.Where("team_id = ? AND role_id NOT IN ?", teamID, roleIDs).Delete(&models.TeamRoleAccessSnapshot{}).Error
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

func (s *service) getPackageIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.TeamFeaturePackage{}).
		Where("team_id = ? AND enabled = ?", teamID, true).
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) resolvePackageSet(ids []uuid.UUID, context string) ([]uuid.UUID, []uuid.UUID, error) {
	if len(ids) == 0 {
		return []uuid.UUID{}, []uuid.UUID{}, nil
	}
	packages, err := s.getPackagesByIDs(ids)
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
		if !ok || !contextAllowsPackage(context, pkg.ContextType) {
			continue
		}
		if _, ok := seenDirect[id]; !ok {
			seenDirect[id] = struct{}{}
			validDirect = append(validDirect, id)
		}
		if err := s.expandPackageID(id, context, packageMap, bundleChildrenMap, visited, seenExpanded, &expanded); err != nil {
			return nil, nil, err
		}
	}
	return validDirect, expanded, nil
}

func (s *service) expandPackageID(packageID uuid.UUID, context string, packageMap map[uuid.UUID]models.FeaturePackage, bundleChildrenMap map[uuid.UUID][]uuid.UUID, visited map[uuid.UUID]struct{}, seenExpanded map[uuid.UUID]struct{}, expanded *[]uuid.UUID) error {
	if _, ok := visited[packageID]; ok {
		return nil
	}
	visited[packageID] = struct{}{}
	pkg, ok := packageMap[packageID]
	if !ok || !contextAllowsPackage(context, pkg.ContextType) {
		return nil
	}
	if pkg.PackageType == "bundle" {
		for _, childID := range bundleChildrenMap[packageID] {
			if err := s.expandPackageID(childID, context, packageMap, bundleChildrenMap, visited, seenExpanded, expanded); err != nil {
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

func (s *service) getMenuIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	type packageMenuRow struct {
		PackageID uuid.UUID
		MenuID    uuid.UUID
	}
	var rows []packageMenuRow
	if err := s.db.Model(&models.FeaturePackageMenu{}).
		Select("package_id", "menu_id").
		Where("package_id IN ?", packageIDs).
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

func (s *service) getPackageIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Where("role_id = ? AND enabled = ?", roleID, true).
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) getRelevantRoleIDs(teamID uuid.UUID) ([]uuid.UUID, error) {
	roleIDs := make([]uuid.UUID, 0)
	var directRoleIDs []uuid.UUID
	if err := s.db.Model(&models.Role{}).
		Where("tenant_id = ? AND status = ?", teamID, "normal").
		Distinct("id").
		Pluck("id", &directRoleIDs).Error; err != nil {
		return nil, err
	}
	roleIDs = append(roleIDs, directRoleIDs...)
	var assignedRoleIDs []uuid.UUID
	if err := s.db.Model(&models.UserRole{}).
		Where("tenant_id = ?", teamID).
		Distinct("role_id").
		Pluck("role_id", &assignedRoleIDs).Error; err != nil {
		return nil, err
	}
	identityRoleIDs, err := s.getTenantIdentityRoleIDs(teamID)
	if err != nil {
		return nil, err
	}
	return mergeActionIDs(roleIDs, assignedRoleIDs, identityRoleIDs), nil
}

func (s *service) getTenantIdentityRoleIDs(teamID uuid.UUID) ([]uuid.UUID, error) {
	var roleCodes []string
	if err := s.db.Model(&models.TenantMember{}).
		Where("tenant_id = ? AND status = ?", teamID, "active").
		Distinct("role_code").
		Pluck("role_code", &roleCodes).Error; err != nil {
		return nil, err
	}
	if len(roleCodes) == 0 {
		return []uuid.UUID{}, nil
	}

	var roleIDs []uuid.UUID
	err := s.db.Model(&models.Role{}).
		Where("tenant_id IS NULL AND status = ? AND code IN ?", "normal", roleCodes).
		Distinct("id").
		Pluck("id", &roleIDs).Error
	return roleIDs, err
}

func (s *service) getInheritedRoleMap(roleIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	result := make(map[uuid.UUID]bool, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, nil
	}
	type roleTenantRow struct {
		ID       uuid.UUID
		TenantID *uuid.UUID
	}
	var rows []roleTenantRow
	if err := s.db.Model(&models.Role{}).
		Select("id", "tenant_id").
		Where("id IN ?", roleIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row.TenantID == nil || *row.TenantID == uuid.Nil
	}
	return result, nil
}

func (s *service) getPackagesByIDs(seedIDs []uuid.UUID) ([]models.FeaturePackage, error) {
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
	if err := s.db.Where("id IN ? AND status = ?", queue, "normal").Find(&items).Error; err != nil {
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

func (s *service) getBlockedActionIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.TeamBlockedAction{}).
		Where("team_id = ?", teamID).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) getBlockedMenuIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.TeamBlockedMenu{}).
		Where("team_id = ?", teamID).
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

func (s *service) getHiddenMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.RoleHiddenMenu{}).
		Where("role_id = ?", roleID).
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

func contextAllowsPackage(targetContext, packageContext string) bool {
	if packageContext == "" {
		return true
	}
	switch targetContext {
	case "team":
		return packageContext == "team" || packageContext == "common"
	case "platform":
		return packageContext == "platform" || packageContext == "common"
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
