<template>
  <ElDrawer
    :model-value="modelValue"
    title="菜单分组"
    size="720px"
    direction="rtl"
    @update:model-value="handleClose"
  
    class="config-drawer">
    <div class="group-drawer">
      <div class="group-form-card">
        <div class="group-form-title">{{ editingId ? '编辑分组' : '新增分组' }}</div>
        <ElForm ref="formRef" :model="form" :rules="rules" label-width="72px">
          <ElRow :gutter="16">
            <ElCol :span="16">
              <ElFormItem label="名称" prop="name">
                <ElInput v-model="form.name" placeholder="请输入分组名称" />
              </ElFormItem>
            </ElCol>
            <ElCol :span="8">
              <ElFormItem label="排序" prop="sortOrder">
                <ElInputNumber v-model="form.sortOrder" :min="0" controls-position="right" class="w-full" />
              </ElFormItem>
            </ElCol>
          </ElRow>
        </ElForm>
        <div class="group-form-actions">
          <ElButton v-if="editingId" @click="resetForm">取消编辑</ElButton>
          <ElButton type="primary" :loading="saving" @click="handleSave">
            {{ editingId ? '保存分组' : '新增分组' }}
          </ElButton>
        </div>
      </div>

      <ElTable :data="pagedItems" border v-loading="loading" class="group-table">
        <ElTableColumn prop="name" label="分组名称" min-width="220" />
        <ElTableColumn prop="sortOrder" label="排序" width="90" align="center" />
        <ElTableColumn label="操作" width="160" align="center">
          <template #default="{ row }">
            <div class="group-table-actions">
              <ElButton link type="primary" @click="startEdit(row)">编辑</ElButton>
              <ElButton link type="danger" @click="handleDelete(row)">删除</ElButton>
            </div>
          </template>
        </ElTableColumn>
      </ElTable>
      <WorkspacePagination
        v-if="items.length > 0"
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.size"
        :total="items.length"
        compact
      />
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, reactive, ref, watch } from 'vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessageBox } from 'element-plus'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'

  interface Props {
    modelValue: boolean
    loading?: boolean
    saving?: boolean
    items: Api.SystemManage.MenuManageGroupItem[]
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (
      e: 'save',
      payload: Api.SystemManage.MenuManageGroupSaveParams & {
        id?: string
      }
    ): void
    (e: 'delete', id: string): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    loading: false,
    saving: false,
    items: () => []
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()
  const editingId = ref('')
  const pagination = reactive({
    current: 1,
    size: 10
  })
  const form = reactive({
    name: '',
    sortOrder: 0
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }]
  }

  const pagedItems = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return props.items.slice(start, start + pagination.size)
  })

  function resetForm() {
    editingId.value = ''
    form.name = ''
    form.sortOrder = 0
    formRef.value?.clearValidate()
  }

  function handleClose() {
    emit('update:modelValue', false)
    resetForm()
    pagination.current = 1
  }

  function startEdit(row: Api.SystemManage.MenuManageGroupItem) {
    editingId.value = row.id
    form.name = row.name
    form.sortOrder = row.sortOrder ?? 0
  }

  async function handleSave() {
    if (!formRef.value) return
    await formRef.value.validate()
    emit('save', {
      id: editingId.value || undefined,
      name: form.name.trim(),
      sort_order: Number(form.sortOrder ?? 0),
      status: 'normal'
    })
  }

  async function handleDelete(row: Api.SystemManage.MenuManageGroupItem) {
    await ElMessageBox.confirm(`确认删除分组“${row.name}”吗？`, '提示', { type: 'warning' })
    emit('delete', row.id)
    if (editingId.value === row.id) {
      resetForm()
    }
  }

  watch(
    () => [props.modelValue, props.items.length],
    () => {
      pagination.current = 1
    }
  )
</script>

<style lang="scss" scoped>
  .group-drawer {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .group-form-card {
    border: 1px solid var(--el-border-color-light);
    border-radius: 14px;
    padding: 18px 18px 10px;
    background: linear-gradient(180deg, #fafcff 0%, #f5f7fb 100%);
  }

  .group-form-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--el-text-color-primary);
    margin-bottom: 14px;
  }

  .group-form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
  }

  .group-table-actions {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .w-full {
    width: 100%;
  }
</style>
