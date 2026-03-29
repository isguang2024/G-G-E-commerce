package system

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

type SystemHandler struct {
	logger           *zap.Logger
	cache            *cachepkg.Cache
	fastEnterService *fastEnterService
	messageService   *messageService
}

func NewSystemHandler(logger *zap.Logger, cache *cachepkg.Cache, fastEnterService *fastEnterService, messageService *messageService) *SystemHandler {
	return &SystemHandler{
		logger:           logger,
		cache:            cache,
		fastEnterService: fastEnterService,
		messageService:   messageService,
	}
}

type fastEnterSaveRequest struct {
	Applications []FastEnterApplication `json:"applications"`
	QuickLinks   []FastEnterQuickLink   `json:"quickLinks"`
	MinWidth     int                    `json:"minWidth"`
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

	if h.cache != nil {
		if err := h.cache.Set(ctx, cacheKey, resp, 30*24*time.Hour); err != nil {
			h.logger.Warn("Failed to cache view pages", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(resp))
}

func (h *SystemHandler) GetFastEnterConfig(c *gin.Context) {
	config, err := h.fastEnterService.GetConfig()
	if err != nil {
		h.logger.Error("Failed to get fast enter config", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "获取快捷入口配置失败")
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, dto.SuccessResponse(config))
}

func (h *SystemHandler) UpdateFastEnterConfig(c *gin.Context) {
	var req fastEnterSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.Response(errcode.ErrParamInvalid)
		c.JSON(status, resp)
		return
	}

	config, err := h.fastEnterService.SaveConfig(FastEnterConfig{
		Applications: req.Applications,
		QuickLinks:   req.QuickLinks,
		MinWidth:     req.MinWidth,
	})
	if err != nil {
		h.logger.Error("Failed to update fast enter config", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "保存快捷入口配置失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(config))
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
		rel = filepath.ToSlash(rel)

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
