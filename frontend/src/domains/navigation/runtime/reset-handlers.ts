import { resetRouteInitState } from '@/domains/navigation/runtime/guard-state'

let resetNavigationRuntimeHandler: (() => void) | null = null

export function registerNavigationRuntimeResetHandler(handler: () => void): void {
  resetNavigationRuntimeHandler = handler
}

export function resetRouterStateNow(): void {
  resetNavigationRuntimeHandler?.()
  resetRouteInitState()
}

export function resetRouterState(delay: number): void {
  setTimeout(() => {
    resetRouterStateNow()
  }, delay)
}
