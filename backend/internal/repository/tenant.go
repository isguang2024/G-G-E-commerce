package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// TenantRepository 团队（租户）仓储
type TenantRepository interface {
	List(offset, limit int, name, status string) ([]model.Tenant, int64, error)
	GetByID(id uuid.UUID) (*model.Tenant, error)
	Create(t *model.Tenant) error
	Update(t *model.Tenant) error
	Delete(id uuid.UUID) error
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) List(offset, limit int, name, status string) ([]model.Tenant, int64, error) {
	var list []model.Tenant
	query := r.db.Model(&model.Tenant{})
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

func (r *tenantRepository) GetByID(id uuid.UUID) (*model.Tenant, error) {
	var t model.Tenant
	err := r.db.Where("id = ?", id).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *tenantRepository) Create(t *model.Tenant) error {
	return r.db.Create(t).Error
}

func (r *tenantRepository) Update(t *model.Tenant) error {
	return r.db.Model(t).Updates(t).Error
}

func (r *tenantRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Tenant{}, id).Error
}
