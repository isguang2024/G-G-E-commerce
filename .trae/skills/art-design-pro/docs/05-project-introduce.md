# 项目结构 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/essentials/project-introduce.html

```
├── src
│   ├── api                     # API 接口相关代码
│   │   ├── articleApi.ts       # 文章相关的 API 接口定义
│   │   ├── menuApi.ts          # 菜单相关的 API 接口定义
│   │   ├── modules             # API 模块化目录
│   │   └── usersApi.ts         # 用户相关的 API 接口定义
│   ├── App.vue                 # Vue 根组件
│   ├── assets                  # 静态资源目录
│   │   ├── fonts               # 字体文件
│   │   ├── icons               # 图标文件
│   │   ├── img                 # 图片资源
│   │   ├── styles              # 全局 CSS/SCSS 样式文件
│   │   └── svg                 # SVG 图标资源
│   ├── components              # 组件目录
│   │   ├── core                # 系统组件（Art Design Pro 组件库）
│   │   └── custom              # 自定义组件（开发者组件库）
│   ├── composables             # Vue 3 Composable 函数
│   │   ├── useAuth.ts          # 认证相关逻辑
│   │   ├── useChart.ts         # 图表相关逻辑
│   │   ├── useCommon.ts        # 通用的 Composable 函数
│   │   └── useTheme.ts         # 主题切换逻辑
│   ├── config                  # 项目配置目录
│   │   ├── assets              # 静态资源配置
│   │   └── index.ts            # 全局配置文件
│   ├── directives              # Vue 自定义指令
│   │   ├── highlight.ts        # 高亮指令
│   │   ├── permission.ts       # 权限指令
│   │   └── ripple.ts           # 波纹效果指令
│   ├── enums                   # 枚举定义
│   ├── locales                 # 国际化（i18n）资源
│   ├── main.ts                 # 项目主入口文件
│   ├── mock                    # Mock 数据目录
│   ├── router                  # Vue Router 路由相关代码
│   │   ├── guards              # 路由守卫
│   │   ├── routes              # 路由定义
│   │   └── utils               # 路由工具函数
│   ├── store                   # Pinia 状态管理
│   │   └── modules             # 分模块的状态管理
│   ├── types                   # TypeScript 类型定义
│   ├── utils                   # 工具函数目录
│   │   ├── browser             # 浏览器相关工具
│   │   ├── constants           # 常量定义
│   │   ├── http                # HTTP 请求工具
│   │   ├── navigation          # 导航相关工具
│   │   └── storage             # 存储相关工具
│   └── views                   # 页面组件目录
├── tsconfig.json               # TypeScript 配置文件
└── vite.config.ts              # Vite 配置文件
```

## 核心目录说明

### components/core/
这是 Art Design Pro 的核心组件库，包含所有可复用的业务组件：
- **表格**：ArtTable, ArtTableHeader
- **表单**：ArtForm, ArtSearchBar, ArtButtonTable
- **图表**：ArtStatsCard, ArtLineChart, ArtBarChart
- **布局**：ArtPageContent, ArtBreadcrumb, ArtHeaderBar

### views/
页面组件目录，所有业务页面都应该放在这里。

### router/routes/
- **staticRoutes.ts**: 静态路由（登录、404等）
- **asyncRoutes.ts**: 动态路由（业务页面）

### config/
项目配置，包括系统名称、主题配置等。
