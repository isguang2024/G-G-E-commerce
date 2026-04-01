# 变更日志

## 2026-04-01 旧前端文档归档与根目录清空

### 本次改动

- 新建 `frontend/docs/legacy` 作为旧技术栈文档归档入口，按 `root/` 与 `docs/` 两层收纳原根目录与原 `docs/` 全量 Markdown。
- 为归档副本补充了归档索引，并将归档区内原本指向根目录与 `docs/` 的 Markdown 链接统一改写到新位置，保证历史文档仍可跳转查阅。
- 删除根目录旧文档 `AGENTS.md`、`PROJECT_FRAMEWORK.md`、`FRONTEND_GUIDELINE.md` 以及原 `docs/` 目录内容，完成旧栈文档出根目录收口。

### 下次方向

- 等新技术栈确定后，在根目录与新的正式文档目录中重建当前生效的协作约束、项目框架和前端规范，不直接复用旧栈文案。
- 视需要继续整理 `frontend/docs/legacy` 的索引颗粒度，例如补充按主题分类或迁移映射说明。
- 本次未执行自动化测试；已完成的验证仅包括文件落点、链接改写和目录清理检查。

## 2026-03-31 访问链路菜单迁移治理

### 本次改动

- 处理 `backend/cmd/migrate/main.go` 中访问链路（AccessTrace）相关迁移的幂等与重复问题：移除两个临时修复迁移（`20260331_access_trace_package_binding`、`20260331_access_trace_parent_rebind_fix`），将入口保留为 `20260331_access_trace_navigation_seed`。
- 将 `Access Trace` 菜单/页面显示名调整为中文 `访问链路测试`，并在菜单查找中统一按 `name + space_key + path` 做幂等匹配，减少未来重复插入与误绑定风险。
- 新增 `menuSeedQuery` 统一查询函数并用于 `ensureMenuSeed`、`syncMenuSeed`，使菜单种子入库在空间/路径一致性下更稳定。

### 下次方向

- 建议执行一次 `go run ./cmd/migrate` 完成新迁移落库（用户环境已安装最新代码后可直接执行），确认 `菜单管理`、`页面管理` 中显示为中文“访问链路测试”，并确认页面链路无重复。
- 后续如有新增类似菜单迁移，统一按 `space_key + path + name` 编写查询条件，避免与历史同名菜单发生串线。

> 只记录大修改、重要节点或功能级收口；过程性推进已转写到专题文档。

## 2026-03-31 菜单备份去历史兼容与页面空间收紧

### 本次改动

- 菜单备份移除历史全局兼容语义，前后端仅保留 `space / global` 两种正式范围。
- 页面空间相关文案收紧为“独立页绑定”，避免继续把 `spaceKey` 讲成页面主归属语义。
- 相关专题文档已同步压缩，确保新约束优先落在当前规则里。

### 下次方向

- 后续如需扩展，只在少量独立页上增加明确绑定，不再把空间能力放大成通用页面模型。
- 菜单备份恢复链路可继续回归一次，确认当前空间备份与全空间备份的提示和行为一致。

## 2026-03-31 文档与日志收口

### 本次改动

- 精简 `AGENTS.md` 的日志规则，明确 `change-log.md` 只记大修改、重要节点或功能级收口。
- 压缩专题文档中的重复背景、阶段性解释和执行结果，删掉过程型说明。
- 将碎片化历史记录收束到专题文档中的当前结论，避免变更日志继续膨胀。

### 下次方向

- 后续新增文档内容优先写“当前规则”和“验收入口”，少写过程回放。
- 再出现关键架构节点时，只追加一条里程碑即可。

## 2026-03-31 单空间与菜单主链收口

### 本次改动

- 单空间运行模式、菜单空间切换、默认空间回退和菜单备份范围已形成稳定主链。
- 运行时导航、页面注册、菜单权限与消息链路统一围绕当前空间和后端正式接口收口。

### 下次方向

- 继续只保留对业务有用的收口记录。
- 需要长期保留的结论，回写专题文档，不依赖日志存活。

## 2026-03-30 系统收尾与空间底座成型

### 本次改动

- 系统管理、消息系统、菜单-页面关系、空间化底座完成主链收口。
- 默认单域可运行，多空间与 Host 绑定能力保留但不做过度扩散。

### 下次方向

- 真实业务模块进入后，再按业务回归暴露新的系统问题。
- 若后续启用真实多 Host，再补登录回跳、请求基址与 Cookie 域策略。

## 2026-03-31 页面空间切换收口与列表精简

### 本次改动

- 去掉页面管理页顶栏的空间选择器，当前列表默认固定在业务主空间上下文。
- 删除“空间视角”列，列表默认只保留页面、路径、组件、挂接对象、归属链路、链路状态、排序、访问状态、更新时间与操作。
- 收紧页面管理页的空间同步逻辑，去掉路由空间参数联动与废弃样式。

### 下次方向

- 若后续确认页面只保留默认空间，可继续收敛页面表单中的空间辅助字段与候选加载。
- 必要时再把“挂接对象 / 归属链路”进一步合并到详情区，继续降低表格横向宽度。
- 未验证项：本次未执行自动化测试。

## 2026-03-31 空间可见文案统一

### 本次改动

- 将页面相关表单中的“独立页绑定”统一改为“空间可见”，减少“绑定”带来的归属歧义。
- 同步收紧提示文案，明确该字段只控制可见范围，不改变页面类型、路径、权限或菜单挂载。

### 下次方向

- 如果后续确认 `全局页` 不应再填写空间可见范围，可进一步按页面类型隐藏或禁用该字段。
- 若要继续压缩页面概念，可把空间可见、挂接对象和归属链路拆到详情区，减少主表字段数量。

- 未验证项：本次未执行自动化测试。

## 2026-03-31 全空间可见与空间多选

### 本次改动

- 后端页面写入接口新增 `space_keys` 多选字段，`[]` 明确表示全空间可见，避免再回退成单空间。
- 页面、逻辑分组、普通分组三类表单统一使用 `空间可见` 多选，增加 `全空间可见` 选项。
- 页面列表和运行时读取继续兼容旧的 `space_key` 语义，但优先使用 `spaceKeys` 和 `spaceScope`。

### 下次方向

- 如果确认全局页不需要单独维护空间候选，可继续按页面类型隐藏空间多选控件。
- 后续再把页面管理页的表格字段做一次收敛，优先保留最常用的操作字段。

- 未验证项：前端联调尚未完整跑通，已完成后端 `go test ./internal/modules/system/page/...`。

## 2026-03-31 Markdown 文档结构收口

### 本次改动

- 收紧了根目录文档边界，在 [AGENTS.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/AGENTS.md) 和 [PROJECT_FRAMEWORK.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/PROJECT_FRAMEWORK.md) 中明确根目录只保留长期稳定约束，`docs/` 只保留有效专题、固定清单和少量关键里程碑。
- 重写了 [docs/README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/README.md) 的索引方式，去掉重复目录式描述，改为“核心入口 / 保留原则 / 删除原则 / 维护规则”。
- 删除了 `docs/system-wrapup-plan.md` 与 `docs/system-compatibility-audit.md` 两份阶段性或重复性文档，避免与主框架和专题文档并存造成多版本语义。

### 下次方向

- 继续检查 `docs/` 中是否还存在“阶段性计划、重复总结、已被专题吸收”的文档，优先按同一规则继续收口。
- 后续如果菜单/页面链路的兼容边界继续稳定，可再评估是否进一步压缩到单一正式设计文档中。

- 未验证项：本次仅做 Markdown 文档收口，未执行代码测试。

## 2026-03-31 Markdown 文档结构收口-第二轮

### 本次改动

- 将 `docs/scaffold-navigation-access-redesign.md` 的剩余有效信息并入 [menu-page-management-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/menu-page-management-design.md)，补齐了“新空间初始化不复制页面定义”和统一验收口径。
- 删除 `docs/scaffold-navigation-access-redesign.md`，避免与正式设计文档、实施现状文档形成第三份重复口径。
- 同步更新了 [PROJECT_FRAMEWORK.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/PROJECT_FRAMEWORK.md) 和 [docs/README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/README.md) 的文档入口与删除原则。

### 下次方向

- 继续评估菜单/页面链路的实现快照是否还需要独立文档；如果兼容边界完全稳定，可直接并入正式设计文档。
- 后续新增文档时，优先先判断它属于“正式设计 / 当前实现 / 固定清单 / 关键里程碑”中的哪一类，不再新增摘要型重复文档。

- 未验证项：本次仅做 Markdown 文档收口，未执行代码测试。

## 2026-03-31 Markdown 文档结构收口-第三轮

### 本次改动

- 将 `docs/menu-page-management-implementation-plan.md` 中仍有价值的“当前实现落点、兼容边界、最低验证矩阵”并入 [menu-page-management-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/menu-page-management-design.md)。
- 删除 `docs/menu-page-management-implementation-plan.md`，把菜单/页面链路从“正式设计 + 实现快照”继续压缩为单一主文档，避免同一主题继续分散在两份文档里。
- 同步更新了 [PROJECT_FRAMEWORK.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/PROJECT_FRAMEWORK.md)、[docs/README.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/README.md) 和本日志中的旧引用。

### 下次方向

- 继续观察 [space-host-architecture-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/space-host-architecture-design.md) 是否存在过多未来态描述；如果与当前单域基线偏离过大，可继续拆分“当前约束”和“远期演进”。
- 后续新增文档时，优先先补正式设计文档中的“当前实现落点 / 兼容边界 / 验收口径”，避免再单独创建实现快照型文档。

- 未验证项：本次仅做 Markdown 文档收口，未执行代码测试。

## 2026-03-31 Markdown 文档结构收口-第四轮

### 本次改动

- 重写了 [space-host-architecture-design.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/docs/space-host-architecture-design.md)，删除过多的远期拆分方案、分阶段计划和未来字段推演，只保留当前可执行约束、已落地底座、运行规则与允许的演进方向。
- 将“空间是运行时模型，Host 是可选部署映射”的边界明确成正式结论，避免后续再把子域设计当成当前实施要求。
- 本轮未再新增或删除文档文件，但显著压缩了单篇文档体量，使 `docs/` 更接近“当前规则手册”而不是“架构脑暴记录”。

### 下次方向

