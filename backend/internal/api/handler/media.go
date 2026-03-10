package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MediaHandler 媒体处理器
type MediaHandler struct {
	// TODO: 注入 MediaService
}

// NewMediaHandler 创建媒体处理器
func NewMediaHandler() *MediaHandler {
	return &MediaHandler{}
}

// Upload 上传媒体文件
func (h *MediaHandler) Upload(c *gin.Context) {
	// TODO: 实现上传逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"url": "placeholder"},
	})
}

// List 媒体列表
func (h *MediaHandler) List(c *gin.Context) {
	// TODO: 实现媒体列表逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    []interface{}{},
	})
}

// Delete 删除媒体
func (h *MediaHandler) Delete(c *gin.Context) {
	// TODO: 实现删除媒体逻辑
	id := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    gin.H{"id": id},
	})
}
