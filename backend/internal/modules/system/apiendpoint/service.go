package apiendpoint

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/modules/system/user"
	"github.com/maben/backend/internal/pkg/apiendpointaccess"
	"github.com/maben/backend/internal/pkg/permissionseed"
)

var (
	ErrNoStaleCleanupSelection = errors.New("请选择要清理的失效 API")
	ErrStaleCleanupTargetGone  = errors.New("所选 API 已不是失效状态，请刷新后重试")
)

// endpointStatusStale is the value written to api_endpoints.status by the seed
// ensure step when an existing row no longer matches any OpenAPI operation.
const endpointStatusStale = "stale"

type ListRequest struct {
	Current           int
	Size              int
	PermissionKey     string
	PermissionPattern string
	Keyword           string
	Method            string
	Path              string
	CategoryID        string
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
	AppKey  string
}

type Service interface {
	List(req *ListRequest) ([]user.APIEndpoint, int64, error)
	Overview(appKey string) (*EndpointOverview, error)
	ListRuntimeStates(endpoints []user.APIEndpoint) map[uuid.UUID]EndpointRuntimeState
	ListStale(req *StaleListRequest) ([]user.APIEndpoint, int64, error)
	ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error)
	GetUnregisteredScanConfig() (UnregisteredScanConfig, error)
	SaveUnregisteredScanConfig(config UnregisteredScanConfig) (UnregisteredScanConfig, error)
	ListBindingsByEndpointCodes(endpointCodes []string) ([]user.APIEndpointPermissionBinding, error)
	ListBindings(endpointCode string) ([]user.APIEndpointPermissionBinding, error)
	ListCategories() ([]user.APIEndpointCategory, error)
	Save(endpoint *user.APIEndpoint, permissionKeys []string, currentAppKey string) (*user.APIEndpoint, error)
	SaveCategory(item *user.APIEndpointCategory) (*user.APIEndpointCategory, error)
	Sync() (*SyncSummary, error)
	CleanupStale(endpointIDs []uuid.UUID, appKey string) (int, error)
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
	logger         *zap.Logger
	env            string
	endpointAccess apiendpointaccess.Service
}

