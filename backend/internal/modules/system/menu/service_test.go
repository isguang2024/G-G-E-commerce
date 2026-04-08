package menu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	"github.com/gg-ecommerce/backend/internal/modules/system/user"
)

func TestMergeSystemMenuIDsIncludesJWTAndPublicMenus(t *testing.T) {
	permissionID := uuid.New()
	jwtID := uuid.New()
	publicID := uuid.New()

	merged := (&menuService{}).mergeSystemMenuIDs(
		[]user.Menu{
			{ID: permissionID, Meta: models.MetaJSON{"accessMode": "permission"}},
			{ID: jwtID, Meta: models.MetaJSON{"accessMode": "jwt"}},
			{ID: publicID, Meta: models.MetaJSON{"accessMode": "public"}},
		},
		[]uuid.UUID{permissionID},
	)

	got := make(map[uuid.UUID]struct{}, len(merged))
	for _, id := range merged {
		got[id] = struct{}{}
	}

	for _, want := range []uuid.UUID{permissionID, jwtID, publicID} {
		if _, ok := got[want]; !ok {
			t.Fatalf("merged menu ids missing %s", want)
		}
	}
}

func TestFilterTreeByMenuIDsSkipsDisabledMenus(t *testing.T) {
	rootID := uuid.New()
	disabledChildID := uuid.New()

	tree := []*user.Menu{
		{
			ID:   rootID,
			Name: "System",
			Meta: models.MetaJSON{"isEnable": true},
			Children: []*user.Menu{
				{
					ID:   disabledChildID,
					Name: "TeamAll",
					Meta: models.MetaJSON{"isEnable": false},
				},
			},
		},
	}

	filtered := (&menuService{}).filterTreeByMenuIDs(tree, map[uuid.UUID]struct{}{
		rootID:          {},
		disabledChildID: {},
	})

	if len(filtered) != 1 {
		t.Fatalf("len(filtered) = %d, want 1", len(filtered))
	}
	if len(filtered[0].Children) != 0 {
		t.Fatalf("len(filtered[0].Children) = %d, want 0 for disabled child", len(filtered[0].Children))
	}
}

func TestBuildPromotedMenuFullPathMapRemapsChildren(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	grandChildID := uuid.New()

	flat := []user.Menu{
		{ID: rootID, Path: "dashboard"},
		{ID: childID, ParentID: &rootID, Path: "console"},
		{ID: grandChildID, ParentID: &childID, Path: "user-center"},
	}

	got := buildPromotedMenuFullPathMap(flat, rootID, nil)

	if _, exists := got[rootID]; exists {
		t.Fatalf("promoted path map should not contain deleted root %s", rootID)
	}
	if got[childID] != "/console" {
		t.Fatalf("promoted child path = %q, want %q", got[childID], "/console")
	}
	if got[grandChildID] != "/console/user-center" {
		t.Fatalf("promoted grand child path = %q, want %q", got[grandChildID], "/console/user-center")
	}
}

func TestValidateMenuPromoteTargetRejectsSelfAndDescendants(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()

	flat := []user.Menu{
		{ID: rootID, Path: "dashboard"},
		{ID: childID, ParentID: &rootID, Path: "console"},
	}

	if err := validateMenuPromoteTarget(flat, rootID, rootID); err == nil {
		t.Fatalf("validateMenuPromoteTarget() should reject self as target")
	}
	if err := validateMenuPromoteTarget(flat, rootID, childID); err == nil {
		t.Fatalf("validateMenuPromoteTarget() should reject descendants as target")
	}
}
