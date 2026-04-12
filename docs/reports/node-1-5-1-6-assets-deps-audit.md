# Node 1.5 + 1.6 资源与依赖审计报告

- 审计时间：2026-04-12
- 审计范围：`frontend/public`、`frontend/src/assets`、`frontend/package.json` 依赖
- 执行方式：本地静态扫描 + `git grep` 引用核查 + `pnpm dlx depcheck --json`

## 1) `frontend/public` 资源核查

### 1.1 目录存在性

- 结论：`frontend/public` **不存在**（`Get-ChildItem frontend/public` 报路径不存在）。
- 影响判断：当前项目静态资源不走 `public` 目录，主要位于 `frontend/src/assets`。

### 1.2 疑似过期资源（按源码静态引用核查）

通过 `git grep` 在 `frontend/src` + `frontend/index.html` 检索文件名与相对路径，以下资源未发现引用：

- `frontend/src/assets/images/avatar/avatar7.webp`
- `frontend/src/assets/images/avatar/avatar8.webp`
- `frontend/src/assets/images/avatar/avatar9.webp`
- `frontend/src/assets/images/ceremony/xc.png`
- `frontend/src/assets/images/login/lf_icon2.webp`

说明：

- 以上为“静态未引用”结果，不排除运行时字符串拼接、远程配置下发或历史分支保留用途。
- 当前可作为“候选删除清单”，建议先进入灰度（见第 3 节）。

## 2) 依赖删除风险核查（`depcheck`）

### 2.1 扫描结果概览

执行：`pnpm dlx depcheck --json`（`pnpm exec depcheck` 因未安装 depcheck 不可用，已改用 dlx）。

`depcheck` 标记为“未使用”的依赖：

- dependencies：
  - `@popperjs/core`
  - `@vue/reactivity`
  - `qrcode.vue`
  - `tailwindcss`
- devDependencies：
  - `@commitlint/cli`
  - `@commitlint/config-conventional`
  - `@typescript-eslint/eslint-plugin`
  - `@typescript-eslint/parser`
  - `@vue/compiler-sfc`
  - `cz-git`
  - `stylelint-config-html`
  - `stylelint-config-recess-order`
  - `stylelint-config-recommended-scss`
  - `stylelint-config-recommended-vue`
  - `stylelint-config-standard`
  - `vue-demi`

### 2.2 误报与可信度判定

`depcheck` 同时报告了大量 alias 缺失（如 `@styles/*`、`@imgs/*`、`@views/*`），属于典型路径别名误报；并且存在 `invalidFiles`（`frontend/src/components/core/forms/art-drag-verify/index.vue` 解析报错），会降低未使用结论的可信度。

结合配置文件复核后的判定：

- 高概率误报/不建议删：
  - `tailwindcss`（`src/assets/styles/core/tailwind.css` 中 `@import 'tailwindcss'`，且 `vite.config.ts` 使用 `@tailwindcss/vite`）。
  - `@commitlint/config-conventional`、`@commitlint/cli`（`frontend/commitlint.config.cjs` 与提交流相关）。
  - `cz-git`（`package.json` 的 commitizen path 指向 `node_modules/cz-git`）。
  - `stylelint-config-*`（`frontend/.stylelintrc.cjs` 的 `extends` 直接使用）。
  - `@vue/compiler-sfc`（Vue SFC 编译链常用依赖，建议保留）。
- 可优先复核候选（相对更可能可删）：
  - `@vue/reactivity`（仅在锁文件/`package.json` 出现，源码未检索到直接使用）。
  - `qrcode.vue`（仅见 `src/env.d.ts` module 声明，未见实际 import）。
  - `@typescript-eslint/eslint-plugin`、`@typescript-eslint/parser`（当前 ESLint 配置使用 `typescript-eslint` 聚合包 + `tseslint.parser`，可做一次移除验证）。
  - `@popperjs/core`（可能由 UI 库间接依赖，需验证直接依赖是否必要）。
  - `vue-demi`（常为三方包间接依赖，直接依赖必要性待验证）。

## 3) 可执行建议（按风险分级）

1. 资源侧（低风险先行）  
   先对以下文件做“临时移出 + 全量构建/冒烟”验证，再决定正式删除：
   - `avatar7.webp`、`avatar8.webp`、`avatar9.webp`
   - `ceremony/xc.png`
   - `login/lf_icon2.webp`

2. 依赖侧（分批验证，不一次性删）  
   仅先试删高置信候选：`@vue/reactivity`、`qrcode.vue`、`@typescript-eslint/eslint-plugin`、`@typescript-eslint/parser`。  
   每次删除后立即执行：
   - `pnpm install`
   - `pnpm lint`
   - `pnpm exec vue-tsc --noEmit`
   - `pnpm build`
   若任一步失败，立即回滚该依赖删除。

3. 保守项暂不动  
   `tailwindcss`、`stylelint-config-*`、`commitlint` 相关、`cz-git`、`@vue/compiler-sfc` 先保持不变，避免破坏构建与提交流程。

4. 提升后续扫描准确度  
   先修复 `art-drag-verify/index.vue` 的语法解析问题后，再复跑 depcheck；并在 depcheck 配置中补充 alias 识别/忽略项，减少误报噪音。
