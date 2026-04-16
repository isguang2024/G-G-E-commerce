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
	ErrTypeNotFound                = errors.New("字典类型不存在")
	ErrBuiltinReadonly             = errors.New("内置字典不可删除")
	ErrCodeDuplicate               = errors.New("字典编码已存在")
	ErrItemNotFound                = errors.New("字典项不存在")
	ErrItemLabelRequired           = errors.New("字典项标签不能为空")
	ErrItemValueRequired           = errors.New("字典项值不能为空")
	ErrItemValueDuplicate          = errors.New("同批次内字典项值不能重复")
	ErrItemBuiltinReadonly         = errors.New("内置字典项不可删除")
	ErrItemDeleteRequiresSuspended = errors.New("请先停用字典项，再执行删除")
	ErrItemBuiltinValueImmutable   = errors.New("内置字典项的值不可修改")
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
	Label       string
	Value       string
	Description string
	Extra       models.MetaJSON
	IsDefault   bool
	IsBuiltin   bool
	Status      string
	SortOrder   int
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

	existingItems, err := s.repo.ListItems(ctx, dictTypeID)
	if err != nil {
		return nil, err
	}
	existingByValue := make(map[string]models.DictItem, len(existingItems))
	for _, item := range existingItems {
		existingByValue[strings.TrimSpace(item.Value)] = item
	}
	inputByValue := make(map[string]DictItemInput, len(inputs))
	for _, in := range inputs {
		inputByValue[strings.TrimSpace(in.Value)] = in
	}
	for _, item := range existingItems {
		in, ok := inputByValue[strings.TrimSpace(item.Value)]
		if !ok {
			if item.IsBuiltin {
				return nil, ErrItemBuiltinReadonly
			}
			if strings.TrimSpace(item.Status) != "suspended" {
				return nil, ErrItemDeleteRequiresSuspended
			}
			continue
		}
		if item.IsBuiltin && strings.TrimSpace(in.Value) != strings.TrimSpace(item.Value) {
			return nil, ErrItemBuiltinValueImmutable
		}
	}

	items := make([]models.DictItem, len(inputs))
	for i, in := range inputs {
		status := in.Status
		if status == "" {
			status = "normal"
		}
		existingItem, existed := existingByValue[strings.TrimSpace(in.Value)]
		items[i] = models.DictItem{
			Label:       in.Label,
			Value:       in.Value,
			Description: strings.TrimSpace(in.Description),
			IsBuiltin:   existed && existingItem.IsBuiltin || (!existed && in.IsBuiltin),
			Extra:       in.Extra,
			IsDefault:   in.IsDefault,
			Status:      status,
			SortOrder:   in.SortOrder,
		}
	}
	result, err := s.repo.BatchReplaceItems(ctx, dictTypeID, items)
	if err != nil {
		return nil, err
	}
	s.invalidateCache(dt.Code)
	return result, nil
}

