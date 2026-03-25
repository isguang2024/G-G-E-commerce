<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :rules="rules"
    :showExpand="true"
    :defaultExpanded="true"
    @search="handleSearch"
    @reset="handleReset"
  />
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'

  interface SearchForm {
    source: string
    method: string
    path: string
    keyword: string
    permissionKey: string
    contextScope: string
    featureKind: string
    status: string
    hasPermissionKey: string
  }

  interface Props {
    modelValue: SearchForm
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
      label: '注册方式',
      key: 'source',
      type: 'select',
      props: {
        clearable: true,
        options: [
          { label: '自动注册', value: 'sync' },
          { label: '手工补录', value: 'manual' },
          { label: '初始种子', value: 'seed' }
        ]
      }
    },
    {
      label: 'Method',
      key: 'method',
      type: 'select',
      props: {
        clearable: true,
        options: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((value) => ({
          label: value,
          value
        }))
      }
    },
    {
      label: '路径',
      key: 'path',
      type: 'input',
      props: { clearable: true, placeholder: '按路径搜索' }
    },
    {
      label: '关键词',
      key: 'keyword',
      type: 'input',
      props: { clearable: true, placeholder: '按摘要/处理器搜索' }
    },
    {
      label: '权限键',
      key: 'permissionKey',
      type: 'input',
      props: { clearable: true, placeholder: '按权限键搜索' }
    },
    {
      label: '团队上下文',
      key: 'contextScope',
      type: 'select',
      props: {
        clearable: true,
        options: [
          { label: '可选', value: 'optional' },
          { label: '必需', value: 'required' },
          { label: '禁止', value: 'forbidden' }
        ]
      }
    },
    {
      label: '功能归属',
      key: 'featureKind',
      type: 'select',
      props: {
        clearable: true,
        options: [
          { label: '系统', value: 'system' },
          { label: '业务', value: 'business' }
        ]
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        clearable: true,
        options: [
          { label: '正常', value: 'normal' },
          { label: '停用', value: 'suspended' }
        ]
      }
    },
    {
      label: '权限键',
      key: 'hasPermissionKey',
      type: 'select',
      props: {
        clearable: true,
        options: [
          { label: '有权限键', value: 'true' },
          { label: '无权限键', value: 'false' }
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
