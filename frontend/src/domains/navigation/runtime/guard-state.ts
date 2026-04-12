let pendingLoading = false
let routeInitFailed = false
let routeInitInProgress: Promise<unknown> | null = null
const routeRefreshAttempted = new Set<string>()

export function getPendingLoading(): boolean {
  return pendingLoading
}

export function setPendingLoading(value: boolean): void {
  pendingLoading = value
}

export function resetPendingLoading(): void {
  pendingLoading = false
}

export function getRouteInitFailed(): boolean {
  return routeInitFailed
}

export function setRouteInitFailed(value: boolean): void {
  routeInitFailed = value
}

export function getRouteInitInProgress(): Promise<unknown> | null {
  return routeInitInProgress
}

export function setRouteInitInProgress(value: Promise<unknown> | null): void {
  routeInitInProgress = value
}

export function hasRouteRefreshAttempted(path: string): boolean {
  return routeRefreshAttempted.has(path)
}

export function addRouteRefreshAttempted(path: string): void {
  routeRefreshAttempted.add(path)
}

export function clearRouteRefreshAttempted(path: string): void {
  routeRefreshAttempted.delete(path)
}

export function resetRouteInitState(): void {
  routeInitFailed = false
  routeInitInProgress = null
  routeRefreshAttempted.clear()
}
