import { Body1, Caption1, makeStyles, tokens } from '@fluentui/react-components'
import type { RelationSourceRecord } from '@/shared/types/admin'

const useStyles = makeStyles({
  root: {
    display: 'grid',
    gap: '10px',
  },
  item: {
    display: 'grid',
    gap: '4px',
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  title: {
    color: tokens.colorNeutralForeground2,
  },
  caption: {
    color: tokens.colorNeutralForeground3,
  },
})

export function RelationSourceList({
  items,
  entityLabel = '实体',
}: {
  items: RelationSourceRecord[]
  entityLabel?: string
}) {
  const styles = useStyles()

  return (
    <div className={styles.root}>
      {items.map((item) => (
        <div key={item.entityId} className={styles.item}>
          <Body1 className={styles.title}>
            {entityLabel}：{item.entityId}
          </Body1>
          <Caption1 className={styles.caption}>
            来源功能包：{item.packageIds.length ? item.packageIds.join('、') : '无'}
          </Caption1>
        </div>
      ))}
    </div>
  )
}
