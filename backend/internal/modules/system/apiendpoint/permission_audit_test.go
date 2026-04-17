package apiendpoint

import "testing"

func TestBuildPermissionProfileForJWTEndpoint(t *testing.T) {
	profile := buildPermissionProfile("GET", "/api/v1/messages/inbox", nil, nil)
	if profile.BindingMode != permissionPatternSelfJWT {
		t.Fatalf("binding mode = %q, want %q", profile.BindingMode, permissionPatternSelfJWT)
	}
	if profile.Note != "登录态自服务接口，无需权限键" {
		t.Fatalf("note = %q", profile.Note)
	}
}

func TestBuildPermissionProfileForGlobalJWTEndpoint(t *testing.T) {
	profile := buildPermissionProfile("GET", "/api/v1/runtime/navigation", nil, nil)
	if profile.BindingMode != permissionPatternGlobalJWT {
		t.Fatalf("binding mode = %q, want %q", profile.BindingMode, permissionPatternGlobalJWT)
	}
	if profile.Note != "登录态全局接口，无需权限键" {
		t.Fatalf("note = %q", profile.Note)
	}
}

func TestBuildPermissionProfileForCrossContextSharedEndpoint(t *testing.T) {
	profile := buildPermissionProfile("POST", "/api/v1/messages/dispatch", []string{"message.manage", "collaboration.message.manage"}, nil)
	if profile.BindingMode != permissionPatternCrossContextShared {
		t.Fatalf("binding mode = %q, want %q", profile.BindingMode, permissionPatternCrossContextShared)
	}
	if !profile.SharedAcrossContexts {
		t.Fatalf("expected shared across contexts")
	}
	if len(profile.Contexts) != 2 {
		t.Fatalf("contexts len = %d, want 2", len(profile.Contexts))
	}
}
