<template>
  <ElDialog
    v-model="visible"
    :title="`功能包权限配置 - ${packageName}`"
    width="980px"
    destroy-on-close
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        为功能包选择包含的功能权限。保存时会自动展开父级节点，并校验功能包与功能权限的上下文类型一致。
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
        footer-text="功能包保存的是能力集合，提交前会自动展开父级选择。"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
  import {
    fetchGetFeaturePackageActions,
    fetchGetPermissionActionList,
    fetchSetFeaturePackageActions
  } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    contextType?: 'platform' | 'team' | string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    packageId: '',
    packageName: '',
    contextType: 'team'
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
    allActions.value.filter((item) => (item.contextType || 'team') === (props.contextType || 'team'))
  )

  const contextLabel = computed(() => (props.contextType === 'platform' ? '平台' : '团队'))

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
        fetchGetPermissionActionList({
          current: 1,
          size: 1000,
          status: 'normal',
          contextType: props.contextType || 'team'
        }),
        fetchGetFeaturePackageActions(props.packageId)
      ])
      allActions.value = actionsRes?.records || []
      selectedIds.value = [...(currentRes?.action_ids || [])]
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包权限失败')
    } finally {
      loading.value = false
    }
  }

  async function handleSave() {
    if (!props.packageId) return
    saving.value = true
    try {
      await fetchSetFeaturePackageActions(props.packageId, expandSelectedValues(selectedIds.value, filteredActions.value))
      ElMessage.success('功能包权限已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存功能包权限失败')
    } finally {
      saving.value = false
    }
  }

  function expandSelectedValues(values: string[], actions: Api.SystemManage.PermissionActionItem[]) {
    const result = new Set<string>()
    const featureMap = new Map<string, Api.SystemManage.PermissionActionItem[]>()
    const moduleMap = new Map<string, Api.SystemManage.PermissionActionItem[]>()

    actions.forEach((action) => {
      const featureKey = `${action.featureKind || 'business'}`
      const moduleKey = `${action.moduleCode || action.resourceCode || 'default'}`
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

  .summary-card {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
