<template>
  <div class="message-template-page art-full-height">
    <AdminWorkspaceHero :title="pageTitle" :description="pageDescription" :metrics="heroMetrics">
      <div class="message-template-hero__actions">
        <ElButton @click="loadTemplates" :loading="loading" v-ripple>刷新</ElButton>
        <ElButton type="primary" @click="openCreateDrawer" v-ripple>新建模板</ElButton>
      </div>
    </AdminWorkspaceHero>

    <MessageWorkspaceNav :scope="props.scope" current="template" />

    <ElAlert
      v-if="loadError"
      class="message-template-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="message-template-shell art-card">
      <header class="message-template-shell__toolbar">
        <div class="message-template-shell__toolbar-main">
          <div class="message-template-shell__title">模板列表</div>
          <p>{{ toolbarDescription }}</p>
        </div>
        <div class="message-template-shell__toolbar-side">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索模板标识、名称或描述"
            @keyup.enter="handleFilterChange"
            @clear="handleFilterChange"
          >
            <template #append>
              <ElButton @click="handleFilterChange">查询</ElButton>
            </template>
          </ElInput>
        </div>
      </header>

      <div v-loading="loading" class="message-template-board">
        <button
          v-for="item in list"
          :key="item.id"
          type="button"
          class="message-template-card"
          @click="openEditDrawer(item)"
        >
          <div class="message-template-card__head">
            <div>
              <h3>{{ item.name }}</h3>
              <p>{{ item.template_key }}</p>
            </div>
            <div class="message-template-card__tags">
              <ElTag size="small" effect="plain">{{ resolveScopeLabel(item) }}</ElTag>
              <ElTag size="small" effect="plain">{{
                resolveMessageTypeLabel(item.message_type)
              }}</ElTag>
              <ElTag
                size="small"
                :type="item.status === 'disabled' ? 'info' : 'success'"
                effect="plain"
              >
                {{ item.status === 'disabled' ? '停用' : '正常' }}
              </ElTag>
            </div>
          </div>

          <p class="message-template-card__desc">{{ item.description || '未填写模板说明' }}</p>

          <div class="message-template-card__meta">
            <span>{{ resolveAudienceLabel(item.audience_type) }}</span>
            <span>{{ item.editable ? '可编辑' : '只读' }}</span>
            <span>{{ formatTime(item.updated_at || item.created_at) }}</span>
          </div>

          <div class="message-template-card__preview">
            <div class="message-template-card__preview-title">{{
              item.title_template || '未填写标题模板'
            }}</div>
            <div class="message-template-card__preview-text">{{
              plainTextFromHtml(item.summary_template || item.content_template) ||
              '未填写摘要或正文模板'
            }}</div>
          </div>
        </button>

        <ElEmpty v-if="!loading && !list.length" description="当前范围下还没有消息模板" />
      </div>

      <footer class="message-template-shell__footer">
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
      :title="drawerTitle"
      size="620px"
      destroy-on-close
      append-to-body
      class="message-template-drawer"
    >
      <template v-if="drawerModel">
        <div class="message-template-drawer__summary">
          <div>
            <div class="message-template-drawer__summary-title">{{
              drawerModel.name || '未命名模板'
            }}</div>
            <div class="message-template-drawer__summary-text">
              {{ drawerReadOnly ? '当前模板为继承只读模板，只能查看不能修改。' : drawerScopeText }}
            </div>
          </div>
          <ElTag effect="plain">{{ drawerCollaborationWorkspaceBadge }}</ElTag>
        </div>

        <div class="message-template-drawer__form">
          <section class="message-template-drawer__block">
            <div class="message-template-drawer__block-header">
              <h4>基础信息</h4>
              <p>先确定模板标识、消息类型和发送对象，保证发送页能直接复用。</p>
            </div>

            <ElFormItem label="模板名称">
              <ElInput
                v-model="drawerModel.name"
                :disabled="drawerReadOnly"
                placeholder="例如：全局公告模板"
              />
            </ElFormItem>

            <ElFormItem label="模板标识">
              <ElInput
                v-model="drawerModel.template_key"
                :disabled="drawerReadOnly"
                placeholder="例如：announcement-default"
              />
              <div class="field-hint"
                >保存时会按作用域自动补全局或协作空间前缀，不需要手动写完整前缀。</div
              >
            </ElFormItem>

            <div class="message-template-drawer__grid">
              <ElFormItem label="消息类型">
                <ElRadioGroup v-model="drawerModel.message_type" :disabled="drawerReadOnly">
                  <ElRadioButton
                    v-for="item in messageTypeOptions"
                    :key="item.value"
                    :value="item.value"
                  >
                    {{ item.label }}
                  </ElRadioButton>
                </ElRadioGroup>
              </ElFormItem>

              <ElFormItem label="发送对象">
                <ElSelect v-model="drawerModel.audience_type" :disabled="drawerReadOnly">
                  <ElOption
                    v-for="item in availableAudienceOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </ElSelect>
              </ElFormItem>
            </div>

            <ElFormItem label="模板说明">
              <ElInput
                v-model="drawerModel.description"
                type="textarea"
                :disabled="drawerReadOnly"
                :rows="3"
                resize="vertical"
                placeholder="简要说明适用场景和推荐对象"
              />
            </ElFormItem>
          </section>

          <section class="message-template-drawer__block">
            <div class="message-template-drawer__block-header">
              <h4>内容模板</h4>
              <p>摘要保持纯文本，正文保留富文本，更符合消息公告和站内信后台的编辑习惯。</p>
            </div>

            <ElFormItem label="标题模板">
              <ElInput
                v-model="drawerModel.title_template"
                :disabled="drawerReadOnly"
                maxlength="120"
                show-word-limit
                placeholder="例如：全局维护通知"
              />
            </ElFormItem>

            <ElFormItem label="摘要模板">
              <ElInput
                v-model="drawerModel.summary_template"
                type="textarea"
                :rows="4"
                :disabled="drawerReadOnly"
                maxlength="200"
                show-word-limit
                resize="vertical"
                placeholder="顶部消息面板和列表摘要优先展示这里，建议一句话说清重点。"
              />
              <div class="field-hint"
                >摘要只保留纯文本，更适合消息中心列表、副标题和顶部预览展示。</div
              >
            </ElFormItem>

            <ElFormItem label="正文模板">
              <ArtWangEditor
                v-model="drawerModel.content_template"
                height="320px"
                :disabled="drawerReadOnly"
                placeholder="正文支持富文本和超链接。"
              />
            </ElFormItem>

            <ElFormItem label="状态">
              <ElRadioGroup v-model="drawerModel.status" :disabled="drawerReadOnly">
                <ElRadioButton value="normal">正常</ElRadioButton>
                <ElRadioButton value="disabled">停用</ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
          </section>
        </div>
      </template>

      <template #footer>
        <div class="message-template-drawer__footer">
          <ElButton @click="drawerVisible = false">关闭</ElButton>
          <ElButton v-if="!drawerReadOnly" type="primary" :loading="saving" @click="saveTemplate"
            >保存模板</ElButton
          >
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
  import ArtWangEditor from '@/components/core/forms/art-wang-editor/index.vue'
  import MessageWorkspaceNav from '@/views/message/modules/message-workspace-nav.vue'
  import {
    fetchCreateMessageTemplate,
    fetchGetMessageTemplateList,
    fetchUpdateMessageTemplate
  } from '@/api/message'
  import { useMessageWorkspace } from '@/views/message/modules/useMessageWorkspace'

  defineOptions({ name: 'MessageTemplateConsole' })

  const props = defineProps<{
    scope: 'global' | 'collaboration'
  }>()

  const loading = ref(false)
  const loadError = ref('')
  const saving = ref(false)
  const list = ref<Api.Message.MessageTemplateItem[]>([])
  const drawerVisible = ref(false)
  const drawerReadOnly = ref(false)
  const drawerEditingId = ref('')
  const filters = reactive({
    keyword: ''
  })
  const pagination = reactive({
    current: 1,
    size: 12,
    total: 0
  })

  interface TemplateDrawerModel {
    template_key: string
    name: string
    description: string
    message_type: Api.Message.BoxType
    audience_type: Api.Message.AudienceType
    title_template: string
    summary_template: string
    content_template: string
    status: 'normal' | 'disabled' | string
  }

  const drawerModel = ref<TemplateDrawerModel | null>(null)

  const {
    isCollaborationScope,
    skipAuthWorkspaceHeader,
    currentCollaborationName,
    currentWorkspaceName,
    currentWorkspaceLabel,
    ensureAuthWorkspaceContext,
    plainTextFromHtml,
    formatTime
  } = useMessageWorkspace(props.scope)

  const pageTitle = computed(() => (isCollaborationScope.value ? '协作空间消息模板' : '消息模板'))
  const pageDescription = computed(() =>
    isCollaborationScope.value
      ? '查看全局模板并维护当前协作空间可复用的消息模板，发送页会直接复用这里的标题、摘要和正文。'
      : '统一维护全局消息模板，供全局发信页复用；标题、摘要和正文都在模板层准备好。'
  )
  const toolbarDescription = computed(() =>
    isCollaborationScope.value
      ? `全局模板会以只读方式展示；当前 ${currentWorkspaceLabel.value} 下的协作空间视图 ${currentCollaborationName.value} 可维护自己的协作空间模板。`
      : '全局模板会直接出现在全局消息发送页中，建议用少量稳定模板覆盖高频场景。'
  )
  const heroMetrics = computed(() => [
    { label: '模板总数', value: pagination.total },
    {
      label: isCollaborationScope.value ? '协作空间模板' : '全局模板',
      value: list.value.filter(
        (item) => item.owner_scope === (isCollaborationScope.value ? 'collaboration' : 'global')
      ).length
    },
    { label: '可编辑', value: list.value.filter((item) => item.editable).length }
  ])

  const drawerTitle = computed(
    () =>
      `${drawerEditingId.value ? '编辑' : '新建'}${isCollaborationScope.value ? '协作空间' : '全局'}模板`
  )
  const drawerScopeText = computed(() =>
    isCollaborationScope.value
      ? `保存后该模板只属于 ${currentWorkspaceName.value} 下的协作空间视图 ${currentCollaborationName.value}，不会影响全局模板。`
      : '保存后该模板会作为全局模板在当前上下文内复用。'
  )
  const drawerCollaborationWorkspaceBadge = computed(() =>
    isCollaborationScope.value ? '协作空间模板' : '全局模板'
  )

  const messageTypeOptions = [
    { label: '通知', value: 'notice' },
    { label: '消息', value: 'message' },
    { label: '待办', value: 'todo' }
  ]

  const availableAudienceOptions = computed(() =>
    isCollaborationScope.value
      ? [
          {
            label: '当前协作空间成员',
            value: 'collaboration_users' as Api.Message.AudienceType
          }
        ]
      : [
          { label: '所有用户', value: 'all_users' as Api.Message.AudienceType },
          {
            label: '协作空间管理员',
            value: 'collaboration_admins' as Api.Message.AudienceType
          },
          {
            label: '指定协作空间成员',
            value: 'collaboration_users' as Api.Message.AudienceType
          }
        ]
  )

  const createDefaultModel = (): TemplateDrawerModel => ({
    template_key: '',
    name: '',
    description: '',
    message_type: 'notice',
    audience_type: isCollaborationScope.value ? 'collaboration_users' : 'all_users',
    title_template: '',
    summary_template: '',
    content_template: '',
    status: 'normal'
  })

  const resolveScopeLabel = (item: Api.Message.MessageTemplateItem) => {
    if (item.owner_scope === 'collaboration') {
      const workspaceName = `${(item as unknown as Record<string, unknown>).owner_workspace_name || ''}`.trim()
      return workspaceName
        ? `协作空间 · ${workspaceName}`
        : '协作空间模板'
    }
    return '全局模板'
  }

  const resolveMessageTypeLabel = (value: Api.Message.BoxType) =>
    messageTypeOptions.find((item) => item.value === value)?.label || '通知'

  const resolveAudienceLabel = (value: Api.Message.AudienceType) => {
    if (value === 'all_users') return '所有用户'
    if (value === 'collaboration_admins') return '协作空间管理员'
    return '协作空间成员'
  }

  const loadTemplates = async () => {
    loading.value = true
    loadError.value = ''
    try {
      ensureAuthWorkspaceContext()
      const result = await fetchGetMessageTemplateList(
        {
          keyword: filters.keyword || undefined,
          current: pagination.current,
          size: pagination.size
        },
        { skipAuthWorkspaceHeader: skipAuthWorkspaceHeader.value }
      )
      list.value = result.records || []
      pagination.total = result.total || 0
    } catch {
      list.value = []
      pagination.total = 0
      loadError.value = '消息模板暂时不可用，稍后重试或刷新状态。'
    } finally {
      loading.value = false
    }
  }

  const handleFilterChange = async () => {
    pagination.current = 1
    await loadTemplates()
  }

  const openCreateDrawer = () => {
    drawerEditingId.value = ''
    drawerReadOnly.value = false
    drawerModel.value = createDefaultModel()
    drawerVisible.value = true
  }

  const openEditDrawer = (item: Api.Message.MessageTemplateItem) => {
    drawerEditingId.value = item.id
    drawerReadOnly.value = !item.editable
    drawerModel.value = {
      template_key: item.template_key
        .replace(/^platform\./, '')
        .replace(/^collaboration_workspace\.[^.]+\./, '')
        .replace(/^collaboration\.[^.]+\./, ''),
      name: item.name,
      description: item.description || '',
      message_type: item.message_type,
      audience_type: item.audience_type,
      title_template: item.title_template || '',
      summary_template: item.summary_template || '',
      content_template: item.content_template || '',
      status: item.status
    }
    drawerVisible.value = true
  }

  const saveTemplate = async () => {
    if (!drawerModel.value) return
    if (!drawerModel.value.name.trim()) {
      ElMessage.warning('请先填写模板名称')
      return
    }
    saving.value = true
    try {
      if (drawerEditingId.value) {
        await fetchUpdateMessageTemplate(drawerEditingId.value, drawerModel.value, {
          skipAuthWorkspaceHeader: skipAuthWorkspaceHeader.value
        })
      } else {
        await fetchCreateMessageTemplate(drawerModel.value, {
          skipAuthWorkspaceHeader: skipAuthWorkspaceHeader.value
        })
      }
      drawerVisible.value = false
      await loadTemplates()
    } catch {
      ElMessage.error('保存消息模板失败')
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadTemplates()
  })

  watch(
    () => [pagination.current, pagination.size],
    ([current, size], [oldCurrent, oldSize]) => {
      if (current === oldCurrent && size === oldSize) return
      loadTemplates()
    }
  )
