# Change Log

# 2026-04-02 frontend-fluentV2 全量页面收口范围表与壳层状态统一

### 本次改动
- 补充第 8 版全量页面收口的范围文档 [frontend-fluentV2/docs/page-inventory.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/docs/page-inventory.md)，把 `frontend/src/views/**/*.vue`、docker PostgreSQL 里的 `menus` / `ui_pages` 与当前 React 承接页统一到一份清单里，后续页面迁移以此为准。
- 将运行时导航错误态与壳层初始化错误态从 `MessageBar` 收口为 `PageStatusBanner`，让页面级状态提示继续跟 Fluent 2 Web / Teams 的低噪风格保持一致。
- 外链页空 URL 的提示也切到统一页面状态条，减少公共页与业务页之间的视觉分裂。

### 下次方向
- 后续新增或拆分页面时，先回到 page-inventory 清单补范围，再做 React 页面或模块实现，避免只盯单一域。
- 继续把重页里的模块向 `features/<domain>` 下沉，确保第 8 版是“全量页面 + 全量模块”收口，而不是只完成几个热点域。

# 2026-04-02 仓库规范补充：数据库临时迁移与默认种子

### 本次改动
- 补充仓库级协作约定：当代码修改需要调整数据库时，可以先增加一个临时迁移并执行验证；如果迁移已经完成目标、当前数据库也已稳定，就删除这份临时迁移，避免把一次性修复脚本长期保留在迁移链里。
- 同步明确默认初始化数据应优先通过 seed / ensure 逻辑完善，迁移只承担一次性结构变更或历史数据修正，不把长期默认状态反复写进迁移步骤。

### 下次方向
- 后续凡是涉及数据库改动的任务，先判断是结构变更还是默认初始化补齐；前者可用临时迁移，后者优先补 seed。
- 如果某个临时迁移已经执行成功并且不再需要回放，就在代码库里删除它，避免后续启动重复触发历史修正。

# 2026-04-02 frontend-fluentV2 菜单重复节点接口核查与前端去重

### 本次改动
- 通过真实登录后的运行时导航接口核查到重复菜单来源：后端 `menu_tree` 里确实返回了两条同名 `PageManagement` 节点，一条路径为 `/system/page`，另一条为相对路径 `page`，两者在前端解析后最终指向同一菜单入口。
- 前端在 `buildNavigationItems` 中增加了同级保守去重，按 `path + label` 只保留一条，避免接口里的重复菜单直接渲染成两个“页面治理”。
- 菜单图标继续沿“接口没给就不渲染”的策略，不再回到默认图标兜底；本轮仅收口重复节点，不改后端菜单契约。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 若后端后续要修数据源，应从菜单表里消除同名同路径的重复记录，而不是继续依赖前端兜底。
- 继续用真实菜单数据回归其它分支，确认前端这层保守去重不会误伤合法的同名不同层级菜单。

# 2026-04-02 frontend-fluentV2 菜单图标仅在后端提供时渲染

### 本次改动
- 调整运行时菜单图标策略：如果后端菜单接口没有提供 `icon`，前端不再自行推断或补默认图标，而是直接不渲染图标，避免不同菜单因为前端兜底规则过粗而看起来一样。
- 将导航类型里的 `icon` 改为可选字段，并同步修正菜单搜索和侧栏渲染，保证“没图标就不画”在搜索结果、桌面展开树、收缩态级联浮层和移动端抽屉里都一致生效。
- 继续保留菜单的无限层级、自动扩宽、长标题截断与悬浮提示等已有行为，只收口图标表现，不回退结构层逻辑。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 回归真实菜单数据，确认后端哪些节点本来就不提供图标后，前端确实完全不渲染，而不是又被别的兜底逻辑重新补出来。
- 若后续还要增强图标表现，只能优先让后端补明确图标字段，再由前端按字段渲染，不再做模糊推断。

## 2026-04-02 frontend-fluentV2 第七版长标题宽度收窄

### 本次改动
- 进一步收紧长标题对菜单宽度的影响：侧栏宽度估算系数降低、三级及以下层级的额外宽度减小，收缩态浮层最大宽度也继续下调，避免一个很长的菜单标题把整列菜单撑得过宽。
- 菜单标题在侧栏和收缩浮层里继续保持单行截断逻辑，优先保住整列宽度和层级结构，不再为了显示完整长标题让菜单块整体变宽。
- 三级及以下节点的缩进和引导线仍然保持收浅，配合更小的字体和块高，整体更偏轻量微软风格。
- 仍然保持后端运行时导航契约和菜单空间接口不变，仅调整前端宽度估算和视觉表现。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 用更长的真实标题继续回归，重点确认现在的宽度上限是否足够保守，以及单行截断是否仍然可读。
- 如果后面还要继续压宽度，优先再收紧一级和二级节点的文本占位，而不是再放宽容器。

## 2026-04-02 frontend-fluentV2 第七版菜单父子重复修复与三级压浅

### 本次改动
- 针对菜单里出现“父节点展开后又重复显示成子节点”的问题，前端在 `buildNavigationItems` 阶段增加了保守去重，自动剔除与父节点 `routeId`、`path` 或 `label` 完全一致的自重复子节点，避免同一个菜单项被当成父子两份渲染。
- 三级及以下菜单层级的视觉再次收浅：缩进、引导线、块高和字号都继续压低，减少深层目录一层层往里堆叠的厚重感。
- 收缩态级联浮层同步减轻阴影和字号，并把宽度上限收窄，避免长标题把浮层撑得过宽。
- 这轮仍然保持导航契约、运行时 DTO 和菜单空间接口不变，只修前端侧的树清理与视觉层级表现。
- 已通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 用真实菜单数据继续回归“父子重复”是否彻底消失，尤其关注那些后台树本身带有别名或自引用痕迹的节点。
- 再看一轮三级及以下的视觉层级，如果仍然显深，可以继续收缩引导线和子节点字号，但先不要把信息层级压没。

## 2026-04-02 frontend-fluentV2 第七版菜单视觉细化

### 本次改动
- 在无限层级和自动扩宽基础上继续收口菜单视觉：侧栏主项、子项、收缩态 rail 和级联浮层都下调了字号、块高和缩进强度，整体从“深、厚、重”调整为更轻的微软企业后台风格。
- 桌面展开态的侧栏宽度动画从原来的快切节奏调整为更柔和的过渡，避免深层菜单展开时侧栏显得生硬。
- 桌面展开树的子节点间距、边界留白和层级引导线都收浅，减少菜单块一层层往里压的深度感。
- 收缩态级联浮层增加了更轻的阴影和边框，子菜单项字体也同步缩小，避免浮层显得过厚。
- 移动端抽屉头部和树体间距也做了轻量收敛，保证统一视觉节奏。

### 下次方向
- 继续用深层菜单和长标题回归视觉细节，重点看展开/收起动画是否还可以再轻一点，以及层级引导线是否需要继续减弱。
- 如果后续仍觉得“菜单块比较深”，可以再做一轮仅针对一级/二级/三级的字号和留白分级，但这轮先停在轻量化收口。

## 2026-04-02 frontend-fluentV2 第七版菜单导航自动扩宽优化

### 本次改动
- 在无限层级导航基础上继续收口菜单宽度问题：桌面展开侧栏不再固定为 `252px`，而是会根据当前可见导航树的层级深度和节点标题长度自动计算推荐宽度，展开深层节点时会平滑增宽，避免子节点较多或缩进较深时可视空间不足。
- 桌面收缩态的级联浮层也改为按当前层级菜单项内容自适应宽度，不再固定使用单一 `240px-300px` 的窄浮层；子节点标题较长时，当前浮层会自动放宽。
- 这轮仍然保持后端运行时导航 DTO、菜单空间接口和搜索模型不变，只调整了 `AppShell` 与 `SideNav` 的渲染宽度策略。
- 已完成 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 校验。

