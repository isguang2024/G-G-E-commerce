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
	HostCount        int      `json:"host_count"`
	Hosts            []string `json:"hosts,omitempty"`
	MenuCount        int      `json:"menu_count"`
	PageCount        int      `json:"page_count"`
	AccessMode       string   `json:"access_mode"`
	AllowedRoleCodes []string `json:"allowed_role_codes"`
}

type HostBindingRecord struct {
	models.MenuSpaceHostBinding
	SpaceName string `json:"space_name"`
}

type CurrentResponse struct {
	Space         SpaceRecord        `json:"space"`
	Binding       *HostBindingRecord `json:"binding,omitempty"`
	ResolvedBy    string             `json:"resolved_by"`
	RequestHost   string             `json:"request_host"`
	AccessGranted bool               `json:"access_granted"`
}

type InitializeResult struct {
	SourceSpaceKey         string `json:"source_space_key"`
	TargetSpaceKey         string `json:"target_space_key"`
	ForceReinitialized     bool   `json:"force_reinitialized"`
	ClearedMenuCount       int    `json:"cleared_menu_count"`
	ClearedPageCount       int    `json:"cleared_page_count"`
	ClearedPackageMenuLink int    `json:"cleared_package_menu_link_count"`
	CreatedMenuCount       int    `json:"created_menu_count"`
	CreatedPageCount       int    `json:"created_page_count"`
	CreatedPackageMenuLink int    `json:"created_package_menu_link_count"`
}

type SaveSpaceRequest struct {
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
	ListSpaces() ([]SpaceRecord, error)
	GetCurrent(host string, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CurrentResponse, error)
	ListHostBindings() ([]HostBindingRecord, error)
	GetMode() (string, error)
	SaveMode(mode string) (string, error)
	SaveSpace(req *SaveSpaceRequest) (*SpaceRecord, error)
	SaveHostBinding(req *SaveHostBindingRequest) (*HostBindingRecord, error)
	InitializeFromDefault(targetSpaceKey string, force bool, actorUserID *uuid.UUID) (*InitializeResult, error)
}

type service struct {
	db        *gorm.DB
	refresher permissionrefresh.Service
	logger    *zap.Logger
}

func NewService(db *gorm.DB, refresher permissionrefresh.Service, logger *zap.Logger) Service {
	return &service{db: db, refresher: refresher, logger: logger}
}

func EnsureDefaultMenuSpace(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	defaultSpace := models.MenuSpace{
		SpaceKey:        DefaultMenuSpaceKey,
		Name:            "默认菜单空间",
		Description:     "兼容当前单域单菜单运行模式",
		DefaultHomePath: "/dashboard/console",
		IsDefault:       true,
		Status:          "normal",
		Meta:            models.MetaJSON{},
	}
	var existing models.MenuSpace
	err := db.Where("space_key = ?", DefaultMenuSpaceKey).First(&existing).Error
	if err == nil {
		updates := map[string]interface{}{
			"is_default": true,
			"status":     "normal",
		}
		return db.Model(&models.MenuSpace{}).Where("space_key = ?", DefaultMenuSpaceKey).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&defaultSpace).Error
}

