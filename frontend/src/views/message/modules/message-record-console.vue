<template>
  <div class="message-record-page art-full-height">
    <AdminWorkspaceHero :title="pageTitle" :description="pageDescription" :metrics="heroMetrics">
      <div class="message-record-hero__actions">
        <ElButton @click="loadRecords" :loading="loading" v-ripple>刷新</ElButton>
      </div>
    </AdminWorkspaceHero>

    <MessageWorkspaceNav :scope="props.scope" current="record" />

    <ElAlert
      v-if="loadError"
      class="message-record-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="message-record-shell art-card">
      <header class="message-record-toolbar">
        <ElInput
          v-model="filters.keyword"
          clearable
          placeholder="搜索标题、摘要、正文或发送人"
          @keyup.enter="handleFilterChange"
          @clear="handleFilterChange"
        >
          <template #append>
            <ElButton @click="handleFilterChange">查询</ElButton>
          </template>
        </ElInput>

        <ElSelect
          v-model="filters.messageType"
          clearable
          placeholder="消息类型"
          @change="handleFilterChange"
        >
          <ElOption label="通知" value="notice" />
          <ElOption label="消息" value="message" />
          <ElOption label="待办" value="todo" />
        </ElSelect>

        <ElSelect
          v-model="filters.audienceType"
          clearable
          placeholder="发送对象"
          @change="handleFilterChange"
        >
          <ElOption
            v-for="item in audienceOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </header>

      <ElTable
        v-loading="loading"
        :data="list"
        class="message-record-table"
        @row-click="openDetail"
      >
        <ElTableColumn label="标题" min-width="260">
          <template #default="{ row }">
            <div class="message-record-title-cell">
              <div class="message-record-title-cell__title">{{ row.title }}</div>
              <div class="message-record-title-cell__sub">
                {{ plainTextFromHtml(row.summary || row.content) || '未填写摘要或正文' }}
              </div>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="类型" width="110">
          <template #default="{ row }">
            <ElTag effect="plain">{{ resolveMessageTypeLabel(row.message_type) }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="对象" width="140">
          <template #default="{ row }">
            <div class="message-record-simple-cell">
              <span>{{ resolveAudienceLabel(row.audience_type) }}</span>
              <small>{{
                row.target_collaboration_workspace_name || row.target_tenant_name || '全局范围'
              }}</small>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="投递情况" width="160">
          <template #default="{ row }">
            <div class="message-record-simple-cell">
              <span>{{ row.delivery_count }} 人</span>
              <small>已读 {{ row.read_count }} / 未读 {{ row.unread_count }}</small>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="模板 / 发送人" min-width="180">
          <template #default="{ row }">
            <div class="message-record-simple-cell">
              <span>{{ row.template_name || '直接发送' }}</span>
              <small>{{ row.sender_name || '系统' }}</small>
            </div>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="120">
          <template #default="{ row }">
            <ElTag size="small" effect="plain" :type="resolveRecordStatusTagType(row.status)">
              {{ resolveRecordStatusLabel(row.status) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="时间" width="150">
          <template #default="{ row }">
            <div class="message-record-simple-cell">
              <span>{{ formatTime(row.published_at || row.created_at) }}</span>
              <small>{{ row.scope_type === 'team' ? '协作空间发送' : '平台发送' }}</small>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>

      <footer class="message-record-footer">
        <WorkspacePagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :total="pagination.total"
          layout="prev, pager, next"
        />
      </footer>
    </section>

    <ElDrawer
      v-model="drawerVisible"
      title="发送详情"
      size="720px"
      destroy-on-close
      append-to-body
      class="message-record-drawer"
    >
      <template v-if="activeRecord">
        <div class="message-record-drawer__summary">
          <div>
            <div class="message-record-drawer__title">{{ activeRecord.title }}</div>
            <div class="message-record-drawer__meta">
              <span>{{ resolveMessageTypeLabel(activeRecord.message_type) }}</span>
              <span>{{ resolveAudienceLabel(activeRecord.audience_type) }}</span>
              <span>{{
                activeRecord.target_collaboration_workspace_name ||
                activeRecord.target_tenant_name ||
                (activeRecord.scope_type === 'team' ? currentTeamName : '平台范围')
              }}</span>
            </div>
          </div>
          <ElTag effect="plain">{{
            activeRecord.scope_type === 'team' ? '协作空间发送' : '平台发送'
          }}</ElTag>
        </div>

        <div class="message-record-drawer__stats">
          <div class="message-record-stat">
            <strong>{{ activeRecord.delivery_count }}</strong>
            <span>投递人数</span>
          </div>
          <div class="message-record-stat">
            <strong>{{ activeRecord.read_count }}</strong>
            <span>已读</span>
          </div>
          <div class="message-record-stat">
            <strong>{{ activeRecord.unread_count }}</strong>
            <span>未读</span>
          </div>
          <div class="message-record-stat">
            <strong>{{ activeRecord.pending_todo_count }}</strong>
            <span>待处理</span>
          </div>
        </div>

        <div class="message-record-delivery-block">
          <div class="message-record-delivery-block__header">
            <h4>投递明细</h4>
            <span>{{ activeRecord.deliveries?.length || 0 }} 条</span>
          </div>
          <ElAlert
            v-if="detailError && !detailLoading"
            class="message-record-detail-alert"
            type="info"
            :closable="false"
            show-icon
            :title="detailError"
          />
          <ElSkeleton v-if="detailLoading" :rows="5" animated />
          <div v-else-if="activeRecord.deliveries?.length" class="message-record-delivery-list">
            <article
              v-for="delivery in activeRecord.deliveries"
              :key="delivery.id"
              class="message-record-delivery-card"
            >
              <div class="message-record-delivery-card__top">
                <div>
                  <strong>{{ delivery.recipient_name || '未命名用户' }}</strong>
                  <p>{{ delivery.recipient_team_name || '平台用户' }}</p>
                </div>
                <ElTag size="small" effect="plain">{{
                  resolveDeliveryStatusLabel(delivery.delivery_status)
                }}</ElTag>
              </div>
              <div class="message-record-delivery-card__meta">
                <span>来源组：{{ delivery.source_group_name || '直接命中' }}</span>
                <span>规则：{{ resolveSourceRuleLabel(delivery) }}</span>
                <span>目标项：{{ resolveSourceTargetLabel(delivery) }}</span>
                <span v-if="resolveSourceOriginLabel(delivery)"
                  >来源链：{{ resolveSourceOriginLabel(delivery) }}</span
                >
                <span v-if="delivery.todo_status && delivery.todo_status !== 'pending'">
                  待办：{{ resolveTodoStatusLabel(delivery.todo_status) }}
                </span>
                <span
                  >最近动作：{{
                    formatTime(delivery.last_action_at || delivery.done_at || delivery.read_at)
                  }}</span
                >
              </div>
            </article>
          </div>
          <ElEmpty v-else description="暂无投递明细" />
        </div>

        <div class="message-record-rich-text">
          <h4>摘要</h4>
          <div
            class="rich-text-content"
            v-html="renderRichText(activeRecord.summary, '未填写摘要')"
            @click="handleRichTextClick"
          ></div>
        </div>

        <div class="message-record-rich-text">
          <h4>正文</h4>
          <div
            class="rich-text-content"
            v-html="renderRichText(activeRecord.content, '未填写正文')"
            @click="handleRichTextClick"
          ></div>
        </div>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { useRouter } from 'vue-router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import MessageWorkspaceNav from '@/views/message/modules/message-workspace-nav.vue'
  import { fetchGetDispatchRecordDetail, fetchGetDispatchRecordList } from '@/api/message'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { handleRichTextLinkNavigation } from '@/utils/navigation/rich-text'
  import { useMessageWorkspace } from '@/views/message/modules/useMessageWorkspace'

  defineOptions({ name: 'MessageRecordConsole' })

  const props = defineProps<{
    scope: 'platform' | 'team'
  }>()

  const router = useRouter()
  const menuSpaceStore = useMenuSpaceStore()
  const {
    isTeamScope,
    skipTenantHeader,
    currentTeamName,
    ensureTeamContext,
    plainTextFromHtml,
    formatTime
  } = useMessageWorkspace(props.scope)
  const loading = ref(false)
  const loadError = ref('')
  const detailLoading = ref(false)
  const detailError = ref('')
  const drawerVisible = ref(false)
  const list = ref<Api.Message.DispatchRecordItem[]>([])
  const activeRecord = ref<Api.Message.DispatchRecordDetail | null>(null)
  const summary = reactive<Api.Message.DispatchRecordSummary>({
    total_messages: 0,
    total_deliveries: 0,
    read_deliveries: 0,
    todo_messages: 0
  })
  const filters = reactive({
    keyword: '',
    messageType: '' as Api.Message.BoxType | '',
    audienceType: '' as Api.Message.AudienceType | ''
  })
  const pagination = reactive({
    current: 1,
    size: 12,
    total: 0
  })

  const pageTitle = computed(() => (isTeamScope.value ? '协作空间发送记录' : '消息发送记录'))
  const pageDescription = computed(() =>
    isTeamScope.value
      ? '查看当前协作空间发出的通知、消息和待办投递情况，重点看已读率和待处理数量。'
      : '查看平台侧消息发送结果和投递情况，判断哪些消息已读、未读或仍处于待处理状态。'
  )

  const audienceOptions = computed(() =>
    isTeamScope.value
      ? [
          { label: '当前协作空间成员', value: 'tenant_users' as Api.Message.AudienceType },
          { label: '指定成员', value: 'specified_users' as Api.Message.AudienceType },
          { label: '接收组', value: 'recipient_group' as Api.Message.AudienceType },
          { label: '角色规则', value: 'role' as Api.Message.AudienceType },
          { label: '功能包规则', value: 'feature_package' as Api.Message.AudienceType }
        ]
      : [
          { label: '所有用户', value: 'all_users' as Api.Message.AudienceType },
          { label: '协作空间管理员', value: 'tenant_admins' as Api.Message.AudienceType },
          { label: '指定协作空间成员', value: 'tenant_users' as Api.Message.AudienceType },
          { label: '指定用户', value: 'specified_users' as Api.Message.AudienceType },
          { label: '接收组', value: 'recipient_group' as Api.Message.AudienceType },
          { label: '角色规则', value: 'role' as Api.Message.AudienceType },
          { label: '功能包规则', value: 'feature_package' as Api.Message.AudienceType }
        ]
  )

  const heroMetrics = computed(() => [
    { label: '发送次数', value: summary.total_messages },
    { label: '总投递', value: summary.total_deliveries },
    { label: '已读', value: summary.read_deliveries },
    { label: '待办消息', value: summary.todo_messages }
  ])

  const renderRichText = (value?: string, fallback = '') => {
    const target = `${value || ''}`.trim()
    return plainTextFromHtml(target) ? target : `<p>${fallback}</p>`
  }

  const resolveMessageTypeLabel = (value: Api.Message.BoxType) => {
    if (value === 'message') return '消息'
    if (value === 'todo') return '待办'
    return '通知'
  }

  const resolveAudienceLabel = (value: Api.Message.AudienceType) => {
    if (value === 'all_users') return '所有用户'
    if (value === 'tenant_admins') return '协作空间管理员'
    if (value === 'specified_users') return '指定用户'
    if (value === 'recipient_group') return '接收组'
    if (value === 'role') return '角色规则'
    if (value === 'feature_package') return '功能包规则'
    return '协作空间成员'
  }

  const resolveDeliveryStatusLabel = (value?: string) => {
    if (value === 'read') return '已读'
    return '未读'
  }

  const resolveRecordStatusLabel = (value?: string) => {
    if (value === 'queued') return '排队中'
    if (value === 'processing') return '发送中'
    if (value === 'failed') return '发送失败'
    return '已发送'
  }

  const resolveRecordStatusTagType = (value?: string) => {
    if (value === 'queued' || value === 'processing') return 'warning'
    if (value === 'failed') return 'danger'
    return 'success'
  }

  const resolveTodoStatusLabel = (value?: string) => {
    if (value === 'done') return '已完成'
    if (value === 'ignored') return '已忽略'
    return '待处理'
  }

  const resolveSourceRuleLabel = (delivery: Api.Message.DispatchRecordDeliveryItem) => {
    if (delivery.source_rule_label) {
      return delivery.source_rule_label.replace(
        /\s*·\s*(workspace_role_binding|legacy_user_role|membership_identity)\s*$/u,
        ''
      )
    }
    if (delivery.source_rule_type === 'all_users') return '所有用户'
    if (delivery.source_rule_type === 'tenant_admins') return '协作空间管理员'
    if (delivery.source_rule_type === 'tenant_users') return '协作空间成员'
    if (delivery.source_rule_type === 'specified_users') return '指定用户'
    if (delivery.source_rule_type === 'recipient_group') return '接收组'
    if (delivery.source_rule_type === 'feature_package') return '功能包规则'
    if (delivery.source_rule_type === 'role') return '角色规则'
    return '直接命中'
  }

  const resolveSourceTargetLabel = (delivery: Api.Message.DispatchRecordDeliveryItem) => {
    if (delivery.source_target_type === 'user')
      return delivery.source_target_value || delivery.recipient_user_id
    if (delivery.source_target_type === 'tenant_users')
      return delivery.recipient_team_name || currentTeamName.value
    if (delivery.source_target_type === 'tenant_admins')
      return delivery.recipient_team_name || currentTeamName.value
    if (delivery.source_target_type === 'role') return delivery.source_target_value || '角色规则'
    if (delivery.source_target_type === 'feature_package')
      return delivery.source_target_value || '功能包规则'
    if (delivery.source_target_type === 'all_users') return '所有用户'
    return delivery.source_target_value || '直接命中'
  }

  const resolveSourceOriginLabel = (delivery: Api.Message.DispatchRecordDeliveryItem) => {
    const label = delivery.source_rule_label || ''
    if (label.includes('workspace_role_binding')) return 'workspace_role_binding'
    if (label.includes('legacy_user_role')) return 'legacy_user_role'
    if (label.includes('membership_identity')) return 'membership_identity'
    return ''
  }

  const loadRecords = async () => {
    loading.value = true
    loadError.value = ''
    try {
      ensureTeamContext()
      const result = await fetchGetDispatchRecordList(
        {
          keyword: filters.keyword || undefined,
          message_type: filters.messageType || undefined,
          audience_type: filters.audienceType || undefined,
          current: pagination.current,
          size: pagination.size
        },
        { skipTenantHeader: skipTenantHeader.value }
      )
      list.value = result.records || []
      pagination.total = result.total || 0
      Object.assign(summary, result.summary || {})
    } catch {
      list.value = []
      pagination.total = 0
      Object.assign(summary, {
        total_messages: 0,
        total_deliveries: 0,
        read_deliveries: 0,
        todo_messages: 0
      })
      loadError.value = '发送记录暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  const handleFilterChange = async () => {
    pagination.current = 1
    await loadRecords()
  }

  const openDetail = async (row: Api.Message.DispatchRecordItem) => {
    drawerVisible.value = true
    detailLoading.value = true
    detailError.value = ''
    activeRecord.value = {
      ...row,
      deliveries: []
    }
    try {
      activeRecord.value = await fetchGetDispatchRecordDetail(row.id, {
        skipTenantHeader: skipTenantHeader.value
      })
    } catch {
      detailError.value = '发送详情暂时不可用，稍后重试。'
    } finally {
      detailLoading.value = false
    }
  }

  const handleRichTextClick = async (event: MouseEvent) => {
    await handleRichTextLinkNavigation(event, {
      router,
      spaceResolver: menuSpaceStore
    })
  }

  onMounted(() => {
    loadRecords()
  })

  watch(
    () => [pagination.current, pagination.size],
    ([current, size], [oldCurrent, oldSize]) => {
      if (current === oldCurrent && size === oldSize) return
      loadRecords()
    }
  )
</script>

<style scoped lang="scss">
  .message-record-page {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .message-record-inline-alert {
    margin-top: 0;
  }

  .message-record-hero__actions {
    display: flex;
    gap: 12px;
  }

  .message-record-shell {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 18px;
    border-radius: 24px;
  }

  .message-record-toolbar {
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) 180px 180px;
    gap: 12px;
  }

  .message-record-table {
    width: 100%;
  }

  .message-record-title-cell,
  .message-record-simple-cell {
    display: grid;
    gap: 4px;
  }

  .message-record-title-cell__title {
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-record-title-cell__sub,
  .message-record-simple-cell small {
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-record-footer {
    display: flex;
    justify-content: flex-end;
  }

  .message-record-drawer__summary {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.96), rgb(255 255 255 / 0.98));
  }

  .message-record-drawer__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-record-drawer__meta {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    margin-top: 6px;
    font-size: 12px;
    color: #64748b;
  }

  .message-record-drawer__stats {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
    margin-top: 16px;
  }

  .message-record-stat {
    display: grid;
    gap: 4px;
    padding: 14px 12px;
    border-radius: 16px;
    background: rgb(248 250 252 / 0.92);
    text-align: center;
  }

  .message-record-stat strong {
    font-size: 20px;
    color: #0f172a;
  }

  .message-record-stat span {
    font-size: 12px;
    color: #64748b;
  }

  .message-record-rich-text {
    margin-top: 16px;
    padding: 16px;
    border-radius: 18px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-record-delivery-block {
    margin-top: 16px;
    padding: 16px;
    border-radius: 18px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-record-detail-alert {
    margin-bottom: 12px;
  }

  .message-record-delivery-block__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .message-record-delivery-block__header h4 {
    margin: 0;
    font-size: 13px;
    color: #0f172a;
  }

  .message-record-delivery-block__header span {
    font-size: 12px;
    color: #64748b;
  }

  .message-record-delivery-list {
    display: grid;
    gap: 12px;
  }

  .message-record-delivery-card {
    padding: 14px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.96);
  }

  .message-record-delivery-card__top {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .message-record-delivery-card__top strong {
    display: block;
    font-size: 13px;
    color: #0f172a;
  }

  .message-record-delivery-card__top p {
    margin: 4px 0 0;
    font-size: 12px;
    color: #64748b;
  }

  .message-record-delivery-card__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-top: 12px;
    font-size: 12px;
    color: #64748b;
  }

  .message-record-rich-text h4 {
    margin: 0 0 12px;
    font-size: 13px;
    color: #0f172a;
  }

  .rich-text-content {
    font-size: 13px;
    line-height: 1.8;
    color: #334155;
  }

  .rich-text-content :deep(a) {
    color: #2563eb;
    text-decoration: underline;
  }

  @media (max-width: 1080px) {
    .message-record-toolbar {
      grid-template-columns: 1fr;
    }

    .message-record-drawer__stats {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }
</style>

