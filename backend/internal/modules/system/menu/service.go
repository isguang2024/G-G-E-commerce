package menu

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/maben/backend/internal/api/dto"
	apppkg "github.com/maben/backend/internal/modules/system/app"
	"github.com/maben/backend/internal/modules/system/models"
	page "github.com/maben/backend/internal/modules/system/page"
	spaceutil "github.com/maben/backend/internal/modules/system/space"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/permissionrefresh"
	"github.com/maben/backend/internal/pkg/platformaccess"
	"github.com/maben/backend/internal/pkg/platformroleaccess"
)

var (
	ErrMenuNotFound          = errors.New("菜单不存在")
	ErrMenuSystemProtected   = errors.New("系统默认菜单不可删除")
	ErrMenuDeleteModeInvalid = errors.New("无效的菜单删除方式")
	ErrMenuHasChildren       = errors.New("该菜单存在子菜单，请选择删除策略")
)

type MenuService interface {
	GetTree(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error)
	Create(req *dto.MenuCreateRequest) (*user.Menu, error)
	Update(id uuid.UUID, req *dto.MenuUpdateRequest) error
	DeletePreview(id uuid.UUID, mode string, targetParentID *uuid.UUID) (*MenuDeletePreview, error)
	Delete(id uuid.UUID, mode string, targetParentID *uuid.UUID) error
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

// ---------------------------------------------------------------------------
// Menu CRUD
// ---------------------------------------------------------------------------

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
	kind, component, meta := sanitizeMenuPayloadByKind(req.Kind, req.Component, req.Meta)
	appKey := normalizeMenuAppKey(req.AppKey)
	spaceKey := strings.TrimSpace(req.MenuSpaceKey)
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
			MenuSpaceKey:  normalizeMenuSpaceKey(spaceKey),
			MenuKey:       menuKey,
			ParentMenuKey: parentMenuKey,
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
		s.logger.Warn("menu create: snapshot refresh failed (non-fatal)", zap.Error(err))
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
		if strings.TrimSpace(req.MenuSpaceKey) == "" {
			return nil
		}

		parentID, err := parseOptionalUUID(req.ParentID)
		if err != nil {
			return err
		}
		if parentID != nil && *parentID == id {
			return errors.New("不能将上级设为自己")
		}
		layoutMenus, err := s.menuRepo.ListByAppAndSpace(definition.AppKey, normalizeMenuSpaceKey(req.MenuSpaceKey))
		if err != nil {
			return err
		}
		if parentID != nil && s.isDescendant(layoutMenus, id, *parentID) {
			return errors.New("不能将上级设为自身子级（会造成循环）")
		}
		parentMenuKey, err := s.resolveParentMenuKey(tx, definition.AppKey, req.MenuSpaceKey, parentID)
		if err != nil {
			return err
		}
		placement := &models.SpaceMenuPlacement{
			AppKey:        definition.AppKey,
			MenuSpaceKey:  normalizeMenuSpaceKey(req.MenuSpaceKey),
			MenuKey:       nextMenuKey,
			ParentMenuKey: parentMenuKey,
			SortOrder:     req.SortOrder,
			Hidden:        req.Hidden,
		}
		return tx.Unscoped().Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "app_key"},
				{Name: "menu_space_key"},
				{Name: "menu_key"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"parent_menu_key": parentMenuKey,
				"sort_order":      req.SortOrder,
				"hidden":          req.Hidden,
				"updated_at":      time.Now(),
				"deleted_at":      nil,
			}),
		}).Create(placement).Error
	}); err != nil {
		return err
	}
	page.InvalidateRuntimeCache()
	invalidateMenuCaches()
	if err := s.refreshAllMenuSnapshots(); err != nil {
		s.logger.Warn("menu update: snapshot refresh failed (non-fatal)", zap.Error(err))
	}
	return nil
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
	if err := s.refreshAllMenuSnapshots(); err != nil {
		s.logger.Warn("menu delete: snapshot refresh failed (non-fatal)", zap.Error(err))
	}
	return nil
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

// ---------------------------------------------------------------------------
// shared private helpers
// ---------------------------------------------------------------------------

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
	type result struct {
		Total int64
	}
	var r result
	err := s.db.Raw(`
		SELECT (
			SELECT COUNT(*) FROM feature_package_menus WHERE menu_id IN ?
		) + (
			SELECT COUNT(*) FROM role_hidden_menus WHERE menu_id IN ?
		) + (
			SELECT COUNT(*) FROM collaboration_workspace_blocked_menus WHERE menu_id IN ?
		) + (
			SELECT COUNT(*) FROM user_hidden_menus WHERE menu_id IN ?
		) AS total`,
		menuIDs, menuIDs, menuIDs, menuIDs,
	).Scan(&r).Error
	if err != nil {
		return 0, err
	}
	return int(r.Total), nil
}

func (s *menuService) refreshAllMenuSnapshots() error {
	if s.refresher == nil || s.db == nil {
		return nil
	}
	if err := s.refresher.RefreshAllCollaborationWorkspaces(); err != nil {
		return err
	}
	if err := s.refresher.RefreshAllPersonalWorkspaceRoles(); err != nil {
		return err
	}
	return s.refresher.RefreshAllPersonalWorkspaceUsers()
}

func (s *menuService) cleanupMenuRelationsByIDs(tx *gorm.DB, menuIDs []uuid.UUID) error {
	if tx == nil || len(menuIDs) == 0 {
		return nil
	}
	statements := []string{
		"DELETE FROM feature_package_menus WHERE menu_id IN ?",
		"DELETE FROM role_hidden_menus WHERE menu_id IN ?",
		"DELETE FROM collaboration_workspace_blocked_menus WHERE menu_id IN ?",
		"DELETE FROM user_hidden_menus WHERE menu_id IN ?",
	}
	for _, statement := range statements {
		if err := tx.Exec(statement, menuIDs).Error; err != nil {
			return err
		}
	}
	return nil
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
			Where("app_key = ? AND menu_space_key = ? AND menu_key = ?", normalizeMenuAppKey(appKey), normalizeMenuSpaceKey(spaceKey), parent.MenuKey).
			Count(&placementCount).Error; err != nil {
			return "", err
		}
		if placementCount == 0 {
			return "", fmt.Errorf("上级菜单未放入当前空间")
		}
	}
	return strings.TrimSpace(parent.MenuKey), nil
}

// ---------------------------------------------------------------------------
// package-level helpers
// ---------------------------------------------------------------------------

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
	if strings.TrimSpace(component) != "" {
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

const (
	menuDeleteModeSingle          = "single"
	menuDeleteModeCascade         = "cascade"
	menuDeleteModePromoteChildren = "promote_children"
)

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
	target := spaceutil.NormalizeMenuSpaceKey(spaceKey)
	if target == "" {
		return flat
	}
	result := make([]user.Menu, 0, len(flat))
	for _, menu := range flat {
		if spaceutil.NormalizeMenuSpaceKey(menu.MenuSpaceKey) != target {
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
