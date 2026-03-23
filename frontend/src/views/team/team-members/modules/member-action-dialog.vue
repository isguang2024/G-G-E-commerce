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
        title="这里配置的是团队内个人功能权限覆盖。页面只展示该成员当前角色功能包展开后的可用能力；默认继承角色权限，单独允许或单独拒绝只用于少量例外场景。"
      />

      <div class="summary-header">
        <ElTag effect="plain" round>成员 {{ member?.userName || '-' }}</ElTag>
        <ElTag type="info" effect="plain" round>基础角色 {{ assignedGlobalRoleCount }}</ElTag>
        <ElTag type="success" effect="plain" round>团队自定义 {{ assignedCustomRoleCount }}</ElTag>
        <ElTag type="warning" effect="plain" round>功能包展开 {{ derivedActionCount }}</ElTag>
        <ElTag type="primary" effect="plain" round>团队补充 {{ manualActionCount }}</ElTag>
      </div>

      <div v-if="derivedActions.length || manualActions.length" class="source-detail-grid">
        <div v-if="derivedActions.length" class="source-card source-card--derived">
          <div class="source-header">
            <div class="source-title">功能包展开能力</div>
            <ElButton
              v-if="selectedDerivedPackage"
              type="warning"
              text
              @click="goToFeaturePackagePage(selectedDerivedPackage)"
            >
              前往功能包页
            </ElButton>
          </div>
          <div v-if="derivedSourcePackages.length" class="package-filter-row">
            <ElTag
              :type="selectedDerivedPackageId ? 'info' : 'warning'"
              effect="plain"
              round
              class="package-filter-tag"
              @click="selectedDerivedPackageId = ''"
            >
              全部功能包
            </ElTag>
            <ElTag
              v-for="item in derivedSourcePackages"
              :key="item.id"
              :type="selectedDerivedPackageId === item.id ? 'warning' : 'info'"
              effect="plain"
              round
              class="package-filter-tag"
              @click="selectedDerivedPackageId = selectedDerivedPackageId === item.id ? '' : item.id"
            >
              {{ item.name }}
            </ElTag>
          </div>
          <div class="source-tags">
            <ElTag
              v-for="item in filteredDerivedActions"
              :key="item.id"
              type="warning"
              effect="plain"
              round
              :title="buildDerivedSourceText(item.id)"
            >
              {{ item.name }}
            </ElTag>
          </div>
        </div>

        <div v-if="manualActions.length" class="source-card source-card--manual">
          <div class="source-title">团队补充能力</div>
          <div class="source-tags">
            <ElTag
              v-for="item in manualActions"
              :key="item.id"
              type="primary"
              effect="plain"
              round
            >
              {{ item.name }}
            </ElTag>
          </div>
        </div>
      </div>

      <section class="control-panel">
        <div class="summary">
          <ElTag effect="plain" class="summary-tag">已开通 {{ actions.length }}</ElTag>
          <ElTag effect="plain" class="summary-tag summary-tag--baseline">角色基线 {{ roleAllowCount }}</ElTag>
          <ElTag effect="plain" class="summary-tag summary-tag--allow">单独允许 {{ overrideAllowCount }}</ElTag>
          <ElTag effect="plain" class="summary-tag summary-tag--deny">单独拒绝 {{ overrideDenyCount }}</ElTag>
        </div>

        <div v-if="assignedRoles.length" class="role-chip-row">
          <span class="role-chip-label">当前角色</span>
          <ElTag
            v-for="role in assignedRoles"
            :key="role.roleId"
            :type="role.isGlobal ? 'info' : 'success'"
            effect="plain"
            round
          >
            {{ role.roleName }}
          </ElTag>
        </div>

        <div class="toolbar">
          <ElInput
            v-model="filters.keyword"
            clearable
            placeholder="搜索权限名称/权限键/模块归属/兼容编码"
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
              <span>显示 ID/说明</span>
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

      <ActionPermissionTreePanel
        ref="treePanelRef"
        :loading="loading"
        :tree-data="treeData"
        empty-description="当前团队未开通功能权限"
      >
        <template #node="{ data }">
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
      </ActionPermissionTreePanel>
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
  import ActionPermissionTreePanel from '@/components/business/permission/action-permission-tree-panel.vue'
  import {
    fetchGetMyTeam,
    fetchGetMyTeamMemberActions,
    fetchGetMyTeamMemberRoles,
    fetchGetMyTeamRoleActions,
    fetchGetMyTeamRoles,
    fetchSetMyTeamMemberActions
  } from '@/api/team'
  import { fetchGetTeamFeaturePackages } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'

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
    roleEffectText?: string
  }

  const props = defineProps<Props>()
  const emit = defineEmits(['success'])
  const router = useRouter()

  const visible = ref(false)
  const loading = ref(false)
  const submitting = ref(false)
  const actions = ref<PermissionActionItem[]>([])
  const treePanelRef = ref<InstanceType<typeof ActionPermissionTreePanel>>()
  const expandedKeys = ref<string[]>([])
  const effectMap = reactive<Record<string, EffectValue>>({})
  const roleEffectMap = reactive<Record<string, EffectValue>>({})
  const assignedRoles = ref<Api.SystemManage.RoleListItem[]>([])
  const derivedActionIds = ref<string[]>([])
  const manualActionIds = ref<string[]>([])
  const derivedSourceMap = ref<Record<string, string[]>>({})
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedDerivedPackageId = ref('')
  const derivedActionCount = ref(0)
  const manualActionCount = ref(0)
  const filters = reactive({
    keyword: '',
    featureKind: '',
    overrideState: '',
    onlyOverrides: false,
    showRemark: false,
    compact: false
  })

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
  const assignedGlobalRoleCount = computed(() => assignedRoles.value.filter((role) => role.isGlobal).length)
  const assignedCustomRoleCount = computed(() => assignedRoles.value.filter((role) => !role.isGlobal).length)
  const derivedActions = computed(() => {
    const idSet = new Set(derivedActionIds.value)
    return actions.value.filter((item) => idSet.has(item.id))
  })
  const manualActions = computed(() => {
    const idSet = new Set(manualActionIds.value)
    return actions.value.filter((item) => idSet.has(item.id))
  })
  const derivedSourcePackages = computed(() => {
    const packageIdSet = new Set(Object.values(derivedSourceMap.value).flat())
    return featurePackages.value.filter((item) => packageIdSet.has(item.id))
  })
  const filteredDerivedActions = computed(() => {
    if (!selectedDerivedPackageId.value) return derivedActions.value
    return derivedActions.value.filter((item) => (derivedSourceMap.value[item.id] || []).includes(selectedDerivedPackageId.value))
  })
  const selectedDerivedPackage = computed(
    () => featurePackages.value.find((item) => item.id === selectedDerivedPackageId.value) || null
  )

  const filteredGroups = computed<FeatureGroup[]>(() => buildPermissionGroups(filteredActions.value))

  const treeData = computed<PermissionTreeNode[]>(() =>
    buildPermissionTree(filteredGroups.value, (action) => ({
      meta: buildActionMeta(action),
      roleEffectText: buildRoleEffectText(action.id)
    })) as PermissionTreeNode[]
  )

  function buildActionMeta(action: PermissionActionItem) {
    if (!filters.showRemark) {
      return ''
    }
    const parts = [action.permissionKey || `${action.resourceCode}:${action.actionCode}`]
    if (action.description) {
      parts.push(action.description)
    }
    return parts.join('  ')
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
      treePanelRef.value?.setExpandedKeys(nextKeys)
    })
  }

  function expandAll() {
    const nextKeys = buildAllExpandedKeys(treeData.value)
    expandedKeys.value = nextKeys
    treePanelRef.value?.expandAll()
  }

  function collapseAll() {
    expandedKeys.value = []
    treePanelRef.value?.collapseAll()
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
    assignedRoles.value = availableRoles
      .filter((role) => roleIdSet.has(role.roleId))
      .sort((left, right) => {
        if (left.isGlobal === right.isGlobal) return left.roleName.localeCompare(right.roleName, 'zh-CN')
        return left.isGlobal ? -1 : 1
      })
    const roleIds = assignedRoles.value.map((role) => role.roleId)

    if (roleIds.length === 0) return

    const roleActionsList = await Promise.all(roleIds.map((roleId) => fetchGetMyTeamRoleActions(roleId)))
    roleActionsList.forEach((result) => {
      ;(result?.action_ids || []).forEach((actionId) => {
        const current = roleEffectMap[actionId] || ''
        if (current === 'deny') return
        roleEffectMap[actionId] = 'allow'
      })
    })
  }

  async function open() {
    if (!props.member?.userId) return
    visible.value = true
    loading.value = true
    try {
      const [memberActions, team] = await Promise.all([
        fetchGetMyTeamMemberActions(props.member.userId),
        fetchGetMyTeam()
      ])
      const packageRes = team?.id ? await fetchGetTeamFeaturePackages(team.id) : { packages: [] }
      actions.value = memberActions.availableActions || []
      derivedActionIds.value = [...(memberActions.availableActionIds || [])]
      manualActionIds.value = []
      derivedSourceMap.value = Object.fromEntries(
        (memberActions.derivedSources || []).map((item) => [item.actionId, item.packageIds])
      )
      featurePackages.value = packageRes?.packages || []
      selectedDerivedPackageId.value = ''
      derivedActionCount.value = memberActions.availableActionIds?.length || 0
      manualActionCount.value = 0
      Object.keys(effectMap).forEach((key) => delete effectMap[key])
      actions.value.forEach((item) => {
        effectMap[item.id] = ''
      })
      ;(memberActions.actions || []).forEach((item) => {
        effectMap[item.actionId] = item.effect
      })

      await loadRoleBaseline()

      Object.assign(filters, {
        keyword: '',
        featureKind: '',
        overrideState: '',
        onlyOverrides: false,
        showRemark: false,
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

  function buildDerivedSourceText(actionId: string) {
    const packageIdSet = new Set(derivedSourceMap.value[actionId] || [])
    const names = featurePackages.value.filter((item) => packageIdSet.has(item.id)).map((item) => item.name)
    return names.length ? `来源功能包：${names.join('、')}` : '来源功能包未命名'
  }

  function goToFeaturePackagePage(item: Api.SystemManage.FeaturePackageItem) {
    router.push({
      name: 'FeaturePackage',
      query: {
        packageKey: item.packageKey,
        contextType: item.contextType || 'team',
        open: 'actions'
      }
    })
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

  .summary-header {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .source-detail-grid {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  }

  .source-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 12px;
    border: 1px solid #e5e7eb;
    background: #fff;
  }

  .source-card--derived {
    border-color: #f3d38a;
    background: #fffaf0;
  }

  .source-card--manual {
    border-color: #bfd3ff;
    background: #f5f9ff;
  }

  .source-title {
    font-size: 13px;
    color: #475569;
  }

  .source-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
  }

  .source-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .package-filter-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .package-filter-tag {
    cursor: pointer;
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

  .role-chip-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
  }

  .role-chip-label {
    font-size: 12px;
    color: #69778a;
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

  .tree-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
    min-height: 34px;
    padding: 6px 8px;
    border-radius: 12px;
  }

  .tree-node--feature {
    background: rgba(246, 249, 253, 0.92);
    border: 1px solid #e5ebf3;
  }

  .tree-node--module {
    background: rgba(255, 255, 255, 0.9);
    border: 1px solid #edf2f7;
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

  .muted-tag {
    font-size: 11px;
    color: #6b7280;
    border-color: #dde3ea;
    background: #f8fafc;
    border-radius: 999px;
  }

  .effect-select {
    width: 110px;
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
