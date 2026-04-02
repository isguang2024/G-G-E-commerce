import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  addMyTeamMember,
  addTeamMember,
  createMyTeamBoundaryRole,
  createTeam,
  deleteMyTeamBoundaryRole,
  deleteTeam,
  fetchMyTeamActionOrigins,
  fetchMyTeamBoundaryRoleActions,
  fetchMyTeamBoundaryRoleMenus,
  fetchMyTeamBoundaryRolePackages,
  fetchMyTeamBoundaryRoles,
  fetchMyTeamMemberRoles,
  fetchMyTeamMembers,
  fetchMyTeamMenuOrigins,
  fetchTeamDetail,
  fetchTeamList,
  fetchTeamMembers,
  removeMyTeamMember,
  removeTeamMember,
  setMyTeamBoundaryRoleActions,
  setMyTeamBoundaryRoleMenus,
  setMyTeamBoundaryRolePackages,
  setMyTeamMemberRoles,
  updateMyTeamBoundaryRole,
  updateMyTeamMemberRole,
  updateTeam,
  updateTeamMemberRole,
} from '@/shared/api/modules/team.api'
import { queryKeys } from '@/shared/api/query-keys'
import type { TeamSavePayload } from '@/shared/types/admin'

function serializeFilters(filters?: Record<string, unknown>) {
  return JSON.stringify(filters || {})
}

export function useTeamListQuery(filters?: Record<string, unknown>) {
  return useQuery({
    queryKey: queryKeys.team.list(serializeFilters(filters)),
    queryFn: () => fetchTeamList(filters),
    placeholderData: (previousData) => previousData,
  })
}

export function useTeamDetailQuery(teamId?: string | null) {
  return useQuery({
    queryKey: queryKeys.team.detail(teamId || ''),
    queryFn: () => fetchTeamDetail(teamId!),
    enabled: Boolean(teamId),
    placeholderData: (previousData) => previousData,
  })
}

export function useTeamMembersQuery(teamId?: string | null) {
  return useQuery({
    queryKey: queryKeys.team.members(teamId || ''),
    queryFn: () => fetchTeamMembers(teamId!),
    enabled: Boolean(teamId),
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamMembersQuery() {
  return useQuery({
    queryKey: queryKeys.team.myTeamMembers,
    queryFn: fetchMyTeamMembers,
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamMemberRolesQuery(userId?: string | null) {
  return useQuery({
    queryKey: ['team', 'myTeamMemberRoles', userId || ''],
    queryFn: () => fetchMyTeamMemberRoles(userId!),
    enabled: Boolean(userId),
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamBoundaryRolesQuery() {
  return useQuery({
    queryKey: queryKeys.team.myTeamBoundaryRoles,
    queryFn: fetchMyTeamBoundaryRoles,
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamActionOriginsQuery() {
  return useQuery({
    queryKey: queryKeys.team.myTeamActionOrigins,
    queryFn: fetchMyTeamActionOrigins,
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamMenuOriginsQuery() {
  return useQuery({
    queryKey: queryKeys.team.myTeamMenuOrigins,
    queryFn: fetchMyTeamMenuOrigins,
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamBoundaryRoleActionsQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.team.myTeamBoundaryRoleActions(roleId || ''),
    queryFn: () => fetchMyTeamBoundaryRoleActions(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamBoundaryRoleMenusQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.team.myTeamBoundaryRoleMenus(roleId || ''),
    queryFn: () => fetchMyTeamBoundaryRoleMenus(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

export function useMyTeamBoundaryRolePackagesQuery(roleId?: string | null) {
  return useQuery({
    queryKey: queryKeys.team.myTeamBoundaryRolePackages(roleId || ''),
    queryFn: () => fetchMyTeamBoundaryRolePackages(roleId!),
    enabled: Boolean(roleId),
    placeholderData: (previousData) => previousData,
  })
}

function invalidateTeamQueries(client: ReturnType<typeof useQueryClient>, teamId?: string, roleId?: string, memberUserId?: string) {
  return Promise.all([
    client.invalidateQueries({ queryKey: ['team', 'list'] }),
    teamId ? client.invalidateQueries({ queryKey: queryKeys.team.detail(teamId) }) : Promise.resolve(),
    teamId ? client.invalidateQueries({ queryKey: queryKeys.team.members(teamId) }) : Promise.resolve(),
    client.invalidateQueries({ queryKey: queryKeys.team.myTeamMembers }),
    client.invalidateQueries({ queryKey: queryKeys.team.myTeamBoundaryRoles }),
    client.invalidateQueries({ queryKey: queryKeys.team.myTeamActionOrigins }),
    client.invalidateQueries({ queryKey: queryKeys.team.myTeamMenuOrigins }),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.team.myTeamBoundaryRoleActions(roleId) }) : Promise.resolve(),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.team.myTeamBoundaryRoleMenus(roleId) }) : Promise.resolve(),
    roleId ? client.invalidateQueries({ queryKey: queryKeys.team.myTeamBoundaryRolePackages(roleId) }) : Promise.resolve(),
    memberUserId ? client.invalidateQueries({ queryKey: ['team', 'myTeamMemberRoles', memberUserId] }) : Promise.resolve(),
  ])
}

export function useCreateTeamMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createTeam,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['team', 'list'] })
    },
  })
}

export function useUpdateTeamMutation(teamId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: TeamSavePayload) => updateTeam(teamId, payload),
    onSuccess: async () => {
      await invalidateTeamQueries(client, teamId)
    },
  })
}

export function useDeleteTeamMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deleteTeam,
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['team', 'list'] })
    },
  })
}

export function useAddTeamMemberMutation(teamId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, roleCode }: { userId: string; roleCode?: string }) => addTeamMember(teamId, userId, roleCode),
    onSuccess: async () => {
      await invalidateTeamQueries(client, teamId)
    },
  })
}