</script>

<style scoped lang="scss">
  .message-template-page {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .message-template-inline-alert {
    margin-top: 0;
  }

  .message-template-hero__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .message-template-shell {
    padding: 18px;
    border-radius: 24px;
  }

  .message-template-shell__toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding-bottom: 14px;
    border-bottom: 1px solid rgb(226 232 240 / 0.85);
  }

  .message-template-shell__toolbar-main {
    display: grid;
    gap: 4px;
  }

  .message-template-shell__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-template-shell__toolbar p {
    margin: 0;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-template-shell__toolbar-side {
    width: min(360px, 100%);
  }

  .message-template-board {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
    padding-top: 16px;
  }

  .message-template-card {
    display: grid;
    gap: 12px;
    width: 100%;
    padding: 16px;
    border: 1px solid rgb(226 232 240 / 0.92);
    border-radius: 20px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.96));
    text-align: left;
    transition:
      transform 0.18s ease,
      border-color 0.18s ease,
      box-shadow 0.18s ease;
  }

  .message-template-card:hover {
    border-color: rgb(59 130 246 / 0.3);
    box-shadow: 0 16px 30px rgb(15 23 42 / 0.08);
    transform: translateY(-1px);
  }

  .message-template-card__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .message-template-card__head h3 {
    margin: 0;
    font-size: 16px;
    color: #0f172a;
  }

  .message-template-card__head p {
    margin: 6px 0 0;
    font-size: 11px;
    color: #94a3b8;
  }

  .message-template-card__tags {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  .message-template-card__desc {
    margin: 0;
    font-size: 12px;
    line-height: 1.7;
    color: #64748b;
  }

  .message-template-card__meta {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    font-size: 11px;
    color: #475569;
  }

  .message-template-card__preview {
    padding: 14px;
    border-radius: 16px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-template-card__preview-title {
    font-size: 13px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-template-card__preview-text {
    margin-top: 8px;
    font-size: 12px;
    line-height: 1.7;
    color: #64748b;
    display: -webkit-box;
    overflow: hidden;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
  }

  .message-template-shell__footer {
    display: flex;
    justify-content: flex-end;
    padding-top: 16px;
  }

  .message-template-drawer__summary {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.96), rgb(255 255 255 / 0.98));
  }

  .message-template-drawer__summary-title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-template-drawer__summary-text,
  .field-hint {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-template-drawer__form {
    display: grid;
    gap: 14px;
    margin-top: 16px;
  }

  .message-template-drawer__block {
    display: grid;
    gap: 14px;
    padding: 18px;
    border: 1px solid rgb(226 232 240 / 0.88);
    border-radius: 20px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.92));
  }

  .message-template-drawer__block-header {
    display: grid;
    gap: 4px;
  }

  .message-template-drawer__block-header h4 {
    margin: 0;
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-template-drawer__block-header p {
    margin: 0;
    font-size: 12px;
    line-height: 1.7;
    color: #64748b;
  }

  .message-template-drawer__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 14px;
  }

  .message-template-drawer__footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    width: 100%;
  }

  :deep(.message-template-drawer .el-form-item) {
    margin-bottom: 0;
  }

  @media (max-width: 1120px) {
    .message-template-board {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .message-template-shell__toolbar {
      flex-direction: column;
      align-items: stretch;
    }

    .message-template-drawer__grid {
      grid-template-columns: 1fr;
    }
  }
</style>


