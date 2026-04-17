# 前端可观测性规范（Frontend Observability Spec）

> 统一 MaBen 多租户 SaaS 平台前端的三件事：
> 1. 关键结果区的 `data-testid` 命名约定（供 E2E/手工观测用）
> 2. 前端表单错误语义（`el-form-item` + `data-testid` + 后端结构化 error）
> 3. 后端结构化 `error.code` / `error.details.field` 扩展指南
>
> **适用范围**：一切新写的 Vue 页面、整改中的高风险页面（`docs/archive/high-risk-remediation-matrix.md`）
> 其他业务域推荐照抄，本规范是最低约束。
>
> **真源**：
> - 后端码表：`backend/internal/api/apperr/codes.go`（新增/修改后运行 `go run ./cmd/gen-permissions`）
> - 前端码表：`frontend/src/api/v5/error-codes.ts`（AUTO-GENERATED，禁止手改）
> - Error schema：`backend/api/openapi/components/common.yaml` `Error` 定义
> - 前端消息展示策略：`frontend/src/utils/http/error.ts`

---

## 1. `data-testid` 命名约定

### 1.1 为什么需要 `data-testid`

`tsk_01KP5BKHXP6MYKWC9BKMR1` 深测时使用了基于文本/CSS 选择器的通用脚本——8/8 页面可进入，但"保存校验文本"、"发送预览体"、"访问链路节点"这类关键状态都**没有稳定抓手**，截图 diff 是唯一证据。
引入 `data-testid` 是最低成本的提升：

- Playwright/WebdriverIO 的 `page.getByTestId(...)` 命中稳定，不随 i18n 文案变化
- 人工排障时也能在 DevTools `$('[data-testid=...]')` 秒定位

### 1.2 命名规则

格式：`<scope>-<node>[-<modifier>]`

| 段 | 含义 | 示例 |
|----|------|------|
| `scope` | 页面/领域前缀，使用短横线单词 | `register-entry`、`login-template`、`api-sync`、`trace`、`send` |
| `node`  | 节点类型，固定名词 | `row`、`cell`、`field-error`、`preview`、`status`、`summary`、`node` |
| `modifier` | 可选修饰，通常是业务键 | `[key=foo]`、`[app=crm]`、`[step=2]` |

- 全小写、短横线连字、**不要混用下划线/驼峰**。
- 业务键放入属性值（例如 `data-testid="login-template-row" data-key="default"`），**不要塞进 `data-testid` 自身**，否则选择器变脆。

### 1.3 高风险整改专用基线表

下列 `data-testid` 名称直接与 `docs/archive/high-risk-remediation-matrix.md` 中的整改节点一一对应，整改 PR 必须落地到这一套命名：

| 整改节点 | 页面 | `data-testid` | 用途 |
|----------|------|---------------|------|
| `fix-register-entry` | `/system/register-entry` | `register-entry-field-error` | 表单字段错误挂点（`el-form-item` 内） |
| `fix-register-entry` | 同上 | `register-entry-row` | 保存后新行定位（配合 `data-code=<code>`） |
| `fix-login-template` | `/system/login-page-template` | `login-template-row` | 模板列表行（配合 `data-key=<key>`） |
| `fix-login-template` | 同上 | `login-template-preview` | 实时预览容器 |
| `fix-app-management` | `/system/app` | `app-card` | App 卡片（配合 `data-app-key=`） |
| `fix-app-management` | 同上 | `app-field-error` | 认证配置表单字段错误 |
| `fix-menu-space` | `/system/menu-space` | `menu-space-card` | 空间卡片 |
| `fix-menu-space` | 同上 | `host-conflict-reason` | Host 冲突结构化提示 |
| `fix-access-trace` | `/system/access-trace` | `trace-summary` | 汇总统计区 |
| `fix-access-trace` | 同上 | `trace-node` | 链路节点（配合 `data-node-id=`） |
| `fix-api-management` | `/system/api-endpoint` | `api-sync-result` | 同步/注册结果摘要 |
| `fix-api-management` | 同上 | `api-endpoint-row` | 端点列表行 |
| `fix-system-message-required` | `/system/message` | `send-field-error` | 发送表单必填错误 |
| `fix-system-message-confirm` | 同上 | `send-confirm-dialog` | 二次确认对话框 |
| `fix-system-message-confirm` | 同上 | `send-status` | 发送状态标签 |
| `fix-collab-message` | `/collaboration/message` | `send-preview`、`send-status` | 同上，协作态版本 |

