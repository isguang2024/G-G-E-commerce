package repository

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// RoleRepository 全局角色仓储（仅 roles 表，无 tenant）
type RoleRepository interface {
	List(offset, limit int, code, name, description, startTime, endTime string, enabled *bool) ([]model.Role, int64, error)
	ListByScope(scope string, offset, limit int, code, name, description, startTime, endTime string, enabled *bool) ([]model.Role, int64, error)
	GetByID(id uuid.UUID) (*model.Role, error)
	GetByCode(code string) (*model.Role, error)
	GetByScopeID(scopeID uuid.UUID) ([]model.Role, error)
	Create(role *model.Role) error
	Update(role *model.Role) error
	UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) List(offset, limit int, code, name, description, startTime, endTime string, enabled *bool) ([]model.Role, int64, error) {
	var list []model.Role
	query := r.db.Model(&model.Role{}).Preload("Scope")
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}
	// 兼容旧的 enabled 参数，转换为 status 查询
	if enabled != nil {
		if *enabled {
			query = query.Where("status = ?", "normal")
		} else {
			query = query.Where("status = ?", "suspended")
		}
	}
	// 创建时间范围过滤：startTime、endTime 格式为 YYYY-MM-DD
	if startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02", endTime); err == nil {
			// 结束日期为当天 23:59:59，使用 < nextDay 实现「含当天」的查询
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	err := query.Order("sort_order ASC, created_at ASC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *roleRepository) ListByScope(scope string, offset, limit int, code, name, description, startTime, endTime string, enabled *bool) ([]model.Role, int64, error) {
	var list []model.Role
	// 通过Scope关联查询，先找到scope code对应的scope，再查询roles
	var scopeModel model.Scope
	if err := r.db.Where("code = ?", scope).First(&scopeModel).Error; err != nil {
		return nil, 0, err
	}
	query := r.db.Model(&model.Role{}).Preload("Scope").Where("scope_id = ?", scopeModel.ID)
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if description != "" {
		query = query.Where("description LIKE ?", "%"+description+"%")
	}
	// 兼容旧的 enabled 参数，转换为 status 查询
	if enabled != nil {
		if *enabled {
			query = query.Where("status = ?", "normal")
		} else {
			query = query.Where("status = ?", "suspended")
		}
	}
	// 创建时间范围过滤
	if startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endTime != "" {
		if t, err := time.Parse("2006-01-02", endTime); err == nil {
			query = query.Where("created_at < ?", t.AddDate(0, 0, 1))
		}
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	err := query.Order("sort_order ASC, created_at ASC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *roleRepository) GetByID(id uuid.UUID) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Scope").Where("id = ?", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByCode(code string) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("code = ?", code).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) Update(role *model.Role) error {
	// 使用Select指定要更新的字段，确保scope_id等字段能正确更新
	return r.db.Model(role).Select("name", "description", "scope_id", "sort_order", "updated_at").Updates(role).Error
}

func (r *roleRepository) UpdateWithMap(id uuid.UUID, updates map[string]interface{}) error {
	// 先检查记录是否存在
	var role model.Role
	if err := r.db.Where("id = ?", id).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return err
	}
	
	// 执行更新，使用 Select 显式指定要更新的字段，确保所有字段都能更新
	updateFields := make([]string, 0, len(updates))
	for k := range updates {
		updateFields = append(updateFields, k)
	}
	
	result := r.db.Model(&model.Role{}).Where("id = ?", id).Select(updateFields).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	
	// 检查是否真的更新了记录
	if result.RowsAffected == 0 {
		// 如果记录存在但没有更新，可能是值没有变化
		// 这种情况下不应该返回错误，因为 GORM 的 Updates 不会更新相同值的字段
		// 但我们可以记录一个警告
	}
	return nil
}

func (r *roleRepository) GetByScopeID(scopeID uuid.UUID) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("scope_id = ?", scopeID).Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Role{}, id).Error
}
