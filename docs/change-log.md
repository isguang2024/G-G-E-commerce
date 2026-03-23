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

## 2026-03-23 团队边界减法语义收口

### 本次改动
- 团队列表中的“补充权限”入口正式改名为“团队边界”，团队边界弹窗整体改成减法语义：说明、统计、来源明细、保存提示都围绕“功能包展开 - 团队屏蔽”表达，不再把团队层误导成另一套开通入口。
- 团队成员个人权限弹窗继续降级成“权限例外（兼容）”语义，明确它只用于少量历史兼容例外，当前主模型仍然是“团队开包 + 角色裁剪”。
- 总览文档同步补充：团队页入口与弹窗还需要继续按“团队边界减法”统一心智，这是当前第二优先级之一。

### 下次方向
- 继续把团队角色菜单/权限页面从“正向勾选”推进到真正的减法裁剪心智，减少新模型与旧正向表的混用感。
- 继续收平台用户例外链，让平台侧主链稳定在“角色包并集 + 用户直绑包 + 用户隐藏菜单”，兼容例外入口只保留少量历史用途。

## 2026-03-24 团队角色裁剪交互收口

### 本次改动
- 团队角色菜单裁剪弹窗新增“已保留/已屏蔽”实时统计，并提供“全部保留/全部屏蔽”批量操作，保存时改为直接使用当前保留集合，强化减法边界心智。
- 团队角色权限裁剪弹窗同步新增“已保留/已屏蔽”统计与一键全保留/全屏蔽操作，文案明确未保留项将视为角色屏蔽。
- 同步清理团队角色与成员例外弹窗残留的“可配/功能权限”旧文案，统一到“裁剪/边界/兼容例外”术语体系。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`；本轮未启动服务。

### 下次方向
- 继续把团队角色裁剪结果按来源分组展示（功能包展开保留 vs 角色显式屏蔽），减少排查时的认知成本。
- 推进平台用户兼容例外入口的进一步降级，保持平台侧主链稳定在“角色包并集 + 用户直绑包 + 用户隐藏菜单”。

## 2026-03-24 团队角色屏蔽结果可见与平台用户包约束显式化

### 本次改动
- 团队角色菜单裁剪弹窗新增“当前角色已屏蔽菜单”，团队角色权限裁剪弹窗新增“当前角色已屏蔽能力”，减法结果不再只靠统计数字推断。
- 平台“用户权限例外”页新增功能包约束模式/兼容回退模式显式提示，并把候选能力来源功能包标签直接展示到叶子节点和角色继承页，进一步压实“先绑包，再做兼容例外”的心智。
- `docs/permission-overall-summary.md` 已同步补充：平台用户例外页需要明确区分功能包约束模式与兼容回退模式。

### 下次方向
- 继续把平台用户例外页做成真正的兼容副入口，优先补“无功能包时回退风险提示”和“来源功能包跳转”。
- 继续做团队侧来源分组，把功能包展开保留与角色显式屏蔽同时可追溯，减少排查时在多个弹窗之间来回切换。

## 2026-03-24 平台功能包菜单绑定开放与上下文三态收口

### 本次改动
- 后端已放开平台功能包绑定菜单，功能包菜单集合不再只允许团队上下文配置；组合包仍然禁止直接绑定菜单。
- 功能包管理页已把上下文展示收口成真实三态：平台、团队、平台/团队，并新增双上下文统计；平台与双上下文功能包现在都可进入“绑定菜单”。
- 功能包菜单弹窗文案同步改成“当前上下文入口”，不再把平台功能包误描述为团队专用。已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续处理文档与实现的另一条高优先级错位：团队边界目前只有屏蔽权限写链，还缺团队屏蔽菜单的正式读写与刷新链。
- 平台用户主链仍需进一步收口到“角色包并集 + 用户直绑包 + 用户隐藏菜单”，后续应继续弱化 `user_action_permissions` 的可写入口。

## 2026-03-24 团队成员兼容例外降级为只读审计

### 本次改动
- 团队成员“权限例外（兼容）”页已降级为只读审计视图：批量覆盖按钮、分组覆盖按钮、单项允许/拒绝编辑和保存入口都已退出页面主交互。
- 当前页面仍保留来源功能包、团队边界、角色基线、既有例外结果的可视化能力，便于审计历史数据，但不再鼓励继续走成员级例外写链。
- `docs/permission-overall-summary.md` 与 `docs/permission-package-design.md` 已同步写明：团队成员个人例外处于迁移期只读审计状态。已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续处理团队边界“屏蔽菜单”缺口，把团队边界从“只有权限减法”补成“菜单 + 权限”完整减法模型。
- 后续再决定是否彻底下线团队成员个人例外写接口，避免后端仍然保留可写兼容链。

## 2026-03-24 团队角色来源跳转增强与规范文档对齐

### 本次改动
- 团队角色菜单裁剪页与权限裁剪页的“展开能力”和“已屏蔽能力”标签都支持直接点击跳转来源功能包，排查链路从“看结果”缩短到“定位配置”。
- 平台“用户权限例外（兼容）”页新增未绑定功能包时的“前往绑定功能包”引导按钮，并在角色继承区域补充“功能包约束/兼容回退”状态标签，进一步明确该页是兼容入口而非主入口。
- `docs/permission-package-design.md` 与 `docs/permission-overall-summary.md` 已同步更新当前实现状态，清理重复错位条目，并补充团队侧与平台侧兼容入口的现状说明。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`；本轮未启动服务。