func (s *service) ListSpaces() ([]SpaceRecord, error) {
	if err := EnsureDefaultMenuSpace(s.db); err != nil {
		return nil, err
	}
	var spaces []models.MenuSpace
	if err := s.db.Order("is_default DESC, created_at ASC").Find(&spaces).Error; err != nil {
		return nil, err
	}
	var bindings []models.MenuSpaceHostBinding
	if err := s.db.Where("deleted_at IS NULL").Order("created_at ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	menuCountMap, err := s.loadMenuCountMap()
	if err != nil {
		return nil, err
	}
	pageCountMap, err := s.loadPageCountMap()
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]string, len(spaces))
	for _, binding := range bindings {
		key := NormalizeSpaceKey(binding.SpaceKey)
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

func (s *service) GetCurrent(host string, requestedSpaceKey string, userID *uuid.UUID, tenantID *uuid.UUID) (*CurrentResponse, error) {
	if err := EnsureDefaultMenuSpace(s.db); err != nil {
		return nil, err
	}
	if IsSingleSpaceMode(s.db) {
		explicit := NormalizeSpaceKey(requestedSpaceKey)
		if explicit != "" && explicit != DefaultMenuSpaceKey {
			ok, checkErr := spaceExists(s.db, explicit)
			if checkErr != nil {
				return nil, checkErr
			}
			if ok {
				allowed, accessErr := CanAccessSpace(s.db, userID, tenantID, explicit)
				if accessErr != nil {
					return nil, accessErr
				}
				if allowed {
					spaceRecord, err := s.getSpaceRecord(explicit)
					if err != nil {
						return nil, err
					}
					return &CurrentResponse{
						Space:         *spaceRecord,
						Binding:       nil,
						ResolvedBy:    "single_mode_explicit",
						RequestHost:   NormalizeHost(host),
						AccessGranted: true,
					}, nil
				}
			}
		}
		spaceRecord, err := s.getSpaceRecord(DefaultMenuSpaceKey)
		if err != nil {
			return nil, err
		}
		return &CurrentResponse{
			Space:         *spaceRecord,
			Binding:       nil,
			ResolvedBy:    "single_mode",
			RequestHost:   NormalizeHost(host),
			AccessGranted: true,
		}, nil
	}
	explicit := NormalizeSpaceKey(requestedSpaceKey)
	resolvedKey, binding, err := ResolveSpaceKeyByHost(s.db, host)
	if err != nil {
		return nil, err
	}
	resolvedBy := "default"
	if explicit == DefaultMenuSpaceKey {
		resolvedKey = explicit
		binding = nil
		resolvedBy = "explicit"
	} else if explicit != "" {
		ok, checkErr := spaceExists(s.db, explicit)
		if checkErr != nil {
			return nil, checkErr
		}
		if ok {
			resolvedKey = explicit
			binding = nil
			resolvedBy = "explicit"
		}
	} else if binding != nil {
		resolvedBy = "host"
	}
	accessGranted, accessErr := CanAccessSpace(s.db, userID, tenantID, resolvedKey)
	if accessErr != nil {
		return nil, accessErr
	}
	if !accessGranted {
		resolvedKey = DefaultMenuSpaceKey
		binding = nil
		resolvedBy = "fallback_default"
	}

	spaceRecord, err := s.getSpaceRecord(resolvedKey)
	if err != nil {
		return nil, err
	}

	var bindingRecord *HostBindingRecord
	if binding != nil {
		bindingRecord = &HostBindingRecord{
			MenuSpaceHostBinding: *binding,
		}
		if spaceRecord.SpaceKey != "" {
			bindingRecord.SpaceName = spaceRecord.Name
		}
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

func (s *service) ListHostBindings() ([]HostBindingRecord, error) {
	if err := EnsureDefaultMenuSpace(s.db); err != nil {
		return nil, err
	}
	var bindings []models.MenuSpaceHostBinding
	if err := s.db.Order("created_at ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	spaceMap, err := s.loadSpaceNameMap()
	if err != nil {
		return nil, err
	}
	records := make([]HostBindingRecord, 0, len(bindings))
	for _, item := range bindings {
		records = append(records, HostBindingRecord{
			MenuSpaceHostBinding: item,
			SpaceName:            spaceMap[NormalizeSpaceKey(item.SpaceKey)],
		})
	}
	return records, nil
}

func (s *service) SaveSpace(req *SaveSpaceRequest) (*SpaceRecord, error) {
	if req == nil {
		return nil, fmt.Errorf("space request is required")
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
				Where("space_key <> ?", record.SpaceKey).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "space_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"description",
				"default_home_path",
				"is_default",
				"status",
				"meta",
				"updated_at",
			}),
		}).Create(&record).Error
	}); err != nil {
		return nil, err
	}

	return s.getSpaceRecord(key)
}

