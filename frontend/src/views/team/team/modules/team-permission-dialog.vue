<template>
  <ElDialog
    v-model="visible"
    :title="`团队功能权限 - ${teamName}`"
    width="980px"
    destroy-on-close
    class="team-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        配置团队可使用的功能能力。关闭的功能不会对该团队成员开放，也不会出现在团队相关入口中。
      </div>

      <div class="team-summary">
        <ElTag effect="plain" round>团队 {{ teamName }}</ElTag>
        <ElTag type="success" effect="plain" round>已开通 {{ selectedIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>总计 {{ permissionActions.length }}</ElTag>
      </div>

      <PermissionActionWorkbench
        mode="team"
        :actions="permissionActions"
        :selected-ids="selectedIds"
        :loading="loading"
        search-placeholder="搜索团队可开通功能"
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
import PermissionActionWorkbench from '@/components/business/permission/PermissionActionWorkbench.vue'
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
    ElMessage.error(error?.message || '加载团队权限失败')
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
    await fetchSetTeamActions(props.teamId, selectedIds.value)
    ElMessage.success('团队功能权限已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存团队权限失败')
  } finally {
    saving.value = false
  }
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

.team-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
