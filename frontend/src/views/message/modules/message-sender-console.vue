<template>
  <div class="message-sender-page art-full-height">
    <AdminWorkspaceHero :title="pageTitle" :description="pageDescription" :metrics="heroMetrics">
      <div class="message-sender-hero__actions">
        <ElButton @click="loadSenders" :loading="loading" v-ripple>刷新</ElButton>
        <ElButton type="primary" @click="openCreateDrawer" v-ripple>新建发送人</ElButton>
      </div>
    </AdminWorkspaceHero>

    <MessageWorkspaceNav :scope="props.scope" current="sender" />

    <ElAlert
      v-if="loadError"
      class="message-sender-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="message-sender-shell art-card">
      <header class="message-sender-shell__toolbar">
        <div>
          <div class="message-sender-shell__title">发送人列表</div>
          <p>{{ toolbarDescription }}</p>
        </div>
      </header>

      <div v-loading="loading" class="message-sender-board">
        <button
          v-for="item in pagedList"
          :key="item.id"
          type="button"
          class="message-sender-card"
          @click="openEditDrawer(item)"
        >
          <div class="message-sender-card__head">
            <div>
              <h3>{{ item.name }}</h3>
              <p>{{ item.description || '未填写发送人说明' }}</p>
            </div>
            <div class="message-sender-card__tags">
              <ElTag v-if="item.is_default" type="success" effect="plain" size="small">默认</ElTag>
              <ElTag
                size="small"
                :type="item.status === 'disabled' ? 'info' : 'primary'"
                effect="plain"
              >
                {{ item.status === 'disabled' ? '停用' : '正常' }}
              </ElTag>
            </div>
          </div>

          <div class="message-sender-card__preview">
            <div class="message-sender-card__label">消息中心展示</div>
            <div class="message-sender-card__name">{{ item.name }}</div>
          </div>

          <div class="message-sender-card__meta">
            <span>{{ item.scope_type === 'collaboration' ? '协作空间发送人' : '平台发送人' }}</span>
            <span>{{ formatTime(item.updated_at || item.created_at) }}</span>
          </div>
        </button>

        <ElEmpty v-if="!loading && !list.length" description="当前还没有可用发送人" />
      </div>

      <WorkspacePagination
        v-if="list.length > 0"
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.size"
        :total="list.length"
      />
    </section>

    <ElDrawer
      v-model="drawerVisible"
      :title="drawerEditingId ? '编辑发送人' : '新建发送人'"
      size="520px"
      destroy-on-close
      append-to-body
    >
      <template v-if="drawerModel">
        <div class="message-sender-drawer__summary">
          <div>
            <div class="message-sender-drawer__title">{{ drawerModel.name || '未命名发送人' }}</div>
            <div class="message-sender-drawer__text">{{ drawerSummary }}</div>
          </div>
          <ElTag effect="plain">{{ isTeamScope ? '协作空间发送人' : '平台发送人' }}</ElTag>
        </div>

        <div class="message-sender-drawer__form">
          <ElFormItem label="发送人名称">
            <ElInput
              v-model="drawerModel.name"
              maxlength="40"
              show-word-limit
              placeholder="例如：平台 / 平台管理 / 平台空间 / 协作空间"
            />
            <div class="field-hint">消息中心和右上角消息面板都会展示这个发送人名称。</div>
          </ElFormItem>

          <ElFormItem label="发送人说明">
            <ElInput
              v-model="drawerModel.description"
              type="textarea"
              :rows="3"
              placeholder="例如：平台统一通知发送身份"
            />
          </ElFormItem>

          <ElFormItem label="头像地址">
            <ElInput
              v-model="drawerModel.avatar_url"
              placeholder="可选，后续消息头像展示可直接复用"
            />
          </ElFormItem>

          <div class="message-sender-drawer__grid">
            <ElFormItem label="是否默认">
              <ElSwitch v-model="drawerModel.is_default" />
            </ElFormItem>

            <ElFormItem label="状态">
              <ElRadioGroup v-model="drawerModel.status">
                <ElRadioButton value="normal">正常</ElRadioButton>
                <ElRadioButton value="disabled">停用</ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
          </div>
        </div>
      </template>

      <template #footer>
        <div class="message-sender-drawer__footer">
          <ElButton @click="drawerVisible = false">关闭</ElButton>
          <ElButton type="primary" :loading="saving" @click="saveSender">保存发送人</ElButton>
        </div>
      </template>
    </ElDrawer>
  </div>
</template>

