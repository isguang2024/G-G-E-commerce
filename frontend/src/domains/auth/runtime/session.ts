import { useMenuSpaceStore } from '@/domains/app-runtime/menu-space'
import { fetchGetUserInfo } from '@/domains/auth/api'
import {
  registerSessionRuntimeHandlers,
  type RestoreSessionOptions
} from '@/domains/auth/runtime/session-handlers'
import { useUserStore } from '@/domains/auth/store'
import {
  hasPersonalWorkspaceAccessByUserInfo,
  useCollaborationStore
} from '@/store/modules/collaboration'
import { useWorkspaceStore } from '@/store/modules/workspace'

function mapBackendRolesToFrontend(data: {
  roles?: Array<{ code?: string }> | string[]
  is_super_admin?: boolean
}): string[] {
  if (data.is_super_admin) return ['R_SUPER']
  return ['R_USER']
}

function buildFrontendUserInfo(data: Api.Auth.UserInfo): Api.Auth.UserInfo {
  const roles = mapBackendRolesToFrontend(data)
  return {
    ...data,
    userId: data.id,
    userName: data.username || data.email,
    avatar: data.avatar_url,
    roles,
    buttons: data.buttons || [],
    actions: data.actions || []
  }
}

export async function refreshSessionContext(
  options: RestoreSessionOptions = {}
): Promise<Api.Auth.UserInfo> {
  const userStore = useUserStore()
  const collaborationStore = useCollaborationStore()
  const workspaceStore = useWorkspaceStore()
  const menuSpaceStore = useMenuSpaceStore()
  const data = options.prefetchedUser ?? (await fetchGetUserInfo())
  const frontendUserInfo = buildFrontendUserInfo(data)

  userStore.syncLoginUserIdentity(`${frontendUserInfo.userId || frontendUserInfo.id || ''}`.trim())
  userStore.setUserInfo(frontendUserInfo)
  userStore.setLoginStatus(true)
  userStore.checkAndClearWorktabs()

  collaborationStore.setPersonalWorkspaceAccess(
    hasPersonalWorkspaceAccessByUserInfo(frontendUserInfo)
  )
  menuSpaceStore.syncRuntimeHost()

  await Promise.all([
    (async () => {
      await menuSpaceStore.refreshRuntimeConfig(options.forceRefresh !== false)
      await menuSpaceStore.syncResolvedCurrentSpace(options.preferredSpaceKey || '')
    })(),
    collaborationStore.loadMyCollaborations({
      preferredCollaborationId: data.current_collaboration_workspace_id || '',
      preferredLegacyCollaborationId:
        data.collaboration_workspace_id || data.current_collaboration_workspace_id || '',
      preferredWorkspaceId: data.current_auth_workspace_id || '',
      preferredWorkspaceType: data.current_auth_workspace_type || '',
      preferPersonalWorkspace: hasPersonalWorkspaceAccessByUserInfo(frontendUserInfo)
    })
  ])

  if (
    !options.skipWorkspaceReconcile &&
    (workspaceStore.currentAuthWorkspaceType !== (data.current_auth_workspace_type || 'personal') ||
      workspaceStore.currentAuthWorkspaceId !== (data.current_auth_workspace_id || ''))
  ) {
    return refreshSessionContext({
      ...options,
      prefetchedUser: data,
      skipWorkspaceReconcile: true
    })
  }

  return data
}

export async function restoreSession(
  options: RestoreSessionOptions = {}
): Promise<Api.Auth.UserInfo | null> {
  const userStore = useUserStore()
  if (!userStore.isLogin || !userStore.accessToken) {
    return null
  }
  return refreshSessionContext(options)
}

registerSessionRuntimeHandlers({
  restoreSession
})
