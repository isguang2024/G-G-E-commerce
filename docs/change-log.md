> 历史说明：本文件较早记录中提到的 `permission-package-design.md` 已于 2026-03-25 收敛下线，相关设计说明统一以 `permission-overall-summary.md` 为准。

## 2026-03-23 权限系统设计文档重写

### 本次改动
- 按最新确认的“功能包优先”模型，重写了权限总览文档；其中旧 `permission-package-design.md` 的内容现已并入 [permission-overall-summary.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/docs/permission-overall-summary.md)。
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
- 更新权限总览文档，将平台用户减法裁剪收口为“用户隐藏菜单”，并新增“快照刷新机制”章节。
- 更新 `docs/permission-overall-summary.md`，补充快照刷新触发源清单，明确 `platform` 与 `team` 需分上下文刷新。
- 触发源已覆盖：功能包绑定、组合包关系、菜单启停、角色包、用户直绑包、角色减法、用户隐藏菜单、团队开包、团队边界减法、团队角色裁剪、成员角色、功能包状态变更。

### 下次方向
- 按新规范为平台用户、平台角色、团队边界、团队角色四条写链补统一刷新入口。
- 将快照刷新从文档规则落实到代码实现，避免出现规则正确但刷新不全的情况。

## 2026-03-23 统一快照刷新入口接入首批写链

### 本次改动
- 新增 `backend/internal/pkg/permissionrefresh/service.go`，把团队刷新、平台用户刷新、按功能包传播刷新、按菜单传播刷新统一收口为显式编排入口；第一阶段仍采用同步刷新，不引入事件总线。
- 功能包、平台角色、平台用户、菜单更新四条关键写链已开始接入统一刷新入口：功能包变更会传播到相关团队/平台主体，平台角色包与减法裁剪变更会刷新关联用户，平台用户绑包/分配角色/旧个人权限变更会刷新该用户，菜单更新/删除会走按菜单传播刷新。
- 同步补强了权限总览与规范文档，进一步明确“刷新必须在写事务成功后触发”；验证通过 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

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
- 权限总览文档已同步写明：团队成员个人例外处于迁移期只读审计状态。已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。

### 下次方向
- 继续处理团队边界“屏蔽菜单”缺口，把团队边界从“只有权限减法”补成“菜单 + 权限”完整减法模型。
- 后续再决定是否彻底下线团队成员个人例外写接口，避免后端仍然保留可写兼容链。

## 2026-03-24 团队角色来源跳转增强与规范文档对齐

### 本次改动
- 团队角色菜单裁剪页与权限裁剪页的“展开能力”和“已屏蔽能力”标签都支持直接点击跳转来源功能包，排查链路从“看结果”缩短到“定位配置”。
- 平台“用户权限例外（兼容）”页新增未绑定功能包时的“前往绑定功能包”引导按钮，并在角色继承区域补充“功能包约束/兼容回退”状态标签，进一步明确该页是兼容入口而非主入口。
- 权限总览文档已同步更新当前实现状态，清理重复错位条目，并补充团队侧与平台侧兼容入口的现状说明。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`；本轮未启动服务。

### 下次方向
- 继续把平台用户兼容入口降级为纯例外配置，并补齐“未绑定功能包时保存风险”防误操作策略。
- 推进团队角色/成员页对旧兼容字段的依赖清理，逐步转向“包展开 + 减法裁剪 + 快照刷新”单一心智。

## 2026-03-24 团队菜单边界主链落地

### 本次改动
- 后端已补齐团队菜单边界接口：新增团队菜单快照读取、菜单来源返回和菜单边界保存，团队菜单边界现在正式以“功能包展开菜单 - 团队屏蔽菜单”工作。
- 前端团队页新增“菜单边界”入口，并新增团队菜单边界弹窗；当前支持按功能包来源查看菜单、按来源功能包跳转，以及“全部保留/全部屏蔽”批量操作。
- 文档已同步收口：权限总览文档已明确记录团队边界已形成“菜单 + 权限”双减法模型。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build`。本轮未启动服务。

### 下次方向
- 继续收平台用户主链，把平台侧稳定到“全局角色包并集 + 用户直绑包 + 用户隐藏菜单”的正式模型，进一步弱化 `user_action_permissions` 的可写入口。
- 继续把团队成员与团队角色页面里的旧兼容字段降级为审计输出，前端主配置入口统一收成“功能包开通 + 菜单/权限减法裁剪 + 快照刷新”。

## 2026-03-24 团队菜单边界主入口落地

### 本次改动
- 新增团队菜单边界前端主入口：团队列表操作菜单补齐“菜单边界”，并挂载新弹窗 `frontend/src/views/team/team/modules/team-menu-permission-dialog.vue`。
- 新增团队菜单边界前端 API：`fetchGetTeamMenus`、`fetchGetTeamMenuOrigins`、`fetchSetTeamMenus`，并补齐团队边界返回字段（含 `expandedPackageIds`、`blockedActionIds`）。
- 团队菜单边界弹窗支持来源分组、功能包筛选、功能包跳转、全保留/全屏蔽批量操作，保存语义固定为“功能包展开菜单 - 团队屏蔽菜单”。
- 同步更新权限总览文档，将团队边界正式收口为“菜单减法 + 权限减法”双边界模型。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续收 platform 主链，将平台用户侧正式固定到“全局角色包并集 + 用户直绑包 + 用户隐藏菜单”，进一步弱化 `user_action_permissions` 兼容入口。
- 把团队边界与团队角色页面的来源视图继续统一成同一套交互（来源分组 + 目标包跳转 + 屏蔽明细），减少排障切换成本。
- 在后续一节推进数据库快照表落地，逐步减少运行时现算路径，确保刷新触发后读链稳定命中快照。

## 2026-03-24 平台用户菜单裁剪主链落地

### 本次改动
- 后端已补齐平台用户菜单裁剪接口，用户菜单读写现在基于平台快照的“候选菜单 + 最终菜单”双层结构工作，支持从功能包展开菜单中反推出用户隐藏菜单，并在保存后统一刷新平台用户快照。
- 平台快照已补充候选菜单层，避免“已隐藏菜单无法恢复”的问题；用户管理页新增“菜单裁剪”入口，并新增 `user-menu-selector-dialog.vue`，当前支持功能包来源筛选、目标功能包跳转、全部保留/全部隐藏批量操作。
- 平台用户菜单裁剪已成为正式入口，平台用户“权限例外（兼容）”继续保留但降级为兼容副入口；权限总览文档已同步写明这一点。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

### 下次方向
- 继续把平台用户“权限例外（兼容）”写链再降一级，明确只承担少量历史例外能力，不再与主链并列。
- 推进团队角色和平台角色两侧来源视图统一，把“包展开结果 / 显式屏蔽结果 / 最终生效结果”做成同一套查看交互。
- 在平台和团队两侧主入口都稳定后，开始把现算路径进一步收口成数据库快照表，减少运行时临时拼装。

## 2026-03-24 平台用户兼容例外降级为只读审计

### 本次改动
- 平台用户“权限例外（兼容）”页已从可写入口降级为只读审计视图：当前保留历史 allow/deny 例外、角色继承、来源功能包与结果预览，但退出保存与交互修改，不再和平台主链并列。
- 用户管理页操作文案同步收口为“权限例外审计”，强调平台侧主配置应优先使用“功能包 + 菜单裁剪”；平台角色权限页也新增了功能包摘要、菜单已屏蔽数、能力已关闭数，开始统一平台侧来源与减法心智。
- 权限总览文档已同步更新当前状态。已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过；本轮未启动服务。

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
- 继续整理并压缩旧设计文档沉淀的重复段落，保持权限总览只表达当前正式状态与实现差距。
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

## 2026-03-24 旧权限表物理删除与边界字段最终收口

### 本次改动
- 后端已正式删除旧权限表与历史快照列：`tenant_action_permissions`、`team_manual_action_permissions`、`role_menus`、`role_action_permissions` 与 `team_access_snapshots.manual_action_ids` 不再存在于运行时模型、AutoMigrate 与迁移主链中。
- `backend/cmd/migrate/main.go` 已补齐 `20260324_drop_legacy_permission_tables`，并修复文件中文乱码问题；旧 seed、rebind、scope 清理逻辑也已同步收口，不再继续引用已删除旧表。
- 前端角色边界 API 包装层与类型定义继续收口为 `expanded_package_ids + available_* + hidden/disabled_* + derived_sources` 单主链；README 与设计文档已改成“旧表已物理删除”的现势口径。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 继续扫一轮非主页面与低频接口，确认没有新的边界页再把 `package_ids` 或历史审计别名带回角色/团队边界主链。
- 如继续做权限改版收尾，下一步优先把刷新入口、快照回填任务和迁移文档再并口，减少后续维护时对历史阶段的误读。

## 2026-03-24 最后一轮旧数据拔除

### 本次改动
- 平台角色与团队角色边界接口已去掉旧兼容输出：后端不再返回 `package_ids`、`has_menu_boundary`、`has_package_boundary` 这类旧字段，前后端统一只按 `expanded_package_ids + available_* + hidden/disabled_* + derived_sources` 主链消费。
- 平台角色快照服务已移除 `HasPackageConfig` 历史判断，`platform_role_access_snapshots` 运行时模型不再保留该列；迁移新增 `20260324_drop_legacy_snapshot_columns`，同步删除 `platform_role_access_snapshots.has_package_config` 与 `team_role_access_snapshots.has_menu_boundary`。
- 这轮目标是给后续权限系统新增/改造功能清空历史负担，旧快照列、旧边界响应字段和残余兼容判断已继续做物理收口。
- 待执行：`go run ./cmd/migrate`，随后统一跑后后端与前端校验。

### 下次方向
- 进入权限系统新功能开发前，只需再做一轮库结构与全仓校验确认；如果迁移已执行通过，后续可以默认按新快照主链直接开发，不再为旧边界别名兜底。
- 新增接口或页面时继续坚持“功能包展开 + 减法裁剪 + 快照读取”，不要重新引入旧快照字段、旧正向表语义或边界兼容别名。

## 2026-03-24 权限文档缺口整理

### 本次改动
- 重写了 `docs/permission-overall-summary.md` 和 `docs/permission-package-design.md` 里“当前实现状态”的表达，不再只写抽象原则，改为明确区分“已落地能力”和“仍缺功能”。
- 已确认功能包集合不是完全未做：底层已有 `feature_package_bundles`、`package_type`、运行时递归展开和刷新联动；真正缺的是组合包关系管理 API、前端页面、约束校验、展开预览和初始化运营能力。
- 两份文档都已补上下一阶段建议，明确后续权限系统扩展前应先补齐组合包产品能力，再进入常规新功能接入。
- 本轮未执行测试；仅整理文档与缺口结论，无运行时代码变更。

### 下次方向
- 直接进入组合包能力补齐：先做后端子包关系读写接口与校验，再做功能包页面的组合包配置弹窗和展开预览。
- 组合包补齐后，再开始新增权限模块、基础包与默认授予策略，避免新功能继续绕开组合包体系单独堆配置。

## 2026-03-24 功能包组合闭环与管理员默认包切换

### 本次改动
- 后端补齐了组合包基础包读写接口，`featurepackage` 服务新增子包校验、上下文兼容判断、删除联动刷新和 `/feature-packages/:id/children` 路由，组合包正式可按“基础包集合”维护。
- 迁移初始化补上了内置 `platform.admin_bundle` 组合包及其默认子包关系，平台管理员默认授予从三个基础包改为单个组合包，并清理 `admin` 角色上的旧基础包遗留绑定。
- 前端功能包管理页改成“基础包 / 组合包”双页签；基础包继续管理功能范围和菜单，组合包改为独立“配置基础包”弹窗，避免旧的基础包直配心智回流。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 全部通过。

### 下次方向
- 继续补组合包展开预览、组合来源展示和角色/团队授予页的组合包可视化，方便后续新增权限功能时直接围绕组合包开发。
- 再扫 `permission-overall-summary` / `permission-package-design`，把“组合包已落地、仍缺预览/运营能力”的现状补齐，避免文档还停留在缺口阶段。

## 2026-03-24 组合包展开预览与文档状态回填

### 本次改动
- 新增共享组件 `FeaturePackageGrantPreview`，平台角色、团队角色、团队开通三个授予弹窗现在都能直接预览“直绑包 -> 展开基础包 -> 展开权限/菜单”的结果，并附带来源提示。
- 来源面板跳转到功能包管理时已带上 `packageType`，点击组合包来源不会再误落到基础包页签。
- `permission-overall-summary` 与 `permission-package-design` 已回填为当前现状：组合包关系 API、管理页、双页签、默认管理员组合包、授予页展开预览都已落地，剩余缺口改成运营视图、循环校验和影响范围总览。
- 已验证：`frontend/pnpm exec vue-tsc --noEmit`、`frontend/pnpm build` 通过；本轮未改后端逻辑，未再执行 `go test`。

### 下次方向
- 继续做组合包详情级运营视图，把“组合包引用了哪些基础包、影响了哪些角色/团队、展开出哪些菜单/权限”整理成可筛选页面，而不只停留在弹窗预览。
- 补循环依赖检测、停用影响提示和最小自动化测试，避免组合包正式进入运营后再把约束问题带回运行时。

## 2026-03-24 ?????????

