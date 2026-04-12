<template>
  <ElDrawer
    v-model="visible"
    :title="dialogType === 'add' ? '新增字典类型' : '编辑字典类型'"
    size="420px"
    direction="rtl"
    @close="handleClose"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="80px">
      <ElFormItem label="编码" prop="code">
        <ElInput
          v-model="form.code"
          placeholder="如 gender, page_status"
          :disabled="dialogType === 'edit' && isBuiltin"
        />
      </ElFormItem>
      <ElFormItem label="名称" prop="name">
        <ElInput v-model="form.name" placeholder="如 性别、页面状态" />
      </ElFormItem>
      <ElFormItem label="描述">
        <ElInput v-model="form.description" type="textarea" :rows="3" placeholder="字典用途说明" />
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
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">确定</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { ref, reactive, computed, watch } from 'vue'
  import type { FormInstance, FormRules } from 'element-plus'
  import { ElMessage } from 'element-plus'
  import {
    fetchCreateDictType,
    fetchUpdateDictType,
    type DictTypeSummary
  } from '@/api/system-manage/dictionary'

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit'
    typeData?: DictTypeSummary
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    typeData: undefined
  })

  const emit = defineEmits<Emits>()
  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const isBuiltin = computed(() => props.typeData?.is_builtin ?? false)

  const form = reactive({
    code: '',
    name: '',
    description: '',
    status: 'normal' as string,
    sort_order: 0
  })

  const rules: FormRules = {
    code: [
      { required: true, message: '请输入编码', trigger: 'blur' },
      { max: 100, message: '编码最长 100 字符', trigger: 'blur' }
    ],
    name: [
      { required: true, message: '请输入名称', trigger: 'blur' },
      { max: 200, message: '名称最长 200 字符', trigger: 'blur' }
    ]
  }

  function initForm() {
    if (props.dialogType === 'edit' && props.typeData) {
      form.code = props.typeData.code
      form.name = props.typeData.name
      form.description = props.typeData.description ?? ''
      form.status = props.typeData.status
      form.sort_order = props.typeData.sort_order ?? 0
    } else {
      form.code = ''
      form.name = ''
      form.description = ''
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

    submitting.value = true
    try {
      const body = {
        code: form.code,
        name: form.name,
        description: form.description || undefined,
        status: form.status as 'normal' | 'suspended',
        sort_order: form.sort_order
      }

      if (props.dialogType === 'add') {
        await fetchCreateDictType(body)
        ElMessage.success('新增成功')
      } else if (props.typeData) {
        await fetchUpdateDictType(props.typeData.id, body)
        ElMessage.success('更新成功')
      }
      emit('success')
      handleClose()
    } catch (error) {
      if (error instanceof Error) {
        ElMessage.error(error.message)
      }
    } finally {
      submitting.value = false
    }
  }
</script>
