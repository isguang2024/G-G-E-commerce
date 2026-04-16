package upload

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/maben/backend/internal/modules/system/models"
)

const (
	extraSchemaVersionV1 = "v1"

	configExtraScopeProvider = "provider.extra"
	configExtraScopeBucket   = "bucket.extra"
	configExtraScopeKey      = "upload_key.extra_schema"
	configExtraScopeRule     = "upload_rule.extra_schema"

	extraSchemaFieldTypeString  = "string"
	extraSchemaFieldTypeNumber  = "number"
	extraSchemaFieldTypeBoolean = "boolean"
	extraSchemaFieldTypeObject  = "object"
	extraSchemaFieldTypeSelect  = "select"
)

type DriverExtraField struct {
	Key          string
	Label        string
	Type         string
	Description  string
	Placeholder  string
	Required     bool
	DefaultValue any
}

type DriverExtraRegistry struct {
	Driver        string
	ProviderExtra []DriverExtraField
	BucketExtra   []DriverExtraField
}

type extraSchemaOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type extraSchemaField struct {
	Key          string              `json:"key"`
	Label        string              `json:"label"`
	Type         string              `json:"type"`
	Required     bool                `json:"required,omitempty"`
	Placeholder  string              `json:"placeholder,omitempty"`
	Description  string              `json:"description,omitempty"`
	DefaultValue any                 `json:"default_value,omitempty"`
	Options      []extraSchemaOption `json:"options,omitempty"`
}

type extraSchemaDocument struct {
	Version string             `json:"version"`
	Fields  []extraSchemaField `json:"fields,omitempty"`
}

var (
	allowedExtraSchemaFieldTypes = map[string]struct{}{
		extraSchemaFieldTypeString:  {},
		extraSchemaFieldTypeNumber:  {},
		extraSchemaFieldTypeBoolean: {},
		extraSchemaFieldTypeObject:  {},
		extraSchemaFieldTypeSelect:  {},
	}
	driverExtraRegistry = map[string]DriverExtraRegistry{
		models.UploadProviderDriverLocal: {
			Driver:        models.UploadProviderDriverLocal,
			ProviderExtra: []DriverExtraField{},
			BucketExtra:   []DriverExtraField{},
		},
		UploadProviderDriverAliyunOSS: {
			Driver: UploadProviderDriverAliyunOSS,
			ProviderExtra: []DriverExtraField{
				{Key: ossProviderExtraUseCNameKey, Label: "启用 CNAME", Type: extraSchemaFieldTypeBoolean, Description: "使用自定义域名访问 OSS", DefaultValue: false},
				{Key: ossProviderExtraUsePathStyleKey, Label: "启用 Path-Style", Type: extraSchemaFieldTypeBoolean, Description: "兼容部分网关或代理场景", DefaultValue: false},
				{Key: ossProviderExtraInsecureSkipVerifyKey, Label: "跳过 TLS 校验", Type: extraSchemaFieldTypeBoolean, Description: "仅用于内网测试或自签证书场景", DefaultValue: false},
				{Key: ossProviderExtraDisableSSLKey, Label: "禁用 HTTPS", Type: extraSchemaFieldTypeBoolean, Description: "改走 HTTP 访问 OSS", DefaultValue: false},
				{Key: ossProviderExtraConnectTimeoutMsKey, Label: "连接超时(ms)", Type: extraSchemaFieldTypeNumber, Description: "SDK 建连超时时间"},
				{Key: ossProviderExtraReadWriteTimeoutMsKey, Label: "读写超时(ms)", Type: extraSchemaFieldTypeNumber, Description: "SDK 上传下载超时时间"},
				{Key: ossProviderExtraRetryMaxAttemptsKey, Label: "最大重试次数", Type: extraSchemaFieldTypeNumber, Description: "SDK 请求重试上限"},
				{Key: ossProviderExtraSTSRoleARNKey, Label: "STS Role ARN", Type: extraSchemaFieldTypeString, Description: "开启 STS 临时凭证时必填"},
				{Key: ossProviderExtraSTSExternalIDKey, Label: "STS External ID", Type: extraSchemaFieldTypeString, Description: "按需配置的外部 ID"},
				{Key: ossProviderExtraSTSSessionNameKey, Label: "STS Session Name", Type: extraSchemaFieldTypeString, Description: "AssumeRole 会话名", DefaultValue: defaultOSSSTSRoleSessionName},
				{Key: ossProviderExtraSTSDurationSecondsKey, Label: "STS 凭证时长(秒)", Type: extraSchemaFieldTypeNumber, Description: "临时凭证有效期", DefaultValue: defaultOSSSTSRoleDurationSeconds},
				{Key: ossProviderExtraSTSPolicyKey, Label: "STS Policy", Type: extraSchemaFieldTypeObject, Description: "AssumeRole 附加策略 JSON"},
				{Key: ossProviderExtraSTSEndpointKey, Label: "STS Endpoint", Type: extraSchemaFieldTypeString, Description: "自定义 STS 接入点"},
			},
			BucketExtra: []DriverExtraField{
				{Key: ossBucketExtraSuccessStatusKey, Label: "直传成功状态码", Type: extraSchemaFieldTypeString, Description: "浏览器表单直传成功后返回的 HTTP 状态码", DefaultValue: defaultOSSDirectUploadSuccessAction},
				{Key: ossBucketExtraContentDispositionKey, Label: "Content-Disposition", Type: extraSchemaFieldTypeString, Description: "对象默认下载头"},
				{Key: ossBucketExtraCallbackKey, Label: "上传回调配置", Type: extraSchemaFieldTypeString, Description: "OSS 回调内容，通常是 base64 JSON"},
				{Key: ossBucketExtraCallbackVarKey, Label: "上传回调变量", Type: extraSchemaFieldTypeString, Description: "OSS 回调变量配置，通常是 base64 JSON"},
			},
		},
	}
	explicitFieldKeysByScope = map[string]map[string]struct{}{
		configExtraScopeKey: {
			"bucket_id":                   {},
			"key":                         {},
			"name":                        {},
			"path_template":               {},
			"default_rule_key":            {},
			"max_size_bytes":              {},
			"allowed_mime_types":          {},
			"upload_mode":                 {},
			"is_frontend_visible":         {},
			"permission_key":              {},
			"fallback_key":                {},
			"client_accept":               {},
			"direct_size_threshold_bytes": {},
			"visibility":                  {},
			"status":                      {},
			"meta":                        {},
		},
		configExtraScopeRule: {
			"rule_key":            {},
			"name":                {},
			"sub_path":            {},
			"filename_strategy":   {},
			"max_size_bytes":      {},
			"allowed_mime_types":  {},
			"process_pipeline":    {},
			"mode_override":       {},
			"visibility_override": {},
			"client_accept":       {},
			"is_default":          {},
			"status":              {},
			"meta":                {},
		},
	}
)