- 如果还要继续收口，下一步应转向 `docs/change-log.md` 本身，按“同日重复主题合并”的原则压缩历史记录。
- 后续新写专题文档时，默认只写“当前结论、稳定边界、验收口径”，把远期方案控制在必要的 1 到 2 段内。

- 未验证项：本次仅做 Markdown 文档收口，未执行代码测试。

## 2026-03-31 页面空间多选全面收口

### 本次改动

- 页面、逻辑分组、普通分组三类表单统一为多选语义，并新增 `全空间可见` 选项。
- 后端 `page` 模块兼容 `space_keys` 写入，空数组视为全空间可见，不再误回退到默认空间。
- 前端请求类型补齐了 `space_keys` 字段，三套表单已统一到同一份提交模型。

### 下次方向

- 如果确认所有独立页都默认全空间可见，可进一步把多选控件折叠成开关，减少录入成本。
- 页面管理页的表格可以继续收缩，把少用列下放到详情区。

- 未验证项：前端联调未做完整回归，仅完成后端 `go test ./internal/modules/system/page/...`。

## 2026-03-31 删除结果页与异常页菜单种子

### 本次改动

- 从默认菜单种子中移除了 `Result`、`Exception` 及其子菜单，不再把结果页和异常页作为菜单初始化。
- 在迁移代码中新增删除迁移，确保历史库里的结果页与异常页菜单树会被清掉，不会再次回灌。
- 前端静态路由仍保留 `403/404/500` 和结果页组件，页面本体继续可用，但不再依赖菜单配置。

### 下次方向

- 如果后续发现页面管理列表里还残留这类系统页，可继续从页面注册表层面做二次清理。
- 需要时再把系统级路由与业务菜单的边界说明补进专题文档，避免后续再次混入菜单种子。

- 未验证项：未执行全量迁移回归，仅完成 `go test ./internal/pkg/permissionseed/...` 与 `go test ./cmd/migrate`。

## 2026-03-31 整改收口-菜单与全局页联调

### 本次改动

- 完成菜单删除弹窗的工程化修复：纠正 `menu-delete-dialog` 模板乱码和语法错误，补齐三种删除策略（含“子菜单提到指定父菜单”）的说明与确认联动，新增“删除”手工确认约束，删除模式对齐宽度，并默认将“删除当前及全部子菜单”下沉到末位。
- 前端与后端联动方面，保持已完成的菜单删除预览/执行策略链路（含子树迁移、权限与路由映射清理）可用状态，并通过 `go test ./internal/modules/system/menu/... ./internal/modules/system/navigation/...` 与 `npm run build` 验证。
- 收敛了全局页/结果异常页等链路的绑定清洗逻辑：全局页在后端保存与查询时统一清空 `parent_menu_id / parent_page_key / active_menu_path`，并增加历史数据清理迁移。
- 头像入口跳转和用户中心路径使用统一 resolver，修复了部分 `/user-center` 直达和菜单跳转到错误路径的问题。

### 下次方向

- 继续逐页验证“菜单删除预览数值”和实际执行结果一致性，补一组端到端删除场景回归（尤其是有页面挂载与权限关联的子树）。
- 对现有前端菜单/页面文案中的历史编码片段做一次统一扫描清理，建议逐文件比对，优先处理影响操作按钮与异常页展示的内容。
- 后续如有空闲，增加 UI 自动化用例覆盖“提到指定父菜单”的树选择与降级清算路径，降低回归风险。

## 2026-03-31 数据库残留清理与回归补强

### 本次改动

- 直接连接 PostgreSQL 容器核验全局页残留，确认 `ui_pages` 中 `page_type='global'` 的记录已无 `parent_menu_id / parent_page_key / active_menu_path` 残留。
- 清理菜单表中遗留的 `Result / Exception` 菜单树，总计删除 59 条残留节点，删除前未发现页面挂载或权限关系表引用。
- 补充了菜单删除路径重映射与全局页绑定归零的单测，并通过 `go test ./internal/modules/system/menu/...`、`go test ./internal/modules/system/page/...`、`go test ./internal/modules/system/navigation/...` 验证。

### 下次方向

- 如果后续继续收口数据库历史数据，优先按“菜单树残留、页面挂载残留、权限关系残留”三类顺序核验，避免只删主表不清子树。
- 需要时再把这次数据库清理结果回写到专题文档，减少后续重复排查成本。
## 2026-03-31 访问链路测试页接入（导航与页面）

### 本次改动

- 新增后端访问链路接口：`GET /api/v1/pages/access-trace`，基于现有运行时权限上下文输出用户、角色、可见菜单与页面可见性判定。
- 新增迁移任务 `20260331_access_trace_navigation_seed`：自动创建/修正 `AccessTrace` 菜单与 `system.access_trace.manage` 页面，并刷新访问快照，避免手工二次配置。
- 前端新增“访问链路测试”页面（`/system/access-trace`），支持按用户、团队、页面维度测试并展示链路结果。

### 下次方向

- 建议补一组后端接口测试，覆盖“无团队上下文/有团队上下文/页面不存在”三类场景，稳定返回语义。
- 建议补前端 E2E 用例验证菜单迁移后入口可达与链路结果展示。
- 未验证项：前端全量 `tsc` 当前受仓库内既有缺失文件影响未通过，本次仅完成后端模块编译验证。

## 2026-03-31 访问链路测试团队角色接口修正

### 本次改动

- 新增后端接口 `GET /api/v1/tenants/:id/roles`，允许具备 `tenant.manage` 权限的管理员按指定团队获取可分配角色，返回基础团队角色和该团队自定义角色。
- 访问链路测试页改为在选中团队后调用指定团队角色接口，不再错误依赖 `my-team/roles` 的当前登录团队上下文。
- 已完成 `go test ./internal/modules/system/tenant/...`、`pnpm exec vue-tsc --noEmit` 验证。

### 下次方向

- 建议继续补一个后端接口测试，覆盖“团队不存在”“无权限”“团队含自定义角色”三类场景，避免后续接口回退。
- 如果后续还有别的后台页需要管理员跨团队查看角色，也统一复用这个新接口，不再走 `my-team/*` 语义。

## 2026-03-31 API 注册机制补漏

### 本次改动

- 调整 API 注册机制：带 `RouteMeta` 的受管接口在没有固定 route code 时，自动派生稳定 code，不再因为忘记维护固定码表而被识别成“未注册 API”。
- 补充 `apiregistry` 单测，覆盖固定码、显式码和自动派生码三类场景，确保后续新增接口不会再次静默漏注册。
- 已完成 `go test ./internal/pkg/apiregistry/... ./internal/modules/system/apiendpoint/...` 验证。

### 下次方向

- 如果还要继续压缩 API 管理页里的“未绑定权限键”噪音，建议给 JWT/自服务接口增加“无需权限键”的显式标识，而不是一律当成缺陷。
- 可继续补一轮全量路由审计，把确实需要独立权限键的接口单独列出来收口，例如媒体类接口是否需要正式纳入权限体系。

## 2026-03-31 API 管理权限结构审计增强

### 本次改动

- 为 API 管理补充权限结构诊断：接口现在会明确区分 `公开接口`、`登录态全局接口`、`登录态自服务接口`、`开放 API Key 接口`、`单权限接口`、`多权限共享`、`跨上下文共享`，不再把“无权限键”一律当成缺陷。
- 新增后端筛选能力，支持按权限结构过滤 API；`message.manage + team.message.manage` 这类消息接口会被标记为“跨上下文共享”，用于表达平台与团队共用同一接口而非错误重复。
- 前端 API 管理页新增权限结构列、权限结构筛选和概览指标（无权限键、共享接口、跨上下文共享），并通过 `go test ./internal/modules/system/apiendpoint/... ./internal/modules/system/user/... ./internal/pkg/apiregistry/...` 与 `pnpm exec vue-tsc --noEmit` 验证。

### 下次方向

- 建议继续审计仍属于 `登录态自服务接口` 但是否应升级为正式权限能力的接口，优先确认媒体类和其他通用工具类接口。
- 如需进一步收口权限设计，可新增“权限键未被页面/功能包/API 消费”的反向审计，把真正冗余的权限键筛出来处理。

## 2026-03-31 功能权限消费审计增强

### 本次改动

- 为功能权限列表补充消费审计模型，后端现可聚合每个权限键被 `API / 页面 / 功能包` 消费的数量，并输出 `未被消费`、`仅 API`、`仅页面`、`仅功能包`、`多方复用` 等使用结构。
- 新增重复判定能力：按权限语义族自动识别 `跨上下文镜像` 与 `疑似重复`；像 `message.manage` / `team.message.manage` 这类会明确标成镜像权限，而不是简单重复。
- 前端“功能权限”页新增审计摘要、消费结构列、重复检查列，以及 `消费情况 / 重复判定` 搜索筛选；同时补齐 `navigation` 测试桩缺失的 `GetAccessTrace`，并通过 `go test ./...`、`pnpm exec vue-tsc --noEmit` 验证。

### 下次方向

- 建议继续补“前端静态按钮权限引用”扫描，这样能把 `v-action` 等代码层直接消费也纳入权限键审计，形成比当前数据库侧更完整的全景图。
- 若后续要做权限收缩，可优先处理“未被消费 + 非内置”的权限键，再逐步复核同上下文疑似重复键，避免一次性误删镜像权限。

## 2026-03-31 功能权限安全清理与列表布局修复

### 本次改动

- 新增后端安全清理能力：`POST /api/v1/permission-actions/cleanup-unused` 只删除“未被 API、页面、功能包消费”且“非内置”的功能权限，避免把镜像权限或仍在使用的键误删。
- 前端“功能权限”页新增“清理未消费自定义权限”按钮，并在清理后自动刷新审计摘要与列表。
- 修复功能权限页表格布局：卡片改为 `flex + min-height + overflow` 结构，恢复分页可见和列表超宽时的滚动体验；已通过 `go test ./...`、`pnpm exec vue-tsc --noEmit` 验证。

### 下次方向

- 如果要继续真正收口数据，建议下一步把“疑似重复”里的同上下文权限做逐项复核，确认是否合并、重命名或补充职责说明。
- 可继续补一层“删除预览 / 清理候选”展示，先给出待删键列表和影响说明，再执行清理，减少批量操作顾虑。

