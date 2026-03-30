import type { Router } from 'vue-router'

type SpaceNavigationResolver = {
  resolveSpaceNavigationTarget: (
    targetPath: string,
    spaceKey?: string
  ) => { mode: 'router' | 'location'; target: string }
}

const IGNORE_PROTOCOL_PATTERN = /^(mailto:|tel:|javascript:)/i

export function resolveRichTextInternalPath(href?: string) {
  const target = `${href || ''}`.trim()
  if (!target || IGNORE_PROTOCOL_PATTERN.test(target)) {
    return ''
  }
  if (target.startsWith('#/')) {
    return target.slice(1)
  }
  if (target.startsWith('/')) {
    return target
  }
  if (typeof window === 'undefined') {
    return ''
  }
  try {
    const url = new URL(target, window.location.href)
    if (!/^https?:$/i.test(url.protocol)) {
      return ''
    }
    if (url.origin !== window.location.origin) {
      return ''
    }
    if (url.hash.startsWith('#/')) {
      return url.hash.slice(1)
    }
    return `${url.pathname || '/'}${url.search || ''}${url.hash || ''}`
  } catch (error) {
    return ''
  }
}

export async function handleRichTextLinkNavigation(
  event: MouseEvent,
  options: {
    router: Router
    spaceResolver: SpaceNavigationResolver
    spaceKey?: string
  }
) {
  const anchor = (event.target as HTMLElement | null)?.closest('a')
  if (!anchor) {
    return false
  }
  const href = `${anchor.getAttribute('href') || ''}`.trim()
  const internalPath = resolveRichTextInternalPath(href)
  if (!internalPath) {
    return false
  }

  const openInNewTab =
    anchor.getAttribute('target') === '_blank' ||
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.button === 1

  const nextTarget = options.spaceResolver.resolveSpaceNavigationTarget(internalPath, options.spaceKey)
  event.preventDefault()

  if (nextTarget.mode === 'router' && !openInNewTab) {
    await options.router.push(nextTarget.target)
    return true
  }

  if (typeof window !== 'undefined') {
    if (openInNewTab) {
      window.open(nextTarget.target, '_blank', 'noopener')
    } else {
      window.location.assign(nextTarget.target)
    }
  }
  return true
}
