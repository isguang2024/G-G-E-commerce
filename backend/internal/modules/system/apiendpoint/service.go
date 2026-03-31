package apiendpoint

import (
	"crypto/sha1"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/apiendpointaccess"
	"github.com/gg-ecommerce/backend/internal/pkg/apiregistry"
)

var (
	ErrNoStaleCleanupSelection = errors.New("请选择要清理的失效 API")
	ErrStaleCleanupTargetGone  = errors.New("所选 API 已不是失效状态，请刷新后重试")
)

type ListRequest struct {
	Current           int
	Size              int
	PermissionKey     string
	PermissionPattern string
	Keyword           string
	Method            string
	Path              string
	CategoryID        string
	ContextScope      string
	Source            string
	FeatureKind       string
	Status            string
	HasPermission     *bool
	HasCategory       *bool
}

type UnregisteredRouteListRequest struct {
	Current    int
	Size       int
	Method     string
	Path       string
	Keyword    string
	OnlyNoMeta bool
}

type UnregisteredRouteItem struct {
	Method  string                 `json:"method"`
	Path    string                 `json:"path"`
	Spec    string                 `json:"spec"`
	Handler string                 `json:"handler"`
	HasMeta bool                   `json:"has_meta"`
	Meta    map[string]interface{} `json:"meta"`
}

type EndpointRuntimeState struct {
	RuntimeExists bool   `json:"runtime_exists"`
	Stale         bool   `json:"stale"`
	StaleReason   string `json:"stale_reason,omitempty"`
}

type EndpointCategoryCount struct {
	CategoryID string `json:"category_id"`
	Count      int64  `json:"count"`
}

type EndpointOverview struct {
	TotalCount              int64                   `json:"total_count"`
	UncategorizedCount      int64                   `json:"uncategorized_count"`
	StaleCount              int64                   `json:"stale_count"`
	NoPermissionCount       int64                   `json:"no_permission_count"`
	SharedPermissionCount   int64                   `json:"shared_permission_count"`
	CrossContextSharedCount int64                   `json:"cross_context_shared_count"`
	CategoryCounts          []EndpointCategoryCount `json:"category_counts"`
}

type StaleListRequest struct {
	Current int
	Size    int
}

type Service interface {
	List(req *ListRequest) ([]user.APIEndpoint, int64, error)
	Overview() (*EndpointOverview, error)
	ListRuntimeStates(endpoints []user.APIEndpoint) map[uuid.UUID]EndpointRuntimeState
	ListStale(req *StaleListRequest) ([]user.APIEndpoint, int64, error)
	ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error)
	GetUnregisteredScanConfig() (UnregisteredScanConfig, error)
	SaveUnregisteredScanConfig(config UnregisteredScanConfig) (UnregisteredScanConfig, error)
	ListBindingsByEndpointCodes(endpointCodes []string) ([]user.APIEndpointPermissionBinding, error)
	ListBindings(endpointCode string) ([]user.APIEndpointPermissionBinding, error)
	ListCategories() ([]user.APIEndpointCategory, error)
	Save(endpoint *user.APIEndpoint, permissionKeys []string) (*user.APIEndpoint, error)
	SaveCategory(item *user.APIEndpointCategory) (*user.APIEndpointCategory, error)
	UpdateContextScope(endpointID uuid.UUID, contextScope string) (*user.APIEndpoint, error)
	Sync() error
	CleanupStale(endpointIDs []uuid.UUID) (int, error)
}

const unregisteredScanConfigSettingKey = "api.unregistered_scan_config"

type UnregisteredScanConfig struct {
	Enabled              bool   `json:"enabled"`
	FrequencyMinutes     int    `json:"frequency_minutes"`
	DefaultCategoryID    string `json:"default_category_id"`
	DefaultPermissionKey string `json:"default_permission_key"`
	MarkAsNoPermission   bool   `json:"mark_as_no_permission"`
}

type service struct {
	db             *gorm.DB
	repo           user.APIEndpointRepository
	categoryRepo   user.APIEndpointCategoryRepository
	bindingRepo    user.APIEndpointPermissionBindingRepository
	router         *gin.Engine
	logger         *zap.Logger
	env            string
	endpointAccess apiendpointaccess.Service
}

