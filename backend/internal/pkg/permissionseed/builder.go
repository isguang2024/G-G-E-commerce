package permissionseed

import (
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type DeploymentBuilder struct {
	db                    *gorm.DB
	logger                *zap.Logger
	menus                 []MenuSeed
	apiEndpointCategories []APIEndpointCategorySeed
	permissionGroups      []PermissionGroupSeed
	permissionKeys        []PermissionKeySeed
	featurePackages       []FeaturePackageSeed
	featurePackageBundles []FeaturePackageBundleSeed
	rolePackageBindings   []RolePackageBindingSeed
}

func NewDeploymentBuilder(db *gorm.DB, logger *zap.Logger) *DeploymentBuilder {
	return &DeploymentBuilder{
		db:     db,
		logger: logger,
	}
}

func (b *DeploymentBuilder) WithCoreDefaults() *DeploymentBuilder {
	if b == nil {
		return nil
	}
	return b.
		WithDefaultMenus().
		WithDefaultAPIEndpointCategories().
		WithDefaultPermissionGroups().
		WithDefaultPermissionKeys().
		WithDefaultFeaturePackages().
		WithDefaultFeaturePackageBundles().
		WithDefaultRolePackageBindings()
}

func (b *DeploymentBuilder) WithDefaultAPIEndpointCategories() *DeploymentBuilder {
	b.apiEndpointCategories = append([]APIEndpointCategorySeed(nil), DefaultAPIEndpointCategories()...)
	return b
}

func (b *DeploymentBuilder) WithDefaultMenus() *DeploymentBuilder {
	b.menus = append([]MenuSeed(nil), DefaultMenus()...)
	return b
}

func (b *DeploymentBuilder) WithDefaultPermissionKeys() *DeploymentBuilder {
	b.permissionKeys = append([]PermissionKeySeed(nil), DefaultPermissionKeys()...)
	return b
}

func (b *DeploymentBuilder) WithDefaultPermissionGroups() *DeploymentBuilder {
	b.permissionGroups = append([]PermissionGroupSeed(nil), DefaultPermissionGroups()...)
	return b
}

func (b *DeploymentBuilder) WithDefaultFeaturePackages() *DeploymentBuilder {
	b.featurePackages = append([]FeaturePackageSeed(nil), DefaultFeaturePackages()...)
	return b
}

func (b *DeploymentBuilder) WithDefaultFeaturePackageBundles() *DeploymentBuilder {
	b.featurePackageBundles = append([]FeaturePackageBundleSeed(nil), DefaultFeaturePackageBundles()...)
	return b
}

func (b *DeploymentBuilder) WithDefaultRolePackageBindings() *DeploymentBuilder {
	b.rolePackageBindings = append([]RolePackageBindingSeed(nil), DefaultRolePackageBindings()...)
	return b
}

func (b *DeploymentBuilder) Summary() map[string]interface{} {
	return map[string]interface{}{
		"default_menu_count":             len(b.menus),
		"default_api_category_count":     len(b.apiEndpointCategories),
		"default_permission_group_count": len(b.permissionGroups),
		"default_permission_key_count":   len(b.permissionKeys),
		"default_feature_package_count":  len(b.featurePackages),
		"default_feature_bundle_count":   len(b.featurePackageBundles),
		"default_role_binding_count":     len(b.rolePackageBindings),
	}
}

func (b *DeploymentBuilder) LogSummary() {
	if b == nil || b.logger == nil {
		return
	}
	summary := b.Summary()
	b.logger.Info("Permission deployment summary",
		zap.Int("default_menu_count", summary["default_menu_count"].(int)),
		zap.Int("default_api_category_count", summary["default_api_category_count"].(int)),
		zap.Int("default_permission_group_count", summary["default_permission_group_count"].(int)),
		zap.Int("default_permission_key_count", summary["default_permission_key_count"].(int)),
		zap.Int("default_feature_package_count", summary["default_feature_package_count"].(int)),
		zap.Int("default_feature_bundle_count", summary["default_feature_bundle_count"].(int)),
		zap.Int("default_role_binding_count", summary["default_role_binding_count"].(int)),
	)
}

func NormalizeRouteModule(path string) string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return ""
	}
	segments := strings.Split(trimmed, "/")
	if len(segments) >= 3 {
		return segments[2]
	}
	return segments[len(segments)-1]
}
