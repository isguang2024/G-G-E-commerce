// observability_prometheus_test.go: 覆盖 GetObservabilityMetricsPrometheus 的
// 两条接受标准——「格式符合 openmetrics」与「Noop 全零仍可抓取」。
// 放在 handlers 包内（非 _test 包）以直接访问未导出 audit 字段，避免搭一套
// gin + evaluator 中间件脚手架仅为了验证一段字符串输出。
package handlers

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/gg-ecommerce/backend/api/gen"
	"github.com/gg-ecommerce/backend/internal/modules/observability/audit"
)

// TestGetObservabilityMetricsPrometheus_UnauthenticatedReturns401 —— 未登录上下文应返回 401。
// 验证 handler 不依赖 ogen 中间件独立兜底 auth。
func TestGetObservabilityMetricsPrometheus_UnauthenticatedReturns401(t *testing.T) {
	h := &APIHandler{audit: audit.Noop{}}
	res, err := h.GetObservabilityMetricsPrometheus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	unauth, ok := res.(*gen.GetObservabilityMetricsPrometheusUnauthorized)
	if !ok {
		t.Fatalf("want Unauthorized, got %T", res)
	}
	if unauth.Code != 401 {
		t.Errorf("Unauthorized.Code = %d, want 401", unauth.Code)
	}
}

// TestGetObservabilityMetricsPrometheus_NoopAllZeros —— Noop 模式 4 项指标全 0,
// 且 openmetrics 格式完整（每个指标有 # HELP + # TYPE + 数值行）。
func TestGetObservabilityMetricsPrometheus_NoopAllZeros(t *testing.T) {
	h := &APIHandler{audit: audit.Noop{}}
	ctx := context.WithValue(context.Background(), CtxUserID, uuid.New().String())

	res, err := h.GetObservabilityMetricsPrometheus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	okResp, ok := res.(*gen.GetObservabilityMetricsPrometheusOK)
	if !ok {
		t.Fatalf("want OK, got %T", res)
	}
	if okResp.Data == nil {
		t.Fatal("OK.Data reader is nil")
	}
	b, err := io.ReadAll(okResp.Data)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	body := string(b)

	// 4 项指标必须全部出现,且 HELP/TYPE/样本三行齐备。
	expected := []string{
		"# HELP audit_queue_depth",
		"# TYPE audit_queue_depth gauge",
		"audit_queue_depth 0",
		"# HELP audit_queue_capacity",
		"# TYPE audit_queue_capacity gauge",
		"audit_queue_capacity 0",
		"# HELP audit_events_accepted_total",
		"# TYPE audit_events_accepted_total counter",
		"audit_events_accepted_total 0",
		"# HELP audit_events_dropped_total",
		"# TYPE audit_events_dropped_total counter",
		"audit_events_dropped_total 0",
	}
	for _, needle := range expected {
		if !strings.Contains(body, needle) {
			t.Errorf("body missing %q; got:\n%s", needle, body)
		}
	}

	// openmetrics 要求每一行以 \n 结束 —— 粗略校验至少 12 行（4 指标 × 3）。
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(lines) < 12 {
		t.Fatalf("expected >= 12 lines, got %d: %v", len(lines), lines)
	}
	for i, ln := range lines {
		if strings.HasSuffix(ln, " ") {
			t.Errorf("line %d has trailing space: %q", i, ln)
		}
	}
}
