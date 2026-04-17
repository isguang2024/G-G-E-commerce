package page

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// firstNonEmpty returns the first non-blank string from the given values.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if target := strings.TrimSpace(value); target != "" {
			return target
		}
	}
	return ""
}

// buildRuntimePageRecords converts a slice of Records into the JSON shapes
// expected by the runtime page registry API.
func buildRuntimePageRecords(items []Record) []gin.H {
	if len(items) == 0 {
		return []gin.H{}
	}
	pageMap := make(map[string]Record, len(items))
	for _, item := range items {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		pageMap[pageKey] = item
	}

	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		if normalizePageType(item.PageType) == "group" {
			continue
		}
		result = append(result, buildRuntimePageRecord(flattenRuntimePageRecord(item, pageMap)))
	}
	return result
}

func buildRuntimePageRecord(item Record) gin.H {
	node := gin.H{
		"page_key":   item.PageKey,
		"name":       item.Name,
		"route_path": item.RoutePath,
	}
	if item.Meta != nil {
		if values, ok := pageMetaMenuSpaceKeys(item.Meta).([]string); ok && len(values) > 0 {
			node["menu_space_keys"] = values
			node["page_space_bindings"] = buildRuntimePageSpaceBindings(values, item.VisibilityScope)
		} else if values, ok := pageMetaMenuSpaceKeys(item.Meta).([]interface{}); ok && len(values) > 0 {
			node["menu_space_keys"] = values
			node["page_space_bindings"] = buildRuntimePageSpaceBindings(interfaceStrings(values), item.VisibilityScope)
		}
		if scope, ok := item.Meta["menuSpaceScope"].(string); ok && strings.TrimSpace(scope) != "" {
			node["space_scope"] = scope
		}
	}
	if scope := strings.TrimSpace(item.VisibilityScope); scope != "" && scope != "inherit" {
		node["visibility_scope"] = scope
	}
	if routeName := strings.TrimSpace(item.RouteName); routeName != "" && routeName != strings.TrimSpace(item.PageKey) {
		node["route_name"] = routeName
	}
	if component := strings.TrimSpace(item.Component); component != "" {
		node["component"] = component
	}
	if pageType := strings.TrimSpace(item.PageType); pageType != "" && pageType != "inner" {
		node["page_type"] = pageType
	}
	if item.ParentMenuID != nil {
		node["parent_menu_id"] = item.ParentMenuID.String()
	}
	if parentPageKey := strings.TrimSpace(item.ParentPageKey); parentPageKey != "" {
		node["parent_page_key"] = parentPageKey
	}
	if activeMenuPath := strings.TrimSpace(item.ActiveMenuPath); activeMenuPath != "" {
		node["active_menu_path"] = activeMenuPath
	}
	if breadcrumbMode := strings.TrimSpace(item.BreadcrumbMode); breadcrumbMode != "" && breadcrumbMode != "inherit_menu" {
		node["breadcrumb_mode"] = breadcrumbMode
	}
	if accessMode := strings.TrimSpace(item.AccessMode); accessMode != "" && accessMode != "inherit" {
		node["access_mode"] = accessMode
	}
	if permissionKey := strings.TrimSpace(item.PermissionKey); permissionKey != "" {
		node["permission_key"] = permissionKey
	}
	if item.KeepAlive {
		node["keep_alive"] = true
	}
	if item.IsFullPage {
		node["is_full_page"] = true
	}
	if status := strings.TrimSpace(item.Status); status != "" && status != "normal" {
		node["status"] = status
	}

	meta := gin.H{}
	if item.Meta != nil {
		if value, ok := item.Meta["isIframe"].(bool); ok && value {
			meta["isIframe"] = true
		}
		if value, ok := item.Meta["isHideTab"].(bool); ok && value {
			meta["isHideTab"] = true
		}
		if value, ok := item.Meta["link"].(string); ok && strings.TrimSpace(value) != "" {
			meta["link"] = strings.TrimSpace(value)
		}
	}
	if len(meta) > 0 {
		node["meta"] = meta
	}

	return node
}

func buildRuntimePageSpaceBindings(menuSpaceKeys []string, visibilityScope string) []gin.H {
	keys := uniqueSortedStrings(menuSpaceKeys)
	if len(keys) == 0 {
		return nil
	}
	source := "explicit_binding"
	if strings.TrimSpace(visibilityScope) == pageVisibilityScopeInherit {
		source = "inherited_binding"
	}
	result := make([]gin.H, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		result = append(result, gin.H{
			"menu_space_key": key,
			"source":         source,
		})
	}
	return result
}

