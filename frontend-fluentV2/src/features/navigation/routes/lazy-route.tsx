import { Suspense, lazy } from 'react'
import type { ComponentType, ReactNode } from 'react'
import { Spinner } from '@fluentui/react-components'

function RouteFallback() {
  return <Spinner label="正在加载页面模块" />
}

export function withRouteSuspense(element: ReactNode) {
  return <Suspense fallback={<RouteFallback />}>{element}</Suspense>
}

export function createLazyRouteElement(
  loader: () => Promise<Record<string, unknown>>,
  exportName: string,
  routeId: string,
) {
  const LazyComponent = lazy(async () => {
    const mod = await loader()
    return {
      default: mod[exportName] as ComponentType<{ routeId: string }>,
    }
  })

  return withRouteSuspense(<LazyComponent routeId={routeId} />)
}