### 下次方向
- 用更长标题、更多深层节点和至少两个真实 `menu space` 继续回归自动扩宽策略，确认宽度上限、空间隔离和运行时导航刷新后的表现都稳定。
- 如后续需要更强的可视反馈，可继续补“当前正在扩宽到哪一级”的层级提示，但这轮先不叠加额外交互。

## 2026-04-02 frontend-fluentV2 第七版菜单导航无限层级修复

### 本次改动
- 将 `SideNav` 从固定两层结构重构为递归导航树，桌面展开态现在支持任意深度内联展开；当前路由祖先链会自动展开，并与用户手动展开状态合并。
- 桌面收缩态改为递归级联浮层，一级图标可打开首层菜单，带子节点的菜单项可继续向右打开下一级，不再只支持单层子菜单。
- 移动端导航改为近全屏抽屉树，支持任意深度展开；点击叶子节点后会自动关闭抽屉并完成路由跳转。
- `useShellStore` 新增按 `menu space` 持久化的展开状态与剪枝动作，收缩/展开切换后仍保留当前空间下的手动展开记忆。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit`、`pnpm --dir frontend-fluentV2 build`，并用浏览器实测桌面展开、桌面收缩级联和移动端抽屉三种导航态。

### 下次方向
- 用至少两个真实 `menu space` 继续回归展开状态隔离，确认切空间后展开记忆不会串空间。
- 继续清理现有控制台网络噪音和 HMR 历史日志，重点区分后端接口异常与导航壳层本身的问题。

## 2026-04-02 frontend-fluentV2 第七版团队联调收口与治理页回归增强

### 本次改动
- 第七版优先解决团队域联调收口问题。后端 `my-team` 只读接口在“当前账号暂无团队上下文”时，不再直接返回 404 业务错误，而是统一改成 `200 + 空结果`，覆盖团队边界角色列表、动作来源、菜单来源、团队成员列表和成员边界角色列表，避免团队治理页在无团队场景下把网络错误暴露给前端。
- 在此基础上继续补齐“默认团队回退”语义：`/api/v1/user/info` 和 `my-team` 在没有显式 `X-Tenant-ID` 时，会回退到当前用户的默认团队成员关系，使前端能够恢复 `current_tenant_id` 并继续使用团队治理主链。
- 前端 `system/team-roles-permissions` 与 `team/team-members` 已同步改成业务空态承接：当当前账号没有团队归属时，页面会显示明确的引导区和团队入口卡，而不是停留在通用错误页；相关保存按钮也会进入清晰的禁用态。
- 消息治理页继续做第七版回归增强：`system/message-template`、`system/message-sender`、`system/message-recipient-group` 现在补齐了页面级成功/失败反馈、URL 选中恢复，以及“删除能力待后端开放”的禁用说明，不再让没有后端删除契约的能力看起来像是前端漏接。
- `system/feature-package` 继续做治理页主链收口：基础信息保存、删除、子包关系、动作关系、菜单关系、团队关系都补上了统一反馈，并在当前选中功能包失效后自动清理 URL，减少列表与右侧详情不同步。
- 运行时导航标题继续收口：左侧导航构建时优先使用 `RuntimeNavItem.title` 而不是后端 label key，进一步减少后端翻译键直接暴露到 Fluent 2 壳层。
- 进一步为 `/system`、`/team`、`/workspace`、`/message` 等目录根节点补上本地组标题映射，减少侧栏和面包屑根节点继续显示 `menus.*` 式后端 key。
- 本轮联调后已重新启动后端服务，并用真实接口确认 `/api/v1/tenants/my-team/boundary/roles`、`/api/v1/tenants/my-team/action-origins`、`/api/v1/tenants/my-team/menu-origins`、`/api/v1/tenants/my-team/members` 在无团队场景下已经全部返回 200。

### 破坏性调整
- `my-team` 只读查询接口在无团队场景下的 HTTP 语义已从 `404 + ErrNoTeam` 改为 `200 + 空结果`。旧前端如果明确依赖 404 来判断“暂无团队”，需要改为根据当前数据是否为空、或根据当前用户租户上下文判断。
- `user/info` 和 `my-team` 新增了默认团队成员关系回退。若旧逻辑假定只有显式 `X-Tenant-ID` 才会出现 `current_tenant_id`，需要接受当前账号会自动进入默认团队上下文。
- 消息模板、发送人、收件组页现在会在当前选中项从列表中消失时主动清理 URL 里的选中参数；后续若继续在这些页面上做操作回滚或缓存补丁，必须保持 URL 与当前详情同步。

### 需要人工回归验证
- 当前管理员账号虽然已经可以自动回退到默认 `my-team` 上下文，但后续继续做团队深回归时仍建议使用专门的联调团队数据，避免直接污染系统默认团队。
- `workspace/inbox`、消息调度、模板/发送人/收件组页需要继续在真实业务数据下验证保存、刷新和回执是否完全符合预期。
- `system/feature-package` 仍需继续做真实数据下的关系配置回归，确认子包、动作、菜单、团队四类关系在保存后都能稳定回刷。

## 2026-04-02 frontend-fluentV2 第六版主链回归与前后端联调收口

### 本次改动
- 在第五版结构化详情与域级懒加载基础上继续推进第六版，优先把消息域、团队域、系统治理域的高频主链补成真正可回归的闭环，而不是只停留在结构化工作台层面。
- `workspace/inbox` 已补齐已读、批量已读、待办处理三条主链的成功/失败反馈，并在列表刷新后自动清理失效选中项；详情区补上了优先级、租户上下文、待办状态和更新时间等结构化信息。
- `system/message` 调度台现在会在发送成功后自动把新记录切成当前选中记录，并展示结构化回执；消息工作台仍继续共用系统域和团队域同一套 feature，不再复制两套逻辑。
- `team/team` 补上删除团队、成员加入、成员移除、成员角色更新等主链反馈；`team/team-members` 补上成员增删改和边界角色分配反馈，并在成员被移除后自动清理 URL 中失效的当前选中项。
- `system/team-roles-permissions` 补上边界角色创建、删除、动作授权、菜单授权、功能包授权的统一反馈；`system/user` 补上用户创建、删除、角色分配、功能包授权、菜单授权和权限快照刷新的统一反馈。
- 前后端联调上继续做非破坏式收口：前端团队边界角色创建接口改为 `/api/v1/tenants/my-team/boundary/roles`，后端同时新增 POST `/api/v1/tenants/my-team/boundary/roles` 别名，兼容第六版前端主链而不破坏旧路径。

### 破坏性调整
- 团队边界角色创建链路已从前端旧的 `/api/v1/tenants/my-team/roles` 对齐为 `/api/v1/tenants/my-team/boundary/roles`；后续若继续扩团队边界角色逻辑，必须沿 `boundary/roles` 这套语义继续走，不要再把成员角色和边界角色混回同一路径。
- 多个治理页新增了统一的页面级反馈条，后续如果继续补操作反馈，必须保持“操作成功/失败 -> 页面级反馈 + query invalidate”这一模式，不要回退到只靠控制台或隐式刷新。

### 需要人工回归验证
- 当前运行中的后端实例仍对 `/api/v1/tenants/my-team/action-origins`、`/api/v1/tenants/my-team/menu-origins`、`/api/v1/tenants/my-team/boundary/roles` 返回 404；仓库源码已包含这些路由，说明联调环境需要重启或重新部署到最新后端。
- `workspace/inbox` 需要继续用真实业务数据验证：未读消息点击后自动已读、批量已读、待办完成/忽略后列表与详情是否同步收口。
- `team/team`、`team/team-members`、`system/team-roles-permissions` 需要继续用真实团队上下文回归：团队删除、成员移除、边界角色授权和来源说明是否都符合预期。
- `system/user` 需要继续验证真实用户数据下的角色分配、菜单授权、功能包授权和权限快照刷新链路。

## 2026-04-02 frontend-fluentV2 第五版业务深化、结构化详情与域级懒加载

### 本次改动
- 在第四版全量页面接入基线上继续做第五版深化，优先补消息域、团队域和系统治理域的结构化详情，不再让右侧详情区继续依赖原始 JSON 文本框或只读拼接文本框。
- 消息域完成第一波工作台深化：`workspace/inbox`、`system/message*`、`team/message*` 继续共用一套消息 feature，调度页现在能展示结构化目标实体预览和结构化发送回执；记录详情页补齐投递汇总、状态时间线、投递明细和 payload 摘要卡，收件组页也补上了匹配模式、预计人数和目标范围表格。
- 团队域完成第一波结构化治理收口：`team/team-members` 成员详情改为属性区 + 角色区 + 操作区；`system/team-roles-permissions` 改成三栏治理台，并直接接入 `/api/v1/tenants/my-team/menu-origins` 与 `/api/v1/tenants/my-team/action-origins` 的来源说明面板，统一展示动作、菜单和功能包授权来源。
- 系统治理域继续去 JSON 化：`system/access-trace` 改成结构化链路摘要与记录表；`system/feature-package` 的 impact preview 改成指标卡 + 属性区；`system/user` 的权限诊断改成结构化诊断摘要、角色链与来源包列表。
- 路由层开始第五版稳态收口：认证页、公共静态页和各业务域页面全部切到按域懒加载，并在 `vite.config.ts` 中增加 `manualChunks`，构建产物已拆成 `auth / dashboard / workspace / message / system / team` 六个主业务 chunk。
- 同步更新 `frontend-fluentV2/README.md`，将当前阶段正式改写为第五版，并把当前已完成能力、剩余风险与第六版建议收口到专题说明里。

### 破坏性调整
- `frontend-fluentV2` 的本地路由定义不再默认直接静态 import 页面组件，而是通过域级 lazy route 包装；后续新增页面时必须继续沿用 `createLazyRouteElement`，不要再把大批页面直接同步打进首包。
- `message.api.ts`、`access.api.ts`、`system.api.ts`、`team.api.ts` 的 adapter 返回值已进一步稳定化，页面侧如果继续直接假定原始 DTO 结构，会和当前稳定内部类型脱节。

### 需要人工回归验证
- 当前运行中的后端实例对 `/api/v1/tenants/my-team/action-origins`、`/api/v1/tenants/my-team/menu-origins`、`/api/v1/tenants/my-team/boundary/roles` 返回了 404；仓库源码里这些路由存在，需确认联调环境是否已重启到最新后端。
- 消息域需要继续做真实业务回归：收件箱待办处理、消息调度成功回执、模板/发送人/收件组编辑、记录明细查看。
- 团队域需要继续做真实业务回归：新增成员、修改角色、边界角色授权，以及来源说明在真实团队上下文中的联动。
- 域级拆包已经生效，但 `auth` 和 `vendor-fluent` 仍然偏大，下一轮需要继续细拆并检查是否还有可以下沉到更细 chunk 的页面依赖。

## 2026-04-02 frontend-fluentV2 第四版完善收尾与第五版起步

### 本次改动
- 对第四版已迁移页面做了一轮后端页面注册审计：确认 `system/access-trace` 已由独立命名迁移维护，不需要重复回灌到 `DefaultPages`；同时将 `system/more`、`team/more` 补入非菜单直达页 `UIPage` 种子，并新增 `20260402_message_more_page_seed` 命名迁移，保证新环境和存量环境都能同步到页面注册表。
- 前端运行时导航继续在第四版架构内做小步收口：导航项标题、页面标题与运行时面包屑现在优先使用本地 React 路由元数据，减少后端翻译键或旧标题直接暴露到 Fluent 2 壳层。
- 更新 `frontend-fluentV2/README.md`，明确第四版收尾结果、后端页面迁移审计结论，以及第五版起步范围，避免后续继续把“哪些页需要进后端页面注册”当成模糊问题反复判断。
- 继续推进第五版稳定化，修复了 [SideNav.tsx](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluentV2/src/features/shell/components/SideNav.tsx) 中将多个 `makeStyles` class 以模板字符串拼接后再传给 Fluent `Button` 的问题，统一改为 `mergeClasses(...)`，收掉了开发态的 `mergeClasses()` 控制台错误噪音。

### 破坏性调整
- 后端 `permissionseed.DefaultPages()` 新增了 `system.message.more`、`team.message.more` 两个页面种子；若有脚本依赖旧页面列表，需要同步接受这两条新页面注册记录。
- 前端导航标题展示策略改为“本地路由优先，后端运行时标题兜底”；如果后续新增本地路由时希望保留后端标题，需显式调整本地 `shellTitle`。

### 需要人工回归验证
- 重新执行 `backend/cmd/migrate` 后，确认 `system/more`、`team/more` 已进入 `ui_pages`，并在系统页面治理页中可被查询到。
- 登录后检查左侧导航、页签、面包屑和页面头部标题，确认已实现页面优先显示 React 侧中文标题，而不是 `menus.*` 形式的旧键名。
- 对 `system/message`、`team/message` 及其 `more` 页再做一轮联动回归，确认新增页面注册不会影响消息域现有权限和激活菜单逻辑。
- 继续抽查其他壳层交互页，尤其是移动端侧栏、消息页、菜单页和页签交互，确认后续新增样式时不再回退到字符串拼接 className 的写法。

## 2026-04-02 frontend-fluentV2 第四版全量页面迁移版

### 本次改动
- 在第三版真实认证、真实运行时导航、真实菜单治理基线上继续增量演进，没有重做壳层、认证或请求分层，而是把 `route registry` 拆成按域组织的本地路由清单，开始承接 Vue 侧既有页面的 React 全量对应物。
- 新增第四版真实页面矩阵：控制台、收件箱、个人中心、页面治理、接口治理、快捷入口、菜单空间、访问轨迹、角色、用户、权限动作、功能包、团队边界角色，以及系统域与团队域的消息、模板、发送人、接收组、记录、更多入口页都已接入真实 React 页面，不再依赖占位页兜底。
- 继续沿用 `API module -> adapter -> query/mutation hook -> page` 分层，补齐 dashboard、inbox、system、access、message、team 等域服务，页面层不再直接解析后端 DTO，也不再把 response envelope 当数组直接消费。
- 新增一批 Fluent 2 工作台页范式：控制台采用摘要型 workbench，收件箱和消息页采用三栏协作布局，系统/团队治理页采用列表 + 详情/编辑的治理工作面，视觉方向统一为微软企业后台风格。
- 扩展公共静态路由，补齐 `/403`、`/404`、`/500`、`/result/success`、`/result/fail`、`/outside/iframe/*` 等真实 React 对应页，并让 `register`、`forgot-password` 成为正式认证页的一部分。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 通过；同时在新的 dev 端口上完成登录后全量受保护路由烟测，新增页面未再落入占位页或错误边界。

### 破坏性调整
- `frontend-fluentV2/src/features/navigation/route-registry.tsx` 不再维护单文件内联本地路由，后续新增本地已实现页面时需要继续按域写入 `core-routes.tsx`、`system-routes.tsx`、`team-routes.tsx`。
- 多个治理页现在直接依赖规范化后的 relation envelope（`ids / items / inherited / records`）；若后续改动 adapter 返回结构，必须同步调整页面而不是在页面里补 DTO 兼容。

### 需要人工回归验证
- `workspace/inbox`、系统消息页、团队消息页的发送、已读、待办动作是否在真实业务数据下全部符合预期。
- `system/page`、`system/api-endpoint`、`system/role`、`system/user`、`system/action-permission`、`system/feature-package`、`team/team`、`team/team-members` 的创建/编辑主链虽然已接通，但仍需结合真实业务数据做一轮逐页 CRUD 回归。
- 开发环境控制台仍存在 `mergeClasses()` 噪音，说明部分 className 组合方式还需要继续清理；虽然不阻塞构建，但应作为下一轮稳定化目标。
- 公共页中的注册、忘记密码、异常页与外链页已接入 React 路由，仍建议在完全游客态下再做一轮独立回归，确认登录态恢复和重定向不会干扰公共路由。

## 2026-04-02 frontend-fluentV2 第三版稳定化与菜单治理编辑版

### 本次改动
- 在第二版真实认证链路上补齐 refresh token 闭环，接入单飞刷新与 refresh 失败统一清会话回登录，页面层不再感知 token 刷新细节。
- 将认证启动阶段状态收口为更明确的 bootstrap 流程，并为导航、菜单树、详情和关联页面查询补上更稳定的 query key 与 placeholderData 策略，降低启动和切空间时的闪动。
- 增补统一空间级失效策略：切换菜单空间或执行菜单 create/update/delete 后，会集中刷新运行时导航、菜单树、详情、关联页面与分组查询。
- 将 `system/menu` 从第二版“真实浏览 / 只读详情”推进为第三版“受控编辑版”，支持 URL 恢复 `spaceKey / selectedMenuId / keyword`，并接入顶级/同级/子级菜单创建、核心字段编辑、类型感知表单、管理分组归属修改。
- 接入真实删除预检与删除确认：支持 `single / cascade / promote_children` 模式预览，删除成功后会刷新树、恢复右侧详情上下文并同步 URL。
- 已验证 `pnpm --dir frontend-fluentV2 install`、`pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前生产构建仍存在既有的大 chunk 警告，需要后续按页面或功能块拆分。

### 破坏性调整
- `frontend-fluentV2` 的 query key 已统一重命名分组；若后续新增 query 或 invalidate 逻辑，必须继续沿用 `auth / navigation / menu` 当前命名，不要再添加散落 key。
- `system/menu` 当前选中节点和搜索关键字已进入 URL；若后续修改菜单页路由参数策略，需要同步考虑刷新恢复与删除后回退逻辑。

### 需要人工回归验证
- access token 过期后，多请求并发触发时是否只发生一次 refresh，并且 refresh 失败后能稳定回到登录页。
- 切换菜单空间后，当前页面若为 `system/menu`，URL、树高亮、详情区、表单和删除预检是否都与新空间一致。
- `system/menu` 创建同级 / 子级节点后，树定位与右侧表单切换是否符合预期；删除当前节点后是否总能回落到合理节点或空态。

## 2026-04-02 frontend-fluentV2 第二版接入真实认证、运行时导航与菜单浏览

### 本次改动
- 将 `frontend-fluentV2` 从第一版 mock shell 升级为第二版真实运行时基础层，补齐真实登录、会话恢复、当前用户获取、退出登录、全局 `401` 清会话与 redirect 回跳闭环。
- 重构共享请求层与适配层，新增统一 Axios client、接口模块、Query key、错误模型与 adapter 映射，页面层不再直接消费后端 DTO 细节。
- 接入真实菜单空间与运行时导航：当前空间可持久化，登录后会按空间加载导航树，左侧导航、面包屑、迁移占位页与 `system/menu` 查询上下文会联动刷新。
- 路由继续保持静态壳 + route registry 模式：本地已实现页面正常进入，运行时存在但未迁移页面统一进入上下文化占位页，避免动态组件路径直驱前端实现。
- 将 `system/menu` 从占位页推进为真实浏览版，支持真实菜单树加载、空间联动、搜索、节点详情只读展示、页面绑定信息只读展示与基础空态/错误态渲染。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 通过；当前生产构建仍有既有的大 chunk 告警，但不影响本次接入结果。

### 下次方向
- 第三版优先继续沿第二版的 adapter、query 与 route registry 基础，逐步把 `system/page`、`system/interface`、`system/role` 等治理链路从占位页迁移为真实页面。
- `system/menu` 下一阶段建议补编辑、新增、删除、排序和更完整的关联治理能力，但继续保持危险操作集中在 feature/service 层，不把 DTO 兼容逻辑回灌到页面组件。

# 2026-04-02 frontend-fluentV2 去除 React Fluent 2 显式标识

### 本次改动
- 清理了 `frontend-fluentV2` 中面向用户的迁移线文案，把错误页、初始化提示、404、迁移占位页、欢迎页和系统菜单中的 `React Fluent 2` / `Fluent 2` 显式字样替换为中性表述。
- 同步更新了应用配置里的产品名与副标题，避免壳层顶部和页面空态继续暴露迁移线标签。
- 保持现有壳层逻辑不变，只调整可见文案和品牌描述。

### 下次方向
- 如果后续还要继续淡化迁移痕迹，可以再检查 README、路由元信息和 mock 数据里的剩余品牌措辞。
- 当前功能层没有改动，后续可继续围绕壳层布局、导航体验和真实业务页面迁移推进。

## 2026-04-02 React Fluent 2 页签标签栏补齐右键菜单与轻量标签组

### 本次改动
- 继续增强 `frontend-fluentV2` 的页签壳层，补上了右键上下文菜单与“关闭左侧 / 右侧标签”能力，并将页签状态持久化到本地，支持页面刷新后的恢复。
- 为页签增加了轻量“合并”模式：连续同模块标签会自动并入一个标签组，标签组支持折叠，用于在页签较多时降低横向噪音。
- 当前页签仍保留固定、刷新、关闭其他、拖动重排与横向滚动优化；右侧工具条和右键菜单形成了双入口，不需要把所有操作挤进单个页签按钮。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次页签增强无关。

### 下次方向
- 可继续补固定标签持久化策略、右键菜单中的“关闭当前组 / 展开当前组”，以及标签组的更明确视觉区分。
- 如果后续真的要做浏览器式复杂“标签合并”，建议继续往工作集 / 标签组模型推进，而不是硬套标准 Tablist。

## 2026-04-02 React Fluent 2 页签标签栏增强

### 本次改动
- 在 `frontend-fluentV2` 的页签标签栏基础上继续补齐固定标签、关闭其他标签、刷新当前标签和桌面端拖动重排能力，操作条统一收口在标签栏右侧。
- 新增页签排序与刷新状态管理：固定标签会自动保持在左侧分区，刷新当前标签会重新挂载当前内容区，关闭其他标签时会保留固定标签与当前标签。
- 对标签栏滚动体验做了优化，支持鼠标滚轮横向滚动、当前页签自动滚入可见区域，并在左右边缘增加渐隐提示，减少长标签列表的压迫感。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次页签增强无关。

### 下次方向
- 若继续往浏览器式体验靠拢，下一步建议补“右键菜单”“关闭左侧 / 右侧标签”“固定标签持久化”和“多标签恢复”。
- Fluent 官方 Tablist 更适合少量相关内容切换，并建议超出宽度时使用 overflow menu；页签分组或“合并”应视为自定义壳层模式，后续更适合做成可折叠的标签组，而不是直接复用标准 Tablist。

## 2026-04-02 React Fluent 2 壳层新增页签标签栏

### 本次改动
- 为 `frontend-fluentV2` 的应用壳新增了浏览器式标签栏，位置放在顶部栏下方、页面内容上方，形成“顶部栏 -> 标签栏 -> 页面内容”的三层结构。
- 新增壳层级页签状态：访问已注册路由会自动打开对应标签，相同路径不重复开标签，关闭当前标签时会自动跳到相邻标签。
- 标签能力接入了路由注册表与页面元数据解析，并在退出登录时清空当前会话的已打开标签，避免跨会话残留。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次页签功能无关。

### 下次方向
- 可继续补固定标签、关闭其他标签、刷新当前标签等增强交互，让页签更接近正式后台工作台。
- 若后续接入运行时菜单或 host 驱动路由，可直接复用当前壳层页签模型，只替换标签来源与恢复策略。

## 2026-04-02 React Fluent 2 补齐本地认证闭环

### 本次改动
- 在 `frontend-fluentV2` 中新增本地假认证 store，支持登录、退出登录、记住登录状态以及基于 `localStorage / sessionStorage` 的会话持久化。
- 接入了认证守卫和游客重定向：未登录访问后台壳会自动跳到 `/login`，已登录访问 `/login`、`/register`、`/forgot-password` 会直接回到默认工作区。
- 登录页已支持登录成功后按来源地址回跳，顶部用户菜单新增退出登录，认证页与后台壳形成完整本地闭环。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前构建仍有既有的大 chunk 警告，但与本次认证逻辑无关。

### 下次方向
- 若继续接真实认证，下一步优先把登录提交动作替换为 adapter 请求，并补上令牌刷新、401 处理和登录态失效回退。
- 后续可继续补密码显隐、错误提示、验证码或邮箱确认等细节，同时把当前本地假会话逐步替换成真实会话模型。

## 2026-04-02 React Fluent 2 新增认证页骨架

### 本次改动
- 为 `frontend-fluentV2` 新增了不接真实认证的登录、注册、忘记密码三张独立页面，并通过壳层外路由接入 `/login`、`/register`、`/forgot-password`。
- 认证页统一采用居中单卡片布局，保持 Fluent 2 风格的简洁输入表单，去掉了第三方登录入口和不必要的营销式文案，只保留必要字段、跳转链接和本地示例反馈。
- 新增了认证页公共骨架，统一品牌、留白、底部链接和响应式行为；已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。
- 当前构建仍有既有的大 chunk 警告，但与本次认证页接入无关。

### 下次方向
- 若后续要接真实登录，可先把认证表单提交动作替换为 adapter 调用，再决定是否引入真正的未登录守卫与返回地址逻辑。
- 下一轮可继续补认证页的细节，例如密码可见切换、基础校验文案、登录成功后的 redirect 规则，以及移动端输入体验收口。

## 2026-04-02 React Fluent 2 壳层品牌配色收口

### 本次改动
- 为 `frontend-fluentV2` 新增统一品牌主题源，改用一套更克制的蓝灰企业色品牌梯度生成 Fluent light / dark theme，不再直接沿用默认品牌蓝。
- 同步更新了应用壳主题 Provider、错误边界页、搜索弹层选中态和品牌 logo 渐变，让导航高亮、链接、Badge、欢迎页渐变与 logo 颜色保持一致。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；当前仍存在既有的大 chunk 警告，但与本次配色调整无关。

### 下次方向
- 继续检查深色模式下的新品牌色对比度，重点看顶部栏图标、侧栏激活态和搜索结果高亮是否仍然足够清晰。
- 若后续继续细调，可把状态色、空态插画和少量硬编码图形资产进一步收口到同一品牌色板，避免局部残留旧蓝。

## 2026-04-01 Fluent 2 React 实验场新增 React 组件目录页

### 本次改动
- 基于 [Fluent 2 React 组件目录](https://fluent2.microsoft.design/components/web/react) 为 `frontend-fluent2-lab` 新增了 4 张专门的组件展示页，集中承载命令与导航、表单与选择、反馈与浮层、身份与内容四类组件。
- 新增文件 [FluentReactComponentPages.tsx](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/FluentReactComponentPages.tsx)，使用当前 `@fluentui/react-components` 已安装的稳定导出组件做真实示例，并把官方目录中的其余组件收进补充索引区。
- 同步更新 [catalog.ts](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/lab/catalog.ts)、[App.tsx](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/App.tsx) 和 [README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/README.md)，让实验场总页数增至 54。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过；未做浏览器逐页人工回归。当前新增组件页单文件 chunk 较大，后续如继续扩展可考虑拆分。

### 下次方向
- 继续把组件目录页和现有 Fluent 规范页打通，逐步将组件页从“目录陈列”收口成“组件 + 模式 + 场景”的联动结构。
- 若继续深化，可优先补 `Icon`、`Carousel`、`TagPicker` 等官方目录项的更完整示例，并检查新组件页在暗色主题下的观感与滚动密度。

## 2026-04-01 Fluent 2 React 实验场 Teams 场景骨架深化

### 本次改动
- 继续调整 `frontend-fluent2-lab/src/pages/ScenarioExpansionPages.tsx`，把 Teams 线页面进一步收成更明确的协作工作面，而不是只保留统一壳层。
- 重点强化了会议指挥、一线简报、文件协作、社区公告、交接班、活动运营、伙伴站会等页面的 Teams 风格骨架，引入议程条、会议舞台、成员条、文件审阅板、公告流、交接三栏、run of show 等更明确的结构。
- 这轮改动的目标是让 Teams 页面更像“频道 / 协作 / 会务 / 社区 / 值守”工作面，而不是同一模板换文案。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过；未做浏览器逐页视觉回归。

### 下次方向
- 继续对 Teams 线页面做浏览器回归，优先检查亮暗主题下的品牌紫使用强度、右侧上下文区密度和移动端收口方式。
- 若继续深化，下一步最值得做的是把会议指挥、文件协作、交接班这三类页面对齐到真实 Teams UI Kit 或 Microsoft Teams App Templates 的具体节点。

## 2026-04-01 Fluent 2 React 实验场剩余页面去模板化

### 本次改动
- 继续重构 `frontend-fluent2-lab/src/pages/ScenarioExpansionPages.tsx`，把剩余仍在复用家族骨架的页面继续拆成更明确的专用布局。
- 新增并接入租户总览、事件响应、策略工作室、资产台账、工单台、发布控制台、知识中枢、表单规范台、导航规范台、Token 治理、可访问性评审、交接模式、模板画廊、动效原则、侧栏参考、会议指挥、前线简报、文件协作、入职协作、社区公告、班次交接、直播运营、伙伴站会等专用工作面。
- 这轮调整的目标不是再扩页数，而是让 50 张实验页在结构、节奏和任务语法上真正拉开，减少“同一模板换内容”的重复感。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过；未做浏览器逐页人工视觉回归。

### 下次方向
- 继续从这批页面里挑出更核心的工作面，结合真实 Figma 节点和 Fluent 2 文档进一步收紧层级与细节。
- 后续可优先检查 `ScenarioExpansionPages` 的 chunk 体积与暗色主题观感，再决定是否拆分代码块或继续做视觉校准。

## 2026-04-01 Fluent 2 React 实验场扩展到 50 页

### 本次改动
- 为 `frontend-fluent2-lab` 新增场景数据层，并把新增页从“按来源三套整页模板”重构为“按场景骨架”组织。
- 将 30 张新增实验页重新映射到指挥台、工作台、目录页、规范页、评审页、频道页、线程页、公告页等多种布局家族，减少单纯换文案的重复页面。
- 重写实验场目录注册表并更新 `frontend-fluent2-lab/src/App.tsx`、`frontend-fluent2-lab/README.md`，让入口支持完整 50 页切换，并把总页数展示改为动态值。
- 已验证 `pnpm --dir frontend-fluent2-lab build` 通过。

### 下次方向
- 从新增 30 张页面里继续挑出高价值工作台页，按真实 Figma 节点收紧布局与信息层级，而不是继续只扩数量。
- 若后续继续做视觉回归，可优先检查 `ScenarioExpansionPages` 大包的拆分策略，以及暗色主题下的新页面对比度。

## 2026-04-01 React Fluent 2 基础壳初始化

### 本次改动
- 新建 `frontend-fluentV2/` 独立 React + Fluent 2 工程，并接入 Router、Query、Axios、Zustand 与 Fluent Provider。
- 落地应用壳、顶部栏、空间切换器、侧边导航、面包屑、统一页面容器和迁移占位页。
- 重建根目录当前生效的协作约束、项目框架与前端规范，并新增 `frontend-fluentV2/docs/` 正式文档目录。
- 已验证 `pnpm --dir frontend-fluentV2 install`、`pnpm --dir frontend-fluentV2 exec tsc --noEmit`、`pnpm --dir frontend-fluentV2 build`。
- 已验证 `pnpm --dir frontend-fluentV2 dev` 可启动；由于本机 `9030` 已被占用，Vite 在 2026-04-01 实际回退到 `http://127.0.0.1:9031/`。

### 下次方向
- 优先迁移系统治理链路中的菜单、页面、接口、角色与用户页面。
- 将 mock 查询逐步替换为真实 adapter，同时保持壳层、路由和页面容器不变。
- 若继续收拢历史资料，可为 `frontend/docs/legacy/` 增加分类索引。

## 2026-04-01 React Fluent 2 侧栏壳体调整

### 本次改动
- 将品牌区从顶部栏移入左侧导航顶部，桌面端不再保留单独的“侧栏”收缩按钮。
- 侧栏顶部品牌位改为承担收起/展开交互：展开态显示品牌块，收起态显示右箭头，再次点击可展开。
- 同步收紧了侧栏宽度和头部信息布局，使顶部栏更聚焦当前区域与全局操作。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 继续细调侧栏顶部品牌块、分组标题和激活态样式，使其更接近目标参考图的节奏。
- 下一轮可继续收口菜单项密度、内容区留白以及顶部栏空间切换器的视觉权重。

## 2026-04-01 React Fluent 2 顶部栏与侧栏穿插布局

### 本次改动
- 调整桌面端壳体层级关系，让顶部栏为左侧菜单预留通道，菜单区域本体再向上抬入顶部栏范围。
- 侧栏不再只是位于内容区下方，而是形成“左侧菜单超过顶部栏基线”的视觉关系，更接近目标参考图。
- 同步根据侧栏展开/收起状态动态调整顶部栏左侧偏移，避免全局操作区与菜单区域互相挤压。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 下一轮继续细调菜单卡片顶部阴影、边界和内容区首屏对齐，让穿插关系更自然。
- 可继续收口页面标题区和菜单顶部品牌块的垂直节奏，避免出现“刚好碰线”的中间态。

## 2026-04-01 React Fluent 2 壳体回退到初版

### 本次改动
- 回退了“品牌区移入侧栏”“菜单上抬穿插顶部栏”“品牌区承担收缩交互”等实验性壳体改动。
- 恢复到最开始的稳定结构：品牌区回顶部栏、左侧菜单回归常规侧栏、桌面端独立收缩按钮恢复。
- 保留新工程既有的 Provider、路由、导航树、页面容器和 mock 数据骨架，不回退基础工程能力。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 之后如果要重新设计壳体，建议先用固定线框确认“顶部栏 / 侧栏 / 内容区”的边界，再进入视觉细调。
- 下一轮可以从初版结构上继续优化间距、层级和组件样式，但不再直接做跨区穿插实验。

## 2026-04-01 React Fluent 2 移除顶部栏空间切换

### 本次改动
- 移除了顶部栏里的手动菜单空间切换入口，不再把空间切换作为当前壳体的可见交互。
- 保留 `currentSpace` 状态和按空间过滤导航的底层能力，后续可改为按 host 自动同步，不需要推翻现有骨架。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 下一轮如接入 host 上下文，可直接把空间值写入 shell store，并移除当前 mock 默认空间初始化逻辑。
- 之后顶部栏可继续收口为“品牌 / 当前区域 / 全局操作 / 用户区”的更轻量结构。
## 2026-04-01 Fluent 2 实验场新增基础控件集与图标总览

### 本次改动
- 为 `frontend-fluent2-lab` 新增 `基础控件集` 分类，并将组件展示页、React 组件页统一迁入该分组，避免底层能力继续散落在 `Fluent 2 Web` 分类里。
- 新增 `React 图标总览` 页面，集中展示 `@fluentui/react-icons` 的基础图标 ID，支持按关键字过滤、按首字母分组浏览，并点击复制图标标识。
- 继续补齐组件页中此前只停留在“目录补充”的缺口组件，已落地 `Nav`、`Label`、`Dropdown`、`TagPicker`、`Carousel`、`FluentProvider` 的真实示例，并将 `Icon` 能力明确收口到图标总览页。
- 同步更新 `catalog`、实验场入口、README 与分组计数，已验证 `pnpm --dir frontend-fluent2-lab build` 通过。

### 下次方向
- 继续补图标页的变体使用示例，例如按 `Regular / Filled` 生成推荐导入名，或补充常用图标组合样板。
- 后续可再做一次浏览器回归，重点检查图标页的大数据量滚动、暗色主题对比度、移动端折叠表现，以及组件页新增示例的交互细节。

## 2026-04-01 React Fluent 2 顶栏接入真实菜单搜索

### 本次改动
- 参考旧 Vue 壳层里的 `ArtGlobalSearch`，在 `frontend-fluentV2` 新增了顶栏级菜单搜索弹层，而不是继续保留占位图标按钮。
- 新增导航拍平与搜索模型，按当前可见导航树递归提取叶子菜单，支持按菜单名、分组、路径过滤，并接入最近访问记录。
- 顶栏搜索按钮已可真实打开搜索弹层，并支持 `Ctrl/Cmd + K`、方向键切换、回车跳转；已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`。

### 下次方向
- 继续补搜索结果的高亮标注和分组视觉，让命中原因更直观。
- 若后续接真实 host 上下文或动态菜单，可把当前基于 mock 的搜索数据源替换成运行时导航注册表，而不重写弹层交互。

## 2026-04-01 Fluent 2 实验场新增 React 组合模式页

### 本次改动
- 为 `frontend-fluent2-lab` 的 `基础控件集` 新增三张组合模式页：`React 组合页：导航与命令`、`React 组合页：表单与反馈`、`React 组合页：内容与协作`。
- 这三张页不再按“单组件示例”堆砌，而是把 `Nav`、`Toolbar`、`SearchBox`、`Field`、`TagPicker`、`Dialog`、`Card`、`Persona` 等基础控件收口成真实工作面，减少基础控件集继续千篇一律的风险。
- 已同步更新 `catalog`、实验场入口、README 和页面总数统计，并执行 `pnpm --dir frontend-fluent2-lab build` 验证；当前未做浏览器逐页人工回归。

### 下次方向
- 继续给组合模式页补更贴近真实开发使用的复制能力，例如图标导入名、组件组合代码片段或推荐搭配说明。
- 后续可把这三张组合页继续对齐官方 Fluent 2 React 文档结构，补一轮暗色主题和移动端滚动节奏回归。
## 2026-04-02

- 新增根目录文档 `使用与迁移部署说明.md`，整理了后端、前端、`a/` 目录的本地使用方式和迁移部署步骤。
- 明确了后端依赖的 PostgreSQL、Redis、Elasticsearch 和 MinIO 配置项，以及前端的安装和构建流程。
- 对 `a/` 下的复制副本用途做了说明，便于后续做独立迁移、打包和交付。

## 2026-04-02 React Fluent 2 标签栏新增顶部开关

### 本次改动
- 在 `frontend-fluentV2` 顶部左侧功能区新增“界面设置”入口，并提供“显示标签栏”开关，统一管理标签栏是否显示。
- 壳层在关闭标签栏时不再为后续路由访问自动开标签，仅保留已持久化标签状态，避免彻底丢失当前工作区上下文。
- 同步更新标签栏壳层规则文档，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 通过；本次未额外执行生产构建。

### 下次方向
- 继续收口标签栏体验，可补“关闭当前组”“固定标签优先恢复”“右键菜单边缘避让”等细节。
- 如后续需要更强的工作区能力，可在当前开关基础上继续扩展“默认启用策略”“按用户偏好记忆”和更完整的标签组模型。

## 2026-04-02 React Fluent 2 标签栏改为显式自由分组

### 本次改动
- 将 `frontend-fluentV2` 标签栏从“按模块自动归并”切换为“用户显式组合”，不再按上级菜单或模块名自动成组。
- 在标签右键菜单中新增“与左侧标签成组”“与右侧标签成组”“移出当前组合”“解散当前组合”，并把组合状态持久化到本地。
- 组合折叠后改为显示首个标签标题与数量；标签拖出原组合时会自动脱离原组合，保证分组结构与实际顺序一致。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补“关闭当前组合”“组合重命名”“组合内整体拖动”等更完整的工作集能力。
- 若后续需要更接近浏览器的标签体验，可继续研究组合颜色、固定组合和组合恢复策略，但建议先稳定当前显式分组心智。

## 2026-04-02 React Fluent 2 标签栏补齐即时滑动与组合拖放

### 本次改动
- 将标签拖拽重排改为拖动过程中的即时滑动，不再等到松手后才调整顺序；单标签与组合拖动都采用同一套壳层状态更新方式。
- 将“刷新当前标签”统一改名为“刷新当前页面”，与实际行为保持一致；工具栏按钮和右键菜单文案已同步收口。
- 支持把单个标签拖放到已有组合上加入该组合，组合本身也支持整组拖动重排；专题文档同步更新为“即时滑动 + 组合拖放”的新交互语义。
- 已验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补“关闭当前组合”“组合重命名”“组合整体拖出为单标签”和更细的拖拽目标高亮，让组合交互更接近浏览器工作集。
- 若后续要进一步对齐浏览器标签组，可继续研究组合颜色、组合固定和组合级右键菜单，但建议优先稳定当前拖拽手感与边界行为。

## 2026-04-02 React Fluent 2 标签栏移除刷新入口并修正隔位拖拽

### 本次改动
- 移除了标签栏工具区和右键菜单中的“刷新当前页面”入口，避免标签栏继续承载与浏览器刷新心智冲突的动作。
- 将标签拖拽重排的判定从“路径锁定”改为“按悬停标签中线即时换位”，修复了拖到相邻标签后无法继续回拖、会出现一格隔离的问题。
- 专题文档同步改成当前交互语义，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补组合级右键菜单、组合整体拖出和更明确的拖拽占位高亮，让多标签工作区更接近浏览器标签组体验。
- 后续若要继续增强，可考虑加入拖拽自动横向滚动和组合级固定策略，但建议先观察当前即时换位的实际手感。

## 2026-04-02 React Fluent 2 标签栏改为拖拽合并

### 本次改动
- 去掉了标签右键菜单中的成组入口，将“合并”统一收口到拖拽手势里，不再让同一能力同时存在菜单和手势两套入口。
- 单标签拖动经过目标标签靠近来向一侧的三分之一区域时，会显示合并预览；继续拖深后才执行位置换位，从而把“合并”和“重排”分成两段手势。
- `AppShell` 与标签栏组件已切换到新的 `groupTabs(sourcePath, targetPath)` 接口，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build`；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续补组合级拖拽高亮、组合整体拖出为单标签和拖拽自动横向滚动，让多标签体验更接近浏览器标签组。
- 如果后续要继续深化，可再研究组合级右键菜单，但建议只保留和组合本身强相关的动作，避免再次把合并入口做回菜单里。

## 2026-04-02 React Fluent 2 标签栏移除全部组合逻辑

### 本次改动
- 将 `frontend-fluentV2` 标签栏恢复为纯平标签轨道，删除了标签组合、折叠组合、合并预览、组级右键菜单和相关持久化状态，只保留打开、关闭、固定与拖拽换位。
- `useShellStore` 改回仅维护 `openTabs` 和 `tabsEnabled` 两类标签状态，`AppShell` 也同步移除了分组接线与页面刷新版本号逻辑，避免后续继续被旧分组模型干扰。
- `OpenTabsBar` 已重写为单层标签实现，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 与 `pnpm --dir frontend-fluentV2 build` 通过；生产构建仍存在既有的大 chunk 告警。

### 下次方向
- 继续收拖拽换位的手感和横向滚动体验，但不再重新引入任何组合语义，先把纯标签工作区稳定下来。
- 如果后续还要增强标签栏，优先考虑固定标签恢复、右键菜单边缘避让和移动端展示，不再回到浏览器标签组那条路线。

## 2026-04-02 React Fluent 2 标签栏拖拽内核重构

### 本次改动
- 删除 `frontend-fluentV2` 标签栏里旧的 HTML5 DnD、旧拖拽幽灵层、旧目标判定和旧 FLIP 实现，重写为 `Pointer Events + 自定义拖拽层 + wrapper 级让位动画` 的纯平换位模型。
- 新拖拽逻辑按拖动标签实体当前位置计算换位，源标签在原位只保留占位壳；其他受影响标签在每次顺序变化前记录当前屏幕位置，并从该位置平滑过渡到新位置。
- 同步更新标签栏专题文档，并验证 `pnpm --dir frontend-fluentV2 exec tsc --noEmit` 通过；生产构建待本轮改动后一并复核。

### 下次方向
- 继续观察拖拽手感，重点回看横向自动滚动、边缘拖拽阈值与触控板场景，但不再恢复任何分组或合并逻辑。
- 若后续还要增强标签栏，优先补纯平标签模式下的细节体验，例如拖拽中的滚动辅助和右键菜单边缘避让。 

## 2026-04-02 frontend-fluentV2 第七版导航映射与团队域联调复核

### 本次改动
- 将运行时导航中的相对路径和旧别名统一映射到本地 React 路由，补齐 `Dashboard`、`Console`、`TeamRoot`、`TeamManage`、`TeamMembers`、`TeamRolesAndPermissions`、`TeamMessageManage` 等入口，左侧导航、标签和面包屑不再落到 `#/team`、`#/members`、`#/roles` 这类半成品路径。
- 为 `team/team` 与 `system/team-roles-permissions` 补齐选中项失效后的 URL 清理，避免删除、切换或列表刷新后遗留脏 `selectedTeamId` / `selectedRoleId`。
- 使用全新 Playwright 浏览器会话重新验证 `system/team-roles-permissions` 与 `team/team-members`；在最新后端环境下，`/api/v1/tenants/my-team/boundary/roles`、`/action-origins`、`/menu-origins` 均返回 `200`，页面首次渲染控制台 `0 error`。

### 下次方向
- 继续用真实团队数据做深回归，重点验证团队边界角色写操作、成员边界角色分配和团队消息主链，而不是只停留在读链路。
- 若后端运行时导航后续再新增旧风格相对路径菜单，继续把映射收口在前端导航层，不把别名兼容散落到页面组件里。

## 2026-04-02 frontend-fluentV2 第七版消息调度作用域与团队发送闭环

### 本次改动
- 为消息请求层补齐 `tenantMode` 和 `scope` 驱动，平台消息请求会显式移除 `X-Tenant-ID`，团队消息继续带当前租户头，`system/message` 与 `team/message` 不再误用同一组数据。
- 将消息调度 payload 映射为后端真实字段：`specified_users -> target_user_ids`、`tenant_users/tenant_admins -> target_tenant_ids`、`recipient_group/role/feature_package -> target_group_ids`，并把调度页目标预览改成可勾选卡片，不再只靠手工输入 ID。
- 修正团队消息页初始草稿仍落到 `all_users` 的问题，调度草稿与发送后重置都改为以 `/api/v1/messages/dispatch/options` 的默认受众、默认发送人和默认优先级驱动。
- 后端 `messages/dispatch` 增量修正了 `normalizeTargetTenants`，`specified_users`、`recipient_group`、`role`、`feature_package` 不再被误判为“不支持的发送对象”；前端与浏览器已实测发通系统域和团队域消息。

### 下次方向
- 继续回归消息域的已读、全部已读、待办完成/忽略和记录详情，重点确认左列计数、中列记录和右侧详情在操作后保持同步。
- 如后续还要提升消息工作台，可继续优化调度成功回执与受众摘要，让团队域默认当前团队的回执文案和投递数表达更贴近业务语言。

## 2026-04-02 frontend-fluentV2 增补下次方向记录文档

### 本次改动
- 新增 `frontend-fluentV2/docs/下次方向记录.md`，专门承接当前迁移线尚未完成的后续方向，避免未来继续推进时只依赖对话上下文。
- 在 `AGENTS.md` 中补充了维护规则：该文档存在时持续更新，全部完成后直接删除，不长期保留空文件。

### 下次方向
- 后续每轮完成时继续同步清理该文档中的已完成项，确保它只保留真实未完成事项。
- 暂不把这条规则下沉到全局技能；如果后续多个仓库都需要同样机制，再考虑把它抽成技能行为。

## 2026-04-02 frontend-fluentV2 下次方向记录改为持续备案

### 本次改动
- 将 `frontend-fluentV2/docs/下次方向记录.md` 的语义从“本轮下次方向”收口为“持续维护的方向备案”，明确要求后续只按条目增删，不整份覆盖。
- 同步更新 `AGENTS.md`，强调该文档允许跨轮次、跨模块持续保留未完成事项；只有全部完成时才删除。

### 下次方向
- 后续继续推进时，只修改本次涉及的事项条目；完成则删除，未完成则保留，不再把整份文档当成每轮收尾模板重写。
- 若未来多个仓库都要采用同一机制，再考虑把这条规则抽进技能，而不是当前就提升为全局行为。

## 2026-04-02 frontend-fluentV2 父菜单保留页面入口

### 本次改动
- 将侧栏点击规则收回到“有路径的父菜单仍可作为页面入口”，不再因为存在子节点就一刀切禁用跳转。
- 收缩态级联浮层同时补出父菜单自身的可点击菜单项，避免父菜单的页面入口在浮层中被子项覆盖掉。

### 下次方向
- 继续回归带子菜单的父节点，确认展开态、收缩态和移动端抽屉三种状态下，父节点页面入口与子菜单展开都符合后台菜单配置语义。
- 后续若仍出现“同名但不同语义”的菜单项，优先回查后端种子和迁移，而不是继续在前端做更激进的隐藏。

## 2026-04-02 frontend-fluentV2 第8版 Fluent 2 Web / Teams 页面收口

### 本次改动
- 新增 `PageStatusBanner`，把消息、团队等治理页的成功/失败反馈统一成更轻的 Fluent 2 提示条，减少各页自行拼接 MessageBar 的重复样式。
- 收口 `PageContainer`、`SectionCard`、`WorkbenchLayouts` 的视觉参数，整体压浅阴影、边框和间距，使页面更接近 Fluent 2 Web 的低噪后台风格。
- 为 `EntityPageLayout` 补齐更完整的标题/说明/元信息/操作区结构，作为后续页面模块化拆分时的统一骨架。
- 已将消息域、团队域中多处页面反馈接入新的统一提示组件，并完成 `tsc` 与 `build` 验证通过。
- `system/menu`、`message/workspace`、`team/team`、`system/user`、`system/team-roles-permissions`、`system/feature-package` 等高频治理页正在继续切换到统一的页面反馈组件，减少旧式 MessageBar 视觉碎片。

### 下次方向
- 继续把 `system/menu`、`system/page`、`system/user`、`team/team`、`message/*` 等大页拆回域级组件和对话框，减少 `pages/*` 的单文件体积。
- 继续按 Fluent 2 Web / Teams 视觉规范统一各页的 section、面板、列表和详情区，不再让单页自己决定深阴影和厚块布局。

## 2026-04-02 frontend-fluentV2 菜单治理页组件拆分

### 本次改动
- 将 `SystemMenuPage` 中的树节点与只读字段抽出为 `features/system/components/MenuTreeNode.tsx`，页面本体保留状态、编辑和删除流程。
- 继续把系统菜单页中的错误/提示块统一到 `PageStatusBanner`，减少页面内零散的反馈实现。
- 保持 `tsc` 与 `build` 通过，`system/menu` 页面结构开始向域组件化收口。

### 下次方向
- 继续把 `system/menu` 里的编辑表单、删除弹窗、详情摘要进一步拆成域组件。
- 再继续拆 `message/*`、`team/*`、`system/page`、`system/user` 的内部模块，减少大页体积并统一页面骨架。

## 2026-04-02 frontend-fluentV2 第8版全量页面清单与重页下沉

### 本次改动
- 新增并完善 `frontend-fluentV2/docs/page-inventory.md`，将 Vue 全量页面清单、docker 数据库中的 `menus / ui_pages` 以及当前 React 承接关系统一写成一份范围表，明确第 8 版不再只围绕 `system/menu`、`message/*`、`team/*` 推进。
- 重写 `frontend-fluentV2/README.md`，将说明口径更新到第 8 版：以 `Fluent UI React v9 + Fluent 2 Web + Teams` 为统一规范，以路由装配层 + 域组件 + 真实 API 为当前主线。
- 重写 `frontend-fluentV2/docs/architecture.md`，移除过期的 `features/session`、mock 主线和旧 query key 描述，改成当前真实认证、真实导航、Query 分层与页面范式。
- 将 `WorkspaceInboxPage` 完全切到 `features/workspace/components/InboxPanels.tsx`，收件箱页面现在只负责 URL 状态与查询编排，左中右三栏实现正式下沉到域组件。
- 将 `MessageCatalogPages.tsx` 的模板、发送人、收件组、记录、更多入口整体下沉到 `features/message/components/MessageCatalogWorkspaces.tsx`，`pages/message/MessageCatalogPages.tsx` 退回到轻量装配层。
- 将 `TeamPages.tsx` 的团队、成员、更多入口整体下沉到 `features/team/components/TeamWorkspaces.tsx`，`pages/team/TeamPages.tsx` 退回到轻量装配层。
- 将 `SystemMenuPage` 的右侧摘要、编辑、关联页面与删除确认整体下沉到 `features/system/components/SystemMenuPanels.tsx`，菜单治理页正式形成“树区 + 右侧域面板”的拆分结构。
- 已重新通过 `pnpm --dir frontend-fluentV2 exec tsc --noEmit`、`pnpm --dir frontend-fluentV2 build` 和 `go build ./...` 校验。

### 下次方向
- 继续将 `system/page`、`system/user`、`system/role`、`system/action-permission`、`system/api-endpoint`、`system/feature-package` 中残留的内联模块继续拆到 `features/system/components|dialogs|drawers`。
- 继续用真实菜单树、页面表和 Vue 文件清单回归，确认没有遗漏子页面、入口卡片和模块级交互。
## 2026-04-02

- 清空 `frontend-fluentV2/src/pages`，保留壳层与导航骨架，准备重新重构第 8 版页面体系。
- 删除 `frontend-fluentV2/docs` 专题文档，重置为干净目录，后续只保留新的重构文档。
- 清理了 `frontend-fluentV2/src/features/navigation/routes/core-routes.tsx`、`system-routes.tsx`、`team-routes.tsx` 的页面依赖，并将 `AppRouter` 收敛为壳层入口。
