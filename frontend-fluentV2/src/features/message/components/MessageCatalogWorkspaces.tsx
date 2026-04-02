import { useEffect, useState } from 'react'
import { Badge, Body1, Button, Checkbox, Field, Input, Spinner, Textarea, makeStyles } from '@fluentui/react-components'
import { Add20Regular, Save20Regular } from '@fluentui/react-icons'
import { useSearchParams } from 'react-router-dom'
import { PageContainer } from '@/features/shell/components/PageContainer'
import {
  useCreateMessageRecipientGroupMutation,
  useCreateMessageSenderMutation,
  useCreateMessageTemplateMutation,
  useMessageRecordDetailQuery,
  useMessageRecipientGroupsQuery,
  useMessageRecordsQuery,
  useMessageSendersQuery,
  useMessageTemplatesQuery,
  useUpdateMessageRecipientGroupMutation,
  useUpdateMessageSenderMutation,
  useUpdateMessageTemplateMutation,
} from '@/features/message/message.service'
import { EmptyState, ErrorState, LoadingState } from '@/shared/ui/AsyncState'
import { DetailTimeline } from '@/shared/ui/DetailTimeline'
import { LinkCardGrid } from '@/shared/ui/LinkCardGrid'
import { MetricGrid } from '@/shared/ui/MetricGrid'
import { PageStatusBanner } from '@/shared/ui/PageStatusBanner'
import { PropertyGrid } from '@/shared/ui/PropertyGrid'
import { SectionCard } from '@/shared/ui/SectionCard'
import { SimpleTable } from '@/shared/ui/SimpleTable'
import { TwoPaneWorkbench, WorkbenchStack } from '@/shared/ui/WorkbenchLayouts'
import type {
  MessageRecipientGroupSavePayload,
  MessageSenderSavePayload,
  MessageTemplateSavePayload,
} from '@/shared/types/message-center'

const useStyles = makeStyles({
  form: {
    display: 'grid',
    gap: '12px',
  },
  formGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  checklist: {
    display: 'grid',
    gap: '8px',
  },
  content: {
    whiteSpace: 'pre-wrap',
  },
})

function updateSearchParams(
  searchParams: URLSearchParams,
  setSearchParams: ReturnType<typeof useSearchParams>[1],
  patch: Record<string, string | null | undefined>,
) {
  const next = new URLSearchParams(searchParams)
  Object.entries(patch).forEach(([key, value]) => {
    if (!value) next.delete(key)
    else next.set(key, value)
  })
  setSearchParams(next, { replace: true })
}

function createTemplateDraft(): MessageTemplateSavePayload {
  return {
    templateKey: '',
    name: '',
    description: '',
    messageType: 'notice',
    audienceType: 'all_users',
    titleTemplate: '',
    summaryTemplate: '',
    contentTemplate: '',
    status: 'normal',
  }
}

function createSenderDraft(): MessageSenderSavePayload {
  return {
    name: '',
    description: '',
    avatarUrl: '',
    isDefault: false,
    status: 'normal',
  }
}

function createGroupDraft(): MessageRecipientGroupSavePayload {
  return {
    groupKey: '',
    name: '',
    description: '',
    status: 'normal',
  }
}

function stringifyValue(value: unknown) {
  if (value === null || value === undefined || value === '') return '-'
  if (Array.isArray(value)) return value.length ? value.map((item) => `${item}`).join('、') : '-'
  if (typeof value === 'object') return JSON.stringify(value)
  return `${value}`
}

function buildPayloadItems(input: Record<string, unknown>) {
  return Object.entries(input || {}).map(([key, value]) => ({
    label: key,
    value: stringifyValue(value),
  }))
}

interface PageFeedbackState {
  intent: 'success' | 'error'
  title: string
  description: string
}

