import { Body1, Caption1, makeStyles, tokens } from '@fluentui/react-components'
import type { MessageTimelineItem } from '@/shared/types/message-center'

const useStyles = makeStyles({
  root: {
    display: 'grid',
    gap: '10px',
  },
  item: {
    display: 'grid',
    gridTemplateColumns: '10px minmax(0, 1fr)',
    gap: '10px',
    alignItems: 'start',
  },
  rail: {
    width: '10px',
    display: 'grid',
    justifyItems: 'center',
    gap: '6px',
    paddingTop: '4px',
  },
  dot: {
    width: '10px',
    height: '10px',
    borderRadius: '999px',
    backgroundColor: tokens.colorBrandBackground,
  },
  line: {
    width: '2px',
    minHeight: '24px',
    backgroundColor: tokens.colorNeutralStroke2,
  },
  content: {
    display: 'grid',
    gap: '4px',
    paddingBottom: '6px',
  },
  label: {
    color: tokens.colorNeutralForeground2,
  },
  value: {
    color: tokens.colorNeutralForeground3,
  },
})

function resolveToneColor(tone?: MessageTimelineItem['tone']) {
  switch (tone) {
    case 'success':
      return tokens.colorPaletteGreenBackground3
    case 'warning':
      return tokens.colorPaletteYellowBackground3
    case 'danger':
      return tokens.colorPaletteRedBackground3
    case 'neutral':
      return tokens.colorNeutralStrokeAccessible
    default:
      return tokens.colorBrandBackground
  }
}

export function DetailTimeline({ items }: { items: MessageTimelineItem[] }) {
  const styles = useStyles()

  return (
    <div className={styles.root}>
      {items.map((item, index) => (
        <div key={item.id} className={styles.item}>
          <div className={styles.rail}>
            <div className={styles.dot} style={{ backgroundColor: resolveToneColor(item.tone) }} />
            {index < items.length - 1 ? <div className={styles.line} /> : null}
          </div>
          <div className={styles.content}>
            <Body1 className={styles.label}>{item.label}</Body1>
            <Caption1 className={styles.value}>{item.value || '-'}</Caption1>
          </div>
        </div>
      ))}
    </div>
  )
}
