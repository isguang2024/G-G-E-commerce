<template>
  <ElDialog
    v-model="visible"
    :title="`功能包绑定菜单 - ${packageName}`"
    width="560px"
    align-center
    destroy-on-close
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里不是编辑菜单定义，而是从全量菜单中选择并绑定到当前功能包。菜单定义仍在全量菜单管理页维护；功能包只负责决定当前上下文是否拥有这些模块入口。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>功能包 {{ packageName }}</ElTag>
        <ElTag type="warning" effect="plain" round>上下文 {{ formatContextType(contextType) }}</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ checkedCount }}</ElTag>
      </div>

      <ElScrollbar height="68vh">
        <ElTree
          ref="treeRef"
          :data="menuList"
          show-checkbox
          node-key="id"
          :default-expand-all="expandAll"
          :props="defaultProps"
        />
      </ElScrollbar>
    </div>

    <template #footer>
      <ElButton @click="toggleExpand">{{ expandAll ? '全部收起' : '全部展开' }}</ElButton>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { fetchGetMenuTreeAll, fetchGetFeaturePackageMenus, fetchSetFeaturePackageMenus } from '@/api/system-manage'
  import { formatMenuTitle } from '@/utils/router'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    contextType?: 'platform' | 'team' | string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    packageId: '',
    packageName: '',
    contextType: 'team'
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }>()

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const treeRef = ref()
  const loading = ref(false)
  const saving = ref(false)
  const expandAll = ref(true)
  const menuList = ref<any[]>([])
  const checkedCount = ref(0)

  const defaultProps = {
    children: 'children',
    label: (data: any) => formatMenuTitle(data.meta?.title) || data.label || data.name || ''
  }

  watch(
    () => props.modelValue,
    async (open) => {
      if (open) {
        await loadData()
      }
    }
  )

  async function loadData() {
    if (!props.packageId) return
    loading.value = true
    try {
      const [menus, assigned] = await Promise.all([
        fetchGetMenuTreeAll(),
        fetchGetFeaturePackageMenus(props.packageId)
      ])
      menuList.value = Array.isArray(menus) ? menus : []
      checkedCount.value = assigned?.menu_ids?.length || 0
      await nextTick()
      treeRef.value?.setCheckedKeys(assigned?.menu_ids || [])
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包绑定菜单失败')
    } finally {
      loading.value = false
    }
  }

  function toggleExpand() {
    const tree = treeRef.value
    if (!tree?.store?.nodesMap) return
    Object.values(tree.store.nodesMap).forEach((node: any) => {
      node.expanded = !expandAll.value
    })
    expandAll.value = !expandAll.value
  }

  async function handleSave() {
    if (!props.packageId) return
    saving.value = true
    try {
      const menuIds = treeRef.value?.getCheckedKeys?.() || []
      await fetchSetFeaturePackageMenus(props.packageId, menuIds)
      checkedCount.value = menuIds.length
      ElMessage.success('功能包绑定菜单已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存功能包绑定菜单失败')
    } finally {
      saving.value = false
    }
  }

  function formatContextType(contextType?: string) {
    if (contextType === 'platform') return '平台'
    if (contextType === 'team') return '团队'
    if (contextType === 'platform,team') return '平台/团队'
    return contextType || '-'
  }
</script>

<style scoped lang="scss">
  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .dialog-note {
    color: #6b7280;
    line-height: 1.6;
  }

  .summary-card {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
