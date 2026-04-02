import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  addPermissionActionEndpoint,
  assignUserRoles,
  createFeaturePackage,
  createPermissionAction,
  createPermissionGroup,
  createRole,
  createUser,
  deleteFeaturePackage,
  deletePermissionAction,
  deleteRole,
  deleteUser,
  fetchFeaturePackageActions,
  fetchFeaturePackageChildren,
  fetchFeaturePackageDetail,
  fetchFeaturePackageImpactPreview,
  fetchFeaturePackageList,
  fetchFeaturePackageMenus,
  fetchFeaturePackageRelationTree,
  fetchFeaturePackageTeams,
  fetchPermissionActionConsumers,
  fetchPermissionActionDetail,
  fetchPermissionActionEndpoints,
  fetchPermissionActionList,
  fetchPermissionGroups,
  fetchRoleActions,
  fetchRoleDetail,
  fetchRoleList,
  fetchRoleMenus,
  fetchRolePackages,
  fetchUserDetail,
  fetchUserList,
  fetchUserMenus,
  fetchUserPackages,
  fetchUserPermissionDiagnosis,
  refreshUserPermissionSnapshot,
  removePermissionActionEndpoint,
  updateFeaturePackage,
  updateFeaturePackageActions,
  updateFeaturePackageChildren,
  updateFeaturePackageMenus,
  updateFeaturePackageTeams,
  updatePermissionAction,
  updatePermissionGroup,
  updateRole,
  updateRoleActions,
  updateRoleMenus,
  updateRolePackages,
  updateUser,
  updateUserMenus,
  updateUserPackages,
} from '@/shared/api/modules/access.api'
import { queryKeys } from '@/shared/api/query-keys'
import type {
  FeaturePackageSavePayload,
  PermissionActionSavePayload,
  PermissionGroupSummary,
  RoleSavePayload,
  UserSavePayload,
} from '@/shared/types/admin'

function serializeFilters(filters?: Record<string, unknown>) {
  return JSON.stringify(filters || {})
}

export function useRoleListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.role.list(serializeFilters(filters)),
    queryFn: () => fetchRoleList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useRoleDetailQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.role.detail(roleId || ''),
    queryFn: () => fetchRoleDetail(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

export function useRolePackagesQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.role.packages(roleId || ''),
    queryFn: () => fetchRolePackages(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

export function useRoleActionsQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.role.actions(roleId || ''),
    queryFn: () => fetchRoleActions(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

export function useRoleMenusQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.role.menus(roleId || ''),
    queryFn: () => fetchRoleMenus(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

function invalidateRoleQueries(client: ReturnType<typeof useQueryClient>, roleId?: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: ['role', 'list'] }),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.role.detail(roleId) }) : Promise.resolve(),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.role.packages(roleId) }) : Promise.resolve(),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.role.actions(roleId) }) : Promise.resolve(),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.role.menus(roleId) }) : Promise.resolve(),
  ])
}

export function useCreateRoleMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createRole,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['role', 'list'] })
    },
  })
}

export function useUpdateRoleMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: RoleSavePayload) => updateRole(roleId, payload),
    onSuccess: async () => {
      await invalidateRoleQueries(client, roleId)
    },
  })
}

export function useDeleteRoleMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deleteRole,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['role', 'list'] })
    },
  })
}

export function useUpdateRolePackagesMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (packageIds: string[]) => updateRolePackages(roleId, packageIds),
    onSuccess: async () => {
      await invalidateRoleQueries(client, roleId)
    },
  })
}

export function useUpdateRoleActionsMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (actionIds: string[]) => updateRoleActions(roleId, actionIds),
    onSuccess: async () => {
      await invalidateRoleQueries(client, roleId)
    },
  })
}

export function useUpdateRoleMenusMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (menuIds: string[]) => updateRoleMenus(roleId, menuIds),
    onSuccess: async () => {
      await invalidateRoleQueries(client, roleId)
    },
  })
}

export function useUserListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.user.list(serializeFilters(filters)),
    queryFn: () => fetchUserList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useUserDetailQuery(userId?: string | null) {
  return useQuery({
    queryKey: queryKeys.user.detail(userId || ''),
    queryFn: () => fetchUserDetail(userId!),
    enabled: Boolean(userId),
    placeholderData: (previousData) => previousData,
  })
}

export function useUserPackagesQuery(userId?: string | null) {
  return useQuery({
    queryKey: queryKeys.user.packages(userId || ''),
    queryFn: () => fetchUserPackages(userId!),
    enabled: Boolean(userId),
    placeholderData: (previousData) => previousData,
  })
}

export function useUserMenusQuery(userId?: string | null) {
  return useQuery({
    queryKey: queryKeys.user.menus(userId || ''),
    queryFn: () => fetchUserMenus(userId!),
    enabled: Boolean(userId),
    placeholderData: (previousData) => previousData,
  })
}

