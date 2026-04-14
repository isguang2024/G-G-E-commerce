package register

import "strings"

// EffectiveRegisterContext 是注册入口命中后的有效上下文。
// 入口内联完整注册决策，不再从 policy 合并。
type EffectiveRegisterContext struct {
	EntryCode                string
	EntryName                string
	EntryAppKey              string
	LoginPageKey             string
	RegisterSource           string
	IsSystemReserved         bool

	// 注册后去向
	TargetURL                string
	TargetAppKey             string
	TargetNavigationSpaceKey string
	TargetHomePath           string

	// 注册规则
	AllowPublicRegister bool
	RequireInvite       bool
	RequireEmailVerify  bool
	RequireCaptcha      bool
	AutoLogin           bool
	AgreementVersion    string

	// 验证码
	CaptchaProvider string
	CaptchaSiteKey  string

	// 注册决策：绑定的角色 code 和功能包 key
	RoleCodes          []string
	FeaturePackageKeys []string
}

type LoginPageContext struct {
	AppKey         string
	LoginPageKey   string
	LoginUiMode    string
	SsoMode        string
	ResolvedBy     string
	PageScene      string
	TargetAppKey   string
	EntryCode      string
	EntryName      string
	RegisterPath   string
	RegisterAppKey string
	TemplateName   string
	TemplateConfig map[string]interface{}
}

type ResolveLoginPageContextInput struct {
	Host         string
	Path         string
	TargetAppKey string
	LoginPageKey string
	PageScene    string
}

// LandingInfo 注册/登录成功后前端应当跳转的目标。
// 优先级：URL > AppKey+HomePath(+NavigationSpaceKey) > 来源回源 > 全局默认入口。
type LandingInfo struct {
	URL                string `json:"url,omitempty"`
	AppKey             string `json:"app_key"`
	NavigationSpaceKey string `json:"navigation_space_key"`
	HomePath           string `json:"home_path"`
}

// PostAuthLandingInput 统一认证后 landing 解析的输入。
// 5 条认证链路（login, register-auto-login, register-pending, callback, social）
// 均使用此结构，确保跳转优先级一致。
type PostAuthLandingInput struct {
	// 入口级：来自 EffectiveRegisterContext 或 entry 配置
	EntryTargetURL                string
	EntryTargetAppKey             string
	EntryTargetNavigationSpaceKey string
	EntryTargetHomePath           string
	// 请求级：来自 login/callback 请求参数（来源 app）
	SourceAppKey             string
	SourceNavigationSpaceKey string
	SourceHomePath           string
}

// IsSafeRedirectURL 校验 URL 安全性，阻止 javascript:/data:/非 HTTP(S) 协议。
func IsSafeRedirectURL(rawURL string) bool {
	u := strings.TrimSpace(rawURL)
	if u == "" {
		return false
	}
	// 相对路径始终安全
	if strings.HasPrefix(u, "/") {
		return true
	}
	lower := strings.ToLower(u)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return true
	}
	return false
}

// ResolvePostAuthLanding 统一解析认证后跳转目标。
// 优先级链：
//  1. entry 显式 target_url（最高，直接外跳，需通过安全校验）
//  2. entry 显式 target_app_key + home_path + navigation_space_key
//  3. 请求来源 source_app_key + source_home_path + source_navigation_space_key
//  4. 空 landing（前端 fallback）
func ResolvePostAuthLanding(in PostAuthLandingInput) *LandingInfo {
	// 优先级 1: entry target_url（校验安全性）
	if in.EntryTargetURL != "" && IsSafeRedirectURL(in.EntryTargetURL) {
		return &LandingInfo{URL: in.EntryTargetURL}
	}
	// 优先级 2: entry target_app_key
	if in.EntryTargetAppKey != "" {
		return &LandingInfo{
			AppKey:             in.EntryTargetAppKey,
			NavigationSpaceKey: in.EntryTargetNavigationSpaceKey,
			HomePath:           in.EntryTargetHomePath,
		}
	}
	// 优先级 3: 请求来源
	if in.SourceAppKey != "" {
		return &LandingInfo{
			AppKey:             in.SourceAppKey,
			NavigationSpaceKey: in.SourceNavigationSpaceKey,
			HomePath:           in.SourceHomePath,
		}
	}
	// 优先级 4: 空 landing，前端决定
	return &LandingInfo{}
}
