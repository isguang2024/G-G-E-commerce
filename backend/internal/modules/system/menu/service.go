package menu

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	apppkg "github.com/gg-ecommerce/backend/internal/modules/system/app"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	page "github.com/gg-ecommerce/backend/internal/modules/system/page"
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
	"github.com/gg-ecommerce/backend/internal/pkg/platformaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/platformroleaccess"
)

var (
	ErrMenuNotFound          = errors.New("menu not found")
	ErrMenuSystemProtected   = errors.New("系统默认菜单不可删除")
	ErrMenuGroupNotFound     = errors.New("menu group not found")
	ErrMenuGroupInUse        = errors.New("menu group in use")
	ErrMenuDeleteModeInvalid = errors.New("无效的菜单删除方式")
	ErrMenuHasChildren       = errors.New("该菜单存在子菜单，请选择删除策略")
	ErrMenuBackupAppMismatch = errors.New("menu backup app mismatch")
)

type MenuService interface {
	GetTree(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error)
	Create(req *dto.MenuCreateRequest) (*user.Menu, error)
	Update(id uuid.UUID, req *dto.MenuUpdateRequest) error
	DeletePreview(id uuid.UUID, mode string, targetParentID *uuid.UUID) (*MenuDeletePreview, error)
	Delete(id uuid.UUID, mode string, targetParentID *uuid.UUID) error
	ListGroups() ([]user.MenuManageGroup, error)
	CreateGroup(req *dto.MenuManageGroupCreateRequest) (*user.MenuManageGroup, error)
	UpdateGroup(id uuid.UUID, req *dto.MenuManageGroupUpdateRequest) error
	DeleteGroup(id uuid.UUID) error
	// 菜单备份相关方法
	CreateBackup(name, description, scopeType, appKey, spaceKey string, createdBy *uuid.UUID) error
	ListBackups(appKey, spaceKey string) ([]*user.MenuBackup, error)
	DeleteBackup(id uuid.UUID, appKey string) error
	RestoreBackup(id uuid.UUID, appKey string) error
}

type menuService struct {
	db        *gorm.DB
	menuRepo  user.MenuRepository
	refresher permissionrefresh.Service
	logger    *zap.Logger
}

type MenuDeletePreview struct {
	Mode                  string `json:"mode"`
	MenuCount             int    `json:"menu_count"`
	ChildCount            int    `json:"child_count"`
	AffectedPageCount     int    `json:"affected_page_count"`
	AffectedRelationCount int    `json:"affected_relation_count"`
}

func NewMenuService(db *gorm.DB, menuRepo user.MenuRepository, refresher permissionrefresh.Service, logger *zap.Logger) MenuService {
	return &menuService{db: db, menuRepo: menuRepo, refresher: refresher, logger: logger}
}

func (s *menuService) GetTree(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error) {
	flat, err := s.menuRepo.ListByAppAndSpace(normalizeMenuAppKey(appKey), normalizeMenuSpaceKey(spaceKey))
	if err != nil {
		return nil, err
	}
	normalizeMenuListKinds(flat)
	tree := user.BuildTree(flat, nil)
	if all {
		return tree, nil
	}
	merged := s.mergeSystemMenuIDs(flat, allowedMenuIDs)
	visibleIDs := s.visibleMenuIDs(flat, merged)
	return s.filterTreeByMenuIDs(tree, visibleIDs), nil
}

func (s *menuService) mergeSystemMenuIDs(flat []user.Menu, allowed []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{})
	for _, id := range allowed {
		set[id] = struct{}{}
	}

	// 对 jwt/public 菜单做默认放行，不依赖功能权限分配。
	for _, menu := range flat {
		accessMode := menuAccessMode(menu.Meta)
		if accessMode == "jwt" || accessMode == "public" {
			set[menu.ID] = struct{}{}
		}
	}

	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

func menuAccessMode(meta map[string]interface{}) string {
	if meta == nil {
		return "permission"
	}
	raw, ok := meta["accessMode"]
	if !ok {
		return "permission"
	}
	value, ok := raw.(string)
	if !ok {
		return "permission"
	}
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "public", "jwt", "permission":
		return value
	default:
		return "permission"
	}
}

func filterMenusBySpace(flat []user.Menu, spaceKey string) []user.Menu {
	target := spaceutil.NormalizeSpaceKey(spaceKey)
	if target == "" {
		return flat
	}
	result := make([]user.Menu, 0, len(flat))
	for _, menu := range flat {
		if spaceutil.NormalizeSpaceKey(menu.SpaceKey) != target {
			continue
		}
		result = append(result, menu)
	}
	return result
}

func filterMenusByApp(flat []user.Menu, appKey string) []user.Menu {
	target := normalizeMenuAppKey(appKey)
	if target == "" {
		return flat
	}
	result := make([]user.Menu, 0, len(flat))
	for _, menu := range flat {
		if normalizeMenuAppKey(menu.AppKey) != target {
			continue
		}
		result = append(result, menu)
	}
	return result
}

func (s *menuService) visibleMenuIDs(flat []user.Menu, allowed []uuid.UUID) map[uuid.UUID]struct{} {
	idToParent := make(map[uuid.UUID]*uuid.UUID)
	for i := range flat {
		idToParent[flat[i].ID] = flat[i].ParentID
	}
	visible := make(map[uuid.UUID]struct{})
	for _, id := range allowed {
		for cur := &id; cur != nil; {
			visible[*cur] = struct{}{}
			pid := idToParent[*cur]
			if pid == nil {
				break
			}
			cur = pid
		}
	}
	return visible
}

func (s *menuService) filterTreeByMenuIDs(nodes []*user.Menu, visible map[uuid.UUID]struct{}) []*user.Menu {
	var out []*user.Menu
	for _, n := range nodes {
		if _, ok := visible[n.ID]; !ok {
			continue
		}
		// 检查菜单是否启用
		if n.Meta != nil {
			if isEnable, ok := n.Meta["isEnable"].(bool); ok && !isEnable {
				continue
			}
		}
		clone := *n
		clone.Children = s.filterTreeByMenuIDs(n.Children, visible)
		out = append(out, &clone)
	}
	return out
}