func (s *service) SaveHostBinding(req *SaveHostBindingRequest) (*HostBindingRecord, error) {
	if req == nil {
		return nil, fmt.Errorf("host binding request is required")
	}
	host := NormalizeHost(req.Host)
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}
	spaceKey := NormalizeSpaceKey(req.SpaceKey)
	if spaceKey == "" {
		spaceKey = DefaultMenuSpaceKey
	}
	ok, err := spaceExists(s.db, spaceKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("menu space not found: %s", spaceKey)
	}
	var targetSpace models.MenuSpace
	if err := s.db.Select("space_key", "status").Where("space_key = ?", spaceKey).First(&targetSpace).Error; err != nil {
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
	var existingBinding models.MenuSpaceHostBinding
	if err := s.db.Where("host = ? AND deleted_at IS NULL", host).First(&existingBinding).Error; err == nil {
		if NormalizeSpaceKey(existingBinding.SpaceKey) != spaceKey {
			return nil, fmt.Errorf("该 Host 已绑定到菜单空间 %s，请先解除原绑定后再调整", NormalizeSpaceKey(existingBinding.SpaceKey))
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
	binding := models.MenuSpaceHostBinding{
		SpaceKey:    spaceKey,
		Host:        host,
		Description: strings.TrimSpace(req.Description),
		IsDefault:   req.IsDefault,
		Status:      status,
		Meta:        meta,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if binding.IsDefault {
			if err := tx.Model(&models.MenuSpaceHostBinding{}).
				Where("space_key = ? AND host <> ?", spaceKey, host).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "host"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"space_key",
				"description",
				"is_default",
				"status",
				"meta",
				"updated_at",
			}),
		}).Create(&binding).Error
	}); err != nil {
		return nil, err
	}

	return s.getHostBindingRecord(host)
}

