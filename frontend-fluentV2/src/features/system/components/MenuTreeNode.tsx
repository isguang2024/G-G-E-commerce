import type { ReactNode } from 'react'
import { Badge, Caption1, makeStyles, tokens } from '@fluentui/react-components'
import type { MenuNode } from '@/shared/types/menu'

const useStyles = makeStyles({
  treeButton: {
    width: '100%',
    justifyContent: 'flex-start',
    display: 'flex',
    gap: '10px',
    minHeight: '48px',
    textAlign: 'left',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    padding: '8px 12px',
    color: tokens.colorNeutralForeground2,
  },
  treeButtonActive: {
    backgroundColor: tokens.colorBrandBackground2,
    border: `1px solid ${tokens.colorBrandStroke1}`,
    color: tokens.colorBrandForeground1,
  },
  treeIndent: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    width: '18px',
    flexShrink: 0,
  },
  treeMeta: {
    display: 'grid',
    gap: '4px',
    minWidth: 0,
    flex: 1,
  },
  treeLabel: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  treeSubline: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '6px',
    alignItems: 'center',
  },
  treePath: {
    color: tokens.colorNeutralForeground3,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
})

export function MenuReadonlyField({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ display: 'grid', gap: 4, padding: 12, borderRadius: tokens.borderRadiusLarge, backgroundColor: tokens.colorNeutralBackground2 }}>
      <Caption1 style={{ color: tokens.colorNeutralForeground3 }}>{label}</Caption1>
      <div style={{ wordBreak: 'break-word' }}>{value || '-'}</div>
    </div>
  )
}

export function MenuTreeNode({
  item,
  level,
  selectedId,
  expandedIds,
  onSelect,
  onToggleExpand,
  expandIcon,
}: {
  item: MenuNode
  level: number
  selectedId: string | null
  expandedIds: Set<string>
  onSelect: (nodeId: string) => void
  onToggleExpand: (nodeId: string) => void
  expandIcon: { expanded: ReactNode; collapsed: ReactNode }
}) {
  const styles = useStyles()
  const hasChildren = item.children.length > 0
  const expanded = expandedIds.has(item.id)
  const isActive = selectedId === item.id

  return (
    <>
      <button
        className={isActive ? `${styles.treeButton} ${styles.treeButtonActive}` : styles.treeButton}
        style={{ paddingLeft: `${12 + level * 18}px` }}
        type="button"
        onClick={() => onSelect(item.id)}
      >
        <span
          className={styles.treeIndent}
          onClick={(event) => {
            if (!hasChildren) return
            event.stopPropagation()
            onToggleExpand(item.id)
          }}
        >
          {hasChildren ? (expanded ? expandIcon.expanded : expandIcon.collapsed) : null}
        </span>
        <div className={styles.treeMeta}>
          <span className={styles.treeLabel}>{item.title || item.name}</span>
          <div className={styles.treeSubline}>
            <Caption1 className={styles.treePath}>{item.path || '无路径'}</Caption1>
            <Badge appearance="tint" size="small">
              {item.kind}
            </Badge>
            {item.hidden ? (
              <Badge color="important" appearance="outline" size="small">
                隐藏
              </Badge>
            ) : null}
            {item.manageGroup?.name ? (
              <Badge appearance="outline" size="small">
                {item.manageGroup.name}
              </Badge>
            ) : null}
          </div>
        </div>
      </button>
      {hasChildren && expanded
        ? item.children.map((child) => (
            <MenuTreeNode
              key={child.id}
              item={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggleExpand={onToggleExpand}
              expandIcon={expandIcon}
            />
          ))
        : null}
    </>
  )
}
