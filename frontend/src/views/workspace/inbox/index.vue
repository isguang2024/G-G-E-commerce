<template>
  <div class="message-center-page art-full-height">
    <AdminWorkspaceHero
      :title="`消息中心 · ${workspaceName}`"
      description="统一查看当前空间通知、直接消息和待处理事项。当前授权工作空间决定权限来源，协作空间视图只作为协作空间边界的兼容派生。"
      :metrics="heroMetrics"
    >
      <div class="message-center-hero__actions">
        <ElButton @click="reloadData" :loading="loading" v-ripple>刷新</ElButton>
        <ElButton
          type="primary"
          plain
          @click="markCurrentBoxRead"
          :disabled="activeBoxType === 'todo'"
          v-ripple
        >
          当前分类全部已读
        </ElButton>
      </div>
    </AdminWorkspaceHero>

    <ElAlert
      v-if="loadError"
      class="message-center-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="message-center-shell">
      <aside class="message-center-sidebar">
        <div class="message-center-sidebar__header">消息分类</div>
        <button
          v-for="item in boxOptions"
          :key="item.value"
          type="button"
          class="message-center-filter"
          :class="{ 'is-active': activeBoxType === item.value }"
          @click="switchBoxType(item.value)"
        >
          <span>{{ item.label }}</span>
          <strong>{{ item.count }}</strong>
        </button>

        <div class="message-center-sidebar__switch">
          <span>仅看未读</span>
          <ElSwitch v-model="filters.unreadOnly" @change="handleFilterChange" />
        </div>
      </aside>

      <div class="message-center-list">
        <header class="message-center-list__toolbar">
          <div>
            <h3>{{ activeBoxLabel }}</h3>
            <p>{{
              filters.unreadOnly
                ? `当前只显示 ${workspaceLabel} 下的未读内容`
                : `按时间倒序显示 ${workspaceLabel} 下的最近消息`
            }}</p>
          </div>
          <span class="message-center-list__count">共 {{ pagination.total }} 条</span>
        </header>

        <div v-loading="loading" class="message-center-list__body">
          <button
            v-for="item in list"
            :key="item.id"
            type="button"
            class="message-center-item"
            :class="{
              'is-active': selectedId === item.id,
              'is-unread': item.delivery_status === 'unread'
            }"
            @click="selectMessage(item)"
          >
            <div class="message-center-item__top">
              <div class="message-center-item__title-wrap">
                <span class="message-center-item__title">{{ item.title }}</span>
                <span
                  v-if="item.delivery_status === 'unread'"
                  class="message-center-item__dot"
                ></span>
              </div>
              <span class="message-center-item__time">{{
                formatTime(item.last_action_at || item.published_at || item.created_at)
              }}</span>
            </div>
            <p class="message-center-item__summary">{{ resolveSummary(item) }}</p>
            <div class="message-center-item__meta">
              <span class="message-center-chip">{{ resolveSender(item) }}</span>
              <span
                v-if="resolveCollaborationWorkspaceTag(item)"
                class="message-center-chip is-collaboration-workspace"
                >{{ resolveCollaborationWorkspaceTag(item) }}</span
              >
              <span class="message-center-chip is-soft">{{ resolveTypeLabel(item) }}</span>
            </div>
          </button>

          <ElEmpty v-if="!loading && !list.length" description="当前筛选下没有消息" />
        </div>

        <footer class="message-center-list__footer">
          <ElPagination
            v-model:current-page="pagination.current"
            v-model:page-size="pagination.size"
            layout="prev, pager, next"
            :total="pagination.total"
            @current-change="handlePageChange"
          />
        </footer>
      </div>

      <section class="message-center-detail">
        <div v-if="detail" class="message-center-detail__inner" v-loading="detailLoading">
          <header class="message-center-detail__header">
            <div>
              <div class="message-center-detail__eyebrow">{{ resolveTypeLabel(detail) }}</div>
              <h2>{{ detail.title }}</h2>
              <p
                >{{ resolveSender(detail) }} ·
                {{ formatTime(detail.published_at || detail.created_at) }}</p
              >
            </div>
            <div class="message-center-detail__status">
              <span
                v-if="resolveCollaborationWorkspaceTag(detail)"
                class="message-center-chip is-collaboration-workspace"
                >{{ resolveCollaborationWorkspaceTag(detail) }}</span
              >
              <span class="message-center-chip">{{
                detail.delivery_status === 'unread' ? '未读' : '已读'
              }}</span>
              <span v-if="detail.box_type === 'todo'" class="message-center-chip is-danger">
                {{ resolveTodoStatus(detail.todo_status) }}
              </span>
            </div>
          </header>

          <div
            class="message-center-detail__summary rich-text-content"
            v-html="renderRichText(detail.summary, '这条消息没有额外摘要。')"
            @click="handleRichTextClick"
          ></div>
          <ElAlert
            v-if="detailError && !detailLoading"
            class="message-center-detail-alert"
            type="info"
            :closable="false"
            show-icon
            :title="detailError"
          />
          <article
            class="message-center-detail__content rich-text-content"
            v-html="renderRichText(detail.content, '这条消息没有正文内容。')"
            @click="handleRichTextClick"
          ></article>

          <div class="message-center-detail__actions">
            <ElButton
              v-if="detail.box_type === 'todo'"
              type="success"
              plain
              @click="handleTodoAction('done')"
              v-ripple
            >
              标记完成
            </ElButton>
            <ElButton
              v-if="detail.box_type === 'todo'"
              plain
              @click="handleTodoAction('ignored')"
              v-ripple
            >
              忽略待办
            </ElButton>
          </div>
        </div>

        <ElEmpty v-else description="从左侧选择一条消息查看详情" />
      </section>
    </section>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRoute, useRouter } from 'vue-router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { fetchGetInboxList, fetchGetInboxDetail } from '@/api/message'
  import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
  import { useMessageStore } from '@/store/modules/message'
  import { useCollaborationStore } from '@/store/modules/collaboration'
  import { useWorkspaceStore } from '@/store/modules/workspace'
  import { handleRichTextLinkNavigation } from '@/domains/navigation/utils/rich-text'

  defineOptions({ name: 'WorkspaceInbox' })

  type BoxValue = Api.Message.BoxType | ''

  const route = useRoute()
  const router = useRouter()
  const menuSpaceStore = useMenuSpaceStore()
  const messageStore = useMessageStore()
  const collaborationStore = useCollaborationStore()
  const workspaceStore = useWorkspaceStore()

  const loading = ref(false)
  const loadError = ref('')
  const detailLoading = ref(false)
  const detailError = ref('')
  const list = ref<Api.Message.InboxItem[]>([])
  const detail = ref<Api.Message.InboxDetail | null>(null)
  const selectedId = ref('')

  const filters = reactive({
    boxType: '' as BoxValue,
    unreadOnly: false
  })

  const pagination = reactive({
    current: 1,
    size: 20,
    total: 0
  })

  const boxOptions = computed(() => [
    { label: '全部', value: '' as BoxValue, count: messageStore.summary.unread_total },
    { label: '通知', value: 'notice' as BoxValue, count: messageStore.summary.notice_count },
    { label: '消息', value: 'message' as BoxValue, count: messageStore.summary.message_count },
    { label: '待办', value: 'todo' as BoxValue, count: messageStore.summary.todo_count }
  ])

  const activeBoxType = computed(() => filters.boxType)
  const activeBoxLabel = computed(
    () => boxOptions.value.find((item) => item.value === filters.boxType)?.label || '全部'
  )

  const heroMetrics = computed(() => [
    { label: '未读总数', value: messageStore.summary.unread_total },
    { label: '通知', value: messageStore.summary.notice_count },
    { label: '消息', value: messageStore.summary.message_count },
    { label: '待办', value: messageStore.summary.todo_count }
  ])
  const workspaceLabel = computed(() =>
    workspaceStore.currentAuthWorkspaceType === 'collaboration' ? '协作空间' : '个人空间'
  )
  const workspaceName = computed(
    () => workspaceStore.currentAuthWorkspace?.name || '当前授权工作空间'
  )

  const formatTime = (value?: string) => {
    if (!value) return '刚刚'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date)
  }

  const resolveSender = (item: Api.Message.InboxItem) => {
    if (item.sender_name_snapshot) return item.sender_name_snapshot
    if (item.sender_type === 'system') return '系统'
    if (item.sender_type === 'service') return item.sender_service_key || '服务'
    if (item.sender_type === 'automation') return '自动任务'
    return '站内消息'
  }

  const resolveTypeLabel = (item: Api.Message.InboxItem) => {
    if (item.box_type === 'notice') return '通知'
    if (item.box_type === 'message') return '消息'
    return '待办'
  }

  const plainTextFromHtml = (value?: string) => {
    const target = `${value || ''}`.trim()
    if (!target) return ''
    if (typeof window === 'undefined') {
      return target
        .replace(/<[^>]+>/g, ' ')
        .replace(/&nbsp;/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
    }
    const parser = new DOMParser()
    const doc = parser.parseFromString(target, 'text/html')
    return (doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  }

  const renderRichText = (value?: string, fallback = '') => {
    const target = `${value || ''}`.trim()
    return plainTextFromHtml(target) ? target : `<p>${fallback}</p>`
  }

  const resolveSummary = (item: Api.Message.InboxItem) => {
    return plainTextFromHtml(item.summary) || plainTextFromHtml(item.content) || '点击查看详情'
  }

  const resolveCollaborationWorkspaceName = (item: Api.Message.InboxItem) => {
    const collaborationWorkspaceId =
      item.scope_type === 'collaboration'
        ? item.scope_id ||
          item.recipient_collaboration_workspace_id ||
          item.target_collaboration_workspace_id ||
          item.recipient_collaboration_workspace_id
        : item.recipient_collaboration_workspace_id ||
          item.target_collaboration_workspace_id ||
          item.recipient_collaboration_workspace_id ||
          item.scope_id
    if (!collaborationWorkspaceId) return ''
    return (
      collaborationStore.collaborationList.find(
        (workspace) => workspace.id === collaborationWorkspaceId
      )?.name || ''
    )
  }

  const resolveCollaborationWorkspaceTag = (item: Api.Message.InboxItem) => {
    const collaborationWorkspaceName = resolveCollaborationWorkspaceName(item)
    return collaborationWorkspaceName ? `协作空间视图 · ${collaborationWorkspaceName}` : ''
  }

  const resolveTodoStatus = (value?: string) => {
    if (value === 'done') return '已完成'
    if (value === 'ignored') return '已忽略'
    return '待处理'
  }

  const syncRouteQuery = (deliveryId?: string) => {
    router.replace({
      path: '/workspace/inbox',
      query: {
        boxType: filters.boxType || undefined,
        unreadOnly: filters.unreadOnly ? '1' : undefined,
        deliveryId: deliveryId || undefined
      }
    })
  }

  const loadList = async () => {
    loading.value = true
    loadError.value = ''
    try {
      const response = await fetchGetInboxList({
        box_type: filters.boxType || undefined,
        unread_only: filters.unreadOnly,
        current: pagination.current,
        size: pagination.size
      })
      list.value = response.records || []
      pagination.total = response.total || 0

      const queryDeliveryId = `${route.query.deliveryId || ''}`.trim()
      const nextSelectedId =
        (queryDeliveryId &&
          list.value.some((item) => item.id === queryDeliveryId) &&
          queryDeliveryId) ||
        (selectedId.value &&
          list.value.some((item) => item.id === selectedId.value) &&
          selectedId.value) ||
        list.value[0]?.id ||
        ''

      if (nextSelectedId) {
        try {
          await loadDetail(nextSelectedId, false)
        } catch {
          const fallbackId = list.value[0]?.id || ''
          if (fallbackId && fallbackId !== nextSelectedId) {
            await loadDetail(fallbackId, false)
          } else {
            selectedId.value = ''
            detail.value = null
            syncRouteQuery()
          }
        }
      } else {
        selectedId.value = ''
        detail.value = null
        syncRouteQuery()
      }
    } catch {
      list.value = []
      pagination.total = 0
      selectedId.value = ''
      detail.value = null
      loadError.value = '消息中心暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  const loadDetail = async (deliveryId: string, markIfUnread = false) => {
    if (!deliveryId) return
    detailLoading.value = true
    detailError.value = ''
    try {
      const nextDetail = await fetchGetInboxDetail(deliveryId)
      detail.value = nextDetail
      selectedId.value = deliveryId
      syncRouteQuery(deliveryId)

      if (markIfUnread && nextDetail.delivery_status === 'unread') {
        await messageStore.markRead(deliveryId)
        if (detail.value) detail.value.delivery_status = 'read'
        const target = list.value.find((item) => item.id === deliveryId)
        if (target) target.delivery_status = 'read'
      }
    } catch (error) {
      detailError.value = '消息详情暂时不可用，稍后重试。'
      throw error
    } finally {
      detailLoading.value = false
    }
  }

  const reloadData = async () => {
    await messageStore.loadSummary(true)
    await loadList()
  }

  const handleFilterChange = async () => {
    pagination.current = 1
    await reloadData()
  }

  const handlePageChange = async () => {
    await loadList()
  }

  const switchBoxType = async (value: BoxValue) => {
    filters.boxType = value
    await handleFilterChange()
  }

  const selectMessage = async (item: Api.Message.InboxItem) => {
    try {
      await loadDetail(item.id, true)
    } catch {
      return
    }
  }

  const markCurrentBoxRead = async () => {
    if (filters.boxType === 'todo') return
    try {
      await messageStore.markReadAll(filters.boxType || undefined)
      await reloadData()
      ElMessage.success('当前分类已全部标记为已读')
    } catch {
      ElMessage.error('批量已读失败')
    }
  }

  const handleTodoAction = async (action: 'done' | 'ignored') => {
    if (!detail.value) return
    try {
      await messageStore.handleTodo(detail.value.id, action)
      await reloadData()
      await loadDetail(detail.value.id, false)
      ElMessage.success(action === 'done' ? '待办已完成' : '待办已忽略')
    } catch {
      ElMessage.error('处理待办失败')
    }
  }

  const handleRichTextClick = async (event: MouseEvent) => {
    await handleRichTextLinkNavigation(event, {
      router,
      spaceResolver: menuSpaceStore
    })
  }

  const applyRouteQuery = () => {
    const queryBoxType = `${route.query.boxType || ''}`.trim()
    filters.boxType =
      queryBoxType === 'notice' || queryBoxType === 'message' || queryBoxType === 'todo'
        ? (queryBoxType as BoxValue)
        : ''
    filters.unreadOnly =
      route.query.unreadOnly === '1' || `${route.query.unreadOnly || ''}`.toLowerCase() === 'true'
  }

  watch(
    () => route.query,
    () => {
      applyRouteQuery()
    }
  )

  onMounted(async () => {
    applyRouteQuery()
    await reloadData()
  })
</script>

<style scoped lang="scss">
  .message-center-page {
    min-height: 100%;
  }

  .message-center-inline-alert {
    margin-top: 0;
  }

  .message-center-hero__actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
  }

  .message-center-shell {
    display: grid;
    grid-template-columns: 220px minmax(320px, 420px) minmax(380px, 1fr);
    gap: 14px;
    min-height: calc(100vh - 230px);
  }

  .message-center-sidebar,
  .message-center-list,
  .message-center-detail {
    min-width: 0;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 22px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.98));
  }

  .message-center-sidebar {
    padding: 16px;
  }

  .message-center-sidebar__header {
    margin-bottom: 12px;
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #64748b;
  }

  .message-center-filter {
    display: flex;
    width: 100%;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    padding: 12px 14px;
    border: 1px solid transparent;
    border-radius: 16px;
    background: rgb(248 250 252 / 0.9);
    color: #0f172a;
    transition: 0.2s ease;
  }

  .message-center-filter.is-active {
    border-color: rgb(125 211 252 / 0.9);
    background: linear-gradient(135deg, rgb(239 246 255 / 0.95), rgb(236 253 245 / 0.9));
  }

  .message-center-sidebar__switch {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 18px;
    padding-top: 16px;
    border-top: 1px solid rgb(226 232 240 / 0.9);
    font-size: 13px;
    color: #475569;
  }

  .message-center-list {
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .message-center-list__toolbar,
  .message-center-detail__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 18px 20px 14px;
    border-bottom: 1px solid rgb(226 232 240 / 0.85);
  }

  .message-center-list__toolbar h3,
  .message-center-detail__header h2 {
    margin: 0;
    font-size: 20px;
    line-height: 1.2;
    color: #0f172a;
  }

  .message-center-list__toolbar p,
  .message-center-detail__header p {
    margin: 6px 0 0;
    font-size: 12px;
    color: #64748b;
  }

  .message-center-list__count {
    font-size: 12px;
    color: #64748b;
  }

  .message-center-list__body {
    flex: 1;
    overflow: auto;
    padding: 8px;
  }

  .message-center-item {
    width: 100%;
    margin-bottom: 8px;
    padding: 14px;
    text-align: left;
    border: 1px solid transparent;
    border-radius: 18px;
    background: rgb(255 255 255 / 0.85);
    transition: 0.2s ease;
  }

  .message-center-item:hover,
  .message-center-item.is-active {
    border-color: rgb(191 219 254 / 0.95);
    background: linear-gradient(135deg, rgb(248 250 252 / 1), rgb(239 246 255 / 0.92));
  }

  .message-center-item.is-unread {
    box-shadow: inset 3px 0 0 #3b82f6;
  }

  .message-center-item__top,
  .message-center-item__meta {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .message-center-item__title-wrap {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .message-center-item__title {
    overflow: hidden;
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .message-center-item__dot {
    width: 7px;
    height: 7px;
    border-radius: 999px;
    background: #ef4444;
  }

  .message-center-item__time {
    flex-shrink: 0;
    font-size: 11px;
    color: #94a3b8;
  }

  .message-center-item__summary {
    margin: 8px 0 12px;
    display: -webkit-box;
    overflow: hidden;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .message-center-chip {
    display: inline-flex;
    align-items: center;
    padding: 3px 9px;
    border-radius: 999px;
    background: rgb(226 232 240 / 0.72);
    font-size: 11px;
    color: #334155;
  }

  .message-center-chip.is-soft {
    background: rgb(241 245 249 / 0.95);
  }

  .message-center-chip.is-collaboration-workspace {
    background: rgb(219 234 254 / 0.9);
    color: #1d4ed8;
  }

  .message-center-chip.is-danger {
    background: rgb(254 226 226 / 1);
    color: #b91c1c;
  }

  .message-center-list__footer {
    display: flex;
    justify-content: flex-end;
    padding: 12px 16px 16px;
    border-top: 1px solid rgb(226 232 240 / 0.8);
  }

  .message-center-detail {
    overflow: hidden;
  }

  .message-center-detail__inner {
    height: 100%;
    overflow: auto;
  }

  .message-center-detail__eyebrow {
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #2563eb;
  }

  .message-center-detail__status {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .message-center-detail__summary {
    margin: 18px 20px 0;
    padding: 14px 16px;
    border-radius: 18px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-center-detail-alert {
    margin: 16px 20px 0;
  }

  .message-center-detail__content {
    padding: 18px 20px 0;
  }

  .rich-text-content {
    font-size: 14px;
    line-height: 1.8;
    color: #0f172a;
    word-break: break-word;
  }

  .rich-text-content :deep(p) {
    margin: 0 0 0.9em;
  }

  .rich-text-content :deep(p:last-child) {
    margin-bottom: 0;
  }

  .rich-text-content :deep(a) {
    color: #2563eb;
    text-decoration: underline;
  }

  .rich-text-content :deep(ul),
  .rich-text-content :deep(ol) {
    margin: 0 0 0.9em;
    padding-left: 1.35em;
  }

  .message-center-detail__actions {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    padding: 22px 20px 24px;
  }

  @media (max-width: 1280px) {
    .message-center-shell {
      grid-template-columns: 220px 1fr;
    }

    .message-center-detail {
      grid-column: 1 / -1;
      min-height: 320px;
    }
  }

  @media (max-width: 860px) {
    .message-center-shell {
      grid-template-columns: 1fr;
      min-height: auto;
    }
  }
</style>