## 2026-03-31 权限键与 API 注册全量修复

### 本次改动

- 新增迁移 `20260331_permission_api_registry_full_repair`，对历史页面键误入权限键的残留数据做归一化处理，重点将 `system.message.manage` 这类脏权限键迁回正式权限键并清理残留记录。
- 补齐正式 API 分类种子 `message`、`navigation`，同时将默认 API 分类、权限分组、权限键从“仅创建缺失项”升级为“按种子持续回写元数据”，修复了 `message.manage` 被错误标成团队上下文等历史脏数据。
- 修正 API 注册同步逻辑：`syncRoutesInternal` 现在会对已有受管接口执行真正的回写更新，而不是只插入新接口；配合消息模块和运行时导航模块的路由元数据修正，`/api/v1/messages/*` 统一归类到 `message`，`/api/v1/runtime/navigation` 归类到 `navigation`。
- 为 `api_endpoint_permission_bindings` 增加去重迁移和唯一索引，避免后续重复插入同一 `(endpoint_code, permission_key)` 绑定；已执行 `go test ./...` 与 `go run ./cmd/migrate` 验证，并回查确认消息权限键、消息 API 分类和消息接口绑定均已归一化。

### 下次方向

- 建议继续把其余历史接口分类做一轮同样的“代码元数据 vs 数据库现状”对账，尤其是全局运行时接口和跨上下文共享接口，避免再出现代码已修正但注册表未回写的情况。
- 如果后续要开放自定义业务权限键，建议补一层“上下文推导/校验”规则说明，明确 `platform / team / common` 三种语义的创建边界，减少人工录入歧义。

## 2026-03-31 权限上下文校验与默认空间回收

### 本次改动

- 在功能权限创建/更新链路补充后端上下文校验：内置权限键强制使用规范上下文，自定义权限键按 `platform / team / common` 明确约束前缀和模块归属，避免再把平台键、团队键和通用业务键混写进库。
- 调整功能权限前端弹窗默认上下文为 `common`，并补充“平台/团队/通用”填写提示，减少录入时的语义歧义。
- 新增迁移 `20260331_restore_default_space_only`，清理历史 `ops` 演示菜单空间、其系统菜单克隆、相关备份与空间绑定；同时把 `init-demo` 默认空间改回 `default`，避免后续再自动生成“运营空间”副本。
- 已执行 `go test ./...`、`pnpm exec vue-tsc --noEmit`、`go run ./cmd/migrate`，并回查确认当前仅保留 `default` 菜单空间。

### 下次方向

- 建议继续把“自定义权限键”与“功能包上下文”联动校验补到功能包层，避免上下文合法但功能包归属仍混乱。
- 如果后续还要保留多空间能力，建议把 `ops` 这类演示空间改成显式命令参数开启，而不是任何默认初始化路径都可能生成。

## 2026-03-31 演示数据全量清理与注入路径移除

### 本次改动

- 移除迁移任务 `20260329_demo_messages_seed`，并删除 `migrate` 中对应的 `seedDemoMessages` 注入函数，避免后续迁移再次写入演示消息、演示模板和演示接收组。
- 删除命令入口 [backend/cmd/init-demo/main.go](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/backend/cmd/init-demo/main.go)，彻底移除演示账号、演示团队、演示消息的代码注入路径。
- 直接连接数据库执行脏数据清理，删除了演示消息链路与演示主体数据（消息投递 52、消息 35、接收组目标 96、接收组 2、模板 2、发送人 4、演示团队 1、演示用户 3），并清理 `ops` 空间残留（菜单空间 1、菜单 132）。
- 对有效数据再次做一致性检查：`feature_package_menus` 与 `ui_pages.parent_menu_id` 无活跃孤儿引用；当前仅保留 `default` 菜单空间。

### 下次方向

- 建议增加一个受保护的“数据治理脚本”目录，把本次清理 SQL 做成可重复执行的幂等任务，后续可直接复用，不再临时拼接脚本。
- 建议在 CI 增加“禁止演示种子回灌”静态检查（例如扫描 `demo.`、`_demo` 关键字），防止后续改动再次引入测试数据注入路径。

## 2026-04-01 全量权限治理与团队消息能力改造（一次性落地）

### 本次改动

- 团队消息发送链路完成边界修复：团队上下文的模板查询改为仅团队模板；团队发送页成员下拉统一只显示成员名；模板提示文案改为团队专属，并在配置加载时先重置选项避免上下文切换残留。
- 后端补齐治理接口主干：
  - 功能包新增“关系树”查询接口，返回树结构、循环依赖诊断、孤立基础包提示。
  - 功能权限新增“消费明细”接口，统一返回 API/页面/功能包/角色引用数据。
  - API 管理新增“未注册扫描配置”读写接口，配置落库到 `system_settings`（`api.unregistered_scan_config`）。
  - 权限刷新服务新增增量刷新统计能力，支持返回受影响包/角色/团队/用户数量与耗时。
- 功能包配置接口（组合关系、权限、菜单、团队分配、团队包分配、更新/删除）统一回传 `refresh_stats`，前端后续可直接消费增量刷新结果。
- 前端管理端增强：
  - 功能包管理新增“包关系树”入口与可视化展示。
  - 功能权限管理“消费结构”新增明细弹窗，可查看 API/页面/功能包/角色链路。
  - API 管理新增“扫描配置”弹窗（启停、频率、默认分类、默认权限键、无权限标记）。
- 已完成验证：`go test ./...`、`pnpm exec vue-tsc --noEmit` 全通过。

### 下次方向

- 继续补齐本次计划中尚未完成的深层能力：功能包版本历史与回滚、权限批量模板化操作、未注册 API 自动任务执行器（当前为配置落库阶段）。
- 建议把 `refresh_stats` 在前端保存成功提示中统一透出（角色/团队/用户与耗时），并补充“菜单/动作二次裁剪”提示文案，形成可追溯闭环。

## 2026-04-01 全量权限治理二阶段并入（审计/版本/模拟/影响预览）

### 本次改动

- 新增独立治理数据结构并完成迁移接入：`risk_operation_audits`、`feature_package_versions`、`permission_batch_templates`，已纳入 `AutoMigrate` 与索引创建。
- 功能包能力补齐：
  - 新增接口：影响预览、版本历史、版本回滚、最近变更查询。
  - 功能包写操作（更新、删包、配置基础包/功能范围/菜单/团队、团队包分配）统一接入风险审计落库，并在关键配置变更后自动写入版本快照。
  - 回滚流程复用同一保存链路并返回 `refresh_stats`。
- 功能权限治理补齐：
  - 新增接口：影响预览、批量治理、批量模板保存/列表、风险审计查询。
  - 删除/更新/批量变更接入风险审计落库。
- 权限模拟独立入口：
  - 后端新增 `/api/v1/pages/permission-simulator`（复用既有访问链路口径）。
  - 前端新增独立页面组件 `/system/permission-simulator`（系统菜单独立入口），并支持调用模拟接口。
  - 菜单种子与命名迁移新增“权限模拟”入口，默认空间可直接落地。
- 前端交互治理：
  - 功能包相关弹窗保存成功提示统一为“本次增量刷新：角色 X、团队 Y、用户 Z、耗时 N ms”。
  - 功能包删除与权限停用/删除接入“影响预览 + 二次确认”。
  - API 管理“扫描配置”升级为实验态文案，增加禁用态说明，避免误判为已自动调度。
- 已完成验证：`go test ./...`、`pnpm exec vue-tsc --noEmit`、`go run ./cmd/migrate/main.go`。

### 下次方向

- 前端补齐“功能包版本历史抽屉+可视化差异+回滚入口”和“权限批量治理操作面板”（当前后端接口已就绪）。
- 基于现有 `risk_operation_audits` 增加统一“治理审计页”，支持按对象类型聚合检索，形成跨模块追溯闭环。

## 2026-04-01 全站前端页面现代化重构（公共底座收口）

### 本次改动

- 统一全局设计底座：补充页面底色、表面层、文本层级、阴影、圆角与状态色 token，并收口 `page-content`、`art-table-card`、`art-card` 等全局容器样式。
- 强化公共组件视觉一致性：
  - `AdminWorkspaceHero` 改为更明确的工作台头部层级。
  - `ArtSearchBar`、`ArtTableHeader`、`ArtTable`、`ArtStatsCard`、`ArtForm` 全面重做间距、边界感、按钮优先级和表格/筛选/统计的层次。
  - `AppContextBadge`、`PermissionSummaryTags`、`FeaturePackageGrantPreview`、`PermissionSourcePanels`、`ArtResultPage` 做了统一视觉收口。
- 重点页面完成统一收口：
  - 消息发送页、角色页、团队角色页、登录页、全局页面容器、结果页等完成视觉与交互基线升级。
  - 主要列表页因为公共组件统一而同步受益，减少了页面之间的风格漂移。
- 已完成验证：`pnpm exec vue-tsc --noEmit` 通过。

### 下次方向

- 继续补齐未逐页收口的系统页头与页面骨架，优先处理菜单、页面、团队、消息中心等仍有局部旧样式的页面。
- 建议再做一次真实浏览器手工回归，重点确认 1440 / 1600 / 1920 宽屏下的表格密度、筛选折叠和消息预览区域是否达到预期。

## 2026-04-01 全站视觉底座轻量化收口

### 本次改动

- 收紧全局视觉变量：页面底色、表面层、边框、阴影和高亮状态改成更浅、更克制的企业后台风格，整体从“厚卡片、偏亮边框”调整为“浅底色、轻阴影、柔边界”。
- 统一布局壳层观感：`page-content`、`art-table-card`、`AdminWorkspaceHero`、`ArtHeaderBar`、侧边栏菜单都改成更平滑的圆角、阴影和留白节奏，减少装饰性渐变，向参考图的简约风格靠拢。
- 修复页面容器裁切风险：把 `page-content` 与全局页面壳层统一为 `overflow: visible` / `min-height` 方案，避免长列表和底部区域再次被父容器截断。
- 已完成代码检查：`pnpm --dir frontend exec vue-tsc --noEmit` 通过；本轮未重新完成浏览器回归，因为浏览器会话在验证阶段已失效，后续需要重新打开后再看 1440 / 1600 / 1920 三档效果。

### 下次方向

