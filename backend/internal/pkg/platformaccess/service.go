package platformaccess

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

var (
	platformPublicMenuCacheMu        sync.RWMutex
	platformPublicMenuCacheIDs       []uuid.UUID
	platformPublicMenuCacheExpiresAt time.Time
)

type Snapshot struct {
	RoleIDs            []uuid.UUID
	RolePackageIDs     []uuid.UUID
	UserPackageIDs     []uuid.UUID
	DirectPackageIDs   []uuid.UUID
	ExpandedPackageIDs []uuid.UUID
	ActionIDs          []uuid.UUID
	ActionSourceMap    map[uuid.UUID][]uuid.UUID
	AvailableMenuIDs   []uuid.UUID
	AvailableMenuMap   map[uuid.UUID][]uuid.UUID
	MenuIDs            []uuid.UUID
	MenuSourceMap      map[uuid.UUID][]uuid.UUID
	HiddenMenuIDs      []uuid.UUID
	DisabledActionIDs  []uuid.UUID
	HasPackageConfig   bool
}

type Service interface {
	GetSnapshot(userID uuid.UUID) (*Snapshot, error)
	RefreshSnapshot(userID uuid.UUID) (*Snapshot, error)
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func InvalidatePublicMenuCache() {
	platformPublicMenuCacheMu.Lock()
	defer platformPublicMenuCacheMu.Unlock()
	platformPublicMenuCacheIDs = nil
	platformPublicMenuCacheExpiresAt = time.Time{}
}

func (s *service) GetSnapshot(userID uuid.UUID) (*Snapshot, error) {
	snapshot, err := s.loadSnapshot(userID)
	if err != nil {
		return nil, err
	}
	if snapshot != nil {
		return snapshot, nil
	}
	return s.RefreshSnapshot(userID)
}

func (s *service) RefreshSnapshot(userID uuid.UUID) (*Snapshot, error) {
	snapshot, err := s.calculateSnapshot(userID)
	if err != nil {
		return nil, err
	}
	if err := s.saveSnapshot(userID, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (s *service) calculateSnapshot(userID uuid.UUID) (*Snapshot, error) {
	roleIDs, err := s.getGlobalRoleIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	rolePackageIDs, err := s.getPackageIDsByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	userPackageIDs, err := s.getPackageIDsByUserID(userID)
	if err != nil {
		return nil, err
	}
	directPackageIDs := mergeUUIDs(rolePackageIDs, userPackageIDs)
	expandedPackageIDs, err := s.expandPackageIDs(directPackageIDs, "platform")
	if err != nil {
		return nil, err
	}
	actionIDs, actionSourceMap, err := s.getActionIDsByPackageIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	disabledActionIDs, err := s.getDisabledActionIDsByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	actionIDs = subtractUUIDs(actionIDs, disabledActionIDs)
	actionSourceMap = filterSourceMap(actionSourceMap, actionIDs)

	menuIDs, menuSourceMap, err := s.getEffectiveMenuContributionByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}
	userExpandedPackageIDs, err := s.expandPackageIDs(userPackageIDs, "platform")
	if err != nil {
		return nil, err
	}
	userMenuIDs, userMenuSourceMap, err := s.getMenuIDsByPackageIDs(userExpandedPackageIDs)
	if err != nil {
		return nil, err
	}
	menuIDs = mergeUUIDs(menuIDs, userMenuIDs)
	menuSourceMap = mergeSourceMaps(menuSourceMap, userMenuSourceMap)
	publicMenuIDs, err := s.getPublicMenuIDs()
	if err != nil {
		return nil, err
	}
	menuIDs = mergeUUIDs(menuIDs, publicMenuIDs)
	availableMenuIDs := append([]uuid.UUID{}, menuIDs...)
	availableMenuMap := filterSourceMap(menuSourceMap, availableMenuIDs)
	hiddenMenuIDs, err := s.getUserHiddenMenuIDs(userID)
	if err != nil {
		return nil, err
	}
	menuIDs = subtractUUIDs(menuIDs, hiddenMenuIDs)
	menuSourceMap = filterSourceMap(menuSourceMap, menuIDs)

	return &Snapshot{
		RoleIDs:            roleIDs,
		RolePackageIDs:     rolePackageIDs,
		UserPackageIDs:     userPackageIDs,
		DirectPackageIDs:   directPackageIDs,
		ExpandedPackageIDs: expandedPackageIDs,
		ActionIDs:          actionIDs,
		ActionSourceMap:    actionSourceMap,
		AvailableMenuIDs:   availableMenuIDs,
		AvailableMenuMap:   availableMenuMap,
		MenuIDs:            menuIDs,
		MenuSourceMap:      menuSourceMap,
		HiddenMenuIDs:      hiddenMenuIDs,
		DisabledActionIDs:  disabledActionIDs,
		HasPackageConfig:   len(directPackageIDs) > 0,
	}, nil
}

func (s *service) loadSnapshot(userID uuid.UUID) (*Snapshot, error) {
	var record models.PlatformUserAccessSnapshot
	if err := s.db.Where("user_id = ?", userID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &Snapshot{
		RoleIDs:            uuidStringsToIDs(record.RoleIDs),
		RolePackageIDs:     uuidStringsToIDs(record.RolePackageIDs),
		UserPackageIDs:     uuidStringsToIDs(record.UserPackageIDs),
		DirectPackageIDs:   uuidStringsToIDs(record.DirectPackageIDs),
		ExpandedPackageIDs: uuidStringsToIDs(record.ExpandedPackageIDs),
		ActionIDs:          uuidStringsToIDs(record.ActionIDs),
		ActionSourceMap:    sourceMapStringsToUUIDs(record.ActionSourceMap),
		AvailableMenuIDs:   uuidStringsToIDs(record.AvailableMenuIDs),
		AvailableMenuMap:   sourceMapStringsToUUIDs(record.AvailableMenuMap),
		MenuIDs:            uuidStringsToIDs(record.MenuIDs),
		MenuSourceMap:      sourceMapStringsToUUIDs(record.MenuSourceMap),
		HiddenMenuIDs:      uuidStringsToIDs(record.HiddenMenuIDs),
		DisabledActionIDs:  uuidStringsToIDs(record.DisabledActionIDs),
		HasPackageConfig:   record.HasPackageConfig,
	}, nil
}

func (s *service) saveSnapshot(userID uuid.UUID, snapshot *Snapshot) error {
	record := models.PlatformUserAccessSnapshot{
		UserID:             userID,
		RoleIDs:            idsToUUIDStrings(snapshot.RoleIDs),
		RolePackageIDs:     idsToUUIDStrings(snapshot.RolePackageIDs),
		UserPackageIDs:     idsToUUIDStrings(snapshot.UserPackageIDs),
		DirectPackageIDs:   idsToUUIDStrings(snapshot.DirectPackageIDs),
		ExpandedPackageIDs: idsToUUIDStrings(snapshot.ExpandedPackageIDs),
		ActionIDs:          idsToUUIDStrings(snapshot.ActionIDs),
		ActionSourceMap:    sourceMapUUIDsToStrings(snapshot.ActionSourceMap),
		AvailableMenuIDs:   idsToUUIDStrings(snapshot.AvailableMenuIDs),
		AvailableMenuMap:   sourceMapUUIDsToStrings(snapshot.AvailableMenuMap),
		MenuIDs:            idsToUUIDStrings(snapshot.MenuIDs),
		MenuSourceMap:      sourceMapUUIDsToStrings(snapshot.MenuSourceMap),
		HiddenMenuIDs:      idsToUUIDStrings(snapshot.HiddenMenuIDs),
		DisabledActionIDs:  idsToUUIDStrings(snapshot.DisabledActionIDs),
		HasPackageConfig:   snapshot.HasPackageConfig,
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		UpdateAll: true,
	}).Create(&record).Error
}

func (s *service) getGlobalRoleIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	err := s.db.Model(&models.UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.tenant_id IS NULL").
		Where("roles.status = ?", "normal").
		Distinct("user_roles.role_id").
		Pluck("user_roles.role_id", &roleIDs).Error
	return roleIDs, err
}

func (s *service) getPackageIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Where("role_id IN ? AND enabled = ?", roleIDs, true).
		Distinct("package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) getPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.UserFeaturePackage{}).
		Where("user_id = ? AND enabled = ?", userID, true).
		Distinct("package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) expandPackageIDs(packageIDs []uuid.UUID, context string) ([]uuid.UUID, error) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	packageMap, bundleChildrenMap, err := s.loadPackageGraph(packageIDs)
	if err != nil {
		return nil, err
	}
	return expandPackageIDsFromGraph(packageIDs, context, packageMap, bundleChildrenMap), nil
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

func (s *service) getDisabledActionIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	var actionIDs []uuid.UUID
	err := s.db.Model(&models.RoleDisabledAction{}).
		Where("role_id IN ?", roleIDs).
		Distinct("action_id").
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (s *service) getEffectiveMenuContributionByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID, error) {
	if len(roleIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}, nil
	}
	rolePackageMap, allPackageIDs, err := s.getPackageIDsByRoleIDsMap(roleIDs)
	if err != nil {
		return nil, nil, err
	}
	packageMap, bundleChildrenMap, err := s.loadPackageGraph(allPackageIDs)
	if err != nil {
		return nil, nil, err
	}
	roleExpandedPackageMap := make(map[uuid.UUID][]uuid.UUID, len(roleIDs))
	allExpandedPackageIDs := make([]uuid.UUID, 0, len(allPackageIDs))
	expandedSeen := make(map[uuid.UUID]struct{}, len(allPackageIDs))
	for _, roleID := range roleIDs {
		expandedPackageIDs := expandPackageIDsFromGraph(rolePackageMap[roleID], "platform", packageMap, bundleChildrenMap)
		roleExpandedPackageMap[roleID] = expandedPackageIDs
		for _, packageID := range expandedPackageIDs {
			if _, ok := expandedSeen[packageID]; ok {
				continue
			}
			expandedSeen[packageID] = struct{}{}
			allExpandedPackageIDs = append(allExpandedPackageIDs, packageID)
		}
	}
	packageMenuMap, err := s.getPackageMenuMapByPackageIDs(allExpandedPackageIDs)
	if err != nil {
		return nil, nil, err
	}
	roleHiddenMenuMap, err := s.getHiddenMenuIDsByRoleIDsMap(roleIDs)
	if err != nil {
		return nil, nil, err
	}
	menuIDs := make([]uuid.UUID, 0)
	menuSeen := make(map[uuid.UUID]struct{})
	sourceMap := make(map[uuid.UUID][]uuid.UUID)
	for _, roleID := range roleIDs {
		roleMenuIDs, roleSourceMap := buildMenuContributionByPackageIDs(roleExpandedPackageMap[roleID], packageMenuMap)
		roleHiddenMenuIDs := roleHiddenMenuMap[roleID]
		effectiveRoleMenuIDs := subtractUUIDs(roleMenuIDs, roleHiddenMenuIDs)
		for _, menuID := range effectiveRoleMenuIDs {
			if _, ok := menuSeen[menuID]; ok {
				continue
			}
			menuSeen[menuID] = struct{}{}
			menuIDs = append(menuIDs, menuID)
		}
		sourceMap = mergeSourceMaps(sourceMap, filterSourceMap(roleSourceMap, effectiveRoleMenuIDs))
	}
	return menuIDs, sourceMap, nil
}

func (s *service) getPackageIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := s.db.Model(&models.RoleFeaturePackage{}).
		Where("role_id = ? AND enabled = ?", roleID, true).
		Distinct("package_id").
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (s *service) getPackageIDsByRoleIDsMap(roleIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, []uuid.UUID, error) {
	result := make(map[uuid.UUID][]uuid.UUID, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, []uuid.UUID{}, nil
	}
	var rows []struct {
		RoleID    uuid.UUID
		PackageID uuid.UUID
	}
	if err := s.db.Model(&models.RoleFeaturePackage{}).
		Select("role_id, package_id").
		Where("role_id IN ? AND enabled = ?", roleIDs, true).
		Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	for _, roleID := range roleIDs {
		result[roleID] = []uuid.UUID{}
	}
	allPackageIDs := make([]uuid.UUID, 0, len(rows))
	allSeen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		result[row.RoleID] = appendIfMissing(result[row.RoleID], row.PackageID)
		if _, ok := allSeen[row.PackageID]; ok {
			continue
		}
		allSeen[row.PackageID] = struct{}{}
		allPackageIDs = append(allPackageIDs, row.PackageID)
	}
	return result, allPackageIDs, nil
}

func (s *service) getHiddenMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := s.db.Model(&models.RoleHiddenMenu{}).
		Where("role_id = ?", roleID).
		Distinct("menu_id").
		Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (s *service) getHiddenMenuIDsByRoleIDsMap(roleIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	result := make(map[uuid.UUID][]uuid.UUID, len(roleIDs))
	if len(roleIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		RoleID uuid.UUID
		MenuID uuid.UUID
	}
	if err := s.db.Model(&models.RoleHiddenMenu{}).
		Select("role_id, menu_id").
		Where("role_id IN ?", roleIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, roleID := range roleIDs {
		result[roleID] = []uuid.UUID{}
	}
	for _, row := range rows {
		result[row.RoleID] = appendIfMissing(result[row.RoleID], row.MenuID)
	}
	return result, nil
}

func (s *service) getUserHiddenMenuIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	var userHidden []uuid.UUID
	if err := s.db.Model(&models.UserHiddenMenu{}).
		Where("user_id = ?", userID).
		Distinct("menu_id").
		Pluck("menu_id", &userHidden).Error; err != nil {
		return nil, err
	}
	return userHidden, nil
}

func (s *service) getPublicMenuIDs() ([]uuid.UUID, error) {
	platformPublicMenuCacheMu.RLock()
	if time.Now().Before(platformPublicMenuCacheExpiresAt) {
		cached := append([]uuid.UUID{}, platformPublicMenuCacheIDs...)
		platformPublicMenuCacheMu.RUnlock()
		return cached, nil
	}
	platformPublicMenuCacheMu.RUnlock()

	var menus []models.Menu
	if err := s.db.Select("id", "meta").Find(&menus).Error; err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0)
	for _, item := range menus {
		if !isMenuEnabled(item.Meta) {
			continue
		}
		if isPublicMenu(item.Meta) {
			result = append(result, item.ID)
		}
	}
	platformPublicMenuCacheMu.Lock()
	platformPublicMenuCacheIDs = append([]uuid.UUID{}, result...)
	platformPublicMenuCacheExpiresAt = time.Now().Add(30 * time.Second)
	platformPublicMenuCacheMu.Unlock()
	return result, nil
}

