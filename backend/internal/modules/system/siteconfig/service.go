package siteconfig

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/internal/config"
	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/pkg/appctx"
	pkgLogger "github.com/maben/backend/internal/pkg/logger"
)

// 值来源
const (
	ResolveSourceApp     = "app"
	ResolveSourceGlobal  = "global"
	ResolveSourceDefault = "default"

	ScopeTypeContext = "context"
	ScopeTypeGlobal  = "global"
	ScopeTypeApp     = "app"
	ScopeTypeAll     = "all"
)

// Service 参数管理服务接口。
type Service interface {
	// Resolve 按 keys/set_codes 批量解析（支持 context/global/app 作用域）。
	Resolve(ctx context.Context, req ResolveRequest) (*ResolveResult, error)
	// Lookup 按作用域解析单个参数。
	Lookup(ctx context.Context, req LookupRequest) (*LookupResult, error)
	// ListConfigs 列参数项。
	ListConfigs(ctx context.Context, tenantID string, req ListRequest) ([]models.SiteConfig, error)
	GetConfig(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfig, error)
	UpsertConfig(ctx context.Context, cfg *models.SiteConfig) error
	DeleteConfig(ctx context.Context, tenantID string, id uuid.UUID) error

	ListSets(ctx context.Context, tenantID string) ([]SetWithItems, error)
	GetSet(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfigSet, error)
	UpsertSet(ctx context.Context, set *models.SiteConfigSet) error
	DeleteSet(ctx context.Context, tenantID string, id uuid.UUID) error
	UpdateSetItems(ctx context.Context, tenantID string, setID uuid.UUID, keys []string) error

	// Cache 返回底层缓存（可为 nil），用于在主程序里启动 Pub/Sub。
	Cache() *resolvedConfigCache
}

// ResolveRequest 批量解析入参。
type ResolveRequest struct {
	ScopeType string
	ScopeKey  string
	Keys      []string
	SetCodes  []string
}

// ResolveResult 批量解析结果。
type ResolveResult struct {
	Items     map[string]ResolvedItem `json:"items"`
	Version   string                  `json:"version"`
	ScopeType string                  `json:"scope_type"`
	ScopeKey  string                  `json:"scope_key"`
}

// LookupRequest 单键解析入参。
type LookupRequest struct {
	ScopeType string
	ScopeKey  string
	ConfigKey string
}

// LookupResult 单键解析结果。
type LookupResult struct {
	ConfigKey string       `json:"config_key"`
	Item      ResolvedItem `json:"item"`
	ScopeType string       `json:"scope_type"`
	ScopeKey  string       `json:"scope_key"`
}

// ListRequest 管理端参数查询入参。
type ListRequest struct {
	ScopeType string
	ScopeKey  string
}

// ResolvedItem 单个 key 的解析结果。
type ResolvedItem struct {
	Value     models.MetaJSON `json:"value"`
	Source    string          `json:"source"`
	ValueType string          `json:"value_type"`
	Sets      []string        `json:"sets"`
}

// SetWithItems 集合 + 其关联的 config_keys。
type SetWithItems struct {
	Set        models.SiteConfigSet `json:"set"`
	ConfigKeys []string             `json:"config_keys"`
}

// service 实现。
type service struct {
	repo   *Repository
	cache  *resolvedConfigCache
	logger *zap.Logger
}

// NewService 装配 Repository 和可选的 Redis 缓存。
func NewService(repo *Repository, cfg *config.Config, logger *zap.Logger) Service {
	s := &service{
		repo:   repo,
		logger: logger,
	}
	if cfg != nil {
		cache, err := newResolvedConfigCache(cfg, logger)
		if err != nil {
			if logger != nil {
				logger.Warn("initialize site config cache failed", zap.Error(err))
			}
		} else if cache != nil {
			s.cache = cache
		}
	}
	return s
}

// NewServiceWithCache 测试用：注入指定缓存。
func NewServiceWithCache(repo *Repository, cache *resolvedConfigCache, logger *zap.Logger) Service {
	return &service{repo: repo, cache: cache, logger: logger}
}

func (s *service) Cache() *resolvedConfigCache { return s.cache }

// ---- Resolve ----

