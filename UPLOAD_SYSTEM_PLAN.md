# 文件上传系统 + 配置中心 — 实施计划书

> 版本：v1.0  ｜  日期：2026-04-16  ｜  状态：规划定稿，尚未启动执行
> 任务主键：`tsk_01KP930A1038SPNGAMZ912`（项目：`prj_01KP87XR1XEA58497KYTV8`）

---

## 0. TL;DR（一页纸结论）

- **做一件事**：把 MaBen 现有散落的上传代码，统一成一个「配置中心 + 上传服务 + 前端 SDK」三件套，对外只暴露一个稳定的「Key」。
- **统一心智模型**：`Provider → Bucket → UploadKey → UploadKeyRule` 四级，上游越静态（基础设施），下游越业务（规则）。其中 **UploadKey 是容器（大 Key），Rule 是可组合子策略（小 Key）**，这是整个系统的锚点。
- **统一调用方式**：前端永远只写 `upload('ABC')` 或 `upload('ABC.avatar')`，后端所有校验/处理/落盘/回调全部由配置驱动。
- **支持全场景**：本地盘 / 阿里云 OSS / 腾讯云 COS / S3 兼容；直传（STS / Presigned / POST Policy）+ 中转（relay）两种形态共存；多租户 + AK/SK 加密 + 默认 Key 回退。
- **全量 6 阶段、259 个任务节点**已在 Task Tree 落地（详见 §9 索引），本文档是它的「可读面」。**任务线现在还没开跑**。

---

## 1. 目标与非目标

### 1.1 目标（In-Scope）

1. **三层抽象落地**：`StorageProvider / StorageBucket / UploadKey / UploadKeyRule` 四张核心表 + 一张 `UploadRecord` 审计表。
2. **多 Driver 支持**：Local（开发+私有化）、Aliyun OSS、Tencent COS、S3-Compatible（AWS S3 / MinIO / R2 / ...），外加 Custom driver 骨架。
3. **两种上传形态并存**：
   - 前端直传（STS / Presigned / POST Policy，含 provider 回调签名校验）
   - 后端中转（relay，含完整处理管道）
   - 按 `driver capability × key.upload_mode × 文件尺寸` 的决策矩阵自动选择。
4. **规则引擎**：
   - **Validator Pipeline**：mime / size / dimension / duration / regex / dedup / CEL 表达式。
   - **Processor Pipeline**：ImageCompress / Watermark / Thumbnail / VideoTranscode / Antivirus / Moderation，含同步 vs 异步分流。
5. **默认 Key + Fallback**：Key 不存在时可回退到 `App.capabilities.default_upload_key`；支持 `strict / fallback / override` 三策略。
6. **前端可见性与权限**：`is_frontend_visible=false` 的 Key 不出现在前端接口；基于 `UploadKey.permission_key` 的细粒度动态权限。
7. **多租户隔离**：三张核心表带 `tenant_id`，`object_key` 模板强制含 `{tenant_id}` 前缀；所有查询显式过滤。
8. **AK/SK 加密**：`SecretCipher` 接口 + AES-256-GCM（版本前缀 `cipher_v1::`），预留 KMS 接入；支持主密钥轮换、双主密钥过渡窗口。
9. **前后端统一 SDK**：
   - 后端：`UploadService` + OpenAPI 契约（admin + frontend 两套 domain）。
   - 前端：`useUpload` composable / 扁平字符串 key（`ABC.avatar`）与对象参数两种形态。
10. **管理端 UI**：Providers / Buckets / UploadKeys / Rules / Records / 统计面板，配 Permission 联动选择器。

### 1.2 非目标（Out-of-Scope，v1.0）

- ❌ **运行时动态注册 Driver / Processor**：MVP 只开放编译期注册；未来再谈沙箱、资源上限、权限隔离。
- ❌ **KMS 正式对接**：只预留 `SecretCipher` 接口 + env 主密钥实现。
- ❌ **非 admin 跨租户资源池**：单租户 `default` 起步，但模型与查询层已预留 `tenant_id`。
- ❌ **WangEditor 以外的富文本迁移**：只迁移现有编辑器；其他业务集成交到应用层。
- ❌ **内置 CDN 管理**：只做 `public_url` 模板拼接，不做 CDN 刷新。

---

## 2. 核心概念