func NewService(db *gorm.DB, repo user.APIEndpointRepository, categoryRepo user.APIEndpointCategoryRepository, bindingRepo user.APIEndpointPermissionBindingRepository, router *gin.Engine, logger *zap.Logger, env string, endpointAccess apiendpointaccess.Service) Service {
	return &service{
		db:             db,
		repo:           repo,
		categoryRepo:   categoryRepo,
		bindingRepo:    bindingRepo,
		router:         router,
		logger:         logger,
		env:            strings.TrimSpace(env),
		endpointAccess: endpointAccess,
	}
}

func (s *service) List(req *ListRequest) ([]user.APIEndpoint, int64, error) {
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	params := &user.APIEndpointListParams{
		Method:            strings.ToUpper(strings.TrimSpace(req.Method)),
		PermissionKey:     strings.TrimSpace(req.PermissionKey),
		PermissionPattern: strings.TrimSpace(req.PermissionPattern),
		Keyword:           strings.TrimSpace(req.Keyword),
		Path:              strings.TrimSpace(req.Path),
		CategoryID:        strings.TrimSpace(req.CategoryID),
		ContextScope:      strings.TrimSpace(req.ContextScope),
		Source:            strings.TrimSpace(req.Source),
		FeatureKind:       strings.TrimSpace(req.FeatureKind),
		Status:            strings.TrimSpace(req.Status),
		HasPermission:     req.HasPermission,
		HasCategory:       req.HasCategory,
	}
	if params.PermissionKey != "" {
		endpointCodes, err := s.bindingRepo.ListEndpointCodesByPermissionKey(params.PermissionKey)
		if err != nil {
			return nil, 0, err
		}
		if len(endpointCodes) == 0 {
			return []user.APIEndpoint{}, 0, nil
		}
		params.EndpointCodes = endpointCodes
	}
	return s.repo.List((req.Current-1)*req.Size, req.Size, params)
}

func (s *service) Overview() (*EndpointOverview, error) {
	overview := &EndpointOverview{
		CategoryCounts: make([]EndpointCategoryCount, 0),
	}
	if err := s.db.Model(&user.APIEndpoint{}).Count(&overview.TotalCount).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&user.APIEndpoint{}).Where("category_id IS NULL").Count(&overview.UncategorizedCount).Error; err != nil {
		return nil, err
	}

	type categoryCountRow struct {
		CategoryID *uuid.UUID `gorm:"column:category_id"`
		Count      int64      `gorm:"column:count"`
	}
	rows := make([]categoryCountRow, 0)
	if err := s.db.
		Model(&user.APIEndpoint{}).
		Select("category_id, COUNT(*) AS count").
		Where("category_id IS NOT NULL").
		Group("category_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.CategoryID == nil {
			continue
		}
		overview.CategoryCounts = append(overview.CategoryCounts, EndpointCategoryCount{
			CategoryID: row.CategoryID.String(),
			Count:      row.Count,
		})
	}

	syncCandidates, err := s.listSyncRuntimeCandidates()
	if err != nil {
		return nil, err
	}
	runtimeStates := s.ListRuntimeStates(syncCandidates)
	overview.StaleCount = int64(len(filterStaleEndpoints(syncCandidates, runtimeStates)))

	endpoints := make([]user.APIEndpoint, 0)
	if err := s.db.Model(&user.APIEndpoint{}).Select("id", "code", "path").Find(&endpoints).Error; err != nil {
		return nil, err
	}
	endpointCodes := make([]string, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if code := strings.TrimSpace(endpoint.Code); code != "" {
			endpointCodes = append(endpointCodes, code)
		}
	}
	bindings, err := s.bindingRepo.ListByEndpointCodes(endpointCodes)
	if err != nil {
		return nil, err
	}
	bindingMap := make(map[string][]string, len(endpointCodes))
	for _, item := range bindings {
		code := strings.TrimSpace(item.EndpointCode)
		if code == "" {
			continue
		}
		bindingMap[code] = append(bindingMap[code], item.PermissionKey)
	}
	for _, endpoint := range endpoints {
		profile := buildPermissionProfile(endpoint.Path, bindingMap[endpoint.Code])
		switch profile.BindingMode {
		case permissionPatternNone, permissionPatternPublic, permissionPatternGlobalJWT, permissionPatternSelfJWT, permissionPatternAPIKey:
			overview.NoPermissionCount++
		case permissionPatternShared:
			overview.SharedPermissionCount++
		case permissionPatternCrossContextShared:
			overview.SharedPermissionCount++
			overview.CrossContextSharedCount++
		}
	}
	return overview, nil
}

