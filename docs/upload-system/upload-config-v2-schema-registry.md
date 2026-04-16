# 上传配置中心二期：扩展参数注册与持久化约束

## 1. 目标

- 明确哪些参数必须走显式字段，哪些参数允许进入 `extra / extra_schema`。
- 给不同 driver 的扩展参数提供统一注册入口，避免页面和 driver 各自维护一套键名。
- 给 `upload_keys.extra_schema` / `upload_key_rules.extra_schema` 提供最小结构校验，避免写成不可消费的 JSON 黑箱。

## 2. 边界定义

### 2.1 显式字段

以下字段属于运行时主链路语义，禁止再挪进 `extra_schema.fields[].key`：

- UploadKey：
  - `upload_mode`
  - `is_frontend_visible`
  - `permission_key`
  - `fallback_key`
  - `client_accept`
  - `direct_size_threshold_bytes`
  - 以及已有主字段：`path_template`、`visibility`、`allowed_mime_types` 等
- UploadKeyRule：
  - `mode_override`
  - `visibility_override`
  - `client_accept`
  - 以及已有主字段：`filename_strategy`、`allowed_mime_types`、`is_default` 等

这些字段由后端 `EffectiveConfig`、权限过滤和运行时 prepare/complete 逻辑直接消费，必须保持强类型与稳定语义。

### 2.2 `extra`

- `storage_providers.extra`
  - 存 driver 级实际配置值，例如阿里云 OSS 的 `sts_role_arn`、`use_cname`
- `storage_buckets.extra`
  - 存 bucket 级实际配置值，例如 `success_action_status`、`callback`

`extra` 是真实值存储槽，不负责描述 UI。

### 2.3 `extra_schema`

- `upload_keys.extra_schema`
- `upload_key_rules.extra_schema`

当前版本中，`extra_schema` 负责描述“自定义参数定义”，供管理面展示说明、基础校验和后续动态表单扩展使用；不承载运行时主链路关键字段。

## 3. 后端注册表

代码入口：

- `backend/internal/modules/system/upload/schema_registry.go`

当前提供两类能力：

1. driver 扩展参数注册表
   - `LookupDriverExtraRegistry(driver)`
   - `DriverExtraDefaults(driver, scope)`
2. `extra_schema` 结构归一化与校验
   - `NormalizeUploadKeyExtraSchema`
   - `NormalizeUploadRuleExtraSchema`

## 4. 当前内置 driver 注册内容

### 4.1 local

- `provider.extra`：无内置扩展键
- `bucket.extra`：无内置扩展键

### 4.2 aliyun_oss

- `provider.extra`
  - `use_cname`
  - `use_path_style`
  - `insecure_skip_verify`
  - `disable_ssl`
  - `connect_timeout_ms`
  - `read_write_timeout_ms`
  - `retry_max_attempts`
  - `sts_role_arn`
  - `sts_external_id`
  - `sts_session_name`
  - `sts_duration_seconds`
  - `sts_policy`
  - `sts_endpoint`
- `bucket.extra`
  - `success_action_status`
  - `content_disposition`
  - `callback`
  - `callback_var`

默认值也统一在注册表里声明，例如：

- `sts_session_name = gge-upload`
- `sts_duration_seconds = 3600`
- `success_action_status = 200`

## 5. `extra_schema` 最小结构

```json
{
  "version": "v1",
  "fields": [
    {
      "key": "callback_scene",
      "label": "回调场景",
      "type": "string",
      "required": false,
      "placeholder": "editor",
      "description": "上传完成后用于业务识别的场景值"
    }
  ]
}
```

约束如下：

- `version` 目前只支持 `v1`
- `fields[].key` 必填且不能重复
- `fields[].key` 不能和显式运行时字段冲突
- `fields[].type` 支持：
  - `string`
  - `number`
  - `boolean`
  - `object`
  - `select`
- `select` 类型必须提供 `options`
- `label` 为空时自动回退为 `key`
- `type` 为空时自动回退为 `string`

## 6. 持久化策略

- 创建 / 更新 UploadKey 时，服务层统一走 `NormalizeUploadKeyExtraSchema`
- 创建 / 更新 Rule 时，服务层统一走 `NormalizeUploadRuleExtraSchema`
- 空值统一归一化为 `{}`，避免数据库里出现不稳定的 `null`

## 7. 后续扩展原则

- 新增 driver 时，先补注册表，再补 UI 模板，不直接把键名散落在页面里。
- 需要结构化表单时，优先从注册表和 `extra_schema` 推导，不重新发明一套配置协议。
- 若未来需要 UploadKey / Rule 层“自定义参数实际值”，应新增独立 `extra` 存储槽，不复用 `extra_schema`。
