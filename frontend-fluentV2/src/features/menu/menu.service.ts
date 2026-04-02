import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  createMenu,
  deleteMenu,
  fetchMenuDeletePreview,
  fetchMenuManageGroups,
  fetchMenuTree,
  fetchRuntimePages,
  updateMenu,
} from '@/shared/api/modules/menu.api'
import { invalidateSpaceScopedQueries, queryClient } from '@/shared/api/query-client'
import { queryKeys } from '@/shared/api/query-keys'
import type {
  MenuDeletePreview,
  MenuMutationDraft,
  MenuNode,
  MenuNodeDetail,
  MenuPageBinding,
} from '@/shared/types/menu'

export function flattenMenuTree(items: MenuNode[]): MenuNode[] {
  return items.flatMap((item) => [item, ...flattenMenuTree(item.children)])
}

export function findMenuNode(items: MenuNode[], menuId: string | null | undefined) {
  if (!menuId) {
    return null
  }

  return flattenMenuTree(items).find((item) => item.id === menuId) || null
}

export function filterMenuTree(items: MenuNode[], keyword: string): MenuNode[] {
  const query = keyword.trim().toLowerCase()
  if (!query) {
    return items
  }

  return items.reduce<MenuNode[]>((result, item) => {
    const children = filterMenuTree(item.children, query)
    const target = `${item.title} ${item.name} ${item.path} ${item.component} ${item.manageGroup?.name || ''}`.toLowerCase()
    if (target.includes(query) || children.length > 0) {
      result.push({
        ...item,
        children,
      })
    }
    return result
  }, [])
}

export function buildMenuNodeDetail(
  node: MenuNode,
  tree: MenuNode[],
  runtimePages: MenuPageBinding[],
): MenuNodeDetail {
  const flattened = flattenMenuTree(tree)
  const parent = node.parentId
    ? flattened.find((item) => item.id === node.parentId)
    : undefined
  const linkedPages = runtimePages.filter((item) => item.parentMenuId === node.id)
  const permissionKeys = [
    `${node.meta.permissionKey || ''}`.trim(),
    `${node.meta.requiredAction || ''}`.trim(),
    ...linkedPages.map((item) => `${item.permissionKey || ''}`.trim()),
  ].filter(Boolean)

  return {
    ...node,
    parent: parent
      ? {
          id: parent.id,
          title: parent.title,
          path: parent.path,
        }
      : undefined,
    childCount: node.children.length,
    linkedPages,
    permissionKeys: [...new Set(permissionKeys)],
  }
}

export function createMenuDraft(
  input: {
    mode: 'create-root' | 'create-sibling' | 'create-child' | 'edit'
    currentSpaceKey: string
    node?: MenuNode | null
  },
): MenuMutationDraft {
  const node = input.node || null
  const baseSpaceKey = node?.spaceKey || input.currentSpaceKey

  if (input.mode === 'edit' && node) {
    return {
      parentId: node.parentId || null,
      manageGroupId: node.manageGroupId || null,
      spaceKey: baseSpaceKey,
      kind: node.kind,
      path: node.path,
      name: node.name,
      title: node.title,
      component: node.component,
      icon: node.icon,
      sortOrder: node.sortOrder,
      hidden: node.hidden,
      accessMode: `${node.meta.accessMode || 'permission'}`.trim() || 'permission',
      activePath: `${node.meta.activePath || ''}`.trim(),
      externalLink: `${node.meta.link || ''}`.trim(),
      keepAlive: Boolean(node.meta.keepAlive),
      fixedTab: Boolean(node.meta.fixedTab),
      isFullPage: Boolean(node.meta.isFullPage),
    }
  }

  return {
    parentId:
      input.mode === 'create-child'
        ? node?.id || null
        : input.mode === 'create-sibling'
          ? node?.parentId || null
          : null,
    manageGroupId: node?.manageGroupId || null,
    spaceKey: baseSpaceKey,
    kind: node?.kind === 'external' ? 'external' : 'entry',
    path: '',
    name: '',
    title: '',
    component: '',
    icon: node?.icon || '',
    sortOrder: node ? node.sortOrder + 10 : 0,
    hidden: false,
    accessMode: 'permission',
    activePath: '',
    externalLink: '',
    keepAlive: false,
    fixedTab: false,
    isFullPage: false,
  }
}

