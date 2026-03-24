<template>
  <ElDialog
    v-model="visible"
    :title="`关联接口 - ${permissionName}`"
    width="1080px"
    destroy-on-close
  >
    <ArtTable :loading="loading" :data="data" :columns="columns" />

    <template #footer>
      <ElButton @click="visible = false">关闭</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, h, ref, watch } from 'vue'
  import { ElMessage, ElTag } from 'element-plus'
  import { fetchGetPermissionActionEndpoints } from '@/api/system-manage'

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
  const data = ref<Api.SystemManage.APIEndpointItem[]>([])

  const columns = [
    {
      prop: 'spec',
      label: '接口规格',
      minWidth: 320,
      showOverflowTooltip: true,
      formatter: (row: Api.SystemManage.APIEndpointItem) => row.spec || `${row.method} ${row.path}`
    },
    { prop: 'module', label: '模块', width: 120 },
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
        const current = config[row.authMode as keyof typeof config] || { type: 'info', text: row.authMode || '-' }
        return h(ElTag, { type: current.type as 'success' | 'info' | 'warning' | 'danger' }, () => current.text)
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
      const res = await fetchGetPermissionActionEndpoints(permissionId)
      data.value = res.records || []
    } catch (error: any) {
      ElMessage.error(error?.message || '加载关联接口失败')
    } finally {
      loading.value = false
    }
  }
</script>