func (s *service) InitializeFromDefault(targetSpaceKey string, force bool, actorUserID *uuid.UUID) (*InitializeResult, error) {
	targetKey := NormalizeSpaceKey(targetSpaceKey)
	if targetKey == "" {
		return nil, fmt.Errorf("target space is required")
	}
	if targetKey == DefaultMenuSpaceKey {
		return nil, fmt.Errorf("默认菜单空间无需初始化")
	}
	ok, err := spaceExists(s.db, targetKey)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("menu space not found: %s", targetKey)
	}

	result := &InitializeResult{
		SourceSpaceKey:     DefaultMenuSpaceKey,
		TargetSpaceKey:     targetKey,
		ForceReinitialized: force,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		var existingMenuCount int64
		if err := tx.Model(&models.Menu{}).
			Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, targetKey).
			Count(&existingMenuCount).Error; err != nil {
			return err
		}
		if existingMenuCount > 0 && !force {
			return fmt.Errorf("目标空间已存在菜单，请先清空后再初始化")
		}

		var existingPageCount int64
		if err := tx.Model(&models.UIPage{}).
			Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, targetKey).
			Count(&existingPageCount).Error; err != nil {
			return err
		}
		if existingPageCount > 0 && !force {
			return fmt.Errorf("目标空间已存在页面，请先清空后再初始化")
		}
		if force {
			var targetMenus []models.Menu
			if err := tx.Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, targetKey).
				Find(&targetMenus).Error; err != nil {
				return err
			}
			targetMenuIDs := make([]uuid.UUID, 0, len(targetMenus))
			for _, item := range targetMenus {
				targetMenuIDs = append(targetMenuIDs, item.ID)
			}
			if len(targetMenuIDs) > 0 {
				var linkCount int64
				if err := tx.Model(&models.FeaturePackageMenu{}).
					Where("menu_id IN ?", targetMenuIDs).
					Count(&linkCount).Error; err != nil {
					return err
				}
				result.ClearedPackageMenuLink = int(linkCount)
				if err := tx.Where("menu_id IN ?", targetMenuIDs).Delete(&models.FeaturePackageMenu{}).Error; err != nil {
					return err
				}
			}
			if existingPageCount > 0 {
				result.ClearedPageCount = int(existingPageCount)
				if err := tx.Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, targetKey).
					Delete(&models.UIPage{}).Error; err != nil {
					return err
				}
			}
			if existingMenuCount > 0 {
				result.ClearedMenuCount = int(existingMenuCount)
				if err := tx.Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, targetKey).
					Delete(&models.Menu{}).Error; err != nil {
					return err
				}
			}
		}

		var sourceMenus []models.Menu
		if err := tx.Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, DefaultMenuSpaceKey).
			Order("sort_order ASC, created_at ASC").
			Find(&sourceMenus).Error; err != nil {
			return err
		}

		menuIDMap := make(map[uuid.UUID]uuid.UUID, len(sourceMenus))
		clonedMenus := make([]models.Menu, 0, len(sourceMenus))
		for _, item := range sourceMenus {
			newID := uuid.New()
			menuIDMap[item.ID] = newID
			clonedMenus = append(clonedMenus, models.Menu{
				ID:            newID,
				ParentID:      nil,
				ManageGroupID: item.ManageGroupID,
				SpaceKey:      targetKey,
				Path:          item.Path,
				Name:          item.Name,
				Component:     item.Component,
				Title:         item.Title,
				Icon:          item.Icon,
				SortOrder:     item.SortOrder,
				Hidden:        item.Hidden,
				Meta:          cloneMetaJSON(item.Meta),
			})
		}
		for index, item := range sourceMenus {
			if item.ParentID == nil {
				continue
			}
			if newParentID, ok := menuIDMap[*item.ParentID]; ok {
				clonedMenus[index].ParentID = &newParentID
			}
		}
		if len(clonedMenus) > 0 {
			if err := tx.Create(&clonedMenus).Error; err != nil {
				return err
			}
		}
		result.CreatedMenuCount = len(clonedMenus)

		var sourcePages []models.UIPage
		if err := tx.Where("COALESCE(NULLIF(space_key, ''), ?) = ?", DefaultMenuSpaceKey, DefaultMenuSpaceKey).
			Order("sort_order ASC, created_at ASC").
			Find(&sourcePages).Error; err != nil {
			return err
		}
		spaceSuffix := sanitizeSpaceKeySegment(targetKey)
		pageKeyMap := make(map[string]string, len(sourcePages))
		for _, item := range sourcePages {
			pageKeyMap[item.PageKey] = fmt.Sprintf("%s.%s", item.PageKey, spaceSuffix)
		}
		clonedPages := make([]models.UIPage, 0, len(sourcePages))
		for _, item := range sourcePages {
			cloned := models.UIPage{
				ID:                uuid.New(),
				PageKey:           pageKeyMap[item.PageKey],
				Name:              item.Name,
				RouteName:         buildClonedRouteName(item.RouteName, spaceSuffix),
				RoutePath:         item.RoutePath,
				Component:         item.Component,
				SpaceKey:          targetKey,
				PageType:          item.PageType,
				Source:            item.Source,
				ModuleKey:         item.ModuleKey,
				SortOrder:         item.SortOrder,
				ParentMenuID:      nil,
				ParentPageKey:     remapStringKey(pageKeyMap, item.ParentPageKey),
				DisplayGroupKey:   remapStringKey(pageKeyMap, item.DisplayGroupKey),
				ActiveMenuPath:    item.ActiveMenuPath,
				BreadcrumbMode:    item.BreadcrumbMode,
				AccessMode:        item.AccessMode,
				PermissionKey:     item.PermissionKey,
				InheritPermission: item.InheritPermission,
				KeepAlive:         item.KeepAlive,
				IsFullPage:        item.IsFullPage,
				Status:            item.Status,
				Meta:              cloneMetaJSON(item.Meta),
			}
			if item.ParentMenuID != nil {
				if newParentMenuID, ok := menuIDMap[*item.ParentMenuID]; ok {
					cloned.ParentMenuID = &newParentMenuID
				}
			}
			clonedPages = append(clonedPages, cloned)
		}
		if len(clonedPages) > 0 {
			if err := tx.Create(&clonedPages).Error; err != nil {
				return err
			}
		}
		result.CreatedPageCount = len(clonedPages)

		var sourcePackageMenus []models.FeaturePackageMenu
		sourceMenuIDs := make([]uuid.UUID, 0, len(sourceMenus))
		for _, item := range sourceMenus {
			sourceMenuIDs = append(sourceMenuIDs, item.ID)
		}
		if len(sourceMenuIDs) > 0 {
			if err := tx.Where("menu_id IN ?", sourceMenuIDs).Find(&sourcePackageMenus).Error; err != nil {
				return err
			}
		}
		clonedPackageMenus := make([]models.FeaturePackageMenu, 0, len(sourcePackageMenus))
		seenPackageMenuPairs := make(map[string]struct{}, len(sourcePackageMenus))
		for _, item := range sourcePackageMenus {
			newMenuID, ok := menuIDMap[item.MenuID]
			if !ok {
				continue
			}
			pairKey := item.PackageID.String() + ":" + newMenuID.String()
			if _, exists := seenPackageMenuPairs[pairKey]; exists {
				continue
			}
			seenPackageMenuPairs[pairKey] = struct{}{}
			clonedPackageMenus = append(clonedPackageMenus, models.FeaturePackageMenu{
				PackageID: item.PackageID,
				MenuID:    newMenuID,
			})
		}
		if len(clonedPackageMenus) > 0 {
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&clonedPackageMenus).Error; err != nil {
				return err
			}
		}
		result.CreatedPackageMenuLink = len(clonedPackageMenus)
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

