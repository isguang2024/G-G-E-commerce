<template>
  <ElDrawer
    v-model="visible"
    :title="`协作空间角色功能包 - ${roleTitle}`"
    size="1280px"
    destroy-on-close
    class="business-dialog config-drawer"
    direction="rtl"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        {{
          props.roleData?.isGlobal
            ? '基础协作空间角色默认继承当前协作空间已开通的功能包，这里展示完整功能包列表和继承结果。'
            : '角色功能包是当前协作空间角色的主配置入口。后续菜单权限和角色权限都会在这里已绑定的功能包范围内继续配置。'
        }}
      </div>

      <div class="summary-card">
        <ElTag type="primary" effect="light" round>角色 {{ roleTitle }}</ElTag>
        <ElTag type="info" effect="light" round>
          {{ props.roleData?.isGlobal ? '基础角色' : '协作空间自定义' }}
        </ElTag>
        <ElTag type="info" effect="light" round>
          {{ inherited ? '继承协作空间功能包' : '角色独立功能包' }}
        </ElTag>
        <ElTag type="info" effect="light" round>已选 {{ selectedPackageIds.length }}</ElTag>
        <ElTag type="info" effect="light" round>当前筛选 {{ filteredPackages.length }}</ElTag>
      </div>

      <div class="toolbar-row">
        <ElInput
          v-model="keyword"
          clearable
          placeholder="搜索功能包名称、编码或说明"
          class="toolbar-search"
        />
        <ElSelect v-model="contextFilter" class="toolbar-select">
          <ElOption label="全部适用空间" value="" />
          <ElOption label="个人空间" value="personal" />
          <ElOption label="协作空间" value="collaboration" />
          <ElOption label="通用" value="common" />
        </ElSelect>
        <ElSelect v-model="selectionFilter" class="toolbar-select">
          <ElOption label="全部" value="all" />
          <ElOption label="已开通" value="selected" />
          <ElOption label="未开通" value="unselected" />
        </ElSelect>
        <ElSelect v-model="statusFilter" class="toolbar-select">
          <ElOption label="全部状态" value="" />
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="disabled" />
        </ElSelect>
      </div>

      <ElTable :data="pagedPackages" border max-height="520" row-key="id">
        <ElTableColumn type="expand" width="56">
          <template #default="{ row }">
            <div class="expand-panel">
              <FeaturePackageGrantPreview
                :package-id="row.id"
                :package-item="row"
                :packages="packages"
              />
            </div>
          </template>
        </ElTableColumn>
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
        <ElTableColumn prop="name" label="功能包名称" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="功能包种类" width="120">
          <template #default="{ row }">
            <ElTag :type="getPackageTypeTagType(row.packageType)" effect="light" round>
              {{ formatPackageType(row.packageType) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="上下文" width="120">
          <template #default="{ row }">
            <ElTag :type="getContextTagType(row.workspaceScope)" effect="light" round>
              {{ formatContext(row.workspaceScope) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="100">
          <template #default="{ row }">
            <ElTag :type="getStatusTagType(row.status)" effect="light" round>
              {{ row.status === 'normal' ? '正常' : '停用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="说明" min-width="280" show-overflow-tooltip />
      </ElTable>
      <WorkspacePagination
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.size"
        :total="filteredPackages.length"
        compact
      />
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton
        v-if="!props.roleData?.isGlobal"
        type="primary"
        :loading="saving"
        @click="handleSave"
      >
        保存
      </ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import FeaturePackageGrantPreview from '@/components/business/permission/FeaturePackageGrantPreview.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import {
    fetchGetMyCollaborationWorkspaceBoundaryPackages,
    fetchGetMyCollaborationWorkspaceBoundaryRolePackages,
    fetchSetMyCollaborationWorkspaceBoundaryRolePackages
  } from '@/api/collaboration-workspace'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
    appKey?: string
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
  const contextFilter = ref('')
  const selectionFilter = ref<'all' | 'selected' | 'unselected'>('selected')
  const statusFilter = ref('normal')
  const inherited = ref(false)
  const packages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedPackageIds = ref<string[]>([])
  const pagination = ref({
    current: 1,
    size: 10
  })
  const roleTitle = computed(() => props.roleData?.roleName || '')
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())

  const filteredPackages = computed(() => {
    const currentKeyword = keyword.value.trim().toLowerCase()
    return packages.value.filter((item) => {
      if (
        currentKeyword &&
        ![item.packageKey, item.name, item.description]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
          .includes(currentKeyword)
      ) {
        return false
      }

      if (statusFilter.value && item.status !== statusFilter.value) {
        return false
      }
      if (contextFilter.value && item.workspaceScope !== contextFilter.value) {
        return false
      }

      const isSelected = selectedPackageIds.value.includes(item.id)
      if (selectionFilter.value === 'selected' && !isSelected) return false
      if (selectionFilter.value === 'unselected' && isSelected) return false
      return true
    })
  })

  const pagedPackages = computed(() => {
    const start = (pagination.value.current - 1) * pagination.value.size
    return filteredPackages.value.slice(start, start + pagination.value.size)
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) {
        void loadData()
      } else {
        resetFilters()
      }
    }
  )

  async function loadData() {
    if (!props.roleData?.roleId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    loading.value = true
    resetFilters()
    try {
      const [listRes, roleRes] = await Promise.all([
        fetchGetMyCollaborationWorkspaceBoundaryPackages(currentAppKey.value),
        fetchGetMyCollaborationWorkspaceBoundaryRolePackages(
          props.roleData.roleId,
          currentAppKey.value
        )
      ])
      packages.value = listRes?.packages || []
      selectedPackageIds.value = [...(roleRes?.package_ids || [])]
      inherited.value = Boolean(roleRes?.inherited)
      pagination.value.current = 1
      if (!selectedPackageIds.value.length) {
        selectionFilter.value = 'all'
      }
    } catch (error: any) {
      ElMessage.error(error?.message || '加载协作空间角色功能包失败')
    } finally {
      loading.value = false
    }
  }

  function resetFilters() {
    keyword.value = ''
    contextFilter.value = ''
    selectionFilter.value = 'selected'
    statusFilter.value = 'normal'
    pagination.value.current = 1
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

  function formatPackageType(packageType?: string) {
    return packageType === 'bundle' ? '组合包' : '基础包'
  }

  function getPackageTypeTagType(packageType?: string) {
    return packageType === 'bundle' ? 'warning' : 'primary'
  }

  function formatContext(workspaceScope?: string) {
    if (workspaceScope === 'all' || workspaceScope === 'common') return '通用'
    if (workspaceScope === 'personal') return '个人空间'
    return '协作空间'
  }

  function getContextTagType(workspaceScope?: string) {
    if (workspaceScope === 'all' || workspaceScope === 'common') return 'primary'
    if (workspaceScope === 'personal') return 'success'
    return 'warning'
  }

  function getStatusTagType(status?: string) {
    if (status === 'normal') return 'success'
    if (status === 'disabled') return 'info'
    return 'danger'
  }

  watch([keyword, contextFilter, selectionFilter, statusFilter], () => {
    pagination.value.current = 1
  })

  async function handleSave() {
    if (!props.roleData?.roleId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    saving.value = true
    try {
      await fetchSetMyCollaborationWorkspaceBoundaryRolePackages(
        props.roleData.roleId,
        selectedPackageIds.value,
        currentAppKey.value
      )
      ElMessage.success('协作空间角色功能包已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存协作空间角色功能包失败')
    } finally {
      saving.value = false
    }
  }
</script>

<style scoped lang="scss">
  :deep(.business-dialog .el-dialog) {
    border-radius: 16px;
    border: 1px solid var(--default-border);
    background: var(--default-box-color);
    box-shadow: 0 12px 32px rgba(0, 0, 0, 0.08);
  }

  :deep(.business-dialog .el-dialog__header) {
    margin-right: 0;
    padding: 22px 24px 12px;
    border-bottom: 1px solid var(--default-border);
  }

  :deep(.business-dialog .el-dialog__title) {
    color: var(--art-gray-900);
    font-size: 20px;
    font-weight: 600;
  }

  :deep(.business-dialog .el-dialog__body) {
    padding: 22px 24px 18px;
  }

  :deep(.business-dialog .el-dialog__footer) {
    padding: 14px 24px 22px;
    border-top: 1px solid var(--default-border);
  }

  :deep(.business-dialog .el-table) {
    --el-table-border-color: var(--default-border);
    --el-table-header-bg-color: var(--default-bg-color);
    --el-table-row-hover-bg-color: var(--art-hover-color);
    border-radius: 14px;
    overflow: hidden;
  }

  :deep(.business-dialog .el-table th.el-table__cell) {
    color: var(--art-gray-700);
    font-weight: 600;
  }

  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 18px;
  }

  .dialog-note {
    color: var(--art-gray-700);
    line-height: 1.75;
  }

  .summary-card {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .toolbar-row {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
  }

  .toolbar-search {
    width: 360px;
    max-width: 100%;
  }

  .toolbar-select {
    width: 160px;
  }

  .expand-panel {
    padding: 16px 18px;
    background: var(--default-bg-color);
    border-top: 1px solid var(--default-border);
  }

  @media (max-width: 960px) {
    .toolbar-row {
      flex-direction: column;
    }

    .toolbar-search,
    .toolbar-select {
      width: 100%;
    }
  }
</style>
