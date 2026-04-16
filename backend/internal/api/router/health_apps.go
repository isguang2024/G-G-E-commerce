package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

type appHealthItem struct {
	AppKey           string `json:"app_key"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	AuthMode         string `json:"auth_mode"`
	FrontendEntryURL string `json:"frontend_entry_url"`
	BackendEntryURL  string `json:"backend_entry_url"`
	HealthCheckURL   string `json:"health_check_url"`
	HasCapabilities  bool   `json:"has_capabilities"`
	Healthy          bool   `json:"healthy"`
}

func registerAppHealthRoutes(r *gin.Engine, db *gorm.DB, logger *zap.Logger) {
	if r == nil {
		return
	}
	r.GET("/health/apps", func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "degraded",
				"message": "database unavailable",
			})
			return
		}
		var apps []models.App
		if err := db.Select("app_key", "name", "status", "auth_mode", "frontend_entry_url", "backend_entry_url", "health_check_url", "capabilities").
			Where("deleted_at IS NULL").
			Order("app_key ASC").
			Find(&apps).Error; err != nil {
			if logger != nil {
				logger.Error("load app registry failed", zap.Error(err))
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "load app registry failed",
			})
			return
		}

		items := make([]appHealthItem, 0, len(apps))
		healthyCount := 0
		for _, app := range apps {
			item := appHealthItem{
				AppKey:           strings.TrimSpace(app.AppKey),
				Name:             strings.TrimSpace(app.Name),
				Status:           strings.TrimSpace(app.Status),
				AuthMode:         strings.TrimSpace(app.AuthMode),
				FrontendEntryURL: strings.TrimSpace(app.FrontendEntryURL),
				BackendEntryURL:  strings.TrimSpace(app.BackendEntryURL),
				HealthCheckURL:   strings.TrimSpace(app.HealthCheckURL),
				HasCapabilities:  len(app.Capabilities) > 0,
			}
			item.Healthy = item.Status != "disabled" && item.HealthCheckURL != ""
			if item.Healthy {
				healthyCount++
			}
			items = append(items, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"registry": gin.H{
				"total":         len(items),
				"healthy_count": healthyCount,
			},
			"apps": items,
		})
	})
}