func interfaceStrings(values []interface{}) []string {
	if len(values) == 0 {
		return nil
	}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if text, ok := value.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func flattenRuntimePageRecord(item Record, pageMap map[string]Record) Record {
	flattened := item
	flattened.RoutePath = resolveRuntimeOutputRoutePath(item, pageMap, map[string]struct{}{})
	flattened.ParentPageKey = resolveNearestRuntimeParentPageKey(item, pageMap)
	if mode, permissionKey, ok := resolveRuntimeGroupAccessOverride(item, pageMap); ok {
		flattened.AccessMode = mode
		flattened.PermissionKey = permissionKey
	}
	return flattened
}

func resolveRuntimeOutputRoutePath(
	page Record,
	pageMap map[string]Record,
	seen map[string]struct{},
) string {
	pageKey := strings.TrimSpace(page.PageKey)
	if pageKey != "" {
		if _, ok := seen[pageKey]; ok {
			return ""
		}
		seen[pageKey] = struct{}{}
		defer delete(seen, pageKey)
	}

	rawRoutePath := strings.TrimSpace(page.RoutePath)
	if rawRoutePath == "" {
		return resolveRuntimeOutputBasePath(page, pageMap, seen)
	}
	if strings.HasPrefix(rawRoutePath, "http://") || strings.HasPrefix(rawRoutePath, "https://") {
		return rawRoutePath
	}
	if strings.HasPrefix(rawRoutePath, "/") && !isSingleSegmentRuntimePath(rawRoutePath) {
		return normalizeRoutePath(rawRoutePath)
	}

	basePath := resolveRuntimeOutputBasePath(page, pageMap, seen)
	segment := strings.TrimLeft(rawRoutePath, "/")
	if basePath != "" && !strings.HasPrefix(basePath, "http://") && !strings.HasPrefix(basePath, "https://") {
		return buildMenuFullPath(segment, basePath)
	}
	return normalizeRoutePath(segment)
}

func resolveRuntimeOutputBasePath(
	page Record,
	pageMap map[string]Record,
	seen map[string]struct{},
) string {
	if activeMenuPath := normalizeRoutePath(page.ActiveMenuPath); activeMenuPath != "" {
		return activeMenuPath
	}
	parentPageKey := strings.TrimSpace(page.ParentPageKey)
	if parentPageKey == "" {
		return ""
	}
	parentPage, ok := pageMap[parentPageKey]
	if !ok {
		return ""
	}
	return resolveRuntimeOutputRoutePath(parentPage, pageMap, seen)
}

func resolveNearestRuntimeParentPageKey(page Record, pageMap map[string]Record) string {
	parentPageKey := strings.TrimSpace(page.ParentPageKey)
	for parentPageKey != "" {
		parentPage, ok := pageMap[parentPageKey]
		if !ok {
			return ""
		}
		if normalizePageType(parentPage.PageType) != "group" {
			return parentPage.PageKey
		}
		parentPageKey = strings.TrimSpace(parentPage.ParentPageKey)
	}
	return ""
}

func resolveRuntimeGroupAccessOverride(
	page Record,
	pageMap map[string]Record,
) (string, string, bool) {
	if normalizeAccessMode(page.AccessMode) != "inherit" {
		return "", "", false
	}

	parentPageKey := strings.TrimSpace(page.ParentPageKey)
	for parentPageKey != "" {
		parentPage, ok := pageMap[parentPageKey]
		if !ok {
			return "", "", false
		}
		if normalizePageType(parentPage.PageType) != "group" {
			return "", "", false
		}

		mode := normalizeAccessMode(parentPage.AccessMode)
		switch mode {
		case "public", "jwt":
			return mode, "", true
		case "permission":
			return mode, strings.TrimSpace(parentPage.PermissionKey), true
		default:
			parentPageKey = strings.TrimSpace(parentPage.ParentPageKey)
		}
	}
	return "", "", false
}

func isSingleSegmentRuntimePath(path string) bool {
	normalized := strings.Trim(strings.TrimSpace(path), "/")
	return normalized != "" && !strings.Contains(normalized, "/")
}