func (s *menuService) Create(req *dto.MenuCreateRequest) (*user.Menu, error) {
	if req == nil {
		return nil, ErrMenuNotFound
	}
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &pid
	}
	manageGroupID, err := parseOptionalUUID(req.ManageGroupID)
	if err != nil {
		return nil, err
	}
	kind, component, meta := sanitizeMenuPayloadByKind(req.Kind, req.Component, req.Meta)
	appKey := normalizeMenuAppKey(req.AppKey)
	spaceKey := strings.TrimSpace(req.SpaceKey)
	menuKey := strings.TrimSpace(req.Name)
	if menuKey == "" {
		return nil, fmt.Errorf("menu_key is required")
	}
	menuID := uuid.New()
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var duplicateCount int64
		if err := tx.Model(&models.MenuDefinition{}).
			Where("app_key = ? AND menu_key = ? AND deleted_at IS NULL", appKey, menuKey).
			Count(&duplicateCount).Error; err != nil {
			return err
		}
		if duplicateCount > 0 {
			return fmt.Errorf("menu_key already exists")
		}
		definition := &models.MenuDefinition{
			ID:           menuID,
			AppKey:       appKey,
			MenuKey:      menuKey,
			Kind:         kind,
			Path:         req.Path,
			Name:         req.Name,
			Component:    component,
			DefaultTitle: req.Title,
			DefaultIcon:  req.Icon,
			Status:       "normal",
			Meta:         models.MetaJSON(meta),
		}
		if err := tx.Create(definition).Error; err != nil {
			return err
		}
		if strings.TrimSpace(spaceKey) == "" {
			return nil
		}
		parentMenuKey, err := s.resolveParentMenuKey(tx, appKey, spaceKey, parentID)
		if err != nil {
			return err
		}
		placement := &models.SpaceMenuPlacement{
			AppKey:        appKey,
			SpaceKey:      normalizeMenuSpaceKey(spaceKey),
			MenuKey:       menuKey,
			ParentMenuKey: parentMenuKey,
			ManageGroupID: manageGroupID,
			SortOrder:     req.SortOrder,
			Hidden:        req.Hidden,
		}
		return tx.Create(placement).Error
	}); err != nil {
		return nil, err
	}
	page.InvalidateRuntimeCache()
	invalidateMenuCaches()
	if err := s.refreshAllMenuSnapshots(); err != nil {
		return nil, err
	}
	return s.menuRepo.GetByID(menuID)
}

func (s *menuService) Update(id uuid.UUID, req *dto.MenuUpdateRequest) error {
	if req == nil {
		return ErrMenuNotFound
	}
	definition, err := s.loadMenuDefinitionByID(id, normalizeMenuAppKey(req.AppKey))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}
	if normalizeMenuAppKey(req.AppKey) != normalizeMenuAppKey(definition.AppKey) {
		return ErrMenuNotFound
	}
	manageGroupID, err := parseOptionalUUID(req.ManageGroupID)
	if err != nil {
		return err
	}
	kind, component, meta := sanitizeMenuPayloadByKind(req.Kind, req.Component, req.Meta)
	oldMenuKey := strings.TrimSpace(definition.MenuKey)
	nextMenuKey := oldMenuKey
	if strings.TrimSpace(req.Name) != "" {
		nextMenuKey = strings.TrimSpace(req.Name)
	}
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if nextMenuKey != oldMenuKey {
			var duplicateCount int64
			if err := tx.Model(&models.MenuDefinition{}).
				Where("app_key = ? AND menu_key = ? AND id <> ? AND deleted_at IS NULL", definition.AppKey, nextMenuKey, definition.ID).
				Count(&duplicateCount).Error; err != nil {
				return err
			}
			if duplicateCount > 0 {
				return fmt.Errorf("menu_key already exists")
			}
		}

		updates := map[string]interface{}{
			"menu_key":      nextMenuKey,
			"kind":          kind,
			"path":          req.Path,
			"name":          req.Name,
			"component":     component,
			"default_title": req.Title,
			"default_icon":  req.Icon,
		}
		if req.Meta != nil {
			updates["meta"] = models.MetaJSON(meta)
		}
		if err := tx.Model(&models.MenuDefinition{}).
			Where("id = ? AND app_key = ?", definition.ID, definition.AppKey).
			Updates(updates).Error; err != nil {
			return err
		}
		if nextMenuKey != oldMenuKey {
			if err := tx.Model(&models.SpaceMenuPlacement{}).
				Where("app_key = ? AND menu_key = ?", definition.AppKey, oldMenuKey).
				Update("menu_key", nextMenuKey).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.SpaceMenuPlacement{}).
				Where("app_key = ? AND parent_menu_key = ?", definition.AppKey, oldMenuKey).
				Update("parent_menu_key", nextMenuKey).Error; err != nil {
				return err
			}
		}
		if req.ManageGroupID != nil {
			if err := tx.Model(&models.SpaceMenuPlacement{}).
				Where("app_key = ? AND menu_key = ?", definition.AppKey, nextMenuKey).
				Update("manage_group_id", manageGroupID).Error; err != nil {
				return err
			}
		}
		if strings.TrimSpace(req.SpaceKey) == "" {
			return nil
		}

		parentID, err := parseOptionalUUID(req.ParentID)
		if err != nil {
			return err
		}
		if parentID != nil && *parentID == id {
			return errors.New("不能将上级设为自己")
		}
		layoutMenus, err := s.menuRepo.ListByAppAndSpace(definition.AppKey, normalizeMenuSpaceKey(req.SpaceKey))
		if err != nil {
			return err
		}
		if parentID != nil && s.isDescendant(layoutMenus, id, *parentID) {
			return errors.New("不能将上级设为自身子级（会造成循环）")
		}
		parentMenuKey, err := s.resolveParentMenuKey(tx, definition.AppKey, req.SpaceKey, parentID)
		if err != nil {
			return err
		}
		placement := &models.SpaceMenuPlacement{
			AppKey:        definition.AppKey,
			SpaceKey:      normalizeMenuSpaceKey(req.SpaceKey),
			MenuKey:       nextMenuKey,
			ParentMenuKey: parentMenuKey,
			ManageGroupID: manageGroupID,
			SortOrder:     req.SortOrder,
			Hidden:        req.Hidden,
		}
		return tx.Unscoped().Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "app_key"},
				{Name: "space_key"},
				{Name: "menu_key"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"parent_menu_key": parentMenuKey,
				"sort_order":      req.SortOrder,
				"hidden":          req.Hidden,
				"manage_group_id": manageGroupID,
				"updated_at":      time.Now(),
				"deleted_at":      nil,
			}),
		}).Create(placement).Error
	}); err != nil {
		return err
	}
	page.InvalidateRuntimeCache()
	invalidateMenuCaches()
	return s.refreshAllMenuSnapshots()
}

func (s *menuService) ListGroups() ([]user.MenuManageGroup, error) {
	var groups []user.MenuManageGroup
	err := s.db.Order("sort_order ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

func (s *menuService) CreateGroup(req *dto.MenuManageGroupCreateRequest) (*user.MenuManageGroup, error) {
	group := &user.MenuManageGroup{
		Name:      strings.TrimSpace(req.Name),
		SortOrder: req.SortOrder,
		Status:    normalizeMenuGroupStatus(req.Status),
	}
	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (s *menuService) UpdateGroup(id uuid.UUID, req *dto.MenuManageGroupUpdateRequest) error {
	var group user.MenuManageGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuGroupNotFound
		}
		return err
	}
	return s.db.Model(&group).Updates(map[string]interface{}{
		"name":       strings.TrimSpace(req.Name),
		"sort_order": req.SortOrder,
		"status":     normalizeMenuGroupStatus(req.Status),
	}).Error
}

func (s *menuService) DeleteGroup(id uuid.UUID) error {
	var group user.MenuManageGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuGroupNotFound
		}
		return err
	}

	var count int64
	if err := s.db.Model(&models.SpaceMenuPlacement{}).Where("manage_group_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrMenuGroupInUse
	}

	return s.db.Delete(&group).Error
}

func (s *menuService) isDescendant(flat []user.Menu, ancestorID, targetID uuid.UUID) bool {
	idToParent := make(map[uuid.UUID]*uuid.UUID)
	for i := range flat {
		idToParent[flat[i].ID] = flat[i].ParentID
	}
	cur := &targetID
	for {
		parentID := idToParent[*cur]
		if parentID == nil {
			return false
		}
		if *parentID == ancestorID {
			return true
		}
		cur = parentID
	}
}

