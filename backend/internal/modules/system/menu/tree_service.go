package menu

import (
	"strings"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/user"
)

// GetTree returns the menu tree, optionally filtered to only menus visible
// to the provided allowedMenuIDs (plus any jwt/public menus).
func (s *menuService) GetTree(all bool, allowedMenuIDs []uuid.UUID, appKey, spaceKey string) ([]*user.Menu, error) {
	normalizedAppKey := normalizeMenuAppKey(appKey)
	normalizedSpaceKey := strings.TrimSpace(spaceKey)
	if normalizedSpaceKey != "" {
		normalizedSpaceKey = normalizeMenuSpaceKey(normalizedSpaceKey)
	}
	flat, err := s.menuRepo.ListByAppAndSpace(normalizedAppKey, normalizedSpaceKey)
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

// BuildMenuTree is the exported helper used by other packages.
func BuildMenuTree(flat []user.Menu, parentID *uuid.UUID) []*user.Menu {
	return user.BuildTree(flat, parentID)
}

// ---------------------------------------------------------------------------
// private tree helpers
// ---------------------------------------------------------------------------

func (s *menuService) mergeSystemMenuIDs(flat []user.Menu, allowed []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{})
	for _, id := range allowed {
		set[id] = struct{}{}
	}
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

