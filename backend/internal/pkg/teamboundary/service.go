package teamboundary

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type Snapshot struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	DerivedIDs         []uuid.UUID
	DerivedMap         map[uuid.UUID][]uuid.UUID
	BlockedIDs         []uuid.UUID
	ManualIDs          []uuid.UUID
	EffectiveIDs       []uuid.UUID
	FromCache          bool
}

type RoleSnapshot struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	ActionIDs          []uuid.UUID
	ActionSourceMap    map[uuid.UUID][]uuid.UUID
	MenuIDs            []uuid.UUID
	MenuSourceMap      map[uuid.UUID][]uuid.UUID
	HasMenuBoundary    bool
	Inherited          bool
}

type Service interface {
	GetSnapshot(teamID uuid.UUID) (*Snapshot, error)
	GetEffectiveActionSet(teamID uuid.UUID) (map[uuid.UUID]bool, error)
	GetRoleSnapshot(teamID, roleID uuid.UUID, inheritAll bool) (*RoleSnapshot, error)
	RefreshCache(teamID uuid.UUID) (*Snapshot, error)
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
	manualIDs, err := s.getManualActionIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	effectiveIDs := subtractIDs(mergeActionIDs(derivedIDs, manualIDs), blockedIDs)
	fromCache := false
	if len(packageIDs) == 0 && len(blockedIDs) == 0 && len(manualIDs) == 0 && len(effectiveIDs) == 0 {
		cacheIDs, cacheErr := s.getCachedActionIDsByTeamID(teamID)
		if cacheErr != nil {
			return nil, cacheErr
		}
		if len(cacheIDs) > 0 {
			effectiveIDs = cacheIDs
			fromCache = true
		}
	}
	return &Snapshot{
		PackageIDs:         packageIDs,
		ExpandedPackageIDs: expandedPackageIDs,
		DerivedIDs:         derivedIDs,
		DerivedMap:         derivedMap,
		BlockedIDs:         blockedIDs,
		ManualIDs:          manualIDs,
		EffectiveIDs:       effectiveIDs,
		FromCache:          fromCache,
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

func (s *service) GetRoleSnapshot(teamID, roleID uuid.UUID, inheritAll bool) (*RoleSnapshot, error) {
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
	actionIDs, actionSourceMap, err := s.getDerivedActionIDs(effectiveExpandedPackageIDs)
	if err != nil {
		return nil, err
	}
	teamBlockedActionIDs, err := s.getBlockedActionIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	actionIDs = subtractIDs(actionIDs, teamBlockedActionIDs)
	actionSourceMap = filterSourceMap(actionSourceMap, actionIDs)

	roleDisabledActionIDs, err := s.getDisabledActionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	actionIDs = subtractIDs(actionIDs, roleDisabledActionIDs)
	actionSourceMap = filterSourceMap(actionSourceMap, actionIDs)

	menuIDs, menuSourceMap, err := s.getMenuIDsByPackageIDs(effectiveExpandedPackageIDs)
	if err != nil {
		return nil, err
	}
	teamBlockedMenuIDs, err := s.getBlockedMenuIDsByTeamID(teamID)
	if err != nil {
		return nil, err
	}
	menuIDs = subtractIDs(menuIDs, teamBlockedMenuIDs)
	menuSourceMap = filterSourceMap(menuSourceMap, menuIDs)

	roleHiddenMenuIDs, err := s.getHiddenMenuIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	menuIDs = subtractIDs(menuIDs, roleHiddenMenuIDs)
	menuSourceMap = filterSourceMap(menuSourceMap, menuIDs)

	return &RoleSnapshot{
		PackageIDs:         effectivePackageIDs,
		ExpandedPackageIDs: effectiveExpandedPackageIDs,
		ActionIDs:          actionIDs,
		ActionSourceMap:    actionSourceMap,
		MenuIDs:            menuIDs,
		MenuSourceMap:      menuSourceMap,
		HasMenuBoundary:    len(effectiveExpandedPackageIDs) > 0,
		Inherited:          inherited,
	}, nil
}

func (s *service) RefreshCache(teamID uuid.UUID) (*Snapshot, error) {
	snapshot, err := s.GetSnapshot(teamID)
	if err != nil {
		return nil, err
	}
	if err := s.replaceCachedActionIDs(teamID, snapshot.EffectiveIDs); err != nil {
		return nil, err
	}
	snapshot.FromCache = false
	return snapshot, nil
}

func (s *service) getDerivedActionIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	derivedMap := make(map[uuid.UUID][]uuid.UUID)
	for _, packageID := range packageIDs {
		actionIDs, err := s.getActionIDsByPackageID(packageID)
		if err != nil {
			return nil, nil, err
		}
		for _, actionID := range actionIDs {
			derivedMap[actionID] = appendDerivedPackageID(derivedMap[actionID], packageID)
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			result = append(result, actionID)
		}
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
	validDirect := make([]uuid.UUID, 0, len(ids))
	seenDirect := make(map[uuid.UUID]struct{}, len(ids))
	expanded := make([]uuid.UUID, 0, len(ids))
	seenExpanded := make(map[uuid.UUID]struct{}, len(ids))
	visited := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		pkg, err := s.getPackageByID(id)
		if err != nil {
			return nil, nil, err
		}
		if pkg == nil || !contextAllowsPackage(context, pkg.ContextType) {
			continue
		}
		if _, ok := seenDirect[id]; !ok {
			seenDirect[id] = struct{}{}
			validDirect = append(validDirect, id)
		}
		if err := s.expandPackageID(id, context, visited, seenExpanded, &expanded); err != nil {
			return nil, nil, err
		}
	}
	return validDirect, expanded, nil
}

func (s *service) expandPackageID(packageID uuid.UUID, context string, visited map[uuid.UUID]struct{}, seenExpanded map[uuid.UUID]struct{}, expanded *[]uuid.UUID) error {
	if _, ok := visited[packageID]; ok {
		return nil
	}
	visited[packageID] = struct{}{}
	pkg, err := s.getPackageByID(packageID)
	if err != nil {
		return err
	}
	if pkg == nil || !contextAllowsPackage(context, pkg.ContextType) {
		return nil
	}
	if pkg.PackageType == "bundle" {
		childIDs, err := s.getChildPackageIDs(packageID)
		if err != nil {
			return err
		}
		for _, childID := range childIDs {
			if err := s.expandPackageID(childID, context, visited, seenExpanded, expanded); err != nil {
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

func (s *service) getPackageByID(packageID uuid.UUID) (*models.FeaturePackage, error) {
	var item models.FeaturePackage
	err := s.db.Where("id = ? AND status = ?", packageID, "normal").First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (s *service) getChildPackageIDs(packageID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := s.db.Model(&models.FeaturePackageBundle{}).
		Where("package_id = ?", packageID).
		Pluck("child_package_id", &ids).Error
	return ids, err
}

func (s *service) getActionIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.FeaturePackageAction{}).
		Where("package_id = ?", packageID).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) getMenuIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	if len(packageIDs) == 0 {
		return menuIDs, map[uuid.UUID][]uuid.UUID{}, nil
	}
	err := s.db.Model(&models.FeaturePackageMenu{}).
		Distinct("menu_id").
		Where("package_id IN ?", packageIDs).
		Pluck("menu_id", &menuIDs).Error
	if err != nil {
		return nil, nil, err
	}
	sourceMap := make(map[uuid.UUID][]uuid.UUID)
	for _, packageID := range packageIDs {
		packageMenuIDs, menuErr := s.getMenuIDsByPackageID(packageID)
		if menuErr != nil {
			return nil, nil, menuErr
		}
		for _, menuID := range packageMenuIDs {
			sourceMap[menuID] = appendDerivedPackageID(sourceMap[menuID], packageID)
		}
	}
	return menuIDs, sourceMap, nil
}

func (s *service) getMenuIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.FeaturePackageMenu{}).
		Where("package_id = ?", packageID).
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (s *service) getPackageIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Where("role_id = ? AND enabled = ?", roleID, true).
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) getManualActionIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.TeamManualActionPermission{}).
		Where("tenant_id = ? AND enabled = ?", teamID, true).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
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

func (s *service) getCachedActionIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.TenantActionPermission{}).
		Where("tenant_id = ? AND enabled = ?", teamID, true).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) replaceCachedActionIDs(teamID uuid.UUID, actionIDs []uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ?", teamID).Delete(&models.TenantActionPermission{}).Error; err != nil {
			return err
		}
		if len(actionIDs) == 0 {
			return nil
		}
		records := make([]models.TenantActionPermission, 0, len(actionIDs))
		seen := make(map[uuid.UUID]struct{}, len(actionIDs))
		for _, actionID := range actionIDs {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			records = append(records, models.TenantActionPermission{
				TenantID: teamID,
				ActionID: actionID,
				Enabled:  true,
			})
		}
		if len(records) == 0 {
			return nil
		}
		return tx.Create(&records).Error
	})
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
		return packageContext == "team" || packageContext == "platform,team"
	case "platform":
		return packageContext == "platform" || packageContext == "platform,team"
	default:
		return false
	}
}
