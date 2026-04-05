package space

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
)

type SpaceRecord struct {
	models.MenuSpace
	HostCount int      `json:"host_count"`
	Hosts     []string `json:"hosts,omitempty"`
	MenuCount int      `json:"menu_count"`
	// PageCount 统计的是独立页暴露数量，不再代表“复制到该空间的页面定义数量”。
	PageCount        int      `json:"page_count"`
	AccessMode       string   `json:"access_mode"`
	AllowedRoleCodes []string `json:"allowed_role_codes"`
}

type HostBindingRecord struct {
	ID          uuid.UUID       `json:"id"`
	AppKey      string          `json:"app_key"`
	Host        string          `json:"host"`
	SpaceKey    string          `json:"space_key"`
	SpaceName   string          `json:"space_name"`
	Description string          `json:"description"`
	IsDefault   bool            `json:"is_default"`
	Status      string          `json:"status"`
	Meta        models.MetaJSON `json:"meta"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CurrentResponse struct {
	Space         SpaceRecord        `json:"space"`
	Binding       *HostBindingRecord `json:"binding,omitempty"`
	ResolvedBy    string             `json:"resolved_by"`
	RequestHost   string             `json:"request_host"`
	AccessGranted bool               `json:"access_granted"`
}

type InitializeResult struct {
	SourceSpaceKey     string `json:"source_space_key"`
	TargetSpaceKey     string `json:"target_space_key"`
	ForceReinitialized bool   `json:"force_reinitialized"`
	ClearedMenuCount   int    `json:"cleared_menu_count"`
	// ClearedPageCount / CreatedPageCount 仅表示独立页暴露绑定数量，兼容旧接口命名保留。
	ClearedPageCount       int `json:"cleared_page_count"`
	ClearedPackageMenuLink int `json:"cleared_package_menu_link_count"`
	CreatedMenuCount       int `json:"created_menu_count"`
	CreatedPageCount       int `json:"created_page_count"`
	CreatedPackageMenuLink int `json:"created_package_menu_link_count"`
}

type SaveSpaceRequest struct {
	AppKey           string                 `json:"app_key"`
	SpaceKey         string                 `json:"space_key"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	DefaultHomePath  string                 `json:"default_home_path"`
	IsDefault        bool                   `json:"is_default"`
	Status           string                 `json:"status"`
	AccessMode       string                 `json:"access_mode"`
	AllowedRoleCodes []string               `json:"allowed_role_codes"`
	Meta             map[string]interface{} `json:"meta"`
}

type SaveHostBindingRequest struct {
	AppKey      string                 `json:"app_key"`
	Host        string                 `json:"host"`
	SpaceKey    string                 `json:"space_key"`
	Description string                 `json:"description"`
	IsDefault   bool                   `json:"is_default"`
	Status      string                 `json:"status"`
	Meta        map[string]interface{} `json:"meta"`
}

const (
	hostBindingSchemeHTTP  = "http"
	hostBindingSchemeHTTPS = "https"

	hostBindingAuthInherit      = "inherit_host"
	hostBindingAuthCentralized  = "centralized_login"
	hostBindingAuthSharedCookie = "shared_cookie"

	hostBindingCookieInherit      = "inherit"
	hostBindingCookieHostOnly     = "host_only"
	hostBindingCookieParentDomain = "parent_domain"
)

type Service interface {
	ListSpaces(appKey string) ([]SpaceRecord, error)
	GetCurrent(appKey string, host string, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CurrentResponse, error)
	ListHostBindings(appKey string) ([]HostBindingRecord, error)
	GetMode() (string, error)
	SaveMode(mode string) (string, error)
	SaveSpace(appKey string, req *SaveSpaceRequest) (*SpaceRecord, error)
	SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error)
	InitializeFromDefault(appKey string, targetSpaceKey string, force bool, actorUserID *uuid.UUID) (*InitializeResult, error)
}

type service struct {
	db        *gorm.DB
	refresher permissionrefresh.Service
	logger    *zap.Logger
}

func NewService(db *gorm.DB, refresher permissionrefresh.Service, logger *zap.Logger) Service {
	return &service{db: db, refresher: refresher, logger: logger}
}