export function useUserPermissionDiagnosisQuery(userId?: string | null) {
  return useQuery({
    queryKey: queryKeys.user.permissionDiagnosis(userId || ''),
    queryFn: () => fetchUserPermissionDiagnosis(userId!),
    enabled: Boolean(userId),
    placeholderData: (previousData) => previousData,
  })
}

function invalidateUserQueries(client: ReturnType<typeof useQueryClient>, userId?: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: ['user', 'list'] }),
    userId ? client.invalidateQueries({ queryKey: queryKeys.user.detail(userId) }) : Promise.resolve(),
    userId ? client.invalidateQueries({ queryKey: queryKeys.user.packages(userId) }) : Promise.resolve(),
    userId ? client.invalidateQueries({ queryKey: queryKeys.user.menus(userId) }) : Promise.resolve(),
    userId ? client.invalidateQueries({ queryKey: queryKeys.user.permissionDiagnosis(userId) }) : Promise.resolve(),
  ])
}

export function useCreateUserMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createUser,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['user', 'list'] })
    },
  })
}

export function useUpdateUserMutation(userId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: UserSavePayload) => updateUser(userId, payload),
    onSuccess: async () => {
      await invalidateUserQueries(client, userId)
    },
  })
}

export function useDeleteUserMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deleteUser,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['user', 'list'] })
    },
  })
}

export function useUpdateUserPackagesMutation(userId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (packageIds: string[]) => updateUserPackages(userId, packageIds),
    onSuccess: async () => {
      await invalidateUserQueries(client, userId)
    },
  })
}

export function useUpdateUserMenusMutation(userId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (menuIds: string[]) => updateUserMenus(userId, menuIds),
    onSuccess: async () => {
      await invalidateUserQueries(client, userId)
    },
  })
}

export function useAssignUserRolesMutation(userId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (roleIds: string[]) => assignUserRoles(userId, roleIds),
    onSuccess: async () => {
      await invalidateUserQueries(client, userId)
    },
  })
}

export function useRefreshUserPermissionSnapshotMutation(userId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: () => refreshUserPermissionSnapshot(userId),
    onSuccess: async () => {
      await invalidateUserQueries(client, userId)
    },
  })
}

export function usePermissionActionListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.permission.list(serializeFilters(filters)),
    queryFn: () => fetchPermissionActionList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function usePermissionGroupsQuery() {
  return useQuery({
    queryKey: queryKeys.permission.groups,
    queryFn: fetchPermissionGroups,
    placeholderData: (previousData) => previousData,
  })
}

export function usePermissionActionDetailQuery(actionId?: string | null) {
  return useQuery({
    queryKey: queryKeys.permission.detail(actionId || ''),
    queryFn: () => fetchPermissionActionDetail(actionId!),
    enabled: Boolean(actionId),
    placeholderData: (previousData) => previousData,
  })
}

export function usePermissionActionEndpointsQuery(actionId?: string | null) {
  return useQuery({
    queryKey: queryKeys.permission.endpoints(actionId || ''),
    queryFn: () => fetchPermissionActionEndpoints(actionId!),
    enabled: Boolean(actionId),
    placeholderData: (previousData) => previousData,
  })
}

export function usePermissionActionConsumersQuery(actionId?: string | null) {
  return useQuery({
    queryKey: queryKeys.permission.consumers(actionId || ''),
    queryFn: () => fetchPermissionActionConsumers(actionId!),
    enabled: Boolean(actionId),
    placeholderData: (previousData) => previousData,
  })
}

function invalidatePermissionQueries(client: ReturnType<typeof useQueryClient>, actionId?: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: ['permission', 'list'] }),
    client.invalidateQueries({ queryKey: queryKeys.permission.groups }),
    actionId ? client.invalidateQueries({ queryKey: queryKeys.permission.detail(actionId) }) : Promise.resolve(),
    actionId ? client.invalidateQueries({ queryKey: queryKeys.permission.endpoints(actionId) }) : Promise.resolve(),
    actionId ? client.invalidateQueries({ queryKey: queryKeys.permission.consumers(actionId) }) : Promise.resolve(),
  ])
}

export function useCreatePermissionGroupMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createPermissionGroup,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.permission.groups })
    },
  })
}

export function useUpdatePermissionGroupMutation(groupId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: Partial<PermissionGroupSummary>) => updatePermissionGroup(groupId, payload),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: queryKeys.permission.groups })
    },
  })
}

export function useCreatePermissionActionMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createPermissionAction,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['permission', 'list'] })
    },
  })
}

export function useUpdatePermissionActionMutation(actionId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: PermissionActionSavePayload) => updatePermissionAction(actionId, payload),
    onSuccess: async () => {
      await invalidatePermissionQueries(client, actionId)
    },
  })
}

