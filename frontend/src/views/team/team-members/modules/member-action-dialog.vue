<template>
  <ElDialog
    v-model="visible"
    :title="`成员功能权限 - ${member?.userName || ''}`"
    width="760px"
    destroy-on-close
  >
    <div v-loading="loading">
      <ElAlert
        type="info"
        :closable="false"
        class="mb-4"
        title="这里配置的是团队内个人覆盖权限。未设置时沿用角色权限；允许或拒绝会覆盖角色结果。"
      />
      <ElEmpty v-if="actions.length === 0" description="当前团队未开通功能权限" />
      <ElTable v-else :data="actions" border max-height="480">
        <ElTableColumn prop="name" label="权限名称" min-width="180" />
        <ElTableColumn prop="resourceCode" label="资源编码" min-width="140" />
        <ElTableColumn prop="actionCode" label="动作编码" min-width="150" />
        <ElTableColumn label="个人覆盖" width="200">
          <template #default="{ row }">
            <ElSelect v-model="effectMap[row.id]" clearable placeholder="继承角色" style="width: 150px">
              <ElOption label="允许" value="allow" />
              <ElOption label="拒绝" value="deny" />
            </ElSelect>
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
  import {
    fetchGetMyTeamActions,
    fetchGetMyTeamMemberActions,
    fetchSetMyTeamMemberActions
  } from '@/api/team'
  import { ElMessage } from 'element-plus'

  interface Props {
    member: Api.SystemManage.TeamMemberItem | null
  }

  const props = defineProps<Props>()
  const emit = defineEmits(['success'])

  const visible = ref(false)
  const loading = ref(false)
  const submitting = ref(false)
  const actions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const effectMap = reactive<Record<string, 'allow' | 'deny' | ''>>({})

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
      for (const item of memberActions) {
        effectMap[item.actionId] = item.effect
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '获取成员功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

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
