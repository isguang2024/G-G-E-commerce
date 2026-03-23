package platformaccess

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
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

	menuIDs, menuSourceMap, err := s.getMenuIDsByPackageIDs(expandedPackageIDs)
	if err != nil {
		return nil, err
	}
	publicMenuIDs, err := s.getPublicMenuIDs()
	if err != nil {
		return nil, err
	}
	menuIDs = mergeUUIDs(menuIDs, publicMenuIDs)
	availableMenuIDs := append([]uuid.UUID{}, menuIDs...)
	availableMenuMap := filterSourceMap(menuSourceMap, availableMenuIDs)
	hiddenMenuIDs, err := s.getHiddenMenuIDs(userID, roleIDs)
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
	err := s.db.Where("id = ? AND status = ?", packageID, "normal").First(&item).Error
	if err != nil {
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

func (s *service) getHiddenMenuIDs(userID uuid.UUID, roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	hiddenMenuIDs := make([]uuid.UUID, 0)
	if len(roleIDs) > 0 {
		var roleHidden []uuid.UUID
		if err := s.db.Model(&models.RoleHiddenMenu{}).
			Where("role_id IN ?", roleIDs).
			Distinct("menu_id").
			Pluck("menu_id", &roleHidden).Error; err != nil {
			return nil, err
		}
		hiddenMenuIDs = mergeUUIDs(hiddenMenuIDs, roleHidden)
	}
	var userHidden []uuid.UUID
	if err := s.db.Model(&models.UserHiddenMenu{}).
		Where("user_id = ?", userID).
		Distinct("menu_id").
		Pluck("menu_id", &userHidden).Error; err != nil {
		return nil, err
	}
	return mergeUUIDs(hiddenMenuIDs, userHidden), nil
}

func (s *service) getPublicMenuIDs() ([]uuid.UUID, error) {
	var menus []models.Menu
	if err := s.db.Find(&menus).Error; err != nil {
		return nil, err
	}
	result := make([]uuid.UUID, 0)
	for _, item := range menus {
		if isPublicMenu(item.Meta) {
			result = append(result, item.ID)
		}
	}
	return result, nil
}

func isPublicMenu(meta models.MetaJSON) bool {
	if meta == nil {
		return false
	}
	for _, key := range []string{"isPublic", "public", "globalVisible"} {
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
		return packageContext == "platform" || packageContext == "platform,team"
	case "team":
		return packageContext == "team" || packageContext == "platform,team"
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
