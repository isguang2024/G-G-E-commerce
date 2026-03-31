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

