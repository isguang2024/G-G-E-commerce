package page

import (
	"testing"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func TestNormalizeLegacyGlobalPageClearsBindings(t *testing.T) {
	menuID := uuid.New()
	parentPageKey := "dashboard.console"

	got := normalizeLegacyGlobalPage(models.UIPage{
		PageType:       "global",
		ParentMenuID:   &menuID,
		ParentPageKey:  parentPageKey,
		ActiveMenuPath: "/dashboard/console/user-center",
	})

	if got.ParentMenuID != nil {
		t.Fatalf("ParentMenuID = %#v, want nil", got.ParentMenuID)
	}
	if got.ParentPageKey != "" {
		t.Fatalf("ParentPageKey = %q, want empty", got.ParentPageKey)
	}
	if got.ActiveMenuPath != "" {
		t.Fatalf("ActiveMenuPath = %q, want empty", got.ActiveMenuPath)
	}
}

func TestResolvePageVisibilityScopeForGlobalPage(t *testing.T) {
	got := resolvePageVisibilityScope(
		models.UIPage{
			PageType: models.PageTypeGlobal,
		},
		nil,
	)

	if got != pageVisibilityScopeApp {
		t.Fatalf("resolvePageVisibilityScope(global) = %q, want %q", got, pageVisibilityScopeApp)
	}
}

func TestResolvePageVisibilityScopeForGlobalBoundPage(t *testing.T) {
	got := resolvePageVisibilityScope(
		models.UIPage{
			PageType: models.PageTypeGlobal,
		},
		[]string{"ops"},
	)

	if got != pageVisibilityScopeSpaces {
		t.Fatalf("resolvePageVisibilityScope(global-bound) = %q, want %q", got, pageVisibilityScopeSpaces)
	}
}

func TestResolvePageVisibilityScopeForInheritedInnerPage(t *testing.T) {
	menuID := uuid.New()
	got := resolvePageVisibilityScope(
		models.UIPage{
			PageType:     models.PageTypeInner,
			ParentMenuID: &menuID,
		},
		[]string{"ops"},
	)

	if got != pageVisibilityScopeInherit {
		t.Fatalf("resolvePageVisibilityScope(inner-with-parent) = %q, want %q", got, pageVisibilityScopeInherit)
	}
}

func TestResolvePageVisibilityScopeForStandaloneBoundPage(t *testing.T) {
	got := resolvePageVisibilityScope(
		models.UIPage{
			PageType: models.PageTypeStandalone,
		},
		[]string{"ops"},
	)

	if got != pageVisibilityScopeSpaces {
		t.Fatalf("resolvePageVisibilityScope(standalone-bound) = %q, want %q", got, pageVisibilityScopeSpaces)
	}
}

func TestApplyResolvedPageSpaceSetsVisibilityScope(t *testing.T) {
	item := models.UIPage{
		PageType: models.PageTypeStandalone,
		Meta:     models.MetaJSON{},
	}

	applyResolvedPageSpace(&item, []string{"ops"})

	if item.VisibilityScope != pageVisibilityScopeSpaces {
		t.Fatalf("VisibilityScope = %q, want %q", item.VisibilityScope, pageVisibilityScopeSpaces)
	}
}

func TestApplyResolvedPageSpaceClearsLegacySpaceKey(t *testing.T) {
	item := models.UIPage{
		PageType: models.PageTypeGlobal,
		SpaceKey: "default",
		Meta:     models.MetaJSON{},
	}

	applyResolvedPageSpace(&item, []string{"ops"})

	if item.SpaceKey != "" {
		t.Fatalf("SpaceKey = %q, want empty", item.SpaceKey)
	}
	if item.VisibilityScope != pageVisibilityScopeSpaces {
		t.Fatalf("VisibilityScope = %q, want %q", item.VisibilityScope, pageVisibilityScopeSpaces)
	}
}
