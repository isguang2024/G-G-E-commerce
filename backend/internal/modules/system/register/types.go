package register

// EffectiveRegisterContext 是注册入口命中后与注册策略合并得到的有效上下文。
// 其中 *bool 字段若在 RegisterEntry 中非空则覆盖 policy 同名字段。
type EffectiveRegisterContext struct {
	EntryCode                string
	EntryName                string
	EntryAppKey              string
	LoginPageKey             string
	RegisterSource           string
	PolicyCode               string
	TargetAppKey             string
	TargetNavigationSpaceKey string
	TargetHomePath           string
	AllowPublicRegister      bool
	RequireInvite            bool
	RequireEmailVerify       bool
	RequireCaptcha           bool
	AutoLogin                bool
	AgreementVersion         string
	// 人机验证提供商及公开 site_key（require_captcha=true 时有效）
	CaptchaProvider string
	CaptchaSiteKey  string
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
type LandingInfo struct {
	AppKey             string `json:"app_key"`
	NavigationSpaceKey string `json:"navigation_space_key"`
	HomePath           string `json:"home_path"`
}