### 2.1 三层抽象（四级模型）

```
StorageProvider      基础设施层   ← 凭证、endpoint、容量（SaaS 租户通常只读）
      │
      └──  StorageBucket        物理容器   ← 一个 provider 下的一个桶 / 本地根目录
                │
                └──  UploadKey          业务容器（"大 Key"）
                           │              绑定 bucket / 根路径 / 权限 / 默认 Rule
                           │              前端/后端只认这一层
                           │
                           └──  UploadKeyRule   子策略（"小 Key"）
                                                  覆盖：mime / size / 处理管道 /
                                                  子路径 / 命名策略 / 可见性 …
```

**读向与写向的解耦**：
- 写端（上传）：`业务代码` → `UploadKey [.rule]` → 解析为 `EffectiveConfig` → 落到 Driver。
- 读端（展示）：`UploadRecord.public_url` 或 Bucket `public_url_template`，**完全不依赖 Key**。

### 2.2 「大 Key / 小 Key(Rule)」语义（★ 全系统锚点）

> 这是整个系统最容易被人误解的一层，写在最前面。

- **UploadKey（大 Key）**：稳定、少量、业务语义强。例：`ABC`、`USER_AVATAR`、`PRODUCT_MEDIA`。
  - 承载：`bucket_id`、`root_path`、`permission_key`、`upload_mode（direct_only/relay_only/both）`、`is_frontend_visible`、一组默认 rule 值。
- **UploadKeyRule（小 Key）**：同一 Key 下的可切换子形态，数量可多可少。例：`ABC.avatar`、`ABC.thumb`、`ABC.original`。
  - 承载：mime 白名单、size 限制、命名模板、处理管道、子路径、可见性 override…
  - **继承关系**：Rule 字段为 `null/unset` 时继承 Key 的默认值；非空即覆盖。
- **前端调用形态（三选一，全部合法）**：
  ```ts
  upload(file, 'ABC')                    // 使用 Key + 默认 rule
  upload(file, 'ABC.avatar')             // 扁平字符串（key.rule）
  upload(file, { key: 'ABC', rule: 'avatar' })  // 对象参数
  ```
- **Rule 不存在时的处理**：回落到 Key 默认 rule（带 `onFallback` 回调通知）；若 Key 都不存在，按 §2.3 的 fallback 策略处理。

### 2.3 数据模型一览（详见阶段 2）

| 表 | 角色 | 关键字段 |
|---|---|---|
| `storage_provider` | 基础设施账号（含加密 AK/SK） | `driver_type`, `credentials_encrypted`, `region`, `status` |
| `storage_bucket` | 物理容器 | `provider_id`, `name`, `public_url_template`, `tenant_id` |
| `upload_key` | 业务容器（大 Key） | `bucket_id`, `root_path`, `upload_mode`, `permission_key`, `is_frontend_visible`, `default_rule`, `tenant_id` |
| `upload_key_rule` | 子策略（小 Key） | `upload_key_id`, `rule_key`, `mime`, `size`, `pipeline`, `path_template` |
| `upload_record` | 审计/去重/生命周期 | `upload_key_id`, `rule_key`, `object_key`, `sha256`, `size`, `uploader`, `tenant_id` |

---

## 3. 架构总览

### 3.1 组件依赖（自下而上）