function MessageTemplatePage({
  routeId,
  scope,
  title,
}: {
  routeId: string
  scope: 'platform' | 'team'
  title: string
}) {
  const styles = useStyles()
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedId = searchParams.get('selectedTemplateId') || ''
  const isCreating = searchParams.get('mode') === 'new'
  const listQuery = useMessageTemplatesQuery(scope)
  const createMutation = useCreateMessageTemplateMutation(scope)
  const updateMutation = useUpdateMessageTemplateMutation(scope, selectedId || '')
  const [draft, setDraft] = useState<MessageTemplateSavePayload>(createTemplateDraft())
  const [feedback, setFeedback] = useState<PageFeedbackState | null>(null)
  const selected = listQuery.data?.records.find((item) => item.id === selectedId) || null

  useEffect(() => {
    if (isCreating) {
      setDraft(createTemplateDraft())
      return
    }
    if (selected) {
      setDraft({
        templateKey: selected.templateKey,
        name: selected.name,
        description: selected.description,
        messageType: selected.messageType,
        audienceType: selected.audienceType,
        titleTemplate: selected.titleTemplate,
        summaryTemplate: selected.summaryTemplate,
        contentTemplate: selected.contentTemplate,
        status: selected.status,
      })
    }
  }, [isCreating, selected])

  useEffect(() => {
    if (!selectedId || isCreating || !listQuery.data) return
    if (!selected) {
      updateSearchParams(searchParams, setSearchParams, { selectedTemplateId: '' })
    }
  }, [isCreating, listQuery.data, searchParams, selected, selectedId, setSearchParams])

  async function handleSave() {
    try {
      if (isCreating) {
        const created = await createMutation.mutateAsync(draft)
        updateSearchParams(searchParams, setSearchParams, { selectedTemplateId: created.id, mode: '' })
        setFeedback({
          intent: 'success',
          title: '消息模板已创建',
          description: `模板「${created.name || draft.name || '未命名模板'}」已创建并切换为当前上下文。`,
        })
        return
      }
      await updateMutation.mutateAsync(draft)
      setFeedback({
        intent: 'success',
        title: '消息模板已保存',
        description: `模板「${draft.name || selected?.name || '未命名模板'}」已更新。`,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '保存消息模板失败'
      setFeedback({
        intent: 'error',
        title: '保存消息模板失败',
        description: message,
      })
    }
  }

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <Button appearance="secondary" icon={<Add20Regular />} onClick={() => updateSearchParams(searchParams, setSearchParams, { mode: 'new', selectedTemplateId: '' })}>
            新建模板
          </Button>
          {selectedId && !isCreating ? <Button appearance="secondary" disabled>删除能力待后端开放</Button> : null}
          <Button appearance="primary" icon={<Save20Regular />} disabled={createMutation.isPending || updateMutation.isPending} onClick={() => void handleSave()}>
            {createMutation.isPending || updateMutation.isPending ? <Spinner size="tiny" /> : '保存模板'}
          </Button>
        </>
      }
    >
      {feedback ? <PageStatusBanner intent={feedback.intent} title={feedback.title} description={feedback.description} /> : null}
      <TwoPaneWorkbench
        primary={
          <SectionCard title={title} description="模板治理继续采用列表 + 详情编辑模式，并保留 URL 驱动的当前选中项。">
            {listQuery.isLoading ? <LoadingState label="正在加载模板" /> : null}
            {listQuery.isError ? <ErrorState description={listQuery.error?.message} /> : null}
            {listQuery.data ? (
              <SimpleTable
                columns={[
                  { key: 'name', header: '模板', render: (item) => `${item.name} (${item.templateKey})` },
                  { key: 'type', header: '类型', render: (item) => item.messageType },
                  { key: 'audience', header: '受众', render: (item) => item.audienceType },
                ]}
                items={listQuery.data.records}
                onRowClick={(item) => updateSearchParams(searchParams, setSearchParams, { selectedTemplateId: item.id, mode: '' })}
                rowKey={(item) => item.id}
                selectedRowKey={isCreating ? '' : selectedId}
              />
            ) : null}
          </SectionCard>
        }
        secondary={
          <WorkbenchStack>
            <SectionCard title={isCreating ? '新建模板' : selected ? `编辑模板 · ${selected.name}` : '模板详情'} description="模板元数据和正文模板都保留在同一上下文中编辑。">
              {!isCreating && !selected ? <EmptyState title="选择左侧模板或新建一条记录" /> : null}
              {isCreating || selected ? (
                <div className={styles.form}>
                  {selected ? (
                    <PropertyGrid
                      items={[
                        { label: '归属范围', value: selected.ownerScope || '-' },
                        { label: '状态', value: selected.status || '-' },
                        { label: '更新时间', value: selected.updatedAt || '-' },
                        { label: '受众类型', value: selected.audienceType || '-' },
                      ]}
                    />
                  ) : null}
                  <div className={styles.formGrid}>
                    <Field label="模板键">
                      <Input value={draft.templateKey} onChange={(_, data) => setDraft((prev) => ({ ...prev, templateKey: data.value }))} />
                    </Field>
                    <Field label="名称">
                      <Input value={draft.name} onChange={(_, data) => setDraft((prev) => ({ ...prev, name: data.value }))} />
                    </Field>
                    <Field label="消息类型">
                      <Input value={draft.messageType} onChange={(_, data) => setDraft((prev) => ({ ...prev, messageType: data.value }))} />
                    </Field>
                    <Field label="受众类型">
                      <Input value={draft.audienceType} onChange={(_, data) => setDraft((prev) => ({ ...prev, audienceType: data.value }))} />
                    </Field>
                  </div>
                  <Field label="说明">
                    <Textarea value={draft.description} onChange={(_, data) => setDraft((prev) => ({ ...prev, description: data.value }))} />
                  </Field>
                  <Field label="标题模板">
                    <Textarea value={draft.titleTemplate} onChange={(_, data) => setDraft((prev) => ({ ...prev, titleTemplate: data.value }))} />
                  </Field>
                  <Field label="摘要模板">
                    <Textarea value={draft.summaryTemplate} onChange={(_, data) => setDraft((prev) => ({ ...prev, summaryTemplate: data.value }))} />
                  </Field>
                  <Field label="正文模板">
                    <Textarea resize="vertical" value={draft.contentTemplate} onChange={(_, data) => setDraft((prev) => ({ ...prev, contentTemplate: data.value }))} />
                  </Field>
                </div>
              ) : null}
            </SectionCard>
          </WorkbenchStack>
        }
      />
    </PageContainer>
  )
}

