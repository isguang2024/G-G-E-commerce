<template>
  <ElDrawer
    v-model="visible"
    :title="`协作空间功能包 - ${collaborationWorkspaceName}`"
    size="1280px"
    destroy-on-close
    class="business-dialog config-drawer"
    direction="rtl"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        协作空间功能包由个人空间统一开通。保存后会同步刷新该协作空间的功能权限边界和菜单边界。
      </div>

      <div class="summary-card">
        <ElTag type="primary" effect="light" round>协作空间 {{ collaborationWorkspaceName }}</ElTag>
        <ElTag type="info" effect="light" round>已开通 {{ selectedPackageIds.length }}</ElTag>
        <ElTag type="info" effect="light" round>当前筛选 {{ filteredPackages.length }}</ElTag>
        <ElTag type="info" effect="light" round>全部 {{ packages.length }}</ElTag>
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
      <ElButton type="primary" :loading="saving" @click="handleSave"> 保存 </ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import FeaturePackageGrantPreview from '@/components/business/permission/FeaturePackageGrantPreview.vue'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import {
    fetchGetFeaturePackageOptions,
    fetchGetCollaborationWorkspaceFeaturePackages,
    fetchSetCollaborationWorkspaceFeaturePackages
  } from '@/domains/governance/api'

  interface Props {
    modelValue: boolean
    collaborationWorkspaceId: string
    collaborationWorkspaceName: string
    appKey?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    collaborationWorkspaceId: '',
    collaborationWorkspaceName: ''
  })

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
  const packages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedPackageIds = ref<string[]>([])
  const pagination = ref({
    current: 1,
    size: 10
  })
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())
  const collaborationWorkspaceName = computed(() => props.collaborationWorkspaceName || '')

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
    if (!props.collaborationWorkspaceId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    loading.value = true
    resetFilters()
    try {
      const [listRes, collaborationWorkspaceRes] = await Promise.all([
        fetchGetFeaturePackageOptions({
          workspaceScope: 'all',
          status: 'normal'
        }),
        fetchGetCollaborationWorkspaceFeaturePackages(
          props.collaborationWorkspaceId,
          currentAppKey.value
        )
      ])
      packages.value = listRes?.records || []
      selectedPackageIds.value = [...(collaborationWorkspaceRes?.package_ids || [])]
      pagination.value.current = 1
      if (!selectedPackageIds.value.length) {
        selectionFilter.value = 'all'
      }
    } catch (error: any) {
      ElMessage.error(error?.message || '加载协作空间功能包失败')
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

  async function handleSave() {
    if (!props.collaborationWorkspaceId || !currentAppKey.value) {
      if (!currentAppKey.value) {
        ElMessage.warning('缺少 app 上下文')
      }
      return
    }
    saving.value = true
    try {
      const stats = await fetchSetCollaborationWorkspaceFeaturePackages(
        props.collaborationWorkspaceId,
        selectedPackageIds.value,
        currentAppKey.value
      )
      ElMessage.success(formatRefreshMessage(stats))
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存协作空间功能包失败')
    } finally {
      saving.value = false
    }
  }

  watch([keyword, contextFilter, selectionFilter, statusFilter], () => {
    pagination.value.current = 1
  })

  function formatRefreshMessage(stats?: Api.SystemManage.RefreshStats) {
    return `本次增量刷新：角色 ${stats?.roleCount || 0}、协作空间 ${stats?.collaborationWorkspaceCount || 0}、用户 ${stats?.userCount || 0}、耗时 ${stats?.elapsedMilliseconds || 0} ms`
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
