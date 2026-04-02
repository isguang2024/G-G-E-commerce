import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  type MouseEvent as ReactMouseEvent,
  type PointerEvent as ReactPointerEvent,
} from 'react'
import { Button, Text, Tooltip, makeStyles, mergeClasses, tokens } from '@fluentui/react-components'
import {
  Dismiss16Regular,
  DismissSquareMultiple16Regular,
  Pin16Regular,
  PinOff16Regular,
} from '@fluentui/react-icons'
import type { ShellTab } from '@/shared/types/navigation'

type ContextMenuState = {
  path: string
  x: number
  y: number
}

type MeasuredRect = {
  left: number
  right: number
  width: number
}

type DragCandidate = {
  pointerId: number
  path: string
  label: string
  pinned: boolean
  width: number
  height: number
  sourceLeft: number
  sourceTop: number
  startClientX: number
  startClientY: number
  grabOffsetX: number
  grabOffsetY: number
  startedAt: number
}

type ActiveDragState = DragCandidate & {
  currentClientX: number
  currentClientY: number
}

const DRAG_THRESHOLD_PX = 4
const TAB_FLIP_DURATION_MS = 160
const TAB_FLIP_DRAG_DURATION_MS = 90
const DRAG_FLIP_WARMUP_MS = 140
const TAB_FLIP_EASING = 'cubic-bezier(0.33, 0, 0.67, 1)'
const TAB_MIN_WIDTH = '108px'

