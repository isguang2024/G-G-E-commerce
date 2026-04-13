package register

import (
	"context"
	"errors"

	"gorm.io/gorm"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// Repository 封装注册体系表的只读/写操作。
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

const defaultTenantID = "default"

// FindEntryByHostPath 按 host + path_prefix 命中首选入口：
// 1. 同时匹配 host 与最长 path_prefix
// 2. host 可留空表示任意 host
// 命中失败返回 gorm.ErrRecordNotFound。
func (r *Repository) FindEntryByHostPath(ctx context.Context, host, path string) (*systemmodels.RegisterEntry, error) {
	var list []systemmodels.RegisterEntry
	q := r.db.WithContext(ctx).
		Where("status = ?", "enabled").
		Where("(host = ? OR host = '')", host).
		Order("sort_order ASC, length(path_prefix) DESC")
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	for i := range list {
		e := list[i]
		if e.PathPrefix == "" || (len(path) >= len(e.PathPrefix) && path[:len(e.PathPrefix)] == e.PathPrefix) {
			return &e, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// FindEntryByCode 按 entry_code 精确查找。
func (r *Repository) FindEntryByCode(ctx context.Context, code string) (*systemmodels.RegisterEntry, error) {
	var entry systemmodels.RegisterEntry
	if err := r.db.WithContext(ctx).Where("entry_code = ?", code).First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// FindPolicyByCode 按 policy_code 精确查找。
func (r *Repository) FindPolicyByCode(ctx context.Context, code string) (*systemmodels.RegisterPolicy, error) {
	var p systemmodels.RegisterPolicy
	if err := r.db.WithContext(ctx).Where("policy_code = ?", code).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) FindAppByKey(ctx context.Context, appKey string) (*systemmodels.App, error) {
	var app systemmodels.App
	if err := r.db.WithContext(ctx).
		Where("app_key = ? AND deleted_at IS NULL", appKey).
		First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *Repository) FindLoginPageTemplateByKey(
	ctx context.Context,
	templateKey string,
) (*systemmodels.LoginPageTemplate, error) {
	var item systemmodels.LoginPageTemplate
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND template_key = ? AND status = ? AND deleted_at IS NULL", defaultTenantID, templateKey, "normal").
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) FindDefaultLoginPageTemplate(
	ctx context.Context,
) (*systemmodels.LoginPageTemplate, error) {
	var item systemmodels.LoginPageTemplate
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND scene = ? AND status = ? AND is_default = ? AND deleted_at IS NULL", defaultTenantID, "auth_family", "normal", true).
		Order("updated_at DESC").
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListPolicyFeaturePackages(ctx context.Context, policyCode string) ([]systemmodels.RegisterPolicyFeaturePackage, error) {
	var list []systemmodels.RegisterPolicyFeaturePackage
	if err := r.db.WithContext(ctx).Where("policy_code = ?", policyCode).Order("sort_order ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) ListPolicyRoles(ctx context.Context, policyCode string) ([]systemmodels.RegisterPolicyRole, error) {
	var list []systemmodels.RegisterPolicyRole
	if err := r.db.WithContext(ctx).Where("policy_code = ?", policyCode).Order("sort_order ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// IsNotFound helper.
func IsNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }
