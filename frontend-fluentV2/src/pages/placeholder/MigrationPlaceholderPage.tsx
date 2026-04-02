import { Badge, Body1, makeStyles } from '@fluentui/react-components'
import { useLocation } from 'react-router-dom'
import { useRouteContext } from '@/features/navigation/navigation.service'
import { PageContainer } from '@/features/shell/components/PageContainer'
import type { RouteContext } from '@/shared/types/navigation'
import { SectionCard } from '@/shared/ui/SectionCard'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gap: '18px',
  },
})

export function MigrationPlaceholderPage({
  routeId,
  routeContext,
}: {
  routeId?: string
  routeContext?: RouteContext
}) {
  const styles = useStyles()
  const location = useLocation()
  const routeContextQuery = useRouteContext(routeId)
  const context = routeContext || routeContextQuery.context

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <RouterButtonLink appearance="primary" to="/welcome">
            返回首页
          </RouterButtonLink>
          <RouterButtonLink appearance="secondary" to="/system/menu">
            查看菜单浏览版
          </RouterButtonLink>
        </>
      }
    >
      <div className={styles.grid}>
        <SectionCard title="迁移占位说明" description="当前路由已经接入真实认证、空间和运行时导航，但业务内容尚未迁入 React 版本。">
          <Body1>该页面暂由统一占位页承接，避免运行时导航命中后出现白屏。</Body1>
        </SectionCard>
        <SectionCard title="当前上下文" description="占位页会展示当前路径、空间和运行时来源信息，便于第三版继续逐页迁移。">
          <Body1>页面标题：{context?.title || routeId || '未命名页面'}</Body1>
          <Body1>当前路径：{location.pathname}</Body1>
          <Body1>所属菜单空间：{context?.spaceKey || 'default'}</Body1>
          <Body1>导航来源：{context?.source || 'local'}</Body1>
          <Body1>页面状态：该页面尚未迁移到当前 React 前端。</Body1>
          {context?.pageKey ? <Body1>Page Key：{context.pageKey}</Body1> : null}
          {context?.permissionKey ? <Body1>Permission Key：{context.permissionKey}</Body1> : null}
          {context?.manageGroupName ? <Body1>管理分组：{context.manageGroupName}</Body1> : null}
          <Badge appearance="tint">后续只替换内容区，不推翻当前壳层</Badge>
        </SectionCard>
      </div>
    </PageContainer>
  )
}