const useStyles = makeStyles({
  root: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: '10px',
    minHeight: '40px',
    padding: '0 8px',
    borderBottom: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: 'transparent',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
      paddingLeft: '6px',
      paddingRight: '6px',
    },
  },
  rootDragging: {
    cursor: 'grabbing',
    '& *': {
      cursor: 'grabbing !important',
    },
  },
  scrollerWrap: {
    position: 'relative',
    minWidth: 0,
  },
  scrollerFadeLeft: {
    '::before': {
      content: '""',
      position: 'absolute',
      inset: '0 auto 0 0',
      width: '24px',
      backgroundImage: `linear-gradient(90deg, ${tokens.colorNeutralBackground3} 12%, transparent 100%)`,
      pointerEvents: 'none',
      zIndex: 2,
    },
  },
  scrollerFadeRight: {
    '::after': {
      content: '""',
      position: 'absolute',
      inset: '0 0 0 auto',
      width: '28px',
      backgroundImage: `linear-gradient(270deg, ${tokens.colorNeutralBackground3} 18%, transparent 100%)`,
      pointerEvents: 'none',
      zIndex: 2,
    },
  },
  rail: {
    display: 'flex',
    alignItems: 'flex-end',
    gap: '4px',
    overflowX: 'auto',
    overflowY: 'hidden',
    paddingTop: '4px',
    paddingLeft: '2px',
    paddingRight: '2px',
    scrollbarWidth: 'thin',
    scrollBehavior: 'smooth',
  },
  tabWrap: {
    position: 'relative',
    flex: '0 0 auto',
    display: 'flex',
    alignItems: 'stretch',
    paddingTop: '2px',
    paddingBottom: '0',
    willChange: 'transform',
  },
  tab: {
    position: 'relative',
    flex: '0 0 auto',
    display: 'grid',
    gridTemplateColumns: 'auto minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: '6px',
    minWidth: TAB_MIN_WIDTH,
    maxWidth: '214px',
    height: '34px',
    padding: '0 8px 0 9px',
    borderRadius: '10px 10px 0 0',
    border: `1px solid transparent`,
    borderBottomColor: 'transparent',
    backgroundColor: tokens.colorNeutralBackground3,
    color: tokens.colorNeutralForeground2,
    boxShadow: 'none',
    transitionProperty: 'background-color, border-color, box-shadow, color, opacity, transform',
    transitionDuration: tokens.durationFaster,
    transitionTimingFunction: tokens.curveEasyEase,
    cursor: 'pointer',
    userSelect: 'none',
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1,
      color: tokens.colorNeutralForeground1,
    },
    ':focus-visible': {
      outlineStyle: 'solid',
      outlineWidth: '2px',
      outlineColor: tokens.colorStrokeFocus2,
      outlineOffset: '1px',
    },
    '&:hover [data-tab-action="close"]': {
      opacity: 1,
    },
  },
  tabDragging: {
    cursor: 'grabbing',
  },
  activeTab: {
    backgroundColor: tokens.colorNeutralBackground1,
    color: tokens.colorNeutralForeground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    borderBottom: `1px solid ${tokens.colorNeutralBackground1}`,
    boxShadow: `0 1px 0 ${tokens.colorNeutralBackground1}, 0 -1px 0 ${tokens.colorNeutralStroke2}, 1px 0 0 ${tokens.colorNeutralStroke2}, -1px 0 0 ${tokens.colorNeutralStroke2}`,
    zIndex: 1,
  },
  dragSourcePlaceholder: {
    opacity: 0,
    backgroundColor: 'transparent',
    border: '1px solid transparent',
    boxShadow: 'none',
  },
  dragGhost: {
    position: 'fixed',
    display: 'grid',
    gridTemplateColumns: 'auto minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: '6px',
    minWidth: TAB_MIN_WIDTH,
    height: '34px',
    padding: '0 8px 0 9px',
    borderRadius: '10px 10px 0 0',
    backgroundColor: tokens.colorNeutralBackground1,
    color: tokens.colorNeutralForeground1,
    border: `1px solid ${tokens.colorNeutralStroke1}`,
    borderBottomColor: tokens.colorNeutralBackground1,
    boxShadow: 'none',
    pointerEvents: 'none',
    cursor: 'grabbing',
    zIndex: 60,
  },
  ghostLabel: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    fontSize: tokens.fontSizeBase300,
    lineHeight: tokens.lineHeightBase300,
    fontWeight: tokens.fontWeightMedium,
  },
  pinnedMark: {
    color: tokens.colorBrandForeground1,
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  spacer: {
    width: '10px',
  },
  label: {
    minWidth: 0,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
    fontSize: tokens.fontSizeBase300,
    lineHeight: tokens.lineHeightBase300,
    fontWeight: tokens.fontWeightMedium,
  },
  closeButton: {
    flexShrink: 0,
    minWidth: '16px',
    width: '16px',
    height: '16px',
    paddingLeft: '0',
    paddingRight: '0',
    borderRadius: tokens.borderRadiusSmall,
    color: 'inherit',
    opacity: 0,
    transitionProperty: 'opacity',
    transitionDuration: tokens.durationFaster,
    transitionTimingFunction: tokens.curveEasyEase,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Pressed,
      opacity: 1,
    },
    ':focus-visible': {
      opacity: 1,
    },
  },
  tabVisibleClose: {
    '& [data-tab-action="close"]': {
      opacity: 1,
    },
  },
  utility: {
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
    paddingBottom: '4px',
    paddingLeft: '8px',
    '@media (max-width: 720px)': {
      justifyContent: 'flex-end',
      paddingLeft: '0',
      paddingBottom: '2px',
    },
  },
  utilityButton: {
    minWidth: '32px',
    width: '32px',
    height: '32px',
    paddingLeft: '0',
    paddingRight: '0',
    borderRadius: tokens.borderRadiusCircular,
  },
  menu: {
    position: 'fixed',
    minWidth: '176px',
    display: 'grid',
    gap: '2px',
    padding: '4px',
    borderRadius: '10px',
    backgroundColor: tokens.colorNeutralBackground1,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    boxShadow: tokens.shadow8,
    zIndex: 30,
  },
  menuDivider: {
    height: '1px',
    margin: '3px 0',
    backgroundColor: tokens.colorNeutralStroke2,
  },
  menuItem: {
    justifyContent: 'flex-start',
    minWidth: 'auto',
    height: '30px',
    paddingLeft: '10px',
    paddingRight: '10px',
    fontSize: tokens.fontSizeBase200,
    fontWeight: tokens.fontWeightRegular,
    borderRadius: '6px',
    color: tokens.colorNeutralForeground2,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground1Hover,
      color: tokens.colorNeutralForeground1,
    },
  },
})

function getDistance(startX: number, startY: number, currentX: number, currentY: number) {
  const deltaX = currentX - startX
  const deltaY = currentY - startY
  return Math.hypot(deltaX, deltaY)
}

