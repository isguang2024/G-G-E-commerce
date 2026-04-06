<template>
  <ElDrawer
    v-model="visible"
    :title="drawerTitle"
    size="1240px"
    destroy-on-close
    direction="rtl"
    class="config-drawer"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        {{ noteText }}
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>功能包 {{ packageName }}</ElTag>
        <ElTag type="warning" effect="plain" round>上下文 {{ contextLabel }}</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ selectedIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>总计 {{ filteredActions.length }}</ElTag>
      </div>

      <PermissionActionCascaderPanel
        :actions="filteredActions"
        :selected-ids="selectedIds"
        :footer-text="footerText"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
  import {
    fetchGetFeaturePackageActions,
    fetchGetPermissionActionOptions,
    fetchSetFeaturePackageActions
  } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    contextType?: 'personal' | 'collaboration' | 'common' | string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    packageId: '',
    packageName: '',
    contextType: 'collaboration'
  })

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
  const allActions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedIds = ref<string[]>([])

  const filteredActions = computed(() =>
    allActions.value.filter((item) =>
      supportsActionContext(
        props.contextType || 'collaboration',
        item.contextType || 'collaboration'
      )
    )
  )

  const contextLabel = computed(() => formatContextType(props.contextType))
  const drawerTitle = computed(() => `功能包功能范围配置 - ${props.packageName}`)
  const noteText = computed(() => {
    const scopeLabel = getScopeLabel(props.contextType)
    return `这里配置的是${scopeLabel}启用该功能包后可进入的功能范围，不是直接给角色或成员授予权限。后续角色和成员的权限分配，仍然只能在这批已开通范围内继续细分。`
  })
  const footerText = computed(() => {
    const scopeLabel = getScopeLabel(props.contextType)
    return `这里保存的是${scopeLabel}可开放能力范围，角色和成员分配仍然基于基础功能权限。`
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) loadData()
    }
  )

  async function loadData() {
    if (!props.packageId) return
    loading.value = true
    try {
      const [actionsRes, currentRes] = await Promise.all([
        fetchGetPermissionActionOptions({
          status: 'normal',
          contextType:
            props.contextType === 'common' ? undefined : props.contextType || 'collaboration'
        }),
        fetchGetFeaturePackageActions(props.packageId)
      ])
      allActions.value = actionsRes?.records || []
      selectedIds.value = [...(currentRes?.action_ids || [])]
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包功能范围失败')
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    if (!props.packageId) return
    saving.value = true
    try {
      const stats = await fetchSetFeaturePackageActions(
        props.packageId,
        expandSelectedValues(selectedIds.value, filteredActions.value)
      )
      ElMessage.success(formatRefreshMessage(stats))
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存功能包功能范围失败')
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

  function supportsActionContext(packageContextType: string, actionContextType: string) {
    if (packageContextType === 'common') {
      return (
        actionContextType === 'personal' ||
        actionContextType === 'collaboration' ||
        actionContextType === 'common'
      )
    }
    if (
      packageContextType === 'personal,collaboration' ||
      packageContextType === 'collaboration,personal'
    ) {
      return (
        actionContextType === 'personal' ||
        actionContextType === 'collaboration' ||
        actionContextType === 'common'
      )
    }
    return packageContextType === actionContextType || actionContextType === 'common'
  }

  function formatContextType(contextType?: string) {
    if (contextType === 'personal') return '个人空间'
    if (contextType === 'collaboration') return '协作空间'
    if (contextType === 'common') return '通用'
    if (contextType === 'personal,collaboration' || contextType === 'collaboration,personal')
      return '个人空间/协作空间'
    return contextType || '-'
  }

  function getScopeLabel(contextType?: string) {
    if (contextType === 'personal') return '个人空间'
    if (contextType === 'collaboration') return '协作空间'
    if (contextType === 'common') return '个人空间或协作空间'
    if (contextType === 'personal,collaboration' || contextType === 'collaboration,personal')
      return '个人空间或协作空间'
    return '当前上下文'
  }

  function formatRefreshMessage(stats?: Api.SystemManage.RefreshStats) {
    return `本次增量刷新：角色 ${stats?.roleCount || 0}、协作空间 ${stats?.collaborationWorkspaceCount || 0}、用户 ${stats?.userCount || 0}、耗时 ${stats?.elapsedMilliseconds || 0} ms`
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

  .summary-card {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
