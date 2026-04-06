<template>
  <ArtSearchBar
    v-model="formData"
    :items="formItems"
    :showExpand="true"
    @search="handleSearch"
    @reset="handleReset"
  />
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'

  interface SearchForm {
    keyword: string
    moduleGroupId: string
    featureGroupId: string
    contextType: string
    status: string
    isBuiltin: string
    usagePattern: string
    duplicatePattern: string
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

  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

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
          { label: '协作空间', value: 'team' },
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
      label: '消费情况',
      key: 'usagePattern',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '请选择消费情况',
        options: [
          { label: '未被消费', value: 'unused' },
          { label: '仅 API', value: 'api_only' },
          { label: '仅页面', value: 'page_only' },
          { label: '仅功能包', value: 'package_only' },
          { label: '多方复用', value: 'multi_consumer' }
        ]
      }
    },
    {
      label: '重复判定',
      key: 'duplicatePattern',
      type: 'select',
      props: {
        clearable: true,
        placeholder: '请选择重复判定',
        options: [
          { label: '跨上下文镜像', value: 'cross_context_mirror' },
          { label: '疑似重复', value: 'suspected_duplicate' }
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

  const handleSearch = () => emit('search')

  const handleReset = () => emit('reset')
</script>
