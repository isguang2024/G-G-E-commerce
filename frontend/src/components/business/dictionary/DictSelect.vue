<template>
  <ElSelect
    :model-value="modelValue"
    class="dict-select"
    :placeholder="placeholder"
    :loading="loading"
    :disabled="disabled"
    :clearable="clearable"
    :filterable="true"
    :allow-create="allowCreate"
    :multiple="multiple"
    :collapse-tags="multiple"
    :collapse-tags-tooltip="multiple"
    default-first-option
    reserve-keyword
    @change="handleChange"
  >
    <ElOption
      v-for="item in mergedOptions"
      :key="item.value"
      :label="item.label"
      :value="item.value"
    >
      <div class="dict-select__option">
        <div class="dict-select__content">
          <span>{{ item.label }}</span>
          <span v-if="item.description" class="dict-select__description">{{ item.description }}</span>
        </div>
        <ElTag v-if="item.isDefault" size="small" type="success" effect="plain">默认</ElTag>
      </div>
    </ElOption>
  </ElSelect>
</template>

<script setup lang="ts">
  import { computed, watch } from 'vue'
  import { useDictionary, type DictOption } from '@/hooks/business/useDictionary'

  defineOptions({ name: 'DictSelect' })

  const props = withDefaults(
    defineProps<{
      code: string
      modelValue?: string | string[]
      placeholder?: string
      disabled?: boolean
      clearable?: boolean
      allowCreate?: boolean
      multiple?: boolean
      autoSelectDefault?: boolean
      fallbackOptions?: DictOption[]
    }>(),
    {
      modelValue: '',
      placeholder: '请选择',
      disabled: false,
      clearable: true,
      allowCreate: false,
      multiple: false,
      autoSelectDefault: false,
      fallbackOptions: () => []
    }
  )

  const emit = defineEmits<{
    'update:modelValue': [value: any]
    change: [value: any]
    'create-option': [value: string]
  }>()

  const { options, loading, defaultValue } = useDictionary(props.code)

  const mergedOptions = computed(() => {
    const map = new Map<string, DictOption>()
    for (const item of props.fallbackOptions) {
      const value = `${item.value || ''}`.trim()
      if (!value) continue
      map.set(value, {
        label: `${item.label || value}`.trim() || value,
        value,
        description: `${item.description || ''}`.trim(),
        isDefault: Boolean(item.isDefault),
        extra: item.extra
      })
    }
    for (const item of options.value) {
      const value = `${item.value || ''}`.trim()
      if (!value) continue
      map.set(value, item)
    }
    return Array.from(map.values()).sort((a, b) => Number(b.isDefault) - Number(a.isDefault))
  })

  const mergedDefaultValue = computed(
    () =>
      defaultValue.value ||
      mergedOptions.value.find((item) => item.isDefault)?.value ||
      mergedOptions.value[0]?.value ||
      ''
  )

  watch(
    () => [props.autoSelectDefault, props.modelValue, mergedDefaultValue.value, props.multiple],
    () => {
      if (!props.autoSelectDefault || props.multiple) return
      const current = `${props.modelValue || ''}`.trim()
      if (current) return
      const next = mergedDefaultValue.value
      if (!next) return
      emit('update:modelValue', next)
      emit('change', next)
    },
    { immediate: true }
  )

  function handleChange(value: string | string[]) {
    const nextValue = Array.isArray(value)
      ? value.map((item) => `${item || ''}`.trim()).filter(Boolean)
      : `${value || ''}`.trim()

    emit('update:modelValue', nextValue)
    emit('change', nextValue)

    if (!Array.isArray(nextValue) && nextValue && !mergedOptions.value.some((item) => item.value === nextValue)) {
      emit('create-option', nextValue)
    }
  }
</script>

<style scoped lang="scss">
  .dict-select {
    width: 100%;
  }

  .dict-select__option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
  }

  .dict-select__content {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .dict-select__description {
    font-size: 12px;
    line-height: 1.4;
    color: var(--el-text-color-secondary);
    white-space: normal;
  }
</style>