func (s *service) Resolve(ctx context.Context, req ResolveRequest) (*ResolveResult, error) {
	tenantID := tenantFromCtx(ctx)
	scopeType, scopeKey, appKey, err := resolveRuntimeScope(ctx, req.ScopeType, req.ScopeKey)
	if err != nil {
		return nil, err
	}
	keys := compactStrings(req.Keys)
	setCodes := compactStrings(req.SetCodes)

	cacheKey := siteConfigResolvedCache(tenantID, scopeType, scopeKey, keys, setCodes)
	if s.cache != nil {
		var cached ResolveResult
		if s.cache.Get(ctx, cacheKey, &cached) {
			return &cached, nil
		}
	}

	// 从集合展开 keys
	var setKeyMap map[string][]string // key -> [set_code, ...]
	if len(setCodes) > 0 {
		sets, err := s.repo.GetSetsByCodes(ctx, tenantID, setCodes)
		if err != nil {
			return nil, err
		}
		if len(sets) > 0 {
			setIDs := make([]uuid.UUID, 0, len(sets))
			setCodeByID := make(map[uuid.UUID]string, len(sets))
			for _, set := range sets {
				setIDs = append(setIDs, set.ID)
				setCodeByID[set.ID] = set.SetCode
			}
			items, err := s.repo.ListItemsBySetIDs(ctx, tenantID, setIDs)
			if err != nil {
				return nil, err
			}
			setKeyMap = make(map[string][]string)
			for _, item := range items {
				code := setCodeByID[item.SetID]
				keys = appendUnique(keys, item.ConfigKey)
				setKeyMap[item.ConfigKey] = appendUnique(setKeyMap[item.ConfigKey], code)
			}
		}
	}

	// 反查 key → 所属集合（覆盖用户直接传入的 keys）
	if len(keys) > 0 {
		items, err := s.repo.ListItemsByConfigKey(ctx, tenantID, keys)
		if err != nil {
			return nil, err
		}
		if len(items) > 0 && setKeyMap == nil {
			setKeyMap = make(map[string][]string)
		}
		if len(items) > 0 {
			// 需要知道 set_id -> set_code
			setIDs := uniqueSetIDs(items)
			idToCode, err := s.loadSetCodes(ctx, tenantID, setIDs)
			if err != nil {
				return nil, err
			}
			for _, item := range items {
				code := idToCode[item.SetID]
				if code == "" {
					continue
				}
				setKeyMap[item.ConfigKey] = appendUnique(setKeyMap[item.ConfigKey], code)
			}
		}
	}

	items := make(map[string]ResolvedItem)
	if len(keys) > 0 {
		appLevel, global, err := s.repo.ResolveByKeys(ctx, tenantID, appKey, keys)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			var picked *models.SiteConfig
			source := ResolveSourceDefault
			if appKey != "" {
				if cfg, ok := appLevel[k]; ok {
					picked = cfg
					source = ResolveSourceApp
				}
			}
			if picked == nil {
				if appKey == "" {
					if cfg, ok := global[k]; ok {
						picked = cfg
						source = ResolveSourceGlobal
					}
				} else if cfg, ok := global[k]; ok && canFallbackToGlobal(cfg) {
					picked = cfg
					source = ResolveSourceGlobal
				}
			}
			sets := setKeyMap[k]
			if sets == nil {
				sets = []string{}
			}
			if picked == nil {
				items[k] = ResolvedItem{
					Value:     models.MetaJSON{},
					Source:    source,
					ValueType: models.SiteConfigValueTypeString,
					Sets:      sets,
				}
				continue
			}
			value := picked.ConfigValue
			if value == nil {
				value = models.MetaJSON{}
			}
			items[k] = ResolvedItem{
				Value:     value,
				Source:    source,
				ValueType: picked.ValueType,
				Sets:      sets,
			}
		}
	}

	result := &ResolveResult{
		Items:     items,
		Version:   fingerprintResolveInput(keys, setCodes),
		ScopeType: scopeType,
		ScopeKey:  scopeKey,
	}
	if s.cache != nil {
		s.cache.Set(ctx, cacheKey, result)
	}
	return result, nil
}

func (s *service) Lookup(ctx context.Context, req LookupRequest) (*LookupResult, error) {
	configKey := strings.TrimSpace(req.ConfigKey)
	if configKey == "" {
		return nil, errors.New("config_key is required")
	}
	resolved, err := s.Resolve(ctx, ResolveRequest{
		ScopeType: req.ScopeType,
		ScopeKey:  req.ScopeKey,
		Keys:      []string{configKey},
	})
	if err != nil {
		return nil, err
	}
	item, ok := resolved.Items[configKey]
	if !ok {
		item = ResolvedItem{
			Value:     models.MetaJSON{},
			Source:    ResolveSourceDefault,
			ValueType: models.SiteConfigValueTypeString,
			Sets:      []string{},
		}
	}
	return &LookupResult{
		ConfigKey: configKey,
		Item:      item,
		ScopeType: resolved.ScopeType,
		ScopeKey:  resolved.ScopeKey,
	}, nil
}

