import { Badge, Body1, makeStyles } from '@fluentui/react-components'
import { useLocation } from 'react-router-dom'
import { PageContainer } from '@/features/shell/components/PageContainer'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { useSpacesQuery } from '@/features/navigation/navigation.service'
import { getRouteDefinition } from '@/features/navigation/route-registry'
import { SectionCard } from '@/shared/ui/SectionCard'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gap: '18px',
  },
})

export function MigrationPlaceholderPage({ routeId }: { routeId: string }) {
  const styles = useStyles()
  const location = useLocation()
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const spacesQuery = useSpacesQuery()
  const currentSpace = spacesQuery.data?.find((item) => item.key === currentSpaceKey)
  const routeDefinition = getRouteDefinition(routeId)

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <RouterButtonLink appearance="primary" to="/welcome">
            返回首页
          </RouterButtonLink>
          <RouterButtonLink appearance="secondary" to="/system/menu">
            查看已实现示例
          </RouterButtonLink>
        </>
      }
    >
      <div className={styles.grid}>
        <SectionCard title="迁移占位说明" description="这个路由位置已经在当前壳层中注册，但业务内容尚未迁入。">
          <Body1>该页面尚未迁移到当前前端。</Body1>
        </SectionCard>
        <SectionCard title="当前上下文" description="占位页也要明确展示路径、导航组和当前菜单空间，避免壳层失去上下文。">
          <Body1>页面标题：{routeDefinition?.shellTitle || routeId}</Body1>
          <Body1>路由路径：{location.pathname}</Body1>
          <Body1>所属导航组：{routeDefinition?.group || 'unknown'}</Body1>
          <Body1>当前菜单空间：{currentSpace?.label || currentSpaceKey}</Body1>
          <Badge appearance="tint">后续只替换内容区，不推翻壳层</Badge>
        </SectionCard>
      </div>
    </PageContainer>
  )
}
