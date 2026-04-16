package register

import (
	"context"
	"fmt"
	"strings"

	systemmodels "github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/pkg/permissionseed"
)

// Resolver 负责把 (host, path) 解析为 EffectiveRegisterContext。
type Resolver struct {
	repo *Repository
}

func NewResolver(repo *Repository) *Resolver { return &Resolver{repo: repo} }

// Resolve 解析注册上下文。算法：
// 1. host+path 命中 register_entries
// 2. 未命中则 fallback 到 entry_code=default
// 3. 直接从 entry 内联配置构建有效上下文（不再加载 policy）
func (r *Resolver) Resolve(ctx context.Context, host, path string) (*EffectiveRegisterContext, error) {
	entry, err := r.repo.FindEntryByHostPath(ctx, host, path)
	if err != nil {
		if !IsNotFound(err) {
			return nil, err
		}
		entry, err = r.repo.FindEntryByCode(ctx, permissionseed.DefaultRegisterEntryCode)
		if err != nil {
			return nil, fmt.Errorf("register: no entry matched and default missing: %w", err)
		}
	}
	loginPageKey := strings.TrimSpace(entry.LoginPageKey)
	if loginPageKey == "" {
		app, appErr := r.repo.FindAppByKey(ctx, entry.AppKey)
		if appErr == nil {
			loginPageKey = resolveAppLoginPageKey(app)
		}
	}
	if loginPageKey == "" {
		loginPageKey = permissionseed.DefaultLoginPageTemplateKey
	}
	return &EffectiveRegisterContext{
		EntryCode:                entry.EntryCode,
		EntryName:                entry.Name,
		EntryAppKey:              entry.AppKey,
		LoginPageKey:             loginPageKey,
		RegisterSource:           entry.RegisterSource,
		IsSystemReserved:         entry.IsSystemReserved,
		TargetURL:                entry.TargetURL,
		TargetAppKey:             entry.TargetAppKey,
		TargetNavigationSpaceKey: entry.TargetNavigationSpaceKey,
		TargetHomePath:           entry.TargetHomePath,
		AllowPublicRegister:      entry.AllowPublicRegister,
		RequireInvite:            entry.RequireInvite,
		RequireEmailVerify:       entry.RequireEmailVerify,
		RequireCaptcha:           entry.RequireCaptcha,
		AutoLogin:                entry.AutoLogin,
		CaptchaProvider:          entry.CaptchaProvider,
		CaptchaSiteKey:           entry.CaptchaSiteKey,
		RoleCodes:                []string(entry.RoleCodes),
		FeaturePackageKeys:       []string(entry.FeaturePackageKeys),
	}, nil
}

func (r *Resolver) ResolveLoginPageContext(
	ctx context.Context,
	input ResolveLoginPageContextInput,
) (*LoginPageContext, error) {
	const defaultRegisterPath = permissionseed.DefaultRegisterEntryPathPrefix
	host := strings.TrimSpace(input.Host)
	path := strings.TrimSpace(input.Path)
	targetAppKeyFromInput := strings.TrimSpace(input.TargetAppKey)
	loginPageKeyFromInput := strings.TrimSpace(input.LoginPageKey)
	pageScene := strings.TrimSpace(input.PageScene)
	if pageScene == "" {
		pageScene = "login"
	}
	appKey := permissionseed.AccountPortalAppKey
	loginPageKey := permissionseed.DefaultLoginPageTemplateKey
	loginUiMode := "auth_center_ui"
	ssoMode := "participate"
	resolvedBy := "default_template"
	targetAppKey := appKey
	entryCode := ""
	entryName := ""
	registerAppKey := appKey
	registerPath := defaultRegisterPath
	if path != "" {
		registerPath = path
	}

	if entry, err := r.repo.FindEntryByHostPath(ctx, host, registerPath); err == nil && entry != nil {
		entryCode = entry.EntryCode
		entryName = entry.Name
		registerAppKey = entry.AppKey
		targetAppKey = entry.AppKey
		if trimmed := strings.TrimSpace(entry.LoginPageKey); trimmed != "" {
			loginPageKey = trimmed
			resolvedBy = "register_entry"
		}
	}

	if targetAppKeyFromInput != "" {
		targetAppKey = targetAppKeyFromInput
		registerAppKey = targetAppKeyFromInput
		resolvedBy = "target_app"
	}
	if loginPageKeyFromInput != "" {
		loginPageKey = loginPageKeyFromInput
		resolvedBy = "query"
	}

	app, err := r.repo.FindAppByKey(ctx, registerAppKey)
	if err == nil && app != nil {
		appKey = app.AppKey
		if resolvedBy == "default_template" || resolvedBy == "target_app" {
			if trimmed := resolveAppLoginPageKey(app); trimmed != "" {
				loginPageKey = trimmed
				resolvedBy = "app_capability"
			}
		}
		if trimmed := resolveAppLoginUiMode(app); trimmed != "" {
			loginUiMode = trimmed
		}
		if trimmed := resolveAppSsoMode(app); trimmed != "" {
			ssoMode = trimmed
		}
	}

	var templateName string
	var templateConfig map[string]interface{}
	if template, tmplErr := r.repo.FindLoginPageTemplateByKey(ctx, loginPageKey); tmplErr == nil && template != nil {
		loginPageKey = template.TemplateKey
		templateName = template.Name
		templateConfig = resolveTemplateConfigByScene(map[string]interface{}(template.Config), pageScene)
	} else if template, tmplErr := r.repo.FindDefaultLoginPageTemplate(ctx); tmplErr == nil && template != nil {
		loginPageKey = template.TemplateKey
		templateName = template.Name
		templateConfig = resolveTemplateConfigByScene(map[string]interface{}(template.Config), pageScene)
		if resolvedBy == "default_template" {
			resolvedBy = "default_template_record"
		}
	}

	return &LoginPageContext{
		AppKey:         appKey,
		LoginPageKey:   loginPageKey,
		LoginUiMode:    loginUiMode,
		SsoMode:        ssoMode,
		ResolvedBy:     resolvedBy,
		PageScene:      pageScene,
		TargetAppKey:   targetAppKey,
		EntryCode:      entryCode,
		EntryName:      entryName,
		RegisterPath:   registerPath,
		RegisterAppKey: registerAppKey,
		TemplateName:   templateName,
		TemplateConfig: templateConfig,
	}, nil
}

