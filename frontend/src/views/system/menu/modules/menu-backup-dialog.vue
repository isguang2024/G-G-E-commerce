<template>
  <ElDialog v-model="visible" title="备份菜单" width="500px" destroy-on-close>
    <ElForm :model="form" label-width="80px">
      <ElFormItem label="备份名称" required>
        <ElInput v-model="form.name" placeholder="请输入备份名称" />
      </ElFormItem>
      <ElFormItem label="备份描述">
        <ElInput v-model="form.description" type="textarea" placeholder="请输入备份描述" :rows="3" />
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
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'submit', data: { name: string; description: string }): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    loading: false
  })

  const emit = defineEmits<Emits>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

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
