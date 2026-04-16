package role

import (
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/maben/backend/internal/modules/system/user"
)

func TestNormalizeAppKeysRemovesBlankAndDeduplicates(t *testing.T) {
	got := normalizeAppKeys([]string{
		" platform-admin ",
		"",
		"merchant-console",
		"platform-admin",
		"merchant-console ",
	})

	want := []string{"platform-admin", "merchant-console"}
	if len(got) != len(want) {
		t.Fatalf("len(normalizeAppKeys()) = %d, want %d; got=%v", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("normalizeAppKeys()[%d] = %q, want %q; got=%v", index, got[index], want[index], got)
		}
	}
}

func TestSameStringSetIgnoresOrderAndDuplicates(t *testing.T) {
	if !sameStringSet(
		[]string{"platform-admin", "merchant-console", "platform-admin"},
		[]string{"merchant-console", "platform-admin"},
	) {
		t.Fatal("sameStringSet() = false, want true")
	}
	if sameStringSet(
		[]string{"platform-admin"},
		[]string{"merchant-console"},
	) {
		t.Fatal("sameStringSet() = true, want false")
	}
}

func TestEnsureRoleEffectiveInAppAllowsGlobalRole(t *testing.T) {
	service := &roleService{logger: zap.NewNop()}
	role := &user.Role{
		ID:      uuid.New(),
		AppKeys: nil,
	}

	if err := service.ensureRoleEffectiveInApp(role, "platform-admin"); err != nil {
		t.Fatalf("ensureRoleEffectiveInApp(global) error = %v, want nil", err)
	}
}

func TestEnsureRoleEffectiveInAppAllowsScopedRoleInMatchedApp(t *testing.T) {
	service := &roleService{logger: zap.NewNop()}
	role := &user.Role{
		ID:      uuid.New(),
		AppKeys: []string{"platform-admin", "merchant-console"},
	}

	if err := service.ensureRoleEffectiveInApp(role, " merchant-console "); err != nil {
		t.Fatalf("ensureRoleEffectiveInApp(matched scoped role) error = %v, want nil", err)
	}
}

func TestEnsureRoleEffectiveInAppRejectsOutOfScopeApp(t *testing.T) {
	service := &roleService{logger: zap.NewNop()}
	role := &user.Role{
		ID:      uuid.New(),
		AppKeys: []string{"platform-admin"},
	}

	err := service.ensureRoleEffectiveInApp(role, "merchant-console")
	if err != ErrRoleAppScopeMismatch {
		t.Fatalf("ensureRoleEffectiveInApp(out of scope) error = %v, want %v", err, ErrRoleAppScopeMismatch)
	}
}

