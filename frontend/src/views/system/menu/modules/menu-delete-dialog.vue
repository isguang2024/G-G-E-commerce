<template>
  <ElDialog
    :model-value="visible"
    title="删除菜单"
    width="520px"
    destroy-on-close
    @close="handleClose"
  >
    <div class="menu-delete-dialog">
      <div class="menu-delete-dialog__summary">
        <div class="menu-delete-dialog__title">{{ titleText }}</div>
        <div class="menu-delete-dialog__desc">{{ summaryText }}</div>
      </div>

      <ElRadioGroup v-model="selectedMode" class="menu-delete-dialog__modes">
        <ElRadio
          v-for="item in modeOptions"
          :key="item.value"
          :label="item.value"
          :disabled="item.disabled"
          border
        >
          <div class="menu-delete-dialog__option">
            <div class="menu-delete-dialog__option-title">{{ item.label }}</div>
            <div class="menu-delete-dialog__option-desc">{{ item.description }}</div>
          </div>
        </ElRadio>
      </ElRadioGroup>

      <div v-if="isPromoteMode" class="menu-delete-dialog__field">
        <div class="menu-delete-dialog__field-label">提到指定父菜单</div>
        <ElTreeSelect
          v-model="targetParentId"
          class="menu-delete-dialog__input"
          clearable
          filterable
          :data="treeParentOptions"
          :props="treeSelectProps"
          placeholder="请选择要提到到的父菜单"
        />
      </div>

      <div class="menu-delete-dialog__field">
        <div class="menu-delete-dialog__field-label">确认删除</div>
        <ElInput
          v-model="confirmText"
          class="menu-delete-dialog__input"
          placeholder="请手动输入“删除”后再提交"
          clearable
        />
      </div>
    </div>

    <template #footer>
      <div class="menu-delete-dialog__footer">
        <ElButton @click="handleClose">取消</ElButton>
        <ElButton type="danger" :loading="loading" :disabled="!canConfirm" @click="handleConfirm">
          确认删除
        </ElButton>
      </div>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElButton, ElDialog, ElInput, ElRadio, ElRadioGroup, ElTreeSelect } from 'element-plus'

  type ParentTreeOption = {
    label: string
    value: string
    children?: ParentTreeOption[]
  }

  const props = withDefaults(
    defineProps<{
      visible: boolean
      loading?: boolean
      menuTitle?: string
      childCount?: number
      descendantCount?: number
      affectedPageCount?: number
      affectedRelationCount?: number
      parentOptions?: ParentTreeOption[]
    }>(),
    {
      loading: false,
      menuTitle: '',
      childCount: 0,
      descendantCount: 0,
      affectedPageCount: 0,
      affectedRelationCount: 0,
      parentOptions: () => []
    }
  )

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'confirm', payload: {
      mode: 'single' | 'cascade' | 'promote_children'
      targetParentId?: string | null
    }): void
  }>()

  const selectedMode = ref<'single' | 'cascade' | 'promote_children'>('single')
  const confirmText = ref('')
  const targetParentId = ref('')

  const hasChildren = computed(() => Number(props.childCount || 0) > 0)
  const titleText = computed(() => props.menuTitle?.trim() || '当前菜单')
  const treeSelectProps = {
    label: 'label',
    value: 'value',
    children: 'children',
    disabled: 'disabled'
  }

  const treeParentOptions = computed(() => [
    {
      label: '顶级菜单',
      value: '',
      children: props.parentOptions || []
    }
  ])

  const summaryText = computed(() => {
    if (!hasChildren.value) {
      return `当前菜单无子菜单，删除后将清理 ${props.affectedPageCount} 条页面挂载与 ${props.affectedRelationCount} 项权限关联`
    }
    return `当前菜单有 ${props.childCount} 个直接子菜单，全部 ${props.descendantCount} 个后代节点，预计影响 ${props.affectedPageCount} 个页面和 ${props.affectedRelationCount} 项权限关联`
  })

  const modeOptions = computed(() => {
    const options = [
      {
        value: 'single' as const,
        label: '删除当前菜单',
        description: hasChildren.value
          ? '当前菜单存在子菜单，请先处理子菜单'
          : '只删除当前菜单',
        disabled: hasChildren.value
      },
      {
        value: 'promote_children' as const,
        label: '删除当前菜单，子菜单提到指定父菜单',
        description:
          '删除当前菜单后，把其直接子菜单提到指定父菜单；若不选择父菜单，将默认挂到顶级',
        disabled: !hasChildren.value
      },
      {
        value: 'cascade' as const,
        label: '删除当前菜单及全部子菜单',
        description: '删除当前菜单及其全部后代菜单。仅在确认无需保留子树时使用',
        disabled: false
      }
    ]
    return options
  })

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return
      selectedMode.value = hasChildren.value ? 'cascade' : 'single'
      confirmText.value = ''
      targetParentId.value = ''
    },
    { immediate: true }
  )

  const canConfirm = computed(() => confirmText.value.trim() === '删除')
  const isPromoteMode = computed(() => selectedMode.value === 'promote_children')

  const handleClose = () => {
    emit('update:visible', false)
  }

  const handleConfirm = () => {
    if (!canConfirm.value) return
    emit('confirm', {
      mode: selectedMode.value,
      targetParentId: isPromoteMode.value && targetParentId.value ? targetParentId.value : null
    })
  }
</script>

<style scoped lang="scss">
  .menu-delete-dialog {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .menu-delete-dialog__input {
    width: 100%;
  }

  .menu-delete-dialog__field {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .menu-delete-dialog__field-label {
    font-size: 13px;
    font-weight: 600;
    color: #374151;
  }

  .menu-delete-dialog__summary {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .menu-delete-dialog__title {
    font-size: 15px;
    font-weight: 700;
    color: #111827;
  }

  .menu-delete-dialog__desc {
    font-size: 13px;
    line-height: 1.6;
    color: #6b7280;
  }

  .menu-delete-dialog__modes {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .menu-delete-dialog__modes :deep(.el-radio) {
    width: 100%;
  }

  .menu-delete-dialog__modes :deep(.el-radio.is-bordered) {
    width: 100%;
    padding-right: 12px;
    box-sizing: border-box;
  }

  .menu-delete-dialog__option {
    display: flex;
    flex-direction: column;
    gap: 4px;
    white-space: normal;
    width: 100%;
  }

  .menu-delete-dialog__option-title {
    font-size: 14px;
    font-weight: 600;
    color: #111827;
  }

  .menu-delete-dialog__option-desc {
    font-size: 12px;
    line-height: 1.6;
    color: #6b7280;
  }

  .menu-delete-dialog__footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  :deep(.el-radio.is-bordered) {
    align-items: flex-start;
    height: auto;
    padding: 12px 14px;
    margin-right: 0;
  }

  :deep(.el-radio__label) {
    width: 100%;
    padding-left: 12px;
  }

  :deep(.el-select) {
    width: 100%;
  }

  :deep(.el-tree-select) {
    width: 100%;
  }

  :deep(.el-tree-select .el-input) {
    width: 100%;
  }
</style>

