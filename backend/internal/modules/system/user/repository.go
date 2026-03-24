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
	ListTeamRoles(tenantID uuid.UUID) ([]Role, error)
	Create(role *Role) error
	Update(role *Role) error
	Delete(id uuid.UUID) error
	GetAll() ([]Role, error)
	List() ([]Role, error)
	ListByPage(offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool) ([]Role, int64, error)
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
	err := r.db.Where("id = ?", id).First(&role).Error
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
	err := r.db.Where("code = ? AND tenant_id IS NULL", code).First(&role).Error
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
	err := r.db.Where("code = ? AND tenant_id IS NULL", code).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetByIDs(ids []uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.Where("id IN ?", ids).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) List() ([]Role, error) {
	var roles []Role
	err := r.db.Where("tenant_id IS NULL").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) ListTeamRoles(tenantID uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.
		Where("(tenant_id IS NULL AND code IN ?) OR tenant_id = ?", []string{"team_admin", "team_member"}, tenantID).
		Order("tenant_id IS NULL DESC, sort_order ASC, created_at DESC").
		Find(&roles).Error
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
	err := r.db.Where("tenant_id IS NULL").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) ListByPage(offset, limit int, roleCode, roleName, description, startTime, endTime string, enabled *bool) ([]Role, int64, error) {
	baseQuery := r.db.Model(&Role{}).Where("tenant_id IS NULL")
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
	if err != nil {
		return nil, err
	}
	if tenantID != nil && len(roleIDs) == 0 {
		return r.getTenantIdentityRoleIDs(userID, *tenantID, false)
	}
	return roleIDs, nil
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
	if err != nil {
		return nil, err
	}
	if tenantID != nil && len(codes) == 0 {
		return r.getTenantIdentityRoleCodes(userID, *tenantID, false)
	}
	return codes, nil
}

func (r *userRoleRepository) GetEffectiveRoleIDsByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]uuid.UUID, error) {
	var roleIDs []uuid.UUID
	query := r.db.Model(&UserRole{}).Where("user_id = ?", userID)
	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}
	err := query.Pluck("role_id", &roleIDs).Error
	if err != nil {
		return nil, err
	}
	if tenantID != nil && len(roleIDs) == 0 {
		return r.getTenantIdentityRoleIDs(userID, *tenantID, false)
	}
	return roleIDs, nil
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
		query = query.Where("user_roles.tenant_id = ?", *tenantID)
	}
	err := query.Distinct("user_roles.role_id").Pluck("user_roles.role_id", &roleIDs).Error
	if err != nil {
		return nil, err
	}
	if tenantID != nil && len(roleIDs) == 0 {
		return r.getTenantIdentityRoleIDs(userID, *tenantID, true)
	}
	return roleIDs, nil
}

func (r *userRoleRepository) GetEffectiveRoleCodesByUserAndTenant(userID uuid.UUID, tenantID *uuid.UUID) ([]string, error) {
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
	if err != nil {
		return nil, err
	}
	if tenantID != nil && len(codes) == 0 {
		return r.getTenantIdentityRoleCodes(userID, *tenantID, false)
	}
	return codes, nil
}

func (r *userRoleRepository) getTenantIdentityRoleIDs(userID, tenantID uuid.UUID, onlyActive bool) ([]uuid.UUID, error) {
	roles, err := r.getTenantIdentityRoles(userID, tenantID, onlyActive)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]uuid.UUID, 0, len(roles))
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}
	return roleIDs, nil
}

func (r *userRoleRepository) getTenantIdentityRoleCodes(userID, tenantID uuid.UUID, onlyActive bool) ([]string, error) {
	roles, err := r.getTenantIdentityRoles(userID, tenantID, onlyActive)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	return codes, nil
}

