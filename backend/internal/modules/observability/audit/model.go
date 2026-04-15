package audit

import "time"

// AuditLog 是 audit_logs 表的 GORM 映射。所有字段采用零值友好策略：
// Postgres 列带 NOT NULL DEFAULT，Go 侧直接用非指针类型，减少模板化 NULL 判空。
//
// 约定：
//   - audit_logs 只追加，永不更新/删除（合规审计要求）。因此不加 UpdatedAt/DeletedAt；
//   - CreatedAt 列给 DB 判断写入时间；Ts 是业务事件发生时间（一般与 CreatedAt 相同，
//     但对于异步 worker，两者可能有微小漂移）。
//   - BeforeJSON/AfterJSON/Metadata 用 []byte 持有 jsonb，由 service.Record 序列化。
type AuditLog struct {
	ID           uint64    `gorm:"column:id;primaryKey"`
	Ts           time.Time `gorm:"column:ts;autoCreateTime"`
	RequestID    string    `gorm:"column:request_id"`
	TenantID     string    `gorm:"column:tenant_id"`
	ActorID      string    `gorm:"column:actor_id"`
	ActorType    string    `gorm:"column:actor_type"`
	AppKey       string    `gorm:"column:app_key"`
	WorkspaceID  string    `gorm:"column:workspace_id"`
	Action       string    `gorm:"column:action"`
	ResourceType string    `gorm:"column:resource_type"`
	ResourceID   string    `gorm:"column:resource_id"`
	Outcome      string    `gorm:"column:outcome"`
	ErrorCode    string    `gorm:"column:error_code"`
	HTTPStatus   int       `gorm:"column:http_status"`
	IP           string    `gorm:"column:ip"`
	UserAgent    string    `gorm:"column:user_agent"`
	BeforeJSON   []byte    `gorm:"column:before_json;type:jsonb"`
	AfterJSON    []byte    `gorm:"column:after_json;type:jsonb"`
	Metadata     []byte    `gorm:"column:metadata;type:jsonb"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

// TableName 固定表名，防止 GORM 按结构体名自动复数化猜错。
func (AuditLog) TableName() string { return "audit_logs" }

// 审计事件结果的取值常量。和 DB 列 outcome 的取值保持一致，调用方
// 应使用这些常量而不是魔法字符串。
const (
	OutcomeSuccess = "success"
	OutcomeDenied  = "denied"
	OutcomeError   = "error"
)

// 常见的 actor_type 取值。调用方可以写其它字符串，但这三个是主流取值。
const (
	ActorTypeUser      = "user"
	ActorTypeAPIKey    = "api_key"
	ActorTypeSystem    = "system"
	ActorTypeAnonymous = "anonymous"
)