func (s *menuService) Delete(id uuid.UUID, mode string, targetParentID *uuid.UUID) error {
	targetMenu, err := s.menuRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}

	deleteMode := normalizeMenuDeleteMode(mode)
	if deleteMode == "" {
		return ErrMenuDeleteModeInvalid
	}

	flatMenus, err := s.menuRepo.ListByAppAndSpace(targetMenu.AppKey, "")
	if err != nil {
		return err
	}

	directChildren := collectDirectChildMenus(flatMenus, id)
	if len(directChildren) > 0 && deleteMode == menuDeleteModeSingle {
		return ErrMenuHasChildren
	}
	if deleteMode == menuDeleteModePromoteChildren && targetParentID != nil {
		if err := validateMenuPromoteTarget(flatMenus, id, *targetParentID); err != nil {
			return err
		}
	}

	affectedMenus := collectMenuSubtree(flatMenus, id)
	affectedIDs := make([]uuid.UUID, 0, len(affectedMenus))
	for _, item := range affectedMenus {
		affectedIDs = append(affectedIDs, item.ID)
	}
	menuKeyMap, err := s.loadMenuKeyMapByIDs(targetMenu.AppKey, affectedIDs)
	if err != nil {
		return err
	}
	affectedMenuKeys := make([]string, 0, len(menuKeyMap))
	for _, menuID := range affectedIDs {
		if menuKey := strings.TrimSpace(menuKeyMap[menuID]); menuKey != "" {
			affectedMenuKeys = append(affectedMenuKeys, menuKey)
		}
	}
	targetParentMenuKey := ""
	if targetParentID != nil {
		targetParent, parentErr := s.loadMenuDefinitionByID(*targetParentID, targetMenu.AppKey)
		if parentErr != nil {
			if errors.Is(parentErr, gorm.ErrRecordNotFound) {
				return ErrMenuNotFound
			}
			return parentErr
		}
		targetParentMenuKey = strings.TrimSpace(targetParent.MenuKey)
	}

	oldPathMap := buildMenuFullPathMap(flatMenus)
	oldAffectedPaths := collectMenuPathSet(affectedIDs, oldPathMap)
	newPathMap := map[uuid.UUID]string{}
	if deleteMode == menuDeleteModePromoteChildren {
		newPathMap = buildPromotedMenuFullPathMap(flatMenus, id, targetParentID)
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		switch deleteMode {
		case menuDeleteModeCascade:
			if err := s.cleanupMenuRelationsByIDs(tx, affectedIDs); err != nil {
				return err
			}
			if err := tx.Model(&models.UIPage{}).
				Where("parent_menu_id IN ?", affectedIDs).
				Update("parent_menu_id", nil).Error; err != nil {
				return err
			}
			if err := clearPageActiveMenuPaths(tx, oldAffectedPaths); err != nil {
				return err
			}
			if len(affectedMenuKeys) > 0 {
				if err := tx.Where("app_key = ? AND menu_key IN ?", targetMenu.AppKey, affectedMenuKeys).
					Delete(&models.SpaceMenuPlacement{}).Error; err != nil {
					return err
				}
			}
			if err := tx.Where("app_key = ? AND id IN ?", targetMenu.AppKey, affectedIDs).
				Delete(&models.MenuDefinition{}).Error; err != nil {
				return err
			}
		case menuDeleteModePromoteChildren:
			if err := s.cleanupMenuRelationsByIDs(tx, []uuid.UUID{id}); err != nil {
				return err
			}
			if err := tx.Model(&models.UIPage{}).
				Where("parent_menu_id = ?", id).
				Update("parent_menu_id", nil).Error; err != nil {
				return err
			}
			if targetPath := normalizeManagedMenuPath(oldPathMap[id]); targetPath != "" {
				if err := clearPageActiveMenuPaths(tx, []string{targetPath}); err != nil {
					return err
				}
			}
			if err := remapPageActiveMenuPaths(tx, oldPathMap, newPathMap); err != nil {
				return err
			}
			if err := tx.Model(&models.SpaceMenuPlacement{}).
				Where("app_key = ? AND parent_menu_key = ?", targetMenu.AppKey, menuKeyMap[id]).
				Update("parent_menu_key", targetParentMenuKey).Error; err != nil {
				return err
			}
			if err := tx.Where("app_key = ? AND menu_key = ?", targetMenu.AppKey, menuKeyMap[id]).
				Delete(&models.SpaceMenuPlacement{}).Error; err != nil {
				return err
			}
			if err := tx.Where("app_key = ? AND id = ?", targetMenu.AppKey, id).
				Delete(&models.MenuDefinition{}).Error; err != nil {
				return err
			}
		default:
			if err := s.cleanupMenuRelationsByIDs(tx, []uuid.UUID{id}); err != nil {
				return err
			}
			if err := tx.Model(&models.UIPage{}).
				Where("parent_menu_id = ?", id).
				Update("parent_menu_id", nil).Error; err != nil {
				return err
			}
			if err := clearPageActiveMenuPaths(tx, oldAffectedPaths); err != nil {
				return err
			}
			if err := tx.Where("app_key = ? AND menu_key = ?", targetMenu.AppKey, menuKeyMap[id]).
				Delete(&models.SpaceMenuPlacement{}).Error; err != nil {
				return err
			}
			if err := tx.Where("app_key = ? AND id = ?", targetMenu.AppKey, id).
				Delete(&models.MenuDefinition{}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	s.logger.Info(
		"Menu deleted",
		zap.String("menu_id", id.String()),
		zap.String("menu_title", strings.TrimSpace(targetMenu.Title)),
		zap.String("delete_mode", deleteMode),
		zap.Int("affected_menu_count", len(affectedIDs)),
	)
	page.InvalidateRuntimeCache()
	invalidateMenuCaches()
	return s.refreshAllMenuSnapshots()
}

func (s *menuService) DeletePreview(id uuid.UUID, mode string, targetParentID *uuid.UUID) (*MenuDeletePreview, error) {
	targetMenu, err := s.menuRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuNotFound
		}
		return nil, err
	}

	deleteMode := normalizeMenuDeleteMode(mode)
	if deleteMode == "" {
		return nil, ErrMenuDeleteModeInvalid
	}

	flatMenus, err := s.menuRepo.ListByAppAndSpace(targetMenu.AppKey, "")
	if err != nil {
		return nil, err
	}

	affectedMenus := []user.Menu{}
	childCount := len(collectDirectChildMenus(flatMenus, id))
	switch deleteMode {
	case menuDeleteModeCascade:
		affectedMenus = collectMenuSubtree(flatMenus, id)
	case menuDeleteModePromoteChildren, menuDeleteModeSingle:
		affectedMenus = []user.Menu{{ID: id}}
	}

	affectedIDs := make([]uuid.UUID, 0, len(affectedMenus))
	for _, item := range affectedMenus {
		affectedIDs = append(affectedIDs, item.ID)
	}
	pageCount, err := s.countAffectedPages(flatMenus, affectedIDs, deleteMode, id)
	if err != nil {
		return nil, err
	}
	relationCount, err := s.countAffectedMenuRelations(affectedIDs)
	if err != nil {
		return nil, err
	}

	return &MenuDeletePreview{
		Mode:                  deleteMode,
		MenuCount:             len(affectedIDs),
		ChildCount:            childCount,
		AffectedPageCount:     pageCount,
		AffectedRelationCount: relationCount,
	}, nil
}