func EnsureDefaultMenuSpace(db *gorm.DB, appKey string) error {
	if db == nil {
		return nil
	}
	normalizedAppKey := normalizeAppKey(appKey)
	defaultSpace := models.MenuSpace{
		AppKey:          normalizedAppKey,
		SpaceKey:        DefaultMenuSpaceKey,
		Name:            "默认菜单空间",
		Description:     "兼容当前单域单菜单运行模式",
		DefaultHomePath: "/dashboard/console",
		IsDefault:       true,
		Status:          "normal",
		Meta:            models.MetaJSON{},
	}
	var existing models.MenuSpace
	err := db.Where("app_key = ? AND space_key = ?", normalizedAppKey, DefaultMenuSpaceKey).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"app_key":    normalizedAppKey,
			"is_default": true,
			"status":     "normal",
		}
		return db.Model(&models.MenuSpace{}).Where("app_key = ? AND space_key = ?", normalizedAppKey, DefaultMenuSpaceKey).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&defaultSpace).Error
}

func (s *service) ListSpaces(appKey string) ([]SpaceRecord, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	if err := EnsureDefaultMenuSpace(s.db, normalizedAppKey); err != nil {
		return nil, err
	}
	var spaces []models.MenuSpace
	if err := s.db.Where("app_key = ?", normalizedAppKey).Order("is_default DESC, created_at ASC").Find(&spaces).Error; err != nil {
		return nil, err
	}
	var bindings []models.AppHostBinding
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).Order("created_at ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	menuCountMap, err := s.loadMenuCountMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	pageCountMap, err := s.loadPageCountMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]string, len(spaces))
	for _, binding := range bindings {
		key := NormalizeSpaceKey(binding.DefaultSpaceKey)
		grouped[key] = append(grouped[key], binding.Host)
	}
	records := make([]SpaceRecord, 0, len(spaces))
	for _, item := range spaces {
		key := NormalizeSpaceKey(item.SpaceKey)
		hosts := append([]string(nil), grouped[key]...)
		sort.Strings(hosts)
		record := SpaceRecord{
			MenuSpace:        item,
			HostCount:        len(hosts),
			Hosts:            hosts,
			MenuCount:        menuCountMap[key],
			PageCount:        pageCountMap[key],
			AccessMode:       ExtractSpaceAccessProfile(item.Meta).Mode,
			AllowedRoleCodes: ExtractSpaceAccessProfile(item.Meta).AllowedRoleCodes,
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *service) GetCurrent(appKey string, host string, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CurrentResponse, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	if err := EnsureDefaultMenuSpace(s.db, normalizedAppKey); err != nil {
		return nil, err
	}
	resolvedKey, resolvedBy, err := ResolveCurrentSpaceKey(s.db, normalizedAppKey, host, requestedSpaceKey, userID, tenantID)
	if err != nil {
		return nil, err
	}
	accessGranted, accessErr := CanAccessSpace(s.db, userID, tenantID, resolvedKey)
	if accessErr != nil {
		return nil, accessErr
	}
	spaceRecord, err := s.getSpaceRecord(normalizedAppKey, resolvedKey)
	if err != nil {
		return nil, err
	}

	var bindingRecord *HostBindingRecord
	if resolvedBy == "app_host_binding" || resolvedBy == "legacy_space_host_binding" {
		bindingRecord, _ = s.getHostBindingRecord(normalizedAppKey, host)
	}

	return &CurrentResponse{
		Space:         *spaceRecord,
		Binding:       bindingRecord,
		ResolvedBy:    resolvedBy,
		RequestHost:   NormalizeHost(host),
		AccessGranted: accessGranted || resolvedKey == DefaultMenuSpaceKey,
	}, nil
}

func (s *service) GetMode() (string, error) {
	return CurrentSpaceMode(s.db), nil
}

func (s *service) SaveMode(mode string) (string, error) {
	return SaveCurrentSpaceMode(s.db, mode)
}

func (s *service) ListHostBindings(appKey string) ([]HostBindingRecord, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	if err := EnsureDefaultMenuSpace(s.db, normalizedAppKey); err != nil {
		return nil, err
	}
	var bindings []models.AppHostBinding
	if err := s.db.Where("app_key = ?", normalizedAppKey).Order("created_at ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	spaceMap, err := s.loadSpaceNameMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	records := make([]HostBindingRecord, 0, len(bindings))
	for _, item := range bindings {
		records = append(records, HostBindingRecord{
			ID:          item.ID,
			AppKey:      normalizedAppKey,
			Host:        item.Host,
			SpaceKey:    NormalizeSpaceKey(item.DefaultSpaceKey),
			SpaceName:   spaceMap[NormalizeSpaceKey(item.DefaultSpaceKey)],
			Description: item.Description,
			IsDefault:   item.IsPrimary,
			Status:      item.Status,
			Meta:        item.Meta,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	return records, nil
}

func (s *service) SaveSpace(appKey string, req *SaveSpaceRequest) (*SpaceRecord, error) {
	if req == nil {
		return nil, fmt.Errorf("space request is required")
	}
	normalizedAppKey := normalizeAppKey(strings.TrimSpace(appKey))
	if normalizedAppKey == "" {
		return nil, fmt.Errorf("app_key is required")
	}
	if requestAppKey := normalizeAppKey(strings.TrimSpace(req.AppKey)); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, fmt.Errorf("app_key mismatch")
	}
	key := NormalizeSpaceKey(req.SpaceKey)
	if key == "" {
		key = DefaultMenuSpaceKey
	}
	name := strings.TrimSpace(req.Name)
	if key == DefaultMenuSpaceKey && name == "" {
		name = "默认菜单空间"
	}
	if name == "" {
		return nil, fmt.Errorf("space name is required")
	}
	status := normalizeSpaceStatus(req.Status)
	if status == "" {
		status = "normal"
	}
	if req.IsDefault {
		status = "normal"
	}
	description := strings.TrimSpace(req.Description)
	defaultHomePath := strings.TrimSpace(req.DefaultHomePath)
	if defaultHomePath == "" && key == DefaultMenuSpaceKey {
		defaultHomePath = "/dashboard/console"
	}
	if defaultHomePath != "" && !isValidInternalPath(defaultHomePath) {
		return nil, fmt.Errorf("默认首页必须是以 / 开头的站内路径")
	}
	meta := req.Meta
	if meta == nil {
		meta = models.MetaJSON{}
	}
	accessMode := strings.TrimSpace(req.AccessMode)
	switch accessMode {
	case spaceAccessModePlatform, spaceAccessModeTeam, spaceAccessModeRoleCodes:
	default:
		accessMode = spaceAccessModeAll
	}
	meta["access_mode"] = accessMode
	meta["allowed_role_codes"] = normalizeStringSlice(req.AllowedRoleCodes)

	record := models.MenuSpace{
		AppKey:          normalizedAppKey,
		SpaceKey:        key,
		Name:            name,
		Description:     description,
		DefaultHomePath: defaultHomePath,
		IsDefault:       req.IsDefault || key == DefaultMenuSpaceKey,
		Status:          status,
		Meta:            meta,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if record.IsDefault {
			if err := tx.Model(&models.MenuSpace{}).
				Where("app_key = ? AND space_key <> ?", normalizedAppKey, record.SpaceKey).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "app_key"}, {Name: "space_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"description",
				"default_home_path",
				"is_default",
				"status",
				"meta",
				"updated_at",
			}),
		}).Create(&record).Error; err != nil {
			return err
		}
		if record.IsDefault {
			return tx.Model(&models.App{}).
				Where("app_key = ?", normalizedAppKey).
				Update("default_space_key", record.SpaceKey).Error
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.getSpaceRecord(normalizedAppKey, key)
}

func (s *service) SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error) {
	if req == nil {
		return nil, fmt.Errorf("host binding request is required")
	}
	normalizedAppKey := normalizeAppKey(strings.TrimSpace(appKey))
	if normalizedAppKey == "" {
		return nil, fmt.Errorf("app_key is required")
	}
	if requestAppKey := normalizeAppKey(strings.TrimSpace(req.AppKey)); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, fmt.Errorf("app_key mismatch")
	}
	host := NormalizeHost(req.Host)
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	spaceKey := NormalizeSpaceKey(req.SpaceKey)
	if spaceKey == "" {
		spaceKey = DefaultMenuSpaceKey
	}
	ok, err := spaceExists(s.db, normalizedAppKey, spaceKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("menu space not found: %s", spaceKey)
	}
	var targetSpace models.MenuSpace
	if err := s.db.Select("app_key", "space_key", "status").Where("app_key = ? AND space_key = ?", normalizedAppKey, spaceKey).First(&targetSpace).Error; err != nil {
		return nil, err
	}
	status := normalizeSpaceStatus(req.Status)
	if status == "" {
		status = "normal"
	}
	if targetSpace.Status == "disabled" && status == "normal" {
		return nil, fmt.Errorf("目标菜单空间已停用，无法启用 Host 绑定")
	}
	if !isValidMenuSpaceHost(host) {
		return nil, fmt.Errorf("Host 格式无效，请填写域名、子域名、localhost 或 IP")
	}
	var existingBinding models.AppHostBinding
	if err := s.db.Where("host = ? AND deleted_at IS NULL", host).First(&existingBinding).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	meta := normalizeMeta(req.Meta)
	scheme := normalizeHostBindingScheme(toStringValue(meta["scheme"]))
	meta["scheme"] = scheme
	routePrefix := normalizeHostBindingRoutePrefix(toStringValue(meta["route_prefix"], meta["routePrefix"]))
	if routePrefix != "" && !isValidInternalPath(routePrefix) {
		return nil, fmt.Errorf("路由前缀必须是以 / 开头的站内路径")
	}
	meta["route_prefix"] = routePrefix
	authMode := normalizeHostBindingAuthMode(toStringValue(meta["auth_mode"], meta["authMode"]))
	meta["auth_mode"] = authMode
	loginHost := NormalizeHost(toStringValue(meta["login_host"], meta["loginHost"]))
	if loginHost != "" && !isValidMenuSpaceHost(loginHost) {
		return nil, fmt.Errorf("统一登录 Host 格式无效")
	}
	meta["login_host"] = loginHost
	callbackHost := NormalizeHost(toStringValue(meta["callback_host"], meta["callbackHost"]))
	if callbackHost != "" && !isValidMenuSpaceHost(callbackHost) {
		return nil, fmt.Errorf("登录回调 Host 格式无效")
	}
	meta["callback_host"] = callbackHost
	cookieScopeMode := normalizeHostBindingCookieScopeMode(toStringValue(meta["cookie_scope_mode"], meta["cookieScopeMode"]))
	meta["cookie_scope_mode"] = cookieScopeMode
	cookieDomain := NormalizeHost(toStringValue(meta["cookie_domain"], meta["cookieDomain"]))
	if cookieDomain != "" && !isValidMenuSpaceHost(strings.TrimPrefix(cookieDomain, ".")) {
		return nil, fmt.Errorf("Cookie 域格式无效")
	}
	meta["cookie_domain"] = cookieDomain
	binding := models.AppHostBinding{
		AppKey:          normalizedAppKey,
		Host:            host,
		Description:     strings.TrimSpace(req.Description),
		IsPrimary:       req.IsDefault,
		DefaultSpaceKey: spaceKey,
		Status:          status,
		Meta:            meta,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if binding.IsPrimary {
			if err := tx.Model(&models.AppHostBinding{}).
				Where("app_key = ? AND host <> ? AND deleted_at IS NULL", normalizedAppKey, host).
				Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "host"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"app_key",
				"default_space_key",
				"description",
				"is_primary",
				"status",
				"meta",
				"updated_at",
			}),
		}).Create(&binding).Error; err != nil {
			return err
		}
		if binding.IsPrimary {
			return tx.Model(&models.App{}).
				Where("app_key = ?", normalizedAppKey).
				Update("default_space_key", binding.DefaultSpaceKey).Error
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.getHostBindingRecord(normalizedAppKey, host)
}

func (s *service) InitializeFromDefault(appKey string, targetSpaceKey string, force bool, actorUserID *uuid.UUID) (*InitializeResult, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, fmt.Errorf("app_key is required")
	}
	targetKey := NormalizeSpaceKey(targetSpaceKey)
	if targetKey == "" {
		return nil, fmt.Errorf("target space is required")
	}
	if targetKey == DefaultMenuSpaceKey {
		return nil, fmt.Errorf("默认菜单空间无需初始化")
	}
	ok, err := spaceExists(s.db, normalizedAppKey, targetKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("menu space not found: %s", targetKey)
	}

	sourceSpaceKey, err := loadAppDefaultSpaceKey(s.db, normalizedAppKey)
	if err != nil {
		return nil, err
	}
	if sourceSpaceKey == "" {
		sourceSpaceKey = DefaultMenuSpaceKey
	}
	result := &InitializeResult{
		SourceSpaceKey:     sourceSpaceKey,
		TargetSpaceKey:     targetKey,
		ForceReinitialized: force,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var existingMenuCount int64
		if err := tx.Model(&models.SpaceMenuPlacement{}).
			Where("app_key = ?", normalizedAppKey).
			Where("space_key = ?", targetKey).
			Count(&existingMenuCount).Error; err != nil {
			return err
		}
		if existingMenuCount > 0 && !force {
			return fmt.Errorf("目标空间已存在菜单，请先清空后再初始化")
		}

		if force {
			if existingMenuCount > 0 {
				result.ClearedMenuCount = int(existingMenuCount)
				if err := tx.Where("app_key = ? AND space_key = ?", normalizedAppKey, targetKey).
					Delete(&models.SpaceMenuPlacement{}).Error; err != nil {
					return err
				}
			}
		}

		var sourcePlacements []models.SpaceMenuPlacement
		if err := tx.Where("app_key = ?", normalizedAppKey).
			Where("space_key = ?", sourceSpaceKey).
			Order("sort_order ASC, created_at ASC").
			Find(&sourcePlacements).Error; err != nil {
			return err
		}

		clonedPlacements := make([]models.SpaceMenuPlacement, 0, len(sourcePlacements))
		for _, item := range sourcePlacements {
			clonedPlacements = append(clonedPlacements, models.SpaceMenuPlacement{
				ID:            uuid.New(),
				AppKey:        normalizedAppKey,
				SpaceKey:      targetKey,
				MenuKey:       item.MenuKey,
				ParentMenuKey: item.ParentMenuKey,
				ManageGroupID: item.ManageGroupID,
				SortOrder:     item.SortOrder,
				Hidden:        item.Hidden,
				TitleOverride: item.TitleOverride,
				IconOverride:  item.IconOverride,
				MetaOverride:  cloneMetaJSON(item.MetaOverride),
			})
		}
		if len(clonedPlacements) > 0 {
			if err := tx.Create(&clonedPlacements).Error; err != nil {
				return err
			}
		}
		result.CreatedMenuCount = len(clonedPlacements)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.refresher != nil {
		if err := s.refresher.RefreshAllTeams(); err != nil {
			return nil, err
		}
		if err := s.refresher.RefreshAllPlatformRoles(); err != nil {
			return nil, err
		}
		if err := s.refresher.RefreshAllPlatformUsers(); err != nil {
			return nil, err
		}
	}
	if actorUserID != nil {
		if err := s.notifySpaceOperation(*actorUserID, result); err != nil && s.logger != nil {
			s.logger.Warn("Notify menu space operation failed", zap.String("space_key", targetKey), zap.Error(err))
		}
	}
	return result, nil
}

func (s *service) getSpaceRecord(appKey, spaceKey string) (*SpaceRecord, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	var item models.MenuSpace
	if err := s.db.Where("app_key = ? AND space_key = ?", normalizedAppKey, NormalizeSpaceKey(spaceKey)).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("menu space not found: %s", spaceKey)
		}
		return nil, err
	}
	var bindings []models.AppHostBinding
	if err := s.db.Where("app_key = ? AND default_space_key = ? AND deleted_at IS NULL", normalizedAppKey, item.SpaceKey).Order("created_at ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	hosts := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		hosts = append(hosts, binding.Host)
	}
	sort.Strings(hosts)
	menuCountMap, err := s.loadMenuCountMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	pageCountMap, err := s.loadPageCountMap(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	return &SpaceRecord{
		MenuSpace:        item,
		HostCount:        len(hosts),
		Hosts:            hosts,
		MenuCount:        menuCountMap[NormalizeSpaceKey(item.SpaceKey)],
		PageCount:        pageCountMap[NormalizeSpaceKey(item.SpaceKey)],
		AccessMode:       ExtractSpaceAccessProfile(item.Meta).Mode,
		AllowedRoleCodes: ExtractSpaceAccessProfile(item.Meta).AllowedRoleCodes,
	}, nil
}

