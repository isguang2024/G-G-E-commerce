package router

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/gg-ecommerce/backend/internal/pkg/permissionseed"
)

var (
	bridgeRoutePattern = regexp.MustCompile(`(?m)^\s*(v1|authenticated)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)",\s*(publicBridge|ogenBridge)\)`)
	ginParamPattern    = regexp.MustCompile(`:([A-Za-z0-9_]+)`)
)

func TestOpenAPIOperationsMatchGinBridgeRegistrations(t *testing.T) {
	seed, err := permissionseed.LoadOpenAPISeed()
	if err != nil {
		t.Fatalf("LoadOpenAPISeed() error = %v", err)
	}

	actual, err := loadGinBridgeRoutes()
	if err != nil {
		t.Fatalf("loadGinBridgeRoutes() error = %v", err)
	}

	expected := make(map[string]string, len(seed.Operations))
	for _, op := range seed.Operations {
		if op.AccessMode == "permission" && strings.TrimSpace(op.PermissionKey) == "" {
			t.Fatalf("operation %s %s 缺少 permission_key", strings.ToUpper(op.Method), op.Path)
		}
		expected[routeKey(op.Method, op.Path)] = expectedBridge(op.AccessMode)
	}

	var missing []string
	var unexpected []string
	var mismatched []string

	for key, expectedBridgeName := range expected {
		actualBridgeName, ok := actual[key]
		if !ok {
			missing = append(missing, key)
			continue
		}
		if actualBridgeName != expectedBridgeName {
			mismatched = append(mismatched, fmt.Sprintf("%s => actual=%s expected=%s", key, actualBridgeName, expectedBridgeName))
		}
	}

	for key, actualBridgeName := range actual {
		if _, ok := expected[key]; !ok {
			unexpected = append(unexpected, fmt.Sprintf("%s => %s", key, actualBridgeName))
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

func loadGinBridgeRoutes() (map[string]string, error) {
	routerPath, err := currentRouterPath()
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(routerPath)
	if err != nil {
		return nil, fmt.Errorf("read router.go: %w", err)
	}

	matches := bridgeRoutePattern.FindAllStringSubmatch(string(content), -1)
	routes := make(map[string]string, len(matches))
	for _, match := range matches {
		method := match[2]
		path := normalizeGinPath(match[3])
		bridgeName := match[4]
		routes[routeKey(method, path)] = bridgeName
	}
	return routes, nil
}

func currentRouterPath() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve current file failed")
	}
	return filepath.Join(filepath.Dir(currentFile), "router.go"), nil
}

func normalizeGinPath(path string) string {
	normalized := ginParamPattern.ReplaceAllString(path, `{$1}`)
	return strings.TrimSpace(normalized)
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
		return ""
	}
}
