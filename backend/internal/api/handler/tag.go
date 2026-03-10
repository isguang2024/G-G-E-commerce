package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签处理器
type TagHandler struct {
	// TODO: 注入 TagService
}

// NewTagHandler 创建标签处理器
func NewTagHandler() *TagHandler {
	return &TagHandler{}
}

// List 标签列表
func (h *TagHandler) List(c *gin.Context) {
	// TODO: 实现标签列表逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    []interface{}{},
	})
}

// Create 创建标签
func (h *TagHandler) Create(c *gin.Context) {
	// TODO: 实现创建标签逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": "placeholder"},
	})
}

// Update 更新标签
func (h *TagHandler) Update(c *gin.Context) {
	// TODO: 实现更新标签逻辑
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": id},
	})
}

// Delete 删除标签
func (h *TagHandler) Delete(c *gin.Context) {
	// TODO: 实现删除标签逻辑
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": id},
	})
}
