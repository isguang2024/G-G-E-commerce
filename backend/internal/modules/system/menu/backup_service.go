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

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	page "github.com/gg-ecommerce/backend/internal/modules/system/page"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

// ---------------------------------------------------------------------------
// Backup interface methods
// ---------------------------------------------------------------------------

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

	backupPtrs := make([]*user.MenuBackup, len(backups))
	for i, backup := range backups {
		backupCopy := backup
		backupPtrs[i] = &backupCopy
	}

	return backupPtrs, nil
}

func (s *menuService) DeleteBackup(id uuid.UUID, appKey string) error {
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

	if err := s.cleanupInvalidMenuRelations(); err != nil {
		return err
	}
	if err := s.refreshAllMenuSnapshots(); err != nil {
		return err
	}

	s.logger.Info("Menu backup restored", zap.String("backup_id", id.String()))
	return nil
}

// ---------------------------------------------------------------------------
// private backup helpers
// ---------------------------------------------------------------------------

func (s *menuService) cleanupInvalidMenuRelations() error {
	if s.db == nil {
		return nil
	}
	statements := []string{
		"DELETE FROM feature_package_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
		"DELETE FROM role_hidden_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
		"DELETE FROM collaboration_workspace_blocked_menus WHERE menu_id NOT IN (SELECT id FROM menu_definitions)",
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

// ---------------------------------------------------------------------------
// backup payload types and helpers
// ---------------------------------------------------------------------------

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
	return normalizeMenuSpaceKey("") // returns DefaultMenuSpaceKey
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