```
┌──────────────────────────────────────────────────────────────┐
│ Admin UI (Vue3)        Frontend SDK (useUpload)              │
│    └─ Providers/Buckets/Keys/Rules/Records/Stats CRUD        │
│    └─ Direct/Relay/Multipart uploader                        │
└──────────────────────────────────────────────────────────────┘
                              ↕  OpenAPI (admin + frontend)
┌──────────────────────────────────────────────────────────────┐
│ Handlers 层（ogen）                                          │
│  Providers│Buckets│UploadKeys│Rules│Records│Frontend{Prep..│ │
│                                    │Complete│Relay}│Callback│ │
└──────────────────────────────────────────────────────────────┘
                              ↕
┌──────────────────────────────────────────────────────────────┐
│ UploadService                                                │
│  ├─ UploadKeyResolver  (parse → load → merge → visibility)   │
│  ├─ PathTemplateEngine (DSL 占位符渲染 + 安全过滤)           │
│  ├─ Validator Pipeline (mime/size/dim/dur/regex/dedup/CEL)   │
│  ├─ Processor Pipeline (compress/watermark/transcode/..)     │
│  ├─ QuotaGuard         (租户/key 配额、并发 backpressure)    │
│  └─ EventBus           (upload_record 写入 + webhook 回调)   │
└──────────────────────────────────────────────────────────────┘
                              ↕
┌──────────────────────────────────────────────────────────────┐
│ Driver Registry                                              │
│  ├─ LocalDriver  (relay only, 目录遍历防护, 原子 rename)     │
│  ├─ AliyunDriver (STS / PostPolicy / relay / callback)       │
│  ├─ TencentDriver(STS / PostPolicy / relay / callback)       │
│  ├─ S3Driver     (Presigned PUT / POST / multipart)          │
│  └─ CustomDriver 骨架 + Contract 测试 harness                │
└──────────────────────────────────────────────────────────────┘
                              ↕
┌──────────────────────────────────────────────────────────────┐
│ Repository 层（GORM）  +  SecretCipher（AES-256-GCM）        │
│  Redis 缓存 + 失效广播 ｜  builtin 注册（参考 dictionary）   │
└──────────────────────────────────────────────────────────────┘
```

### 3.2 两种上传流程（简化时序）

**前端直传（direct，e.g. 阿里云 STS）**
```
FE ──(1) POST /upload/prepare { key, file_meta }──► BE UploadService
                                                    ├─ Resolve Key+Rule
                                                    ├─ Validate (mime/size/dim/regex)
                                                    ├─ Render object_key
                                                    └─ Sign STS / Presigned / PostPolicy
FE ◄──(2) { upload_endpoint, credentials, object_key, session_id }
FE ──(3) 直传到 OSS ──────────────────────────────────► Provider
Provider ──(4) callback {object_key, size, hash, mime} ──► BE CallbackHandler
                                                          ├─ 验证签名
                                                          ├─ 触发 Processor Pipeline（同步/异步分流）
                                                          └─ 写 upload_record + event
FE ──(5) POST /upload/complete { session_id } ──► BE FrontendCompleteHandler
FE ◄──(6) { public_url, record_id, processed: bool }
```

**后端中转（relay，e.g. LocalDriver 或强制 relay 策略）**
```
FE ──(1) POST /upload/relay  (multipart, key) ──► BE FrontendRelayHandler
                                                  ├─ Resolve Key+Rule
                                                  ├─ Validate（含 dedup sha256）
                                                  ├─ Processor Pipeline（同步段）
                                                  ├─ Driver.Upload(stream)
                                                  ├─ 写 upload_record
                                                  └─ 异步段 enqueue（如转码/moderation）
FE ◄──(2) { public_url, record_id }
```

---

## 4. 能力清单（9 大配置类别）

> 每一类都能在 UploadKey 设默认值、在 UploadKeyRule 覆盖。详细字段见阶段 1 任务「可配置项分类清单（9 大类）」。

1. **存储**：provider / bucket / root_path / path_template / public_url_template。
2. **校验**：mime 白名单、size 限制、图片尺寸、视频时长、文件名正则、sha256 去重、CEL 表达式。
3. **处理**：图片压缩 / 水印 / 缩略图 / 视频转码 / 病毒扫描 / 内容审核（同步 or 异步）。
4. **访问**：ACL（private/public-read）、签名 URL 有效期、CDN 域名、下载鉴权。
5. **回调**：Provider callback 签名校验、业务 webhook、重试队列、失败降级。
6. **配额**：租户存储总额、Key/Rule 次数配额、并发 backpressure、超限告警。
7. **元数据抽取**：EXIF / 视频帧率 / 音频时长 / 人脸检测点位（仅元数据，**不采集人脸图像**）。
8. **前端控制**：`is_frontend_visible`、permission_key、默认 rule、UI 文案、进度/取消行为。
9. **生命周期**：TTL 删除、归档策略、软删 + 合规保留窗口、导出归档。

---

## 5. 关键设计决策（ADR 速览）

