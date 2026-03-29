package system

import (
	"strings"

	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const fastEnterSettingKey = "ui.fast_enter"

type FastEnterApplication struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IconColor   string `json:"iconColor"`
	Enabled     bool   `json:"enabled"`
	Order       int    `json:"order"`
	RouteName   string `json:"routeName,omitempty"`
	Link        string `json:"link,omitempty"`
}

type FastEnterQuickLink struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	Order     int    `json:"order"`
	RouteName string `json:"routeName,omitempty"`
	Link      string `json:"link,omitempty"`
}

type FastEnterConfig struct {
	Applications []FastEnterApplication `json:"applications"`
	QuickLinks   []FastEnterQuickLink   `json:"quickLinks"`
	MinWidth     int                    `json:"minWidth"`
}

type fastEnterService struct {
	db *gorm.DB
}

func NewFastEnterService(db *gorm.DB) *fastEnterService {
	return &fastEnterService{db: db}
}

func (s *fastEnterService) GetConfig() (FastEnterConfig, error) {
	var setting models.SystemSetting
	err := s.db.Where("key = ?", fastEnterSettingKey).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			defaultConfig := defaultFastEnterConfig()
			return defaultConfig, s.saveConfig(defaultConfig)
		}
		return FastEnterConfig{}, err
	}

	return normalizeFastEnterConfig(setting.Value), nil
}

func (s *fastEnterService) SaveConfig(config FastEnterConfig) (FastEnterConfig, error) {
	normalized := normalizeFastEnterConfig(models.MetaJSON{
		"applications": config.Applications,
		"quickLinks":   config.QuickLinks,
		"minWidth":     config.MinWidth,
	})
	return normalized, s.saveConfig(normalized)
}

func (s *fastEnterService) saveConfig(config FastEnterConfig) error {
	payload := models.MetaJSON{
		"applications": config.Applications,
		"quickLinks":   config.QuickLinks,
		"minWidth":     config.MinWidth,
	}

	var setting models.SystemSetting
	err := s.db.Where("key = ?", fastEnterSettingKey).First(&setting).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			setting = models.SystemSetting{
				Key:    fastEnterSettingKey,
				Value:  payload,
				Status: "normal",
			}
			return s.db.Create(&setting).Error
		}
		return err
	}

	return s.db.Model(&setting).Updates(map[string]interface{}{
		"value":  payload,
		"status": "normal",
	}).Error
}

func defaultFastEnterConfig() FastEnterConfig {
	return FastEnterConfig{
		MinWidth: 1200,
		Applications: []FastEnterApplication{
			{ID: "console", Name: "工作台", Description: "系统概览与数据统计", Icon: "ri:pie-chart-line", IconColor: "#377dff", Enabled: true, Order: 1, RouteName: "Console"},
			{ID: "role", Name: "角色管理", Description: "维护平台角色与角色权限", Icon: "ri:shield-user-line", IconColor: "#0f766e", Enabled: true, Order: 2, RouteName: "Role"},
			{ID: "user", Name: "用户管理", Description: "查看平台账号、角色归属和权限诊断", Icon: "ri:user-settings-line", IconColor: "#2563eb", Enabled: true, Order: 3, RouteName: "User"},
			{ID: "menu", Name: "菜单管理", Description: "维护菜单树、菜单分组和备份", Icon: "ri:menu-line", IconColor: "#f97316", Enabled: true, Order: 4, RouteName: "Menus"},
			{ID: "page", Name: "页面管理", Description: "维护页面注册表和运行时页面", Icon: "ri:layout-4-line", IconColor: "#7c3aed", Enabled: true, Order: 5, RouteName: "PageManagement"},
			{ID: "api-endpoint", Name: "API 管理", Description: "同步 API 注册表与诊断未注册接口", Icon: "ri:route-line", IconColor: "#dc2626", Enabled: true, Order: 6, RouteName: "ApiEndpoint"},
		},
		QuickLinks: []FastEnterQuickLink{
			{ID: "user-center", Name: "个人中心", Enabled: true, Order: 1, RouteName: "UserCenter"},
			{ID: "team-members", Name: "团队成员", Enabled: true, Order: 2, RouteName: "TeamMembers"},
			{ID: "feature-package", Name: "功能包管理", Enabled: true, Order: 3, RouteName: "FeaturePackage"},
		},
	}
}

