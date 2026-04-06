<template>
  <ElDialog
    v-model="visible"
    title="未注册受管页"
    width="980px"
    destroy-on-close
    @close="handleClose"
  >
    <div class="unregistered-toolbar">
      <div class="unregistered-toolbar-summary">
        <span>共 {{ records.length }} 项候选受管页面</span>
        <span class="unregistered-toolbar-note"
          >菜单入口组件已自动排除，这里只显示非菜单直达页候选</span
        >
      </div>
      <div class="unregistered-toolbar-actions">
        <ElInput
          class="unregistered-toolbar-keyword"
          v-model="keyword"
          clearable
          placeholder="搜索组件/标识/路由"
        />
        <ElButton type="primary" :loading="syncing" @click="handleSync">同步受管页</ElButton>
      </div>
    </div>

    <ElTable :data="pagedRecords" height="460px" border>
      <ElTableColumn prop="component" label="组件路径" min-width="220" />
      <ElTableColumn prop="pageKey" label="页面标识" min-width="180" />
      <ElTableColumn prop="name" label="页面名称" min-width="140" />
      <ElTableColumn prop="routePath" label="路由路径" min-width="160" />
      <ElTableColumn prop="parentMenuName" label="推测入口菜单" min-width="120" />
      <ElTableColumn label="操作" width="120" align="center">
        <template #default="{ row }">
          <ElButton type="primary" link @click="handleCreateFromCandidate(row)"
            >创建受管页</ElButton
          >
        </template>
      </ElTableColumn>
    </ElTable>
    <WorkspacePagination
      v-model:current-page="pageState.current"
      v-model:page-size="pageState.size"
      :total="filteredRecords.length"
    />

    <template #footer>
      <ElButton @click="handleClose">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import { fetchGetPageUnregisteredList, fetchSyncPages } from '@/api/system-manage'

  type PageUnregisteredItem = Api.SystemManage.PageUnregisteredItem
  type PageItem = Api.SystemManage.PageItem

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'synced'): void
    (e: 'create-candidate', value: Partial<PageItem>): void
  }

  const props = withDefaults(
    defineProps<{
      modelValue: boolean
      appKey?: string
    }>(),
    {
      modelValue: false
    }
  )

  const emit = defineEmits<Emits>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const syncing = ref(false)
  const records = ref<PageUnregisteredItem[]>([])
  const keyword = ref('')
  const currentAppKey = computed(() => `${props.appKey || ''}`.trim())
  const pageState = reactive({
    current: 1,
    size: 20
  })

  const filteredRecords = computed(() => {
    const key = `${keyword.value || ''}`.trim().toLowerCase()
    if (!key) return records.value
    return records.value.filter((item) => {
      const source = [
        item.component,
        item.pageKey,
        item.name,
        item.routePath,
        item.routeName,
        item.parentMenuName
      ]
        .join(' ')
        .toLowerCase()
      return source.includes(key)
    })
  })

  const pagedRecords = computed(() => {
    const start = (pageState.current - 1) * pageState.size
    return filteredRecords.value.slice(start, start + pageState.size)
  })

  async function loadData() {
    if (loading.value) return
    if (!currentAppKey.value) {
      ElMessage.warning('缺少 app 上下文')
      return
    }
    loading.value = true
    try {
      const res = await fetchGetPageUnregisteredList(currentAppKey.value)
      records.value = res.records || []
      pageState.current = 1
    } catch (error: any) {
      ElMessage.error(error?.message || '加载未注册页面失败')
    } finally {
      loading.value = false
    }
  }

  async function handleSync() {
    if (syncing.value) return
    if (!currentAppKey.value) {
      ElMessage.warning('缺少 app 上下文')
      return
    }
    syncing.value = true
    try {
      const res = await fetchSyncPages(currentAppKey.value)
      ElMessage.success(`同步完成：新增 ${res.createdCount}，跳过 ${res.skippedCount}`)
      emit('synced')
      await loadData()
    } catch (error: any) {
      ElMessage.error(error?.message || '同步失败')
    } finally {
      syncing.value = false
    }
  }

  function handleCreateFromCandidate(row: PageUnregisteredItem) {
    emit('create-candidate', {
      pageKey: row.pageKey,
      name: row.name,
      routeName: row.routeName,
      routePath: row.routePath,
      component: row.component,
      pageType: row.pageType || 'inner',
      moduleKey: row.moduleKey || '',
      parentMenuId: row.parentMenuId || '',
      activeMenuPath: row.activeMenuPath || '',
      source: 'manual',
      accessMode: row.pageType === 'inner' ? 'inherit' : 'jwt'
    })
    visible.value = false
  }

  function handleClose() {
    visible.value = false
  }

  watch(
    () => props.modelValue,
    (value) => {
      if (!value) return
      loadData()
    }
  )

  watch(keyword, () => {
    pageState.current = 1
  })
</script>
<style scoped lang="scss">
  .unregistered-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
  }

  .unregistered-toolbar-summary {
    color: var(--el-text-color-secondary);
    font-size: 13px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .unregistered-toolbar-note {
    font-size: 12px;
  }

  .unregistered-toolbar-actions {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .unregistered-toolbar-keyword {
    width: min(100%, 260px);
  }

  @media (max-width: 768px) {
    .unregistered-toolbar {
      flex-direction: column;
      align-items: stretch;
      gap: 8px;
    }

    .unregistered-toolbar-actions {
      width: 100%;
      justify-content: flex-end;
    }

    .unregistered-toolbar-keyword {
      width: 100%;
      min-width: 0;
    }
  }
</style>