| # | 标题 | 选择 | 代价 / 备注 |
|---|---|---|---|
| ADR-001 | 三层抽象 | Provider→Bucket→UploadKey→Rule | 多一张表，但换来配置 DRY |
| ADR-002 | 规则配置存储 | JSON 列 + JSONSchema 动态校验 | 比宽表灵活；需统一 schema 版本 |
| ADR-003 | 密钥加密 | AES-256-GCM + env 主密钥；抽 `SecretCipher` 接口 | KMS 未来替换只改实现 |
| ADR-004 | 上传模式 | `direct_only / relay_only / both`，默认 both；Local 强制 relay | 决策矩阵由 Resolver 执行 |
| ADR-005 | 路径模板 DSL | 显式白名单占位符 `{tenant_id}/{yyyy}/{mm}/{uid}/{uuid}/{hash:N}/{ext}` | 黑名单防注入 |
| ADR-006 | 处理管道 | 同步段走请求线程；异步段进队列（复用现有 task/queue） | 明确哪些必须同步（校验/去重/轻量）|
| ADR-007 | 权限模型 | `UploadKey.permission_key` 绑定菜单/路由/端点 | 动态权限，需补生成器 |
| ADR-008 | 多租户 | `tenant_id` 列 + object_key 强制前缀 + 查询显式过滤 | 每一层都得走这个约束 |
| ADR-009 | `upload_record` 策略 | 存所有成功上传；失败另表/日志 | 审计 + 去重；体积由 TTL 控制 |
| ADR-010 | 前端 SDK 调用形态 | 扁平字符串 `key.rule` + 对象参数 两种形态共存 | 简单场景用字符串，复杂场景用对象 |

详见阶段 1 任务「ADR 决策记录」下的每份 `docs/adr/adr-00x-*.md`。

---

## 6. 阶段规划（6 阶段，门禁串行）

### 阶段 1 — 规格定稿与架构决策（纸面）
- **任务树根**：`nd_01KP930A10DWMKTGFGNDKP`
- **输出**：需求定稿（11 项）+ 架构图（6 张）+ ADR（10 份）+ 契约预审（OpenAPI / SDK / UI / 权限键）+ 阶段评审会签。
- **验收**：
  - 每条需求都有文字结论；
  - 每份 ADR 列「上下文 / 候选 / 选择 / 代价」；
  - Mermaid/PlantUML 图已入 `docs/`；
  - 团队签字确认方可进入阶段 2。

### 阶段 2 — 数据层与密钥管理（代码基线）
- **任务树根**：`nd_01KP930A1BTNJZ7DW6XDNG`
- **输出**：
  - `config.go` + `MaBen_UPLOAD_*` ENV + config.yaml 样例；
  - 5 表设计 + goose up/down 迁移 + seed 内置数据；
  - Models（含 JSON 包装类型）+ Repository 层（含 builtin 注册、缓存层、审计）；
  - `SecretCipher`（AES-256-GCM + 版本前缀 + 轮换 CLI + 解密缓存 + 泄露应急手册）。
- **验收**：Repo 单元测试 ≥ 80%；加密写入/读取往返一致；goose up/down 幂等。

### 阶段 3 — 存储驱动层（接入真实厂商）
- **任务树根**：`nd_01KP930A1DS674A56WMQV7`
- **输出**：Driver 接口 + Registry + Local / Aliyun / Tencent / S3 + Custom 骨架 + Contract 测试套件（localstack / MinIO 桩）。
- **验收**：每个 driver 过 Contract 测试（happy path / 异常 / Presigned / STS / 并发）；Local 显式拒绝 direct/STS；回调签名验证到位。

### 阶段 4 — 规则引擎与业务服务（核心能力）
- **任务树根**：`nd_01KP930A1DACQFN8TZDCKK`
- **输出**：
  - `PathTemplateEngine`（DSL + 占位符白名单 + 安全过滤）；
  - `Validator Pipeline`（mime/size/dim/dur/regex/dedup/CEL）；
  - `Processor Pipeline`（压缩/水印/缩略图/转码/杀毒/审核 + 同步/异步分流）；
  - `UploadKeyResolver`（key.rule 语法 / EffectiveConfig 合并 / fallback / 可见性）；
  - `UploadService` 核心（`PrepareUpload / CompleteDirectUpload / PerformRelayUpload / Session`）；
  - 配额 + 限流 + 事件/回调/审计。
- **验收**：端到端（驱动+规则+服务）在单元/集成层通；fallback 行为与矩阵一致。