func (s *service) getHostBindingRecord(appKey, host string) (*HostBindingRecord, error) {
	normalizedAppKey := normalizeAppKey(appKey)
	normalizedHost := NormalizeHost(host)
	var binding models.AppHostBinding
	if err := s.db.Where("app_key = ? AND host = ? AND deleted_at IS NULL", normalizedAppKey, normalizedHost).First(&binding).Error; err == nil {
		spaceRecord, _ := s.getSpaceRecord(normalizedAppKey, binding.DefaultSpaceKey)
		spaceName := ""
		if spaceRecord != nil {
			spaceName = spaceRecord.Name
		}
		return &HostBindingRecord{
			ID:          binding.ID,
			AppKey:      normalizedAppKey,
			Host:        binding.Host,
			SpaceKey:    NormalizeSpaceKey(binding.DefaultSpaceKey),
			SpaceName:   spaceName,
			Description: binding.Description,
			IsDefault:   binding.IsPrimary,
			Status:      binding.Status,
			Meta:        binding.Meta,
			CreatedAt:   binding.CreatedAt,
			UpdatedAt:   binding.UpdatedAt,
		}, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var legacy models.MenuSpaceHostBinding
	if err := s.db.Where("host = ? AND deleted_at IS NULL", normalizedHost).First(&legacy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("host binding not found: %s", host)
		}
		return nil, err
	}
	spaceRecord, err := s.getSpaceRecord(normalizedAppKey, legacy.SpaceKey)
	if err != nil {
		return nil, err
	}
	return &HostBindingRecord{
		ID:          legacy.ID,
		AppKey:      normalizedAppKey,
		Host:        legacy.Host,
		SpaceKey:    NormalizeSpaceKey(legacy.SpaceKey),
		SpaceName:   spaceRecord.Name,
		Description: legacy.Description,
		IsDefault:   legacy.IsDefault,
		Status:      legacy.Status,
		Meta:        legacy.Meta,
		CreatedAt:   legacy.CreatedAt,
		UpdatedAt:   legacy.UpdatedAt,
	}, nil
}

