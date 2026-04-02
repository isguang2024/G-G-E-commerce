# Fluent 2 Figma 设计资源

## 推荐资源

- 官方 Figma Community 文件：`Microsoft Fluent 2 Web`
- 链接：`https://www.figma.com/community/file/836828295772957889/microsoft-fluent-2-web`
- 已验证可访问的真实设计文件：`https://www.figma.com/design/sVtHDv2NF7d1dMZFiz5eaX/Microsoft-Fluent-2-Web--Community-`
- 真实 `fileKey`：`sVtHDv2NF7d1dMZFiz5eaX`
- 发布者：`Microsoft`
- 当前可确认的公开信息：
  - 社区页标题为 `Microsoft Fluent 2 Web`
  - 资源类型为 `Design file`
  - 社区标签包含 `#fluent`、`#fluent 2`、`#fluent ui 9`、`#library`、`#microsoft`
  - 社区页显示该资源已被大量用户使用，适合作为官方 Web 风格参考
  - 社区页可跳转到真实设计文件，而不是只能停留在社区壳层

## 补充资源

- 官方 Figma Community 文件：`Microsoft Teams UI Kit`
- 链接：`https://www.figma.com/community/file/916836509871353159/microsoft-teams-ui-kit`
- 已验证可访问的真实设计文件：`https://www.figma.com/design/GFS6tbMoqoyB5MoNO555k9/Microsoft-Teams-UI-Kit--Community-?node-id=1-355`
- 真实 `fileKey`：`GFS6tbMoqoyB5MoNO555k9`
- 发布者：`Microsoft`
- 当前可确认的公开信息：
  - 社区页标题为 `Microsoft Teams UI Kit`
  - 资源类型为 `Design file`
  - 社区标签包含 `#design guidelines`、`#microsoft teams`、`#teams app ecosystem`、`#teams app store`、`#ui kit`
  - 社区页可直接通过 `Open in Figma` 跳到真实设计文件
  - 这套资源更适合补充协作、消息、频道、会议与 Teams 生态入口语义

## 什么时候优先参考它

- 用户明确说要按 Fluent 2 官方 Web 设计稿落界面
- 任务包含 Figma 节点、截图、页面还原、设计对齐
- 需要校准组件层级、排版密度、留白和状态样式

## 什么时候额外参考 `Microsoft Teams UI Kit`

- 页面属于协作产品、通信产品、团队工作台、会话列表或会议场景
- 需要设计频道侧栏、消息流、协作入口、应用商店入口、团队空间等 Teams 风格结构
- 需要借用 Microsoft 产品线里更偏协作语义的页面组织方式

## 在本技能里的使用方式

- 把它当作视觉和组件组织参考，而不是生硬逐像素还原目标。
- 先从设计稿提取层级、栅格、组件组合、状态表达，再结合产品语义做收口。
- 如果设计稿和产品实际任务冲突，优先保留任务清晰度、可访问性和工作流效率。
- 重新抓取社区页时，可优先关注它强调的两个方向：
  - 组件变量与代码实现更加对齐，适合反推 Fluent React v9 的 token 与状态组织方式
  - `Appearance` 相关区域值得重点参考，用于校准外观状态、密度和可配置项表达
- 对 `Microsoft Teams UI Kit`，重点不要放在复制 Teams 品牌外观，而是提取：
  - 协作型信息架构
  - 导航与会话层级
  - 频道、消息、成员和应用入口的优先级
  - 长列表与高频切换场景下的布局稳定性

## 与 `fluent-react-v9` 的配合方式

- `fluent2-frontend-style` 负责视觉、层级、文案和交互规则。
- `fluent-react-v9` 负责 React 组件选型、页面结构和实现路径。
- 当用户同时要“按 Figma 设计落地 React Fluent 页面”时，先用本技能校准视觉，再用 `fluent-react-v9` 选择实际组件和代码模式。
- 优先把 Figma 中的页面结构映射到 `fluent-react-v9/assets/fluent-react-page-starters` 的三类起手式，再做细化。
- 如果参考 `Microsoft Teams UI Kit`，优先映射到：
  - `Tree + Detail Pane` 作为频道/导航与内容区骨架
  - `Workspace Grid + Drawer` 作为协作列表与详情检查面板骨架
  - 表单起手式用于设置、成员管理和应用配置页

## 若具备 Figma 工具

- 用户提供具体 Figma 节点 URL 时，优先读取节点结构、截图和变量定义。
- 优先抽取 spacing、层级、状态、组件组合和文案结构。
- 不要因为设计稿存在装饰性处理，就破坏 Fluent 2 的后台聚焦原则。
- 当前已验证：
  - Figma Community 页面预览可正常访问
  - 真实设计文件 `sVtHDv2NF7d1dMZFiz5eaX` 可从社区页进入
  - `Microsoft Teams UI Kit` 的真实设计文件 `GFS6tbMoqoyB5MoNO555k9` 也可从社区页进入
  - Figma 账号已接入 connector，但当前席位为 `View`
  - 直接以社区文件根节点抓取元数据与截图仍可能超时
  - 节点 `8911:3186` 已成功抓到元数据与截图，对应组件为 `Badge`
  - 该节点元数据表明文件中存在清晰的组件文档结构：`Header`、`Component name`、`Title`、`Variants`、`Isolated` 等区块
  - `Badge` 节点中能直接看到组件变体信息，例如 `Color`、`Size`、`Appearance`、`In office`、`Status`
  - `get_design_context` 与变量定义读取在当前环境下仍可能提示“需要先选中图层”
  - 对 `Microsoft Teams UI Kit`，已拿到真实设计 URL 和默认节点 `1:355`
  - 当前用 connector 直接读取其默认节点元数据仍可能超时
- 当前限制与建议：
  - 社区文件链接本身更适合做公开资源参考；真正用于实现时，应优先切到真实设计文件链接
  - 如果要稳定抓取节点结构，优先提供编辑器内的设计文件链接，并附 `node-id`
  - 如果要把某个页面直接转换为实现稿，优先指定具体 Frame，而不是让工具读取整个根节点
  - 如果 `get_design_context` 提示未选中图层，先退回到 `get_metadata` 与 `get_screenshot` 组合读取，再由具体节点继续收敛实现
  - 对 `Microsoft Teams UI Kit`，建议后续按具体频道页、消息页、应用入口页分别给 `node-id`，不要直接从默认首页节点开始读

## 两套资源如何分工

- `Microsoft Fluent 2 Web`
  - 适合 Web 端 Fluent 2 组件层级、说明页、规范页、治理工作台
- `Microsoft Teams UI Kit`
  - 适合频道、会话、协作、成员、会议与 Teams 风格信息架构
