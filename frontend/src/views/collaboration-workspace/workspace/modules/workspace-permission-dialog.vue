<template>
  <ElDrawer
    v-model="visible"
    :title="`协作空间边界 - ${props.collaborationWorkspaceName}`"
    size="960px"
    destroy-on-close
    class="collaboration-workspace-permission-dialog config-drawer"
    direction="rtl"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里配置的是协作空间边界减法。正式开通能力请通过功能包完成；这里保存的内容只会从功能包展开结果中屏蔽个别权限，不会额外开通新能力。
      </div>

      <PermissionSummaryTags :items="summaryItems" />

      <div v-if="featurePackages.length" class="package-card">
        <div class="package-title">已开通功能包</div>
        <div class="package-help"
          >功能包决定协作空间正式开通的基础能力；当前弹窗只负责从这些能力中屏蔽个别权限。</div
        >
        <div class="package-tags">
          <ElTag v-for="item in featurePackages" :key="item.id" type="success" effect="plain" round>
            {{ item.name }} · {{ item.actionCount ?? 0 }}
          </ElTag>
        </div>
      </div>

      <div class="source-note">
        <span
          >黄色统计表示来自功能包展开的基础能力，蓝色统计表示当前协作空间边界已屏蔽的权限。</span
        >
      </div>

      <PermissionSourcePanels
        v-model="selectedDerivedPackageId"
        :packages="featurePackages"
        :source-map="derivedSourceMap"
        :derived-items="derivedActionItems"
        :blocked-items="blockedActionItems"
        derived-title="功能包展开明细"
        blocked-title="协作空间边界屏蔽明细"
        open="actions"
        blocked-tag-type="primary"
        filtered-blocked-empty-text="当前筛选下暂无协作空间边界屏蔽项"
        empty-title="当前暂无协作空间边界来源"
        empty-text="请先检查协作空间功能包、协作空间上下文或协作空间边界快照是否已经生成。"
      />

      <PermissionActionCascaderPanel
        :actions="permissionActions"
        :selected-ids="selectedIds"
        footer-text="这里只保存协作空间边界结果；提交时会自动反推出需要屏蔽的权限，最终协作空间权限 = 功能包展开 - 协作空间屏蔽。提交前会自动展开父级选择。"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import PermissionSourcePanels from '@/components/business/permission/PermissionSourcePanels.vue'
  import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
  import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'
  import {
    fetchGetCollaborationWorkspaceActions,
    fetchGetCollaborationWorkspaceActionOrigins,
    fetchSetCollaborationWorkspaceActions
  } from '@/api/collaboration-workspace'
  import {
    fetchGetPermissionActionOptions,
    fetchGetCollaborationWorkspaceFeaturePackages
  } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    collaborationWorkspaceId: string
    collaborationWorkspaceName: string
    appKey?: string
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const saving = ref(false)
  const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedIds = ref<string[]>([])
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const candidateActionIds = ref<string[]>([])
  const blockedActionIds = ref<string[]>([])
  const derivedSourceMap = ref<Record<string, string[]>>({})
  const selectedDerivedPackageId = ref('')
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())
  const derivedActions = computed(() => {
    const idSet = new Set(candidateActionIds.value)
    return permissionActions.value.filter((item) => idSet.has(item.id))
  })
  const blockedActions = computed(() => {
    const idSet = new Set(blockedActionIds.value)
    return permissionActions.value.filter((item) => idSet.has(item.id))
  })
  const blockedActionCount = computed(() => blockedActions.value.length)
  const summaryItems = computed(() => [
    { label: '协作空间', value: props.collaborationWorkspaceName || '-' },
    { label: '边界结果', value: selectedIds.value.length, type: 'success' as const },
    { label: '候选', value: candidateActionIds.value.length, type: 'info' as const },
    { label: '功能包', value: featurePackages.value.length },
    { label: '功能包展开', value: candidateActionIds.value.length, type: 'warning' as const },
    { label: '边界已屏蔽', value: blockedActionCount.value, type: 'primary' as const }
  ])
  const derivedActionItems = computed(() =>
    derivedActions.value.map((item) => ({ id: item.id, label: item.name }))
  )
  const blockedActionItems = computed(() =>
    blockedActions.value.map((item) => ({ id: item.id, label: item.name }))
  )

  watch(
    () => props.modelValue,
    (open) => {
      if (open) {
        loadData()
      }
    }
  )

  async function loadData() {
    if (!props.collaborationWorkspaceId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    loading.value = true
    try {
      const [actionsRes, currentRes, packageRes, originsRes] = await Promise.all([
        fetchGetPermissionActionOptions({ status: 'normal' }),
        fetchGetCollaborationWorkspaceActions(props.collaborationWorkspaceId, currentAppKey.value),
        fetchGetCollaborationWorkspaceFeaturePackages(
          props.collaborationWorkspaceId,
          currentAppKey.value
        ),
        fetchGetCollaborationWorkspaceActionOrigins(
          props.collaborationWorkspaceId,
          currentAppKey.value
        )
      ])

      permissionActions.value = actionsRes?.records || []
      selectedIds.value = [...(currentRes?.action_ids || [])]
      featurePackages.value = packageRes?.packages || []
      candidateActionIds.value = [...(originsRes?.derived_action_ids || [])]
      blockedActionIds.value = [...(originsRes?.blocked_action_ids || [])]
      derivedSourceMap.value = Object.fromEntries(
        (originsRes?.derived_sources || []).map((item: any) => [item.action_id, item.package_ids])
      )
      selectedDerivedPackageId.value = ''
    } catch (error: any) {
      ElMessage.error(error?.message || '加载协作空间边界失败')
    } finally {
      loading.value = false
    }
  }

  function handleCancel() {
    visible.value = false
  }

  async function handleSave() {
    if (!props.collaborationWorkspaceId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    saving.value = true
    try {
      await fetchSetCollaborationWorkspaceActions(
        props.collaborationWorkspaceId,
        expandSelectedValues(selectedIds.value, permissionActions.value),
        currentAppKey.value
      )
      ElMessage.success('协作空间边界已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存协作空间边界失败')
    } finally {
      saving.value = false
    }
  }

  function expandSelectedValues(
    values: string[],
    actions: Api.SystemManage.PermissionActionItem[]
  ) {
    const result = new Set<string>()
    const featureMap = new Map<string, Api.SystemManage.PermissionActionItem[]>()
    const moduleMap = new Map<string, Api.SystemManage.PermissionActionItem[]>()

    actions.forEach((action) => {
      const featureKey = `${action.featureGroupId || action.featureKind || 'business'}`
      const moduleKey = `${action.moduleGroupId || action.moduleCode || action.resourceCode || 'default'}`
      const featureValue = `feature:${featureKey}`
      const moduleValue = `module:${featureKey}:${moduleKey}`

      const featureItems = featureMap.get(featureValue) || []
      featureItems.push(action)
      featureMap.set(featureValue, featureItems)

      const moduleItems = moduleMap.get(moduleValue) || []
      moduleItems.push(action)
      moduleMap.set(moduleValue, moduleItems)
    })

    values.forEach((value) => {
      if (featureMap.has(value)) {
        featureMap.get(value)!.forEach((item) => result.add(item.id))
        return
      }
      if (moduleMap.has(value)) {
        moduleMap.get(value)!.forEach((item) => result.add(item.id))
        return
      }
      result.add(value)
    })

    return Array.from(result)
  }
</script>

<style scoped lang="scss">
  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .dialog-note {
    color: #6b7280;
    line-height: 1.6;
  }

  .package-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 12px;
    background: #f8fafc;
  }

  .package-title {
    color: #475569;
    font-size: 13px;
  }

  .package-help {
    color: #64748b;
    font-size: 12px;
    line-height: 1.5;
  }

  .package-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .source-note {
    font-size: 12px;
    color: #64748b;
  }
</style>