### 1.4 例外与禁忌

- 通用组件（`art-button`, `el-input` 等）**不要**在内部私挂 `data-testid`，由调用方按 scope 决定
- **禁止**用 `data-testid` 作为 CSS hook 或事件绑定 key——那会耦合测试与样式
- 动态列表使用 `data-testid` 常量 + 业务键属性组合（见 1.2），不要生成 `foo-row-<id>` 这样的变长 id

---

## 2. 表单错误语义（`el-form-item` + 结构化 error）

### 2.1 目标

深测中 4/4 配置页 "保存校验尝试时未捕获明确的必填校验文本"。根因不是后端没报错，而是：

1. 前端表单没有挂 `rules`，或挂了但 `el-form-item` 没有 `prop` 绑定
2. 后端 400 返回的只是 `{code: 1001, message: "参数错误"}`，没有 `details.field`，前端无处回显

规范要求：**前端 `el-form` 负责格式/必填校验；后端业务校验失败必须回 `details.field`；两端通过 `el-form-item` 的 `error` 属性统一展示**。

### 2.2 Error schema 约定（已在 `common.yaml` 定义）

```yaml
Error:
  type: object
  required: [code, message]
  properties:
    code:
      type: integer    # 1xxxx 参数 · 2xxxx 认证 · 3xxxx 业务 · 5xxxx 服务端
    message:
      type: string     # 面向用户的整体提示
    details:
      type: object
      additionalProperties: true
      nullable: true
      description: 可选上下文，如参数校验失败时携带 {field: reason}
```

**业务语义**：当且仅当 `code ∈ 1xxxx` 或 `code ∈ {CodeConflict 及其派生}` 时，前端必须优先检查 `details.field`，逐字段回显；其他码走全局 `ElMessage`（由 `utils/http/error.ts` 的 `shouldShowErrorMessage` 决定）。

### 2.3 后端侧：返回 `details.field` 的固定写法

在 handler 参数/前置校验处直接挂 `apperr.ParamError`（已有）不足以表达字段，需要新增一个字段版哨兵。建议扩展 `backend/internal/api/apperr/mapper.go`：

```go
// FieldError 表达"单个字段级"校验失败。mapper 翻译为 400 + Error.details.<field>=<reason>。
type FieldError struct {
    Field  string
    Reason string
    Msg    string // 面向用户的整体提示；留空则使用 Reason
}

func (e *FieldError) Error() string {
    if e.Msg != "" { return e.Msg }
    return e.Reason
}
```

在 `doMap` 的哨兵分支加一段：

```go
var fe *apperr.FieldError
if errors.As(err, &fe) {
    details := map[string]any{fe.Field: fe.Reason}
    msg := fe.Msg
    if msg == "" { msg = "参数校验失败" }
    return mapped{http.StatusBadRequest, &gen.Error{
        Code:    CodeParamInvalid,
        Message: msg,
        Details: gen.NewOptNilErrorDetails(details), // 由 ogen 生成的可选包装
    }}
}
```

> `gen.NewOptNilErrorDetails` 的实际名称以生成产物为准；参见 `backend/api/gen/` 对 `Error` 的封装。

Handler 层调用示例（register-entry 新建入口，A1 整改用）：

```go
func (h *APIHandler) CreateRegisterEntry(ctx context.Context, req *gen.CreateRegisterEntryRequest) (*gen.RegisterEntry, error) {
    if strings.TrimSpace(req.Code) == "" {
        return nil, &apperr.FieldError{Field: "code", Reason: "必填", Msg: "入口 Code 不能为空"}
    }
    if !registerEntryCodePattern.MatchString(req.Code) {
        return nil, &apperr.FieldError{Field: "code", Reason: "格式不合法", Msg: "入口 Code 仅允许小写字母数字和短横线"}
    }
    // ... 业务层冲突：
    existing, _ := h.RegisterEntry.FindByCode(ctx, req.Code)
    if existing != nil {
        return nil, &apperr.FieldError{Field: "code", Reason: "已存在", Msg: "入口 Code 已存在"}
    }
    // ...
}
```