// loadSetCodes 按 set_ids 查 set_code。
func (s *service) loadSetCodes(ctx context.Context, tenantID string, ids []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]string{}, nil
	}
	var sets []models.SiteConfigSet
	err := s.repo.db.WithContext(ctx).
		Where("tenant_id = ? AND id IN ?", normalizeTenantID(tenantID), ids).
		Find(&sets).Error
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]string, len(sets))
	for _, set := range sets {
		out[set.ID] = set.SetCode
	}
	return out, nil
}

// ---- Configs CRUD ----

func (s *service) ListConfigs(ctx context.Context, tenantID string, req ListRequest) ([]models.SiteConfig, error) {
	scopeType, scopeKey, err := normalizeListScope(req.ScopeType, req.ScopeKey)
	if err != nil {
		return nil, err
	}
	return s.repo.ListConfigs(ctx, tenantOr(ctx, tenantID), scopeType, scopeKey)
}

func (s *service) GetConfig(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfig, error) {
	return s.repo.GetConfig(ctx, tenantOr(ctx, tenantID), id)
}

func (s *service) UpsertConfig(ctx context.Context, cfg *models.SiteConfig) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	cfg.TenantID = tenantOr(ctx, cfg.TenantID)
	if err := s.repo.UpsertConfig(ctx, cfg); err != nil {
		return err
	}
	s.invalidateOnConfigWrite(ctx, cfg.TenantID, cfg.AppKey, cfg.ConfigKey)
	return nil
}

func (s *service) DeleteConfig(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = tenantOr(ctx, tenantID)
	cfg, err := s.repo.DeleteConfig(ctx, tenantID, id)
	if err != nil {
		return err
	}
	s.invalidateOnConfigWrite(ctx, cfg.TenantID, cfg.AppKey, cfg.ConfigKey)
	return nil
}

// invalidateOnConfigWrite：
//   - 写应用级 (app_key=X, config_key=K)：失效 per-key `:{tenant}:X:K` + resolved 前缀 `:{tenant}:X:`
//   - 写全局 (app_key=”, config_key=K)：失效 per-key `:{tenant}:_global:K` + **所有 app 的 resolved**
//     （可继承参数可能从全局读取；保守起见继续清空所有 app 的缓存）
func (s *service) invalidateOnConfigWrite(ctx context.Context, tenantID, appKey, configKey string) {
	if s.cache == nil {
		return
	}
	keys := []string{siteConfigKeyCache(tenantID, appKey, configKey)}
	var prefixes []string
	if appKey == "" {
		// 全局配置变更 → 所有 app 的 resolved 都要清
		prefixes = []string{siteConfigResolvedPrefix(tenantID)}
	} else {
		prefixes = []string{siteConfigResolvedPrefixApp(tenantID, appKey)}
	}
	s.cache.Invalidate(ctx, keys, prefixes)
}

// ---- Sets CRUD ----