func (s *service) getSpaceRecord(spaceKey string) (*SpaceRecord, error) {
	var item models.MenuSpace
	if err := s.db.Where("space_key = ?", NormalizeSpaceKey(spaceKey)).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("menu space not found: %s", spaceKey)
		}
		return nil, err
	}
	var bindings []models.MenuSpaceHostBinding
	if err := s.db.Where("space_key = ? AND deleted_at IS NULL", item.SpaceKey).Order("created_at ASC").Find(&bindings).Error; err != nil {
		return nil, err
	}
	hosts := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		hosts = append(hosts, binding.Host)
	}
	sort.Strings(hosts)
	menuCountMap, err := s.loadMenuCountMap()
	if err != nil {
		return nil, err
	}
	pageCountMap, err := s.loadPageCountMap()
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

func (s *service) getHostBindingRecord(host string) (*HostBindingRecord, error) {
	var binding models.MenuSpaceHostBinding
	if err := s.db.Where("host = ?", NormalizeHost(host)).First(&binding).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("host binding not found: %s", host)
		}
		return nil, err
	}
	var space models.MenuSpace
	spaceName := ""
	if err := s.db.Where("space_key = ?", NormalizeSpaceKey(binding.SpaceKey)).First(&space).Error; err == nil {
		spaceName = space.Name
	}
	return &HostBindingRecord{
		MenuSpaceHostBinding: binding,
		SpaceName:            spaceName,
	}, nil
}

func (s *service) loadSpaceNameMap() (map[string]string, error) {
	var spaces []models.MenuSpace
	if err := s.db.Select("space_key", "name").Find(&spaces).Error; err != nil {
		return nil, err
	}
	result := make(map[string]string, len(spaces))
	for _, item := range spaces {
		result[NormalizeSpaceKey(item.SpaceKey)] = item.Name
	}
	return result, nil
}

func (s *service) loadMenuCountMap() (map[string]int, error) {
	type aggregate struct {
		SpaceKey string
		Total    int
	}
	var rows []aggregate
	if err := s.db.Model(&models.Menu{}).
		Select("space_key, COUNT(*) AS total").
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

func (s *service) loadPageCountMap() (map[string]int, error) {
	type aggregate struct {
		SpaceKey string
		Total    int
	}
	var rows []aggregate
	if err := s.db.Model(&models.UIPage{}).
		Select("space_key, COUNT(*) AS total").
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

func (s *service) notifySpaceOperation(actorUserID uuid.UUID, result *InitializeResult) error {
	if result == nil || actorUserID == uuid.Nil {
		return nil
	}
	spaceRecord, err := s.getSpaceRecord(result.TargetSpaceKey)
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
		fmt.Sprintf("新建页面：%d", result.CreatedPageCount),
		fmt.Sprintf("新建功能包菜单关联：%d", result.CreatedPackageMenuLink),
	}
	if result.ForceReinitialized {
		actionText = "重新初始化"
		summary = fmt.Sprintf("菜单空间“%s”已完成重新初始化，原有菜单与页面已被默认空间覆盖。", spaceRecord.Name)
		contentLines = append([]string{
			fmt.Sprintf("已清空菜单：%d", result.ClearedMenuCount),
			fmt.Sprintf("已清空页面：%d", result.ClearedPageCount),
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

func sanitizeSpaceKeySegment(value string) string {
	target := strings.TrimSpace(strings.ToLower(value))
	target = strings.ReplaceAll(target, "-", "_")
	target = strings.ReplaceAll(target, ".", "_")
	target = strings.ReplaceAll(target, "/", "_")
	target = strings.Trim(target, "_")
	if target == "" {
		return DefaultMenuSpaceKey
	}
	return target
}

func buildClonedRouteName(routeName string, spaceSuffix string) string {
	target := strings.TrimSpace(routeName)
	if target == "" {
		target = "Page"
	}
	return target + "__" + strings.Title(spaceSuffix)
}

func remapStringKey(mapping map[string]string, source string) string {
	target := strings.TrimSpace(source)
	if target == "" {
		return ""
	}
	if remapped, ok := mapping[target]; ok {
		return remapped
	}
	return ""
}