- 建议下一步继续把菜单管理、页面管理、团队、消息中心这几类仍偏旧的页面壳再统一一层，重点收掉页头、工具栏和表格密度上的细碎差异。
- 浏览器会话恢复后，建议重新做一次真实回归，确认新底座下的侧栏、顶栏、统计卡和长列表滚动体验是否和参考图一致。

## 2026-04-01 旧壳页面继续收口

### 本次改动

- 继续收紧菜单、页面、团队、团队成员和消息工作台的页面壳，统一按钮、表头、表格 hover、辅助文案和卡片边界的视觉语义。
- 菜单管理页、受管页面页、团队页、团队成员页、消息工作台导航都改成更浅的表面层和更克制的高亮方式，减少旧式灰块、厚边框和强蓝底。
- 团队成员页把“快速添加”和“成员列表”改成更明确的分区卡，空状态、加载态和列表容器更接近新的企业后台风格。
- 已完成 `pnpm --dir frontend exec vue-tsc --noEmit`，当前未重新做浏览器回归，因为此前浏览器会话已关闭，需重新打开后再看真实视觉结果。

### 下次方向

- 继续清理剩余仍带局部旧风格的系统页，优先看菜单管理中的分组/批量工具、页面管理中的树形列表层级，以及消息发送页的细节间距。
- 浏览器会话恢复后，建议再做一次 1440 / 1600 / 1920 的真机回归，确认新底座下的页面密度、工具栏和表格滚动条都稳定。

## 2026-04-01 前端设计从 frontend-copy 恢复

### 本次改动

- 按 `frontend-copy` 备份恢复了共享设计层，直接同步回当前前端的全局样式、布局底座、表单/表格公共组件、登录外观和主工作区样式。
- 对不能整文件覆盖的混合组件，仅回退了样式层，保留了当前项目已经接入的业务能力与交互逻辑，避免把新功能一起覆盖掉。
- 已完成 `pnpm --dir frontend exec vue-tsc --noEmit` 验证，当前恢复范围以“公共设计系统” 为主，业务页中没有备份对应物的局部定制样式未做整页回滚。

### 下次方向

- 如果还要继续做“完全视觉回退”，下一步应逐页处理当前业务页中的局部定制样式，例如菜单、页面、团队、消息治理页这些没有备份对应文件的页面。
- 浏览器回归建议重新补一轮，确认恢复后的共享设计层与当前业务页组合后没有出现视觉断层。

## 2026-04-01 业务页局部定制样式继续收口

### 本次改动

- 菜单管理页和受管页面页继续从页面私有壳层收回到公共组件：顶部摘要统一改为 `AdminWorkspaceHero`，表格工具区统一收回 `ArtTableHeader`，减少页面自造按钮组和自定义页头结构。
- 菜单管理页删除了私有搜索/刷新/全屏/列设置按钮实现，改为直接使用公共表格工具栏能力；页面页把原先嵌在表格头中的标题和统计摘要上移，工具区只保留页面治理相关的开关与提示。
- 已完成 `pnpm exec vue-tsc --noEmit` 验证；浏览器回归未完成，因为当前 Playwright 浏览器上下文已关闭，无法直接复用会话。

### 下次方向

- 继续把团队、消息治理、权限治理页里仍然偏私有的页头和局部工具条收回到现有库组件，进一步减少单页样式负担。
- 重新建立浏览器会话后，重点检查菜单管理和受管页面在 1440 / 1600 / 1920 下的页头层级、工具栏密度和表格滚动表现。
## 2026-04-01 业务页局部壳层继续收口

### 本次改动
- 团队成员页把私有标题块收回到 `AdminWorkspaceHero` 和 `ArtTableHeader`，快速添加与成员列表改成统一卡片节奏。
- 功能权限页把页头从表格卡内部移出，统一成和菜单、页面、团队一致的公共页头结构。
- 消息发送页与消息工作台导航继续收紧局部样式，保持编辑区、预览区和导航壳层使用同一套轻量卡片语言。
- 已执行 `pnpm exec vue-tsc --noEmit`，类型检查通过。

### 下次方向
- 继续处理团队功能包、消息中心和其余治理页里仍保留的私有工具条与分区壳。
- 当前未执行真实浏览器回归，后续需补一轮 1440 / 1600 / 1920 分辨率下的视觉与滚动验证。

## 2026-04-01 去掉权限模拟入口并收口迁移

### 本次改动
- 删除了 `权限模拟` 的前端包装页、后端独立 API/路由和菜单种子，只保留 `访问链路测试` 作为唯一入口，避免同类能力重复出现在菜单里。
- 补了迁移清理逻辑，重放历史迁移或新服务器部署时会主动清除旧的 `permission-simulator` 菜单、页面和关联表残留，降低历史数据复现概率。
- 已执行 `go test ./cmd/migrate ./internal/modules/system/page/... ./internal/pkg/permissionseed/...` 和 `npm run build`，后端与前端均通过。

### 下次方向
- 如果还要继续收口权限链路页面，下一步可以检查是否还有旧文档、菜单截图或种子说明保留 `权限模拟` 字样。
- 后续部署前建议再跑一轮数据库残留核验，确认旧服务器升级后的菜单树和页面表都只保留 `访问链路测试` 一条链路。
## 2026-04-01 共享分页条与列表分页统一

### 本次改动
- 新增共享分页组件 `WorkspacePagination`，统一列表页与抽屉/弹窗列表的分页样式和布局。
- 团队成员页、团队成员抽屉、权限分组管理、角色/团队角色/用户/团队功能包弹窗、功能包基础包与开通团队弹窗、菜单备份列表、未注册受管页、消息模板与消息记录页已接入共享分页。
- 已执行 `pnpm exec vue-tsc --noEmit`，类型检查通过。

### 下次方向
- 继续补齐剩余仍使用原生列表但尚未接分页的治理型明细区域，优先处理 API 管理弹窗和个别诊断抽屉。
- 当前未执行真实浏览器回归，后续需确认分页条在抽屉、弹窗和 1440 / 1600 / 1920 宽度下的换行与滚动表现。

## 2026-04-01 修复空邮箱创建用户失败

### 本次改动
- 用户创建和更新链路增加了邮箱 `trim + 非空查重`，避免空字符串直接写库后触发 `idx_users_email` 唯一冲突。
- 新增了 `邮箱已存在` 的错误码，并把后台用户创建、前台注册的邮箱冲突提示拆开，避免统一报成用户名重复。
- 迁移里补了 `users.email` 的部分唯一索引收口，改为仅对非空邮箱生效，保证新服务器重放迁移时不会再把空邮箱当作唯一值。
- 已验证 `go test ./cmd/migrate ./internal/api/errcode ./internal/modules/system/user ./internal/modules/system/auth` 通过。

### 下次方向
- 如果历史库里存在非空邮箱重复，后续还要补一次数据清理，否则新索引创建会失败。
- 还可以继续检查其它“空字符串进入唯一索引”的字段，统一改成部分唯一索引或空值归一。

## 2026-04-01 全站分页与响应式规范一次性收口

### 本次改动

- 统一搜索区底层规范：`ArtSearchBar` 默认参数改为 `label-position="top" / span=8 / gutter=16`，并固定按钮顺序为“查询 -> 重置 -> 展开/收起”。
- 补齐非 `ArtTable` 列表分页：消息发送人、接收组、团队成员抽屉、菜单分组抽屉、角色权限选择、用户权限测试、访问链路角色/页面明细、工作台新用户列表等全部接入 `WorkspacePagination` 与本地切片模型。
- 收敛功能包与访问链路页面响应式布局：去除固定宽度筛选栅格，统计卡改为稳定自适应网格，确保 1920/1600/1440 下密度稳定且分页区可见。
- 全量扫描并清零“原生 `ElTable` 无分页”缺口；新增规则正式写入 [FRONTEND_GUIDELINE.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/FRONTEND_GUIDELINE.md)“分页与响应式规范（强约束）”章节。
- 已验证：`pnpm --dir frontend exec vue-tsc --noEmit` 通过；本次未执行浏览器分辨率 E2E 实测。

### 下次方向

- 继续按新规范审计后续新增页面，禁止再引入固定宽筛选与无分页长列表。
- 真实浏览器回归需补一轮 1920/1600/1440/1024/768 分辨率验证，重点确认抽屉/弹窗分页条可见性与消息治理卡片区稳定性。

## 2026-04-01 搜索区与概览面板间距收口

### 本次改动

- 将搜索区与概览 Hero 的纵向节奏正式写入 [FRONTEND_GUIDELINE.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/FRONTEND_GUIDELINE.md)，统一规定为“搜索区 -> 概览面板 -> 表格卡 -> 分页区”默认 `12px` 节奏，卡片内部模块默认 `16px` 节奏。
- 将功能权限、功能包等页面中明显的 `10px` 面板间距收敛到 `12px`，避免同类页面出现“搜索区、统计区、操作区”贴得过近的问题。
- 保持用户管理、功能包管理等页面的外部 Hero 结构不变，仅调整上下间距和面板节奏，不影响数据或交互逻辑。

### 下次方向

- 继续按规范排查其余系统页与消息页的局部 `10px` 间距残留，优先处理顶部面板和表格卡之间的视觉节奏。
- 如后续还有新增概览面板，默认直接按 `12px / 16px` 套用，避免再出现局部覆盖。

## 2026-04-01 概览面板统一为分割线样式

### 本次改动

- 将概览面板（统计区）统一收口为 `AdminWorkspaceHero` 的“标题/描述在上，分割线在中，指标区在下”结构，避免再使用一组独立小卡拼统计区。
- 清理团队角色与权限页残留的旧统计标签行，统一改为共享概览面板承载指标，减少页面风格分叉。
- 同步更新 [FRONTEND_GUIDELINE.md](/C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend/docs/legacy/root/FRONTEND_GUIDELINE.md) 约束，明确所有概览面板必须复用 `AdminWorkspaceHero`。

### 下次方向

- 继续回扫消息、团队、系统管理页，确保没有页面级自定义统计条回流。
- 新增概览区时直接复用共享 Hero，不再手写小卡拼接型统计块。

## 2026-04-01 面板间距统一为 12px

### 本次改动

