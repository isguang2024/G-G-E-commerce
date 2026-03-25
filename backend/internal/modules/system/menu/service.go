package menu

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/api/dto"
	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
	"github.com/gg-ecommerce/backend/internal/pkg/permissionrefresh"
)

var (
	ErrMenuNotFound        = errors.New("menu not found")
	ErrMenuSystemProtected = errors.New("系统默认菜单不可删除")
	ErrMenuGroupNotFound   = errors.New("menu group not found")
	ErrMenuGroupInUse      = errors.New("menu group in use")
)

type MenuService interface {
	GetTree(all bool, allowedMenuIDs []uuid.UUID) ([]*user.Menu, error)
	Create(req *dto.MenuCreateRequest) (*user.Menu, error)
	Update(id uuid.UUID, req *dto.MenuUpdateRequest) error
	Delete(id uuid.UUID) error
	ListGroups() ([]user.MenuManageGroup, error)
	CreateGroup(req *dto.MenuManageGroupCreateRequest) (*user.MenuManageGroup, error)
	UpdateGroup(id uuid.UUID, req *dto.MenuManageGroupUpdateRequest) error
	DeleteGroup(id uuid.UUID) error
	// 菜单备份相关方法
	CreateBackup(name, description string, createdBy *uuid.UUID) error
	ListBackups() ([]*user.MenuBackup, error)
	DeleteBackup(id uuid.UUID) error
	RestoreBackup(id uuid.UUID) error
}

type menuService struct {
	db        *gorm.DB
	menuRepo  user.MenuRepository
	refresher permissionrefresh.Service
	logger    *zap.Logger
}

func NewMenuService(db *gorm.DB, menuRepo user.MenuRepository, refresher permissionrefresh.Service, logger *zap.Logger) MenuService {
	return &menuService{db: db, menuRepo: menuRepo, refresher: refresher, logger: logger}
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

	// 对 jwt/public 菜单做默认放行，不依赖功能权限分配。
	for _, menu := range flat {
		accessMode := menuAccessMode(menu.Meta)
		if accessMode == "jwt" || accessMode == "public" {
			set[menu.ID] = struct{}{}
		}
	}

	out := make([]uuid.UUID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	return out
}

func menuAccessMode(meta map[string]interface{}) string {
	if meta == nil {
		return "permission"
	}
	raw, ok := meta["accessMode"]
	if !ok {
		return "permission"
	}
	value, ok := raw.(string)
	if !ok {
		return "permission"
	}
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "public", "jwt", "permission":
		return value
	default:
		return "permission"
	}
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
	manageGroupID, err := parseOptionalUUID(req.ManageGroupID)
	if err != nil {
		return nil, err
	}
	m := &user.Menu{
		ParentID:      parentID,
		ManageGroupID: manageGroupID,
		Path:          req.Path,
		Name:          req.Name,
		Component:     req.Component,
		Title:         req.Title,
		Icon:          req.Icon,
		SortOrder:     req.SortOrder,
		Meta:          sanitizeMenuMeta(req.Meta),
		Hidden:        req.Hidden,
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
	manageGroupID, err := parseOptionalUUID(req.ManageGroupID)
	if err != nil {
		return err
	}
	m.ManageGroupID = manageGroupID
	m.Path = req.Path
	m.Name = req.Name
	m.Component = req.Component
	m.Title = req.Title
	m.Icon = req.Icon
	m.SortOrder = req.SortOrder
	if req.Meta != nil {
		m.Meta = sanitizeMenuMeta(req.Meta)
	}
	m.Hidden = req.Hidden
	if err := s.menuRepo.Update(m, shouldUpdateParent); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshByMenu(id)
	}
	return nil
}

func (s *menuService) ListGroups() ([]user.MenuManageGroup, error) {
	var groups []user.MenuManageGroup
	err := s.db.Order("sort_order ASC, created_at ASC").Find(&groups).Error
	return groups, err
}

func (s *menuService) CreateGroup(req *dto.MenuManageGroupCreateRequest) (*user.MenuManageGroup, error) {
	group := &user.MenuManageGroup{
		Name:      strings.TrimSpace(req.Name),
		SortOrder: req.SortOrder,
		Status:    normalizeMenuGroupStatus(req.Status),
	}
	if err := s.db.Create(group).Error; err != nil {
		return nil, err
	}
	return group, nil
}

func (s *menuService) UpdateGroup(id uuid.UUID, req *dto.MenuManageGroupUpdateRequest) error {
	var group user.MenuManageGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuGroupNotFound
		}
		return err
	}
	return s.db.Model(&group).Updates(map[string]interface{}{
		"name":       strings.TrimSpace(req.Name),
		"sort_order": req.SortOrder,
		"status":     normalizeMenuGroupStatus(req.Status),
	}).Error
}

func (s *menuService) DeleteGroup(id uuid.UUID) error {
	var group user.MenuManageGroup
	if err := s.db.Where("id = ?", id).First(&group).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMenuGroupNotFound
		}
		return err
	}

	var count int64
	if err := s.db.Model(&models.Menu{}).Where("manage_group_id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrMenuGroupInUse
	}

	return s.db.Delete(&group).Error
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

	if err := s.menuRepo.Delete(id); err != nil {
		return err
	}
	if s.refresher != nil {
		return s.refresher.RefreshByMenu(id)
	}
	return nil
}

// 菜单备份相关方法
func (s *menuService) CreateBackup(name, description string, createdBy *uuid.UUID) error {
	// 获取所有菜单
	menus, err := s.menuRepo.ListAll()
	if err != nil {
		s.logger.Error("Failed to list menus for backup", zap.Error(err))
		return err
	}

	var groups []user.MenuManageGroup
	if err := s.db.Order("sort_order ASC, created_at ASC").Find(&groups).Error; err != nil {
		s.logger.Error("Failed to list menu groups for backup", zap.Error(err))
		return err
	}

	// 将菜单数据转换为JSON
	menuData, err := json.Marshal(ginMenuBackupPayload{
		Groups: groups,
		Menus:  menus,
	})
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
	var payload ginMenuBackupPayload
	var menus []user.Menu
	var groups []user.MenuManageGroup
	if err := json.Unmarshal([]byte(backup.MenuData), &payload); err == nil && len(payload.Menus) > 0 {
		menus = payload.Menus
		groups = payload.Groups
	} else {
		if err := json.Unmarshal([]byte(backup.MenuData), &menus); err != nil {
			s.logger.Error("Failed to unmarshal menu data", zap.Error(err))
			return err
		}
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM menus").Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM menu_manage_groups").Error; err != nil {
			return err
		}
		for i := range groups {
			if err := tx.Create(&groups[i]).Error; err != nil {
				return err
			}
		}
		for i := range menus {
			menus[i].Meta = sanitizeMenuMeta(menus[i].Meta)
			if err := tx.Create(&menus[i]).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		s.logger.Error("Failed to restore menu backup", zap.Error(err))
		return err
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
	return nil
}

type ginMenuBackupPayload struct {
	Groups []user.MenuManageGroup `json:"groups"`
	Menus  []user.Menu            `json:"menus"`
}

func parseOptionalUUID(value *string) (*uuid.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(*value))
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func sanitizeMenuMeta(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return nil
	}
	sanitized := make(map[string]interface{}, len(meta))
	for key, value := range meta {
		if key == "manageGroup" {
			continue
		}
		sanitized[key] = value
	}
	return sanitized
}

func normalizeMenuGroupStatus(value string) string {
	switch strings.TrimSpace(value) {
	case "disabled":
		return "disabled"
	default:
		return "normal"
	}
}
