# 消息系统设计

> 现状基线：2026-03-29。本文用于替换右上角通知组件的 mock 数据方案，定义站内消息系统的正式模型与分阶段落地方式。

## 1. 现状判断

当前右上角消息组件位于：

- `frontend/src/components/core/layouts/art-notification/index.vue`

现状特点：

- 面板分为 `通知 / 消息 / 代办` 三栏
- 数据全部是前端 mock
- 没有后端消息主表、收件箱表、未读计数接口
- 前端已有可复用的 websocket 客户端：`frontend/src/utils/socket/index.ts`

因此当前要做的不是“换几条假数据”，而是补齐正式消息系统。

## 2. 设计目标

### 2.1 产品目标

- 右上角面板变成真实收件箱，而不是演示组件
- 用户能区分“看看就行”的通知和“必须处理”的待办
- 平台、团队、系统动作都能统一进入一套消息中心
- 后续可以接 websocket 推送，但第一阶段不依赖它才能工作

### 2.2 视觉与交互方向

- visual thesis：像后台工作收件箱，不像社交 feed，也不像公告弹窗堆叠
- content plan：顶部计数与筛选、三类消息列表、消息详情动作、进入完整消息中心
- interaction thesis：
  - 右上角面板维持轻量预览，只展示最近几条与未读数
  - 点击消息后优先直接落到业务页或右侧详情抽屉
  - 待办类消息必须给出明确主操作，不做纯文本提醒

## 3. 正式消息分类

消息系统保留当前三大入口，但语义要正式化：

### 3.1 通知 `notice`

用于“以告知为主，不要求即时处理”的内容。

建议包含：

- 平台公告
- 团队公告
- 版本发布
- 安全提醒
- 账单/配额提醒
- 系统维护通知

特点：

- 允许批量已读
- 可以没有发送者头像
- 可以有范围标签，如“平台”“当前团队”

### 3.2 消息 `message`

用于“有人或某个业务对象直接发给你”的内容。

建议包含：

- 团队邀请
- 成员变更通知
- 评论 / @ 提醒
- 审批结果通知
- 功能包开通结果
- API / 页面 / 菜单变更提醒

特点：

- 需要展示发送者
- 可以跳转到具体业务对象
- 可以按“来自谁 / 来自哪个团队”聚合

### 3.3 待办 `todo`

用于“你必须处理”的动作型消息。

建议包含：

- 待审核
- 待确认邀请
- 待补充配置
- 待处理异常
- 待完成权限绑定
- 待执行系统同步

特点：

- 一定要有主操作
- 必须记录处理状态
- 不建议仅靠“已读”解决，应有“完成 / 忽略 / 延后”语义

## 4. 发送者模型

消息发送者不要只做成“用户头像 + 名字”，否则系统消息会很乱。建议正式区分：

### 4.1 发送者类型 `sender_type`

- `system`
- `platform_user`
- `team_user`
- `service`
- `automation`

解释：

- `system`：系统自动生成，如安全提醒、权限快照刷新结果
- `platform_user`：平台管理员或平台运营人员发送
- `team_user`：团队成员、团队管理员发送
- `service`：业务服务发送，如“页面同步器”“API 注册器”
- `automation`：定时任务、工作流、机器人发送

### 4.2 发送者展示规则

- `system`：展示统一系统徽标，不展示个人头像
- `service` / `automation`：展示服务图标与服务名
- `platform_user` / `team_user`：展示头像、昵称、所属团队

### 4.3 发送者字段

消息主表建议包含：

- `sender_type`
- `sender_user_id`
- `sender_name_snapshot`
- `sender_avatar_snapshot`
- `sender_service_key`

说明：

- 发送时保留快照，避免用户改名后历史消息全变
- 如果是服务类发送者，不依赖用户表

## 5. 右上角消息面板设计

## 5.1 保留三栏，不改成更多 tab

右上角区域空间很小，保持：

- `通知`
- `消息`
- `待办`

不要继续往上叠“公告”“安全”“审批”等更多一级 tab。

更细分类放到：

- 标签
- 图标
- 副标题
- 完整消息中心页筛选

### 5.2 面板项结构

每条消息建议统一为：

- 标题
- 副标题
- 时间
- 未读状态
- 来源标签
- 快捷动作（仅待办显示）

其中：

- 通知：显示图标 + 标题 + 范围标签
- 消息：显示发送者头像 + 标题 + 副标题
- 待办：显示状态色 + 标题 + 主按钮

### 5.3 面板行为

- 顶部显示三类未读数
- 支持“全部已读”，但仅对 `notice` 和 `message` 生效
- `todo` 不支持简单全部已读，避免掩盖未处理事项
- 点击“查看全部”进入完整消息中心，不再留空处理器

## 6. 完整消息中心设计

建议新增独立页面：

- `#/workspace/inbox`

该页承担：

