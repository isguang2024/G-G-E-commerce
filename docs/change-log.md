## 2026-03-23 权限系统设计文档重写

### 本次改动
- 按最新确认的“功能包优先”模型，重写了 [permission-package-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-package-design.md) 和 [permission-overall-summary.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-overall-summary.md)。
- 正式写清了两上下文模型、基础包与组合包、团队边界减法、平台用户减法裁剪、公共菜单、初始化包、快照策略和自动注册边界。
- 文档明确区分了“当前正式设计基线”和“后续仍需按设计实现的代码部分”，避免设计与现状继续混淆。

### 下次方向
- 基于新文档输出数据库结构和迁移方案，明确哪些旧表继续保留、哪些需要降级或替换。
- 按设计补齐平台角色、平台用户、团队边界、团队角色的完整功能包链路。
- 开始把旧的零散权限直配入口逐步降级，统一收口到“先开包，再裁剪”的配置模式。

## 2026-03-23 权限系统第一段底座落地

### 本次改动
- 按新设计把第一段底座真正落到了后端：补齐了基础包与组合包字段、组合包关系、平台用户直绑包、团队屏蔽菜单/权限、角色隐藏菜单/关闭权限、用户隐藏菜单等模型与仓储骨架。
- 命名迁移和自动迁移已实际执行成功，数据库已经创建并回填了第一段所需的新结构，默认功能包种子也开始带 `package_type` 和 `is_builtin`。
- 后端编译验证已通过：`go test ./...`；这轮没有启动服务。

### 下次方向
- 继续切第二段读链路，优先改 `authorization`、`menu`、`user service` 这几个最终读取入口，让平台和团队上下文真正按新模型展开菜单与权限。
- 再把后台配置入口切成“先开包、再裁剪”，逐步降级旧的零散权限直配入口和团队成员个人例外链路。
- 最后再收口快照表和缓存策略，把运行时读取从旧的混合态切到统一快照态。

## 2026-03-23 权限系统第一段底层结构落地

### 本次改动
- 在后端模型、仓储、数据库迁移和功能包服务层补上了第一段新设计底座，包括功能包类型、内置包标记、组合包关系、平台用户直绑包、角色隐藏菜单、角色关闭权限、团队屏蔽菜单、团队屏蔽权限和用户隐藏菜单等结构。
- 功能包服务已开始识别 `base/bundle`、`platform,team` 双上下文和内置包删除保护，并禁止组合包直接绑定菜单与权限，避免后续继续产生与设计冲突的数据。
- 已完成 `go test ./...` 验证；本轮未执行数据库迁移、未启动服务、未跑前端校验。

### 下次方向
- 继续把第二段读链路切到新模型：优先改 platform/team 两个上下文的最终菜单与最终权限计算。
- 为基础包与组合包补齐正式接口与页面，再逐步替换旧的“功能包直接绑权限集合”心智。
- 把团队成员个人例外和旧的团队补充权限链路继续降级，收口到“团队开包 + 团队边界减法 + 角色裁剪”。

## 2026-03-23 权限系统第二段团队读链切换

### 本次改动
- 团队上下文的运行时边界开始切到新模型：`teamboundary` 现在会解析组合包、过滤停用包、读取团队屏蔽菜单/权限，并在角色快照里叠加角色隐藏菜单和角色关闭权限的减法逻辑。
- 团队角色来源已与平台角色彻底隔离，`authorization` 和 `userRoleRepository` 在团队上下文下不再把全局角色混进团队权限计算；团队菜单读取也去掉了无角色时回落旧角色菜单表的分支。
- 团队边界写链先同步了一刀：团队权限保存仍然接收“最终允许权限集合”，但后端已改成反推出需要屏蔽的权限并写入 `team_blocked_actions`，同时清理旧 `team_manual_action_permissions`，后端验证已通过 `go test ./...`。

### 下次方向
- 继续切 platform 上下文，把全局角色包、用户直绑包、用户隐藏菜单/权限减法正式接入运行时快照，逐步摆脱 `role_action_permissions` 和 `user_action_permissions` 的主链地位。
- 团队角色与团队菜单/权限配置页还停留在“旧直配表 + 新边界校验”的混合态，下一轮要把写链真正改成“角色绑定包 + 包内减法”，并同步前端文案和来源展示。
- 再之后再收口旧的团队成员个人权限例外和兼容字段，把团队链完全固定到“团队开包 + 团队减法边界 + 角色裁剪”。

## 2026-03-23 权限系统第二段平台读链首批切换

