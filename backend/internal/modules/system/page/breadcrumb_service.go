package page

import (
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// PreviewBreadcrumb returns the resolved breadcrumb chain for a page.
func (s *service) PreviewBreadcrumb(id uuid.UUID, appKey string) ([]BreadcrumbPreviewItem, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	page, err := s.findPageByID(id, normalizedAppKey)
	if err != nil {
		return nil, err
	}
	menuMap, err := s.loadMenuMap(normalizedAppKey, "")
	if err != nil {
		return nil, err
	}
	pageMap, err := s.loadPageMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	chain, err := s.resolveBreadcrumbChain(page, menuMap, pageMap)
	if err != nil {
		return nil, err
	}
	result := make([]BreadcrumbPreviewItem, 0, len(chain)+1)
	result = append(result, chain...)
	result = append(result, BreadcrumbPreviewItem{
		Type:    "page",
		Title:   page.Name,
		Path:    strings.TrimSpace(page.RoutePath),
		PageKey: page.PageKey,
	})
	return result, nil
}

// ---------------------------------------------------------------------------
// private breadcrumb helpers
// ---------------------------------------------------------------------------

func (s *service) resolveBreadcrumbChain(
	page *models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
) ([]BreadcrumbPreviewItem, error) {
	if page == nil {
		return []BreadcrumbPreviewItem{}, nil
	}
	switch normalizeBreadcrumbMode(page.BreadcrumbMode) {
	case "inherit_page":
		chain, err := s.resolveParentPageBreadcrumbChain(page, menuMap, pageMap, map[string]struct{}{})
		if err != nil {
			return nil, err
		}
		if len(chain) > 0 {
			return chain, nil
		}
		fallthrough
	case "custom":
		fallthrough
	default:
		activePath := s.resolveActiveMenuPath(page, menuMap, pageMap, map[string]struct{}{})
		return resolveMenuBreadcrumbChain(activePath, menuMap), nil
	}
}

func (s *service) resolveParentPageBreadcrumbChain(
	page *models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
	seen map[string]struct{},
) ([]BreadcrumbPreviewItem, error) {
	if page == nil {
		return []BreadcrumbPreviewItem{}, nil
	}
	parentKey := strings.TrimSpace(page.ParentPageKey)
	if parentKey == "" {
		activePath := s.resolveActiveMenuPath(page, menuMap, pageMap, seen)
		return resolveMenuBreadcrumbChain(activePath, menuMap), nil
	}
	if _, ok := seen[parentKey]; ok {
		return nil, fmt.Errorf("%w: 页面面包屑存在循环引用", ErrPageValidation)
	}
	parentPage, ok := pageMap[parentKey]
	if !ok {
		activePath := s.resolveActiveMenuPath(page, menuMap, pageMap, seen)
		return resolveMenuBreadcrumbChain(activePath, menuMap), nil
	}
	seen[parentKey] = struct{}{}
	parentChain, err := s.resolveBreadcrumbChain(&parentPage, menuMap, pageMap)
	delete(seen, parentKey)
	if err != nil {
		return nil, err
	}
	if isRoutelessPageType(parentPage.PageType) || strings.TrimSpace(parentPage.RoutePath) == "" {
		return parentChain, nil
	}
	return append(parentChain, BreadcrumbPreviewItem{
		Type:    "page",
		Title:   parentPage.Name,
		Path:    normalizeRoutePath(parentPage.RoutePath),
		PageKey: parentPage.PageKey,
	}), nil
}

func (s *service) resolveActiveMenuPath(
	page *models.UIPage,
	menuMap map[uuid.UUID]runtimeMenuNode,
	pageMap map[string]models.UIPage,
	seen map[string]struct{},
) string {
	if page == nil {
		return ""
	}
	if activePath := normalizeRoutePath(page.ActiveMenuPath); activePath != "" {
		return activePath
	}
	if page.ParentMenuID != nil {
		if node, ok := menuMap[*page.ParentMenuID]; ok {
			return node.FullPath
		}
	}
	parentKey := strings.TrimSpace(page.ParentPageKey)
	if parentKey == "" {
		return ""
	}
	if _, ok := seen[parentKey]; ok {
		return ""
	}
	parentPage, ok := pageMap[parentKey]
	if !ok {
		return ""
	}
	seen[parentKey] = struct{}{}
	defer delete(seen, parentKey)
	return s.resolveActiveMenuPath(&parentPage, menuMap, pageMap, seen)
}

func resolveMenuBreadcrumbChain(activePath string, menuMap map[uuid.UUID]runtimeMenuNode) []BreadcrumbPreviewItem {
	targetPath := normalizeRoutePath(activePath)
	if targetPath == "" {
		return []BreadcrumbPreviewItem{}
	}
	var target *runtimeMenuNode
	for _, node := range menuMap {
		if normalizeRoutePath(node.FullPath) == targetPath {
			item := node
			target = &item
			break
		}
	}
	if target == nil {
		return []BreadcrumbPreviewItem{}
	}

	chain := make([]BreadcrumbPreviewItem, 0, 4)
	current := target
	for current != nil {
		title := strings.TrimSpace(current.Menu.Title)
		if title == "" {
			title = strings.TrimSpace(current.Menu.Name)
		}
		chain = append(chain, BreadcrumbPreviewItem{
			Type:  "menu",
			Title: title,
			Path:  normalizeRoutePath(current.FullPath),
		})
		if current.Menu.ParentID == nil {
			break
		}
		parent, ok := menuMap[*current.Menu.ParentID]
		if !ok {
			break
		}
		parentCopy := parent
		current = &parentCopy
	}
	reverseBreadcrumbItems(chain)
	return chain
}

func reverseBreadcrumbItems(items []BreadcrumbPreviewItem) {
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
}
