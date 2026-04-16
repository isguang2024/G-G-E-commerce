package permissionseed

import (
	"errors"
	"strings"
	"testing"
)

// TestErrUnknownPermissionKeys_ErrorMessage 固定 ErrUnknownPermissionKeys
// 的错误格式 —— cmd/migrate 的 P0 拦截日志要能 grep 到该前缀，改格式得改这里。
func TestErrUnknownPermissionKeys_ErrorMessage(t *testing.T) {
	err := &ErrUnknownPermissionKeys{Keys: []string{"user.invented", "workspace.ghost"}}
	msg := err.Error()
	const wantPrefix = "openapi seed references permission_keys missing from permission_keys table: "
	if !strings.HasPrefix(msg, wantPrefix) {
		t.Fatalf("error prefix mismatch: got %q", msg)
	}
	if !strings.Contains(msg, "user.invented") || !strings.Contains(msg, "workspace.ghost") {
		t.Fatalf("error message must enumerate missing keys, got %q", msg)
	}
}

// TestErrUnknownPermissionKeys_IsErrorInterface：配合 errors.As 断言，以便
// cmd/migrate 或更高层可以按类型匹配而不是靠字符串。
func TestErrUnknownPermissionKeys_IsErrorInterface(t *testing.T) {
	var err error = &ErrUnknownPermissionKeys{Keys: []string{"a"}}
	var target *ErrUnknownPermissionKeys
	if !errors.As(err, &target) {
		t.Fatalf("errors.As must unwrap ErrUnknownPermissionKeys")
	}
	if len(target.Keys) != 1 || target.Keys[0] != "a" {
		t.Fatalf("errors.As lost Keys payload: %+v", target)
	}
}

// TestSortStrings_Ascending 验证本地 sortStrings 与 sort.Strings 等价
// —— findUnknownPermissionKeys 用它保证错误列表顺序稳定。
func TestSortStrings_Ascending(t *testing.T) {
	cases := [][]string{
		{"c", "a", "b"},
		{"workspace.read", "user.list", "user.create"},
		{},
		{"only"},
	}
	for _, in := range cases {
		cp := append([]string(nil), in...)
		sortStrings(cp)
		for i := 1; i < len(cp); i++ {
			if cp[i-1] > cp[i] {
				t.Fatalf("sortStrings not ascending: %v -> %v", in, cp)
			}
		}
	}
}
