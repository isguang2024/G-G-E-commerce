<template>
  <ElDialog
    v-model="visible"
    :title="`团队功能权限 - ${teamName}`"
    width="960px"
    destroy-on-close
    class="team-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        配置团队可开通的功能权限。父级和子级都可以直接选择，保存时会自动换算成具体权限项。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>团队 {{ teamName }}</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ selectedIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>总计 {{ permissionActions.length }}</ElTag>
      </div>

      <PermissionActionCascaderPanel
        :actions="permissionActions"
        :selected-ids="selectedIds"
        footer-text="团队功能权限只保留开通结果，提交前会自动展开父级选择。"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
import { fetchGetTeamActions, fetchSetTeamActions } from '@/api/team'
import { fetchGetPermissionActionList } from '@/api/system-manage'

interface Props {
  modelValue: boolean
  teamId: string
  teamName: string
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

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      loadData()
    }
  }
)

async function loadData() {
  if (!props.teamId) return
  loading.value = true
  try {
    const [actionsRes, currentRes] = await Promise.all([
      fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal', scopeCode: 'team' }),
      fetchGetTeamActions(props.teamId)
    ])

    permissionActions.value = (actionsRes?.records || []).filter((item) => {
      return item.scopeCode === 'team' || item.requiresTenantContext
    })
    selectedIds.value = [...(currentRes?.actionIds || [])]
  } catch (error: any) {
    ElMessage.error(error?.message || '加载团队功能权限失败')
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  visible.value = false
}

async function handleSave() {
  if (!props.teamId) return
  saving.value = true
  try {
    await fetchSetTeamActions(props.teamId, expandSelectedValues(selectedIds.value, permissionActions.value))
    ElMessage.success('团队功能权限已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存团队功能权限失败')
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
