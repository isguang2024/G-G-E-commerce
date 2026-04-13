package social

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const defaultTenantID = "default"

// Repository 社交登录配置与账号绑定仓储。
// 所有查询必须显式携带 tenant_id，避免跨租户串绑。
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func normalizeTenantID(tenantID string) string {
	if tenantID == "" {
		return defaultTenantID
	}
	return tenantID
}

func (r *Repository) FindProviderByKey(
	ctx context.Context,
	tenantID string,
	providerKey string,
) (*systemmodels.SocialAuthProvider, error) {
	var provider systemmodels.SocialAuthProvider
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider_key = ? AND enabled = ? AND deleted_at IS NULL", normalizeTenantID(tenantID), providerKey, true).
		First(&provider).Error
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *Repository) UpsertProvider(
	ctx context.Context,
	tenantID string,
	provider *systemmodels.SocialAuthProvider,
) error {
	if provider == nil {
		return errors.New("provider is nil")
	}
	provider.TenantID = normalizeTenantID(tenantID)
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "tenant_id"}, {Name: "provider_key"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"provider_name",
				"auth_url",
				"token_url",
				"user_info_url",
				"scope",
				"client_id",
				"client_secret",
				"redirect_uri",
				"enabled",
				"config",
				"updated_at",
				"deleted_at",
			}),
		}).
		Create(provider).Error
}

func (r *Repository) FindByProviderUID(
	ctx context.Context,
	tenantID string,
	providerKey string,
	providerUID string,
) (*systemmodels.UserSocialAccount, error) {
	var account systemmodels.UserSocialAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND provider_key = ? AND provider_uid = ? AND deleted_at IS NULL", normalizeTenantID(tenantID), providerKey, providerUID).
		First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *Repository) FindByUserID(
	ctx context.Context,
	tenantID string,
	userID uuid.UUID,
) ([]systemmodels.UserSocialAccount, error) {
	var list []systemmodels.UserSocialAccount
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", normalizeTenantID(tenantID), userID).
		Order("created_at DESC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *Repository) Create(
	ctx context.Context,
	account *systemmodels.UserSocialAccount,
) error {
	if account == nil {
		return errors.New("account is nil")
	}
	account.TenantID = normalizeTenantID(account.TenantID)
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *Repository) CreateInTx(
	tx *gorm.DB,
	account *systemmodels.UserSocialAccount,
) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	if account == nil {
		return errors.New("account is nil")
	}
	account.TenantID = normalizeTenantID(account.TenantID)
	return tx.Create(account).Error
}

func (r *Repository) UpdateLastLogin(
	ctx context.Context,
	tenantID string,
	id uuid.UUID,
) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&systemmodels.UserSocialAccount{}).
		Where("tenant_id = ? AND id = ? AND deleted_at IS NULL", normalizeTenantID(tenantID), id).
		Updates(map[string]any{
			"last_login_at": now,
			"updated_at":    now,
		}).Error
}

func (r *Repository) DeleteByID(
	ctx context.Context,
	tenantID string,
	id uuid.UUID,
	userID uuid.UUID,
) error {
	return r.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ? AND user_id = ? AND deleted_at IS NULL", normalizeTenantID(tenantID), id, userID).
		Delete(&systemmodels.UserSocialAccount{}).Error
}