func (s *service) getPackageMenuMapByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	result := make(map[uuid.UUID][]uuid.UUID)
	if len(packageIDs) == 0 {
		return result, nil
	}
	var rows []struct {
		PackageID uuid.UUID
		MenuID    uuid.UUID
	}
	if err := s.db.Model(&models.FeaturePackageMenu{}).
		Select("package_id, menu_id").
		Where("package_id IN ?", packageIDs).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.PackageID] = appendIfMissing(result[row.PackageID], row.MenuID)
	}
	return result, nil
}

func (s *service) loadPackageGraph(seedIDs []uuid.UUID) (map[uuid.UUID]models.FeaturePackage, map[uuid.UUID][]uuid.UUID, error) {
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
	if err := s.db.Where("id IN ? AND status = ?", queue, "normal").Find(&packages).Error; err != nil {
		return nil, nil, err
	}
	for _, item := range packages {
		packageMap[item.ID] = item
	}
	return packageMap, bundleChildrenMap, nil
}

func expandPackageIDsFromGraph(seedIDs []uuid.UUID, context string, packageMap map[uuid.UUID]models.FeaturePackage, bundleChildrenMap map[uuid.UUID][]uuid.UUID) []uuid.UUID {
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

func isMenuEnabled(meta models.MetaJSON) bool {
	if meta == nil {
		return true
	}
	if enabled, ok := meta["isEnable"].(bool); ok {
		return enabled
	}
	return true
}

func isPublicMenu(meta models.MetaJSON) bool {
	if meta == nil {
		return false
	}
	if accessMode := menuAccessMode(meta); accessMode == "public" || accessMode == "jwt" {
		return true
	}
	for _, key := range []string{"isPublic", "public", "globalVisible", "publicMenu", "public_menu"} {
		value, ok := meta[key]
		if !ok {
			continue
		}
		flag, ok := value.(bool)
		if ok && flag {
			return true
		}
	}
	return false
}

func menuAccessMode(meta models.MetaJSON) string {
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

func buildMenuContributionByPackageIDs(packageIDs []uuid.UUID, packageMenuMap map[uuid.UUID][]uuid.UUID) ([]uuid.UUID, map[uuid.UUID][]uuid.UUID) {
	if len(packageIDs) == 0 {
		return []uuid.UUID{}, map[uuid.UUID][]uuid.UUID{}
	}
	menuIDs := make([]uuid.UUID, 0)
	menuSeen := make(map[uuid.UUID]struct{})
	sourceMap := make(map[uuid.UUID][]uuid.UUID)
	for _, packageID := range packageIDs {
		for _, menuID := range packageMenuMap[packageID] {
			if _, ok := menuSeen[menuID]; !ok {
				menuSeen[menuID] = struct{}{}
				menuIDs = append(menuIDs, menuID)
			}
			sourceMap[menuID] = appendIfMissing(sourceMap[menuID], packageID)
		}
	}
	return menuIDs, sourceMap
}

func appendIfMissing(current []uuid.UUID, value uuid.UUID) []uuid.UUID {
	for _, item := range current {
		if item == value {
			return current
		}
	}
	return append(current, value)
}

func mergeUUIDs(groups ...[]uuid.UUID) []uuid.UUID {
	result := make([]uuid.UUID, 0)
	seen := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, id := range group {
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			result = append(result, id)
		}
	}
	return result
}

func mergeSourceMaps(base map[uuid.UUID][]uuid.UUID, incoming map[uuid.UUID][]uuid.UUID) map[uuid.UUID][]uuid.UUID {
	if len(base) == 0 && len(incoming) == 0 {
		return map[uuid.UUID][]uuid.UUID{}
	}
	result := make(map[uuid.UUID][]uuid.UUID, len(base)+len(incoming))
	for key, values := range base {
		result[key] = append([]uuid.UUID{}, values...)
	}
	for key, values := range incoming {
		current := result[key]
		for _, value := range values {
			current = appendIfMissing(current, value)
		}
		result[key] = current
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

func contextAllowsPackage(targetContext, packageContext string) bool {
	if packageContext == "" {
		return true
	}
	switch targetContext {
	case "platform":
		return packageContext == "platform" || packageContext == "common"
	case "team":
		return packageContext == "team" || packageContext == "common"
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
