package user

import (
	"errors"

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
	err := r.db.Preload("Roles").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIDs(ids []uuid.UUID) ([]User, error) {
	var users []User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

func (r *userRepository) GetByEmail(email string) (*User, error) {
	var user User
	err := r.db.Preload("Roles").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*User, error) {
	var user User
	err := r.db.Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
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
	return r.db.Delete(&User{}, id).Error
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
	err := baseQuery.Preload("Roles").Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	return users, total, err
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
	err := baseQuery.Offset(offset).Limit(limit).Order("created_at DESC").Find(&roles).Error
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
		Where("tenant_members.tenant_id = ? AND tenant_members.role_code = ?", tenantID, "owner").
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
		if params.UserName != "" {
			var userIDs []uuid.UUID
			r.db.Model(&User{}).Where("username LIKE ?", "%"+params.UserName+"%").Pluck("id", &userIDs)
			if len(userIDs) > 0 {
				query = query.Where("user_id IN ?", userIDs)
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
	ReplaceUserRoles(userID uuid.UUID, tenantID *uuid.UUID, roleIDs []uuid.UUID) error
	AssignRole(userID, roleID uuid.UUID, tenantID *uuid.UUID) error
	SetUserRoles(userID uuid.UUID, roleIDs []uuid.UUID, tenantID *uuid.UUID) error
	RemoveUserRole(userID uuid.UUID, tenantID *uuid.UUID) error
}

type userRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) GetRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID, tenantMemberRepo TenantMemberRepository) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	err := r.db.Model(&UserRole{}).Where("user_id = ?", userID).Pluck("role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *userRoleRepository) GetRoleCodesByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
	var codes []string
	err := r.db.Model(&UserRole{}).
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Pluck("roles.code", &codes).Error
	return codes, err
}

func (r *userRoleRepository) ReplaceUserRoles(userID uuid.UUID, tenantID *uuid.UUID, roleIDs []uuid.UUID) error {
	tx := r.db.Begin()
	if err := tx.Where("user_id = ?", userID).Delete(&UserRole{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	userRoles := make([]UserRole, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		userRoles = append(userRoles, UserRole{UserID: userID, RoleID: roleID})
	}
	if err := tx.Create(&userRoles).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (r *userRoleRepository) AssignRole(userID, roleID uuid.UUID, tenantID *uuid.UUID) error {
	userRole := UserRole{UserID: userID, RoleID: roleID}
	return r.db.Create(&userRole).Error
}

func (r *userRoleRepository) SetUserRoles(userID uuid.UUID, roleIDs []uuid.UUID, tenantID *uuid.UUID) error {
	return r.ReplaceUserRoles(userID, tenantID, roleIDs)
}

func (r *userRoleRepository) RemoveUserRole(userID uuid.UUID, tenantID *uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&UserRole{}).Error
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
	err := r.db.Model(&RoleMenu{}).Where("role_id IN ?", roleIDs).Pluck("menu_id", &menuIDs).Error
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
