package handler

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/api/errcode"
	cachepkg "github.com/gg-ecommerce/backend/internal/pkg/cache"
)

// SystemHandler 系统管理处理器
type SystemHandler struct {
	logger *zap.Logger
	cache  *cachepkg.Cache
}

// NewSystemHandler 创建系统管理处理器
func NewSystemHandler(logger *zap.Logger, cache *cachepkg.Cache) *SystemHandler {
	return &SystemHandler{
		logger: logger,
		cache:  cache,
	}
}

type ViewPageItem struct {
	FilePath      string `json:"filePath"`
	ComponentPath string `json:"componentPath"`
}

type ViewPagesResponse struct {
	Pages       []ViewPageItem `json:"pages"`
	Refreshed   bool           `json:"refreshed"`
	RefreshedAt string         `json:"refreshedAt"`
}

// GetViewPages 枚举 frontend/src/views 下的 .vue 页面文件（支持 Redis 缓存）
func (h *SystemHandler) GetViewPages(c *gin.Context) {
	forceRefresh := c.Query("force") == "1" || strings.ToLower(c.Query("force")) == "true"

	projectRoot, err := findProjectRoot()
	if err != nil {
		h.logger.Error("Failed to find project root", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "无法找到项目根目录")
		c.JSON(status, resp)
		return
	}

	cacheKey := "system:view-pages:frontend-src-views"
	ctx := context.Background()

	// 优先读缓存（除非强制刷新）
	if h.cache != nil && !forceRefresh {
		var cached ViewPagesResponse
		if err := h.cache.Get(ctx, cacheKey, &cached); err == nil {
			cached.Refreshed = false
			c.JSON(http.StatusOK, dto.SuccessResponse(cached))
			return
		}
	}

	pages, err := enumerateViewPages(projectRoot)
	if err != nil {
		h.logger.Error("Failed to enumerate view pages", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "枚举页面文件失败")
		c.JSON(status, resp)
		return
	}

	resp := ViewPagesResponse{
		Pages:       pages,
		Refreshed:   true,
		RefreshedAt: time.Now().Format(time.RFC3339),
	}

	// 写入缓存，TTL 1 个月
	if h.cache != nil {
		if err := h.cache.Set(ctx, cacheKey, resp, 30*24*time.Hour); err != nil {
			h.logger.Warn("Failed to cache view pages", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(resp))
}

// findProjectRoot 查找项目根目录
func findProjectRoot() (string, error) {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 从当前目录向上查找，直到找到包含 frontend 和 backend 目录的目录
	current := wd
	for {
		// 检查是否存在 frontend 和 backend 目录
		frontendPath := filepath.Join(current, "frontend")
		backendPath := filepath.Join(current, "backend")

		if _, err := os.Stat(frontendPath); err == nil {
			if _, err := os.Stat(backendPath); err == nil {
				return current, nil
			}
		}

		// 向上查找
		parent := filepath.Dir(current)
		if parent == current {
			// 已经到达根目录
			break
		}
		current = parent
	}

	// 如果找不到，返回当前目录（可能是开发环境）
	return wd, nil
}

func enumerateViewPages(projectRoot string) ([]ViewPageItem, error) {
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
		rel = filepath.ToSlash(rel) // frontend/src/views/team/team-roles/index.vue

		filePath := "/" + rel
		componentPath := toComponentPath(filePath)

		items = append(items, ViewPageItem{
			FilePath:      filePath,
			ComponentPath: componentPath,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].ComponentPath < items[j].ComponentPath
	})
	return items, nil
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
