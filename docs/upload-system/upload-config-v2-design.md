# 上传配置中心二期设计稿

> 对应任务：`tsk_01KPBFG73F2AG23F2JHDXM`
> 日期：2026-04-17
> 状态：阶段 1 设计稿

## 1. 设计目标

二期配置中心要解决两个问题：

- 让配置者一眼看懂“这一层配置是干什么的、应该怎么配”。
- 让运行时真正从配置中心拿到可执行的上传策略，而不是只做 CRUD 存档。

## 2. 页面信息架构

### 2.1 总体结构

页面继续沿用现有管理面骨架，但每一层都拆成四个固定区块：

1. 基础信息
2. 运行时策略
3. 前端暴露
4. 高级参数

### 2.2 Provider

#### 基础信息

- `provider_key`
- `name`
- `driver`

#### 运行时策略

- `endpoint`
- `region`
- `base_url`
- 默认状态

#### 前端暴露

- 无

#### 高级参数

- `extra`
- 驱动说明卡片
- 推荐配置样例

### 2.3 Bucket

#### 基础信息

- `bucket_key`
- `name`
- `bucket_name`

#### 运行时策略

- `base_path`
- `public_base_url`
- `is_public`
- `status`

#### 前端暴露

- 无直接字段，只展示“该桶承载的 UploadKey 是否允许前端可见”

#### 高级参数

- `extra`
- 访问域名 / callback / 加速 / 内网配置说明

### 2.4 UploadKey

#### 基础信息

- `key`
- `name`
- `bucket_id`
- `path_template`
- `default_rule_key`

#### 运行时策略

- `upload_mode`
- `visibility`
- `fallback_key`
- `max_size_bytes`
- `allowed_mime_types`
- `direct_size_threshold_bytes`

#### 前端暴露

- `is_frontend_visible`
- `permission_key`
- `client_accept`

#### 高级参数

- `meta`
- `extra_schema`
- 当前规则继承提示

### 2.5 Rule

#### 基础信息

- `rule_key`
- `name`
- `sub_path`
- `filename_strategy`

#### 运行时策略

- `mode_override`
- `visibility_override`
- `max_size_bytes`
- `allowed_mime_types`
- `process_pipeline`

#### 前端暴露

- `client_accept`

#### 高级参数

- `meta`
- `extra_schema`
- 覆盖 UploadKey 默认值的对比说明

## 3. 页面交互原则

### 3.1 列表页

- 列表必须优先显示能帮助判断上传链路的字段：
  - Provider：driver、endpoint、默认状态、健康状态
  - Bucket：provider、bucket_name、public_base_url、公开状态
  - UploadKey：upload_mode、frontend_visible、permission_key、fallback_key、默认 rule
  - Rule：mode_override、visibility_override、process_pipeline、is_default

### 3.2 编辑页

- 默认先展示基础配置。
- 高级参数折叠收起，避免低频字段淹没核心配置。
- 只在用户选择特定 driver 或打开高级模式后显示复杂字段。
- 明确展示“继承”和“覆盖”的关系，不让用户猜测最终生效值。

### 3.3 帮助文案

- 每个区块顶部都要有一句话说明：
  - 这一层解决什么问题
  - 哪些字段会影响运行时
  - 哪些字段只影响前端展示

## 4. 驱动差异化表单策略

### 4.1 Local

- 基础字段：
  - `base_url`
- 高级字段：
  - 本地根目录说明
  - 临时目录说明
- 限制：
  - 不支持 direct
  - 不显示 STS / callback / form policy 配置

### 4.2 Aliyun OSS

- 基础字段：
  - `endpoint`
  - `region`
  - `base_url`
  - `access_key`
  - `secret_key`
- 高级字段：
  - `internal_endpoint`
  - `cdn_base_url`
  - `sts_enabled`
  - `sts_role_arn`
  - `sts_session_name`
  - `sts_duration_seconds`
  - `callback_url`
  - `callback_body`
  - `callback_body_type`

### 4.3 后续 COS / S3 扩展规则

- 不新增一套新的 UI 架构。
- 只为 driver 提供：
  - 基础字段模板
  - 高级参数模板
  - 文案模板
  - `extra_schema` 默认样例

## 5. 自定义参数扩展机制

### 5.1 设计原则

- `extra` 负责存值。
- `extra_schema` 负责解释这些值。
- 不把所有未来参数都摊平成固定列。

### 5.2 `extra_schema` 最小结构

```json
{
  "version": "v1",
  "fields": [
    {
      "key": "callback_url",
      "label": "回调地址",
      "type": "string",
      "required": false,
      "placeholder": "https://example.com/upload/callback",
      "description": "对象存储上传完成后触发的回调地址"
    }
  ]
}
```

### 5.3 本期实现边界

- 管理面先支持：
  - JSON 编辑
  - 基础校验
  - 常见模板插入
- 暂不实现完整 schema form engine。
- 后端仅做结构透传与基础校验，不做复杂联动。

## 6. 决策表

| 决策点 | 选择 | 原因 |
| --- | --- | --- |
| 管理面与运行时配置是否共用接口 | 否 | 前端只能看安全视图，不能直接读管理 CRUD |
| 上传模式由谁最终裁决 | 后端 `EffectiveConfig` | 统一 direct/relay/fallback 行为 |
| 扩展参数如何落地 | `extra + extra_schema` | 保持灵活且便于逐步增强 |
| 管理页是否重做 | 否 | 保留现有骨架，渐进增强 |
| 是否本期直接上完整多云 | 否 | 先把扩展位和运行时闭环做扎实 |

## 7. 本期字段矩阵

### 7.1 UploadKey

| 字段 | 含义 | 页面分区 |
| --- | --- | --- |
| `upload_mode` | 默认上传模式 | 运行时策略 |
| `is_frontend_visible` | 是否前端可见 | 前端暴露 |
| `permission_key` | 上传权限键 | 前端暴露 |
| `fallback_key` | 回退目标 key | 运行时策略 |
| `client_accept` | 前端 accept 建议值 | 前端暴露 |
| `direct_size_threshold_bytes` | 直传阈值 | 运行时策略 |
| `extra_schema` | 自定义参数定义 | 高级参数 |

### 7.2 UploadKeyRule

| 字段 | 含义 | 页面分区 |
| --- | --- | --- |
| `mode_override` | 子规则覆盖上传模式 | 运行时策略 |
| `visibility_override` | 子规则覆盖可见性 | 运行时策略 |
| `client_accept` | 子规则前端 accept 建议值 | 前端暴露 |
| `extra_schema` | 子规则自定义参数定义 | 高级参数 |

## 8. 与代码改造的映射关系

- 模型与迁移：
  - 新增 `UploadKey / UploadKeyRule` 字段
- OpenAPI：
  - 管理面 schema 扩容
  - 新增前端安全视图接口
- 后端：
  - resolver 合并新字段
  - prepare 读取模式与阈值
  - runtime visible API 输出安全摘要
- 前端：
  - 配置页增加分区和新增字段
  - 上传 SDK 读取后端返回的真实模式结果
