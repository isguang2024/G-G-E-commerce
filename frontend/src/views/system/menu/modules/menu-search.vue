<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :showExpand="false"
    @reset="handleReset"
    @search="handleSearch"
  />
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import type { FormItem } from '@/components/core/forms/art-form/index.vue'

  interface SearchForm {
    name: string
    route: string
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

  const formItems = computed<FormItem[]>(() => [
    {
      label: '菜单名称',
      key: 'name',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '请输入菜单名称'
      }
    },
    {
      label: '路由地址',
      key: 'route',
      type: 'input',
      props: {
        clearable: true,
        placeholder: '请输入路由地址'
      }
    }
  ])

  const handleSearch = async () => {
    await searchBarRef.value?.validate()
    emit('search')
  }

  const handleReset = () => emit('reset')
</script>