func (s *menuService) countAffectedPages(flatMenus []user.Menu, affectedIDs []uuid.UUID, mode string, rootID uuid.UUID) (int, error) {
	if len(affectedIDs) == 0 {
		return 0, nil
	}
	var pages []models.UIPage
	if err := s.db.Select("id", "parent_menu_id", "active_menu_path").Find(&pages).Error; err != nil {
		return 0, err
	}

	pathMap := buildMenuFullPathMap(flatMenus)
	targetPaths := collectMenuPathSet(affectedIDs, pathMap)
	count := 0
	seen := make(map[uuid.UUID]struct{})
	for _, pageItem := range pages {
		if _, ok := seen[pageItem.ID]; ok {
			continue
		}
		if pageItem.ParentMenuID != nil {
			for _, menuID := range affectedIDs {
				if *pageItem.ParentMenuID == menuID {
					seen[pageItem.ID] = struct{}{}
					count++
					break
				}
			}
			continue
		}
		activePath := normalizeManagedMenuPath(pageItem.ActiveMenuPath)
		if activePath == "" {
			continue
		}
		matched := false
		for _, path := range targetPaths {
			if pathHasPrefix(activePath, path) {
				matched = true
				break
			}
		}
		if matched {
			seen[pageItem.ID] = struct{}{}
			count++
		}
	}
	return count, nil
}

func (s *menuService) countAffectedMenuRelations(menuIDs []uuid.UUID) (int, error) {
	if len(menuIDs) == 0 {
		return 0, nil
	}
	var count int64
	tables := []string{
		"feature_package_menus",
		"role_hidden_menus",
		"team_blocked_menus",
		"user_hidden_menus",
	}
	for _, table := range tables {
		var current int64
		if err := s.db.Table(table).Where("menu_id IN ?", menuIDs).Count(&current).Error; err != nil {
			return 0, err
		}
		count += current
	}
	return int(count), nil
}

// 菜单备份相关方法
func (s *menuService) CreateBackup(name, description, scopeType, appKey, spaceKey string, createdBy *uuid.UUID) error {
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return errors.New("app_key is required")
	}
	appKey = normalizeMenuAppKey(appKey)
	normalizedScopeType := normalizeBackupScopeType(scopeType, spaceKey)
	backupSpaceKey := resolveBackupSpaceKey(normalizedScopeType, spaceKey)

	definitions, err := s.loadMenuDefinitions(appKey)
	if err != nil {
		s.logger.Error("Failed to list menu definitions for backup", zap.Error(err))
		return err
	}
	placements, err := s.loadMenuPlacements(appKey, backupSpaceKey)
	if err != nil {
		s.logger.Error("Failed to list menu placements for backup", zap.Error(err))
		return err
	}
	if backupSpaceKey != "" {
		definitions = filterMenuDefinitionsByPlacements(definitions, placements)
	}

	groups, err := s.loadBackupGroupsByPlacements(placements)
	if err != nil {
		s.logger.Error("Failed to list menu groups for backup", zap.Error(err))
		return err
	}

	menuData, err := json.Marshal(ginMenuBackupPayload{
		Version:     menuBackupPayloadVersion,
		AppKey:      appKey,
		ScopeType:   normalizedScopeType,
		SpaceKey:    backupSpaceKey,
		Groups:      groups,
		Definitions: definitions,
		Placements:  placements,
	})
	if err != nil {
		s.logger.Error("Failed to marshal menu data", zap.Error(err))
		return err
	}

	backup := &user.MenuBackup{
		Name:        name,
		Description: description,
		AppKey:      appKey,
		SpaceKey:    backupSpaceKey,
		MenuData:    string(menuData),
		CreatedBy:   createdBy,
	}

	if err := s.menuRepo.CreateBackup(backup); err != nil {
		s.logger.Error("Failed to create menu backup", zap.Error(err))
		return err
	}

	s.logger.Info("Menu backup created", zap.String("name", name))
	return nil
}

func (s *menuService) ListBackups(appKey, spaceKey string) ([]*user.MenuBackup, error) {
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return nil, errors.New("app_key is required")
	}
	backups, err := s.menuRepo.ListBackups()
	if err != nil {
		s.logger.Error("Failed to list menu backups", zap.Error(err))
		return nil, err
	}

	backups = filterBackupsByApp(backups, appKey)
	backups = filterBackupsBySpace(backups, spaceKey)

	// 转换为指针切片
	backupPtrs := make([]*user.MenuBackup, len(backups))
	for i, backup := range backups {
		backupCopy := backup
		backupPtrs[i] = &backupCopy
	}

	return backupPtrs, nil
}

func (s *menuService) DeleteBackup(id uuid.UUID, appKey string) error {
	// 检查备份是否存在
	backup, err := s.menuRepo.GetBackupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("备份不存在")
		}
		return err
	}
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return errors.New("app_key is required")
	}
	if normalizeMenuAppKey(backup.AppKey) != normalizeMenuAppKey(appKey) {
		return ErrMenuBackupAppMismatch
	}

	if err := s.menuRepo.DeleteBackup(id); err != nil {
		s.logger.Error("Failed to delete menu backup", zap.Error(err))
		return err
	}

	s.logger.Info("Menu backup deleted", zap.String("backup_id", id.String()))
	return nil
}

func (s *menuService) RestoreBackup(id uuid.UUID, appKey string) error {
	// 检查备份是否存在
	backup, err := s.menuRepo.GetBackupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("备份不存在")
		}
		return err
	}
	appKey = strings.TrimSpace(appKey)
	if appKey == "" {
		return errors.New("app_key is required")
	}
	if normalizeMenuAppKey(backup.AppKey) != normalizeMenuAppKey(appKey) {
		return ErrMenuBackupAppMismatch
	}

	// 解析备份的菜单数据
	payload, err := parseMenuBackupPayload(backup.MenuData)
	if err != nil {
		s.logger.Error("Failed to unmarshal menu data", zap.Error(err))
		return err
	}
	groups := payload.Groups
	definitions := payload.Definitions
	placements := payload.Placements
	if len(definitions) == 0 && len(placements) == 0 && len(payload.Menus) > 0 {
		definitions, placements = convertLegacyMenusToDefinitions(payload.Menus)
	}

	backupSpaceKey := normalizeBackupSpaceKey(backup.SpaceKey)
	if backupSpaceKey == "" {
		backupSpaceKey = normalizeBackupSpaceKey(payload.SpaceKey)
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if backupSpaceKey == "" {
			return s.restoreGlobalMenuBackup(tx, backup.AppKey, groups, definitions, placements)
		}
		spacePlacements := filterPlacementsBySpace(placements, backupSpaceKey)
		return s.restoreSpaceMenuBackup(tx, backup.AppKey, backupSpaceKey, groups, definitions, spacePlacements)
	}); err != nil {
		s.logger.Error("Failed to restore menu backup", zap.Error(err))
		return err
	}
	page.InvalidateRuntimeCache()
	invalidateMenuCaches()

	// 清理无效的角色菜单关联
	if err := s.cleanupInvalidMenuRelations(); err != nil {
		return err
	}
	if err := s.refreshAllMenuSnapshots(); err != nil {
		return err
	}

	s.logger.Info("Menu backup restored", zap.String("backup_id", id.String()))
	return nil
}

