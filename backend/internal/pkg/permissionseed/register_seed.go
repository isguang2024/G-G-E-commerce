package permissionseed

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	systemmodels "github.com/gg-ecommerce/backend/internal/modules/system/models"
)

// 注册体系常量。这些值会被 register/resolver 与 register/service 直接引用，
// 务必保持与 DB seed 一致。
const (
	AccountPortalAppKey            = "account-portal"
	AccountPortalDefaultSpaceKey   = "public"
	SelfServiceMenuSpaceKey        = "self-service"
	SelfServiceFeaturePackageKey   = "self_service.basic"
	SelfServiceRoleCode            = "personal.self_user"
	DefaultRegisterPolicyCode      = "default.self"
	DefaultRegisterEntryCode       = "default"
	DefaultRegisterEntryPathPrefix = "/account/auth/register"
	SelfServiceHomePath            = "/self/user-center"
	AccountPortalHomePath          = "/account/auth/login"
)

// EnsureRegisterSystemSeeds 写入注册体系第一期所需的全部默认数据：
// 1. account-portal App + Host/Path 绑定
// 2. account-portal/public 与 platform-admin/self-service MenuSpace
// 3. self_service.basic 功能包 + personal.self_user 角色 + 二者绑定
// 4. default.self 注册策略 + 默认 register_entries
//
// 整个函数幂等，按 unique key 查找后 upsert。
func EnsureRegisterSystemSeeds(db *gorm.DB) error {
	if db == nil {
		return errors.New("permissionseed: db is nil")
	}
	if err := ensureAccountPortalApp(db); err != nil {
		return err
	}
	if err := ensureRegisterMenuSpaces(db); err != nil {
		return err
	}
	if err := ensureRegisterAppHostBindings(db); err != nil {
		return err
	}
	pkgID, err := ensureSelfServiceFeaturePackage(db)
	if err != nil {
		return err
	}
	roleID, err := ensureSelfServiceRole(db)
	if err != nil {
		return err
	}
	if err := ensureSelfServiceRolePackage(db, roleID, pkgID); err != nil {
		return err
	}
	if err := ensureDefaultRegisterPolicy(db); err != nil {
		return err
	}
	if err := ensureDefaultRegisterPolicyBindings(db, pkgID, roleID); err != nil {
		return err
	}
	if err := ensureDefaultRegisterEntry(db); err != nil {
		return err
	}
	return nil
}

func ensureAccountPortalApp(db *gorm.DB) error {
	desired := systemmodels.App{
		AppKey:          AccountPortalAppKey,
		Name:            "注册中心",
		Description:     "公开注册 / 登录 / 邮箱验证 / 找回密码 / 邀请接受 入口承载 App",
		SpaceMode:       "single",
		DefaultSpaceKey: AccountPortalDefaultSpaceKey,
		AuthMode:        "inherit_host",
		Status:          "normal",
		IsDefault:       false,
		Meta:            systemmodels.MetaJSON{},
	}
	var existing systemmodels.App
	err := db.Where("app_key = ?", AccountPortalAppKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"name":              desired.Name,
			"description":       desired.Description,
			"space_mode":        desired.SpaceMode,
			"default_space_key": desired.DefaultSpaceKey,
			"auth_mode":         desired.AuthMode,
			"status":            desired.Status,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&desired).Error
	default:
		return err
	}
}