### 本次改动
- 新增了 [platformaccess/service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/platformaccess/service.go)，把平台上下文下的“全局角色包 + 用户直绑包 + 组合包展开 + 角色隐藏菜单/关闭权限 + 用户隐藏菜单 + 公共菜单”收成统一快照入口。
- [authorization.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/authorization/authorization.go) 现在在平台上下文下会优先读取平台功能包快照生成 `actions` 和接口鉴权结果；仅当用户还没有任何平台功能包绑定时，才回退旧的角色动作表链路。
- [user/service.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/service.go)、[user/module.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/user/module.go)、[menu/module.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/modules/system/menu/module.go) 已接入这条平台快照服务，平台菜单树和用户菜单权限查询开始与平台功能包对齐；后端验证已通过 `go test ./...`，本轮未启动服务。

### 下次方向
- 继续把平台权限读取从“有包就走新链、没包就回退旧表”推进到真正以功能包为主，同时补齐平台用户权限减法的正式存储，不再依赖 `user_action_permissions`。
- 团队角色/平台角色的写链还没有完全改成“绑定包 + 减法裁剪”，下一轮应优先处理角色菜单与角色权限的保存方式，否则运行时和配置页会继续并存两套语义。
- 前端权限配置页还没同步展示这轮新增的 `blocked_action_ids / expanded_package_ids` 等来源信息，联调前需要把页面文案和类型一起收口。 

## 2026-03-23 权限系统第二段平台写链前端闭环

### 本次改动
- 在 [system-manage.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/system-manage.ts) 和 [api.d.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/types/api/api.d.ts) 补齐了平台角色、平台用户的功能包读写 API，统一把返回结构规整到 `package_ids + packages`。
- 新增了 [role-package-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/role/modules/role-package-dialog.vue) 和 [user-package-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/user/modules/user-package-dialog.vue)，并接入 [role/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/role/index.vue) 与 [user/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/user/index.vue)，让平台角色和平台用户都能先绑定功能包。
- 同步把平台角色/用户旧权限弹窗文案改成“先绑功能包，再做菜单/权限裁剪”的新语义；已完成 `pnpm exec vue-tsc --noEmit`、`pnpm build`、`go test ./...`，本轮未启动服务。

### 下次方向
- 继续把平台角色菜单、平台角色功能权限、平台用户功能权限从“旧正向配置”收成“功能包主入口 + 隐藏菜单/关闭权限减法”模型，减少页面心智分裂。
- 给平台角色页和平台用户页补上来源展示，明确当前菜单和权限来自哪些功能包，和团队侧来源视图保持一致。
- 再往下一轮再处理平台用户减法的正式配置入口，逐步降级旧 `user_action_permissions` 在平台侧的主链地位。

## 2026-03-23 平台功能包入口收口
### 本次改动
- 在 [system-manage.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/api/system-manage.ts) 补齐了平台角色与平台用户的功能包查询、保存接口，前端已能直接消费 `/roles/:id/packages` 与 `/users/:id/packages`。
- 新增 [role-package-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/role/modules/role-package-dialog.vue) 和 [user-package-dialog.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/user/modules/user-package-dialog.vue)，并接入 [role/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/role/index.vue) 与 [user/index.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/system/user/index.vue)。
- 用户列表页已去掉旧的“直接功能权限”主操作入口，避免继续和“功能包为主”的新模型并列暴露；已完成 `pnpm exec vue-tsc --noEmit`、`pnpm build`，本轮未启动服务。
### 下次方向
- 继续把平台角色的菜单/权限配置改成真正的包内减法模型，逐步退出 `role_menus` 与 `role_action_permissions` 在平台侧的主配置地位。
- 为平台用户补正式的菜单隐藏入口，让“全局角色包并集 + 用户直绑包 - 用户隐藏菜单”完整落到配置页，而不是只停留在运行时快照。
## 2026-03-23 平台角色减法写链接入
### 本次改动
- 平台角色菜单与功能权限配置开始按“功能包主入口 + 减法裁剪”保存：`backend/internal/modules/system/role/service.go` 在角色已绑定平台功能包时，会先展开基础包范围，再反推出隐藏菜单与关闭权限，并分别写入 `role_hidden_menus`、`role_disabled_actions`。
- 为了不打断现有页面和旧数据，这轮仍保留了“无包走旧链”的兼容逻辑；同时把最终勾选结果镜像回 `role_menus`、`role_action_permissions`，让旧页面和回退链路短期内还能工作，但主语义已经切到新模型。
- 同步补齐了 `backend/internal/modules/system/role/module.go` 的依赖注入；验证通过：`backend` 执行 `go test ./...`，`frontend` 执行 `pnpm exec vue-tsc --noEmit` 与 `pnpm build`。本轮没有启动服务。
### 下次方向
- 继续把平台用户侧从旧“个人允许/拒绝功能权限”链路收成“平台包并集 + 用户隐藏菜单/后续权限减法”的正式模型，逐步退出 `user_action_permissions` 在平台侧的主入口地位。
- 给平台角色配置页补候选范围收口与来源展示，让菜单/权限弹窗直接按已绑定功能包过滤，并能看出当前条目来自哪些功能包，减少继续依赖旧正向表的心智成本。

