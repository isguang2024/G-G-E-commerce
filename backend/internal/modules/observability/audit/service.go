// Package audit 提供业务操作审计日志的写入服务。
//
// 设计要点：
//  1. 对外只暴露一个接口 Recorder + 一个事件结构体 Event，所有 handler / service
//     通过 Recorder.Record(ctx, Event{...}) 写入。ctx 里自动携带 request_id / actor /
//     tenant / app / workspace，无需在 Event 里重复填。
//  2. 持久化走异步 channel + worker 池，不阻塞请求；channel 满时降级为
//     "drop-newest + Warn 日志"，保证 DB 抖动不会撑爆内存。
//  3. 脱敏在入队前同步完成，避免敏感字段以明文进入内存 channel。
//  4. 对单元测试 / CI 友好：当 cfg.Enabled=false 或 DB 为 nil 时，退化为 Noop，
//     业务代码不需要用 if recorder != nil 判空。
package audit

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/gg-ecommerce/backend/internal/pkg/logger"
)

// Event 是一次待写入的审计事件。调用方只需要填业务关心的字段：
// Action（必填）、ResourceType、ResourceID、Outcome、ErrorCode、Before/After。
// request_id / actor / tenant / app / workspace / IP / UA 自动从 ctx 读取。
type Event struct {
	Action       string // 必填：领域.实体.动作，例如 "system.app.create"
	ResourceType string // 资源类型，例如 "app"、"role"
	ResourceID   string // 资源主键，字符串化
	Outcome      string // success | denied | error；为空时按 ErrorCode 自动判定
	ErrorCode    string // 对齐 apperr code；仅 Outcome != success 时有意义
	HTTPStatus   int    // 对应 HTTP 响应码，方便按状态聚合

	// Before / After 是操作前/后的结构化快照。会被 JSON 序列化并脱敏写入。
	// 建议只保留业务字段，避免把整条 GORM model 灌进去（含太多噪音列）。
	Before any
	After  any

	// Metadata 用于承载 Action 特有的扩展字段，例如 dispatch_id / send_count。
	Metadata map[string]any
}

// Recorder 是审计服务对外接口。
//
// Record 约定：永远不返回业务级错误（不阻塞 handler 主流程）；
// 内部失败（JSON 序列化、channel 满、DB 异常）只记 Warn/Error 日志。
// 调用方可以安全地忽略返回值。
type Recorder interface {
	Record(ctx context.Context, e Event)
	Shutdown(ctx context.Context) error
}

// Config 控制 audit service 的运行参数。通过 config.AuditConfig 加载。
type Config struct {
	Enabled       bool
	RedactFields  []string
	QueueSize     int // 异步 channel 缓冲，默认 1024
	Workers       int // 消费 goroutine 数量，默认 2
	FlushInterval time.Duration
	AsyncMode     bool // false = 同步写入（测试友好）；true = channel+worker
}

// DefaultConfig 提供生产默认值。对应 config.example.yaml 里的 audit 默认。
func DefaultConfig() Config {
	return Config{
		Enabled:       true,
		RedactFields:  DefaultRedactFields,
		QueueSize:     1024,
		Workers:       2,
		FlushInterval: time.Second,
		AsyncMode:     true,
	}
}

// service 是 Recorder 的默认实现。
type service struct {
	db       *gorm.DB
	log      *zap.Logger
	cfg      Config
	redactor *redactor

	mu       sync.Mutex
	queue    chan *AuditLog
	wg       sync.WaitGroup
	stopCh   chan struct{}
	shutdown bool
	dropped  uint64 // 累计丢弃数量（channel 满时）
}

// New 构建一个 Recorder。
//   - db 为 nil 或 cfg.Enabled=false 时，返回 Noop，业务侧继续调 Record 不报错；
//   - AsyncMode=true 时启动 workers 个后台消费 goroutine；
//   - 调用方 main 应在 shutdown 阶段 defer recorder.Shutdown(ctx)，保证 channel 里的
//     事件能被 drain 干净再退出。
func New(db *gorm.DB, log *zap.Logger, cfg Config) Recorder {
	if !cfg.Enabled || db == nil {
		return Noop{}
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if log == nil {
		log = zap.NewNop()
	}

	fields := cfg.RedactFields
	if len(fields) == 0 {
		fields = DefaultRedactFields
	}

	s := &service{
		db:       db,
		log:      log.Named("audit"),
		cfg:      cfg,
		redactor: newRedactor(fields),
		stopCh:   make(chan struct{}),
	}
	if cfg.AsyncMode {
		s.queue = make(chan *AuditLog, cfg.QueueSize)
		for i := 0; i < cfg.Workers; i++ {
			s.wg.Add(1)
			go s.runWorker()
		}
	}
	return s
}

// Record 实现 Recorder 接口。ctx / event 组合成 *AuditLog 并入队 / 同步写入。
func (s *service) Record(ctx context.Context, e Event) {
	row, err := s.build(ctx, e)
	if err != nil {
		s.log.Warn("audit.build_failed",
			zap.Error(err),
			zap.String("action", e.Action),
			zap.String("request_id", logger.RequestIDFromContext(ctx)),
		)
		return
	}
	if !s.cfg.AsyncMode {
		s.writeRow(ctx, row)
		return
	}
	select {
	case s.queue <- row:
	default:
		s.mu.Lock()
		s.dropped++
		dropped := s.dropped
		s.mu.Unlock()
		s.log.Warn("audit.queue_full_drop",
			zap.String("action", row.Action),
			zap.Uint64("dropped_total", dropped),
		)
	}
}

// Shutdown 在服务退出时调用，关闭 channel 并等待 workers 把剩余事件写完。
// 带超时 ctx，防止 DB 卡住导致进程不退出。
func (s *service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return nil
	}
	s.shutdown = true
	s.mu.Unlock()

	close(s.stopCh)
	if s.queue != nil {
		close(s.queue)
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return errors.New("audit: shutdown timeout; some events may be lost")
	}
}