func ensureRegisterMenuSpaces(db *gorm.DB) error {
	specs := []systemmodels.MenuSpace{
		{
			AppKey:          AccountPortalAppKey,
			SpaceKey:        AccountPortalDefaultSpaceKey,
			Name:            "公开入口",
			Description:     "account-portal 公开页（注册 / 登录 / 找回密码 / 邀请接受）",
			DefaultHomePath: AccountPortalHomePath,
			IsDefault:       true,
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{},
		},
		{
			AppKey:          systemmodels.DefaultAppKey,
			SpaceKey:        SelfServiceMenuSpaceKey,
			Name:            "自助中心",
			Description:     "自注册用户登录后承载空间，与治理 default 空间隔离",
			DefaultHomePath: SelfServiceHomePath,
			IsDefault:       false,
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{},
		},
	}
	for i := range specs {
		spec := specs[i]
		var existing systemmodels.MenuSpace
		err := db.Where("app_key = ? AND space_key = ?", spec.AppKey, spec.SpaceKey).First(&existing).Error
		switch {
		case err == nil:
			if updateErr := db.Model(&existing).Updates(map[string]interface{}{
				"name":              spec.Name,
				"description":       spec.Description,
				"default_home_path": spec.DefaultHomePath,
				"status":            spec.Status,
			}).Error; updateErr != nil {
				return updateErr
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			if createErr := db.Create(&spec).Error; createErr != nil {
				return createErr
			}
		default:
			return err
		}
	}
	return nil
}

func ensureRegisterAppHostBindings(db *gorm.DB) error {
	// account-portal 路径绑定（本地 / 单域名部署）
	bindings := []systemmodels.AppHostBinding{
		{
			AppKey:          AccountPortalAppKey,
			MatchType:       systemmodels.EntryMatchPathPrefix,
			Host:            "",
			PathPattern:     "/account",
			Priority:        100,
			Description:     "account-portal 路径前缀（本地 / 单域名部署）",
			DefaultSpaceKey: AccountPortalDefaultSpaceKey,
			Status:          "normal",
			IsPrimary:       true,
			Meta:            systemmodels.MetaJSON{},
		},
		{
			AppKey:          AccountPortalAppKey,
			MatchType:       systemmodels.EntryMatchHostExact,
			Host:            "account.example.com",
			PathPattern:     "",
			Priority:        200,
			Description:     "account-portal 子域名示例（生产环境运维启用）",
			DefaultSpaceKey: AccountPortalDefaultSpaceKey,
			Status:          "disabled",
			Meta:            systemmodels.MetaJSON{},
		},
	}
	// platform-admin self-service 路径绑定（Level 2 MenuSpace 入口）
	spaceBindings := []systemmodels.MenuSpaceEntryBinding{
		{
			AppKey:      systemmodels.DefaultAppKey,
			SpaceKey:    SelfServiceMenuSpaceKey,
			MatchType:   systemmodels.EntryMatchPathPrefix,
			Host:        "",
			PathPattern: "/self",
			Priority:    100,
			Description: "platform-admin 自助空间路径前缀",
			Status:      "normal",
			Meta:        systemmodels.MetaJSON{},
		},
	}
	for i := range bindings {
		spec := bindings[i]
		var existing systemmodels.AppHostBinding
		err := db.Where("app_key = ? AND match_type = ? AND host = ? AND path_pattern = ?",
			spec.AppKey, spec.MatchType, spec.Host, spec.PathPattern).First(&existing).Error
		switch {
		case err == nil:
			continue
		case errors.Is(err, gorm.ErrRecordNotFound):
			if createErr := db.Create(&spec).Error; createErr != nil {
				return createErr
			}
		default:
			return err
		}
	}
	for i := range spaceBindings {
		spec := spaceBindings[i]
		var existing systemmodels.MenuSpaceEntryBinding
		err := db.Where("app_key = ? AND space_key = ? AND match_type = ? AND host = ? AND path_pattern = ?",
			spec.AppKey, spec.SpaceKey, spec.MatchType, spec.Host, spec.PathPattern).First(&existing).Error
		switch {
		case err == nil:
			continue
		case errors.Is(err, gorm.ErrRecordNotFound):
			if createErr := db.Create(&spec).Error; createErr != nil {
				return createErr
			}
		default:
			return err
		}
	}
	return nil
}

func ensureSelfServiceFeaturePackage(db *gorm.DB) (uuid.UUID, error) {
	id := StableID("feature-package", SelfServiceFeaturePackageKey)
	desired := systemmodels.FeaturePackage{
		ID:             id,
		AppKey:         systemmodels.DefaultAppKey,
		PackageKey:     SelfServiceFeaturePackageKey,
		PackageType:    "self_service",
		Name:           "自助用户基础包",
		Description:    "自注册用户的最小可用功能集（个人中心 / 收件箱 / 我的协作空间）",
		WorkspaceScope: "personal",
		ContextType:    "personal",
		IsBuiltin:      true,
		Status:         "normal",
		SortOrder:      100,
	}
	var existing systemmodels.FeaturePackage
	err := db.Where("package_key = ?", SelfServiceFeaturePackageKey).First(&existing).Error
	switch {
	case err == nil:
		if updateErr := db.Model(&existing).Updates(map[string]interface{}{
			"name":            desired.Name,
			"description":     desired.Description,
			"workspace_scope": desired.WorkspaceScope,
			"context_type":    desired.ContextType,
			"package_type":    desired.PackageType,
			"is_builtin":      desired.IsBuiltin,
			"status":          desired.Status,
			"sort_order":      desired.SortOrder,
			"app_key":         desired.AppKey,
		}).Error; updateErr != nil {
			return uuid.Nil, updateErr
		}
		return existing.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		if createErr := db.Create(&desired).Error; createErr != nil {
			return uuid.Nil, createErr
		}
		return desired.ID, nil
	default:
		return uuid.Nil, err
	}
}

func ensureSelfServiceRole(db *gorm.DB) (uuid.UUID, error) {
	desired := systemmodels.Role{
		ID:          StableID("role", SelfServiceRoleCode),
		Code:        SelfServiceRoleCode,
		Name:        "个人自助用户",
		Description: "公开注册用户默认角色，仅授予个人空间最小自助能力",
		Status:      "normal",
		IsSystem:    true,
		SortOrder:   100,
	}
	var existing systemmodels.Role
	err := db.Where("code = ? AND collaboration_workspace_id IS NULL", SelfServiceRoleCode).First(&existing).Error
	switch {
	case err == nil:
		if updateErr := db.Model(&existing).Updates(map[string]interface{}{
			"name":        desired.Name,
			"description": desired.Description,
			"status":      desired.Status,
			"is_system":   true,
		}).Error; updateErr != nil {
			return uuid.Nil, updateErr
		}
		return existing.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		if createErr := db.Create(&desired).Error; createErr != nil {
			return uuid.Nil, createErr
		}
		return desired.ID, nil
	default:
		return uuid.Nil, err
	}
}

func ensureSelfServiceRolePackage(db *gorm.DB, roleID, packageID uuid.UUID) error {
	if roleID == uuid.Nil || packageID == uuid.Nil {
		return nil
	}
	var existing systemmodels.RoleFeaturePackage
	err := db.Where("role_id = ? AND package_id = ?", roleID, packageID).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{"enabled": true}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&systemmodels.RoleFeaturePackage{
			ID:        StableID("role-feature-package", SelfServiceRoleCode+":"+SelfServiceFeaturePackageKey),
			RoleID:    roleID,
			PackageID: packageID,
			Enabled:   true,
		}).Error
	default:
		return err
	}
}

