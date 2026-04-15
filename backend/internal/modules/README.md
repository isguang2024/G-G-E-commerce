# backend/internal/modules

`internal/modules` 是后端领域实现层，负责承载业务 service、领域模型访问和模块内的运行时编排。

## 当前目录

| 目录 | 说明 |
| --- | --- |
| `system/` | 系统主域，覆盖认证、用户、角色、菜单、页面、注册、空间、权限等基础能力 |
| `observability/` | 可观测性域，覆盖业务审计日志和前端日志摄取 |

## 使用原则

- 新业务优先判断能否复用 `system/` 里的认证、权限、空间上下文、页面和 API 注册能力。
- 模块文档优先写在模块目录内；上层 README 只做目录导航，不重复堆业务规则。
- API 契约仍以 `backend/api/openapi/` 为唯一真相源，模块层不手写契约绕过生成链。

## 阅读顺序

1. `system/README.md`：系统主域目录职责
2. `observability/README.md`：审计与遥测目录职责
3. 具体子目录内的 `README.md`：局部边界和接入约束
