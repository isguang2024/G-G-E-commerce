<template>
  <ElDialog v-model="visible" :title="`团队角色菜单权限 - ${roleTitle}`" width="520px" align-center @close="handleClose">
    <ElScrollbar height="70vh">
      <ElTree
        ref="treeRef"
        :data="menuList"
        show-checkbox
        node-key="id"
        :default-expand-all="expandAll"
        :props="defaultProps"
      />
    </ElScrollbar>
    <template #footer>
      <ElButton @click="toggleExpand">{{ expandAll ? '全部收起' : '全部展开' }}</ElButton>
      <ElButton @click="handleClose">取消</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { fetchGetMenuTreeAll } from '@/api/system-manage'
  import { fetchGetMyTeamRoleMenus, fetchSetMyTeamRoleMenus } from '@/api/team'
  import { formatMenuTitle } from '@/utils/router'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
  }

  const props = defineProps<Props>()
  const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void; (e: 'success'): void }>()

  const treeRef = ref()
  const expandAll = ref(true)
  const saving = ref(false)
  const menuList = ref<any[]>([])
  const roleTitle = computed(() => props.roleData?.roleName || '')

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  const defaultProps = {
    children: 'children',
    label: (data: any) => formatMenuTitle(data.meta?.title) || data.label || data.name || ''
  }

  watch(
    () => props.modelValue,
    async (open) => {
      if (!open || !props.roleData?.roleId) return
      try {
        const [menus, assigned] = await Promise.all([
          fetchGetMenuTreeAll(),
          fetchGetMyTeamRoleMenus(props.roleData.roleId)
        ])
        menuList.value = Array.isArray(menus) ? menus : []
        await nextTick()
        treeRef.value?.setCheckedKeys(assigned?.menu_ids || [])
      } catch (error: any) {
        ElMessage.error(error?.message || '加载团队角色菜单权限失败')
      }
    }
  )

  function toggleExpand() {
    const tree = treeRef.value
    if (!tree?.store?.nodesMap) return
    Object.values(tree.store.nodesMap).forEach((node: any) => {
      node.expanded = !expandAll.value
    })
    expandAll.value = !expandAll.value
  }

  function handleClose() {
    visible.value = false
    treeRef.value?.setCheckedKeys([])
  }

  async function handleSave() {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      await fetchSetMyTeamRoleMenus(props.roleData.roleId, treeRef.value?.getCheckedKeys?.() || [])
      ElMessage.success('团队角色菜单权限已保存')
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队角色菜单权限失败')
    } finally {
      saving.value = false
    }
  }
</script>