### 下次方向
- 继续把平台用户兼容入口降级为纯例外配置，并补齐“未绑定功能包时保存风险”防误操作策略。
- 推进团队角色/成员页对旧兼容字段的依赖清理，逐步转向“包展开 + 减法裁剪 + 快照刷新”单一心智。

## 2026-03-24 团队菜单边界主链落地

### 本次改动
- 后端已补齐团队菜单边界接口：新增团队菜单快照读取、菜单来源返回和菜单边界保存，团队菜单边界现在正式以“功能包展开菜单 - 团队屏蔽菜单”工作。
- 前端团队页新增“菜单边界”入口，并新增团队菜单边界弹窗；当前支持按功能包来源查看菜单、按来源功能包跳转，以及“全部保留/全部屏蔽”批量操作。
- 文档已同步收口：`permission-overall-summary.md` 与 `permission-package-design.md` 现在都明确记录团队边界已形成“菜单 + 权限”双减法模型。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。本轮未启动服务。

### 下次方向
- 继续收平台用户主链，把平台侧稳定到“全局角色包并集 + 用户直绑包 + 用户隐藏菜单”的正式模型，进一步弱化 `user_action_permissions` 的可写入口。
- 继续把团队成员与团队角色页面里的旧兼容字段降级为审计输出，前端主配置入口统一收成“功能包开通 + 菜单/权限减法裁剪 + 快照刷新”。

## 2026-03-24 团队菜单边界主入口落地

### 本次改动
- 新增团队菜单边界前端主入口：团队列表操作菜单补齐“菜单边界”，并挂载新弹窗 `frontend/src/views/team/team/modules/team-menu-permission-dialog.vue`。
- 新增团队菜单边界前端 API：`fetchGetTeamMenus`、`fetchGetTeamMenuOrigins`、`fetchSetTeamMenus`，并补齐团队边界返回字段（含 `expandedPackageIds`、`blockedActionIds`）。
- 团队菜单边界弹窗支持来源分组、功能包筛选、功能包跳转、全保留/全屏蔽批量操作，保存语义固定为“功能包展开菜单 - 团队屏蔽菜单”。
- 同步更新 `docs/permission-overall-summary.md` 与 `docs/permission-package-design.md`，将团队边界正式收口为“菜单减法 + 权限减法”双边界模型。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续收 platform 主链，将平台用户侧正式固定到“全局角色包并集 + 用户直绑包 + 用户隐藏菜单”，进一步弱化 `user_action_permissions` 兼容入口。
- 把团队边界与团队角色页面的来源视图继续统一成同一套交互（来源分组 + 目标包跳转 + 屏蔽明细），减少排障切换成本。
- 在后续一节推进数据库快照表落地，逐步减少运行时现算路径，确保刷新触发后读链稳定命中快照。

## 2026-03-24 平台用户菜单裁剪主链落地

### 本次改动
- 后端已补齐平台用户菜单裁剪接口，用户菜单读写现在基于平台快照的“候选菜单 + 最终菜单”双层结构工作，支持从功能包展开菜单中反推出用户隐藏菜单，并在保存后统一刷新平台用户快照。
- 平台快照已补充候选菜单层，避免“已隐藏菜单无法恢复”的问题；用户管理页新增“菜单裁剪”入口，并新增 `user-menu-selector-dialog.vue`，当前支持功能包来源筛选、目标功能包跳转、全部保留/全部隐藏批量操作。
- 平台用户菜单裁剪已成为正式入口，平台用户“权限例外（兼容）”继续保留但降级为兼容副入口；`permission-overall-summary.md` 与 `permission-package-design.md` 已同步写明这一点。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把平台用户“权限例外（兼容）”写链再降一级，明确只承担少量历史例外能力，不再与主链并列。
- 推进团队角色和平台角色两侧来源视图统一，把“包展开结果 / 显式屏蔽结果 / 最终生效结果”做成同一套查看交互。
- 在平台和团队两侧主入口都稳定后，开始把现算路径进一步收口成数据库快照表，减少运行时临时拼装。

## 2026-03-24 平台用户兼容例外降级为只读审计

### 本次改动
- 平台用户“权限例外（兼容）”页已从可写入口降级为只读审计视图：当前保留历史 allow/deny 例外、角色继承、来源功能包与结果预览，但退出保存与交互修改，不再和平台主链并列。
- 用户管理页操作文案同步收口为“权限例外审计”，强调平台侧主配置应优先使用“功能包 + 菜单裁剪”；平台角色权限页也新增了功能包摘要、菜单已屏蔽数、能力已关闭数，开始统一平台侧来源与减法心智。
- `permission-overall-summary.md` 与 `permission-package-design.md` 已同步更新当前状态。已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续统一平台角色、团队角色、团队边界三处来源视图，把“包展开结果 / 显式屏蔽结果 / 最终生效结果”做成一致交互。
- 继续评估并逐步下线平台用户 `user_action_permissions` 的可写后端链路，只保留兼容读能力。
- 在平台与团队主入口都稳定后，开始推进数据库快照表落地，减少运行时现算路径。

## 2026-03-24 平台角色来源视图与团队侧交互统一

### 本次改动
- 平台角色菜单权限页补齐了“功能包展开菜单 / 当前角色已屏蔽菜单”来源卡片，支持按来源功能包筛选，并能直接跳转到目标功能包的菜单配置。
- 平台角色功能权限页补齐了“功能包展开能力 / 当前角色已关闭能力”来源卡片，支持按来源功能包筛选，并能直接跳转到目标功能包的能力配置。
- 平台角色页顶部摘要已与团队角色页统一，当前平台与团队两侧都能按“展开结果 / 显式屏蔽结果 / 最终保留结果”同一心智查看来源与减法结果。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续统一平台角色、团队角色、团队边界三处页面的来源展示与筛选交互，减少排查时的页面切换成本。
- 在平台与团队主入口都稳定后，继续把现算路径收口到正式数据库快照表，减少运行时临时拼装。

