package siteconfig

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/pkg/logger"
)

func newTestSiteConfigService(t *testing.T) Service {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open(sqlite) error = %v", err)
	}
	schema := []string{
		`CREATE TABLE IF NOT EXISTS site_configs (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL DEFAULT 'default',
			app_key TEXT NOT NULL DEFAULT '',
			config_key TEXT NOT NULL,
			config_value TEXT NOT NULL DEFAULT '{}',
			value_type TEXT NOT NULL DEFAULT 'string',
			fallback_policy TEXT NOT NULL DEFAULT 'inherit',
			label TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			is_builtin NUMERIC NOT NULL DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'normal',
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS site_config_set_items (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL DEFAULT 'default',
			set_id TEXT NOT NULL,
			config_key TEXT NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("create test schema error = %v", err)
		}
	}
	return NewServiceWithCache(NewRepository(db), nil, zap.NewNop())
}

func mustUpsertConfig(t *testing.T, svc Service, cfg *models.SiteConfig) {
	t.Helper()
	if err := svc.UpsertConfig(context.Background(), cfg); err != nil {
		t.Fatalf("UpsertConfig() error = %v", err)
	}
}

func TestResolveFallsBackToGlobalOnlyWhenAllowed(t *testing.T) {
	svc := newTestSiteConfigService(t)
	ctx := logger.WithApp(context.Background(), "admin")

	mustUpsertConfig(t, svc, &models.SiteConfig{
		AppKey:         "",
		ConfigKey:      "site.title",
		ConfigValue:    models.MetaJSON{"value": "global"},
		ValueType:      models.SiteConfigValueTypeString,
		FallbackPolicy: models.SiteConfigFallbackPolicyInherit,
		Status:         "normal",
	})
	mustUpsertConfig(t, svc, &models.SiteConfig{
		AppKey:         "admin",
		ConfigKey:      "site.title",
		ConfigValue:    models.MetaJSON{"value": "app"},
		ValueType:      models.SiteConfigValueTypeString,
		FallbackPolicy: models.SiteConfigFallbackPolicyStrict,
		Status:         "normal",
	})

	appResolved, err := svc.Resolve(ctx, ResolveRequest{
		ScopeType: ScopeTypeApp,
		ScopeKey:  "admin",
		Keys:      []string{"site.title"},
	})
	if err != nil {
		t.Fatalf("Resolve(app) error = %v", err)
	}
	if got := appResolved.Items["site.title"].Source; got != ResolveSourceApp {
		t.Fatalf("Resolve(app) source = %q, want %q", got, ResolveSourceApp)
	}

	globalResolved, err := svc.Resolve(ctx, ResolveRequest{
		ScopeType: ScopeTypeApp,
		ScopeKey:  "shop",
		Keys:      []string{"site.title"},
	})
	if err != nil {
		t.Fatalf("Resolve(app fallback) error = %v", err)
	}
	if got := globalResolved.Items["site.title"].Source; got != ResolveSourceGlobal {
		t.Fatalf("Resolve(app fallback) source = %q, want %q", got, ResolveSourceGlobal)
	}

	lookupGlobal, err := svc.Lookup(ctx, LookupRequest{
		ScopeType: ScopeTypeGlobal,
		ConfigKey: "site.title",
	})
	if err != nil {
		t.Fatalf("Lookup(global) error = %v", err)
	}
	if got := lookupGlobal.Item.Source; got != ResolveSourceGlobal {
		t.Fatalf("Lookup(global) source = %q, want %q", got, ResolveSourceGlobal)
	}
}

func TestResolveDoesNotFallbackWhenStrict(t *testing.T) {
	svc := newTestSiteConfigService(t)
	ctx := logger.WithApp(context.Background(), "admin")

	mustUpsertConfig(t, svc, &models.SiteConfig{
		AppKey:         "",
		ConfigKey:      "site.strict",
		ConfigValue:    models.MetaJSON{"value": "global"},
		ValueType:      models.SiteConfigValueTypeString,
		FallbackPolicy: models.SiteConfigFallbackPolicyStrict,
		Status:         "normal",
	})

	resolved, err := svc.Resolve(ctx, ResolveRequest{
		ScopeType: ScopeTypeApp,
		ScopeKey:  "admin",
		Keys:      []string{"site.strict"},
	})
	if err != nil {
		t.Fatalf("Resolve(strict) error = %v", err)
	}
	item := resolved.Items["site.strict"]
	if item.Source != ResolveSourceDefault {
		t.Fatalf("Resolve(strict) source = %q, want %q", item.Source, ResolveSourceDefault)
	}
	if len(item.Value) != 0 {
		t.Fatalf("Resolve(strict) value = %#v, want empty", item.Value)
	}
}

func TestResolveContextScopeUsesCurrentApp(t *testing.T) {
	svc := newTestSiteConfigService(t)
	ctx := logger.WithApp(context.Background(), "admin")

	mustUpsertConfig(t, svc, &models.SiteConfig{
		AppKey:         "",
		ConfigKey:      "site.context",
		ConfigValue:    models.MetaJSON{"value": "global"},
		ValueType:      models.SiteConfigValueTypeString,
		FallbackPolicy: models.SiteConfigFallbackPolicyInherit,
		Status:         "normal",
	})

	resolved, err := svc.Resolve(ctx, ResolveRequest{
		ScopeType: ScopeTypeContext,
		Keys:      []string{"site.context"},
	})
	if err != nil {
		t.Fatalf("Resolve(context) error = %v", err)
	}
	if got := resolved.ScopeType; got != ScopeTypeApp {
		t.Fatalf("Resolve(context) scope_type = %q, want %q", got, ScopeTypeApp)
	}
	if got := resolved.ScopeKey; got != "admin" {
		t.Fatalf("Resolve(context) scope_key = %q, want %q", got, "admin")
	}
	if got := resolved.Items["site.context"].Source; got != ResolveSourceGlobal {
		t.Fatalf("Resolve(context) source = %q, want %q", got, ResolveSourceGlobal)
	}
}
