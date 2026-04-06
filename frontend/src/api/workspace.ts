import request from '@/utils/http'

const WORKSPACE_BASE = '/api/v1/workspaces'

function normalizeWorkspace(item: any): Api.SystemManage.WorkspaceItem {
  return {
    id: item?.id || '',
    workspaceType: item?.workspace_type || item?.workspaceType || 'personal',
    name: item?.name || '',
    code: item?.code || '',
    ownerUserId: item?.owner_user_id || item?.ownerUserId || '',
    collaborationWorkspaceId:
      item?.collaboration_workspace_id || item?.collaborationWorkspaceId || '',
    currentCollaborationWorkspaceId:
      item?.current_collaboration_workspace_id ||
      item?.currentCollaborationWorkspaceId ||
      item?.collaboration_workspace_id ||
      item?.collaborationWorkspaceId ||
      '',
    currentCollaborationWorkspaceName:
      item?.current_collaboration_workspace_name ||
      item?.currentCollaborationWorkspaceName ||
      item?.collaboration_workspace_name ||
      item?.collaborationWorkspaceName ||
      '',
    status: item?.status || 'active'
  }
}

export async function fetchGetMyWorkspaces() {
  const res = await request.get<{ records: any[]; total: number }>({
    url: `${WORKSPACE_BASE}/my`,
    skipAuthWorkspaceHeader: true,
    skipCollaborationWorkspaceHeader: true,
    showErrorMessage: false
  })

  return {
    records: (res?.records || []).map(normalizeWorkspace),
    total: res?.total || 0
  }
}

export async function fetchGetCurrentWorkspace() {
  const res = await request.get<any>({
    url: `${WORKSPACE_BASE}/current`,
    skipAuthWorkspaceHeader: true,
    skipCollaborationWorkspaceHeader: true,
    showErrorMessage: false
  })
  return normalizeWorkspace(res)
}

export async function fetchGetWorkspace(workspaceId: string) {
  const res = await request.get<any>({
    url: `${WORKSPACE_BASE}/${workspaceId}`,
    skipAuthWorkspaceHeader: true,
    skipCollaborationWorkspaceHeader: true,
    showErrorMessage: false
  })
  return normalizeWorkspace(res)
}

export function fetchSwitchWorkspace(workspaceId: string) {
  return request.post<{
    auth_workspace_id: string
    auth_workspace_type: string
    current_collaboration_workspace_id?: string
    collaboration_workspace_id?: string
    workspace: Api.SystemManage.WorkspaceItem
  }>({
    url: `${WORKSPACE_BASE}/switch`,
    skipAuthWorkspaceHeader: true,
    skipCollaborationWorkspaceHeader: true,
    data: {
      workspace_id: workspaceId
    }
  })
}
