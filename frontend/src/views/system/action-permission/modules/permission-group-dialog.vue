<template>
  <ElDrawer
    v-model="visible"
    :title="`${groupTypeLabel}管理`"
    size="1160px"
    destroy-on-close
    @open="handleDialogOpen"
    direction="rtl"
    class="config-drawer"
  >
    <div class="group-dialog-body">
      <ElCard shadow="never" class="group-list-card">
        <template #header>
          <div class="group-list-header">
            <span>{{ groupTypeLabel }}列表</span>
            <div class="group-list-actions">
              <ElButton size="small" @click="loadGroupList" :loading="listLoading">刷新</ElButton>
              <ElButton size="small" type="primary" @click="startCreate">新增分组</ElButton>
            </div>
          </div>
        </template>
        <ElTable
          v-loading="listLoading"
          :data="pagedGroupList"
          height="100%"
          highlight-current-row
          @current-change="handleCurrentChange"
        >
          <ElTableColumn prop="code" label="分组编码" min-width="120" show-overflow-tooltip />
          <ElTableColumn prop="name" label="分组名称" min-width="140" show-overflow-tooltip />
          <ElTableColumn label="状态" width="90" align="center">
            <template #default="{ row }">
              <ElTag :type="row.status === 'normal' ? 'success' : 'danger'">
                {{ row.status === 'normal' ? '正常' : '停用' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn prop="sortOrder" label="排序" width="70" align="center" />
          <ElTableColumn label="内置" width="70" align="center">
            <template #default="{ row }">
              <ElTag :type="row.isBuiltin ? 'success' : 'info'" effect="plain">
                {{ row.isBuiltin ? '是' : '否' }}
              </ElTag>
            </template>
          </ElTableColumn>
          <ElTableColumn label="操作" width="72" fixed="right" align="center">
            <template #default="{ row }">
              <ArtButtonMore
                :list="buildOperationList(row)"
                @click="(item) => handleOperation(item, row)"
              />
            </template>
          </ElTableColumn>
        </ElTable>
        <WorkspacePagination
          v-model:current-page="pagination.current"
          v-model:page-size="pagination.size"
          :total="groupList.length"
          compact
        />
      </ElCard>

      <ElCard shadow="never" class="group-form-card">
        <template #header>
          <div class="group-form-header">
            <span>{{ form.id ? `编辑${groupTypeLabel}` : `新增${groupTypeLabel}` }}</span>
            <ElButton text @click="startCreate">清空</ElButton>
          </div>
        </template>
        <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
          <ElFormItem label="分组编码" prop="code">
            <ElInput v-model="form.code" placeholder="例如 role 或 system_feature" />
          </ElFormItem>
          <ElFormItem label="分组名称" prop="name">
            <ElInput v-model="form.name" placeholder="请输入名称" />
          </ElFormItem>
          <ElFormItem label="英文名称">
            <ElInput v-model="form.nameEn" placeholder="可选" />
          </ElFormItem>
          <ElFormItem>
            <template #label>
              <span class="label-help">
                <span>状态</span>
                <ElTooltip
                  content="仅影响分组管理，不影响鉴权判断；权限是否可用以权限键状态为准。"
                  placement="top"
                >
                  <ElIcon class="label-help-icon"><QuestionFilled /></ElIcon>
                </ElTooltip>
              </span>
            </template>
            <ElSelect v-model="form.status" style="width: 100%">
              <ElOption label="正常" value="normal" />
              <ElOption label="停用" value="suspended" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="排序">
            <ElInputNumber v-model="form.sortOrder" :min="0" :max="9999" style="width: 100%" />
          </ElFormItem>
          <ElFormItem label="说明">
            <ElInput v-model="form.description" type="textarea" :rows="4" />
          </ElFormItem>
        </ElForm>
      </ElCard>
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import type { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'
  import { ElIcon, ElMessage, ElMessageBox, ElTooltip } from 'element-plus'
  import {
    fetchCreatePermissionGroup,
    fetchDeletePermissionGroup,
    fetchGetPermissionGroupList,
    fetchUpdatePermissionGroup
  } from '@/domains/governance/api'

  interface Props {
    modelValue: boolean
    groupType: 'module' | 'feature'
    groupData?: Api.SystemManage.PermissionGroupItem
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    groupType: 'module',
    groupData: undefined
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const groupTypeLabel = computed(() => (props.groupType === 'module' ? '模块分组' : '功能分组'))
  const formRef = ref<FormInstance>()
  const listLoading = ref(false)
  const submitting = ref(false)
  const groupList = ref<Api.SystemManage.PermissionGroupItem[]>([])
  const pagination = reactive({
    current: 1,
    size: 10
  })
  const form = reactive({
    id: '',
    code: '',
    name: '',
    nameEn: '',
    description: '',
    status: 'normal',
    sortOrder: 0
  })

  const rules = reactive<FormRules>({
    code: [{ required: true, message: '请输入分组编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }]
  })

  const pagedGroupList = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return groupList.value.slice(start, start + pagination.size)
  })

  function initForm(data?: Api.SystemManage.PermissionGroupItem) {
    Object.assign(form, {
      id: data?.id || '',
      code: data?.code || '',
      name: data?.name || '',
      nameEn: data?.nameEn || '',
      description: data?.description || '',
      status: data?.status || 'normal',
      sortOrder: data?.sortOrder ?? 0
    })
  }

  function startCreate() {
    formRef.value?.clearValidate()
    initForm()
  }

  function startEdit(row: Api.SystemManage.PermissionGroupItem) {
    formRef.value?.clearValidate()
    initForm(row)
  }

  function handleCurrentChange(row?: Api.SystemManage.PermissionGroupItem) {
    if (!row) return
    startEdit(row)
  }

  function buildOperationList(row: Api.SystemManage.PermissionGroupItem): ButtonMoreItem[] {
    const list: ButtonMoreItem[] = [{ key: 'edit', label: '编辑', icon: 'ri:edit-2-line' }]
    list.push({
      key: 'delete',
      label: '删除',
      icon: 'ri:delete-bin-4-line',
      color: '#f56c6c',
      disabled: !!row.isBuiltin
    })
    return list
  }

  function handleOperation(item: ButtonMoreItem, row: Api.SystemManage.PermissionGroupItem) {
    if (item.key === 'edit') {
      startEdit(row)
      return
    }
    if (item.key === 'delete' && !row.isBuiltin) {
      handleDelete(row)
    }
  }

  async function loadGroupList() {
    listLoading.value = true
    try {
      const res = await fetchGetPermissionGroupList({
        current: 1,
        size: 500,
        groupType: props.groupType
      })
      groupList.value = (res.records || []).slice().sort((a, b) => {
        const sortA = a.sortOrder ?? 0
        const sortB = b.sortOrder ?? 0
        if (sortA !== sortB) return sortA - sortB
        return (a.name || '').localeCompare(b.name || '', 'zh-CN')
      })
      pagination.current = 1
    } finally {
      listLoading.value = false
    }
  }

  async function handleDialogOpen() {
    await loadGroupList()
    if (props.groupData?.id) {
      const current = groupList.value.find((item) => item.id === props.groupData?.id)
      if (current) {
        startEdit(current)
        return
      }
      initForm(props.groupData)
      return
    }
    startCreate()
  }

  async function handleDelete(row: Api.SystemManage.PermissionGroupItem) {
    if (row.isBuiltin) return
    try {
      await ElMessageBox.confirm(
        `确定删除${groupTypeLabel.value}「${row.name}」吗？\n若该分组已被功能权限引用，系统将拦截删除。`,
        '删除确认',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
      await fetchDeletePermissionGroup(row.id)
      ElMessage.success('删除成功')
      await loadGroupList()
      if (form.id === row.id) startCreate()
      emit('success')
    } catch (error: any) {
      if (error !== 'cancel') {
        ElMessage.error(error?.message || '删除失败')
      }
    }
  }

  watch(
    () => [props.modelValue, props.groupData, props.groupType],
    async ([opened]) => {
      if (!opened) return
      await handleDialogOpen()
    },
    { deep: true }
  )

  async function handleSubmit() {
    if (!formRef.value) return
    await formRef.value.validate()
    submitting.value = true
    try {
      const payload = {
        code: form.code.trim(),
        name: form.name.trim(),
        name_en: form.nameEn.trim(),
        description: form.description.trim(),
        group_type: props.groupType,
        status: form.status,
        sort_order: form.sortOrder ?? 0
      }
      if (form.id) {
        await fetchUpdatePermissionGroup(form.id, payload)
      } else {
        await fetchCreatePermissionGroup(payload)
      }
      ElMessage.success('分组保存成功')
      await loadGroupList()
      emit('success')
      const current = groupList.value.find((item) => item.code === payload.code)
      if (current) startEdit(current)
    } catch (error: any) {
      ElMessage.error(error?.message || '分组保存失败')
    } finally {
      submitting.value = false
    }
  }
</script>

<style scoped>
  .group-dialog-body {
    display: grid;
    grid-template-columns: 1.2fr 1fr;
    gap: 12px;
    height: calc(100vh - 186px);
    min-height: 560px;
  }

  .group-list-header,
  .group-form-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .group-list-actions {
    display: flex;
    gap: 8px;
  }

  .group-list-card,
  .group-form-card {
    height: 100%;
  }

  .group-list-card :deep(.el-card__body),
  .group-form-card :deep(.el-card__body) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .group-list-card :deep(.el-table) {
    flex: 1;
  }

  .group-form-card :deep(.el-form) {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .group-form-card :deep(.el-form-item:last-child) {
    flex: 1;
    align-items: stretch;
  }

  .group-form-card :deep(.el-form-item:last-child .el-form-item__content) {
    align-items: stretch;
  }

  .group-form-card :deep(.el-textarea),
  .group-form-card :deep(.el-textarea__inner) {
    height: 100%;
    min-height: 180px !important;
  }

  .label-help {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .label-help-icon {
    font-size: 14px;
    color: var(--el-text-color-secondary);
    cursor: help;
  }

  @media (max-width: 1200px) {
    .group-dialog-body {
      grid-template-columns: 1fr;
      height: auto;
      min-height: 0;
    }

    .group-list-card,
    .group-form-card {
      height: auto;
      min-height: 420px;
    }
  }
</style>