### ????
- ?????????????????????????????????????????????/??????????????????????????????
- ??? [FeaturePackageGrantPreview.vue](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/components/business/permission/FeaturePackageGrantPreview.vue) ???????????????????????????????????????
- ???????????????????????????????????????? `frontend/pnpm exec vue-tsc --noEmit` ? `backend/go test ./...` ?????????????????

### ????
- ????????????????????????????????????????????????????????
- ?????????????????????????????????????????????????
## 2026-03-24 功能包弹窗莫兰迪语义配色统一

### 本次改动
- 统一调整平台角色、团队角色、团队功能包、用户功能包四类弹窗与功能包预览组件的标签配色，按基础包、组合包、平台、团队、平台/团队、正常、停用、异常等语义分别映射到低饱和度色板。
- 同步将复选框选中态、主按钮强调色、输入框聚焦态收敛到静谧蓝主色，保证弹窗整体视觉统一，减少此前中性色和多色标签混杂的问题。
- 已通过 `pnpm exec vue-tsc --noEmit` 验证，未执行视觉回归截图比对。

### 下次方向
- 建议补一轮真实数据联调，确认停用、异常、平台/团队混合上下文等少量状态在所有弹窗里的显示是否符合预期。
- 如需继续统一系统视觉，可把这套 tag token 抽到公共样式层，避免多个弹窗各自维护同一套颜色变量。
## 2026-03-24 功能包弹窗回归组件库主题风格

### 本次改动
- 全量回查并调整平台角色、团队角色、团队功能包、用户功能包弹窗及功能包预览组件，移除上一轮新增的硬编码商务配色、按钮和复选框皮肤，恢复到组件库原生语义色与主题变量体系。
- 标签展示改回基于 `ElTag` 的 `type` 语义映射，容器、表格和展开区仅保留布局与 `var(--default-border)`、`var(--default-bg-color)`、`var(--default-box-color)` 等主题变量，重新兼容深色/浅色模式。
- 未改动筛选、展开、保存等业务逻辑；已通过 `pnpm exec vue-tsc --noEmit` 验证。

### 下次方向
- 建议在浅色与深色模式下各回归一次这几类弹窗，重点确认 `ElTag` 的 `primary/success/warning/info/danger` 语义是否完全符合业务认知。
- 如果后续还要做视觉增强，优先通过全局主题变量或组件库主题配置实现，避免再次在单页里写死颜色覆盖。
## 2026-03-24 08:41:05 用户权限例外审计弹窗下线

### 本次改动

- 删除了用户管理页里的“权限例外审计”入口和对应弹窗，保留功能包、菜单裁剪等主链能力不变。
- 移除了前端独占 API `GET/PUT /api/v1/users/:id/actions` 的调用封装，以及后端 `/users/:id/actions` 路由和处理逻辑。
- 清理了用户模块里仅服务该弹窗的注入依赖，避免保留无效接线；已通过 `pnpm exec vue-tsc --noEmit` 和 `go test ./...` 验证。

### 下次方向

- 建议继续检查权限相关页面是否还保留“历史兼容/审计”文案，统一收口到功能包、角色和菜单裁剪主链。
- 若后续需要恢复只读审计能力，建议独立成后台审计页，而不是重新挂回用户配置入口。
## 2026-03-24 团队管理权限键合并与迁移

