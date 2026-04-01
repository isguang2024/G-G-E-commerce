import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Input,
  ProgressBar,
  Subtitle2,
  Title3,
  makeStyles,
  tokens,
} from '@fluentui/react-components';
import { LabBadgeRow, LabStatGrid, LabSurfaceCard } from '../lab/primitives';

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '18px',
  },
  hero: {
    display: 'grid',
    gap: '14px',
    padding: '20px',
    borderRadius: tokens.borderRadiusXLarge,
    background:
      'linear-gradient(135deg, rgba(15,108,189,0.08) 0%, rgba(15,108,189,0.02) 56%, rgba(255,255,255,0.02) 100%)',
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  toolbar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '12px',
    flexWrap: 'wrap',
  },
  actionRow: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
  },
  filters: {
    display: 'grid',
    gridTemplateColumns: '280px 220px 220px',
    gap: '12px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '18px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: '1fr',
    },
  },
  queue: {
    display: 'grid',
    gap: '12px',
  },
  card: {
    display: 'grid',
    gap: '12px',
  },
  cardHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'start',
    gap: '12px',
  },
  detailPanel: {
    display: 'grid',
    gap: '14px',
    alignContent: 'start',
  },
  infoBlock: {
    display: 'grid',
    gap: '6px',
  },
  progressMeta: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '12px',
    alignItems: 'center',
  },
  queueHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: '12px',
  },
});

const items = [
  {
    title: '发布策略调整审批',
    summary: '需要确认消息中心默认订阅范围是否推送到北区团队空间。',
    status: '待审批',
    owner: '消息治理组',
  },
  {
    title: '导航可见性变更',
    summary: '请求将频道工作台入口暴露给试点团队成员。',
    status: '复核中',
    owner: '前端实验线',
  },
  {
    title: '风险告警规则更新',
    summary: '拟提高高风险消息的默认醒目程度并增加二次确认。',
    status: '待审批',
    owner: '平台安全组',
  },
];

export function ApprovalWorkbenchPage() {
  const styles = useStyles();

  return (
    <div className={styles.page}>
      <header className={styles.toolbar}>
        <div>
          <Title3>审批工作台测试页</Title3>
          <Caption1>偏 Fluent 2 Web 的治理页表达，验证筛选区、队列区和右侧详情区的稳定层级。</Caption1>
        </div>
        <div className={styles.actionRow}>
          <Button appearance="secondary">导出队列</Button>
          <Button appearance="primary">创建审批</Button>
        </div>
      </header>

      <section className={styles.hero}>
        <LabBadgeRow>
          <Badge appearance="filled" color="brand">
            Governance workbench
          </Badge>
          <Badge appearance="tint" color="warning">
            6 项待处理
          </Badge>
          <Badge appearance="tint" color="success">
            SLA 92%
          </Badge>
        </LabBadgeRow>
        <Subtitle2>审批线仍然使用同一套 Fluent 组件，但页面语法偏向治理和风险判断，而不是协作流。</Subtitle2>
        <div className={styles.progressMeta}>
          <Caption1>本日队列处理进度</Caption1>
          <Caption1>17 / 24</Caption1>
        </div>
        <ProgressBar value={0.71} />
        <LabStatGrid
          items={[
            { label: '待审批', value: '6', tone: 'warning' },
            { label: '复核中', value: '3', tone: 'brand' },
            { label: '已关闭', value: '11', tone: 'success' },
          ]}
        />
      </section>

      <section className={styles.filters}>
        <Input placeholder="搜索审批主题或申请人" />
        <Input placeholder="审批状态" />
        <Input placeholder="所属模块" />
      </section>

      <div className={styles.layout}>
        <section className={styles.queue}>
          <div className={styles.queueHeader}>
            <Body1Strong>待处理队列</Body1Strong>
            <Badge appearance="tint" color="warning">
              3 条
            </Badge>
          </div>
          {items.map(item => (
            <LabSurfaceCard key={item.title}>
              <article className={styles.card}>
              <div className={styles.cardHeader}>
                <div className={styles.infoBlock}>
                  <Body1Strong>{item.title}</Body1Strong>
                  <Caption1>{item.owner}</Caption1>
                </div>
                <Badge appearance="tint" color={item.status === '待审批' ? 'warning' : 'brand'}>
                  {item.status}
                </Badge>
              </div>
              <Body1>{item.summary}</Body1>
              <div className={styles.actionRow}>
                <Button appearance="subtle">查看详情</Button>
                <Button appearance="subtle">转交</Button>
              </div>
              </article>
            </LabSurfaceCard>
          ))}
        </section>

        <LabSurfaceCard subtle>
          <aside className={styles.detailPanel}>
          <Body1Strong>当前选中项</Body1Strong>
          <Caption1>发布策略调整审批</Caption1>

          <div className={styles.infoBlock}>
            <Caption1>影响范围</Caption1>
            <Body1>北区团队空间、消息中心默认配置、实验场通知样式。</Body1>
          </div>

          <div className={styles.infoBlock}>
            <Caption1>风险说明</Caption1>
            <Body1>如果直接发布，可能导致试点团队收到额外系统消息，需要先确认订阅边界。</Body1>
          </div>

          <div className={styles.actionRow}>
            <Button appearance="primary">批准</Button>
            <Button appearance="secondary">驳回</Button>
            <Button appearance="subtle">请求补充信息</Button>
          </div>
          </aside>
        </LabSurfaceCard>
      </div>
    </div>
  );
}
