// facade.go keeps the remaining message facade plus exported view-page service
// helpers used by ogen handlers.
package system

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	cachepkg "github.com/maben/backend/internal/pkg/cache"
)

// Facade currently only wraps the message service while the remaining
// Phase 4 leftovers have been promoted to explicit services.
type Facade struct {
	messageSvc *messageService
}

// NewFacade builds a Facade. cache may be nil.
func NewFacade(db *gorm.DB, logger *zap.Logger, cache *cachepkg.Cache) *Facade {
	_ = cache
	return &Facade{
		messageSvc: NewMessageService(db, logger),
	}
}

type ViewPagesService interface {
	GetPages(ctx context.Context, force bool) ([]ViewPageItem, bool, time.Time, error)
}

type viewPagesService struct {
	logger *zap.Logger
	cache  *cachepkg.Cache
}

func NewViewPagesService(logger *zap.Logger, cache *cachepkg.Cache) ViewPagesService {
	return &viewPagesService{logger: logger, cache: cache}
}

func (s *viewPagesService) GetPages(ctx context.Context, force bool) ([]ViewPageItem, bool, time.Time, error) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, false, time.Time{}, err
	}
	cacheKey := "system:view-pages:frontend-src-views"
	if s.cache != nil && !force {
		var cached ViewPagesResponse
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached.Pages, false, time.Now(), nil
		}
	}
	pages, err := enumerateViewPagesPublic(projectRoot)
	if err != nil {
		return nil, false, time.Time{}, err
	}
	now := time.Now()
	if s.cache != nil {
		resp := ViewPagesResponse{Pages: pages, Refreshed: true, RefreshedAt: now.Format(time.RFC3339)}
		if err := s.cache.Set(ctx, cacheKey, resp, 30*24*time.Hour); err != nil {
			s.logger.Warn("Failed to cache view pages", zap.Error(err))
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

// ── Shared types (moved from handler.go) ───────────────────────────────────

type ViewPageItem struct {
	FilePath      string `json:"file_path"`
	ComponentPath string `json:"component_path"`
}

type ViewPagesResponse struct {
	Pages       []ViewPageItem `json:"pages"`
	Refreshed   bool           `json:"refreshed"`
	RefreshedAt string         `json:"refreshedAt"`
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

