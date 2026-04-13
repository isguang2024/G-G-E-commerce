type AuthContextHandlers = {
  getAccessToken?: () => string
  getRefreshToken?: () => string
  applySession?: (payload: {
    accessToken: string
    refreshToken?: string
    isLogin?: boolean
  }) => void
  logOut?: () => Promise<void>
}

type WorkspaceContextHandlers = {
  getCurrentAuthWorkspaceId?: () => string
}

type CollaborationContextHandlers = {
  getCurrentCollaborationWorkspaceId?: () => string
  getCurrentContextMode?: () => string
}

type AppContextHandlers = {
  getCurrentRuntimeBackendEntryURL?: () => string
  getEffectiveManagedAppKey?: () => string
  getCurrentRuntimeAppKey?: () => string
  resolveAppAuthMode?: (appKey?: string | null) => string
  resolveAppSsoMode?: (appKey?: string | null) => string
  resolveAppLoginPageKey?: (appKey?: string | null) => string
}

let authContextHandlers: AuthContextHandlers = {}
let workspaceContextHandlers: WorkspaceContextHandlers = {}
let collaborationContextHandlers: CollaborationContextHandlers = {}
let appContextHandlers: AppContextHandlers = {}

export function registerHttpAuthContext(handlers: AuthContextHandlers): void {
  authContextHandlers = handlers
}

export function registerHttpWorkspaceContext(handlers: WorkspaceContextHandlers): void {
  workspaceContextHandlers = handlers
}

export function registerHttpCollaborationContext(
  handlers: CollaborationContextHandlers
): void {
  collaborationContextHandlers = handlers
}

export function registerHttpAppContext(handlers: AppContextHandlers): void {
  appContextHandlers = handlers
}

export function getHttpAccessToken(): string {
  return `${authContextHandlers.getAccessToken?.() || ''}`.trim()
}

export function getHttpRefreshToken(): string {
  return `${authContextHandlers.getRefreshToken?.() || ''}`.trim()
}

export function applyHttpSession(payload: {
  accessToken: string
  refreshToken?: string
  isLogin?: boolean
}): void {
  authContextHandlers.applySession?.(payload)
}

export async function logoutHttpSession(): Promise<void> {
  await authContextHandlers.logOut?.()
}

export function getCurrentAuthWorkspaceId(): string {
  return `${workspaceContextHandlers.getCurrentAuthWorkspaceId?.() || ''}`.trim()
}

export function getCurrentCollaborationWorkspaceId(): string {
  return `${collaborationContextHandlers.getCurrentCollaborationWorkspaceId?.() || ''}`.trim()
}

export function getCurrentContextMode(): string {
  return `${collaborationContextHandlers.getCurrentContextMode?.() || ''}`.trim()
}

export function getCurrentRuntimeBackendEntryURL(): string {
  return `${appContextHandlers.getCurrentRuntimeBackendEntryURL?.() || ''}`.trim()
}

export function getEffectiveManagedAppKey(): string {
  return `${appContextHandlers.getEffectiveManagedAppKey?.() || ''}`.trim()
}

export function getCurrentRuntimeAppKey(): string {
  return `${appContextHandlers.getCurrentRuntimeAppKey?.() || ''}`.trim()
}

export function resolveHttpAppAuthMode(appKey?: string | null): string {
  return `${appContextHandlers.resolveAppAuthMode?.(appKey) || ''}`.trim()
}

export function resolveHttpAppSsoMode(appKey?: string | null): string {
  return `${appContextHandlers.resolveAppSsoMode?.(appKey) || 'participate'}`.trim()
}

export function resolveHttpAppLoginPageKey(appKey?: string | null): string {
  return `${appContextHandlers.resolveAppLoginPageKey?.(appKey) || ''}`.trim()
}
