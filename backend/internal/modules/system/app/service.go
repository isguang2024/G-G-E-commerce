package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spacepkg "github.com/gg-ecommerce/backend/internal/modules/system/space"
	appctx "github.com/gg-ecommerce/backend/internal/pkg/appctx"
	"github.com/gg-ecommerce/backend/internal/pkg/pathmatch"
)

type AppRecord struct {
	models.App
	HostCount      int        `json:"host_count"`
	SpaceCount     int        `json:"space_count"`
	MenuCount      int        `json:"menu_count"`
	PageCount      int        `json:"page_count"`
	PrimaryHosts   []string   `json:"primary_hosts,omitempty"`
	ManifestURL    string     `json:"manifest_url"`
	RuntimeVersion string     `json:"runtime_version"`
	ProbeStatus    string     `json:"probe_status"`
	ProbeTarget    string     `json:"probe_target"`
	ProbeMessage   string     `json:"probe_message"`
	ProbeCheckedAt *time.Time `json:"probe_checked_at,omitempty"`
}

type HostBindingRecord struct {
	models.AppHostBinding
	AppName string `json:"app_name"`
}

type CurrentResponse struct {
	App         AppRecord          `json:"app"`
	Binding     *HostBindingRecord `json:"binding,omitempty"`
	ResolvedBy  string             `json:"resolved_by"`
	RequestHost string             `json:"request_host"`
}

type AppPreflightSummary struct {
	Level         string `json:"level"`
	BlockingCount int    `json:"blocking_count"`
	WarningCount  int    `json:"warning_count"`
	InfoCount     int    `json:"info_count"`
	SuccessCount  int    `json:"success_count"`
}

type AppPreflightCheckItem struct {
	Key    string `json:"key"`
	Title  string `json:"title"`
	Level  string `json:"level"`
	Passed bool   `json:"passed"`
	Value  string `json:"value"`
	Hint   string `json:"hint"`
}

type AppPreflightPreviewItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint"`
}

type AppPreflightResponse struct {
	AppKey         string                    `json:"app_key"`
	Name           string                    `json:"name"`
	RequestHost    string                    `json:"request_host"`
	ManifestURL    string                    `json:"manifest_url"`
	RuntimeVersion string                    `json:"runtime_version"`
	ProbeStatus    string                    `json:"probe_status"`
	ProbeTarget    string                    `json:"probe_target"`
	ProbeMessage   string                    `json:"probe_message"`
	ProbeCheckedAt *time.Time                `json:"probe_checked_at,omitempty"`
	Summary        AppPreflightSummary       `json:"summary"`
	Checks         []AppPreflightCheckItem   `json:"checks"`
	PreviewItems   []AppPreflightPreviewItem `json:"preview_items"`
}