func (r *userRoleRepository) getTenantIdentityRoles(userID, tenantID uuid.UUID, onlyActive bool) ([]Role, error) {
	var member TenantMember
	err := r.db.Where("user_id = ? AND tenant_id = ?", userID, tenantID).First(&member).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []Role{}, nil
		}
		return nil, err
	}
	if member.Status != "active" {
		return []Role{}, nil
	}

	var roles []Role
	query := r.db.Where("code = ? AND tenant_id IS NULL", member.RoleCode)
	if onlyActive {
		query = query.Where("status = ?", "normal")
	}
	err = query.Order("sort_order ASC, created_at DESC").Find(&roles).Error
	return roles, err
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
		SortOrder   int       `gorm:"column:sort_order"`
		IsSystem    bool      `gorm:"column:is_system"`
		CreatedAt   time.Time `gorm:"column:created_at"`
		UpdatedAt   time.Time `gorm:"column:updated_at"`
	}

	var rows []userRoleRow
	if err := r.db.Table("user_roles").
		Select("user_roles.user_id, roles.id, roles.code, roles.name, roles.description, roles.status, roles.priority, roles.sort_order, roles.is_system, roles.created_at, roles.updated_at").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id IN ?", userIDs).
		Where("user_roles.tenant_id IS NULL").
		Scan(&rows).Error; err != nil {
		return err
	}

	roleCache := make(map[uuid.UUID]*Role)
	userRoleIndex := make(map[uuid.UUID]map[uuid.UUID]int, len(users))

	for _, row := range rows {
		target, ok := userIndex[row.UserID]
		if !ok {
			continue
		}
		role, exists := roleCache[row.ID]
		if !exists {
			role = &Role{
				ID:          row.ID,
				Code:        row.Code,
				Name:        row.Name,
				Description: row.Description,
				Status:      row.Status,
				Priority:    row.Priority,
				SortOrder:   row.SortOrder,
				IsSystem:    row.IsSystem,
				CreatedAt:   row.CreatedAt,
				UpdatedAt:   row.UpdatedAt,
			}
			roleCache[row.ID] = role
		}
		if _, exists := userRoleIndex[target.ID]; !exists {
			userRoleIndex[target.ID] = make(map[uuid.UUID]int)
		}
		if roleIdx, exists := userRoleIndex[target.ID][role.ID]; exists {
			target.Roles[roleIdx] = *role
			continue
		}
		userRoleIndex[target.ID][role.ID] = len(target.Roles)
		target.Roles = append(target.Roles, *role)
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

type PermissionActionRepository interface {
	List(offset, limit int, params *PermissionActionListParams) ([]PermissionAction, int64, error)
	GetByID(id uuid.UUID) (*PermissionAction, error)
	GetByIDs(ids []uuid.UUID) ([]PermissionAction, error)
	GetByPermissionKey(permissionKey string) (*PermissionAction, error)
	GetAllEnabled() ([]PermissionAction, error)
	ListDistinctModuleCodes() ([]string, error)
	Create(action *PermissionAction) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
}

type PermissionActionListParams struct {
	Keyword        string
	PermissionKey  string
	Name           string
	ModuleCode     string
	ModuleGroupID  *uuid.UUID
	FeatureGroupID *uuid.UUID
	ContextType    string
	FeatureKind    string
	Status         string
	IsBuiltin      *bool
}

type PermissionGroupRepository interface {
	List(offset, limit int, groupType string, keyword string, status string) ([]PermissionGroup, int64, error)
	GetByID(id uuid.UUID) (*PermissionGroup, error)
	GetByTypeAndCode(groupType, code string) (*PermissionGroup, error)
	ListByType(groupType string) ([]PermissionGroup, error)
	Create(group *PermissionGroup) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
}

type permissionGroupRepository struct {
	db *gorm.DB
}

func NewPermissionGroupRepository(db *gorm.DB) PermissionGroupRepository {
	return &permissionGroupRepository{db: db}
}

func (r *permissionGroupRepository) List(offset, limit int, groupType string, keyword string, status string) ([]PermissionGroup, int64, error) {
	query := r.db.Model(&PermissionGroup{})
	if strings.TrimSpace(groupType) != "" {
		query = query.Where("group_type = ?", strings.TrimSpace(groupType))
	}
	if strings.TrimSpace(keyword) != "" {
		target := "%" + strings.TrimSpace(keyword) + "%"
		query = query.Where("(code LIKE ? OR name LIKE ? OR name_en LIKE ? OR description LIKE ?)", target, target, target, target)
	}
	if strings.TrimSpace(status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(status))
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []PermissionGroup
	err := query.Offset(offset).Limit(limit).Order("group_type ASC, sort_order ASC, created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *permissionGroupRepository) GetByID(id uuid.UUID) (*PermissionGroup, error) {
	var item PermissionGroup
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *permissionGroupRepository) GetByTypeAndCode(groupType, code string) (*PermissionGroup, error) {
	var item PermissionGroup
	err := r.db.Where("group_type = ? AND code = ?", strings.TrimSpace(groupType), strings.TrimSpace(code)).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *permissionGroupRepository) ListByType(groupType string) ([]PermissionGroup, error) {
	var items []PermissionGroup
	query := r.db.Model(&PermissionGroup{})
	if strings.TrimSpace(groupType) != "" {
		query = query.Where("group_type = ?", strings.TrimSpace(groupType))
	}
	err := query.Order("sort_order ASC, created_at DESC").Find(&items).Error
	return items, err
}

func (r *permissionGroupRepository) Create(group *PermissionGroup) error {
	return r.db.Create(group).Error
}

func (r *permissionGroupRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&PermissionGroup{}).Where("id = ?", id).Updates(updates).Error
}

type permissionActionRepository struct {
	db *gorm.DB
}

func NewPermissionActionRepository(db *gorm.DB) PermissionActionRepository {
	return &permissionActionRepository{db: db}
}

func (r *permissionActionRepository) List(offset, limit int, params *PermissionActionListParams) ([]PermissionAction, int64, error) {
	query := r.db.Model(&PermissionAction{}).Preload("ModuleGroup").Preload("FeatureGroup")
	if params != nil {
		if params.Keyword != "" {
			keyword := "%" + params.Keyword + "%"
			query = query.Where(
				"(name LIKE ? OR description LIKE ? OR permission_key LIKE ? OR module_code LIKE ? OR feature_kind LIKE ?)",
				keyword, keyword, keyword, keyword, keyword,
			)
		}
		if params.PermissionKey != "" {
			query = query.Where("permission_key LIKE ?", "%"+params.PermissionKey+"%")
		}
		if params.Name != "" {
			query = query.Where("name LIKE ?", "%"+params.Name+"%")
		}
		if params.ModuleCode != "" {
			query = query.Where("module_code LIKE ?", "%"+params.ModuleCode+"%")
		}
		if params.ModuleGroupID != nil {
			query = query.Where("module_group_id = ?", *params.ModuleGroupID)
		}
		if params.FeatureGroupID != nil {
			query = query.Where("feature_group_id = ?", *params.FeatureGroupID)
		}
		if params.ContextType != "" {
			query = query.Where("context_type = ?", params.ContextType)
		}
		if params.FeatureKind != "" {
			query = query.Where("feature_kind = ?", params.FeatureKind)
		}
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}
		if params.IsBuiltin != nil {
			query = query.Where("is_builtin = ?", *params.IsBuiltin)
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
	err := r.db.Preload("ModuleGroup").Preload("FeatureGroup").Where("id = ?", id).First(&action).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &action, nil
}

func (r *permissionActionRepository) GetByPermissionKey(permissionKey string) (*PermissionAction, error) {
	var action PermissionAction
	err := r.db.Preload("ModuleGroup").Preload("FeatureGroup").Where("permission_keys.permission_key = ?", permissionKey).First(&action).Error
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
	err := r.db.Preload("ModuleGroup").Preload("FeatureGroup").Where("id IN ?", ids).Order("sort_order ASC, created_at DESC").Find(&actions).Error
	return actions, err
}

func (r *permissionActionRepository) GetAllEnabled() ([]PermissionAction, error) {
	var actions []PermissionAction
	err := r.db.Preload("ModuleGroup").Preload("FeatureGroup").Where("status = ?", "normal").Order("sort_order ASC, created_at DESC").Find(&actions).Error
	return actions, err
}

func (r *permissionActionRepository) ListDistinctModuleCodes() ([]string, error) {
	var moduleCodes []string
	err := r.db.Model(&PermissionAction{}).
		Where("COALESCE(permission_keys.module_code, '') <> ''").
		Distinct("permission_keys.module_code").
		Order("permission_keys.module_code ASC").
		Pluck("permission_keys.module_code", &moduleCodes).Error
	return moduleCodes, err
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
	GetByID(id uuid.UUID) (*APIEndpoint, error)
	Create(endpoint *APIEndpoint) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
}

type APIEndpointCategoryRepository interface {
	List() ([]APIEndpointCategory, error)
	GetByID(id uuid.UUID) (*APIEndpointCategory, error)
	GetByCode(code string) (*APIEndpointCategory, error)
	Create(item *APIEndpointCategory) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
}

type FeaturePackageRepository interface {
	List(offset, limit int, params *FeaturePackageListParams) ([]FeaturePackage, int64, error)
	GetByID(id uuid.UUID) (*FeaturePackage, error)
	GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error)
	GetByPackageKey(packageKey string) (*FeaturePackage, error)
	Create(item *FeaturePackage) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
}

type FeaturePackageListParams struct {
	Keyword     string
	PackageKey  string
	PackageType string
	Name        string
	ContextType string
	Status      string
}

type featurePackageRepository struct {
	db *gorm.DB
}

func NewFeaturePackageRepository(db *gorm.DB) FeaturePackageRepository {
	return &featurePackageRepository{db: db}
}

func (r *featurePackageRepository) List(offset, limit int, params *FeaturePackageListParams) ([]FeaturePackage, int64, error) {
	query := r.db.Model(&FeaturePackage{})
	if params != nil {
		if params.Keyword != "" {
			keyword := "%" + params.Keyword + "%"
			query = query.Where("(package_key LIKE ? OR name LIKE ? OR description LIKE ?)", keyword, keyword, keyword)
		}
		if params.PackageKey != "" {
			query = query.Where("package_key LIKE ?", "%"+params.PackageKey+"%")
		}
		if params.PackageType != "" {
			query = query.Where("package_type = ?", params.PackageType)
		}
		if params.Name != "" {
			query = query.Where("name LIKE ?", "%"+params.Name+"%")
		}
		if params.ContextType != "" {
			switch params.ContextType {
			case "platform", "team":
				query = query.Where("(context_type = ? OR context_type = ?)", params.ContextType, "common")
			default:
				query = query.Where("context_type = ?", params.ContextType)
			}
		}
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []FeaturePackage
	err := query.Offset(offset).Limit(limit).Order("sort_order ASC, created_at DESC").Find(&items).Error
	return items, total, err
}

func (r *featurePackageRepository) GetByID(id uuid.UUID) (*FeaturePackage, error) {
	var item FeaturePackage
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *featurePackageRepository) GetByIDs(ids []uuid.UUID) ([]FeaturePackage, error) {
	var items []FeaturePackage
	if len(ids) == 0 {
		return items, nil
	}
	err := r.db.Where("id IN ?", ids).Order("sort_order ASC, created_at DESC").Find(&items).Error
	return items, err
}

func (r *featurePackageRepository) GetByPackageKey(packageKey string) (*FeaturePackage, error) {
	var item FeaturePackage
	err := r.db.Where("package_key = ?", packageKey).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *featurePackageRepository) Create(item *FeaturePackage) error {
	return r.db.Create(item).Error
}

func (r *featurePackageRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&FeaturePackage{}).Where("id = ?", id).Updates(updates).Error
}

func (r *featurePackageRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&FeaturePackage{}, id).Error
}