func (s *service) loadSpaceNameMap(appKey string) (map[string]string, error) {
	var spaces []models.MenuSpace
	if err := s.db.Select("space_key", "name").Where("app_key = ?", normalizeAppKey(appKey)).Find(&spaces).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(spaces))
	for _, item := range spaces {
		result[NormalizeSpaceKey(item.SpaceKey)] = item.Name
	}
	return result, nil
}

func (s *service) loadMenuCountMap(appKey string) (map[string]int, error) {
	type aggregate struct {
		SpaceKey string
		Total    int
	}
	var rows []aggregate
	if err := s.db.Model(&models.SpaceMenuPlacement{}).
		Select("space_key, COUNT(*) AS total").
		Where("app_key = ?", normalizeAppKey(appKey)).
		Where("deleted_at IS NULL").
		Group("space_key").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, item := range rows {
		result[NormalizeSpaceKey(item.SpaceKey)] = item.Total
	}
	return result, nil
}

func (s *service) loadPageCountMap(appKey string) (map[string]int, error) {
	type aggregate struct {
		SpaceKey string
		Total    int
	}
	result := make(map[string]int)

	var bindingRows []aggregate
	if err := s.db.Model(&models.PageSpaceBinding{}).
		Select("space_key, COUNT(*) AS total").
		Where("app_key = ?", normalizeAppKey(appKey)).
		Where("deleted_at IS NULL").
		Group("space_key").
		Scan(&bindingRows).Error; err != nil {
		return nil, err
	}
	for _, item := range bindingRows {
		result[NormalizeSpaceKey(item.SpaceKey)] += item.Total
	}

	var legacyRows []aggregate
	if err := s.db.Model(&models.UIPage{}).
		Select("space_key, COUNT(*) AS total").
		Where("app_key = ?", normalizeAppKey(appKey)).
		Where("deleted_at IS NULL").
		Where("parent_menu_id IS NULL").
		Where("COALESCE(NULLIF(TRIM(parent_page_key), ''), '') = ''").
		Where("COALESCE(NULLIF(TRIM(space_key), ''), '') <> ''").
		Where("LOWER(TRIM(space_key)) <> ?", DefaultMenuSpaceKey).
		Where("NOT EXISTS (SELECT 1 FROM page_space_bindings WHERE page_space_bindings.page_id = ui_pages.id AND page_space_bindings.deleted_at IS NULL)").
		Group("space_key").
		Scan(&legacyRows).Error; err != nil {
		return nil, err
	}
	for _, item := range legacyRows {
		result[NormalizeSpaceKey(item.SpaceKey)] += item.Total
	}
	return result, nil
}

