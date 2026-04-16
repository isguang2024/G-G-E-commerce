# 上传配置中心二期交付说明

> 对应任务：`tsk_01KPBFG73F2AG23F2JHDXM`
> 日期：2026-04-17
> 状态：交付版

## 1. 本次交付范围

本次交付补齐了上传配置中心二期的最小闭环：

- 数据层新增 UploadKey / Rule 运行时字段：
  - `upload_mode`
  - `is_frontend_visible`
  - `permission_key`
  - `fallback_key`
  - `client_accept`
  - `direct_size_threshold_bytes`
  - `extra_schema`
  - `mode_override`
  - `visibility_override`
- OpenAPI 新增前端安全可见上传配置接口：
  - `GET /api/v1/media/upload-keys`
- 后端上传服务支持：
  - 基于 `EffectiveConfig` 决策 `direct / relay / fallback`
  - 基于 `permission_key` 过滤前端可见 UploadKey
  - UploadKey / Rule 的 `extra_schema` 归一化与约束校验
- 前端上传 SDK 支持：
  - 获取可见 UploadKey 列表
  - 解析可见目标
  - 回传 `requestedMode / resolvedMode / fallbackUsed`
- 管理端支持：
  - Provider / Bucket / UploadKey / Rule 分层引导式配置
  - 按 driver 显示差异化字段
  - `extra_schema` 结构化编辑

## 2. 上线步骤

### 2.1 前置确认

- 确认后端配置已具备上传密钥加密所需密钥材料。
- 确认数据库可执行迁移。
- 确认对象存储配置中，回调地址、STS、域名等外部依赖已经在目标环境准备完成。

### 2.2 推荐上线顺序

1. 部署包含新模型、迁移和 OpenAPI 生成物的后端版本。
2. 执行数据库迁移：

```powershell
cd backend
go run ./cmd/migrate
```

3. 确认迁移版本包含 `00032_upload_config_runtime_fields.sql`。
4. 部署前端管理台与上传 SDK 版本。
5. 进入 `/system/upload-config` 补齐或核对 UploadKey / Rule 新字段。
6. 用管理台或现有业务上传入口验证一次：
   - relay 上传
   - direct 上传
   - fallback 到 relay

## 3. 数据迁移说明

### 3.1 结构变更

迁移文件：

- `backend/internal/pkg/database/migrations/00032_upload_config_runtime_fields.sql`

本次迁移只做结构补齐，不写长期默认业务数据。新增列均带安全默认值：

- `upload_keys.upload_mode = 'auto'`
- `upload_keys.is_frontend_visible = false`
- `upload_keys.permission_key = ''`
- `upload_keys.fallback_key = ''`
- `upload_keys.client_accept = []`
- `upload_keys.direct_size_threshold_bytes = 0`
- `upload_keys.extra_schema = {}`
- `upload_key_rules.mode_override = 'inherit'`
- `upload_key_rules.visibility_override = 'inherit'`
- `upload_key_rules.client_accept = []`
- `upload_key_rules.extra_schema = {}`

### 3.2 兼容性影响

- 老数据迁移后默认仍可继续工作：
  - UploadKey 默认为 `auto`
  - 旧管理页未配置的新字段会走默认值
  - 运行时仍可继续使用 relay 上传
- 旧前端如果仍只调用 `prepare / complete / upload`，不会因为新增字段直接中断。
- 新增的 `GET /media/upload-keys` 是增量接口，不会破坏旧调用方。

### 3.3 配置补齐建议

迁移完成后建议优先补以下字段，否则新增能力不会显式生效：

- 对需要被业务前端发现的 UploadKey：
  - `is_frontend_visible = true`
  - `permission_key` 设置为对应业务权限键，或留空表示仅受媒体主权限控制
- 对需要优先 relay 的入口：
  - `upload_mode = relay`
- 对需要超阈值自动转 relay 的入口：
  - `direct_size_threshold_bytes > 0`
- 对需要规则级覆盖的场景：
  - 在 Rule 上补 `mode_override` / `visibility_override`

