---
name: fluent2-frontend-style
description: 按 Fluent 2 设计原则约束产品界面、后台和工作台的视觉、排版、交互与评审。用于页面重设计、视觉方案输出、交互优化、设计评审、界面一致性收敛，以及需要微软企业感、聚焦、低噪和高可读风格时。
---

# Fluent2 Frontend Style

## 概览

把 Fluent 2 当成“设计语言和产品气质”，不要把它收窄成几个控件皮肤。

默认面向产品后台、运营工作台、专业工具界面和高信息密度页面。

## 先写三件事

1. 写视觉主张。
- 用一句话说明页面的气质、材料感和密度。
- 目标通常是清晰、克制、可靠、专业，而不是炫技。

2. 写内容层级。
- 明确首屏先传达什么。
- 明确主工作区、次级上下文和低频操作分别放哪里。

3. 写交互主张。
- 明确焦点如何切换。
- 明确反馈放在哪一层。
- 明确哪些动作需要即时响应，哪些动作需要确认。

## 固定遵循的原则

- `Natural on every platform`：先适配当前平台的阅读和交互习惯，不生搬硬套桌面软件外观。
- `Built for focus`：每个区块只承担一个主任务，减少视觉噪音。
- `One for all, all for one`：局部差异服从整体系统，不为单页发明另一套语法。
- `Unmistakably Microsoft`：整体气质应可靠、干净、克制、强调信息与任务，而不是营销感和装饰感。

需要校准原则时，读取 [references/fluent2-principles-notes.md](references/fluent2-principles-notes.md)。

## 使用顺序

- 先写视觉主张、内容层级和交互主张，再决定颜色、阴影和动效。
- 先决定主工作区与次级面板的关系，再决定卡片、分割和边框。
- 先确认状态反馈、等待体验和跨页面切换，再决定细节文案和动效。
- 需要处理 handoff、等待反馈、材质、排版和可访问性时，读取 [references/fluent2-patterns-and-systems.md](references/fluent2-patterns-and-systems.md)。
- 若用户提供或明确提到 Figma 社区文件 `Microsoft Fluent 2 Web`，优先将其作为视觉与组件层级参考，再结合本技能进行产品化收口。
- 若任务明显带有 Teams、协作、会话、频道、会议、消息或应用生态入口特征，额外参考 `Microsoft Teams UI Kit`，但仍用 Fluent 2 原则约束收口，不直接照搬品牌化细节。


## 后台化落地规则

- 保持低噪音背景、清晰层级和明确主操作。
- 用少量颜色表达状态，不用大面积装饰色块抢焦点。
- 优先用留白、对齐、字号和分割组织信息，而不是堆卡片和厚边框。
- 提高信息密度时，优先压缩装饰而不是压缩可读性。
- 让操作入口贴近任务本身，不把高频动作藏得太深。

## 页面结构规范

- 应用壳优先采用“顶部栏 + 导航区 + 主工作区 + 次级面板”。
- 页面头部优先放标题、状态、摘要和主操作，不堆一排低频按钮。
- 工作区优先承载核心表格、表单、图表或编辑器。
- 次级面板负责详情、属性、帮助和风险说明，不抢主工作区层级。
- 消息反馈按严重性分层：内联提示、局部消息、全局通知。
- 需要按系统层面检查布局、材质、排版和表层关系时，读取 [references/fluent2-surface-layout-checklist.md](references/fluent2-surface-layout-checklist.md)。

## 文案规范

- 标题要像导航和任务名，不要像广告语。
- 辅助说明只解释范围、结果和后果，不写空泛口号。
- 操作文案优先用动词开头，直接说明动作。
- 对危险或不可逆动作，明确写出影响对象和结果。

## 动效规范

- 只保留支持理解和焦点迁移的必要过渡。
- 用动效解释层级变化、展开收起、面板进入和反馈出现。
- 控制时长和幅度，避免让动效本身成为视觉负担。
- 若动效不能提升理解，就删掉。
- 如果页面存在等待状态、跨应用切换或 AI 生成后跳转，优先参考 `Wait UX` 与 `Handoffs` 的做法。

## 评审清单

- 先检查主任务是否在首屏清晰可见。
- 再检查主次层级是否明确，是否存在“所有元素都在抢注意力”。
- 再检查文案是否直接、克制、可操作。
- 再检查状态、反馈、焦点和可访问性是否完整。
- 需要完整清单时，读取 [references/fluent2-ui-review-checklist.md](references/fluent2-ui-review-checklist.md)。

## 实施禁忌

- 不做大面积渐变面板堆叠。
- 不做卡片瀑布流式后台首页。
- 不把营销页语言、夸张主视觉和悬浮装饰带进工作台。
- 不让阴影、边框和图标比数据和操作更抢眼。
- 不为单页局部效果破坏整套产品的一致性。

## 参考资料

- 原则摘要：读取 [references/fluent2-principles-notes.md](references/fluent2-principles-notes.md)
- 设计评审：读取 [references/fluent2-ui-review-checklist.md](references/fluent2-ui-review-checklist.md)
- 模式与系统规则：读取 [references/fluent2-patterns-and-systems.md](references/fluent2-patterns-and-systems.md)
- 布局与表层检查：读取 [references/fluent2-surface-layout-checklist.md](references/fluent2-surface-layout-checklist.md)
- Figma Community 设计资源：读取 [references/fluent2-figma-resource.md](references/fluent2-figma-resource.md)