func (s *service) notifySpaceOperation(actorUserID uuid.UUID, result *InitializeResult) error {
	if result == nil || actorUserID == uuid.Nil {
		return nil
	}
	var space models.MenuSpace
	if err := s.db.Select("app_key").Where("space_key = ?", result.TargetSpaceKey).Order("is_default DESC, updated_at DESC").First(&space).Error; err != nil {
		return err
	}
	spaceRecord, err := s.getSpaceRecord(space.AppKey, result.TargetSpaceKey)
	if err != nil {
		return err
	}
	operatorName := s.loadActorDisplayName(actorUserID)
	actionText := "初始化"
	summary := fmt.Sprintf("菜单空间“%s”已从默认空间完成初始化。", spaceRecord.Name)
	contentLines := []string{
		fmt.Sprintf("操作空间：%s（%s）", spaceRecord.Name, spaceRecord.SpaceKey),
		fmt.Sprintf("操作人：%s", operatorName),
		fmt.Sprintf("新建菜单：%d", result.CreatedMenuCount),
		fmt.Sprintf("新建独立页暴露：%d", result.CreatedPageCount),
		fmt.Sprintf("新建功能包菜单关联：%d", result.CreatedPackageMenuLink),
	}
	if result.ForceReinitialized {
		actionText = "重新初始化"
		summary = fmt.Sprintf("菜单空间“%s”已完成重新初始化，原有菜单树与功能包菜单关联已被默认空间覆盖。", spaceRecord.Name)
		contentLines = append([]string{
			fmt.Sprintf("已清空菜单：%d", result.ClearedMenuCount),
			fmt.Sprintf("已清空独立页暴露：%d", result.ClearedPageCount),
			fmt.Sprintf("已清空功能包菜单关联：%d", result.ClearedPackageMenuLink),
		}, contentLines...)
	}
	now := time.Now()
	message := models.Message{
		ID:                 uuid.New(),
		MessageType:        "notice",
		BizType:            "menu_space",
		ScopeType:          "platform",
		SenderType:         "service",
		SenderUserID:       &actorUserID,
		SenderNameSnapshot: "菜单空间系统",
		SenderServiceKey:   "menu_space",
		AudienceType:       "specified_users",
		AudienceScope:      "platform",
		TargetUserIDs:      []string{actorUserID.String()},
		Title:              fmt.Sprintf("菜单空间%s完成", actionText),
		Summary:            summary,
		Content:            strings.Join(contentLines, "\n"),
		Priority:           "normal",
		ActionType:         "none",
		Status:             "published",
		PublishedAt:        &now,
		Meta: models.MetaJSON{
			"space_key":                       result.TargetSpaceKey,
			"space_name":                      spaceRecord.Name,
			"operation_type":                  map[bool]string{true: "reinitialize", false: "initialize"}[result.ForceReinitialized],
			"operator_user_id":                actorUserID.String(),
			"operator_name":                   operatorName,
			"created_menu_count":              result.CreatedMenuCount,
			"created_page_count":              result.CreatedPageCount,
			"created_package_menu_link_count": result.CreatedPackageMenuLink,
			"cleared_menu_count":              result.ClearedMenuCount,
			"cleared_page_count":              result.ClearedPageCount,
			"cleared_package_menu_link_count": result.ClearedPackageMenuLink,
			"force_reinitialized":             result.ForceReinitialized,
		},
	}
	delivery := models.MessageDelivery{
		ID:              uuid.New(),
		MessageID:       message.ID,
		RecipientUserID: actorUserID,
		BoxType:         "notice",
		DeliveryStatus:  "unread",
		TodoStatus:      "",
		Meta: models.MetaJSON{
			"recipient_username":  operatorName,
			"source_rule_type":    "specified_users",
			"source_rule_label":   "操作者通知",
			"source_target_type":  "user",
			"source_target_value": actorUserID.String(),
		},
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&message).Error; err != nil {
			return err
		}
		return tx.Create(&delivery).Error
	})
}

