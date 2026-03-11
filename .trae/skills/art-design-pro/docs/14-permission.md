# 权限说明 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/in-depth/permission.html

## 权限控制模式

本系统支持两种权限控制模式：
- **基于角色**：通过接口获取用户角色，控制页面访问和按钮显示权限
- **基于菜单**：通过接口获取菜单列表，依据菜单结构控制页面访问和按钮权限

## 配置方式

权限控制模式通过根目录下的 `.env` 文件配置：

```bash
# 权限控制模式（frontend | backend）
VITE_ACCESS_MODE=frontend
```

- `frontend`：前端控制模式
- `backend`：后端控制模式

## 前端控制模式

### 原理

前端维护菜单列表。用户登录后，接口返回角色标识（如 `R_SUPER`）。前端根据角色遍历菜单列表，若菜单的 `roles` 字段包含该角色，则允许访问对应路由。

### 配置示例

菜单配置文件位于：`/src/router/routes/asyncRoutes.ts`

```typescript
[
  {
    id: 4,
    path: "/system/",
    name: "System",
    component: RoutesAlias.Layout,
    meta: {
      title: "menus.system.title",
      icon: "",
      keepAlive: false,
    },
    children: [
      // 仅 R_SUPER 和 R_ADMIN 角色可访问
      {
        id: 41,
        path: "user",
        name: "User",
        component: RoutesAlias.User,
        meta: {
          title: "menus.system.user",
          keepAlive: true,
          roles: ["R_SUPER", "R_ADMIN"],  // 角色权限
        },
      },
      // 未设置 roles，所有用户可访问
      {
        id: 42,
        path: "role",
        name: "Role",
        component: RoutesAlias.Role,
        meta: {
          title: "menus.system.role",
          keepAlive: true,
        },
      },
    ],
  },
];
```

## 后端控制模式

### 原理

后端生成菜单列表。用户登录后，接口返回菜单数据，前端校验后动态注册路由，实现权限控制。

### 注意事项

- 后端返回的菜单数据结构必须与前端定义一致
- 不需要在菜单配置中设置 `roles` 字段

## 按钮权限控制

### 权限码

权限码适用于前端和后端控制模式：
- **前端控制模式**：登录接口需返回权限码列表
- **后端控制模式**：菜单列表需包含 `authList` 字段

#### 配置示例（后端控制模式）

```typescript
[
  {
    id: 44,
    path: "menu",
    name: "Menus",
    component: RoutesAlias.Menu,
    meta: {
      title: "menus.system.menu",
      keepAlive: true,
      authList: [
        { id: 441, title: "新增", authMark: "add" },
        { id: 442, title: "编辑", authMark: "edit" },
      ],
    },
  },
];
```

#### 使用方式

通过系统提供的 `hasAuth` 方法控制按钮显示：

```typescript
import { useAuth } from "@/composables/useAuth";
const { hasAuth } = useAuth();
```

```vue
<ElButton v-if="hasAuth('add')">添加</ElButton>
```

### 自定义指令（v-auth）

在后端控制模式下，可通过自定义指令 `v-auth` 基于 `authList` 的 `authMark` 控制按钮显示：

```vue
<ElButton v-auth="'add'">添加</ElButton>
```

### 自定义指令（v-roles）

可基于用户信息接口中返回的 `roles` 进行权限控制：

```vue
<el-button v-roles="['R_SUPER', 'R_ADMIN']">按钮</el-button>
<el-button v-roles="'R_ADMIN'">按钮</el-button>
```

## 用户接口示例

```typescript
{
  "userId": "1",
  "userName": "Super",
  "roles": ["R_SUPER"],
  "buttons": ["B_CODE1", "B_CODE2", "B_CODE3"]
}
```

## 前后端控制模式对比

### 前端控制模式
- 适用于角色固定的系统
- 后端角色变更需同步更新前端路由配置
- 实现简单，适合小型项目

### 后端控制模式
- 适用于权限复杂的系统
- 后端返回完整菜单列表，前端动态注册路由
- 更灵活，但需确保前后端数据结构一致