function MessageSenderPage({ routeId, scope, title }: { routeId: string; scope: 'platform' | 'team'; title: string }) {
  const styles = useStyles()
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedId = searchParams.get('selectedSenderId') || ''
  const isCreating = searchParams.get('mode') === 'new'
  const listQuery = useMessageSendersQuery(scope)
  const createMutation = useCreateMessageSenderMutation(scope)
  const updateMutation = useUpdateMessageSenderMutation(scope, selectedId || '')
  const [draft, setDraft] = useState<MessageSenderSavePayload>(createSenderDraft())
  const [feedback, setFeedback] = useState<PageFeedbackState | null>(null)
  const selected = listQuery.data?.find((item) => item.id === selectedId) || null

  useEffect(() => {
    if (isCreating) {
      setDraft(createSenderDraft())
      return
    }
    if (selected) {
      setDraft({
        name: selected.name,
        description: selected.description,
        avatarUrl: selected.avatarUrl,
        isDefault: selected.isDefault,
        status: selected.status,
      })
    }
  }, [isCreating, selected])

  useEffect(() => {
    if (!selectedId || isCreating || !listQuery.data) return
    if (!selected) {
      updateSearchParams(searchParams, setSearchParams, { selectedSenderId: '' })
    }
  }, [isCreating, listQuery.data, searchParams, selected, selectedId, setSearchParams])

  async function handleSave() {
    try {
      if (isCreating) {
        const created = await createMutation.mutateAsync(draft)
        updateSearchParams(searchParams, setSearchParams, { selectedSenderId: created.id, mode: '' })
        setFeedback({
          intent: 'success',
          title: '消息发送人已创建',
          description: `发送人「${created.name || draft.name || '未命名发送人'}」已创建。`,
        })
        return
      }
      await updateMutation.mutateAsync(draft)
      setFeedback({
        intent: 'success',
        title: '消息发送人已保存',
        description: `发送人「${draft.name || selected?.name || '未命名发送人'}」已更新。`,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '保存发送人失败'
      setFeedback({
        intent: 'error',
        title: '保存发送人失败',
        description: message,
      })
    }
  }

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <Button appearance="secondary" icon={<Add20Regular />} onClick={() => updateSearchParams(searchParams, setSearchParams, { mode: 'new', selectedSenderId: '' })}>
            新建发送人
          </Button>
          {selectedId && !isCreating ? <Button appearance="secondary" disabled>删除能力待后端开放</Button> : null}
          <Button appearance="primary" icon={<Save20Regular />} disabled={createMutation.isPending || updateMutation.isPending} onClick={() => void handleSave()}>
            {createMutation.isPending || updateMutation.isPending ? <Spinner size="tiny" /> : '保存发送人'}
          </Button>
        </>
      }
    >
      {feedback ? <PageStatusBanner intent={feedback.intent} title={feedback.title} description={feedback.description} /> : null}
      <TwoPaneWorkbench
        primary={
          <SectionCard title={title} description="发送人治理保留列表 + 右侧详情编辑，便于系统域和团队域复用。">
            {listQuery.isLoading ? <LoadingState label="正在加载发送人" /> : null}
            {listQuery.isError ? <ErrorState description={listQuery.error?.message} /> : null}
            {listQuery.data ? (
              <SimpleTable
                columns={[
                  { key: 'name', header: '发送人', render: (item) => item.name },
                  { key: 'scope', header: '归属', render: (item) => item.scopeType },
                  { key: 'default', header: '默认', render: (item) => (item.isDefault ? '是' : '否') },
                ]}
                items={listQuery.data}
                onRowClick={(item) => updateSearchParams(searchParams, setSearchParams, { selectedSenderId: item.id, mode: '' })}
                rowKey={(item) => item.id}
                selectedRowKey={isCreating ? '' : selectedId}
              />
            ) : null}
          </SectionCard>
        }
        secondary={
          <SectionCard title={isCreating ? '新建发送人' : selected ? `编辑发送人 · ${selected.name}` : '发送人详情'} description="发送人身份、说明和默认状态都会在右侧详情区同步更新。">
            {!isCreating && !selected ? <EmptyState title="选择左侧发送人或新建一条记录" /> : null}
            {isCreating || selected ? (
              <div className={styles.form}>
                {selected ? (
                  <PropertyGrid
                    items={[
                      { label: '归属范围', value: selected.scopeType || '-' },
                      { label: '默认发送人', value: selected.isDefault ? '是' : '否' },
                      { label: '状态', value: selected.status || '-' },
                      { label: '更新时间', value: selected.updatedAt || '-' },
                    ]}
                  />
                ) : null}
                <Field label="名称">
                  <Input value={draft.name} onChange={(_, data) => setDraft((prev) => ({ ...prev, name: data.value }))} />
                </Field>
                <Field label="头像地址">
                  <Input value={draft.avatarUrl || ''} onChange={(_, data) => setDraft((prev) => ({ ...prev, avatarUrl: data.value }))} />
                </Field>
                <Field label="说明">
                  <Textarea value={draft.description} onChange={(_, data) => setDraft((prev) => ({ ...prev, description: data.value }))} />
                </Field>
                <Checkbox checked={draft.isDefault} label="设为默认发送人" onChange={(_, data) => setDraft((prev) => ({ ...prev, isDefault: Boolean(data.checked) }))} />
              </div>
            ) : null}
          </SectionCard>
        }
      />
    </PageContainer>
  )
}