func (s *menuService) cleanupInvalidMenuRelations() error {
	if s.db == nil {
		return nil
	}
	statements := []string{
		"DELETE FROM feature_package_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
		"DELETE FROM role_hidden_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
		"DELETE FROM team_blocked_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
		"DELETE FROM user_hidden_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
		"UPDATE ui_pages SET parent_menu_id = NULL WHERE parent_menu_id IS NOT NULL AND parent_menu_id NOT IN (SELECT id FROM menu_definitions)",
	}
	for _, statement := range statements {
		if err := s.db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *menuService) refreshAllMenuSnapshots() error {
	if s.refresher == nil || s.db == nil {
		return nil
	}
	if err := s.refresher.RefreshAllTeams(); err != nil {
		return err
	}
	if err := s.refresher.RefreshAllPlatformRoles(); err != nil {
		return err
	}
	return s.refresher.RefreshAllPlatformUsers()
}

type ginMenuBackupPayload struct {
	Version     string                      `json:"version,omitempty"`
	AppKey      string                      `json:"app_key,omitempty"`
	ScopeType   string                      `json:"scope_type,omitempty"`
	SpaceKey    string                      `json:"space_key,omitempty"`
	Groups      []user.MenuManageGroup      `json:"groups"`
	Definitions []models.MenuDefinition     `json:"definitions,omitempty"`
	Placements  []models.SpaceMenuPlacement `json:"placements,omitempty"`
	Menus       []user.Menu                 `json:"menus,omitempty"`
}

type menuBackupScopeInfo struct {
	ScopeType   string
	ScopeOrigin string
	AppKey      string
	SpaceKey    string
}

const menuBackupPayloadVersion = "menu_backup.v2"

const (
	menuDeleteModeSingle          = "single"
	menuDeleteModeCascade         = "cascade"
	menuDeleteModePromoteChildren = "promote_children"
)

func normalizeBackupSpaceKey(value string) string {
	target := strings.TrimSpace(value)
	if target == "" {
		return ""
	}
	return normalizeMenuSpaceKey(target)
}

func normalizeBackupScopeType(scopeType string, spaceKey string) string {
	switch strings.TrimSpace(strings.ToLower(scopeType)) {
	case "global":
		return "global"
	case "space":
		return "space"
	default:
		if normalizeBackupSpaceKey(spaceKey) != "" {
			return "space"
		}
		return "global"
	}
}

func resolveBackupSpaceKey(scopeType string, spaceKey string) string {
	if normalizeBackupScopeType(scopeType, spaceKey) == "global" {
		return ""
	}
	if normalized := normalizeBackupSpaceKey(spaceKey); normalized != "" {
		return normalized
	}
	return spaceutil.DefaultMenuSpaceKey
}

func resolveMenuBackupScopeInfo(backup user.MenuBackup) menuBackupScopeInfo {
	scopeOrigin := "menu_backup"
	if backupSpaceKey := normalizeBackupSpaceKey(backup.SpaceKey); backupSpaceKey != "" {
		return menuBackupScopeInfo{
			ScopeType:   "space",
			ScopeOrigin: scopeOrigin,
			AppKey:      normalizeMenuAppKey(backup.AppKey),
			SpaceKey:    backupSpaceKey,
		}
	}

	payload, err := parseMenuBackupPayload(backup.MenuData)
	if err != nil {
		return menuBackupScopeInfo{
			ScopeType:   "global",
			ScopeOrigin: scopeOrigin,
			AppKey:      normalizeMenuAppKey(backup.AppKey),
		}
	}

	resolvedScopeType := normalizeBackupScopeType(payload.ScopeType, payload.SpaceKey)
	if resolvedScopeType == "space" {
		return menuBackupScopeInfo{
			ScopeType:   "space",
			ScopeOrigin: scopeOrigin,
			AppKey:      normalizeMenuAppKey(backup.AppKey),
			SpaceKey:    resolveBackupSpaceKey(resolvedScopeType, payload.SpaceKey),
		}
	}

	return menuBackupScopeInfo{
		ScopeType:   "global",
		ScopeOrigin: scopeOrigin,
		AppKey:      normalizeMenuAppKey(backup.AppKey),
	}
}

func filterBackupsBySpace(backups []user.MenuBackup, spaceKey string) []user.MenuBackup {
	targetSpaceKey := normalizeBackupSpaceKey(spaceKey)
	if targetSpaceKey == "" {
		return backups
	}
	filtered := make([]user.MenuBackup, 0, len(backups))
	for _, backup := range backups {
		scopeInfo := resolveMenuBackupScopeInfo(backup)
		if scopeInfo.ScopeType == "global" || scopeInfo.SpaceKey == targetSpaceKey {
			filtered = append(filtered, backup)
		}
	}
	return filtered
}

func filterBackupsByApp(backups []user.MenuBackup, appKey string) []user.MenuBackup {
	targetAppKey := normalizeMenuAppKey(appKey)
	if targetAppKey == "" {
		return backups
	}
	filtered := make([]user.MenuBackup, 0, len(backups))
	for _, backup := range backups {
		if normalizeMenuAppKey(backup.AppKey) == targetAppKey {
			filtered = append(filtered, backup)
		}
	}
	return filtered
}

func parseMenuBackupPayload(raw string) (ginMenuBackupPayload, error) {
	var payload ginMenuBackupPayload
	if err := json.Unmarshal([]byte(raw), &payload); err == nil {
		if payload.Version != "" || payload.SpaceKey != "" || payload.Groups != nil || payload.Menus != nil || payload.Definitions != nil || payload.Placements != nil {
			payload.ScopeType = normalizeBackupScopeType(payload.ScopeType, payload.SpaceKey)
			return payload, nil
		}
	}
	return ginMenuBackupPayload{}, fmt.Errorf("invalid menu backup payload")
}

func (s *menuService) loadBackupGroupsByPlacements(placements []models.SpaceMenuPlacement) ([]user.MenuManageGroup, error) {
	if s.db == nil {
		return nil, nil
	}

	var groups []user.MenuManageGroup
	groupIDs := collectPlacementManageGroupIDs(placements)
	if len(groupIDs) == 0 {
		return []user.MenuManageGroup{}, nil
	}
	if err := s.db.Order("sort_order ASC, created_at ASC").Where("id IN ?", groupIDs).Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func collectPlacementManageGroupIDs(placements []models.SpaceMenuPlacement) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	ids := make([]uuid.UUID, 0)
	for _, placement := range placements {
		if placement.ManageGroupID == nil || *placement.ManageGroupID == uuid.Nil {
			continue
		}
		if _, exists := seen[*placement.ManageGroupID]; exists {
			continue
		}
		seen[*placement.ManageGroupID] = struct{}{}
		ids = append(ids, *placement.ManageGroupID)
	}
	return ids
}

func (s *menuService) restoreGlobalMenuBackup(
	tx *gorm.DB,
	appKey string,
	groups []user.MenuManageGroup,
	definitions []models.MenuDefinition,
	placements []models.SpaceMenuPlacement,
) error {
	if err := tx.Where("app_key = ?", normalizeMenuAppKey(appKey)).Delete(&models.SpaceMenuPlacement{}).Error; err != nil {
		return err
	}
	if err := tx.Where("app_key = ?", normalizeMenuAppKey(appKey)).Delete(&models.MenuDefinition{}).Error; err != nil {
		return err
	}
	if err := s.upsertMenuManageGroups(tx, groups); err != nil {
		return err
	}
	for i := range definitions {
		definitions[i].AppKey = normalizeMenuAppKey(definitions[i].AppKey)
		definitions[i].Kind = normalizeMenuKind(definitions[i].Kind, definitions[i].Component, definitions[i].Meta)
		definitions[i].Meta = sanitizeMenuMeta(definitions[i].Meta)
		if err := tx.Create(&definitions[i]).Error; err != nil {
			return err
		}
	}
	for i := range placements {
		placements[i].AppKey = normalizeMenuAppKey(placements[i].AppKey)
		placements[i].SpaceKey = normalizeMenuSpaceKey(placements[i].SpaceKey)
		if err := tx.Create(&placements[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *menuService) restoreSpaceMenuBackup(
	tx *gorm.DB,
	appKey string,
	spaceKey string,
	groups []user.MenuManageGroup,
	definitions []models.MenuDefinition,
	placements []models.SpaceMenuPlacement,
) error {
	if err := s.upsertMenuManageGroups(tx, groups); err != nil {
		return err
	}
	for i := range definitions {
		definitions[i].AppKey = normalizeMenuAppKey(definitions[i].AppKey)
		definitions[i].Kind = normalizeMenuKind(definitions[i].Kind, definitions[i].Component, definitions[i].Meta)
		definitions[i].Meta = sanitizeMenuMeta(definitions[i].Meta)
		if err := tx.Unscoped().Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"app_key":       definitions[i].AppKey,
				"menu_key":      definitions[i].MenuKey,
				"kind":          definitions[i].Kind,
				"path":          definitions[i].Path,
				"name":          definitions[i].Name,
				"component":     definitions[i].Component,
				"default_title": definitions[i].DefaultTitle,
				"default_icon":  definitions[i].DefaultIcon,
				"status":        definitions[i].Status,
				"meta":          definitions[i].Meta,
				"updated_at":    time.Now(),
				"deleted_at":    nil,
			}),
		}).Create(&definitions[i]).Error; err != nil {
			return err
		}
	}
	if err := s.deleteMenusBySpace(tx, appKey, spaceKey); err != nil {
		return err
	}
	for i := range placements {
		placements[i].AppKey = normalizeMenuAppKey(appKey)
		placements[i].SpaceKey = normalizeMenuSpaceKey(spaceKey)
		if err := tx.Create(&placements[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *menuService) upsertMenuManageGroups(tx *gorm.DB, groups []user.MenuManageGroup) error {
	if len(groups) == 0 {
		return nil
	}
	now := time.Now()
	for i := range groups {
		groups[i].Status = normalizeMenuGroupStatus(groups[i].Status)
		if err := tx.Unscoped().Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":       groups[i].Name,
				"sort_order": groups[i].SortOrder,
				"status":     groups[i].Status,
				"updated_at": now,
				"deleted_at": nil,
			}),
		}).Create(&groups[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *menuService) deleteMenusBySpace(tx *gorm.DB, appKey, spaceKey string) error {
	normalizedSpaceKey := normalizeMenuSpaceKey(spaceKey)
	return tx.Where("app_key = ? AND space_key = ?", normalizeMenuAppKey(appKey), normalizedSpaceKey).
		Delete(&models.SpaceMenuPlacement{}).Error
}

func (s *menuService) loadMenuDefinitions(appKey string) ([]models.MenuDefinition, error) {
	var definitions []models.MenuDefinition
	if err := s.db.Where("app_key = ?", normalizeMenuAppKey(appKey)).
		Order("created_at ASC").
		Find(&definitions).Error; err != nil {
		return nil, err
	}
	return definitions, nil
}

func (s *menuService) loadMenuPlacements(appKey, spaceKey string) ([]models.SpaceMenuPlacement, error) {
	query := s.db.Where("app_key = ?", normalizeMenuAppKey(appKey)).Order("sort_order ASC, created_at ASC")
	if strings.TrimSpace(spaceKey) != "" {
		query = query.Where("space_key = ?", normalizeMenuSpaceKey(spaceKey))
	}
	var placements []models.SpaceMenuPlacement
	if err := query.Find(&placements).Error; err != nil {
		return nil, err
	}
	return placements, nil
}

func (s *menuService) loadMenuDefinitionByID(id uuid.UUID, appKey string) (*models.MenuDefinition, error) {
	var definition models.MenuDefinition
	if err := s.db.Where("id = ? AND app_key = ?", id, normalizeMenuAppKey(appKey)).First(&definition).Error; err != nil {
		return nil, err
	}
	return &definition, nil
}

func (s *menuService) loadMenuKeyMapByIDs(appKey string, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	var definitions []models.MenuDefinition
	if err := s.db.Select("id", "menu_key").
		Where("app_key = ? AND id IN ?", normalizeMenuAppKey(appKey), ids).
		Find(&definitions).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]string, len(definitions))
	for _, definition := range definitions {
		result[definition.ID] = strings.TrimSpace(definition.MenuKey)
	}
	return result, nil
}

