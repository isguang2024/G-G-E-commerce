<template>
  <ElDialog v-model="visible" title="作用域管理" width="60%" align-center @close="handleClose">
    <ElCard shadow="never">
      <div class="mb-4 flex justify-end">
        <ElButton type="primary" @click="showDialog('add')">新增作用域</ElButton>
      </div>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <!-- 作用域编辑弹窗 -->
    <ScopeEditDialog
      v-model="dialogVisible"
      :dialog-type="dialogType"
      :scope-data="currentScopeData"
      @success="refreshData"
    />
  </ElDialog>
</template>

<script setup lang="ts">
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetScopeList, fetchDeleteScope } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import ScopeEditDialog from './scope-edit-dialog.vue'
  import { ElTag, ElMessage, ElMessageBox } from 'element-plus'

  interface Props {
    modelValue: boolean
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  /**
   * 弹窗显示状态双向绑定
   */
  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  type ScopeListItem = Api.SystemManage.ScopeListItem

  const dialogVisible = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentScopeData = ref<ScopeListItem | undefined>(undefined)

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
      apiFn: fetchGetScopeList,
      apiParams: {
        current: 1,
        size: 20
      },
      excludeParams: [],
      columnsFactory: () => [
        {
          prop: 'scopeCode',
          label: '作用域编码',
          minWidth: 120
        },
        {
          prop: 'scopeName',
          label: '作用域名称',
          minWidth: 120
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'sortOrder',
          label: '排序',
          width: 80
        },
        {
          prop: 'createTime',
          label: '创建时间',
          width: 180
        },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row) => {
            const isDefaultScope = ['global', 'team'].includes(row.scopeCode)
            const list = [{ key: 'edit', label: '编辑', icon: 'ri:edit-2-line' }]
            if (!isDefaultScope) {
              list.push({
                key: 'delete',
                label: '删除',
                icon: 'ri:delete-bin-4-line'
              })
            }
            return h('div', [
              h(ArtButtonMore, {
                list,
                onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
              })
            ])
          }
        }
      ]
    }
  })

  const showDialog = (type: 'add' | 'edit', row?: ScopeListItem) => {
    dialogVisible.value = true
    dialogType.value = type
    currentScopeData.value = row
  }

  const buttonMoreClick = (item: ButtonMoreItem, row: ScopeListItem) => {
    switch (item.key) {
      case 'edit':
        showDialog('edit', row)
        break
      case 'delete':
        deleteScope(row)
        break
    }
  }

  const deleteScope = (row: ScopeListItem) => {
    ElMessageBox.confirm(`确定删除作用域"${row.scopeName}"吗？此操作不可恢复！`, '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
      .then(() => fetchDeleteScope(row.scopeId))
      .then(() => {
        ElMessage.success('删除成功')
        refreshData()
      })
      .catch((e) => {
        if (e !== 'cancel') {
          const error = e as any

          // 检查是否有角色关联信息
          // 当 HTTP 状态码是 500 时，error.data 包含整个响应体 { code, message, data }
          // 当 HTTP 状态码是 200 但 code !== 0 时，error.data 直接是后端返回的 data
          let errorData = error?.data

          // 如果 error.data 有嵌套的 data 字段（HTTP 500 的情况）
          if (
            errorData &&
            typeof errorData === 'object' &&
            errorData.data &&
            typeof errorData.data === 'object'
          ) {
            errorData = errorData.data
          }

          // 检查是否有角色关联信息
          if (errorData && typeof errorData === 'object' && Array.isArray(errorData.roles)) {
            const roles = errorData.roles as string[]
            const roleCount = errorData.roleCount || roles.length
            ElMessageBox.alert(
              `无法删除作用域"${row.scopeName}"，该作用域已被 ${roleCount} 个角色使用：\n${roles.join('、')}`,
              '删除失败',
              {
                confirmButtonText: '确定',
                type: 'warning',
                dangerouslyUseHTMLString: false
              }
            )
          } else {
            // 显示后端返回的错误消息（优先使用后端消息，如果没有则使用通用错误消息）
            const errorMessage = error?.message || '删除失败'
            ElMessage.error(errorMessage)
          }
        }
      })
  }

  const handleClose = () => {
    visible.value = false
  }
</script>
