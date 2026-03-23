<template>
  <div class="feature-package-page art-full-height">
    <ArtSearchBar
      v-show="showSearchBar"
      v-model="searchForm"
      :items="searchItems"
      :showExpand="false"
      @search="handleSearch"
      @reset="handleReset"
    />

    <ElCard class="art-table-card" shadow="never" :style="{ marginTop: showSearchBar ? '12px' : '0' }">
      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="handleRefresh"
      >
        <template #left>
          <ElButton v-action="'platform.package.manage'" type="primary" @click="openDialog('add')" v-ripple>
            新增功能包
          </ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <FeaturePackageDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :package-data="currentPackage"
      @success="handleRefresh"
    />

    <FeaturePackageActionsDialog
      v-model="actionsDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'team'"
      @success="handleRefresh"
    />

    <FeaturePackageTeamsDialog
      v-model="teamsDialogVisible"
      :package-id="currentPackage.id || ''"
      :package-name="currentPackage.name || ''"
      :context-type="currentPackage.contextType || 'team'"
      @success="handleRefresh"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, reactive, ref } from 'vue'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteFeaturePackage, fetchGetFeaturePackageList } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'
  import FeaturePackageDialog from './modules/feature-package-dialog.vue'
  import FeaturePackageActionsDialog from './modules/feature-package-actions-dialog.vue'
  import FeaturePackageTeamsDialog from './modules/feature-package-teams-dialog.vue'

  defineOptions({ name: 'FeaturePackage' })

  type PackageItem = Api.SystemManage.FeaturePackageItem

  type SearchForm = {
    keyword: string
    packageKey: string
    name: string
    contextType: string
    status: string
  }
  const showSearchBar = ref(true)
  const dialogVisible = ref(false)
  const actionsDialogVisible = ref(false)
  const teamsDialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentPackage = ref<Partial<PackageItem>>({})

  const searchForm = reactive<SearchForm>({
    keyword: '',
    packageKey: '',
    name: '',
    contextType: '',
    status: ''
  })

  const contextTypeOptions = [
    { label: '全部上下文', value: '' },
    { label: '平台功能包', value: 'platform' },
    { label: '团队功能包', value: 'team' }
  ]

  const statusOptions = [
    { label: '全部状态', value: '' },
    { label: '正常', value: 'normal' },
    { label: '停用', value: 'disabled' }
  ]

  const searchItems = computed<FormItem[]>(() => [
    { label: '关键词', key: 'keyword', type: 'input', props: { placeholder: '名称/编码/描述' } },
    { label: '功能包编码', key: 'packageKey', type: 'input', props: { placeholder: '请输入功能包编码' } },
    { label: '功能包名称', key: 'name', type: 'input', props: { placeholder: '请输入功能包名称' } },
    {
      label: '上下文类型',
      key: 'contextType',
      type: 'select',
      props: { options: contextTypeOptions, clearable: true }
    },
    { label: '状态', key: 'status', type: 'select', props: { options: statusOptions, clearable: true } }
  ])

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetFeaturePackageList,
      apiParams: {
        current: 1,
        size: 20
      },
      columnsFactory: () => [
        { prop: 'packageKey', label: '功能包编码', minWidth: 220, showOverflowTooltip: true },
        { prop: 'name', label: '功能包名称', minWidth: 180, showOverflowTooltip: true },
        {
          prop: 'contextType',
          label: '上下文',
          width: 110,
          formatter: (row: PackageItem) =>
            h(ElTag, { type: row.contextType === 'platform' ? 'success' : 'info' }, () =>
              row.contextType === 'platform' ? '平台' : '团队'
            )
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 220,
          showOverflowTooltip: true,
          formatter: (row: PackageItem) => row.description || '-'
        },
        { prop: 'actionCount', label: '权限数', width: 90, formatter: (row: PackageItem) => row.actionCount ?? 0 },
        { prop: 'teamCount', label: '团队数', width: 90, formatter: (row: PackageItem) => row.teamCount ?? 0 },
        { prop: 'sortOrder', label: '排序', width: 80, formatter: (row: PackageItem) => row.sortOrder ?? 0 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PackageItem) =>
            h(ElTag, { type: row.status === 'normal' ? 'success' : 'warning' }, () =>
              row.status === 'normal' ? '正常' : '停用'
            )
        },
        { prop: 'updatedAt', label: '更新时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 140,
          fixed: 'right',
          formatter: (row: PackageItem) => {
            const list: ButtonMoreItem[] = [
              { key: 'actions', label: '配置权限', icon: 'ri:key-2-line', auth: 'platform.package.manage' },
              {
                key: 'teams',
                label: '开通团队',
                icon: 'ri:team-line',
                auth: 'platform.package.assign',
                disabled: row.contextType !== 'team'
              },
              { key: 'edit', label: '编辑', icon: 'ri:edit-2-line', auth: 'platform.package.manage' },
              { key: 'delete', label: '删除', icon: 'ri:delete-bin-4-line', auth: 'platform.package.manage' }
            ]
            return h(ArtButtonMore, {
              list,
              onClick: (item: ButtonMoreItem) => handleAction(item.key as string, row)
            })
          }
        }
      ]
    }
  })

  function normalizeSearchParams() {
    return {
      keyword: searchForm.keyword.trim() || undefined,
      package_key: searchForm.packageKey.trim() || undefined,
      name: searchForm.name.trim() || undefined,
      context_type: searchForm.contextType || undefined,
      status: searchForm.status || undefined
    }
  }

  async function handleSearch() {
    Object.assign(searchParams, normalizeSearchParams())
    await getData()
  }

  async function handleReset() {
    Object.assign(searchForm, {
      keyword: '',
      packageKey: '',
      name: '',
      contextType: '',
      status: ''
    })
    await resetSearchParams()
  }

  async function handleRefresh() {
    await refreshData()
  }

  function openDialog(type: 'add' | 'edit', row?: PackageItem) {
    dialogType.value = type
    currentPackage.value = row ? { ...row } : {}
    dialogVisible.value = true
  }

  function handleAction(command: string, row: PackageItem) {
    if (command === 'actions') {
      currentPackage.value = { ...row }
      actionsDialogVisible.value = true
      return
    }
    if (command === 'teams') {
      currentPackage.value = { ...row }
      teamsDialogVisible.value = true
      return
    }
    if (command === 'edit') {
      openDialog('edit', row)
      return
    }
    if (command === 'delete') {
      ElMessageBox.confirm(`确定删除功能包「${row.name}」吗？`, '删除确认', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
        .then(() =>
          fetchDeleteFeaturePackage(row.id)
        )
        .then(() => {
          ElMessage.success('删除成功')
          handleRefresh()
        })
        .catch((e) => {
          if (e !== 'cancel') ElMessage.error(e?.message || '删除失败')
        })
    }
  }
</script>