func (s *menuService) resolveParentMenuKey(tx *gorm.DB, appKey, spaceKey string, parentID *uuid.UUID) (string, error) {
	if parentID == nil || *parentID == uuid.Nil {
		return "", nil
	}
	db := s.db
	if tx != nil {
		db = tx
	}
	var parent models.MenuDefinition
	if err := db.Where("id = ? AND app_key = ?", *parentID, normalizeMenuAppKey(appKey)).First(&parent).Error; err != nil {
		return "", err
	}
	if strings.TrimSpace(spaceKey) != "" {
		var placementCount int64
		if err := db.Model(&models.SpaceMenuPlacement{}).
			Where("app_key = ? AND space_key = ? AND menu_key = ?", normalizeMenuAppKey(appKey), normalizeMenuSpaceKey(spaceKey), parent.MenuKey).
			Count(&placementCount).Error; err != nil {
			return "", err
		}
		if placementCount == 0 {
			return "", fmt.Errorf("上级菜单未放入当前空间")
		}
	}
	return strings.TrimSpace(parent.MenuKey), nil
}

func filterMenuDefinitionsByPlacements(
	definitions []models.MenuDefinition,
	placements []models.SpaceMenuPlacement,
) []models.MenuDefinition {
	if len(definitions) == 0 || len(placements) == 0 {
		return []models.MenuDefinition{}
	}
	allowed := make(map[string]struct{}, len(placements))
	for _, placement := range placements {
		allowed[normalizeMenuAppKey(placement.AppKey)+"::"+strings.TrimSpace(placement.MenuKey)] = struct{}{}
	}
	result := make([]models.MenuDefinition, 0, len(definitions))
	for _, definition := range definitions {
		if _, ok := allowed[normalizeMenuAppKey(definition.AppKey)+"::"+strings.TrimSpace(definition.MenuKey)]; ok {
			result = append(result, definition)
		}
	}
	return result
}

func filterPlacementsBySpace(
	placements []models.SpaceMenuPlacement,
	spaceKey string,
) []models.SpaceMenuPlacement {
	target := normalizeMenuSpaceKey(spaceKey)
	result := make([]models.SpaceMenuPlacement, 0, len(placements))
	for _, placement := range placements {
		if normalizeMenuSpaceKey(placement.SpaceKey) != target {
			continue
		}
		result = append(result, placement)
	}
	return result
}