> **新错误码怎么加**：
> 1. `codes.go` 增补常量（保持段号一致：1xxxx/2xxxx/3xxxx/5xxxx）
> 2. `mapper.go` 增 case
> 3. `go run ./cmd/gen-permissions` 重新派生 `frontend/src/api/v5/error-codes.ts`
> 4. 如果是配置类/表单类场景，优先复用 `FieldError`，不要每字段单独一个 code

### 2.4 前端侧：`el-form-item` 统一挂载

**核心原则**：`el-form-item` 必须有 `prop`；错误文案通过 `error` 属性注入（而不是 `rules` 的副作用）；同时在 `el-form-item` 内挂 `data-testid="<scope>-field-error"` 包住错误容器，方便 E2E 抓错。

通用模板（Composition API + `<script setup>`）：

```vue
<template>
  <el-form
    ref="formRef"
    :model="form"
    :rules="rules"
    label-width="120px"
    data-testid="register-entry-form"
  >
    <el-form-item
      label="入口 Code"
      prop="code"
      :error="fieldErrors.code"
      data-testid="register-entry-field-error"
      :data-field="'code'"
    >
      <el-input v-model="form.code" placeholder="例如 default" />
    </el-form-item>

    <el-form-item
      label="名称"
      prop="name"
      :error="fieldErrors.name"
      data-testid="register-entry-field-error"
      :data-field="'name'"
    >
      <el-input v-model="form.name" />
    </el-form-item>

    <el-form-item>
      <el-button type="primary" :loading="saving" @click="submit">保存</el-button>
    </el-form-item>
  </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { v5Client } from '@/api/v5/client'
import { ErrorCodes } from '@/api/v5/error-codes'

const formRef = ref<FormInstance>()
const saving = ref(false)

const form = reactive({ code: '', name: '' })

const rules: FormRules = {
  code: [{ required: true, message: '请输入入口 Code', trigger: 'blur' }],
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

// 后端字段错误回显容器：字段名 → 文案
const fieldErrors = reactive<Record<string, string>>({})

function clearFieldErrors() {
  for (const k of Object.keys(fieldErrors)) delete fieldErrors[k]
}

async function submit() {
  clearFieldErrors()
  if (!(await formRef.value?.validate().catch(() => false))) return

  saving.value = true
  try {
    // openapi-fetch: 返回 { data, error, response }
    const { data, error } = await v5Client.POST('/register-entries', {
      body: { code: form.code, name: form.name }
    })

    if (error) {
      // 1xxxx 参数/业务字段错误：逐字段回显
      if (
        (error.code === ErrorCodes.ParamInvalid || error.code === ErrorCodes.Conflict) &&
        error.details &&
        typeof error.details === 'object'
      ) {
        for (const [field, reason] of Object.entries(error.details as Record<string, string>)) {
          fieldErrors[field] = reason
        }
        return
      }
      // 其他码：交给全局 ElMessage（由 utils/http/error.ts 统一显示）
      throw error
    }

    // data 命中 200/201 分支
    console.info('created', data)
  } finally {
    saving.value = false
  }
}
</script>
```

关键点：

- `el-form-item` **同时**挂 `prop`（前端 rules）与 `:error`（后端字段错误）——两者不冲突，`error` 优先级更高
- `fieldErrors` 在每次提交前清空，避免"上次失败的 code 错误残留"
- 即便后端返回 `CodeConflict(3013)`，只要带 `details.field` 也走字段回显分支——重复名/冲突是最常见的字段场景
- 全局错误由 `shouldShowErrorMessage` 决定是否 toast，页面不重复弹

### 2.5 E2E 断言模板

Playwright 片段（配合 1.3 的命名表）：

```ts
// 必填校验：直接点保存
await page.getByRole('button', { name: '保存' }).click()
await expect(
  page.getByTestId('register-entry-field-error').filter({ has: page.locator('[data-field="code"]') })
).toContainText('请输入入口 Code')

// 后端字段错误：提交已存在的 code
await page.getByLabel('入口 Code').fill('existing-one')
await page.getByRole('button', { name: '保存' }).click()
await expect(
  page.getByTestId('register-entry-field-error').filter({ has: page.locator('[data-field="code"]') })
).toContainText('已存在')
```

---

## 3. 后端结构化 error 扩展指南

### 3.1 新增错误码的触发条件

