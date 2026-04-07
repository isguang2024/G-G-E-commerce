// facade.go — Phase 4 ogen migration helpers. Exposes the unexported
// fastEnterService, messageService and view-page enumeration through a thin
// Facade so internal/api/handlers can call them without re-implementing logic.
package system

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	cachepkg "github.com/gg-ecommerce/backend/internal/pkg/cache"
)

// Facade bundles the legacy unexported services for use by ogen handlers.
type Facade struct {
	logger     *zap.Logger
	cache      *cachepkg.Cache
	fastEnter  *fastEnterService
	messageSvc *messageService
}

// NewFacade builds a Facade. cache may be nil.
func NewFacade(db *gorm.DB, logger *zap.Logger, cache *cachepkg.Cache) *Facade {
	return &Facade{
		logger:     logger,
		cache:      cache,
		fastEnter:  NewFastEnterService(db),
		messageSvc: NewMessageService(db, logger),
	}
}

// ── Fast enter ─────────────────────────────────────────────────────────────

func (f *Facade) GetFastEnterConfig() (FastEnterConfig, error) {
	return f.fastEnter.GetConfig()
}

type FastEnterSaveRequestPublic struct {
	Applications []FastEnterApplication `json:"applications"`
	QuickLinks   []FastEnterQuickLink   `json:"quickLinks"`
	MinWidth     int                    `json:"minWidth"`
}

func (f *Facade) UpdateFastEnterConfig(req FastEnterSaveRequestPublic) (FastEnterConfig, error) {
	return f.fastEnter.SaveConfig(FastEnterConfig{
		Applications: req.Applications,
		QuickLinks:   req.QuickLinks,
		MinWidth:     req.MinWidth,
	})
}

// ── View pages ─────────────────────────────────────────────────────────────

func (f *Facade) GetViewPages(ctx context.Context, force bool) ([]ViewPageItem, bool, time.Time, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, false, time.Time{}, err
	}
	cacheKey := "system:view-pages:frontend-src-views"
	if f.cache != nil && !force {
		var cached ViewPagesResponse
		if err := f.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached.Pages, false, time.Now(), nil
		}
	}
	pages, err := enumerateViewPagesPublic(projectRoot)
	if err != nil {
		return nil, false, time.Time{}, err
	}
	now := time.Now()
	if f.cache != nil {
		resp := ViewPagesResponse{Pages: pages, Refreshed: true, RefreshedAt: now.Format(time.RFC3339)}
		if err := f.cache.Set(ctx, cacheKey, resp, 30*24*time.Hour); err != nil {
			f.logger.Warn("Failed to cache view pages", zap.Error(err))
		}
	}
	return pages, true, now, nil
}

func enumerateViewPagesPublic(projectRoot string) ([]ViewPageItem, error) {
	viewsDir := filepath.Join(projectRoot, "frontend", "src", "views")
	items := make([]ViewPageItem, 0, 256)
	err := filepath.WalkDir(viewsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".vue" {
			return nil
		}
		rel, err := filepath.Rel(projectRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		filePath := "/" + rel
		items = append(items, ViewPageItem{FilePath: filePath, ComponentPath: toComponentPath(filePath)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ComponentPath < items[j].ComponentPath })
	return items, nil
}