func (s *service) runWorker() {
	defer s.wg.Done()
	for row := range s.queue {
		if row == nil {
			continue
		}
		// 用独立 ctx 写 DB，避免 request ctx 提前被取消导致丢审计。
		writeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.writeRow(writeCtx, row)
		cancel()
	}
}

func (s *service) writeRow(ctx context.Context, row *AuditLog) {
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		s.log.Error("audit.write_failed",
			zap.Error(err),
			zap.String("action", row.Action),
			zap.String("request_id", row.RequestID),
		)
	}
}

// build 将 Event + ctx 合成一条 AuditLog，完成：
//  1. 取 ctx 中的 request_id / actor / tenant / app / workspace / client info；
//  2. 对 Before/After/Metadata 做 JSON 序列化 + 敏感字段脱敏；
//  3. Outcome 缺省按 ErrorCode 推断。
func (s *service) build(ctx context.Context, e Event) (*AuditLog, error) {
	if e.Action == "" {
		return nil, errors.New("audit: action is required")
	}
	outcome := e.Outcome
	if outcome == "" {
		if e.ErrorCode != "" {
			outcome = OutcomeError
		} else {
			outcome = OutcomeSuccess
		}
	}

	before, err := s.marshalRedacted(e.Before)
	if err != nil {
		return nil, err
	}
	after, err := s.marshalRedacted(e.After)
	if err != nil {
		return nil, err
	}
	var metadata []byte
	if len(e.Metadata) > 0 {
		metadata, err = s.marshalRedactedMap(e.Metadata)
		if err != nil {
			return nil, err
		}
	} else {
		metadata = []byte(`{}`)
	}

	now := time.Now().UTC()
	row := &AuditLog{
		Ts:           now,
		RequestID:    logger.RequestIDFromContext(ctx),
		TenantID:     logger.TenantFromContext(ctx),
		ActorID:      logger.ActorIDFromContext(ctx),
		ActorType:    logger.ActorTypeFromContext(ctx),
		AppKey:       logger.AppFromContext(ctx),
		WorkspaceID:  logger.WorkspaceFromContext(ctx),
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		Outcome:      outcome,
		ErrorCode:    e.ErrorCode,
		HTTPStatus:   e.HTTPStatus,
		IP:           logger.ClientIPFromContext(ctx),
		UserAgent:    logger.UserAgentFromContext(ctx),
		BeforeJSON:   before,
		AfterJSON:    after,
		Metadata:     metadata,
		CreatedAt:    now,
	}
	return row, nil
}

// marshalRedacted 把任意结构 -> JSON bytes，然后读回 map 做脱敏，再序列化回去。
// nil 时写入字面量 null，DB 列默认 'null'::jsonb 对齐。
func (s *service) marshalRedacted(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}
	decoded = s.redactor.redact(decoded)
	return json.Marshal(decoded)
}

// marshalRedactedMap 是 marshalRedacted 针对顶层 map 的快捷版本。
func (s *service) marshalRedactedMap(m map[string]any) ([]byte, error) {
	copyMap := make(map[string]any, len(m))
	for k, v := range m {
		copyMap[k] = v
	}
	redacted := s.redactor.redact(copyMap)
	return json.Marshal(redacted)
}

// Noop 是一个显式的"什么都不做"的 Recorder，用于关闭审计 / 无 DB 环境。
// 特意放到包级可导出，方便测试 stub：audit.Noop{}。
type Noop struct{}

// Record ignores everything.
func (Noop) Record(_ context.Context, _ Event) {}

// Shutdown 总是立即返回 nil。
func (Noop) Shutdown(_ context.Context) error { return nil }