- 继续收口页面主结构与常用弹窗/抽屉中的块级间距，把仍残留的 `10px` 统一提升为 `12px`，避免搜索区、概览面板、工具栏、表格区之间出现视觉断层。
- 复查并清零前端视图中的页面级 `10px` 间距残留，保留的差异仅作为局部控件微调，不再作为面板节奏使用。
- 已验证：`pnpm --dir frontend exec vue-tsc --noEmit` 通过。

### 下次方向

- 后续新增页面默认直接使用 `12px / 16px` 节奏，避免再次引入新的 10px 页面级间距。
- 若某些局部控件还需更细腻的微调，优先放到组件内部，不要影响页面主节奏。

## 2026-04-01 概览指标改为紧凑包裹

### 本次改动

- 将 `AdminWorkspaceHero` 的概览指标区从均分撑满改为左对齐紧凑包裹，避免一排指标被拉得过散。
- 同步更新前端规范，明确概览面板指标区必须紧凑包裹，不得占满整行两端。
- 该改动会同步影响所有使用共享概览面板的页面，属于全局视觉收口。

### 下次方向

- 后续新增统计指标时默认按紧凑包裹布局接入，不要再单独写均分网格。
- 如需更高信息密度，优先调整指标宽度与换行策略，而不是继续拉伸整行。

## 2026-04-01 菜单页搜索栏与概览间距收口

### 本次改动

- 修复菜单管理页搜索栏与概览面板之间的页面级间距异常，将 Hero 顶部间距收敛到 `12px`，并同步把概览子行间距统一到 `12px`。
- 在前端规范里补充主治理页约束，明确菜单管理、页面管理、功能权限等页的搜索栏与概览区默认保持 `12px`。
- 这次修复只调整页面主节奏，不改数据和交互逻辑。

### 下次方向

- 后续继续回扫其它治理页，避免新的页面再写出不同的搜索栏到概览区间距。
- 若需要更细的视觉层次，优先通过概览面板内部结构调整，而不是增加页面级空隙。

## 2026-04-01 搜索栏自带间距与 Hero 顶边距冲突收口

### 本次改动

- 发现菜单管理页的搜索栏间距实际来自 `ArtSearchBar` 自带底部空隙，页面又额外给 Hero 补了顶部空隙，导致视觉间距翻倍。
- 已将菜单页 Hero 的顶部外边距显式收回为 `0`，保留搜索栏自身的默认间距，避免重复叠加。
- 同步更新前端规范，明确“搜索栏组件自带底部间距时，页面级 Hero / 概览面板不得再重复补顶边距”。

### 下次方向

- 其他治理页如果也同时使用搜索栏和概览面板，需要先确认搜索组件本身是否带间距，再决定是否补页面级 margin。
- 后续新增同类页面时，优先沿用共享规范，不要再在页面级重复叠加空隙。

## 2026-04-01 菜单页顶部栈统一间距

### 本次改动

- 将菜单管理页的“搜索栏 + 概览面板”包进统一的顶部栈容器，由页面容器直接控制 `12px` 间距，避免靠单个组件 margin 叠加导致的视觉错位。
- 这次调整是针对页面结构本身，而不是继续微调 Hero 或搜索栏组件内部样式。
- 菜单页主节奏现在与其它治理页一致，搜索区和统计区之间会稳定保留间隔。

### 下次方向

- 其它同时出现搜索栏和概览面板的治理页，也建议逐页确认是否需要独立的顶部栈容器。
- 新增页面时优先采用“页面壳层 gap 控制间距”的方式，不再依赖单组件 margin 拼接。

## 2026-04-01 搜索栏与概览面板统一页面级顶部栈

### 本次改动

- 全量回扫带有“搜索栏 + 概览面板”的治理页，统一补上页面级顶部栈容器，避免搜索区和统计区间距由单组件 margin 互相叠加。
- 将 `page-top-stack` 抽成全局页面壳层工具类，后续页面只要有这类组合就直接复用，不再各自写 margin。
- 同步更新前端规范，明确“搜索栏 + 概览面板”必须由页面级顶部栈控制间距。

### 下次方向

- 新增治理页优先直接套用 `page-top-stack`，不要再手写独立 margin。
- 若后续还有视觉贴边或空隙过大，优先检查页面级栈是否缺失，再看组件内部样式。

## 2026-04-01 顶部栈与表格卡双重间距清理

### 本次改动

- 确认用户页等页面的异常来自双重间距：`page-top-stack` 已负责搜索区到概览区的节奏，但页面里的 `art-table-card` 仍额外根据 `showSearchBar` 补 `marginTop`，导致展开时空隙过大、收起时节奏不稳定。
- 已将这类页面的手工 `marginTop` 全部移除，并在全局样式里明确：`page-top-stack` 自带底部间距，且其后的 `art-table-card` 不再额外补顶边距。
- 已复查前端视图，相关 `showSearchBar ? '12px' : '0'` 残留为 0；`pnpm --dir frontend exec vue-tsc --noEmit` 通过。

### 下次方向

- 后续新增“搜索栏 + 概览 + 表格”页面时，直接套 `page-top-stack + art-table-card` 结构，不要再在页面级写条件 margin。
- 如果再出现展开/收起后节奏变化，优先检查是否重新引入了组件外的补偿间距。

## 2026-04-01 主面板节奏全局收口

### 本次改动

- 从全局壳层统一收口主面板间距，明确两种合法结构：`page-top-stack + art-table-card`，或 `AdminWorkspaceHero + art-table-card`。
- 为“概览面板直连表格卡”的页面补上全局去重规则，避免 `Hero` 底部 `12px` 与 `art-table-card` 顶部 `12px` 叠加成 `24px`。
- 继续复查系统、团队、消息、工作台页面；当前残留的直连结构已由全局规则覆盖，`pnpm --dir frontend exec vue-tsc --noEmit` 通过。

### 下次方向

- 后续新增页面优先直接使用现有两套合法结构，不要再引入第三种页面节奏写法。
- 如果还出现“几个主面板不对齐”，优先先查全局壳层和页面结构，再看局部组件。

## 2026-04-01 主面板节奏定板与 API 管理页修复

### 本次改动

- 将全站主面板节奏正式定板为 `10px`，范围只覆盖页面壳层上的顶部标签、搜索区、概览面板、告警条、主表格卡与主内容板，不再把组件内部细粒度布局混入同一口径。
- 在全局样式中收口 `art-full-height`、`page-top-stack`、`menu-top-stack`、`art-table-card` 与 `AdminWorkspaceHero` 的组合关系，移除主面板级 `12px / 16px` 叠加来源；同时把用户、功能权限、功能包、菜单空间、消息中心与消息治理页的页面级残留 margin 一并清掉，并把默认展开的搜索栏统一改回默认收起。
- 重写 API 管理页右侧主区模板结构，恢复为稳定的 `page-top-stack -> ElCard -> AdminWorkspaceHero + ArtTableHeader + ArtTable` 结构，消除近期模板层级紊乱带来的页面打不开风险。
- 重新整理 `FRONTEND_GUIDELINE.md`，将其改为前端定板文档，明确主面板标准 `10px`、合法页面结构、搜索区标准、概览面板标准、分页与响应式标准以及禁止项。
- 已执行 `pnpm --dir frontend exec vue-tsc --noEmit`；本次未做完整浏览器逐页回归，未验证项是 API 管理页与所有消息治理页的真实 UI 打开结果。

### 下次方向

- 补一轮真实浏览器回归，重点验证 API 管理页、菜单空间、消息中心和消息治理页在 `1440 / 1600 / 1920` 下的实际间距、分页和搜索展开表现。
- 若后续还发现主面板不对齐，优先先查页面是否偏离三种合法结构，而不是继续追加单页 margin 补丁。

## 2026-04-01 分页组件回退为 frontend-copy 口径

### 本次改动

- 将 `ArtTable` 的分页实现按 `frontend-copy` 备份回退，恢复为原有的容器高度与分页间距写法，不再使用最近加上的自动流式布局逻辑。
- 分页显示、布局对齐和样式节奏全部以备份实现为准，保留 `frontend-copy` 里的默认分页写法，避免继续在组件层追逐新的间距策略。
- 已执行 `pnpm --dir frontend exec vue-tsc --noEmit`，当前类型检查通过。

### 下次方向

- 如果分页仍显得太贴近数据，应优先从页面壳层、卡片内边距或表格上方留白调整，而不是继续改 `ArtTable` 本身。
- 后续同类分页问题统一先对照 `frontend-copy`，保持分页组件实现风格一致。

## 2026-04-01 概览面板改为单行指标与下沉操作区

### 本次改动

- 收紧共享概览组件 `AdminWorkspaceHero` 的结构：标题与说明保留在左侧，统计信息尽量压缩到同一行右侧，避免统计区横向铺散。
- 若页面存在操作按钮，统一下沉到第二行并通过分割线与主信息区隔开，按钮从第二行左侧开始显示，减少顶部区域拥挤感。
- 这次调整会同步影响所有复用该共享概览组件的页面，属于全局视觉收口。

### 下次方向

- 后续新增统计面板优先沿用这套结构，不再把统计卡分散成多行。
- 若个别页面按钮过多导致第二行拥挤，再在按钮组内部做局部换行，不要破坏概览组件的整体节奏。

## 2026-04-01 快捷应用页回退公共面板壳

### 本次改动

- 将快捷应用管理页里自定义的外层壳、左右面板和摘要块回退为公共卡片壳，统一使用系统已有的面板圆角与边框风格。
- 去掉快捷应用页私有的厚边框、渐变背景和过强圆角，让页面结构更贴近当前公共面板体系，减少“浮起来”的视觉断层。
- 这次调整只改快捷应用页的局部视觉壳层，不影响快捷入口数据、抽屉编辑和保存逻辑。

### 下次方向

- 其他仍保留私有面板壳的页面，优先回退到公共卡片组件或公共面板样式，不要继续单页自造圆角和边框。
- 如果后续还想进一步统一圆角，可以把公共卡片壳的圆角值再做一轮全局定标，但不要在单页覆盖。

## 2026-04-01 快捷入口外链修复与宽度配置收口

### 本次改动

- 修复顶部快捷入口的外链点击，改为在点击事件里同步执行 `window.open(..., '_blank', 'noopener,noreferrer')`，避免外链被异步链路或弹窗拦截导致无响应。
- 去掉快捷应用管理页里的“最小显示宽度”配置项与对应展示指标，页面不再允许单独配置该值。
- 将快捷入口显示阈值统一固定为 `1450`，并同步收口默认配置、接口归一化和 store 内部默认值，保持前后口径一致。

