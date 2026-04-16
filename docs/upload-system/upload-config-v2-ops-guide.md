# 上传配置中心二期运维与配置示例

> 对应任务：`tsk_01KPBFG73F2AG23F2JHDXM`
> 日期：2026-04-17
> 状态：交付版

## 1. 配置心智模型

上传配置中心按四层组织，每一层只回答一个问题：

- Provider：怎么接入存储驱动和凭据
- Bucket：文件最终放在哪个桶、通过什么域名访问
- UploadKey：业务入口是什么，前端能不能发现，默认走直传还是中转
- Rule：这个入口下不同文件场景如何做差异化覆盖

推荐配置顺序：

1. 先建 Provider
2. 再建 Bucket
3. 再建 UploadKey
4. 最后按场景补 Rule

## 2. Provider 配置示例

### 2.1 local

适合：

- 本地开发
- 单机环境
- 不要求前端直传

建议：

- `driver = local`
- `base_url` 指向后端静态可访问前缀
- 不开启 direct 相关配置

`provider.extra`：

```json
{}
```

### 2.2 aliyun_oss

适合：

- 需要对象存储托管
- 需要浏览器直传
- 需要 STS / callback / CDN 等能力

基础字段建议：

- `driver = aliyun_oss`
- `endpoint = oss-cn-hangzhou.aliyuncs.com`
- `region = cn-hangzhou`
- `base_url = https://your-cdn.example.com`

`provider.extra` 示例：

```json
{
  "use_cname": true,
  "use_path_style": false,
  "disable_ssl": false,
  "connect_timeout_ms": 5000,
  "read_write_timeout_ms": 10000,
  "retry_max_attempts": 3,
  "sts_role_arn": "acs:ram::1234567890123456:role/upload-direct-role",
  "sts_external_id": "",
  "sts_session_name": "gge-upload",
  "sts_duration_seconds": 3600,
  "sts_policy": "",
  "sts_endpoint": "sts.cn-hangzhou.aliyuncs.com"
}
```

说明：

- 不需要 STS 时，可只保留基础连接参数。
- 开启 STS 后，前端直传会优先使用临时凭证链路。

## 3. Bucket 配置示例

### 3.1 公共读 Bucket

适合：

- 商品图
- 富文本素材
- 页面装饰图

建议：

- `is_public = true`
- `public_base_url` 指向 CDN 或公开域名
- `base_path` 按业务域拆目录，例如 `media/`、`assets/`

`bucket.extra` 示例：

```json
{
  "success_action_status": "200",
  "content_disposition": "inline"
}
```

### 3.2 带 callback 的 Bucket

适合：

- 需要对象存储上传后触发业务回调
- 需要后端二次登记或异步处理

`bucket.extra` 示例：

```json
{
  "success_action_status": "200",
  "callback": "{\"callbackUrl\":\"https://api.example.com/upload/callback\",\"callbackBody\":\"bucket=${bucket}&object=${object}&etag=${etag}\",\"callbackBodyType\":\"application/x-www-form-urlencoded\"}",
  "callback_var": "{\"x:scene\":\"editor\"}"
}
```

说明：

- `callback` / `callback_var` 会进入 OSS 直传表单，不会通过前端可见列表原样暴露。
- 回调地址必须是平台控制的后端地址，不要直接指向不受控外部系统。

## 4. UploadKey 配置示例

### 4.1 富文本图片

目标：

- 前端可发现
- 小图优先直传
- 大图自动回退 relay

建议字段：

```json
{
  "key": "editor.image",
  "upload_mode": "auto",
  "is_frontend_visible": true,
  "permission_key": "content.editor.upload",
  "fallback_key": "media.default",
  "client_accept": ["image/*"],
  "direct_size_threshold_bytes": 5242880,
  "visibility": "public"
}
```

### 4.2 站点配置图片

目标：

- 仅后台页面使用
- 允许前端管理台发现
- 不需要复杂 Rule

建议字段：