## 2026-03-24 默认功能包菜单绑定与默认角色开包补齐

### 本次改动
- 修复了“页面进不去”的初始化根因：默认内置功能包种子此前只绑定了权限，没有绑定菜单，导致新模型下菜单候选范围为空时即使已有权限键也无法进入页面。
- 为默认内置功能包补齐了菜单绑定：平台系统管理包、平台菜单管理包、平台接口管理包、团队成员管理包现在都会同步写入 `feature_package_menus`。
- 新增默认角色功能包初始化：`admin` 自动绑定平台系统/菜单/API 管理包，`team_admin` 自动绑定团队成员管理包，避免默认角色继续停留在旧 `role_menus` 单链路。
- 已验证：`backend/go test ./...`。这轮没有启动服务，也没有执行迁移；需手动运行迁移后数据库才会真正补齐默认包菜单和角色开包关系。

### 下次方向
- 继续处理当前菜单主链的剩余兼容点，逐步降低旧 `role_menus` 在平台侧的上限语义，固定为“功能包展开 + 角色减法”主模型。
- 在默认入口补齐后，继续推进数据库快照表，把菜单与权限读取从现算进一步收口到持久化快照链。
## 2026-03-24 来源页统一与数据库快照表落库

### 本次改动
- 新增统一来源组件 `PermissionSourcePanels`，把平台角色、团队角色、团队边界三处来源页的筛选布局、屏蔽态展示和来源包跳转统一到同一套交互，减少后续维护分叉。
- 新增正式数据库快照表模型并完成迁移落地：
  - `platform_user_access_snapshots`
  - `team_access_snapshots`
- 平台访问服务和团队边界服务已开始支持“读取优先命中快照、刷新时持久化写入快照”，不再只停留在现算/warm-up。
- 已实际执行迁移：`backend/go run ./cmd/migrate`
- 已验证：
  - `backend/go test ./...`
  - `frontend/pnpm exec vue-tsc --noEmit`
  - `frontend/pnpm build`

### 下次方向
- 继续把平台旧直配链再降一级，优先缩小 `user_action_permissions` 和 `role_menus` 在平台侧的主链地位。
- 继续统一团队边界、团队角色、平台角色三处页面的细节反馈，例如相同的摘要顺序、相同的来源态文案、相同的跳转后反馈。

## 2026-03-24 菜单来源页统一与平台用户旧写链停用

### 本次改动
- 平台用户菜单裁剪页与团队菜单边界页已统一接入共享来源面板组件，来源功能包筛选、来源包跳转、屏蔽态标签展示现在使用同一套交互，继续减少前端维护分叉。
- 平台用户“权限例外（兼容）”后端写入口已正式停用，`PUT /api/v1/users/:id/actions` 现在只返回禁止写入提示，避免旧 `user_action_permissions` 链继续写穿平台主链。
- 设计总览与设计基线文档已同步更新，明确平台用户主链固定为“全局角色包并集 + 用户直绑包 + 用户隐藏菜单”，`user_action_permissions` 仅保留读取审计用途。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续统一平台用户菜单裁剪、平台角色裁剪、团队边界、团队角色四处来源页的摘要和状态反馈，让“展开结果 / 显式屏蔽 / 最终结果”完全同构。
- 开始继续推进数据库快照表的真实读链，把平台和团队主链从现算/warm-up 更进一步收口到持久化快照。

## 2026-03-24 运行时菜单读链切到平台与团队快照

### 本次改动
- 菜单树接口与用户菜单权限接口已开始优先读取数据库快照：平台侧读取 `platform_user_access_snapshots`，团队侧读取 `team_access_snapshots`，进一步降低旧 `permissionService` 在运行时的主链地位。
- `permissionService.GetUserMenuIDs` 现在退回兼容回退角色，只有在快照服务不可用时才继续兜底；数据库快照已经开始真正接管菜单入口的正式读链。
- 设计总览和设计基线文档已同步更新，明确“平台/团队快照”正在成为菜单主读链，而不是单纯预热缓存。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把更多运行时读链切到快照，优先看平台/团队最终权限集合与菜单权限接口的旧回退路径，进一步缩小旧 `permissionService` 的职责。
- 继续统一平台用户、平台角色、团队边界、团队角色四处页面的摘要与状态反馈，让“展开结果 / 显式屏蔽 / 最终结果”在交互上完全同构。

## 2026-03-24 来源页摘要组件统一

### 本次改动
- 新增共享摘要组件 `PermissionSummaryTags`，并接入平台用户菜单裁剪、团队菜单边界、团队边界、团队角色菜单裁剪、团队角色权限裁剪、平台角色权限页顶部统计。
- 这些页面现在开始共用同一套摘要顺序和标签语义，避免后续继续出现“同类页面统计字段顺序、颜色、文案各不相同”的维护分叉。
- 设计总览与设计基线文档已同步更新，明确前端来源页当前已经形成“共享来源面板 + 共享摘要组件”的双层统一交互基础。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把平台与团队剩余运行时读链切到数据库快照，优先评估最终权限集合和旧权限服务回退路径的收口空间。
- 继续统一前端来源页剩余的提示文案、空态反馈和跳转后反馈，让四类页面在体验层完全同构。

## 2026-03-24 团队成员菜单读链去旧服务与菜单模块兼容依赖收口

