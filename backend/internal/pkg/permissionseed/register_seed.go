package permissionseed

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	systemmodels "github.com/maben/backend/internal/modules/system/models"
)

// 注册体系常量。这些值会被 register/resolver 与 register/service 直接引用，
// 务必保持与 DB seed 一致。
const (
	AccountPortalAppKey              = "account-portal"
	AccountPortalDefaultMenuSpaceKey = "public"
	DemoAppKey                       = "demo-app"
	DemoAppDefaultMenuSpaceKey       = "demo"
	DemoAppHomePath                  = "/demo/lab"
	SelfServiceMenuSpaceKey          = "self-service"
	SelfServiceFeaturePackageKey     = "self_service.basic"
	SelfServiceRoleCode              = "personal.self_user"
	DefaultRegisterEntryCode         = "default"
	DefaultRegisterEntryPathPrefix   = "/account/auth/register"
	DefaultLoginPageTemplateKey      = "default"
	SelfServiceHomePath              = "/self/user-center"
	AccountPortalHomePath            = "/account/auth/login"
)

// EnsureRegisterSystemSeeds 写入注册体系第一期所需的全部默认数据：
// 1. account-portal App + Host/Path 绑定
// 2. account-portal/public 与 platform-admin/self-service 菜单空间
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
	if err := ensureDemoApp(db); err != nil {
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
	if err := ensureDefaultRegisterEntry(db); err != nil {
		return err
	}
	if err := ensureLoginPageTemplates(db); err != nil {
		return err
	}
	if err := ensureAccountPortalPublicPages(db); err != nil {
		return err
	}
	if err := ensureDemoAppPages(db); err != nil {
		return err
	}
	return nil
}

