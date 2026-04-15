package audit

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

// TestRecorderStats_AsyncDropsWhenQueueFull 验证异步 channel 满时 dropped 正确累加、
// accepted 只计成功入队的那一条，且 QueueDepth / QueueCap 与实际 chan 对齐。
//
// 故意不启动 workers，让 channel 塞满后无人消费，从而触发 default 分支。
// 这避免了对真实 DB 的依赖，纯粹测 Stats() 指标采集逻辑。
func TestRecorderStats_AsyncDropsWhenQueueFull(t *testing.T) {
	s := &service{
		log:      zap.NewNop(),
		cfg:      Config{Enabled: true, AsyncMode: true, QueueSize: 1, Workers: 0},
		redactor: newRedactor(DefaultRedactFields),
		queue:    make(chan *AuditLog, 1),
		stopCh:   make(chan struct{}),
	}
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		s.Record(ctx, Event{Action: "test.action"})
	}
	got := s.Stats()
	if got.QueueCap != 1 {
		t.Fatalf("QueueCap = %d, want 1", got.QueueCap)
	}
	if got.QueueDepth != 1 {
		t.Fatalf("QueueDepth = %d, want 1", got.QueueDepth)
	}
	if got.AcceptedTotal != 1 {
		t.Fatalf("AcceptedTotal = %d, want 1", got.AcceptedTotal)
	}
	if got.DroppedTotal != 4 {
		t.Fatalf("DroppedTotal = %d, want 4", got.DroppedTotal)
	}
}

// TestRecorderStats_BuildFailureDoesNotCount 验证 build 失败（action 为空）既不算 accepted 也不算 dropped。
// 这保证指标不会被「坏入参」污染，对齐注释里"dropped 只记 channel 满"的语义。
func TestRecorderStats_BuildFailureDoesNotCount(t *testing.T) {
	s := &service{
		log:      zap.NewNop(),
		cfg:      Config{Enabled: true, AsyncMode: true, QueueSize: 8, Workers: 0},
		redactor: newRedactor(DefaultRedactFields),
		queue:    make(chan *AuditLog, 8),
		stopCh:   make(chan struct{}),
	}
	s.Record(context.Background(), Event{}) // action 缺失 → build 失败

	got := s.Stats()
	if got.AcceptedTotal != 0 || got.DroppedTotal != 0 {
		t.Fatalf("build-failure Record should not move counters, got %+v", got)
	}
}

// TestRecorderStats_NoopReturnsZero 验证 Noop 实现 Stats() 返回零值结构。
// 线上 audit.Enabled=false / db=nil 时应用会持有 Noop，暴露 /metrics 的 handler
// 必须能安全调用 Stats() 而不 panic。
func TestRecorderStats_NoopReturnsZero(t *testing.T) {
	var r Recorder = Noop{}
	got := r.Stats()
	if got != (Stats{}) {
		t.Fatalf("Noop.Stats() = %+v, want zero value", got)
	}
}