### 本次改动
- 用户菜单权限接口的团队上下文读取已从旧 `permissionService` 迁出，改为直接按“当前成员角色子集 + 团队角色快照”计算最终菜单集合，避免继续把旧菜单权限服务当团队成员菜单主链。
- 菜单树接口补齐了平台超级管理员特判：平台超级管理员现在直接读取启用菜单集合，不再依赖功能包配置才能看到系统入口。
- 菜单模块已停止创建旧菜单权限服务作为菜单树依赖，进一步缩小旧 `permissionService` 在运行时的主链职责。
- 共享来源面板补齐了统一空态与筛选后反馈；当前无来源明细、筛选后无展开项/无屏蔽项时，平台与团队来源页会使用同一套提示。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把最终权限集合等剩余读链逐步切到数据库快照，进一步缩小旧 `permissionService` 和旧正向表的运行时职责。
- 继续统一平台用户、平台角色、团队边界、团队角色四类来源页的空态、提示文案和跳转反馈，保持“展开结果 / 显式屏蔽 / 最终结果”完全同构。

## 2026-03-24 权限键返回统一与来源页术语对齐

### 本次改动
- 超级管理员动作快照现在统一按 `permission_key` 返回，不再优先回落成 `resource:action` 旧串，前端权限判断与展示语义进一步统一。
- 团队角色能力裁剪页已把“当前角色已屏蔽能力”收口为“当前角色已关闭能力”，与平台角色页保持一致。
- 团队菜单边界、团队权限边界的顶部摘要已统一收口为“边界已屏蔽”，避免同类边界页出现多个近义名词。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把最终权限集合的运行时判断往快照/边界服务推进，进一步降低旧正向表在鉴权链中的参与度。
- 继续统一来源页的空态、筛选后反馈和目标包跳转提示，减少平台与团队页面的维护分叉。
## 2026-03-24 平台优先上下文修复管理员空菜单

### 本次改动
- 前端正式补入 `platform/team` 上下文模式：租户 store 新增 `currentContextMode` 与平台能力判断，HTTP 层仅在 `team` 上下文下发送 `X-Tenant-ID`，避免管理员默认带团队头把平台菜单请求打到团队分支。
- 登录初始化、路由初始化和上下文切换器已统一改成“有平台能力则默认进入平台空间，否则进入团队空间”；上下文切换器新增平台入口，同时保留单团队无平台能力时不显示切换器的规则。
- 设计文档已同步写明“平台优先默认上下文”和“只有团队上下文才发送租户头”的正式规则。已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 通过；本轮未启动服务。

### 下次方向
- 继续验证管理员菜单接口在平台上下文下的返回，必要时再排查平台默认角色、默认功能包和快照刷新链是否完整命中。
- 继续统一平台用户、平台角色、团队边界、团队角色四类页面的上下文提示与来源反馈，并推进数据库快照表接管更多运行时读链。
## 2026-03-24 平台角色快照读链接入

### 本次改动
- 新增平台角色快照服务 `backend/internal/pkg/platformroleaccess/service.go`，正式使用 `platform_role_access_snapshots` 持久化平台角色的已绑定包、展开包、候选菜单、最终菜单、候选能力、最终能力与来源映射。
- `permissionrefresh` 已接入平台角色快照刷新：平台角色变更时会先刷新角色快照，再刷新受影响平台用户快照。
- 平台角色菜单/权限边界读取已改为“平台角色快照优先、现算兜底”，平台角色与团队角色的读链模型进一步对齐。
- 已验证：`backend/go test ./...` 通过；本轮未启动服务。

### 下次方向
- 继续把平台与团队最终权限集合的运行时读取往快照表推进，进一步缩小旧现算与旧正向表在鉴权链中的职责。
- 继续整理并压缩 `permission-package-design.md` 中已累计的重复段落，保持设计文档只表达当前正式状态与实现差距。
## 2026-03-24 平台用户权限审计主链收口与设计文档重写

### 本次改动
- 平台用户权限审计接口改成“快照主链 + 兼容审计”双层返回：后端 `GET /users/:id/actions` 现在同时返回快照展开出来的候选/生效权限和历史兼容例外，前端不再依赖全量功能权限列表去二次过滤，用户权限审计页直接围绕快照结果展示。
- 重写了 `docs/permission-package-design.md` 和 `docs/permission-overall-summary.md`，移除累计追加产生的重复段落和历史残留，重新整理为当前正式基线：两个上下文、功能包优先、基础包/组合包、团队减法边界、平台用户菜单裁剪、快照刷新机制。
- 同步补全了最新设计约束：平台用户权限例外仅作审计、团队成员个人例外退出主链、快照刷新必须覆盖功能包关系、组合包关系、菜单启停、角色包变更、用户直绑包变更、角色减法、用户减法、团队开包、团队边界减法、团队角色裁剪、成员角色变更。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把运行时权限读链往快照表收口，优先处理鉴权中仍会回退旧正向表和旧现算的分支，进一步固定到“功能包展开 + 减法裁剪 + 快照读取”主链。
- 继续统一平台用户、平台角色、团队边界、团队角色四类页面的来源交互和术语，减少前端页面继续分叉维护的成本。

## 2026-03-24 旧团队菜单清理、权限键点式收口与功能包双上下文入口补齐

