package logger

import (
	"context"

	"go.uber.org/zap"
)

// ctxKey 是本包私有的 context key 类型。用一个命名 struct 而不是字符串，
// 是为了避免与其它包的 context.Value 冲突 —— Go 官方指引里强调
// context key 必须有明确类型而非裸字符串。
type ctxKey struct{ name string }

// 下面这些 key 是日志/审计/遥测链路共用的上下文键。中间件、handler、
// 后台 worker 都只能通过本包导出的 Setter/Getter 写入与读取，外部不可见。
var (
	requestIDKey   = ctxKey{"request_id"}
	actorIDKey     = ctxKey{"actor_id"}
	actorTypeKey   = ctxKey{"actor_type"}
	tenantIDKey    = ctxKey{"tenant_id"}
	appKeyKey      = ctxKey{"app_key"}
	workspaceIDKey = ctxKey{"workspace_id"}
	userAgentKey   = ctxKey{"user_agent"}
	clientIPKey    = ctxKey{"client_ip"}
)

// WithRequestID 把一个 request_id 注入 ctx。中间件 RequestID() 调用它。
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext 读取 ctx 中的 request_id；不存在时返回空串。
func RequestIDFromContext(ctx context.Context) string { return stringFrom(ctx, requestIDKey) }

// WithActor 把调用者身份 (id + type) 注入 ctx。type 通常为 user/api_key/system。
func WithActor(ctx context.Context, id, typ string) context.Context {
	if id != "" {
		ctx = context.WithValue(ctx, actorIDKey, id)
	}
	if typ != "" {
		ctx = context.WithValue(ctx, actorTypeKey, typ)
	}
	return ctx
}

// ActorIDFromContext 读取 ctx 中的 actor id；缺省返回空串。
func ActorIDFromContext(ctx context.Context) string { return stringFrom(ctx, actorIDKey) }

// ActorTypeFromContext 读取 ctx 中的 actor 类型。未显式设置时返回 "anonymous"。
func ActorTypeFromContext(ctx context.Context) string {
	if v := stringFrom(ctx, actorTypeKey); v != "" {
		return v
	}
	if stringFrom(ctx, actorIDKey) != "" {
		return "user"
	}
	return "anonymous"
}

// WithTenant 注入租户 ID。目前系统里始终是 "default"，但通道预留了多租户扩展点。
func WithTenant(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, tenantIDKey, id)
}

// TenantFromContext 读取租户 ID；未设置时返回 "default" 以贴合 DB 默认值。
func TenantFromContext(ctx context.Context) string {
	if v := stringFrom(ctx, tenantIDKey); v != "" {
		return v
	}
	return "default"
}

// WithApp 注入当前请求绑定的 app_key。
func WithApp(ctx context.Context, key string) context.Context {
	if key == "" {
		return ctx
	}
	return context.WithValue(ctx, appKeyKey, key)
}

// AppFromContext 读取 app_key；可能为空，调用方自行兜底。
func AppFromContext(ctx context.Context) string { return stringFrom(ctx, appKeyKey) }

// WithWorkspace 注入 workspace_id（协作空间 ID）。
func WithWorkspace(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, workspaceIDKey, id)
}

// WorkspaceFromContext 读取 workspace_id；可能为空。
func WorkspaceFromContext(ctx context.Context) string { return stringFrom(ctx, workspaceIDKey) }

// WithClient 注入客户端 IP + UA；一般在请求入口中间件统一调用。
func WithClient(ctx context.Context, ip, ua string) context.Context {
	if ip != "" {
		ctx = context.WithValue(ctx, clientIPKey, ip)
	}
	if ua != "" {
		ctx = context.WithValue(ctx, userAgentKey, ua)
	}
	return ctx
}

// ClientIPFromContext / UserAgentFromContext 返回对应取值；可能为空。
func ClientIPFromContext(ctx context.Context) string  { return stringFrom(ctx, clientIPKey) }
func UserAgentFromContext(ctx context.Context) string { return stringFrom(ctx, userAgentKey) }

func stringFrom(ctx context.Context, key ctxKey) string {
	if ctx == nil {
		return ""
	}
	if v, ok := ctx.Value(key).(string); ok {
		return v
	}
	return ""
}

// base 是全局根 logger；SetBase 在 main 启动时调用一次。尚未初始化时
// Base() 返回 zap.NewNop() 以防空指针。
var base *zap.Logger

// SetBase 挂载全局根 logger。
func SetBase(l *zap.Logger) {
	if l != nil {
		base = l
	}
}

// Base 返回全局根 logger；未初始化时返回 Nop，调用方无需判空。
func Base() *zap.Logger {
	if base == nil {
		return zap.NewNop()
	}
	return base
}

// With 基于 ctx 派生带结构化字段的子 logger —— 把 request_id / actor /
// tenant / app / workspace 等一次性填进去。约定：非空才追加，保持日志紧凑。
// 推荐 handler / service / 背景 goroutine 里统一使用 logger.With(ctx) 记录。
func With(ctx context.Context) *zap.Logger {
	l := Base()
	if ctx == nil {
		return l
	}
	fields := make([]zap.Field, 0, 6)
	if v := stringFrom(ctx, requestIDKey); v != "" {
		fields = append(fields, zap.String("request_id", v))
	}
	if v := stringFrom(ctx, actorIDKey); v != "" {
		fields = append(fields, zap.String("actor_id", v))
	}
	if v := stringFrom(ctx, tenantIDKey); v != "" {
		fields = append(fields, zap.String("tenant_id", v))
	}
	if v := stringFrom(ctx, appKeyKey); v != "" {
		fields = append(fields, zap.String("app_key", v))
	}
	if v := stringFrom(ctx, workspaceIDKey); v != "" {
		fields = append(fields, zap.String("workspace_id", v))
	}
	if len(fields) == 0 {
		return l
	}
	return l.With(fields...)
}

// FromContext 是 With 的语义别名。
func FromContext(ctx context.Context) *zap.Logger { return With(ctx) }
