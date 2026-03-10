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
	Pages      []ViewPageItem `json:"pages"`
	Refreshed  bool           `json:"refreshed"`
	RefreshedAt string        `json:"refreshedAt"`
}

// SaveHiddenPageFileRequest 保存隐藏页面文件请求
type SaveHiddenPageFileRequest struct {
	FileName string `json:"fileName" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Path     string `json:"path" binding:"required"`
}

// SaveHiddenPageFile 保存页面关联文件（内部复用）
func (h *SystemHandler) SaveHiddenPageFile(c *gin.Context) {
	var req SaveHiddenPageFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "参数错误: "+err.Error())
		c.JSON(status, resp)
		return
	}

	// 验证文件路径安全性（防止路径遍历攻击）
	if !isValidHiddenPagePath(req.Path) {
		status, resp := errcode.ResponseWithMsg(errcode.ErrParamInvalid, "无效的文件路径")
		c.JSON(status, resp)
		return
	}

	// 获取项目根目录（相对于backend目录）
	// 假设backend目录在项目根目录下，需要向上找到项目根目录
	projectRoot, err := findProjectRoot()
	if err != nil {
		h.logger.Error("Failed to find project root", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "无法找到项目根目录")
		c.JSON(status, resp)
		return
	}

	// 构建完整文件路径
	fullPath := filepath.Join(projectRoot, req.Path)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		h.logger.Error("Failed to create directory", zap.Error(err), zap.String("dir", dir))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "创建目录失败")
		c.JSON(status, resp)
		return
	}

	// 写入文件
	if err := os.WriteFile(fullPath, []byte(req.Code), 0644); err != nil {
		h.logger.Error("Failed to write file", zap.Error(err), zap.String("path", fullPath))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "写入文件失败")
		c.JSON(status, resp)
		return
	}

	h.logger.Info("Hidden page file saved", zap.String("path", fullPath))

	c.JSON(http.StatusOK, dto.SuccessResponse(gin.H{
		"message": "文件保存成功",
		"path":    fullPath,
	}))
}

// SavePageAssociationFile 保存页面关联文件
func (h *SystemHandler) SavePageAssociationFile(c *gin.Context) {
	h.SaveHiddenPageFile(c)
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

// GetHiddenIndexFile 获取隐藏路由模块 index.ts 文件内容
func (h *SystemHandler) GetHiddenIndexFile(c *gin.Context) {
	projectRoot, err := findProjectRoot()
	if err != nil {
		h.logger.Error("Failed to find project root", zap.Error(err))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "无法找到项目根目录")
		c.JSON(status, resp)
		return
	}

	indexFilePath := filepath.Join(projectRoot, "frontend", "src", "router", "modules", "hidden", "index.ts")
	
	// 读取文件内容
	content, err := os.ReadFile(indexFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，返回空内容（前端会创建）
			c.JSON(http.StatusOK, dto.SuccessResponse(""))
			return
		}
		h.logger.Error("Failed to read index file", zap.Error(err), zap.String("path", indexFilePath))
		status, resp := errcode.ResponseWithMsg(errcode.ErrInternal, "读取文件失败")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse(string(content)))
}

// isValidHiddenPagePath 验证文件路径是否安全
func isValidHiddenPagePath(path string) bool {
	// 只允许保存到 frontend/src/router/modules/hidden/ 目录下
	allowedPrefix := "frontend/src/router/modules/hidden/"
	
	// 规范化路径（使用正斜杠）
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	
	// 检查是否以允许的前缀开头
	if !strings.HasPrefix(normalizedPath, allowedPrefix) {
		return false
	}

	// 检查是否包含路径遍历字符
	if strings.Contains(normalizedPath, "..") {
		return false
	}

	// 检查文件名是否合法
	fileName := filepath.Base(normalizedPath)
	
	// 允许 index.ts 文件
	if fileName == "index.ts" {
		return true
	}
	
	// 检查文件名是否以 _hidepage.ts 结尾
	if !strings.HasSuffix(normalizedPath, "_hidepage.ts") {
		return false
	}

	// 检查文件名是否符合命名规范
	if !isValidFileName(fileName) {
		return false
	}

	return true
}

// isValidFileName 验证文件名是否合法
func isValidFileName(fileName string) bool {
	// 文件名必须匹配模式: {module}_hidepage.ts
	// module 只能包含小写字母、数字、下划线
	validPatterns := []string{
		"team_hidepage.ts",
		"system_hidepage.ts",
	}
	
	for _, pattern := range validPatterns {
		if fileName == pattern {
			return true
		}
	}
	
	// 允许其他模块，但必须符合命名规范
	if strings.HasSuffix(fileName, "_hidepage.ts") {
		moduleName := strings.TrimSuffix(fileName, "_hidepage.ts")
		for _, char := range moduleName {
			if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '_') {
				return false
			}
		}
		return true
	}
	
	return false
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
