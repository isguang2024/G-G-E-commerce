import { useEffect, useMemo, useState } from 'react'
import {
  Body1,
  Button,
  Caption1,
  Field,
  Input,
  MessageBar,
  MessageBarBody,
  Spinner,
  makeStyles,
  tokens,
} from '@fluentui/react-components'
import { bundleIcon, ChevronDown20Filled, ChevronDown20Regular, ChevronRight20Filled, ChevronRight20Regular } from '@fluentui/react-icons'
import { buildMenuNodeDetail, useMenuManageGroupsQuery, useMenuTreeQuery, useRuntimePagesQuery } from '@/features/menu/menu.service'
import { PageContainer } from '@/features/shell/components/PageContainer'
import { useShellStore } from '@/features/shell/store/useShellStore'
import type { MenuNode } from '@/shared/types/menu'
import { SectionCard } from '@/shared/ui/SectionCard'

const ChevronDownIcon = bundleIcon(ChevronDown20Filled, ChevronDown20Regular)
const ChevronRightIcon = bundleIcon(ChevronRight20Filled, ChevronRight20Regular)

const useStyles = makeStyles({
  layout: {
    display: 'grid',
    gridTemplateColumns: 'minmax(320px, 0.9fr) minmax(360px, 1.1fr)',
    gap: '18px',
    '@media (max-width: 1100px)': {
      gridTemplateColumns: '1fr',
    },
  },
  sidebarStack: {
    display: 'grid',
    gap: '16px',
  },
  detailStack: {
    display: 'grid',
    gap: '16px',
  },
  treeToolbar: {
    display: 'grid',
    gap: '12px',
  },
  treeList: {
    display: 'grid',
    gap: '6px',
  },
  treeButton: {
    width: '100%',
    justifyContent: 'flex-start',
    display: 'flex',
    gap: '8px',
    minHeight: '40px',
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
    width: '18px',
    flexShrink: 0,
  },
  treeMeta: {
    display: 'grid',
    gap: '2px',
    minWidth: 0,
  },
  treeLabel: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  treePath: {
    color: tokens.colorNeutralForeground3,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  fieldGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
    },
  },
  detailField: {
    display: 'grid',
    gap: '4px',
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  detailLabel: {
    color: tokens.colorNeutralForeground3,
  },
  detailValue: {
    wordBreak: 'break-word',
  },
  listBlock: {
    display: 'grid',
    gap: '8px',
  },
  listRow: {
    display: 'grid',
    gap: '2px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  emptyBlock: {
    padding: '12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    color: tokens.colorNeutralForeground3,
  },
})

function flattenTree(items: MenuNode[]): MenuNode[] {
  return items.flatMap((item) => [item, ...flattenTree(item.children)])
}

function filterTree(items: MenuNode[], keyword: string): MenuNode[] {
  const query = keyword.trim().toLowerCase()
  if (!query) {
    return items
  }

  return items.reduce<MenuNode[]>((result, item) => {
    const children = filterTree(item.children, query)
    const target = `${item.title} ${item.name} ${item.path} ${item.component}`.toLowerCase()
    if (target.includes(query) || children.length > 0) {
      result.push({
        ...item,
        children,
      })
    }
    return result
  }, [])
}

function TreeNode({
  item,
  level,
  selectedId,
  expandedIds,
  onSelect,
  onToggleExpand,
}: {
  item: MenuNode
  level: number
  selectedId: string | null
  expandedIds: Set<string>
  onSelect: (nodeId: string) => void
  onToggleExpand: (nodeId: string) => void
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
        <span className={styles.treeIndent} onClick={(event) => {
          if (!hasChildren) {
            return
          }
          event.stopPropagation()
          onToggleExpand(item.id)
        }}>
          {hasChildren ? (expanded ? <ChevronDownIcon /> : <ChevronRightIcon />) : null}
        </span>
        <div className={styles.treeMeta}>
          <span className={styles.treeLabel}>{item.title || item.name}</span>
          <Caption1 className={styles.treePath}>{item.path || '无路径'}</Caption1>
        </div>
      </button>
      {hasChildren && expanded
        ? item.children.map((child) => (
            <TreeNode
              key={child.id}
              item={child}
              level={level + 1}
              selectedId={selectedId}
              expandedIds={expandedIds}
              onSelect={onSelect}
              onToggleExpand={onToggleExpand}
            />
          ))
        : null}
    </>
  )
}