func (s *service) ListRuntimeStates(endpoints []user.APIEndpoint) map[uuid.UUID]EndpointRuntimeState {
	result := make(map[uuid.UUID]EndpointRuntimeState, len(endpoints))
	if len(endpoints) == 0 || s.router == nil {
		return result
	}

	index := s.buildRuntimeRouteIndex()
	for _, endpoint := range endpoints {
		result[endpoint.ID] = resolveEndpointRuntimeState(endpoint, index)
	}
	return result
}

func (s *service) ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error) {
	if req == nil {
		req = &UnregisteredRouteListRequest{}
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}
	if s.router == nil {
		return []UnregisteredRouteItem{}, 0, nil
	}

	runtimeRoutes := apiregistry.CollectRuntimeRoutes(s.router.Routes())
	registered, err := s.listPotentiallyRegisteredEndpointsForRoutes(runtimeRoutes)
	if err != nil {
		return nil, 0, err
	}
	registeredCodeSet := make(map[string]struct{}, len(registered))
	legacyRegisteredSet := make(map[string]struct{}, len(registered))
	for _, item := range registered {
		if code := strings.TrimSpace(item.Code); code != "" {
			registeredCodeSet[code] = struct{}{}
			continue
		}
		legacyRegisteredSet[routeSpec(item.Method, item.Path)] = struct{}{}
	}

	methodFilter := strings.ToUpper(strings.TrimSpace(req.Method))
	pathFilter := strings.TrimSpace(req.Path)
	keyword := strings.TrimSpace(req.Keyword)

	items := make([]UnregisteredRouteItem, 0, len(runtimeRoutes))
	for _, route := range runtimeRoutes {
		routeCode := apiregistry.ResolveRouteCode(route.Method, route.Path, routeMetaPointer(route))
		if _, ok := registeredCodeSet[routeCode]; ok {
			continue
		}
		if _, ok := legacyRegisteredSet[routeSpec(route.Method, route.Path)]; ok {
			continue
		}
		if req.OnlyNoMeta && route.HasMeta {
			continue
		}
		if methodFilter != "" && route.Method != methodFilter {
			continue
		}
		if pathFilter != "" && !strings.Contains(route.Path, pathFilter) {
			continue
		}
		if keyword != "" {
			target := strings.ToLower(route.Method + " " + route.Path + " " + route.Handler + " " + route.RouteMeta.Summary + " " + route.RouteMeta.CategoryCode)
			if !strings.Contains(target, strings.ToLower(keyword)) {
				continue
			}
		}
		meta := map[string]interface{}{}
		if route.HasMeta {
			meta["summary"] = route.RouteMeta.Summary
			meta["category_code"] = route.RouteMeta.CategoryCode
			meta["context_scope"] = route.RouteMeta.ContextScope
			meta["source"] = route.RouteMeta.Source
			meta["feature_kind"] = route.RouteMeta.FeatureKind
			meta["permission_keys"] = route.RouteMeta.PermissionKeys
		}
		items = append(items, UnregisteredRouteItem{
			Method:  route.Method,
			Path:    route.Path,
			Spec:    routeSpec(route.Method, route.Path),
			Handler: route.Handler,
			HasMeta: route.HasMeta,
			Meta:    meta,
		})
	}

	total := int64(len(items))
	start := (req.Current - 1) * req.Size
	if start >= len(items) {
		return []UnregisteredRouteItem{}, total, nil
	}
	end := start + req.Size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], total, nil
}

func (s *service) GetUnregisteredScanConfig() (UnregisteredScanConfig, error) {
	var setting models.SystemSetting
	err := s.db.Where("key = ?", unregisteredScanConfigSettingKey).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return defaultUnregisteredScanConfig(), nil
		}
		return UnregisteredScanConfig{}, err
	}
	return normalizeUnregisteredScanConfig(setting.Value), nil
}

