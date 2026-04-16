<template>
  <ElSelect
    :model-value="modelValue"
    class="app-key-select"
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
    @visible-change="handleVisibleChange"
    @change="handleChange"
  >
    <ElOption
      v-for="item in options"
      :key="item.appKey"
      :label="item.label"
      :value="item.appKey"
    >
      <div class="app-key-select__option">
        <div class="app-key-select__meta">
          <span class="app-key-select__name">{{ item.name || item.appKey }}</span>
          <span class="app-key-select__key">{{ item.appKey }}</span>
        </div>
        <ElTag v-if="item.isDefault" size="small" type="success" effect="plain">默认</ElTag>
      </div>
    </ElOption>
  </ElSelect>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { computed, onMounted, ref } from 'vue'
  import { loadAppCatalog, type AppCatalogOption } from '@/domains/governance/app-catalog'

  defineOptions({ name: 'AppKeySelect' })

  const props = withDefaults(
    defineProps<{
      modelValue?: string | string[]
      placeholder?: string
      disabled?: boolean
      clearable?: boolean
      allowCreate?: boolean
      eager?: boolean
      multiple?: boolean
    }>(),
    {
      modelValue: '',
      placeholder: '选择应用',
      disabled: false,
      clearable: true,
      allowCreate: false,
      eager: true,
      multiple: false
    }
  )

  const emit = defineEmits<{
    'update:modelValue': [value: any]
    change: [value: any]
  }>()

  const loading = ref(false)
  const options = ref<AppCatalogOption[]>([])

  const normalizedValue = computed(() => {
    if (Array.isArray(props.modelValue)) {
      return props.modelValue.map((item) => `${item || ''}`.trim()).filter(Boolean)
    }
    return `${props.modelValue || ''}`.trim()
  })

  const ensureOptions = async (force = false) => {
    if (options.value.length > 0 && !force) {
      return
    }
    loading.value = true
    try {
      options.value = await loadAppCatalog(force)
    } catch (error: any) {
      ElMessage.error(error?.message || '加载应用列表失败')
    } finally {
      loading.value = false
    }
  }

  const handleVisibleChange = (visible: boolean) => {
    if (visible) {
      void ensureOptions()
    }
  }

  const handleChange = (value: string | string[]) => {
    const nextValue = Array.isArray(value)
      ? value.map((item) => `${item || ''}`.trim()).filter(Boolean)
      : `${value || ''}`.trim()
    emit('update:modelValue', nextValue)
    emit('change', nextValue)
  }

  onMounted(() => {
    if (props.eager || normalizedValue.value) {
      void ensureOptions()
    }
  })
</script>

<style scoped lang="scss">
  .app-key-select {
    width: 100%;
  }

  .app-key-select__option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
  }

  .app-key-select__meta {
    display: flex;
    align-items: baseline;
    gap: 8px;
    min-width: 0;
  }

  .app-key-select__name {
    color: var(--el-text-color-primary);
    font-weight: 500;
  }

  .app-key-select__key {
    color: var(--el-text-color-secondary);
    font-family: var(--el-font-family-monospace, monospace);
    font-size: 12px;
  }
</style>
