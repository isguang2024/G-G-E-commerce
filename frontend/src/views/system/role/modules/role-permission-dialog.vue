<template>
  <ElDialog
    v-model="visible"
    title="菜单权限"
    width="520px"
    align-center
    class="el-dialog-border"
    @close="handleClose"
  >
    <div v-if="menuLoadError" class="text-red-500 py-2">{{ menuLoadError }}</div>
    <ElScrollbar v-if="!menuLoadError" height="70vh">
      <ElTree
        ref="treeRef"
        :data="processedMenuList"
        show-checkbox
        node-key="id"
        :default-expand-all="isExpandAll"
        :props="defaultProps"
        @check="handleTreeCheck"
      >
        <template #default="{ data }">
          <div style="display: flex; align-items: center">
            <span v-if="data.isAuth">{{ data.label }}</span>
            <span v-else>{{ defaultProps.label(data) }}</span>
          </div>
        </template>
      </ElTree>
    </ElScrollbar>
    <template #footer>
      <ElButton @click="toggleExpandAll">{{ isExpandAll ? '全部收起' : '全部展开' }}</ElButton>
      <ElButton @click="toggleSelectAll" style="margin-left: 8px">{{
        isSelectAll ? '取消全选' : '全部选择'
      }}</ElButton>
      <ElButton type="primary" :loading="saving" @click="savePermission"> 保存 </ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { fetchGetMenuTreeAll, fetchGetRoleMenus, fetchSetRoleMenus } from '@/api/system-manage'
  import { formatMenuTitle } from '@/utils/router'

  type RoleListItem = Api.SystemManage.RoleListItem

  interface Props {
    modelValue: boolean
    roleData?: RoleListItem
  }

  interface Emits {
    (e: 'update:modelValue', value: boolean): void
    (e: 'success'): void
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    roleData: undefined
  })

  const emit = defineEmits<Emits>()

  const treeRef = ref()
  const isExpandAll = ref(true)
  const isSelectAll = ref(false)
  const saving = ref(false)
  const menuTreeRaw = ref<Array<Record<string, any>>>([])
  const menuLoadError = ref('')

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })

  interface MenuNode {
    id?: string
    name?: string
    label?: string
    meta?: { title?: string; authList?: Array<{ authMark: string; title: string }> }
    children?: MenuNode[]
    isAuth?: boolean
    [key: string]: any
  }

  /** 后端菜单树 + 将 authList 展开为展示用子节点（保存时只提交菜单 id） */
  const processedMenuList = computed(() => {
    const processNode = (node: MenuNode): MenuNode => {
      const processed = { ...node }
      if (!processed.id && node.path) processed.id = node.path

      if (node.meta?.authList?.length) {
        const authNodes = node.meta.authList.map((auth) => ({
          id: `${node.id}_auth_${auth.authMark}`,
          name: `${node.name}_auth_${auth.authMark}`,
          label: auth.title,
          isAuth: true
        }))
        processed.children = processed.children ? [...processed.children, ...authNodes] : authNodes
      }
      if (processed.children) {
        processed.children = processed.children.map(processNode)
      }
      return processed
    }
    return (menuTreeRaw.value as MenuNode[]).map(processNode)
  })

  const uuidRe = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

  const collectDescendantMenuIds = (node: MenuNode): string[] => {
    const ids: string[] = []
    const traverse = (children?: MenuNode[]) => {
      children?.forEach((child) => {
        if (child.id && uuidRe.test(child.id)) ids.push(child.id)
        traverse(child.children)
      })
    }
    traverse(node.children)
    return ids
  }

  const normalizeMenuIds = (menuIds: string[]): string[] => {
    const rawSet = new Set(menuIds.filter((id) => typeof id === 'string' && uuidRe.test(id)))
    const normalized = new Set(rawSet)

    const visit = (nodes: MenuNode[]) => {
      nodes.forEach((node) => {
        if (!node.id || !rawSet.has(node.id)) {
          if (node.children?.length) visit(node.children)
          return
        }

        const descendantIds = collectDescendantMenuIds(node)
        if (descendantIds.some((id) => rawSet.has(id))) {
          normalized.delete(node.id)
        }

        if (node.children?.length) visit(node.children)
      })
    }

    visit(processedMenuList.value)
    return Array.from(normalized)
  }

  const defaultProps = {
    children: 'children',
    label: (data: any) => formatMenuTitle(data.meta?.title) || data.label || data.name || ''
  }

  /** 弹窗打开时：拉取完整菜单树 + 该角色已选菜单 ID */
  watch(
    () => [props.modelValue, props.roleData] as const,
    async ([open, role]) => {
      if (!open) return
      menuLoadError.value = ''
      try {
        const list = await fetchGetMenuTreeAll()
        menuTreeRaw.value = Array.isArray(list) ? list : []
      } catch (e: any) {
        menuTreeRaw.value = []
        const msg = e?.message || e?.msg || ''
        menuLoadError.value = msg
          ? `加载菜单树失败：${msg}`
          : '加载菜单树失败，请稍后重试（请确认后端已启动且已执行数据库迁移）'
        return
      }
      if (!role?.roleId) return
      try {
        const res = await fetchGetRoleMenus(role.roleId)
        const ids = normalizeMenuIds(res?.menu_ids ?? [])
        // 等树渲染完成（含新增菜单节点）后再设置勾选，避免新节点未参与导致保存时漏传
        await nextTick()
        await nextTick()
        treeRef.value?.setCheckedKeys(Array.isArray(ids) ? ids : [])
        syncSelectAllState()
      } catch {
        nextTick(() => treeRef.value?.setCheckedKeys([]))
      }
    }
  )

  const handleClose = () => {
    visible.value = false
    treeRef.value?.setCheckedKeys([])
    menuLoadError.value = ''
  }

  /** 仅收集菜单 ID（UUID 格式），排除 auth 子节点 */
  const collectMenuIds = (): string[] => {
    const tree = treeRef.value
    if (!tree) return []
    const checked = (tree.getCheckedKeys() || []) as string[]
    return normalizeMenuIds(checked)
  }

  const savePermission = async () => {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      const menuIds = collectMenuIds()
      await fetchSetRoleMenus(props.roleData.roleId, menuIds)
      ElMessage.success('权限保存成功')
      emit('success')
      handleClose()
    } catch (e) {
      ElMessage.error((e as Error)?.message || '保存失败')
    } finally {
      saving.value = false
    }
  }

  const toggleExpandAll = () => {
    const tree = treeRef.value
    if (!tree?.store?.nodesMap) return
    Object.values(tree.store.nodesMap).forEach((node: any) => {
      node.expanded = !isExpandAll.value
    })
    isExpandAll.value = !isExpandAll.value
  }

  const getAllNodeIds = (nodes: MenuNode[]): string[] => {
    const keys: string[] = []
    const traverse = (list: MenuNode[]) => {
      list.forEach((node) => {
        if (node.id) keys.push(node.id)
        if (node.children?.length) traverse(node.children)
      })
    }
    traverse(nodes)
    return keys
  }

  const syncSelectAllState = () => {
    const tree = treeRef.value
    if (!tree) return
    const checked = (tree.getCheckedKeys() || []) as string[]
    const halfChecked = (tree.getHalfCheckedKeys() || []) as string[]
    const allIds = getAllNodeIds(processedMenuList.value)
    const menuIds = allIds.filter((id) => uuidRe.test(id))
    const totalChecked = new Set([...checked, ...halfChecked])
    isSelectAll.value = menuIds.length > 0 && totalChecked.size >= menuIds.length
  }

  const toggleSelectAll = () => {
    const tree = treeRef.value
    if (!tree) return
    const allIds = getAllNodeIds(processedMenuList.value)
    const uuidRe = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
    const menuIds = allIds.filter((id) => uuidRe.test(id))
    if (!isSelectAll.value) {
      tree.setCheckedKeys(menuIds)
    } else {
      tree.setCheckedKeys([])
    }
    isSelectAll.value = !isSelectAll.value
  }

  const handleTreeCheck = () => {
    syncSelectAllState()
  }
</script>
