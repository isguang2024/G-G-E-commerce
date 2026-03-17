<template>
  <ElDialog
    v-model="visible"
    :title="`数据权限 - ${roleData?.roleName || ''}`"
    width="760px"
    destroy-on-close
  >
    <div v-loading="loading">
      <ElAlert
        type="info"
        :closable="false"
        class="mb-4"
        title="数据权限用于控制某个资源能看到的数据范围；未配置的资源默认不额外放大数据访问范围。"
      />
      <ElEmpty v-if="resources.length === 0" description="当前作用域下暂无可配置资源" />
      <ElTable v-else :data="resources" border max-height="480">
        <ElTableColumn prop="resourceName" label="资源名称" min-width="160" />
        <ElTableColumn prop="resourceCode" label="资源编码" min-width="180" />
        <ElTableColumn label="数据范围" min-width="220">
          <template #default="{ row }">
            <ElSelect v-model="scopeMap[row.resourceCode]" clearable placeholder="未配置" style="width: 180px">
              <ElOption
                v-for="option in availableScopes"
                :key="option.scopeCode"
                :label="option.scopeName"
                :value="option.scopeCode"
              />
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
  import { ElMessage } from 'element-plus'
  import {
    fetchGetRoleDataPermissions,
    fetchSetRoleDataPermissions
  } from '@/api/system-manage'

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
  const resources = ref<Api.SystemManage.RoleDataPermissionResourceItem[]>([])
  const availableScopes = ref<Api.SystemManage.RoleDataPermissionScopeOption[]>([])
  const scopeMap = reactive<Record<string, string>>({})

  async function loadData() {
    if (!props.roleData?.roleId) return
    loading.value = true
    try {
      const res = await fetchGetRoleDataPermissions(props.roleData.roleId)
      resources.value = (res?.resources || []).map((item) => ({
        resourceCode: item.resource_code || '',
        resourceName: item.resource_name || item.resource_code || ''
      }))
      availableScopes.value = (res?.available_scopes || []).map((item) => ({
        scopeCode: item.scope_code || '',
        scopeName: item.scope_name || item.scope_code || ''
      }))
      Object.keys(scopeMap).forEach((key) => delete scopeMap[key])
      for (const item of res?.permissions || []) {
        if (item?.resource_code) {
          scopeMap[item.resource_code] = item.scope_code || ''
        }
      }
    } catch (e: any) {
      ElMessage.error(e?.message || '获取角色数据权限失败')
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
      const permissions = Object.entries(scopeMap)
        .filter(([, scopeCode]) => scopeCode)
        .map(([resourceCode, scopeCode]) => ({
          resource_code: resourceCode,
          scope_code: scopeCode
        }))
      await fetchSetRoleDataPermissions(props.roleData.roleId, permissions)
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