### 本次改动
- 合并了团队管理相关的细粒度权限键：平台侧将 `tenant.member.manage`、`tenant.boundary.manage` 并入 `tenant.manage`，团队侧将 `team.member.assign_role`、`team.member.assign_action` 并入 `team.member.manage`，保留 `team.boundary.manage` 独立存在。
- 同步修改了后端权限映射、团队模块接口鉴权和前端团队页/团队成员页的 `hasAction` 判断，避免继续依赖已废弃的旧权限键。
- 扩展了迁移逻辑，自动回写 `feature_package_actions`、`role_disabled_actions`、`team_blocked_actions`、`user_action_permissions` 等引用表并软删除旧权限键；已执行 `backend/go run cmd/migrate/main.go`、`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 验证通过。

### 下次方向
- 建议继续检查团队相关功能包和角色是否还保留旧文案，避免页面上继续出现“成员角色配置/成员功能权限配置”这类已经被合并的概念。
- 如果后续决定进一步收敛团队边界能力，可以再评估是否把 `team.boundary.manage` 也并入 `team.member.manage` 或单独抽成更清晰的团队管理包。

## 2026-03-24 功能权限查看接口链路补齐

### 本次改动
- 给功能权限管理页补了“查看接口”入口，点击后会跳到 API 管理页并自动按当前权限键过滤，直接查看该权限组下的全部接口。
- 后端 `/api/v1/api-endpoints` 列表新增 `permission_key` 查询支持，前端 API 管理页同步接入路由筛选态和清除筛选入口，不额外新增接口。
- 已验证 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过。

### 下次方向
- 如果后面还要继续增强，可再补一个接口数量列，直接在功能权限列表上展示“关联接口数”。
- 若 API 总量继续增长，再把 `permission_key` 过滤下沉到仓储 SQL 层，避免当前服务层内存过滤持续放大。

## 2026-03-24 功能权限页内查看关联接口

### 本次改动
- 新增后端接口 `GET /api/v1/permission-actions/:id/endpoints`，按功能权限 ID 直接返回该权限组关联的全部接口，避免继续依赖 API 管理页的跳转筛选。
- 功能权限页改成页内弹窗查看接口列表，支持直接查看接口规格、鉴权模式、处理函数和状态，交互路径更短。
- 已验证 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过。

### 下次方向
- 可以继续补“复制接口路径”“按方法筛选”等轻量操作，方便直接在弹窗里排查问题。
- 如果后面确认 API 管理页不再需要 `permission_key` 跳转筛选，可再收掉那层兼容逻辑。 

## 2026-03-24 功能权限旧键与旧文案清理

### 本次改动
- 清理了菜单边界相关的旧点式权限键，把 `tenant.configure_menu_boundary` 并入 `tenant.manage`，把 `team.configure_menu_boundary` 并入 `team.boundary.manage`，不再依赖运行时兼容。
- 补上了 [permissionkey.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/permissionkey/permissionkey.go) 中 `team:configure_menu_boundary` 的源头映射，并重新执行迁移，避免 API 注册同步再次生成旧权限键。
- 已重新执行 `backend/go run cmd/migrate/main.go` 与 `backend/go test ./...`，数据库当前生效集里仅剩 `tenant.manage` 和 `team.boundary.manage`。

### 下次方向
- 建议继续把功能权限列表里的展示名和描述统一收口成权限组语义，减少“获取/配置某接口边界”这类接口化表达。
- 如果要进一步降噪，可以继续把 `feature_package.assign_menu`、`user.assign_menu` 这类自动注册名称也统一成业务名。 
## 2026-03-24 删除团队成员权限例外审计

### 本次改动
- 删除了团队成员页里的“权限例外审计”入口、对应弹窗 [member-action-dialog.vue](/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/src/views/team/team-members/modules/member-action-dialog.vue) 以及相关前端请求封装，团队成员页只保留“分配角色”和“删除”主链操作。
- 删除了后端 `/api/v1/tenants/my-team/members/:userId/actions` 读写接口及其 handler 实现，并清理了 Tenant 模块里仅供该链路使用的仓储注入和死代码，避免界面下线后仍保留无用 API。
- 已执行 `frontend/pnpm exec vue-tsc --noEmit` 与 `backend/go test ./...`，通过；全局检索确认当前仓库不再有该弹窗、接口封装和入口调用残留。

### 下次方向
- 如果后续还要继续收口团队权限体系，可以把功能权限列表里与成员例外审计相关的历史说明文案一并清理，避免页面语义继续暗示这条旧能力存在。
- 若数据库里仍保留该接口历史注册记录，下一步可在迁移里补一条专门的 API 注册表清理，进一步收干净历史痕迹。

## 2026-03-24 团队管理员菜单与前端身份收口

### 本次改动
- 将团队管理员默认功能包 `team.member_admin` 的菜单范围收口为 `TeamRoot`、`TeamMembers`、`TeamRolesAndPermissions`，不再默认挂入平台侧的 `TeamManagement`，避免团队管理员登录后命中平台团队管理链路。
- 新增迁移 `20260324_team_menu_single_context_cleanup`，把现有库里的 `TeamRoot`、`TeamMembers`、`TeamRolesAndPermissions` 菜单统一纠正到团队菜单树下，并重置它们的 `meta`，清掉旧的前端角色限制和遗留动作配置。
- 前端将非超级管理员统一映射为单一身份 `R_USER`，同时把 `hasPlatformAccess` 改为仅识别 `system.* / platform.* / tenant.*` 平台动作前缀，避免团队用户拿到 `team.*` 后被误判成平台态；已执行 `backend/go test ./...`、`backend/go run cmd/migrate/main.go`、`frontend/pnpm exec vue-tsc --noEmit` 验证。

### 下次方向
- 建议用一个仅绑定团队和 `team_admin` 的新账号重新登录，重点验证默认落点是否进入 `/team/*`，以及“团队成员”“团队角色及权限”两个页面是否都不再跳 403。
- 如果后续决定彻底删除平台态团队管理链路，再继续下线 `tenant.manage` 相关页面、接口和功能包，当前这次先完成团队管理员主链收口，未动平台管理员团队管理能力。

## 2026-03-24 登录切账号菜单重建修复

### 本次改动
- 在登录页成功登录后，先调用路由守卫的重置逻辑清空旧账号的动态路由和菜单状态，再写入新 token 和用户信息，避免同一标签页切换账号时继续沿用上一个账号的菜单缓存。
- 这样可以修复“新账号登录后没有重新发起菜单请求，直接落入 /403 或首页”的问题，尤其是从管理员切到团队账号这类跨权限范围切换场景。
- 已执行 `frontend/pnpm exec vue-tsc --noEmit` 验证通过。

### 下次方向
- 建议再观察一次从管理员账号直接切到团队管理员账号的场景，确认登录后一定会重新拉菜单树。
- 如果后续仍有偶发状态残留，可以再把登录成功后的 tenant 状态和 worktab 状态做更彻底的同步重置。
## 2026-03-24 功能权限关联接口与权限组同步修复

### 本次改动
- 修正功能权限关联接口查询逻辑，统一按规范权限组匹配接口，不再因旧 permission_key 文本导致返回空数组。
- 在迁移中新增规范权限组同步，自动把活跃的 permission_actions 名称、描述、上下文、资源码和动作码收口到当前映射定义。
- 已执行 go test ./...、go run cmd/migrate/main.go，并重启 8080 后端进程使新逻辑生效。

### 下次方向
- 建议继续清理功能权限页仍可能残留的旧展示文案，并观察是否还有历史缓存导致的旧项显示。
- 若后续继续合并权限组，保持先改映射、再跑迁移、最后重启服务，避免再次出现列表与关联接口不同步。

## 2026-03-24 API固定GUID与重建流程收口

### 本次改动
- 重构了接口注册核心逻辑：`api_endpoints.code` 改为默认按 `METHOD + PATH` 生成稳定 GUID，`permission_actions.code` 改为按 `permission_key` 生成稳定 GUID，避免每次注册出现随机标识导致关联漂移。
- 在 `apiregistry` 新增 `GETAction/POSTAction/PUTAction/DELETEAction` 与 `*Protected` 封装，路由可直接用权限键注入鉴权中间件并自动补元数据；已落地到 API 管理、角色、用户、功能权限、功能包、菜单、系统模块。
- 新增 `POST /api/v1/api-endpoints/rebuild`：测试环境可一键重建 API/权限/功能包基础数据（保留菜单列表，保留并重建默认管理员、默认角色、默认基础包与组合包），并在重建后自动重新同步 API 注册表。
- 前端 API 管理页新增“重建数据”按钮与二次确认，调用新接口完成重建。
- 已验证 `backend/go test ./internal/pkg/apiregistry ./internal/modules/system/apiendpoint ./internal/modules/system/role ./internal/modules/system/user ./internal/modules/system/permission ./internal/modules/system/featurepackage ./internal/modules/system/menu ./internal/modules/system/system` 与 `frontend/pnpm exec vue-tsc --noEmit` 通过。

### 下次方向
- 建议在测试库先执行一次“重建数据”，重点核对默认功能包菜单挂载与角色包绑定是否符合预期。
- 若后续希望在生产也具备安全重建能力，可再补“白名单环境 + 手工确认令牌 + 审计日志”三重保护，而不是仅按 `env != production` 放行。

## 2026-03-24 Tenant 路由封装收口

### 本次改动
- 将 Tenant 模块中所有需要鉴权的接口统一改为 `GETAction/POSTAction/PUTAction/DELETEAction` 封装，直接绑定 `permission_key`，不再手工维护 `ResourceCode/ActionCode` 与 `RequireAction(...)` 组合写法。
- 团队侧接口统一使用 `team.member.manage` 与 `team.boundary.manage`，平台团队管理接口统一使用 `tenant.manage`，避免历史旧权限键残留导致注册元数据漂移。
- 已执行 `backend/go test ./internal/modules/system/tenant ./internal/modules/system/apiendpoint ./internal/pkg/apiregistry` 验证通过。

### 下次方向
- 建议在测试环境实际调用一次 `POST /api/v1/api-endpoints/rebuild`，确认 Tenant 相关接口在重建后仍能按权限键正确聚合。
- 若后续继续统一接口定义，可按同样方式扫一遍非 system 模块，确保路由注册风格一致。

## 2026-03-24 API 重建联调与兼容修复

### 本次改动
- 为 API 注册新增 `system.api_registry.rebuild` 权限映射，并把默认平台接口管理包描述与默认权限清单补齐（包含 `view/sync/rebuild`）。
- 修复 `apiregistry` 的接口写入逻辑：`upsertEndpoint` 先按 `code` 查找，未命中时再按 `method+path` 回退匹配并更新，解决旧数据升级后 `sync` 触发 `idx_api_endpoints_method_path_unique` 冲突的问题。
- 完成真实联调：执行 `go run cmd/migrate/main.go` 后，已验证 `POST /api/v1/api-endpoints/sync` 与 `POST /api/v1/api-endpoints/rebuild` 返回成功；并验证默认角色（`admin/team_admin/team_member`）、默认功能包（基础包+组合包）和角色绑定重建正常，菜单树接口可正常返回。
- 当前 `rebuild` 接口鉴权仍复用 `system.api_registry.sync`（保证旧环境立即可用），前端“重建数据”按钮也保持该权限键可见，避免新权限未分配时出现 403。

### 下次方向
- 若要彻底拆分高危权限，建议升级 `authorization.resolvePermissionKey` 支持“主权限键 + 备选权限键”而非旧 `resource/action` 双参语义，再把 `rebuild` 路由切回独立 `system.api_registry.rebuild`。
- 可补一条迁移，把现有管理员角色默认补授 `system.api_registry.rebuild`，完成权限平滑切换并移除兼容路径。

## 2026-03-24 Rebuild 独立权限与默认快照刷新

### 本次改动
- 在鉴权层新增 `RequireAnyAction/AuthorizeAny`，并将 `POST /api/v1/api-endpoints/rebuild` 的注册元数据切到独立权限 `system.api_registry.rebuild`；实际放行逻辑同时兼容旧权限 `system.api_registry.sync`，保证旧测试库迁移前后都能过渡访问。
- 迁移脚本补充 `api_endpoint/rebuild` 默认权限种子，把 `platform.api_admin` 默认权限清单更新为 `view/sync/rebuild`，并在迁移结束后主动刷新默认管理员角色与默认功能包快照，避免已有账号继续使用旧权限快照。
- 前端 API 管理页“重建数据”按钮已切换到 `v-action='system.api_registry.rebuild'`；已重新执行 `go test ./internal/pkg/authorization ./internal/modules/system/apiendpoint ./internal/pkg/apiregistry ./internal/pkg/permissionkey`、`pnpm -C frontend exec vue-tsc --noEmit`、`go run cmd/migrate/main.go`，并在重启 8080 最新后端后实测 `POST /api/v1/api-endpoints/sync`、`POST /api/v1/api-endpoints/rebuild` 均返回 `code=0`，按 `permission_key=system.api_registry.rebuild` 可查到重建接口，`platform.api_admin` 已包含该权限。

### 下次方向
- 建议再用非超级管理员但拥有平台接口管理包的账号回归一次前端页面，确认“重建数据”按钮显隐与刷新后的权限缓存完全一致。
- 等旧环境都执行过这次迁移后，可再移除 `rebuild -> sync` 的兼容放行分支，彻底把高风险操作权限独立出来。

## 2026-03-24 团队身份角色回填与团队菜单恢复

### 本次改动
- 修正了团队菜单链路对 `user_roles` 的硬依赖：当团队成员的团队作用域角色记录缺失时，后端现在会回退到 `tenant_members.role_code` 解析默认身份角色，避免旧数据直接导致团队菜单、团队鉴权和团队角色列表全部判空。
- 扩展了团队边界快照构建逻辑，`team_role_access_snapshots` 会把活跃团队成员的身份角色一并纳入；迁移中新增团队身份角色回填，会为现有 `tenant_members` 自动补齐团队作用域 `user_roles`，并刷新对应团队快照。
- 前端平台用户相关接口补充 `skipTenantHeader`，避免在团队上下文里把 `X-Tenant-ID` 混入平台用户功能包/菜单裁剪请求，减少上下文误判。
- 已执行 `go test ./internal/modules/system/user ./internal/pkg/teamboundary ./cmd/migrate`、`pnpm -C frontend exec vue-tsc --noEmit`、`go run cmd/migrate/main.go`，并针对团队 `5767135e-9476-4cf9-922e-496cb6f7e193` 验证：团队作用域 `user_roles` 已回填出 `team_admin`，`team_role_access_snapshots` 已生成，8080 后端已重启到最新代码。

### 下次方向
- 建议直接用该团队管理员账号重新登录，确认侧边栏已能进入 `/team/*`，并核对“团队成员”“团队角色及权限”两个页面是否都恢复正常。
- `platform.api_admin` 属于平台上下文功能包，不会出现在团队角色功能包或团队菜单链路里；如果后续还要继续降误解，建议把平台/团队功能包入口文案再做一次上下文区分。

## 2026-03-24 团队角色功能包并入团队边界接口

### 本次改动
- 在 Tenant 模块新增一组“当前团队功能边界管理”角色接口：`/api/v1/tenants/my-team/boundary/roles` 及其 `packages/menus/actions` 子路由，统一绑定 `team.boundary.manage` 权限，复用原有处理逻辑，不新增重复数据模型。
- 团队角色权限页面改为调用上述边界接口（角色列表、功能包、菜单裁剪、权限裁剪），确保团队管理员即使没有 `team.member.manage`，也可获取并管理角色功能包链路。
- 页面操作按钮按权限分层：新增/编辑/删除角色继续要求 `team.member.manage`；功能包/菜单/权限裁剪走 `team.boundary.manage`，避免误触发 403。
- 已执行 `backend/go test ./internal/modules/system/tenant/...` 与 `frontend/pnpm exec vue-tsc --noEmit`，通过。

### 下次方向
- 建议用团队管理员账号携带 `X-Tenant-ID=5767135e-9476-4cf9-922e-496cb6f7e193` 回归：验证“团队角色及权限”可打开、功能包可保存、菜单裁剪可保存。
- 若后续要彻底去掉旧链路，可在确认前端全部切换后下线 `/my-team/roles/:roleId/*` 上 `team.member.manage` 的同功能接口，减少权限歧义。

## 2026-03-24 团队角色功能包列表权限修复

### 本次改动
- 新增团队边界接口 `GET /api/v1/tenants/my-team/boundary/packages`（权限键 `team.boundary.manage`），返回当前团队已开通且团队上下文可用的功能包列表。
- 团队角色功能包弹窗改为调用团队边界功能包接口，不再调用平台接口 `/api/v1/feature-packages`，解决团队管理员在该页面出现 `code=2003` 的无权限问题。
- 已执行 `backend/go test ./internal/modules/system/tenant/...` 与 `frontend/pnpm exec vue-tsc --noEmit`，通过。

### 下次方向
- 建议用团队管理员账号再次验证“团队角色 -> 功能包”弹窗的加载与保存，确认不再触发 `/feature-packages` 403/2003。
- 若后续还存在角色裁剪链路中的平台接口调用，可继续同样收口到 `team.boundary.manage` 域，彻底避免团队态串到平台权限。

## 2026-03-24 权限与 API 体系硬切换（无兼容）

### 本次改动
- 后端 API 注册链路已硬切换为“纯 API 元信息 + 多权限绑定”模型：`apiendpoint/service` 移除 `rebuild` 全链路与旧重建逻辑，仅保留 `List/Save/Sync/ListBindings`；`Save` 支持多权限键绑定（`api_endpoint_permission_bindings`，`match_mode=ANY`）并对 `method+path` 做唯一校验。
- API 元信息改造继续落地：新增并透传 `group_code/group_name/context_scope/source`；手工创建和编辑 API 接口已可通过 `POST/PUT /api-endpoints` 管理，接口 `code` 使用固定算法生成稳定 GUID（避免重复注册导致标识漂移）。
- 清理旧能力残留：删除 `api_endpoint.rebuild` 默认权限种子与“同步与重建”文案；全仓已无 `api_registry.rebuild`、`/rebuild`、`platform,team`/`team,platform` 残留。
- 前端 API 管理页升级为可管理模型：新增“新增 API/编辑 API”弹窗，支持多权限键、分组、团队上下文策略（`required|forbidden|optional`）、来源配置；列表同步展示多权限键、分组、上下文、来源。
- 前端上下文三态补齐：`context_type` 相关类型、筛选与展示增加 `common`，包括功能权限弹窗、功能权限列表、功能包能力/菜单弹窗及 API 类型声明；功能包 context 兜底逻辑新增 `common.*` 识别。
- 已验证：`go test ./...`（backend）通过；`pnpm exec vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 若要彻底下线 `/my-team/roles/:roleId*` 全链路（含角色 CRUD），需先补齐并切换到新的团队角色 CRUD 路由后再删除旧路由，避免页面“编辑/删除角色”断链。
- 建议补一条专门的破坏性迁移任务，对历史库执行一次性清理与重建（保留 `users/tenants/roles/menus` 主体），并在迁移后做团队管理员实测回归：团队角色功能包、菜单裁剪、权限裁剪与菜单可见性链路。

## 2026-03-24 API 管理分类化与元数据收口

### 本次改动
- API 元信息从旧 `group_code/group_name` 模式切到“分类表 + category_id”模式：新增 `api_endpoint_categories`，支持固定种子 ID、分类编码、中文名、英文名；迁移会自动初始化默认分类并把接口按模块回填分类。
- API 注册同步元数据已改为走分类编码派生，不再写入分组字段；迁移执行后会直接删除 `api_endpoints.group_code/group_name` 旧列，避免继续混用。
- API 管理页已改为 Method 独立首列、路径单独展示、来源中文化、功能归属仅系统/业务、分类可新建（中英文字段）、团队上下文可在列表内直接修改。
- 已验证 `go test ./...`、`pnpm exec vue-tsc --noEmit`、`go run ./cmd/migrate/main.go` 通过。

### 下次方向
- 如果你要进一步做“分类管理”，下一步建议把分类编辑/停用入口直接放进 API 管理页列表，而不是只在弹窗里新建。
- 当前团队上下文在列表内是直接选择器；如果你要更强的视觉提示，可以下一轮改成“标签 + 点击切换”的交互，而不用再动后端接口。

## 2026-03-24 API 自动注册收口与部署 Builder 骨架

### 本次改动
- API 自动注册规则已收口为“仅同步带元数据的路由”；未声明元数据的普通 API 不再自动写入 `api_endpoints`，但如果管理员已手工创建同 `method+path` 的 API 单元，后续同步仍会保留并更新其路由侧信息。
- API 管理后端新增“未注册路由”能力：`GET /api/v1/api-endpoints/unregistered` 会从运行中路由中筛出尚未进入 API 单元的接口，并返回方法、路径、处理器、模块以及是否带元数据，供后续后台手动补录使用。
- 新增 `internal/pkg/permissionseed` 部署初始化骨架：提供固定 UUID 规则、默认 API 分类 Seeds、默认功能键 Seeds、默认功能包 Seeds、默认角色绑定 Seeds，以及部署摘要 Builder，用于后续把初始化链路从迁移主文件中收口出来。
- 权限设计文档已同步补充：明确“带元数据 API 才自动注册”“普通 API 可在后台手动创建 API 单元并绑定路由”“API 管理页必须支持搜索未注册 API”“部署链路采用独立初始化 Builder”。
- 已验证 `backend/go test ./...` 通过。

### 下次方向
- 下一步应把 `permissionseed.DeploymentBuilder` 真正接入 `cmd/migrate`，完成默认功能键、默认功能包、默认角色绑定的统一导入，不再继续把种子硬编码散落在 `cmd/migrate/main.go`。
- 需要继续收口 `permission_actions -> permission_keys` 与 `feature_package_actions -> feature_package_keys`，让“功能键”正式成为主模型，避免旧 action 语义继续扩散。
- 前端 API 管理页下一轮应补上“未注册 API 搜索/选择后创建 API 单元”的交互闭环，并增加可观测提示，区分“自动注册 API”“手动补录 API”“未注册路由”三类状态。

## 2026-03-24 部署 Builder 接入迁移入口

### 本次改动
- 已将 `permissionseed` 新骨架接入 [main.go](/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go) 的初始化入口：默认 API 分类、核心功能键、核心功能包和默认角色绑定现在开始通过统一 Seeds 索引参与初始化，不再完全依赖散落的硬编码创建逻辑。
- `syncAPIRegistry` 现在会基于 `permissionseed.DeploymentBuilder` 输出初始化摘要，补充默认分类数、默认功能键数、默认功能包数、默认角色绑定数、带元数据路由数、未注册路由数，作为部署链路的基础可观测信息。
- 核心默认数据已开始固定化：命中新 Seeds 的功能键和功能包在首次创建时会写入稳定 UUID，后续为正式收口 `permission_keys` / `permission_packages` 做准备。
- 已执行 `backend/go test ./...` 通过；`go run ./cmd/migrate/main.go` 在当前 Windows 环境启动临时可执行文件时被系统拒绝，报错 `Access is denied`，因此本轮未完成命令级迁移验证。

### 下次方向
- 下一步应继续把剩余默认种子从 `cmd/migrate/main.go` 迁入 `permissionseed`，尤其是功能包菜单绑定、功能包功能键绑定和默认角色包绑定逻辑，减少迁移主文件体积。
- 需要补前端“未注册 API”页面闭环，并把自动注册/手动补录/未注册三类状态明确展示出来。
- 等当前环境的 `go run` 执行限制排除后，建议补一次完整迁移回归，确认初始化摘要、默认 UUID 落库和 API 自动注册结果都符合预期。

## 2026-03-24 API 管理页未注册补录闭环

### 本次改动
- 前端 API 管理页新增“未注册 API”入口与弹窗，可直接查询运行时尚未进入 `api_endpoints` 的路由，并支持按 `method/path/module/keyword/only_no_meta` 过滤。
- 未注册路由列表已支持“一键创建 API”，会将路由与元数据自动带入新增 API 表单，减少管理员手工补录时的重复输入。
- `frontend/src/api/system-manage.ts` 与 `frontend/src/types/api/api.d.ts` 已补齐未注册路由列表的请求封装与类型声明；权限设计文档同步追加了补录交互规则。
- 已执行 `frontend/pnpm exec vue-tsc --noEmit` 通过；本轮未执行前端构建与服务联调。

### 下次方向
- 继续补 API 管理页的状态区分与提示，把“自动注册 API / 手工补录 API / 未注册路由”三类状态明确展示出来。
- 可继续补“从未注册路由直接创建分类映射”的轻量能力，但应坚持简单稳定，不把复杂运营逻辑塞进首版弹窗。
- 等下一轮联调时，建议配合真实路由同步和手工补录流程一起回归，确认保存后未注册列表、API 主列表和权限绑定结果一致。

## 2026-03-24 API 管理页状态概览补齐

### 本次改动
- API 管理页新增顶部状态概览，直接展示自动注册、手工补录、初始种子、未注册路由四类数量，提升部署初始化后的可观测性。
- “未注册 API”按钮已同步显示未注册数量，`同步 API` 与保存 API 后会自动刷新概览和未注册统计，减少页面状态滞后。
- 权限设计文档同步补充了 API 管理页状态概览要求，明确这四类数量应作为后台默认观察面板的一部分。
- 已执行 `frontend/pnpm exec vue-tsc --noEmit` 通过；本轮未执行前端构建与浏览器联调。

### 下次方向
- 下一步可继续把 API 主列表做成更明确的状态视图，例如列表级标签或筛选器，区分自动注册、手工补录、初始种子。
- 若后续补前端联调，建议一起验证同步 API、手工补录 API、未注册列表刷新三条路径是否完全一致。
- 后端初始化链路仍需继续收口到 `permissionseed`，尤其是默认功能包与功能键绑定关系，避免页面状态与初始化种子长期脱节。

## 2026-03-24 API 主列表注册方式筛选补齐

### 本次改动
- API 管理页主列表新增“注册方式”快速筛选，可直接切换查看自动注册、手工补录、初始种子三类 API，不再只依赖顶部统计卡片判断来源分布。
- 路径列新增行内标签，直接显示 API 来源与功能归属，让列表明细与顶部概览形成呼应，减少进入编辑弹窗前的信息缺失。
- 权限设计文档同步补充：API 主列表应支持按注册方式快速筛选，作为管理页标准交互的一部分。
- 已执行 `frontend/pnpm exec vue-tsc --noEmit` 通过；本轮未执行前端构建与浏览器联调。

### 下次方向
- 下一步可以继续补主列表的关键字搜索和组合筛选，把模块、来源、路径关键字一起收口到更完整的查询条。
- 若要继续提升可观测性，可以考虑在列表层再增加“是否带权限键”“是否带分类”的轻量筛选，但应保持界面不过载。
- 后端初始化种子仍需继续迁移到 `permissionseed`，前端状态视图只是观察层，不能长期替代初始化链路本身的收口。

## 2026-03-24 permissionseed 全量收口与 API 管理页统一查询

### 本次改动
- `backend/internal/pkg/permissionseed` 已扩展为默认主数据唯一目录，统一承载默认 API 分类、默认功能键、默认功能包、功能包菜单绑定、功能包功能键绑定、组合包包含关系和默认角色功能包绑定。
- `backend/cmd/migrate/main.go` 已改为直接消费 `permissionseed`，移除本地平行默认清单；默认功能键、功能包、组合包、角色绑定和 API 分类初始化都开始走单一 Seeds 来源。
- API 管理页补齐统一查询条，支持按 `method/path/category/contextScope/status` 组合筛选，并继续与模块切片、注册方式切片联动，形成完整查询入口。
- 部署摘要新增默认组合包关系数量；已执行 `backend/go test ./...` 与 `frontend/pnpm exec vue-tsc --noEmit`，通过。

### 下次方向
- 下一步建议继续把 `permission_actions -> permission_keys` 与 `feature_package_actions -> feature_package_keys` 的正式模型收口推进下去，让默认 Seeds 与正式表结构语义一致。
- 若要继续增强 API 管理页，可补“是否有权限键 / 是否有分类”的轻量筛选和空态提示，但应坚持简单稳定，不把页面做成重运营后台。
- 当前仍未执行真实迁移命令与浏览器联调；下一轮建议配合本地数据库跑一次完整 `migrate` 和页面操作回归，确认初始化落库与筛选交互都符合预期。
## 2026-03-24 功能权限旧列彻底摘除与 API 元数据收口

### 本次改动
- 将 `permission_actions` 的正式主链收口为 `permission_key + module_code + module_group_id + feature_group_id + context_type`，后端服务、仓储、鉴权和前端权限适配都不再依赖 `resource_code / action_code` 历史列。
- 迁移程序已实际执行并完成数据库清理，`permission_actions.source`、`permission_actions.resource_code`、`permission_actions.action_code` 三个历史列已物理删除，同时保留默认功能键、功能包和角色绑定数据不变。
- `internal/pkg/apiregistry` 的 `RouteMeta` 与链式 Builder 已同步去掉旧 `resource/action` 元数据入口，API 注册正式只保留 `code/module/summary/category/source/context_scope/permission_keys` 等字段。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过，并校验数据库结果 `permission_groups=13`、`permission_actions=18`、`feature_package_actions=11`、`role_feature_packages(enabled)=2`。

### 下次方向
- 继续把 `apiregistry` 周边调用和页面说明统一到“权限键 + 分组 + 分类”模型，避免后续再次引入旧 `resource/action` 心智。
- 继续把 `cmd/migrate/main.go` 中剩余零散初始化逻辑迁入 `permissionseed`，让部署入口进一步接近单一 Builder/Seeds 结构。
- 如需继续做物理收口，可下一轮推进 `permission_actions -> permission_keys`、`feature_package_actions -> feature_package_keys` 的正式命名迁移。
## 2026-03-24 permissionseed 初始化执行层收口

### 本次改动
- 在 [permissionseed](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/internal/pkg/permissionseed) 新增执行型 Ensure 能力，将默认 API 分类、默认权限分组、默认功能键、默认功能包、默认组合包关系、默认角色功能包绑定统一收口到同一包内，不再只保留静态 Seeds 常量。
- [main.go](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/migrate/main.go) 已改为直接调用 `permissionseed` 的执行函数，迁移入口继续瘦身，职责收口为顺序编排、日志输出和刷新链路串联。
- 为部署摘要新增 `WithCoreDefaults()` 入口，统一装载默认菜单、分类、功能键、功能包、组合包和角色绑定，减少后续初始化场景的重复拼装。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过；已重新构建并执行真实迁移二进制，数据库校验结果为 `permission_groups=13`、`permission_actions=18`、`feature_packages=6`、`feature_package_bundles=3`、`feature_package_actions=11`、`role_feature_packages(enabled)=2`。

### 下次方向
- 继续把 `cmd/migrate/main.go` 中剩余非必要初始化细节迁入 `permissionseed`，尤其是可继续下沉的菜单/角色相关默认化逻辑。
- 推进正式命名收口：`permission_actions -> permission_keys`、`feature_package_actions -> feature_package_keys`，让表名和当前主模型一致。
- 若接下来转向别的主线，当前初始化层已经足够稳定，可优先避免再往迁移入口追加新的种子写库实现。

## 2026-03-24 删除联动、快照补偿与初始化覆盖修复

### 本次改动
- 修复了平台角色、团队角色和团队删除时的关联清理链路，补齐用户角色绑定、角色功能包、角色菜单隐藏、角色功能禁用、角色数据权限、团队/平台快照等从属数据的删除，避免主体删除后残留脏权限。
- 快照读取逻辑已改为“缺失即重算并持久化”，覆盖平台角色快照、团队功能快照、团队菜单快照和团队角色快照，避免首次读取或异常缺失时返回假空权限。
- 初始化 Seeds 不再对内置功能包执行“先删后重建”覆盖，改为仅补齐缺失的默认动作、菜单和组合关系；同时 API 保存时开始校验 `permission_key` 是否真实存在，平台角色绑定功能包时增加上下文校验。
- 前端同步补了 API 管理页全量统计、功能包 `common` 上下文范围选择以及权限工作台按功能分组筛选；已验证 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过，并重新执行真实迁移二进制，数据库校验结果仍为 `permission_groups=13`、`permission_actions=18`、`feature_packages=6`、`feature_package_bundles=3`、`feature_package_actions=11`、`role_feature_packages(enabled)=2`。

### 下次方向
- 继续把固定 ID 策略从默认 Seeds 推到正式表命名收口，优先推进 `permission_actions -> permission_keys` 与 `feature_package_actions -> feature_package_keys`。
- 若后续要进一步增强删除一致性，可补数据库外键和 `ON DELETE CASCADE` 到纯从属表，但业务刷新与跨表补偿仍建议保留在服务层事务里。
- API 管理和权限页面的主链已经稳定，下一轮可以把重心切到你准备转向的新主线，不必再回头做结构级清理。

## 2026-03-24 权限主表正式收口到 permission_keys

### 本次改动
- 将数据库正式主表名从 `permission_actions / feature_package_actions` 收口为 `permission_keys / feature_package_keys`，并在迁移入口增加了 `AutoMigrate` 前重命名补偿，确保老库升级不会先建新空表。
- 后端 ORM、仓储、鉴权、索引维护和命名迁移 SQL 已同步改到新表名；当前 Go 结构名仍保留 `PermissionAction / FeaturePackageAction` 作为兼容层，避免这一轮把服务层和接口语义同时放大。
- 已验证 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过，并重新执行真实迁移二进制；数据库校验结果为仅保留 `permission_keys`、`feature_package_keys` 两张正式表，数据量分别为 `18`、`11`。

### 下次方向
- 下一步可以继续把代码语义层也正式收口，把结构名、Repo 名、Service 名从 `Action` 改成 `Key`，彻底结束兼容过渡态。
- 如果要补数据库约束，建议围绕新表名继续加唯一索引、外键和纯从属级联，不要再往旧表名追加任何新逻辑。
- 当前数据库命名已经稳定，接下来可以安全转向新的主线，不必再回头处理旧 `permission_actions` 命名问题。

## 2026-03-24 权限键代码语义命名收口

### 本次改动
- 将后端内部主链的结构名、仓储名和服务方法名继续从 `Action` 语义收口到 `Key` 语义，重点覆盖 `PermissionKey`、`FeaturePackageKey`、角色权限边界、角色权限设置 DTO 与角色权限服务方法。
- 保持外部兼容面稳定：API 路径仍保留 `/actions`，JSON 字段仍保留 `action_id / action_ids`，避免这一轮同时破坏前端和外部调用。
- 同步修正文档口径，明确系统内部正式模型已经完成 `Key` 化；已执行 `backend/go test ./...` 与 `frontend/pnpm exec vue-tsc --noEmit`，通过。

### 下次方向
- 若后续要继续彻底清理旧心智，可再处理少量兼容残留，例如日志文案、局部变量名和 `RoleDisabledActionRepository` 这类基于历史表语义保留的实现名。
- 如果未来允许接口破坏式升级，再考虑把 `/actions` 路径和 `action_ids` 字段统一切到 `keys` 命名，并配套前端与接口文档一起升级。
- 当前内部主链已经足够稳定，后续可以优先转向新的业务主线，不必再反复回头处理权限命名。 

## 2026-03-24 API 管理页改为左树右表

### 本次改动
- 将 API 管理页重构为左右布局：左侧新增分类树用于筛选全部 API、未分类 API 和具体分类，右侧保留 API 列表、查询条件、注册方式筛选和表格操作。
- 原页面内嵌的分类管理区已移除，统一改为 `Drawer` 抽屉配置；抽屉内同时提供分类列表、启停、编辑和新建表单，避免主页面信息堆叠。
- 同步补了分类树筛选状态持久化、分类计数统计和移动端自适应样式；已验证 `frontend/npm run build` 通过，未单独做浏览器手工点测。

### 下次方向
- 下一步建议在浏览器里补一次真实交互验证，重点看分类树选中态、Drawer 编辑后列表刷新，以及移动端折叠后的可用性。
- 如果后端后续支持真正的多级分类结构，可以直接把当前左侧树从单层统计扩展成真实树形，不需要再推翻这次布局。

## 2026-03-24 默认权限分组固定化

### 本次改动
- 将默认权限分组从“按已有权限键动态推导模块分组”改为“显式固定 Seed 清单”，新增固定模块分组定义并统一使用稳定 UUID，确保部署迁移时可以一键导入。
- 默认内置权限键的功能分组默认值统一收口到 `system`，并继续按模块分组固定绑定到 `role`、`user`、`api_endpoint`、`feature_package`、`tenant_member_admin`、`system_permission` 等显式模块分组。
- 已重新执行真实迁移并校验数据库，`permission_groups` 现为 `15` 条，模块分组名称已更新为中文固定名，内置 `permission_keys` 已绑定固定 `module_group_id / feature_group_id`；同时执行了 `backend/go test ./...` 与 `frontend/pnpm exec vue-tsc --noEmit`，通过。

### 下次方向
- 如果后续要给新业务模块预埋权限主数据，直接往 `permissionseed` 的固定模块分组和默认功能键清单里追加，不要再回到运行时动态推导。
- 前端权限键管理页下一步可以直接展示“系统固定模块分组”与“管理员自建分组”的差异，例如增加 `is_builtin` 只读提示。
- 当前默认分组已经稳定，接下来可以把重心转到新的业务方向，不需要再为迁移一致性回头补分组 ID 规则。

## 2026-03-24 permission_action 正式迁移到 permission_key

### 本次改动
- 补充了新的幂等命名迁移 `20260324_permission_key_code_alignment_v2`，在迁移中先确保 `permission_key` 对应的默认 API 分类、默认模块分组和默认功能键存在，再统一把旧 `permission_action` 引用迁移到 `permission_key`。
- 迁移会同步更新 `api_endpoints.category_id`、`api_endpoints.module`、`permission_keys.module_code`、`permission_keys.module_group_id`，并物理清理旧的 `permission_action` API 分类和模块分组，避免数据库长期保留双份正式编码。
- 文档规则已同步收口到 `permission_key`：默认模块分组、API 分类和功能键模块编码全部以 `permission_key` 为正式值；前端仍保留少量旧值兼容解析，但数据库正式状态已完成切换。
- 已验证：`backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过；重新构建并执行真实迁移二进制后，数据库核验结果为 `permission_groups=15`、`permission_keys=18`、`feature_packages=5`、`feature_package_keys=11`、`role_feature_packages(enabled)=2`，且 `permission_action` 分类与模块分组已清理完成。

### 下次方向
- 如果后续要继续减兼容层，可以逐步移除前端与映射层对 `permission_action` 的兜底解析，只保留 `permission_key` 正式编码。
- 历史文档中若仍出现 `permission_actions` 或 `permission_action` 旧表述，可继续按“历史兼容说明”与“正式模型说明”两层再做一次全文清理。
- 当前数据库主数据和迁移链已稳定，后续新增系统功能请直接在 `permissionseed` 中补固定分组、固定功能键和固定分类，不要再引入动态推导。

## 2026-03-24 API 管理页模块字段改为分类主语义

### 本次改动
- API endpoint 对外接口已去掉 `module` 查询、表单和返回字段，主列表、未注册路由列表、权限关联接口弹窗统一改为按分类展示与筛选。
- 前端类型与请求封装已同步收口到 `category`，API 管理页新增/编辑、未注册补录、分类树统计都不再依赖模块字段，避免页面继续混用两套语义。
- 后端保留底层 `api_endpoints.module` 存量字段用于兼容已有同步链路，但更新 API 时若前端未传该字段会保留历史值，不会被错误覆盖；已验证 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过。

### 下次方向
- 若后续确认底层同步与自动注册也不再需要 `module`，可继续评估 `apiregistry`、模型字段和数据库迁移的彻底清理方案。
- 浏览器侧建议补一次 API 管理页与权限关联弹窗的手工联调，重点确认分类筛选、未注册补录和保存后回显一致。

## 2026-03-24 API endpoint 模块字段彻底删除

### 本次改动
- 继续把 `api_endpoints.module` 从主链中彻底移除：后端模型、API 管理接口、权限关联接口返回、前端类型和页面展示都不再依赖该字段，统一按分类语义工作。
- `internal/pkg/apiregistry` 已去掉 API 注册阶段的模块元数据与模块推导，自动注册现在只使用显式分类编码；未声明分类的路由不再按路径自动推模块。
- 迁移入口已补“先用旧 module 回填 category_id，再删除 module 列”的收口逻辑，并移除命名迁移里对 `api_endpoints.module` 的兼容更新；已验证 `backend/go test ./...`、`frontend/pnpm exec vue-tsc --noEmit` 通过。

### 下次方向
- 下一步建议执行一次真实迁移并检查现有库，确认 `api_endpoints.module` 已被成功删除，且历史 API 的 `category_id` 回填完整。
- 浏览器侧建议补一次 API 管理页、未注册 API 补录和权限关联接口弹窗联调，重点确认分类显示与保存回显一致。

## 2026-03-24 system/action-permission 分组管理弹窗升级

### 任务概述
- 为 `/system/action-permission` 的模块分组与功能分组提供统一弹窗管理能力，支持查看全部、新增、编辑、删除。

### 本次改动
- 将权限分组弹窗升级为“列表 + 表单”一体化管理：左侧展示当前分组类型的全量列表，右侧进行新增/编辑。
- 在分组弹窗中新增删除能力，并对内置分组禁用删除操作，避免误删系统基础分组。
- 新增前端分组删除 API 封装，并将页面入口按钮从“新建分组”调整为“管理分组”，与新交互语义一致。
- 影响范围：`system/action-permission` 页面及其分组弹窗、`system-manage` 分组接口封装。
- 已验证：执行 `pnpm -C frontend run build` 成功（包含 `vue-tsc --noEmit`）。

### 下次方向
- 建议补充分组删除前的“被引用数量”提示（如有权限项绑定时给出拦截/确认策略），降低误操作风险。
- 建议补充分组管理弹窗的自动化 UI 测试（新增/编辑/删除/内置分组限制场景）。
- 未验证项：未在真实后端数据环境逐条手工验证删除分组时的后端约束提示与联动反馈。

## 2026-03-24 系统页面搜索组件统一与默认收起调整

### 任务概述
- 将菜单管理、用户管理相关页面搜索统一为独立搜索组件模式，并将功能权限页搜索默认状态调整为收起。

### 本次改动
- 新增菜单搜索组件 `menu-search.vue`，将菜单管理页的内联搜索栏改为模块化搜索组件接入。
- 用户管理页继续使用既有 `UserSearch` 组件，保持搜索组件化一致性。
- 功能权限搜索组件 `action-permission-search.vue` 的 `defaultExpanded` 由 `true` 改为 `false`，默认收起。
- 影响范围：`system/menu`、`system/action-permission` 的搜索交互与组件组织。
- 已验证：执行 `pnpm -C frontend exec vue-tsc --noEmit` 通过。

### 下次方向
- 建议统一 `ArtTableHeader` 的 `showSearchBar` 开关状态持久化（按页面记忆上次展开状态）。
- 建议补充菜单页和功能权限页的搜索重置/查询联动用例，防止后续重构回归。
- 未验证项：未执行完整 `vite build`（仓库当前存在其它页面模板语法问题，非本次改动引入）。

## 2026-03-24 全局角色新增 custom_params(JSONB) 与编辑页自定义参数

### 任务概述
- 为全局角色增加参数配置能力，后端使用 `jsonb` 存储，并在系统角色编辑页新增“自定义参数”维护入口。

### 本次改动
- 后端模型 `roles` 增加 `custom_params` 字段（`jsonb`，默认 `{}`），并新增命名迁移 `20260324_role_custom_params_jsonb`。
- 角色创建/更新 DTO 增加 `custom_params`，服务层在创建与更新时写入该字段。
- 角色列表/详情接口返回 `customParams`，前端可直接回填编辑。
- 前端角色类型定义补充 `customParams/custom_params`，角色编辑弹窗新增“自定义参数(JSON)”输入框，支持格式化回填与 JSON 对象校验后提交。
- 影响范围：后端迁移、角色模型、角色 CRUD 接口、前端角色编辑页面。
- 已验证：`pnpm -C frontend exec vue-tsc --noEmit` 通过；`go test ./internal/modules/system/role ./internal/api/dto ./internal/modules/system/models ./cmd/migrate` 通过。

### 下次方向
- 建议为 `custom_params` 约束一版业务 schema（键白名单/类型），避免后续被任意结构污染。
- 建议在角色列表页增加“参数摘要”列（可选），方便运营快速识别关键配置。
- 待确认：是否需要对团队角色（tenant 角色）同样开放该参数字段与编辑能力。

## 2026-03-25 功能权限关联接口新增/删除与迁移补齐

### 任务概述
- 为功能权限键补充“关联接口”的新增与删除能力，并确保新增接口在老环境通过迁移自动归属 API 管理并绑定到既有权限键。

### 本次改动
- 后端新增接口：`POST /api/v1/permission-actions/:id/endpoints`（新增关联）与 `DELETE /api/v1/permission-actions/:id/endpoints/:endpointId`（删除关联），均使用 `system.permission.manage` 权限保护。
- 新增请求 DTO：`PermissionKeyEndpointBindRequest`，服务层新增 `AddEndpoint/RemoveEndpoint`，并增加 `ErrAPIEndpointNotFound` 错误语义。
- 扩展 `APIEndpointPermissionBindingRepository`：新增按 `permission_key + endpoint_id` 的增删方法，避免重复绑定并支持精准移除。
- 快照刷新：关联接口变更后，按功能权限键反查受影响功能包并触发 `permissionrefresh.RefreshByPackages`，确保团队/平台权限快照同步。
- 新增命名迁移 `20260325_permission_endpoint_binding_ops`：
  - 自动补齐两条新接口到 `api_endpoints`；
  - 强制归属 `permission_key` 分类（API 管理归属）；
  - 自动写入 `api_endpoint_permission_bindings`，绑定 `system.permission.manage`。
- 前端“关联接口”弹窗增加：
  - 可选接口下拉 + “新增关联接口”按钮；
  - 列表“移除”操作；
  - 调用新增后端接口并联动刷新可选项与当前列表。
- 影响范围：权限键关联接口管理、API 注册表归属与绑定迁移、权限快照刷新链路、功能权限前端操作体验。
- 已验证：
  - `go test ./internal/modules/system/permission ./internal/modules/system/user ./cmd/migrate` 通过；
  - `pnpm -C frontend build` 通过。

### 下次方向
- 建议在“新增关联接口”中加入服务端分页/远程搜索，避免接口数量很大时一次性加载过多数据。
- 建议补充后端用例：覆盖重复绑定幂等、无效 endpointId、迁移重复执行幂等。
- 建议在 API 管理页增加“按权限键反查已绑定接口”入口，便于运维审计。

## 2026-03-25 API 同步与旧库 schema 兼容修复

### 任务概述
- 修复 `POST /api/v1/api-endpoints/sync` 在旧库结构下报错 `api_endpoints.module NOT NULL` 导致 5001 的问题，统一迁移顺序与运行时兼容行为。

### 本次改动
- 调整迁移主流程顺序：先执行 `finalizeAPIEndpointSchema`，再执行 `syncAPIRegistry`，避免旧 `module` 列约束在同步阶段触发插入失败。
- 在 `apiregistry.SyncRoutes` 增加旧结构兼容预处理：若检测到 `api_endpoints.module` 列存在，自动执行：
  - `ALTER COLUMN module DROP NOT NULL`
  - `ALTER COLUMN module SET DEFAULT ''`
- 新增 `hasColumn` 检测函数，用于运行时按表结构安全分支处理。
- 影响范围：API 同步链路、迁移执行顺序、旧库升级兼容。
- 已验证：`go test ./internal/pkg/apiregistry ./internal/modules/system/apiendpoint ./cmd/migrate` 通过。

### 下次方向
- 建议补充集成测试：覆盖“旧库带 module NOT NULL”场景下的 `SyncRoutes` 回归。
- 建议在 API 同步失败响应中附带简化错误摘要（保留安全边界），降低排障成本。
- 待确认：是否在后续版本彻底删除旧 `module` 列相关兼容代码（以迁移完成率为准）。

## 2026-03-25 移除 API 同步旧库兼容逻辑

### 任务概述
- 在确认环境已完成迁移后，移除 `apiregistry.SyncRoutes` 中针对旧 `api_endpoints.module` 列的运行时兼容分支。

### 本次改动
- 删除 `syncRoutesInternal` 入口处的旧库预处理调用。
- 删除旧库兼容函数 `prepareLegacyAPIEndpointSchema` 与 `hasColumn`。
- 保留并依赖迁移链路中的 schema 统一（`finalizeAPIEndpointSchema`）作为唯一升级路径。
- 影响范围：API 同步流程从“兼容旧库+新库”收敛为“仅支持已迁移库”。
- 已验证：`go test ./internal/pkg/apiregistry ./cmd/migrate` 通过。

### 下次方向
- 建议在部署文档中明确“必须先执行迁移再启用同步 API”顺序，避免新环境漏跑迁移。
- 建议增加启动期 schema 健康检查（检测 `api_endpoints.module` 是否仍存在并给出明确告警）。

## 2026-03-25 迁移前置修复（api_endpoints.module 旧约束）

### 任务概述
- 处理旧库在命名迁移阶段写入 `api_endpoints` 时触发 `module NOT NULL` 约束报错，导致迁移中断的问题。

### 本次改动
- 在 `runNamedMigrations` 之前新增前置步骤 `ensureAPIEndpointLegacyModuleNullable`。
- 若检测到旧列 `api_endpoints.module` 存在，则先执行：
  - `ALTER COLUMN module DROP NOT NULL`
  - `ALTER COLUMN module SET DEFAULT ''`
- 确保命名迁移插入 API 注册记录时不再被旧约束阻断，并保持后续 `finalizeAPIEndpointSchema` 统一清理旧列。
- 已验证：`go run ./cmd/migrate` 在当前环境执行成功，最终输出 `Migration completed successfully`。

### 下次方向
- 建议把 `cmd/migrate` 的 SQL 调试日志级别下调，避免迁移日志过长影响故障定位效率。
- 建议补一条自动化校验：迁移后若仍存在 `api_endpoints.module/group_code/group_name` 则直接告警失败。

## 2026-03-25 清理与文档统一（第一批）

### 任务概述
- 按“行为不变”原则执行仓库清理与文档重整。
- 前后端同步清理低风险冗余代码。
- 建立前端高频交互统一规范，并接入 AGENTS 约束链路。

### 本次改动
- 前端清理：
  - `system/user` 页面移除无用导入 `ArtButtonTable`、移除调试 `console.log`、简化操作列渲染包装。
  - `system/menu` 备份列表“操作”由并排按钮改为 `ArtButtonMore` 三点菜单，统一操作列交互风格。
- 后端清理：
  - `menu/handler.go` 将重复匿名授权接口收敛为单一 `menuAuthzService` 类型，减少重复定义。
- 文档重整：
  - 新增 `FRONTEND_GUIDELINE.md`，固化搜索区、操作列、弹窗工具栏、状态/辅助文案规范。
  - `AGENTS.md` 增加前端规范引用条款。
  - 重写 `PROJECT_FRAMEWORK.md` 为可执行清单版本（触发条件、输入/设计/实施检查、完成标准、常见反例）。
  - 新增 `docs/README.md` 作为 docs 索引。
  - 删除重复文档 `docs/permission-package-design.md`，并在 `permission-overall-summary.md` 标注收敛说明。

### 下次方向
- 继续执行“第二批”业务功能分离交付：把当前权限链路功能改动与清理改动拆分提交。
- 在 `views/system` 范围继续收敛搜索组件字段布局与交互一致性（保持业务字段差异）。
- 对剩余候选“无用代码”执行引用级扫描后再删，避免动态场景误删。

### 验证
- 已验证：
  - `go test ./...`（backend）通过。
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`（本次按需求未执行）。

## 2026-03-25 持续优化（前端统一 + 迁移收口）

### 任务概述
- 继续按“行为不变”原则推进前端统一与后端内部收敛。
- 前端聚焦系统管理页的无用代码与操作列统一。
- 后端聚焦迁移职责边界，避免 `AutoMigrate` 混入历史数据修复。

### 本次改动
- 前端：
  - `system/menu` 页面移除无用的 `computed` 导入。
  - `system/role` 权限选择弹窗移除未使用的 `hiddenMenuIdSet`。
  - `action-permission/permission-group-dialog` 的列表操作列统一改为 `ArtButtonMore` 三点菜单。
  - `FRONTEND_GUIDELINE.md` 补充规则级别、标准片段和反例，提升可执行性。
- 后端：
  - `database.AutoMigrate()` 不再执行历史 `user_roles` 数据修复，只保留 schema/index 职责。
  - `cmd/migrate` 新增 `20260325_legacy_user_roles_backfill` 命名迁移，承接旧 `user_roles` 修复。
  - `20260324_permission_key_code_alignment_v2` 去掉重复 `EnsureDefault*` 初始化。
  - `syncCanonicalPermissionKeys` 不再重复回写 `context_type`，降低与 `permission_context_backfill` 的职责重叠。
  - 迁移日志文案从旧 `permission actions` 收口为 `permission keys`，并统一使用 logger 收尾。
- 文档：
  - `docs/README.md` 补充 `FRONTEND_GUIDELINE.md`、`AGENTS.md` 入口和阅读路径。
  - `PROJECT_FRAMEWORK.md` 补充最低验证命令矩阵与未验证项记录格式。

### 下次方向
- 继续清理 `docs/change-log.md` 中对已删除文档的历史坏链引用。
- 继续收敛系统页搜索组件触发契约与默认展开策略，但不与本轮低风险清理混提。
- 若确认线上已完成兼容窗口，可规划下线 `legacy ...string` 权限兼容入参链。

### 验证
- 已验证：
  - `go test ./...`（backend）通过。
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 持续优化（搜索契约与组件日志清理）

### 任务概述
- 继续按“行为不变”原则清理系统页交互分叉与基础组件调试噪音。
- 收口搜索组件事件契约，减少父子双通道传参。

### 本次改动
- 前端：
  - `system/role` 与 `system/user` 的搜索组件统一改为 `emit('search')`，父层直接从 `v-model` 读取当前筛选值。
  - `system/menu/modules/menu-search.vue` 收口到可展开搜索区配置，与当前前端规范一致。
  - 移除基础组件中的明显调试输出：
    - `art-wang-editor` 去掉全屏 `console.log`
    - `art-notification` 去掉“查看全部”占位日志
    - `art-video-player` 去掉播放/暂停日志
    - `art-cutter-img` 去掉下载日志
- 文档：
  - `change-log.md` 顶部补充历史说明，明确 `permission-package-design.md` 已下线并由 `permission-overall-summary.md` 接替。

### 下次方向
- 继续清理 `change-log.md` 中直接出现旧设计文档名的历史条目，但以批量文案收口方式处理。
- 继续收敛系统页搜索区默认展开策略，避免不同页面继续各自漂移。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理链路收敛（访问路径与分组/页面拆分）

### 任务概述
- 修复页面管理中“访问”地址解析仍会偏离实际运行时路由的问题。
- 简化页面管理配置流，将页面分组与页面配置拆开，减少无关字段干扰。

### 本次改动
- 前端：
  - 新增统一页面路径解析工具，管理页访问、列表展示、运行时路由注册三处共用同一套路径推导规则。
  - 页面列表补充菜单路径映射，路由列改为显示最终访问路径，避免只看原始 `routePath` 无法判断真实访问地址。
  - `page-dialog` 拆为分组表单与页面表单：
    - 分组只保留名称、标识、模块、排序、挂载方式、状态。
    - 页面表单按“基本信息 + 挂载方式 + 访问模式 + 高级配置”收口，并增加最终路径预览。
  - 页面挂载关系收敛为两种主流程：
    - 挂到菜单
    - 挂到页面/分组
  - 选上级页面后不再要求重复选择上级菜单，默认沿父页面链继承菜单路径。

### 下次方向
- 补一轮页面管理实际交互冒烟，重点看：
  - 挂菜单页面
  - 挂分组页面
  - 全局页
  - 内嵌页
- 若需要进一步收口，可继续把“高级配置”里的面包屑模式做成更强约束的策略选项，减少人工自由输入。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。
  - 页面管理 UI 实际点击冒烟，原因：本轮未启动前端页面联调。

## 2026-03-25 页面管理表单说明增强（字段帮助与默认回退）

### 任务概述
- 页面管理表单继续收敛配置成本。
- 支持路由名称留空时自动回退到页面标识，并为主要字段补充就地说明。

### 本次改动
- 前端：
  - 新增统一字段说明组件，为页面表单和分组表单的配置项补充问号提示说明。
  - `路由名称` 改为可留空，提交保存时默认回退为 `页面标识`，减少重复填写。
  - 保留现有示例区与字段提示，使“示例 + 字段说明 + 最终路径预览”形成完整配置引导。

### 下次方向
- 可继续把说明文案收敛到配置元数据，避免页面表单和分组表单各自维护文本。
- 若后续继续简化，可把低频字段移动到更明确的“高级配置”分区。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理复制页面能力

### 任务概述
- 在页面管理操作列中新增“复制页面”能力。
- 复制时复用新增弹窗，但按被复制对象预填内容，并显示独立标题。

### 本次改动
- 前端：
  - 页面管理操作菜单新增“复制页面”。
  - 点击后打开新增链路弹窗，标题显示为“复制页面”。
  - 复制模式按原页面预填表单内容，但不带原记录 `id`，并为名称、页面标识、路由名称生成副本默认值。
  - 复制提交走创建接口，不会误覆盖原页面。

### 下次方向
- 可继续补“复制分组”或“复制后自动聚焦唯一字段”的交互优化。
- 若你希望更稳，可再加一层复制时对 `routePath` 的冲突提醒。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理移除继承权限开关

### 本次改动
- 前端：移除页面配置中的“继承权限”开关，页面访问控制只保留 `public / jwt / permission / inherit` 四种访问模式。
- 运行时：`ManagedPageProcessor` 不再读取 `inheritPermission`，当页面选择 `accessMode=inherit` 时，始终沿上级页面、分组链或菜单继续继承权限约束。
- 兼容收口：未注册页面创建候选不再注入 `inheritPermission`，避免继续写入这类冗余配置。

### 下次方向
- 可继续清理后端 `inherit_permission` 的保存和返回字段，把前后端模型彻底统一成“只看 accessMode”。
- 建议补一轮页面继承分组、分组继承菜单的实际冒烟，确认旧数据里存在 `inherit_permission=false` 时也能按新规则正常继承。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面分组支持不挂载与继承节点配置

### 本次改动
- 前端：页面分组新增 `不挂载` 方式，并补充 `基础路径`、`访问模式`、`权限键` 配置。分组虽然不注册运行时页面，但可作为下级页面和下级分组的路径继承节点、权限继承节点。
- 前端：页面管理列表的分组路由展示改为显示解析后的继承路径；当分组配置了基础路径时，列表可直接看到其最终继承前缀。
- 后端：页面保存规则放开“非全局页必须绑定父级”的限制，允许分组和页面在不挂载菜单/上级页面的情况下独立存在；分组现有的 `route_path / access_mode / permission_key` 配置会保留并可继续参与下级继承。

### 下次方向
- 建议补一轮实际联调：`分组(不挂载)+基础路径+权限键 -> 子页面(继承)`、`分组(挂菜单)+基础路径 -> 子页面(相对路径)` 两条链路，确认访问地址与权限校验完全一致。
- 可继续优化页面管理文案，把“基础路径”在列表中再区分为“继承前缀”，减少和真正可访问页面路径的语义混淆。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
  - `go test ./internal/modules/system/page/...`（backend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面列表精简与父分组停用链路修复

### 本次改动
- 前端：页面管理列表名称列去掉 `pageKey` 的同行展示，列表回到更简洁的一层信息展示。
- 前端：`ManagedPageProcessor` 新增父页面链有效性校验。页面若通过 `parentPageKey` 绑定到上级页面或分组，而其父链节点在运行时注册表中缺失（例如上级分组被停用），则该页面不再继续注册为可访问路由。
- 修复结果：避免出现“父分组已停用，但下级页面因回退成独立路径而仍可访问”的错误行为。

### 下次方向
- 可继续在页面管理列表中提示“父链无效/父分组停用”状态，便于管理员在配置页直接看出为什么某个页面不会被运行时注册。
- 建议补一轮联调：停用上级分组后刷新页面，确认子页面路由不再注入、访问直接落到 404 或权限兜底页。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 公开页面未登录访问修复

### 本次改动
- 前端：`ManagedPageProcessor` 现在区分 `public` 与 `jwt`。公开页面允许未登录注册，`jwt` 页面在未登录状态下不会再被误放行。
- 前端：路由守卫新增“未登录先注册公开运行时页面”的链路。访问 `accessMode=public` 的页面时，会先拉取公开运行时页面并完成动态注册，再放行访问；非公开页面仍按原规则跳转登录。
- 前端：登录后会用完整的菜单与运行时页面路由覆盖公开预注册路由，避免公开态与登录态的动态路由长期混用。

### 下次方向
- 建议实际联调两条路径：`public` 页面未登录直达、`jwt` 页面未登录访问应跳登录，确认行为完全分离。
- 如果后续需要让“挂到菜单的公开页面”也能未登录继承菜单路径，可再补 runtime 接口返回父菜单完整路径，避免前端在无菜单上下文时只能依赖显式路径。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 公开页面未登录直达链路补强

### 本次改动
- 后端：`/api/v1/pages/runtime` 现在会为运行时页面补齐可解析的 `active_menu_path`，即使页面挂在菜单或分组下，前端在无菜单上下文时也能推导出完整访问路径。
- 前后端联动后，公开页面未登录预注册时不再依赖前端已有菜单树，`accessMode=public` 的页面可以按完整路径完成动态注册。
- 修复范围覆盖“公开页挂菜单/挂分组”的未登录访问场景，不再因为基础路径解析失败而回退到登录页。

### 下次方向
- 建议继续实测：公开页面使用相对路径挂到菜单、挂到分组、挂到不挂载分组三种场景，确认未登录均可直达。
- 若后续希望进一步减少前端推导，可继续考虑在 runtime 接口里直接返回最终解析后的完整路由路径。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
  - `go test ./internal/modules/system/page/...`（backend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 公开页面未登录访问最终根因修复

### 本次改动
- 后端：将 `/api/v1/pages/runtime` 从 JWT 保护路由组中拆出，改为真正的公开接口。此前虽然页面模块里将其声明为普通 GET，但由于整个页面模块挂在认证路由组下，未登录访问 runtime 实际仍会返回 401，导致前端公开页面注册链提前失败并跳转登录。
- 后端：页面模块新增 `RegisterPublicRoutes`，只负责注册 runtime；其余页面管理接口仍保留在受保护路由组内。
- 前后端配合后，未登录访问 `accessMode=public` 页面时，前端能够成功获取 runtime 页面注册表并完成公开页动态注册，不再被迫跳转到登录页。

### 下次方向
- 建议直接在浏览器里复测具体地址 `/#/2/example-page`，并确认 Network 中 `/api/v1/pages/runtime` 返回 200 而不是 401。
- 若后续担心公开 runtime 接口泄露过多内部页面元数据，可继续做一层“匿名请求仅返回公开可达页面”的过滤逻辑。

### 验证
- 已验证：
  - `go test ./internal/modules/system/page/...`（backend）通过。
  - `go test ./internal/api/router/...`（backend）通过。
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 公开页面继承访问模式修复

### 本次改动
- 前端：`ManagedPageProcessor` 现在会把页面继承解析后的有效访问模式写入运行时路由 `meta.accessMode`。此前子页面原始配置为 `inherit` 时，即使它通过上级分组或菜单实际继承到了 `public`，守卫仍会把它识别成非公开页并跳转登录。
- 前端：公开运行时路由若已按旧状态注册，但当前目标页仍未被识别为公开页，守卫会自动卸载并重建公开运行时路由，避免浏览器会话里残留的旧动态路由元信息继续生效。
- 修复后，像“页面自身 `access_mode=inherit`，上级分组 `access_mode=public`”这类链路，未登录访问时会按有效访问模式 `public` 处理。

### 下次方向
- 建议继续用你当前这条数据复测：子页面 `inherit`，父分组 `public`，直接访问 `/#/2/example-page`，确认不再跳登录。
- 若后续还要进一步稳固，可考虑让 runtime 接口直接返回“effective_access_mode”，前端不再自行二次推导。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理拆分逻辑分组与普通分组

### 本次改动
- 页面管理新增 `display_group` 与 `display_group_key`：保留现有 `group` 作为“逻辑分组”，新增“普通分组”只负责页面管理列表归类，不再参与路径、权限、菜单高亮和面包屑继承。
- 后端 `ui_pages` 模型、保存校验、删除校验、列表装饰和运行时筛选已同步调整；普通分组会落库并可被页面/逻辑分组引用，但会被排除在 runtime 页面注册表之外，避免误进运行时路由链。
- 前端页面管理已拆出“新增逻辑分组”和“新增普通分组”两套入口；页面、逻辑分组弹窗新增普通分组选择，列表树改为“逻辑树 + 普通分组归类”双层结构，现有逻辑分组文案也统一改名。

### 下次方向
- 建议补一轮页面管理实际联调：重点检查“页面挂菜单 + 归属普通分组”“逻辑分组挂菜单 + 归属普通分组”“普通分组删除时引用拦截”这三条链路。
- 若后续还要继续收紧模型，可以再把普通分组的状态语义单独化，例如停用普通分组时自动让页面回到未分组根节点展示。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
  - `go test ./internal/modules/system/page/... ./internal/api/router/...`（backend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理弹窗排版收口

### 本次改动
- 页面、逻辑分组、普通分组三类弹窗统一改为“说明卡片 + 分区表单”结构，示例说明默认折叠，基础信息、路由与渲染、挂载与归属、访问与行为分区更清晰，减少原先一整屏平铺字段带来的阅读负担。
- 页面弹窗将最终路径改成独立预览框，并把低频项收进“高级配置”折叠区；普通分组弹窗同步补齐示例折叠和滚动容器，整体视觉与页面/逻辑分组保持同一套样式语言。
- 这次调整只改前端排版和交互层次，不改页面管理的保存字段与业务逻辑，重点是让配置流程更清晰、更简约。

### 下次方向
- 建议继续实际过一轮页面管理主流程：新增页面、复制页面、新增逻辑分组、新增普通分组，确认不同弹窗在中等分辨率下的纵向空间和滚动体验都稳定。
- 若后续还要继续收紧界面，可以再把字段帮助文案抽成统一配置元数据，减少各弹窗内部的静态文本重复。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理弹窗切换为右侧抽屉

### 本次改动
- 页面、逻辑分组、普通分组这三类配置弹窗统一从 `Dialog` 切换为 `Drawer`，交互改为从右侧滑出，减少列表页被整块模态遮住后的割裂感，配置过程也更接近“侧边编辑”。
- 抽屉宽度沿用原先三套配置表单的有效宽度，并同步调整 `body/footer` 的滚动与底部按钮样式，保证长表单在抽屉模式下仍可稳定滚动，底部操作区固定且清晰。
- 这次不改表单字段、校验和保存逻辑，只替换承载容器和对应样式层，避免把视觉交互调整和业务行为调整混在一起。

### 下次方向
- 建议直接联调一次新增页面、编辑逻辑分组、复制普通分组三条链路，重点确认抽屉关闭、再次打开、表单滚动位置和底部按钮区域在实际浏览器中的手感。
- 如果后续还要继续压缩操作成本，可以再把抽屉标题区补成“主标题 + 副说明”的固定头部样式，进一步统一系统管理侧边编辑体验。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面挂载后最终路径预览不联动修复

### 本次改动
- 修复 `Cascader` 选值在特定场景下可能返回数组导致菜单 ID 解析失败的问题：路径解析工具和页面/逻辑分组抽屉都增加了菜单值归一化逻辑，统一取最终节点 ID 参与路径计算与保存。
- 调整路径解析规则：`/单段路径` 在挂载菜单或上级页面时不再被强制当成绝对路径，而是按相对段参与拼接，保证选择挂载目标后“最终路径”能够即时变化。
- 本次改动覆盖页面管理的“路径预览 + 提交保存”两条链路，避免预览和实际落库行为不一致。

### 下次方向
- 建议联调两条典型路径：`/example-page`（单段前置斜杠）和 `/report/detail`（多段绝对路径），确认在“挂菜单/挂页面”两种模式下都符合预期。
- 若后续继续强化可用性，可在路由路径输入框旁补充“当前按相对还是绝对解析”的即时提示，降低配置歧义。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 编辑页面挂载后最终路径不变化修复

### 本次改动
- 继续修复编辑链路：页面与逻辑分组弹窗在“编辑态”下计算最终路径时，统一对菜单选择值做归一化（兼容 `Cascader` 字符串值和数组值），并在提交时同样写入归一化后的菜单 ID，避免“预览不变 / 保存错误”。
- 路径解析工具进一步对前导斜杠的单段路径做兼容处理，挂载菜单或上级页面后可按相对段参与拼接，确保编辑场景与新增场景行为一致。
- 修复影响范围覆盖“最终路径预览 + 编辑提交”双链路，重点解决你反馈的“新增正常、编辑异常”。

### 下次方向
- 建议针对已有历史页面做一轮编辑回归：分别测试挂菜单、挂页面、切换挂载方式三种操作，确认最终路径实时联动。
- 若后续还要增强容错，可在打开编辑时对历史 `route_path` 做一次规则提示，区分“绝对路径（固定）”与“相对路径（会继承）”。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 菜单路径映射补齐父级链路

### 本次改动
- 修复页面管理中菜单路径映射只记录当前节点 `path` 的问题。此前像“系统管理 / 角色管理”这类菜单，前端拿到的是子节点自身的 `role`，没有把父级 `/system` 拼进去，导致最终路径预览和列表显示都可能退化成 `/role`。
- 页面编辑弹窗、逻辑分组弹窗、页面管理列表三处菜单路径映射已统一改为递归拼装完整菜单链路，使用父级路径 + 当前节点路径生成完整路径，再参与最终路径计算和列表展示。
- 这次修复直接对应你截图里的场景：选择二级菜单后，最终路径不再只用叶子节点路径。

### 下次方向
- 建议再实测一轮多级菜单：二级菜单、三级菜单分别挂载页面，确认最终路径和页面列表显示都已变成完整链路。
- 若后续想把这层逻辑前移，也可以考虑让后端菜单选项接口直接返回 `fullPath`，前端就不再重复递归组装。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 未注册页面注册时锁定组件路径

### 本次改动
- 未注册页面从候选列表进入“创建页面”时，会在默认数据里带上内部标记，前端据此将组件路径视为固定扫描结果，不再允许在注册弹窗里修改组件入口。
- 路由路径仍保持可编辑，并且继续支持随挂载菜单或上级页面变化进行自动相对化处理，避免组件入口和访问路径两个概念被混在一起。
- 这次调整只影响“未注册页面 -> 创建页面”链路，不改变普通页面新增、编辑、复制时组件路径的可编辑行为。

### 下次方向
- 建议实际回归一条未注册页面：确认组件路径为只读、路由路径仍可改，并在切换挂载菜单后最终路径能继续联动。
- 若后续还想继续压缩误操作，可把未注册来源的页面标识、路由名称也增加轻量提示，提醒它们来自扫描默认值。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理抽屉提示收敛

### 本次改动
- 页面、逻辑分组、普通分组三类抽屉移除了和问号帮助重复的字段下方说明，重点保留动态价值更高的路径预览提示，避免整屏出现重复文案。
- 未注册页面来源改为在页面抽屉顶部使用轻量标签提示“未注册来源，组件路径固定”，替代原先组件字段下方的大段说明文字。
- 这次调整只收敛抽屉提示层级，不改变字段含义、校验规则和保存逻辑。

### 下次方向
- 建议你顺手再看一轮页面抽屉、逻辑分组抽屉、普通分组抽屉，确认信息密度是否已合适；如果还要继续压缩，可以把示例区默认改成更轻的折叠按钮样式。
- 如果后续发现某个动态提示仍然不够直观，可以只针对那个状态补充，不再回到字段下方堆文本的方式。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 页面管理列表改为数据优先视图

### 本次改动
- 页面管理列表头部改成“标题 + 统计 + 收口操作”结构，新增页面/逻辑分组/普通分组被收进统一新增入口，配合同步、未注册、展开分组等操作，减少原先整排按钮对数据区的干扰。
- 表格列重新分级：类型、归属、排序和高级配置合并进名称主列，访问模式和状态合并为同一列，只保留最终路径、组件入口、更新时间和操作等高价值信息，让用户先看到核心数据，再看辅助属性。
- 这次调整聚焦“简单、商务、一眼看清数据”，不改后端接口和页面管理数据模型。

### 下次方向
- 建议你实际看一轮真实数据量更大的场景，重点确认第一列的信息密度和统计块是否已经足够清晰；如果还需要更商务化，可以继续把状态和访问列做成更统一的芯片样式。
- 若后续想进一步减少横向滚动，可以继续考虑把组件入口也降为悬浮展示或次级展开信息。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-25 菜单体系移除内页类型

### 本次改动
- 菜单管理前端移除了“内页”相关入口与筛选：不再展示“显示内页”开关，不再允许在菜单弹窗中选择或模板化创建内页，菜单弹窗只保留“菜单入口”配置。
- 菜单列表中的类型判定改为仅区分“目录/菜单”，并统一由菜单表单写入 `meta.isInnerPage = false`，避免再产生新的菜单内页数据。
- 后端菜单服务增加统一兜底：`GetTree` 过滤历史 `isInnerPage=true` 菜单；`Create/Update/RestoreBackup` 对 `meta` 做标准化，强制 `isInnerPage=false`，从接口层彻底收口菜单内页。

### 下次方向
- 建议增加一次数据库迁移，将历史 `menus.meta.isInnerPage=true` 数据迁移到 `ui_pages` 或做归档清理，避免数据库长期残留旧模型记录。
- 建议在菜单备份恢复后追加一次“旧内页记录统计”日志，便于运维确认是否仍存在待迁移数据。

### 验证
- 已验证：
  - `pnpm -C frontend exec vue-tsc --noEmit`（frontend）通过。
  - `go test ./internal/modules/system/menu/...`（backend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-26 菜单内页残留代码清理

### 本次改动
- 菜单管理链路进一步去除“内页”残留语义：清理了菜单域中的内页兼容字段与筛选逻辑，菜单配置只保留入口菜单能力。
- 角色权限配置弹窗（菜单权限页签）移除了“显示内页”开关及相关筛选/数据结构字段，菜单权限树只按隐藏、内嵌、启用状态筛选。
- 后端菜单服务删除了菜单内页专用过滤与元数据清洗辅助函数，菜单服务恢复为纯菜单树处理逻辑，不再携带菜单内页兼容分支。

### 下次方向
- 建议补一次手工回归：菜单管理、角色菜单权限、团队菜单边界、用户菜单裁剪四个页面各点一轮，确认筛选与保存行为一致。
- 若后续要做“菜单元数据白名单”，可在后端 `MenuCreate/Update` 层再加统一校验，避免非菜单字段被写入 `meta`。

### 验证
- 已验证：
  - `pnpm -C frontend exec vue-tsc --noEmit`（frontend）通过。
  - `go test ./internal/modules/system/menu/...`（backend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-26 菜单管理分组独立表全量部署

### 本次改动
- 菜单管理分组从临时 `meta.manageGroup` 升级为独立数据模型：新增 `menu_manage_groups` 表、`menus.manage_group_id` 关联字段、唯一索引与排序索引，并将其纳入 `AutoMigrate`。
- 后端菜单模块补齐了分组 CRUD、菜单读写关联、备份/恢复携带分组数据，以及 `20260326_menu_manage_groups_backfill` 命名迁移，用于把旧 `meta.manageGroup` 回填到新表并清理旧字段。
- 前端菜单管理页接入了独立分组链路：菜单抽屉改为选择正式分组、菜单列表按分组生成仅管理页可见的虚拟分组层展示，并新增“菜单分组”抽屉做新增、编辑、删除与排序管理。
- 同步修正了路由类型定义，让运行时路由可稳定承载 UUID 菜单/页面数据，避免前端在 `id` 类型上继续被旧的数字约束卡住。

### 下次方向
- 建议补一轮实际页面回归：菜单管理新增/编辑菜单、分组排序、删除分组占用校验、菜单备份恢复分组数据四条链路各走一遍。
- 如果后续要继续做菜单管理体验优化，可以把分组行的展开/收起状态做本地持久化，避免管理页刷新后每次重新展开。
- 目前迁移已兼容旧 `meta.manageGroup` 回填；若确认线上不存在旧写入来源，后续可以删掉服务层和迁移中的兼容清洗说明，只保留新模型。

### 验证
- 已验证：
  - `pnpm exec vue-tsc --noEmit`（frontend）通过。
  - `go test ./...`（backend）通过。
- 未验证：
  - 前端构建 `vite build`，原因：本轮按既定约束未执行。

## 2026-03-26 配置型大弹窗统一改为抽屉

### 本次改动
- 将系统管理、团队管理中的配置型大弹窗统一替换为右侧抽屉，覆盖功能权限、功能包、角色、团队角色、用户、团队等主配置入口；保留未注册列表、预览、备份等轻量浏览型弹窗继续使用 Dialog。
- 补齐了 API 管理页中的内联配置容器，将“新增/编辑 API”和“加入/移除权限键”也统一改为抽屉，避免同类页面交互形态继续混用。
- 更新 FRONTEND_GUIDELINE.md，明确“大表单、绑定关系、树选择、批量配置统一使用 Drawer；轻量确认/预览/列表型弹层保留 Dialog”的前端规范。
- 已验证 pnpm exec vue-tsc --noEmit（frontend）通过；本轮未执行前端构建 ite build。

### 下次方向
- 建议再补一轮抽屉统一样式收口，例如标题区、底部按钮区、滚动区高度和说明文案层级，避免不同模块虽然都改成抽屉但视觉密度不一致。
- 可继续盘点 rontend/src/views/system/api-endpoint/index.vue、rontend/src/views/system/menu/index.vue 这类页面内联弹层，判断哪些轻量交互未来也应拆成独立抽屉组件，减少首页文件继续膨胀。

## 2026-03-26 用户管理新增权限测试诊断工具

### 本次改动
- 在用户管理模块新增两个后端接口：`GET /api/v1/users/:id/permission-diagnosis` 和 `POST /api/v1/users/:id/permission-refresh`，用于查看平台/团队上下文权限快照、测试单个权限键、并手动刷新快照。
- 诊断结果接入了现有平台快照与团队边界快照，返回当前快照刷新时间、权限键测试结果、命中快照情况、来源功能包，以及团队上下文下各角色的命中/禁用链路。
- 前端用户管理页操作列新增“权限测试”，使用独立抽屉承载上下文选择、权限键测试、快照摘要、来源功能包与角色链路展示，避免把复杂判断压到运行时鉴权链路里。
- 已验证 `go test ./...`（backend）与 `pnpm exec vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 可继续补“只展示该用户所在团队”的上下文选择，而不是当前全团队列表，提高诊断效率。
- 若后续要让模块分组/功能分组停用真正影响运行时权限，应继续把鉴权与快照刷新链路统一到“有效状态”口径，再让本诊断抽屉同步展示这层差异。
- 本轮未执行浏览器联调和前端构建；下一步建议实际点开用户管理页验证平台/团队两种上下文与刷新按钮交互。

## 2026-03-26 API 停用接入运行时鉴权缓存

### 本次改动
- 新增 `backend/internal/pkg/apiendpointaccess/service.go`，维护 `method + path -> API 状态` 的运行时内存缓存，并提供统一中间件用于拦截已停用 API。
- 在 `backend/internal/api/router/router.go` 中将该中间件挂到 `/api/v1` 与 `/open/v1` 分组，登录、刷新令牌、仅 JWT 接口、权限接口都会统一受 API 状态控制；接口已停用时直接返回“当前 API 已停用”。
- 在 API 管理模块中接入缓存刷新：`创建/更新 API` 与 `同步 API 注册表` 成功后立即刷新运行时缓存，避免为接口状态校验去读权限快照或在每次请求时查数据库。
- 已验证 `go test ./...`（backend）通过。

### 下次方向
- 可继续补一层 API 状态诊断信息，在权限测试或 API 管理页中展示“运行时缓存状态 / 最近刷新时间”，便于排查接口已落库但缓存未更新的异常场景。
- 若后续要支持多实例部署，需要把当前单实例内存缓存升级为广播刷新或短周期失效重载，否则不同实例之间的 API 状态生效时机会不一致。

## 2026-03-26 用户权限测试补充超级管理员直通说明与菜单测试

### 本次改动
- 修正用户权限测试的诊断语义：当用户为超级管理员时，即使权限键未命中当前快照，也会因超级管理员直通而返回通过；后端现在会明确返回该状态，前端会展示“直通放行 / 超级管理员”与对应说明，避免把“未命中但通过”误判成快照异常。
- 在用户权限测试抽屉中新增“菜单测试”面板，按当前平台或团队上下文加载该用户最终可见菜单，并使用三级级联只读面板展示，支持搜索菜单标题或路由，以及显示隐藏、内嵌、启用、路径等过滤开关。
- 前端补充用户权限诊断与菜单树的类型归一化，统一处理后端返回的诊断字段和菜单树节点结构。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 可继续在菜单测试里增加“命中来源”展示，例如菜单来自公共页面、平台快照还是团队角色快照，方便定位菜单可见性的真正来源。
- 若后续还要扩展权限测试，可把功能权限与菜单测试统一成 Tab 结构，并增加复制结果、导出诊断快照、按角色钻取等操作，减少单屏信息堆叠。

## 2026-03-26 后端查询优化首轮收口

### 本次改动
- 将 API 管理列表的权限绑定查询从逐条 `ListBindings` 改为按 `endpointIDs` 一次批量加载，消除接口列表页的 N+1 查询。
- 将 API 按权限键筛选、权限关联接口列表改为数据库侧 `id IN` 过滤，不再先拉全量接口再在内存裁剪；同时给 `apiEndpointRepo` 和 `tenantRepo` 补了 `GetByIDs` 以支撑批量校验。
- 将功能包服务中的组合包子包校验、团队绑定校验、团队功能包读取改为批量读取；将团队边界快照中的包展开、动作推导、菜单推导、角色继承判断改为批量预取，减少快照刷新时的递归逐条 SQL。
- 已验证 `go test ./...`（backend）通过。

### 下次方向
- 还可以继续收 `teamboundary` 里“全表取 bundle 再展开”的实现，改成按根包闭包递归或一次性 CTE，避免功能包规模继续增大后出现全表扫描。
- 可继续补基准日志或 SQL 统计，对团队快照刷新、角色快照刷新、API 列表这三条热点链路做优化前后对比，避免后续修改把查询数量又带回去。

## 2026-03-26 权限测试补充拒绝层级与平台团队作用域说明

### 本次改动
- 在用户权限测试后端诊断中新增结构化字段，明确返回 `拒绝层级`、`拒绝原因`、`成员状态`、`边界链路状态`、`角色链路状态`，用于区分“角色命中但最终被成员或边界挡住”的场景。
- 在权限测试抽屉中补充团队上下文链路展示，将原先模糊的“权限未通过”拆成可定位的诊断信息：团队成员校验、团队边界校验、角色链路是否命中或禁用。
- 在用户功能包与用户菜单裁剪抽屉中补充平台/团队作用域说明，明确平台功能包和平台菜单裁剪只影响平台上下文，不直接决定团队内权限与菜单。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 可继续把角色链路与权限测试结果联动高亮，例如当“角色链路命中但团队成员失效”时，直接在角色表格上方给出冲突提示，减少人工比对。
- 如果后续还要做团队权限排查，建议再加入“当前团队成员记录”和“命中的团队功能包列表”两个块，让权限来源、边界来源和成员状态在同一屏内闭环。

## 2026-03-26 权限测试只展示用户所在团队并补齐团队链路

### 本次改动
- 新增 `GET /api/v1/users/:id/teams`，团队上下文下拉不再读取全量团队，而是只展示该用户实际所在的团队。
- 在用户权限诊断响应中新增 `team_member` 与 `team_packages`，前端权限测试页新增“当前团队成员记录”和“当前团队功能包”，把成员、边界、角色、功能包四段链路串起来。
- 前端用户权限测试抽屉改为使用用户专属团队列表，并保留单团队自动选中逻辑，减少错误选择无关团队造成的误判。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 可以继续把“团队功能包”区分为“团队已开通”和“本次权限命中的来源功能包”，避免团队开通很多包时难以看出真正来源。
- 若后续还要做运维排障，建议在权限测试里再补“当前团队成员更新时间 / 来源身份角色”，把成员来源也显式展示出来。

## 2026-03-26 运行时菜单树按默认值瘦身

### 本次改动
- 将普通 `/menus/tree` 返回改为运行时瘦身结构：顶层只保留必要字段，`meta` 中只有 `true`、非空字符串、非空数组字段才返回，默认值不再重复传输。
- 保留 `all=1` 的完整菜单树返回不变，确保菜单管理、角色菜单配置、功能包绑定菜单等管理侧功能继续拿到完整字段。
- 前端菜单运行时解析统一改为“显式为真才生效”，即缺失字段与 `false` 同义，避免后端为了兼容默认值继续传大量冗余布尔字段。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 可继续把运行时菜单树里的 `icon`、`roles` 做成更细的按需输出，例如仅在存在角色限制时返回。
- 若菜单量继续增加，建议再补版本号或缓存键，让前端只在菜单配置变更后重新拉取树结构。

## 2026-03-26 运行时页面注册表缓存与公开页轻量加载

### 本次改动
- 在 `backend/internal/modules/system/page` 增加进程内运行时页面缓存，缓存全量运行时页面与公开运行时页面两份结果，默认 TTL 为 24 小时；普通 `/pages/runtime` 与新增 `/pages/runtime/public` 现在优先走内存，不再每次重复查 `ui_pages`、菜单和页面关联数据。
- 页面创建、更新、删除，以及菜单创建、更新、删除、备份恢复时都会主动失效这份缓存，保证页面挂载菜单或父链变化后可以实时同步新的运行时结果。
- 前端公开页守卫改为请求 `/api/v1/pages/runtime/public`，未登录访问公开页时不再拉取整份运行时页面注册表，只加载公开页面及其必要父链。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 当前这轮缓存是共享运行时注册表缓存，没有按用户拆分；这是刻意保留的，因为原始页面注册表本身与用户无关，若后续要继续减小登录时的计算量，更合适的是对“用户已展开的运行时路由结果”做前端或网关侧按用户缓存。
- 若页面规模继续增长，可继续补 `/pages/runtime/match` 按路径懒匹配接口，让已登录和未登录场景都支持按需补注册，而不是在首次进入时加载完整页面集合。

## 2026-03-26 运行时页面接口继续瘦身

### 本次改动
- 将 `/api/v1/pages/runtime` 与 `/api/v1/pages/runtime/public` 的返回改为运行时轻量结构，只保留页面注册和路径解析真正需要的字段，去掉来源、模块、展示名、时间戳、软删除等管理态字段。
- 对布尔与默认值字段采用“显式为真才返回”的口径：`keep_alive`、`is_full_page` 仅在为 `true` 时输出，`breadcrumb_mode` 与 `access_mode` 仅在偏离默认值时输出，`meta` 仅保留 `isIframe`、`isHideTab`、`link` 等运行时需要的非默认值。
- 前端运行时页面归一化继续兼容缺省字段，公开页和已登录路由注册链路无需额外改造。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 若还要继续压缩数据量，可把 `route_name`、`id` 这类仅用于调试或重复检测的字段再做一次必要性梳理，按运行时真实依赖决定是否继续保留。
- 下一步建议补 `/pages/runtime/match`，把“接口轻量化”和“按路径懒匹配”接起来，进一步减少首次进入时的传输量与前端构建成本。

## 2026-03-26 运行时页面读取阶段拍平分组链

### 本次改动
- 调整运行时页面接口读取逻辑：`/api/v1/pages/runtime` 与 `/api/v1/pages/runtime/public` 现在只返回实际可注册页面，不再把 `group` 分组节点直接吐给前端。
- 读取阶段会先把分组父链拍平，再输出页面的最终 `route_path`、最近的非分组 `parent_page_key`，并在需要时把分组上的访问模式覆盖到子页面，避免前端为了路径和权限继承继续依赖分组节点。
- 这样公开页和运行时页面注册表都可以只消费最终页面集合，前端不再需要拿到“分组 + 页面”混合列表来自己拼路径。
- 已验证 `go test ./...`（backend）与 `pnpm -C frontend exec -- vue-tsc --noEmit`（frontend）通过。

### 下次方向
- 还可以继续把运行时 DTO 再压一轮，例如评估 `route_name` 是否只在与 `page_key` 不同且确有命名跳转依赖时才保留。
- 下一步建议补 `/pages/runtime/match`，让运行时接口不仅拍平分组链，还支持按访问路径懒匹配，进一步减少首次加载量。