func (s *service) ListSets(ctx context.Context, tenantID string) ([]SetWithItems, error) {
	tenantID = tenantOr(ctx, tenantID)
	sets, err := s.repo.ListSets(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if len(sets) == 0 {
		return []SetWithItems{}, nil
	}
	setIDs := make([]uuid.UUID, 0, len(sets))
	for _, set := range sets {
		setIDs = append(setIDs, set.ID)
	}
	items, err := s.repo.ListItemsBySetIDs(ctx, tenantID, setIDs)
	if err != nil {
		return nil, err
	}
	keysBySet := make(map[uuid.UUID][]string)
	for _, item := range items {
		keysBySet[item.SetID] = append(keysBySet[item.SetID], item.ConfigKey)
	}
	out := make([]SetWithItems, 0, len(sets))
	for _, set := range sets {
		keys := keysBySet[set.ID]
		if keys == nil {
			keys = []string{}
		}
		out = append(out, SetWithItems{Set: set, ConfigKeys: keys})
	}
	return out, nil
}

func (s *service) GetSet(ctx context.Context, tenantID string, id uuid.UUID) (*models.SiteConfigSet, error) {
	return s.repo.GetSet(ctx, tenantOr(ctx, tenantID), id)
}

func (s *service) UpsertSet(ctx context.Context, set *models.SiteConfigSet) error {
	if set == nil {
		return errors.New("set is nil")
	}
	set.TenantID = tenantOr(ctx, set.TenantID)
	if err := s.repo.UpsertSet(ctx, set); err != nil {
		return err
	}
	s.invalidateOnSetWrite(ctx, set.TenantID)
	return nil
}

func (s *service) DeleteSet(ctx context.Context, tenantID string, id uuid.UUID) error {
	tenantID = tenantOr(ctx, tenantID)
	_, err := s.repo.DeleteSet(ctx, tenantID, id)
	if err != nil {
		return err
	}
	s.invalidateOnSetWrite(ctx, tenantID)
	return nil
}

func (s *service) UpdateSetItems(ctx context.Context, tenantID string, setID uuid.UUID, keys []string) error {
	tenantID = tenantOr(ctx, tenantID)
	if _, err := s.repo.GetSet(ctx, tenantID, setID); err != nil {
		return err
	}
	if err := s.repo.ReplaceSetItems(ctx, tenantID, setID, keys); err != nil {
		return err
	}
	s.invalidateOnSetWrite(ctx, tenantID)
	return nil
}

// invalidateOnSetWrite：集合 / items 变更后清掉对应 tenant 的所有 resolved，并刷新 sets 索引缓存。
func (s *service) invalidateOnSetWrite(ctx context.Context, tenantID string) {
	if s.cache == nil {
		return
	}
	s.cache.Invalidate(ctx,
		[]string{siteConfigSetsCache(tenantID)},
		[]string{siteConfigResolvedPrefix(tenantID)},
	)
}

// ---- helpers ----

func uniqueSetIDs(items []models.SiteConfigSetItem) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(items))
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.SetID]; ok {
			continue
		}
		seen[item.SetID] = struct{}{}
		out = append(out, item.SetID)
	}
	return out
}

func appendUnique(list []string, item string) []string {
	for _, v := range list {
		if v == item {
			return list
		}
	}
	return append(list, item)
}

func normalizeListScope(scopeType, scopeKey string) (string, string, error) {
	switch normalizeListScopeType(scopeType) {
	case ScopeTypeAll:
		return ScopeTypeAll, "", nil
	case ScopeTypeApp:
		target := appctx.NormalizeExplicitAppKey(scopeKey)
		if target == "" {
			return "", "", errors.New("scope_key is required when scope_type=app")
		}
		return ScopeTypeApp, appctx.NormalizeAppKey(target), nil
	default:
		return ScopeTypeGlobal, "", nil
	}
}

func normalizeListScopeType(scopeType string) string {
	switch strings.TrimSpace(strings.ToLower(scopeType)) {
	case ScopeTypeAll:
		return ScopeTypeAll
	case ScopeTypeApp:
		return ScopeTypeApp
	default:
		return ScopeTypeGlobal
	}
}

func canFallbackToGlobal(cfg *models.SiteConfig) bool {
	if cfg == nil {
		return false
	}
	return cfg.FallbackPolicyOrDefault() == models.SiteConfigFallbackPolicyInherit
}

func resolveRuntimeScope(ctx context.Context, scopeType, scopeKey string) (resolvedScopeType, resolvedScopeKey, appKey string, err error) {
	targetScopeType := strings.TrimSpace(strings.ToLower(scopeType))
	targetScopeKey := appctx.NormalizeExplicitAppKey(scopeKey)
	switch targetScopeType {
	case "":
		if targetScopeKey != "" {
			targetScopeType = ScopeTypeApp
		} else {
			targetScopeType = ScopeTypeContext
		}
	case ScopeTypeContext, ScopeTypeGlobal, ScopeTypeApp:
		// valid
	default:
		return "", "", "", fmt.Errorf("invalid scope_type: %s", scopeType)
	}

	switch targetScopeType {
	case ScopeTypeGlobal:
		return ScopeTypeGlobal, "", "", nil
	case ScopeTypeApp:
		if targetScopeKey == "" {
			return "", "", "", errors.New("scope_key is required when scope_type=app")
		}
		appKey = appctx.NormalizeAppKey(targetScopeKey)
		return ScopeTypeApp, appKey, appKey, nil
	default:
		appKey = appctx.NormalizeAppKey(pkgLogger.AppFromContext(ctx))
		return ScopeTypeApp, appKey, appKey, nil
	}
}

// tenantFromCtx 从 context 取 tenant。
func tenantFromCtx(ctx context.Context) string {
	return pkgLogger.TenantFromContext(ctx)
}

func tenantOr(ctx context.Context, fallback string) string {
	if s := strings.TrimSpace(fallback); s != "" {
		return s
	}
	return tenantFromCtx(ctx)
}
