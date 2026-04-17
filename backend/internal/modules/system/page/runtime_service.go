package page

import (
	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/models"
)

// ListRuntime returns runtime-visible pages for the given user and space.
func (s *service) ListRuntime(appKey, host, requestedMenuSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]Record, error) {
	return s.loadRuntimeRecords(normalizeAppKey(appKey), host, requestedMenuSpaceKey, userID, collaborationWorkspaceID)
}

// ListRuntimePublic returns publicly visible runtime pages.
func (s *service) ListRuntimePublic(appKey, host, requestedMenuSpaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) ([]Record, error) {
	return s.loadPublicRuntimeRecords(normalizeAppKey(appKey), host, requestedMenuSpaceKey, userID, collaborationWorkspaceID)
}

// ResolveCompiledAccessContext builds the compiled access context for a user/space.
func (s *service) ResolveCompiledAccessContext(appKey, spaceKey string, userID *uuid.UUID, collaborationWorkspaceID *uuid.UUID) (*CompiledAccessContext, error) {
	return s.buildCompiledAccessContextForSpace(normalizeAppKey(appKey), spaceKey, userID, collaborationWorkspaceID)
}

// ListRuntimeWithAccess returns runtime pages filtered by a pre-built access context.
func (s *service) ListRuntimeWithAccess(appKey, spaceKey string, accessCtx *CompiledAccessContext) ([]Record, error) {
	return s.loadRuntimeRecordsWithAccess(normalizeAppKey(appKey), spaceKey, accessCtx)
}

// buildRuntimeRecords loads all active pages and applies the managed-page model for the given space.
func (s *service) buildRuntimeRecords(appKey, spaceKey string) ([]Record, map[uuid.UUID]runtimeMenuNode, error) {
	var items []models.UIPage
	if err := s.db.Where("app_key = ?", normalizeAppKey(appKey)).Where("status = ? AND page_type <> ?", "normal", "display_group").
		Order("sort_order ASC, created_at ASC").
		Find(&items).Error; err != nil {
		return nil, nil, err
	}
	records, err := s.decorateRecords(items)
	if err != nil {
		return nil, nil, err
	}
	menuMap, err := s.loadMenuMap(normalizeAppKey(appKey), spaceKey)
	if err != nil {
		return nil, nil, err
	}
	pageMap, err := s.loadPageMap(normalizeAppKey(appKey))
	if err != nil {
		return nil, nil, err
	}
	bindingMap, err := loadPageSpaceBindingMap(s.db, normalizeAppKey(appKey))
	if err != nil {
		return nil, nil, err
	}
	filtered := s.applyManagedPageModel(records, spaceKey, menuMap, pageMap, bindingMap)
	return filtered, menuMap, nil
}

func (s *service) applyManagedPageModel(
	records []Record,
	spaceKey string,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
	bindingMap map[uuid.UUID][]string,
) []Record {
	filtered := make([]Record, 0, len(records))
	for index := range records {
		record := records[index]
		if isMenuBackedEntryPage(record.UIPage, menuMap) {
			continue
		}
		resolvedMenuSpaceKeys := resolvePageMenuSpaceKeys(record.UIPage, pageMap, menuMap, bindingMap, map[string]struct{}{})
		applyResolvedPageSpace(&record.UIPage, resolvedMenuSpaceKeys)
		record.ActiveMenuPath = s.resolveActiveMenuPath(
			&record.UIPage,
			menuMap,
			pageMap,
			map[string]struct{}{},
		)
		if !isPageVisibleInSpace(record.UIPage, spaceKey) {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}
