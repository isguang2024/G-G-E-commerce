import { Body1, makeStyles } from '@fluentui/react-components'
import { useLocation } from 'react-router-dom'
import { SectionCard } from '@/shared/ui/SectionCard'
import { RouterButtonLink } from '@/shared/ui/RouterButtonLink'

const useStyles = makeStyles({
  root: {
    display: 'grid',
    gap: '18px',
  },
})

export function NotFoundPage() {
  const styles = useStyles()
  const location = useLocation()

  return (
    <div className={styles.root}>
      <SectionCard
        title="页面未找到"
        description="当前路径没有匹配到系统中的路由定义。"
        actions={
          <>
            <RouterButtonLink appearance="primary" to="/welcome">
              返回首页
            </RouterButtonLink>
            <RouterButtonLink appearance="secondary" to="/workspace">
              打开工作台
            </RouterButtonLink>
          </>
        }
      >
        <Body1>未匹配路径：{location.pathname}</Body1>
      </SectionCard>
    </div>
  )
}