func ensureAccountPortalPublicPages(db *gorm.DB) error {
	specs := []systemmodels.UIPage{
		{
			AppKey:          AccountPortalAppKey,
			PageKey:         "account_portal.auth.login",
			Name:            "账号登录",
			RouteName:       "Login",
			RoutePath:       "/account/auth/login",
			Component:       "/account-portal/auth/login",
			PageType:        "standalone",
			VisibilityScope: "app",
			Source:          "manual",
			ModuleKey:       "auth",
			SortOrder:       10,
			AccessMode:      "public",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"authScene": "login"},
		},
		{
			AppKey:          AccountPortalAppKey,
			PageKey:         "account_portal.auth.register",
			Name:            "账号注册",
			RouteName:       "Register",
			RoutePath:       "/account/auth/register",
			Component:       "/account-portal/auth/register",
			PageType:        "standalone",
			VisibilityScope: "app",
			Source:          "manual",
			ModuleKey:       "auth",
			SortOrder:       11,
			AccessMode:      "public",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"authScene": "register"},
		},
		{
			AppKey:          AccountPortalAppKey,
			PageKey:         "account_portal.auth.forget_password",
			Name:            "找回密码",
			RouteName:       "ForgetPassword",
			RoutePath:       "/account/auth/forget-password",
			Component:       "/account-portal/auth/forget-password",
			PageType:        "standalone",
			VisibilityScope: "app",
			Source:          "manual",
			ModuleKey:       "auth",
			SortOrder:       12,
			AccessMode:      "public",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"authScene": "forget-password"},
		},
		{
			AppKey:          AccountPortalAppKey,
			PageKey:         "account_portal.auth.callback",
			Name:            "认证回调",
			RouteName:       "AuthCallback",
			RoutePath:       "/account/auth/callback",
			Component:       "/auth/callback",
			PageType:        "standalone",
			VisibilityScope: "app",
			Source:          "manual",
			ModuleKey:       "auth",
			SortOrder:       13,
			AccessMode:      "public",
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{"authScene": "callback"},
		},
	}
	for i := range specs {
		spec := specs[i]
		var existing systemmodels.UIPage
		err := db.Where("app_key = ? AND page_key = ?", spec.AppKey, spec.PageKey).First(&existing).Error
		switch {
		case err == nil:
			if updateErr := db.Model(&existing).Updates(map[string]interface{}{
				"name":             spec.Name,
				"route_name":       spec.RouteName,
				"route_path":       spec.RoutePath,
				"component":        spec.Component,
				"page_type":        spec.PageType,
				"visibility_scope": spec.VisibilityScope,
				"source":           spec.Source,
				"module_key":       spec.ModuleKey,
				"sort_order":       spec.SortOrder,
				"access_mode":      spec.AccessMode,
				"status":           spec.Status,
				"meta":             spec.Meta,
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

func ensureAccountPortalApp(db *gorm.DB) error {
	desired := systemmodels.App{
		AppKey:              AccountPortalAppKey,
		Name:                "认证中心",
		Description:         "公开注册 / 登录 / 邮箱验证 / 找回密码 / 邀请接受 入口承载 App",
		SpaceMode:           "single",
		DefaultMenuSpaceKey: AccountPortalDefaultMenuSpaceKey,
		AuthMode:            "inherit_host",
		FrontendEntryURL:    "/account/auth/login",
		BackendEntryURL:     "",
		HealthCheckURL:      "/health",
		Status:              "normal",
		IsDefault:           false,
		Capabilities:        systemmodels.DefaultAccountPortalCapabilities(),
		Meta:                systemmodels.MetaJSON{},
	}
	var existing systemmodels.App
	err := db.Where("app_key = ?", AccountPortalAppKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"name":                   desired.Name,
			"description":            desired.Description,
			"space_mode":             desired.SpaceMode,
			"default_menu_space_key": desired.DefaultMenuSpaceKey,
			"auth_mode":              desired.AuthMode,
			"frontend_entry_url":     desired.FrontendEntryURL,
			"backend_entry_url":      desired.BackendEntryURL,
			"health_check_url":       desired.HealthCheckURL,
			"capabilities":           desired.Capabilities,
			"status":                 desired.Status,
			"meta":                   desired.Meta,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&desired).Error
	default:
		return err
	}
}

func ensureDemoApp(db *gorm.DB) error {
	desired := systemmodels.App{
		AppKey:              DemoAppKey,
		Name:                "Demo App",
		Description:         "Phase A path_prefix 多 APP 验证应用",
		SpaceMode:           "single",
		DefaultMenuSpaceKey: DemoAppDefaultMenuSpaceKey,
		AuthMode:            "shared_cookie",
		FrontendEntryURL:    DemoAppHomePath,
		BackendEntryURL:     "",
		HealthCheckURL:      "/health",
		Status:              "normal",
		IsDefault:           false,
		Capabilities: systemmodels.MetaJSON{
			"auth": systemmodels.MetaJSON{
				"is_auth_center": false,
				"login_strategy": "shared_cookie",
			},
		},
		Meta: systemmodels.MetaJSON{},
	}
	var existing systemmodels.App
	err := db.Where("app_key = ?", DemoAppKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"name":                   desired.Name,
			"description":            desired.Description,
			"space_mode":             desired.SpaceMode,
			"default_menu_space_key": desired.DefaultMenuSpaceKey,
			"auth_mode":              desired.AuthMode,
			"frontend_entry_url":     desired.FrontendEntryURL,
			"backend_entry_url":      desired.BackendEntryURL,
			"health_check_url":       desired.HealthCheckURL,
			"capabilities":           desired.Capabilities,
			"status":                 desired.Status,
			"meta":                   desired.Meta,
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
			MenuSpaceKey:    AccountPortalDefaultMenuSpaceKey,
			Name:            "公开入口",
			Description:     "account-portal 公开页（注册 / 登录 / 找回密码 / 邀请接受）",
			DefaultHomePath: AccountPortalHomePath,
			IsDefault:       true,
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{},
		},
		{
			AppKey:          DemoAppKey,
			MenuSpaceKey:    DemoAppDefaultMenuSpaceKey,
			Name:            "Demo 空间",
			Description:     "Phase A path_prefix 多 APP 验证空间",
			DefaultHomePath: DemoAppHomePath,
			IsDefault:       true,
			Status:          "normal",
			Meta:            systemmodels.MetaJSON{},
		},
		{
			AppKey:          systemmodels.DefaultAppKey,
			MenuSpaceKey:    SelfServiceMenuSpaceKey,
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
		err := db.Where("app_key = ? AND menu_space_key = ?", spec.AppKey, spec.MenuSpaceKey).First(&existing).Error
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
			AppKey:              AccountPortalAppKey,
			MatchType:           systemmodels.EntryMatchPathPrefix,
			Host:                "",
			PathPattern:         "/account",
			Priority:            100,
			Description:         "account-portal 路径前缀（本地 / 单域名部署）",
			DefaultMenuSpaceKey: AccountPortalDefaultMenuSpaceKey,
			Status:              "normal",
			IsPrimary:           true,
			Meta:                systemmodels.MetaJSON{},
		},
		{
			AppKey:              AccountPortalAppKey,
			MatchType:           systemmodels.EntryMatchHostExact,
			Host:                "account.example.com",
			PathPattern:         "",
			Priority:            200,
			Description:         "account-portal 子域名示例（生产环境运维启用）",
			DefaultMenuSpaceKey: AccountPortalDefaultMenuSpaceKey,
			Status:              "disabled",
			Meta:                systemmodels.MetaJSON{},
		},
	}
	// platform-admin self-service 路径绑定（Level 2 菜单空间入口）
	spaceBindings := []systemmodels.MenuSpaceEntryBinding{
		{
			AppKey:       systemmodels.DefaultAppKey,
			MenuSpaceKey: SelfServiceMenuSpaceKey,
			MatchType:    systemmodels.EntryMatchPathPrefix,
			Host:         "",
			PathPattern:  "/self",
			Priority:     100,
			Description:  "platform-admin 自助空间路径前缀",
			Status:       "normal",
			Meta:         systemmodels.MetaJSON{},
		},
	}
	for i := range bindings {
		spec := bindings[i]
		var existing systemmodels.AppHostBinding
		err := db.Where("app_key = ? AND match_type = ? AND host = ? AND path_pattern = ?",
			spec.AppKey, spec.MatchType, spec.Host, spec.PathPattern).First(&existing).Error
		switch {
		case err == nil:
			if updateErr := db.Model(&existing).Updates(map[string]interface{}{
				"priority":               spec.Priority,
				"description":            spec.Description,
				"default_menu_space_key": spec.DefaultMenuSpaceKey,
				"status":                 spec.Status,
				"is_primary":             spec.IsPrimary,
				"meta":                   spec.Meta,
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
	for i := range spaceBindings {
		spec := spaceBindings[i]
		var existing systemmodels.MenuSpaceEntryBinding
		err := db.Where("app_key = ? AND menu_space_key = ? AND match_type = ? AND host = ? AND path_pattern = ?",
			spec.AppKey, spec.MenuSpaceKey, spec.MatchType, spec.Host, spec.PathPattern).First(&existing).Error
		switch {
		case err == nil:
			if updateErr := db.Model(&existing).Updates(map[string]interface{}{
				"priority":    spec.Priority,
				"description": spec.Description,
				"status":      spec.Status,
				"meta":        spec.Meta,
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
	err := db.Model(&systemmodels.Role{}).
		Where("code = ?", SelfServiceRoleCode).
		Where("NOT EXISTS (SELECT 1 FROM role_scopes rs WHERE rs.role_id = roles.id AND rs.deleted_at IS NULL AND rs.scope_type <> ?)", systemmodels.ScopeTypeGlobal).
		First(&existing).Error
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

func ensureDefaultRegisterEntry(db *gorm.DB) error {
	desired := systemmodels.RegisterEntry{
		ID:             StableID("register-entry", DefaultRegisterEntryCode),
		AppKey:         AccountPortalAppKey,
		EntryCode:      DefaultRegisterEntryCode,
		Name:           "默认公开注册入口",
		Description:    "系统保留兜底入口：当未命中其它 register_entries 时使用",
		Host:           "",
		PathPrefix:     DefaultRegisterEntryPathPrefix,
		RegisterSource: "self",
		LoginPageKey:   DefaultLoginPageTemplateKey,
		Status:         "enabled",
		// 注册规则（从 default.self 策略内联）
		AllowPublicRegister: false,
		AutoLogin:           true,
		// 注册后去向
		TargetAppKey:             systemmodels.DefaultAppKey,
		TargetNavigationSpaceKey: SelfServiceMenuSpaceKey,
		TargetHomePath:           SelfServiceHomePath,
		// 注册决策：绑定
		RoleCodes:          systemmodels.StringList{SelfServiceRoleCode},
		FeaturePackageKeys: systemmodels.StringList{SelfServiceFeaturePackageKey},
		// 系统保留
		IsSystemReserved: true,
		SortOrder:        0,
		Remark:           "兜底入口：当未命中其它 register_entries 时使用",
	}
	var existing systemmodels.RegisterEntry
	err := db.Where("entry_code = ?", DefaultRegisterEntryCode).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"app_key":                     desired.AppKey,
			"name":                        desired.Name,
			"description":                 desired.Description,
			"path_prefix":                 desired.PathPrefix,
			"register_source":             desired.RegisterSource,
			"login_page_key":              desired.LoginPageKey,
			"status":                      desired.Status,
			"target_app_key":              desired.TargetAppKey,
			"target_navigation_space_key": desired.TargetNavigationSpaceKey,
			"target_home_path":            desired.TargetHomePath,
			"role_codes":                  desired.RoleCodes,
			"feature_package_keys":        desired.FeaturePackageKeys,
			"is_system_reserved":          desired.IsSystemReserved,
			"sort_order":                  desired.SortOrder,
			"remark":                      desired.Remark,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&desired).Error
	default:
		return err
	}
}

func ensureLoginPageTemplates(db *gorm.DB) error {
	specs := []systemmodels.LoginPageTemplate{
		{
			ID:          StableID("login-page-template", DefaultLoginPageTemplateKey),
			TenantID:    "default",
			TemplateKey: DefaultLoginPageTemplateKey,
			Name:        "默认认证模板",
			Scene:       "auth_family",
			AppScope:    "shared",
			Status:      "normal",
			IsDefault:   true,
			Config: systemmodels.MetaJSON{
				"theme": systemmodels.MetaJSON{
					"primaryColor": "#409EFF",
					"borderRadius": "8px",
				},
				"features": systemmodels.MetaJSON{
					"socialLogin":    false,
					"rememberMe":     true,
					"forgetPassword": true,
					"register":       true,
				},
				"social": systemmodels.MetaJSON{
					"items": []any{
						systemmodels.MetaJSON{
							"key":  GitHubProviderKey,
							"name": "GitHub",
							"icon": "mdi:github",
							"url":  "/auth/oauth/github/authorize",
						},
					},
				},
				"pages": systemmodels.MetaJSON{
					"login": systemmodels.MetaJSON{
						"texts": systemmodels.MetaJSON{
							"title":      "欢迎回来",
							"subTitle":   "继续完成登录或注册流程",
							"buttonText": "登录",
						},
					},
					"register": systemmodels.MetaJSON{
						"texts": systemmodels.MetaJSON{
							"title":      "创建账号",
							"subTitle":   "继续完成注册流程",
							"buttonText": "注册",
						},
						"features": systemmodels.MetaJSON{
							"register": true,
						},
					},
					"forget_password": systemmodels.MetaJSON{
						"texts": systemmodels.MetaJSON{
							"title":               "找回密码",
							"subTitle":            "请输入账号后继续下一步重置流程",
							"buttonText":          "继续",
							"secondaryButtonText": "返回登录",
						},
						"features": systemmodels.MetaJSON{
							"forgetPassword": true,
						},
					},
				},
			},
			Meta: systemmodels.MetaJSON{
				"allowed_theme_variants": []string{"classic"},
			},
		},
		{
			ID:          StableID("login-page-template", "aurora"),
			TenantID:    "default",
			TemplateKey: "aurora",
			Name:        "Aurora 认证模板",
			Scene:       "auth_family",
			AppScope:    "shared",
			Status:      "normal",
			IsDefault:   false,
			Config: systemmodels.MetaJSON{
				"theme": systemmodels.MetaJSON{
					"primaryColor": "#7C4DFF",
					"borderRadius": "14px",
				},
				"features": systemmodels.MetaJSON{
					"socialLogin":    false,
					"rememberMe":     true,
					"forgetPassword": true,
					"register":       true,
				},
				"social": systemmodels.MetaJSON{
					"items": []any{
						systemmodels.MetaJSON{
							"key":  GitHubProviderKey,
							"name": "GitHub",
							"icon": "mdi:github",
							"url":  "/auth/oauth/github/authorize",
						},
					},
				},
				"pages": systemmodels.MetaJSON{
					"login": systemmodels.MetaJSON{
						"texts": systemmodels.MetaJSON{
							"title":      "欢迎回来",
							"subTitle":   "继续完成登录或注册流程",
							"buttonText": "登录",
						},
					},
					"register": systemmodels.MetaJSON{
						"texts": systemmodels.MetaJSON{
							"title":      "创建账号",
							"subTitle":   "继续完成注册流程",
							"buttonText": "注册",
						},
					},
					"forget_password": systemmodels.MetaJSON{
						"texts": systemmodels.MetaJSON{
							"title":               "找回密码",
							"subTitle":            "请输入账号后继续下一步重置流程",
							"buttonText":          "继续",
							"secondaryButtonText": "返回登录",
						},
					},
				},
			},
			Meta: systemmodels.MetaJSON{
				"allowed_theme_variants": []string{"aurora"},
			},
		},
	}

	for i := range specs {
		spec := specs[i]
		var existing systemmodels.LoginPageTemplate
		err := db.Where("tenant_id = ? AND template_key = ?", spec.TenantID, spec.TemplateKey).First(&existing).Error
		switch {
		case err == nil:
			if updateErr := db.Model(&existing).Updates(map[string]interface{}{
				"name":       spec.Name,
				"scene":      spec.Scene,
				"app_scope":  spec.AppScope,
				"status":     spec.Status,
				"is_default": spec.IsDefault,
				"config":     spec.Config,
				"meta":       spec.Meta,
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

func ensureDemoAppPages(db *gorm.DB) error {
	spec := systemmodels.UIPage{
		AppKey:          DemoAppKey,
		PageKey:         "demo_app.lab.index",
		Name:            "Demo 验证页",
		RouteName:       "DemoAppLab",
		RoutePath:       DemoAppHomePath,
		Component:       "/demo/lab",
		PageType:        "standalone",
		VisibilityScope: "app",
		Source:          "manual",
		ModuleKey:       "demo",
		SortOrder:       10,
		AccessMode:      "public",
		Status:          "normal",
		Meta: systemmodels.MetaJSON{
			"purpose": "phase-a-validation",
		},
	}

	var existing systemmodels.UIPage
	err := db.Where("app_key = ? AND page_key = ?", spec.AppKey, spec.PageKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"name":             spec.Name,
			"route_name":       spec.RouteName,
			"route_path":       spec.RoutePath,
			"component":        spec.Component,
			"page_type":        spec.PageType,
			"visibility_scope": spec.VisibilityScope,
			"source":           spec.Source,
			"module_key":       spec.ModuleKey,
			"sort_order":       spec.SortOrder,
			"access_mode":      spec.AccessMode,
			"status":           spec.Status,
			"meta":             spec.Meta,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&spec).Error
	default:
		return err
	}
}
