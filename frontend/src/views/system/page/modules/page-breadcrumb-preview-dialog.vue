<template>
  <ElDialog
    v-model="visible"
    title="面包屑预览"
    width="560px"
    destroy-on-close
    @close="handleClose"
  >
    <div v-loading="loading" class="breadcrumb-preview-body">
      <ElEmpty v-if="!items.length" description="暂无可预览数据" />
      <template v-else>
        <div
          v-for="(item, index) in items"
          :key="`${item.type}-${item.pageKey || item.path || index}`"
          class="breadcrumb-preview-item"
        >
          <ElTag size="small" :type="item.type === 'menu' ? 'primary' : 'success'" effect="plain">
            {{ item.type === 'menu' ? '菜单' : '页面' }}
          </ElTag>
          <span class="breadcrumb-preview-title">{{ item.title || '-' }}</span>
          <span class="breadcrumb-preview-path">{{ item.path || '-' }}</span>
          <span v-if="index < items.length - 1" class="breadcrumb-preview-separator">/</span>
        </div>
      </template>
    </div>
    <template #footer>
      <ElButton @click="handleClose">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { fetchGetPageBreadcrumbPreview } from '@/domains/governance/api'

  type PageBreadcrumbPreviewItem = Api.SystemManage.PageBreadcrumbPreviewItem

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
  }

  const props = withDefaults(
    defineProps<{
      modelValue: boolean
      pageId?: string
    }>(),
    {
      modelValue: false,
      pageId: ''
    }
  )

  const emit = defineEmits<Emits>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const loading = ref(false)
  const items = ref<PageBreadcrumbPreviewItem[]>([])

  async function loadData() {
    const pageId = `${props.pageId || ''}`.trim()
    if (!pageId) {
      items.value = []
      return
    }
    loading.value = true
    try {
      const res = await fetchGetPageBreadcrumbPreview(pageId)
      items.value = res.items || []
    } catch (error: any) {
      items.value = []
      ElMessage.error(error?.message || '加载面包屑预览失败')
    } finally {
      loading.value = false
    }
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
</script>

<style scoped lang="scss">
  .breadcrumb-preview-body {
    min-height: 120px;
  }

  .breadcrumb-preview-item {
    display: flex;
    align-items: center;
    gap: 8px;
    min-height: 34px;
  }

  .breadcrumb-preview-title {
    font-size: 14px;
    color: var(--el-text-color-primary);
  }

  .breadcrumb-preview-path {
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .breadcrumb-preview-separator {
    margin-left: 2px;
    color: var(--el-text-color-secondary);
  }
</style>
