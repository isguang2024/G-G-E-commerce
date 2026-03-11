# 路由和菜单 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/essentials/route.html

## 路由类型

项目中的路由分为两类：**静态路由** 和 **动态路由**。

- **静态路由**：无需权限即可访问的基础页面路由（登录页、404等）
- **动态路由**：需要权限控制的业务页面路由（用户管理、菜单管理等）

## 静态路由

配置位置：`src/router/routes/staticRoutes.ts`

```typescript
export const staticRoutes: AppRouteRecordRaw[] = [
  {
    path: RoutesAlias.Login,
    name: "Login",
    component: () => import("@views/auth/login/index.vue"),
    meta: { title: "menus.login.title", isHideTab: true, setTheme: true },
  },
  {
    path: "/exception",
    component: Home,
    name: "Exception",
    children: [
      {
        path: RoutesAlias.Exception404,
        name: "Exception404",
        component: () => import("@views/exception/404/index.vue"),
        meta: { title: "404" },
      }
    ]
  }
];
```

## 动态路由

配置位置：`src/router/routes/asyncRoutes.ts`

```typescript
export const asyncRoutes: AppRouteRecord[] = [
  {
    name: "Dashboard",
    path: "/dashboard/",
    component: RoutesAlias.Layout,
    meta: {
      title: "menus.dashboard.title",
      icon: "&#xe721;",
    },
    children: [
      {
        path: "console",
        name: "Console",
        component: RoutesAlias.Dashboard,
        meta: {
          title: "menus.dashboard.console",
          keepAlive: false,
          fixedTab: true,
        },
      }
    ],
  }
];
```

## 路由元信息（meta）

```typescript
meta: {
  title: string;           // 路由标题
  icon?: string;           // 路由图标
  showBadge?: boolean;     // 是否显示徽章
  showTextBadge?: string;  // 文本徽章
  isHide?: boolean;        // 是否在菜单中隐藏
  isHideTab?: boolean;     // 是否在标签页中隐藏
  link?: string;           // 外部链接
  isIframe?: boolean;      // 是否为 iframe
  keepAlive?: boolean;     // 是否缓存
  roles?: string[];        // 角色权限
  fixedTab?: boolean;      // 是否固定标签页
  isFullPage?: boolean;    // 是否为全屏页面
  activePath?: string;     // 激活的菜单路径
}
```

## 新建页面步骤

### 1. 创建页面文件

在 `/src/views/` 目录下创建页面：

```vue
<template>
  <div class="page-content">
    <h1>test page</h1>
  </div>
</template>
```

**注意**：使用 `class="page-content"` 让页面高度占满屏幕剩余高度。

### 2. 注册路由

在 `src/router/routes/asyncRoutes.ts` 中添加路由配置：

```typescript
// 一级路由
{
  path: "/test/index",
  name: "Test",
  component: "/test/index",
  meta: {
    title: "测试页",
    keepAlive: true,
  },
}

// 多级路由
{
  name: "Form",
  path: "/form/",
  component: RoutesAlias.Layout,
  meta: {
    title: "表单",
    icon: "&#xe721;",
  },
  children: [
    {
      path: "basic",
      name: "Basic",
      component: "/form/basic",
      meta: {
        title: "基础表单",
        keepAlive: true,
      },
    }
  ],
}
```

### 3. 访问页面

访问 `http://localhost:3006/form/basic` 即可查看新建的页面。
