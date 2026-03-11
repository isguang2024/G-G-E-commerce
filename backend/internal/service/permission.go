package service

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
	"github.com/gg-ecommerce/backend/internal/repository"
)

// PermissionService 权限计算服务
type PermissionService interface {
	// GetUserMenuIDs 计算用户可见的菜单ID集合
	// 规则：
	// 1. 用户状态必须为 active
	// 2. 只计算状态为 normal 的角色
	// 3. 同一作用域内：按优先级合并（高优先级覆盖低优先级）
	// 4. 不同作用域间：权限相加
	GetUserMenuIDs(userID uuid.UUID) ([]uuid.UUID, error)
}

type permissionService struct {
	userRepo       repository.UserRepository
	roleMenuRepo   repository.RoleMenuRepository
	db             *gorm.DB
}

// NewPermissionService 创建权限计算服务
func NewPermissionService(
	userRepo repository.UserRepository,
	roleMenuRepo repository.RoleMenuRepository,
	db *gorm.DB,
) PermissionService {
	return &permissionService{
		userRepo:     userRepo,
		roleMenuRepo: roleMenuRepo,
		db:           db,
	}
}

// GetUserMenuIDs 计算用户可见的菜单ID集合
func (s *permissionService) GetUserMenuIDs(userID uuid.UUID) ([]uuid.UUID, error) {
	// 1. 获取用户及其角色（含 Scope）
	var user model.User
	err := s.db.Preload("Roles.Scope").First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}

	// 2. 检查用户状态
	if user.Status != "active" {
		return []uuid.UUID{}, nil
	}

	// 3. 过滤已启用的角色，并收集角色信息
	var enabledRoles []model.Role
	rolePriorityMap := make(map[uuid.UUID]int)      // 角色ID -> 优先级
	roleScopeMap := make(map[uuid.UUID]string)       // 角色ID -> 作用域编码
	for _, r := range user.Roles {
		if r.Status == "normal" {
			enabledRoles = append(enabledRoles, r)
			rolePriorityMap[r.ID] = r.Priority
			scopeCode := r.Scope.Code
			if scopeCode == "" {
				scopeCode = "global"
			}
			roleScopeMap[r.ID] = scopeCode
		}
	}
	if len(enabledRoles) == 0 {
		return []uuid.UUID{}, nil
	}

	// 4. 收集所有角色ID，一次性查询所有菜单关联
	roleIDs := make([]uuid.UUID, len(enabledRoles))
	for i, r := range enabledRoles {
		roleIDs[i] = r.ID
	}

	// 一次性获取所有角色的菜单关联（优化：减少N次查询为1次）
	roleMenuList, err := s.roleMenuRepo.GetMenusByRoleIDs(roleIDs)
	if err != nil {
		return nil, err
	}

	// 5. 按作用域分组，同时带上优先级
	// map[scopeCode]map[menuID]优先级
	scopeMenuPriority := make(map[string]map[uuid.UUID]int)
	for _, rm := range roleMenuList {
		scopeCode := roleScopeMap[rm.RoleID]
		if scopeMenuPriority[scopeCode] == nil {
			scopeMenuPriority[scopeCode] = make(map[uuid.UUID]int)
		}
		// 同一作用域内，高优先级覆盖低优先级
		currentPriority := rolePriorityMap[rm.RoleID]
		if existingPriority, exists := scopeMenuPriority[scopeCode][rm.MenuID]; !exists || currentPriority > existingPriority {
			scopeMenuPriority[scopeCode][rm.MenuID] = currentPriority
		}
	}

	// 6. 合并所有作用域的权限
	allMenuIDs := make(map[uuid.UUID]struct{})
	for _, menuPriority := range scopeMenuPriority {
		for menuID := range menuPriority {
			allMenuIDs[menuID] = struct{}{}
		}
	}

	// 7. 转换为切片
	result := make([]uuid.UUID, 0, len(allMenuIDs))
	for id := range allMenuIDs {
		result = append(result, id)
	}

	return result, nil
}
