export interface SwitchAppPayload {
  appKey?: string | null
  name?: string | null
  frontendEntryUrl?: string | null
  backendEntryUrl?: string | null
  healthCheckUrl?: string | null
  authMode?: string | null
  capabilities?: Record<string, unknown> | null
  meta?: Record<string, unknown> | null
  defaultSpaceKey?: string | null
}

type AppContextRuntimeHandlers = {
  ensureRuntimeAppKey?: () => Promise<string>
  switchApp?: (payload: SwitchAppPayload) => Promise<void>
}

let handlers: AppContextRuntimeHandlers = {}

export function registerAppContextRuntimeHandlers(
  nextHandlers: AppContextRuntimeHandlers
): void {
  handlers = nextHandlers
}

export async function ensureRuntimeAppKeyViaHandler(): Promise<string> {
  if (!handlers.ensureRuntimeAppKey) {
    throw new Error('[AppContextRuntime] ensureRuntimeAppKey handler 未注册')
  }
  return handlers.ensureRuntimeAppKey()
}

export async function switchAppViaHandler(payload: SwitchAppPayload): Promise<void> {
  if (!handlers.switchApp) {
    throw new Error('[AppContextRuntime] switchApp handler 未注册')
  }
  return handlers.switchApp(payload)
}
