<template>
  <ElDrawer
    v-model="visible"
    :title="`协作空间角色权限裁剪 - ${roleTitle}`"
    size="980px"
    destroy-on-close
    direction="rtl"
    class="config-drawer"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        {{
          props.roleData?.isGlobal
            ? '基础协作空间角色默认继承当前协作空间已开通功能包，这里只读展示最终可用角色权限结果。'
            : '请先为角色绑定功能包，再在角色功能包展开范围内配置权限裁剪。'
        }}
      </div>

      <PermissionSummaryTags :items="summaryItems" />

      <div v-if="featurePackages.length" class="package-summary-card">
        <div class="source-title">当前角色功能包</div>
        <div class="source-tags">
          <ElTag v-for="item in featurePackages" :key="item.id" type="success" effect="plain" round>
            {{ item.name }}
          </ElTag>
        </div>
      </div>

      <PermissionSourcePanels
        v-model="selectedDerivedPackageId"
        :packages="featurePackages"
        :source-map="derivedSourceMap"
        :derived-items="derivedActionItems"
        :blocked-items="blockedActionItems"
        derived-title="功能包展开能力"
        blocked-title="当前角色已关闭能力"
        open="actions"
        filtered-blocked-empty-text="当前筛选下暂无角色显式关闭能力"
        empty-title="当前暂无角色能力来源"
        empty-text="请先为角色绑定功能包，或检查当前协作空间快照是否已经刷新。"
      />

      <PermissionActionCascaderPanel
        :actions="actions"
        :selected-ids="selectedIds"
        footer-text="这里只保存角色在功能包范围内的保留结果；未保留项将视为被角色屏蔽。"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton v-if="!props.roleData?.isGlobal" @click="keepAll">全部保留</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" @click="blockAll">全部屏蔽</ElButton>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton
        v-if="!props.roleData?.isGlobal"
        type="primary"
        :loading="saving"
        @click="handleSave"
        >保存</ElButton
      >
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
    fetchGetMyCollaborationWorkspaceBoundaryRoleActions,
    fetchGetMyCollaborationWorkspaceBoundaryRolePackages,
    fetchSetMyCollaborationWorkspaceBoundaryRoleActions
  } from '@/api/collaboration-workspace'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
    appKey?: string
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const loading = ref(false)
  const saving = ref(false)
  const actions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedIds = ref<string[]>([])
  const candidateActionIds = ref<string[]>([])
  const derivedSourceMap = ref<Record<string, string[]>>({})
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedDerivedPackageId = ref('')
  const candidateActionCount = ref(0)
  const inherited = ref(false)
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())
  const derivedActions = computed(() => {
    const idSet = new Set(candidateActionIds.value)
    return actions.value.filter((item) => idSet.has(item.id))
  })
  const blockedActions = computed(() => {
    const selectedIdSet = new Set(selectedIds.value)
    return actions.value.filter((item) => !selectedIdSet.has(item.id))
  })
  const derivedActionItems = computed(() =>
    derivedActions.value.map((item) => ({ id: item.id, label: item.name }))
  )
  const blockedActionItems = computed(() =>
    blockedActions.value.map((item) => ({ id: item.id, label: item.name }))
  )

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })
  const roleTitle = computed(() => props.roleData?.roleName || '')
  const blockedCount = computed(() => Math.max(actions.value.length - selectedIds.value.length, 0))
  const summaryItems = computed(() => [
    { label: '角色', value: roleTitle.value || '-' },
    {
      label: '类型',
      value: props.roleData?.isGlobal ? '基础角色' : '协作空间自定义',
      type: props.roleData?.isGlobal ? ('info' as const) : ('success' as const)
    },
    { label: '角色功能包', value: featurePackages.value.length, type: 'warning' as const },
    {
      label: '继承模式',
      value: inherited.value ? '继承协作空间功能包' : '角色独立功能包',
      type: 'primary' as const
    },
    { label: '功能包展开', value: candidateActionCount.value, type: 'warning' as const },
    { label: '可裁剪', value: actions.value.length, type: 'success' as const },
    { label: '已保留', value: selectedIds.value.length, type: 'success' as const },
    { label: '已关闭', value: blockedCount.value, type: 'danger' as const }
  ])

  watch(
    () => props.modelValue,
    async (open) => {
      if (!open || !props.roleData?.roleId || !currentAppKey.value) {
        if (open && !currentAppKey.value) {
          ElMessage.warning('缺少 app 上下文')
        }
        return
      }
      loading.value = true
      try {
        const [packagesRes, selected] = await Promise.all([
          fetchGetMyCollaborationWorkspaceBoundaryRolePackages(
            props.roleData.roleId,
            currentAppKey.value
          ),
          fetchGetMyCollaborationWorkspaceBoundaryRoleActions(
            props.roleData.roleId,
            currentAppKey.value
          )
        ])
        actions.value = selected?.actions || []
        selectedIds.value = [...(selected?.action_ids || [])]
        candidateActionIds.value = [...(selected?.available_action_ids || [])]
        derivedSourceMap.value = Object.fromEntries(
          (selected?.derived_sources || []).map((item: any) => [item.action_id, item.package_ids])
        )
        featurePackages.value = packagesRes?.packages || []
        selectedDerivedPackageId.value = ''
        inherited.value = Boolean(packagesRes?.inherited)
        candidateActionCount.value = selected?.available_action_ids?.length || 0
      } catch (error: any) {
        ElMessage.error(error?.message || '加载协作空间角色权限裁剪失败')
      } finally {
        loading.value = false
      }
    }
  )

  async function handleSave() {
    if (!props.roleData?.roleId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    saving.value = true
    try {
      await fetchSetMyCollaborationWorkspaceBoundaryRoleActions(
        props.roleData.roleId,
        selectedIds.value,
        currentAppKey.value
      )
      ElMessage.success('协作空间角色权限裁剪已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存协作空间角色权限裁剪失败')
    } finally {
      saving.value = false
    }
  }

  function keepAll() {
    selectedIds.value = actions.value.map((item) => item.id)
  }

  function blockAll() {
    selectedIds.value = []
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

  .package-summary-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 12px 14px;
    border-radius: 12px;
    border: 1px solid #d6e8d8;
    background: #f7fcf7;
  }

  .source-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
