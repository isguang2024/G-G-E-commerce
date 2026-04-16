package page

import (
	"testing"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/models"
)

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

