package middleware

import (
	"testing"

	"github.com/maben/backend/internal/pkg/permissionseed"
)

// TestResolvePermissionKey_PrimaryPathEndpointCode 主路径必须走 endpoint_code：
// 给一个 operation_id 对应的 endpoint_code 在 ByEndpointCode 里有命中时，
// 不得走 fallback，且返回的 viaFallback=false。这是 P2-2 的核心契约。
func TestResolvePermissionKey_PrimaryPathEndpointCode(t *testing.T) {
	lookup := &permissionseed.PermissionLookup{
		ByEndpointCode: map[string]string{
			"code-users-list": "user.read",
		},
		OperationToEndpointCode: map[string]string{
			"listUsers": "code-users-list",
		},
		// 故意放一个不同的 permission_key 到 deprecated map 里：如果主路径
		// 正确走 endpoint_code，则 fallback 这个值不会被返回。
		ByOperationIDDeprecated: map[string]string{
			"listUsers": "WRONG.should.never.be.returned",
		},
	}
	key, code, viaFallback := resolvePermissionKey(lookup, "listUsers")
	if key != "user.read" {
		t.Fatalf("expected primary key 'user.read', got %q", key)
	}
	if code != "code-users-list" {
		t.Fatalf("expected endpoint_code 'code-users-list', got %q", code)
	}
	if viaFallback {
		t.Fatalf("main path must not trigger fallback")
	}
}

// TestResolvePermissionKey_FallbackWhenBindingMissing：ByEndpointCode 没有命中
// 但 ByOperationIDDeprecated 有 —— 模拟 bindings 与 seed 暂时不同步的情况。
// 应当返回 fallback 的 key 并标记 viaFallback=true 让 middleware 记 warn。
func TestResolvePermissionKey_FallbackWhenBindingMissing(t *testing.T) {
	lookup := &permissionseed.PermissionLookup{
		ByEndpointCode:          map[string]string{}, // 主路径缺失
		OperationToEndpointCode: map[string]string{"listUsers": "code-users-list"},
		ByOperationIDDeprecated: map[string]string{"listUsers": "user.read"},
	}
	key, code, viaFallback := resolvePermissionKey(lookup, "listUsers")
	if key != "user.read" {
		t.Fatalf("expected fallback to return 'user.read', got %q", key)
	}
	if code != "code-users-list" {
		t.Fatalf("expected endpoint_code still resolved for logging, got %q", code)
	}
	if !viaFallback {
		t.Fatalf("expected viaFallback=true when primary miss + fallback hit")
	}
}

// TestResolvePermissionKey_RenameOperationIDKeepsPermission：端点 method+path
// 不变，operation_id 被重命名。ByEndpointCode 仍然以 endpoint_code（method+path
// 派生）为 key，因此 seed 重建后主路径仍命中，不出现权限失控。
func TestResolvePermissionKey_RenameOperationIDKeepsPermission(t *testing.T) {
	// 原 operation_id = "listUsers"，重命名后 = "getAllUsers"；
	// endpoint_code（method+path 派生）保持不变。
	lookup := &permissionseed.PermissionLookup{
		ByEndpointCode: map[string]string{
			"code-users-list": "user.read",
		},
		OperationToEndpointCode: map[string]string{
			"getAllUsers": "code-users-list", // seed 重建后已是新名
		},
		ByOperationIDDeprecated: map[string]string{
			"getAllUsers": "user.read", // 同样以新名索引
		},
	}
	key, _, viaFallback := resolvePermissionKey(lookup, "getAllUsers")
	if key != "user.read" {
		t.Fatalf("expected permission intact after operation_id rename, got %q", key)
	}
	if viaFallback {
		t.Fatalf("primary path must still hit after rename since endpoint_code is stable")
	}
}

// TestResolvePermissionKey_PublicOpNoKey：operation 未声明 permission_key
// （如 public / authenticated-only op），两条路径都应 miss；key 为空，
// middleware 应放行。
func TestResolvePermissionKey_PublicOpNoKey(t *testing.T) {
	lookup := &permissionseed.PermissionLookup{
		ByEndpointCode:          map[string]string{},
		OperationToEndpointCode: map[string]string{"healthCheck": "code-health"},
		ByOperationIDDeprecated: map[string]string{},
	}
	key, code, viaFallback := resolvePermissionKey(lookup, "healthCheck")
	if key != "" {
		t.Fatalf("expected empty key for public op, got %q", key)
	}
	if code != "code-health" {
		t.Fatalf("endpoint_code should still resolve for logging, got %q", code)
	}
	if viaFallback {
		t.Fatalf("no key anywhere must not flag fallback")
	}
}

// TestOpenAPISeedPermissionLookup_Derivation：跑通真实 seed，保证：
//   - 每条 op 的 OperationToEndpointCode 都能映回 ByEndpointCode（当 permission_key 非空）
//   - endpoint_code 与 P0-2 ensure 步骤 (StableID("openapi-api-endpoint", METHOD+" "+path)) 一致
func TestOpenAPISeedPermissionLookup_Derivation(t *testing.T) {
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		t.Fatalf("LoadOpenAPISeed: %v", err)
	}
	lookup := seed.PermissionLookup()

	if len(lookup.OperationToEndpointCode) == 0 {
		t.Fatalf("OperationToEndpointCode must not be empty")
	}

	for _, op := range seed.Operations {
		code := lookup.OperationToEndpointCode[op.OperationID]
		if code == "" {
			t.Fatalf("op %s: missing endpoint_code mapping", op.OperationID)
		}
		if op.PermissionKey == "" {
			// public / authenticated 无键，主路径应 miss。
			if _, ok := lookup.ByEndpointCode[code]; ok {
				t.Fatalf("op %s: no permission_key but ByEndpointCode has entry", op.OperationID)
			}
			continue
		}
		if got := lookup.ByEndpointCode[code]; got != op.PermissionKey {
			t.Fatalf("op %s: ByEndpointCode[%s]=%q, want %q", op.OperationID, code, got, op.PermissionKey)
		}
	}
}
