package menu

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/internal/modules/system/models"
	spaceutil "github.com/gg-ecommerce/backend/internal/modules/system/space"
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

func TestFilterBackupsBySpaceKeepsLegacyGlobalBackup(t *testing.T) {
	defaultBackupID := uuid.New()
	globalBackupID := uuid.New()
	opsBackupID := uuid.New()

	backups := filterBackupsBySpace([]user.MenuBackup{
		{ID: defaultBackupID, SpaceKey: "default"},
		{ID: globalBackupID, SpaceKey: ""},
		{ID: opsBackupID, SpaceKey: "ops"},
	}, "default")

	if len(backups) != 2 {
		t.Fatalf("len(backups) = %d, want 2", len(backups))
	}

	got := map[uuid.UUID]struct{}{}
	for _, item := range backups {
		got[item.ID] = struct{}{}
	}
	for _, want := range []uuid.UUID{defaultBackupID, globalBackupID} {
		if _, exists := got[want]; !exists {
			t.Fatalf("filtered backups missing %s", want)
		}
	}
	if _, exists := got[opsBackupID]; exists {
		t.Fatalf("ops backup %s should not be returned for default space filter", opsBackupID)
	}
}

func TestResolveMenuBackupScopeInfoDistinguishesOrigins(t *testing.T) {
	spacePayload, err := json.Marshal(ginMenuBackupPayload{
		Version:   menuBackupPayloadVersion,
		ScopeType: "space",
		SpaceKey:  "ops",
		Menus: []user.Menu{
			{ID: uuid.New(), SpaceKey: "ops", Name: "OpsMenu"},
		},
	})
	if err != nil {
		t.Fatalf("marshal space payload failed: %v", err)
	}

	globalPayload, err := json.Marshal(ginMenuBackupPayload{
		Version:   menuBackupPayloadVersion,
		ScopeType: "global",
		Menus: []user.Menu{
			{ID: uuid.New(), SpaceKey: "default", Name: "DefaultMenu"},
		},
	})
	if err != nil {
		t.Fatalf("marshal global payload failed: %v", err)
	}

	legacyPayload, err := json.Marshal([]user.Menu{
		{ID: uuid.New(), SpaceKey: "default", Name: "LegacyMenu"},
	})
	if err != nil {
		t.Fatalf("marshal legacy payload failed: %v", err)
	}

	testCases := []struct {
		name   string
		backup user.MenuBackup
		want   menuBackupScopeInfo
	}{
		{
			name: "space column",
			backup: user.MenuBackup{
				SpaceKey: "ops",
				MenuData: string(spacePayload),
			},
			want: menuBackupScopeInfo{
				ScopeType:   "space",
				ScopeOrigin: "menu_backup",
				AppKey:      "platform-admin",
				SpaceKey:    "ops",
			},
		},
		{
			name: "explicit global payload",
			backup: user.MenuBackup{
				MenuData: string(globalPayload),
			},
			want: menuBackupScopeInfo{
				ScopeType:   "global",
				ScopeOrigin: "menu_backup",
				AppKey:      "platform-admin",
			},
		},
		{
			name: "legacy raw array payload",
			backup: user.MenuBackup{
				MenuData: string(legacyPayload),
			},
			want: menuBackupScopeInfo{
				ScopeType:   "global",
				ScopeOrigin: "menu_backup",
				AppKey:      "platform-admin",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveMenuBackupScopeInfo(tc.backup)
			if got != tc.want {
				t.Fatalf("resolveMenuBackupScopeInfo() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestFilterBackupsBySpaceUsesPayloadScopeFallback(t *testing.T) {
	spacePayload, err := json.Marshal(ginMenuBackupPayload{
		Version:   menuBackupPayloadVersion,
		ScopeType: "space",
		SpaceKey:  "ops",
		Menus: []user.Menu{
			{ID: uuid.New(), SpaceKey: "ops", Name: "OpsMenu"},
		},
	})
	if err != nil {
		t.Fatalf("marshal space payload failed: %v", err)
	}

	globalPayload, err := json.Marshal(ginMenuBackupPayload{
		Version:   menuBackupPayloadVersion,
		ScopeType: "global",
		Menus: []user.Menu{
			{ID: uuid.New(), SpaceKey: "default", Name: "DefaultMenu"},
		},
	})
	if err != nil {
		t.Fatalf("marshal global payload failed: %v", err)
	}

	legacyPayload, err := json.Marshal([]user.Menu{
		{ID: uuid.New(), SpaceKey: "default", Name: "LegacyMenu"},
	})
	if err != nil {
		t.Fatalf("marshal legacy payload failed: %v", err)
	}

	defaultColumnBackupID := uuid.New()
	payloadOnlySpaceBackupID := uuid.New()
	explicitGlobalBackupID := uuid.New()
	legacyGlobalBackupID := uuid.New()

	backups := filterBackupsBySpace([]user.MenuBackup{
		{ID: defaultColumnBackupID, SpaceKey: "default", MenuData: string(globalPayload)},
		{ID: payloadOnlySpaceBackupID, MenuData: string(spacePayload)},
		{ID: explicitGlobalBackupID, MenuData: string(globalPayload)},
		{ID: legacyGlobalBackupID, MenuData: string(legacyPayload)},
	}, "ops")

	got := map[uuid.UUID]struct{}{}
	for _, item := range backups {
		got[item.ID] = struct{}{}
	}

	for _, want := range []uuid.UUID{payloadOnlySpaceBackupID, explicitGlobalBackupID, legacyGlobalBackupID} {
		if _, exists := got[want]; !exists {
			t.Fatalf("filtered backups missing %s", want)
		}
	}
	if _, exists := got[defaultColumnBackupID]; exists {
		t.Fatalf("default space backup %s should not be returned for ops space filter", defaultColumnBackupID)
	}
}

func TestParseMenuBackupPayloadSupportsV2AndLegacyFormats(t *testing.T) {
	spaceMenu := user.Menu{
		ID:       uuid.New(),
		SpaceKey: "ops",
		Name:     "OpsMenu",
	}
	v2Raw, err := json.Marshal(ginMenuBackupPayload{
		Version:   menuBackupPayloadVersion,
		ScopeType: "space",
		SpaceKey:  "ops",
		Menus:     []user.Menu{spaceMenu},
	})
	if err != nil {
		t.Fatalf("marshal v2 payload failed: %v", err)
	}

	parsedV2, err := parseMenuBackupPayload(string(v2Raw))
	if err != nil {
		t.Fatalf("parse v2 payload failed: %v", err)
	}
	if parsedV2.SpaceKey != "ops" {
		t.Fatalf("parsedV2.SpaceKey = %q, want %q", parsedV2.SpaceKey, "ops")
	}
	if parsedV2.ScopeType != "space" {
		t.Fatalf("parsedV2.ScopeType = %q, want %q", parsedV2.ScopeType, "space")
	}
	if len(parsedV2.Menus) != 1 || parsedV2.Menus[0].Name != "OpsMenu" {
		t.Fatalf("parsedV2.Menus = %#v, want one OpsMenu", parsedV2.Menus)
	}

	legacyRaw, err := json.Marshal([]user.Menu{spaceMenu})
	if err != nil {
		t.Fatalf("marshal legacy payload failed: %v", err)
	}

	if _, err := parseMenuBackupPayload(string(legacyRaw)); err == nil {
		t.Fatalf("parse legacy payload should fail for raw array payload")
	}
}

func TestNormalizeBackupScopeTypeSupportsExplicitGlobalScope(t *testing.T) {
	if got := normalizeBackupScopeType("global", "default"); got != "global" {
		t.Fatalf("normalizeBackupScopeType(global, default) = %q, want global", got)
	}
	if got := resolveBackupSpaceKey("global", "default"); got != "" {
		t.Fatalf("resolveBackupSpaceKey(global, default) = %q, want empty", got)
	}
}

func TestResolveBackupSpaceKeyDefaultsToDefaultSpace(t *testing.T) {
	if got := normalizeBackupScopeType("", ""); got != "global" {
		t.Fatalf("normalizeBackupScopeType(empty, empty) = %q, want global", got)
	}
	if got := normalizeBackupScopeType("space", ""); got != "space" {
		t.Fatalf("normalizeBackupScopeType(space, empty) = %q, want space", got)
	}
	if got := resolveBackupSpaceKey("space", ""); got != spaceutil.DefaultMenuSpaceKey {
		t.Fatalf(
			"resolveBackupSpaceKey(space, empty) = %q, want %q",
			got,
			spaceutil.DefaultMenuSpaceKey,
		)
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
