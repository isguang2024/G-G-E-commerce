# 开发必读文档 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/must-read.html

## 接口对接

默认返回以下格式，如需修改请到 `src/typings/http.d.ts` 文件修改：

```typescript
interface BaseResponse<T = unknown> {
  code: number;    // 状态码
  msg: string;     // 消息
  data: T;         // 数据
}
```

## 网络请求

默认返回 data 中的数据而不是整个响应体：

```typescript
try {
  const { token, refreshToken } = await fetchLogin({
    userName: username,
    password,
  });
} catch (error) {
  if (error instanceof HttpError) {
    // 这里可以根据状态码进行不同的处理
  }
}
```

## 菜单数据（asyncRoutes.ts）

- `RoutesAlias.Layout` 指向的是布局容器
- 后端返回的菜单数据中，component 字段需要指向 `/index/index`
- `roles` 字段用于前端控制模式

**前端模式**：通过获取用户信息接口返回的 roles 跟菜单数据 asyncRoutes 中的 roles 进行对比实现菜单过滤

**后端模式**：直接通过接口返回对应角色的菜单即可，不需要返回 `roles` 字段

示例：
```typescript
{
  name: 'Dashboard',
  path: '/dashboard',
  component: RoutesAlias.Layout,
  meta: {
    title: 'menus.dashboard.title',
    icon: '&#xe721;',
    roles: ['R_SUPER', 'R_ADMIN']  // 前端模式需要
  },
  children: [
    {
      path: 'console',
      name: 'Console',
      component: RoutesAlias.Dashboard,
      meta: {
        title: 'menus.dashboard.console',
        keepAlive: false,
        fixedTab: true
      }
    }
  ]
}
```

## 打包大小说明

- **完整版项目**：约 10MB
- **精简版项目**：约 5MB

项目默认开启 gzip 压缩，因此会额外生成 .gz 文件：

关闭 gzip 时，实际打包体积约 4.5MB

开启 gzip 后，产物体积更小（浏览器请求时会优先加载 .gz 文件）

**进一步优化方案**：

若对体积有更高要求，可通过以下方式优化，可轻易降至 3.5MB 左右：
- 精简或替换图标库
- 移除非必要图片资源
- 减少第三方库依赖，或替换为更轻量的方案