func (s *service) SaveUnregisteredScanConfig(config UnregisteredScanConfig) (UnregisteredScanConfig, error) {
	normalized := normalizeUnregisteredScanConfig(models.MetaJSON{
		"enabled":                config.Enabled,
		"frequency_minutes":      config.FrequencyMinutes,
		"default_category_id":    strings.TrimSpace(config.DefaultCategoryID),
		"default_permission_key": strings.TrimSpace(config.DefaultPermissionKey),
		"mark_as_no_permission":  config.MarkAsNoPermission,
	})
	payload := models.MetaJSON{
		"enabled":                normalized.Enabled,
		"frequency_minutes":      normalized.FrequencyMinutes,
		"default_category_id":    normalized.DefaultCategoryID,
		"default_permission_key": normalized.DefaultPermissionKey,
		"mark_as_no_permission":  normalized.MarkAsNoPermission,
	}

	var setting models.SystemSetting
	err := s.db.Where("key = ?", unregisteredScanConfigSettingKey).First(&setting).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			record := models.SystemSetting{
				Key:    unregisteredScanConfigSettingKey,
				Value:  payload,
				Status: "normal",
			}
			if createErr := s.db.Create(&record).Error; createErr != nil {
				return UnregisteredScanConfig{}, createErr
			}
			return normalized, nil
		}
		return UnregisteredScanConfig{}, err
	}
	if err := s.db.Model(&setting).Updates(map[string]interface{}{
		"value":  payload,
		"status": "normal",
	}).Error; err != nil {
		return UnregisteredScanConfig{}, err
	}
	return normalized, nil
}

func (s *service) ListStale(req *StaleListRequest) ([]user.APIEndpoint, int64, error) {
	if req == nil {
		req = &StaleListRequest{}
	}
	if req.Current <= 0 {
		req.Current = 1
	}
	if req.Size <= 0 {
		req.Size = 20
	}

	syncCandidates, err := s.listSyncRuntimeCandidates()
	if err != nil {
		return nil, 0, err
	}
	runtimeStates := s.ListRuntimeStates(syncCandidates)
	staleEndpoints := filterStaleEndpoints(syncCandidates, runtimeStates)
	total := int64(len(staleEndpoints))
	start := (req.Current - 1) * req.Size
	if start >= len(staleEndpoints) {
		return []user.APIEndpoint{}, total, nil
	}
	end := start + req.Size
	if end > len(staleEndpoints) {
		end = len(staleEndpoints)
	}
	return staleEndpoints[start:end], total, nil
}

func (s *service) Sync() error {
	if s.router == nil {
		return nil
	}
	if err := apiregistry.SyncRoutes(s.db, s.logger, s.router.Routes()); err != nil {
		return err
	}
	return s.refreshRuntimeCache()
}