func resolveTemplateConfigByScene(config map[string]interface{}, pageScene string) map[string]interface{} {
	if len(config) == 0 {
		return nil
	}
	base := cloneMap(config)
	delete(base, "pages")

	sceneKey := normalizeSceneKey(pageScene)
	if sceneKey == "" {
		return base
	}

	pages, ok := asMap(config["pages"])
	if !ok {
		return base
	}

	sceneConfig, ok := asMap(pages[sceneKey])
	if !ok && sceneKey == "forget_password" {
		sceneConfig, ok = asMap(pages["forgetPassword"])
	}
	if !ok {
		return base
	}
	mergeMap(base, sceneConfig)
	return base
}

func normalizeSceneKey(pageScene string) string {
	switch strings.TrimSpace(pageScene) {
	case "login":
		return "login"
	case "register":
		return "register"
	case "forget_password", "forgetPassword":
		return "forget_password"
	default:
		return ""
	}
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		if nested, ok := asMap(v); ok {
			dst[k] = cloneMap(nested)
			continue
		}
		dst[k] = v
	}
	return dst
}

func mergeMap(dst map[string]interface{}, src map[string]interface{}) {
	for k, v := range src {
		srcMap, srcIsMap := asMap(v)
		if !srcIsMap {
			dst[k] = v
			continue
		}
		if existing, ok := asMap(dst[k]); ok {
			mergeMap(existing, srcMap)
			dst[k] = existing
			continue
		}
		dst[k] = cloneMap(srcMap)
	}
}

func asMap(v interface{}) (map[string]interface{}, bool) {
	switch typed := v.(type) {
	case map[string]interface{}:
		return typed, true
	case systemmodels.MetaJSON:
		return map[string]interface{}(typed), true
	default:
		return nil, false
	}
}

func resolveAppLoginPageKey(app *systemmodels.App) string {
	if app == nil {
		return ""
	}
	authConfig := normalizeMetaSection(app.Capabilities, "auth")
	return strings.TrimSpace(readMetaString(authConfig, "login_page_key", "loginPageKey"))
}

func resolveAppLoginUiMode(app *systemmodels.App) string {
	if app == nil {
		return ""
	}
	authConfig := normalizeMetaSection(app.Capabilities, "auth")
	mode := strings.TrimSpace(readMetaString(authConfig, "login_ui_mode", "loginUiMode"))
	switch mode {
	case "auth_center_custom", "local_ui":
		return mode
	default:
		return "auth_center_ui"
	}
}

func resolveAppSsoMode(app *systemmodels.App) string {
	if app == nil {
		return ""
	}
	authConfig := normalizeMetaSection(app.Capabilities, "auth")
	mode := strings.TrimSpace(readMetaString(authConfig, "sso_mode", "ssoMode"))
	switch mode {
	case "reauth", "isolated":
		return mode
	default:
		return "participate"
	}
}

func normalizeMetaSection(meta systemmodels.MetaJSON, key string) systemmodels.MetaJSON {
	if len(meta) == 0 {
		return systemmodels.MetaJSON{}
	}
	raw, ok := meta[key]
	if !ok {
		return systemmodels.MetaJSON{}
	}
	switch typed := raw.(type) {
	case map[string]interface{}:
		return systemmodels.MetaJSON(typed)
	case systemmodels.MetaJSON:
		return typed
	default:
		return systemmodels.MetaJSON{}
	}
}

func readMetaString(meta systemmodels.MetaJSON, keys ...string) string {
	for _, key := range keys {
		if value, ok := meta[key]; ok {
			if text, ok := value.(string); ok {
				if trimmed := strings.TrimSpace(text); trimmed != "" {
					return trimmed
				}
			}
		}
	}
	return ""
}

