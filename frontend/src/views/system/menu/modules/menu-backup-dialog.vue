<template>
  <ElDialog v-model="visible" :title="dialogTitle" width="500px" destroy-on-close>
    <ElAlert
      class="mb-4"
      type="info"
      :closable="false"
      :title="alertTitle"
      :description="alertDescription"
    />
    <ElForm :model="form" label-width="80px">
      <ElFormItem label="备份名称" required>
        <ElInput v-model="form.name" placeholder="请输入备份名称" />
      </ElFormItem>
      <ElFormItem label="备份描述">
        <ElInput
          v-model="form.description"
          type="textarea"
          placeholder="请输入备份描述"
          :rows="3"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="visible = false">取消</ElButton>
        <ElButton type="primary" :loading="loading" @click="handleSubmit">确认备份</ElButton>
      </span>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, reactive, watch } from 'vue'
  import { ElMessage } from 'element-plus'

  interface Props {
    modelValue: boolean
    loading?: boolean
    currentSpaceName?: string
    scopeType?: 'space' | 'global'
    dialogTitle?: string
    alertTitle?: string
    alertDescription?: string
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'submit', data: { name: string; description: string }): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    loading: false,
    currentSpaceName: '默认空间',
    scopeType: 'space'
  })

  const emit = defineEmits<Emits>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const dialogTitle = computed(
    () =>
      props.dialogTitle || (props.scopeType === 'global' ? '备份全部空间菜单' : '备份当前空间菜单')
  )

  const alertTitle = computed(
    () =>
      props.alertTitle ||
      (props.scopeType === 'global'
        ? '当前将创建全局菜单备份'
        : `当前将备份菜单空间：${props.currentSpaceName}`)
  )

  const alertDescription = computed(
    () =>
      props.alertDescription ||
      (props.scopeType === 'global'
        ? '该备份会保存全部空间的菜单树与菜单分组，恢复时会覆盖所有空间菜单。该入口用于正式全局备份，不再依赖省略 space_key 的兼容语义。'
        : '该备份只保存当前空间菜单树及其引用的菜单分组，避免多空间场景下一次备份覆盖所有空间。')
  )

  const form = reactive({
    name: '',
    description: ''
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) {
        form.name = ''
        form.description = ''
      }
    }
  )

  function handleSubmit() {
    const name = form.name.trim()
    if (!name) {
      ElMessage.warning('请输入备份名称')
      return
    }
    emit('submit', {
      name,
      description: form.description.trim()
    })
  }
</script>
