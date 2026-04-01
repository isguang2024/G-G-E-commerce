import { Badge, Body1, makeStyles } from '@fluentui/react-components'
import { PageContainer } from '@/features/shell/components/PageContainer'
import { SectionCard } from '@/shared/ui/SectionCard'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(3, minmax(0, 1fr))',
    gap: '18px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: '1fr',
    },
  },
  item: {
    display: 'grid',
    gap: '8px',
  },
})

export function SystemHomePage({ routeId }: { routeId: string }) {
  const styles = useStyles()

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <RouterButtonLink appearance="primary" to="/system/menu">
            进入菜单管理
          </RouterButtonLink>
          <RouterButtonLink appearance="secondary" to="/system/page">
            查看页面管理占位
          </RouterButtonLink>
        </>
      }
    >
      <div className={styles.grid}>
        <SectionCard title="导航治理" description="菜单、页面、接口和功能包仍然保留为治理主线，只是先统一用占位页承接。">
          <div className={styles.item}>
            <Badge appearance="filled" color="brand">
              当前重点
            </Badge>
            <Body1>菜单管理已落地为首个治理型示例页，用来验证标题区、操作区和内容区节奏。</Body1>
          </div>
        </SectionCard>
        <SectionCard title="角色与用户" description="权限体系先保留导航位置和 metadata，后续再接真实 API 与列表页模式。">
          <Body1>不在本期接入真实登录、真实权限和菜单裁剪。</Body1>
        </SectionCard>
        <SectionCard title="后续承接方式" description="每个治理页都应该只替换内容区实现，而不是重写整套容器。">
          <Body1>首期把 Provider、路由、导航、主题和占位体系先固定下来。</Body1>
        </SectionCard>
      </div>
    </PageContainer>
  )
}