func (s *service) loadActorDisplayName(userID uuid.UUID) string {
	if userID == uuid.Nil {
		return "未知用户"
	}
	var user struct {
		Username string `gorm:"column:username"`
		Nickname string `gorm:"column:nickname"`
	}
	if err := s.db.Model(&models.User{}).
		Select("username", "nickname").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return userID.String()
	}
	if strings.TrimSpace(user.Nickname) != "" {
		return strings.TrimSpace(user.Nickname)
	}
	if strings.TrimSpace(user.Username) != "" {
		return strings.TrimSpace(user.Username)
	}
	return userID.String()
}

func normalizeSpaceStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "disabled", "normal":
		return strings.TrimSpace(strings.ToLower(value))
	default:
		return "normal"
	}
}

func isValidInternalPath(value string) bool {
	target := strings.TrimSpace(value)
	if target == "" || strings.Contains(target, "://") {
		return false
	}
	return strings.HasPrefix(target, "/")
}

func isValidMenuSpaceHost(value string) bool {
	target := strings.TrimSpace(strings.ToLower(value))
	if target == "" || strings.Contains(target, "/") || strings.Contains(target, "://") {
		return false
	}
	if target == "localhost" {
		return true
	}
	if ip := net.ParseIP(target); ip != nil {
		return true
	}
	parts := strings.Split(target, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, ch := range part {
			if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
				continue
			}
			return false
		}
	}
	return true
}

