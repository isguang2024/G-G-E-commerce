<template>
  <ElDrawer
    v-model="visible"
    :title="`配置协作空间内部角色 - ${member?.userName || ''}`"
    size="600px"
    destroy-on-close
    direction="rtl"
    class="config-drawer"
  >
    <div v-loading="loading" class="role-dialog-content">
      <div class="mb-4 text-gray-500 text-sm">
        这里配置的是该成员在当前协作空间内生效的权限角色，不是成员身份。基础协作空间角色由系统提供，协作空间自定义角色仅在当前协作空间内生效。
      </div>

      <div class="summary-row">
        <ElTag effect="plain" round>总计 {{ allRoles.length }}</ElTag>
        <ElTag type="info" effect="plain" round>基础角色 {{ globalRoleCount }}</ElTag>
        <ElTag type="success" effect="plain" round>协作空间自定义 {{ customRoleCount }}</ElTag>
        <ElTag type="warning" effect="plain" round>已选 {{ selectedRoleIds.length }}</ElTag>
      </div>

      <ElCheckboxGroup v-model="selectedRoleIds" class="flex flex-col gap-2">
        <div v-for="role in allRoles" :key="role.roleId" class="role-item">
          <ElCheckbox
            :value="role.roleId"
            :disabled="isCollaborationWorkspaceAdminRole(role.roleCode)"
          >
            <div class="flex items-center gap-2">
              <span class="font-medium">{{ role.roleName }}</span>
              <ElTag :type="role.isGlobal ? 'info' : 'success'" size="small">
                {{ role.isGlobal ? '基础角色' : '协作空间自定义' }}
              </ElTag>
            </div>
            <div v-if="role.description" class="text-xs text-gray-400 mt-1 pl-6">
              {{ role.description }}
            </div>
          </ElCheckbox>
        </div>
      </ElCheckboxGroup>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit"> 保存更改 </ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import {
    fetchGetMyCollaborationMemberRoles,
    fetchSetMyCollaborationMemberRoles,
    fetchGetMyCollaborationRoles
  } from '@/api/collaboration'
  import { ElMessage } from 'element-plus'

  interface Props {
    member: Api.SystemManage.CollaborationWorkspaceMemberItem | null
  }

  const props = defineProps<Props>()
  const emit = defineEmits(['success'])

  const visible = ref(false)
  const loading = ref(false)
  const submitting = ref(false)
  const allRoles = ref<Api.SystemManage.RoleListItem[]>([])
  const selectedRoleIds = ref<string[]>([])
  const memberRoleCodes = ref<string[]>([])
  const globalRoleCount = computed(() => allRoles.value.filter((role) => role.isGlobal).length)
  const customRoleCount = computed(() => allRoles.value.filter((role) => !role.isGlobal).length)

  function isCollaborationWorkspaceAdminRole(roleCode: string): boolean {
    // 如果成员已经是协作空间管理员，则禁用 collaboration_admin 角色的选择
    // 协作空间管理员角色不能被移除
    return (
      memberRoleCodes.value.includes('collaboration_admin') &&
      roleCode === 'collaboration_admin'
    )
  }

  async function open() {
    if (!props.member) return
    visible.value = true
    loading.value = true
    try {
      // 1. 获取所有可选角色（全局+本协作空间）
      const rolesRes = await fetchGetMyCollaborationRoles()
      allRoles.value = [...(rolesRes || [])].sort((left, right) => {
        if (left.isGlobal === right.isGlobal)
          return left.roleName.localeCompare(right.roleName, 'zh-CN')
        return left.isGlobal ? -1 : 1
      })

      // 2. 获取该成员当前已有的角色
      const memberRolesRes = await fetchGetMyCollaborationMemberRoles(props.member.userId)
      selectedRoleIds.value = memberRolesRes.role_ids || []

      // 3. 获取成员的角色编码，用于判断是否为管理员
      memberRoleCodes.value = props.member.roleCode ? [props.member.roleCode] : []
    } catch (e: any) {
      ElMessage.error(e?.message || '获取角色信息失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

  async function handleSubmit() {
    if (!props.member) return
    submitting.value = true
    try {
      await fetchSetMyCollaborationMemberRoles(props.member.userId, selectedRoleIds.value)
      ElMessage.success('协作空间内部角色已更新')
      emit('success')
      visible.value = false
    } catch (e: any) {
      ElMessage.error(e?.message || '更新协作空间内部角色失败')
    } finally {
      submitting.value = false
    }
  }

  defineExpose({ open })
</script>

<style scoped>
  .role-dialog-content {
    max-height: 400px;
    overflow-y: auto;
  }
  .role-item {
    padding: 8px;
    border-radius: 4px;
    transition: background 0.2s;
  }
  .summary-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }
  .role-item:hover {
    background-color: var(--el-fill-color-light);
  }
  :deep(.el-checkbox) {
    height: auto;
    display: flex;
    align-items: flex-start;
  }
  :deep(.el-checkbox__label) {
    padding-top: 2px;
  }
</style>
