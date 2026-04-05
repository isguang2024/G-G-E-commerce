package page

import (
	"sort"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const (
	pageVisibilityScopeInherit = "inherit"
	pageVisibilityScopeApp     = "app"
	pageVisibilityScopeSpaces  = "spaces"
)

func loadPageSpaceBindingMap(dbQuery *gorm.DB, appKey string) (map[uuid.UUID][]string, error) {
	var bindings []models.PageSpaceBinding
	if err := dbQuery.Where("app_key = ?", normalizeAppKey(appKey)).Find(&bindings).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID][]string, len(bindings))
	for _, item := range bindings {
		pageID := item.PageID
		spaceKey := normalizeSpaceKey(item.SpaceKey)
		if pageID == uuid.Nil || spaceKey == "" {
			continue
		}
		result[pageID] = append(result[pageID], spaceKey)
	}
	for id, keys := range result {
		result[id] = uniqueSortedStrings(keys)
	}
	return result, nil
}

func resolvePageSpaceKeys(
	page models.UIPage,
	pageMap map[string]models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	bindingMap map[uuid.UUID][]string,
	seen map[string]struct{},
) []string {
	if page.ParentMenuID != nil {
		if node, ok := menuMap[*page.ParentMenuID]; ok {
			return []string{normalizeSpaceKey(node.Menu.SpaceKey)}
		}
	}
	if parentPageKey := strings.TrimSpace(page.ParentPageKey); parentPageKey != "" {
		if _, ok := seen[parentPageKey]; ok {
			return []string{}
		}
		if parentPage, ok := pageMap[parentPageKey]; ok {
			seen[parentPageKey] = struct{}{}
			defer delete(seen, parentPageKey)
			return resolvePageSpaceKeys(parentPage, pageMap, menuMap, bindingMap, seen)
		}
	}
	if keys := uniqueSortedStrings(bindingMap[page.ID]); len(keys) > 0 {
		return keys
	}
	return []string{}
}

func isPageVisibleInSpace(item models.UIPage, targetSpaceKey string) bool {
	if readPageVisibilityScope(item) == pageVisibilityScopeApp {
		return true
	}
	keys := readPageSpaceKeys(item)
	if len(keys) == 0 {
		return true
	}
	target := normalizeSpaceKey(targetSpaceKey)
	for _, key := range keys {
		if normalizeSpaceKey(key) == target {
			return true
		}
	}
	return false
}

func resolvePageVisibilityScope(item models.UIPage, keys []string) string {
	if item.ParentMenuID != nil || strings.TrimSpace(item.ParentPageKey) != "" {
		return pageVisibilityScopeInherit
	}
	if len(uniqueSortedStrings(keys)) > 0 {
		return pageVisibilityScopeSpaces
	}
	return pageVisibilityScopeApp
}

func readPageVisibilityScope(item models.UIPage) string {
	switch strings.TrimSpace(item.VisibilityScope) {
	case pageVisibilityScopeInherit, pageVisibilityScopeApp, pageVisibilityScopeSpaces:
		return strings.TrimSpace(item.VisibilityScope)
	}
	if item.Meta != nil {
		if scope, ok := item.Meta["spaceScope"].(string); ok {
			switch strings.TrimSpace(scope) {
			case "bound":
				return pageVisibilityScopeSpaces
			case "global":
				return pageVisibilityScopeApp
			case pageVisibilityScopeInherit, pageVisibilityScopeApp, pageVisibilityScopeSpaces:
				return strings.TrimSpace(scope)
			}
		}
	}
	if item.ParentMenuID != nil || strings.TrimSpace(item.ParentPageKey) != "" {
		return pageVisibilityScopeInherit
	}
	if item.Meta != nil {
		if values, ok := item.Meta["spaceKeys"].([]string); ok && len(uniqueSortedStrings(values)) > 0 {
			return pageVisibilityScopeSpaces
		}
		if values, ok := item.Meta["spaceKeys"].([]interface{}); ok {
			result := make([]string, 0, len(values))
			for _, value := range values {
				if text, ok := value.(string); ok {
					result = append(result, normalizeSpaceKey(text))
				}
			}
			if len(uniqueSortedStrings(result)) > 0 {
				return pageVisibilityScopeSpaces
			}
		}
	}
	return pageVisibilityScopeApp
}

func readPageSpaceKeys(item models.UIPage) []string {
	if readPageVisibilityScope(item) == pageVisibilityScopeApp {
		return []string{}
	}
	if item.Meta != nil {
		if values, ok := item.Meta["spaceKeys"].([]string); ok {
			return uniqueSortedStrings(values)
		}
		if values, ok := item.Meta["spaceKeys"].([]interface{}); ok {
			result := make([]string, 0, len(values))
			for _, value := range values {
				if text, ok := value.(string); ok {
					result = append(result, normalizeSpaceKey(text))
				}
			}
			return uniqueSortedStrings(result)
		}
	}
	return []string{}
}

func applyResolvedPageSpace(item *models.UIPage, keys []string) {
	if item == nil {
		return
	}
	keys = uniqueSortedStrings(keys)
	visibilityScope := resolvePageVisibilityScope(*item, keys)
	item.VisibilityScope = visibilityScope
	if item.Meta == nil {
		item.Meta = models.MetaJSON{}
	}
	delete(item.Meta, "spaceKeys")
	delete(item.Meta, "spaceScope")
	switch visibilityScope {
	case pageVisibilityScopeInherit:
		item.SpaceKey = ""
		if len(keys) > 0 {
			item.Meta["spaceKeys"] = keys
		}
		item.Meta["spaceScope"] = pageVisibilityScopeInherit
	case pageVisibilityScopeSpaces:
		item.SpaceKey = ""
		item.Meta["spaceKeys"] = keys
		item.Meta["spaceScope"] = pageVisibilityScopeSpaces
	default:
		item.SpaceKey = ""
		item.Meta["spaceScope"] = pageVisibilityScopeApp
	}
}

func isMenuBackedEntryPage(item models.UIPage, menuMap map[uuid.UUID]runtimeMenuNode) bool {
	if item.ParentMenuID == nil {
		return false
	}
	if isRoutelessPageType(item.PageType) {
		return false
	}
	node, ok := menuMap[*item.ParentMenuID]
	if !ok {
		return false
	}
	if resolveMenuKind(node.Menu) != models.MenuKindEntry {
		return false
	}
	return normalizeRoutePath(item.RoutePath) == normalizeRoutePath(node.FullPath) &&
		strings.TrimSpace(item.Component) != "" &&
		strings.TrimSpace(item.Component) == strings.TrimSpace(node.Menu.Component)
}

func resolveMenuKind(item models.Menu) string {
	switch strings.TrimSpace(item.Kind) {
	case models.MenuKindDirectory, models.MenuKindEntry, models.MenuKindExternal:
		return strings.TrimSpace(item.Kind)
	}
	if item.Meta != nil {
		if link, ok := item.Meta["link"].(string); ok && strings.TrimSpace(link) != "" {
			return models.MenuKindExternal
		}
	}
	if strings.TrimSpace(item.Component) != "" && strings.TrimSpace(item.Component) != "/index/index" {
		return models.MenuKindEntry
	}
	return models.MenuKindDirectory
}

func uniqueSortedStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		key := normalizeSpaceKey(item)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
