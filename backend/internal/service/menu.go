package service

import (
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
)

var ErrMenuNotFound = errors.New("menu not found")
var ErrMenuSystemProtected = errors.New("系统默认菜单不可删除")

// MenuService 菜单服务
type MenuService interface {
	// GetTree all=true 返回全部；all=false 时按 allowedMenuIDs 过滤（角色-菜单权限），nil 表示无权限
	GetTree(all bool, allowedMenuIDs []uuid.UUID) ([]*model.Menu, error)
	Create(req *dto.MenuCreateRequest) (*model.Menu, error)
	Update(id uuid.UUID, req *dto.MenuUpdateRequest) error
	Delete(id uuid.UUID) error
}

type menuService struct {
	menuRepo repository.MenuRepository
	logger   *zap.Logger
}

// NewMenuService 创建菜单服务
func NewMenuService(menuRepo repository.MenuRepository, logger *zap.Logger) MenuService {
	return &menuService{menuRepo: menuRepo, logger: logger}
}

// GetTree 获取菜单树；all=true 返回全部；all=false 时仅返回 allowedMenuIDs 及其祖先节点；系统默认菜单（IsSystem）始终并入可见集合
func (s *menuService) GetTree(all bool, allowedMenuIDs []uuid.UUID) ([]*model.Menu, error) {
	flat, err := s.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	tree := repository.BuildTree(flat, nil)
	if all {
		return tree, nil
	}
	// 系统默认菜单始终加入侧栏（与用户已有权限取并集）
	merged := s.mergeSystemMenuIDs(flat, allowedMenuIDs)
	visibleIDs := s.visibleMenuIDs(flat, merged)
	return s.filterTreeByMenuIDs(tree, visibleIDs), nil
}

// mergeSystemMenuIDs 将 IsSystem 的菜单 ID 并入 allowed，保证系统菜单始终可见
func (s *menuService) mergeSystemMenuIDs(flat []model.Menu, allowed []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{})
	for _, id := range allowed {
		set[id] = struct{}{}
	}
	for _, m := range flat {
		if m.IsSystem {
			set[m.ID] = struct{}{}
		}
	}
	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

// visibleMenuIDs 计算可见菜单 ID 集合：allowedMenuIDs + 其所有祖先
func (s *menuService) visibleMenuIDs(flat []model.Menu, allowed []uuid.UUID) map[uuid.UUID]struct{} {
	idToParent := make(map[uuid.UUID]*uuid.UUID)
	for i := range flat {
		idToParent[flat[i].ID] = flat[i].ParentID
	}
	visible := make(map[uuid.UUID]struct{})
	for _, id := range allowed {
		for cur := &id; cur != nil; {
			visible[*cur] = struct{}{}
			pid := idToParent[*cur]
			if pid == nil {
				break
			}
			cur = pid
		}
	}
	return visible
}

func (s *menuService) filterTreeByMenuIDs(nodes []*model.Menu, visible map[uuid.UUID]struct{}) []*model.Menu {
	var out []*model.Menu
	for _, n := range nodes {
		if _, ok := visible[n.ID]; !ok {
			continue
		}
		clone := *n
		clone.Children = s.filterTreeByMenuIDs(n.Children, visible)
		out = append(out, &clone)
	}
	return out
}

// Create 创建菜单
func (s *menuService) Create(req *dto.MenuCreateRequest) (*model.Menu, error) {
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &pid
	}
	m := &model.Menu{
		ParentID:  parentID,
		Path:      req.Path,
		Name:      req.Name,
		Component: req.Component,
		Title:     req.Title,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		Meta:      req.Meta,
		Hidden:    req.Hidden,
	}
	if err := s.menuRepo.Create(m); err != nil {
		return nil, err
	}
	return s.menuRepo.GetByID(m.ID)
}

// Update 更新菜单
func (s *menuService) Update(id uuid.UUID, req *dto.MenuUpdateRequest) error {
	m, err := s.menuRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}
	shouldUpdateParent := true // 总是更新 parent_id，因为前端明确发送了这个字段
	if req.ParentID != nil {
		if *req.ParentID == "" {
			m.ParentID = nil
			s.logger.Info("Menu parent cleared", zap.String("menu_id", id.String()))
		} else {
			pid, err := uuid.Parse(*req.ParentID)
			if err != nil {
				s.logger.Error("Invalid parent_id", zap.String("parent_id", *req.ParentID), zap.Error(err))
				return err
			}
			if pid == id {
				return errors.New("不能将上级设为自己")
			}
			flat, _ := s.menuRepo.ListAll()
			if s.isDescendant(flat, id, pid) {
				return errors.New("不能将上级设为自身子级（会造成循环）")
			}
			m.ParentID = &pid
			s.logger.Info("Menu parent updated", zap.String("menu_id", id.String()), zap.String("parent_id", pid.String()))
		}
	} else {
		// 前端发送了 null，表示设置为顶级菜单
		m.ParentID = nil
		s.logger.Info("Menu parent set to top level", zap.String("menu_id", id.String()))
	}
	// 编辑时提交的字段一律按提交值更新，空字符串表示清空该字段（不再“为空则不修改”）
	m.Path = req.Path
	m.Name = req.Name
	m.Component = req.Component
	m.Title = req.Title
	m.Icon = req.Icon
	m.SortOrder = req.SortOrder
	if req.Meta != nil {
		m.Meta = req.Meta
	}
	m.Hidden = req.Hidden
	return s.menuRepo.Update(m, shouldUpdateParent)
}

// isDescendant 判断 targetID 是否是 ancestorID 的子孙节点（用于防止循环）
func (s *menuService) isDescendant(flat []model.Menu, ancestorID, targetID uuid.UUID) bool {
	idToParent := make(map[uuid.UUID]*uuid.UUID)
	for i := range flat {
		idToParent[flat[i].ID] = flat[i].ParentID
	}
	cur := &targetID
	for {
		parentID := idToParent[*cur]
		if parentID == nil {
			return false
		}
		if *parentID == ancestorID {
			return true
		}
		cur = parentID
	}
}

// Delete 删除菜单；系统默认菜单（IsSystem）不可删除
func (s *menuService) Delete(id uuid.UUID) error {
	m, err := s.menuRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}
	if m.IsSystem {
		return ErrMenuSystemProtected
	}
	return s.menuRepo.Delete(id)
}
