<template>
  <ElDialog v-model="visible" :title="`角色功能包 - ${roleTitle}`" width="920px" destroy-on-close>
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        功能包是平台角色的主开通入口。角色绑定功能包后，包内菜单和权限默认生效；后续菜单可见性和权限裁剪都应在已绑定功能包范围内进行。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>角色 {{ roleTitle }}</ElTag>
        <ElTag type="warning" effect="plain" round>平台上下文</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ selectedPackageIds.length }}</ElTag>
      </div>

      <ElInput
        v-model="keyword"
        clearable
        placeholder="搜索功能包名称、编码或说明"
        class="toolbar-search"
      />

      <ElTable :data="filteredPackages" border max-height="420">
        <ElTableColumn width="60">
          <template #default="{ row }">
            <ElCheckbox
              :model-value="selectedPackageIds.includes(row.id)"
              @change="toggleSelection(row.id, $event)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="packageKey" label="功能包编码" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="name" label="功能包名称" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="上下文" width="120">
          <template #default="{ row }">
            <ElTag :type="row.contextType === 'team' ? 'success' : 'warning'" effect="plain">
              {{ formatContext(row.contextType) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="100">
          <template #default="{ row }">
            <ElTag :type="row.status === 'normal' ? 'success' : 'info'" effect="plain">
              {{ row.status === 'normal' ? '正常' : '停用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="说明" min-width="240" show-overflow-tooltip />
      </ElTable>
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
  import {
    fetchGetFeaturePackageList,
    fetchGetRolePackages,
    fetchSetRolePackages
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
  const saving = ref(false)
  const keyword = ref('')
  const packages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedPackageIds = ref<string[]>([])

  const roleTitle = computed(() => props.roleData?.roleName || '')

  const filteredPackages = computed(() => {
    const currentKeyword = keyword.value.trim().toLowerCase()
    if (!currentKeyword) return packages.value
    return packages.value.filter((item) =>
      [item.packageKey, item.name, item.description]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(currentKeyword)
    )
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) {
        loadData()
      }
    }
  )

  async function loadData() {
    if (!props.roleData?.roleId) return
    loading.value = true
    try {
      const [listRes, roleRes] = await Promise.all([
        fetchGetFeaturePackageList({ current: 1, size: 1000, contextType: 'platform' }),
        fetchGetRolePackages(props.roleData.roleId)
      ])
      packages.value = listRes?.records || []
      selectedPackageIds.value = [...(roleRes?.package_ids || [])]
    } catch (error: any) {
      ElMessage.error(error?.message || '加载角色功能包失败')
    } finally {
      loading.value = false
    }
  }

  function toggleSelection(packageId: string, checked: boolean | string | number) {
    if (checked) {
      if (!selectedPackageIds.value.includes(packageId)) {
        selectedPackageIds.value = [...selectedPackageIds.value, packageId]
      }
      return
    }
    selectedPackageIds.value = selectedPackageIds.value.filter((item) => item !== packageId)
  }

  function formatContext(contextType?: string) {
    if (contextType === 'platform,team') return '平台/团队'
    if (contextType === 'team') return '团队'
    return '平台'
  }

  async function handleSave() {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      await fetchSetRolePackages(props.roleData.roleId, selectedPackageIds.value)
      ElMessage.success('角色功能包已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存角色功能包失败')
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

  .summary-card {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .toolbar-search {
    width: 320px;
  }
</style>
