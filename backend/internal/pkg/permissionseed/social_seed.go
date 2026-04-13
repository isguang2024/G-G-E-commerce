package permissionseed

import (
	"errors"

	"gorm.io/gorm"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
)

const GitHubProviderKey = "github"

// EnsureSocialAuthProviders 幂等初始化社交登录 Provider 种子。
// 默认仅写入 GitHub provider 且 enabled=false，避免误开通。
func EnsureSocialAuthProviders(db *gorm.DB) error {
	if db == nil {
		return errors.New("db is nil")
	}

	spec := systemmodels.SocialAuthProvider{
		ID:           StableID("social-provider", "default:"+GitHubProviderKey),
		TenantID:     "default",
		ProviderKey:  GitHubProviderKey,
		ProviderName: "GitHub",
		AuthURL:      "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		UserInfoURL:  "https://api.github.com/user",
		Scope:        "read:user user:email",
		Enabled:      false,
		Config:       systemmodels.MetaJSON{},
	}

	var existing systemmodels.SocialAuthProvider
	err := db.Where("tenant_id = ? AND provider_key = ?", spec.TenantID, spec.ProviderKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]any{
			"provider_name": spec.ProviderName,
			"auth_url":      spec.AuthURL,
			"token_url":     spec.TokenURL,
			"user_info_url": spec.UserInfoURL,
			"scope":         spec.Scope,
			"config":        spec.Config,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&spec).Error
	default:
		return err
	}
}