```json
{
  "key": "site.asset",
  "upload_mode": "auto",
  "is_frontend_visible": true,
  "permission_key": "system.site_config.manage",
  "fallback_key": "",
  "client_accept": ["image/*"],
  "direct_size_threshold_bytes": 3145728,
  "visibility": "public"
}
```

### 4.3 仅后端中转的私有文件

目标：

- 不暴露前端可见列表
- 始终 relay

建议字段：

```json
{
  "key": "internal.private-file",
  "upload_mode": "relay",
  "is_frontend_visible": false,
  "permission_key": "",
  "fallback_key": "",
  "client_accept": [],
  "direct_size_threshold_bytes": 0,
  "visibility": "private"
}
```

## 5. Rule 配置示例

### 5.1 图片规则

```json
{
  "rule_key": "image",
  "mode_override": "auto",
  "visibility_override": "public",
  "client_accept": ["image/png", "image/jpeg", "image/webp"],
  "max_size_bytes": 10485760
}
```

### 5.2 原始文件规则

```json
{
  "rule_key": "file",
  "mode_override": "relay",
  "visibility_override": "private",
  "client_accept": [".pdf", ".docx", ".xlsx"],
  "max_size_bytes": 52428800
}
```

## 6. `extra_schema` 使用建议

`extra_schema` 不是存值槽，而是“如何解释自定义参数”的定义槽。

适合放入：

- 前端说明型参数
- 业务扩展字段定义
- 未来动态表单所需的字段描述

不适合放入：

- `upload_mode`
- `permission_key`
- `fallback_key`
- `is_frontend_visible`
- 其他运行时主链路关键字段

UploadKey `extra_schema` 示例：

```json
{
  "version": "v1",
  "fields": [
    {
      "key": "scene",
      "label": "上传场景",
      "type": "string",
      "required": false,
      "placeholder": "editor",
      "description": "供业务在后续链路中识别来源场景"
    },
    {
      "key": "review_policy",
      "label": "审核策略",
      "type": "select",
      "required": false,
      "options": [
        {"label": "默认", "value": "default"},
        {"label": "严格", "value": "strict"}
      ]
    }
  ]
}
```

## 7. 常见配置建议

### 7.1 什么场景优先 relay

- 本地盘
- 需要后端先做同步处理
- 需要避免浏览器直接接触对象存储域名
- 文件体积小且没必要多一次 prepare

### 7.2 什么场景适合 direct

- OSS / 后续 S3 / COS 等对象存储
- 纯文件写入，无需后端先加工
- 浏览器上传体积较大，希望降低后端带宽压力

### 7.3 如何设置 fallback

- direct 依赖外部 STS、策略签名或对象存储配置时，建议配置一个 relay 型 `fallback_key`
- fallback 目标应尽量复用同一业务域的默认 UploadKey，避免跨域配置绕路太远

## 8. 日常排障清单

- Provider 健康检查失败：
  - 先核对 `endpoint / region / AK / SK`
  - 再核对 `provider.extra` 中 SSL / CNAME / PathStyle 参数
- 前端看不到 UploadKey：
  - 核对 `is_frontend_visible`
  - 核对 UploadKey `status`
  - 核对当前用户是否具备 `permission_key`
- 直传被回退为 relay：
  - 核对 driver 是否支持 direct
  - 核对 `upload_mode`
  - 核对 `direct_size_threshold_bytes`
  - 核对 Rule 是否设置了 `mode_override = relay`
- callback 失败：
  - 核对 OSS 回调地址可达性
  - 核对 bucket.extra 中 callback JSON
  - 核对后端 callback 验签日志

## 9. 后续扩展约束

- 新增 driver 时，先补后端 schema registry，再补前端 driver registry。
- 不为新 driver 复制一套新页面结构，只补：
  - 基础字段模板
  - 高级参数模板
  - 文案与示例
- 未来若要支持 UploadKey / Rule 自定义参数实际值，新增独立 `extra` 存值槽，不复用 `extra_schema`。