### 本次改动
- 扩大了迁移阶段的旧菜单清理范围：在保留当前 `TeamRoot` 结构的前提下，自动删除历史残留的旧 `/team` 根菜单树，修复菜单管理中“团队管理”重复的问题。
- 权限键旧回退链已收口：`permissionkey.FromLegacy` 对未知 legacy 不再回退成 `resource:action`，而是统一输出点式 key；同时新增命名迁移清理现存冒号格式权限键。
- 功能包编辑弹窗补齐了 `platform,team` 选项，正式开放双上下文功能包的创建与编辑入口。
- 已执行验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`、`backend/go run ./cmd/migrate`。

### 下次方向
- 继续检查并收口菜单与权限页面里仍残留的旧命名和旧兼容展示，进一步减少新旧模型混用感。
- 继续把运行时读链往数据库快照表推进，缩小旧正向表与旧现算分支在鉴权主链中的职责。

## 2026-03-24 团队角色支持双上下文功能包

### 本次改动
- 修复了团队角色保存功能包时错误拒绝 `platform,team` 双上下文包的问题；团队角色现在和团队本身一样，允许绑定 `team` 与 `platform,team` 功能包，只有纯 `platform` 功能包仍然禁止进入团队链路。
- 团队角色功能包弹窗与团队功能包弹窗的上下文展示已统一，`platform,team` 现在明确显示为“平台/团队”，避免看起来像纯团队包。
- 设计文档已同步补充正式规则：团队和团队角色都可以使用 `team` 与 `platform,team` 功能包，纯 `platform` 包不能进入团队授予链。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续清理旧权限键与旧兼容回退链，进一步统一成点式 `permission_key` 与功能包主链。
- 继续把运行时权限读链往快照表推进，缩小旧正向表和现算分支在平台/团队鉴权中的职责。

## 2026-03-24 基础团队角色功能包全量可见修复

### 本次改动
- 修复了基础团队角色功能包弹窗只显示“当前已继承包”、看起来像缺包的问题；现在会只读展示当前团队全部可用功能包，并保留继承结果勾选态。
- 同步补充了设计基线：基础团队角色虽然不可直接编辑功能包，但查看页必须展示团队全部可用包，不能只展示已继承子集。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续清理旧权限键与旧兼容回退链，进一步统一成点式 `permission_key` 与功能包主链。
- 继续把运行时权限读链往快照表推进，缩小旧正向表和现算分支在平台/团队鉴权中的职责。

## 2026-03-24 点式权限键运行时收口

### 本次改动
- 收口了旧 `resource:action` 回退链：后端 `permissionkey.Normalize`、鉴权快照输出与前端权限判断现在都会优先归一化为点式 `permission_key`，不再把旧冒号串继续传播到运行时与展示层。
- 前端系统管理与团队接口的权限键回退已统一改为点式格式；历史接口即使还返回冒号格式，也会在进入前端前先被归一化。
- 设计文档已同步补充正式规则：权限键统一使用点式格式，冒号格式仅作为兼容输入。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把运行时权限读链往快照表推进，缩小旧正向表和现算分支在平台/团队鉴权中的职责。
- 再检查平台角色、团队角色和团队边界页是否还存在旧兼容字段参与判断的分支，继续压缩到功能包展开与快照主链。

## 2026-03-24 权限注册上下文回填修复

### 本次改动
- 修复了权限自动注册链未显式写入 `context_type` 的问题：`permissionkey` 映射已补齐平台系统权限的上下文，API 注册表同步在创建和更新 `permission_actions` 时都会显式写入 `context_type`，不再依赖模型默认值 `team`。
- 新增并执行了命名迁移 `20260324_permission_context_backfill`，把数据库里已错误写成团队上下文的 `system.*`、`tenant.*`、`platform.*` 权限统一回填为平台上下文，同时保留 `team.*` 为团队上下文。
- 已完成样本核查：`system.user.manage`、`system.user.assign_role`、`system.menu.manage`、`system.menu.backup`、`system.page_catalog.view`、`tenant.manage`、`tenant.boundary.manage` 现已为 `platform`，`team.member.manage` 仍为 `team`。已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`、`backend/go run ./cmd/migrate`。

### 下次方向
- 保留既定方向，继续把运行时权限读链往快照表推进，缩小旧正向表和现算分支在平台/团队鉴权中的职责。
- 继续检查平台角色、团队角色、团队边界页里是否还存在旧兼容字段参与判断的分支，继续压缩到“功能包展开 + 减法裁剪 + 快照读取”的单主链。

## 2026-03-24 平台鉴权切换到快照强主链

### 本次改动
- `authorization` 在平台上下文下已改为快照强主链：平台用户权限判定直接读取平台用户快照中的最终 `ActionIDs`，不再按 `HasPackageConfig` 回退旧正向权限表。
- 平台用户权限快照接口输出也同步改为无条件使用平台快照结果，减少平台侧新旧规则混用。
- 已同步更新设计文档，明确平台上下文鉴权应直接读取快照；已验证：`backend/go test ./...`。

### 下次方向
- 继续把团队侧最终权限判定中仍依赖旧正向表的回退分支压缩到最小，进一步固定到“功能包展开 + 减法裁剪 + 快照读取”主链。
- 继续统一平台角色、团队角色、团队边界来源页的筛选和状态反馈，减少页面维护分叉。

## 2026-03-24 团队鉴权切换到角色快照主链

### 本次改动
- `authorization` 在团队上下文下已改为角色快照主链：团队权限判定直接读取团队角色快照动作集合，不再在“团队边界未配置”时回退旧角色正向权限表。
- 团队权限快照接口输出也同步改为统一读取团队角色快照结果，不再混用团队成员个人例外和旧正向权限链。
- 已同步更新设计文档，明确团队上下文鉴权与权限快照输出应直接走角色快照结果；已验证：`backend/go test ./...`。