- 全量消息列表
- 搜索
- 类型筛选
- 已读 / 未读筛选
- 平台 / 团队范围筛选
- 消息详情
- 待办处理

页面结构建议：

1. 左侧分类
- 全部
- 通知
- 消息
- 待办
- 已归档

2. 中间列表
- 紧凑列表
- 支持未读高亮
- 支持批量操作

3. 右侧详情
- 正文
- 发送者
- 关联对象
- 处理动作

右上角面板只做“预览入口”，完整处理放在消息中心。

## 7. 数据模型设计

第一阶段建议拆成两张主表，而不是一张表硬塞所有状态。

### 7.1 消息主表 `messages`

职责：

- 保存消息内容与发送者信息
- 描述消息类型、范围、优先级、动作入口

建议字段：

- `id`
- `message_type`：`notice | message | todo`
- `biz_type`：如 `team_invite`、`api_sync_result`、`security_alert`
- `scope_type`：`platform | team | user`
- `scope_id`
- `sender_type`
- `sender_user_id`
- `sender_name_snapshot`
- `sender_avatar_snapshot`
- `sender_service_key`
- `title`
- `summary`
- `content`
- `priority`：`low | normal | high | urgent`
- `action_type`：`route | external_link | api | none`
- `action_target`
- `status`：`draft | published | revoked | expired`
- `published_at`
- `expired_at`
- `meta`

### 7.2 收件箱表 `message_deliveries`

职责：

- 每个用户一条收件记录
- 管理未读、已读、归档、待办状态

建议字段：

- `id`
- `message_id`
- `recipient_user_id`
- `recipient_team_id`
- `box_type`：冗余自 `message_type`
- `delivery_status`：`unread | read | archived`
- `todo_status`：`pending | processing | done | ignored`
- `read_at`
- `archived_at`
- `done_at`
- `last_action_at`
- `meta`

### 7.3 为什么不只做一张表

因为一条平台公告可能发给很多人：

- 消息内容是一份
- 每个人的已读状态不同

所以必须拆成：

- 内容
- 投递结果

## 8. 接收范围设计

消息投递范围建议支持：

- 单用户
- 多用户
- 单团队全员
- 指定角色用户
- 平台管理员
- 当前团队管理员

第一阶段不要做“复杂规则引擎”，先支持显式投递：

- 指定用户 ID
- 指定团队 ID
- 指定角色编码

真正写入收件箱时展开成用户级记录。

## 9. 权限与可见性

消息中心不建议沿用菜单裁剪语义，而应独立判断：

- 谁能发
- 谁能看
- 谁能处理

建议权限键：

- `system.message.view`
- `system.message.manage`
- `system.message.send`
- `system.message.template.manage`

其中：

- 普通用户默认只有 `view`
- 平台管理员有 `manage/send`
- 团队管理员后续可有团队域发送权限，但第一阶段先不开放后台发信页

## 10. 接口设计

第一阶段建议最少补这些接口：

### 10.1 用户侧接口

- `GET /api/v1/messages/inbox/summary`
- `GET /api/v1/messages/inbox`
- `GET /api/v1/messages/inbox/:deliveryId`
- `POST /api/v1/messages/inbox/:deliveryId/read`
- `POST /api/v1/messages/inbox/read-all`
- `POST /api/v1/messages/inbox/:deliveryId/todo-action`

### 10.2 管理侧接口

- `GET /api/v1/messages`
- `POST /api/v1/messages`
- `GET /api/v1/messages/:id`
- `POST /api/v1/messages/:id/publish`
- `POST /api/v1/messages/:id/revoke`

第一阶段即使不做完整后台发信页，也建议先把管理接口边界定好。

## 11. 推送策略

第一阶段：

- 登录后拉一次摘要
- 打开面板时拉消息列表
- 处理动作后局部刷新

第二阶段：

- 接入 websocket 推送未读数变化
- 接入新消息提醒

第三阶段：

- 接入更细的业务事件流
- 支持后台实时协作消息

也就是说：

- 先拉模式
- 再推模式

不要一上来就让消息系统绑定 websocket 才能用。

## 12. 与当前项目的推荐落地顺序

### 12.1 第一阶段

- 保留右上角三栏交互
- 用正式接口替换 mock
- 只做“收件箱查看 + 已读 + 待办处理”
- 新增完整消息中心页 `#/workspace/inbox`

### 12.2 第二阶段

- 增加平台发送能力
- 增加消息模板
- 增加团队公告与邀请消息

### 12.3 第三阶段

- websocket 未读推送
- 更细粒度业务事件接入
- 团队管理员发送能力

## 13. 当前建议结论

一句话定稿：

- 右上角继续保留 `通知 / 消息 / 待办`
- 消息发送者必须支持 `系统 / 用户 / 服务`
- 后端采用 `消息主表 + 用户收件箱表`
- 第一阶段先做拉取式真实收件箱和完整消息中心，不先做复杂推送

这套设计和当前项目最匹配，改动顺序也最稳。
