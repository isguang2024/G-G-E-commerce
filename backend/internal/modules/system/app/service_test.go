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
	if err.Error() != "app_key is required" {
		t.Fatalf("SaveHostBinding() error = %q, want %q", err.Error(), "app_key is required")
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
	if err.Error() != "app_key mismatch" {
		t.Fatalf("SaveHostBinding() error = %q, want %q", err.Error(), "app_key mismatch")
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