function MessageRecipientGroupPage({ routeId, scope, title }: { routeId: string; scope: 'platform' | 'team'; title: string }) {
  const styles = useStyles()
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedId = searchParams.get('selectedGroupId') || ''
  const isCreating = searchParams.get('mode') === 'new'
  const listQuery = useMessageRecipientGroupsQuery(scope)
  const createMutation = useCreateMessageRecipientGroupMutation(scope)
  const updateMutation = useUpdateMessageRecipientGroupMutation(scope, selectedId || '')
  const [draft, setDraft] = useState<MessageRecipientGroupSavePayload>(createGroupDraft())
  const [feedback, setFeedback] = useState<PageFeedbackState | null>(null)
  const selected = listQuery.data?.find((item) => item.id === selectedId) || null

  useEffect(() => {
    if (isCreating) {
      setDraft(createGroupDraft())
      return
    }
    if (selected) {
      setDraft({
        groupKey: selected.groupKey,
        name: selected.name,
        description: selected.description,
        status: selected.status,
      })
    }
  }, [isCreating, selected])

  useEffect(() => {
    if (!selectedId || isCreating || !listQuery.data) return
    if (!selected) {
      updateSearchParams(searchParams, setSearchParams, { selectedGroupId: '' })
    }
  }, [isCreating, listQuery.data, searchParams, selected, selectedId, setSearchParams])

  async function handleSave() {
    try {
      if (isCreating) {
        const created = await createMutation.mutateAsync(draft)
        updateSearchParams(searchParams, setSearchParams, { selectedGroupId: created.id, mode: '' })
        setFeedback({
          intent: 'success',
          title: '收件组已创建',
          description: `收件组「${created.name || draft.name || '未命名分组'}」已创建。`,
        })
        return
      }
      await updateMutation.mutateAsync(draft)
      setFeedback({
        intent: 'success',
        title: '收件组已保存',
        description: `收件组「${draft.name || selected?.name || '未命名分组'}」已更新。`,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '保存收件组失败'
      setFeedback({
        intent: 'error',
        title: '保存收件组失败',
        description: message,
      })
    }
  }

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <Button appearance="secondary" icon={<Add20Regular />} onClick={() => updateSearchParams(searchParams, setSearchParams, { mode: 'new', selectedGroupId: '' })}>
            新建分组
          </Button>
          {selectedId && !isCreating ? <Button appearance="secondary" disabled>删除能力待后端开放</Button> : null}
          <Button appearance="primary" icon={<Save20Regular />} disabled={createMutation.isPending || updateMutation.isPending} onClick={() => void handleSave()}>
            {createMutation.isPending || updateMutation.isPending ? <Spinner size="tiny" /> : '保存分组'}
          </Button>
        </>
      }
    >
      {feedback ? <PageStatusBanner intent={feedback.intent} title={feedback.title} description={feedback.description} /> : null}
      <TwoPaneWorkbench
        primary={
          <SectionCard title={title} description="收件组页承接固定受众集合、匹配模式和目标摘要。">
            {listQuery.isLoading ? <LoadingState label="正在加载收件组" /> : null}
            {listQuery.isError ? <ErrorState description={listQuery.error?.message} /> : null}
            {listQuery.data ? (
              <SimpleTable
                columns={[
                  { key: 'name', header: '分组', render: (item) => item.name },
                  { key: 'matchMode', header: '匹配模式', render: (item) => item.matchMode || '-' },
                  { key: 'estimated', header: '预计覆盖', render: (item) => `${item.estimatedCount || item.memberCount}` },
                ]}
                items={listQuery.data}
                onRowClick={(item) => updateSearchParams(searchParams, setSearchParams, { selectedGroupId: item.id, mode: '' })}
                rowKey={(item) => item.id}
                selectedRowKey={isCreating ? '' : selectedId}
              />
            ) : null}
          </SectionCard>
        }
        secondary={
          <WorkbenchStack>
            <SectionCard title={isCreating ? '新建收件组' : selected ? `编辑收件组 · ${selected.name}` : '收件组详情'} description="右侧详情区会同时展示分组元信息和当前目标集合摘要。">
              {!isCreating && !selected ? <EmptyState title="选择左侧收件组或新建一条记录" /> : null}
              {isCreating || selected ? (
                <div className={styles.form}>
                  {selected ? (
                    <PropertyGrid
                      items={[
                        { label: '匹配模式', value: selected.matchMode || '-' },
                        { label: '预计覆盖', value: `${selected.estimatedCount || selected.memberCount}` },
                        { label: '状态', value: selected.status || '-' },
                        { label: '更新时间', value: selected.updatedAt || '-' },
                      ]}
                    />
                  ) : null}
                  <Field label="分组键">
                    <Input value={draft.groupKey} onChange={(_, data) => setDraft((prev) => ({ ...prev, groupKey: data.value }))} />
                  </Field>
                  <Field label="名称">
                    <Input value={draft.name} onChange={(_, data) => setDraft((prev) => ({ ...prev, name: data.value }))} />
                  </Field>
                  <Field label="说明">
                    <Textarea value={draft.description} onChange={(_, data) => setDraft((prev) => ({ ...prev, description: data.value }))} />
                  </Field>
                </div>
              ) : null}
            </SectionCard>

            {selected?.targets.length ? (
              <SectionCard title="目标摘要" description="当前收件组包含的目标实体及排序摘要。">
                <SimpleTable
                  columns={[
                    { key: 'label', header: '目标', render: (item) => item.targetLabel },
                    { key: 'type', header: '类型', render: (item) => item.targetType || '-' },
                    { key: 'value', header: '值', render: (item) => item.targetValue || '-' },
                    { key: 'sort', header: '排序', render: (item) => `${item.sortOrder}` },
                  ]}
                  items={selected.targets}
                  rowKey={(item) => item.id}
                />
              </SectionCard>
            ) : null}
          </WorkbenchStack>
        }
      />
    </PageContainer>
  )
}