type FeaturePackageActionRepository interface {
	GetActionIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error)
	GetPackageIDsByActionID(actionID uuid.UUID) ([]uuid.UUID, error)
	CountByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	ReplacePackageActions(packageID uuid.UUID, actionIDs []uuid.UUID) error
	DeleteByPackageID(packageID uuid.UUID) error
	DeleteByActionID(actionID uuid.UUID) error
}

type featurePackageActionRepository struct {
	db *gorm.DB
}

func NewFeaturePackageActionRepository(db *gorm.DB) FeaturePackageActionRepository {
	return &featurePackageActionRepository{db: db}
}

func (r *featurePackageActionRepository) GetActionIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := r.db.Model(&FeaturePackageAction{}).Where("package_id = ?", packageID).Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (r *featurePackageActionRepository) GetPackageIDsByActionID(actionID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := r.db.Model(&FeaturePackageAction{}).Where("action_id = ?", actionID).Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (r *featurePackageActionRepository) CountByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	result := make(map[uuid.UUID]int64, len(packageIDs))
	if len(packageIDs) == 0 {
		return result, nil
	}
	type row struct {
		PackageID uuid.UUID
		Total     int64
	}
	var rows []row
	if err := r.db.Model(&FeaturePackageAction{}).
		Select("package_id, COUNT(*) AS total").
		Where("package_id IN ?", packageIDs).
		Group("package_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.PackageID] = item.Total
	}
	return result, nil
}

