package user

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	List(offset, limit int, username, userPhone, userEmail, status, roleID, id, registerSource, invitedBy string) ([]User, int64, error)
	GetByID(id uuid.UUID) (*User, error)
	GetByIDs(ids []uuid.UUID) ([]User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	Create(user *User) error
	Update(user *User) error
	Delete(id uuid.UUID) error
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
	UpdateLastLogin(id uuid.UUID, ip string) error
	ReplaceRoles(userID uuid.UUID, roleIDs []uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id uuid.UUID) (*User, error) {
	var user User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.loadGlobalRoles([]*User{&user}); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIDs(ids []uuid.UUID) ([]User, error) {
	var users []User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}
	if err := r.loadGlobalRoles(userSlicePointers(users)); err != nil {
		return nil, err
	}
	return users, err
}

func (r *userRepository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.loadGlobalRoles([]*User{&user}); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	if err := r.loadGlobalRoles([]*User{&user}); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *User) error {
	return r.db.Model(user).Updates(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&UserRole{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&UserActionPermission{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&TenantMember{}).Error; err != nil {
			return err
		}
		return tx.Delete(&User{}, id).Error
	})
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) UpdateLastLogin(id uuid.UUID, ip string) error {
	now := r.db.NowFunc()
	return r.db.Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

func (r *userRepository) List(offset, limit int, username, userPhone, userEmail, status, roleID, id, registerSource, invitedBy string) ([]User, int64, error) {
	baseQuery := r.db.Model(&User{})
	if id != "" {
		baseQuery = baseQuery.Where("id = ?", id)
	}
	if username != "" {
		baseQuery = baseQuery.Where("username LIKE ?", "%"+username+"%")
	}
	if userPhone != "" {
		baseQuery = baseQuery.Where("phone LIKE ?", "%"+userPhone+"%")
	}
	if userEmail != "" {
		baseQuery = baseQuery.Where("email LIKE ?", "%"+userEmail+"%")
	}
	if status != "" {
		baseQuery = baseQuery.Where("status = ?", status)
	}
	if registerSource != "" {
		baseQuery = baseQuery.Where("register_source = ?", registerSource)
	}
	if invitedBy != "" {
		baseQuery = baseQuery.Where("invited_by = ?", invitedBy)
	}
	if roleID != "" {
		baseQuery = baseQuery.Joins("JOIN user_roles ON users.id = user_roles.user_id").Where("user_roles.role_id = ?", roleID)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []User
	err := baseQuery.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	if err := r.loadGlobalRoles(userSlicePointers(users)); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepository) ReplaceRoles(userID uuid.UUID, roleIDs []uuid.UUID) error {
	tx := r.db.Begin()

	if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	userRoles := make([]UserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, UserRole{
			UserID: userID,
			RoleID: roleID,
		})
	}

	if err := tx.Create(&userRoles).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

type RoleRepository interface {
	GetByID(id uuid.UUID) (*Role, error)
	GetByCode(code string) (*Role, error)
	FindByCode(code string) ([]Role, error)
	GetByIDs(ids []uuid.UUID) ([]Role, error)
	GetByScopeID(scopeID uuid.UUID) ([]Role, error)
	Create(role *Role) error
	Update(role *Role) error
	Delete(id uuid.UUID) error
	GetAll() ([]Role, error)
	GetByScope(scope string) ([]Role, error)
	List() ([]Role, error)
	ListByPage(offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool) ([]Role, int64, error)
	ListByScope(scope string, offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool) ([]Role, int64, error)
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByID(id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.Preload("Scope").Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByCode(code string) (*Role, error) {
	var role Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByCode(code string) ([]Role, error) {
	var roles []Role
	err := r.db.Where("code = ?", code).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByIDs(ids []uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) List() ([]Role, error) {
	var roles []Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByScopeID(scopeID uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.Where("scope_id = ?", scopeID).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Create(role *Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) Update(role *Role) error {
	return r.db.Model(role).Updates(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Role{}, id).Error
}

func (r *roleRepository) GetAll() ([]Role, error) {
	var roles []Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByScope(scope string) ([]Role, error) {
	var roles []Role
	err := r.db.Joins("JOIN scopes ON roles.scope_id = scopes.id").Where("scopes.code = ?", scope).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) ListByPage(offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool) ([]Role, int64, error) {
	return r.listWithScope(offset, limit, roleCode, roleName, description, startTime, endTime, enabled, "")
}

func (r *roleRepository) ListByScope(scope string, offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool) ([]Role, int64, error) {
	return r.listWithScope(offset, limit, roleCode, roleName, description, startTime, endTime, enabled, scope)
}

func (r *roleRepository) listWithScope(offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool, scope string) ([]Role, int64, error) {
	baseQuery := r.db.Model(&Role{}).Preload("Scope")
	if roleCode != "" {
		baseQuery = baseQuery.Where("code LIKE ?", "%"+roleCode+"%")
	}
	if roleName != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+roleName+"%")
	}
	if description != "" {
		baseQuery = baseQuery.Where("description LIKE ?", "%"+description+"%")
	}
	if startTime != "" {
		baseQuery = baseQuery.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		baseQuery = baseQuery.Where("created_at <= ?", endTime)
	}
	if enabled != nil {
		if *enabled {
			baseQuery = baseQuery.Where("status = ?", "normal")
		} else {
			baseQuery = baseQuery.Where("status = ?", "disabled")
		}
	}
	if scope != "" {
		baseQuery = baseQuery.Joins("JOIN scopes ON roles.scope_id = scopes.id").Where("scopes.code = ?", scope)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var roles []Role
	err := baseQuery.Offset(offset).Limit(limit).Order("sort_order ASC, created_at DESC").Find(&roles).Error
	return roles, total, err
}

func (r *roleRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&Role{}).Where("id = ?", id).Updates(updates).Error
}

type ScopeRepository interface {
	GetByID(id uuid.UUID) (*Scope, error)
	GetByCode(code string) (*Scope, error)
	Create(scope *Scope) error
	Update(scope *Scope) error
	Delete(id uuid.UUID) error
	GetAll() ([]Scope, error)
	List(offset, limit int, code, name string) ([]Scope, int64, error)
}

type scopeRepository struct {
	db *gorm.DB
}

func NewScopeRepository(db *gorm.DB) ScopeRepository {
	return &scopeRepository{db: db}
}

func (r *scopeRepository) GetByID(id uuid.UUID) (*Scope, error) {
	var scope Scope
	err := r.db.Where("id = ?", id).First(&scope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &scope, nil
}

func (r *scopeRepository) GetByCode(code string) (*Scope, error) {
	var scope Scope
	err := r.db.Where("code = ?", code).First(&scope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &scope, nil
}

func (r *scopeRepository) Create(scope *Scope) error {
	return r.db.Create(scope).Error
}

func (r *scopeRepository) Update(scope *Scope) error {
	return r.db.Model(scope).Updates(scope).Error
}

func (r *scopeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Scope{}, id).Error
}

func (r *scopeRepository) GetAll() ([]Scope, error) {
	var scopes []Scope
	err := r.db.Order("sort_order ASC").Find(&scopes).Error
	return scopes, err
}

func (r *scopeRepository) List(offset, limit int, code, name string) ([]Scope, int64, error) {
	baseQuery := r.db.Model(&Scope{})
	if code != "" {
		baseQuery = baseQuery.Where("code LIKE ?", "%"+code+"%")
	}
	if name != "" {
		baseQuery = baseQuery.Where("name LIKE ?", "%"+name+"%")
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var scopes []Scope
	err := baseQuery.Offset(offset).Limit(limit).Order("sort_order ASC").Find(&scopes).Error
	return scopes, total, err
}

type MenuRepository interface {
	GetByID(id uuid.UUID) (*Menu, error)
	GetChildren(parentID uuid.UUID) ([]Menu, error)
	Create(menu *Menu) error
	Update(menu *Menu, updateParent bool) error
	Delete(id uuid.UUID) error
	ListAll() ([]Menu, error)
	GetByIDs(ids []uuid.UUID) ([]Menu, error)
	// 菜单备份相关方法
	CreateBackup(backup *MenuBackup) error
	GetBackupByID(id uuid.UUID) (*MenuBackup, error)
	ListBackups() ([]MenuBackup, error)
	DeleteBackup(id uuid.UUID) error
	DeleteAllMenus() error
}

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) GetByID(id uuid.UUID) (*Menu, error) {
	var menu Menu
	err := r.db.Where("id = ?", id).First(&menu).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &menu, nil
}

func (r *menuRepository) GetChildren(parentID uuid.UUID) ([]Menu, error) {
	var menus []Menu
	err := r.db.Where("parent_id = ?", parentID).Find(&menus).Error
	return menus, err
}

func (r *menuRepository) Create(menu *Menu) error {
	return r.db.Create(menu).Error
}

func (r *menuRepository) Update(menu *Menu, updateParent bool) error {
	if updateParent {
		return r.db.Model(menu).Updates(map[string]interface{}{
			"parent_id":  menu.ParentID,
			"path":       menu.Path,
			"name":       menu.Name,
			"component":  menu.Component,
			"title":      menu.Title,
			"icon":       menu.Icon,
			"sort_order": menu.SortOrder,
			"meta":       menu.Meta,
			"hidden":     menu.Hidden,
		}).Error
	}
	return r.db.Model(menu).Updates(map[string]interface{}{
		"path":       menu.Path,
		"name":       menu.Name,
		"component":  menu.Component,
		"title":      menu.Title,
		"icon":       menu.Icon,
		"sort_order": menu.SortOrder,
		"meta":       menu.Meta,
		"hidden":     menu.Hidden,
	}).Error
}

func (r *menuRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Menu{}, id).Error
}

func (r *menuRepository) ListAll() ([]Menu, error) {
	var menus []Menu
	err := r.db.Order("sort_order ASC").Find(&menus).Error
	return menus, err
}

func (r *menuRepository) GetByIDs(ids []uuid.UUID) ([]Menu, error) {
	var menus []Menu
	err := r.db.Where("id IN ?", ids).Find(&menus).Error
	return menus, err
}

// 菜单备份相关方法
func (r *menuRepository) CreateBackup(backup *MenuBackup) error {
	return r.db.Create(backup).Error
}

func (r *menuRepository) GetBackupByID(id uuid.UUID) (*MenuBackup, error) {
	var backup MenuBackup
	err := r.db.Where("id = ?", id).First(&backup).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &backup, nil
}

func (r *menuRepository) ListBackups() ([]MenuBackup, error) {
	var backups []MenuBackup
	err := r.db.Order("created_at DESC").Find(&backups).Error
	return backups, err
}

func (r *menuRepository) DeleteBackup(id uuid.UUID) error {
	return r.db.Delete(&MenuBackup{}, id).Error
}

func (r *menuRepository) DeleteAllMenus() error {
	// 只删除所有菜单，不删除角色菜单关联
	// 角色菜单关联会在 cleanupInvalidRoleMenus 中清理无效关联
	return r.db.Exec("DELETE FROM menus").Error
}

func BuildTree(menus []Menu, parentID *uuid.UUID) []*Menu {
	var tree []*Menu
	for i := range menus {
		if (parentID == nil && menus[i].ParentID == nil) ||
			(parentID != nil && menus[i].ParentID != nil && *menus[i].ParentID == *parentID) {
			children := BuildTree(menus, &menus[i].ID)
			menus[i].Children = children
			tree = append(tree, &menus[i])
		}
	}
	return tree
}

type TenantRepository interface {
	GetByID(id uuid.UUID) (*Tenant, error)
	Create(tenant *Tenant) error
	Update(tenant *Tenant) error
	Delete(id uuid.UUID) error
	List(offset, limit int, name, status string) ([]Tenant, int64, error)
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) GetByID(id uuid.UUID) (*Tenant, error) {
	var tenant Tenant
	err := r.db.Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Create(tenant *Tenant) error {
	return r.db.Create(tenant).Error
}

func (r *tenantRepository) Update(tenant *Tenant) error {
	return r.db.Model(tenant).Updates(tenant).Error
}

func (r *tenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Tenant{}, id).Error
}

func (r *tenantRepository) List(offset, limit int, name, status string) ([]Tenant, int64, error) {
	var total int64
	query := r.db.Model(&Tenant{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tenants []Tenant
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&tenants).Error
	return tenants, total, err
}

func (r *tenantRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&Tenant{}).Where("id = ?", id).Updates(updates).Error
}

type TenantMemberRepository interface {
	GetByUserAndTenant(userID, tenantID uuid.UUID) (*TenantMember, error)
	GetByUserID(userID uuid.UUID) (*TenantMember, error)
	GetTenantsByUserID(userID uuid.UUID) ([]Tenant, error)
	GetAdminUsersByTenantID(tenantID uuid.UUID) ([]User, error)
	List(tenantID uuid.UUID, params *MemberSearchParams) ([]TenantMember, error)
	Create(member *TenantMember) error
	Delete(id uuid.UUID) error
	DeleteByUserAndTenant(userID, tenantID uuid.UUID) error
	UpdateRole(id uuid.UUID, roleCode string) error
}

type tenantMemberRepository struct {
	db *gorm.DB
}

func NewTenantMemberRepository(db *gorm.DB) TenantMemberRepository {
	return &tenantMemberRepository{db: db}
}

func (r *tenantMemberRepository) GetByUserAndTenant(userID, tenantID uuid.UUID) (*TenantMember, error) {
	var member TenantMember
	err := r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *tenantMemberRepository) GetTenantsByUserID(userID uuid.UUID) ([]Tenant, error) {
	var tenants []Tenant
	err := r.db.Joins("JOIN tenant_members ON tenants.id = tenant_members.tenant_id").
		Where("tenant_members.user_id = ?", userID).
		Find(&tenants).Error
	return tenants, err
}

func (r *tenantMemberRepository) GetAdminUsersByTenantID(tenantID uuid.UUID) ([]User, error) {
	var users []User
	err := r.db.Joins("JOIN tenant_members ON users.id = tenant_members.user_id").
		Where("tenant_members.tenant_id = ? AND tenant_members.role_code = ?", tenantID, "team_admin").
		Find(&users).Error
	return users, err
}

func (r *tenantMemberRepository) GetByUserID(userID uuid.UUID) (*TenantMember, error) {
	var member TenantMember
	err := r.db.Where("user_id = ?", userID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &member, nil
}

func (r *tenantMemberRepository) List(tenantID uuid.UUID, params *MemberSearchParams) ([]TenantMember, error) {
	query := r.db.Where("tenant_id = ?", tenantID)
	if params != nil {
		if params.UserID != "" {
			query = query.Where("user_id = ?", params.UserID)
		}
		if params.UserName != "" {
			var userIDs []uuid.UUID
			r.db.Model(&User{}).
				Where("username LIKE ? OR nickname LIKE ?", "%"+params.UserName+"%", "%"+params.UserName+"%").
				Pluck("id", &userIDs)
			if len(userIDs) > 0 {
				query = query.Where("user_id IN ?", userIDs)
			} else {
				query = query.Where("1 = 0")
			}
		}
		if params.RoleCode != "" {
			query = query.Where("role_code = ?", params.RoleCode)
		}
	}
	var members []TenantMember
	err := query.Find(&members).Error
	return members, err
}

func (r *tenantMemberRepository) Create(member *TenantMember) error {
	return r.db.Create(member).Error
}

func (r *tenantMemberRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&TenantMember{}, "id = ?", id).Error
}

func (r *tenantMemberRepository) DeleteByUserAndTenant(userID, tenantID uuid.UUID) error {
	return r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).Delete(&TenantMember{}).Error
}

func (r *tenantMemberRepository) UpdateRole(id uuid.UUID, roleCode string) error {
	return r.db.Model(&TenantMember{}).Where("id = ?", id).Update("role_code", roleCode).Error
}

type UserRoleRepository interface {
	GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID, tenantMemberRepo TenantMemberRepository) ([]uuid.UUID, error)
	GetRoleCodesByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error)
	GetEffectiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
	GetEffectiveActiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error)
	GetEffectiveRoleCodesByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error)
	ReplaceUserRoles(userID uuid.UUID, tenantID *uuid.UUID, roleIDs []uuid.UUID) error
	AssignRole(userID, roleID uuid.UUID, tenantID *uuid.UUID) error
	SetUserRoles(userID uuid.UUID, roleIDs []uuid.UUID, tenantID *uuid.UUID) error
	RemoveUserRole(userID uuid.UUID, tenantID *uuid.UUID) error
	RemoveRolesByCodes(userID uuid.UUID, tenantID *uuid.UUID, roleCodes []string) error
}

type userRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID, tenantMemberRepo TenantMemberRepository) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	query := r.db.Model(&UserRole{}).Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	err := query.Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRoleRepository) GetRoleCodesByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	var codes []string
	query := r.db.Model(&UserRole{}).
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("user_roles.tenant_id IS NULL")
	} else {
		query = query.Where("user_roles.tenant_id = ?", *tenantID)
	}
	err := query.Pluck("roles.code", &codes).Error
	return codes, err
}

