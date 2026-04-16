package permission

import (
	"testing"

	"github.com/google/uuid"

	"github.com/maben/backend/internal/modules/system/user"
)

func TestBuildPermissionAuditProfileForUnusedKey(t *testing.T) {
	item := user.PermissionKey{
		ID:            uuid.New(),
		PermissionKey: "system.page.sync",
	}

	profile := buildPermissionAuditProfile(item, permissionUsageCounters{}, permissionDuplicateProfile{})
	if profile.UsagePattern != permissionUsagePatternUnused {
		t.Fatalf("usage pattern = %q, want %q", profile.UsagePattern, permissionUsagePatternUnused)
	}
	if profile.UsageNote != "未被 API、页面或功能包消费" {
		t.Fatalf("usage note = %q", profile.UsageNote)
	}
}

func TestBuildPermissionAuditProfileForMultiConsumerKey(t *testing.T) {
	item := user.PermissionKey{
		ID:            uuid.New(),
		PermissionKey: "system.page.manage",
	}

	profile := buildPermissionAuditProfile(item, permissionUsageCounters{
		APICount:  3,
		PageCount: 1,
	}, permissionDuplicateProfile{})
	if profile.UsagePattern != permissionUsagePatternMultiConsumer {
		t.Fatalf("usage pattern = %q, want %q", profile.UsagePattern, permissionUsagePatternMultiConsumer)
	}
	if profile.UsageNote != "API、页面复合消费" {
		t.Fatalf("usage note = %q", profile.UsageNote)
	}
}

func TestBuildPermissionDuplicateProfilesForCrossContextMirror(t *testing.T) {
	profiles := buildPermissionDuplicateProfiles([]permissionDuplicateSource{
		{ID: uuid.New(), PermissionKey: "message.manage", ContextType: "platform"},
		{ID: uuid.New(), PermissionKey: "collaboration_workspace.message.manage", ContextType: "collaboration"},
	})

	profile, ok := profiles["message.manage"]
	if !ok {
		t.Fatalf("missing duplicate profile for message.manage")
	}
	if profile.Pattern != permissionDuplicatePatternCrossContextMirror {
		t.Fatalf("pattern = %q, want %q", profile.Pattern, permissionDuplicatePatternCrossContextMirror)
	}
}

func TestBuildPermissionDuplicateProfilesForSameContextDuplicate(t *testing.T) {
	profiles := buildPermissionDuplicateProfiles([]permissionDuplicateSource{
		{ID: uuid.New(), PermissionKey: "system.report.manage", ContextType: "platform"},
		{ID: uuid.New(), PermissionKey: "personal.report.manage", ContextType: "personal"},
	})

	profile, ok := profiles["system.report.manage"]
	if !ok {
		t.Fatalf("missing duplicate profile for system.report.manage")
	}
	if profile.Pattern != permissionDuplicatePatternSuspectedDuplicate {
		t.Fatalf("pattern = %q, want %q", profile.Pattern, permissionDuplicatePatternSuspectedDuplicate)
	}
}