### 阶段 5 — API 契约 / Handlers / SDK / 管理端
- **任务树根**：`nd_01KP930A1EBH4HTT4CNR14`
- **输出**：
  - OpenAPI（admin + frontend 两套 domain + `x-permission` 注解）；
  - ogen 服务端 + openapi-typescript 客户端；
  - Handlers（Providers/Buckets/UploadKeys/Rules/Records + Frontend Prepare/Complete/Relay/Session + Callback + Stats）；
  - 前端 SDK（`useUpload` / direct / relay / multipart / 扁平 key / 对象参数 / 错误与 fallback 提示）；
  - 前端组件（UploadField / FileBrowser / 进度 + 取消 / WangEditor 迁移 / 旧接口 deprecation 层）；
  - 管理端 UI（Providers/Buckets/UploadKeys/Rules 嵌套 + 排序/Records/Stats/Permission 联动/JSONSchema 向导/i18n）；
  - E2E（STS / PostPolicy / relay / watermark / 不可见 key 403 / key missing fallback / 429 超限）。
- **验收**：7 条 E2E 全绿；前端 typecheck / 后端 `go vet` 全过；无 openapi lint 告警。

### 阶段 6 — 迁移 / 测试 / 运维 / 交付
- **任务树根**：`nd_01KP930A1FM0TEKPDSAYH5`
- **输出**：
  - 旧系统迁移（盘点调用链 / 旧 endpoint→新 SDK 映射 / 老路径回溯兼容 / 老数据写入 `upload_record` / Deprecation 公告）；
  - 测试矩阵（单测 ≥ 80% / 每 driver 集成 / 负载 / 安全 / Playwright 回归 / STS 与回调渗透）；
  - 运维可观测（Metrics + Grafana / 日志 / 告警 / 密钥轮换 SOP / 备份恢复演练 / 灰度发布 / Runbook）；
  - 文档（ADR 汇总 / API reference 自动生成 / 前端集成指南 / 运维手册 / 扩展开发 / CHANGELOG）；
  - 发布（版本号 / 检查表 / 构建产物 / 上线窗口 / 回滚预案 / 48h 盯盘 / 收尾总结）。
- **验收**：灰度→全量→48h 零 P0；旧调用全部切换或 Deprecation 生效。

---

## 7. 里程碑与门禁（Sign-off Gates）

| 门禁 | 时点 | 评审人 | 阻塞条件 |
|---|---|---|---|
| G1 | 阶段 1 完成 | 架构 + 产品 + 安全 | 任一 ADR 未签字 / 契约草案未过 review |
| G2 | 阶段 2 完成 | 后端 Tech Lead + DBA | 覆盖率 < 80% / 密钥往返失败 / 迁移非幂等 |
| G3 | 阶段 3 完成 | 后端 + SRE | 任一 driver 未过 Contract；Local 直传未拒绝 |
| G4 | 阶段 4 完成 | 后端 Tech Lead | fallback 矩阵与 ADR-004 不一致；Pipeline 执行器未过同步/异步分流测试 |
| G5 | 阶段 5 完成 | 全栈 + 产品 | 7 条 E2E 不全绿；openapi lint 有 error |
| G6 | 阶段 6 完成 | SRE + QA + 产品 | 灰度期内出现 P0 / 回滚预案未演练 / 48h 盯盘未完成 |

> 除 G1 外，不允许跳过门禁。G1 可并行起草后续阶段的代码分支，但不可合并 main。

---

## 8. 风险与回滚

| 风险 | 影响 | 预案 |
|---|---|---|
| AK/SK 主密钥泄露 | 全部 provider 凭证需要轮换 | 双主密钥过渡 + 轮换 CLI 批量重加密（阶段 2 内置） |
| Provider 回调签名被伪造 | 任意 object_key 被"认领" | 回调必须验签；`session_id` 一次性；record 落表前二次校验 `sha256` |
| 路径模板注入 | 写入预期外目录 | DSL 白名单占位符 + 黑名单字符过滤（`..`、绝对路径、控制符）|
| 大文件 relay 打爆内存 | 后端 OOM | 流式 + multipart chunked；Local driver 原子 rename；超大默认走 direct |
| 异步处理队列阻塞 | 转码/审核堆积 | 按 key.rule 分队列 + 上限 backpressure + 降级到只做关键校验 |
| 旧系统切换失败 | 业务中断 | 老路径回溯兼容层保留 ≥ 1 版本；SDK 侧暗切换 + Feature Flag |
| OSS 账号被封 / 欠费 | 全线直传失败 | Fallback 到 relay 到备用 bucket（配置级切换，不发版）|

