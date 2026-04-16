package dictionary

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
)

var (
	ErrTypeNotFound       = errors.New("字典类型不存在")
	ErrBuiltinReadonly    = errors.New("内置字典不可删除")
	ErrCodeDuplicate      = errors.New("字典编码已存在")
	ErrItemLabelRequired  = errors.New("字典项标签不能为空")
	ErrItemValueRequired  = errors.New("字典项值不能为空")
	ErrItemValueDuplicate = errors.New("同批次内字典项值不能重复")
)

// Service handles dictionary business logic.
type Service struct {
	repo   *Repository
	logger *zap.Logger

	// Local memory cache for GetByCodes (code → cached items).
	cacheMu  sync.RWMutex
	cache    map[string]*cacheEntry
	cacheTTL time.Duration
}

type cacheEntry struct {
	items     []models.DictItem
	expiresAt time.Time
}

func NewService(db *gorm.DB, logger *zap.Logger) *Service {
	return &Service{
		repo:     NewRepository(db),
		logger:   logger,
		cache:    make(map[string]*cacheEntry),
		cacheTTL: 1 * time.Hour,
	}
}

// ─── Dict Type management ────────────────────────────────────────────────────

type DictItemInput struct {
	Label     string
	Value     string
	Extra     models.MetaJSON
	IsDefault bool
	Status    string
	SortOrder int
}

func (s *Service) ListTypes(ctx context.Context, current, size int, keyword, status string) ([]models.DictType, int64, error) {
	if current < 1 {
		current = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (current - 1) * size
	return s.repo.ListTypes(ctx, offset, size, keyword, status)
}

func (s *Service) GetType(ctx context.Context, id uuid.UUID) (*models.DictType, []models.DictItem, error) {
	dt, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, ErrTypeNotFound
		}
		return nil, nil, err
	}
	items, err := s.repo.ListItems(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return dt, items, nil
}

func (s *Service) CreateType(ctx context.Context, code, name, description, status string, sortOrder int) (*models.DictType, error) {
	// Check duplicate code
	_, err := s.repo.GetTypeByCode(ctx, code)
	if err == nil {
		return nil, ErrCodeDuplicate
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if status == "" {
		status = "normal"
	}
	dt := &models.DictType{
		Code:        code,
		Name:        name,
		Description: description,
		Status:      status,
		SortOrder:   sortOrder,
	}
	if err := s.repo.CreateType(ctx, dt); err != nil {
		return nil, err
	}
	return dt, nil
}

func (s *Service) UpdateType(ctx context.Context, id uuid.UUID, name, description, status string, sortOrder int) (*models.DictType, error) {
	dt, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTypeNotFound
		}
		return nil, err
	}
	dt.Name = name
	dt.Description = description
	if status != "" {
		dt.Status = status
	}
	dt.SortOrder = sortOrder
	if err := s.repo.UpdateType(ctx, dt); err != nil {
		return nil, err
	}
	// Invalidate cache for this type's code
	s.invalidateCache(dt.Code)
	return dt, nil
}

func (s *Service) DeleteType(ctx context.Context, id uuid.UUID) error {
	dt, err := s.repo.GetTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTypeNotFound
		}
		return err
	}
	if dt.IsBuiltin {
		return ErrBuiltinReadonly
	}
	// Delete items first, then type
	if err := s.repo.DeleteItemsByTypeID(ctx, id); err != nil {
		return fmt.Errorf("delete items: %w", err)
	}
	if err := s.repo.DeleteType(ctx, id); err != nil {
		return err
	}
	s.invalidateCache(dt.Code)
	return nil
}

// ─── Dict Item management ────────────────────────────────────────────────────

func (s *Service) ListItems(ctx context.Context, dictTypeID uuid.UUID) ([]models.DictItem, error) {
	// Verify type exists
	if _, err := s.repo.GetTypeByID(ctx, dictTypeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTypeNotFound
		}
		return nil, err
	}
	return s.repo.ListItems(ctx, dictTypeID)
}

func (s *Service) SaveItems(ctx context.Context, dictTypeID uuid.UUID, inputs []DictItemInput) ([]models.DictItem, error) {
	dt, err := s.repo.GetTypeByID(ctx, dictTypeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTypeNotFound
		}
		return nil, err
	}
	// Validate inputs
	seen := make(map[string]struct{}, len(inputs))
	for _, in := range inputs {
		if strings.TrimSpace(in.Label) == "" {
			return nil, ErrItemLabelRequired
		}
		if strings.TrimSpace(in.Value) == "" {
			return nil, ErrItemValueRequired
		}
		if _, dup := seen[in.Value]; dup {
			return nil, fmt.Errorf("%w: %s", ErrItemValueDuplicate, in.Value)
		}
		seen[in.Value] = struct{}{}
	}

	items := make([]models.DictItem, len(inputs))
	for i, in := range inputs {
		status := in.Status
		if status == "" {
			status = "normal"
		}
		items[i] = models.DictItem{
			Label:     in.Label,
			Value:     in.Value,
			Extra:     in.Extra,
			IsDefault: in.IsDefault,
			Status:    status,
			SortOrder: in.SortOrder,
		}
	}
	result, err := s.repo.BatchReplaceItems(ctx, dictTypeID, items)
	if err != nil {
		return nil, err
	}
	s.invalidateCache(dt.Code)
	return result, nil
}

// ─── Consumer (batch query by codes) ─────────────────────────────────────────

