import { Badge, Body1, Button, Field, Input, MessageBar, MessageBarBody, Switch, makeStyles, tokens } from '@fluentui/react-components'
import { PageContainer } from '@/features/shell/components/PageContainer'
import { SectionCard } from '@/shared/ui/SectionCard'

const useStyles = makeStyles({
  grid: {
    display: 'grid',
    gridTemplateColumns: '1.1fr 0.9fr',
    gap: '18px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: '1fr',
    },
  },
  toolbar: {
    display: 'flex',
    gap: '10px',
    flexWrap: 'wrap',
  },
  filterGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '14px',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
    },
  },
  switches: {
    display: 'grid',
    gap: '12px',
  },
  previewTable: {
    display: 'grid',
    gap: '8px',
  },
  row: {
    display: 'grid',
    gridTemplateColumns: '1.2fr 0.8fr 0.6fr',
    gap: '12px',
    alignItems: 'center',
    padding: '12px 14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    '@media (max-width: 640px)': {
      gridTemplateColumns: '1fr',
    },
  },
  headerRow: {
    color: tokens.colorNeutralForeground3,
    fontSize: tokens.fontSizeBase200,
  },
})

export function SystemMenuPage({ routeId }: { routeId: string }) {
  const styles = useStyles()

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <div className={styles.toolbar}>
          <Button appearance="primary">创建菜单</Button>
          <Button appearance="secondary">更多操作</Button>
        </div>
      }
    >
      <MessageBar>
        <MessageBarBody>这里故意不接真实 API，只验证治理页的结构、节奏与扩展边界。</MessageBarBody>
      </MessageBar>

      <div className={styles.grid}>
        <SectionCard title="筛选与工具区" description="沿用 Vue 里仍然合理的结构职责：筛选区、开关区、工具区各自独立。">
          <div className={styles.filterGrid}>
            <Field label="菜单名称">
              <Input placeholder="例如：菜单管理" />
            </Field>
            <Field label="路由路径">
              <Input placeholder="例如：/system/menu" />
            </Field>
          </div>
          <div className={styles.switches}>
            <Switch label="显示隐藏菜单" defaultChecked />
            <Switch label="显示内嵌菜单" defaultChecked />
            <Switch label="显示启用菜单" defaultChecked />
          </div>
        </SectionCard>

        <SectionCard title="治理预览" description="当前只展示治理页骨架，不实现真实表格行为。">
          <div className={styles.previewTable}>
            <div className={`${styles.row} ${styles.headerRow}`}>
              <span>菜单名称</span>
              <span>所属分组</span>
              <span>状态</span>
            </div>
            {[
              ['菜单管理', '系统管理', '已实现'],
              ['页面管理', '系统管理', '待迁移'],
              ['功能包管理', '系统管理', '待迁移'],
            ].map((item) => (
              <div key={item[0]} className={styles.row}>
                <Body1>{item[0]}</Body1>
                <Badge appearance="tint">{item[1]}</Badge>
                <Badge appearance={item[2] === '已实现' ? 'filled' : 'outline'} color={item[2] === '已实现' ? 'success' : 'brand'}>
                  {item[2]}
                </Badge>
              </div>
            ))}
          </div>
        </SectionCard>
      </div>
    </PageContainer>
  )
}