## 2026-03-23 平台角色边界返回与前端候选收口
### 本次改动
- `backend/internal/modules/system/role/handler.go` 开始返回平台角色菜单/权限的边界型结果，在保留 `menu_ids`、`action_ids` 兼容字段的同时，补了 `available_*`、`expanded_package_ids`、`hidden/disabled_*`、`derived_sources`，让前端能感知“包展开范围”和“减法结果”。
- `frontend/src/api/system-manage.ts`、`frontend/src/types/api/api.d.ts`、`frontend/src/views/system/role/modules/role-permission-selector-dialog.vue` 已接到这组边界字段；平台角色权限弹窗现在会按已绑定功能包过滤菜单和权限候选项，而不是继续对全量资源做直配。
- 已验证：`backend` 的 `go test ./...`、`frontend` 的 `pnpm exec vue-tsc --noEmit`、`pnpm build`。本轮没有启动服务。
### 下次方向
- 继续切平台用户侧，把用户页收成“平台角色包并集 + 用户直绑包 + 用户隐藏菜单”的主链，并逐步退出旧的 `user_action_permissions` 主入口。
- 再补平台角色/平台用户的来源展示，让条目能直接看到来自哪些功能包，进一步统一平台侧和团队侧的配置心智。
## 2026-03-23 快照刷新机制规范补充

### 本次改动
- 更新 `docs/permission-package-design.md`，将平台用户减法裁剪收口为“用户隐藏菜单”，并新增“快照刷新机制”章节。
- 更新 `docs/permission-overall-summary.md`，补充快照刷新触发源清单，明确 `platform` 与 `team` 需分上下文刷新。
- 触发源已覆盖：功能包绑定、组合包关系、菜单启停、角色包、用户直绑包、角色减法、用户隐藏菜单、团队开包、团队边界减法、团队角色裁剪、成员角色、功能包状态变更。

### 下次方向
- 按新规范为平台用户、平台角色、团队边界、团队角色四条写链补统一刷新入口。
- 将快照刷新从文档规则落实到代码实现，避免出现规则正确但刷新不全的情况。

## 2026-03-23 统一快照刷新入口接入首批写链

### 本次改动
- 新增 `backend/internal/pkg/permissionrefresh/service.go`，把团队刷新、平台用户刷新、按功能包传播刷新、按菜单传播刷新统一收口为显式编排入口；第一阶段仍采用同步刷新，不引入事件总线。
- 功能包、平台角色、平台用户、菜单更新四条关键写链已开始接入统一刷新入口：功能包变更会传播到相关团队/平台主体，平台角色包与减法裁剪变更会刷新关联用户，平台用户绑包/分配角色/旧个人权限变更会刷新该用户，菜单更新/删除会走按菜单传播刷新。
- 同步补强了规范文档：`docs/permission-package-design.md` 与 `docs/permission-overall-summary.md` 进一步明确“刷新必须在写事务成功后触发”；验证通过 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续把团队角色写链、团队边界菜单减法、成员角色变更接入统一刷新入口，避免团队侧仍有零散 `RefreshCache` 直接调用。
- 平台用户仍保留旧 `user_action_permissions` 过渡入口，下一轮应继续降级它的主链地位，并把用户侧收成“角色包并集 + 用户直绑包 + 用户隐藏菜单”。

## 2026-03-23 团队刷新漏点收口与平台用户兼容入口降级

### 本次改动
- 团队成员加入、移除、身份角色变更现在统一通过刷新入口回刷团队快照；团队成员角色分配、团队角色包、团队角色菜单减法、团队角色权限减法、团队边界权限减法也都不再直接散落调用 `RefreshCache`。
- 功能权限删除链从“只刷新受影响团队”升级成“按关联功能包反查受影响主体后统一刷新”，避免平台角色、平台用户或团队侧遗漏脏快照。
- 文档已补充“当前已接入统一刷新入口的主链”与“平台用户 `user_action_permissions` 仍为兼容入口”的说明；前端旧弹窗与接口注释同步降级成“用户权限例外”语义。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。本轮未启动服务。

### 下次方向
- 继续收平台用户旧 `user_action_permissions` 链，逐步把用户侧正式配置收成“全局角色包并集 + 用户直绑包 + 用户隐藏菜单”的单主链。
- 团队侧仍有部分旧正向表作为兼容输出存在，下一轮应继续把团队角色/团队成员页面的心智收成“先包后减法”，减少旧接口继续被当成主配置入口。
- 等平台与团队两侧的刷新主链都稳定后，再考虑把平台快照从现算/warm-up 收成真正的数据库快照表。 