export function OpenTabsBar({
  tabs,
  activeTab,
  activePath,
  onSelect,
  onClose,
  onTogglePinned,
  onCloseOthers,
  onCloseLeft,
  onCloseRight,
  onReorder,
}: {
  tabs: ShellTab[]
  activeTab?: ShellTab
  activePath: string
  onSelect: (path: string) => void
  onClose: (path: string) => void
  onTogglePinned: (path: string) => void
  onCloseOthers: (path: string) => void
  onCloseLeft: (path: string) => void
  onCloseRight: (path: string) => void
  onReorder: (sourcePath: string, targetPath: string) => void
}) {
  const styles = useStyles()
  const scrollerRef = useRef<HTMLDivElement>(null)
  const tabRefs = useRef(new Map<string, HTMLDivElement>())
  const tabsRef = useRef(tabs)
  const previousRectsRef = useRef(new Map<string, DOMRect>())
  const pendingFlipRectsRef = useRef<Map<string, DOMRect> | null>(null)
  const pendingDragRef = useRef<DragCandidate | null>(null)
  const activeDragRef = useRef<ActiveDragState | null>(null)
  const dragReorderCountRef = useRef(0)
  const lastReorderIntentRef = useRef<string | null>(null)
  const suppressClickUntilRef = useRef(0)
  const [dragState, setDragState] = useState<ActiveDragState | null>(null)
  const [trackingPointerId, setTrackingPointerId] = useState<number | null>(null)
  const [canScrollLeft, setCanScrollLeft] = useState(false)
  const [canScrollRight, setCanScrollRight] = useState(false)
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null)

  const closable = tabs.length > 1
  const contextTab = tabs.find((item) => item.path === contextMenu?.path)

  useEffect(() => {
    tabsRef.current = tabs
  }, [tabs])

  function setCurrentDrag(nextState: ActiveDragState | null) {
    activeDragRef.current = nextState
    setDragState(nextState)
  }

  function captureCurrentRects() {
    const rects = new Map<string, DOMRect>()

    tabsRef.current.forEach((tab) => {
      const element = tabRefs.current.get(tab.path)
      if (element) {
        rects.set(tab.path, element.getBoundingClientRect())
      }
    })

    return rects
  }

  useLayoutEffect(() => {
    const previousRects = pendingFlipRectsRef.current ?? previousRectsRef.current
    const tabElements = tabs
      .map((tab) => {
        const element = tabRefs.current.get(tab.path)
        if (!element) {
          return null
        }

        return {
          path: tab.path,
          element,
        }
      })
      .filter((item): item is { path: string; element: HTMLDivElement } => Boolean(item))
    const nextRects = new Map<string, DOMRect>()

    tabElements.forEach(({ element }) => {
      element.getAnimations().forEach((animation) => animation.cancel())
    })

    tabElements.forEach(({ path, element }) => {
      nextRects.set(path, element.getBoundingClientRect())
    })

    const currentDrag = activeDragRef.current
    const shouldAnimateDuringDrag = currentDrag
      ? performance.now() - currentDrag.startedAt >= DRAG_FLIP_WARMUP_MS &&
        dragReorderCountRef.current > 1
      : true
    const animationDuration =
      currentDrag && shouldAnimateDuringDrag ? TAB_FLIP_DRAG_DURATION_MS : TAB_FLIP_DURATION_MS

    tabElements.forEach(({ path, element }) => {
      const previousRect = previousRects.get(path)
      const nextRect = nextRects.get(path)
      if (!element || !previousRect) {
        return
      }
      if (!nextRect) {
        return
      }
      const deltaX = previousRect.left - nextRect.left

      if (Math.abs(deltaX) < 1) {
        return
      }

      if (currentDrag && !shouldAnimateDuringDrag) {
        return
      }

      element.animate(
        [
          { transform: `translateX(${deltaX}px)` },
          { transform: 'translateX(0)' },
        ],
        {
          duration: animationDuration,
          easing: TAB_FLIP_EASING,
        },
      )
    })

    pendingFlipRectsRef.current = null
    previousRectsRef.current = nextRects
  }, [tabs])

  useEffect(() => {
    const element = scrollerRef.current
    if (!element) {
      return
    }

    const currentElement = element

    function syncScrollState() {
      setCanScrollLeft(currentElement.scrollLeft > 8)
      setCanScrollRight(
        currentElement.scrollLeft + currentElement.clientWidth < currentElement.scrollWidth - 8,
      )
    }

    syncScrollState()
    currentElement.addEventListener('scroll', syncScrollState)
    window.addEventListener('resize', syncScrollState)

    return () => {
      currentElement.removeEventListener('scroll', syncScrollState)
      window.removeEventListener('resize', syncScrollState)
    }
  }, [tabs])

  useEffect(() => {
    if (dragState) {
      return
    }

    const activeElement = scrollerRef.current?.querySelector<HTMLElement>(
      `[data-tab-path="${CSS.escape(activePath)}"]`,
    )
    activeElement?.scrollIntoView({ block: 'nearest', inline: 'nearest', behavior: 'smooth' })
  }, [activePath, dragState, tabs])

  useEffect(() => {
    if (!contextMenu) {
      return
    }

    function handleCloseMenu() {
      setContextMenu(null)
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setContextMenu(null)
      }
    }

    window.addEventListener('pointerdown', handleCloseMenu)
    window.addEventListener('resize', handleCloseMenu)
    window.addEventListener('scroll', handleCloseMenu, true)
    window.addEventListener('keydown', handleKeyDown)

    return () => {
      window.removeEventListener('pointerdown', handleCloseMenu)
      window.removeEventListener('resize', handleCloseMenu)
      window.removeEventListener('scroll', handleCloseMenu, true)
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [contextMenu])

  useEffect(() => {
    if (!dragState) {
      return
    }

    const previousUserSelect = document.body.style.userSelect
    const previousCursor = document.body.style.cursor

    document.body.style.userSelect = 'none'
    document.body.style.cursor = 'grabbing'

    return () => {
      document.body.style.userSelect = previousUserSelect
      document.body.style.cursor = previousCursor
    }
  }, [dragState])

  useEffect(() => {
    if (trackingPointerId === null) {
      return
    }

    function cleanupDragState() {
      pendingDragRef.current = null
      setCurrentDrag(null)
      setTrackingPointerId(null)
      lastReorderIntentRef.current = null
      dragReorderCountRef.current = 0
    }

    function syncDragIntent(nextState: ActiveDragState) {
      const currentTabs = tabsRef.current
      const sourceTab = currentTabs.find((item) => item.path === nextState.path)
      if (!sourceTab) {
        lastReorderIntentRef.current = null
        return
      }

      const sourcePartition = currentTabs.filter(
        (item) => Boolean(item.pinned) === Boolean(sourceTab.pinned),
      )
      const sourceIndex = sourcePartition.findIndex((item) => item.path === nextState.path)
      if (sourceIndex < 0) {
        lastReorderIntentRef.current = null
        return
      }

      const candidateRects = sourcePartition
        .filter((item) => item.path !== nextState.path)
        .map((item) => {
          const element = tabRefs.current.get(item.path)
          if (!element) {
            return null
          }

          const rect = element.getBoundingClientRect()
          return {
            path: item.path,
            rect: {
              left: rect.left,
              right: rect.right,
              width: rect.width,
            } satisfies MeasuredRect,
          }
        })
        .filter((item): item is { path: string; rect: MeasuredRect } => Boolean(item))

      if (!candidateRects.length) {
        lastReorderIntentRef.current = null
        return
      }

      const currentLeft = nextState.sourceLeft + (nextState.currentClientX - nextState.startClientX)
      const currentCenter = currentLeft + nextState.width / 2
      const desiredIndex = candidateRects.reduce((count, item) => {
        const midpoint = item.rect.left + item.rect.width / 2
        return midpoint < currentCenter ? count + 1 : count
      }, 0)

      if (desiredIndex === sourceIndex) {
        lastReorderIntentRef.current = null
        return
      }

      const targetPath =
        desiredIndex > sourceIndex
          ? candidateRects[Math.min(desiredIndex - 1, candidateRects.length - 1)]?.path
          : candidateRects[desiredIndex]?.path

      if (!targetPath) {
        lastReorderIntentRef.current = null
        return
      }

      const nextIntent = `${nextState.path}|${targetPath}|${desiredIndex}`
      if (lastReorderIntentRef.current === nextIntent) {
        return
      }

      pendingFlipRectsRef.current = captureCurrentRects()
      lastReorderIntentRef.current = nextIntent
      dragReorderCountRef.current += 1
      onReorder(nextState.path, targetPath)
    }

    function handlePointerMove(event: PointerEvent) {
      if (event.pointerId !== trackingPointerId) {
        return
      }

      const pendingDrag = pendingDragRef.current
      const currentDrag = activeDragRef.current

      if (!currentDrag && pendingDrag) {
        const distance = getDistance(
          pendingDrag.startClientX,
          pendingDrag.startClientY,
          event.clientX,
          event.clientY,
        )

        if (distance < DRAG_THRESHOLD_PX) {
          return
        }

        const startedDrag: ActiveDragState = {
          ...pendingDrag,
          currentClientX: event.clientX,
          currentClientY: event.clientY,
        }

        setContextMenu(null)
        suppressClickUntilRef.current = Number.MAX_SAFE_INTEGER
        setCurrentDrag(startedDrag)
        event.preventDefault()
        syncDragIntent(startedDrag)
        return
      }

      if (!currentDrag) {
        return
      }

      const nextState: ActiveDragState = {
        ...currentDrag,
        currentClientX: event.clientX,
        currentClientY: event.clientY,
      }

      setCurrentDrag(nextState)
      event.preventDefault()
      syncDragIntent(nextState)
    }

    function handlePointerEnd(event: PointerEvent) {
      if (event.pointerId !== trackingPointerId) {
        return
      }

      if (activeDragRef.current) {
        suppressClickUntilRef.current = performance.now() + 320
        event.preventDefault()
      } else {
        suppressClickUntilRef.current = 0
      }

      cleanupDragState()
    }

    window.addEventListener('pointermove', handlePointerMove, { passive: false })
    window.addEventListener('pointerup', handlePointerEnd)
    window.addEventListener('pointercancel', handlePointerEnd)

    return () => {
      window.removeEventListener('pointermove', handlePointerMove)
      window.removeEventListener('pointerup', handlePointerEnd)
      window.removeEventListener('pointercancel', handlePointerEnd)
    }
  }, [onReorder, trackingPointerId])

  if (!tabs.length) {
    return null
  }

  function beginAnimatedMutation(callback: () => void) {
    pendingFlipRectsRef.current = captureCurrentRects()
    callback()
  }

  function handleWheel(event: React.WheelEvent<HTMLDivElement>) {
    const element = scrollerRef.current
    if (!element) {
      return
    }

    if (Math.abs(event.deltaY) > Math.abs(event.deltaX)) {
      element.scrollLeft += event.deltaY
      event.preventDefault()
    }
  }

  function handleTabContextMenu(event: ReactMouseEvent<HTMLDivElement>, path: string) {
    event.preventDefault()
    event.stopPropagation()
    setContextMenu({
      path,
      x: event.clientX,
      y: event.clientY,
    })
  }

  function handleTabPointerDown(event: ReactPointerEvent<HTMLDivElement>, tab: ShellTab) {
    if (event.button !== 0) {
      return
    }

    if ((event.target as HTMLElement).closest('[data-tab-action]')) {
      return
    }

    const rect = event.currentTarget.getBoundingClientRect()

    pendingDragRef.current = {
      pointerId: event.pointerId,
      path: tab.path,
      label: tab.label,
      pinned: Boolean(tab.pinned),
      width: rect.width,
      height: rect.height,
      sourceLeft: rect.left,
      sourceTop: rect.top,
      startClientX: event.clientX,
      startClientY: event.clientY,
      grabOffsetX: event.clientX - rect.left,
      grabOffsetY: event.clientY - rect.top,
      startedAt: performance.now(),
    }

    setTrackingPointerId(event.pointerId)

    if (tab.path !== activePath) {
      onSelect(tab.path)
    }
  }

  function renderTab(tab: ShellTab) {
    const active = tab.path === activePath
    const isDragging = dragState?.path === tab.path

    return (
      <div
        className={styles.tabWrap}
        key={tab.path}
        data-tab-path={tab.path}
        ref={(node) => {
          if (node) {
            tabRefs.current.set(tab.path, node)
          } else {
            tabRefs.current.delete(tab.path)
          }
        }}
      >
        <div
          role="tab"
          aria-selected={active}
          aria-label={tab.label}
          tabIndex={active ? 0 : -1}
          className={mergeClasses(
            styles.tab,
            Boolean(dragState) && styles.tabDragging,
            active && styles.tabVisibleClose,
            active && styles.activeTab,
            isDragging && styles.dragSourcePlaceholder,
          )}
          onClick={() => {
            if (performance.now() < suppressClickUntilRef.current) {
              return
            }

            onSelect(tab.path)
          }}
          onContextMenu={(event) => handleTabContextMenu(event, tab.path)}
          onDoubleClick={() =>
            beginAnimatedMutation(() => {
              onTogglePinned(tab.path)
            })
          }
          onKeyDown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault()
              onSelect(tab.path)
            }
          }}
          onPointerDown={(event) => handleTabPointerDown(event, tab)}
        >
          {tab.pinned ? (
            <span className={styles.pinnedMark}>
              <Pin16Regular />
            </span>
          ) : (
            <span className={styles.spacer} />
          )}
          <Text className={styles.label} title={tab.label}>
            {tab.label}
          </Text>
          {closable && !tab.pinned ? (
            <Button
              className={mergeClasses(styles.closeButton, Boolean(dragState) && styles.tabDragging)}
              appearance="subtle"
              aria-label={`关闭${tab.label}`}
              icon={<Dismiss16Regular />}
              data-tab-action="close"
              onClick={(event) => {
                event.stopPropagation()
                beginAnimatedMutation(() => {
                  onClose(tab.path)
                })
              }}
            />
          ) : (
            <span className={styles.spacer} />
          )}
        </div>
      </div>
    )
  }

  return (
    <div className={mergeClasses(styles.root, Boolean(dragState) && styles.rootDragging)}>
      <div
        className={mergeClasses(
          styles.scrollerWrap,
          canScrollLeft && styles.scrollerFadeLeft,
          canScrollRight && styles.scrollerFadeRight,
        )}
      >
        <div
          ref={scrollerRef}
          className={styles.rail}
          role="tablist"
          aria-label="已打开页面"
          onWheel={handleWheel}
        >
          {tabs.map((tab) => renderTab(tab))}
        </div>
      </div>

      <div className={styles.utility}>
        <Tooltip
          content={activeTab?.pinned ? '取消固定当前标签' : '固定当前标签'}
          relationship="label"
        >
          <Button
            className={styles.utilityButton}
            appearance="subtle"
            aria-label={activeTab?.pinned ? '取消固定当前标签' : '固定当前标签'}
            disabled={!activeTab}
            icon={activeTab?.pinned ? <PinOff16Regular /> : <Pin16Regular />}
            onClick={() =>
              activeTab &&
              beginAnimatedMutation(() => {
                onTogglePinned(activeTab.path)
              })
            }
          />
        </Tooltip>
        <Tooltip content="关闭其他标签" relationship="label">
          <Button
            className={styles.utilityButton}
            appearance="subtle"
            aria-label="关闭其他标签"
            disabled={!activeTab || tabs.length <= 1}
            icon={<DismissSquareMultiple16Regular />}
            onClick={() =>
              activeTab &&
              beginAnimatedMutation(() => {
                onCloseOthers(activeTab.path)
              })
            }
          />
        </Tooltip>
      </div>

      {dragState ? (
        <div
          className={styles.dragGhost}
          style={{
            width: `${dragState.width}px`,
            height: `${dragState.height}px`,
            left: `${dragState.currentClientX - dragState.grabOffsetX}px`,
            top: `${dragState.sourceTop}px`,
          }}
        >
          {dragState.pinned ? (
            <span className={styles.pinnedMark}>
              <Pin16Regular />
            </span>
          ) : (
            <span className={styles.spacer} />
          )}
          <Text className={styles.ghostLabel}>{dragState.label}</Text>
          <span className={styles.spacer} />
        </div>
      ) : null}

      {contextMenu && contextTab ? (
        <div
          className={styles.menu}
          style={{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }}
          onPointerDown={(event) => event.stopPropagation()}
        >
          <Button
            className={styles.menuItem}
            appearance="subtle"
            icon={contextTab.pinned ? <PinOff16Regular /> : <Pin16Regular />}
            onClick={() => {
              beginAnimatedMutation(() => {
                onTogglePinned(contextTab.path)
              })
              setContextMenu(null)
            }}
          >
            {contextTab.pinned ? '取消固定' : '固定标签'}
          </Button>
          <div className={styles.menuDivider} />
          <Button
            className={styles.menuItem}
            appearance="subtle"
            onClick={() => {
              beginAnimatedMutation(() => {
                onCloseOthers(contextTab.path)
              })
              setContextMenu(null)
            }}
          >
            关闭其他标签
          </Button>
          <Button
            className={styles.menuItem}
            appearance="subtle"
            onClick={() => {
              beginAnimatedMutation(() => {
                onCloseLeft(contextTab.path)
              })
              setContextMenu(null)
            }}
          >
            关闭左侧标签
          </Button>
          <Button
            className={styles.menuItem}
            appearance="subtle"
            onClick={() => {
              beginAnimatedMutation(() => {
                onCloseRight(contextTab.path)
              })
              setContextMenu(null)
            }}
          >
            关闭右侧标签
          </Button>
          <Button
            className={styles.menuItem}
            appearance="subtle"
            disabled={contextTab.pinned || tabs.length <= 1}
            icon={<Dismiss16Regular />}
            onClick={() => {
              beginAnimatedMutation(() => {
                onClose(contextTab.path)
              })
              setContextMenu(null)
            }}
          >
            关闭当前标签
          </Button>
        </div>
      ) : null}
    </div>
  )
}
