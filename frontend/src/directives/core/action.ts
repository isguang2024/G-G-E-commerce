import { watchEffect } from 'vue'
import { useUserStore } from '@/store/modules/user'
import { App, Directive, DirectiveBinding } from 'vue'

interface ActionBinding extends DirectiveBinding {
  value: string
}

interface ActionHTMLElement extends HTMLElement {
  __actionStop__?: () => void
}

function updateActionPermission(el: HTMLElement, binding: ActionBinding): void {
  const userStore = useUserStore()
  const userInfo = userStore.getUserInfo
  const isSuperAdmin = Boolean(userInfo?.is_super_admin)
  const actionList = userInfo?.actions || []
  toggleElement(el, isSuperAdmin || actionList.includes(binding.value))
}

function mountWatcher(el: ActionHTMLElement, binding: ActionBinding): void {
  unmountWatcher(el)
  el.__actionStop__ = watchEffect(() => {
    updateActionPermission(el, binding)
  })
}

function unmountWatcher(el: ActionHTMLElement): void {
  if (el.__actionStop__) {
    el.__actionStop__()
    delete el.__actionStop__
  }
}

function toggleElement(el: HTMLElement, visible: boolean): void {
  if (visible) {
    const originalDisplay = el.dataset.authDisplay || ''
    el.style.display = originalDisplay
    return
  }
  if (!el.dataset.authDisplay) {
    el.dataset.authDisplay = el.style.display
  }
  el.style.display = 'none'
}

const actionDirective: Directive = {
  mounted(el, binding) {
    mountWatcher(el as ActionHTMLElement, binding as ActionBinding)
  },
  updated(el, binding) {
    mountWatcher(el as ActionHTMLElement, binding as ActionBinding)
  },
  unmounted(el) {
    unmountWatcher(el as ActionHTMLElement)
  }
}

export function setupActionDirective(app: App): void {
  app.directive('action', actionDirective)
}