func convertLegacyMenusToDefinitions(
	menus []user.Menu,
) ([]models.MenuDefinition, []models.SpaceMenuPlacement) {
	if len(menus) == 0 {
		return []models.MenuDefinition{}, []models.SpaceMenuPlacement{}
	}

	keyByLegacyID := make(map[uuid.UUID]string, len(menus))
	usedKeys := make(map[string]int)
	for _, item := range menus {
		appKey := normalizeMenuAppKey(item.AppKey)
		baseKey := strings.TrimSpace(item.Name)
		if baseKey == "" {
			baseKey = "legacy-" + strings.ToLower(strings.ReplaceAll(item.ID.String(), "-", ""))
		}
		baseKey = strings.ToLower(strings.TrimSpace(baseKey))
		if baseKey == "" {
			baseKey = "legacy-" + strings.ToLower(strings.ReplaceAll(item.ID.String(), "-", ""))
		}
		composite := appKey + "::" + baseKey
		if seen, ok := usedKeys[composite]; ok {
			seen++
			usedKeys[composite] = seen
			baseKey = fmt.Sprintf("%s-%d", baseKey, seen)
		} else {
			usedKeys[composite] = 1
		}
		keyByLegacyID[item.ID] = baseKey
	}

	definitions := make([]models.MenuDefinition, 0, len(menus))
	seenDefinitions := make(map[string]struct{}, len(menus))
	for _, item := range menus {
		appKey := normalizeMenuAppKey(item.AppKey)
		menuKey := keyByLegacyID[item.ID]
		composite := appKey + "::" + menuKey
		if _, ok := seenDefinitions[composite]; ok {
			continue
		}
		seenDefinitions[composite] = struct{}{}
		definitions = append(definitions, models.MenuDefinition{
			ID:           item.ID,
			AppKey:       appKey,
			MenuKey:      menuKey,
			Kind:         normalizeMenuKind(item.Kind, item.Component, item.Meta),
			Path:         item.Path,
			Name:         item.Name,
			Component:    item.Component,
			DefaultTitle: item.Title,
			DefaultIcon:  item.Icon,
			Status:       "normal",
			Meta:         models.MetaJSON(sanitizeMenuMeta(item.Meta)),
		})
	}

	placements := make([]models.SpaceMenuPlacement, 0, len(menus))
	for _, item := range menus {
		parentMenuKey := ""
		if item.ParentID != nil {
			parentMenuKey = keyByLegacyID[*item.ParentID]
		}
		placements = append(placements, models.SpaceMenuPlacement{
			AppKey:        normalizeMenuAppKey(item.AppKey),
			SpaceKey:      normalizeMenuSpaceKey(item.SpaceKey),
			MenuKey:       keyByLegacyID[item.ID],
			ParentMenuKey: parentMenuKey,
			ManageGroupID: item.ManageGroupID,
			SortOrder:     item.SortOrder,
			Hidden:        item.Hidden,
		})
	}
	return definitions, placements
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func sanitizeMenuMeta(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return nil
	}
	sanitized := make(map[string]interface{}, len(meta))
	for key, value := range meta {
		if key == "manageGroup" {
			continue
		}
		sanitized[key] = value
	}
	return sanitized
}

func sanitizeMenuPayloadByKind(kind string, component string, meta map[string]interface{}) (string, string, map[string]interface{}) {
	sanitizedMeta := sanitizeMenuMeta(meta)
	resolvedKind := normalizeMenuKind(kind, component, sanitizedMeta)
	sanitizedComponent := strings.TrimSpace(component)

	switch resolvedKind {
	case models.MenuKindDirectory:
		sanitizedComponent = ""
		if sanitizedMeta != nil {
			delete(sanitizedMeta, "link")
			delete(sanitizedMeta, "isIframe")
		}
	case models.MenuKindExternal:
		sanitizedComponent = ""
		if sanitizedMeta != nil {
			delete(sanitizedMeta, "isIframe")
		}
	default:
		if sanitizedMeta != nil {
			delete(sanitizedMeta, "link")
		}
	}

	return resolvedKind, sanitizedComponent, sanitizedMeta
}

func normalizeMenuKind(kind string, component string, meta map[string]interface{}) string {
	target := strings.TrimSpace(strings.ToLower(kind))
	switch target {
	case models.MenuKindDirectory, models.MenuKindEntry, models.MenuKindExternal:
		return target
	}
	if meta != nil {
		if link := strings.TrimSpace(toStringValue(meta["link"])); link != "" {
			return models.MenuKindExternal
		}
	}
	if strings.TrimSpace(component) != "" && strings.TrimSpace(component) != "/index/index" {
		return models.MenuKindEntry
	}
	return models.MenuKindDirectory
}

func normalizeMenuListKinds(items []user.Menu) {
	for i := range items {
		items[i].Kind = normalizeMenuKind(items[i].Kind, items[i].Component, items[i].Meta)
	}
}

func normalizeMenuSpaceKey(value string) string {
	target := strings.ToLower(strings.TrimSpace(value))
	if target == "" {
		return spaceutil.DefaultMenuSpaceKey
	}
	return target
}

func normalizeMenuAppKey(value string) string {
	return apppkg.NormalizeAppKey(value)
}

func normalizeMenuGroupStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "disabled":
		return "disabled"
	default:
		return "normal"
	}
}

func invalidateMenuCaches() {
	platformaccess.InvalidatePublicMenuCache()
	platformroleaccess.InvalidatePublicMenuCache()
}

func normalizeMenuDeleteMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", menuDeleteModeSingle:
		return menuDeleteModeSingle
	case menuDeleteModeCascade:
		return menuDeleteModeCascade
	case menuDeleteModePromoteChildren:
		return menuDeleteModePromoteChildren
	default:
		return ""
	}
}

func collectDirectChildMenus(flat []user.Menu, parentID uuid.UUID) []user.Menu {
	items := make([]user.Menu, 0)
	for _, item := range flat {
		if item.ParentID != nil && *item.ParentID == parentID {
			items = append(items, item)
		}
	}
	return items
}

func collectMenuSubtree(flat []user.Menu, rootID uuid.UUID) []user.Menu {
	childrenMap := make(map[uuid.UUID][]user.Menu)
	menuMap := make(map[uuid.UUID]user.Menu, len(flat))
	for _, item := range flat {
		menuMap[item.ID] = item
		if item.ParentID != nil {
			childrenMap[*item.ParentID] = append(childrenMap[*item.ParentID], item)
		}
	}

	root, ok := menuMap[rootID]
	if !ok {
		return []user.Menu{}
	}

	result := make([]user.Menu, 0, 8)
	var walk func(user.Menu)
	walk = func(item user.Menu) {
		result = append(result, item)
		for _, child := range childrenMap[item.ID] {
			walk(child)
		}
	}
	walk(root)
	return result
}

func (s *menuService) cleanupMenuRelationsByIDs(tx *gorm.DB, menuIDs []uuid.UUID) error {
	if tx == nil || len(menuIDs) == 0 {
		return nil
	}
	statements := []string{
		"DELETE FROM feature_package_menus WHERE menu_id IN ?",
		"DELETE FROM role_hidden_menus WHERE menu_id IN ?",
		"DELETE FROM team_blocked_menus WHERE menu_id IN ?",
		"DELETE FROM user_hidden_menus WHERE menu_id IN ?",
	}
	for _, statement := range statements {
		if err := tx.Exec(statement, menuIDs).Error; err != nil {
			return err
		}
	}
	return nil
}

