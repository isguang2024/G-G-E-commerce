# 注册入口重构 — 上线切换与回滚预案

> AUTH-REG-ENTRY-REFACTOR 上线操作手册。

---

## 1. 上线前提

- [x] 全量代码审计通过（6/6 阶段无阻塞问题）
- [x] `go build ./...` 零错误
- [x] `vue-tsc --noEmit` 零错误
- [ ] 回归场景矩阵核心链路手动验证通过（B1-B2, C1-C4, E1-E5）

---

## 2. 上线步骤

### Step 1: 数据库迁移

```bash
# 执行迁移（幂等，已有表结构不会重复创建）
cd backend && go run cmd/migrate/main.go
```

**验证**：
- `register_entries` 表含 `target_url`, `target_navigation_space_key`, `target_home_path`, `role_codes`, `feature_package_keys` 等列
- `register_entries` 表已无 `policy_code` 列
- `users` 表已无 `register_policy_code` 列
- `register_policies`, `register_policy_roles`, `register_policy_feature_packages` 表已删除
- `default` 入口记录存在且 `is_system_reserved=true`

### Step 2: Permission Seed 刷新

```bash
# 确保 OpenAPI 定义的权限键已写入 DB
cd backend && go run cmd/migrate/main.go
```

**验证**：`permission_keys` 表含注册入口相关的 CRUD 权限键；`RegisterPolicy` 菜单和 `system.register_policy.*` 权限键已不存在。

### Step 3: 部署后端

正常部署后端服务。新 Resolver 逻辑向后兼容：
- 现有 `default` 入口继续生效
- 未配置的 host+path 自动 fallback 到 default
- `RegisterPolicy` 表已完全删除，所有配置已内联到 `RegisterEntry`

### Step 4: 部署前端

正常部署前端构建产物。
- 治理端入口管理页面已改为半屏抽屉编辑（ElDrawer 50% 宽度）
- **注册策略页已删除**（菜单入口、Vue 组件均已移除），策略功能已内联到入口
- 认证页链路已统一为 `gotoAfterAuth`，向后兼容

### Step 5: 验证核心链路

按回归场景矩阵 (`docs/guides/register-entry-regression-matrix.md`) 执行：
1. 默认入口注册 → auto-login → 跳转 (B1)
2. 自定义入口 pending 注册 → 登录页 → 登录 (B2, D1-D2)
3. 跳转优先级 target_url > entry app_key > source app_key (C1-C3)
4. Social callback login + register (E5, F1-F3)

---

## 3. 灰度观察项

| 指标 | 正常范围 | 异常阈值 | 观察方式 |
|------|----------|----------|----------|
| 注册成功率 | >95% | <90% 连续 5 分钟 | 后端日志 `register failed` 频率 |
| 登录跳转成功 | 100% 跳出登录页 | 用户卡在登录页 | 前端 `[AuthFlow]` console 日志 |
| GetRegisterContext 延迟 | <100ms | >500ms | API 监控 |
| default 入口 fallback | 正常触发 | 频繁 404 / resolver error | 后端 `no entry matched` 日志 |

---

## 4. 回滚预案

### 场景 A: 前端跳转异常

**症状**：用户注册/登录后无法跳转，卡在登录页或白屏。

**回滚操作**：
1. 回滚前端到上一个稳定版本
2. 后端无需回滚（API 向后兼容）

**影响范围**：仅前端认证页链路

### 场景 B: 入口解析错误

**症状**：GetRegisterContext 返回错误，注册页无法加载。

**回滚操作**：
1. 检查 `register_entries` 表 default 入口是否存在
2. 如 default 入口丢失，执行恢复：
```sql
INSERT INTO register_entries (
  app_key, entry_code, name, host, path_prefix,
  register_source, status, allow_public_register,
  auto_login, is_system_reserved, sort_order
) VALUES (
  'account-portal', 'default', '默认注册入口', '', '/account/auth/register',
  'self', 'enabled', true,
  true, true, 0
) ON CONFLICT (entry_code) DO NOTHING;
```
3. 如 Resolver 代码有 bug，回滚后端到上一个版本

### 场景 C: 注册事务失败

**症状**：注册提交后报错，用户无法创建。

**回滚操作**：
1. 检查后端日志定位失败步骤（用户创建 / 角色绑定 / 功能包绑定 / 快照写入）
2. 事务保证原子性，无半成品数据
3. 如为角色/功能包 code 无效，修正入口配置即可，无需回滚代码

### 场景 D: 全量回滚

**操作**：
1. 回滚前端到上一版本
2. 回滚后端到上一版本
3. 数据库无需回滚（新增列不影响旧代码，旧代码不读取新列）

---

## 5. 数据兼容性说明

| 项目 | 说明 |
|------|------|
| `register_entries` 新增列 | 有默认值，旧数据兼容 |
| `register_entries.policy_code` | 已删除（00025 迁移） |
| `users.register_policy_code` | 已删除（00025 迁移） |
| `register_policies` / `register_policy_roles` / `register_policy_feature_packages` | 已完全删除（00025 迁移），所有配置已内联到入口 |
| `user.register_policy_snapshot` | 新注册用户冻结快照，已有用户该字段为空（正常） |
| `StringList` (JSONB) 列 | `role_codes`, `feature_package_keys` 默认 `'[]'`，兼容旧数据 |

---

## 6. 策略模板迁移

> **已不再需要。** `register_policies` 表已在 00025 迁移中完全删除，所有配置已内联到 `register_entries`。
