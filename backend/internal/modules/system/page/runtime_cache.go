package page

import (
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const runtimePageCacheTTL = 24 * time.Hour

type runtimePageCacheSnapshot struct {
	all       []Record
	public    []Record
	expiresAt time.Time
}

type runtimePageCacheStore struct {
	mu       sync.RWMutex
	snapshot *runtimePageCacheSnapshot
}

var globalRuntimePageCache runtimePageCacheStore

func InvalidateRuntimeCache() {
	globalRuntimePageCache.invalidate()
}

func (s *service) loadRuntimeRecords() ([]Record, error) {
	return globalRuntimePageCache.get(s, false)
}

func (s *service) loadPublicRuntimeRecords() ([]Record, error) {
	return globalRuntimePageCache.get(s, true)
}

func (c *runtimePageCacheStore) get(s *service, publicOnly bool) ([]Record, error) {
	now := time.Now()

	c.mu.RLock()
	snapshot := c.snapshot
	if snapshot != nil && now.Before(snapshot.expiresAt) {
		records := snapshot.all
		if publicOnly {
			records = snapshot.public
		}
		c.mu.RUnlock()
		return cloneRuntimeRecords(records), nil
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	snapshot = c.snapshot
	if snapshot != nil && now.Before(snapshot.expiresAt) {
		records := snapshot.all
		if publicOnly {
			records = snapshot.public
		}
		return cloneRuntimeRecords(records), nil
	}

	all, menuMap, err := s.buildRuntimeRecords()
	if err != nil {
		return nil, err
	}
	publicRecords := buildPublicRuntimeRecords(all, menuMap)
	c.snapshot = &runtimePageCacheSnapshot{
		all:       cloneRuntimeRecords(all),
		public:    cloneRuntimeRecords(publicRecords),
		expiresAt: now.Add(runtimePageCacheTTL),
	}

	if publicOnly {
		return cloneRuntimeRecords(publicRecords), nil
	}
	return cloneRuntimeRecords(all), nil
}

func (c *runtimePageCacheStore) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot = nil
}

func cloneRuntimeRecords(items []Record) []Record {
	if len(items) == 0 {
		return []Record{}
	}
	result := make([]Record, len(items))
	copy(result, items)
	return result
}

func buildPublicRuntimeRecords(
	all []Record,
	menuMap map[uuid.UUID]runtimeMenuNode,
) []Record {
	if len(all) == 0 {
		return []Record{}
	}

	pageMap := make(map[string]Record, len(all))
	for _, item := range all {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		pageMap[pageKey] = item
	}

	included := make(map[string]struct{}, len(all))
	accessCache := make(map[string]string, len(all))
	resolving := make(map[string]struct{})

	var resolveAccessMode func(pageKey string) string
	resolveAccessMode = func(pageKey string) string {
		normalizedKey := strings.TrimSpace(pageKey)
		if normalizedKey == "" {
			return "jwt"
		}
		if cached, ok := accessCache[normalizedKey]; ok {
			return cached
		}
		if _, ok := resolving[normalizedKey]; ok {
			return "jwt"
		}

		item, ok := pageMap[normalizedKey]
		if !ok {
			return "jwt"
		}

		resolving[normalizedKey] = struct{}{}
		defer delete(resolving, normalizedKey)

		mode := normalizeAccessMode(item.AccessMode)
		switch mode {
		case "public", "jwt", "permission":
		case "inherit":
			parentPageKey := strings.TrimSpace(item.ParentPageKey)
			if parentPageKey != "" {
				mode = resolveAccessMode(parentPageKey)
			} else {
				mode = resolveMenuAccessMode(item.ParentMenuID, menuMap)
			}
		default:
			mode = "jwt"
		}
		if mode == "" {
			mode = "jwt"
		}
		accessCache[normalizedKey] = mode
		return mode
	}

	var includeAncestors func(pageKey string)
	includeAncestors = func(pageKey string) {
		normalizedKey := strings.TrimSpace(pageKey)
		if normalizedKey == "" {
			return
		}
		if _, ok := included[normalizedKey]; ok {
			return
		}
		item, ok := pageMap[normalizedKey]
		if !ok {
			return
		}
		included[normalizedKey] = struct{}{}
		if parentPageKey := strings.TrimSpace(item.ParentPageKey); parentPageKey != "" {
			includeAncestors(parentPageKey)
		}
	}

	for _, item := range all {
		pageKey := strings.TrimSpace(item.PageKey)
		if pageKey == "" {
			continue
		}
		if resolveAccessMode(pageKey) == "public" {
			includeAncestors(pageKey)
		}
	}

	result := make([]Record, 0, len(included))
	for _, item := range all {
		if _, ok := included[strings.TrimSpace(item.PageKey)]; ok {
			result = append(result, item)
		}
	}
	return result
}

func resolveMenuAccessMode(
	parentMenuID *uuid.UUID,
	menuMap map[uuid.UUID]runtimeMenuNode,
) string {
	if parentMenuID == nil {
		return "jwt"
	}
	node, ok := menuMap[*parentMenuID]
	if !ok {
		return "jwt"
	}
	raw, ok := node.Menu.Meta["accessMode"]
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
