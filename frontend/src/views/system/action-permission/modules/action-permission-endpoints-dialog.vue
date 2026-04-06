<template>
  <ElDrawer
    v-model="visible"
    :title="`关联接口 - ${permissionName}`"
    size="1080px"
    destroy-on-close
    direction="rtl"
    class="config-drawer"
  >
    <div class="toolbar">
      <ElSelect
        v-model="selectedEndpointCode"
        class="candidate-endpoint-select"
        filterable
        remote
        reserve-keyword
        clearable
        :loading="candidateLoading"
        placeholder="选择要关联的接口"
        :remote-method="handleCandidateSearch"
        @visible-change="handleCandidateVisibleChange"
      >
        <ElOption
          v-for="item in candidateEndpoints"
          :key="item.code"
          :label="item.spec || `${item.method} ${item.path}`"
          :value="item.code"
        >
          <div class="candidate-option">
            <span class="candidate-option-spec">{{
              item.spec || `${item.method} ${item.path}`
            }}</span>
            <span v-if="item.summary" class="candidate-option-summary">{{ item.summary }}</span>
          </div>
        </ElOption>
        <template #footer>
          <div class="candidate-dropdown-footer">
            <span class="candidate-pagination-text">
              共 {{ candidateTotal }} 条，可选第 {{ candidateCurrent }} /
              {{ candidatePageCount }} 页
            </span>
            <ElButton text :disabled="candidateCurrent <= 1" @click="changeCandidatePage(-1)">
              上一页
            </ElButton>
            <ElButton
              text
              :disabled="candidateCurrent >= candidatePageCount || candidatePageCount === 0"
              @click="changeCandidatePage(1)"
            >
              下一页
            </ElButton>
          </div>
        </template>
      </ElSelect>
      <ElButton type="primary" :loading="adding" @click="handleAddEndpoint">新增关联接口</ElButton>
    </div>
    <ArtTable :loading="loading" :data="data" :columns="columns" />

    <template #footer>
      <ElButton @click="visible = false">关闭</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, h, ref, watch } from 'vue'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import {
    fetchAddPermissionActionEndpoint,
    fetchDeletePermissionActionEndpoint,
    fetchGetApiEndpointCategories,
    fetchGetApiEndpointList,
    fetchGetPermissionActionEndpoints
  } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    permissionId: string
    permissionName: string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    permissionId: '',
    permissionName: ''
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const adding = ref(false)
  const data = ref<Api.SystemManage.APIEndpointItem[]>([])
  const categoryNameMap = ref<Record<string, string>>({})
  const candidateEndpoints = ref<Api.SystemManage.APIEndpointItem[]>([])
  const candidateLoading = ref(false)
  const candidateKeyword = ref('')
  const candidateCurrent = ref(1)
  const candidateSize = ref(20)
  const candidateTotal = ref(0)
  const selectedEndpointCode = ref('')
  const candidateInFlightKey = ref('')
  const candidateLastLoadedKey = ref('')
  const candidatePageCount = computed(() =>
    Math.max(1, Math.ceil((candidateTotal.value || 0) / candidateSize.value))
  )

  const columns = [
    {
      prop: 'spec',
      label: '接口规格',
      minWidth: 320,
      showOverflowTooltip: true,
      formatter: (row: Api.SystemManage.APIEndpointItem) => row.spec || `${row.method} ${row.path}`
    },
    {
      prop: 'category',
      label: '分类',
      width: 140,
      formatter: (row: Api.SystemManage.APIEndpointItem) => {
        const categoryId = row.categoryId || ''
        return row.category?.name || (categoryId ? categoryNameMap.value[categoryId] : '') || '-'
      }
    },
    {
      prop: 'authMode',
      label: '鉴权模式',
      width: 110,
      formatter: (row: Api.SystemManage.APIEndpointItem) => {
        const config = {
          public: { type: 'info', text: '公开' },
          jwt: { type: 'warning', text: '仅登录' },
          permission: { type: 'success', text: '功能权限' },
          api_key: { type: 'danger', text: 'API Key' }
        } as const
        const current = config[row.authMode as keyof typeof config] || {
          type: 'info',
          text: row.authMode || '-'
        }
        return h(
          ElTag,
          { type: current.type as 'success' | 'info' | 'warning' | 'danger' },
          () => current.text
        )
      }
    },
    { prop: 'summary', label: '说明', minWidth: 180, showOverflowTooltip: true },
    { prop: 'handler', label: '处理函数', minWidth: 260, showOverflowTooltip: true },
    {
      prop: 'status',
      label: '状态',
      width: 90,
      formatter: (row: Api.SystemManage.APIEndpointItem) =>
        h(ElTag, { type: row.status === 'normal' ? 'success' : 'danger' }, () =>
          row.status === 'normal' ? '正常' : '停用'
        )
    },
    {
      prop: 'operation',
      label: '操作',
      width: 100,
      fixed: 'right' as const,
      formatter: (row: Api.SystemManage.APIEndpointItem) =>
        h(
          ElButton,
          {
            text: true,
            type: 'danger',
            onClick: () => handleRemoveEndpoint(row)
          },
          () => '移除'
        )
    }
  ]

  watch(
    () => [props.modelValue, props.permissionId],
    ([open, permissionId]) => {
      if (open && permissionId) {
        loadData(permissionId as string)
      }
    },
    { immediate: true }
  )

  async function loadData(permissionId: string) {
    loading.value = true
    try {
      const [res, categories] = await Promise.all([
        fetchGetPermissionActionEndpoints(permissionId),
        fetchGetApiEndpointCategories()
      ])
      categoryNameMap.value = (categories.records || []).reduce(
        (acc, item) => {
          if (item.id) acc[item.id] = item.name || item.code || '-'
          return acc
        },
        {} as Record<string, string>
      )
      data.value = res.records || []
    } catch (error: any) {
      ElMessage.error(error?.message || '加载关联接口失败')
    } finally {
      loading.value = false
    }
  }

  async function loadCandidateEndpoints(options?: {
    resetPage?: boolean
    keyword?: string
    force?: boolean
  }) {
    if (options?.resetPage) {
      candidateCurrent.value = 1
    }
    if (typeof options?.keyword === 'string') {
      candidateKeyword.value = options.keyword
      candidateCurrent.value = 1
    }
    const requestKey = `${candidateCurrent.value}|${candidateSize.value}|${candidateKeyword.value || ''}`
    if (!options?.force && requestKey === candidateLastLoadedKey.value) {
      return
    }
    if (candidateLoading.value && requestKey === candidateInFlightKey.value) {
      return
    }
    candidateInFlightKey.value = requestKey
    candidateLoading.value = true
    try {
      const res = await fetchGetApiEndpointList({
        current: candidateCurrent.value,
        size: candidateSize.value,
        status: 'normal',
        keyword: candidateKeyword.value || undefined
      })
      const linked = new Set((data.value || []).map((item) => item.code))
      candidateTotal.value = res.total || 0
      candidateEndpoints.value = (res.records || []).filter(
        (item) => item.code && !linked.has(item.code)
      )
      candidateLastLoadedKey.value = requestKey
    } catch (error: any) {
      ElMessage.error(error?.message || '加载可关联接口失败')
    } finally {
      candidateLoading.value = false
      if (candidateInFlightKey.value === requestKey) {
        candidateInFlightKey.value = ''
      }
    }
  }

  async function handleCandidateSearch(keyword: string) {
    await loadCandidateEndpoints({ keyword: keyword || '' })
  }

  async function handleCandidateVisibleChange(visible: boolean) {
    if (!visible) return
    await loadCandidateEndpoints({ resetPage: true, force: true })
  }

  async function changeCandidatePage(offset: number) {
    const target = candidateCurrent.value + offset
    if (target < 1 || target > candidatePageCount.value) {
      return
    }
    candidateCurrent.value = target
    await loadCandidateEndpoints()
  }

  async function handleAddEndpoint() {
    if (!props.permissionId) return
    if (!selectedEndpointCode.value) {
      ElMessage.warning('请先选择接口')
      return
    }
    adding.value = true
    try {
      await fetchAddPermissionActionEndpoint(props.permissionId, selectedEndpointCode.value)
      ElMessage.success('新增关联接口成功')
      selectedEndpointCode.value = ''
      await loadData(props.permissionId)
      await loadCandidateEndpoints({ force: true })
    } catch (error: any) {
      ElMessage.error(error?.message || '新增关联接口失败')
    } finally {
      adding.value = false
    }
  }

  function handleRemoveEndpoint(row: Api.SystemManage.APIEndpointItem) {
    if (!props.permissionId) return
    if (!row.code) {
      ElMessage.warning('当前接口缺少固定编码，请先重建 API 注册表')
      return
    }
    ElMessageBox.confirm(
      `确定移除接口「${row.spec || `${row.method} ${row.path}`}」吗？`,
      '移除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
      .then(() => fetchDeletePermissionActionEndpoint(props.permissionId, row.code))
      .then(async () => {
        ElMessage.success('移除成功')
        await loadData(props.permissionId)
        await loadCandidateEndpoints({ force: true })
      })
      .catch((error) => {
        if (error !== 'cancel') ElMessage.error(error?.message || '移除失败')
      })
  }
</script>

<style scoped>
  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 12px;
  }

  .candidate-endpoint-select {
    width: min(100%, 520px);
  }

  .candidate-dropdown-footer {
    display: flex;
    align-items: center;
    gap: 8px;
    justify-content: space-between;
    padding: 8px 12px 4px;
    border-top: 1px solid var(--el-border-color-lighter);
    color: var(--el-text-color-secondary);
    font-size: 12px;
  }

  .candidate-pagination-text {
    margin-right: 6px;
  }

  .candidate-option {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .candidate-option-spec {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .candidate-option-summary {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  @media (max-width: 768px) {
    .toolbar {
      flex-wrap: wrap;
    }

    .candidate-endpoint-select {
      width: 100%;
    }
  }
</style>
