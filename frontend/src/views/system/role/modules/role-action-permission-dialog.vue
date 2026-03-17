<template>
  <ElDialog
    v-model="visible"
    :title="`功能权限 - ${roleData?.roleName || ''}`"
    width="760px"
    destroy-on-close
  >
    <div v-loading="loading">
      <ElEmpty v-if="actions.length === 0" description="暂无已注册功能权限" />
      <ElTable v-else :data="actions" border max-height="480">
        <ElTableColumn prop="name" label="权限名称" min-width="180" />
        <ElTableColumn prop="resourceCode" label="资源编码" min-width="140" />
        <ElTableColumn prop="actionCode" label="动作编码" min-width="140" />
        <ElTableColumn label="作用域" width="90">
          <template #default="{ row }">
            <ElTag :type="row.scopeCode === 'team' ? 'success' : 'primary'">
              {{ row.scopeName || (row.scopeCode === 'team' ? '团队' : '平台') }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="授权效果" width="200">
          <template #default="{ row }">
            <ElSelect v-model="effectMap[row.id]" style="width: 140px" clearable>
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
    fetchGetPermissionActionList,
    fetchGetRoleActions,
    fetchSetRoleActions
  } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
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
  const effectMap = reactive<Record<string, 'allow' | 'deny' | ''>>({})

  async function loadData() {
    if (!props.roleData?.roleId) return
    loading.value = true
    try {
      const [actionList, roleActionRes] = await Promise.all([
        fetchGetPermissionActionList({
          current: 1,
          size: 500,
          scopeCode: props.roleData.scopeCode || props.roleData.scope
        }),
        fetchGetRoleActions(props.roleData.roleId)
      ])
      actions.value = actionList.records || []
      Object.keys(effectMap).forEach((key) => delete effectMap[key])
      for (const item of roleActionRes?.actions || []) {
        effectMap[item.action_id] = item.effect
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '获取角色功能权限失败')
      visible.value = false
    } finally {
      loading.value = false
    }
  }

  watch(
    () => visible.value,
    (opened) => {
      if (opened) loadData()
    }
  )

  async function handleSubmit() {
    if (!props.roleData?.roleId) return
    submitting.value = true
    try {
      const payload = Object.entries(effectMap)
        .filter(([, effect]) => effect)
        .map(([actionId, effect]) => ({
          action_id: actionId,
          effect: effect as 'allow' | 'deny'
        }))
      await fetchSetRoleActions(props.roleData.roleId, payload)
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