func (s *Service) CreateItem(ctx context.Context, dictTypeID uuid.UUID, input DictItemInput) (*models.DictItem, error) {
	if _, err := s.repo.GetTypeByID(ctx, dictTypeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTypeNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(input.Label) == "" {
		return nil, ErrItemLabelRequired
	}
	if strings.TrimSpace(input.Value) == "" {
		return nil, ErrItemValueRequired
	}
	existing, err := s.repo.ListItems(ctx, dictTypeID)
	if err != nil {
		return nil, err
	}
	for _, item := range existing {
		if strings.EqualFold(strings.TrimSpace(item.Value), strings.TrimSpace(input.Value)) {
			return nil, fmt.Errorf("%w: %s", ErrItemValueDuplicate, input.Value)
		}
	}
	item := &models.DictItem{
		DictTypeID:  dictTypeID,
		Label:       strings.TrimSpace(input.Label),
		Value:       strings.TrimSpace(input.Value),
		Description: strings.TrimSpace(input.Description),
		Extra:       input.Extra,
		IsDefault:   input.IsDefault,
		IsBuiltin:   false,
		Status:      firstNonEmptyStatus(input.Status),
		SortOrder:   input.SortOrder,
	}
	if err := s.repo.CreateItem(ctx, item); err != nil {
		return nil, err
	}
	if dt, err := s.repo.GetTypeByID(ctx, dictTypeID); err == nil {
		s.invalidateCache(dt.Code)
	}
	return item, nil
}

func (s *Service) UpdateItem(ctx context.Context, dictTypeID, itemID uuid.UUID, input DictItemInput) (*models.DictItem, error) {
	item, err := s.repo.GetItemByID(ctx, dictTypeID, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrItemNotFound
		}
		return nil, err
	}
	if strings.TrimSpace(input.Label) == "" {
		return nil, ErrItemLabelRequired
	}
	if strings.TrimSpace(input.Value) == "" {
		return nil, ErrItemValueRequired
	}
	if item.IsBuiltin && strings.TrimSpace(input.Value) != strings.TrimSpace(item.Value) {
		return nil, ErrItemBuiltinValueImmutable
	}
	existing, err := s.repo.ListItems(ctx, dictTypeID)
	if err != nil {
		return nil, err
	}
	for _, sibling := range existing {
		if sibling.ID == item.ID {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(sibling.Value), strings.TrimSpace(input.Value)) {
			return nil, fmt.Errorf("%w: %s", ErrItemValueDuplicate, input.Value)
		}
	}
	item.Label = strings.TrimSpace(input.Label)
	item.Value = strings.TrimSpace(input.Value)
	item.Description = strings.TrimSpace(input.Description)
	item.Extra = input.Extra
	item.IsDefault = input.IsDefault
	item.Status = firstNonEmptyStatus(input.Status)
	item.SortOrder = input.SortOrder
	if err := s.repo.UpdateItem(ctx, item); err != nil {
		return nil, err
	}
	if dt, err := s.repo.GetTypeByID(ctx, dictTypeID); err == nil {
		s.invalidateCache(dt.Code)
	}
	return item, nil
}

func (s *Service) DeleteItem(ctx context.Context, dictTypeID, itemID uuid.UUID) error {
	item, err := s.repo.GetItemByID(ctx, dictTypeID, itemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrItemNotFound
		}
		return err
	}
	if item.IsBuiltin {
		return ErrItemBuiltinReadonly
	}
	if strings.TrimSpace(item.Status) != "suspended" {
		return ErrItemDeleteRequiresSuspended
	}
	if err := s.repo.DeleteItem(ctx, dictTypeID, itemID); err != nil {
		return err
	}
	if dt, err := s.repo.GetTypeByID(ctx, dictTypeID); err == nil {
		s.invalidateCache(dt.Code)
	}
	return nil
}

func firstNonEmptyStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "normal"
	}
	return strings.TrimSpace(status)
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
			{Label: "正常", Value: "normal", Description: "默认可用状态，参与正常展示与选择。", SortOrder: 1},
			{Label: "停用", Value: "suspended", Description: "禁用状态，通常用于下线或临时关闭。", SortOrder: 2},
		},
	},
	{
		Code: "gender", Name: "性别", Description: "性别选项",
		Items: []DictItemInput{
			{Label: "男", Value: "male", Description: "男性用户标识。", SortOrder: 1},
			{Label: "女", Value: "female", Description: "女性用户标识。", SortOrder: 2},
		},
	},
	{
		Code: "page_type", Name: "页面类型", Description: "页面类型枚举",
		Items: []DictItemInput{
			{Label: "分组", Value: "group", Description: "用于组织页面层级的目录型节点。", SortOrder: 1},
			{Label: "展示分组", Value: "display_group", Description: "仅用于前端展示分组，不直接承载页面。", SortOrder: 2},
			{Label: "内页", Value: "inner", Description: "挂载在菜单体系中的业务页面。", SortOrder: 3},
			{Label: "独立页", Value: "standalone", Description: "不依赖菜单层级的独立访问页面。", SortOrder: 4},
		},
	},
	{
		Code: "access_mode", Name: "访问模式", Description: "页面/接口访问模式",
		Items: []DictItemInput{
			{Label: "继承", Value: "inherit", Description: "继承父级或上游配置决定访问控制。", SortOrder: 1},
			{Label: "公开", Value: "public", Description: "无需登录即可访问。", SortOrder: 2},
			{Label: "JWT认证", Value: "jwt", Description: "要求已登录并持有有效 JWT。", SortOrder: 3},
			{Label: "权限", Value: "permission", Description: "要求通过权限点校验后访问。", SortOrder: 4},
		},
	},
	{
		Code: "page_source", Name: "页面来源", Description: "页面创建来源",
		Items: []DictItemInput{
			{Label: "手动", Value: "manual", Description: "由管理员手工创建或维护。", SortOrder: 1},
			{Label: "同步", Value: "sync", Description: "由系统同步流程自动写入。", SortOrder: 2},
			{Label: "种子", Value: "seed", Description: "来自默认 seed 或初始化脚本。", SortOrder: 3},
			{Label: "远程", Value: "remote", Description: "来自远端系统或注册中心。", SortOrder: 4},
		},
	},
	{
		Code: "http_method", Name: "HTTP方法", Description: "HTTP 请求方法",
		Items: []DictItemInput{
			{Label: "GET", Value: "GET", Description: "读取资源，不应产生副作用。", SortOrder: 1},
			{Label: "POST", Value: "POST", Description: "创建资源或提交动作请求。", SortOrder: 2},
			{Label: "PUT", Value: "PUT", Description: "整体更新资源。", SortOrder: 3},
			{Label: "PATCH", Value: "PATCH", Description: "局部更新资源。", SortOrder: 4},
			{Label: "DELETE", Value: "DELETE", Description: "删除资源。", SortOrder: 5},
		},
	},
	{
		Code: "message_type", Name: "消息类型", Description: "消息分类",
		Items: []DictItemInput{
			{Label: "通知", Value: "notice", Description: "广播型通知，强调告知。", SortOrder: 1},
			{Label: "消息", Value: "message", Description: "普通消息沟通记录。", SortOrder: 2},
			{Label: "待办", Value: "todo", Description: "需要用户处理的动作项。", SortOrder: 3},
		},
	},
	{
		Code: "workspace_plan", Name: "空间套餐", Description: "工作空间套餐等级",
		Items: []DictItemInput{
			{Label: "免费版", Value: "free", Description: "默认基础套餐，适合轻量使用场景。", SortOrder: 1},
			{Label: "专业版", Value: "pro", Description: "提供更多协作与管理能力的进阶套餐。", SortOrder: 2},
			{Label: "企业版", Value: "enterprise", Description: "面向组织治理与定制化能力的企业套餐。", SortOrder: 3},
		},
	},
	{
		Code: "register_source", Name: "注册来源", Description: "用户注册来源标识",
		Items: []DictItemInput{
			{Label: "自注册", Value: "self", Description: "用户通过公开注册页自主完成注册。", SortOrder: 1, IsDefault: true},
			{Label: "邀请注册", Value: "invite", Description: "用户通过邀请码或邀请链路进入注册。", SortOrder: 2},
			{Label: "管理员添加", Value: "admin", Description: "由后台管理员直接创建用户。", SortOrder: 3},
			{Label: "邮箱注册", Value: "email", Description: "通过邮箱验证流程完成注册。", SortOrder: 4},
			{Label: "短信注册", Value: "sms", Description: "通过短信验证码流程完成注册。", SortOrder: 5},
			{Label: "第三方登录注册", Value: "oauth", Description: "通过第三方身份提供商完成注册。", SortOrder: 6},
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
				Label:       in.Label,
				Value:       in.Value,
				Description: strings.TrimSpace(in.Description),
				IsDefault:   in.IsDefault,
				IsBuiltin:   true,
				Status:      "normal",
				SortOrder:   in.SortOrder,
			}
		}
		if _, err := s.repo.BatchReplaceItems(ctx, dt.ID, items); err != nil {
			return fmt.Errorf("seed items for %s: %w", bd.Code, err)
		}
		s.logger.Info("seeded builtin dict", zap.String("code", bd.Code), zap.Int("items", len(items)))
	}
	return nil
}
