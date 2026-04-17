package router

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
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
//   2. 从生产 newOgenBridges 取到同源 bridge（不再在测试里自己实现一份）；
//   3. 调用 SetupRouter 实际使用的那段 mountOpenAPIBridgeRoutes；
//   4. 枚举 engine.Routes()，按函数指针识别桥接类型，和 seed 对账：
//      method + path + 期望的 bridge 都要对得上。
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
		{"/collaboration/current/members/{userId}/roles", "/collaboration/current/members/:userId/roles"},
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

// TestNewOgenBridgesProducesDistinctHandlers 保护 newOgenBridges 的核心契约：
// 返回的两个 handler 必须是不同的函数指针，否则 registerAndSnapshot 无法
// 区分 public 与 authenticated 挂载，整个对账测试就会退化成空检查。
func TestNewOgenBridgesProducesDistinctHandlers(t *testing.T) {
	pub, ogen := newOgenBridges(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	if pub == nil || ogen == nil {
		t.Fatalf("newOgenBridges returned nil handler(s): public=%v ogen=%v", pub == nil, ogen == nil)
	}
	if reflect.ValueOf(pub).Pointer() == reflect.ValueOf(ogen).Pointer() {
		t.Fatalf("newOgenBridges collapsed public/ogen into the same function pointer")
	}
}

// registerAndSnapshot builds a minimal gin.Engine, runs the production
// mount logic against the given ops, and returns a map
// `"METHOD /openapi-style-path" -> "publicBridge" | "ogenBridge"` for
// comparison with the seed. Any conflict inside Gin's radix tree will
// panic at register time and surface as an error here.
//
// 桥接函数来源于生产的 newOgenBridges —— 这是对 "测试和生产 bridge 不漂移"
// 的结构化保障：如果生产代码改用别的 bridge 构造路径绕过 newOgenBridges，
// 测试里 HandlerFunc 指针匹配不上，会整体报 missing。
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

	// 关键：从生产工厂取 bridge，确保和 SetupRouter 跑的是同一段代码。
	// ogen 底层 http.Handler 在注册阶段不会被调用，传一个 no-op 即可。
	publicBridge, ogenBridge := newOgenBridges(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	publicPtr := reflect.ValueOf(publicBridge).Pointer()
	ogenPtr := reflect.ValueOf(ogenBridge).Pointer()

	mountOpenAPIBridgeRoutes(v1, authenticated, ops, publicBridge, ogenBridge, zap.NewNop())

	out := make(map[string]string, len(ops))
	for _, ri := range engine.Routes() {
		// 去掉 /api/v1 前缀，回到 seed 的路径视角；再把 :id 翻回 {id} 便于
		// 直接和 seed.Operations 比对。
		trimmed := strings.TrimPrefix(ri.Path, "/api/v1")
		openapiPath := ginParamPattern.ReplaceAllString(trimmed, `{$1}`)

		bridge := ""
		handlerPtr := reflect.ValueOf(ri.HandlerFunc).Pointer()
		switch handlerPtr {
		case publicPtr:
			bridge = "publicBridge"
		case ogenPtr:
			bridge = "ogenBridge"
		default:
			// 非 seed 注册的路由（OAuth、health 等）不进入对账表。
			continue
		}
		out[routeKey(ri.Method, openapiPath)] = bridge
	}
	return out, nil
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
		// mountOpenAPIBridgeRoutes 现在对未知 access_mode 直接 logger.Fatal，
		// 不会产出任何路由；这里返回 "unknown" 保证断言显式失败而非被兜底。
		return "unknown"
	}
}
