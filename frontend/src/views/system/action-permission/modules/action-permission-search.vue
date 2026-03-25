<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :rules="rules"
    :showExpand="true"
    :defaultExpanded="false"
    @search="handleSearch"
    @reset="handleReset"
  />
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'

  interface SearchForm {
    keyword: string
    moduleGroupId: string
    featureGroupId: string
    contextType: string
    status: string
    isBuiltin: string
  }

  interface OptionItem {
    label: string
    value: string
  }

  interface Props {
    modelValue: SearchForm
    moduleGroupOptions: OptionItem[]
    featureGroupOptions: OptionItem[]
  }

  interface Emits {
    (e: 'update:modelValue', value: SearchForm): void
    (e: 'search'): void
    (e: 'reset'): void
  }

  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()
  const searchBarRef = ref()

  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  const rules = {}

  const formItems = computed<FormItem[]>(() => [
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '名称/描述/权限键'
      }
    },
    {
      label: '模块分组',
      key: 'moduleGroupId',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '请选择模块分组',
        options: props.moduleGroupOptions
      }
    },
    {
      label: '功能分组',
      key: 'featureGroupId',
      type: 'select',
      props: {
        clearable: true,
        filterable: true,
        placeholder: '请选择功能分组',
        options: props.featureGroupOptions
      }
    },
    {
      label: '上下文',
      key: 'contextType',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '请选择上下文',
        options: [
          { label: '平台', value: 'platform' },
          { label: '团队', value: 'team' },
          { label: '通用', value: 'common' }
        ]
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '请选择状态',
        options: [
          { label: '正常', value: 'normal' },
          { label: '停用', value: 'suspended' }
        ]
      }
    },
    {
      label: '是否内置',
      key: 'isBuiltin',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '请选择是否内置',
        options: [
          { label: '内置', value: 'true' },
          { label: '自定义', value: 'false' }
        ]
      }
    }
  ])

  const handleSearch = async () => {
    await searchBarRef.value?.validate()
    emit('search')
  }

  const handleReset = () => emit('reset')
</script>
