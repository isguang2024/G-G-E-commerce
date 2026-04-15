/**
 * 组件加载器
 *
 * 负责动态加载 Vue 组件
 *
 * @module router/core/ComponentLoader
 * @author Art Design Pro Team
 */

import { h } from 'vue'
import { logger } from '@/utils/logger'

export class ComponentLoader {
  private modules: Record<string, () => Promise<any>>
  private readonly componentPathAliases: Array<[string, string]> = []
  private readonly missingPathCache = new Set<string>()

  constructor() {
    // 动态导入 views 目录下所有 .vue 组件
    this.modules = import.meta.glob('../../../views/**/*.vue')
  }

  /**
   * 加载组件
   */
  load(componentPath: string): () => Promise<any> {
    if (!componentPath) {
      return this.createEmptyComponent()
    }

    const resolvedPath = this.resolveComponentPath(componentPath)
    const { fullPath, fullPathWithIndex } = this.buildCandidatePaths(resolvedPath)

    // 先尝试直接路径，再尝试添加/index的路径
    const module = this.modules[fullPath] || this.modules[fullPathWithIndex]

    if (!module) {
      if (import.meta.env.DEV && !this.missingPathCache.has(resolvedPath)) {
        logger.debug('navigation.component_loader.missing_component', {
          componentPath,
          resolvedPath,
          fullPath,
          fullPathWithIndex
        })
        this.missingPathCache.add(resolvedPath)
      }
      return this.createErrorComponent(componentPath)
    }

    return module
  }

  /**
   * 检查组件路径是否可被加载
   */
  exists(componentPath: string): boolean {
    if (!componentPath) return false
    const resolvedPath = this.resolveComponentPath(componentPath)
    const { fullPath, fullPathWithIndex } = this.buildCandidatePaths(resolvedPath)
    return !!(this.modules[fullPath] || this.modules[fullPathWithIndex])
  }

  /**
   * 加载布局组件
   */
  loadLayout(): () => Promise<any> {
    return () => import('@/views/index/index.vue')
  }

  /**
   * 加载 iframe 组件
   */
  loadIframe(): () => Promise<any> {
    return () => import('@/views/outside/Iframe.vue')
  }

  /**
   * 清理 APP 切换期间的组件解析缓存，避免跨 APP 的历史 miss 污染日志与调试判断。
   */
  clearCache(): void {
    this.missingPathCache.clear()
  }

  /**
   * 创建空组件
   */
  private createEmptyComponent(): () => Promise<any> {
    return () =>
      Promise.resolve({
        render() {
          return h('div', {})
        }
      })
  }

  /**
   * 创建错误提示组件
   */
  private createErrorComponent(componentPath: string): () => Promise<any> {
    return () =>
      Promise.resolve({
        render() {
          return h('div', { class: 'route-error' }, `组件未找到: ${componentPath}`)
        }
      })
  }

  /**
   * 构建组件候选路径
   */
  private buildCandidatePaths(componentPath: string): {
    fullPath: string
    fullPathWithIndex: string
  } {
    const normalizedPath = componentPath.startsWith('/') ? componentPath : `/${componentPath}`
    return {
      fullPath: `../../../views${normalizedPath}.vue`,
      fullPathWithIndex: `../../../views${normalizedPath}/index.vue`
    }
  }

  private resolveComponentPath(componentPath: string): string {
    const normalizedPath = componentPath.startsWith('/') ? componentPath : `/${componentPath}`
    const matchedAlias = this.componentPathAliases.find(([from]) => {
      return normalizedPath === from || normalizedPath.startsWith(`${from}/`)
    })
    if (!matchedAlias) return normalizedPath
    const [from, to] = matchedAlias
    return `${to}${normalizedPath.slice(from.length)}`
  }
}