func (r *userRoleRepository) GetEffectiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	query := r.db.Model(&UserRole{}).Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}
	err := query.Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRoleRepository) GetEffectiveActiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	query := r.db.Model(&UserRole{}).
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("roles.status = ?", "normal")
	if tenantID == nil {
		query = query.Where("user_roles.tenant_id IS NULL")
	} else {
		query = query.Where("user_roles.tenant_id IS NULL OR user_roles.tenant_id = ?", *tenantID)
	}
	err := query.Distinct("user_roles.role_id").Pluck("user_roles.role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRoleRepository) GetEffectiveRoleCodesByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	var codes []string
	query := r.db.Model(&UserRole{}).
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("user_roles.tenant_id IS NULL")
	} else {
		query = query.Where("user_roles.tenant_id IS NULL OR user_roles.tenant_id = ?", *tenantID)
	}
	err := query.Pluck("roles.code", &codes).Error
	return codes, err
}

func (r *userRoleRepository) ReplaceUserRoles(userID uuid.UUID, tenantID *uuid.UUID, roleIDs []uuid.UUID) error {
	tx := r.db.Begin()
	deleteQuery := tx.Where("user_id = ?", userID)
	if tenantID == nil {
		deleteQuery = deleteQuery.Where("tenant_id IS NULL")
	} else {
		deleteQuery = deleteQuery.Where("tenant_id = ?", *tenantID)
	}
	if err := deleteQuery.Delete(&UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	userRoles := make([]UserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, UserRole{UserID: userID, RoleID: roleID, TenantID: tenantID})
	}
	if len(userRoles) == 0 {
		return tx.Commit().Error
	}
	if err := tx.Create(&userRoles).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *userRoleRepository) AssignRole(userID, roleID uuid.UUID, tenantID *uuid.UUID) error {
	query := r.db.Model(&UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	userRole := UserRole{UserID: userID, RoleID: roleID, TenantID: tenantID}
	return r.db.Create(&userRole).Error
}

func (r *userRoleRepository) SetUserRoles(userID uuid.UUID, roleIDs []uuid.UUID, tenantID *uuid.UUID) error {
	return r.ReplaceUserRoles(userID, tenantID, roleIDs)
}

func (r *userRoleRepository) RemoveUserRole(userID uuid.UUID, tenantID *uuid.UUID) error {
	query := r.db.Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	return query.Delete(&UserRole{}).Error
}

func (r *userRoleRepository) RemoveRolesByCodes(userID uuid.UUID, tenantID *uuid.UUID, roleCodes []string) error {
	if len(roleCodes) == 0 {
		return nil
	}
	query := r.db.Where("user_id = ?", userID).
		Where("role_id IN (?)",
			r.db.Model(&Role{}).Select("id").Where("code IN ?", roleCodes),
		)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	return query.Delete(&UserRole{}).Error
}

func (r *userRepository) loadGlobalRoles(users []*User) error {
	if len(users) == 0 {
		return nil
	}

	userIDs := make([]uuid.UUID, 0, len(users))
	userIndex := make(map[uuid.UUID]*User, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}
		user.Roles = nil
		userIDs = append(userIDs, user.ID)
		userIndex[user.ID] = user
	}

	type userRoleRow struct {
		UserID      uuid.UUID `gorm:"column:user_id"`
		ID          uuid.UUID `gorm:"column:id"`
		Code        string    `gorm:"column:code"`
		Name        string    `gorm:"column:name"`
		Description string    `gorm:"column:description"`
		Status      string    `gorm:"column:status"`
		Priority    int       `gorm:"column:priority"`
		ScopeID     uuid.UUID `gorm:"column:scope_id"`
		SortOrder   int       `gorm:"column:sort_order"`
		IsSystem    bool      `gorm:"column:is_system"`
		CreatedAt   time.Time `gorm:"column:created_at"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
	}

	var rows []userRoleRow
	if err := r.db.Table("user_roles").
		Select("user_roles.user_id, roles.id, roles.code, roles.name, roles.description, roles.status, roles.priority, roles.scope_id, roles.sort_order, roles.is_system, roles.created_at, roles.updated_at").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id IN ?", userIDs).
		Where("user_roles.tenant_id IS NULL").
		Scan(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		target, ok := userIndex[row.UserID]
		if !ok {
			continue
		}
		target.Roles = append(target.Roles, Role{
			ID:          row.ID,
			Code:        row.Code,
			Name:        row.Name,
			Description: row.Description,
			Status:      row.Status,
			Priority:    row.Priority,
			ScopeID:     row.ScopeID,
			SortOrder:   row.SortOrder,
			IsSystem:    row.IsSystem,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return nil
}

func userSlicePointers(users []User) []*User {
	result := make([]*User, 0, len(users))
	for i := range users {
		result = append(result, &users[i])
	}
	return result
}

type RoleMenuRepository interface {
	GetMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error)
	GetMenuIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error)
	SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error
}

type roleMenuRepository struct {
	db *gorm.DB
}

func NewRoleMenuRepository(db *gorm.DB) RoleMenuRepository {
	return &roleMenuRepository{db: db}
}

func (r *roleMenuRepository) GetMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := r.db.Model(&RoleMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *roleMenuRepository) GetMenuIDsByRoleIDs(roleIDs []uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	if len(roleIDs) == 0 {
		return menuIDs, nil
	}
	err := r.db.Model(&RoleMenu{}).Distinct("menu_id").Where("role_id IN ?", roleIDs).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *roleMenuRepository) SetRoleMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error {
	tx := r.db.Begin()
	if err := tx.Where("role_id = ?", roleID).Delete(&RoleMenu{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	roleMenus := make([]RoleMenu, 0, len(menuIDs))
	for _, menuID := range menuIDs {
		roleMenus = append(roleMenus, RoleMenu{RoleID: roleID, MenuID: menuID})
	}
	if err := tx.Create(&roleMenus).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

type PermissionActionRepository interface {
	List(offset, limit int, params *PermissionActionListParams) ([]PermissionAction, int64, error)
	GetByID(id uuid.UUID) (*PermissionAction, error)
	GetByIDs(ids []uuid.UUID) ([]PermissionAction, error)
	GetByResourceAndAction(resourceCode, actionCode, scopeCode string) (*PermissionAction, error)
	GetAllEnabled() ([]PermissionAction, error)
	ListDistinctResourceCodesByScope(scopeCode string) ([]string, error)
	Create(action *PermissionAction) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
}

type PermissionActionListParams struct {
	Keyword               string
	Name                  string
	ResourceCode          string
	ActionCode            string
	ModuleCode            string
	Category              string
	Source                string
	FeatureKind           string
	ScopeID               *uuid.UUID
	ScopeCode             string
	Status                string
	RequiresTenantContext *bool
}

type permissionActionRepository struct {
	db *gorm.DB
}

func NewPermissionActionRepository(db *gorm.DB) PermissionActionRepository {
	return &permissionActionRepository{db: db}
}

func (r *permissionActionRepository) List(offset, limit int, params *PermissionActionListParams) ([]PermissionAction, int64, error) {
	query := r.db.Model(&PermissionAction{}).Preload("Scope")
	if params != nil {
		if params.Keyword != "" {
			keyword := "%" + params.Keyword + "%"
			query = query.Where(
				"(name LIKE ? OR description LIKE ? OR resource_code LIKE ? OR action_code LIKE ? OR module_code LIKE ? OR category LIKE ? OR feature_kind LIKE ?)",
				keyword, keyword, keyword, keyword, keyword, keyword, keyword,
			)
		}
		if params.Name != "" {
			query = query.Where("name LIKE ?", "%"+params.Name+"%")
		}
		if params.ResourceCode != "" {
			query = query.Where("resource_code LIKE ?", "%"+params.ResourceCode+"%")
		}
		if params.ActionCode != "" {
			query = query.Where("action_code LIKE ?", "%"+params.ActionCode+"%")
		}
		if params.ModuleCode != "" {
			query = query.Where("module_code LIKE ?", "%"+params.ModuleCode+"%")
		}
		if params.Category != "" {
			query = query.Where("category LIKE ?", "%"+params.Category+"%")
		}
		if params.Source != "" {
			query = query.Where("source = ?", params.Source)
		}
		if params.FeatureKind != "" {
			query = query.Where("feature_kind = ?", params.FeatureKind)
		}
		if params.ScopeID != nil {
			query = query.Where("scope_id = ?", *params.ScopeID)
		}
		if params.ScopeCode != "" {
			query = query.Joins("JOIN scopes ON permission_actions.scope_id = scopes.id").Where("scopes.code = ?", params.ScopeCode)
		}
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}
		if params.RequiresTenantContext != nil {
			query = query.Where("requires_tenant_context = ?", *params.RequiresTenantContext)
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var actions []PermissionAction
	err := query.Offset(offset).Limit(limit).Order("sort_order ASC, created_at DESC").Find(&actions).Error
	return actions, total, err
}

func (r *permissionActionRepository) GetByID(id uuid.UUID) (*PermissionAction, error) {
	var action PermissionAction
	err := r.db.Preload("Scope").Where("id = ?", id).First(&action).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &action, nil
}

func (r *permissionActionRepository) GetByIDs(ids []uuid.UUID) ([]PermissionAction, error) {
	var actions []PermissionAction
	if len(ids) == 0 {
		return actions, nil
	}
	err := r.db.Preload("Scope").Where("id IN ?", ids).Order("sort_order ASC, created_at DESC").Find(&actions).Error
	return actions, err
}

func (r *permissionActionRepository) GetByResourceAndAction(resourceCode, actionCode, scopeCode string) (*PermissionAction, error) {
	var action PermissionAction
	query := r.db.Preload("Scope").Where("permission_actions.resource_code = ? AND permission_actions.action_code = ?", resourceCode, actionCode)
	if trimmedScopeCode := strings.TrimSpace(scopeCode); trimmedScopeCode != "" {
		query = query.Joins("JOIN scopes ON permission_actions.scope_id = scopes.id").Where("scopes.code = ?", trimmedScopeCode)
	}
	err := query.First(&action).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &action, nil
}

func (r *permissionActionRepository) GetAllEnabled() ([]PermissionAction, error) {
	var actions []PermissionAction
	err := r.db.Preload("Scope").Where("status = ?", "normal").Order("sort_order ASC, created_at DESC").Find(&actions).Error
	return actions, err
}

func (r *permissionActionRepository) ListDistinctResourceCodesByScope(scopeCode string) ([]string, error) {
	var resourceCodes []string
	query := r.db.Model(&PermissionAction{}).Distinct("permission_actions.resource_code")
	if scopeCode != "" {
		query = query.Joins("JOIN scopes ON permission_actions.scope_id = scopes.id").Where("scopes.code = ?", scopeCode)
	}
	err := query.Order("permission_actions.resource_code ASC").Pluck("permission_actions.resource_code", &resourceCodes).Error
	return resourceCodes, err
}

func (r *permissionActionRepository) Create(action *PermissionAction) error {
	return r.db.Create(action).Error
}

func (r *permissionActionRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&PermissionAction{}).Where("id = ?", id).Updates(updates).Error
}

func (r *permissionActionRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&PermissionAction{}, id).Error
}

type RoleActionPermissionRepository interface {
	GetByRoleID(roleID uuid.UUID) ([]RoleActionPermission, error)
	GetByRoleIDs(roleIDs []uuid.UUID) ([]RoleActionPermission, error)
	GetByRoleIDsAndAction(roleIDs []uuid.UUID, actionID uuid.UUID) ([]RoleActionPermission, error)
	SetRoleActions(roleID uuid.UUID, actions []RoleActionPermission) error
	DeleteByRoleID(roleID uuid.UUID) error
	DeleteByActionID(actionID uuid.UUID) error
}

type roleActionPermissionRepository struct {
	db *gorm.DB
}

func NewRoleActionPermissionRepository(db *gorm.DB) RoleActionPermissionRepository {
	return &roleActionPermissionRepository{db: db}
}

func (r *roleActionPermissionRepository) GetByRoleID(roleID uuid.UUID) ([]RoleActionPermission, error) {
	var records []RoleActionPermission
	err := r.db.Where("role_id = ?", roleID).Find(&records).Error
	return records, err
}

func (r *roleActionPermissionRepository) GetByRoleIDs(roleIDs []uuid.UUID) ([]RoleActionPermission, error) {
	var records []RoleActionPermission
	if len(roleIDs) == 0 {
		return records, nil
	}
	err := r.db.Where("role_id IN ?", roleIDs).Find(&records).Error
	return records, err
}

func (r *roleActionPermissionRepository) GetByRoleIDsAndAction(roleIDs []uuid.UUID, actionID uuid.UUID) ([]RoleActionPermission, error) {
	var records []RoleActionPermission
	if len(roleIDs) == 0 {
		return records, nil
	}
	err := r.db.Where("role_id IN ? AND action_id = ?", roleIDs, actionID).Find(&records).Error
	return records, err
}

func (r *roleActionPermissionRepository) SetRoleActions(roleID uuid.UUID, actions []RoleActionPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RoleActionPermission{}).Error; err != nil {
			return err
		}
		if len(actions) == 0 {
			return nil
		}
		return tx.Create(&actions).Error
	})
}

func (r *roleActionPermissionRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&RoleActionPermission{}).Error
}

func (r *roleActionPermissionRepository) DeleteByActionID(actionID uuid.UUID) error {
	return r.db.Where("action_id = ?", actionID).Delete(&RoleActionPermission{}).Error
}

type RoleDataPermissionRepository interface {
	GetByRoleID(roleID uuid.UUID) ([]RoleDataPermission, error)
	ReplaceRoleDataPermissions(roleID uuid.UUID, permissions []RoleDataPermission) error
	DeleteByRoleID(roleID uuid.UUID) error
}

type roleDataPermissionRepository struct {
	db *gorm.DB
}

func NewRoleDataPermissionRepository(db *gorm.DB) RoleDataPermissionRepository {
	return &roleDataPermissionRepository{db: db}
}

func (r *roleDataPermissionRepository) GetByRoleID(roleID uuid.UUID) ([]RoleDataPermission, error) {
	var records []RoleDataPermission
	err := r.db.Where("role_id = ?", roleID).Order("resource_code ASC").Find(&records).Error
	return records, err
}

func (r *roleDataPermissionRepository) ReplaceRoleDataPermissions(roleID uuid.UUID, permissions []RoleDataPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RoleDataPermission{}).Error; err != nil {
			return err
		}
		if len(permissions) == 0 {
			return nil
		}
		return tx.Create(&permissions).Error
	})
}

func (r *roleDataPermissionRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&RoleDataPermission{}).Error
}

type TenantActionPermissionRepository interface {
	GetEnabledActionIDsByTenantID(tenantID uuid.UUID) ([]uuid.UUID, error)
	ReplaceTenantActions(tenantID uuid.UUID, actionIDs []uuid.UUID) error
	CountByTenantID(tenantID uuid.UUID) (int64, error)
	IsActionEnabled(tenantID, actionID uuid.UUID) (bool, error)
	DeleteByActionID(actionID uuid.UUID) error
}

type tenantActionPermissionRepository struct {
	db *gorm.DB
}

func NewTenantActionPermissionRepository(db *gorm.DB) TenantActionPermissionRepository {
	return &tenantActionPermissionRepository{db: db}
}

func (r *tenantActionPermissionRepository) GetEnabledActionIDsByTenantID(tenantID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := r.db.Model(&TenantActionPermission{}).
		Where("tenant_id = ? AND enabled = ?", tenantID, true).
		Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (r *tenantActionPermissionRepository) ReplaceTenantActions(tenantID uuid.UUID, actionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("tenant_id = ?", tenantID).Delete(&TenantActionPermission{}).Error; err != nil {
			return err
		}
		if len(actionIDs) == 0 {
			return nil
		}
		records := make([]TenantActionPermission, 0, len(actionIDs))
		seen := make(map[uuid.UUID]struct{}, len(actionIDs))
		for _, actionID := range actionIDs {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			records = append(records, TenantActionPermission{TenantID: tenantID, ActionID: actionID, Enabled: true})
		}
		if len(records) == 0 {
			return nil
		}
		return tx.Create(&records).Error
	})
}

func (r *tenantActionPermissionRepository) CountByTenantID(tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&TenantActionPermission{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *tenantActionPermissionRepository) IsActionEnabled(tenantID, actionID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&TenantActionPermission{}).
		Where("tenant_id = ? AND action_id = ? AND enabled = ?", tenantID, actionID, true).
		Count(&count).Error
	return count > 0, err
}

func (r *tenantActionPermissionRepository) DeleteByActionID(actionID uuid.UUID) error {
	return r.db.Where("action_id = ?", actionID).Delete(&TenantActionPermission{}).Error
}

type UserActionPermissionRepository interface {
	GetByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]UserActionPermission, error)
	GetEffectiveByUserAndAction(userID uuid.UUID, tenantID *uuid.UUID, actionID uuid.UUID) ([]UserActionPermission, error)
	ReplaceUserActions(userID uuid.UUID, tenantID *uuid.UUID, actions []UserActionPermission) error
	DeleteByActionID(actionID uuid.UUID) error
}

type userActionPermissionRepository struct {
	db *gorm.DB
}

func NewUserActionPermissionRepository(db *gorm.DB) UserActionPermissionRepository {
	return &userActionPermissionRepository{db: db}
}

func (r *userActionPermissionRepository) GetByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]UserActionPermission, error) {
	var records []UserActionPermission
	query := r.db.Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	err := query.Find(&records).Error
	return records, err
}

func (r *userActionPermissionRepository) GetEffectiveByUserAndAction(userID uuid.UUID, tenantID *uuid.UUID, actionID uuid.UUID) ([]UserActionPermission, error) {
	var records []UserActionPermission
	query := r.db.Where("user_id = ? AND action_id = ?", userID, actionID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id IS NULL OR tenant_id = ?", *tenantID)
	}
	err := query.Find(&records).Error
	return records, err
}

func (r *userActionPermissionRepository) ReplaceUserActions(userID uuid.UUID, tenantID *uuid.UUID, actions []UserActionPermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		query := tx.Where("user_id = ?", userID)
		if tenantID == nil {
			query = query.Where("tenant_id IS NULL")
		} else {
			query = query.Where("tenant_id = ?", *tenantID)
		}
		if err := query.Delete(&UserActionPermission{}).Error; err != nil {
			return err
		}
		if len(actions) == 0 {
			return nil
		}
		return tx.Create(&actions).Error
	})
}

func (r *userActionPermissionRepository) DeleteByActionID(actionID uuid.UUID) error {
	return r.db.Where("action_id = ?", actionID).Delete(&UserActionPermission{}).Error
}

type APIEndpointRepository interface {
	List(offset, limit int, params *APIEndpointListParams) ([]APIEndpoint, int64, error)
	Upsert(endpoint *APIEndpoint) error
	GetByMethodAndPath(method, path string) (*APIEndpoint, error)
}

type APIEndpointListParams struct {
	Method                string
	Path                  string
	Module                string
	FeatureKind           string
	ResourceCode          string
	ActionCode            string
	ScopeCode             string
	RequiresTenantContext *bool
	Status                string
}

type apiEndpointRepository struct {
	db *gorm.DB
}

func NewAPIEndpointRepository(db *gorm.DB) APIEndpointRepository {
	return &apiEndpointRepository{db: db}
}

func (r *apiEndpointRepository) List(offset, limit int, params *APIEndpointListParams) ([]APIEndpoint, int64, error) {
	query := r.db.Model(&APIEndpoint{}).Preload("Scope")
	if params != nil {
		if params.Method != "" {
			query = query.Where("method = ?", params.Method)
		}
		if params.Path != "" {
			query = query.Where("path LIKE ?", "%"+params.Path+"%")
		}
		if params.Module != "" {
			query = query.Where("module LIKE ?", "%"+params.Module+"%")
		}
		if params.FeatureKind != "" {
			query = query.Where("feature_kind = ?", params.FeatureKind)
		}
		if params.ResourceCode != "" {
			query = query.Where("resource_code LIKE ?", "%"+params.ResourceCode+"%")
		}
		if params.ActionCode != "" {
			query = query.Where("action_code LIKE ?", "%"+params.ActionCode+"%")
		}
		if params.ScopeCode != "" {
			query = query.Joins("LEFT JOIN scopes ON api_endpoints.scope_id = scopes.id").Where("scopes.code = ?", params.ScopeCode)
		}
		if params.RequiresTenantContext != nil {
			query = query.Where("requires_tenant_context = ?", *params.RequiresTenantContext)
		}
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var endpoints []APIEndpoint
	err := query.Offset(offset).Limit(limit).Order("module ASC, path ASC, method ASC").Find(&endpoints).Error
	return endpoints, total, err
}

func (r *apiEndpointRepository) GetByMethodAndPath(method, path string) (*APIEndpoint, error) {
	var endpoint APIEndpoint
	err := r.db.Where("method = ? AND path = ?", method, path).First(&endpoint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &endpoint, nil
}

func (r *apiEndpointRepository) Upsert(endpoint *APIEndpoint) error {
	if endpoint == nil {
		return nil
	}
	updates := map[string]interface{}{
		"module":                  endpoint.Module,
		"feature_kind":            endpoint.FeatureKind,
		"handler":                 endpoint.Handler,
		"summary":                 endpoint.Summary,
		"resource_code":           endpoint.ResourceCode,
		"action_code":             endpoint.ActionCode,
		"scope_id":                endpoint.ScopeID,
		"requires_tenant_context": endpoint.RequiresTenantContext,
		"status":                  endpoint.Status,
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing APIEndpoint
		err := tx.Where("method = ? AND path = ?", endpoint.Method, endpoint.Path).First(&existing).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return tx.Create(endpoint).Error
			}
			return err
		}
		return tx.Model(&existing).Updates(updates).Error
	})
}
