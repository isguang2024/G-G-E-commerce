import {
  Avatar,
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Divider,
  Input,
  Subtitle2,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { LabBadgeRow, LabStatGrid } from '../lab/primitives';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '18px',
  },
  hero: {
    display: 'grid',
    gap: '12px',
    padding: '18px 20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.10) 0%, rgba(15,108,189,0.03) 56%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroRow: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '14px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  heroText: {
    display: 'grid',
    gap: '6px',
  },
  toolbar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '12px',
    flexWrap: 'wrap',
  },
  actions: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: '320px minmax(0, 1fr)',
    gap: '18px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  rail: {
    display: 'grid',
    gap: '14px',
  },
  railCard: {
    display: 'grid',
    gap: '12px',
    padding: '16px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  filterList: {
    display: 'grid',
    gap: '8px',
  },
  filterItem: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  content: {
    display: 'grid',
    gap: '14px',
  },
  messageCard: {
    display: 'grid',
    gap: '12px',
    padding: '18px',
    borderRadius: tokens.borderRadiusXLarge,
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    boxShadow: tokens.shadow2,
  },
  messageHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '12px',
  },
  messageMeta: {
    display: 'flex',
    gap: '12px',
    alignItems: 'start',
  },
  messageText: {
    color: tokens.colorNeutralForeground2,
  },
  messageFooter: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '10px',
    flexWrap: 'wrap',
  },
  listShell: {
    display: 'grid',
    gridTemplateColumns: '320px minmax(0, 1fr)',
    gap: '18px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  rightPanel: {
    display: 'grid',
    gap: '14px',
  },
});

const messages = [
  {
    title: '权限快照刷新已完成',
    body: '平台管理员角色变更后，相关团队边界与用户访问快照已重新生成。',
    sender: 'System Bot',
    time: '2 分钟前',
    status: '成功',
  },
  {
    title: '审批流程需要你确认',
    body: '北区运营团队提交了新的告警策略，当前等待你确认是否发布到生产环境。',
    sender: '流程中心',
    time: '12 分钟前',
    status: '待处理',
  },
  {
    title: '消息订阅范围已更新',
    body: '你已被加入“风险告警”和“运行日报”两个消息分组，建议检查默认接收偏好。',
    sender: '通知中心',
    time: '今天 09:10',
    status: '信息',
  },
];

export function MessageCenterPage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <header className={styles.toolbar}>
        <div>
          <Title3>消息中心测试页</Title3>
          <Caption1>验证活动流、筛选区、状态标签和消息密度在 Fluent 2 下的组合方式。</Caption1>
        </div>
        <div className={styles.actions}>
          <Button appearance="secondary">全部标记已读</Button>
          <Button appearance="primary">新建订阅</Button>
        </div>
      </header>

      <section className={styles.hero}>
        <LabBadgeRow>
          <Badge appearance="filled" color="brand">
            Message center
          </Badge>
          <Badge appearance="tint" color="warning">
            6 条待处理
          </Badge>
          <Badge appearance="tint" color="success">
            14 条系统通知
          </Badge>
        </LabBadgeRow>
        <div className={styles.heroRow}>
          <div className={styles.heroText}>
            <Subtitle2>把消息、任务和通知放进同一条信息流，先看优先级，再决定处理路径。</Subtitle2>
            <Body1>这一页重点验证消息密度、状态标签和快速筛选的层级，不让通知区把正文挤掉。</Body1>
          </div>
          <LabStatGrid
            items={[
              { label: '全部消息', value: '28', tone: 'brand' },
              { label: '待处理', value: '6', tone: 'warning' },
              { label: '系统通知', value: '14', tone: 'success' },
            ]}
          />
        </div>
      </section>

      <div className={styles.listShell}>
        <aside className={styles.rail}>
          <section className={styles.railCard}>
            <Subtitle2>快速筛选</Subtitle2>
            <Input placeholder="搜索消息或发送方" />
            <Divider />
            <div className={styles.filterList}>
              <div className={styles.filterItem}>
                <Body1Strong>全部消息</Body1Strong>
                <Badge appearance="filled">28</Badge>
              </div>
              <div className={styles.filterItem}>
                <Body1Strong>待处理</Body1Strong>
                <Badge appearance="tint" color="warning">
                  6
                </Badge>
              </div>
              <div className={styles.filterItem}>
                <Body1Strong>系统通知</Body1Strong>
                <Badge appearance="tint">14</Badge>
              </div>
            </div>
          </section>
        </aside>

        <section className={styles.content}>
          {messages.map(message => (
            <article key={message.title} className={styles.messageCard}>
              <div className={styles.messageHeader}>
                <div className={styles.messageMeta}>
                  <Avatar name={message.sender} color="brand" />
                  <div>
                    <Body1Strong>{message.title}</Body1Strong>
                    <Caption1>{message.sender}</Caption1>
                  </div>
                </div>
                <Badge
                  appearance="tint"
                  color={
                    message.status === '成功'
                      ? 'success'
                      : message.status === '待处理'
                        ? 'warning'
                        : 'informative'
                  }
                >
                  {message.status}
                </Badge>
              </div>
              <Body1 className={styles.messageText}>{message.body}</Body1>
              <div className={styles.messageFooter}>
                <Caption1>{message.time}</Caption1>
                <div className={styles.actions}>
                  <Button appearance="subtle">查看详情</Button>
                  <Button appearance="subtle">稍后处理</Button>
                </div>
              </div>
            </article>
          ))}
        </section>
      </div>
    </div>
  );
}