### 下次方向
- 继续把团队边界与团队角色页面里仍使用旧兼容字段的判断分支压缩到最小，统一到“功能包展开 + 减法裁剪 + 快照读取”语义。
- 继续推进平台角色快照从“快照优先、现算兜底”向更稳定的持久化快照单读链收口。

## 2026-03-24 平台角色边界退出旧正向直配回退

### 本次改动
- 平台角色菜单与权限边界读取已改为统一走角色快照结果：`GetRoleMenuBoundary/GetRoleActionBoundary` 不再在无包场景回退 `role_menus` / `role_action_permissions`。
- 平台角色菜单与权限保存也统一按快照可用边界做减法裁剪，不再把无包场景写回旧正向表，进一步压缩了平台侧新旧模型混用。
- 已同步更新两份设计文档，并验证：`backend/go test ./...` 通过。

### 下次方向
- 继续收团队边界与团队角色页面里残留的旧兼容字段判断，统一来源展示与状态语义。
- 继续推进平台角色与平台用户快照从“计算兜底”向更稳定的持久化快照单读链靠拢。

## 2026-03-24 连续推进小节1 团队鉴权主链收口

### 本次改动
- 团队上下文的 `authorization` 权限判定已统一改为团队角色快照主链，不再在团队边界未配置场景回退旧角色正向权限表。
- 团队权限快照输出也改为统一读取团队角色快照结果，不再混用团队成员个人例外与旧正向链。
- 已验证：`backend/go test ./...`。

### 下次方向
- 继续收平台角色前端页中的 `has_*` 兼容判断，统一按快照可用集合过滤。
- 继续把快照缺失时的临时现算路径改成“刷新并持久化”。

## 2026-03-24 连续推进小节2 平台角色页边界过滤统一

### 本次改动
- 平台角色权限页已移除 `has_menu_boundary` / `has_package_boundary` 兜底判断，菜单与能力列表统一按快照返回的可用集合过滤。
- 角色动作预选逻辑不再回退全量动作集合，统一按可用集合筛选，减少页面与后端主链语义偏差。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续收团队边界页面中的旧 `manual` 命名，统一为“边界屏蔽”语义。
- 推进快照服务缺失自动刷新持久化。

## 2026-03-24 连续推进小节3 团队边界页面命名收口

### 本次改动
- 团队边界页中旧 `manual*` 命名已改为 `blocked*`，统一成“团队边界屏蔽”语义，避免把减法边界误解为手工补充能力。
- 来源面板传参同步改为 `blockedActionItems`，界面语义与后端 `blocked` 模型一致。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续把平台/团队快照服务改成“缺失即刷新并落库”。
- 收口平台角色服务里剩余的旧正向回退分支。

## 2026-03-24 连续推进小节4 快照缺失自动刷新持久化

### 本次改动
- 平台用户快照 `platformaccess.GetSnapshot` 改为快照缺失直接 `RefreshSnapshot` 并持久化。
- 平台角色快照 `platformroleaccess.GetSnapshot` 同步改为缺失即刷新持久化。
- 团队边界快照 `teamboundary.GetSnapshot/GetMenuSnapshot/GetRoleSnapshot` 改为缺失即刷新并写回快照表，减少运行时临时现算。
- 已验证：`backend/go test ./...`。

### 下次方向
- 收口平台角色服务里的旧 `role_menus` / `role_action_permissions` 回退。
- 继续统一来源页状态反馈。

## 2026-03-24 连续推进小节5 平台角色服务退出旧正向回退

### 本次改动
- 平台角色服务 `GetRoleMenuBoundary/GetRoleActionBoundary` 已改为统一读取角色快照边界，不再在无包场景回退 `role_menus` / `role_action_permissions`。
- 平台角色 `SetRoleMenus/SetRoleActions` 已统一按快照可用边界做减法裁剪，不再把无包场景写回旧正向表。
- 已同步更新设计文档与总览文档；已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续统一团队边界、团队角色、平台角色三类来源页的筛选/空态/跳转反馈，降低维护分叉。
- 继续压缩旧兼容字段在接口响应中的主链参与度，保持“功能包展开 + 减法裁剪 + 快照读取”单主链。

## 2026-03-24 来源页统一与团队边界 blocked 主字段
### 本次改动
- 重写了共享来源面板组件，统一功能包筛选、目标功能包跳转、筛选后空态和全局空态反馈，减少平台角色、团队角色、团队边界、平台用户页面的交互分叉。
- 团队边界动作来源正式以 lockedActionIds 作为前端主字段，manualActionIds 仅保留兼容读取；团队边界页已切换到 blocked 主链。
- 平台角色、团队角色、团队边界、平台用户来源页新增统一空态文案与筛选后反馈，继续收口到“功能包展开 + 显式裁剪 + 最终结果”的同构交互。
- 已验证：ackend/go test ./...、rontend/pnpm exec vue-tsc --noEmit、rontend/pnpm build。
### 下次方向
- 继续把团队边界、团队角色、平台角色页里残留的旧兼容字段判断压到最小，进一步固定到快照主链。
- 继续推进运行时权限读链往持久化快照收口，缩小旧正向表和现算回退在鉴权链中的职责。

## 2026-03-24 权限上下文与点式权限键收口

### 本次改动
- 修复了权限自动注册和权限列表展示中的上下文漂移问题：后端映射、迁移回填、前端归一化三条链统一补齐 `context_type` 推导，平台系统权限不再错误显示为团队上下文。
- 修复了功能权限页仍出现冒号/旧兼容格式的问题：权限列表接口和前端统一优先使用点式 `permission_key`，历史冒号格式只保留兼容输入。
- 已实际执行数据库迁移 `go run ./cmd/migrate`，并验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 继续沿当前主线把运行时权限读链往快照表推进，进一步缩小旧正向表和现算分支在平台/团队鉴权中的职责。
- 继续检查团队边界、团队角色、平台角色页面里残留的旧兼容字段判断，进一步固定到“功能包展开 + 减法裁剪 + 快照读取”的单主链。

