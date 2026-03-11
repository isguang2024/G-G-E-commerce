package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// List 分页列表（可选按用户名、手机号、邮箱、状态、角色ID、用户ID、注册来源、邀请人筛选）
	List(offset, limit int, username, userPhone, userEmail, status, roleID, id, registerSource, invitedBy string) ([]model.User, int64, error)
	// GetByID 根据ID获取用户
	GetByID(id uuid.UUID) (*model.User, error)
	// GetByIDs 批量获取用户
	GetByIDs(ids []uuid.UUID) ([]model.User, error)
	// GetByEmail 根据邮箱获取用户
	GetByEmail(email string) (*model.User, error)
	// GetByUsername 根据用户名获取用户
	GetByUsername(username string) (*model.User, error)
	// Create 创建用户
	Create(user *model.User) error
	// Update 更新用户
	Update(user *model.User) error
	// Delete 删除用户（软删除）
	Delete(id uuid.UUID) error
	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(email string) (bool, error)
	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(username string) (bool, error)
	// UpdateLastLogin 更新最后登录时间和IP
	UpdateLastLogin(id uuid.UUID, ip string) error
	// ReplaceRoles 替换用户角色
	ReplaceRoles(userID uuid.UUID, roleIDs []uuid.UUID) error
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// GetByID 根据ID获取用户（包含角色信息）
func (r *userRepository) GetByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByIDs 批量获取用户（用于查询邀请人信息）
func (r *userRepository) GetByIDs(ids []uuid.UUID) ([]model.User, error) {
	var users []model.User
	err := r.db.Where("id IN ?", ids).Find(&users).Error
	return users, err
}

// GetByEmail 根据邮箱获取用户（包含角色信息）
func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 根据用户名获取用户（包含角色信息）
func (r *userRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Roles").Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &user, nil
}


// Create 创建用户
func (r *userRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// Update 更新用户
func (r *userRepository) Update(user *model.User) error {
	return r.db.Model(user).Updates(user).Error
}

// Delete 删除用户（软删除）
func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, id).Error
}

// ExistsByEmail 检查邮箱是否存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsByUsername 检查用户名是否存在
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateLastLogin 更新最后登录时间和IP
func (r *userRepository) UpdateLastLogin(id uuid.UUID, ip string) error {
	now := r.db.NowFunc()
	return r.db.Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"last_login_ip": ip,
		}).Error
}

// List 分页列表，预加载角色
func (r *userRepository) List(offset, limit int, username, userPhone, userEmail, status, roleID, id, registerSource, invitedBy string) ([]model.User, int64, error) {
	baseQuery := r.db.Model(&model.User{})
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
	
	// 调试：输出查询条件
	// fmt.Printf("DEBUG: username=%s, userPhone=%s, userEmail=%s, status=%s, roleID=%s\n", username, userPhone, userEmail, status, roleID)
	
	var total int64
	var list []model.User
	
	// 如果指定了角色ID，通过子查询获取拥有该角色的用户ID列表
	if roleID != "" {
		// 验证 roleID 是否为有效的 UUID
		roleUUID, err := uuid.Parse(roleID)
		if err != nil {
			return nil, 0, err
		}
		// 子查询：获取拥有指定角色的用户ID（去重）
		// 使用 Model 方式创建子查询
		subQuery := r.db.Model(&model.UserRole{}).
			Select("DISTINCT user_id").
			Where("role_id = ? AND deleted_at IS NULL", roleUUID)
		
		// 在主查询中使用子查询过滤用户
		query := baseQuery.Where("id IN (?)", subQuery)
		
		// 计算总数
		if err := query.Count(&total).Error; err != nil {
			return nil, 0, err
		}
		if limit <= 0 {
			limit = 20
		}
		// 查询用户列表并预加载角色
		err = query.Preload("Roles").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
		return list, total, err
	}
	
	// 没有角色过滤条件时的普通查询
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	err := baseQuery.Preload("Roles").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// ReplaceRoles 替换用户全局角色
func (r *userRepository) ReplaceRoles(userID uuid.UUID, roleIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 先真正删除软删除的记录，避免唯一键冲突
		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		for _, roleID := range roleIDs {
			if err := tx.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