- **必须新码**：出现了现有枚举无法表达的**稳定业务语义**（例如 `host_conflict`、`space_key_invalid`）
- **不要新码**：只是文案不同但语义相同（例如 `ParamInvalid` 已覆盖）——换 `Message` 即可

### 3.2 三步扩码（按 `apperr/codes.go` 顶部注释）

1. **加常量**：`backend/internal/api/apperr/codes.go`，按段号放置，注释写"面向用户的含义"。例如：
   ```go
   // 3xxxx 业务 / 资源
   CodeHostConflict     = 3022 // 入口 Host 已被占用
   CodeSpaceKeyInvalid  = 3023 // 协作空间 Key 不合法
   ```
2. **加翻译**：`backend/internal/api/apperr/mapper.go`，加 `errors.Is(err, xxxModule.ErrHostConflict)` 分支，返回对应 `CodeXxx`。
3. **重新派生前端码表**：
   ```bash
   cd backend && go run ./cmd/gen-permissions
   ```
   产物 `frontend/src/api/v5/error-codes.ts` 带 `AUTO-GENERATED` 标记，不要手改。

### 3.3 何时用 `FieldError`，何时用独立 code

| 场景 | 选 |
|------|----|
| 单字段必填/格式/长度问题 | `FieldError{Field:"x", Reason:"..."}`，不新增 code |
| 多字段组合冲突但想一键展示 | 新增 3xxxx code，`message` 写整体提示，`details` 可选地挂多字段 |
| 资源级冲突需要前端分支处理（例如禁用某按钮） | 新增 3xxxx code，前端用 `ErrorCodes.Xxx` 判断 |
| 权限/认证问题 | 永远在 2xxxx 段，不要走 `FieldError` |

### 3.4 `details` 的兼容规则

- `details` **始终可选**。老的 handler 只返 `{code, message}` 合法，不需要回填。
- 新加 `details` 字段**只能新增 key**，删 key 等于破坏契约，必须走 spec 兼容流程（参见 `backend/api/openapi/README.md`）。
- `details` 的 key 名统一 snake_case（与数据库列/API JSON 一致）。前端 `el-form-item` 的 `prop` 如果是 camelCase，由页面自己做 `snake_case → camelCase` 的映射（建议用 `lodash.camelCase` 单次转换）。

---

## 4. 自查清单（新页面 / 整改 PR 提交前）

| # | 检查项 | 证据 |
|---|--------|------|
| 1 | 页面关键结果区（表格行、表单错误、状态标签、预览区）都挂了 `data-testid` | grep PR diff |
| 2 | `data-testid` 遵循 `<scope>-<node>[-<modifier>]` 命名，未混驼峰/下划线 | grep |
| 3 | 所有 `el-form-item` 都有 `prop`，且错误通过 `:error="fieldErrors[x]"` 回显 | 代码审查 |
| 4 | 提交前 `fieldErrors` 清空；后端返 `details.field` 能被正确映射 | 手动/E2E |
| 5 | 新增 error code 已经 `go run ./cmd/gen-permissions` 并提交 `error-codes.ts` | git status |
| 6 | 若新 code 需要 toast，`utils/http/error.ts` 的 `showCodes/hideCodes` 已更新 | 代码审查 |
| 7 | 至少一个 Playwright 断言使用 `getByTestId` 命中新挂载的 `data-testid` | E2E spec diff |

---

## 5. 前端日志契约（telemetry）

> 本节与 `docs/guides/logging-spec.md` §4–§5 对齐。日志系统的完整契约（后端 zap、
> audit、ingest 限流、表结构）见 logging-spec.md；本节只抽出**前端页面侧必须遵守**的
> 最小约束。

### 5.1 入口：一个 logger，零 `console.*`

- 业务代码 **不得** 写 `console.log / console.warn / console.error`，所有日志统一走
  `import { logger } from '@/utils/logger'`。
- `console.*` 只出现在 `frontend/src/utils/logger/index.ts` 内部（开发体验用），
  其他文件里出现即视为规范违反。
- `debug` 级别只打控制台，`info/warn/error` 批量上报 `/api/v1/telemetry/logs`。

### 5.2 事件名（event）命名