func (s *service) CleanupStale(endpointIDs []uuid.UUID) (int, error) {
	if s.router == nil {
		return 0, nil
	}
	if len(endpointIDs) == 0 {
		return 0, ErrNoStaleCleanupSelection
	}

	endpoints, err := s.listSyncRuntimeCandidates()
	if err != nil {
		return 0, err
	}
	if len(endpoints) == 0 {
		return 0, nil
	}

	runtimeStates := s.ListRuntimeStates(endpoints)
	staleEndpoints := selectStaleEndpointsForCleanup(endpoints, runtimeStates, endpointIDs)
	if len(staleEndpoints) == 0 {
		return 0, ErrStaleCleanupTargetGone
	}
	staleIDs := make([]uuid.UUID, 0, len(staleEndpoints))
	staleCodes := make([]string, 0, len(staleEndpoints))
	for _, endpoint := range staleEndpoints {
		staleIDs = append(staleIDs, endpoint.ID)
		if code := strings.TrimSpace(endpoint.Code); code != "" {
			staleCodes = append(staleCodes, code)
		}
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(staleCodes) > 0 {
			if err := tx.Unscoped().Where("endpoint_code IN ?", staleCodes).Delete(&user.APIEndpointPermissionBinding{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("id IN ?", staleIDs).Delete(&user.APIEndpoint{}).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}

	if err := s.refreshRuntimeCache(); err != nil {
		return 0, err
	}
	return len(staleEndpoints), nil
}

func (s *service) ListBindings(endpointCode string) ([]user.APIEndpointPermissionBinding, error) {
	return s.bindingRepo.ListByEndpointCode(endpointCode)
}

func (s *service) ListBindingsByEndpointCodes(endpointCodes []string) ([]user.APIEndpointPermissionBinding, error) {
	return s.bindingRepo.ListByEndpointCodes(endpointCodes)
}

func (s *service) ListCategories() ([]user.APIEndpointCategory, error) {
	return s.categoryRepo.List()
}

func (s *service) Save(endpoint *user.APIEndpoint, permissionKeys []string) (*user.APIEndpoint, error) {
	if endpoint == nil {
		return nil, errors.New("endpoint is nil")
	}
	endpoint.Method = strings.ToUpper(strings.TrimSpace(endpoint.Method))
	endpoint.Path = strings.TrimSpace(endpoint.Path)
	if endpoint.Method == "" || endpoint.Path == "" {
		return nil, errors.New("method 和 path 不能为空")
	}
	if endpoint.ContextScope == "" {
		endpoint.ContextScope = "optional"
	}
	switch endpoint.ContextScope {
	case "required", "forbidden", "optional":
	default:
		return nil, errors.New("context_scope 仅支持 required|forbidden|optional")
	}
	if endpoint.Source == "" {
		endpoint.Source = "manual"
	}
	switch endpoint.Source {
	case "sync", "seed", "manual":
	default:
		return nil, errors.New("source 仅支持 sync|seed|manual")
	}
	if endpoint.FeatureKind == "" {
		endpoint.FeatureKind = "system"
	}
	switch endpoint.FeatureKind {
	case "system", "business":
	default:
		return nil, errors.New("feature_kind 仅支持 system|business")
	}
	if endpoint.Status == "" {
		endpoint.Status = "normal"
	}
	if endpoint.CategoryID != nil && *endpoint.CategoryID != uuid.Nil {
		if _, err := s.categoryRepo.GetByID(*endpoint.CategoryID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("分类不存在")
			}
			return nil, err
		}
	} else {
		endpoint.CategoryID = nil
	}

	var oldMethod string
	var oldPath string
	var current *user.APIEndpoint
	if endpoint.ID == uuid.Nil {
		endpoint.Code = resolveEndpointCodeForSave(endpoint, nil)
		existsByCode, err := s.repo.GetByCode(endpoint.Code)
		if err == nil && existsByCode != nil {
			return nil, errors.New("code 已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		exists, err := s.repo.GetByMethodAndPath(endpoint.Method, endpoint.Path)
		if err == nil && exists != nil {
			return nil, errors.New("method+path 已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := s.repo.Create(endpoint); err != nil {
			return nil, err
		}
	} else {
		var err error
		current, err = s.repo.GetByID(endpoint.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("API 不存在")
			}
			return nil, err
		}
		endpoint.Code = resolveEndpointCodeForSave(endpoint, current)
		oldMethod = current.Method
		oldPath = current.Path
		existsByCode, err := s.repo.GetByCode(endpoint.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if existsByCode != nil && existsByCode.ID != endpoint.ID {
			return nil, errors.New("code 已存在")
		}
		exists, err := s.repo.GetByMethodAndPath(endpoint.Method, endpoint.Path)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if exists != nil && exists.ID != endpoint.ID {
			return nil, errors.New("method+path 已存在")
		}
		updates := map[string]interface{}{
			"code":          endpoint.Code,
			"method":        endpoint.Method,
			"path":          endpoint.Path,
			"feature_kind":  endpoint.FeatureKind,
			"handler":       endpoint.Handler,
			"summary":       endpoint.Summary,
			"category_id":   endpoint.CategoryID,
			"context_scope": endpoint.ContextScope,
			"source":        endpoint.Source,
			"status":        endpoint.Status,
		}
		if err := s.repo.UpdateWithMap(endpoint.ID, updates); err != nil {
			return nil, err
		}
	}

	keys := make([]string, 0, len(permissionKeys))
	seen := make(map[string]struct{}, len(permissionKeys))
	actionRepo := user.NewPermissionKeyRepository(s.db)
	for _, key := range permissionKeys {
		target := strings.TrimSpace(key)
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		if _, err := actionRepo.GetByPermissionKey(target); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("存在未注册的功能权限键")
			}
			return nil, err
		}
		keys = append(keys, target)
	}
	items := make([]user.APIEndpointPermissionBinding, 0, len(keys))
	for idx, key := range keys {
		items = append(items, user.APIEndpointPermissionBinding{
			EndpointCode:  endpoint.Code,
			PermissionKey: key,
			MatchMode:     "ANY",
			SortOrder:     idx,
		})
	}
	if err := s.bindingRepo.ReplaceByEndpointCode(endpoint.Code, items); err != nil {
		return nil, err
	}
	if err := s.refreshRuntimeCacheForSave(oldMethod, oldPath, endpoint.Method, endpoint.Path); err != nil {
		return nil, err
	}

	return s.repo.GetByID(endpoint.ID)
}

func (s *service) SaveCategory(item *user.APIEndpointCategory) (*user.APIEndpointCategory, error) {
	if item == nil {
		return nil, errors.New("category is nil")
	}
	item.Code = strings.TrimSpace(item.Code)
	item.Name = strings.TrimSpace(item.Name)
	item.NameEn = strings.TrimSpace(item.NameEn)
	if item.Code == "" || item.Name == "" || item.NameEn == "" {
		return nil, errors.New("分类编码、中文名、英文名不能为空")
	}
	if item.Status == "" {
		item.Status = "normal"
	}
	if item.ID == uuid.Nil {
		exists, err := s.categoryRepo.GetByCode(item.Code)
		if err == nil && exists != nil {
			return nil, errors.New("分类编码已存在")
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err := s.categoryRepo.Create(item); err != nil {
			return nil, err
		}
		return s.categoryRepo.GetByID(item.ID)
	}
	exists, err := s.categoryRepo.GetByCode(item.Code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if exists != nil && exists.ID != item.ID {
		return nil, errors.New("分类编码已存在")
	}
	if err := s.categoryRepo.UpdateWithMap(item.ID, map[string]interface{}{
		"code":       item.Code,
		"name":       item.Name,
		"name_en":    item.NameEn,
		"sort_order": item.SortOrder,
		"status":     item.Status,
	}); err != nil {
		return nil, err
	}
	return s.categoryRepo.GetByID(item.ID)
}

func (s *service) UpdateContextScope(endpointID uuid.UUID, contextScope string) (*user.APIEndpoint, error) {
	switch strings.TrimSpace(contextScope) {
	case "required", "forbidden", "optional":
	default:
		return nil, errors.New("context_scope 仅支持 required|forbidden|optional")
	}
	if err := s.repo.UpdateWithMap(endpointID, map[string]interface{}{
		"context_scope": strings.TrimSpace(contextScope),
	}); err != nil {
		return nil, err
	}
	return s.repo.GetByID(endpointID)
}

func deriveStableEndpointCode(method, path string) string {
	targetMethod := strings.ToUpper(strings.TrimSpace(method))
	targetPath := strings.TrimSpace(path)
	return uuid.NewHash(sha1.New(), uuid.NameSpaceURL, []byte("api-endpoint:"+targetMethod+" "+targetPath), 5).String()
}

type runtimeRouteIndex struct {
	codeSet map[string]struct{}
	specSet map[string]struct{}
}

func (s *service) buildRuntimeRouteIndex() runtimeRouteIndex {
	index := runtimeRouteIndex{
		codeSet: make(map[string]struct{}),
		specSet: make(map[string]struct{}),
	}
	if s.router == nil {
		return index
	}

	runtimeRoutes := apiregistry.CollectRuntimeRoutes(s.router.Routes())
	for _, route := range runtimeRoutes {
		index.specSet[routeSpec(route.Method, route.Path)] = struct{}{}
		if code := strings.TrimSpace(apiregistry.ResolveRouteCode(route.Method, route.Path, routeMetaPointer(route))); code != "" {
			index.codeSet[code] = struct{}{}
		}
	}
	return index
}

func resolveEndpointRuntimeState(endpoint user.APIEndpoint, index runtimeRouteIndex) EndpointRuntimeState {
	code := strings.TrimSpace(endpoint.Code)
	if code != "" {
		if _, ok := index.codeSet[code]; ok {
			return EndpointRuntimeState{RuntimeExists: true}
		}
	}

	if _, ok := index.specSet[routeSpec(endpoint.Method, endpoint.Path)]; ok {
		return EndpointRuntimeState{RuntimeExists: true}
	}

	state := EndpointRuntimeState{}
	if shouldMarkEndpointStale(endpoint) {
		state.Stale = true
		state.StaleReason = "源码中已不存在该自动同步 API"
	}
	return state
}

func shouldMarkEndpointStale(endpoint user.APIEndpoint) bool {
	if strings.TrimSpace(endpoint.Source) != "sync" {
		return false
	}
	code := strings.TrimSpace(endpoint.Code)
	if code == "" {
		return false
	}
	if apiregistry.IsFixedManagedRouteCode(code) {
		return true
	}
	_, err := uuid.Parse(code)
	return err == nil
}

func selectStaleEndpointsForCleanup(endpoints []user.APIEndpoint, runtimeStates map[uuid.UUID]EndpointRuntimeState, endpointIDs []uuid.UUID) []user.APIEndpoint {
	staleEndpoints := filterStaleEndpoints(endpoints, runtimeStates)
	if len(staleEndpoints) == 0 || len(endpointIDs) == 0 {
		return []user.APIEndpoint{}
	}
	selectedSet := make(map[uuid.UUID]struct{}, len(endpointIDs))
	for _, endpointID := range endpointIDs {
		if endpointID == uuid.Nil {
			continue
		}
		selectedSet[endpointID] = struct{}{}
	}
	if len(selectedSet) == 0 {
		return []user.APIEndpoint{}
	}

	result := make([]user.APIEndpoint, 0, len(selectedSet))
	for _, endpoint := range staleEndpoints {
		if _, ok := selectedSet[endpoint.ID]; ok {
			result = append(result, endpoint)
		}
	}
	return result
}

func filterStaleEndpoints(endpoints []user.APIEndpoint, runtimeStates map[uuid.UUID]EndpointRuntimeState) []user.APIEndpoint {
	if len(endpoints) == 0 {
		return []user.APIEndpoint{}
	}
	result := make([]user.APIEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if state, ok := runtimeStates[endpoint.ID]; ok && state.Stale {
			result = append(result, endpoint)
		}
	}
	return result
}

func (s *service) listSyncRuntimeCandidates() ([]user.APIEndpoint, error) {
	items := make([]user.APIEndpoint, 0)
	err := s.db.
		Model(&user.APIEndpoint{}).
		Select("id", "code", "method", "path", "category_id", "status", "source").
		Where("source = ?", "sync").
		Order("path ASC, method ASC").
		Find(&items).Error
	return items, err
}

func (s *service) listPotentiallyRegisteredEndpointsForRoutes(runtimeRoutes []apiregistry.RuntimeRoute) ([]user.APIEndpoint, error) {
	if len(runtimeRoutes) == 0 {
		return []user.APIEndpoint{}, nil
	}

	codeSet := make(map[string]struct{}, len(runtimeRoutes))
	methodSet := make(map[string]struct{}, len(runtimeRoutes))
	pathSet := make(map[string]struct{}, len(runtimeRoutes))
	for _, route := range runtimeRoutes {
		if code := strings.TrimSpace(apiregistry.ResolveRouteCode(route.Method, route.Path, routeMetaPointer(route))); code != "" {
			codeSet[code] = struct{}{}
		}
		methodSet[route.Method] = struct{}{}
		pathSet[route.Path] = struct{}{}
	}

	codes := make([]string, 0, len(codeSet))
	for code := range codeSet {
		codes = append(codes, code)
	}
	methods := make([]string, 0, len(methodSet))
	for method := range methodSet {
		methods = append(methods, method)
	}
	paths := make([]string, 0, len(pathSet))
	for path := range pathSet {
		paths = append(paths, path)
	}

	items := make([]user.APIEndpoint, 0, len(runtimeRoutes))
	seen := make(map[uuid.UUID]struct{}, len(runtimeRoutes))
	appendDistinct := func(batch []user.APIEndpoint) {
		for _, item := range batch {
			if _, ok := seen[item.ID]; ok {
				continue
			}
			seen[item.ID] = struct{}{}
			items = append(items, item)
		}
	}

	if len(codes) > 0 {
		batch := make([]user.APIEndpoint, 0, len(codes))
		if err := s.db.
			Model(&user.APIEndpoint{}).
			Select("id", "code", "method", "path").
			Where("code IN ?", codes).
			Find(&batch).Error; err != nil {
			return nil, err
		}
		appendDistinct(batch)
	}
	if len(methods) > 0 && len(paths) > 0 {
		batch := make([]user.APIEndpoint, 0, len(runtimeRoutes))
		if err := s.db.
			Model(&user.APIEndpoint{}).
			Select("id", "code", "method", "path").
			Where("method IN ? AND path IN ?", methods, paths).
			Find(&batch).Error; err != nil {
			return nil, err
		}
		appendDistinct(batch)
	}
	return items, nil
}

func routeSpec(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func routeMetaPointer(route apiregistry.RuntimeRoute) *apiregistry.RouteMeta {
	if !route.HasMeta {
		return nil
	}
	meta := route.RouteMeta
	return &meta
}

func resolveEndpointCodeForSave(endpoint *user.APIEndpoint, current *user.APIEndpoint) string {
	if current != nil {
		if code := strings.TrimSpace(current.Code); code != "" {
			return code
		}
	}
	if endpoint != nil {
		if code := strings.TrimSpace(endpoint.Code); code != "" {
			return code
		}
		return deriveStableEndpointCode(endpoint.Method, endpoint.Path)
	}
	return ""
}

func (s *service) refreshRuntimeCache() error {
	if s.endpointAccess == nil {
		return nil
	}
	return s.endpointAccess.Refresh()
}

func (s *service) refreshRuntimeCacheForSave(oldMethod, oldPath, newMethod, newPath string) error {
	if s.endpointAccess == nil {
		return nil
	}

	oldKeyMethod := strings.ToUpper(strings.TrimSpace(oldMethod))
	oldKeyPath := strings.TrimSpace(oldPath)
	newKeyMethod := strings.ToUpper(strings.TrimSpace(newMethod))
	newKeyPath := strings.TrimSpace(newPath)

	if oldKeyMethod != "" && oldKeyPath != "" && (oldKeyMethod != newKeyMethod || oldKeyPath != newKeyPath) {
		s.endpointAccess.RemoveByRoute(oldKeyMethod, oldKeyPath)
	}

	return s.endpointAccess.RefreshByRoute(newKeyMethod, newKeyPath)
}

func defaultUnregisteredScanConfig() UnregisteredScanConfig {
	return UnregisteredScanConfig{
		Enabled:              false,
		FrequencyMinutes:     60,
		DefaultCategoryID:    "",
		DefaultPermissionKey: "",
		MarkAsNoPermission:   false,
	}
}

func normalizeUnregisteredScanConfig(raw models.MetaJSON) UnregisteredScanConfig {
	result := defaultUnregisteredScanConfig()
	if target, ok := raw["enabled"].(bool); ok {
		result.Enabled = target
	}
	switch value := raw["frequency_minutes"].(type) {
	case int:
		if value >= 5 && value <= 1440 {
			result.FrequencyMinutes = value
		}
	case float64:
		target := int(value)
		if target >= 5 && target <= 1440 {
			result.FrequencyMinutes = target
		}
	}
	if target, ok := raw["default_category_id"].(string); ok {
		result.DefaultCategoryID = strings.TrimSpace(target)
	}
	if target, ok := raw["default_permission_key"].(string); ok {
		result.DefaultPermissionKey = strings.TrimSpace(target)
	}
	if target, ok := raw["mark_as_no_permission"].(bool); ok {
		result.MarkAsNoPermission = target
	}
	return result
}