## 2026-03-24 团队角色减法裁剪与运行时快照读链收口

### 本次改动
- 团队角色快照补齐了 `available_* / hidden_* / disabled_*` 边界字段，并新增迁移 `20260324_team_role_access_snapshot_boundary_fields`，把团队角色菜单/能力的候选范围与最终结果一起固化到快照表。
- 团队角色菜单与能力接口改为按“功能包展开候选 - 角色裁剪”读写 `role_hidden_menus` / `role_disabled_actions`，保存后同步清空旧 `role_menus` / `role_action_permissions` 正向表，避免页面显示与运行时鉴权继续错位。
- 平台/团队运行时菜单与能力读取进一步收口到快照：`authorization`、`menu`、`user` 相关链路不再把旧正向表作为主判断；团队边界前端页也移除了 camelCase/旧兼容字段折叠，统一按快照响应字段消费。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`。未执行 `frontend/pnpm build`。

### 下次方向
- 继续检查平台角色和其余权限来源页是否还残留旧兼容响应折叠，尤其是共享 API 包装层里未被主页面使用但仍在保留的旧别名。
- 继续推进平台用户、团队成员等运行时消费点对快照 source map / 边界字段的统一，逐步压缩旧正向表只保留迁移清理职责。

## 2026-03-24 快照主链收尾小节6 鉴权死代码与包装层双轨清理

### 本次改动
- 删除了 `authorization.go` 中已失效的旧“用户覆盖 + 角色正向表”辅助函数，避免核心鉴权文件继续传递旧主链信号。
- 收口 `frontend/src/api/team.ts` 的 my-team action 包装函数为 snake_case 主字段；同步统一平台角色 API 包装层的 `actions/packages` 读取形态，减少页面侧再次补兼容分支的空间。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`。未执行 `frontend/pnpm build`。

### 下次方向
- 继续检查 `system-manage.ts` 与团队成员相关包装函数里是否还有非主页面使用但会污染语义边界的双轨字段。
- 继续把团队角色里仅用于清空旧表的依赖显式标成 legacy cleanup，降低后续误读成本。

## 2026-03-24 快照主链收尾小节7 团队角色 legacy cleanup 显式化

### 本次改动
- 团队角色 handler/module 中，旧 `role_menus` / `role_action_permissions` 仓库依赖已显式重命名为 `legacy*CleanupRepo`，明确它们只用于保存后清空旧正向表，不参与运行时主链判断。
- 已验证：`backend/go test ./...`。未执行前端校验与构建。

### 下次方向
- 继续清理团队成员与平台用户包装层里不必要的 camelCase 投影，尽量把快照字段统一在 API 层就固定下来。
- 视需要把 tenant handler 里 legacy cleanup 调用再补一层短注释，进一步降低后续维护误判。

## 2026-03-24 快照主链收尾小节8 平台用户与团队成员包装层收口

