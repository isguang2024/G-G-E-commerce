# node 1.4：`frontend/src/utils` 与 `frontend/src/types` 未使用候选审查

## 范围与方法
- 审查范围：`frontend/src/utils/**`、`frontend/src/types/**`。
- 检索方式：PowerShell + UTF-8 全局检索（`Get-ChildItem | Select-String -Encoding UTF8`），核对“文件导出符号 -> 真实导入点”。
- 结论口径：
  - `可删`：当前仓库内无任何业务引用，且文件职责为可选/占位，不影响编译主链。
  - `保留`：存在明确业务引用，或属于类型系统/构建链必要声明文件。
  - `需确认后删除`：当前仓库无引用，但可能是对外 API 面/历史兼容入口。

## 可删清单（高置信）
- `frontend/src/utils/ui/iconify-loader.ts`
  - 依据：
    - 仓库内未检索到任何 `iconify-loader` 导入或调用。
    - 文件内容全部为注释示例，未导出运行时代码。
    - `utils/ui/index.ts` 未导出该模块，`main.ts` 也未引入。
  - 建议：可直接删除；若后续需要离线图标，再按需恢复为真实实现。

## 需确认后删除（中置信）
- `frontend/src/utils/socket/index.ts`
  - 依据：
    - 检索未发现 `WebSocketClient` 的外部使用点（仅文件内自引用）。
    - 但该文件通过 `frontend/src/utils/index.ts` 的 `export * from './socket'` 暴露在公共 utils 出口。
  - 风险：
    - 若仓库外部消费者依赖 `@/utils` 的 socket 导出，直接删除可能造成兼容性破坏。
  - 建议：先确认是否存在仓库外消费；若无，可与 `utils/index.ts` 一并收口删除。

## 保留清单（已有明确引用或构建必需）
- `frontend/src/utils/http/v5.ts`
  - 依据：`domains/governance/api/_shared.ts` 引用 `createUnifiedV5HttpError`、`unwrapV5Response`。
- `frontend/src/utils/sys/error-handle.ts`
  - 依据：`main.ts` 引入并执行 `setupErrorHandle(app)`。
- `frontend/src/utils/sys/console.ts`
  - 依据：`main.ts` side-effect 导入 `@utils/sys/console.ts`。
- `frontend/src/utils/table/tableUtils.ts`
  - 依据：`hooks/core/useTable.ts` 大量引用其导出能力。
- `frontend/src/utils/ui/animation.ts`
  - 依据：`App.vue` 与多个布局/登录组件直接引用。
- `frontend/src/types/component/chart.ts`
  - 依据：多个图表组件与 `hooks/core/useChart.ts` 直接引用类型。
- `frontend/src/types/index.ts`、`types/router/index.ts`、`types/store/index.ts`、`types/config/index.ts`、`types/component/index.ts`、`types/common/index.ts`、`types/common/response.ts`
  - 依据：存在 `@/types`、`@/types/router`、`@/types/config`、`@/types/component` 等多处稳定导入链。
- `frontend/src/types/import/auto-imports.d.ts`、`frontend/src/types/import/components.d.ts`
  - 依据：自动生成声明文件，属于构建/IDE 类型提示链路，不能按“业务零导入”处理。
- `frontend/src/types/api/api.d.ts`
  - 依据：全局 `Api` 命名空间声明文件，项目内大量 `Api.*` 类型依赖其全局注入。

## 本轮结论
- 当前高置信“可删”仅 1 项：`utils/ui/iconify-loader.ts`。
- `utils/socket/index.ts` 属“未使用但可能是公共出口”的灰区，建议确认消费面后再删。
- 其余本轮误报候选已通过调用点核验，建议保留。
