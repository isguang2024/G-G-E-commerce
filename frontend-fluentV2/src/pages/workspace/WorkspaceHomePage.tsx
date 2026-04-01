import { Badge, Body1, Button, makeStyles, tokens } from '@fluentui/react-components'
import { PageContainer } from '@/features/shell/components/PageContainer'
import { SectionCard } from '@/shared/ui/SectionCard'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  layout: {
    display: 'grid',
    gridTemplateColumns: '1.3fr 0.7fr',
    gap: '18px',
    '@media (max-width: 1024px)': {
      gridTemplateColumns: '1fr',
    },
  },
  focusList: {
    display: 'grid',
    gap: '12px',
  },
  focusItem: {
    display: 'grid',
    gap: '6px',
    padding: '16px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  split: {
    display: 'grid',
    gap: '16px',
  },
})

export function WorkspaceHomePage({ routeId }: { routeId: string }) {
  const styles = useStyles()

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <RouterButtonLink appearance="primary" to="/workspace/inbox">
            打开占位页
          </RouterButtonLink>
          <Button appearance="secondary">刷新 mock</Button>
        </>
      }
    >
      <div className={styles.layout}>
        <SectionCard title="当前工作面" description="这里模拟 Vue 工作台里常见的标题区 + 主工作区 + 次级说明区结构。">
          <div className={styles.focusList}>
            <div className={styles.focusItem}>
              <Badge appearance="tint" color="success">
                已实现
              </Badge>
              <Body1>工作台首页容器</Body1>
              <Body1>页面标题区、主卡片、按钮动作和次级说明已经归入统一壳层。</Body1>
            </div>
            <div className={styles.focusItem}>
              <Badge appearance="tint">后续接入</Badge>
              <Body1>收件中心、消息流和待办链路</Body1>
              <Body1>当前只保留入口位置和占位页，不接消息 API。</Body1>
            </div>
          </div>
        </SectionCard>

        <div className={styles.split}>
          <SectionCard title="当前上下文" description="菜单空间和当前上下文会留在壳层，不下沉到业务页内部。">
            <Body1>后续接真实租户、团队和消息链路时，只替换数据来源。</Body1>
          </SectionCard>
          <SectionCard title="推荐下一步" description="迁移工作台类页面时优先复用这个容器，而不是重写页面头部。">
            <Body1>建议下一批优先迁移：消息中心、团队总览、团队成员。</Body1>
          </SectionCard>
        </div>
      </div>
    </PageContainer>
  )
}