### 下次方向

- 如果后续还要继续收口快捷入口，优先处理顶部弹出层的布局密度和链接 hover 反馈，不再扩展新的可配置项。
- 若线上历史配置里仍带旧 `minWidth`，当前实现会忽略该值；后续如需彻底清理，可再补一次服务端配置归档。

## 2026-04-01 快捷入口全链路打通与配置迁移入库

### 本次改动

- 补齐快捷入口后端配置升级链路：在 `fast-enter` 配置归一化时自动合并缺失的默认项，确保旧库配置会补齐“项目文档”“技术支持”两个外链入口，并统一将宽度阈值收口到 `1450`。
- 新增命名迁移 `20260401_fast_enter_config_seed`，执行时会读取现有 `ui.fast_enter` 配置、按当前默认口径归一化后重新写回数据库，避免旧库保留 1200 宽度和缺失外链项。
- 真实验证了快捷入口接口读写：管理员登录后 `GET /api/v1/system/fast-enter` 已返回补齐后的配置，`PUT /api/v1/system/fast-enter` 可成功写入；同时做了“新增临时链接 -> 回读确认 -> 恢复原配置”的往返验证，确认新增、删除、保存已真正落库。
- 已执行 `go run cmd/migrate/main.go`、`go test ./...`、`pnpm --dir frontend exec vue-tsc --noEmit`。

### 下次方向

- 继续做一轮真实浏览器回归，重点确认快捷入口顶部弹层在登录后的显示结果、外链点击行为以及管理页“新增/编辑/删除 -> 保存 -> 顶部回显”的完整交互。
- 若后续还要演进快捷入口配置，优先维持“默认种子 + 迁移归一化 + 前端草稿预览”这一条链路，不再引入前后端双份不一致的默认值。

## 2026-04-01 新增 Fluent 技能

### 本次改动

- 在全局技能目录新增 `fluent-react-v9` 与 `fluent2-frontend-style` 两个技能，分别覆盖 Fluent UI React v9 的工程实施，以及 Fluent 2 的视觉与交互约束。
- 为两个技能补齐中文 `SKILL.md`、`agents/openai.yaml` 和四份参考资料，内容聚焦后台与工作台场景，不绑定当前仓库结构。
- 已执行技能结构校验；本次未做真实子代理前向触发验证，未验证项是技能在多轮复杂对话下的实际触发效果。

### 下次方向

- 后续可继续补充更具体的后台页面配方，例如多步表单、审计页、监控页和设置中心。
- 若后面发现两个技能触发边界仍有重叠，可进一步收紧 frontmatter 描述和参考文件分工。

## 2026-04-01 Fluent React 技能补充 Storybook 索引

### 本次改动

- 为 `fluent-react-v9` 技能新增 Storybook 文档索引参考文件，收录 `Concepts -> Developer -> Quick Start` 及其相关开发者章节，并整理了主题、组件、工具、动效和迁移目录的使用顺序。
- 在技能正文与官方说明参考里补充 Storybook 导航入口，要求实现 Fluent UI React v9 时优先按文档树定位资料，不再依赖零散旧博客。
- 已重新执行技能结构校验；本次未逐页人工核对 Storybook 所有组件文档正文，未验证项是目录项未来变更后的同步维护。

### 下次方向

- 后续可继续把高频组件如 `DataGrid`、`Drawer`、`Nav`、`Tree`、`Field` 的后台落地经验沉淀成更具体的配方。
- 如果 Storybook 目录未来有明显调整，优先更新这份索引文件，再同步调整 `fluent-react-v9` 的正文入口。

## 2026-04-01 Fluent 技能进一步完善

### 本次改动

- 继续增强 `fluent-react-v9`，新增实现检查清单和高频组件后台配方，把 `FluentProvider`、`Nav`、`Tree`、`Toolbar`、`DataGrid`、`Drawer`、`Dialog`、`Field`、`MessageBar` 等组件的选型与页面组合规则收敛到参考文件。
- 继续增强 `fluent2-frontend-style`，新增模式与系统规则、布局与表层检查清单，把 `Handoffs`、`Wait UX`、`Motion`、`Accessibility`、`Design tokens` 与后台工作台页面结构串起来。
- 同步更新两个技能正文，把“先读什么、何时切到哪份参考、如何做实现和评审”写得更明确，降低技能只停留在资料索引层的风险。
- 已再次执行技能结构校验；本次仍未做真实子代理长对话前向验证，未验证项是技能在更复杂、多轮任务中的长期触发稳定性。

### 下次方向

- 可继续把 `DataGrid`、`Drawer`、`Nav`、`Tree`、`Field` 等组件分别沉淀成独立专题，附上更具体的后台代码模式。
- 若后续开始实际用这些技能生成页面，可以根据真实对话结果继续压缩或重写触发描述，进一步减少职责重叠。

## 2026-04-01 Fluent React 技能拆分高频组件专题

### 本次改动

- 为 `fluent-react-v9` 新增 `DataGrid`、`Drawer`、`Nav`、`Tree`、`Field` 五份独立专题参考，分别整理了官方 Storybook 文档中的组件定位、最佳实践、后台使用场景和常见误用。
- 同步更新技能正文、实现检查清单和高频组件配方，加入“何时深入读单组件专题”的入口，避免技能只停留在粗粒度建议。
- 这次专题重点吸收了官方文档中的关键判断：例如 `DataGrid` 是 `Table` 与 `useTableFeatures` 之上的高阶能力，`Drawer` 分为 overlay 与 inline，`Nav` 适合两级导航，`Tree` 适合层级数据，`Field` 负责标签与校验语义包装。
- 已再次执行技能结构校验；本次未做基于真实代码仓的页面生成演练，未验证项是这些专题在真实产出代码时的细节覆盖度。

### 下次方向

- 后续可继续把这些专题扩成带示意代码的参考文件，尤其是 `DataGrid + Drawer`、`Tree + Detail Pane`、`Field + Form Section` 等后台高频组合。
- 如果开始实际用技能生成 Fluent 页面，可根据产出质量继续补反例和“何时不要用该组件”的更细判断。

## 2026-04-01 Fluent React 技能补充代码模式

### 本次改动

- 为 `fluent-react-v9` 新增三份代码模式参考：工作区列表 + 抽屉、树 + 详情面板、分组表单，分别服务 `DataGrid + Drawer`、`Tree + Detail Pane`、`Field + Form Section` 这三类后台高频页面。
- 同步更新技能正文和实施检查清单，把“遇到高频后台模式时优先读取代码模式文件”的入口接进主工作流。
- 代码骨架明确标注为结构示意，强调布局、职责和组件搭配顺序，避免把技能写成容易过时的硬编码 props 清单。
- 已再次执行技能结构校验；本次未在真实 React 项目里编译这些示意骨架，未验证项是示意代码在当前包版本下的逐行可编译性。

### 下次方向

- 后续可以把这些示意骨架迁移成真正可编译的小型模板工程或 assets。
- 若开始实战用技能生成页面，可根据生成结果继续收紧哪些 props 必须写、哪些结构必须保留。

## 2026-04-01 Fluent 技能补充可编译模板与 Figma 资源

### 本次改动

- 为 `fluent-react-v9` 新增 `assets/fluent-react-page-starters` 小型模板工程，包含 `package.json`、Vite 基础配置、三份可直接运行的 Fluent React 页面起手式，以及一个可切换示例页的应用入口。
- 模板覆盖三类高频后台页面：`DataGrid + Drawer` 工作区列表页、`Tree + Detail Pane` 树形治理页、`Field + Form Section` 分组表单页，并在技能正文中加入资产入口。
- 为 `fluent2-frontend-style` 新增 Figma Community 资源参考，明确可搭配 `Microsoft Fluent 2 Web` 设计文件作为视觉与组件层级基准，并说明它与 `fluent-react-v9` 的配合方式。
- 本次计划在临时目录执行模板编译验证；若验证失败需继续修正模板。当前未验证项是模板在真实业务项目中的二次扩展体验。

### 下次方向

- 如果模板验证通过，后续可继续把它扩成多页面、多布局的小型 starter 包。
- 若开始用具体 Figma 节点落地页面，可继续补节点级映射与设计到组件的转换规则。

## 2026-04-01 Fluent 2 测试页面实验场

### 本次改动

- 在仓库根目录新增独立实验目录 `frontend-fluent2-lab`，直接复用 `fluent-react-v9` 技能沉淀的小型模板资产，不修改现有 `frontend` Vue 工程。
- 将实验目录收口为一套可切换的测试页面集合，包含 `列表 + 抽屉`、`树 + 详情`、`分组表单` 三类后台高频页面，并补充 `README.md`、本地 `.npmrc` 和更清晰的实验场文案。
- 解决了当前机器 `pnpm` 全局 `symlink=false` 导致依赖无法解析的问题，通过实验目录本地 `.npmrc` 强制使用 `hoisted` 链接方式，随后完成 `pnpm build`。
- 已完成真实验证：`pnpm build` 通过，且本地启动后已在浏览器打开 `http://127.0.0.1:9016/` 确认首页与默认测试页面结构正常。

### 下次方向

- 选一个具体 Figma Frame，把这套实验场中的某一页替换成更贴近设计稿的真实测试页面，优先从 `Workspace Grid + Drawer` 或 `Tree + Detail Pane` 开始。
- 若实验场要继续扩展，可再补应用壳、登录页和主题切换，再决定是否进入正式 `frontend-fluent2` 项目。

## 2026-04-01 Teams 响应式设计测试页

### 本次改动

- 在 `frontend-fluent2-lab` 中新增基于 `Microsoft Teams UI Kit` 的真实设计测试页，并将其设为实验场默认首页，文件位于 `src/pages/TeamsResponsivenessPage.tsx`。
- 这张页面没有复刻 Figma 编辑器，而是提炼了 `Responsiveness` 设计页的三栏结构、响应式说明层级和设备预览表达，重新组织为适合 Fluent React 实验场的内容页。
- 同步更新了实验场入口切换逻辑与 README，并再次完成 `pnpm build`；随后通过本地 `http://127.0.0.1:9016/` 实际打开页面，确认默认首页已切换到该测试页。

