<template>
  <ElDialog
    v-model="visible"
    :title="`团队功能权限 - ${teamName}`"
    width="760px"
    destroy-on-close
  >
    <div v-loading="loading">
      <ElEmpty v-if="actions.length === 0" description="暂无已注册功能权限" />
      <ElTable v-else :data="actions" border max-height="480">
        <ElTableColumn prop="name" label="权限名称" min-width="180" />
        <ElTableColumn prop="resourceCode" label="资源编码" min-width="130" />
        <ElTableColumn prop="actionCode" label="动作编码" min-width="150" />
        <ElTableColumn label="作用域" width="90">
          <template #default="{ row }">
            <ElTag :type="row.scopeCode === 'team' ? 'success' : 'primary'">
              {{ row.scopeName || (row.scopeCode === 'team' ? '团队' : '平台') }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="已开通" width="100">
          <template #default="{ row }">
            <ElSwitch v-model="selectedMap[row.id]" />
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetPermissionActionList } from '@/api/system-manage'
  import { fetchGetTeamActions, fetchSetTeamActions } from '@/api/team'
  import { ElMessage } from 'element-plus'

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
  const submitting = ref(false)
  const actions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedMap = reactive<Record<string, boolean>>({})

  async function loadData() {
    if (!props.teamId) return
    loading.value = true
    try {
      const [actionList, teamActions] = await Promise.all([
        fetchGetPermissionActionList({ current: 1, size: 500, scopeCode: 'team' }),
        fetchGetTeamActions(props.teamId)
      ])
      actions.value = actionList.records || []
      Object.keys(selectedMap).forEach((key) => delete selectedMap[key])
      for (const action of actions.value) {
        selectedMap[action.id] = teamActions.actionIds.includes(action.id)
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '获取团队功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

  watch(
    () => [visible.value, props.teamId],
    ([opened]) => {
      if (opened) loadData()
    }
  )

  async function handleSubmit() {
    submitting.value = true
    try {
      const actionIds = Object.keys(selectedMap).filter((id) => selectedMap[id])
      await fetchSetTeamActions(props.teamId, actionIds)
      ElMessage.success('保存成功')
      emit('success')
      visible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    } finally {
      submitting.value = false
    }
  }
</script>
