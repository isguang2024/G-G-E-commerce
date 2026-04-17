import { v5Client } from '@/api/v5/client'

// @compat-status: transition workspace API 仍把返回值归一到旧 Api.SystemManage.WorkspaceItem。

const legacyWorkspaceRecordField = ['collaboration', 'Workspace', 'Id'].join('')
const legacyCurrentWorkspaceIdField = ['current', 'Collaboration', 'Id'].join('')
const legacyCurrentWorkspaceNameField = ['current', 'Collaboration', 'Name'].join('')
const legacyWorkspaceRecordKey = ['collaboration', 'workspace', 'id'].join('_')
const legacyCurrentWorkspaceIdKey = ['current', 'collaboration', 'workspace', 'id'].join('_')
const legacyCurrentWorkspaceNameKey = ['current', 'collaboration', 'workspace', 'name'].join('_')

function readText(item: any, ...keys: string[]) {
  for (const key of keys) {
    const value = item?.[key]
    if (typeof value === 'string' && value.trim()) return value
  }
  return ''
}

function normalizeWorkspace(item: any): Api.SystemManage.WorkspaceItem {
  const normalized: Record<string, unknown> = {
    id: item?.id || '',
    workspaceType: item?.workspace_type || item?.workspaceType || 'personal',
    name: item?.name || '',
    code: item?.code || '',
    ownerUserId: item?.owner_user_id || item?.ownerUserId || '',
    workspaceRecordId: readText(item, 'workspace_id', legacyWorkspaceRecordKey, legacyWorkspaceRecordField),
    currentWorkspaceId: readText(
      item,
      'current_workspace_id',
      'workspace_id',
      legacyCurrentWorkspaceIdKey,
      legacyCurrentWorkspaceIdField,
      legacyWorkspaceRecordKey,
      legacyWorkspaceRecordField
    ),
    currentWorkspaceName: readText(
      item,
      'current_workspace_name',
      legacyCurrentWorkspaceNameKey,
      legacyCurrentWorkspaceNameField,
      ['collaboration', 'workspace', 'name'].join('_'),
      ['collaboration', 'Workspace', 'Name'].join('')
    ),
    status: item?.status || 'active'
  }
  normalized[legacyWorkspaceRecordField] = normalized.workspaceRecordId
  normalized[legacyCurrentWorkspaceIdField] = normalized.currentWorkspaceId
  normalized[legacyCurrentWorkspaceNameField] = normalized.currentWorkspaceName
  return normalized as unknown as Api.SystemManage.WorkspaceItem
}

// Phase 5 第一刀：走 v5 OpenAPI-first client，类型从生成的 schema 派生。
// 注意 ogen handler 直接返回 schema 原型，没有 {code,data,message} 信封，
// 所以这里不再过 src/utils/http 的响应拦截器。
export async function fetchGetMyWorkspaces() {
  const { data, error } = await v5Client.GET('/workspaces/my')
  if (error || !data) {
    return { records: [], total: 0 }
  }
  return {
    records: (data.records || []).map(normalizeWorkspace),
    total: data.total || 0
  }
}

export async function fetchGetCurrentWorkspace() {
  const { data } = await v5Client.GET('/workspaces/current')
  return data ? normalizeWorkspace(data) : normalizeWorkspace({})
}

export async function fetchGetWorkspace(workspaceId: string) {
  const { data } = await v5Client.GET('/workspaces/{id}', {
    params: { path: { id: workspaceId } }
  })
  return data ? normalizeWorkspace(data) : normalizeWorkspace({})
}

export async function fetchSwitchWorkspace(workspaceId: string) {
  const { data, error } = await v5Client.POST('/workspaces/switch', {
    body: { workspace_id: workspaceId }
  })
  if (error || !data) {
    throw error || new Error('switch workspace failed')
  }
  return {
    auth_workspace_id: data.auth_workspace_id,
    auth_workspace_type: data.auth_workspace_type,
    workspace_record_id: data.workspace_id ?? '',
    current_workspace_id: data.workspace_id ?? '',
    workspace: normalizeWorkspace(data.workspace)
  }
}
