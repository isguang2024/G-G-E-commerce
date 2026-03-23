<template>
  <ElDialog
    v-model="visible"
    :title="`团队角色功能权限 - ${roleTitle}`"
    width="980px"
    destroy-on-close
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        {{ props.roleData?.isGlobal ? '基础团队角色仅支持查看当前团队内的生效能力。' : '这里只能配置当前团队已开通能力范围内的角色功能权限。' }}
      </div>

      <PermissionActionCascaderPanel
        :actions="actions"
        :selected-ids="selectedIds"
        footer-text="保存后该团队角色只会在当前团队上下文中生效。"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
  import { fetchGetMyTeamActions, fetchGetMyTeamRoleActions, fetchSetMyTeamRoleActions } from '@/api/team'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void; (e: 'success'): void }>()

  const loading = ref(false)
  const saving = ref(false)
  const actions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedIds = ref<string[]>([])

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })
  const roleTitle = computed(() => props.roleData?.roleName || '')

  watch(
    () => props.modelValue,
    async (open) => {
      if (!open || !props.roleData?.roleId) return
      loading.value = true
      try {
        const [boundary, selected] = await Promise.all([
          fetchGetMyTeamActions(),
          fetchGetMyTeamRoleActions(props.roleData.roleId)
        ])
        actions.value = boundary.actions || []
        selectedIds.value = [...(selected?.action_ids || [])]
      } catch (error: any) {
        ElMessage.error(error?.message || '加载团队角色功能权限失败')
      } finally {
        loading.value = false
      }
    }
  )

  async function handleSave() {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      await fetchSetMyTeamRoleActions(props.roleData.roleId, selectedIds.value)
      ElMessage.success('团队角色功能权限已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队角色功能权限失败')
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
</style>