- 第一个参数是稳定的 `dot-case` 事件名：`domain.entity.action[.outcome]`
- **不要** 传面向用户的句子（"保存失败"），不要拼业务变量进 event；上下文放第二参数 map
- 常见事件段命名约定：

  | 段前缀        | 语义                            | 示例                           |
  |---------------|---------------------------------|--------------------------------|
  | `http.`       | HTTP 请求相关                   | `http.error`、`http.request_cancelled` |
  | `sys.`        | Vue / 脚本 / Promise / 资源错误 | `sys.vue_error`、`sys.promise_rejection` |
  | `<page>.`     | 页面级业务事件                   | `register-entry.save_success`  |
  | `<domain>.`   | 跨页领域事件                     | `auth.login_denied`            |

### 5.3 上报字段（snake_case）

```ts
logger.error('http.error', {
  url: '/api/v1/apps',
  status: 500,
  err: error,          // Error 对象会被自动序列化进 entry.error
  // 其他业务字段...
})
```

- 字段命名 **snake_case**（例如 `request_id`、`session_id`），与后端 OpenAPI
  `TelemetryLogEntry` 对齐；不要用 camelCase
- 敏感字段（`password / token / authorization / cookie / phone / ...`）由 logger
  内部自动脱敏替换为 `[REDACTED]`；**新增敏感 key 必须同步更新前后端两套**
  （`utils/logger/index.ts` `REDACT_KEYS` + `backend/.../audit/redact.go` `DefaultRedactFields`）
- Error 对象可以直接作为 `logger.error('...', err)` 的第二参数传入

### 5.4 用户 / 路由 hook

以下两处是规范要求的单一挂点，不得在其他文件重复注入：

| 触发          | 代码位置                                  | 调用                                |
|---------------|-------------------------------------------|-------------------------------------|
| 登录成功/切换 | `frontend/src/domains/auth/store.ts` `setUserInfo` | `logger.setUser(userId)`          |
| 登出          | 同上，`clearSessionState` 末尾            | `logger.setUser('')`                |
| 路由切换      | `frontend/src/router/guards/afterEach.ts` | `logger.setRoute(to.fullPath)`      |

### 5.5 HTTP 错误的统一出口

- 所有 axios 异常 / `openapi-fetch` error 分支由 `frontend/src/utils/http/error.ts` 统一
  调用 `logger.error('http.error', {...})`；业务页面 **不要** 再重复上报同一份错误
- 仅当业务需要加独有字段时再在页面里额外 `logger.info/warn` 一条业务事件（不要重复 error）
- ERR_CANCELED / 用户主动取消：用 `logger.debug('http.request_cancelled', {...})`，不产生 toast，
  也不上报到生产端点

### 5.6 页面级业务事件示例

```ts
// 保存成功
logger.info('register-entry.save_success', { code, mode: 'create' })

// 打开敏感操作确认框
logger.info('system-message.send_confirm_open', { dispatchId })

// 本地前置校验失败（不是 HTTP 错误，不要 logger.error）
logger.warn('register-entry.validate_failed', { field: 'code', reason: 'required' })
```

### 5.7 自查清单（附加到本文件 §4 的通用清单）

| # | 检查项                                                          | 证据            |
|---|-----------------------------------------------------------------|-----------------|
| A | 页面所有错误 / 异常处理路径走 `logger.error`，无 `console.*`     | grep diff       |
| B | 事件名遵循 `dot-case`，无中文句子 / camelCase                    | grep            |
| C | 上报字段使用 snake_case（`session_id` 而非 `sessionId`）          | 代码审查         |
| D | 新增敏感字段同时加到前端 `REDACT_KEYS` 与后端 `DefaultRedactFields` | diff         |
| E | 没有在业务组件里直接订阅或 import logger 的内部类 `Logger`       | grep            |

---

## 6. 相关文档

- `docs/guides/logging-spec.md`——全栈日志系统真源（zap / audit / telemetry / ingest）
- `docs/archive/high-risk-remediation-matrix.md`——驱动本规范落地的 8 页整改清单
- `docs/API_OPENAPI_FIXED_FLOW.md`——新接口 spec 变更的固定步骤
- `backend/api/openapi/README.md`——OpenAPI 多文件结构与错误码体系真源
- `backend/internal/api/apperr/codes.go`——错误码唯一真源
- `frontend/src/utils/http/error.ts`——前端全局错误显示策略（`showCodes`/`hideCodes`）
- `frontend/src/utils/logger/index.ts`——前端 logger 单例、批量传输、脱敏实现