func (r *featurePackageActionRepository) ReplacePackageActions(packageID uuid.UUID, actionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("package_id = ?", packageID).Delete(&FeaturePackageAction{}).Error; err != nil {
			return err
		}
		if len(actionIDs) == 0 {
			return nil
		}
		items := make([]FeaturePackageAction, 0, len(actionIDs))
		seen := make(map[uuid.UUID]struct{}, len(actionIDs))
		for _, actionID := range actionIDs {
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			items = append(items, FeaturePackageAction{PackageID: packageID, ActionID: actionID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *featurePackageActionRepository) DeleteByPackageID(packageID uuid.UUID) error {
	return r.db.Where("package_id = ?", packageID).Delete(&FeaturePackageAction{}).Error
}

func (r *featurePackageActionRepository) DeleteByActionID(actionID uuid.UUID) error {
	return r.db.Where("action_id = ?", actionID).Delete(&FeaturePackageAction{}).Error
}

type FeaturePackageMenuRepository interface {
	GetMenuIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error)
	GetMenuIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error)
	CountByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	ReplacePackageMenus(packageID uuid.UUID, menuIDs []uuid.UUID) error
	DeleteByPackageID(packageID uuid.UUID) error
	DeleteByMenuID(menuID uuid.UUID) error
}

type featurePackageMenuRepository struct {
	db *gorm.DB
}

func NewFeaturePackageMenuRepository(db *gorm.DB) FeaturePackageMenuRepository {
	return &featurePackageMenuRepository{db: db}
}

func (r *featurePackageMenuRepository) GetMenuIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := r.db.Model(&FeaturePackageMenu{}).Where("package_id = ?", packageID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *featurePackageMenuRepository) GetMenuIDsByPackageIDs(packageIDs []uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	if len(packageIDs) == 0 {
		return menuIDs, nil
	}
	err := r.db.Model(&FeaturePackageMenu{}).Distinct("menu_id").Where("package_id IN ?", packageIDs).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *featurePackageMenuRepository) CountByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	result := make(map[uuid.UUID]int64, len(packageIDs))
	if len(packageIDs) == 0 {
		return result, nil
	}
	type row struct {
		PackageID uuid.UUID
		Total     int64
	}
	var rows []row
	if err := r.db.Model(&FeaturePackageMenu{}).
		Select("package_id, COUNT(*) AS total").
		Where("package_id IN ?", packageIDs).
		Group("package_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.PackageID] = item.Total
	}
	return result, nil
}

