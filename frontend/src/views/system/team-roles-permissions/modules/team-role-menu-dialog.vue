<template>
  <ElDialog v-model="visible" :title="`团队角色菜单权限 - ${roleTitle}`" width="620px" align-center @close="handleClose">
    <div class="dialog-shell">
      <div class="dialog-note">
        {{
          props.roleData?.isGlobal
            ? '基础团队角色默认继承当前团队功能包的菜单范围，这里只读查看最终角色菜单。'
            : '请先绑定角色功能包。这里只展示当前角色功能包展开范围内可配置的菜单。'
        }}
      </div>
      <div class="summary-row">
        <ElTag effect="plain" round>角色 {{ roleTitle }}</ElTag>
        <ElTag type="success" effect="plain" round>功能包 {{ featurePackages.length }}</ElTag>
        <ElTag type="primary" effect="plain" round>{{ inherited ? '继承团队功能包' : '角色独立功能包' }}</ElTag>
        <ElTag type="warning" effect="plain" round>可配菜单 {{ availableMenuIds.length }}</ElTag>
      </div>
      <div v-if="featurePackages.length" class="package-tags">
        <ElTag v-for="item in featurePackages" :key="item.id" type="success" effect="plain" round>
          {{ item.name }}
        </ElTag>
      </div>
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
    </div>
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
  import { fetchGetMyTeamRoleMenus, fetchGetMyTeamRolePackages, fetchSetMyTeamRoleMenus } from '@/api/team'
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
  const availableMenuIds = ref<string[]>([])
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const inherited = ref(false)
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
        const [menus, assigned, packagesRes] = await Promise.all([
          fetchGetMenuTreeAll(),
          fetchGetMyTeamRoleMenus(props.roleData.roleId),
          fetchGetMyTeamRolePackages(props.roleData.roleId)
        ])
        availableMenuIds.value = assigned?.available_menu_ids || []
        featurePackages.value = packagesRes?.packages || []
        inherited.value = Boolean(packagesRes?.inherited)
        menuList.value = filterMenuTreeByAllowedIds(Array.isArray(menus) ? menus : [], new Set(availableMenuIds.value))
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

  function filterMenuTreeByAllowedIds(source: any[], allowed: Set<string>): any[] {
    if (!allowed.size) return []
    return source
      .map((item: any) => {
        const children: any[] = filterMenuTreeByAllowedIds(item.children || [], allowed)
        if (!allowed.has(item.id) && children.length === 0) return null
        return {
          ...item,
          children
        }
      })
      .filter(Boolean) as any[]
  }
</script>

<style scoped lang="scss">
  .dialog-shell {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .dialog-note {
    color: #6b7280;
    line-height: 1.6;
  }

  .summary-row,
  .package-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
</style>
