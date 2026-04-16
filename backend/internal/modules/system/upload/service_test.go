package upload

import (
	"testing"
	"time"

	"github.com/maben/backend/internal/modules/system/models"
)

func TestBuildRelativeDir(t *testing.T) {
	now := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	got := buildRelativeDir("public-media", "{yyyy}/{mm}/{dd}", "images", now)
	want := "public-media/2026/04/16/images"
	if got != want {
		t.Fatalf("buildRelativeDir() = %q, want %q", got, want)
	}
}

func TestJoinURLPath(t *testing.T) {
	got := joinURLPath("/uploads/", "/public-media/")
	if got != "/uploads/public-media" {
		t.Fatalf("joinURLPath() = %q, want %q", got, "/uploads/public-media")
	}
}

func TestNormalizeTenantIDUsesConfiguredDefault(t *testing.T) {
	setDefaultTenantID("tenant-a")
	t.Cleanup(func() {
		setDefaultTenantID("default")
	})

	if got := normalizeTenantID(""); got != "tenant-a" {
		t.Fatalf("normalizeTenantID() = %q, want %q", got, "tenant-a")
	}
}

func TestServiceHelpers(t *testing.T) {
	if got := buildPublicURL("", "2026/04/16/demo.png"); got != "/uploads/2026/04/16/demo.png" {
		t.Fatalf("buildPublicURL() = %q", got)
	}
	if got := normalizeContentType("", "demo.png"); got != "image/png" {
		t.Fatalf("normalizeContentType() = %q, want image/png", got)
	}
	if got := normalizedExt("avatar", "image/png"); got != ".png" {
		t.Fatalf("normalizedExt() = %q, want .png", got)
	}
	if !mimeAllowed("image/png", models.StringList{"image/*"}, nil) {
		t.Fatalf("mimeAllowed() should allow image/png with wildcard")
	}
	if mimeAllowed("text/plain", models.StringList{"image/*"}, nil) {
		t.Fatalf("mimeAllowed() should reject text/plain")
	}
	if got := getMetaString(models.MetaJSON{"driver": "local"}, "driver", "fallback"); got != "local" {
		t.Fatalf("getMetaString() = %q, want local", got)
	}
	if got := getMetaString(models.MetaJSON{"driver": 1}, "driver", "fallback"); got != "fallback" {
		t.Fatalf("getMetaString() fallback = %q, want fallback", got)
	}
	if got := joinURLPath("", ""); got != "/uploads" {
		t.Fatalf("joinURLPath(empty, empty) = %q, want /uploads", got)
	}
	if got := joinURLPath("/", "child"); got != "/child" {
		t.Fatalf("joinURLPath(root, child) = %q, want /child", got)
	}
}

