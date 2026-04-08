package page

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// ListUnregistered returns page candidates found in the frontend views directory
// that are not yet registered in the database.
func (s *service) ListUnregistered(appKey string) ([]UnregisteredRecord, error) {
	return s.buildUnregisteredRecords(normalizeAppKey(appKey))
}

// Sync creates database records for every unregistered page candidate.
func (s *service) Sync(appKey string) (*SyncResult, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	items, err := s.buildUnregisteredRecords(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	result := &SyncResult{
		CreatedKeys: make([]string, 0, len(items)),
	}
	for _, item := range items {
		req := &SaveRequest{
			AppKey:            normalizedAppKey,
			PageKey:           item.PageKey,
			Name:              item.Name,
			RouteName:         item.RouteName,
			RoutePath:         item.RoutePath,
			Component:         item.Component,
			PageType:          item.PageType,
			VisibilityScope:   item.VisibilityScope,
			Source:            "sync",
			ModuleKey:         item.ModuleKey,
			ParentMenuID:      item.ParentMenuID,
			ActiveMenuPath:    item.ActiveMenuPath,
			BreadcrumbMode:    "inherit_menu",
			AccessMode:        deriveSyncedPageAccessMode(item.PageType),
			InheritPermission: boolPtr(item.PageType == "inner"),
			KeepAlive:         boolPtr(false),
			IsFullPage:        boolPtr(false),
			Status:            "normal",
			Meta:              map[string]interface{}{},
		}
		if _, err := s.Create(req); err != nil {
			return nil, err
		}
		result.CreatedCount++
		result.CreatedKeys = append(result.CreatedKeys, item.PageKey)
	}
	result.SkippedCount = len(items) - result.CreatedCount
	return result, nil
}

// ---------------------------------------------------------------------------
// private sync helpers
// ---------------------------------------------------------------------------

func (s *service) buildUnregisteredRecords(appKey string) ([]UnregisteredRecord, error) {
	viewPages, err := enumerateManagedViewPages()
	if err != nil {
		return nil, err
	}

	var existingPages []models.UIPage
	if err := s.db.Where("app_key = ?", normalizeAppKey(appKey)).Find(&existingPages).Error; err != nil {
		return nil, err
	}
	pageComponentSet := make(map[string]struct{}, len(existingPages))
	pageKeySet := make(map[string]struct{}, len(existingPages))
	routeNameSet := make(map[string]struct{}, len(existingPages))
	for _, item := range existingPages {
		pageComponentSet[strings.TrimSpace(item.Component)] = struct{}{}
		pageKeySet[strings.TrimSpace(item.PageKey)] = struct{}{}
		routeNameSet[strings.TrimSpace(item.RouteName)] = struct{}{}
	}

	menuMap, err := s.loadMenuMap(appKey, "")
	if err != nil {
		return nil, err
	}
	menuComponentSet := make(map[string]struct{}, len(menuMap))
	for _, node := range menuMap {
		component := strings.TrimSpace(node.Menu.Component)
		if component != "" {
			menuComponentSet[component] = struct{}{}
		}
	}

	result := make([]UnregisteredRecord, 0, len(viewPages))
	routePathSet := make(map[string]struct{})
	for _, page := range viewPages {
		component := strings.TrimSpace(page.Component)
		if component == "" {
			continue
		}
		if _, ok := pageComponentSet[component]; ok {
			continue
		}
		if _, ok := menuComponentSet[component]; ok {
			continue
		}

		candidate := deriveUnregisteredRecord(page, menuMap, pageKeySet, routeNameSet, routePathSet)
		pageKeySet[candidate.PageKey] = struct{}{}
		routeNameSet[candidate.RouteName] = struct{}{}
		routePathSet[candidate.RoutePath] = struct{}{}
		result = append(result, candidate)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Component < result[j].Component
	})
	return result, nil
}

