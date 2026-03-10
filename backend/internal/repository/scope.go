package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/model"
)

// ScopeRepository 作用域仓储
type ScopeRepository interface {
	List(offset, limit int, code, name string) ([]model.Scope, int64, error)
	GetByID(id uuid.UUID) (*model.Scope, error)
	GetByCode(code string) (*model.Scope, error)
	Create(scope *model.Scope) error
	Update(scope *model.Scope) error
	Delete(id uuid.UUID) error
	GetAll() ([]model.Scope, error)
}

type scopeRepository struct {
	db *gorm.DB
}

func NewScopeRepository(db *gorm.DB) ScopeRepository {
	return &scopeRepository{db: db}
}

func (r *scopeRepository) List(offset, limit int, code, name string) ([]model.Scope, int64, error) {
	var list []model.Scope
	query := r.db.Model(&model.Scope{})
	if code != "" {
		query = query.Where("code LIKE ?", "%"+code+"%")
	}
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
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

func (r *scopeRepository) GetByID(id uuid.UUID) (*model.Scope, error) {
	var scope model.Scope
	err := r.db.Where("id = ?", id).First(&scope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &scope, nil
}

func (r *scopeRepository) GetByCode(code string) (*model.Scope, error) {
	var scope model.Scope
	err := r.db.Where("code = ?", code).First(&scope).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, err
	}
	return &scope, nil
}

func (r *scopeRepository) Create(scope *model.Scope) error {
	return r.db.Create(scope).Error
}

func (r *scopeRepository) Update(scope *model.Scope) error {
	return r.db.Model(scope).Updates(scope).Error
}

func (r *scopeRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Scope{}, id).Error
}

func (r *scopeRepository) GetAll() ([]model.Scope, error) {
	var list []model.Scope
	err := r.db.Order("sort_order ASC, created_at ASC").Find(&list).Error
	return list, err
}
