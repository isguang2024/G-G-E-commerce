package audit

import "strings"

// DefaultRedactFields 列出审计 JSON payload 里默认要屏蔽的敏感字段。
// 调用方可以通过 config.audit.redact_fields 追加项目私有字段。
var DefaultRedactFields = []string{
	"password",
	"pwd",
	"secret",
	"token",
	"access_token",
	"refresh_token",
	"api_key",
	"authorization",
	"cookie",
	"credit_card",
	"card_no",
	"id_card",
	"phone",
	"mobile",
}

// redactor 把一组不区分大小写的敏感字段名做成 map，递归遍历 JSON 兼容结构
// （map[string]any / []any / 基础类型）时做替换。输入结构原地被修改。
type redactor struct {
	keys map[string]struct{}
}

// newRedactor 预处理字段集合，统一转小写以便大小写无关匹配。
func newRedactor(fields []string) *redactor {
	r := &redactor{keys: make(map[string]struct{}, len(fields))}
	for _, f := range fields {
		f = strings.ToLower(strings.TrimSpace(f))
		if f == "" {
			continue
		}
		r.keys[f] = struct{}{}
	}
	return r
}

// RedactorPublic 对外暴露的只读接口，供 observability 其他子包（telemetry）复用。
// 故意不暴露构造函数以外的实现细节，使用者把自己的 map/slice 传进来即可。
type RedactorPublic interface {
	Redact(v any) any
}

// NewPublicRedactor 创建一个可脱敏输入结构的 RedactorPublic。用于 telemetry 等
// 邻近模块需要同款规则时，避免复制粘贴 redact 逻辑。
func NewPublicRedactor(fields []string) RedactorPublic {
	return publicRedactor{inner: newRedactor(fields)}
}

type publicRedactor struct{ inner *redactor }

func (p publicRedactor) Redact(v any) any { return p.inner.redact(v) }

// redact 会在输入 value 上原地递归屏蔽敏感字段（map 键命中就替换为 "[REDACTED]"）。
// 返回处理后的 value，便于链式调用。slice / map 之外的类型原样返回。
func (r *redactor) redact(v any) any {
	if r == nil || v == nil {
		return v
	}
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if _, hit := r.keys[strings.ToLower(k)]; hit {
				if val == nil {
					continue
				}
				t[k] = "[REDACTED]"
				continue
			}
			t[k] = r.redact(val)
		}
		return t
	case []any:
		for i, val := range t {
			t[i] = r.redact(val)
		}
		return t
	default:
		return v
	}
}