func (r *featurePackageMenuRepository) ReplacePackageMenus(packageID uuid.UUID, menuIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("package_id = ?", packageID).Delete(&FeaturePackageMenu{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		items := make([]FeaturePackageMenu, 0, len(menuIDs))
		seen := make(map[uuid.UUID]struct{}, len(menuIDs))
		for _, menuID := range menuIDs {
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			items = append(items, FeaturePackageMenu{PackageID: packageID, MenuID: menuID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *featurePackageMenuRepository) DeleteByPackageID(packageID uuid.UUID) error {
	return r.db.Where("package_id = ?", packageID).Delete(&FeaturePackageMenu{}).Error
}

func (r *featurePackageMenuRepository) DeleteByMenuID(menuID uuid.UUID) error {
	return r.db.Where("menu_id = ?", menuID).Delete(&FeaturePackageMenu{}).Error
}

type TeamFeaturePackageRepository interface {
	GetPackageIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error)
	GetTeamIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error)
	CountByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID]int64, error)
	ReplaceTeamPackages(teamID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
	DeleteByPackageID(packageID uuid.UUID) error
}

type teamFeaturePackageRepository struct {
	db *gorm.DB
}

func NewTeamFeaturePackageRepository(db *gorm.DB) TeamFeaturePackageRepository {
	return &teamFeaturePackageRepository{db: db}
}

func (r *teamFeaturePackageRepository) GetPackageIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := r.db.Model(&TeamFeaturePackage{}).
		Where("team_id = ? AND enabled = ?", teamID, true).
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (r *teamFeaturePackageRepository) GetTeamIDsByPackageID(packageID uuid.UUID) ([]uuid.UUID, error) {
	var teamIDs []uuid.UUID
	err := r.db.Model(&TeamFeaturePackage{}).
		Where("package_id = ? AND enabled = ?", packageID, true).
		Pluck("team_id", &teamIDs).Error
	return teamIDs, err
}

func (r *teamFeaturePackageRepository) CountByPackageIDs(packageIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	result := make(map[uuid.UUID]int64, len(packageIDs))
	if len(packageIDs) == 0 {
		return result, nil
	}
	type row struct {
		PackageID uuid.UUID
		Total     int64
	}
	var rows []row
	if err := r.db.Model(&TeamFeaturePackage{}).
		Select("package_id, COUNT(*) AS total").
		Where("package_id IN ? AND enabled = ?", packageIDs, true).
		Group("package_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, item := range rows {
		result[item.PackageID] = item.Total
	}
	return result, nil
}

func (r *teamFeaturePackageRepository) ReplaceTeamPackages(teamID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("team_id = ?", teamID).Delete(&TeamFeaturePackage{}).Error; err != nil {
			return err
		}
		if len(packageIDs) == 0 {
			return nil
		}
		items := make([]TeamFeaturePackage, 0, len(packageIDs))
		seen := make(map[uuid.UUID]struct{}, len(packageIDs))
		now := time.Now()
		for _, packageID := range packageIDs {
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			items = append(items, TeamFeaturePackage{
				TeamID:    teamID,
				PackageID: packageID,
				Enabled:   true,
				GrantedBy: grantedBy,
				GrantedAt: &now,
			})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *teamFeaturePackageRepository) DeleteByPackageID(packageID uuid.UUID) error {
	return r.db.Where("package_id = ?", packageID).Delete(&TeamFeaturePackage{}).Error
}

type RoleFeaturePackageRepository interface {
	GetPackageIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error)
	ReplaceRolePackages(roleID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
	DeleteByRoleID(roleID uuid.UUID) error
	DeleteByPackageID(packageID uuid.UUID) error
}

type roleFeaturePackageRepository struct {
	db *gorm.DB
}

func NewRoleFeaturePackageRepository(db *gorm.DB) RoleFeaturePackageRepository {
	return &roleFeaturePackageRepository{db: db}
}

func (r *roleFeaturePackageRepository) GetPackageIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := r.db.Model(&RoleFeaturePackage{}).
		Where("role_id = ? AND enabled = ?", roleID, true).
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (r *roleFeaturePackageRepository) ReplaceRolePackages(roleID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RoleFeaturePackage{}).Error; err != nil {
			return err
		}
		if len(packageIDs) == 0 {
			return nil
		}
		items := make([]RoleFeaturePackage, 0, len(packageIDs))
		seen := make(map[uuid.UUID]struct{}, len(packageIDs))
		now := time.Now()
		for _, packageID := range packageIDs {
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			items = append(items, RoleFeaturePackage{
				RoleID:    roleID,
				PackageID: packageID,
				Enabled:   true,
				GrantedBy: grantedBy,
				GrantedAt: &now,
			})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *roleFeaturePackageRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&RoleFeaturePackage{}).Error
}

func (r *roleFeaturePackageRepository) DeleteByPackageID(packageID uuid.UUID) error {
	return r.db.Where("package_id = ?", packageID).Delete(&RoleFeaturePackage{}).Error
}

type FeaturePackageBundleRepository interface {
	GetChildPackageIDs(packageID uuid.UUID) ([]uuid.UUID, error)
	GetParentPackageIDs(childPackageID uuid.UUID) ([]uuid.UUID, error)
	ReplaceChildPackages(packageID uuid.UUID, childPackageIDs []uuid.UUID) error
	DeleteByPackageID(packageID uuid.UUID) error
	DeleteByChildPackageID(childPackageID uuid.UUID) error
}

type featurePackageBundleRepository struct {
	db *gorm.DB
}

func NewFeaturePackageBundleRepository(db *gorm.DB) FeaturePackageBundleRepository {
	return &featurePackageBundleRepository{db: db}
}

func (r *featurePackageBundleRepository) GetChildPackageIDs(packageID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&FeaturePackageBundle{}).Where("package_id = ?", packageID).Pluck("child_package_id", &ids).Error
	return ids, err
}

func (r *featurePackageBundleRepository) GetParentPackageIDs(childPackageID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&FeaturePackageBundle{}).Where("child_package_id = ?", childPackageID).Pluck("package_id", &ids).Error
	return ids, err
}

