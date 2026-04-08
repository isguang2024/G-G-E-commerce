package app

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/pathmatch"
)

type AppRecord struct {
	models.App
	HostCount    int      `json:"host_count"`
	SpaceCount   int      `json:"space_count"`
	MenuCount    int      `json:"menu_count"`
	PageCount    int      `json:"page_count"`
	PrimaryHosts []string `json:"primary_hosts,omitempty"`
}

type HostBindingRecord struct {
	models.AppHostBinding
	AppName string `json:"app_name"`
}

type CurrentResponse struct {
	App         AppRecord          `json:"app"`
	Binding     *HostBindingRecord `json:"binding,omitempty"`
	ResolvedBy  string             `json:"resolved_by"`
	RequestHost string             `json:"request_host"`
}

type SaveAppRequest struct {
	AppKey          string                 `json:"app_key"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	SpaceMode       string                 `json:"space_mode"`
	DefaultSpaceKey string                 `json:"default_space_key"`
	AuthMode        string                 `json:"auth_mode"`
	Status          string                 `json:"status"`
	IsDefault       bool                   `json:"is_default"`
	Meta            map[string]interface{} `json:"meta"`
}

type SaveHostBindingRequest struct {
	ID              string                 `json:"id"`
	AppKey          string                 `json:"app_key"`
	MatchType       string                 `json:"match_type"`
	Host            string                 `json:"host"`
	PathPattern     string                 `json:"path_pattern"`
	Priority        int                    `json:"priority"`
	Description     string                 `json:"description"`
	IsPrimary       bool                   `json:"is_primary"`
	DefaultSpaceKey string                 `json:"default_space_key"`
	Status          string                 `json:"status"`
	Meta            map[string]interface{} `json:"meta"`
}

type MenuSpaceEntryBindingRecord struct {
	models.MenuSpaceEntryBinding
	AppName   string `json:"app_name"`
	SpaceName string `json:"space_name"`
}

type SaveMenuSpaceEntryBindingRequest struct {
	ID          string                 `json:"id"`
	AppKey      string                 `json:"app_key"`
	SpaceKey    string                 `json:"space_key"`
	MatchType   string                 `json:"match_type"`
	Host        string                 `json:"host"`
	PathPattern string                 `json:"path_pattern"`
	Priority    int                    `json:"priority"`
	Description string                 `json:"description"`
	IsPrimary   bool                   `json:"is_primary"`
	Status      string                 `json:"status"`
	Meta        map[string]interface{} `json:"meta"`
}

type Service interface {
	ListApps() ([]AppRecord, error)
	GetCurrent(host, requestedAppKey string) (*CurrentResponse, error)
	SaveApp(req *SaveAppRequest) (*AppRecord, error)
	ListHostBindings(appKey string) ([]HostBindingRecord, error)
	SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error)
	DeleteHostBinding(appKey, id string) error
	ListMenuSpaceEntryBindings(appKey string) ([]MenuSpaceEntryBindingRecord, error)
	SaveMenuSpaceEntryBinding(appKey string, req *SaveMenuSpaceEntryBindingRequest) (*MenuSpaceEntryBindingRecord, error)
	DeleteMenuSpaceEntryBinding(appKey, id string) error
}

type service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func NormalizeAppKey(value string) string {
	return appctx.NormalizeAppKey(value)
}

func normalizeAppSpaceMode(value string) string {
	switch strings.TrimSpace(value) {
	case "multiple", "multi":
		return "multi"
	default:
		return "single"
	}
}

func RequestAppKey(c *gin.Context) string {
	return appctx.RequestAppKey(c)
}

func CurrentAppKey(c *gin.Context) string {
	return appctx.CurrentAppKey(c)
}

// ResolveAppByHost 兼容旧签名（仅 host），等价于 ResolveAppEntry(db, host, "", requestedAppKey)。
func ResolveAppByHost(db *gorm.DB, host string, requestedAppKey string) (string, *models.AppHostBinding, string, error) {
	return ResolveAppEntry(db, host, "", requestedAppKey)
}

// ResolveAppEntry Level 1 入口解析：按 host + path 匹配 APP。
func ResolveAppEntry(db *gorm.DB, host, path, requestedAppKey string) (string, *models.AppHostBinding, string, error) {
	if db == nil {
		return models.DefaultAppKey, nil, "fallback_default", nil
	}
	if err := ensureDefaultApp(db); err != nil {
		return models.DefaultAppKey, nil, "", err
	}

	explicit := NormalizeAppKey(requestedAppKey)
	if explicit != "" {
		ok, err := appExists(db, explicit)
		if err != nil {
			return models.DefaultAppKey, nil, "", err
		}
		if ok {
			return explicit, nil, "explicit", nil
		}
	}

	normalizedHost := pathmatch.NormalizeHost(host)
	normalizedPath := pathmatch.NormalizePath(path)

	// 加载所有启用绑定，按具体度排序。
	var bindings []models.AppHostBinding
	if err := db.Where("status = ? AND deleted_at IS NULL", "normal").Find(&bindings).Error; err != nil {
		return models.DefaultAppKey, nil, "", err
	}
	matched := matchAppEntryBinding(bindings, normalizedHost, normalizedPath)
	if matched != nil {
		ok, appErr := appExists(db, matched.AppKey)
		if appErr != nil {
			return models.DefaultAppKey, nil, "", appErr
		}
		if ok {
			return NormalizeAppKey(matched.AppKey), matched, "entry_binding", nil
		}
	}

	// 兼容旧菜单空间 Host 绑定。
	if normalizedHost != "" {
		var legacyBinding models.MenuSpaceHostBinding
		err := db.Where("host = ? AND status = ? AND deleted_at IS NULL", normalizedHost, "normal").First(&legacyBinding).Error
		if err == nil {
			var space models.MenuSpace
			spaceErr := db.Where("space_key = ? AND deleted_at IS NULL", legacyBinding.SpaceKey).Order("is_default DESC, created_at ASC").First(&space).Error
			if spaceErr == nil {
				appKey := NormalizeAppKey(space.AppKey)
				if appKey == "" {
					appKey = models.DefaultAppKey
				}
				ok, appErr := appExists(db, appKey)
				if appErr != nil {
					return models.DefaultAppKey, nil, "", appErr
				}
				if ok {
					return appKey, nil, "legacy_space_host_binding", nil
				}
			} else if spaceErr != nil && !errors.Is(spaceErr, gorm.ErrRecordNotFound) {
				return models.DefaultAppKey, nil, "", spaceErr
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DefaultAppKey, nil, "", err
		}
	}

	defaultApp, err := loadDefaultApp(db)
	if err != nil {
		return models.DefaultAppKey, nil, "", err
	}
	return defaultApp.AppKey, nil, "default_app", nil
}

// ResolveMenuSpaceEntry Level 2 入口解析：按 host + path 在 App 内匹配菜单空间。
// 单空间 App 直接返回 App 默认空间，不做匹配。
func ResolveMenuSpaceEntry(db *gorm.DB, appKey, host, path string) (string, string, error) {
	if db == nil {
		return "", "fallback_default", nil
	}
	normalizedAppKey := NormalizeAppKey(appKey)
	if normalizedAppKey == "" {
		return "", "fallback_default", nil
	}
	var app models.App
	if err := db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "fallback_default", nil
		}
		return "", "", err
	}
	defaultSpace := strings.TrimSpace(app.DefaultSpaceKey)
	if app.SpaceMode != "multi" {
		// 单空间短路。
		return defaultSpace, "single_space_app", nil
	}

	normalizedHost := pathmatch.NormalizeHost(host)
	normalizedPath := pathmatch.NormalizePath(path)

	var bindings []models.MenuSpaceEntryBinding
	if err := db.Where("app_key = ? AND status = ? AND deleted_at IS NULL", normalizedAppKey, "normal").Find(&bindings).Error; err != nil {
		return defaultSpace, "", err
	}
	matched := matchMenuSpaceEntryBinding(bindings, normalizedHost, normalizedPath)
	if matched != nil {
		spaceKey := strings.TrimSpace(matched.SpaceKey)
		// P2: 校验目标空间存在且未被禁用，避免遗留绑定指向失效空间。
		var space models.MenuSpace
		err := db.Select("space_key", "status").
			Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", normalizedAppKey, spaceKey).
			First(&space).Error
		if err == nil && space.Status != "disabled" {
			return spaceKey, "entry_binding", nil
		}
		// 空间不存在或已禁用，回退默认空间。
	}
	return defaultSpace, "default_space", nil
}

// matchAppEntryBinding 在所有候选 binding 中按具体度排序，返回第一个命中的。
func matchAppEntryBinding(bindings []models.AppHostBinding, host, path string) *models.AppHostBinding {
	type scored struct {
		idx   int
		score int
	}
	candidates := make([]scored, 0, len(bindings))
	for i := range bindings {
		b := bindings[i]
		hostPattern := pathmatch.NormalizeHostPattern(b.MatchType, b.Host)
		if !pathmatch.MatchHost(b.MatchType, hostPattern, host) {
			continue
		}
		needPath := b.MatchType == pathmatch.PathPrefix || b.MatchType == pathmatch.HostAndPath
		if needPath && !pathmatch.MatchPath(b.PathPattern, path) {
			continue
		}
		s := pathmatch.PatternSpecificity(b.MatchType, hostPattern, b.PathPattern) + b.Priority*10
		candidates = append(candidates, scored{idx: i, score: s})
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	winner := bindings[candidates[0].idx]
	return &winner
}

func matchMenuSpaceEntryBinding(bindings []models.MenuSpaceEntryBinding, host, path string) *models.MenuSpaceEntryBinding {
	type scored struct {
		idx   int
		score int
	}
	candidates := make([]scored, 0, len(bindings))
	for i := range bindings {
		b := bindings[i]
		hostPattern := pathmatch.NormalizeHostPattern(b.MatchType, b.Host)
		if !pathmatch.MatchHost(b.MatchType, hostPattern, host) {
			continue
		}
		needPath := b.MatchType == pathmatch.PathPrefix || b.MatchType == pathmatch.HostAndPath
		if needPath && !pathmatch.MatchPath(b.PathPattern, path) {
			continue
		}
		s := pathmatch.PatternSpecificity(b.MatchType, hostPattern, b.PathPattern) + b.Priority*10
		candidates = append(candidates, scored{idx: i, score: s})
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	winner := bindings[candidates[0].idx]
	return &winner
}

func (s *service) ListApps() ([]AppRecord, error) {
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	var apps []models.App
	if err := s.db.Where("deleted_at IS NULL").Order("is_default DESC, created_at ASC").Find(&apps).Error; err != nil {
		return nil, err
	}

	hostCounts, primaryHosts, err := loadStringCountAndHosts(s.db, &models.AppHostBinding{}, "app_key", "host", "is_primary = true")
	if err != nil {
		return nil, err
	}
	spaceCounts, err := loadStringCounts(s.db, &models.MenuSpace{}, "app_key")
	if err != nil {
		return nil, err
	}
	menuCounts, err := loadStringCounts(s.db, &models.MenuDefinition{}, "app_key")
	if err != nil {
		return nil, err
	}
	pageCounts, err := loadStringCounts(s.db, &models.UIPage{}, "app_key")
	if err != nil {
		return nil, err
	}

	records := make([]AppRecord, 0, len(apps))
	for _, item := range apps {
		key := NormalizeAppKey(item.AppKey)
		records = append(records, AppRecord{
			App:          item,
			HostCount:    hostCounts[key],
			SpaceCount:   spaceCounts[key],
			MenuCount:    menuCounts[key],
			PageCount:    pageCounts[key],
			PrimaryHosts: primaryHosts[key],
		})
	}
	return records, nil
}

func (s *service) GetCurrent(host, requestedAppKey string) (*CurrentResponse, error) {
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	appKey, binding, resolvedBy, err := ResolveAppByHost(s.db, host, requestedAppKey)
	if err != nil {
		return nil, err
	}
	record, err := s.getAppRecord(appKey)
	if err != nil {
		return nil, err
	}
	var hostBindingRecord *HostBindingRecord
	if binding != nil {
		hostBindingRecord = &HostBindingRecord{
			AppHostBinding: *binding,
			AppName:        record.Name,
		}
	}
	return &CurrentResponse{
		App:         *record,
		Binding:     hostBindingRecord,
		ResolvedBy:  resolvedBy,
		RequestHost: spacepkg.NormalizeHost(host),
	}, nil
}

func (s *service) SaveApp(req *SaveAppRequest) (*AppRecord, error) {
	if req == nil {
		return nil, errors.New("应用参数不能为空")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	appKey := NormalizeAppKey(req.AppKey)
	if appKey == "" {
		return nil, errors.New("应用标识不能为空")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("应用名称不能为空")
	}
	defaultSpaceKey := spacepkg.NormalizeSpaceKey(req.DefaultSpaceKey)
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	authMode := strings.TrimSpace(req.AuthMode)
	if authMode == "" {
		authMode = "inherit_host"
	}

	payload := models.App{
		AppKey:          appKey,
		Name:            name,
		Description:     strings.TrimSpace(req.Description),
		SpaceMode:       normalizeAppSpaceMode(req.SpaceMode),
		DefaultSpaceKey: defaultSpaceKey,
		AuthMode:        authMode,
		Status:          status,
		IsDefault:       req.IsDefault || appKey == models.DefaultAppKey,
		Meta:            models.MetaJSON(req.Meta),
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing models.App
		err := tx.Where("app_key = ? AND deleted_at IS NULL", appKey).First(&existing).Error
		switch {
		case err == nil:
			if payload.DefaultSpaceKey == "" {
				payload.DefaultSpaceKey = spacepkg.NormalizeSpaceKey(existing.DefaultSpaceKey)
			}
			if payload.DefaultSpaceKey == "" {
				payload.DefaultSpaceKey = models.DefaultMenuSpaceKey
			}
			if err := ensureSpaceExistsForApp(tx, appKey, payload.DefaultSpaceKey); err != nil {
				return err
			}
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"name":              payload.Name,
				"description":       payload.Description,
				"space_mode":        payload.SpaceMode,
				"default_space_key": payload.DefaultSpaceKey,
				"auth_mode":         payload.AuthMode,
				"status":            payload.Status,
				"is_default":        payload.IsDefault,
				"meta":              payload.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			payload.DefaultSpaceKey = models.DefaultMenuSpaceKey
			if createErr := tx.Create(&payload).Error; createErr != nil {
				return createErr
			}
		default:
			return err
		}
		if err := spacepkg.EnsureDefaultMenuSpace(tx, appKey); err != nil {
			return err
		}
		if payload.DefaultSpaceKey == models.DefaultMenuSpaceKey {
			if err := tx.Model(&models.MenuSpace{}).
				Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", appKey, models.DefaultMenuSpaceKey).
				Updates(map[string]interface{}{
					"is_default": true,
					"status":     "normal",
				}).Error; err != nil {
				return err
			}
		}
		if payload.IsDefault {
			if err := tx.Model(&models.App{}).
				Where("app_key <> ? AND deleted_at IS NULL", appKey).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.getAppRecord(appKey)
}

func (s *service) ListHostBindings(appKey string) ([]HostBindingRecord, error) {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	query := s.db.Model(&models.AppHostBinding{}).Where("deleted_at IS NULL").Where("app_key = ?", normalizedAppKey)
	var items []models.AppHostBinding
	if err := query.Order("is_primary DESC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	apps, err := s.ListApps()
	if err != nil {
		return nil, err
	}
	appNames := make(map[string]string, len(apps))
	for _, item := range apps {
		appNames[NormalizeAppKey(item.AppKey)] = item.Name
	}
	records := make([]HostBindingRecord, 0, len(items))
	for _, item := range items {
		records = append(records, HostBindingRecord{
			AppHostBinding: item,
			AppName:        appNames[NormalizeAppKey(item.AppKey)],
		})
	}
	return records, nil
}

func normalizeMatchType(value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		v = pathmatch.HostExact
	}
	switch v {
	case pathmatch.HostExact, pathmatch.HostSuffix, pathmatch.PathPrefix, pathmatch.HostAndPath:
		return v, nil
	}
	return "", fmt.Errorf("不支持的匹配类型: %s", value)
}

// validateEntryRule 校验匹配类型与字段组合，并返回规范化后的 host / path。
func validateEntryRule(matchType, host, pathPattern string) (string, string, error) {
	mt, err := normalizeMatchType(matchType)
	if err != nil {
		return "", "", err
	}
	normalizedHost := pathmatch.NormalizeHostPattern(mt, host)
	normalizedPath := pathmatch.NormalizePathPattern(pathPattern)
	switch mt {
	case pathmatch.HostExact, pathmatch.HostSuffix:
		if normalizedHost == "" {
			return "", "", errors.New("Host 不能为空")
		}
		normalizedPath = ""
	case pathmatch.PathPrefix:
		if normalizedPath == "" {
			return "", "", errors.New("路径模式不能为空")
		}
		normalizedHost = ""
	case pathmatch.HostAndPath:
		if normalizedHost == "" || normalizedPath == "" {
			return "", "", errors.New("host_and_path 类型必须同时填写 Host 和路径模式")
		}
	}
	if normalizedPath != "" {
		if _, err := pathmatch.CompilePathPattern(normalizedPath); err != nil {
			return "", "", fmt.Errorf("路径模式编译失败: %w", err)
		}
	}
	return normalizedHost, normalizedPath, nil
}

func (s *service) SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error) {
	if req == nil {
		return nil, errors.New("入口绑定参数不能为空")
	}
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	if requestAppKey := appctx.NormalizeExplicitAppKey(req.AppKey); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, errors.New("app_key 不匹配")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	if ok, err := appExists(s.db, normalizedAppKey); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("应用不存在")
	}
	matchType, err := normalizeMatchType(req.MatchType)
	if err != nil {
		return nil, err
	}
	host, pathPattern, err := validateEntryRule(matchType, req.Host, req.PathPattern)
	if err != nil {
		return nil, err
	}
	defaultSpaceKey := spacepkg.NormalizeSpaceKey(req.DefaultSpaceKey)
	if defaultSpaceKey == "" {
		return nil, errors.New("default_space_key 不能为空")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}

	binding := models.AppHostBinding{
		AppKey:          normalizedAppKey,
		MatchType:       matchType,
		Host:            host,
		PathPattern:     pathPattern,
		Priority:        req.Priority,
		Description:     strings.TrimSpace(req.Description),
		IsPrimary:       req.IsPrimary,
		DefaultSpaceKey: defaultSpaceKey,
		Status:          status,
		Meta:            models.MetaJSON(req.Meta),
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureSpaceExistsForApp(tx, normalizedAppKey, defaultSpaceKey); err != nil {
			return err
		}
		// 唯一性：(match_type, host, path_pattern) 全局唯一。
		conflictQuery := tx.Where("match_type = ? AND host = ? AND path_pattern = ? AND deleted_at IS NULL", matchType, host, pathPattern)
		if strings.TrimSpace(req.ID) != "" {
			conflictQuery = conflictQuery.Where("id <> ?", req.ID)
		}
		var conflictCount int64
		if err := conflictQuery.Model(&models.AppHostBinding{}).Count(&conflictCount).Error; err != nil {
			return err
		}
		if conflictCount > 0 {
			return errors.New("已存在同匹配类型 + Host + 路径的入口绑定")
		}

		if strings.TrimSpace(req.ID) != "" {
			var existing models.AppHostBinding
			// P1: 限定 app_key 防止跨 App 改绑定。
			if err := tx.Where("id = ? AND app_key = ? AND deleted_at IS NULL", req.ID, normalizedAppKey).First(&existing).Error; err != nil {
				return err
			}
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"app_key":           binding.AppKey,
				"match_type":        binding.MatchType,
				"host":              binding.Host,
				"path_pattern":      binding.PathPattern,
				"priority":          binding.Priority,
				"description":       binding.Description,
				"is_primary":        binding.IsPrimary,
				"default_space_key": binding.DefaultSpaceKey,
				"status":            binding.Status,
				"meta":              binding.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
			binding.ID = existing.ID
		} else {
			if createErr := tx.Create(&binding).Error; createErr != nil {
				return createErr
			}
		}
		if binding.IsPrimary {
			if err := tx.Model(&models.AppHostBinding{}).
				Where("app_key = ? AND id <> ? AND deleted_at IS NULL", binding.AppKey, binding.ID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	items, err := s.ListHostBindings(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == binding.ID {
			return &items[i], nil
		}
	}
	if len(items) > 0 {
		return &items[0], nil
	}
	return nil, errors.New("保存入口绑定后读取失败")
}

func (s *service) DeleteHostBinding(appKey, id string) error {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("id 不能为空")
	}
	return s.db.Where("id = ? AND app_key = ? AND deleted_at IS NULL", id, normalizedAppKey).
		Delete(&models.AppHostBinding{}).Error
}

func (s *service) ListMenuSpaceEntryBindings(appKey string) ([]MenuSpaceEntryBindingRecord, error) {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	var items []models.MenuSpaceEntryBinding
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).
		Order("priority DESC, is_primary DESC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	// app + space name 索引
	var app models.App
	_ = s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&app).Error
	var spaceList []models.MenuSpace
	_ = s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).Find(&spaceList).Error
	spaceNames := make(map[string]string, len(spaceList))
	for _, sp := range spaceList {
		spaceNames[sp.SpaceKey] = sp.Name
	}
	records := make([]MenuSpaceEntryBindingRecord, 0, len(items))
	for _, item := range items {
		records = append(records, MenuSpaceEntryBindingRecord{
			MenuSpaceEntryBinding: item,
			AppName:               app.Name,
			SpaceName:             spaceNames[item.SpaceKey],
		})
	}
	return records, nil
}

func (s *service) SaveMenuSpaceEntryBinding(appKey string, req *SaveMenuSpaceEntryBindingRequest) (*MenuSpaceEntryBindingRecord, error) {
	if req == nil {
		return nil, errors.New("菜单空间入口绑定参数不能为空")
	}
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	if requestAppKey := appctx.NormalizeExplicitAppKey(req.AppKey); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, errors.New("app_key 不匹配")
	}
	spaceKey := spacepkg.NormalizeSpaceKey(req.SpaceKey)
	if spaceKey == "" {
		return nil, errors.New("space_key 不能为空")
	}
	// 单空间 App 不允许配置 Level 2。
	var app models.App
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&app).Error; err != nil {
		return nil, err
	}
	if app.SpaceMode != "multi" {
		return nil, errors.New("单空间 App 无需配置菜单空间入口绑定")
	}
	if err := ensureSpaceExistsForApp(s.db, normalizedAppKey, spaceKey); err != nil {
		return nil, err
	}

	matchType, err := normalizeMatchType(req.MatchType)
	if err != nil {
		return nil, err
	}
	host, pathPattern, err := validateEntryRule(matchType, req.Host, req.PathPattern)
	if err != nil {
		return nil, err
	}

	// Level 2 不能超出 Level 1 任何一条规则的范围。
	// 加载 App 的所有 Level 1 绑定，若全部 host 均非空且 child host 都不在范围内 → 拒绝。
	var l1Bindings []models.AppHostBinding
	if err := s.db.Where("app_key = ? AND status = ? AND deleted_at IS NULL", normalizedAppKey, "normal").Find(&l1Bindings).Error; err != nil {
		return nil, err
	}
	if len(l1Bindings) > 0 {
		// 至少匹配一条 Level 1 规则的 host & path 范围
		ok := false
		for _, l1 := range l1Bindings {
			if !pathmatch.IsHostInScope(l1.MatchType, l1.Host, matchType, host) {
				continue
			}
			if !pathmatch.IsPathInScope(l1.PathPattern, pathPattern) {
				continue
			}
			ok = true
			break
		}
		if !ok {
			return nil, errors.New("菜单空间入口绑定必须落在 APP 入口规则范围内")
		}
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	binding := models.MenuSpaceEntryBinding{
		AppKey:      normalizedAppKey,
		SpaceKey:    spaceKey,
		MatchType:   matchType,
		Host:        host,
		PathPattern: pathPattern,
		Priority:    req.Priority,
		IsPrimary:   req.IsPrimary,
		Description: strings.TrimSpace(req.Description),
		Status:      status,
		Meta:        models.MetaJSON(req.Meta),
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		conflictQuery := tx.Where("app_key = ? AND match_type = ? AND host = ? AND path_pattern = ? AND deleted_at IS NULL",
			normalizedAppKey, matchType, host, pathPattern)
		if strings.TrimSpace(req.ID) != "" {
			conflictQuery = conflictQuery.Where("id <> ?", req.ID)
		}
		var conflictCount int64
		if err := conflictQuery.Model(&models.MenuSpaceEntryBinding{}).Count(&conflictCount).Error; err != nil {
			return err
		}
		if conflictCount > 0 {
			return errors.New("已存在同匹配类型 + Host + 路径的菜单空间入口绑定")
		}
		if strings.TrimSpace(req.ID) != "" {
			var existing models.MenuSpaceEntryBinding
			// P1: 限定 app_key 防止跨 App 改绑定。
			if err := tx.Where("id = ? AND app_key = ? AND deleted_at IS NULL", req.ID, normalizedAppKey).First(&existing).Error; err != nil {
				return err
			}
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"space_key":    binding.SpaceKey,
				"match_type":   binding.MatchType,
				"host":         binding.Host,
				"path_pattern": binding.PathPattern,
				"priority":     binding.Priority,
				"is_primary":   binding.IsPrimary,
				"description":  binding.Description,
				"status":       binding.Status,
				"meta":         binding.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
			binding.ID = existing.ID
		} else {
			if err := tx.Create(&binding).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	items, err := s.ListMenuSpaceEntryBindings(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == binding.ID {
			return &items[i], nil
		}
	}
	return nil, errors.New("保存菜单空间入口绑定后读取失败")
}

func (s *service) DeleteMenuSpaceEntryBinding(appKey, id string) error {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("id 不能为空")
	}
	return s.db.Where("id = ? AND app_key = ? AND deleted_at IS NULL", id, normalizedAppKey).
		Delete(&models.MenuSpaceEntryBinding{}).Error
}

func ensureSpaceExistsForApp(db *gorm.DB, appKey string, spaceKey string) error {
	normalizedAppKey := NormalizeAppKey(appKey)
	normalizedSpaceKey := spacepkg.NormalizeSpaceKey(spaceKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	if normalizedSpaceKey == "" {
		return errors.New("default_space_key 不能为空")
	}
	if normalizedSpaceKey == models.DefaultMenuSpaceKey {
		return spacepkg.EnsureDefaultMenuSpace(db, normalizedAppKey)
	}
	var count int64
	if err := db.Model(&models.MenuSpace{}).
		Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", normalizedAppKey, normalizedSpaceKey).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("默认空间不存在，请先在高级空间配置中创建")
	}
	return nil
}

func (s *service) getAppRecord(appKey string) (*AppRecord, error) {
	records, err := s.ListApps()
	if err != nil {
		return nil, err
	}
	target := NormalizeAppKey(appKey)
	for i := range records {
		if NormalizeAppKey(records[i].AppKey) == target {
			record := records[i]
			return &record, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func ensureDefaultApp(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var existing models.App
	err := db.Where("app_key = ? AND deleted_at IS NULL", models.DefaultAppKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"name":              models.DefaultAppName,
			"space_mode":        "multi",
			"default_space_key": models.DefaultMenuSpaceKey,
			"status":            "normal",
			"is_default":        true,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&models.App{
			AppKey:          models.DefaultAppKey,
			Name:            models.DefaultAppName,
			Description:     "当前内置管理员后台应用",
			SpaceMode:       "multi",
			DefaultSpaceKey: models.DefaultMenuSpaceKey,
			AuthMode:        "inherit_host",
			Status:          "normal",
			IsDefault:       true,
			Meta:            models.MetaJSON{},
		}).Error
	default:
		return err
	}
}

func loadDefaultApp(db *gorm.DB) (*models.App, error) {
	if err := ensureDefaultApp(db); err != nil {
		return nil, err
	}
	var item models.App
	err := db.Where("is_default = ? AND deleted_at IS NULL", true).Order("updated_at DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = db.Where("app_key = ? AND deleted_at IS NULL", models.DefaultAppKey).First(&item).Error
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func appExists(db *gorm.DB, appKey string) (bool, error) {
	if db == nil {
		return false, nil
	}
	var count int64
	if err := db.Model(&models.App{}).Where("app_key = ? AND deleted_at IS NULL", NormalizeAppKey(appKey)).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func loadStringCounts(db *gorm.DB, model interface{}, keyColumn string) (map[string]int, error) {
	type countRow struct {
		Key   string `gorm:"column:key"`
		Total int64  `gorm:"column:total"`
	}
	rows := make([]countRow, 0)
	if err := db.Model(model).
		Select(keyColumn + " AS key, COUNT(*) AS total").
		Where("deleted_at IS NULL").
		Group(keyColumn).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[NormalizeAppKey(row.Key)] = int(row.Total)
	}
	return result, nil
}

func loadStringCountAndHosts(db *gorm.DB, model interface{}, keyColumn string, hostColumn string, primaryFilter string) (map[string]int, map[string][]string, error) {
	counts, err := loadStringCounts(db, model, keyColumn)
	if err != nil {
		return nil, nil, err
	}
	type hostRow struct {
		Key  string `gorm:"column:key"`
		Host string `gorm:"column:host"`
	}
	rows := make([]hostRow, 0)
	query := db.Model(model).
		Select(keyColumn + " AS key, " + hostColumn + " AS host").
		Where("deleted_at IS NULL")
	if strings.TrimSpace(primaryFilter) != "" {
		query = query.Where(primaryFilter)
	}
	if err := query.Order(hostColumn + " ASC").Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	hosts := make(map[string][]string)
	for _, row := range rows {
		key := NormalizeAppKey(row.Key)
		hosts[key] = append(hosts[key], row.Host)
	}
	return counts, hosts, nil
}
