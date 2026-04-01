import { type ReactNode, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Body1Strong,
  Button,
  Caption1,
  Dialog,
  DialogSurface,
  Input,
  makeStyles,
  mergeClasses,
  tokens,
} from '@fluentui/react-components'
import { ArrowEnterLeft20Regular, Search20Regular } from '@fluentui/react-icons'
import {
  buildMenuSearchEntries,
  filterMenuSearchEntries,
  MENU_SEARCH_HISTORY_KEY,
  MENU_SEARCH_HISTORY_LIMIT,
  type MenuSearchEntry,
} from '@/features/navigation/menu-search'
import { AppIcon } from '@/shared/ui/AppIcon'
import type { NavigationItem } from '@/shared/types/navigation'

const useStyles = makeStyles({
  surface: {
    width: 'min(680px, calc(100vw - 32px))',
    maxHeight: 'calc(100vh - 72px)',
    alignSelf: 'start',
    marginTop: '24px',
    padding: '14px 16px 12px',
    borderRadius: tokens.borderRadiusXLarge,
    boxShadow: tokens.shadow64,
  },
  body: {
    display: 'grid',
    gap: '10px',
  },
  searchBar: {
    width: '100%',
  },
  input: {
    width: '100%',
    minHeight: '40px',
    '& input': {
      fontSize: tokens.fontSizeBase300,
    },
  },
  resultsShell: {
    display: 'grid',
    gap: '8px',
  },
  sectionHeader: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: '12px',
    color: tokens.colorNeutralForeground3,
  },
  resultPane: {
    display: 'grid',
    gap: '6px',
    alignContent: 'start',
    minHeight: '84px',
    maxHeight: '248px',
    overflowY: 'auto',
    paddingRight: '2px',
  },
  resultPaneIdle: {
    minHeight: '0',
    maxHeight: '0',
    overflow: 'hidden',
  },
  groupTitle: {
    padding: '4px 6px 0',
    color: tokens.colorNeutralForeground3,
  },
  resultButton: {
    display: 'grid',
    gridTemplateColumns: '20px minmax(0, 1fr) auto',
    alignItems: 'center',
    gap: '12px',
    width: '100%',
    minHeight: '46px',
    padding: '10px 14px',
    borderRadius: tokens.borderRadiusLarge,
    border: `1px solid ${tokens.colorNeutralStroke1}`,
    backgroundColor: tokens.colorNeutralBackground2,
    textAlign: 'left',
    color: tokens.colorNeutralForeground2,
    transitionProperty: 'background-color, border-color',
    transitionDuration: tokens.durationFast,
    transitionTimingFunction: tokens.curveEasyEase,
    ':hover': {
      backgroundColor: tokens.colorNeutralBackground2Hover,
    },
  },
  resultButtonActive: {
    border: `1px solid ${tokens.colorBrandStroke1}`,
    backgroundColor: tokens.colorBrandBackground2,
    color: tokens.colorNeutralForeground1,
  },
  resultMain: {
    minWidth: 0,
    display: 'grid',
    gap: '2px',
  },
  resultTitle: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  trail: {
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  activeMeta: {
    color: tokens.colorNeutralForeground2,
  },
  highlight: {
    padding: '0 2px',
    borderRadius: tokens.borderRadiusSmall,
    backgroundColor: tokens.colorBrandBackground2,
    color: tokens.colorBrandForeground1,
  },
  highlightOnActive: {
    padding: '0 2px',
    borderRadius: tokens.borderRadiusSmall,
    backgroundColor: tokens.colorNeutralBackgroundInverted,
    color: tokens.colorNeutralForegroundInverted,
  },
  resultAction: {
    color: 'inherit',
    opacity: 0.8,
  },
  empty: {
    display: 'grid',
    gap: '4px',
    padding: '10px 4px 6px',
    color: tokens.colorNeutralForeground3,
  },
  footer: {
    display: 'flex',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: '16px',
    flexWrap: 'wrap',
    paddingTop: '12px',
    borderTop: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  footerCopy: {
    display: 'flex',
    alignItems: 'center',
    color: tokens.colorNeutralForeground3,
  },
  hintRow: {
    display: 'flex',
    alignItems: 'center',
    gap: '14px',
    flexWrap: 'wrap',
    color: tokens.colorNeutralForeground3,
  },
  hintItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
  },
  keycap: {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    minWidth: '22px',
    height: '22px',
    padding: '0 6px',
    borderRadius: tokens.borderRadiusMedium,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground2,
  },
})

