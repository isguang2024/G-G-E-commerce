import { Badge, Body1, Caption1, makeStyles, tokens } from '@fluentui/react-components'
import { PageContainer } from '@/features/shell/components/PageContainer'
import { SectionCard } from '@/shared/ui/SectionCard'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  heroGrid: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr',
    gap: '18px',
    '@media (max-width: 1024px)': {
      gridTemplateColumns: '1fr',
    },
  },
  banner: {
    display: 'grid',
    gap: '14px',
    padding: '24px',
    borderRadius: tokens.borderRadiusXLarge,
    background: `linear-gradient(135deg, ${tokens.colorBrandBackground2} 0%, ${tokens.colorNeutralBackground1} 100%)`,
    boxShadow: tokens.shadow8,
  },
  statGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
    },
  },
  stat: {
    display: 'grid',
    gap: '4px',
    padding: '16px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  statValue: {
    fontSize: tokens.fontSizeHero700,
    lineHeight: tokens.lineHeightHero700,
    fontWeight: tokens.fontWeightSemibold,
  },
  linkList: {
    display: 'grid',
    gap: '10px',
  },
})

export function WelcomePage({ routeId }: { routeId: string }) {
  const styles = useStyles()

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <RouterButtonLink appearance="primary" to="/workspace">
            查看工作台壳层
          </RouterButtonLink>
          <RouterButtonLink appearance="secondary" to="/system/menu">
            查看菜单管理占位
          </RouterButtonLink>
        </>
      }
    >
      <div className={styles.heroGrid}>
        <div className={styles.banner}>
          <Badge appearance="filled" color="brand">
            业务基础壳
          </Badge>
          <Body1>
            当前实验线不复刻 Vue 页面，只保留信息架构、菜单空间、导航分组和页面工作区这些未来仍然成立的结构职责。
          </Body1>
          <div className={styles.statGrid}>
            <div className={styles.stat}>
              <Caption1>首期已落地</Caption1>
              <div className={styles.statValue}>4</div>
              <Body1>示例页面</Body1>
            </div>
            <div className={styles.stat}>
              <Caption1>保留结构</Caption1>
              <div className={styles.statValue}>5</div>
              <Body1>壳层职责</Body1>
            </div>
            <div className={styles.stat}>
              <Caption1>当前策略</Caption1>
              <div className={styles.statValue}>Mock</div>
              <Body1>先稳住壳层和数据边界</Body1>
            </div>
          </div>
        </div>

        <SectionCard title="本期约束" description="只做工程骨架、路由壳、导航、主题和占位页，不接入真实业务能力。">
          <div className={styles.linkList}>
            <Body1>不修改现有 `frontend/` Vue 工程。</Body1>
            <Body1>不复制旧目录、不翻译旧页面、不接真实 API。</Body1>
            <Body1>后续只替换 adapter 和页面实现，不推翻壳层。</Body1>
          </div>
        </SectionCard>
      </div>

      <SectionCard title="建议迁移顺序" description="先迁壳层稳定收益高的模块，再逐步接入真实数据。">
        <div className={styles.linkList}>
          <Body1>1. 工作台与系统首页的标题区、容器和卡片布局。</Body1>
          <Body1>2. 系统治理链路中的菜单、页面、接口管理页。</Body1>
          <Body1>3. 团队与消息域的三栏/列表/详情型工作区。</Body1>
        </div>
      </SectionCard>
    </PageContainer>
  )
}
