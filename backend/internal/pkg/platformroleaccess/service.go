package platformroleaccess

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

type Snapshot struct {
	PackageIDs         []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	AvailableActionIDs []uuid.UUID
	ActionSourceMap    map[uuid.UUID][]uuid.UUID
	DisabledActionIDs  []uuid.UUID
	EffectiveActionIDs []uuid.UUID
	AvailableMenuIDs   []uuid.UUID
	MenuSourceMap      map[uuid.UUID][]uuid.UUID
	HiddenMenuIDs      []uuid.UUID
	EffectiveMenuIDs   []uuid.UUID
	HasPackageConfig   bool
}

type Service interface {
	GetSnapshot(roleID uuid.UUID) (*Snapshot, error)
	RefreshSnapshot(roleID uuid.UUID) (*Snapshot, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func (s *service) GetSnapshot(roleID uuid.UUID) (*Snapshot, error) {
	snapshot, err := s.loadSnapshot(roleID)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return s.RefreshSnapshot(roleID)
}

func (s *service) RefreshSnapshot(roleID uuid.UUID) (*Snapshot, error) {
	snapshot, err := s.calculateSnapshot(roleID)
	if err != nil {
		return nil, err
	}
	if err := s.saveSnapshot(roleID, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *service) calculateSnapshot(roleID uuid.UUID) (*Snapshot, error) {
	packageIDs, err := s.getPackageIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	expandedPackageIDs, err := s.expandPackageIDs(packageIDs, "platform")
	if err != nil {
		return nil, err
	}

	availableActionIDs, actionSourceMap, err := s.getActionIDsByPackageIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	disabledActionIDs, err := s.getDisabledActionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	effectiveActionIDs := subtractUUIDs(availableActionIDs, disabledActionIDs)

	availableMenuIDs, menuSourceMap, err := s.getMenuIDsByPackageIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	publicMenuIDs, err := s.getPublicMenuIDs()
	if err != nil {
		return nil, err
	}
	availableMenuIDs = mergeUUIDs(availableMenuIDs, publicMenuIDs)
	hiddenMenuIDs, err := s.getHiddenMenuIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	effectiveMenuIDs := subtractUUIDs(availableMenuIDs, hiddenMenuIDs)

	return &Snapshot{
		PackageIDs:         packageIDs,
		ExpandedPackageIDs: expandedPackageIDs,
		AvailableActionIDs: availableActionIDs,
		ActionSourceMap:    filterSourceMap(actionSourceMap, availableActionIDs),
		DisabledActionIDs:  disabledActionIDs,
		EffectiveActionIDs: effectiveActionIDs,
		AvailableMenuIDs:   availableMenuIDs,
		MenuSourceMap:      filterSourceMap(menuSourceMap, availableMenuIDs),
		HiddenMenuIDs:      hiddenMenuIDs,
		EffectiveMenuIDs:   effectiveMenuIDs,
		HasPackageConfig:   len(packageIDs) > 0,
	}, nil
}

func (s *service) loadSnapshot(roleID uuid.UUID) (*Snapshot, error) {
	var record models.PlatformRoleAccessSnapshot
	if err := s.db.Where("role_id = ?", roleID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &Snapshot{
		PackageIDs:         uuidStringsToIDs(record.PackageIDs),
		ExpandedPackageIDs: uuidStringsToIDs(record.ExpandedPackageIDs),
		AvailableActionIDs: uuidStringsToIDs(record.AvailableActionIDs),
		ActionSourceMap:    sourceMapStringsToUUIDs(record.ActionSourceMap),
		DisabledActionIDs:  uuidStringsToIDs(record.DisabledActionIDs),
		EffectiveActionIDs: uuidStringsToIDs(record.EffectiveActionIDs),
		AvailableMenuIDs:   uuidStringsToIDs(record.AvailableMenuIDs),
		MenuSourceMap:      sourceMapStringsToUUIDs(record.MenuSourceMap),
		HiddenMenuIDs:      uuidStringsToIDs(record.HiddenMenuIDs),
		EffectiveMenuIDs:   uuidStringsToIDs(record.EffectiveMenuIDs),
		HasPackageConfig:   record.HasPackageConfig,
	}, nil
}

func (s *service) saveSnapshot(roleID uuid.UUID, snapshot *Snapshot) error {
	record := models.PlatformRoleAccessSnapshot{
		RoleID:             roleID,
		PackageIDs:         idsToUUIDStrings(snapshot.PackageIDs),
		ExpandedPackageIDs: idsToUUIDStrings(snapshot.ExpandedPackageIDs),
		AvailableActionIDs: idsToUUIDStrings(snapshot.AvailableActionIDs),
		ActionSourceMap:    sourceMapUUIDsToStrings(snapshot.ActionSourceMap),
		DisabledActionIDs:  idsToUUIDStrings(snapshot.DisabledActionIDs),
		EffectiveActionIDs: idsToUUIDStrings(snapshot.EffectiveActionIDs),
		AvailableMenuIDs:   idsToUUIDStrings(snapshot.AvailableMenuIDs),
		MenuSourceMap:      sourceMapUUIDsToStrings(snapshot.MenuSourceMap),
		HiddenMenuIDs:      idsToUUIDStrings(snapshot.HiddenMenuIDs),
		EffectiveMenuIDs:   idsToUUIDStrings(snapshot.EffectiveMenuIDs),
		HasPackageConfig:   snapshot.HasPackageConfig,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) getPackageIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Where("role_id = ? AND enabled = ?", roleID, true).
		Distinct("package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) expandPackageIDs(packageIDs []uuid.UUID, context string) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	result := make([]uuid.UUID, 0, len(packageIDs))
	seenExpanded := make(map[uuid.UUID]struct{}, len(packageIDs))
	visited := make(map[uuid.UUID]struct{}, len(packageIDs))
	for _, packageID := range packageIDs {
		if err := s.expandPackageID(packageID, context, visited, seenExpanded, &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *service) expandPackageID(packageID uuid.UUID, context string, visited map[uuid.UUID]struct{}, seenExpanded map[uuid.UUID]struct{}, result *[]uuid.UUID) error {
	if _, ok := visited[packageID]; ok {
		return nil
	}
	visited[packageID] = struct{}{}

	var item models.FeaturePackage
	if err := s.db.Where("id = ? AND status = ?", packageID, "normal").First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	if !contextAllowsPackage(context, item.ContextType) {
		return nil
	}
	if item.PackageType == "bundle" {
		var childIDs []uuid.UUID
		if err := s.db.Model(&models.FeaturePackageBundle{}).
			Where("package_id = ?", packageID).
			Pluck("child_package_id", &childIDs).Error; err != nil {
			return err
		}
		for _, childID := range childIDs {
			if err := s.expandPackageID(childID, context, visited, seenExpanded, result); err != nil {
				return err
			}
		}
		return nil
	}
	if _, ok := seenExpanded[packageID]; ok {
		return nil
	}
	seenExpanded[packageID] = struct{}{}
	*result = append(*result, packageID)
	return nil
}

func (s *service) getActionIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	var rows []struct {
		PackageID uuid.UUID
		ActionID  uuid.UUID
	}
	if err := s.db.Model(&models.FeaturePackageAction{}).
		Select("package_id, action_id").
		Where("package_id IN ?", packageIDs).
		Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	result := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	sourceMap := make(map[uuid.UUID][]uuid.UUID)
	for _, row := range rows {
		sourceMap[row.ActionID] = appendIfMissing(sourceMap[row.ActionID], row.PackageID)
		if _, ok := seen[row.ActionID]; ok {
			continue
		}
		seen[row.ActionID] = struct{}{}
		result = append(result, row.ActionID)
	}
	return result, sourceMap, nil
}

func (s *service) getMenuIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	var rows []struct {
		PackageID uuid.UUID
		MenuID    uuid.UUID
	}
	if err := s.db.Model(&models.FeaturePackageMenu{}).
		Select("package_id, menu_id").
		Where("package_id IN ?", packageIDs).
		Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	result := make([]uuid.UUID, 0, len(rows))
	seen := make(map[uuid.UUID]struct{}, len(rows))
	sourceMap := make(map[uuid.UUID][]uuid.UUID)
	for _, row := range rows {
		sourceMap[row.MenuID] = appendIfMissing(sourceMap[row.MenuID], row.PackageID)
		if _, ok := seen[row.MenuID]; ok {
			continue
		}
		seen[row.MenuID] = struct{}{}
		result = append(result, row.MenuID)
	}
	return result, sourceMap, nil
}

func (s *service) getDisabledActionIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.RoleDisabledAction{}).
		Where("role_id = ?", roleID).
		Distinct("action_id").
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) getHiddenMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.RoleHiddenMenu{}).
		Where("role_id = ?", roleID).
		Distinct("menu_id").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (s *service) getPublicMenuIDs() ([]uuid.UUID, error) {
	var menus []models.Menu
	if err := s.db.Find(&menus).Error; err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0)
	for _, item := range menus {
		if item.Meta == nil {
			continue
		}
		if !isMenuEnabled(item.Meta) {
			continue
		}
		if isPublicMenu(item.Meta) {
			result = append(result, item.ID)
		}
	}
	return result, nil
}

func isMenuEnabled(meta map[string]interface{}) bool {
	if meta == nil {
		return true
	}
	if enabled, ok := meta["isEnable"].(bool); ok {
		return enabled
	}
	return true
}

func isPublicMenu(meta map[string]interface{}) bool {
	if meta == nil {
		return false
	}
	if public, ok := meta["isPublic"].(bool); ok {
		return public
	}
	return false
}

func contextAllowsPackage(context, packageContext string) bool {
	switch packageContext {
	case "", context:
		return true
	case "platform,team", "team,platform":
		return context == "platform" || context == "team"
	default:
		return false
	}
}

func mergeUUIDs(first, second []uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(first)+len(second))
	seen := make(map[uuid.UUID]struct{}, len(first)+len(second))
	for _, collection := range [][]uuid.UUID{first, second} {
		for _, item := range collection {
			if item == uuid.Nil {
				continue
			}
			if _, ok := seen[item]; ok {
				continue
			}
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func subtractUUIDs(source, blocked []uuid.UUID) []uuid.UUID {
	if len(source) == 0 {
		return []uuid.UUID{}
	}
	if len(blocked) == 0 {
		return append([]uuid.UUID{}, source...)
	}
	blockedSet := make(map[uuid.UUID]struct{}, len(blocked))
	for _, item := range blocked {
		blockedSet[item] = struct{}{}
	}
	result := make([]uuid.UUID, 0, len(source))
	seen := make(map[uuid.UUID]struct{}, len(source))
	for _, item := range source {
		if _, ok := blockedSet[item]; ok {
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

func filterSourceMap(sourceMap map[uuid.UUID][]uuid.UUID, allowedIDs []uuid.UUID) map[uuid.UUID][]uuid.UUID {
	if len(sourceMap) == 0 || len(allowedIDs) == 0 {
		return map[uuid.UUID][]uuid.UUID{}
	}
	allowedSet := make(map[uuid.UUID]struct{}, len(allowedIDs))
	for _, item := range allowedIDs {
		allowedSet[item] = struct{}{}
	}
	result := make(map[uuid.UUID][]uuid.UUID)
	for id, packageIDs := range sourceMap {
		if _, ok := allowedSet[id]; !ok {
			continue
		}
		result[id] = append([]uuid.UUID{}, packageIDs...)
	}
	return result
}

func appendIfMissing(items []uuid.UUID, value uuid.UUID) []uuid.UUID {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func uuidStringsToIDs(items []string) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if id, err := uuid.Parse(item); err == nil {
			result = append(result, id)
		}
	}
	return result
}

func idsToUUIDStrings(items []uuid.UUID) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item == uuid.Nil {
			continue
		}
		result = append(result, item.String())
	}
	return result
}

func sourceMapStringsToUUIDs(source map[string][]string) map[uuid.UUID][]uuid.UUID {
	result := make(map[uuid.UUID][]uuid.UUID, len(source))
	for key, values := range source {
		id, err := uuid.Parse(key)
		if err != nil {
			continue
		}
		result[id] = uuidStringsToIDs(values)
	}
	return result
}

func sourceMapUUIDsToStrings(source map[uuid.UUID][]uuid.UUID) map[string][]string {
	result := make(map[string][]string, len(source))
	for key, values := range source {
		result[key.String()] = idsToUUIDStrings(values)
	}
	return result
}