func ensureDefaultRegisterPolicy(db *gorm.DB) error {
	desired := systemmodels.RegisterPolicy{
		ID:                       StableID("register-policy", DefaultRegisterPolicyCode),
		AppKey:                   AccountPortalAppKey,
		PolicyCode:               DefaultRegisterPolicyCode,
		Name:                     "默认自注册策略",
		Description:              "公开注册默认策略：注册成功后进入 platform-admin/self-service 空间",
		TargetAppKey:             systemmodels.DefaultAppKey,
		TargetNavigationSpaceKey: SelfServiceMenuSpaceKey,
		TargetHomePath:           SelfServiceHomePath,
		DefaultWorkspaceType:     "personal",
		Status:                   "enabled",
		AllowPublicRegister:      false,
		RequireInvite:            false,
		RequireEmailVerify:       false,
		RequireCaptcha:           false,
		AutoLogin:                true,
	}
	var existing systemmodels.RegisterPolicy
	err := db.Where("policy_code = ?", DefaultRegisterPolicyCode).First(&existing).Error
	switch {
	case err == nil:
		// 不覆盖 allow_public_register / 各开关：尊重运维已修改的值
		return db.Model(&existing).Updates(map[string]interface{}{
			"app_key":                     desired.AppKey,
			"name":                        desired.Name,
			"description":                 desired.Description,
			"target_app_key":              desired.TargetAppKey,
			"target_navigation_space_key": desired.TargetNavigationSpaceKey,
			"target_home_path":            desired.TargetHomePath,
			"default_workspace_type":      desired.DefaultWorkspaceType,
			"status":                      desired.Status,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&desired).Error
	default:
		return err
	}
}

func ensureDefaultRegisterPolicyBindings(db *gorm.DB, packageID, roleID uuid.UUID) error {
	if packageID != uuid.Nil {
		var existing systemmodels.RegisterPolicyFeaturePackage
		err := db.Where("policy_code = ? AND package_id = ? AND workspace_scope = ?",
			DefaultRegisterPolicyCode, packageID, "personal").First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := db.Create(&systemmodels.RegisterPolicyFeaturePackage{
				PolicyCode:     DefaultRegisterPolicyCode,
				PackageID:      packageID,
				WorkspaceScope: "personal",
				SortOrder:      0,
			}).Error; createErr != nil {
				return createErr
			}
		} else if err != nil {
			return err
		}
	}
	if roleID != uuid.Nil {
		var existing systemmodels.RegisterPolicyRole
		err := db.Where("policy_code = ? AND role_id = ? AND workspace_scope = ?",
			DefaultRegisterPolicyCode, roleID, "personal").First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := db.Create(&systemmodels.RegisterPolicyRole{
				PolicyCode:     DefaultRegisterPolicyCode,
				RoleID:         roleID,
				WorkspaceScope: "personal",
				SortOrder:      0,
			}).Error; createErr != nil {
				return createErr
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func ensureDefaultRegisterEntry(db *gorm.DB) error {
	desired := systemmodels.RegisterEntry{
		ID:             StableID("register-entry", DefaultRegisterEntryCode),
		AppKey:         AccountPortalAppKey,
		EntryCode:      DefaultRegisterEntryCode,
		Name:           "默认公开注册入口",
		Host:           "",
		PathPrefix:     DefaultRegisterEntryPathPrefix,
		RegisterSource: "self",
		PolicyCode:     DefaultRegisterPolicyCode,
		Status:         "enabled",
		SortOrder:      0,
		Remark:         "兜底入口：当未命中其它 register_entries 时使用",
	}
	var existing systemmodels.RegisterEntry
	err := db.Where("entry_code = ?", DefaultRegisterEntryCode).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"app_key":         desired.AppKey,
			"name":            desired.Name,
			"path_prefix":     desired.PathPrefix,
			"register_source": desired.RegisterSource,
			"policy_code":     desired.PolicyCode,
			"status":          desired.Status,
			"sort_order":      desired.SortOrder,
			"remark":          desired.Remark,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&desired).Error
	default:
		return err
	}
}
