package app

import (
	"testing"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func TestSaveHostBindingRequiresExplicitAppKey(t *testing.T) {
	svc := &service{}

	_, err := svc.SaveHostBinding("", &SaveHostBindingRequest{Host: "admin.example.com"})
	if err == nil {
		t.Fatalf("SaveHostBinding() error = nil, want explicit app_key error")
	}
	if err.Error() != "app_key 不能为空" {
		t.Fatalf("SaveHostBinding() error = %q, want %q", err.Error(), "app_key 不能为空")
	}
}

func TestSaveHostBindingRejectsBodyAppKeyMismatch(t *testing.T) {
	svc := &service{}

	_, err := svc.SaveHostBinding(models.DefaultAppKey, &SaveHostBindingRequest{
		AppKey: "merchant-console",
		Host:   "admin.example.com",
	})
	if err == nil {
		t.Fatalf("SaveHostBinding() error = nil, want app_key mismatch error")
	}
	if err.Error() != "app_key 不匹配" {
		t.Fatalf("SaveHostBinding() error = %q, want %q", err.Error(), "app_key 不匹配")
	}
}

func TestSaveHostBindingAllowsEmptyBodyAppKeyWithoutDefaultMismatch(t *testing.T) {
	svc := &service{}

	_, err := svc.SaveHostBinding("merchant-console", &SaveHostBindingRequest{
		Host:            "admin.example.com",
		DefaultSpaceKey: "ops",
	})
	if err == nil {
		t.Fatalf("SaveHostBinding() error = nil, want downstream validation error")
	}
	if err.Error() != "应用不存在" {
		t.Fatalf("SaveHostBinding() error = %q, want %q", err.Error(), "应用不存在")
	}
}

func TestCollectAllowedRedirectHostsIncludesBindingAndCallbackHost(t *testing.T) {
	items := []HostBindingRecord{
		{
			AppHostBinding: models.AppHostBinding{
				Host: "admin.example.com",
				Meta: models.MetaJSON{
					"callback_host": "auth.example.com",
				},
			},
		},
		{
			AppHostBinding: models.AppHostBinding{
				Host: "admin.example.com",
			},
		},
	}

	got := collectAllowedRedirectHosts(items)
	if len(got) != 2 {
		t.Fatalf("collectAllowedRedirectHosts() len = %d, want 2", len(got))
	}
	if got[0] != "admin.example.com" || got[1] != "auth.example.com" {
		t.Fatalf("collectAllowedRedirectHosts() = %#v", got)
	}
}

func TestBuildAppPreflightSummaryTracksHighestSeverity(t *testing.T) {
	summary := buildAppPreflightSummary([]AppPreflightCheckItem{
		{Level: "success"},
		{Level: "info"},
		{Level: "warning"},
		{Level: "blocking"},
	})

	if summary.Level != "blocking" {
		t.Fatalf("summary.Level = %q, want %q", summary.Level, "blocking")
	}
	if summary.SuccessCount != 1 || summary.InfoCount != 1 || summary.WarningCount != 1 || summary.BlockingCount != 1 {
		t.Fatalf("summary counts = %#v", summary)
	}
}

func TestNormalizeGovernanceMetaPreservesUnknownKeys(t *testing.T) {
	meta, err := normalizeGovernanceMeta(map[string]interface{}{
		"note": "keep",
		"env_profiles": map[string]interface{}{
			"dev": map[string]interface{}{
				"frontend_base": " http://127.0.0.1:5174 ",
			},
		},
		"feature_flags": map[string]interface{}{
			"shared_cookie": map[string]interface{}{
				"default": false,
				"prod":    true,
			},
		},
		"sensitive_config": map[string]interface{}{
			"secret_refs": []interface{}{" oidc/client-secret ", "gateway/token"},
		},
	})
	if err != nil {
		t.Fatalf("normalizeGovernanceMeta() error = %v", err)
	}
	if meta["note"] != "keep" {
		t.Fatalf("meta[note] = %#v, want keep", meta["note"])
	}
	envProfiles, ok := meta["env_profiles"].(models.MetaJSON)
	if !ok {
		t.Fatalf("meta[env_profiles] type = %T, want models.MetaJSON", meta["env_profiles"])
	}
	devProfile, ok := envProfiles["dev"].(models.MetaJSON)
	if !ok || devProfile["frontend_base"] != "http://127.0.0.1:5174" {
		t.Fatalf("env_profiles.dev = %#v", envProfiles["dev"])
	}
}

func TestNormalizeGovernanceMetaRejectsInvalidFeatureFlags(t *testing.T) {
	_, err := normalizeGovernanceMeta(map[string]interface{}{
		"feature_flags": map[string]interface{}{
			"shared_cookie": "enabled",
		},
	})
	if err == nil {
		t.Fatalf("normalizeGovernanceMeta() error = nil, want feature flag validation error")
	}
	if err.Error() != "meta.feature_flags.shared_cookie 必须是布尔值或环境覆盖对象" {
		t.Fatalf("normalizeGovernanceMeta() error = %q", err.Error())
	}
}

func TestNormalizeGovernanceMetaRejectsInvalidSensitiveConfigArray(t *testing.T) {
	_, err := normalizeGovernanceMeta(map[string]interface{}{
		"sensitive_config": map[string]interface{}{
			"secret_refs": []interface{}{"oidc/client-secret", 1},
		},
	})
	if err == nil {
		t.Fatalf("normalizeGovernanceMeta() error = nil, want sensitive config validation error")
	}
	if err.Error() != "meta.sensitive_config.secret_refs 数组元素必须全部为字符串" {
		t.Fatalf("normalizeGovernanceMeta() error = %q", err.Error())
	}
}
