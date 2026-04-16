# 上传配置中心二期现状差距与实施范围

> 对应任务：`tsk_01KPBFG73F2AG23F2JHDXM`
> 日期：2026-04-17
> 状态：阶段 1 基线文档

## 1. 当前链路现状

当前上传系统已经具备可运行的管理面和运行时链路：

- 管理面：
  - `StorageProvider / StorageBucket / UploadKey / UploadKeyRule` 四层模型已存在。
  - 管理接口位于 `backend/api/openapi/domains/storage_admin/`。
  - 管理页面位于 `frontend/src/views/system/upload-config/index.vue`。
- 运行时：
  - 中转上传走 `POST /media/upload`。
  - 直传准备走 `POST /media/prepare`。
  - 直传收口走 `POST /media/complete`。
  - 前端 SDK 已支持 `auto / direct / relay` 三种模式。
- 驱动：
  - 当前仅内置 `local` 与 `aliyun_oss`。
  - `aliyun_oss` 已支持服务端上传与直传准备。

这意味着“上传配置中心”已经不是空壳，当前缺的不是基础 CRUD，而是“配置如何真正驱动运行时”和“如何让不同上传场景一眼看懂怎么配”。

## 2. 当前主要差距

### 2.1 配置字段缺口

当前 `UploadKey` 仅覆盖路径、大小、类型、可见性等基础字段，缺少二期所需的运行时驱动字段：

- `upload_mode`
  - 当前只有前端 SDK 本地 `auto | direct | relay` 开关。
  - 后端配置中心还不能声明“只允许直传 / 只允许中转 / 自动协商”。
- `is_frontend_visible`
  - 当前没有“业务前端可见”的显式开关。
  - 导致管理面配置与运行时暴露策略无法分离。
- `permission_key`
  - 当前上传配置没有自己的业务权限键。
  - 运行时无法按 UploadKey 做细粒度权限控制。
- `fallback_key`
  - 当前只有系统默认上传 Key 兜底。
  - 无法对单个业务 Key 指定显式回退目标。
- `accept` / `direct` / `relay` 等前端展示配置
  - 当前只能从大小和 mime 推导，不足以支撑低代码式引导。
- `extra_schema`
  - 当前有 `extra/meta` 扩展槽，但没有 schema 驱动的参数定义。
  - 结果是“能存 JSON，但不知道怎么让用户安全、清晰地配置”。

### 2.2 运行时接口缺口

- 当前业务前端只能：
  - 直接知道某个固定 `UploadKey` 后调用上传 SDK。
  - 或者依赖后端 `prepare` 动态决策 direct/relay。
- 当前缺少一个“安全的前端可见配置接口”：
  - 只暴露允许前端看到的 UploadKey 列表。
  - 只返回前端必需信息，不返回管理面明文配置。

因此当前链路支持“前端上传”，但还不支持“前端以配置中心为驱动，自助发现可用上传场景”。

### 2.3 管理页体验缺口

当前页面已经有四层结构雏形，但仍偏“字段表单罗列”，缺少面向业务配置者的引导：

- Provider 页只区分 `local` 和 `aliyun_oss`，没有驱动差异化表单分区。
- Bucket 页缺少“公开访问、回源、回调、加速域名”的可视化说明。
- UploadKey 页缺少“这个配置给谁用、怎么上传、前端是否可见、权限如何控制”的核心信息。
- Rule 页缺少“覆盖关系”和“与默认规则的优先级”提示。
- `extra/meta` 还是隐式扩展位，没有“自定义参数”编辑体验。

### 2.4 驱动扩展缺口

- 当前内置驱动只有 `local` 和 `aliyun_oss`。
- `extra` 已可承载高级参数，但缺少统一的驱动注册配置描述。
- 后续接入 `cos / s3` 时，若继续硬编码表单，会快速失控。

## 3. 二期实施目标

本期不做“大而全”的多云平台，而是把上传配置中心补到可扩展、可理解、可驱动运行时的最小闭环。

### 3.1 本期必须完成

- 补齐 `UploadKey` 运行时驱动字段：
  - `upload_mode`
  - `is_frontend_visible`
  - `permission_key`
  - `fallback_key`
  - `client_accept`
  - `direct_size_threshold_bytes`
  - `extra_schema`
- 补齐 `UploadKeyRule` 扩展字段：
  - `visibility_override`
  - `mode_override`
  - `client_accept`
  - `extra_schema`
- 新增业务前端可见配置接口：
  - 返回前端允许看到的 UploadKey 列表与规则摘要。
  - 不返回 Provider/Bucket 明细与 AK/SK 等敏感信息。
- 让 `PrepareUpload` 真正读取 `EffectiveConfig` 中的上传模式与回退策略。
- 管理页改造成“四层分区 + 说明区 + 高级参数区”的配置体验。
- 支持 `extra_schema` 驱动的“自定义参数”配置，先用通用 JSON 编辑器 + 结构化说明落地。

### 3.2 本期明确不做

- 不在本期接入完整 `COS / S3` 实现。
- 不做完整 JSONSchema Form Engine。
- 不做前端任意动态渲染上传器。
- 不把管理面配置原样暴露给业务前端。
- 不引入新的权限体系，仍复用现有 `permission_key` 机制。

## 4. 配置分层与展示原则

### 4.1 Provider

面向“存储接入”：

- 驱动类型
- endpoint / region / base_url
- AK / SK
- 高级参数 `extra`
- 驱动说明与示例

展示原则：

- 基础连接信息和高级参数分区展示。
- 不同 driver 使用不同文案和推荐模板。

### 4.2 Bucket

面向“物理存储空间”：

- bucket_name / base_path / public_base_url
- 公开访问开关
- 回调、CDN、加速、内网接入等高级参数

展示原则：

- 让配置者先看懂“文件实际存哪”和“最终对外怎么访问”。

### 4.3 UploadKey

面向“业务上传入口”：

- 业务标识
- 归属 bucket
- 默认路径模板
- 默认规则
- 上传模式
- 前端是否可见
- 权限键
- 回退 key
- 前端 accept 与阈值配置
- 自定义参数 schema

展示原则：

- UploadKey 页面优先回答四个问题：
  - 谁在用这个 Key
  - 前端能不能看到
  - 走直传还是中转
  - 权限和回退怎么处理

### 4.4 Rule

面向“子策略覆盖”：

- 子路径
- 文件名策略
- 大小 / mime / pipeline
- 模式覆盖
- 可见性覆盖
- 自定义参数 schema

展示原则：

- 强调“只改差异项”，未配置即继承 UploadKey 默认值。

## 5. 关键实施决策

### 5.1 管理面与运行时分离

- 管理面继续使用 `storage_admin`。
- 运行时新增独立“前端可见配置”接口。
- 业务前端只消费运行时安全视图，不直接读管理面 CRUD 数据。

### 5.2 扩展参数以 schema 驱动，不再继续堆散字段

- `extra` 继续作为实际存储槽。
- `extra_schema` 负责描述这些参数如何展示、校验和解释。
- 本期先实现“结构化元信息 + JSON 编辑器”的最小版本。

### 5.3 上传模式以后端 EffectiveConfig 为准

- 前端 `mode=auto/direct/relay` 只保留为调用偏好。
- 最终是否可 direct，由后端结合：
  - driver capability
  - upload_key.upload_mode
  - rule.mode_override
  - 文件大小阈值
  - 回退策略
 统一裁决。

## 6. 字段矩阵（本期新增）

### 6.1 UploadKey

| 字段 | 类型 | 作用 |
| --- | --- | --- |
| `upload_mode` | enum(`auto`,`direct`,`relay`) | 默认上传模式 |
| `is_frontend_visible` | boolean | 是否出现在前端可见配置接口 |
| `permission_key` | string | 业务上传权限键 |
| `fallback_key` | string | 当前 Key 不可用时的回退目标 |
| `client_accept` | string[] | 前端上传器 accept 建议值 |
| `direct_size_threshold_bytes` | bigint | 超过阈值自动转 relay |
| `extra_schema` | jsonb | UploadKey 自定义参数定义 |