### 本次改动
- `frontend/src/api/system-manage.ts` 的平台用户权限/菜单包装层改成“输入兼容、输出固定”：继续兼容后端历史 camelCase/旧字段读取，但对页面只输出 `available_* / derived_* / has_package_config` 这套 snake_case 主字段。
- `frontend/src/views/system/user/modules/user-permission-selector-dialog.vue` 与 `frontend/src/views/system/user/modules/user-menu-selector-dialog.vue` 已同步去掉 `snapshot/effective/compat` 多级 fallback，改为直接消费包装层固定后的快照候选字段，保留兼容审计语义但不再让旧别名参与页面主判断。
- `frontend/src/api/team.ts` 与 `frontend/src/views/team/team-members/modules/member-action-dialog.vue` 已统一团队成员动作例外响应为 snake_case，避免团队成员页继续传播 camelCase 包装别名。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`。未执行 `frontend/pnpm build`。

### 下次方向
- 继续检查平台用户兼容审计页是否还能把局部状态命名和说明文案再贴近“快照候选 + 例外覆盖”语义，进一步降低旧主链暗示。
- 继续把平台/团队快照缺失时的刷新与落库策略统一到单套服务入口，减少运行时现算分支。
## 2026-03-24 快照主链一次性收口 ABC

### 本次改动
- 一次性执行了前端包装层、团队边界/团队角色消费页、平台用户兼容审计页的主链收口：页面只再消费固定 snake_case 快照字段，旧 `snapshot/effective/compat` 回退和多余 camelCase 投影不再参与主判断。
- 后端把平台用户与团队边界来源接口继续压到“候选集 + 减法裁剪 + 快照结果”，移除了不再被前端主链消费的 `manual/effective/package_ids` 等冗余输出，并把团队 manual 清理仓库显式标成 legacy cleanup。
- 运行时链路继续瘦身：清掉了 `menu/user` 模块里不再参与主链的旧依赖注入，团队角色菜单/能力响应不再暴露 `has_*` 决策字段，快照服务只保留仍有业务意义的继承信息。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 继续检查 `backend/internal/modules/system/models/model.go` 与相关快照表字段，把仅剩的历史列和运行时实际消费语义再分层标注，降低后续误读成本。
- 如果继续收尾权限改版，下一步优先看平台/团队快照刷新入口是否还能再并口，进一步压缩“缺失即现算”的残余分支。

## 2026-03-24 快照模型降噪与刷新入口并口

### 本次改动
- 团队快照模型继续按迁移期语义降噪：`team_access_snapshots.manual_action_ids` 在 Go 模型层改名为 `LegacyManualActionIDs`，明确它只剩兼容/迁移职责；团队角色快照里的 `has_menu_boundary` 已从运行时模型移除，不再作为主链判断信号。
- 团队快照服务对外刷新入口统一为 `RefreshSnapshot`，并把权限、功能包、团队动作保存等残留调用全部并到这一个入口，减少“RefreshCache/缺失即现算”双重语义继续扩散。
- 共享权限来源面板的 blocked 卡片样式语义从 `manual` 收口为 `trimmed`，继续贴近“功能包展开 + 减法裁剪 + 快照读取”的单主链表达。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 继续检查共享权限组件与少量非主页面 API 是否还保留旧字段命名或旧语义别名，避免前端新页面再次把兼容字段带回主链。
- 继续沿 `model.go` 和快照表字段梳理其余历史列，能降级成兼容/迁移语义的继续降级，不能删库的先在模型层与服务层彻底去主链化。

## 2026-03-24 审计页与角色边界 API 语义收口

### 本次改动
- 平台用户与团队成员历史审计页继续去旧语义：用户权限/菜单弹窗里的 `compat-banner` 改成 `audit-banner`，成员权限例外页与入口文案改成“权限例外审计”，局部 `manual*` 状态和样式也统一收口到团队边界裁剪语义。
- 平台角色、团队角色边界 API 包装层继续固定主字段：`expanded_package_ids` 现在优先作为主输出，必要时向下兼容 `package_ids`；类型定义中也把 `package_ids` 明确标成迁移兼容字段，减少新页面继续把它当主链字段消费。
- 用户与团队成员动作审计接口的包装和后端命名同步降噪：前端 `fetchGetUserActions`/`fetchSetUserActions` 改成审计语义说明，`*Overrides` 仅保留兼容别名；后端用户动作接口摘要与 handler 内部变量也统一改成 audit/审计命名。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 继续清团队角色 action 弹窗和团队边界页里仍偏旧的本地变量名，把 `derived` 与“候选集”语义再彻底拆开，避免主页面心智继续混用。
- 继续沿前后端快照模型与 API 返回清剩余 legacy 别名，优先处理还可能影响新页面开发的字段命名和注释，而不是只改文案。

## 2026-03-24 旧现算分支与候选集命名全量收口

### 本次改动
- 团队动作快照运行时已彻底切到“功能包展开 - 团队屏蔽”单主链：`teamboundary` 不再把旧 `team_manual_action_permissions` 或旧 `tenant_action_permissions` 缓存结果并入 `effective_ids`，`from_cache` 也已从后端响应、前端 API 类型和页面提示里完全移除。
- 鉴权侧同步去掉了旧 legacy 参与：`authorization` 不再把 `LegacyManualIDs` 视为团队边界已配置条件，团队动作来源接口只再暴露 `derived_* / blocked_*`，避免旧正向表继续影响平台/团队运行时判断。
- 前端主页面继续一次性统一候选集命名：平台用户菜单裁剪、团队边界、团队菜单边界、团队角色能力裁剪、团队成员审计页里的 `derived*` 局部状态已经收口成 `candidate* / available*` 语义；平台用户动作页也直接改用 `fetchGetUserActions`，删除 `*Overrides` 兼容别名。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 如果继续往下收尾，下一步就不该再停留在命名层，而是直接清数据库迁移与 repository 里仅剩的旧 `team_manual_action_permissions` / `tenant_action_permissions` 兼容职责，准备彻底下线旧表。
- 再扫一轮 `model.go`、`database.go` 和迁移脚本，把目前只剩清理职责的历史列、历史索引和历史仓库显式标成 legacy cleanup，最后为删库做准备。
## 2026-03-24 快照缺失分支并口与历史字段继续降级

### 本次改动
- 团队边界快照、团队角色快照、平台角色快照读取全部改为严格走快照表；快照缺失时不再在读取链路上现算补链，团队 `manual_action_ids` 也彻底降级为仅落库兼容字段，不再参与运行时计算。
- 平台角色/团队角色边界服务与前端 API 包装继续收口到 `expanded_package_ids + available_* + hidden/disabled_* + derived_sources` 主字段；团队成员与团队管理接口移除了 `role` 旧兼容入参，只保留 `role_code`。
- README 与类型定义同步去掉旧主链暗示；本轮尚未执行校验，待统一跑 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续清 `model.go`、迁移脚本和 AutoMigrate 中仅剩的 legacy 列/旧表说明，把能降级为迁移清理职责的字段全部显式标成 legacy cleanup。
- 继续扫共享权限组件和非主页面 API，避免 `package_ids`、旧审计别名或历史命名再次从前端包装层回流到新页面。

## 2026-03-24 ???????????????
### ????
- ?????????????????AutoMigrate????????/??????????? `tenant_action_permissions`?`team_manual_action_permissions`?`role_menus`?`role_action_permissions` ? `team_access_snapshots.manual_action_ids` ???????
- `backend/cmd/migrate/main.go` ?? `20260324_drop_legacy_permission_tables` ????????????????????? seed???? rebinding ?? scope ?????`main.go` ???????/???????? UTF-8 ???
- ????`backend/go test ./...`?`frontend/pnpm exec vue-tsc --noEmit`?`frontend/pnpm build` ?????
### ????
- ??????????????????/????????????????????????????????
- ??????????????????????? API???????????????????