func NewService(db *gorm.DB, repo user.APIEndpointRepository, categoryRepo user.APIEndpointCategoryRepository, bindingRepo user.APIEndpointPermissionBindingRepository, logger *zap.Logger, env string, endpointAccess apiendpointaccess.Service) Service {
	return &service{
		db:             db,
		repo:           repo,
		categoryRepo:   categoryRepo,
		bindingRepo:    bindingRepo,
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

func (s *service) Overview(appKey string) (*EndpointOverview, error) {
	_ = appKey
	overview := &EndpointOverview{
		CategoryCounts: make([]EndpointCategoryCount, 0),
	}
	baseQuery := s.db.Model(&user.APIEndpoint{})
	if err := baseQuery.Count(&overview.TotalCount).Error; err != nil {
		return nil, err
	}

	// 统一未分类桶：category_id IS NULL 与 category_id = uncategorized.id 都归入 UncategorizedCount，
	// 前端只渲染一个"未分类"节点，不重复。若 uncategorized 分类尚未 seed（全新库），仅计 NULL。
	var uncategorizedID *uuid.UUID
	if cat, err := s.categoryRepo.GetByCode("uncategorized"); err == nil && cat != nil {
		id := cat.ID
		uncategorizedID = &id
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	uncategorizedQuery := s.db.Model(&user.APIEndpoint{})
	if uncategorizedID != nil {
		uncategorizedQuery = uncategorizedQuery.Where("category_id IS NULL OR category_id = ?", *uncategorizedID)
	} else {
		uncategorizedQuery = uncategorizedQuery.Where("category_id IS NULL")
	}
	if err := uncategorizedQuery.Count(&overview.UncategorizedCount).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&user.APIEndpoint{}).Where("status = ?", endpointStatusStale).Count(&overview.StaleCount).Error; err != nil {
		return nil, err
	}

	type categoryCountRow struct {
		CategoryID *uuid.UUID `gorm:"column:category_id"`
		Count      int64      `gorm:"column:count"`
	}
	rows := make([]categoryCountRow, 0)
	categoryCountsQuery := s.db.Model(&user.APIEndpoint{}).
		Select("category_id, COUNT(*) AS count").
		Where("category_id IS NOT NULL")
	if uncategorizedID != nil {
		// 排除 uncategorized 分类本身，避免它与 UncategorizedCount 双计。
		categoryCountsQuery = categoryCountsQuery.Where("category_id <> ?", *uncategorizedID)
	}
	if err := categoryCountsQuery.
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
		profile := buildPermissionProfile(endpoint.Method, endpoint.Path, bindingMap[endpoint.Code], nil)
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

// ListRuntimeStates derives each endpoint's runtime state from its persisted
// `status` column. The column is authoritative — seed-driven ensure writes
// `stale` to rows that no longer match any OpenAPI operation, and resets
// everything else to `normal`.
func (s *service) ListRuntimeStates(endpoints []user.APIEndpoint) map[uuid.UUID]EndpointRuntimeState {
	result := make(map[uuid.UUID]EndpointRuntimeState, len(endpoints))
	for _, endpoint := range endpoints {
		if strings.TrimSpace(endpoint.Status) == endpointStatusStale {
			result[endpoint.ID] = EndpointRuntimeState{
				Stale:       true,
				StaleReason: "OpenAPI spec 中已不存在该端点",
			}
			continue
		}
		result[endpoint.ID] = EndpointRuntimeState{RuntimeExists: true}
	}
	return result
}

// ListUnregisteredRoutes always returns an empty page. Routes are now derived
// from the OpenAPI spec + seed, so every mounted route is by definition already
// registered in api_endpoints after a successful migration.
func (s *service) ListUnregisteredRoutes(req *UnregisteredRouteListRequest) ([]UnregisteredRouteItem, int64, error) {
	return []UnregisteredRouteItem{}, 0, nil
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
	_ = req.AppKey

	var total int64
	if err := s.db.Model(&user.APIEndpoint{}).Where("status = ?", endpointStatusStale).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]user.APIEndpoint, 0)
	if err := s.db.Model(&user.APIEndpoint{}).
		Where("status = ?", endpointStatusStale).
		Order("path ASC, method ASC").
		Offset((req.Current - 1) * req.Size).
		Limit(req.Size).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Sync re-runs the seed ensure pipeline against the current OpenAPI seed,
// then refreshes the runtime permission cache. Previously this shelled out to
// the now-deleted Registrar-based SyncRoutes that walked gin.Routes(); the
// seed-driven flow is the sole source of truth.
func (s *service) Sync() (*SyncSummary, error) {
	processed, err := permissionseed.EnsureOpenAPIEndpoints(s.db)
	if err != nil {
		return nil, err
	}
	if _, err := permissionseed.EnsureOpenAPIPermissionBindings(s.db); err != nil {
		return nil, err
	}
	if err := s.refreshRuntimeCache(); err != nil {
		return nil, err
	}

	var total int64
	if err := s.db.Model(&user.APIEndpoint{}).Count(&total).Error; err != nil {
		return nil, err
	}
	summary := &SyncSummary{
		Processed: processed,
		Total:     int(total),
	}
	if s.logger != nil {
		s.logger.Info("API endpoints synced via OpenAPI seed ensure",
			zap.Int("processed", summary.Processed),
			zap.Int("total", summary.Total),
		)
	}
	return summary, nil
}

// CleanupStale deletes the stale rows selected by endpointIDs. A row is
// considered stale when its `status` column is "stale" — set by the seed
// ensure pass for operations no longer present in the OpenAPI spec.
func (s *service) CleanupStale(endpointIDs []uuid.UUID, appKey string) (int, error) {
	_ = appKey
	if len(endpointIDs) == 0 {
		return 0, ErrNoStaleCleanupSelection
	}

	filtered := make([]uuid.UUID, 0, len(endpointIDs))
	for _, id := range endpointIDs {
		if id != uuid.Nil {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return 0, ErrStaleCleanupTargetGone
	}

	var stale []user.APIEndpoint
	if err := s.db.
		Where("status = ? AND id IN ?", endpointStatusStale, filtered).
		Find(&stale).Error; err != nil {
		return 0, err
	}
	if len(stale) == 0 {
		return 0, ErrStaleCleanupTargetGone
	}

	staleIDs := make([]uuid.UUID, 0, len(stale))
	staleCodes := make([]string, 0, len(stale))
	for _, endpoint := range stale {
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
		return tx.Where("id IN ?", staleIDs).Delete(&user.APIEndpoint{}).Error
	}); err != nil {
		return 0, err
	}

	if err := s.refreshRuntimeCache(); err != nil {
		return 0, err
	}
	return len(stale), nil
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

func (s *service) Save(endpoint *user.APIEndpoint, permissionKeys []string, currentAppKey string) (*user.APIEndpoint, error) {
	if endpoint == nil {
		return nil, errors.New("端点参数不能为空")
	}
	endpoint.Method = strings.ToUpper(strings.TrimSpace(endpoint.Method))
	endpoint.Path = strings.TrimSpace(endpoint.Path)
	if endpoint.Method == "" || endpoint.Path == "" {
		return nil, errors.New("method 和 path 不能为空")
	}
	if endpoint.Status == "" {
		endpoint.Status = "normal"
	}
	_ = currentAppKey
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
			"code":        endpoint.Code,
			"method":      endpoint.Method,
			"path":        endpoint.Path,
			"handler":     endpoint.Handler,
			"summary":     endpoint.Summary,
			"category_id": endpoint.CategoryID,
			"status":      endpoint.Status,
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
		return nil, errors.New("分类参数不能为空")
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