## 4. 回滚策略

### 4.1 应用层回滚

如果只是页面或上传体验异常，优先回滚应用版本，不立刻回滚数据库结构。

原因：

- 新列均带默认值，旧代码通常不会因为数据库多列而崩溃。
- 先回滚应用版本，能最大限度缩短故障窗口。

### 4.2 数据库回滚

确需回滚结构时，可执行 `00032` 的 Down：

- 删除 UploadKey 新增字段
- 删除 Rule 新增字段

注意事项：

- 若上线后已使用这些字段写入了业务配置，执行 Down 会丢失对应配置。
- 因此数据库回滚前，应先导出受影响的 `upload_keys` / `upload_key_rules` 数据。

### 4.3 配置回退建议

比起直接回滚结构，更推荐使用配置回退：

- 关闭 `is_frontend_visible`
- 将 `upload_mode` 统一切回 `relay`
- 清空有问题的 `fallback_key`
- 暂停有问题的 Provider / Bucket / UploadKey / Rule

## 5. 权限与安全审计结论

本次按代码链路做了一轮烟测式审计，当前未发现阻塞交付的高危泄露问题，结论如下。

### 5.1 管理面接口

- `storage_admin` 全量接口统一挂在 `system.upload.config.manage` 权限下。
- 管理面 Provider 列表与详情返回的是脱敏后的 `access_key_masked / secret_key_masked`，不是明文。
- Provider 更新时传入的 `access_key` / `secret_key` 会在仓储层加密后再落库。

### 5.2 前端可见上传配置接口

- `GET /api/v1/media/upload-keys` 必须已认证。
- 服务层会先过滤 `is_frontend_visible = true` 且 `status = ready` 的 UploadKey。
- 若 UploadKey 声明了 `permission_key`，还会与当前用户已授权权限键做交集校验。
- 返回字段仅包含：
  - `key`
  - `name`
  - `uploadMode`
  - `visibility`
  - `clientAccept`
  - `maxSizeBytes`
  - `directSizeThresholdBytes`
  - `fallbackKey`
  - `rules`
- 不返回 Provider / Bucket 的明文配置，也不返回 AK / SK / STS 密钥材料。

### 5.3 直传与 callback

- OSS callback 校验器已校验公钥地址来源与签名合法性。
- bucket 级 callback 配置只参与直传表单组装，不会通过前端可见列表直接暴露。

### 5.4 当前剩余风险

- 运行时上传接口与可见列表目前仍挂在 `system.media.manage` 下，适合后台管理场景；若后续要开放给更广泛业务前端，需要再拆细读写权限。
- 本轮安全烟测以代码审计和现有测试为主，未额外补浏览器态与外网对象存储联调。

## 6. 验证记录

本次功能交付过程中已完成：

- `cd backend && go run ./cmd/migrate`
- `cd backend && go test ./internal/modules/system/upload -count=1`
- `cd backend && go test ./internal/api/handlers -count=1`
- `cd frontend && pnpm exec vue-tsc --noEmit`

前端相关文件的 `eslint` 已无 error，仍有历史 `no-explicit-any` warning，未在本次任务中顺手清理。

## 7. 发布检查清单

- 后端迁移已执行且版本正确。
- UploadKey 新字段已补齐到目标场景。
- 需要前端发现的 UploadKey 已打开 `is_frontend_visible`。
- 需要细粒度鉴权的 UploadKey 已配置 `permission_key`。
- OSS 场景已验证 relay / direct / fallback 三类链路。
- 管理台已验证 Provider / Bucket / UploadKey / Rule 的保存与回显。

## 8. 后续建议

- 将 `system.media.manage` 进一步拆为“上传使用”和“媒体管理”，降低业务前台接入门槛。
- 为 `COS / S3` 增加 driver registry、默认 `extra` 模板和管理页引导卡片。
- 若后续需要真正按 `extra_schema` 录入业务值，再为 UploadKey / Rule 增加独立 `extra` 存值槽。