func enumerateManagedViewPages() ([]scannedViewPage, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, err
	}
	viewsDir := filepath.Join(projectRoot, "frontend", "src", "views")
	items := make([]scannedViewPage, 0, 128)

	err = filepath.WalkDir(viewsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".vue" {
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		filePath := "/" + filepath.ToSlash(rel)
		component := toComponentPath(filePath)
		if !isManagedViewComponent(component) {
			return nil
		}
		items = append(items, scannedViewPage{
			FilePath:  filePath,
			Component: component,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

func deriveUnregisteredRecord(
	item scannedViewPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageKeySet map[string]struct{},
	routeNameSet map[string]struct{},
	routePathSet map[string]struct{},
) UnregisteredRecord {
	routePath := normalizeRoutePath(item.Component)
	moduleKey := deriveModuleKey(item.Component)
	name := derivePageDisplayName(item.Component)
	pageKey := ensureUniqueValue(derivePageKey(item.Component), pageKeySet)
	routeName := ensureUniqueValue(deriveRouteName(item.Component), routeNameSet)
	if _, ok := routePathSet[routePath]; ok {
		routePath = ensureUniqueRoutePath(routePath, routePathSet)
	}

	parentMenuID, parentMenuName, activeMenuPath := guessParentMenu(routePath, menuMap)
	pageType := models.PageTypeStandalone
	visibilityScope := pageVisibilityScopeApp
	if parentMenuID != "" {
		pageType = models.PageTypeInner
		visibilityScope = pageVisibilityScopeInherit
	}

	return UnregisteredRecord{
		FilePath:        item.FilePath,
		Component:       item.Component,
		PageKey:         pageKey,
		Name:            name,
		RouteName:       routeName,
		RoutePath:       routePath,
		PageType:        pageType,
		VisibilityScope: visibilityScope,
		ModuleKey:       moduleKey,
		ParentMenuID:    parentMenuID,
		ParentMenuName:  parentMenuName,
		ActiveMenuPath:  activeMenuPath,
	}
}

func guessParentMenu(routePath string, menuMap map[uuid.UUID]runtimeMenuNode) (string, string, string) {
	targetPath := normalizeRoutePath(routePath)
	bestDepth := -1
	bestID := ""
	bestName := ""
	bestPath := ""
	for _, node := range menuMap {
		menuPath := normalizeRoutePath(node.FullPath)
		if menuPath == "" || menuPath == "/" {
			continue
		}
		if !strings.HasPrefix(targetPath, menuPath) {
			continue
		}
		if targetPath != menuPath && !strings.HasPrefix(targetPath, menuPath+"/") {
			continue
		}
		depth := strings.Count(menuPath, "/")
		if depth <= bestDepth {
			continue
		}
		bestDepth = depth
		bestID = node.Menu.ID.String()
		bestName = firstNonEmpty(strings.TrimSpace(node.Menu.Title), strings.TrimSpace(node.Menu.Name))
		bestPath = menuPath
	}
	return bestID, bestName, bestPath
}

func deriveSyncedPageAccessMode(pageType string) string {
	if normalizePageType(pageType) == "inner" {
		return "inherit"
	}
	return "jwt"
}

func ensureUniqueValue(base string, used map[string]struct{}) string {
	target := strings.TrimSpace(base)
	if target == "" {
		target = "page"
	}
	if _, ok := used[target]; !ok {
		return target
	}
	for idx := 2; idx < 10000; idx++ {
		candidate := target + "_" + strconv.Itoa(idx)
		if _, ok := used[candidate]; !ok {
			return candidate
		}
	}
	return target + "_" + uuid.NewString()[:8]
}

func ensureUniqueRoutePath(base string, used map[string]struct{}) string {
	target := normalizeRoutePath(base)
	if _, ok := used[target]; !ok {
		return target
	}
	for idx := 2; idx < 10000; idx++ {
		candidate := normalizeRoutePath(target + "-" + strconv.Itoa(idx))
		if _, ok := used[candidate]; !ok {
			return candidate
		}
	}
	return normalizeRoutePath(target + "-" + uuid.NewString()[:8])
}

func derivePageKey(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return "page"
	}
	normalized := make([]string, 0, len(segments))
	for _, segment := range segments {
		normalized = append(normalized, sanitizeSegment(segment))
	}
	return strings.Join(normalized, ".")
}

func deriveRouteName(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return "Page"
	}
	builder := strings.Builder{}
	for _, segment := range segments {
		builder.WriteString(toPascalCase(segment))
	}
	result := builder.String()
	if result == "" {
		return "Page"
	}
	return result
}

func derivePageDisplayName(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return "未命名页面"
	}
	return humanizeSegment(segments[len(segments)-1])
}

func deriveModuleKey(component string) string {
	segments := splitComponentSegments(component)
	if len(segments) == 0 {
		return ""
	}
	return sanitizeSegment(segments[0])
}

func splitComponentSegments(component string) []string {
	target := strings.Trim(component, "/")
	if target == "" {
		return []string{}
	}
	segments := strings.Split(target, "/")
	result := make([]string, 0, len(segments))
	for _, segment := range segments {
		segment = strings.TrimSpace(segment)
		if segment == "" || segment == "index" {
			continue
		}
		result = append(result, segment)
	}
	return result
}

func humanizeSegment(segment string) string {
	parts := splitWords(segment)
	if len(parts) == 0 {
		return "未命名页面"
	}
	for idx, part := range parts {
		parts[idx] = strings.Title(part)
	}
	return strings.Join(parts, " ")
}

func toPascalCase(segment string) string {
	parts := splitWords(segment)
	if len(parts) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			builder.WriteString(part[1:])
		}
	}
	return builder.String()
}

func splitWords(value string) []string {
	replacer := strings.NewReplacer("-", " ", "_", " ", ".", " ")
	normalized := replacer.Replace(strings.TrimSpace(value))
	return strings.Fields(normalized)
}

func sanitizeSegment(segment string) string {
	target := strings.TrimSpace(segment)
	if target == "" {
		return "page"
	}
	var builder strings.Builder
	for _, r := range target {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(unicode.ToLower(r))
		case r == '-' || r == '_' || r == '.':
			builder.WriteRune('_')
		}
	}
	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "page"
	}
	return result
}

func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	current := wd
	for {
		frontendPath := filepath.Join(current, "frontend")
		backendPath := filepath.Join(current, "backend")
		if _, err := os.Stat(frontendPath); err == nil {
			if _, err := os.Stat(backendPath); err == nil {
				return current, nil
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return wd, nil
}

func toComponentPath(filePath string) string {
	withoutPrefix := strings.TrimPrefix(filePath, "/frontend/src/views")
	withoutExt := strings.TrimSuffix(withoutPrefix, ".vue")
	normalized := strings.TrimSuffix(withoutExt, "/index")
	normalized = strings.ReplaceAll(normalized, "//", "/")
	if normalized == "" {
		return "/"
	}
	if !strings.HasPrefix(normalized, "/") {
		return "/" + normalized
	}
	return normalized
}

func isManagedViewComponent(component string) bool {
	target := normalizeRoutePath(component)
	switch {
	case target == "", target == "/", target == "/index", target == "/outside/Iframe":
		return false
	case strings.Contains(target, "/modules/"):
		return false
	case strings.HasPrefix(target, "/auth/"),
		strings.HasPrefix(target, "/exception/"),
		strings.HasPrefix(target, "/result/"):
		return false
	default:
		return true
	}
}