### 6.2 UploadKeyRule

| 字段 | 类型 | 作用 |
| --- | --- | --- |
| `mode_override` | enum(`inherit`,`direct`,`relay`) | 子规则覆盖上传模式 |
| `visibility_override` | enum(`inherit`,`public`,`private`) | 子规则覆盖可见性 |
| `client_accept` | string[] | 子规则前端 accept 建议值 |
| `extra_schema` | jsonb | 子规则自定义参数定义 |

## 7. 执行顺序

后续按下面顺序推进，不逆序跳步：

1. 阶段 1：差距盘点、范围收口、字段矩阵与页面分层方案
2. 阶段 2：模型 + migration + OpenAPI 契约升级
3. 阶段 3：后端 resolver / service / runtime visible API
4. 阶段 4：前端配置中心 + 上传 SDK 适配
5. 阶段 5：生成链、联编、验证、收尾

### 7.1 seed / migration 职责边界

- migration：
  - 只负责结构变更与历史数据一次性修正
  - 可以给新增列补安全的数据库默认值，保证旧数据升级不炸
- `EnsureDefaultSeeds`：
  - 负责脚手架默认 Provider / Bucket / UploadKey / Rule 的长期默认状态
  - 新增字段的业务语义默认值必须在 seed 中显式写出，不能长期隐式依赖数据库列默认值
- 本期新增字段中：
  - `upload_mode`
  - `is_frontend_visible`
  - `client_accept`
  - `mode_override`
  - `visibility_override`
  - `extra_schema`
  这些字段都需要在默认 seed 中显式落值，避免后续数据库默认值调整时影响内置场景行为

## 8. 兼容策略

### 8.1 已有 UploadKey 数据兼容

- 现有 `upload_keys.visibility` 保持不变，继续作为对象可见性来源。
- 新增字段提供默认值，避免旧数据失效：
  - `upload_mode = auto`
  - `is_frontend_visible = false`
  - `permission_key = ''`
  - `fallback_key = ''`
  - `client_accept = []`
  - `direct_size_threshold_bytes = 0`
  - `extra_schema = {}`
- 旧 Key 不会因为本次升级自动暴露给业务前端，必须显式打开 `is_frontend_visible`。

### 8.2 已有前端上传调用兼容

- 保留现有 `uploadMediaWithPrepare(file, { key, rule, mode })` 调用方式。
- 前端的 `mode=auto/direct/relay` 继续可用，但变成“调用偏好”：
  - `auto`：按后端裁决执行。
  - `direct`：仅当后端允许 direct 时成功，否则报错。
  - `relay`：跳过 prepare，直接中转。
- 这样既不破坏旧调用，又让后端配置真正接管运行时策略。

### 8.3 已有管理页兼容

- 管理页不推翻重做，继续沿用现有三大 Tab + Rule 抽屉骨架。
- 二期只做结构升级：
  - 增加四层说明区
  - 增加“基础配置 / 运行时策略 / 前端暴露 / 高级参数”分组
  - 增加自定义参数编辑区
- 先保证现有 CRUD 行为不回归，再补增强字段的交互体验。

### 8.4 驱动扩展兼容

- 本期先不新增 COS/S3 真正落地驱动。
- 但 OpenAPI、模型和页面要预留 `driver extra + extra_schema` 扩展位。
- 后续新增驱动时，只补：
  - driver factory
  - schema 默认模板
  - driver 差异化表单配置
  不再反复改主流程。

## 9. 直接影响文件

- `backend/internal/modules/system/models/upload.go`
- `backend/internal/pkg/database/migrations/`
- `backend/api/openapi/domains/storage_admin/`
- `backend/api/openapi/domains/media/`
- `backend/internal/modules/system/upload/`
- `backend/internal/api/handlers/storage_admin.go`
- `backend/internal/api/handlers/media.go`
- `frontend/src/domains/upload-config/`
- `frontend/src/domains/upload/`
- `frontend/src/views/system/upload-config/index.vue`
