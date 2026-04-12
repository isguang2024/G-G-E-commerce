export interface RestoreSessionOptions {
  preferredSpaceKey?: string
  prefetchedUser?: Api.Auth.UserInfo
  forceRefresh?: boolean
  skipWorkspaceReconcile?: boolean
}

type SessionRuntimeHandlers = {
  restoreSession?: (options?: RestoreSessionOptions) => Promise<Api.Auth.UserInfo | null>
}

let handlers: SessionRuntimeHandlers = {}

export function registerSessionRuntimeHandlers(nextHandlers: SessionRuntimeHandlers): void {
  handlers = nextHandlers
}

export async function restoreSessionViaHandler(
  options: RestoreSessionOptions = {}
): Promise<Api.Auth.UserInfo | null> {
  if (!handlers.restoreSession) {
    throw new Error('[AuthRuntime] restoreSession handler 未注册')
  }
  return handlers.restoreSession(options)
}
