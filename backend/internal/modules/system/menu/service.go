package menu

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/database"
)

var (
	ErrMenuNotFound        = errors.New("menu not found")
	ErrMenuSystemProtected = errors.New("系统默认菜单不可删除")
)

type MenuService interface {
	GetTree(all bool, allowedMenuIDs []uuid.UUID) ([]*user.Menu, error)
	Create(req *dto.MenuCreateRequest) (*user.Menu, error)
	Update(id uuid.UUID, req *dto.MenuUpdateRequest) error
	Delete(id uuid.UUID) error
	// 菜单备份相关方法
	CreateBackup(name, description string, createdBy *uuid.UUID) error
	ListBackups() ([]*user.MenuBackup, error)
	DeleteBackup(id uuid.UUID) error
	RestoreBackup(id uuid.UUID) error
}

type menuService struct {
	menuRepo user.MenuRepository
	logger   *zap.Logger
}

func NewMenuService(menuRepo user.MenuRepository, logger *zap.Logger) MenuService {
	return &menuService{menuRepo: menuRepo, logger: logger}
}

func (s *menuService) GetTree(all bool, allowedMenuIDs []uuid.UUID) ([]*user.Menu, error) {
	flat, err := s.menuRepo.ListAll()
	if err != nil {
		return nil, err
	}
	tree := user.BuildTree(flat, nil)
	if all {
		return tree, nil
	}
	merged := s.mergeSystemMenuIDs(flat, allowedMenuIDs)
	visibleIDs := s.visibleMenuIDs(flat, merged)
	return s.filterTreeByMenuIDs(tree, visibleIDs), nil
}

func (s *menuService) mergeSystemMenuIDs(flat []user.Menu, allowed []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{})
	for _, id := range allowed {
		set[id] = struct{}{}
	}
	// 移除自动添加所有系统菜单的逻辑，只保留用户有权限的菜单
	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

func (s *menuService) visibleMenuIDs(flat []user.Menu, allowed []uuid.UUID) map[uuid.UUID]struct{} {
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

func (s *menuService) filterTreeByMenuIDs(nodes []*user.Menu, visible map[uuid.UUID]struct{}) []*user.Menu {
	var out []*user.Menu
	for _, n := range nodes {
		if _, ok := visible[n.ID]; !ok {
			continue
		}
		// 检查菜单是否启用
		if n.Meta != nil {
			if isEnable, ok := n.Meta["isEnable"].(bool); ok && !isEnable {
				continue
			}
		}
		clone := *n
		clone.Children = s.filterTreeByMenuIDs(n.Children, visible)
		out = append(out, &clone)
	}
	return out
}

func (s *menuService) Create(req *dto.MenuCreateRequest) (*user.Menu, error) {
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		pid, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, err
		}
		parentID = &pid
	}
	m := &user.Menu{
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

func (s *menuService) Update(id uuid.UUID, req *dto.MenuUpdateRequest) error {
	m, err := s.menuRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}
	shouldUpdateParent := true
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
		m.ParentID = nil
		s.logger.Info("Menu parent set to top level", zap.String("menu_id", id.String()))
	}
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

func (s *menuService) isDescendant(flat []user.Menu, ancestorID, targetID uuid.UUID) bool {
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

func (s *menuService) Delete(id uuid.UUID) error {
	_, err := s.menuRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuNotFound
		}
		return err
	}

	// 检查是否有子菜单
	children, err := s.menuRepo.GetChildren(id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return errors.New("该菜单存在子菜单，无法删除")
	}

	return s.menuRepo.Delete(id)
}

// 菜单备份相关方法
func (s *menuService) CreateBackup(name, description string, createdBy *uuid.UUID) error {
	// 获取所有菜单
	menus, err := s.menuRepo.ListAll()
	if err != nil {
		s.logger.Error("Failed to list menus for backup", zap.Error(err))
		return err
	}

	// 将菜单数据转换为JSON
	menuData, err := json.Marshal(menus)
	if err != nil {
		s.logger.Error("Failed to marshal menu data", zap.Error(err))
		return err
	}

	// 创建备份
	backup := &user.MenuBackup{
		Name:        name,
		Description: description,
		MenuData:    string(menuData),
		CreatedBy:   createdBy,
	}

	if err := s.menuRepo.CreateBackup(backup); err != nil {
		s.logger.Error("Failed to create menu backup", zap.Error(err))
		return err
	}

	s.logger.Info("Menu backup created", zap.String("name", name))
	return nil
}

func (s *menuService) ListBackups() ([]*user.MenuBackup, error) {
	backups, err := s.menuRepo.ListBackups()
	if err != nil {
		s.logger.Error("Failed to list menu backups", zap.Error(err))
		return nil, err
	}

	// 转换为指针切片
	backupPtrs := make([]*user.MenuBackup, len(backups))
	for i, backup := range backups {
		backupCopy := backup
		backupPtrs[i] = &backupCopy
	}

	return backupPtrs, nil
}

func (s *menuService) DeleteBackup(id uuid.UUID) error {
	// 检查备份是否存在
	_, err := s.menuRepo.GetBackupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("备份不存在")
		}
		return err
	}

	if err := s.menuRepo.DeleteBackup(id); err != nil {
		s.logger.Error("Failed to delete menu backup", zap.Error(err))
		return err
	}

	s.logger.Info("Menu backup deleted", zap.String("backup_id", id.String()))
	return nil
}

func (s *menuService) RestoreBackup(id uuid.UUID) error {
	// 检查备份是否存在
	backup, err := s.menuRepo.GetBackupByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("备份不存在")
		}
		return err
	}

	// 解析备份的菜单数据
	var menus []user.Menu
	if err := json.Unmarshal([]byte(backup.MenuData), &menus); err != nil {
		s.logger.Error("Failed to unmarshal menu data", zap.Error(err))
		return err
	}

	// 清空现有菜单
	if err := s.menuRepo.DeleteAllMenus(); err != nil {
		s.logger.Error("Failed to delete all menus", zap.Error(err))
		return err
	}

	// 重新创建菜单
	for i := range menus {
		if err := s.menuRepo.Create(&menus[i]); err != nil {
			s.logger.Error("Failed to create menu during restore", zap.Error(err))
			return err
		}
	}

	// 清理无效的角色菜单关联
	if err := s.cleanupInvalidRoleMenus(); err != nil {
		s.logger.Warn("Failed to cleanup invalid role-menus after restore", zap.Error(err))
	}

	s.logger.Info("Menu backup restored", zap.String("backup_id", id.String()))
	return nil
}

// 清理无效的角色菜单关联（删除关联表中不存在于菜单表内的关联）
func (s *menuService) cleanupInvalidRoleMenus() error {
	// 删除角色菜单关联表中不存在于菜单表内的关联
	if err := database.DB.Exec("DELETE FROM role_menus WHERE menu_id NOT IN (SELECT id FROM menus)").Error; err != nil {
		s.logger.Error("Failed to cleanup invalid role-menus", zap.Error(err))
		return err
	}
	s.logger.Info("Invalid role-menus cleaned up")
	return nil
}
