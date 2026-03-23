<template>
  <ElDialog v-model="visible" :title="`团队角色功能包 - ${roleTitle}`" width="920px" destroy-on-close>
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        {{
          props.roleData?.isGlobal
            ? '基础团队角色默认继承当前团队已开通功能包，这里只读查看继承结果。'
            : '角色功能包是当前团队角色的主配置入口。后续菜单权限和角色权限都只能在这里已绑定的功能包范围内配置。'
        }}
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>角色 {{ roleTitle }}</ElTag>
        <ElTag :type="props.roleData?.isGlobal ? 'info' : 'success'" effect="plain" round>
          {{ props.roleData?.isGlobal ? '基础角色' : '团队自定义' }}
        </ElTag>
        <ElTag type="primary" effect="plain" round>{{ inherited ? '继承团队功能包' : '角色独立功能包' }}</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ selectedPackageIds.length }}</ElTag>
      </div>

      <ElInput v-model="keyword" clearable placeholder="搜索功能包名称或编码" class="toolbar-search" />

      <ElTable :data="filteredPackages" border max-height="420">
        <ElTableColumn width="60">
          <template #default="{ row }">
            <ElCheckbox
              :model-value="selectedPackageIds.includes(row.id)"
              :disabled="Boolean(props.roleData?.isGlobal)"
              @change="toggleSelection(row.id, $event)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="packageKey" label="功能包编码" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="name" label="功能包名称" min-width="160" show-overflow-tooltip />
        <ElTableColumn label="上下文" width="100">
          <template #default="{ row }">
            <ElTag :type="row.contextType === 'platform' ? 'warning' : 'success'">
              {{ row.contextType === 'platform' ? '平台' : '团队' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="描述" min-width="220" show-overflow-tooltip />
      </ElTable>
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
  import { fetchGetFeaturePackageList } from '@/api/system-manage'
  import { fetchGetMyTeamRolePackages, fetchSetMyTeamRolePackages } from '@/api/team'

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
  const inherited = ref(false)
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
        fetchGetFeaturePackageList({ current: 1, size: 1000, contextType: 'team', status: 'normal' }),
        fetchGetMyTeamRolePackages(props.roleData.roleId)
      ])
      const allowedIds = new Set(roleRes?.packages?.map((item) => item.id) || [])
      packages.value = (listRes?.records || []).filter((item) => allowedIds.has(item.id) || !props.roleData?.isGlobal)
      selectedPackageIds.value = [...(roleRes?.package_ids || [])]
      inherited.value = Boolean(roleRes?.inherited)
    } catch (error: any) {
      ElMessage.error(error?.message || '加载团队角色功能包失败')
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

  async function handleSave() {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      await fetchSetMyTeamRolePackages(props.roleData.roleId, selectedPackageIds.value)
      ElMessage.success('团队角色功能包已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队角色功能包失败')
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
