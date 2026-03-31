package page

import (
	"testing"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
)

func TestResolveRuntimeMenuInheritedAccessDeniesWhenMountedMenuIsMissing(t *testing.T) {
	menuID := uuid.New()
	page := Record{
		UIPage: models.UIPage{
			PageKey:        "team.detail",
			ParentMenuID:   &menuID,
			ActiveMenuPath: "/team/all",
		},
	}

	decision := resolveRuntimeMenuInheritedAccess(
		page,
		&runtimeAccessContext{Authenticated: true},
		map[uuid.UUID]runtimeMenuNode{},
		map[string]runtimeMenuNode{},
	)

	if decision.allowed {
		t.Fatalf("decision.allowed = true, want false when mounted menu is not visible")
	}
}

func TestResolveRuntimeMenuInheritedAccessHonorsShowVisibilityMode(t *testing.T) {
	menuID := uuid.New()
	page := Record{
		UIPage: models.UIPage{
			PageKey:        "team.detail",
			ParentMenuID:   &menuID,
			ActiveMenuPath: "/team/all",
		},
	}
	ctx := &runtimeAccessContext{
		Authenticated: true,
		ActionKeys:    map[string]struct{}{},
	}
	node := runtimeMenuNode{
		Menu: models.Menu{
			ID:       menuID,
			SpaceKey: "default",
			Meta: models.MetaJSON{
				"accessMode":           "permission",
				"requiredAction":       "team.read",
				"actionVisibilityMode": "show",
			},
		},
		FullPath: "/team/all",
	}

	decision := resolveRuntimeMenuInheritedAccess(
		page,
		ctx,
		map[uuid.UUID]runtimeMenuNode{menuID: node},
		map[string]runtimeMenuNode{"/team/all": node},
	)

	if !decision.allowed {
		t.Fatalf("decision.allowed = false, want true when menu keeps route visible in show mode")
	}
	if decision.effectiveMode != "permission" {
		t.Fatalf("decision.effectiveMode = %q, want permission", decision.effectiveMode)
	}
}
