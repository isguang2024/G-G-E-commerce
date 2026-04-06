<template>
  <PageGroupDialog
    v-if="resolvedPageType === 'group'"
    :key="dialogKey"
    :model-value="modelValue"
    :dialog-type="dialogType"
    :page-data="pageData"
    :app-key="appKey"
    :menu-spaces="menuSpaces"
    :initial-parent-page-key="initialParentPageKey"
    :initial-parent-menu-id="initialParentMenuId"
    :initial-page-type="initialPageType"
    :default-data="defaultData"
    @update:model-value="emit('update:modelValue', $event)"
    @success="emit('success')"
  />
  <PageDisplayGroupDialog
    v-else-if="resolvedPageType === 'display_group'"
    :key="dialogKey"
    :model-value="modelValue"
    :dialog-type="dialogType"
    :page-data="pageData"
    :app-key="appKey"
    :menu-spaces="menuSpaces"
    :initial-parent-page-key="initialParentPageKey"
    :initial-parent-menu-id="initialParentMenuId"
    :initial-page-type="initialPageType"
    :default-data="defaultData"
    @update:model-value="emit('update:modelValue', $event)"
    @success="emit('success')"
  />
  <PageEntryDialog
    v-else
    :key="dialogKey"
    :model-value="modelValue"
    :dialog-type="dialogType"
    :page-data="pageData"
    :app-key="appKey"
    :menu-spaces="menuSpaces"
    :initial-parent-page-key="initialParentPageKey"
    :initial-parent-menu-id="initialParentMenuId"
    :initial-page-type="initialPageType"
    :default-data="defaultData"
    @update:model-value="emit('update:modelValue', $event)"
    @success="emit('success')"
  />
</template>

<script setup lang="ts">
  import { computed, toRefs } from 'vue'
  import PageDisplayGroupDialog from './page-display-group-dialog.vue'
  import PageGroupDialog from './page-group-dialog.vue'
  import PageEntryDialog from './page-entry-dialog.vue'

  type PageItem = Api.SystemManage.PageItem

  interface Props {
    modelValue: boolean
    dialogType: 'add' | 'edit' | 'copy'
    pageData?: Partial<PageItem>
    appKey?: string
    menuSpaces?: Api.SystemManage.MenuSpaceItem[]
    initialParentPageKey?: string
    initialParentMenuId?: string
    initialPageType?: PageItem['pageType']
    defaultData?: Partial<PageItem>
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    dialogType: 'add',
    pageData: undefined,
    menuSpaces: () => [],
    initialParentPageKey: '',
    initialParentMenuId: '',
    initialPageType: 'inner',
    defaultData: undefined
  })

  const emit = defineEmits<Emits>()

  const resolvedPageType = computed(
    () =>
      props.pageData?.pageType || props.defaultData?.pageType || props.initialPageType || 'inner'
  )

  const dialogKey = computed(() =>
    [
      props.dialogType,
      resolvedPageType.value,
      props.pageData?.id || '',
      props.defaultData?.pageKey || 'new'
    ].join(':')
  )

  const {
    modelValue,
    dialogType,
    pageData,
    appKey,
    initialParentPageKey,
    initialParentMenuId,
    initialPageType,
    defaultData
  } = toRefs(props)
</script>
