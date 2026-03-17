import type { App } from 'vue'
import { setupActionDirective } from './core/action'
import { setupHighlightDirective } from './business/highlight'
import { setupRippleDirective } from './business/ripple'
import { setupRolesDirective } from './core/roles'

export function setupGlobDirectives(app: App) {
  setupActionDirective(app) // 功能权限指令
  setupRolesDirective(app) // 角色权限指令
  setupHighlightDirective(app) // 高亮指令
  setupRippleDirective(app) // 水波纹指令
}