func (s *Service) GetByCodes(ctx context.Context, codes []string) (map[string][]models.DictItem, error) {
	// Check cache first
	result := make(map[string][]models.DictItem, len(codes))
	var missingCodes []string

	s.cacheMu.RLock()
	now := time.Now()
	for _, code := range codes {
		if entry, ok := s.cache[code]; ok && now.Before(entry.expiresAt) {
			result[code] = entry.items
		} else {
			missingCodes = append(missingCodes, code)
		}
	}
	s.cacheMu.RUnlock()

	if len(missingCodes) == 0 {
		return result, nil
	}

	// Fetch missing from DB
	fetched, err := s.repo.GetItemsByTypeCodes(ctx, missingCodes)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.cacheMu.Lock()
	expiresAt := now.Add(s.cacheTTL)
	for code, items := range fetched {
		s.cache[code] = &cacheEntry{items: items, expiresAt: expiresAt}
		result[code] = items
	}
	s.cacheMu.Unlock()

	return result, nil
}

func (s *Service) invalidateCache(code string) {
	s.cacheMu.Lock()
	delete(s.cache, code)
	s.cacheMu.Unlock()
}

// CountItemsBatch exposes repository's batch count for handler use.
func (s *Service) CountItemsBatch(ctx context.Context, typeIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	return s.repo.CountItemsBatch(ctx, typeIDs)
}

// CountItems returns item count for a single type.
func (s *Service) CountItems(ctx context.Context, typeID uuid.UUID) (int64, error) {
	return s.repo.CountItems(ctx, typeID)
}

// ─── Seed built-in dictionaries ──────────────────────────────────────────────

type builtinDict struct {
	Code        string
	Name        string
	Description string
	Items       []DictItemInput
}

var builtinDicts = []builtinDict{
	{
		Code: "common_status", Name: "通用状态", Description: "通用启用/停用状态",
		Items: []DictItemInput{
			{Label: "正常", Value: "normal", SortOrder: 1},
			{Label: "停用", Value: "suspended", SortOrder: 2},
		},
	},
	{
		Code: "gender", Name: "性别", Description: "性别选项",
		Items: []DictItemInput{
			{Label: "男", Value: "male", SortOrder: 1},
			{Label: "女", Value: "female", SortOrder: 2},
		},
	},
	{
		Code: "page_type", Name: "页面类型", Description: "页面类型��举",
		Items: []DictItemInput{
			{Label: "分组", Value: "group", SortOrder: 1},
			{Label: "展示分组", Value: "display_group", SortOrder: 2},
			{Label: "内页", Value: "inner", SortOrder: 3},
			{Label: "独立页", Value: "standalone", SortOrder: 4},
		},
	},
	{
		Code: "access_mode", Name: "访问模式", Description: "页面/接口访问模式",
		Items: []DictItemInput{
			{Label: "继承", Value: "inherit", SortOrder: 1},
			{Label: "公开", Value: "public", SortOrder: 2},
			{Label: "JWT认证", Value: "jwt", SortOrder: 3},
			{Label: "权限", Value: "permission", SortOrder: 4},
		},
	},
	{
		Code: "page_source", Name: "页面来源", Description: "页面创建来源",
		Items: []DictItemInput{
			{Label: "手动", Value: "manual", SortOrder: 1},
			{Label: "同步", Value: "sync", SortOrder: 2},
			{Label: "种子", Value: "seed", SortOrder: 3},
			{Label: "远程", Value: "remote", SortOrder: 4},
		},
	},
	{
		Code: "http_method", Name: "HTTP方法", Description: "HTTP 请求方法",
		Items: []DictItemInput{
			{Label: "GET", Value: "GET", SortOrder: 1},
			{Label: "POST", Value: "POST", SortOrder: 2},
			{Label: "PUT", Value: "PUT", SortOrder: 3},
			{Label: "PATCH", Value: "PATCH", SortOrder: 4},
			{Label: "DELETE", Value: "DELETE", SortOrder: 5},
		},
	},
	{
		Code: "message_type", Name: "消息类型", Description: "消息分类",
		Items: []DictItemInput{
			{Label: "通知", Value: "notice", SortOrder: 1},
			{Label: "消息", Value: "message", SortOrder: 2},
			{Label: "待办", Value: "todo", SortOrder: 3},
		},
	},
	{
		Code: "workspace_plan", Name: "空间套餐", Description: "工作空间套餐等级",
		Items: []DictItemInput{
			{Label: "免费版", Value: "free", SortOrder: 1},
			{Label: "专业版", Value: "pro", SortOrder: 2},
			{Label: "企业版", Value: "enterprise", SortOrder: 3},
		},
	},
}

// EnsureBuiltinDicts idempotently creates built-in dictionaries.
// Call this during migration or server startup.
func (s *Service) EnsureBuiltinDicts(ctx context.Context) error {
	for _, bd := range builtinDicts {
		existing, err := s.repo.GetTypeByCode(ctx, bd.Code)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("check builtin dict %s: %w", bd.Code, err)
		}
		if existing != nil {
			// Already exists — skip
			continue
		}
		dt := &models.DictType{
			Code:        bd.Code,
			Name:        bd.Name,
			Description: bd.Description,
			IsBuiltin:   true,
			Status:      "normal",
		}
		if err := s.repo.CreateType(ctx, dt); err != nil {
			return fmt.Errorf("create builtin dict %s: %w", bd.Code, err)
		}
		items := make([]models.DictItem, len(bd.Items))
		for i, in := range bd.Items {
			items[i] = models.DictItem{
				Label:     in.Label,
				Value:     in.Value,
				Status:    "normal",
				SortOrder: in.SortOrder,
			}
		}
		if _, err := s.repo.BatchReplaceItems(ctx, dt.ID, items); err != nil {
			return fmt.Errorf("seed items for %s: %w", bd.Code, err)
		}
		s.logger.Info("seeded builtin dict", zap.String("code", bd.Code), zap.Int("items", len(items)))
	}
	return nil
}

