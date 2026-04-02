import type { ReactNode } from 'react'
import { Body1, Button, makeStyles, tokens } from '@fluentui/react-components'
import { useNavigate } from 'react-router-dom'
import { SectionCard } from '@/shared/ui/SectionCard'

const useStyles = makeStyles({
  root: {
    minHeight: 'calc(100vh - 180px)',
    display: 'grid',
    alignContent: 'center',
  },
  code: {
    fontSize: '64px',
    lineHeight: '64px',
    fontWeight: tokens.fontWeightSemibold,
    color: tokens.colorBrandForeground1,
  },
  actions: {
    display: 'flex',
    gap: '12px',
    flexWrap: 'wrap',
  },
})

export function StatusPage({
  code,
  title,
  description,
  extra,
  primaryLabel = '返回首页',
}: {
  code: string
  title: string
  description: string
  extra?: ReactNode
  primaryLabel?: string
}) {
  const styles = useStyles()
  const navigate = useNavigate()

  return (
    <div className={styles.root}>
      <SectionCard
        title={title}
        description={description}
        actions={
          <div className={styles.actions}>
            <Button appearance="primary" onClick={() => navigate('/dashboard/console')}>
              {primaryLabel}
            </Button>
            <Button appearance="secondary" onClick={() => navigate(-1)}>
              返回上一页
            </Button>
          </div>
        }
      >
        <div className={styles.code}>{code}</div>
        {extra ? extra : <Body1>如果当前页面来自运行时菜单，请检查本地路由实现和后端页面配置。</Body1>}
      </SectionCard>
    </div>
  )
}
