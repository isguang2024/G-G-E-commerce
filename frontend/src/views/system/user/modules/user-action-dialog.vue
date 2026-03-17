<template>
  <ElDialog
    v-model="visible"
    :title="`用户功能权限 - ${userData?.nickName || userData?.userName || ''}`"
    width="760px"
    destroy-on-close
  >
    <div v-loading="loading">
      <ElAlert
        type="info"
        :closable="false"
        class="mb-4"
        title="这里配置的是平台级个人覆盖权限。未设置时沿用角色权限；允许或拒绝会覆盖角色结果。"
      />
      <ElEmpty v-if="actions.length === 0" description="暂无已注册的平台级功能权限" />
      <ElTable v-else :data="actions" border max-height="480">
        <ElTableColumn prop="name" label="权限名称" min-width="180" />
        <ElTableColumn prop="resourceCode" label="资源编码" min-width="140" />
        <ElTableColumn prop="actionCode" label="动作编码" min-width="150" />
        <ElTableColumn label="个人覆盖" width="200">
          <template #default="{ row }">
            <ElSelect
              v-model="effectMap[row.id]"
              clearable
              placeholder="继承角色"
              style="width: 150px"
            >
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
    fetchGetUserActions,
    fetchSetUserActions
  } from '@/api/system-manage'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    userData?: Api.SystemManage.UserListItem
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
    if (!props.userData?.id) return
    loading.value = true
    try {
      const [actionList, userActions] = await Promise.all([
        fetchGetPermissionActionList({ current: 1, size: 500, scopeCode: 'global' }),
        fetchGetUserActions(props.userData.id)
      ])
      actions.value = (actionList.records || []).filter((item) => !item.requiresTenantContext)
      Object.keys(effectMap).forEach((key) => delete effectMap[key])
      for (const item of userActions) {
        effectMap[item.actionId] = item.effect
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '获取用户功能权限失败')
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
    if (!props.userData?.id) return
    submitting.value = true
    try {
      const payload = Object.entries(effectMap)
        .filter(([, effect]) => effect)
        .map(([actionId, effect]) => ({
          action_id: actionId,
          effect: effect as 'allow' | 'deny'
        }))
      await fetchSetUserActions(props.userData.id, payload)
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
