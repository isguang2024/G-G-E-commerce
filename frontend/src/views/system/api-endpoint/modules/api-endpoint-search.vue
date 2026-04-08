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
    method: string
    path: string
    keyword: string
    permissionKey: string
    permissionPattern: string
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

  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  const formItems = computed<FormItem[]>(() => [
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
      label: '权限结构',
      key: 'permissionPattern',
      type: 'select',
      props: {
        clearable: true,
        options: [
          { label: '无权限键', value: 'none' },
          { label: '公开接口', value: 'public' },
          { label: '登录态全局接口', value: 'global_jwt' },
          { label: '登录态自服务接口', value: 'self_jwt' },
          { label: '开放 API Key 接口', value: 'api_key' },
          { label: '单权限接口', value: 'single' },
          { label: '多权限共享', value: 'shared' },
          { label: '跨空间共享', value: 'cross_context_shared' }
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
      label: '权限键状态',
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

  const handleSearch = () => emit('search')

  const handleReset = () => emit('reset')
</script>