type SaveAppRequest struct {
	AppKey           string                 `json:"app_key"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	SpaceMode        string                 `json:"space_mode"`
	DefaultSpaceKey  string                 `json:"default_space_key"`
	AuthMode         string                 `json:"auth_mode"`
	FrontendEntryURL string                 `json:"frontend_entry_url"`
	BackendEntryURL  string                 `json:"backend_entry_url"`
	HealthCheckURL   string                 `json:"health_check_url"`
	ManifestURL      string                 `json:"manifest_url"`
	RuntimeVersion   string                 `json:"runtime_version"`
	Capabilities     map[string]interface{} `json:"capabilities"`
	Status           string                 `json:"status"`
	IsDefault        bool                   `json:"is_default"`
	Meta             map[string]interface{} `json:"meta"`
}

type SaveHostBindingRequest struct {
	ID              string                 `json:"id"`
	AppKey          string                 `json:"app_key"`
	MatchType       string                 `json:"match_type"`
	Host            string                 `json:"host"`
	PathPattern     string                 `json:"path_pattern"`
	Priority        int                    `json:"priority"`
	Description     string                 `json:"description"`
	IsPrimary       bool                   `json:"is_primary"`
	DefaultSpaceKey string                 `json:"default_space_key"`
	Status          string                 `json:"status"`
	Meta            map[string]interface{} `json:"meta"`
}

type MenuSpaceEntryBindingRecord struct {
	models.MenuSpaceEntryBinding
	AppName   string `json:"app_name"`
	SpaceName string `json:"space_name"`
}

type SaveMenuSpaceEntryBindingRequest struct {
	ID          string                 `json:"id"`
	AppKey      string                 `json:"app_key"`
	SpaceKey    string                 `json:"space_key"`
	MatchType   string                 `json:"match_type"`
	Host        string                 `json:"host"`
	PathPattern string                 `json:"path_pattern"`
	Priority    int                    `json:"priority"`
	Description string                 `json:"description"`
	IsPrimary   bool                   `json:"is_primary"`
	Status      string                 `json:"status"`
	Meta        map[string]interface{} `json:"meta"`
}

type Service interface {
	ListApps() ([]AppRecord, error)
	GetCurrent(host, requestedAppKey string) (*CurrentResponse, error)
	GetAppPreflight(appKey, requestHost string) (*AppPreflightResponse, error)
	SaveApp(req *SaveAppRequest) (*AppRecord, error)
	DeleteApp(appKey string) error
	ListHostBindings(appKey string) ([]HostBindingRecord, error)
	SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error)
	DeleteHostBinding(appKey, id string) error
	ListMenuSpaceEntryBindings(appKey string) ([]MenuSpaceEntryBindingRecord, error)
	SaveMenuSpaceEntryBinding(appKey string, req *SaveMenuSpaceEntryBindingRequest) (*MenuSpaceEntryBindingRecord, error)
	DeleteMenuSpaceEntryBinding(appKey, id string) error
}

type service struct {
	db *gorm.DB
}

var appProbeHTTPClient = &http.Client{Timeout: 1500 * time.Millisecond}

func NewService(db *gorm.DB) Service {
	return &service{db: db}
}

func NormalizeAppKey(value string) string {
	return appctx.NormalizeAppKey(value)
}

func normalizeAppSpaceMode(value string) string {
	switch strings.TrimSpace(value) {
	case "multiple", "multi":
		return "multi"
	default:
		return "single"
	}
}

func RequestAppKey(c *gin.Context) string {
	return appctx.RequestAppKey(c)
}

func CurrentAppKey(c *gin.Context) string {
	return appctx.CurrentAppKey(c)
}

// ResolveAppByHost 兼容旧签名（仅 host），等价于 ResolveAppEntry(db, host, "", requestedAppKey)。
func ResolveAppByHost(db *gorm.DB, host string, requestedAppKey string) (string, *models.AppHostBinding, string, error) {
	return ResolveAppEntry(db, host, "", requestedAppKey)
}

// ResolveAppEntry Level 1 入口解析：按 host + path 匹配 APP。
func ResolveAppEntry(db *gorm.DB, host, path, requestedAppKey string) (string, *models.AppHostBinding, string, error) {
	if db == nil {
		return models.DefaultAppKey, nil, "fallback_default", nil
	}
	if err := ensureDefaultApp(db); err != nil {
		return models.DefaultAppKey, nil, "", err
	}

	explicit := NormalizeAppKey(requestedAppKey)
	if explicit != "" {
		ok, err := appExists(db, explicit)
		if err != nil {
			return models.DefaultAppKey, nil, "", err
		}
		if ok {
			return explicit, nil, "explicit", nil
		}
	}

	normalizedHost := pathmatch.NormalizeHost(host)
	normalizedPath := pathmatch.NormalizePath(path)

	// 加载所有启用绑定，按具体度排序。
	var bindings []models.AppHostBinding
	if err := db.Where("status = ? AND deleted_at IS NULL", "normal").Find(&bindings).Error; err != nil {
		return models.DefaultAppKey, nil, "", err
	}
	matched := matchAppEntryBinding(bindings, normalizedHost, normalizedPath)
	if matched != nil {
		ok, appErr := appExists(db, matched.AppKey)
		if appErr != nil {
			return models.DefaultAppKey, nil, "", appErr
		}
		if ok {
			return NormalizeAppKey(matched.AppKey), matched, "entry_binding", nil
		}
	}

	defaultApp, err := loadDefaultApp(db)
	if err != nil {
		return models.DefaultAppKey, nil, "", err
	}
	return defaultApp.AppKey, nil, "default_app", nil
}

// ResolveMenuSpaceEntry Level 2 入口解析：按 host + path 在 App 内匹配菜单空间。
// 单空间 App 直接返回 App 默认空间，不做匹配。
func ResolveMenuSpaceEntry(db *gorm.DB, appKey, host, path string) (string, string, error) {
	if db == nil {
		return "", "fallback_default", nil
	}
	normalizedAppKey := NormalizeAppKey(appKey)
	if normalizedAppKey == "" {
		return "", "fallback_default", nil
	}
	var app models.App
	if err := db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "fallback_default", nil
		}
		return "", "", err
	}
	defaultSpace := strings.TrimSpace(app.DefaultSpaceKey)
	if app.SpaceMode != "multi" {
		// 单空间短路。
		return defaultSpace, "single_space_app", nil
	}

	normalizedHost := pathmatch.NormalizeHost(host)
	normalizedPath := pathmatch.NormalizePath(path)

	var bindings []models.MenuSpaceEntryBinding
	if err := db.Where("app_key = ? AND status = ? AND deleted_at IS NULL", normalizedAppKey, "normal").Find(&bindings).Error; err != nil {
		return defaultSpace, "", err
	}
	matched := matchMenuSpaceEntryBinding(bindings, normalizedHost, normalizedPath)
	if matched != nil {
		spaceKey := strings.TrimSpace(matched.SpaceKey)
		// P2: 校验目标空间存在且未被禁用，避免遗留绑定指向失效空间。
		var space models.MenuSpace
		err := db.Select("space_key", "status").
			Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", normalizedAppKey, spaceKey).
			First(&space).Error
		if err == nil && space.Status != "disabled" {
			return spaceKey, "entry_binding", nil
		}
		// 空间不存在或已禁用，回退默认空间。
	}
	return defaultSpace, "default_space", nil
}

// matchAppEntryBinding 在所有候选 binding 中按具体度排序，返回第一个命中的。
func matchAppEntryBinding(bindings []models.AppHostBinding, host, path string) *models.AppHostBinding {
	type scored struct {
		idx   int
		score int
	}
	candidates := make([]scored, 0, len(bindings))
	for i := range bindings {
		b := bindings[i]
		hostPattern := pathmatch.NormalizeHostPattern(b.MatchType, b.Host)
		if !pathmatch.MatchHost(b.MatchType, hostPattern, host) {
			continue
		}
		needPath := b.MatchType == pathmatch.PathPrefix || b.MatchType == pathmatch.HostAndPath
		if needPath && !pathmatch.MatchPath(b.PathPattern, path) {
			continue
		}
		s := pathmatch.PatternSpecificity(b.MatchType, hostPattern, b.PathPattern) + b.Priority*10
		candidates = append(candidates, scored{idx: i, score: s})
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	winner := bindings[candidates[0].idx]
	return &winner
}

func matchMenuSpaceEntryBinding(bindings []models.MenuSpaceEntryBinding, host, path string) *models.MenuSpaceEntryBinding {
	type scored struct {
		idx   int
		score int
	}
	candidates := make([]scored, 0, len(bindings))
	for i := range bindings {
		b := bindings[i]
		hostPattern := pathmatch.NormalizeHostPattern(b.MatchType, b.Host)
		if !pathmatch.MatchHost(b.MatchType, hostPattern, host) {
			continue
		}
		needPath := b.MatchType == pathmatch.PathPrefix || b.MatchType == pathmatch.HostAndPath
		if needPath && !pathmatch.MatchPath(b.PathPattern, path) {
			continue
		}
		s := pathmatch.PatternSpecificity(b.MatchType, hostPattern, b.PathPattern) + b.Priority*10
		candidates = append(candidates, scored{idx: i, score: s})
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
	winner := bindings[candidates[0].idx]
	return &winner
}

func (s *service) ListApps() ([]AppRecord, error) {
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	var apps []models.App
	if err := s.db.Where("deleted_at IS NULL").Order("is_default DESC, created_at ASC").Find(&apps).Error; err != nil {
		return nil, err
	}

	hostCounts, primaryHosts, err := loadStringCountAndHosts(s.db, &models.AppHostBinding{}, "app_key", "host", "is_primary = true")
	if err != nil {
		return nil, err
	}
	spaceCounts, err := loadStringCounts(s.db, &models.MenuSpace{}, "app_key")
	if err != nil {
		return nil, err
	}
	menuCounts, err := loadStringCounts(s.db, &models.MenuDefinition{}, "app_key")
	if err != nil {
		return nil, err
	}
	pageCounts, err := loadStringCounts(s.db, &models.UIPage{}, "app_key")
	if err != nil {
		return nil, err
	}

	records := make([]AppRecord, 0, len(apps))
	for _, item := range apps {
		key := NormalizeAppKey(item.AppKey)
		probe := probeAppHealth(item, primaryHosts[key])
		records = append(records, AppRecord{
			App:            item,
			HostCount:      hostCounts[key],
			SpaceCount:     spaceCounts[key],
			MenuCount:      menuCounts[key],
			PageCount:      pageCounts[key],
			PrimaryHosts:   primaryHosts[key],
			ManifestURL:    appManifestURL(item.Meta),
			RuntimeVersion: appRuntimeVersion(item.Meta),
			ProbeStatus:    probe.Status,
			ProbeTarget:    probe.Target,
			ProbeMessage:   probe.Message,
			ProbeCheckedAt: probe.CheckedAt,
		})
	}
	return records, nil
}

func (s *service) GetCurrent(host, requestedAppKey string) (*CurrentResponse, error) {
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	appKey, binding, resolvedBy, err := ResolveAppByHost(s.db, host, requestedAppKey)
	if err != nil {
		return nil, err
	}
	record, err := s.getAppRecord(appKey)
	if err != nil {
		return nil, err
	}
	var hostBindingRecord *HostBindingRecord
	if binding != nil {
		hostBindingRecord = &HostBindingRecord{
			AppHostBinding: *binding,
			AppName:        record.Name,
		}
	}
	return &CurrentResponse{
		App:         *record,
		Binding:     hostBindingRecord,
		ResolvedBy:  resolvedBy,
		RequestHost: spacepkg.NormalizeHost(host),
	}, nil
}

func (s *service) GetAppPreflight(appKey, requestHost string) (*AppPreflightResponse, error) {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	record, err := s.getAppRecord(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	hostBindings, err := s.ListHostBindings(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	spaceEntryBindings := make([]MenuSpaceEntryBindingRecord, 0)
	if record.SpaceMode == "multi" {
		spaceEntryBindings, err = s.ListMenuSpaceEntryBindings(normalizedAppKey)
		if err != nil {
			return nil, err
		}
	}

	primaryBinding := pickPrimaryHostBinding(hostBindings)
	allowedRedirectHosts := collectAllowedRedirectHosts(hostBindings)
	checks := make([]AppPreflightCheckItem, 0, 8)
	checks = append(checks,
		newPreflightCheck(
			"binding",
			"入口绑定",
			primaryBinding != nil,
			"warning",
			describePrimaryBinding(primaryBinding),
			func() string {
				if primaryBinding != nil {
					return fmt.Sprintf("当前共 %d 条 APP 入口规则，主入口可作为接入命中基准。", len(hostBindings))
				}
				return "缺少主入口绑定，治理台只能依赖默认 App 回退。"
			}(),
		),
		newPreflightCheck(
			"frontend_entry",
			"前端入口",
			strings.TrimSpace(record.FrontendEntryURL) != "",
			"warning",
			firstNonEmpty(strings.TrimSpace(record.FrontendEntryURL), "继承当前地址"),
			func() string {
				if strings.TrimSpace(record.FrontendEntryURL) != "" {
					return "前端入口已声明，可作为登录后或切换后的首跳落点。"
				}
				return "未声明前端入口，仍会依赖当前地址或守卫逻辑推断。"
			}(),
		),
		newPreflightCheck(
			"backend_entry",
			"后端入口",
			strings.TrimSpace(record.BackendEntryURL) != "",
			"info",
			firstNonEmpty(strings.TrimSpace(record.BackendEntryURL), "沿用主站 API"),
			func() string {
				if strings.TrimSpace(record.BackendEntryURL) != "" {
					return "后端入口已声明，适合独立网关或跨域 API 场景。"
				}
				return "未声明后端入口，表示当前仍与主站共用 API 网关。"
			}(),
		),
		newPreflightCheck(
			"manifest",
			"远端清单",
			strings.TrimSpace(record.ManifestURL) != "",
			"info",
			firstNonEmpty(strings.TrimSpace(record.ManifestURL), "未配置"),
			func() string {
				if strings.TrimSpace(record.ManifestURL) != "" {
					return "已声明 manifest 地址，治理后台可直接展示远端接入清单来源。"
				}
				return "未声明 manifest 地址时，远端页来源仍只能回落到页面 meta 推断。"
			}(),
		),
		newPreflightCheck(
			"runtime_version",
			"运行版本",
			strings.TrimSpace(record.RuntimeVersion) != "",
			"info",
			firstNonEmpty(strings.TrimSpace(record.RuntimeVersion), "未声明"),
			func() string {
				if strings.TrimSpace(record.RuntimeVersion) != "" {
					return "已声明运行版本，可作为治理比对与 smoke 输出的版本锚点。"
				}
				return "未声明运行版本时，治理后台无法直接比对远端接入版本。"
			}(),
		),
		newPreflightCheck(
			"health_check",
			"运行探针",
			strings.TrimSpace(record.HealthCheckURL) != "",
			"info",
			firstNonEmpty(strings.TrimSpace(record.HealthCheckURL), "未配置"),
			func() string {
				if strings.TrimSpace(record.HealthCheckURL) != "" {
					return "已声明健康检查地址，可供后续聚合探测和治理观测使用。"
				}
				return "未声明健康检查地址，后续统一探测只能显示为未配置。"
			}(),
		),
		newPreflightCheck(
			"capabilities",
			"能力声明",
			len(record.Capabilities) > 0,
			"info",
			fmt.Sprintf("%d 个一级分组", len(record.Capabilities)),
			func() string {
				if len(record.Capabilities) > 0 {
					return "已声明 runtime/navigation/integration 等能力，可继续下沉到运行时消费。"
				}
				return "尚未声明能力分组，后续特性开关和壳层行为仍需靠约定。"
			}(),
		),
	)

	authMode := strings.TrimSpace(record.AuthMode)
	if authMode == "" {
		authMode = "inherit_host"
	}
	authHint := fmt.Sprintf("当前认证模式为 %s。", authMode)
	authValue := authMode
	authPassed := true
	authLevel := "success"
	if authMode == "centralized_login" {
		authValue = firstNonEmpty(strings.Join(allowedRedirectHosts, ", "), "未登记 callback host")
		authPassed = len(allowedRedirectHosts) > 0
		authLevel = "warning"
		if authPassed {
			authHint = "centralized_login 已具备可解释的回调白名单来源，回跳域名不再只能靠前端约定。"
		} else {
			authHint = "centralized_login 缺少可用于 redirect_uri 校验的 host，回调链路存在拒绝风险。"
		}
	} else if authMode == "shared_cookie" {
		authHint = "shared_cookie 仍需补真实登录/刷新/登出主链，本次只检查认证中心配置完整度。"
		authLevel = "info"
	}
	checks = append(checks, newPreflightCheck("auth_mode", "认证预检查", authPassed, authLevel, authValue, authHint))

	if record.SpaceMode == "multi" {
		checks = append(checks, newPreflightCheck(
			"space_entries",
			"菜单空间入口",
			len(spaceEntryBindings) > 0,
			"info",
			fmt.Sprintf("%d 条 Level 2 规则", len(spaceEntryBindings)),
			func() string {
				if len(spaceEntryBindings) > 0 {
					return "多空间 APP 已配置菜单空间入口，可进一步校验是否落在 APP 入口规则范围内。"
				}
				return "未配置菜单空间入口时，将统一回退到 APP 默认空间。"
			}(),
		))
	}

	previews := []AppPreflightPreviewItem{
		{
			Label: "诊断标签",
			Value: fmt.Sprintf("app=%s · auth=%s · probe=%s", record.AppKey, authMode, record.ProbeStatus),
			Hint:  fmt.Sprintf("结合 request_host=%s 与空间入口规则，可快速定位当前是入口解析问题、认证策略问题还是探针未配置问题。", spacepkg.NormalizeHost(requestHost)),
		},
		{
			Label: "入口命中",
			Value: describePrimaryBinding(primaryBinding),
			Hint: func() string {
				if primaryBinding != nil {
					return fmt.Sprintf("按 %s 规则进入 %s。", primaryBinding.MatchType, record.AppKey)
				}
				return "当前没有 APP 入口规则，只能依赖默认 App 回退。"
			}(),
		},
		{
			Label: "Manifest",
			Value: firstNonEmpty(strings.TrimSpace(record.ManifestURL), "未配置"),
			Hint: func() string {
				if strings.TrimSpace(record.ManifestURL) != "" {
					return "远端页面与运行版本应优先以该 manifest 为治理真相源。"
				}
				return "未配置 manifest 时，远端页治理仍只能依赖页面级 remote contract。"
			}(),
		},
		{
			Label: "首跳落点",
			Value: firstNonEmpty(strings.TrimSpace(record.FrontendEntryURL), "继承当前地址"),
			Hint: func() string {
				if authMode == "centralized_login" {
					return "登录前先进入认证中心，回调交换 token 后再跳回这里。"
				}
				return "登录后将以这个入口或当前 URL 作为首跳落点。"
			}(),
		},
		{
			Label: "健康探针",
			Value: func() string {
				if strings.TrimSpace(record.ProbeStatus) == "" || record.ProbeStatus == "missing" {
					return firstNonEmpty(strings.TrimSpace(record.HealthCheckURL), "未配置")
				}
				if strings.TrimSpace(record.ProbeTarget) != "" {
					return fmt.Sprintf("%s · %s", record.ProbeStatus, record.ProbeTarget)
				}
				return record.ProbeStatus
			}(),
			Hint: func() string {
				if strings.TrimSpace(record.ProbeMessage) != "" {
					return record.ProbeMessage
				}
				if strings.TrimSpace(record.HealthCheckURL) != "" {
					return "该地址可作为后续治理聚合、观测与 smoke probe 的落点。"
				}
				return "未配置探针地址时，治理后台无法展示统一运行状态。"
			}(),
		},
	}
	if authMode == "centralized_login" {
		previews = append(previews, AppPreflightPreviewItem{
			Label: "认证回跳",
			Value: firstNonEmpty(strings.Join(allowedRedirectHosts, ", "), "未登记 callback host"),
			Hint:  "redirect_uri 校验当前复用 APP host binding 与 callback_host 元数据。",
		})
	}

	return &AppPreflightResponse{
		AppKey:         record.AppKey,
		Name:           record.Name,
		RequestHost:    spacepkg.NormalizeHost(requestHost),
		ManifestURL:    record.ManifestURL,
		RuntimeVersion: record.RuntimeVersion,
		ProbeStatus:    record.ProbeStatus,
		ProbeTarget:    record.ProbeTarget,
		ProbeMessage:   record.ProbeMessage,
		ProbeCheckedAt: record.ProbeCheckedAt,
		Summary:        buildAppPreflightSummary(checks),
		Checks:         checks,
		PreviewItems:   previews,
	}, nil
}

func (s *service) SaveApp(req *SaveAppRequest) (*AppRecord, error) {
	if req == nil {
		return nil, errors.New("应用参数不能为空")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	appKey := NormalizeAppKey(req.AppKey)
	if appKey == "" {
		return nil, errors.New("应用标识不能为空")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("应用名称不能为空")
	}
	defaultSpaceKey := spacepkg.NormalizeSpaceKey(req.DefaultSpaceKey)
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	authMode := strings.TrimSpace(req.AuthMode)
	if authMode == "" {
		authMode = "inherit_host"
	}
	normalizedMeta, err := normalizeGovernanceMeta(req.Meta)
	if err != nil {
		return nil, err
	}
	applyAppRemoteContract(normalizedMeta, req.ManifestURL, req.RuntimeVersion)

	payload := models.App{
		AppKey:           appKey,
		Name:             name,
		Description:      strings.TrimSpace(req.Description),
		SpaceMode:        normalizeAppSpaceMode(req.SpaceMode),
		DefaultSpaceKey:  defaultSpaceKey,
		AuthMode:         authMode,
		FrontendEntryURL: strings.TrimSpace(req.FrontendEntryURL),
		BackendEntryURL:  strings.TrimSpace(req.BackendEntryURL),
		HealthCheckURL:   strings.TrimSpace(req.HealthCheckURL),
		Capabilities:     normalizeMetaJSON(req.Capabilities),
		Status:           status,
		IsDefault:        req.IsDefault || appKey == models.DefaultAppKey,
		Meta:             normalizedMeta,
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		var existing models.App
		err := tx.Where("app_key = ? AND deleted_at IS NULL", appKey).First(&existing).Error
		switch {
		case err == nil:
			if payload.DefaultSpaceKey == "" {
				payload.DefaultSpaceKey = spacepkg.NormalizeSpaceKey(existing.DefaultSpaceKey)
			}
			if payload.DefaultSpaceKey == "" {
				payload.DefaultSpaceKey = models.DefaultMenuSpaceKey
			}
			if err := ensureSpaceExistsForApp(tx, appKey, payload.DefaultSpaceKey); err != nil {
				return err
			}
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"name":               payload.Name,
				"description":        payload.Description,
				"space_mode":         payload.SpaceMode,
				"default_space_key":  payload.DefaultSpaceKey,
				"auth_mode":          payload.AuthMode,
				"frontend_entry_url": payload.FrontendEntryURL,
				"backend_entry_url":  payload.BackendEntryURL,
				"health_check_url":   payload.HealthCheckURL,
				"capabilities":       payload.Capabilities,
				"status":             payload.Status,
				"is_default":         payload.IsDefault,
				"meta":               payload.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			payload.DefaultSpaceKey = models.DefaultMenuSpaceKey
			if createErr := tx.Create(&payload).Error; createErr != nil {
				return createErr
			}
		default:
			return err
		}
		if err := spacepkg.EnsureDefaultMenuSpace(tx, appKey); err != nil {
			return err
		}
		if payload.DefaultSpaceKey == models.DefaultMenuSpaceKey {
			if err := tx.Model(&models.MenuSpace{}).
				Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", appKey, models.DefaultMenuSpaceKey).
				Updates(map[string]interface{}{
					"is_default": true,
					"status":     "normal",
				}).Error; err != nil {
				return err
			}
		}
		if payload.IsDefault {
			if err := tx.Model(&models.App{}).
				Where("app_key <> ? AND deleted_at IS NULL", appKey).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return s.getAppRecord(appKey)
}

// DeleteApp 级联删除一个 App 及其所有子数据。
// 内置默认 App（is_default=true）不允许删除。
func (s *service) DeleteApp(appKey string) error {
	normalizedAppKey := NormalizeAppKey(appKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	var existing models.App
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("应用不存在")
		}
		return err
	}
	if existing.IsDefault {
		return errors.New("内置默认应用不允许删除")
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Level 1 入口绑定
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.AppHostBinding{}).Error; err != nil {
			return fmt.Errorf("删除 APP 入口绑定失败: %w", err)
		}
		// 2. Level 2 菜单空间入口绑定
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.MenuSpaceEntryBinding{}).Error; err != nil {
			return fmt.Errorf("删除菜单空间入口绑定失败: %w", err)
		}
		// 3. 角色 App 范围
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.RoleAppScope{}).Error; err != nil {
			return fmt.Errorf("删除角色 App 范围失败: %w", err)
		}
		// 4. 页面空间绑定 + 页面
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.PageSpaceBinding{}).Error; err != nil {
			return fmt.Errorf("删除页面空间绑定失败: %w", err)
		}
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.UIPage{}).Error; err != nil {
			return fmt.Errorf("删除页面失败: %w", err)
		}
		// 5. 菜单空间布局 + 菜单定义
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.SpaceMenuPlacement{}).Error; err != nil {
			return fmt.Errorf("删除菜单布局失败: %w", err)
		}
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.MenuDefinition{}).Error; err != nil {
			return fmt.Errorf("删除菜单定义失败: %w", err)
		}
		// 6. 菜单空间
		if err := tx.Where("app_key = ?", normalizedAppKey).Delete(&models.MenuSpace{}).Error; err != nil {
			return fmt.Errorf("删除菜单空间失败: %w", err)
		}
		// 7. App 主记录
		if err := tx.Where("app_key = ? AND is_default = false", normalizedAppKey).Delete(&models.App{}).Error; err != nil {
			return fmt.Errorf("删除 App 失败: %w", err)
		}
		return nil
	})
}

func (s *service) ListHostBindings(appKey string) ([]HostBindingRecord, error) {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	query := s.db.Model(&models.AppHostBinding{}).Where("deleted_at IS NULL").Where("app_key = ?", normalizedAppKey)
	var items []models.AppHostBinding
	if err := query.Order("is_primary DESC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	apps, err := s.ListApps()
	if err != nil {
		return nil, err
	}
	appNames := make(map[string]string, len(apps))
	for _, item := range apps {
		appNames[NormalizeAppKey(item.AppKey)] = item.Name
	}
	records := make([]HostBindingRecord, 0, len(items))
	for _, item := range items {
		records = append(records, HostBindingRecord{
			AppHostBinding: item,
			AppName:        appNames[NormalizeAppKey(item.AppKey)],
		})
	}
	return records, nil
}

func normalizeMatchType(value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		v = pathmatch.HostExact
	}
	switch v {
	case pathmatch.HostExact, pathmatch.HostSuffix, pathmatch.PathPrefix, pathmatch.HostAndPath:
		return v, nil
	}
	return "", fmt.Errorf("不支持的匹配类型: %s", value)
}

// validateEntryRule 校验匹配类型与字段组合，并返回规范化后的 host / path。
func validateEntryRule(matchType, host, pathPattern string) (string, string, error) {
	mt, err := normalizeMatchType(matchType)
	if err != nil {
		return "", "", err
	}
	normalizedHost := pathmatch.NormalizeHostPattern(mt, host)
	normalizedPath := pathmatch.NormalizePathPattern(pathPattern)
	switch mt {
	case pathmatch.HostExact, pathmatch.HostSuffix:
		if normalizedHost == "" {
			return "", "", errors.New("Host 不能为空")
		}
		normalizedPath = ""
	case pathmatch.PathPrefix:
		if normalizedPath == "" {
			return "", "", errors.New("路径模式不能为空")
		}
		normalizedHost = ""
	case pathmatch.HostAndPath:
		if normalizedHost == "" || normalizedPath == "" {
			return "", "", errors.New("host_and_path 类型必须同时填写 Host 和路径模式")
		}
	}
	if normalizedPath != "" {
		if _, err := pathmatch.CompilePathPattern(normalizedPath); err != nil {
			return "", "", fmt.Errorf("路径模式编译失败: %w", err)
		}
	}
	return normalizedHost, normalizedPath, nil
}

func (s *service) SaveHostBinding(appKey string, req *SaveHostBindingRequest) (*HostBindingRecord, error) {
	if req == nil {
		return nil, errors.New("入口绑定参数不能为空")
	}
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	if requestAppKey := appctx.NormalizeExplicitAppKey(req.AppKey); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, errors.New("app_key 不匹配")
	}
	if err := ensureDefaultApp(s.db); err != nil {
		return nil, err
	}
	if ok, err := appExists(s.db, normalizedAppKey); err != nil {
		return nil, err
	} else if !ok {
		return nil, errors.New("应用不存在")
	}
	matchType, err := normalizeMatchType(req.MatchType)
	if err != nil {
		return nil, err
	}
	host, pathPattern, err := validateEntryRule(matchType, req.Host, req.PathPattern)
	if err != nil {
		return nil, err
	}
	defaultSpaceKey := spacepkg.NormalizeSpaceKey(req.DefaultSpaceKey)
	if defaultSpaceKey == "" {
		return nil, errors.New("default_space_key 不能为空")
	}
	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}

	binding := models.AppHostBinding{
		AppKey:          normalizedAppKey,
		MatchType:       matchType,
		Host:            host,
		PathPattern:     pathPattern,
		Priority:        req.Priority,
		Description:     strings.TrimSpace(req.Description),
		IsPrimary:       req.IsPrimary,
		DefaultSpaceKey: defaultSpaceKey,
		Status:          status,
		Meta:            models.MetaJSON(req.Meta),
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := ensureSpaceExistsForApp(tx, normalizedAppKey, defaultSpaceKey); err != nil {
			return err
		}
		// 唯一性：(match_type, host, path_pattern) 全局唯一。
		conflictQuery := tx.Where("match_type = ? AND host = ? AND path_pattern = ? AND deleted_at IS NULL", matchType, host, pathPattern)
		if strings.TrimSpace(req.ID) != "" {
			conflictQuery = conflictQuery.Where("id <> ?", req.ID)
		}
		var conflictCount int64
		if err := conflictQuery.Model(&models.AppHostBinding{}).Count(&conflictCount).Error; err != nil {
			return err
		}
		if conflictCount > 0 {
			return errors.New("已存在同匹配类型 + Host + 路径的入口绑定")
		}

		if strings.TrimSpace(req.ID) != "" {
			var existing models.AppHostBinding
			// P1: 限定 app_key 防止跨 App 改绑定。
			if err := tx.Where("id = ? AND app_key = ? AND deleted_at IS NULL", req.ID, normalizedAppKey).First(&existing).Error; err != nil {
				return err
			}
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"app_key":           binding.AppKey,
				"match_type":        binding.MatchType,
				"host":              binding.Host,
				"path_pattern":      binding.PathPattern,
				"priority":          binding.Priority,
				"description":       binding.Description,
				"is_primary":        binding.IsPrimary,
				"default_space_key": binding.DefaultSpaceKey,
				"status":            binding.Status,
				"meta":              binding.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
			binding.ID = existing.ID
		} else {
			if createErr := tx.Create(&binding).Error; createErr != nil {
				return createErr
			}
		}
		if binding.IsPrimary {
			if err := tx.Model(&models.AppHostBinding{}).
				Where("app_key = ? AND id <> ? AND deleted_at IS NULL", binding.AppKey, binding.ID).
				Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	items, err := s.ListHostBindings(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == binding.ID {
			return &items[i], nil
		}
	}
	if len(items) > 0 {
		return &items[0], nil
	}
	return nil, errors.New("保存入口绑定后读取失败")
}

func (s *service) DeleteHostBinding(appKey, id string) error {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("id 不能为空")
	}
	return s.db.Where("id = ? AND app_key = ? AND deleted_at IS NULL", id, normalizedAppKey).
		Delete(&models.AppHostBinding{}).Error
}

func (s *service) ListMenuSpaceEntryBindings(appKey string) ([]MenuSpaceEntryBindingRecord, error) {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	var items []models.MenuSpaceEntryBinding
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).
		Order("priority DESC, is_primary DESC, created_at ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	// app + space name 索引
	var app models.App
	_ = s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&app).Error
	var spaceList []models.MenuSpace
	_ = s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).Find(&spaceList).Error
	spaceNames := make(map[string]string, len(spaceList))
	for _, sp := range spaceList {
		spaceNames[sp.SpaceKey] = sp.Name
	}
	records := make([]MenuSpaceEntryBindingRecord, 0, len(items))
	for _, item := range items {
		records = append(records, MenuSpaceEntryBindingRecord{
			MenuSpaceEntryBinding: item,
			AppName:               app.Name,
			SpaceName:             spaceNames[item.SpaceKey],
		})
	}
	return records, nil
}

func (s *service) SaveMenuSpaceEntryBinding(appKey string, req *SaveMenuSpaceEntryBindingRequest) (*MenuSpaceEntryBindingRecord, error) {
	if req == nil {
		return nil, errors.New("菜单空间入口绑定参数不能为空")
	}
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return nil, errors.New("app_key 不能为空")
	}
	if requestAppKey := appctx.NormalizeExplicitAppKey(req.AppKey); requestAppKey != "" && requestAppKey != normalizedAppKey {
		return nil, errors.New("app_key 不匹配")
	}
	spaceKey := spacepkg.NormalizeSpaceKey(req.SpaceKey)
	if spaceKey == "" {
		return nil, errors.New("space_key 不能为空")
	}
	// 单空间 App 不允许配置 Level 2。
	var app models.App
	if err := s.db.Where("app_key = ? AND deleted_at IS NULL", normalizedAppKey).First(&app).Error; err != nil {
		return nil, err
	}
	if app.SpaceMode != "multi" {
		return nil, errors.New("单空间 App 无需配置菜单空间入口绑定")
	}
	if err := ensureSpaceExistsForApp(s.db, normalizedAppKey, spaceKey); err != nil {
		return nil, err
	}

	matchType, err := normalizeMatchType(req.MatchType)
	if err != nil {
		return nil, err
	}
	host, pathPattern, err := validateEntryRule(matchType, req.Host, req.PathPattern)
	if err != nil {
		return nil, err
	}

	// Level 2 不能超出 Level 1 任何一条规则的范围。
	// 加载 App 的所有 Level 1 绑定，若全部 host 均非空且 child host 都不在范围内 → 拒绝。
	var l1Bindings []models.AppHostBinding
	if err := s.db.Where("app_key = ? AND status = ? AND deleted_at IS NULL", normalizedAppKey, "normal").Find(&l1Bindings).Error; err != nil {
		return nil, err
	}
	if len(l1Bindings) > 0 {
		// 至少匹配一条 Level 1 规则的 host & path 范围
		ok := false
		for _, l1 := range l1Bindings {
			if !pathmatch.IsHostInScope(l1.MatchType, l1.Host, matchType, host) {
				continue
			}
			if !pathmatch.IsPathInScope(l1.PathPattern, pathPattern) {
				continue
			}
			ok = true
			break
		}
		if !ok {
			return nil, errors.New("菜单空间入口绑定必须落在 APP 入口规则范围内")
		}
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "normal"
	}
	binding := models.MenuSpaceEntryBinding{
		AppKey:      normalizedAppKey,
		SpaceKey:    spaceKey,
		MatchType:   matchType,
		Host:        host,
		PathPattern: pathPattern,
		Priority:    req.Priority,
		IsPrimary:   req.IsPrimary,
		Description: strings.TrimSpace(req.Description),
		Status:      status,
		Meta:        models.MetaJSON(req.Meta),
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		conflictQuery := tx.Where("app_key = ? AND match_type = ? AND host = ? AND path_pattern = ? AND deleted_at IS NULL",
			normalizedAppKey, matchType, host, pathPattern)
		if strings.TrimSpace(req.ID) != "" {
			conflictQuery = conflictQuery.Where("id <> ?", req.ID)
		}
		var conflictCount int64
		if err := conflictQuery.Model(&models.MenuSpaceEntryBinding{}).Count(&conflictCount).Error; err != nil {
			return err
		}
		if conflictCount > 0 {
			return errors.New("已存在同匹配类型 + Host + 路径的菜单空间入口绑定")
		}
		if strings.TrimSpace(req.ID) != "" {
			var existing models.MenuSpaceEntryBinding
			// P1: 限定 app_key 防止跨 App 改绑定。
			if err := tx.Where("id = ? AND app_key = ? AND deleted_at IS NULL", req.ID, normalizedAppKey).First(&existing).Error; err != nil {
				return err
			}
			if updateErr := tx.Model(&existing).Updates(map[string]interface{}{
				"space_key":    binding.SpaceKey,
				"match_type":   binding.MatchType,
				"host":         binding.Host,
				"path_pattern": binding.PathPattern,
				"priority":     binding.Priority,
				"is_primary":   binding.IsPrimary,
				"description":  binding.Description,
				"status":       binding.Status,
				"meta":         binding.Meta,
			}).Error; updateErr != nil {
				return updateErr
			}
			binding.ID = existing.ID
		} else {
			if err := tx.Create(&binding).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	items, err := s.ListMenuSpaceEntryBindings(normalizedAppKey)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].ID == binding.ID {
			return &items[i], nil
		}
	}
	return nil, errors.New("保存菜单空间入口绑定后读取失败")
}

func (s *service) DeleteMenuSpaceEntryBinding(appKey, id string) error {
	normalizedAppKey := appctx.NormalizeExplicitAppKey(appKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("id 不能为空")
	}
	return s.db.Where("id = ? AND app_key = ? AND deleted_at IS NULL", id, normalizedAppKey).
		Delete(&models.MenuSpaceEntryBinding{}).Error
}

func ensureSpaceExistsForApp(db *gorm.DB, appKey string, spaceKey string) error {
	normalizedAppKey := NormalizeAppKey(appKey)
	normalizedSpaceKey := spacepkg.NormalizeSpaceKey(spaceKey)
	if normalizedAppKey == "" {
		return errors.New("app_key 不能为空")
	}
	if normalizedSpaceKey == "" {
		return errors.New("default_space_key 不能为空")
	}
	if normalizedSpaceKey == models.DefaultMenuSpaceKey {
		return spacepkg.EnsureDefaultMenuSpace(db, normalizedAppKey)
	}
	var count int64
	if err := db.Model(&models.MenuSpace{}).
		Where("app_key = ? AND space_key = ? AND deleted_at IS NULL", normalizedAppKey, normalizedSpaceKey).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("默认空间不存在，请先在高级空间配置中创建")
	}
	return nil
}

func (s *service) getAppRecord(appKey string) (*AppRecord, error) {
	records, err := s.ListApps()
	if err != nil {
		return nil, err
	}
	target := NormalizeAppKey(appKey)
	for i := range records {
		if NormalizeAppKey(records[i].AppKey) == target {
			record := records[i]
			return &record, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func ensureDefaultApp(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var existing models.App
	err := db.Where("app_key = ? AND deleted_at IS NULL", models.DefaultAppKey).First(&existing).Error
	switch {
	case err == nil:
		return db.Model(&existing).Updates(map[string]interface{}{
			"name":               models.DefaultAppName,
			"space_mode":         "multi",
			"default_space_key":  models.DefaultMenuSpaceKey,
			"status":             "normal",
			"frontend_entry_url": "/",
			"backend_entry_url":  "",
			"health_check_url":   "/health",
			"capabilities":       models.DefaultPlatformAdminCapabilities(),
			"is_default":         true,
		}).Error
	case errors.Is(err, gorm.ErrRecordNotFound):
		return db.Create(&models.App{
			AppKey:           models.DefaultAppKey,
			Name:             models.DefaultAppName,
			Description:      "当前内置管理员后台应用",
			SpaceMode:        "multi",
			DefaultSpaceKey:  models.DefaultMenuSpaceKey,
			AuthMode:         "inherit_host",
			FrontendEntryURL: "/",
			BackendEntryURL:  "",
			HealthCheckURL:   "/health",
			Status:           "normal",
			IsDefault:        true,
			Capabilities:     models.DefaultPlatformAdminCapabilities(),
			Meta:             models.MetaJSON{},
		}).Error
	default:
		return err
	}
}

func normalizeMetaJSON(value map[string]interface{}) models.MetaJSON {
	if len(value) == 0 {
		return models.MetaJSON{}
	}
	return models.MetaJSON(value)
}

func appManifestURL(meta models.MetaJSON) string {
	if len(meta) == 0 {
		return ""
	}
	for _, key := range []string{"manifest_url", "manifestUrl", "remote_manifest_url", "remoteManifestUrl"} {
		if value := strings.TrimSpace(fmt.Sprint(meta[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func appRuntimeVersion(meta models.MetaJSON) string {
	if len(meta) == 0 {
		return ""
	}
	for _, key := range []string{"runtime_version", "runtimeVersion", "version"} {
		if value := strings.TrimSpace(fmt.Sprint(meta[key])); value != "" && value != "<nil>" {
			return value
		}
	}
	return ""
}

func appProbeStatus(item models.App) string {
	if strings.TrimSpace(item.HealthCheckURL) == "" {
		return "missing"
	}
	return "configured"
}

type appProbeResult struct {
	Status    string
	Target    string
	Message   string
	CheckedAt *time.Time
}

func applyAppRemoteContract(meta models.MetaJSON, manifestURL string, runtimeVersion string) {
	if meta == nil {
		return
	}
	upsertMetaString(meta, "manifest_url", strings.TrimSpace(manifestURL))
	delete(meta, "manifestUrl")
	delete(meta, "remote_manifest_url")
	delete(meta, "remoteManifestUrl")
	upsertMetaString(meta, "runtime_version", strings.TrimSpace(runtimeVersion))
	delete(meta, "runtimeVersion")
	delete(meta, "version")
}

func upsertMetaString(meta models.MetaJSON, key string, value string) {
	if meta == nil {
		return
	}
	if value == "" {
		delete(meta, key)
		return
	}
	meta[key] = value
}

func probeAppHealth(item models.App, primaryHosts []string) appProbeResult {
	healthPath := strings.TrimSpace(item.HealthCheckURL)
	if healthPath == "" {
		return appProbeResult{Status: "missing", Message: "未配置 health_check_url"}
	}

	targets := collectAppProbeTargets(item, primaryHosts)
	if len(targets) == 0 {
		return appProbeResult{Status: "unreachable", Message: "缺少可探测 host 或绝对探针地址"}
	}

	var (
		lastMessage  string
		timeoutCount int
	)
	for _, target := range targets {
		status, message := probeHealthTarget(target)
		checkedAt := time.Now()
		if status == "healthy" {
			return appProbeResult{
				Status:    status,
				Target:    target,
				Message:   message,
				CheckedAt: &checkedAt,
			}
		}
		if status == "timeout" {
			timeoutCount++
		}
		lastMessage = message
	}

	checkedAt := time.Now()
	if timeoutCount == len(targets) {
		return appProbeResult{
			Status:    "timeout",
			Target:    targets[0],
			Message:   firstNonEmpty(lastMessage, "探针请求超时"),
			CheckedAt: &checkedAt,
		}
	}
	return appProbeResult{
		Status:    "unreachable",
		Target:    targets[0],
		Message:   firstNonEmpty(lastMessage, "探针请求失败"),
		CheckedAt: &checkedAt,
	}
}

func collectAppProbeTargets(item models.App, primaryHosts []string) []string {
	healthPath := strings.TrimSpace(item.HealthCheckURL)
	if healthPath == "" {
		return nil
	}
	if absoluteProbeURL(healthPath) {
		return []string{healthPath}
	}

	targets := make([]string, 0, len(primaryHosts))
	for _, rawHost := range primaryHosts {
		normalizedHost := strings.TrimSpace(rawHost)
		if normalizedHost == "" {
			continue
		}
		scheme := "http"
		if strings.HasPrefix(strings.ToLower(normalizedHost), "https://") {
			scheme = "https"
			normalizedHost = strings.TrimPrefix(strings.TrimPrefix(normalizedHost, "https://"), "http://")
		} else if strings.HasPrefix(strings.ToLower(normalizedHost), "http://") {
			normalizedHost = strings.TrimPrefix(strings.TrimPrefix(normalizedHost, "http://"), "https://")
		}
		targets = append(targets, fmt.Sprintf("%s://%s%s", scheme, normalizedHost, ensureLeadingSlash(healthPath)))
	}
	return dedupeStrings(targets)
}

func probeHealthTarget(target string) (string, string) {
	req, err := http.NewRequest(http.MethodGet, target, nil)
	if err != nil {
		return "unreachable", fmt.Sprintf("探针地址非法: %v", err)
	}
	resp, err := appProbeHTTPClient.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return "timeout", "探针请求超时"
		}
		return "unreachable", err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return "healthy", fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
	return "unreachable", fmt.Sprintf("HTTP %d", resp.StatusCode)
}

func absoluteProbeURL(value string) bool {
	parsed, err := url.Parse(strings.TrimSpace(value))
	return err == nil && parsed != nil && parsed.Scheme != "" && parsed.Host != ""
}

func ensureLeadingSlash(value string) string {
	if strings.HasPrefix(value, "/") {
		return value
	}
	return "/" + value
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeGovernanceMeta(value map[string]interface{}) (models.MetaJSON, error) {
	normalized := normalizeMetaJSON(value)
	if len(normalized) == 0 {
		return models.MetaJSON{}, nil
	}

	out := models.MetaJSON{}
	for key, raw := range normalized {
		if key == "sensitive_config" || key == "feature_flags" {
			continue
		}
		out[key] = raw
	}

	if raw, ok := normalized["env_profiles"]; ok {
		section, err := normalizeEnvProfiles(raw)
		if err != nil {
			return nil, err
		}
		out["env_profiles"] = section
	}
	return out, nil
}

func normalizeEnvProfiles(raw interface{}) (models.MetaJSON, error) {
	items, ok := toMetaObject(raw)
	if !ok {
		return nil, errors.New("meta.env_profiles 必须是对象，且每个环境节点都必须是对象")
	}
	out := models.MetaJSON{}
	for key, value := range items {
		name := strings.TrimSpace(key)
		if name == "" {
			return nil, errors.New("meta.env_profiles 环境名不能为空")
		}
		profile, ok := toMetaObject(value)
		if !ok {
			return nil, fmt.Errorf("meta.env_profiles.%s 必须是对象", name)
		}
		normalizedValue, err := normalizeJSONObject(profile)
		if err != nil {
			return nil, fmt.Errorf("meta.env_profiles.%s %w", name, err)
		}
		out[name] = normalizedValue
	}
	return out, nil
}

func normalizeJSONObject(input map[string]interface{}) (models.MetaJSON, error) {
	out := models.MetaJSON{}
	for key, value := range input {
		name := strings.TrimSpace(key)
		if name == "" {
			return nil, errors.New("对象键不能为空")
		}
		normalizedValue, err := normalizeLooseJSONValue(value)
		if err != nil {
			return nil, fmt.Errorf("%s", err)
		}
		out[name] = normalizedValue
	}
	return out, nil
}

func normalizeLooseJSONValue(value interface{}) (interface{}, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case string:
		return strings.TrimSpace(typed), nil
	case bool, float64, float32, int, int32, int64, uint, uint32, uint64:
		return typed, nil
	case map[string]interface{}:
		return normalizeJSONObject(typed)
	case []interface{}:
		out := make([]interface{}, 0, len(typed))
		for _, item := range typed {
			normalized, err := normalizeLooseJSONValue(item)
			if err != nil {
				return nil, err
			}
			out = append(out, normalized)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("包含不支持的值类型 %T", value)
	}
}

func toMetaObject(value interface{}) (map[string]interface{}, bool) {
	if value == nil {
		return map[string]interface{}{}, true
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		return typed, true
	case models.MetaJSON:
		out := make(map[string]interface{}, len(typed))
		for key, item := range typed {
			out[key] = item
		}
		return out, true
	default:
		return nil, false
	}
}

func newPreflightCheck(key, title string, passed bool, failLevel string, value, hint string) AppPreflightCheckItem {
	level := "success"
	if !passed {
		level = strings.TrimSpace(failLevel)
		if level == "" {
			level = "warning"
		}
	}
	return AppPreflightCheckItem{
		Key:    key,
		Title:  title,
		Level:  level,
		Passed: passed,
		Value:  strings.TrimSpace(value),
		Hint:   strings.TrimSpace(hint),
	}
}

func buildAppPreflightSummary(checks []AppPreflightCheckItem) AppPreflightSummary {
	summary := AppPreflightSummary{Level: "success"}
	for _, item := range checks {
		switch item.Level {
		case "blocking":
			summary.BlockingCount++
		case "warning":
			summary.WarningCount++
		case "info":
			summary.InfoCount++
		default:
			summary.SuccessCount++
		}
	}
	switch {
	case summary.BlockingCount > 0:
		summary.Level = "blocking"
	case summary.WarningCount > 0:
		summary.Level = "warning"
	case summary.InfoCount > 0:
		summary.Level = "info"
	default:
		summary.Level = "success"
	}
	return summary
}

func pickPrimaryHostBinding(items []HostBindingRecord) *HostBindingRecord {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		if items[i].IsPrimary {
			return &items[i]
		}
	}
	return &items[0]
}

func describePrimaryBinding(item *HostBindingRecord) string {
	if item == nil {
		return "未配置"
	}
	host := strings.TrimSpace(item.Host)
	pathPattern := strings.TrimSpace(item.PathPattern)
	switch item.MatchType {
	case pathmatch.PathPrefix:
		return firstNonEmpty(pathPattern, "未配置")
	case pathmatch.HostAndPath:
		if host != "" && pathPattern != "" {
			return host + pathPattern
		}
		return firstNonEmpty(host, pathPattern, "未配置")
	default:
		if host != "" {
			return host
		}
		return firstNonEmpty(pathPattern, "未配置")
	}
}

func collectAllowedRedirectHosts(items []HostBindingRecord) []string {
	seen := make(map[string]struct{}, len(items)*2)
	out := make([]string, 0, len(items)*2)
	for _, item := range items {
		for _, candidate := range []string{
			strings.TrimSpace(item.Host),
			strings.TrimSpace(metaString(item.Meta, "callback_host", "callbackHost")),
		} {
			normalized := strings.TrimSpace(candidate)
			if normalized == "" {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			out = append(out, normalized)
		}
	}
	sort.Strings(out)
	return out
}

func metaString(meta models.MetaJSON, keys ...string) string {
	for _, key := range keys {
		if value, ok := meta[key]; ok {
			if text, ok := value.(string); ok {
				return strings.TrimSpace(text)
			}
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func loadDefaultApp(db *gorm.DB) (*models.App, error) {
	if err := ensureDefaultApp(db); err != nil {
		return nil, err
	}
	var item models.App
	err := db.Where("is_default = ? AND deleted_at IS NULL", true).Order("updated_at DESC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = db.Where("app_key = ? AND deleted_at IS NULL", models.DefaultAppKey).First(&item).Error
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func appExists(db *gorm.DB, appKey string) (bool, error) {
	if db == nil {
		return false, nil
	}
	var count int64
	if err := db.Model(&models.App{}).Where("app_key = ? AND deleted_at IS NULL", NormalizeAppKey(appKey)).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func loadStringCounts(db *gorm.DB, model interface{}, keyColumn string) (map[string]int, error) {
	type countRow struct {
		Key   string `gorm:"column:key"`
		Total int64  `gorm:"column:total"`
	}
	rows := make([]countRow, 0)
	if err := db.Model(model).
		Select(keyColumn + " AS key, COUNT(*) AS total").
		Where("deleted_at IS NULL").
		Group(keyColumn).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[NormalizeAppKey(row.Key)] = int(row.Total)
	}
	return result, nil
}

func loadStringCountAndHosts(db *gorm.DB, model interface{}, keyColumn string, hostColumn string, primaryFilter string) (map[string]int, map[string][]string, error) {
	counts, err := loadStringCounts(db, model, keyColumn)
	if err != nil {
		return nil, nil, err
	}
	type hostRow struct {
		Key  string `gorm:"column:key"`
		Host string `gorm:"column:host"`
	}
	rows := make([]hostRow, 0)
	query := db.Model(model).
		Select(keyColumn + " AS key, " + hostColumn + " AS host").
		Where("deleted_at IS NULL")
	if strings.TrimSpace(primaryFilter) != "" {
		query = query.Where(primaryFilter)
	}
	if err := query.Order(hostColumn + " ASC").Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	hosts := make(map[string][]string)
	for _, row := range rows {
		key := NormalizeAppKey(row.Key)
		hosts[key] = append(hosts[key], row.Host)
	}
	return counts, hosts, nil
}
