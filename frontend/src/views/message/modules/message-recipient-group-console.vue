<template>
  <div class="message-group-page art-full-height">
    <AdminWorkspaceHero :title="pageTitle" :description="pageDescription" :metrics="heroMetrics">
      <div class="message-group-hero__actions">
        <ElButton @click="loadGroups" :loading="loading" v-ripple>刷新</ElButton>
        <ElButton type="primary" @click="openCreateDrawer" v-ripple>新建接收组</ElButton>
      </div>
    </AdminWorkspaceHero>

    <MessageWorkspaceNav :scope="props.scope" current="group" />

    <ElAlert
      v-if="loadError"
      class="message-group-inline-alert"
      type="info"
      :closable="false"
      show-icon
      :title="loadError"
    />

    <section class="message-group-shell art-card">
      <header class="message-group-shell__toolbar">
        <div>
          <div class="message-group-shell__title">接收组列表</div>
          <p>{{ toolbarDescription }}</p>
        </div>
      </header>

      <div v-loading="loading" class="message-group-board">
        <button
          v-for="item in pagedList"
          :key="item.id"
          type="button"
          class="message-group-card"
          @click="openEditDrawer(item)"
        >
          <div class="message-group-card__head">
            <div>
              <h3>{{ item.name }}</h3>
              <p>{{ item.description || '未填写接收组说明' }}</p>
            </div>
            <div class="message-group-card__tags">
              <ElTag effect="plain" size="small">{{
                item.match_mode === 'manual' ? '手动组' : item.match_mode
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

          <div class="message-group-card__metrics">
            <div>
              <span>预估人数</span>
              <strong>{{ item.estimated_count || 0 }}</strong>
            </div>
            <div>
              <span>规则条目</span>
              <strong>{{ item.targets?.length || 0 }}</strong>
            </div>
          </div>

          <div class="message-group-card__targets">
            <div class="message-group-card__label">当前接收范围</div>
            <p>{{ summarizeTargets(item.targets) }}</p>
          </div>

          <div class="message-group-card__meta">
            <span>{{ item.scope_type === 'collaboration' ? '协作空间接收组' : '平台接收组' }}</span>
            <span>{{ formatTime(item.updated_at || item.created_at) }}</span>
          </div>
        </button>

        <ElEmpty v-if="!loading && !list.length" description="当前还没有可用接收组" />
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
      :title="drawerEditingId ? '编辑接收组' : '新建接收组'"
      size="720px"
      destroy-on-close
      append-to-body
    >
      <template v-if="drawerModel">
        <div class="message-group-drawer__summary">
          <div>
            <div class="message-group-drawer__title">{{ drawerModel.name || '未命名接收组' }}</div>
            <div class="message-group-drawer__text">{{ drawerSummary }}</div>
          </div>
          <div class="message-group-drawer__summary-tags">
            <ElTag effect="plain">{{
              isCollaborationScope ? '协作空间接收组' : '平台接收组'
            }}</ElTag>
            <ElTag effect="plain" type="info">预估 {{ estimatedRecipients }} 人</ElTag>
          </div>
        </div>

        <div class="message-group-drawer__form">
          <div class="message-group-drawer__grid">
            <ElFormItem label="接收组名称">
              <ElInput
                v-model="drawerModel.name"
                maxlength="60"
                show-word-limit
                placeholder="例如：平台重点关注用户 / 当前协作空间管理员"
              />
            </ElFormItem>

            <ElFormItem label="状态">
              <ElRadioGroup v-model="drawerModel.status">
                <ElRadioButton value="normal">正常</ElRadioButton>
                <ElRadioButton value="disabled">停用</ElRadioButton>
              </ElRadioGroup>
            </ElFormItem>
          </div>

          <ElFormItem label="接收组说明">
            <ElInput
              v-model="drawerModel.description"
              type="textarea"
              :rows="3"
              placeholder="说明这个接收组服务什么场景，方便发送页直接选择。"
            />
          </ElFormItem>

          <div class="message-group-drawer__rules">
            <header class="message-group-drawer__rules-header">
              <div>
                <div class="message-group-drawer__rules-title">接收规则</div>
                <p
                  >当前支持指定用户、协作空间成员、协作空间管理员、角色、功能包五类规则。后续标签和组合条件仍然继续收口到接收组。</p
                >
              </div>
              <ElButton @click="appendTarget" v-ripple>新增规则</ElButton>
            </header>

            <div v-if="drawerModel.targets.length" class="message-group-rule-list">
              <div
                v-for="(item, index) in drawerModel.targets"
                :key="item.local_id"
                class="message-group-rule-item"
              >
                <div class="message-group-rule-item__toolbar">
                  <div class="message-group-rule-item__index">规则 {{ index + 1 }}</div>
                  <ElButton link type="danger" @click="removeTarget(index)">删除</ElButton>
                </div>

                <div class="message-group-rule-item__grid">
                  <ElFormItem label="规则类型">
                    <ElSelect v-model="item.target_type" @change="handleTargetTypeChange(item)">
                      <ElOption value="user" label="指定用户" />
                      <ElOption
                        value="collaboration_workspace_users"
                        :label="isCollaborationScope ? '当前协作空间成员' : '指定协作空间成员'"
                      />
                      <ElOption
                        value="collaboration_workspace_admins"
                        :label="isCollaborationScope ? '当前协作空间管理员' : '指定协作空间管理员'"
                      />
                      <ElOption value="role" label="按角色命中" />
                      <ElOption value="feature_package" label="按功能包命中" />
                    </ElSelect>
                  </ElFormItem>

                  <ElFormItem label="排序">
                    <ElInputNumber
                      v-model="item.sort_order"
                      :min="1"
                      :max="999"
                      controls-position="right"
                    />
                  </ElFormItem>
                </div>

                <ElFormItem v-if="item.target_type === 'user'" label="接收用户">
                  <ElSelect v-model="item.user_id" filterable placeholder="请选择用户">
                    <ElOption
                      v-for="user in userOptions"
                      :key="user.id"
                      :label="
                        user.collaboration_workspace_name ||
                        user.current_collaboration_workspace_name
                          ? `${user.display_name} · ${user.collaboration_workspace_name || user.current_collaboration_workspace_name}`
                          : user.display_name
                      "
                      :value="user.id"
                    />
                  </ElSelect>
                </ElFormItem>

                <ElFormItem
                  v-else-if="
                    item.target_type === 'collaboration_workspace_users' ||
                    item.target_type === 'collaboration_workspace_admins'
                  "
                  label="目标协作空间"
                >
                  <div v-if="isCollaborationScope" class="message-group-fixed-target">
                    <strong>{{ currentCollaborationWorkspaceName }}</strong>
                    <span>协作空间侧规则固定作用于当前协作空间。</span>
                  </div>
                  <ElSelect
                    v-else
                    v-model="item.collaborationWorkspaceId"
                    filterable
                    placeholder="请选择协作空间"
                  >
                    <ElOption
                      v-for="workspace in collaborationWorkspaceOptions"
                      :key="workspace.id"
                      :label="workspace.name"
                      :value="workspace.id"
                    />
                  </ElSelect>
                </ElFormItem>

                <ElFormItem v-else-if="item.target_type === 'role'" label="角色">
                  <ElSelect v-model="item.role_code" filterable placeholder="请选择角色">
                    <ElOption
                      v-for="role in roleOptions"
                      :key="role.code"
                      :label="role.name"
                      :value="role.code"
                    />
                  </ElSelect>
                </ElFormItem>

                <ElFormItem v-else-if="item.target_type === 'feature_package'" label="功能包">
                  <ElSelect v-model="item.package_key" filterable placeholder="请选择功能包">
                    <ElOption
                      v-for="pkg in featurePackageOptions"
                      :key="pkg.package_key"
                      :label="pkg.name"
                      :value="pkg.package_key"
                    />
                  </ElSelect>
                </ElFormItem>
              </div>
            </div>

            <ElEmpty v-else description="还没有规则，先添加一条接收规则。" />
          </div>

          <div class="message-group-drawer__reserved">
            <div class="message-group-drawer__rules-title">预留能力</div>
            <p
              >下一阶段继续在接收组里扩标签、组合条件和更复杂的命中规则，发送页只负责选择对象，不再承担规则编辑。</p
            >
          </div>
        </div>
      </template>

      <template #footer>
        <div class="message-group-drawer__footer">
          <ElButton @click="drawerVisible = false">关闭</ElButton>
          <ElButton type="primary" :loading="saving" @click="saveGroup">保存接收组</ElButton>
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
    fetchCreateMessageRecipientGroup,
    fetchGetMessageDispatchOptions,
    fetchGetMessageRecipientGroupList,
    fetchUpdateMessageRecipientGroup
  } from '@/api/message'
  import { useMessageWorkspace } from '@/views/message/modules/useMessageWorkspace'

  defineOptions({ name: 'MessageRecipientGroupConsole' })

  const props = defineProps<{
    scope: 'platform' | 'collaboration'
  }>()

  interface DrawerTargetModel {
    local_id: string
    target_type:
      | 'user'
      | 'collaboration_workspace_users'
      | 'collaboration_workspace_admins'
      | 'role'
      | 'feature_package'
      | string
    user_id: string
    collaborationWorkspaceId: string
    role_code: string
    package_key: string
    sort_order: number
  }

  interface DrawerGroupModel {
    name: string
    description: string
    status: 'normal' | 'disabled' | string
    targets: DrawerTargetModel[]
  }

  const {
    isCollaborationScope,
    skipCollaborationWorkspaceHeader,
    currentCollaborationWorkspaceId,
    currentCollaborationWorkspaceName,
    currentWorkspaceName,
    currentWorkspaceLabel,
    ensureCollaborationWorkspaceContext,
    formatTime
  } = useMessageWorkspace(props.scope)

  const loading = ref(false)
  const loadError = ref('')
  const saving = ref(false)
  const sequence = ref(1)
  const list = ref<Api.Message.MessageRecipientGroupItem[]>([])
  const pagination = reactive({
    current: 1,
    size: 8
  })
  const drawerVisible = ref(false)
  const drawerEditingId = ref('')
  const drawerModel = ref<DrawerGroupModel | null>(null)
  const userOptions = ref<Api.Message.DispatchUserOption[]>([])
  const collaborationWorkspaceOptions = ref<Api.Message.DispatchCollaborationWorkspaceOption[]>([])
  const roleOptions = ref<Api.Message.DispatchRoleOption[]>([])
  const featurePackageOptions = ref<Api.Message.DispatchFeaturePackageOption[]>([])

  const pageTitle = computed(() => (isCollaborationScope.value ? '协作空间接收组' : '接收组管理'))
  const pageDescription = computed(() =>
    isCollaborationScope.value
      ? `维护 ${currentWorkspaceName.value} 下 ${currentCollaborationWorkspaceName.value} 的消息接收组，用于协作空间管理员快速向固定成员组合发送消息。`
      : '维护平台消息接收组，把指定用户、指定协作空间成员和指定协作空间管理员收口到统一的发送对象配置里。'
  )
  const toolbarDescription = computed(() =>
    isCollaborationScope.value
      ? `协作空间接收组只作用于当前协作空间消息发送页（${currentWorkspaceLabel.value}）。`
      : '平台接收组可给平台发信台直接复用，也为后续角色、功能包等条件匹配预留统一扩展位。'
  )
  const heroMetrics = computed(() => [
    { label: '接收组总数', value: list.value.length },
    { label: '正常状态', value: list.value.filter((item) => item.status === 'normal').length },
    {
      label: '手动规则',
      value: list.value.reduce((sum, item) => sum + (item.targets?.length || 0), 0)
    }
  ])
  const drawerSummary = computed(() =>
    isCollaborationScope.value
      ? `保存后会作为 ${currentCollaborationWorkspaceName.value}（${currentWorkspaceName.value}）的可选接收组。`
      : '保存后会作为平台消息发送页的可选接收组。'
  )
  const estimatedRecipients = computed(() =>
    estimateDrawerRecipients(drawerModel.value?.targets || [])
  )

  const nextLocalId = () => {
    const value = sequence.value
    sequence.value += 1
    return `target-${value}`
  }

  const createTarget = (
    target?: Partial<Api.Message.MessageRecipientGroupTargetItem>
  ): DrawerTargetModel => ({
    local_id: nextLocalId(),
    target_type: target?.target_type || 'user',
    user_id: target?.user_id || '',
    collaborationWorkspaceId:
      target?.collaboration_workspace_id ||
      target?.collaboration_workspace_id ||
      (isCollaborationScope.value ? currentCollaborationWorkspaceId.value || '' : ''),
    role_code: target?.role_code || '',
    package_key: target?.package_key || '',
    sort_order: target?.sort_order || sequence.value
  })

  const createDefaultModel = (): DrawerGroupModel => ({
    name: '',
    description: '',
    status: 'normal',
    targets: [createTarget()]
  })

  const resolveTargetCollaborationWorkspaceId = (
    item?: Api.Message.MessageRecipientGroupTargetItem | DrawerTargetModel
  ) => {
    if (!item) return ''
    if ('collaborationWorkspaceId' in item) return item.collaborationWorkspaceId || ''
    return item.collaboration_workspace_id || item.collaboration_workspace_id || ''
  }

  const summarizeTarget = (
    item?: Api.Message.MessageRecipientGroupTargetItem | DrawerTargetModel
  ) => {
    if (!item) return '未配置'
    if (item.target_type === 'user') {
      const name =
        'user_name' in item && item.user_name
          ? item.user_name
          : userOptions.value.find((user) => user.id === item.user_id)?.display_name
      return name ? `指定用户 · ${name}` : '指定用户'
    }
    if (item.target_type === 'role') {
      const name =
        'role_name' in item && item.role_name
          ? item.role_name
          : roleOptions.value.find((role) => role.code === item.role_code)?.name
      return name ? `角色规则 · ${name}` : '角色规则'
    }
    if (item.target_type === 'feature_package') {
      const name =
        'package_name' in item && item.package_name
          ? item.package_name
          : featurePackageOptions.value.find((pkg) => pkg.package_key === item.package_key)?.name
      return name ? `功能包规则 · ${name}` : '功能包规则'
    }
    const collaborationWorkspaceName =
      'collaboration_workspace_name' in item && item.collaboration_workspace_name
        ? item.collaboration_workspace_name
        : collaborationWorkspaceOptions.value.find(
            (workspace) => workspace.id === resolveTargetCollaborationWorkspaceId(item)
          )?.name || currentCollaborationWorkspaceName.value
    if (item.target_type === 'collaboration_workspace_admins') {
      return `${collaborationWorkspaceName} · 协作空间管理员`
    }
    return `${collaborationWorkspaceName} · 协作空间成员`
  }

  const summarizeTargets = (targets?: Api.Message.MessageRecipientGroupTargetItem[]) => {
    if (!targets?.length) return '尚未配置接收规则'
    return targets
      .slice(0, 3)
      .map((item) => summarizeTarget(item))
      .join('、')
  }

  const estimateDrawerRecipients = (targets: DrawerTargetModel[]) => {
    if (!targets.length) return 0
    const seen = new Set<string>()
    targets.forEach((item) => {
      if (item.target_type === 'user' && item.user_id) {
        seen.add(item.user_id)
        return
      }
      if (item.target_type === 'role' && item.role_code) {
        seen.add(`role:${item.role_code}`)
        return
      }
      if (item.target_type === 'feature_package' && item.package_key) {
        seen.add(`package:${item.package_key}`)
        return
      }
      if (!isCollaborationScope.value) return
      if (item.target_type === 'collaboration_workspace_admins') {
        userOptions.value
          .filter(
            (user) =>
              (user.collaboration_workspace_name || user.current_collaboration_workspace_name) ===
              currentCollaborationWorkspaceName.value
          )
          .forEach((user) => seen.add(user.id))
        return
      }
      if (item.target_type === 'collaboration_workspace_users') {
        userOptions.value.forEach((user) => seen.add(user.id))
      }
    })
    return seen.size || targets.length
  }

  const loadDispatchHelpers = async () => {
    const data = await fetchGetMessageDispatchOptions({
      skipCollaborationWorkspaceHeader: skipCollaborationWorkspaceHeader.value
    })
    userOptions.value = data.users || []
    collaborationWorkspaceOptions.value = data.collaboration_workspaces || data.teams || []
    roleOptions.value = data.roles || []
    featurePackageOptions.value = data.feature_packages || []
  }

  const loadGroups = async () => {
    loading.value = true
    loadError.value = ''
    try {
      ensureCollaborationWorkspaceContext()
      await loadDispatchHelpers()
      const result = await fetchGetMessageRecipientGroupList({
        skipCollaborationWorkspaceHeader: skipCollaborationWorkspaceHeader.value
      })
      list.value = result.records || []
      pagination.current = 1
    } catch {
      list.value = []
      loadError.value = '接收组暂时不可用，稍后重试或刷新状态。'
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

  const openEditDrawer = (item: Api.Message.MessageRecipientGroupItem) => {
    drawerEditingId.value = item.id
    drawerModel.value = {
      name: item.name,
      description: item.description || '',
      status: item.status,
      targets: (item.targets || []).map((target) => createTarget(target))
    }
    if (!drawerModel.value.targets.length) {
      drawerModel.value.targets = [createTarget()]
    }
    drawerVisible.value = true
  }

  const appendTarget = () => {
    if (!drawerModel.value) return
    drawerModel.value.targets.push(createTarget())
  }

  const removeTarget = (index: number) => {
    if (!drawerModel.value) return
    drawerModel.value.targets.splice(index, 1)
    if (!drawerModel.value.targets.length) {
      drawerModel.value.targets.push(createTarget())
    }
  }

  const handleTargetTypeChange = (item: DrawerTargetModel) => {
    item.user_id = ''
    item.collaborationWorkspaceId = isCollaborationScope.value
      ? currentCollaborationWorkspaceId.value || ''
      : ''
    item.role_code = ''
    item.package_key = ''
  }

  const saveGroup = async () => {
    if (!drawerModel.value) return
    if (!drawerModel.value.name.trim()) {
      ElMessage.warning('请先填写接收组名称')
      return
    }
    if (!drawerModel.value.targets.length) {
      ElMessage.warning('请至少配置一条接收规则')
      return
    }
    for (const item of drawerModel.value.targets) {
      if (item.target_type === 'user' && !item.user_id) {
        ElMessage.warning('指定用户规则必须选择接收用户')
        return
      }
      if (
        !isCollaborationScope.value &&
        (item.target_type === 'collaboration_workspace_users' ||
          item.target_type === 'collaboration_workspace_admins') &&
        !item.collaborationWorkspaceId
      ) {
        ElMessage.warning('协作空间规则必须选择目标协作空间')
        return
      }
      if (item.target_type === 'role' && !item.role_code) {
        ElMessage.warning('角色规则必须选择角色')
        return
      }
      if (item.target_type === 'feature_package' && !item.package_key) {
        ElMessage.warning('功能包规则必须选择功能包')
        return
      }
    }
    saving.value = true
    try {
      const payload: Api.Message.MessageRecipientGroupSaveParams = {
        name: drawerModel.value.name.trim(),
        description: drawerModel.value.description.trim(),
        match_mode: 'manual',
        status: drawerModel.value.status,
        targets: drawerModel.value.targets.map((item, index) => ({
          target_type: item.target_type,
          user_id: item.target_type === 'user' ? item.user_id || undefined : undefined,
          collaboration_workspace_id:
            item.target_type === 'collaboration_workspace_users' ||
            item.target_type === 'collaboration_workspace_admins'
              ? isCollaborationScope.value
                ? currentCollaborationWorkspaceId.value || undefined
                : item.collaborationWorkspaceId || undefined
              : undefined,
          role_code: item.target_type === 'role' ? item.role_code || undefined : undefined,
          package_key:
            item.target_type === 'feature_package' ? item.package_key || undefined : undefined,
          sort_order: item.sort_order || index + 1
        }))
      }
      if (drawerEditingId.value) {
        await fetchUpdateMessageRecipientGroup(drawerEditingId.value, payload, {
          skipCollaborationWorkspaceHeader: skipCollaborationWorkspaceHeader.value
        })
      } else {
        await fetchCreateMessageRecipientGroup(payload, {
          skipCollaborationWorkspaceHeader: skipCollaborationWorkspaceHeader.value
        })
      }
      drawerVisible.value = false
      await loadGroups()
    } catch {
      ElMessage.error('保存接收组失败')
    } finally {
      saving.value = false
    }
  }

  onMounted(() => {
    loadGroups()
  })

  watch(
    () => pagination.size,
    () => {
      pagination.current = 1
    }
  )
</script>

<style scoped lang="scss">
  .message-group-page {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .message-group-inline-alert {
    margin-top: 0;
  }

  .message-group-hero__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .message-group-shell {
    padding: 18px;
    border-radius: 24px;
  }

  .message-group-shell__toolbar {
    padding-bottom: 14px;
    border-bottom: 1px solid rgb(226 232 240 / 0.85);
  }

  .message-group-shell__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-group-shell__toolbar p,
  .message-group-drawer__text {
    margin: 6px 0 0;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-group-board {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
    padding-top: 16px;
  }

  .message-group-card {
    display: grid;
    gap: 14px;
    width: 100%;
    padding: 16px;
    border: 1px solid rgb(226 232 240 / 0.92);
    border-radius: 20px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.96));
    text-align: left;
  }

  .message-group-card__head,
  .message-group-drawer__summary,
  .message-group-drawer__rules-header,
  .message-group-rule-item__toolbar {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }

  .message-group-card__head h3,
  .message-group-drawer__title {
    margin: 0;
    font-size: 16px;
    color: #0f172a;
  }

  .message-group-card__head p,
  .message-group-card__targets p,
  .message-group-drawer__rules-header p,
  .message-group-drawer__reserved p,
  .message-group-fixed-target span {
    margin: 6px 0 0;
    font-size: 12px;
    line-height: 1.6;
    color: #64748b;
  }

  .message-group-card__tags,
  .message-group-card__meta,
  .message-group-drawer__footer,
  .message-group-drawer__summary-tags {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .message-group-card__metrics {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .message-group-card__metrics div,
  .message-group-drawer__reserved,
  .message-group-fixed-target {
    padding: 14px;
    border-radius: 16px;
    background: rgb(248 250 252 / 0.92);
  }

  .message-group-card__metrics span,
  .message-group-card__label {
    display: block;
    font-size: 11px;
    color: #94a3b8;
  }

  .message-group-card__metrics strong {
    display: block;
    margin-top: 6px;
    font-size: 20px;
    color: #0f172a;
  }

  .message-group-card__meta {
    font-size: 11px;
    color: #475569;
  }

  .message-group-drawer__summary {
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: linear-gradient(180deg, rgb(248 250 252 / 0.96), rgb(255 255 255 / 0.98));
  }

  .message-group-drawer__form {
    display: grid;
    gap: 16px;
    margin-top: 16px;
  }

  .message-group-drawer__grid,
  .message-group-rule-item__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0 14px;
  }

  .message-group-drawer__rules,
  .message-group-rule-item {
    display: grid;
    gap: 14px;
  }

  .message-group-drawer__rules-title,
  .message-group-rule-item__index {
    font-size: 14px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-group-rule-list {
    display: grid;
    gap: 12px;
  }

  .message-group-rule-item {
    padding: 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.95));
  }

  .message-group-drawer__footer {
    justify-content: flex-end;
    width: 100%;
  }

  @media (max-width: 1080px) {
    .message-group-board {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 768px) {
    .message-group-drawer__grid,
    .message-group-rule-item__grid {
      grid-template-columns: 1fr;
    }
  }
</style>