### 下次方向

- 从同一套 Teams 设计文件中继续抽更像真实业务界面的 Frame，例如频道页、消息页或应用入口页，而不是停留在设计指南页。
- 如果这张设计测试页方向稳定，下一步优先把 `Tree + Detail Pane` 页面替换成基于具体 Figma Frame 的协作型工作台页面。

## 2026-04-01 扩充 Fluent 2 实验场测试样式

### 本次改动

- 在 `frontend-fluent2-lab` 中新增三类测试页面：`应用壳 + 仪表盘`、`登录页`、`消息中心`，分别用于验证 Fluent 2 企业后台壳层、登录入口和活动流样式。
- 更新实验场入口切换与默认首页，使 `frontend-fluent2-lab` 打开后优先展示更完整的应用壳测试页，而不再只展示单一参考页。
- 同步更新 `README.md`，并再次执行 `pnpm build` 确认所有测试页面可通过类型检查与打包。
- 已启动本地开发服务，当前可通过 `http://127.0.0.1:9016/` 访问实验场；本次未完成 Playwright 浏览器内复验，未验证项是各页面切换后的逐页可视确认。

### 下次方向

- 继续把新增的 `消息中心` 和 `树 + 详情` 页面替换成来自具体 Figma Frame 的测试稿，逐步减少“概念样式页”占比。
- 如果实验场方向稳定，可补主题切换、顶层应用壳导航态和一个更接近业务的频道/协作工作台页面。

## 2026-04-01 新增 Teams 频道工作台测试页

### 本次改动

- 在 `frontend-fluent2-lab` 中新增 `TeamsChannelWorkspacePage`，把 Teams 风格的频道、活动流、关联任务和右侧成员栏收口成更像真实业务界面的协作工作台测试页。
- 同步更新实验场入口切换和 README，使实验场现在覆盖“设计参考页 + 应用壳 + 登录页 + 消息中心 + 协作工作台 + 通用后台骨架”这一组更完整的测试样式。
- 已再次执行 `pnpm build`，确认新增页面通过类型检查和打包；本次未做浏览器逐页点击切换回归，未验证项是所有入口按钮在实际 UI 中的逐一切换效果。

### 下次方向

- 继续把 `TeamsChannelWorkspacePage` 对齐到更具体的 Figma Frame，优先补频道顶部状态区、正文模块和右栏信息卡的更细节层级。
- 如果要继续扩实验场，下一步建议加主题切换或一个更接近审批/治理链路的工作台页。

## 2026-04-01 同时补齐 Fluent 2 Web 与 Teams 测试页

### 本次改动

- 在 `frontend-fluent2-lab` 中新增 `FluentComponentGalleryPage` 与 `ApprovalWorkbenchPage`，分别覆盖偏 `Microsoft Fluent 2 Web` 的组件/说明型页面，以及偏治理链路的审批工作台页面。
- 更新实验场入口与 README，使实验场现在同时覆盖两条设计源：`Microsoft Fluent 2 Web` 与 `Microsoft Teams UI Kit`，不再只偏向单一风格来源。
- 已再次执行 `pnpm build`，确认新增页面均可通过类型检查和打包；本次未重新做浏览器逐页切换回归，未验证项是新增入口在运行中的逐页点击表现。

### 下次方向

- 继续从 `Microsoft Fluent 2 Web` 中抽更具体的组件文档页或工作区页，把 `FluentComponentGalleryPage` 再收口成更贴近真实设计节点的版本。
- 继续从 `Teams` 设计线中推进一个更接近消息/频道/审批联动的具体 Frame，逐步减少抽象概念页比例。

## 2026-04-01 继续细化 Fluent 2 Web 与 Teams 页面层级

### 本次改动

- 在 `frontend-fluent2-lab` 中新增 `FluentSpecWorkspacePage` 与 `TeamsConversationPage`，分别用于验证偏 `Microsoft Fluent 2 Web` 的规范工作区布局，以及偏 `Teams` 的线程列表、主讨论区和右栏信息栏布局。
- 更新实验场入口与 README，使两条设计线都从“概念页”向“更具体的页面层级”推进，而不是只停留在组件展示或总体风格。
- 期间修正了 `TeamsConversationPage` 中 `makeStyles` 的边框覆盖写法，并再次执行 `pnpm build` 确认页面可正常打包。

### 下次方向

- 继续把 `TeamsConversationPage` 与 `TeamsChannelWorkspacePage` 做成一组联动页，例如共享相近的导航、状态和信息卡节奏。
- 继续把 `FluentSpecWorkspacePage` 往更具体的 `Microsoft Fluent 2 Web` 节点映射收口，减少占位示意块比例。

## 2026-04-01 实验场扩充到 20 张测试页

### 本次改动

- 在 `frontend-fluent2-lab` 中新增 8 张测试页：`分析总览`、`项目看板`、`设置中心`、`成员目录`、`日程规划`、`资源库`、`通知偏好`、`审计时间线`，并全部接入实验场入口。
- 目前实验场已达到 20 张测试页，覆盖数据总览、应用壳、文档工作区、审批治理、Teams 频道与线程、消息流、设置、资源卡、表单、目录、时间线等多种 Fluent 2 页面模式。
- 已再次执行 `pnpm build`，确认 20 张测试页全部可通过类型检查和打包；本次未逐页做浏览器视觉核对，未验证项是所有入口在运行时的逐页观感和切换细节。

### 下次方向

- 从 20 张测试页中筛出 4 到 6 张最值得继续深化的页面，逐步从“样式覆盖”转向“高质量样板页”。
- 优先对 `Teams` 线和 `Fluent 2 Web` 线各选 2 张页面，继续按具体 Figma 节点做更强约束的细化落地。

## 2026-04-01 收口共享 Fluent 基座并精修四张重点页

### 本次改动

- 在 `frontend-fluent2-lab` 中新增 `src/lab/catalog.ts` 与 `src/lab/primitives.tsx`，把 20 张测试页的来源分组、页面元信息和共享卡片/章节/统计条基元收口成可复用层。
- 重写实验场入口 `src/App.tsx`，加入 `共享基座 / Fluent 2 Web / Teams` 三类分组切换、亮暗主题切换，并把默认首页切到 `Teams 频道工作台`，让实验场更接近真实设计评审入口。
- 精修 `TeamsChannelWorkspacePage`、`TeamsConversationPage`、`FluentSpecWorkspacePage`、`ApprovalWorkbenchPage` 四张高价值页面，统一接入共享基元和更明确的状态、统计、层级表达。
- 已执行 `pnpm build`，确认共享层改造后实验场仍可通过类型检查和打包；本次未重启浏览器逐页验收，未验证项是暗色模式下 20 张页面的逐页观感一致性。

### 下次方向

- 继续把另外几张高价值页面接入共享基元，例如 `消息中心`、`应用壳 + 仪表盘`、`树 + 详情`，逐步减少独立页面里重复的布局样式定义。
- 选取 `Fluent 2 Web` 和 `Teams` 各一个更具体的 Figma Frame，对四张重点页中的至少两张做更强约束的设计映射，而不是继续停留在抽象样式层。

## 2026-04-01 修正实验场暗色模式下的偏白问题

### 本次改动

- 修正 `frontend-fluent2-lab` 中多张页面的浅色硬编码背景与白底渐变，重点处理了 `应用壳 + 仪表盘`、`Fluent 组件展示`、`Fluent 规范工作区`、`Teams 频道工作台`、`Teams 消息线程`、`登录页`、`分析总览`。
- 将固定的 `#f3f2f1`、`#f6f7fb`、`#fbfbfc`、`rgba(255,255,255,...)` 等颜色替换为 Fluent 主题 token，并把图表和首屏卡片的渐变统一收口为品牌色透明层加中性色底。
- 已执行 `pnpm build`，并再次扫描实验场源码确认相关浅色硬编码已清空；本次未通过浏览器逐页人工检查暗色观感，未验证项是所有页面在暗色模式下的细节对比度与视觉平衡。

### 下次方向

- 继续做一次暗色模式专项收口，重点检查标题层级、分隔线强度、品牌蓝在深色底上的饱和度，以及统计卡片在深色背景下的抢眼程度。
- 如果下一轮继续优化视觉，可给实验场加入更明确的亮暗主题 token 分层，而不是只依赖默认 `webLightTheme / webDarkTheme`。

## 2026-04-01 接入 Microsoft 365 UI Kit 作为第三个 Figma 设计源

### 本次改动

- 将 `https://www.figma.com/community/file/1314695480773948455` 以 `Microsoft 365 UI Kit` 的身份接入 Fluent 风格技能体系，补充到 [`fluent2-figma-resource.md`](C:/Users/Administrator/.codex/skills/fluent2-frontend-style/references/fluent2-figma-resource.md)。
- 同步更新 [`fluent2-frontend-style/SKILL.md`](C:/Users/Administrator/.codex/skills/fluent2-frontend-style/SKILL.md)，明确它适用于 `Microsoft 365` 生态中的跨产品工作流、上下文侧载、应用入口和避免上下文切换的场景。
- 当前已确认社区页标题为 `Microsoft 365 UI Kit`，但尚未稳定拿到真实设计文件 URL 与节点级 `node-id`；因此这次先完成本地索引与使用边界接入，未直接映射到实验场页面。

### 下次方向

- 等浏览器或 Figma 会话稳定后，继续补抓 `Microsoft 365 UI Kit` 的真实设计文件链接、`fileKey` 和可复用的节点级入口。
- 若后续实验场开始做 Outlook / M365 任务流 / 应用侧载类页面，可把它作为 `Fluent 2 Web` 与 `Teams` 之外的第三条布局参考线接入页面模板。

## 2026-04-01 新增 Microsoft 365 任务流测试页

### 本次改动

- 在 `frontend-fluent2-lab` 中新增 [`Microsoft365ContextFlowPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/Microsoft365ContextFlowPage.tsx)，用 `Microsoft 365 UI Kit` 的场景语义做了一张跨应用任务流测试页。
- 同步更新 [`catalog.ts`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/lab/catalog.ts)、[`App.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/App.tsx) 与 [`README.md`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/README.md)，把实验场正式扩成 `共享基座 / Fluent 2 Web / Teams / Microsoft 365` 四类分组，并把页面总数更新为 21。
- 新页面重点验证 `Continue in ...` / `Open in ...` 这类 handoff 文案、上下文载荷、跨应用任务流和右侧应用关联栏，而不是单一后台页面语法。
- 已执行 `pnpm build`，确认新增页面和新分组可正常通过类型检查与打包；本次未做浏览器逐页可视回归，未验证项是新页面在运行中的首屏观感和与其他 20 页的切换节奏。