export function useDeletePermissionActionMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deletePermissionAction,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['permission', 'list'] })
    },
  })
}

export function useAddPermissionActionEndpointMutation(actionId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (endpointCode: string) => addPermissionActionEndpoint(actionId, endpointCode),
    onSuccess: async () => {
      await invalidatePermissionQueries(client, actionId)
    },
  })
}

export function useRemovePermissionActionEndpointMutation(actionId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (endpointCode: string) => removePermissionActionEndpoint(actionId, endpointCode),
    onSuccess: async () => {
      await invalidatePermissionQueries(client, actionId)
    },
  })
}

export function useFeaturePackageListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.featurePackage.list(serializeFilters(filters)),
    queryFn: () => fetchFeaturePackageList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageDetailQuery(packageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.featurePackage.detail(packageId || ''),
    queryFn: () => fetchFeaturePackageDetail(packageId!),
    enabled: Boolean(packageId),
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageChildrenQuery(packageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.featurePackage.children(packageId || ''),
    queryFn: () => fetchFeaturePackageChildren(packageId!),
    enabled: Boolean(packageId),
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageActionsQuery(packageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.featurePackage.actions(packageId || ''),
    queryFn: () => fetchFeaturePackageActions(packageId!),
    enabled: Boolean(packageId),
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageMenusQuery(packageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.featurePackage.menus(packageId || ''),
    queryFn: () => fetchFeaturePackageMenus(packageId!),
    enabled: Boolean(packageId),
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageTeamsQuery(packageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.featurePackage.teams(packageId || ''),
    queryFn: () => fetchFeaturePackageTeams(packageId!),
    enabled: Boolean(packageId),
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageRelationTreeQuery() {
  return useQuery({
    queryKey: queryKeys.featurePackage.relationTree,
    queryFn: fetchFeaturePackageRelationTree,
    placeholderData: (previousData) => previousData,
  })
}

export function useFeaturePackageImpactPreviewQuery(packageId?: string | null) {
  return useQuery({
    queryKey: queryKeys.featurePackage.impactPreview(packageId || ''),
    queryFn: () => fetchFeaturePackageImpactPreview(packageId!),
    enabled: Boolean(packageId),
    placeholderData: (previousData) => previousData,
  })
}

function invalidateFeaturePackageQueries(client: ReturnType<typeof useQueryClient>, packageId?: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: ['featurePackage', 'list'] }),
    client.invalidateQueries({ queryKey: queryKeys.featurePackage.relationTree }),
    packageId ? client.invalidateQueries({ queryKey: queryKeys.featurePackage.detail(packageId) }) : Promise.resolve(),
    packageId ? client.invalidateQueries({ queryKey: queryKeys.featurePackage.children(packageId) }) : Promise.resolve(),
    packageId ? client.invalidateQueries({ queryKey: queryKeys.featurePackage.actions(packageId) }) : Promise.resolve(),
    packageId ? client.invalidateQueries({ queryKey: queryKeys.featurePackage.menus(packageId) }) : Promise.resolve(),
    packageId ? client.invalidateQueries({ queryKey: queryKeys.featurePackage.teams(packageId) }) : Promise.resolve(),
    packageId ? client.invalidateQueries({ queryKey: queryKeys.featurePackage.impactPreview(packageId) }) : Promise.resolve(),
  ])
}

export function useCreateFeaturePackageMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createFeaturePackage,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['featurePackage', 'list'] })
    },
  })
}

export function useUpdateFeaturePackageMutation(packageId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: FeaturePackageSavePayload) => updateFeaturePackage(packageId, payload),
    onSuccess: async () => {
      await invalidateFeaturePackageQueries(client, packageId)
    },
  })
}

export function useDeleteFeaturePackageMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deleteFeaturePackage,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['featurePackage', 'list'] })
    },
  })
}

export function useUpdateFeaturePackageChildrenMutation(packageId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (packageIds: string[]) => updateFeaturePackageChildren(packageId, packageIds),
    onSuccess: async () => {
      await invalidateFeaturePackageQueries(client, packageId)
    },
  })
}

export function useUpdateFeaturePackageActionsMutation(packageId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (actionIds: string[]) => updateFeaturePackageActions(packageId, actionIds),
    onSuccess: async () => {
      await invalidateFeaturePackageQueries(client, packageId)
    },
  })
}

export function useUpdateFeaturePackageMenusMutation(packageId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (menuIds: string[]) => updateFeaturePackageMenus(packageId, menuIds),
    onSuccess: async () => {
      await invalidateFeaturePackageQueries(client, packageId)
    },
  })
}

export function useUpdateFeaturePackageTeamsMutation(packageId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (teamIds: string[]) => updateFeaturePackageTeams(packageId, teamIds),
    onSuccess: async () => {
      await invalidateFeaturePackageQueries(client, packageId)
    },
  })
}