func collectMenuPathSet(menuIDs []uuid.UUID, pathMap map[uuid.UUID]string) []string {
	paths := make([]string, 0, len(menuIDs))
	seen := make(map[string]struct{}, len(menuIDs))
	for _, menuID := range menuIDs {
		path := normalizeManagedMenuPath(pathMap[menuID])
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		paths = append(paths, path)
	}
	return paths
}

func clearPageActiveMenuPaths(tx *gorm.DB, deletedPaths []string) error {
	if tx == nil || len(deletedPaths) == 0 {
		return nil
	}

	var pages []models.UIPage
	if err := tx.Select("id", "active_menu_path").
		Where("active_menu_path <> ''").
		Find(&pages).Error; err != nil {
		return err
	}
	for _, pageItem := range pages {
		activePath := normalizeManagedMenuPath(pageItem.ActiveMenuPath)
		if activePath == "" || !matchesDeletedMenuPath(activePath, deletedPaths) {
			continue
		}
		if err := tx.Model(&models.UIPage{}).
			Where("id = ?", pageItem.ID).
			Update("active_menu_path", "").Error; err != nil {
			return err
		}
	}
	return nil
}

func remapPageActiveMenuPaths(tx *gorm.DB, oldPathMap map[uuid.UUID]string, newPathMap map[uuid.UUID]string) error {
	if tx == nil || len(oldPathMap) == 0 || len(newPathMap) == 0 {
		return nil
	}

	replacements := make(map[string]string)
	for menuID, oldPath := range oldPathMap {
		newPath := normalizeManagedMenuPath(newPathMap[menuID])
		oldNormalized := normalizeManagedMenuPath(oldPath)
		if oldNormalized == "" || newPath == "" || oldNormalized == newPath {
			continue
		}
		replacements[oldNormalized] = newPath
	}
	if len(replacements) == 0 {
		return nil
	}

	var pages []models.UIPage
	if err := tx.Select("id", "active_menu_path").
		Where("active_menu_path <> ''").
		Find(&pages).Error; err != nil {
		return err
	}
	for _, pageItem := range pages {
		activePath := normalizeManagedMenuPath(pageItem.ActiveMenuPath)
		if activePath == "" {
			continue
		}
		nextPath := activePath
		matched := false
		for oldPath, newPath := range replacements {
			if !pathHasPrefix(activePath, oldPath) {
				continue
			}
			nextPath = strings.TrimSuffix(newPath, "/") + strings.TrimPrefix(activePath, oldPath)
			matched = true
			break
		}
		if !matched || nextPath == activePath {
			continue
		}
		if err := tx.Model(&models.UIPage{}).
			Where("id = ?", pageItem.ID).
			Update("active_menu_path", nextPath).Error; err != nil {
			return err
		}
	}
	return nil
}

func matchesDeletedMenuPath(activePath string, deletedPaths []string) bool {
	for _, item := range deletedPaths {
		if pathHasPrefix(activePath, item) {
			return true
		}
	}
	return false
}

func pathHasPrefix(path string, prefix string) bool {
	target := normalizeManagedMenuPath(path)
	base := normalizeManagedMenuPath(prefix)
	if target == "" || base == "" {
		return false
	}
	return target == base || strings.HasPrefix(target, strings.TrimRight(base, "/")+"/")
}

func buildPromotedMenuFullPathMap(flat []user.Menu, deletedID uuid.UUID, targetParentID *uuid.UUID) map[uuid.UUID]string {
	overrides := make(map[uuid.UUID]*uuid.UUID)
	for _, item := range flat {
		if item.ParentID == nil {
			continue
		}
		parentID := *item.ParentID
		overrides[item.ID] = &parentID
	}
	for _, item := range flat {
		if item.ParentID != nil && *item.ParentID == deletedID {
			overrides[item.ID] = targetParentID
		}
	}
	delete(overrides, deletedID)
	return buildMenuFullPathMapWithParents(flat, overrides, deletedID)
}

func validateMenuPromoteTarget(flat []user.Menu, menuID, targetParentID uuid.UUID) error {
	if targetParentID == menuID {
		return ErrMenuHasChildren
	}
	ancestorMap := make(map[uuid.UUID]*uuid.UUID, len(flat))
	for _, item := range flat {
		ancestorMap[item.ID] = item.ParentID
	}
	cur := &targetParentID
	for cur != nil {
		if *cur == menuID {
			return ErrMenuHasChildren
		}
		cur = ancestorMap[*cur]
	}
	return nil
}

func buildMenuFullPathMap(flat []user.Menu) map[uuid.UUID]string {
	return buildMenuFullPathMapWithParents(flat, nil, uuid.Nil)
}

func buildMenuFullPathMapWithParents(
	flat []user.Menu,
	parentOverrides map[uuid.UUID]*uuid.UUID,
	skipID uuid.UUID,
) map[uuid.UUID]string {
	menuMap := make(map[uuid.UUID]user.Menu, len(flat))
	for _, item := range flat {
		if skipID != uuid.Nil && item.ID == skipID {
			continue
		}
		menuMap[item.ID] = item
	}

	result := make(map[uuid.UUID]string, len(menuMap))
	visiting := make(map[uuid.UUID]struct{})
	var resolve func(uuid.UUID) string
	resolve = func(menuID uuid.UUID) string {
		if path, ok := result[menuID]; ok {
			return path
		}
		item, ok := menuMap[menuID]
		if !ok {
			return ""
		}
		if _, ok := visiting[menuID]; ok {
			return normalizeManagedMenuPath(item.Path)
		}
		visiting[menuID] = struct{}{}
		defer delete(visiting, menuID)

		var parentID *uuid.UUID
		if parentOverrides != nil {
			if override, ok := parentOverrides[menuID]; ok {
				parentID = override
			} else {
				parentID = item.ParentID
			}
		} else {
			parentID = item.ParentID
		}

		parentPath := ""
		if parentID != nil {
			parentPath = resolve(*parentID)
		}
		fullPath := joinManagedMenuPath(item.Path, parentPath)
		result[menuID] = fullPath
		return fullPath
	}

	for menuID := range menuMap {
		resolve(menuID)
	}
	return result
}

func joinManagedMenuPath(path string, parentPath string) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return normalizeManagedMenuPath(parentPath)
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	if strings.HasPrefix(target, "/") {
		return normalizeManagedMenuPath(target)
	}
	base := normalizeManagedMenuPath(parentPath)
	if base == "" || base == "/" {
		return normalizeManagedMenuPath("/" + target)
	}
	return normalizeManagedMenuPath(strings.TrimRight(base, "/") + "/" + strings.TrimLeft(target, "/"))
}

func normalizeManagedMenuPath(path string) string {
	target := strings.TrimSpace(path)
	if target == "" {
		return ""
	}
	if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
		return target
	}
	normalized := "/" + strings.TrimLeft(target, "/")
	normalized = strings.ReplaceAll(normalized, "//", "/")
	if normalized != "/" {
		normalized = strings.TrimRight(normalized, "/")
	}
	return normalized
}