function renderHighlightedText(text: string, query: string, highlightClassName: string): ReactNode {
  const normalizedQuery = query.trim()
  if (!normalizedQuery) {
    return text
  }

  const lowerText = text.toLowerCase()
  const lowerQuery = normalizedQuery.toLowerCase()
  const parts: ReactNode[] = []
  let cursor = 0

  while (cursor < text.length) {
    const matchIndex = lowerText.indexOf(lowerQuery, cursor)
    if (matchIndex === -1) {
      parts.push(text.slice(cursor))
      break
    }

    if (matchIndex > cursor) {
      parts.push(text.slice(cursor, matchIndex))
    }

    const endIndex = matchIndex + normalizedQuery.length
    parts.push(
      <mark key={`${matchIndex}-${endIndex}`} className={highlightClassName}>
        {text.slice(matchIndex, endIndex)}
      </mark>,
    )
    cursor = endIndex
  }

  return parts
}

function buildDisplayTrail(entry: MenuSearchEntry) {
  const segments = [entry.groupLabel, ...entry.trail].filter(Boolean)
  return segments.filter((segment, index) => {
    if (index === 0) {
      return true
    }

    return segment.trim() !== segments[index - 1]?.trim()
  })
}

function readHistoryPaths() {
  if (typeof window === 'undefined') {
    return []
  }

  try {
    const raw = window.localStorage.getItem(MENU_SEARCH_HISTORY_KEY)
    if (!raw) return []

    const parsed = JSON.parse(raw)
    return Array.isArray(parsed) ? parsed.filter((item): item is string => typeof item === 'string') : []
  } catch {
    return []
  }
}

function writeHistoryPaths(paths: string[]) {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(MENU_SEARCH_HISTORY_KEY, JSON.stringify(paths))
}