func normalizeFastEnterConfig(raw models.MetaJSON) FastEnterConfig {
	defaultConfig := defaultFastEnterConfig()
	result := FastEnterConfig{
		Applications: normalizeFastEnterApplications(raw["applications"], defaultConfig.Applications),
		QuickLinks:   normalizeFastEnterQuickLinks(raw["quickLinks"], defaultConfig.QuickLinks),
		MinWidth:     normalizeFastEnterMinWidth(raw["minWidth"], defaultConfig.MinWidth),
	}
	if len(result.Applications) == 0 {
		result.Applications = defaultConfig.Applications
	}
	if len(result.QuickLinks) == 0 {
		result.QuickLinks = defaultConfig.QuickLinks
	}
	return result
}

func normalizeFastEnterMinWidth(raw interface{}, fallback int) int {
	switch value := raw.(type) {
	case int:
		if value >= 960 && value <= 2400 {
			return value
		}
	case float64:
		target := int(value)
		if target >= 960 && target <= 2400 {
			return target
		}
	}
	return fallback
}

func normalizeFastEnterApplications(raw interface{}, fallback []FastEnterApplication) []FastEnterApplication {
	switch typed := raw.(type) {
	case []FastEnterApplication:
		return append([]FastEnterApplication(nil), typed...)
	case []models.MetaJSON:
		items := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			items = append(items, item)
		}
		raw = items
	}
	items, ok := raw.([]interface{})
	if !ok {
		return append([]FastEnterApplication(nil), fallback...)
	}
	result := make([]FastEnterApplication, 0, len(items))
	for index, item := range items {
		record, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := strings.TrimSpace(toString(record["name"]))
		if name == "" {
			continue
		}
		id := strings.TrimSpace(toString(record["id"]))
		if id == "" {
			id = "app-" + strings.TrimSpace(name)
		}
		result = append(result, FastEnterApplication{
			ID:          id,
			Name:        name,
			Description: strings.TrimSpace(toString(record["description"])),
			Icon:        fallbackString(record["icon"], "ri:apps-2-line"),
			IconColor:   fallbackString(record["iconColor"], "#377dff"),
			Enabled:     toBoolDefault(record["enabled"], true),
			Order:       toIntDefault(record["order"], index+1),
			RouteName:   strings.TrimSpace(toString(record["routeName"])),
			Link:        strings.TrimSpace(toString(record["link"])),
		})
	}
	return result
}

func normalizeFastEnterQuickLinks(raw interface{}, fallback []FastEnterQuickLink) []FastEnterQuickLink {
	switch typed := raw.(type) {
	case []FastEnterQuickLink:
		return append([]FastEnterQuickLink(nil), typed...)
	case []models.MetaJSON:
		items := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			items = append(items, item)
		}
		raw = items
	}
	items, ok := raw.([]interface{})
	if !ok {
		return append([]FastEnterQuickLink(nil), fallback...)
	}
	result := make([]FastEnterQuickLink, 0, len(items))
	for index, item := range items {
		record, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		name := strings.TrimSpace(toString(record["name"]))
		if name == "" {
			continue
		}
		id := strings.TrimSpace(toString(record["id"]))
		if id == "" {
			id = "link-" + strings.TrimSpace(name)
		}
		result = append(result, FastEnterQuickLink{
			ID:        id,
			Name:      name,
			Enabled:   toBoolDefault(record["enabled"], true),
			Order:     toIntDefault(record["order"], index+1),
			RouteName: strings.TrimSpace(toString(record["routeName"])),
			Link:      strings.TrimSpace(toString(record["link"])),
		})
	}
	return result
}

func toString(value interface{}) string {
	switch target := value.(type) {
	case string:
		return target
	default:
		return ""
	}
}

func fallbackString(value interface{}, fallback string) string {
	target := strings.TrimSpace(toString(value))
	if target == "" {
		return fallback
	}
	return target
}

func toBoolDefault(value interface{}, fallback bool) bool {
	target, ok := value.(bool)
	if !ok {
		return fallback
	}
	return target
}

func toIntDefault(value interface{}, fallback int) int {
	switch target := value.(type) {
	case int:
		if target > 0 {
			return target
		}
	case float64:
		if int(target) > 0 {
			return int(target)
		}
	}
	return fallback
}
