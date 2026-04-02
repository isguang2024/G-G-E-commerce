import { useQuery } from '@tanstack/react-query'
import { fetchMenuManageGroups, fetchMenuTree, fetchRuntimePages } from '@/shared/api/modules/menu.api'
import { queryKeys } from '@/shared/api/query-keys'
import type { MenuNode, MenuNodeDetail, MenuPageBinding } from '@/shared/types/menu'

function flattenMenuTree(items: MenuNode[]): MenuNode[] {
  return items.flatMap((item) => [item, ...flattenMenuTree(item.children)])
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

export function useMenuTreeQuery(spaceKey: string) {
  return useQuery({
    queryKey: queryKeys.menu.tree(spaceKey),
    queryFn: () => fetchMenuTree(spaceKey),
    enabled: Boolean(spaceKey),
  })
}

export function useMenuManageGroupsQuery() {
  return useQuery({
    queryKey: queryKeys.menu.groups,
    queryFn: fetchMenuManageGroups,
  })
}

export function useRuntimePagesQuery(spaceKey: string) {
  return useQuery({
    queryKey: queryKeys.menu.runtimePages(spaceKey),
    queryFn: () => fetchRuntimePages(spaceKey),
    enabled: Boolean(spaceKey),
  })
}
