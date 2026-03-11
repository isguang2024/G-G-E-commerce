# 环境变量配置 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/essentials/env-variables.html

## 说明

环境变量位于项目根目录下：
- `.env` - 适用于所有环境
- `.env.development` - 仅适用于开发环境
- `.env.production` - 仅适用于生产环境

## .env

- **作用**：适用于所有环境，里面定义的变量会在任何环境下都能访问
- **用法**：一般放置一些通用的配置，比如 API 基础地址、应用名称等

## .env.development

- **作用**：仅适用于开发环境。当你运行 `pnpm dev` 时，Vue 会加载这个文件中的环境变量
- **用法**：适合放置开发阶段的配置，比如本地 API 地址、调试设置等

## .env.production

- **作用**：仅适用于生产环境。当你运行 `pnpm build` 时，Vue 会加载这个文件中的环境变量
- **用法**：适合放置生产阶段的配置，比如生产 API 地址、禁用调试模式等

## 自定义环境变量

**自定义环境变量以 `VITE_` 开头**

比如：`VITE_PORT`

你可以在项目代码中这样访问它们：
```typescript
console.log(import.meta.env.VITE_PORT);
```

## 环境配置说明

### .env（通用）
```bash
# 版本号
VITE_VERSION = 2.4.1.1

# 端口号
VITE_PORT = 3006

# 网站地址前缀
VITE_BASE_URL = /art-design-pro/

# API 地址前缀
VITE_API_URL = https://m1.apifoxmock.com/m1/6400575-6097373-default

# 权限模式（frontend | backend）
VITE_ACCESS_MODE = frontend

# 跨域请求时是否携带 Cookie
VITE_WITH_CREDENTIALS = false

# 是否打开路由信息
VITE_OPEN_ROUTE_INFO = false

# 锁屏加密密钥
VITE_LOCK_ENCRYPT_KEY = jfsfjk1938jfj
```

### .env.development（开发环境）
```bash
# 网站地址前缀
VITE_BASE_URL = /

# API 请求基础路径（开发环境通常为代理前缀）
VITE_API_URL = /api

# 本地开发代理的目标后端地址
VITE_API_PROXY_URL = https://m1.apifoxmock.com/m1/6400575-6097373-default

# Delete console
VITE_DROP_CONSOLE = false
```

### .env.production（生产环境）
```bash
# 网站地址前缀
VITE_BASE_URL = /art-design-pro/

# API 地址前缀
VITE_API_URL = https://m1.apifoxmock.com/m1/6400575-6097373-default

# Delete console
VITE_DROP_CONSOLE = true
```