---

## 9. 任务树索引（可直接跳转）

- **任务 ID**：`tsk_01KP930A1038SPNGAMZ912`
- **项目 ID**：`prj_01KP87XR1XEA58497KYTV8`
- **任务 Key**：`upload-system`
- **总节点数**：259（含 6 个 stage 根）

| 阶段 | Stage 节点 ID | 直接子组 | 叶子节点数 |
|---|---|---|---|
| 1 规格定稿 | `nd_01KP930A10DWMKTGFGNDKP` | 需求定稿 / 架构图 / ADR / 契约预审 | 31 + 1 评审 |
| 2 数据层 | `nd_01KP930A1BTNJZ7DW6XDNG` | 配置 / 数据模型 / 迁移 / Models / Repository / 密钥 | 43 + 1 评审 |
| 3 驱动层 | `nd_01KP930A1DS674A56WMQV7` | 接口 / Local / Aliyun / Tencent / S3 / Custom / Registry / Contract | 48 + 1 评审 |
| 4 核心能力 | `nd_01KP930A1DACQFN8TZDCKK` | Path DSL / Validator / Processor / Resolver / Service / 配额 / 事件 | 51 + 1 评审 |
| 5 API/UI | `nd_01KP930A1EBH4HTT4CNR14` | OpenAPI / Handlers / SDK / 组件 / 管理端 / E2E | 50 + 1 评审 |
| 6 交付 | `nd_01KP930A1FM0TEKPDSAYH5` | 迁移 / 测试 / 运维 / 文档 / 发布 | 30 + 1 收官 |

> 完整树可通过 `task_tree_tree_view(task_id, stage_node_id)` 即时拉取。

---

## 10. 术语表

| 术语 | 含义 |
|---|---|
| **Provider** | 一个对象存储账号/本地根（aliyun / tencent / s3 / local） |
| **Bucket** | Provider 下的一个物理容器 |
| **UploadKey（大 Key）** | 业务容器，稳定的对外标识（`ABC`、`USER_AVATAR`） |
| **UploadKeyRule（小 Key）** | Key 下的可切换子策略（`ABC.avatar`） |
| **EffectiveConfig** | Key 默认值 + Rule 覆盖 + 运行时 override 后的最终配置 |
| **Driver Capability** | Driver 声明的能力位：`direct_put / post_policy / sts / callback / multipart` |
| **Upload Mode** | `direct_only / relay_only / both`（Key 级别默认策略） |
| **Path Template DSL** | `{tenant_id}/{yyyy}/{mm}/{uid}/{uuid}/{hash:N}/{ext}` 等白名单占位符 |
| **SecretCipher** | AK/SK 加解密接口，v1 用 AES-256-GCM，预留 KMS |
| **Validator Pipeline** | 上传前的校验链（mime/size/…/CEL） |
| **Processor Pipeline** | 上传后的处理链（压缩/水印/转码/审核） |
| **Builtin** | 系统内置、不可删除的资源（参考 dictionary 模块） |
| **Fallback** | Key / Rule 不存在时的回退策略（`strict / fallback / override`） |

---

## 11. 执行准备检查表（尚未启动）

- [ ] 本文档通过一次团队 review（架构 + 产品 + 安全 + SRE）
- [ ] 确认 `App.capabilities.default_upload_key` 字段与业务 App 清单匹配
- [ ] 确认 env 主密钥派发方式（运维侧 SOP）
- [ ] 确认 Aliyun / Tencent / S3 三家的测试账号已就绪（或使用 MinIO / localstack 替代）
- [ ] Task Tree 259 个节点的依赖排序一次性扫描（保证可执行叶子的起点无空洞）

> **当前状态：任务线（Task Tree）已建成但尚未开跑。** 等 review 通过、上述检查表打勾之后，从阶段 1 的「需求定稿」组开始 claim 第一个可执行叶子。

---

_本文档由 Claude 基于 Task Tree 自动生成并持续同步；如 Task Tree 结构变动，请重新导出。_

