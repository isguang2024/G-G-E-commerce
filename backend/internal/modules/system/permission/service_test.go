package permission

import (
	"errors"
	"testing"
)

func TestDeriveContextTypeDefaultsToCommonForCustomKeys(t *testing.T) {
	if got := deriveContextType("product.order.audit", "product"); got != "common" {
		t.Fatalf("deriveContextType() = %q, want %q", got, "common")
	}
}

func TestValidatePermissionContext(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		permissionKey string
		moduleCode    string
		contextType   string
		wantErr       bool
	}{
		{
			name:          "builtin personal key rejects collaboration context",
			permissionKey: "message.manage",
			moduleCode:    "message",
			contextType:   "collaboration",
			wantErr:       true,
		},
		{
			name:          "platform custom key with reserved prefix passes",
			permissionKey: "personal.notice.dispatch",
			moduleCode:    "notice",
			contextType:   "platform",
			wantErr:       false,
		},
		{
			name:          "collaboration custom key with reserved prefix passes",
			permissionKey: "collaboration_workspace.notice.dispatch",
			moduleCode:    "notice",
			contextType:   "collaboration",
			wantErr:       false,
		},
		{
			name:          "common custom key passes",
			permissionKey: "product.notice.dispatch",
			moduleCode:    "product",
			contextType:   "common",
			wantErr:       false,
		},
		{
			name:          "common custom key cannot pretend to be platform",
			permissionKey: "product.notice.dispatch",
			moduleCode:    "product",
			contextType:   "platform",
			wantErr:       true,
		},
		{
			name:          "platform module cannot use common",
			permissionKey: "personal.menu.audit",
			moduleCode:    "menu",
			contextType:   "common",
			wantErr:       true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validatePermissionContext(tc.permissionKey, tc.moduleCode, tc.contextType)
			if tc.wantErr {
				if !errors.Is(err, ErrPermissionContextInvalid) {
					t.Fatalf("validatePermissionContext() error = %v, want ErrPermissionContextInvalid", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("validatePermissionContext() error = %v, want nil", err)
			}
		})
	}
}
