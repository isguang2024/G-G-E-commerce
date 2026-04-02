import {
  Badge,
  Body1,
  Button,
  Caption1,
  Field,
  Input,
  Switch,
  mergeClasses,
  makeStyles,
  tokens,
} from '@fluentui/react-components'
import { PropertyGrid } from '@/shared/ui/PropertyGrid'
import { EmptyState, ErrorState, LoadingState } from '@/shared/ui/AsyncState'
import { PageStatusBanner } from '@/shared/ui/PageStatusBanner'
import { SectionCard } from '@/shared/ui/SectionCard'
import type { InboxMessageDetail, InboxThread } from '@/shared/types/message-center'

const useStyles = makeStyles({
  list: {
    display: 'grid',
    gap: '10px',
  },
  categoryButton: {
    width: '100%',
    justifyContent: 'space-between',
  },
  messageList: {
    display: 'grid',
    gap: '8px',
  },
  messageCard: {
    display: 'grid',
    gap: '8px',
    width: '100%',
    textAlign: 'left',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
  },
  messageCardActive: {
    backgroundColor: tokens.colorBrandBackground2,
    border: `1px solid ${tokens.colorBrandStroke1}`,
  },
  titleRow: {
    display: 'flex',
    justifyContent: 'space-between',
    gap: '10px',
    alignItems: 'start',
  },
  summary: {
    color: tokens.colorNeutralForeground3,
  },
  metaRow: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  detailBody: {
    display: 'grid',
    gap: '14px',
  },
  content: {
    whiteSpace: 'pre-wrap',
    lineHeight: tokens.lineHeightBase400,
  },
  softText: {
    color: tokens.colorNeutralForeground3,
  },
  toolbar: {
    display: 'grid',
    gap: '12px',
  },
})

export function resolveBoxLabel(value: string) {
  if (value === 'notice') return '通知'
  if (value === 'message') return '消息'
  if (value === 'todo') return '待办'
  return '全部'
}

export function resolveSummaryText(summary: string, fallback: string) {
  return summary.trim() || fallback.trim() || '点击查看详情'
}

export function formatInboxTime(value?: string) {
  if (!value) return '刚刚'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(date)
}

export function InboxCategoryPanel({
  boxOptions,
  boxType,
  unreadOnly,
  onSelectBox,
  onToggleUnreadOnly,
}: {
  boxOptions: Array<{ key: string; label: string; count: number }>
  boxType: string
  unreadOnly: boolean
  onSelectBox: (boxKey: string) => void
  onToggleUnreadOnly: (checked: boolean) => void
}) {
  const styles = useStyles()

  return (
    <SectionCard title="消息分类" description="按收件箱语义切换通知、消息和待办。">
      <div className={styles.list}>
        {boxOptions.map((item) => (
          <Button
            key={item.key || 'all'}
            appearance={boxType === item.key ? 'primary' : 'secondary'}
            className={styles.categoryButton}
            onClick={() => onSelectBox(item.key)}
          >
            {item.label}
            <Badge appearance="tint">{item.count}</Badge>
          </Button>
        ))}

        <Field label="仅看未读">
          <Switch checked={unreadOnly} onChange={(_, data) => onToggleUnreadOnly(Boolean(data.checked))} />
        </Field>
      </div>
    </SectionCard>
  )
}

export function InboxListPanel({
  title,
  keyword,
  records,
  selectedId,
  isLoading,
  errorMessage,
  onKeywordChange,
  onSelectMessage,
}: {
  title: string
  keyword: string
  records: InboxThread[]
  selectedId: string
  isLoading: boolean
  errorMessage?: string
  onKeywordChange: (value: string) => void
  onSelectMessage: (record: InboxThread) => void
}) {
  const styles = useStyles()

  return (
    <SectionCard title={title} description="真实收件链路已经接入，刷新后仍可恢复当前选中消息。">
      <div className={styles.toolbar}>
        <Input placeholder="按标题或摘要搜索" value={keyword} onChange={(_, data) => onKeywordChange(data.value)} />
        {isLoading ? <LoadingState label="正在加载消息列表" /> : null}
        {errorMessage ? <ErrorState description={errorMessage} /> : null}
        {!isLoading && !errorMessage && !records.length ? (
          <EmptyState title="当前筛选下没有消息" description="可以切换分类或关闭未读过滤。" />
        ) : null}
        <div className={styles.messageList}>
          {records.map((item) => (
            <button
              key={item.id}
              className={mergeClasses(styles.messageCard, selectedId === item.id ? styles.messageCardActive : undefined)}
              type="button"
              onClick={() => onSelectMessage(item)}
            >
              <div className={styles.titleRow}>
                <Body1>{item.title || '未命名消息'}</Body1>
                <Caption1>{formatInboxTime(item.sentAt || item.createdAt)}</Caption1>
              </div>
              <Body1 className={styles.summary}>{resolveSummaryText(item.summary, item.messageType)}</Body1>
              <div className={styles.metaRow}>
                <Badge appearance="outline">{resolveBoxLabel(item.boxType)}</Badge>
                {!item.read ? (
                  <Badge color="danger" appearance="filled">
                    未读
                  </Badge>
                ) : null}
                {item.todoStatus ? <Badge appearance="tint">{item.todoStatus}</Badge> : null}
              </div>
            </button>
          ))}
        </div>
      </div>
    </SectionCard>
  )
}

export function InboxDetailPanel({
  detail,
  isLoading,
  errorMessage,
  onTodoDone,
  onTodoIgnore,
  todoPending,
}: {
  detail: InboxMessageDetail | null
  isLoading: boolean
  errorMessage?: string
  onTodoDone: () => void
  onTodoIgnore: () => void
  todoPending: boolean
}) {
  const styles = useStyles()

  return (
    <SectionCard title="消息详情" description="详情区直接承接正文、待办动作和关联 payload。">
      {!detail && !isLoading && !errorMessage ? (
        <EmptyState title="选择一条消息查看详情" description="当前页面会把选中项同步到 URL，刷新后仍可恢复。" />
      ) : null}
      {isLoading ? <LoadingState label="正在加载消息详情" /> : null}
      {errorMessage ? <ErrorState description={errorMessage} /> : null}
      {detail ? (
        <div className={styles.detailBody}>
          <div className={styles.titleRow}>
            <div>
              <Body1>{detail.title}</Body1>
              <Caption1 className={styles.softText}>
                {detail.senderName || '系统发送'} · {formatInboxTime(detail.sentAt || detail.createdAt)}
              </Caption1>
            </div>
            <Badge appearance="tint">{resolveBoxLabel(detail.boxType)}</Badge>
          </div>
          {detail.summary ? <PageStatusBanner intent="info" title="消息摘要" description={detail.summary} /> : null}
          <div className={styles.content}>{detail.content || '当前消息没有正文。'}</div>
          <PropertyGrid
            items={[
              { label: '优先级', value: detail.priority || '-' },
              { label: '租户上下文', value: detail.tenantName || '-' },
              { label: '待办状态', value: detail.todoStatus || '-' },
              { label: '更新时间', value: detail.updatedAt || '-' },
            ]}
          />
          {detail.boxType === 'todo' ? (
            <div className={styles.metaRow}>
              <Button appearance="primary" disabled={todoPending} onClick={onTodoDone}>
                标记完成
              </Button>
              <Button appearance="secondary" disabled={todoPending} onClick={onTodoIgnore}>
                忽略待办
              </Button>
            </div>
          ) : null}
        </div>
      ) : null}
    </SectionCard>
  )
}