func normalizeHostBindingScheme(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case hostBindingSchemeHTTP:
		return hostBindingSchemeHTTP
	default:
		return hostBindingSchemeHTTPS
	}
}

func normalizeHostBindingRoutePrefix(value string) string {
	target := strings.TrimSpace(value)
	if target == "" {
		return ""
	}
	target = strings.ReplaceAll(target, "\\", "/")
	target = "/" + strings.TrimLeft(target, "/")
	if target != "/" {
		target = strings.TrimRight(target, "/")
	}
	return target
}

func normalizeHostBindingAuthMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case hostBindingAuthCentralized:
		return hostBindingAuthCentralized
	case hostBindingAuthSharedCookie:
		return hostBindingAuthSharedCookie
	default:
		return hostBindingAuthInherit
	}
}

func normalizeHostBindingCookieScopeMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case hostBindingCookieHostOnly:
		return hostBindingCookieHostOnly
	case hostBindingCookieParentDomain:
		return hostBindingCookieParentDomain
	default:
		return hostBindingCookieInherit
	}
}

func toStringValue(values ...interface{}) string {
	for _, raw := range values {
		if text, ok := raw.(string); ok {
			if strings.TrimSpace(text) != "" {
				return text
			}
		}
	}
	return ""
}

func normalizeMeta(meta map[string]interface{}) models.MetaJSON {
	if meta == nil {
		return models.MetaJSON{}
	}
	result := make(models.MetaJSON, len(meta))
	for key, value := range meta {
		result[key] = value
	}
	return result
}

func cloneMetaJSON(meta models.MetaJSON) models.MetaJSON {
	if meta == nil {
		return models.MetaJSON{}
	}
	result := make(models.MetaJSON, len(meta))
	for key, value := range meta {
		result[key] = value
	}
	return result
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		target := strings.TrimSpace(value)
		if target != "" {
			return target
		}
	}
	return ""
}