### 下次方向

- 继续补一张 `Microsoft 365` 风格的应用入口或侧载面板页，形成 `任务流 + 应用入口` 的最小页面组。
- 一旦拿到 `Microsoft 365 UI Kit` 的真实设计文件链接和节点级入口，再把这张测试页从“场景语义映射”推进到“具体节点约束映射”。

## 2026-04-01 补充 Microsoft 365 应用入口页并记录抓取结论

### 本次改动

- 继续尝试抓取 `Microsoft 365 UI Kit` 的公开设计信息，确认社区页标题为 `Microsoft 365 UI Kit`，且公开描述明确强调“跨 Microsoft 365 生态设计应用、提供核心组件、场景模板和最佳实践，并让用户留在当前工作流中”。
- 同步把这次抓取结果回写到 [`fluent2-figma-resource.md`](C:/Users/Administrator/.codex/skills/fluent2-frontend-style/references/fluent2-figma-resource.md)，并补充 `hub_files` 公开接口当前返回 `202` 且无内容这一限制。
- 在 `frontend-fluent2-lab` 中新增 [`Microsoft365AppLauncherPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/Microsoft365AppLauncherPage.tsx)，把 `Microsoft 365` 设计线扩成 `任务流 + 应用入口` 两张页面的最小页面组。
- 同步更新 [`catalog.ts`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/lab/catalog.ts)、[`App.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/App.tsx) 与 [`README.md`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/README.md)，实验场页面总数增至 22。
- 已执行 `pnpm build`，确认第二张 `Microsoft 365` 页面和新的分组入口可正常打包；本次未做浏览器逐页人工验收，未验证项是两张 `Microsoft 365` 页面之间的切换观感和暗色模式细节。

### 下次方向

- 优先继续获取 `Microsoft 365 UI Kit` 的真实设计文件链接、`fileKey` 和节点级入口，避免后续页面一直停留在语义映射层。
- 如果抓取仍受限，下一轮可以先做一张更偏 `Outlook sidecar` 或 `Loop 协作面板` 的页面，把 `Microsoft 365` 这条设计线再向具体业务形态推进一步。

## 2026-04-01 确认 Microsoft 365 UI Kit 的真实设计文件入口

### 本次改动

- 用户提供了 `Microsoft 365 UI Kit` 的真实设计文件链接：`https://www.figma.com/design/QsOLI0O1ZiRvYS2lA3wKdA/Microsoft-365-UI-Kit--Community-`，据此确认其真实 `fileKey` 为 `QsOLI0O1ZiRvYS2lA3wKdA`。
- 已将真实设计文件 URL 与 `fileKey` 回写到 [`fluent2-figma-resource.md`](C:/Users/Administrator/.codex/skills/fluent2-frontend-style/references/fluent2-figma-resource.md)，把该资源从“只有社区链接”升级为“有真实设计入口”的设计源。
- 同时验证到：当前 connector 直接读取该文件根节点 `0:1` 的元数据会超时，因此后续需要优先提供具体 `node-id` 或页面 Frame，再做节点级抓取。

### 下次方向

- 继续从 `Microsoft 365 UI Kit` 中拿到一个具体 `node-id`，优先建议抓“应用入口”或“侧载面板”相关 Frame。
- 一旦拿到具体节点，就把现有的 `Microsoft365ContextFlowPage` 或 `Microsoft365AppLauncherPage` 之一升级成更贴近真实设计稿的版本。

## 2026-04-01 回退 Microsoft 365 设计线

### 本次改动

- 按当前决策移除了 `frontend-fluent2-lab` 中新增的两张 `Microsoft 365` 页面：`Microsoft365ContextFlowPage` 与 `Microsoft365AppLauncherPage`，并同步删掉 `catalog.ts`、`App.tsx`、`README.md` 中对应的入口与分组说明。
- 同步回退 `fluent2-frontend-style` 技能里关于 `Microsoft 365 UI Kit` 的额外触发规则，并从 [`fluent2-figma-resource.md`](C:/Users/Administrator/.codex/skills/fluent2-frontend-style/references/fluent2-figma-resource.md) 删除第三个设计源的整段记录，恢复为只保留 `Microsoft Fluent 2 Web` 与 `Microsoft Teams UI Kit` 两条线。
- 已执行 `pnpm build`，确认实验场回退后仍可正常通过类型检查与打包；同时复扫实验场源码，已无 `m365` 或 `Microsoft 365` 相关残留引用。

### 下次方向

- 后续继续聚焦 `Microsoft Fluent 2 Web` 与 `Microsoft Teams UI Kit` 两条设计线，不再分散到第三条风格参考。
- 如果下一轮继续扩实验场，优先把现有 20 张页面进一步贴合这两套设计源，而不是再引入新的生态模板。

## 2026-04-01 收口应用壳、Teams 频道与审批页的视觉层级

### 本次改动

- 继续对 [`AppShellShowcasePage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/AppShellShowcasePage.tsx)、[`TeamsChannelWorkspacePage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/TeamsChannelWorkspacePage.tsx)、[`TeamsConversationPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/TeamsConversationPage.tsx)、[`ApprovalWorkbenchPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/ApprovalWorkbenchPage.tsx) 做了更强的 Fluent 视觉收口。
- 这轮重点增强了 hero banner、状态芯片、分区背景、右栏层级和主工作区对比度，让实验场从“可用测试页”进一步靠近“样板页”。
- 已执行 `pnpm build` 并通过；本次未做浏览器逐页人工复验，未验证项是这些页面在暗色模式和窄屏下的最终观感。

### 下次方向

- 继续把 `MessageCenterPage`、`FluentComponentGalleryPage` 和 `LoginShowcasePage` 拉到同一套视觉语法。
- 如果下一轮继续优化，建议直接做一次浏览器回归，确认首页、Teams 频道页和审批页在亮暗主题下的整体平衡。

## 2026-04-01 收口消息中心、组件展示与登录页的 Fluent 语法

### 本次改动

- 继续收口 [`MessageCenterPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/MessageCenterPage.tsx)、[`FluentComponentGalleryPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/FluentComponentGalleryPage.tsx)、[`LoginShowcasePage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/LoginShowcasePage.tsx)，把三张页统一到更完整的 Fluent 2 样板语法。
- 这轮重点补了 banner 式导语、统计信息、双栏说明区和更稳定的背景层级，避免暗色模式下局部偏白，也减少了页面之间的视觉断裂。
- 已重新执行 `pnpm build` 并通过；本次未做浏览器逐页回归，未验证项是 20 张页面在暗色模式下的整体一致性和个别高密度页面的可读性。

### 下次方向

- 下一步建议做一次浏览器回归，只看首页、Teams 频道页、审批页、消息中心和登录页的暗色表现。
- 如果回归结果稳定，就停止扩页，转为从现有 20 张中挑 4 到 6 张继续打磨细节。

## 2026-04-01 对齐 Fluent 2 Web 的应用壳与组件文档页

### 本次改动

- 继续收紧 [`AppShellShowcasePage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/AppShellShowcasePage.tsx) 和 [`FluentComponentGalleryPage.tsx`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2-lab/src/pages/FluentComponentGalleryPage.tsx)，把两张页调整成更接近 Fluent 2 Web 截图里的应用壳与文档式组件页结构。
- 应用壳页补了更明确的左侧导航、面包屑、内容工作区、预览框和 Copy me / variant 收口，组件展示页补了左侧目录、中部文档说明、预览框和 Variants 区，减少了此前偏仪表盘化的表达。
- 已重新执行 `pnpm build` 并通过；本次未做浏览器逐页回归，未验证项是这两张页在实际窗口里的细节观感和与其他 Fluent 页面的一致性。

### 下次方向

- 下一步建议只做浏览器视觉回归，重点看应用壳、组件文档页和登录页在亮暗主题下的层级与留白。
- 如果视觉稳定，就继续把其余 Fluent 2 Web 页面统一到同一套壳层和文档语法，不再增加新的页面类型。

## 2026-04-01 新增 frontend-fluent2 一期 React Fluent 2 并行实验线

### 本次改动

- 在仓库根目录新增独立子应用 [`frontend-fluent2`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2)，采用 `React + TypeScript + Vite + Fluent UI React v9 + TanStack Query + Zustand + Axios` 建立并行的一期前端基座，没有替换或侵入现有 [`frontend`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend) Vue 工程。
- 先按当前分支真实后端语义接入认证、运行时导航、菜单空间与菜单治理接口，建立了登录页、会话恢复、应用壳、运行时导航、空间切换、占位页与 [`system/menu`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2/src/features/menu/pages/MenuWorkbenchPage.tsx) 工作台。
- 路由采用“静态壳路由 + 本地 route registry + 占位页回退”，只对已实现路径注册真实页面，其余运行时路径统一进入占位页，避免前端对白名单外路径白屏或死链。
- 请求层统一处理 `Authorization`、`X-Tenant-ID`、业务错误与 `401` 清会话回跳登录；所有接口响应先经 adapter 映射为前端稳定类型，页面层不直接扩散后端字段形态。
- 已新增 [`frontend-fluent2/README.md`](C:/Users/Administrator/Documents/GitHub/G-G-E-commerce/frontend-fluent2/README.md) 说明运行方式，并完成 `pnpm --dir frontend-fluent2 install`、`pnpm --dir frontend-fluent2 exec tsc --noEmit`、`pnpm --dir frontend-fluent2 build` 验证；本次未完成真实后端联调与浏览器级 UI 冒烟。

### 下次方向

- 下一步优先做真实后端联调与浏览器冒烟，重点覆盖登录成功/失败、401 自动退出、菜单空间切换、运行时导航刷新、菜单创建编辑删除与占位页回退。
- 继续扩展 `route registry` 与已实现页面集合，后续可按同一基座平滑接入 `system/page`、`system/role`、`system/user`，但仍保持“前端只消费后端权限结果”的边界。




