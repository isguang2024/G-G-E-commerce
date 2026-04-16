package page

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/models"
	"github.com/maben/backend/internal/pkg/permissionkey"
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

// ── intersectVisibleMenusByPermissionKey 三条关键分支 ─────────────────────────

func TestIntersectVisibleMenusByPermissionKey(t *testing.T) {
	// 三个菜单：
	//   audit: 声明 PermissionKey=observability.audit.read
	//   telemetry: 声明 PermissionKey=observability.telemetry.read
	//   plain: 未声明 PermissionKey（空），保持兼容
	auditID := uuid.New()
	telemetryID := uuid.New()
	plainID := uuid.New()

	menuMap := map[uuid.UUID]runtimeMenuNode{
		auditID: {
			Menu: models.Menu{
				ID:            auditID,
				Name:          "AuditLog",
				PermissionKey: "observability.audit.read",
			},
		},
		telemetryID: {
			Menu: models.Menu{
				ID:            telemetryID,
				Name:          "TelemetryLog",
				PermissionKey: "observability.telemetry.read",
			},
		},
		plainID: {
			Menu: models.Menu{
				ID:   plainID,
				Name: "Plain",
			},
		},
	}
	visibleMenuIDs := []uuid.UUID{auditID, telemetryID, plainID}

	assertContainsAll := func(t *testing.T, got []uuid.UUID, want ...uuid.UUID) {
		t.Helper()
		gotSet := map[uuid.UUID]struct{}{}
		for _, id := range got {
			gotSet[id] = struct{}{}
		}
		for _, id := range want {
			if _, ok := gotSet[id]; !ok {
				t.Errorf("missing expected menu id %s in %v", id, got)
			}
		}
	}

	t.Run("低权限账号缺 action_key: AuditLog/TelemetryLog 被剔除, Plain 保留", func(t *testing.T) {
		actionKeys := map[string]struct{}{} // 无任何 action
		result := intersectVisibleMenusByPermissionKey(menuMap, visibleMenuIDs, actionKeys)
		sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
		if len(result) != 1 {
			t.Fatalf("expected 1 menu (plain), got %d: %v", len(result), result)
		}
		if result[0] != plainID {
			t.Fatalf("expected plainID %s, got %s", plainID, result[0])
		}
	})

	t.Run("高权限账号具备 audit.read: AuditLog 可见, TelemetryLog 仍被剔除", func(t *testing.T) {
		actionKeys := map[string]struct{}{
			permissionkey.Normalize("observability.audit.read"): {},
		}
		result := intersectVisibleMenusByPermissionKey(menuMap, visibleMenuIDs, actionKeys)
		assertContainsAll(t, result, auditID, plainID)
		for _, id := range result {
			if id == telemetryID {
				t.Fatalf("telemetry menu should be filtered out when missing telemetry.read")
			}
		}
	})

	t.Run("同时具备 audit.read+telemetry.read: 三个菜单全可见", func(t *testing.T) {
		actionKeys := map[string]struct{}{
			permissionkey.Normalize("observability.audit.read"):     {},
			permissionkey.Normalize("observability.telemetry.read"): {},
		}
		result := intersectVisibleMenusByPermissionKey(menuMap, visibleMenuIDs, actionKeys)
		if len(result) != 3 {
			t.Fatalf("expected all 3 menus visible, got %d: %v", len(result), result)
		}
		assertContainsAll(t, result, auditID, telemetryID, plainID)
	})

	t.Run("menuMap 找不到的 id 按现状保留, 不崩", func(t *testing.T) {
		ghostID := uuid.New()
		result := intersectVisibleMenusByPermissionKey(menuMap, []uuid.UUID{ghostID, plainID}, map[string]struct{}{})
		assertContainsAll(t, result, ghostID, plainID)
	})
}