function MessageRecordPage({ routeId, scope, title }: { routeId: string; scope: 'platform' | 'team'; title: string }) {
  const styles = useStyles()
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedRecordId = searchParams.get('selectedRecordId') || ''
  const listQuery = useMessageRecordsQuery(scope, { current: 1, size: 40 })
  const detailQuery = useMessageRecordDetailQuery(scope, selectedRecordId || undefined)

  useEffect(() => {
    if (!selectedRecordId || !listQuery.data) return
    const stillExists = listQuery.data.records.some((item) => item.id === selectedRecordId)
    if (!stillExists) {
      updateSearchParams(searchParams, setSearchParams, { selectedRecordId: '' })
    }
  }, [listQuery.data, searchParams, selectedRecordId, setSearchParams])

  const recordMetrics = detailQuery.data
    ? [
        { id: 'deliveries', label: '投递总数', value: `${detailQuery.data.deliveryCount}`, tone: 'brand' as const },
        { id: 'read', label: '已读', value: `${detailQuery.data.readCount}`, tone: 'success' as const },
        { id: 'unread', label: '未读', value: `${detailQuery.data.unreadCount}`, tone: 'warning' as const },
        { id: 'todo', label: '待办处理中', value: `${detailQuery.data.pendingTodoCount}`, tone: 'neutral' as const },
      ]
    : []

  return (
    <PageContainer routeId={routeId}>
      <TwoPaneWorkbench
        primary={
          <SectionCard title={title} description="消息记录页聚焦投递历史、状态和 payload 摘要。">
            {listQuery.isLoading ? <LoadingState label="正在加载消息记录" /> : null}
            {listQuery.isError ? <ErrorState description={listQuery.error?.message} /> : null}
            {listQuery.data ? (
              <SimpleTable
                columns={[
                  { key: 'title', header: '标题', render: (item) => item.title || '-' },
                  { key: 'type', header: '类型', render: (item) => item.messageType },
                  { key: 'status', header: '状态', render: (item) => item.status },
                  { key: 'deliveries', header: '投递', render: (item) => `${item.deliveryCount}` },
                ]}
                items={listQuery.data.records}
                onRowClick={(item) => updateSearchParams(searchParams, setSearchParams, { selectedRecordId: item.id })}
                rowKey={(item) => item.id}
                selectedRowKey={selectedRecordId}
              />
            ) : null}
          </SectionCard>
        }
        secondary={
          <WorkbenchStack>
            <SectionCard title="记录详情" description="消息内容、投递摘要、时间线和 payload 摘要在这里统一查看。">
              {!selectedRecordId ? <EmptyState title="选择左侧记录查看详情" /> : null}
              {selectedRecordId && detailQuery.isLoading ? <LoadingState label="正在加载记录详情" /> : null}
              {selectedRecordId && detailQuery.isError ? <ErrorState description={detailQuery.error.message} /> : null}
              {detailQuery.data ? (
                <div className={styles.form}>
                  <div>
                    <Body1>{detailQuery.data.title || '未命名消息'}</Body1>
                    <div className={styles.checklist}>
                      <Badge appearance="tint">{detailQuery.data.messageType}</Badge>
                      <Badge appearance="outline">{detailQuery.data.status}</Badge>
                    </div>
                  </div>
                  {detailQuery.data.summary ? <Textarea readOnly resize="vertical" value={detailQuery.data.summary} /> : null}
                  <PropertyGrid
                    items={[
                      { label: '发送人', value: detailQuery.data.senderName || '-' },
                      { label: '模板', value: detailQuery.data.templateName || '-' },
                      { label: '受众类型', value: detailQuery.data.audienceType || '-' },
                      { label: '目标团队', value: detailQuery.data.targetTenantName || '-' },
                    ]}
                  />
                  <div className={styles.content}>{detailQuery.data.content || '当前消息没有正文。'}</div>
                </div>
              ) : null}
            </SectionCard>

            {detailQuery.data ? (
              <SectionCard title="投递指标" description="投递、已读和待办状态统一展示。">
                <MetricGrid metrics={recordMetrics} />
              </SectionCard>
            ) : null}

            {detailQuery.data?.timeline.length ? (
              <SectionCard title="状态时间线" description="消息生命周期中的关键时间点。">
                <DetailTimeline items={detailQuery.data.timeline} />
              </SectionCard>
            ) : null}

            {detailQuery.data?.deliveries.length ? (
              <SectionCard title="投递明细" description="收件人、状态和来源规则都会显示在这里。">
                <SimpleTable
                  columns={[
                    { key: 'recipient', header: '收件人', render: (item) => item.recipientName || '-' },
                    { key: 'team', header: '团队', render: (item) => item.recipientTeamName || '-' },
                    { key: 'status', header: '状态', render: (item) => item.deliveryStatus || '-' },
                    { key: 'source', header: '来源', render: (item) => item.sourceRuleLabel || item.sourceGroupName || '-' },
                  ]}
                  items={detailQuery.data.deliveries}
                  rowKey={(item) => item.id}
                />
              </SectionCard>
            ) : null}

            {detailQuery.data && buildPayloadItems(detailQuery.data.payload).length ? (
              <SectionCard title="Payload 摘要" description="保留 payload 联调信息，但不再用整段 JSON 覆盖详情区。">
                <PropertyGrid items={buildPayloadItems(detailQuery.data.payload)} />
              </SectionCard>
            ) : null}
          </WorkbenchStack>
        }
      />
    </PageContainer>
  )
}

