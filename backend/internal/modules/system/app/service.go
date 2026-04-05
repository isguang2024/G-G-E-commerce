package app

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
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
	DefaultSpaceKey string                 `json:"default_space_key"`
	AuthMode        string                 `json:"auth_mode"`
	Status          string                 `json:"status"`
	IsDefault       bool                   `json:"is_default"`
	Meta            map[string]interface{} `json:"meta"`
}

type SaveHostBindingRequest struct {
	AppKey          string                 `json:"app_key"`
	Host            string                 `json:"host"`
	Description     string                 `json:"description"`
	IsPrimary       bool                   `json:"is_primary"`
	DefaultSpaceKey string                 `json:"default_space_key"`
	Status          string                 `json:"status"`
	Meta            map[string]interface{} `json:"meta"`
}

type Service interface {
	ListApps() ([]AppRecord, error)
	GetCurrent(host, requestedAppKey string) (*CurrentResponse, error)
	SaveApp(req *SaveAppRequest) (*AppRecord, error)
	ListHostBindings(appKey string) ([]HostBindingRecord, error)
	SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error)
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

func RequestAppKey(c *gin.Context) string {
	return appctx.RequestAppKey(c)
}

func CurrentAppKey(c *gin.Context) string {
	return appctx.CurrentAppKey(c)
}

func ResolveAppByHost(db *gorm.DB, host string, requestedAppKey string) (string, *models.AppHostBinding, string, error) {
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

	normalizedHost := spacepkg.NormalizeHost(host)
	if normalizedHost != "" {
		var binding models.AppHostBinding
		err := db.Where("host = ? AND status = ? AND deleted_at IS NULL", normalizedHost, "normal").First(&binding).Error
		if err == nil {
			ok, appErr := appExists(db, binding.AppKey)
			if appErr != nil {
				return models.DefaultAppKey, nil, "", appErr
			}
			if ok {
				return NormalizeAppKey(binding.AppKey), &binding, "host_binding", nil
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DefaultAppKey, nil, "", err
		}

		// 兼容旧的菜单空间 Host 绑定，优先级低于 App 绑定。
		var legacyBinding models.MenuSpaceHostBinding
		err = db.Where("host = ? AND status = ? AND deleted_at IS NULL", normalizedHost, "normal").First(&legacyBinding).Error
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
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return models.DefaultAppKey, nil, "", err
		}
	}

	defaultApp, err := loadDefaultApp(db)
	if err != nil {
		return models.DefaultAppKey, nil, "", err
	}
	return defaultApp.AppKey, nil, "default_app", nil
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
	if defaultSpaceKey == "" {
		return nil, errors.New("default_space_key is required")
	}
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
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"name":              payload.Name,
				"description":       payload.Description,
				"default_space_key": payload.DefaultSpaceKey,
				"auth_mode":         payload.AuthMode,
				"status":            payload.Status,
				"is_default":        payload.IsDefault,
				"meta":              payload.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if createErr := tx.Create(&payload).Error; createErr != nil {
				return createErr
			}
		default:
			return err
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
		return nil, errors.New("app_key is required")
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

func (s *service) SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error) {
	if req == nil {
		return nil, errors.New("Host 绑定参数不能为空")
	}
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key is required")
	}
	if requestAppKey := appctx.NormalizeExplicitAppKey(req.AppKey); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, errors.New("app_key mismatch")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	if ok, err := appExists(s.db, normalizedAppKey); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("应用不存在")
	}
	host := spacepkg.NormalizeHost(req.Host)
	if host == "" {
		return nil, errors.New("Host 不能为空")
	}
	defaultSpaceKey := spacepkg.NormalizeSpaceKey(req.DefaultSpaceKey)
	if defaultSpaceKey == "" {
		return nil, errors.New("default_space_key is required")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}

	binding := models.AppHostBinding{
		AppKey:          normalizedAppKey,
		Host:            host,
		Description:     strings.TrimSpace(req.Description),
		IsPrimary:       req.IsPrimary,
		DefaultSpaceKey: defaultSpaceKey,
		Status:          status,
		Meta:            models.MetaJSON(req.Meta),
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing models.AppHostBinding
		err := tx.Where("host = ? AND deleted_at IS NULL", host).First(&existing).Error
		switch {
		case err == nil:
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"app_key":           binding.AppKey,
				"description":       binding.Description,
				"is_primary":        binding.IsPrimary,
				"default_space_key": binding.DefaultSpaceKey,
				"status":            binding.Status,
				"meta":              binding.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if createErr := tx.Create(&binding).Error; createErr != nil {
				return createErr
			}
		default:
			return err
		}
		if binding.IsPrimary {
			if err := tx.Model(&models.AppHostBinding{}).
				Where("app_key = ? AND host <> ? AND deleted_at IS NULL", binding.AppKey, binding.Host).
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
		if spacepkg.NormalizeHost(items[i].Host) == host {
			return &items[i], nil
		}
	}
	return nil, errors.New("保存 Host 绑定后读取失败")
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
			"default_space_key": models.DefaultMenuSpaceKey,
			"status":            "normal",
			"is_default":        true,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&models.App{
			AppKey:          models.DefaultAppKey,
			Name:            models.DefaultAppName,
			Description:     "当前内置管理员后台应用",
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
