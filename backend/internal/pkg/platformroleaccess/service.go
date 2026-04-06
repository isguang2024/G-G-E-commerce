package platformroleaccess

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
)

var (
	platformRolePublicMenuCacheMu        sync.RWMutex
	platformRolePublicMenuCacheIDs       = map[string][]uuid.UUID{}
	platformRolePublicMenuCacheExpiresAt = map[string]time.Time{}
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
}

type Service interface {
	GetSnapshot(roleID uuid.UUID, appKey ...string) (*Snapshot, error)
	RefreshSnapshot(roleID uuid.UUID, appKey ...string) (*Snapshot, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func InvalidatePublicMenuCache() {
	platformRolePublicMenuCacheMu.Lock()
	defer platformRolePublicMenuCacheMu.Unlock()
	platformRolePublicMenuCacheIDs = map[string][]uuid.UUID{}
	platformRolePublicMenuCacheExpiresAt = map[string]time.Time{}
}

func (s *service) GetSnapshot(roleID uuid.UUID, appKey ...string) (*Snapshot, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	snapshot, err := s.loadSnapshot(roleID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return s.RefreshSnapshot(roleID, resolvedAppKey)
}

func (s *service) RefreshSnapshot(roleID uuid.UUID, appKey ...string) (*Snapshot, error) {
	resolvedAppKey := resolveAppKey(appKey...)
	snapshot, err := s.calculateSnapshot(roleID, resolvedAppKey)
	if err != nil {
		return nil, err
	}
	if err := s.saveSnapshot(roleID, resolvedAppKey, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *service) calculateSnapshot(roleID uuid.UUID, appKey string) (*Snapshot, error) {
	packageIDs, err := s.getPackageIDsByRoleID(roleID, appKey)
	if err != nil {
		return nil, err
	}
	expandedPackageIDs, err := s.expandPackageIDs(packageIDs, "all", appKey)
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

	availableMenuIDs, menuSourceMap, err := s.getMenuIDsByPackageIDs(expandedPackageIDs, appKey)
	if err != nil {
		return nil, err
	}
	publicMenuIDs, err := s.getPublicMenuIDs(appKey)
	if err != nil {
		return nil, err
	}
	availableMenuIDs = mergeUUIDs(availableMenuIDs, publicMenuIDs)
	hiddenMenuIDs, err := s.getHiddenMenuIDsByRoleID(roleID, appKey)
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
	}, nil
}

func (s *service) loadSnapshot(roleID uuid.UUID, appKey string) (*Snapshot, error) {
	var record models.PersonalWorkspaceRoleAccessSnapshot
	if err := s.db.Where("app_key = ? AND role_id = ?", appctx.NormalizeAppKey(appKey), roleID).First(&record).Error; err != nil {
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
	}, nil
}

func (s *service) saveSnapshot(roleID uuid.UUID, appKey string, snapshot *Snapshot) error {
	record := models.PersonalWorkspaceRoleAccessSnapshot{
		AppKey:             appctx.NormalizeAppKey(appKey),
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
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app_key"}, {Name: "role_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) getPackageIDsByRoleID(roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Joins("JOIN feature_packages ON feature_packages.id = role_feature_packages.package_id").
		Where("role_id = ? AND enabled = ?", roleID, true).
		Where("feature_packages.app_key = ? AND feature_packages.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Distinct("package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) expandPackageIDs(packageIDs []uuid.UUID, context string, appKey string) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	packageMap, bundleChildrenMap, err := s.loadPackageGraph(packageIDs, appKey)
	if err != nil {
		return nil, err
	}
	return expandRolePackageIDsFromGraph(packageIDs, context, packageMap, bundleChildrenMap), nil
}

func (s *service) getActionIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	var rows []struct {
		PackageID uuid.UUID
		ActionID  uuid.UUID
	}
	if err := s.db.Model(&models.FeaturePackageKey{}).
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

func (s *service) getMenuIDsByPackageIDs(packageIDs []uuid.UUID, appKey string) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	var rows []struct {
		PackageID uuid.UUID
		MenuID    uuid.UUID
	}
	if err := s.db.Model(&models.FeaturePackageMenu{}).
		Select("feature_package_menus.package_id, feature_package_menus.menu_id").
		Joins("JOIN menu_definitions ON menu_definitions.id = feature_package_menus.menu_id").
		Where("feature_package_menus.package_id IN ?", packageIDs).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
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

func (s *service) getHiddenMenuIDsByRoleID(roleID uuid.UUID, appKey string) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.RoleHiddenMenu{}).
		Joins("JOIN menu_definitions ON menu_definitions.id = role_hidden_menus.menu_id").
		Where("role_hidden_menus.role_id = ?", roleID).
		Where("menu_definitions.app_key = ? AND menu_definitions.deleted_at IS NULL", appctx.NormalizeAppKey(appKey)).
		Distinct("menu_id").
		Pluck("role_hidden_menus.menu_id", &menuIDs).Error
	return menuIDs, err
}

func (s *service) getPublicMenuIDs(appKey string) ([]uuid.UUID, error) {
	normalizedAppKey := appctx.NormalizeAppKey(appKey)
	platformRolePublicMenuCacheMu.RLock()
	if expiresAt, ok := platformRolePublicMenuCacheExpiresAt[normalizedAppKey]; ok && time.Now().Before(expiresAt) {
		cached := append([]uuid.UUID{}, platformRolePublicMenuCacheIDs[normalizedAppKey]...)
		platformRolePublicMenuCacheMu.RUnlock()
		return cached, nil
	}
	platformRolePublicMenuCacheMu.RUnlock()

	var menus []models.MenuDefinition
	if err := s.db.Select("id", "meta").Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).Find(&menus).Error; err != nil {
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
	platformRolePublicMenuCacheMu.Lock()
	platformRolePublicMenuCacheIDs[normalizedAppKey] = append([]uuid.UUID{}, result...)
	platformRolePublicMenuCacheExpiresAt[normalizedAppKey] = time.Now().Add(30 * time.Second)
	platformRolePublicMenuCacheMu.Unlock()
	return result, nil
}

func (s *service) loadPackageGraph(seedIDs []uuid.UUID, appKey string) (map[uuid.UUID]models.FeaturePackage, map[uuid.UUID][]uuid.UUID, error) {
	packageMap := make(map[uuid.UUID]models.FeaturePackage)
	if len(seedIDs) == 0 {
		return packageMap, map[uuid.UUID][]uuid.UUID{}, nil
	}
	var bundleRows []models.FeaturePackageBundle
	if err := s.db.Model(&models.FeaturePackageBundle{}).
		Select("package_id", "child_package_id").
		Find(&bundleRows).Error; err != nil {
		return nil, nil, err
	}
	bundleChildrenMap := make(map[uuid.UUID][]uuid.UUID)
	queue := make([]uuid.UUID, 0, len(seedIDs))
	seen := make(map[uuid.UUID]struct{}, len(seedIDs))
	for _, id := range seedIDs {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		queue = append(queue, id)
	}
	for index := 0; index < len(queue); index++ {
		current := queue[index]
		for _, row := range bundleRows {
			if row.PackageID != current {
				continue
			}
			bundleChildrenMap[row.PackageID] = appendIfMissing(bundleChildrenMap[row.PackageID], row.ChildPackageID)
			if _, ok := seen[row.ChildPackageID]; ok {
				continue
			}
			seen[row.ChildPackageID] = struct{}{}
			queue = append(queue, row.ChildPackageID)
		}
	}
	var packages []models.FeaturePackage
	if err := s.db.Where("app_key = ? AND id IN ? AND status = ?", appctx.NormalizeAppKey(appKey), queue, "normal").Find(&packages).Error; err != nil {
		return nil, nil, err
	}
	for _, item := range packages {
		packageMap[item.ID] = item
	}
	return packageMap, bundleChildrenMap, nil
}

func expandRolePackageIDsFromGraph(seedIDs []uuid.UUID, context string, packageMap map[uuid.UUID]models.FeaturePackage, bundleChildrenMap map[uuid.UUID][]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(seedIDs))
	seenExpanded := make(map[uuid.UUID]struct{}, len(seedIDs))
	visited := make(map[uuid.UUID]struct{}, len(seedIDs))
	var visit func(packageID uuid.UUID)
	visit = func(packageID uuid.UUID) {
		if _, ok := visited[packageID]; ok {
			return
		}
		visited[packageID] = struct{}{}
		item, ok := packageMap[packageID]
		if !ok || !contextAllowsPackage(context, item.ContextType) {
			return
		}
		if item.PackageType == "bundle" {
			for _, childID := range bundleChildrenMap[packageID] {
				visit(childID)
			}
			return
		}
		if _, ok := seenExpanded[packageID]; ok {
			return
		}
		seenExpanded[packageID] = struct{}{}
		result = append(result, packageID)
	}
	for _, packageID := range seedIDs {
		visit(packageID)
	}
	return result
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
	if accessMode := menuAccessMode(meta); accessMode == "public" || accessMode == "jwt" {
		return true
	}
	for _, key := range []string{"isPublic", "public", "globalVisible", "publicMenu", "public_menu"} {
		flag, ok := meta[key].(bool)
		if ok && flag {
			return true
		}
	}
	return false
}

func menuAccessMode(meta map[string]interface{}) string {
	if meta == nil {
		return "permission"
	}
	value, ok := meta["accessMode"].(string)
	if !ok {
		return "permission"
	}
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "public", "jwt":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "permission"
	}
}

func contextAllowsPackage(context, packageContext string) bool {
	if context == "all" {
		return true
	}
	switch packageContext {
	case "", context:
		return true
	case "common":
		return context == "personal" || context == "collaboration"
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

func emptySnapshot() *Snapshot {
	return &Snapshot{
		PackageIDs:         []uuid.UUID{},
		ExpandedPackageIDs: []uuid.UUID{},
		AvailableActionIDs: []uuid.UUID{},
		ActionSourceMap:    map[uuid.UUID][]uuid.UUID{},
		DisabledActionIDs:  []uuid.UUID{},
		EffectiveActionIDs: []uuid.UUID{},
		AvailableMenuIDs:   []uuid.UUID{},
		MenuSourceMap:      map[uuid.UUID][]uuid.UUID{},
		HiddenMenuIDs:      []uuid.UUID{},
		EffectiveMenuIDs:   []uuid.UUID{},
	}
}

func resolveAppKey(appKey ...string) string {
	if len(appKey) == 0 {
		return appctx.NormalizeAppKey("")
	}
	return appctx.NormalizeAppKey(appKey[0])
}