function ReadonlyField({ label, value }: { label: string; value: string }) {
  const styles = useStyles()
  return (
    <div className={styles.detailField}>
      <Caption1 className={styles.detailLabel}>{label}</Caption1>
      <Body1 className={styles.detailValue}>{value || '-'}</Body1>
    </div>
  )
}

export function SystemMenuPage({ routeId }: { routeId: string }) {
  const styles = useStyles()
  const currentSpaceKey = useShellStore((state) => state.currentSpaceKey)
  const menuTreeQuery = useMenuTreeQuery(currentSpaceKey)
  const runtimePagesQuery = useRuntimePagesQuery(currentSpaceKey)
  const manageGroupsQuery = useMenuManageGroupsQuery()
  const [keyword, setKeyword] = useState('')
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null)
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set())

  const filteredTree = useMemo(
    () => filterTree(menuTreeQuery.data || [], keyword),
    [keyword, menuTreeQuery.data],
  )
  const flattenedTree = useMemo(() => flattenTree(filteredTree), [filteredTree])
  const detail = useMemo(() => {
    const target = flattenedTree.find((item) => item.id === selectedNodeId) || flattenedTree[0]
    if (!target) {
      return null
    }
    return buildMenuNodeDetail(target, menuTreeQuery.data || [], runtimePagesQuery.data || [])
  }, [flattenedTree, menuTreeQuery.data, runtimePagesQuery.data, selectedNodeId])

  useEffect(() => {
    if (!filteredTree.length) {
      setSelectedNodeId(null)
      return
    }

    if (!selectedNodeId || !flattenedTree.some((item) => item.id === selectedNodeId)) {
      setSelectedNodeId(flattenedTree[0]?.id || null)
    }
  }, [filteredTree.length, flattenedTree, selectedNodeId])

  useEffect(() => {
    if (!filteredTree.length) {
      setExpandedIds(new Set())
      return
    }

    setExpandedIds(new Set(flattenTree(filteredTree).map((item) => item.id)))
  }, [currentSpaceKey, filteredTree])

  function toggleExpand(nodeId: string) {
    setExpandedIds((current) => {
      const next = new Set(current)
      if (next.has(nodeId)) {
        next.delete(nodeId)
      } else {
        next.add(nodeId)
      }
      return next
    })
  }

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <Button appearance="secondary" onClick={() => menuTreeQuery.refetch()}>
            刷新菜单树
          </Button>
          <Button disabled appearance="primary">
            第二版保持只读
          </Button>
        </>
      }
    >
      <div className={styles.layout}>
        <div className={styles.sidebarStack}>
          <SectionCard title="菜单树浏览" description="按当前菜单空间加载真实菜单树，搜索后保留父子结构。">
            <div className={styles.treeToolbar}>
              <Field label="搜索菜单">
                <Input placeholder="按名称、路径、组件搜索" value={keyword} onChange={(_, data) => setKeyword(data.value)} />
              </Field>
              <Body1>当前空间：{currentSpaceKey}</Body1>
            </div>
            {menuTreeQuery.isLoading ? <Spinner label="正在加载菜单树" /> : null}
            {menuTreeQuery.isError ? (
              <MessageBar intent="error">
                <MessageBarBody>菜单树加载失败，请检查后端菜单接口。</MessageBarBody>
              </MessageBar>
            ) : null}
            {!menuTreeQuery.isLoading && !menuTreeQuery.isError && filteredTree.length === 0 ? (
              <div className={styles.emptyBlock}>当前空间没有匹配的菜单节点。</div>
            ) : (
              <div className={styles.treeList}>
                {filteredTree.map((item) => (
                  <TreeNode
                    key={item.id}
                    item={item}
                    level={0}
                    selectedId={selectedNodeId}
                    expandedIds={expandedIds}
                    onSelect={setSelectedNodeId}
                    onToggleExpand={toggleExpand}
                  />
                ))}
              </div>
            )}
          </SectionCard>
        </div>

        <div className={styles.detailStack}>
          <SectionCard title="菜单详情" description="第二版只提供只读详情，用于核对真实结构、页面绑定和权限上下文。">
            {!detail ? (
              <div className={styles.emptyBlock}>请选择左侧菜单节点查看详情。</div>
            ) : (
              <>
                <div className={styles.fieldGrid}>
                  <ReadonlyField label="菜单名称" value={detail.name} />
                  <ReadonlyField label="标题" value={detail.title} />
                  <ReadonlyField label="路径" value={detail.path} />
                  <ReadonlyField label="组件" value={detail.component || '-'} />
                  <ReadonlyField label="类型" value={detail.kind} />
                  <ReadonlyField label="图标" value={detail.icon || '-'} />
                  <ReadonlyField label="排序" value={`${detail.sortOrder}`} />
                  <ReadonlyField label="隐藏状态" value={detail.hidden ? '隐藏' : '显示'} />
                  <ReadonlyField label="所属空间" value={detail.spaceKey} />
                  <ReadonlyField label="管理分组" value={detail.manageGroup?.name || '-'} />
                  <ReadonlyField label="父级菜单" value={detail.parent ? `${detail.parent.title} (${detail.parent.path || '-'})` : '顶级菜单'} />
                  <ReadonlyField label="子节点数量" value={`${detail.childCount}`} />
                </div>
                <div className={styles.listBlock}>
                  <Body1>权限要求</Body1>
                  {detail.permissionKeys.length ? (
                    detail.permissionKeys.map((item) => (
                      <div key={item} className={styles.listRow}>
                        <Body1>{item}</Body1>
                      </div>
                    ))
                  ) : (
                    <div className={styles.emptyBlock}>当前节点未解析到显式权限键。</div>
                  )}
                </div>
              </>
            )}
          </SectionCard>

          <SectionCard title="关联页面信息" description="展示当前菜单关联的受管页面与访问模式，作为第三版编辑能力的基础。">
            {runtimePagesQuery.isLoading ? <Spinner label="正在加载页面绑定" /> : null}
            {detail?.linkedPages.length ? (
              detail.linkedPages.map((item) => (
                <div key={item.pageKey} className={styles.listRow}>
                  <Body1>{item.name || item.pageKey}</Body1>
                  <Caption1>Page Key：{item.pageKey}</Caption1>
                  <Caption1>路由：{item.routePath}</Caption1>
                  <Caption1>组件：{item.component || '-'}</Caption1>
                  <Caption1>访问模式：{item.accessMode || '-'}</Caption1>
                  <Caption1>权限键：{item.permissionKey || '-'}</Caption1>
                </div>
              ))
            ) : (
              <div className={styles.emptyBlock}>当前节点没有关联的受管页面。</div>
            )}
          </SectionCard>

          <SectionCard title="Meta 与分组" description="保留核心只读信息，避免把后端原始字段散落到页面其他区域。">
            <div className={styles.listBlock}>
              <Body1>菜单分组</Body1>
              {manageGroupsQuery.data?.length ? (
                manageGroupsQuery.data.map((item) => (
                  <div key={item.id} className={styles.listRow}>
                    <Body1>{item.name}</Body1>
                    <Caption1>排序：{item.sortOrder}</Caption1>
                    <Caption1>状态：{item.status}</Caption1>
                  </div>
                ))
              ) : (
                <div className={styles.emptyBlock}>当前未返回菜单管理分组，页面会继续按普通菜单树展示。</div>
              )}
            </div>
            <div className={styles.listBlock}>
              <Body1>Meta 核心字段</Body1>
              {detail ? (
                Object.entries(detail.meta).length ? (
                  Object.entries(detail.meta).map(([key, value]) => (
                    <div key={key} className={styles.listRow}>
                      <Body1>{key}</Body1>
                      <Caption1>{typeof value === 'string' ? value : JSON.stringify(value)}</Caption1>
                    </div>
                  ))
                ) : (
                  <div className={styles.emptyBlock}>当前节点没有额外 meta 字段。</div>
                )
              ) : null}
            </div>
          </SectionCard>
        </div>
      </div>
    </PageContainer>
  )
}
