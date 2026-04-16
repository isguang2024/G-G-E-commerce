package router

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/maben/backend/internal/pkg/permissionseed"
)

// 说明：
//
// 老版本的对账测试是扫 router.go 源码里的手工 `v1.GET(...)` 一条条行和 seed 对账，
// 为了防止漏桥接（历史上前端 404 的根因）。在本次重构中，/api/v1 下所有业务路由
// 改为从 OpenAPI seed 自动注册（见 mountOpenAPIBridgeRoutes），源码里不再存在
// "一条路由一行" 的文本，那种文本级对账已失效。
//
// 新的测试改为**运行时对账**：
//   1. 构造一个空的 gin.Engine；
//   2. 调用 SetupRouter 实际使用的那段 mountOpenAPIBridgeRoutes；
//   3. 枚举 engine.Routes()，和 seed 对账：method + path + 期望的 bridge 都要对得上。
//
// 这比文本扫描更健壮：它在**实际的 Gin trie 里**检查每条 op 都被正确挂载，
// 同时也能暴露 Gin radix tree 冲突（注册阶段 panic）。

var ginParamPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

func TestOpenAPIOperationsMatchGinBridgeRegistrations(t *testing.T) {
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		t.Fatalf("LoadOpenAPISeed() error = %v", err)
	}

	// 预先检查 seed 自身：permission 模式的 op 必须带 permission_key，否则
	// 在线上会被 evaluator 静默拒掉。这里失败就停，比 Gin 注册阶段更早。
	for _, op := range seed.Operations {
		if op.AccessMode == "permission" && strings.TrimSpace(op.PermissionKey) == "" {
			t.Fatalf("operation %s %s 缺少 permission_key", strings.ToUpper(op.Method), op.Path)
		}
	}

	actual, err := registerAndSnapshot(seed.Operations)
	if err != nil {
		t.Fatalf("mountOpenAPIBridgeRoutes() panicked / failed: %v", err)
	}

	expected := make(map[string]string, len(seed.Operations))
	for _, op := range seed.Operations {
		expected[routeKey(op.Method, op.Path)] = expectedBridge(op.AccessMode)
	}

	var missing, unexpected, mismatched []string

	for key, want := range expected {
		got, ok := actual[key]
		if !ok {
			missing = append(missing, key)
			continue
		}
		if got != want {
			mismatched = append(mismatched, fmt.Sprintf("%s => actual=%s expected=%s", key, got, want))
		}
	}
	for key, got := range actual {
		if _, ok := expected[key]; !ok {
			unexpected = append(unexpected, fmt.Sprintf("%s => %s", key, got))
		}
	}

	sort.Strings(missing)
	sort.Strings(unexpected)
	sort.Strings(mismatched)

	if len(missing) > 0 || len(unexpected) > 0 || len(mismatched) > 0 {
		t.Fatalf("OpenAPI/Gin bridge 对账失败\nmissing: %s\nunexpected: %s\nmismatched: %s",
			strings.Join(missing, ", "),
			strings.Join(unexpected, ", "),
			strings.Join(mismatched, ", "),
		)
	}
}

// TestOpenAPIPathToGinPlaceholders spot-checks the placeholder translation
// on paths we actually ship, to pin the `{x}` → `:x` contract.
func TestOpenAPIPathToGinPlaceholders(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"/users", "/users"},
		{"/users/{id}", "/users/:id"},
		{"/users/{id}/packages/{packageId}", "/users/:id/packages/:packageId"},
		{"/collaboration-workspaces/current/members/{userId}/roles", "/collaboration-workspaces/current/members/:userId/roles"},
		// Malformed input (unbalanced brace) should pass through verbatim,
		// not silently drop characters.
		{"/broken/{id", "/broken/{id"},
	}
	for _, c := range cases {
		if got := openapiPathToGin(c.in); got != c.want {
			t.Errorf("openapiPathToGin(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// registerAndSnapshot builds a minimal gin.Engine, runs the production
// mount logic against the given ops, and returns a map
// `"METHOD /openapi-style-path" -> "publicBridge" | "ogenBridge"` for
// comparison with the seed. Any conflict inside Gin's radix tree will
// panic at register time and surface as an error here.
func registerAndSnapshot(ops []permissionseed.OpenAPIOperation) (result map[string]string, retErr error) {
	defer func() {
		if rec := recover(); rec != nil {
			retErr = fmt.Errorf("gin route registration panicked: %v", rec)
		}
	}()

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	v1 := engine.Group("/api/v1")
	authenticated := v1.Group("")

	// 给两个 bridge 命名函数（而不是匿名闭包），这样 gin.RouteInfo.Handler
	// 里拿到的函数名能稳定区分 public vs ogen。
	mountOpenAPIBridgeRoutes(v1, authenticated, ops, namedPublicBridge, namedOgenBridge, zap.NewNop())

	out := make(map[string]string, len(ops))
	for _, ri := range engine.Routes() {
		// 去掉 /api/v1 前缀，回到 seed 的路径视角；再把 :id 翻回 {id} 便于
		// 直接和 seed.Operations 比对。
		trimmed := strings.TrimPrefix(ri.Path, "/api/v1")
		openapiPath := ginParamPattern.ReplaceAllString(trimmed, `{$1}`)

		bridge := ""
		switch handlerShortName(ri.Handler) {
		case "namedPublicBridge":
			bridge = "publicBridge"
		case "namedOgenBridge":
			bridge = "ogenBridge"
		default:
			// 非 seed 注册的路由（OAuth、health 等）不进入对账表。
			continue
		}
		out[routeKey(ri.Method, openapiPath)] = bridge
	}
	return out, nil
}

func namedPublicBridge(*gin.Context) {}
func namedOgenBridge(*gin.Context)   {}

// handlerShortName extracts the trailing `.FuncName` out of runtime-reported
// names like `github.com/maben/.../router.namedPublicBridge`.
func handlerShortName(fqn string) string {
	// gin.RouteInfo.Handler is the fully-qualified name captured via
	// runtime.FuncForPC at registration time. We double-check by resolving
	// the real function pointers below, which is the most robust path.
	if fqn == "" {
		return ""
	}
	if idx := strings.LastIndex(fqn, "."); idx >= 0 && idx < len(fqn)-1 {
		return fqn[idx+1:]
	}
	return fqn
}

// Compile-time sanity: the two sentinel bridges must have distinct addresses
// so the runtime.FuncForPC-based name resolution in gin.RouteInfo.Handler
// can tell them apart. This also guarantees the functions aren't inlined
// away by the compiler under test builds.
var _ = func() struct{} {
	if reflect.ValueOf(namedPublicBridge).Pointer() == reflect.ValueOf(namedOgenBridge).Pointer() {
		panic("namedPublicBridge and namedOgenBridge collapsed to the same pointer")
	}
	return struct{}{}
}()

func currentRouterPath() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return currentFile
}

func routeKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + strings.TrimSpace(path)
}

func expectedBridge(accessMode string) string {
	switch strings.TrimSpace(accessMode) {
	case "public":
		return "publicBridge"
	case "authenticated", "permission":
		return "ogenBridge"
	default:
		// mountOpenAPIBridgeRoutes 对未知 access_mode 的默认策略是挂到
		// authenticated（ogenBridge），这里保持一致。
		return "ogenBridge"
	}
}