export function MenuSearchDialog({
  items,
  open,
  onOpenChange,
}: {
  items: NavigationItem[]
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const styles = useStyles()
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement | null>(null)
  const [inputValue, setInputValue] = useState('')
  const [query, setQuery] = useState('')
  const [composing, setComposing] = useState(false)
  const [activeIndex, setActiveIndex] = useState(0)
  const [historyPaths, setHistoryPaths] = useState<string[]>(() => readHistoryPaths())

  const entries = useMemo(() => buildMenuSearchEntries(items), [items])
  const results = useMemo(() => filterMenuSearchEntries(entries, query), [entries, query])
  const historyEntries = useMemo(
    () =>
      historyPaths
        .map((path) => entries.find((entry) => entry.path === path))
        .filter((entry): entry is MenuSearchEntry => Boolean(entry)),
    [entries, historyPaths],
  )
  const visibleEntries = query.trim() ? results : historyEntries

  useEffect(() => {
    writeHistoryPaths(historyPaths)
  }, [historyPaths])

  useEffect(() => {
    setActiveIndex(0)
  }, [query, open])

  useEffect(() => {
    if (!open) {
      setInputValue('')
      setQuery('')
      setComposing(false)
      return
    }

    const timeoutId = window.setTimeout(() => {
      inputRef.current?.focus()
    }, 40)

    return () => window.clearTimeout(timeoutId)
  }, [open])

  useEffect(() => {
    if (activeIndex > 0 && activeIndex >= visibleEntries.length) {
      setActiveIndex(Math.max(visibleEntries.length - 1, 0))
    }
  }, [activeIndex, visibleEntries.length])

  function closeDialog() {
    onOpenChange(false)
  }

  function updateHistory(entry: MenuSearchEntry) {
    setHistoryPaths((current) =>
      [entry.path, ...current.filter((path) => path !== entry.path)].slice(0, MENU_SEARCH_HISTORY_LIMIT),
    )
  }

  function handleSelect(entry: MenuSearchEntry) {
    updateHistory(entry)
    closeDialog()
    navigate(entry.path)
  }

  function handleKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    if (!open) return
    if (composing || event.nativeEvent.isComposing) return

    if (event.key === 'ArrowDown' && visibleEntries.length > 0) {
      event.preventDefault()
      setActiveIndex((current) => (current + 1) % visibleEntries.length)
    }

    if (event.key === 'ArrowUp' && visibleEntries.length > 0) {
      event.preventDefault()
      setActiveIndex((current) => (current - 1 + visibleEntries.length) % visibleEntries.length)
    }

    if (event.key === 'Enter' && visibleEntries[activeIndex]) {
      event.preventDefault()
      handleSelect(visibleEntries[activeIndex])
    }
  }

  return (
    <Dialog open={open} onOpenChange={(_, data) => onOpenChange(data.open)}>
      <DialogSurface className={styles.surface} onKeyDown={handleKeyDown}>
        <div className={styles.body}>
          <div className={styles.searchBar}>
            <Input
              className={styles.input}
              contentBefore={<Search20Regular />}
              contentAfter={<ArrowEnterLeft20Regular />}
              placeholder="搜索页面"
              ref={inputRef}
              value={inputValue}
              onChange={(event, data) => {
                const nativeEvent = event.nativeEvent as InputEvent & { isComposing?: boolean }
                setInputValue(data.value)

                if (composing || nativeEvent.isComposing) {
                  return
                }

                setQuery(data.value)
              }}
              onCompositionStart={() => setComposing(true)}
              onCompositionEnd={(event) => {
                const confirmedValue = event.currentTarget.value
                setComposing(false)
                setInputValue(confirmedValue)
                setQuery(confirmedValue)
              }}
            />
          </div>

          <div className={styles.resultsShell}>
            {query.trim() || historyEntries.length > 0 ? (
              <div className={styles.sectionHeader}>
                <Caption1>{query.trim() ? `匹配结果 ${results.length}` : '最近访问'}</Caption1>
                {!query.trim() && historyEntries.length > 0 ? (
                  <Button appearance="subtle" size="small" onClick={() => setHistoryPaths([])}>
                    清空
                  </Button>
                ) : null}
              </div>
            ) : null}

            <div className={mergeClasses(styles.resultPane, !query.trim() && historyEntries.length === 0 && styles.resultPaneIdle)}>
              {visibleEntries.length > 0 ? (
                visibleEntries.map((entry, index) => {
                  const previousEntry = visibleEntries[index - 1]
                  const showGroupTitle = !previousEntry || previousEntry.group !== entry.group
                  const active = index === activeIndex

                  return (
                    <div key={entry.path}>
                      {showGroupTitle ? <Caption1 className={styles.groupTitle}>{entry.groupLabel}</Caption1> : null}
                      <button
                        className={mergeClasses(styles.resultButton, active && styles.resultButtonActive)}
                        type="button"
                        onClick={() => handleSelect(entry)}
                        onMouseEnter={() => setActiveIndex(index)}
                      >
                        <AppIcon icon={entry.icon} />
                        <div className={styles.resultMain}>
                          <Body1Strong className={styles.resultTitle}>
                            {renderHighlightedText(
                              entry.label,
                              query,
                              active ? styles.highlightOnActive : styles.highlight,
                            )}
                          </Body1Strong>
                          <Caption1 className={mergeClasses(styles.trail, active && styles.activeMeta)}>
                            {renderHighlightedText(
                              buildDisplayTrail(entry).join(' / '),
                              query,
                              active ? styles.highlightOnActive : styles.highlight,
                            )}
                          </Caption1>
                        </div>
                        <span className={styles.resultAction}>
                          <ArrowEnterLeft20Regular />
                        </span>
                      </button>
                    </div>
                  )
                })
              ) : query.trim() ? (
                <div className={styles.empty}>
                  <Body1Strong>没有匹配的菜单项</Body1Strong>
                  <Caption1>试试搜索系统管理、消息模板或用户管理。</Caption1>
                </div>
              ) : null}
            </div>
          </div>

          <div className={styles.footer}>
            <div className={styles.footerCopy}>
              <Caption1>支持 Ctrl + K 打开，方向键切换，回车跳转。</Caption1>
            </div>
            <div className={styles.hintRow}>
              <div className={styles.hintItem}>
                <span className={styles.keycap}>↵</span>
                <Caption1>选择</Caption1>
              </div>
              <div className={styles.hintItem}>
                <span className={styles.keycap}>↑</span>
                <span className={styles.keycap}>↓</span>
                <Caption1>切换</Caption1>
              </div>
              <div className={styles.hintItem}>
                <span className={styles.keycap}>Esc</span>
                <Caption1>关闭</Caption1>
              </div>
            </div>
          </div>
        </div>
      </DialogSurface>
    </Dialog>
  )
}