function MessageMorePage({ routeId, scope, title }: { routeId: string; scope: 'platform' | 'team'; title: string }) {
  const base = scope === 'platform' ? '/system' : '/team'
  const prefix = scope === 'platform' ? '系统域' : '团队域'

  return (
    <PageContainer routeId={routeId}>
      <SectionCard title={title} description="消息域的次级入口被整理成目录卡片，便于按任务进入相应治理页。">
        <LinkCardGrid
          items={[
            { id: 'dispatch', title: `${prefix}消息调度`, description: '用于发起真实消息调度并查看最近发送记录。', to: `${base}/message`, actionLabel: '打开调度台' },
            { id: 'template', title: `${prefix}消息模板`, description: '维护消息标题、摘要和正文模板。', to: `${base}/message-template`, actionLabel: '进入模板治理' },
            { id: 'sender', title: `${prefix}消息发送人`, description: '维护消息发送人身份和默认发送配置。', to: `${base}/message-sender`, actionLabel: '进入发送人治理' },
            { id: 'group', title: `${prefix}收件组`, description: '治理固定受众集合、匹配模式和目标摘要。', to: `${base}/message-recipient-group`, actionLabel: '进入收件组治理' },
            { id: 'record', title: `${prefix}消息记录`, description: '查看投递状态、时间线和 payload 摘要。', to: `${base}/message-record`, actionLabel: '查看消息记录' },
          ]}
        />
      </SectionCard>
    </PageContainer>
  )
}

