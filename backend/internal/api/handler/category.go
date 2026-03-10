package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 分类处理器
type CategoryHandler struct {
	// TODO: 注入 CategoryService
}

// NewCategoryHandler 创建分类处理器
func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

// GetTree 获取分类树
func (h *CategoryHandler) GetTree(c *gin.Context) {
	// TODO: 实现分类树逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    []interface{}{},
	})
}

// List 分类列表
func (h *CategoryHandler) List(c *gin.Context) {
	// TODO: 实现分类列表逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    []interface{}{},
	})
}

// Get 分类详情
func (h *CategoryHandler) Get(c *gin.Context) {
	// TODO: 实现分类详情逻辑
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": id},
	})
}

// Create 创建分类
func (h *CategoryHandler) Create(c *gin.Context) {
	// TODO: 实现创建分类逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": "placeholder"},
	})
}

// Update 更新分类
func (h *CategoryHandler) Update(c *gin.Context) {
	// TODO: 实现更新分类逻辑
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": id},
	})
}

// Delete 删除分类
func (h *CategoryHandler) Delete(c *gin.Context) {
	// TODO: 实现删除分类逻辑
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": id},
	})
}
