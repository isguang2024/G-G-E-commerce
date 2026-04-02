import { Badge, Body1, Button, makeStyles, tokens } from '@fluentui/react-components'
import { ArrowClockwise20Regular } from '@fluentui/react-icons'
import { MetricGrid } from '@/shared/ui/MetricGrid'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'
import { SectionCard } from '@/shared/ui/SectionCard'
import type { DashboardSummary } from '@/shared/types/admin'

const useStyles = makeStyles({
  hero: {
    display: 'grid',
    gap: '12px',
    padding: '24px',
    borderRadius: tokens.borderRadiusXLarge,
    background: `linear-gradient(135deg, ${tokens.colorBrandBackground2} 0%, ${tokens.colorNeutralBackground1} 52%, ${tokens.colorNeutralBackground2} 100%)`,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  heroTitle: {
    margin: 0,
    fontSize: tokens.fontSizeHero800,
    lineHeight: tokens.lineHeightHero800,
    fontWeight: tokens.fontWeightSemibold,
  },
  actionGrid: {
    display: 'grid',
    gap: '12px',
  },
  actionRow: {
    display: 'flex',
    gap: '12px',
    flexWrap: 'wrap',
  },
  list: {
    display: 'grid',
    gap: '10px',
  },
  listItem: {
    display: 'grid',
    gap: '6px',
    padding: '14px 16px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  softText: {
    color: tokens.colorNeutralForeground3,
  },
})

export function AboutProjectModule({ summary }: { summary: DashboardSummary }) {
  const styles = useStyles()

  return (
    <div className={styles.hero}>
      <Badge appearance="filled" color="brand">
        第 8 版全量收口
      </Badge>
      <h2 className={styles.heroTitle}>用同一套 Fluent 2 工作台承接控制台、治理台和协作区。</h2>
      <Body1 className={styles.softText}>
        当前操作用户为 {summary.currentUserName}，工作空间为 {summary.currentSpaceLabel}。本版目标是把 Vue 全部页面与模块能力收口到 React。
      </Body1>
    </div>
  )
}

export function DynamicStatsModule({ metrics }: { metrics: Parameters<typeof MetricGrid>[0]['metrics'] }) {
  return <MetricGrid metrics={metrics} />
}

export function ActiveUserModule({ summary }: { summary: DashboardSummary }) {
  const styles = useStyles()

  return (
    <SectionCard title="活跃上下文" description="当前登录用户、空间与运行时导航上下文由真实链路驱动。">
      <div className={styles.list}>
        <div className={styles.listItem}>
          <Body1>当前用户</Body1>
          <Body1 className={styles.softText}>{summary.currentUserName}</Body1>
        </div>
        <div className={styles.listItem}>
          <Body1>当前空间</Body1>
          <Body1 className={styles.softText}>{summary.currentSpaceLabel}</Body1>
        </div>
      </div>
    </SectionCard>
  )
}

export function CardListModule() {
  const styles = useStyles()

  return (
    <SectionCard title="本期重点页面" description="优先保证首页、收件箱和系统治理链路可直接进入。">
      <div className={styles.list}>
        <div className={styles.listItem}>
          <Body1>工作区收件箱</Body1>
          <Body1 className={styles.softText}>统一查看未读通知、直接消息和待办动作，详情区直接连接真实收件链路。</Body1>
        </div>
        <div className={styles.listItem}>
          <Body1>系统治理主链</Body1>
          <Body1 className={styles.softText}>菜单、页面、接口、权限、功能包、角色和用户统一收口到同一套工作台模式。</Body1>
        </div>
        <div className={styles.listItem}>
          <Body1>团队与消息协作</Body1>
          <Body1 className={styles.softText}>系统域与团队域复用同一批消息模块，差异只体现在作用域和可用选项。</Body1>
        </div>
      </div>
    </SectionCard>
  )
}

export function NewUserModule({ summary }: { summary: DashboardSummary }) {
  const styles = useStyles()

  return (
    <SectionCard title="待跟进规模" description="用当前运行时上下文快速判断本版收口的范围。">
      <div className={styles.list}>
        <div className={styles.listItem}>
          <Body1>可见导航</Body1>
          <Body1 className={styles.softText}>{summary.visibleMenuCount} 个运行时入口</Body1>
        </div>
        <div className={styles.listItem}>
          <Body1>受管页面</Body1>
          <Body1 className={styles.softText}>{summary.managedPageCount} 条数据库注册页</Body1>
        </div>
      </div>
    </SectionCard>
  )
}

export function SalesOverviewModule({
  onRefresh,
}: {
  onRefresh: () => void
}) {
  const styles = useStyles()

  return (
    <SectionCard
      title="工作台动作"
      description="围绕真实入口组织高频动作，低频动作继续收入各治理页内部。"
      actions={
        <Button appearance="subtle" icon={<ArrowClockwise20Regular />} onClick={onRefresh}>
          刷新摘要
        </Button>
      }
    >
      <div className={styles.actionGrid}>
        <div className={styles.actionRow}>
          <RouterButtonLink appearance="primary" to="/workspace/inbox">
            打开收件箱
          </RouterButtonLink>
          <RouterButtonLink appearance="secondary" to="/system/menu">
            打开菜单治理
          </RouterButtonLink>
        </div>
        <div className={styles.actionRow}>
          <RouterButtonLink appearance="secondary" to="/system/page">
            页面治理
          </RouterButtonLink>
          <RouterButtonLink appearance="secondary" to="/system/api-endpoint">
            接口治理
          </RouterButtonLink>
        </div>
      </div>
    </SectionCard>
  )
}

export function TodoListModule({ summary }: { summary: DashboardSummary }) {
  const styles = useStyles()

  return (
    <SectionCard title="待办与快捷入口" description="消息、快捷入口和快速链接在控制台首屏统一可见。">
      <div className={styles.list}>
        <div className={styles.listItem}>
          <Body1>未读消息</Body1>
          <Body1 className={styles.softText}>{summary.unreadInboxCount} 条</Body1>
        </div>
        <div className={styles.listItem}>
          <Body1>快捷应用</Body1>
          <Body1 className={styles.softText}>{summary.fastEntryCount} 项已启用</Body1>
        </div>
        <div className={styles.listItem}>
          <Body1>快速链接</Body1>
          <Body1 className={styles.softText}>{summary.quickLinkCount} 条已启用</Body1>
        </div>
      </div>
    </SectionCard>
  )
}
