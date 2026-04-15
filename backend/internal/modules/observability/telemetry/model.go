package telemetry

import "time"

// TelemetryLog 是 telemetry_logs 表的 GORM 映射。
// 字段与 docs/guides/logging-spec.md §4 里的前端契约对齐，只保留查询/观测
// 必需的列；其余细节统一放进 payload jsonb，保留前端上报时的原貌。
//
// 和 audit_logs 一样是 append-only：运维侧可以 TRUNCATE / 按分区清理，但应用
// 层不提供更新/删除入口。
type TelemetryLog struct {
	ID        uint64    `gorm:"column:id;primaryKey"`
	Ts        time.Time `gorm:"column:ts;autoCreateTime"`
	RequestID string    `gorm:"column:request_id"`
	SessionID string    `gorm:"column:session_id"`
	TenantID  string    `gorm:"column:tenant_id"`
	ActorID   string    `gorm:"column:actor_id"`
	AppKey    string    `gorm:"column:app_key"`
	Level     string    `gorm:"column:level"`
	Event     string    `gorm:"column:event"`
	Message   string    `gorm:"column:message"`
	URL       string    `gorm:"column:url"`
	UserAgent string    `gorm:"column:user_agent"`
	IP        string    `gorm:"column:ip"`
	Release   string    `gorm:"column:release"`
	Payload   []byte    `gorm:"column:payload;type:jsonb"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 固定表名，防止 GORM 按结构体名自动复数化猜错。
func (TelemetryLog) TableName() string { return "telemetry_logs" }

// 合法 level 值；和前端 LogLevel 同构。超出枚举的字段会在入库前被归并到 "info"，
// 避免前端任意值把 DB 列炸开（列长度 16）。
const (
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
)