export function SystemMessageTemplateWorkspace({ routeId }: { routeId: string }) {
  return <MessageTemplatePage routeId={routeId} scope="platform" title="系统消息模板" />
}

export function TeamMessageTemplateWorkspace({ routeId }: { routeId: string }) {
  return <MessageTemplatePage routeId={routeId} scope="team" title="团队消息模板" />
}

export function SystemMessageSenderWorkspace({ routeId }: { routeId: string }) {
  return <MessageSenderPage routeId={routeId} scope="platform" title="系统消息发送人" />
}

export function TeamMessageSenderWorkspace({ routeId }: { routeId: string }) {
  return <MessageSenderPage routeId={routeId} scope="team" title="团队消息发送人" />
}

export function SystemMessageRecipientGroupWorkspace({ routeId }: { routeId: string }) {
  return <MessageRecipientGroupPage routeId={routeId} scope="platform" title="系统收件组" />
}

export function TeamMessageRecipientGroupWorkspace({ routeId }: { routeId: string }) {
  return <MessageRecipientGroupPage routeId={routeId} scope="team" title="团队收件组" />
}

export function SystemMessageRecordWorkspace({ routeId }: { routeId: string }) {
  return <MessageRecordPage routeId={routeId} scope="platform" title="系统消息记录" />
}

export function TeamMessageRecordWorkspace({ routeId }: { routeId: string }) {
  return <MessageRecordPage routeId={routeId} scope="team" title="团队消息记录" />
}

export function SystemMessageMoreWorkspace({ routeId }: { routeId: string }) {
  return <MessageMorePage routeId={routeId} scope="platform" title="系统消息更多入口" />
}

export function TeamMessageMoreWorkspace({ routeId }: { routeId: string }) {
  return <MessageMorePage routeId={routeId} scope="team" title="团队消息更多入口" />
}