export function useMenuTreeQuery(spaceKey: string) {
  return useQuery({
    queryKey: queryKeys.menu.tree(spaceKey),
    queryFn: () => fetchMenuTree(spaceKey),
    enabled: Boolean(spaceKey),
    placeholderData: (previousData) => previousData,
  })
}

export function useMenuManageGroupsQuery(spaceKey: string) {
  return useQuery({
    queryKey: queryKeys.menu.manageGroups(spaceKey),
    queryFn: fetchMenuManageGroups,
    enabled: Boolean(spaceKey),
    placeholderData: (previousData) => previousData,
  })
}

export function useRuntimePagesQuery(spaceKey: string) {
  return useQuery({
    queryKey: queryKeys.menu.runtimePages(spaceKey),
    queryFn: () => fetchRuntimePages(spaceKey),
    enabled: Boolean(spaceKey),
    placeholderData: (previousData) => previousData,
  })
}

export function useMenuNodeDetailQuery(spaceKey: string, menuId: string | null) {
  const client = useQueryClient()

  return useQuery({
    queryKey: queryKeys.menu.detail(menuId || '', spaceKey),
    enabled: Boolean(spaceKey && menuId),
    queryFn: async () => {
      const [tree, runtimePages] = await Promise.all([
        client.ensureQueryData({
          queryKey: queryKeys.menu.tree(spaceKey),
          queryFn: () => fetchMenuTree(spaceKey),
        }),
        client.ensureQueryData({
          queryKey: queryKeys.menu.runtimePages(spaceKey),
          queryFn: () => fetchRuntimePages(spaceKey),
        }),
      ])

      const node = findMenuNode(tree, menuId)
      if (!node) {
        throw new Error('当前菜单节点不存在或已被删除')
      }

      return buildMenuNodeDetail(node, tree, runtimePages)
    },
    placeholderData: (previousData) => previousData,
  })
}

export function useMenuNodePagesQuery(spaceKey: string, menuId: string | null) {
  const client = useQueryClient()

  return useQuery({
    queryKey: queryKeys.menu.pages(menuId || '', spaceKey),
    enabled: Boolean(spaceKey && menuId),
    queryFn: async () => {
      const runtimePages = await client.ensureQueryData({
        queryKey: queryKeys.menu.runtimePages(spaceKey),
        queryFn: () => fetchRuntimePages(spaceKey),
      })

      return runtimePages.filter((item) => item.parentMenuId === menuId)
    },
    placeholderData: (previousData) => previousData,
  })
}

export function useMenuDeletePreviewQuery(
  menuId: string | null,
  payload: { mode: MenuDeletePreview['mode']; targetParentId?: string | null },
) {
  return useQuery({
    queryKey: ['menu', 'deletePreview', menuId, payload.mode, payload.targetParentId || ''],
    enabled: Boolean(menuId),
    queryFn: () =>
      fetchMenuDeletePreview(menuId!, {
        mode: payload.mode,
        targetParentId: payload.targetParentId,
      }),
    placeholderData: (previousData) => previousData,
  })
}

async function refreshMenuSpaceQueries(spaceKey: string) {
  await invalidateSpaceScopedQueries(spaceKey)
  await queryClient.invalidateQueries({
    queryKey: queryKeys.navigation.menuSpaces,
  })
}

export function useCreateMenuMutation() {
  return useMutation({
    mutationFn: createMenu,
    onSuccess: async (_result, variables) => {
      await refreshMenuSpaceQueries(variables.spaceKey)
    },
  })
}

export function useUpdateMenuMutation(menuId: string) {
  return useMutation({
    mutationFn: (draft: MenuMutationDraft) => updateMenu(menuId, draft),
    onSuccess: async (_result, variables) => {
      await refreshMenuSpaceQueries(variables.spaceKey)
      await queryClient.invalidateQueries({
        queryKey: queryKeys.menu.detail(menuId, variables.spaceKey),
      })
      await queryClient.invalidateQueries({
        queryKey: queryKeys.menu.pages(menuId, variables.spaceKey),
      })
    },
  })
}

export function useDeleteMenuMutation() {
  return useMutation({
    mutationFn: ({ menuId, mode, targetParentId }: { menuId: string; mode: MenuDeletePreview['mode']; targetParentId?: string | null }) =>
      deleteMenu(menuId, { mode, targetParentId }),
  })
}
