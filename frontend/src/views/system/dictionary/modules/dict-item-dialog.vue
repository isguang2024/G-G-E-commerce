<template>
  <ElDrawer
    v-model="visible"
    :title="dialogType === 'add' ? '新增字典项' : '编辑字典项'"
    size="400px"
    direction="rtl"
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="80px">
      <ElFormItem label="标签" prop="label">
        <ElInput v-model="form.label" placeholder="显示文本，如「男」" />
      </ElFormItem>
      <ElFormItem label="值" prop="value">
        <ElInput v-model="form.value" placeholder="存储值，如 male" />
      </ElFormItem>
      <ElFormItem label="默认">
        <ElSwitch v-model="form.is_default" />
      </ElFormItem>
      <ElFormItem label="状态">
        <ElSelect v-model="form.status" style="width: 100%">
          <ElOption label="正常" value="normal" />
          <ElOption label="停用" value="suspended" />
        </ElSelect>
      </ElFormItem>
      <ElFormItem label="排序">
        <ElInputNumber v-model="form.sort_order" :min="0" :max="9999" controls-position="right" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton type="primary" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, watch } from 'vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import type { DictItemSummary } from '@/api/system-manage/dictionary'

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    itemData?: DictItemSummary
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success', item: DictItemSummary): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    itemData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const form = reactive({
    label: '',
    value: '',
    is_default: false,
    status: 'normal' as string,
    sort_order: 0
  })

  const rules: FormRules = {
    label: [
      { required: true, message: '请输入标签', trigger: 'blur' },
      { max: 200, message: '标签最长 200 字符', trigger: 'blur' }
    ],
    value: [
      { required: true, message: '请输入值', trigger: 'blur' },
      { max: 200, message: '值最长 200 字符', trigger: 'blur' }
    ]
  }

  function initForm() {
    if (props.dialogType === 'edit' && props.itemData) {
      form.label = props.itemData.label
      form.value = props.itemData.value
      form.is_default = props.itemData.is_default ?? false
      form.status = props.itemData.status
      form.sort_order = props.itemData.sort_order ?? 0
    } else {
      form.label = ''
      form.value = ''
      form.is_default = false
      form.status = 'normal'
      form.sort_order = 0
    }
  }

  watch(
    () => props.modelValue,
    (val) => {
      if (val) initForm()
    }
  )

  function handleClose() {
    visible.value = false
    formRef.value?.resetFields()
  }

  async function handleSubmit() {
    if (!formRef.value) return
    try {
      await formRef.value.validate()
    } catch {
      return
    }

    const item: DictItemSummary = {
      id: props.itemData?.id || `_new_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
      label: form.label,
      value: form.value,
      is_default: form.is_default,
      status: form.status,
      sort_order: form.sort_order
    }
    emit('success', item)
    handleClose()
  }
</script>