func LookupDriverExtraRegistry(driver string) (DriverExtraRegistry, bool) {
	item, ok := driverExtraRegistry[strings.TrimSpace(driver)]
	return item, ok
}

func ListDriverExtraRegistries() []DriverExtraRegistry {
	items := make([]DriverExtraRegistry, 0, len(driverExtraRegistry))
	for _, item := range driverExtraRegistry {
		items = append(items, item)
	}
	return items
}

func DriverExtraDefaults(driver, scope string) models.MetaJSON {
	registry, ok := LookupDriverExtraRegistry(driver)
	if !ok {
		return models.MetaJSON{}
	}
	var fields []DriverExtraField
	switch scope {
	case configExtraScopeProvider:
		fields = registry.ProviderExtra
	case configExtraScopeBucket:
		fields = registry.BucketExtra
	default:
		return models.MetaJSON{}
	}
	defaults := models.MetaJSON{}
	for _, field := range fields {
		if field.DefaultValue == nil {
			continue
		}
		defaults[field.Key] = field.DefaultValue
	}
	return defaults
}

func NormalizeUploadKeyExtraSchema(doc models.MetaJSON) (models.MetaJSON, error) {
	return normalizeExtraSchemaDocument(configExtraScopeKey, doc)
}

func NormalizeUploadRuleExtraSchema(doc models.MetaJSON) (models.MetaJSON, error) {
	return normalizeExtraSchemaDocument(configExtraScopeRule, doc)
}

func normalizeExtraSchemaDocument(scope string, doc models.MetaJSON) (models.MetaJSON, error) {
	if len(doc) == 0 {
		return models.MetaJSON{}, nil
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("%s marshal failed: %w", scope, err)
	}
	var parsed extraSchemaDocument
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return nil, fmt.Errorf("%s must be a JSON object: %w", scope, err)
	}
	parsed.Version = firstNonEmpty(strings.TrimSpace(parsed.Version), extraSchemaVersionV1)
	if parsed.Version != extraSchemaVersionV1 {
		return nil, fmt.Errorf("%s version %q is not supported", scope, parsed.Version)
	}
	seen := make(map[string]struct{}, len(parsed.Fields))
	for i := range parsed.Fields {
		field := &parsed.Fields[i]
		field.Key = strings.TrimSpace(field.Key)
		if field.Key == "" {
			return nil, fmt.Errorf("%s field key is required", scope)
		}
		if isExplicitFieldKey(scope, field.Key) {
			return nil, fmt.Errorf("%s field %q conflicts with explicit runtime field", scope, field.Key)
		}
		if _, exists := seen[field.Key]; exists {
			return nil, fmt.Errorf("%s field %q is duplicated", scope, field.Key)
		}
		seen[field.Key] = struct{}{}
		field.Label = firstNonEmpty(strings.TrimSpace(field.Label), field.Key)
		field.Type = firstNonEmpty(strings.TrimSpace(field.Type), extraSchemaFieldTypeString)
		if _, ok := allowedExtraSchemaFieldTypes[field.Type]; !ok {
			return nil, fmt.Errorf("%s field %q has unsupported type %q", scope, field.Key, field.Type)
		}
		field.Placeholder = strings.TrimSpace(field.Placeholder)
		field.Description = strings.TrimSpace(field.Description)
		if field.Type == extraSchemaFieldTypeSelect && len(field.Options) == 0 {
			return nil, fmt.Errorf("%s field %q requires options for select type", scope, field.Key)
		}
		for j := range field.Options {
			option := &field.Options[j]
			option.Value = strings.TrimSpace(option.Value)
			if option.Value == "" {
				return nil, fmt.Errorf("%s field %q has empty option value", scope, field.Key)
			}
			option.Label = firstNonEmpty(strings.TrimSpace(option.Label), option.Value)
		}
	}
	normalizedPayload, err := json.Marshal(parsed)
	if err != nil {
		return nil, fmt.Errorf("%s normalize failed: %w", scope, err)
	}
	var normalized models.MetaJSON
	if err := json.Unmarshal(normalizedPayload, &normalized); err != nil {
		return nil, fmt.Errorf("%s decode failed: %w", scope, err)
	}
	return normalized, nil
}

func isExplicitFieldKey(scope, key string) bool {
	reserved, ok := explicitFieldKeysByScope[scope]
	if !ok {
		return false
	}
	_, exists := reserved[strings.TrimSpace(key)]
	return exists
}
