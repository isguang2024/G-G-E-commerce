# frontend-platform

GGE 多 App 前端平台，基于 [Vue Vben Admin 5](https://doc.vben.pro/) monorepo 改造。Vben 负责壳体机制（布局、路由、权限、请求、状态），不做魔改；后端保持自己的业务契约；中间通过 **`@gge/runtime-*` adapter 包**做翻译，供多个 App 复用。

## 设计原则

- **不魔改 Vben 源码**：`packages/@core/`、`packages/effects/`、`packages/@vben/*` 按上游原样升级。
- **不逼后端兼容 Vben demo**：后端返回自己的 `/runtime/navigation`、`/auth/*`、`/auth/me` 契约。
- **中间放 adapter**：后端契约 → Vben 结构的翻译全部沉淀到 `packages/runtime-*`，**无 UI 依赖**。
- **App 层只放 App 专属**：`pageMap`、`layoutMap`、登录页 UI、`.env`、少量 UI toast 注入。
- **UI 作为回调注入**：共享包接受 `onError / onSuccess` 等回调，App 层传 `ElMessage / Notification`，下个 App 可以传 Naive/Ant 的等价实现。

## 分层

```
frontend-platform/
├── packages/                          上游 + 自研共享层
│   ├── @core/ effects/ @vben/*        Vben 原生，不改
│   ├── runtime-navigation/            后端 manifest → Vben 路由
│   ├── runtime-auth/                  登录 API + centralized login + 归一化
│   ├── runtime-user/                  /auth/me → UserInfo
│   └── runtime-request/               请求客户端 + 刷新 token（注入式）
└── apps/
    ├── web-ele/                       Element Plus 版电商后台（当前）
    │   └── src/
    │       ├── api/          薄包装：调 createXxxApi(requestClient)
    │       ├── store/auth    业务流程粘合（跳转、toast 注入）
    │       ├── router/       pageMap + layoutMap + guards
    │       ├── layouts/      BasicLayout / IFrameView 适配
    │       ├── views/        业务页面
    │       └── locales/      文案
    ├── web-naive/ (未来)      Naive UI 版后台
    └── web-*/ (未来)          其他业务后台
```

## @gge/runtime-* 包清单

| 包 | 职责 | UI 依赖 | 后端契约 |
| --- | --- | --- | --- |
| `@gge/runtime-navigation` | manifest → Vben 路由/菜单 | 无 | `/runtime/navigation` |
| `@gge/runtime-auth` | 登录/刷新/注册/callback/登录页上下文 API + centralized login session | 无 | `/auth/login` `/auth/refresh` `/auth/callback/exchange` `/auth/logout` `/auth/login-page-context` `/auth/register` `/auth/register-context` |
| `@gge/runtime-user` | `/auth/me` → `UserInfo` 归一化 | 无 | `/auth/me` |
| `@gge/runtime-request` | 请求客户端工厂（token、刷新、重认证、错误） | 无（toast 注入） | 依赖 `@vben/request` |

**加新 App 怎么做**：新建 `apps/web-xxx/`，`package.json` 里 `dependencies` 加 `@gge/runtime-*: workspace:*`，在 `api/` 和 `request.ts` 里调用对应的 `createXxxApi(client)`，UI toast 用自己的 UI 库。**不用复制任何登录/请求/用户逻辑代码**。

## 本地开发

```bash
pnpm install
pnpm dev:ele          # 启动 web-ele
pnpm build:ele        # 构建 web-ele
pnpm --filter @vben/web-ele run typecheck
```

App 专属环境变量在 `apps/web-ele/.env*`：

```env
VITE_APP_TITLE=GGE 平台后台
VITE_APP_NAMESPACE=gge-web-ele
VITE_RUNTIME_APP_KEY=platform-admin       # 后端 AppKey，决定加载哪份菜单/权限
VITE_RUNTIME_SPACE_KEY=                   # 空间/租户 key（未来切换空间时用）
```

## 目录约定

- **共享层不得** import 任何 UI 库（element-plus / naive-ui / ant-design-vue）。
- **共享层不得** 直接读取 `preferences` / `useAccessStore` 单例；依赖项通过函数参数注入。
- **App 层** 不得写后端契约翻译逻辑；发现重复代码立即上抽到 `packages/runtime-*`。
- **Vben 壳层** 保持零改动；需要扩展走 slot / meta / 注入回调。

## 当前状态

Step 1（基础 adapter 抽出）已落地：

- ✅ `@gge/runtime-navigation` —— 早期已抽，维持不变
- ✅ `@gge/runtime-auth` —— 新增，auth API + centralized login
- ✅ `@gge/runtime-user` —— 新增，`/auth/me` 归一化
- ✅ `@gge/runtime-request` —— 新增，请求客户端 + 刷新（UI 注入式）
- ✅ `apps/web-ele/src/api/*`、`auth/centralized-login.ts`、`api/request.ts` 已改为薄粘合

⚠️ 环境历史遗留：`pnpm install` 在 `tsdown` 构建 stub 时报错 (`ansis` 解析失败)，与本次重构无关。待 install 能完整跑通后再执行 `pnpm --filter @vben/web-ele run typecheck` 做正式验证。

后续：
- Step 2 —— 抽 `@gge/runtime-access`（路由守卫 + generateAccess 工厂）
- Step 3 —— 新增 `@gge/runtime-context`（workspace/space 响应式 store + header 切换器）

## 相关文档

- Vben 官方：<https://doc.vben.pro/>
- 后端契约：`../backend/api/openapi/`
- 迁移进度：`../docs/change-log.md`
