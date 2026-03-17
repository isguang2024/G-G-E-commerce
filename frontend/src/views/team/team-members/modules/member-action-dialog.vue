<template>
  <ElDialog
    v-model="visible"
    :title="`成员功能权限 - ${member?.userName || ''}`"
    width="920px"
    destroy-on-close
  >
    <div v-loading="loading" class="member-action-dialog">
      <ElAlert
        type="info"
        :closable="false"
        class="dialog-alert"
        title="这里配置的是团队内个人功能权限覆盖。默认继承团队角色权限；单独允许或单独拒绝只用于少量例外场景。"
      />

      <section class="control-panel">
        <div class="toolbar">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索权限名称/权限键/模块归属/分类"
            class="toolbar-input"
          />
          <ElSelect v-model="filters.featureKind" clearable placeholder="功能归属" class="toolbar-select">
            <ElOption label="全部归属" value="" />
            <ElOption label="系统功能" value="system" />
            <ElOption label="业务功能" value="business" />
          </ElSelect>
          <ElSelect v-model="filters.overrideState" clearable placeholder="覆盖状态" class="toolbar-select">
            <ElOption label="全部状态" value="" />
            <ElOption label="仅继承角色" value="inherit" />
            <ElOption label="仅单独允许" value="allow" />
            <ElOption label="仅单独拒绝" value="deny" />
            <ElOption label="仅已覆盖" value="overridden" />
          </ElSelect>
        </div>

        <div class="option-row">
          <div class="summary">
            <ElTag effect="plain" class="summary-tag">可见 {{ filteredActionCount }} / {{ actions.length }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--baseline">角色允许 {{ roleAllowCount }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--allow">单独允许 {{ overrideAllowCount }}</ElTag>
            <ElTag effect="plain" class="summary-tag summary-tag--deny">单独拒绝 {{ overrideDenyCount }}</ElTag>
          </div>

          <div class="option-switches">
            <label class="option-item">
              <span>仅看例外</span>
              <ElSwitch v-model="filters.onlyOverrides" />
            </label>
            <label class="option-item">
              <span>查看备注</span>
              <ElSwitch v-model="filters.showRemark" />
            </label>
            <label class="option-item">
              <span>紧凑模式</span>
              <ElSwitch v-model="filters.compact" />
            </label>
          </div>
        </div>

        <div class="batch-bar" v-if="filteredActionIds.length > 0">
          <span class="batch-label">批量处理当前筛选结果</span>
          <ElButton size="small" text @click="expandAll">全部展开</ElButton>
          <ElButton size="small" text @click="collapseAll">全部收起</ElButton>
          <ElButton size="small" text @click="applyEffects(filteredActionIds, 'allow')">批量单独允许</ElButton>
          <ElButton size="small" text @click="applyEffects(filteredActionIds, 'deny')">批量单独拒绝</ElButton>
          <ElButton size="small" text @click="applyEffects(filteredActionIds, '')">批量继承角色</ElButton>
        </div>
      </section>

      <ElEmpty v-if="!loading && filteredGroups.length === 0" description="当前团队未开通功能权限" />

      <div v-else class="tree-wrapper">
        <ElTree
          ref="treeRef"
          :data="treeData"
          node-key="key"
          :props="treeProps"
          :default-expanded-keys="expandedKeys"
          :expand-on-click-node="true"
          :highlight-current="false"
          class="permission-tree"
        >
          <template #default="{ data }">
            <div v-if="data.nodeType === 'feature'" class="tree-node tree-node--feature" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions">
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'allow')">本组单独允许</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'deny')">本组单独拒绝</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, '')">本组继承角色</ElButton>
              </div>
            </div>

            <div v-else-if="data.nodeType === 'module'" class="tree-node tree-node--module" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions">
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'allow')">模块单独允许</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, 'deny')">模块单独拒绝</ElButton>
                <ElButton size="small" text @click.stop="applyEffects(data.actionIds, '')">模块继承角色</ElButton>
              </div>
            </div>

            <div v-else class="tree-node tree-node--action" :class="{ 'is-compact': filters.compact }">
              <div class="node-main">
                <div class="node-title">{{ data.label }}</div>
                <div v-if="data.meta" class="node-subtitle">{{ data.meta }}</div>
              </div>
              <div class="node-actions node-actions--leaf">
                <ElTag size="small" effect="plain" class="muted-tag">{{ data.scopeText }}</ElTag>
                <ElTag size="small" effect="plain" class="muted-tag">{{ data.roleEffectText }}</ElTag>
                <ElSelect
                  v-model="effectMap[data.actionId]"
                  class="effect-select"
                  :class="{
                    'effect-select--allow': effectMap[data.actionId] === 'allow',
                    'effect-select--deny': effectMap[data.actionId] === 'deny',
                    'effect-select--empty': !effectMap[data.actionId]
                  }"
                  size="small"
                  placeholder="继承角色"
                >
                  <ElOption label="继承角色" value="" />
                  <ElOption label="单独允许" value="allow" />
                  <ElOption label="单独拒绝" value="deny" />
                </ElSelect>
              </div>
            </div>
          </template>
        </ElTree>
      </div>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import {
    buildAllExpandedKeys,
    buildDefaultExpandedKeys,
    buildPermissionGroups,
    buildPermissionTree,
    type FeatureGroup
  } from '@/components/business/permission/permission-tree'
  import {
    fetchGetMyTeamActions,
    fetchGetMyTeamMemberActions,
    fetchGetMyTeamMemberRoles,
    fetchGetMyTeamRoleActions,
    fetchGetMyTeamRoles,
    fetchSetMyTeamMemberActions
  } from '@/api/team'
  import { formatScopeLabel } from '@/utils/permission/scope'
  import { ElMessage } from 'element-plus'

  interface Props {
    member: Api.SystemManage.TeamMemberItem | null
  }

  type EffectValue = 'allow' | 'deny' | ''
  type PermissionActionItem = Api.SystemManage.PermissionActionItem
  interface PermissionTreeNode {
    key: string
    label: string
    nodeType: 'feature' | 'module' | 'action'
    meta: string
    children?: PermissionTreeNode[]
    actionIds: string[]
    actionId?: string
    scopeText?: string
    roleEffectText?: string
  }

  const props = defineProps<Props>()
  const emit = defineEmits(['success'])

  const visible = ref(false)
  const loading = ref(false)
  const submitting = ref(false)
  const actions = ref<PermissionActionItem[]>([])
  const treeRef = ref()
  const expandedKeys = ref<string[]>([])
  const effectMap = reactive<Record<string, EffectValue>>({})
  const roleEffectMap = reactive<Record<string, EffectValue>>({})
  const filters = reactive({
    keyword: '',
    featureKind: '',
    overrideState: '',
    onlyOverrides: false,
    showRemark: true,
    compact: false
  })

  const treeProps = {
    children: 'children',
    label: 'label'
  }

  const filteredActions = computed(() => {
    const keyword = filters.keyword.trim().toLowerCase()
    return actions.value.filter((item) => {
      if (filters.featureKind && item.featureKind !== filters.featureKind) {
        return false
      }

      const overrideEffect = effectMap[item.id] || ''
      if (filters.onlyOverrides && !overrideEffect) {
        return false
      }

      switch (filters.overrideState) {
        case 'inherit':
          if (overrideEffect) return false
          break
        case 'allow':
          if (overrideEffect !== 'allow') return false
          break
        case 'deny':
          if (overrideEffect !== 'deny') return false
          break
        case 'overridden':
          if (!overrideEffect) return false
          break
      }

      if (!keyword) {
        return true
      }

      const haystack = [
        item.name,
        item.permissionKey,
        item.moduleCode,
        item.category,
        item.resourceCode,
        item.actionCode,
        item.description
      ]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()

      return haystack.includes(keyword)
    })
  })

  const filteredActionIds = computed(() => filteredActions.value.map((item) => item.id))
  const filteredActionCount = computed(() => filteredActions.value.length)
  const roleAllowCount = computed(() => Object.values(roleEffectMap).filter((item) => item === 'allow').length)
  const overrideAllowCount = computed(() => Object.values(effectMap).filter((item) => item === 'allow').length)
  const overrideDenyCount = computed(() => Object.values(effectMap).filter((item) => item === 'deny').length)

  const filteredGroups = computed<FeatureGroup[]>(() => buildPermissionGroups(filteredActions.value))

  const treeData = computed<PermissionTreeNode[]>(() =>
    buildPermissionTree(filteredGroups.value, (action) => ({
      meta: buildActionMeta(action),
      scopeText: formatScopeLabel(action.scopeCode, action.scopeName),
      roleEffectText: buildRoleEffectText(action.id)
    })) as PermissionTreeNode[]
  )

  function buildActionMeta(action: PermissionActionItem) {
    const parts = [action.permissionKey || `${action.resourceCode}:${action.actionCode}`]
    if (filters.showRemark && action.description) {
      parts.push(action.description)
    }
    return parts.filter(Boolean).join('  ')
  }

  function buildRoleEffectText(actionId: string) {
    const effect = roleEffectMap[actionId] || ''
    return effect === 'allow' ? '角色允许' : effect === 'deny' ? '角色拒绝' : '角色未配'
  }

  function applyEffects(actionIds: string[], effect: EffectValue) {
    actionIds.forEach((id) => {
      effectMap[id] = effect
    })
  }

  function syncExpandedKeys() {
    const nextKeys = buildDefaultExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    nextTick(() => {
      treeRef.value?.setExpandedKeys(nextKeys)
    })
  }

  function expandAll() {
    const nextKeys = buildAllExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    treeRef.value?.setExpandedKeys(nextKeys)
  }

  function collapseAll() {
    expandedKeys.value = []
    treeRef.value?.setExpandedKeys([])
  }

  async function loadRoleBaseline() {
    if (!props.member?.userId) return
    Object.keys(roleEffectMap).forEach((key) => delete roleEffectMap[key])

    const [memberRolesRes, availableRoles] = await Promise.all([
      fetchGetMyTeamMemberRoles(props.member.userId),
      fetchGetMyTeamRoles()
    ])
    const roleIdSet = new Set([
      ...(memberRolesRes?.global_role_ids || []),
      ...(memberRolesRes?.team_role_ids || []),
      ...(memberRolesRes?.role_ids || [])
    ])
    const roleIds = availableRoles
      .filter((role) => roleIdSet.has(role.roleId))
      .map((role) => role.roleId)

    if (roleIds.length === 0) return

    const roleActionsList = await Promise.all(roleIds.map((roleId) => fetchGetMyTeamRoleActions(roleId)))
    roleActionsList.forEach((result) => {
      ;(result?.actions || []).forEach((item) => {
        const current = roleEffectMap[item.action_id] || ''
        if (current === 'deny') return
        if (item.effect === 'deny') {
          roleEffectMap[item.action_id] = 'deny'
        } else if (item.effect === 'allow') {
          roleEffectMap[item.action_id] = 'allow'
        }
      })
    })
  }

  async function open() {
    if (!props.member?.userId) return
    visible.value = true
    loading.value = true
    try {
      const [teamActions, memberActions] = await Promise.all([
        fetchGetMyTeamActions(),
        fetchGetMyTeamMemberActions(props.member.userId)
      ])
      actions.value = teamActions.actions || []
      Object.keys(effectMap).forEach((key) => delete effectMap[key])
      actions.value.forEach((item) => {
        effectMap[item.id] = ''
      })
      memberActions.forEach((item) => {
        effectMap[item.actionId] = item.effect
      })

      await loadRoleBaseline()

      Object.assign(filters, {
        keyword: '',
        featureKind: '',
        overrideState: '',
        onlyOverrides: false,
        showRemark: true,
        compact: false
      })
      syncExpandedKeys()
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

  watch(filteredGroups, () => {
    syncExpandedKeys()
  })

  async function handleSubmit() {
    if (!props.member?.userId) return
    submitting.value = true
    try {
      const payload = Object.entries(effectMap)
        .filter(([, effect]) => effect)
        .map(([actionId, effect]) => ({
          action_id: actionId,
          effect: effect as 'allow' | 'deny'
        }))
      await fetchSetMyTeamMemberActions(props.member.userId, payload)
      ElMessage.success('保存成功')
      emit('success')
      visible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    } finally {
      submitting.value = false
    }
  }

  defineExpose({ open })
</script>

<style scoped>
  .member-action-dialog {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .dialog-alert {
    margin-bottom: 0;
  }

  .control-panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 14px;
    border: 1px solid #e7edf5;
    border-radius: 16px;
    background:
      linear-gradient(180deg, rgba(248, 251, 255, 0.92) 0%, rgba(255, 255, 255, 0.98) 100%);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.9),
      0 8px 24px rgba(15, 23, 42, 0.04);
  }

  .toolbar {
    display: grid;
    grid-template-columns: minmax(320px, 1fr) 150px 150px;
    gap: 12px;
    align-items: center;
  }

  .toolbar-input,
  .toolbar-select {
    width: 100%;
  }

  .option-row,
  .summary,
  .batch-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
  }

  .option-row {
    justify-content: space-between;
    gap: 12px;
  }

  .summary,
  .batch-bar {
    gap: 8px;
  }

  .summary-tag {
    height: 28px;
    padding: 0 10px;
    font-size: 12px;
    color: #5f6b7a;
    border-color: #d7dde6;
    background: #f8fafc;
    border-radius: 999px;
  }

  .summary-tag--baseline {
    color: #4d607e;
    border-color: #d7e0ef;
    background: #f4f7fc;
  }

  .summary-tag--allow {
    color: #316b56;
    border-color: #cde5d9;
    background: #f2fbf6;
  }

  .summary-tag--deny {
    color: #925f64;
    border-color: #efd6da;
    background: #fff6f7;
  }

  .option-switches {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
    gap: 10px;
  }

  .option-item {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    min-height: 34px;
    padding: 0 12px;
    border: 1px solid #e2e8f0;
    border-radius: 999px;
    background: rgba(255, 255, 255, 0.88);
    color: #526075;
    font-size: 12px;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.72);
  }

  .batch-label {
    font-size: 12px;
    color: #69778a;
  }

  .tree-wrapper {
    border: 1px solid #e5ebf3;
    border-radius: 18px;
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(249, 251, 254, 0.96) 100%);
    box-shadow:
      inset 0 1px 0 rgba(255, 255, 255, 0.8),
      0 10px 28px rgba(15, 23, 42, 0.04);
    padding: 10px;
  }

  .permission-tree {
    max-height: 520px;
    overflow: auto;
    padding-right: 2px;
  }

  :deep(.permission-tree .el-tree-node__content) {
    height: auto;
    min-height: 34px;
    margin: 2px 0;
    padding: 0;
    border-radius: 14px;
  }

  :deep(.permission-tree .el-tree-node__expand-icon) {
    color: #8a94a6;
    font-size: 12px;
  }

  .tree-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    padding: 8px 10px;
    border-radius: 14px;
    transition:
      background-color 0.18s ease,
      box-shadow 0.18s ease;
  }

  .tree-node:hover {
    background: rgba(244, 247, 252, 0.92);
  }

  .tree-node--feature {
    padding: 10px 12px;
    background:
      linear-gradient(135deg, rgba(248, 251, 255, 0.96) 0%, rgba(242, 246, 252, 0.9) 100%);
    border: 1px solid #e6edf6;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
  }

  .tree-node--module {
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.84);
    border: 1px solid #edf2f7;
  }

  .tree-node--action {
    padding: 6px 8px;
  }

  .node-main {
    min-width: 0;
    flex: 1;
  }

  .node-title {
    font-size: 12px;
    font-weight: 600;
    color: #243144;
    letter-spacing: 0.01em;
  }

  .node-subtitle {
    font-size: 11px;
    line-height: 1.45;
    color: #7b889c;
    margin-top: 2px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .node-actions {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    justify-content: flex-end;
  }

  .node-actions--leaf {
    gap: 4px;
  }

  .muted-tag {
    font-size: 11px;
    color: #6b7280;
    border-color: #dde3ea;
    background: #f8fafc;
    border-radius: 999px;
  }

  .effect-select {
    width: 92px;
  }

  :deep(.effect-select .el-input__wrapper) {
    border-radius: 999px;
    box-shadow: 0 0 0 1px #dde5ef inset;
    background: #f8fafc;
  }

  :deep(.effect-select .el-input__inner) {
    font-size: 12px;
    color: #415064;
  }

  :deep(.effect-select--allow .el-input__wrapper) {
    background: #eefaf3;
    box-shadow: 0 0 0 1px #bfe6ce inset;
  }

  :deep(.effect-select--allow .el-input__inner) {
    color: #1f6a4d;
  }

  :deep(.effect-select--deny .el-input__wrapper) {
    background: #fff3f4;
    box-shadow: 0 0 0 1px #efc8cf inset;
  }

  :deep(.effect-select--deny .el-input__inner) {
    color: #9b4956;
  }

  .is-compact {
    padding-top: 4px;
    padding-bottom: 4px;
  }

  .is-compact .node-subtitle {
    display: none;
  }

  @media (max-width: 900px) {
    .toolbar {
      grid-template-columns: 1fr;
    }

    .option-row {
      flex-direction: column;
      align-items: flex-start;
    }

    .option-switches {
      justify-content: flex-start;
    }

    .tree-node {
      flex-direction: column;
      align-items: flex-start;
    }

    .node-actions {
      justify-content: flex-start;
    }
  }
</style>