<script setup lang="ts">
  import { computed, onMounted, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import MessageWorkspaceNav from '@/views/message/modules/message-workspace-nav.vue'
  import {
    fetchCreateMessageSender,
    fetchGetMessageSenderList,
    fetchUpdateMessageSender
  } from '@/api/message'
  import { useMessageWorkspace } from '@/views/message/modules/useMessageWorkspace'

  defineOptions({ name: 'MessageSenderConsole' })

  const props = defineProps<{
    scope: 'platform' | 'collaboration'
  }>()

  interface SenderDrawerModel {
    name: string
    description: string
    avatar_url: string
    is_default: boolean
    status: 'normal' | 'disabled' | string
  }

  const {
    isTeamScope,
    skipTenantHeader,
    currentTeamName,
    currentWorkspaceName,
    currentWorkspaceLabel,
    ensureTeamContext,
    formatTime
  } = useMessageWorkspace(props.scope)

  const loading = ref(false)
  const loadError = ref('')
  const saving = ref(false)
  const list = ref<Api.Message.MessageSenderItem[]>([])
  const pagination = reactive({
    current: 1,
    size: 8
  })
  const drawerVisible = ref(false)
  const drawerEditingId = ref('')
  const drawerModel = ref<SenderDrawerModel | null>(null)

  const pageTitle = computed(() => (isTeamScope.value ? '协作空间发送人' : '发送人管理'))
  const pageDescription = computed(() =>
    isTeamScope.value
      ? `维护 ${currentWorkspaceName.value} 下 ${currentTeamName.value} 使用的协作空间消息发送人，默认发送人建议保持“协作空间”或更具体的协作空间身份。`
      : '维护平台消息发送人，默认发送人为“平台”，也可以扩展为平台管理、平台空间等发送身份。'
  )
  const toolbarDescription = computed(() =>
    isTeamScope.value
      ? '协作空间侧发送人只作用于当前协作空间消息发送页。'
      : '平台侧发送人只作用于平台消息发送页。'
  )
  const heroMetrics = computed(() => [
    { label: '发送人总数', value: list.value.length },
    { label: '默认发送人', value: list.value.find((item) => item.is_default)?.name || '未设置' },
    { label: '正常状态', value: list.value.filter((item) => item.status === 'normal').length }
  ])
  const drawerSummary = computed(() =>
    isTeamScope.value
      ? `保存后会作为 ${currentWorkspaceLabel.value} 下 ${currentTeamName.value} 的可选发送人。`
      : '保存后会作为平台消息发送页的可选发送人。'
  )

  const createDefaultModel = (): SenderDrawerModel => ({
    name: isTeamScope.value ? '协作空间' : '平台',
    description: '',
    avatar_url: '',
    is_default: list.value.length === 0,
    status: 'normal'
  })

  const loadSenders = async () => {
    loading.value = true
    loadError.value = ''
    try {
      ensureTeamContext()
      const result = await fetchGetMessageSenderList({
        skipTenantHeader: skipTenantHeader.value
      })
      list.value = result.records || []
      pagination.current = 1
    } catch {
      list.value = []
      loadError.value = '发送人列表暂时不可用，稍后重试或刷新状态。'
      pagination.current = 1
    } finally {
      loading.value = false
    }
  }

  const pagedList = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return list.value.slice(start, start + pagination.size)
  })

  const openCreateDrawer = () => {
    drawerEditingId.value = ''
    drawerModel.value = createDefaultModel()
    drawerVisible.value = true
  }

  const openEditDrawer = (item: Api.Message.MessageSenderItem) => {
    drawerEditingId.value = item.id
    drawerModel.value = {
      name: item.name,
      description: item.description || '',
      avatar_url: item.avatar_url || '',
      is_default: !!item.is_default,
      status: item.status
    }
    drawerVisible.value = true
  }

  const saveSender = async () => {
    if (!drawerModel.value) return
    if (!drawerModel.value.name.trim()) {
      ElMessage.warning('请先填写发送人名称')
      return
    }
    saving.value = true
    try {
      if (drawerEditingId.value) {
        await fetchUpdateMessageSender(drawerEditingId.value, drawerModel.value, {
          skipTenantHeader: skipTenantHeader.value
        })
      } else {
        await fetchCreateMessageSender(drawerModel.value, {
          skipTenantHeader: skipTenantHeader.value
        })
      }
      drawerVisible.value = false
      await loadSenders()
    } catch {
      ElMessage.error('保存发送人失败')
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadSenders()
  })

  watch(
    () => pagination.size,
    () => {
      pagination.current = 1
    }
  )
</script>

<style scoped lang="scss">
  .message-sender-page {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .message-sender-inline-alert {
    margin-top: 0;
  }

  .message-sender-hero__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .message-sender-shell {
    padding: 18px;
    border-radius: 24px;
  }

  .message-sender-shell__toolbar {
    padding-bottom: 14px;
    border-bottom: 1px solid rgb(226 232 240 / 0.85);
  }

  .message-sender-shell__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-sender-shell__toolbar p,
  .message-sender-drawer__text,
  .field-hint {
    margin: 6px 0 0;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-sender-board {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
    padding-top: 16px;
  }

  .message-sender-card {
    display: grid;
    gap: 12px;
    width: 100%;
    padding: 16px;
    border: 1px solid rgb(226 232 240 / 0.92);
    border-radius: 20px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.96));
    text-align: left;
  }

  .message-sender-card__head,
  .message-sender-drawer__summary {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .message-sender-card__head h3,
  .message-sender-drawer__title {
    margin: 0;
    font-size: 16px;
    color: #0f172a;
  }

  .message-sender-card__head p {
    margin: 6px 0 0;
    font-size: 12px;
    color: #64748b;
  }

  .message-sender-card__tags,
  .message-sender-card__meta,
  .message-sender-drawer__footer {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .message-sender-card__preview {
    padding: 14px;
    border-radius: 16px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-sender-card__label {
    font-size: 11px;
    color: #94a3b8;
  }

  .message-sender-card__name {
    margin-top: 6px;
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-sender-card__meta {
    font-size: 11px;
    color: #475569;
  }

  .message-sender-drawer__summary {
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.96), rgb(255 255 255 / 0.98));
  }

  .message-sender-drawer__form {
    display: grid;
    gap: 14px;
    margin-top: 16px;
  }

  .message-sender-drawer__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 14px;
  }

  .message-sender-drawer__footer {
    justify-content: flex-end;
    width: 100%;
  }

  @media (max-width: 1080px) {
    .message-sender-board {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .message-sender-drawer__grid {
      grid-template-columns: 1fr;
    }
  }
</style>