func (r *featurePackageBundleRepository) ReplaceChildPackages(packageID uuid.UUID, childPackageIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("package_id = ?", packageID).Delete(&FeaturePackageBundle{}).Error; err != nil {
			return err
		}
		if len(childPackageIDs) == 0 {
			return nil
		}
		items := make([]FeaturePackageBundle, 0, len(childPackageIDs))
		seen := make(map[uuid.UUID]struct{}, len(childPackageIDs))
		for _, childPackageID := range childPackageIDs {
			if _, ok := seen[childPackageID]; ok {
				continue
			}
			seen[childPackageID] = struct{}{}
			items = append(items, FeaturePackageBundle{PackageID: packageID, ChildPackageID: childPackageID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *featurePackageBundleRepository) DeleteByPackageID(packageID uuid.UUID) error {
	return r.db.Where("package_id = ?", packageID).Delete(&FeaturePackageBundle{}).Error
}

func (r *featurePackageBundleRepository) DeleteByChildPackageID(childPackageID uuid.UUID) error {
	return r.db.Where("child_package_id = ?", childPackageID).Delete(&FeaturePackageBundle{}).Error
}

type UserFeaturePackageRepository interface {
	GetPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
	ReplaceUserPackages(userID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error
	DeleteByUserID(userID uuid.UUID) error
	DeleteByPackageID(packageID uuid.UUID) error
}

type userFeaturePackageRepository struct {
	db *gorm.DB
}

func NewUserFeaturePackageRepository(db *gorm.DB) UserFeaturePackageRepository {
	return &userFeaturePackageRepository{db: db}
}

func (r *userFeaturePackageRepository) GetPackageIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error) {
	var packageIDs []uuid.UUID
	err := r.db.Model(&UserFeaturePackage{}).
		Where("user_id = ? AND enabled = ?", userID, true).
		Pluck("package_id", &packageIDs).Error
	return packageIDs, err
}

func (r *userFeaturePackageRepository) ReplaceUserPackages(userID uuid.UUID, packageIDs []uuid.UUID, grantedBy *uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&UserFeaturePackage{}).Error; err != nil {
			return err
		}
		if len(packageIDs) == 0 {
			return nil
		}
		now := time.Now()
		items := make([]UserFeaturePackage, 0, len(packageIDs))
		seen := make(map[uuid.UUID]struct{}, len(packageIDs))
		for _, packageID := range packageIDs {
			if _, ok := seen[packageID]; ok {
				continue
			}
			seen[packageID] = struct{}{}
			items = append(items, UserFeaturePackage{
				UserID:    userID,
				PackageID: packageID,
				Enabled:   true,
				GrantedBy: grantedBy,
				GrantedAt: &now,
			})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *userFeaturePackageRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&UserFeaturePackage{}).Error
}

func (r *userFeaturePackageRepository) DeleteByPackageID(packageID uuid.UUID) error {
	return r.db.Where("package_id = ?", packageID).Delete(&UserFeaturePackage{}).Error
}

type RoleHiddenMenuRepository interface {
	GetMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error)
	ReplaceRoleHiddenMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error
	DeleteByRoleID(roleID uuid.UUID) error
	DeleteByMenuID(menuID uuid.UUID) error
}

type roleHiddenMenuRepository struct {
	db *gorm.DB
}

func NewRoleHiddenMenuRepository(db *gorm.DB) RoleHiddenMenuRepository {
	return &roleHiddenMenuRepository{db: db}
}

func (r *roleHiddenMenuRepository) GetMenuIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := r.db.Model(&RoleHiddenMenu{}).Where("role_id = ?", roleID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *roleHiddenMenuRepository) ReplaceRoleHiddenMenus(roleID uuid.UUID, menuIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RoleHiddenMenu{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		items := make([]RoleHiddenMenu, 0, len(menuIDs))
		seen := make(map[uuid.UUID]struct{}, len(menuIDs))
		for _, menuID := range menuIDs {
			if menuID == uuid.Nil {
				continue
			}
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			items = append(items, RoleHiddenMenu{RoleID: roleID, MenuID: menuID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *roleHiddenMenuRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&RoleHiddenMenu{}).Error
}

func (r *roleHiddenMenuRepository) DeleteByMenuID(menuID uuid.UUID) error {
	return r.db.Where("menu_id = ?", menuID).Delete(&RoleHiddenMenu{}).Error
}

type RoleDisabledActionRepository interface {
	GetActionIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error)
	ReplaceRoleDisabledActions(roleID uuid.UUID, actionIDs []uuid.UUID) error
	DeleteByRoleID(roleID uuid.UUID) error
	DeleteByActionID(actionID uuid.UUID) error
}

type roleDisabledActionRepository struct {
	db *gorm.DB
}

func NewRoleDisabledActionRepository(db *gorm.DB) RoleDisabledActionRepository {
	return &roleDisabledActionRepository{db: db}
}

func (r *roleDisabledActionRepository) GetActionIDsByRoleID(roleID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := r.db.Model(&RoleDisabledAction{}).Where("role_id = ?", roleID).Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (r *roleDisabledActionRepository) ReplaceRoleDisabledActions(roleID uuid.UUID, actionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&RoleDisabledAction{}).Error; err != nil {
			return err
		}
		if len(actionIDs) == 0 {
			return nil
		}
		items := make([]RoleDisabledAction, 0, len(actionIDs))
		seen := make(map[uuid.UUID]struct{}, len(actionIDs))
		for _, actionID := range actionIDs {
			if actionID == uuid.Nil {
				continue
			}
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			items = append(items, RoleDisabledAction{RoleID: roleID, ActionID: actionID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *roleDisabledActionRepository) DeleteByRoleID(roleID uuid.UUID) error {
	return r.db.Where("role_id = ?", roleID).Delete(&RoleDisabledAction{}).Error
}

func (r *roleDisabledActionRepository) DeleteByActionID(actionID uuid.UUID) error {
	return r.db.Where("action_id = ?", actionID).Delete(&RoleDisabledAction{}).Error
}

type TeamBlockedMenuRepository interface {
	GetMenuIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error)
	ReplaceTeamBlockedMenus(teamID uuid.UUID, menuIDs []uuid.UUID) error
	DeleteByTeamID(teamID uuid.UUID) error
	DeleteByMenuID(menuID uuid.UUID) error
}

type teamBlockedMenuRepository struct {
	db *gorm.DB
}

func NewTeamBlockedMenuRepository(db *gorm.DB) TeamBlockedMenuRepository {
	return &teamBlockedMenuRepository{db: db}
}

func (r *teamBlockedMenuRepository) GetMenuIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := r.db.Model(&TeamBlockedMenu{}).Where("team_id = ?", teamID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *teamBlockedMenuRepository) ReplaceTeamBlockedMenus(teamID uuid.UUID, menuIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("team_id = ?", teamID).Delete(&TeamBlockedMenu{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		items := make([]TeamBlockedMenu, 0, len(menuIDs))
		seen := make(map[uuid.UUID]struct{}, len(menuIDs))
		for _, menuID := range menuIDs {
			if menuID == uuid.Nil {
				continue
			}
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			items = append(items, TeamBlockedMenu{TeamID: teamID, MenuID: menuID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *teamBlockedMenuRepository) DeleteByTeamID(teamID uuid.UUID) error {
	return r.db.Where("team_id = ?", teamID).Delete(&TeamBlockedMenu{}).Error
}

func (r *teamBlockedMenuRepository) DeleteByMenuID(menuID uuid.UUID) error {
	return r.db.Where("menu_id = ?", menuID).Delete(&TeamBlockedMenu{}).Error
}

type TeamBlockedActionRepository interface {
	GetActionIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error)
	ReplaceTeamBlockedActions(teamID uuid.UUID, actionIDs []uuid.UUID) error
	DeleteByTeamID(teamID uuid.UUID) error
	DeleteByActionID(actionID uuid.UUID) error
}

type teamBlockedActionRepository struct {
	db *gorm.DB
}

func NewTeamBlockedActionRepository(db *gorm.DB) TeamBlockedActionRepository {
	return &teamBlockedActionRepository{db: db}
}

func (r *teamBlockedActionRepository) GetActionIDsByTeamID(teamID uuid.UUID) ([]uuid.UUID, error) {
	var actionIDs []uuid.UUID
	err := r.db.Model(&TeamBlockedAction{}).Where("team_id = ?", teamID).Pluck("action_id", &actionIDs).Error
	return actionIDs, err
}

func (r *teamBlockedActionRepository) ReplaceTeamBlockedActions(teamID uuid.UUID, actionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("team_id = ?", teamID).Delete(&TeamBlockedAction{}).Error; err != nil {
			return err
		}
		if len(actionIDs) == 0 {
			return nil
		}
		items := make([]TeamBlockedAction, 0, len(actionIDs))
		seen := make(map[uuid.UUID]struct{}, len(actionIDs))
		for _, actionID := range actionIDs {
			if actionID == uuid.Nil {
				continue
			}
			if _, ok := seen[actionID]; ok {
				continue
			}
			seen[actionID] = struct{}{}
			items = append(items, TeamBlockedAction{TeamID: teamID, ActionID: actionID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *teamBlockedActionRepository) DeleteByTeamID(teamID uuid.UUID) error {
	return r.db.Where("team_id = ?", teamID).Delete(&TeamBlockedAction{}).Error
}

func (r *teamBlockedActionRepository) DeleteByActionID(actionID uuid.UUID) error {
	return r.db.Where("action_id = ?", actionID).Delete(&TeamBlockedAction{}).Error
}

type UserHiddenMenuRepository interface {
	GetMenuIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error)
	ReplaceUserHiddenMenus(userID uuid.UUID, menuIDs []uuid.UUID) error
	DeleteByUserID(userID uuid.UUID) error
	DeleteByMenuID(menuID uuid.UUID) error
}

type userHiddenMenuRepository struct {
	db *gorm.DB
}

func NewUserHiddenMenuRepository(db *gorm.DB) UserHiddenMenuRepository {
	return &userHiddenMenuRepository{db: db}
}

func (r *userHiddenMenuRepository) GetMenuIDsByUserID(userID uuid.UUID) ([]uuid.UUID, error) {
	var menuIDs []uuid.UUID
	err := r.db.Model(&UserHiddenMenu{}).Where("user_id = ?", userID).Pluck("menu_id", &menuIDs).Error
	return menuIDs, err
}

func (r *userHiddenMenuRepository) ReplaceUserHiddenMenus(userID uuid.UUID, menuIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&UserHiddenMenu{}).Error; err != nil {
			return err
		}
		if len(menuIDs) == 0 {
			return nil
		}
		items := make([]UserHiddenMenu, 0, len(menuIDs))
		seen := make(map[uuid.UUID]struct{}, len(menuIDs))
		for _, menuID := range menuIDs {
			if menuID == uuid.Nil {
				continue
			}
			if _, ok := seen[menuID]; ok {
				continue
			}
			seen[menuID] = struct{}{}
			items = append(items, UserHiddenMenu{UserID: userID, MenuID: menuID})
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}

func (r *userHiddenMenuRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&UserHiddenMenu{}).Error
}

func (r *userHiddenMenuRepository) DeleteByMenuID(menuID uuid.UUID) error {
	return r.db.Where("menu_id = ?", menuID).Delete(&UserHiddenMenu{}).Error
}

type APIEndpointListParams struct {
	Method        string
	PermissionKey string
	Keyword       string
	Path          string
	Module        string
	CategoryID    string
	ContextScope  string
	Source        string
	FeatureKind   string
	Status        string
	HasPermission *bool
	HasCategory   *bool
}

type apiEndpointRepository struct {
	db *gorm.DB
}

func NewAPIEndpointRepository(db *gorm.DB) APIEndpointRepository {
	return &apiEndpointRepository{db: db}
}

func (r *apiEndpointRepository) List(offset, limit int, params *APIEndpointListParams) ([]APIEndpoint, int64, error) {
	query := r.db.Model(&APIEndpoint{})
	if params != nil {
		if params.Method != "" {
			query = query.Where("method = ?", params.Method)
		}
		if params.Keyword != "" {
			keyword := "%" + params.Keyword + "%"
			query = query.Where("(path LIKE ? OR module LIKE ? OR summary LIKE ? OR handler LIKE ?)", keyword, keyword, keyword, keyword)
		}
		if params.Path != "" {
			query = query.Where("path LIKE ?", "%"+params.Path+"%")
		}
		if params.Module != "" {
			query = query.Where("module LIKE ?", "%"+params.Module+"%")
		}
		if params.CategoryID != "" {
			query = query.Where("category_id = ?", params.CategoryID)
		}
		if params.ContextScope != "" {
			query = query.Where("context_scope = ?", params.ContextScope)
		}
		if params.Source != "" {
			query = query.Where("source = ?", params.Source)
		}
		if params.FeatureKind != "" {
			query = query.Where("feature_kind = ?", params.FeatureKind)
		}
		if params.Status != "" {
			query = query.Where("status = ?", params.Status)
		}
		if params.HasPermission != nil {
			if *params.HasPermission {
				query = query.Where("EXISTS (SELECT 1 FROM api_endpoint_permission_bindings b WHERE b.endpoint_id = api_endpoints.id)")
			} else {
				query = query.Where("NOT EXISTS (SELECT 1 FROM api_endpoint_permission_bindings b WHERE b.endpoint_id = api_endpoints.id)")
			}
		}
		if params.HasCategory != nil {
			if *params.HasCategory {
				query = query.Where("category_id IS NOT NULL")
			} else {
				query = query.Where("category_id IS NULL")
			}
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

func (r *apiEndpointRepository) GetByID(id uuid.UUID) (*APIEndpoint, error) {
	var endpoint APIEndpoint
	err := r.db.Where("id = ?", id).First(&endpoint).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &endpoint, nil
}

func (r *apiEndpointRepository) Create(endpoint *APIEndpoint) error {
	return r.db.Create(endpoint).Error
}

func (r *apiEndpointRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&APIEndpoint{}).Where("id = ?", id).Updates(updates).Error
}

func (r *apiEndpointRepository) Upsert(endpoint *APIEndpoint) error {
	if endpoint == nil {
		return nil
	}
	updates := map[string]interface{}{
		"module":        endpoint.Module,
		"feature_kind":  endpoint.FeatureKind,
		"handler":       endpoint.Handler,
		"summary":       endpoint.Summary,
		"category_id":   endpoint.CategoryID,
		"context_scope": endpoint.ContextScope,
		"source":        endpoint.Source,
		"status":        endpoint.Status,
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

type apiEndpointCategoryRepository struct {
	db *gorm.DB
}

func NewAPIEndpointCategoryRepository(db *gorm.DB) APIEndpointCategoryRepository {
	return &apiEndpointCategoryRepository{db: db}
}

func (r *apiEndpointCategoryRepository) List() ([]APIEndpointCategory, error) {
	var items []APIEndpointCategory
	err := r.db.Order("sort_order ASC, created_at ASC").Find(&items).Error
	return items, err
}

func (r *apiEndpointCategoryRepository) GetByID(id uuid.UUID) (*APIEndpointCategory, error) {
	var item APIEndpointCategory
	err := r.db.Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *apiEndpointCategoryRepository) GetByCode(code string) (*APIEndpointCategory, error) {
	var item APIEndpointCategory
	err := r.db.Where("code = ?", code).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &item, nil
}

func (r *apiEndpointCategoryRepository) Create(item *APIEndpointCategory) error {
	return r.db.Create(item).Error
}

func (r *apiEndpointCategoryRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	return r.db.Model(&APIEndpointCategory{}).Where("id = ?", id).Updates(updates).Error
}

type APIEndpointPermissionBindingRepository interface {
	ListByEndpointIDs(endpointIDs []uuid.UUID) ([]APIEndpointPermissionBinding, error)
	ListByEndpointID(endpointID uuid.UUID) ([]APIEndpointPermissionBinding, error)
	ListEndpointIDsByPermissionKey(permissionKey string) ([]uuid.UUID, error)
	ReplaceByEndpointID(endpointID uuid.UUID, items []APIEndpointPermissionBinding) error
}

type apiEndpointPermissionBindingRepository struct {
	db *gorm.DB
}

func NewAPIEndpointPermissionBindingRepository(db *gorm.DB) APIEndpointPermissionBindingRepository {
	return &apiEndpointPermissionBindingRepository{db: db}
}

func (r *apiEndpointPermissionBindingRepository) ListByEndpointIDs(endpointIDs []uuid.UUID) ([]APIEndpointPermissionBinding, error) {
	if len(endpointIDs) == 0 {
		return []APIEndpointPermissionBinding{}, nil
	}
	var items []APIEndpointPermissionBinding
	err := r.db.Where("endpoint_id IN ?", endpointIDs).Order("sort_order ASC, created_at ASC").Find(&items).Error
	return items, err
}

func (r *apiEndpointPermissionBindingRepository) ListByEndpointID(endpointID uuid.UUID) ([]APIEndpointPermissionBinding, error) {
	var items []APIEndpointPermissionBinding
	err := r.db.Where("endpoint_id = ?", endpointID).Order("sort_order ASC, created_at ASC").Find(&items).Error
	return items, err
}

func (r *apiEndpointPermissionBindingRepository) ListEndpointIDsByPermissionKey(permissionKey string) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Model(&APIEndpointPermissionBinding{}).
		Where("permission_key = ?", permissionKey).
		Pluck("endpoint_id", &ids).Error
	return ids, err
}

func (r *apiEndpointPermissionBindingRepository) ReplaceByEndpointID(endpointID uuid.UUID, items []APIEndpointPermissionBinding) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("endpoint_id = ?", endpointID).Delete(&APIEndpointPermissionBinding{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		return tx.Create(&items).Error
	})
}
