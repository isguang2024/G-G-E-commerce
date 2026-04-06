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
			PageKey:        "collaboration_workspace.detail",
			ParentMenuID:   &menuID,
			ActiveMenuPath: "/collaboration-workspace/all",
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
			PageKey:        "collaboration_workspace.detail",
			ParentMenuID:   &menuID,
			ActiveMenuPath: "/collaboration-workspace/all",
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
				"requiredAction":       "collaboration_workspace.read",
				"actionVisibilityMode": "show",
			},
		},
		FullPath: "/collaboration-workspace/all",
	}

	decision := resolveRuntimeMenuInheritedAccess(
		page,
		ctx,
		map[uuid.UUID]runtimeMenuNode{menuID: node},
		map[string]runtimeMenuNode{"/collaboration-workspace/all": node},
	)

	if !decision.allowed {
		t.Fatalf("decision.allowed = false, want true when menu keeps route visible in show mode")
	}
	if decision.effectiveMode != "permission" {
		t.Fatalf("decision.effectiveMode = %q, want permission", decision.effectiveMode)
	}
}