export function useRemoveTeamMemberMutation(teamId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (userId: string) => removeTeamMember(teamId, userId),
    onSuccess: async () => {
      await invalidateTeamQueries(client, teamId)
    },
  })
}

export function useUpdateTeamMemberRoleMutation(teamId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, roleCode }: { userId: string; roleCode: string }) => updateTeamMemberRole(teamId, userId, roleCode),
    onSuccess: async (_result, variables) => {
      await invalidateTeamQueries(client, teamId, undefined, variables.userId)
    },
  })
}

export function useAddMyTeamMemberMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, roleCode }: { userId: string; roleCode?: string }) => addMyTeamMember(userId, roleCode),
    onSuccess: async () => {
      await invalidateTeamQueries(client)
    },
  })
}

export function useRemoveMyTeamMemberMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: removeMyTeamMember,
    onSuccess: async () => {
      await invalidateTeamQueries(client)
    },
  })
}

export function useUpdateMyTeamMemberRoleMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: ({ userId, roleCode }: { userId: string; roleCode: string }) => updateMyTeamMemberRole(userId, roleCode),
    onSuccess: async (_result, variables) => {
      await invalidateTeamQueries(client, undefined, undefined, variables.userId)
    },
  })
}

export function useSetMyTeamMemberRolesMutation(userId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (roleIds: string[]) => setMyTeamMemberRoles(userId, roleIds),
    onSuccess: async () => {
      await invalidateTeamQueries(client, undefined, undefined, userId)
    },
  })
}

export function useCreateMyTeamBoundaryRoleMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: createMyTeamBoundaryRole,
    onSuccess: async () => {
      await invalidateTeamQueries(client)
    },
  })
}

export function useUpdateMyTeamBoundaryRoleMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (payload: { roleName: string; roleCode: string; description: string }) =>
      updateMyTeamBoundaryRole(roleId, payload),
    onSuccess: async () => {
      await invalidateTeamQueries(client, undefined, roleId)
    },
  })
}

export function useDeleteMyTeamBoundaryRoleMutation() {
  const client = useQueryClient()
  return useMutation({
    mutationFn: deleteMyTeamBoundaryRole,
    onSuccess: async () => {
      await invalidateTeamQueries(client)
    },
  })
}

export function useSetMyTeamBoundaryRoleActionsMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (actionIds: string[]) => setMyTeamBoundaryRoleActions(roleId, actionIds),
    onSuccess: async () => {
      await invalidateTeamQueries(client, undefined, roleId)
    },
  })
}

export function useSetMyTeamBoundaryRoleMenusMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (menuIds: string[]) => setMyTeamBoundaryRoleMenus(roleId, menuIds),
    onSuccess: async () => {
      await invalidateTeamQueries(client, undefined, roleId)
    },
  })
}

export function useSetMyTeamBoundaryRolePackagesMutation(roleId: string) {
  const client = useQueryClient()
  return useMutation({
    mutationFn: (packageIds: string[]) => setMyTeamBoundaryRolePackages(roleId, packageIds),
    onSuccess: async () => {
      await invalidateTeamQueries(client, undefined, roleId)
    },
  })
}
