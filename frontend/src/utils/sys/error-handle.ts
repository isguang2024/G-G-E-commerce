/**
 * 全局错误处理模块
 *
 * 提供统一的错误捕获和处理机制
 *
 * ## 主要功能
 *
 * - Vue 运行时错误捕获（组件错误、生命周期错误等）
 * - 全局脚本错误捕获（语法错误、运行时错误等）
 * - Promise 未捕获错误处理（unhandledrejection）
 * - 静态资源加载错误监控（图片、脚本、样式等）
 * - 错误日志记录和上报
 * - 统一的错误处理入口
 *
 * ## 使用场景
 * - 应用启动时安装全局错误处理器
 * - 捕获和记录所有类型的错误
 * - 错误上报到监控平台
 * - 提升应用稳定性和可维护性
 * - 问题排查和调试
 *
 * ## 错误类型
 *
 * - VueError: Vue 组件相关错误
 * - ScriptError: JavaScript 脚本错误
 * - PromiseError: Promise 未捕获的 rejection
 * - ResourceError: 静态资源加载失败
 *
 * @module utils/sys/error-handle
 * @author Art Design Pro Team
 */
import type { App } from 'vue'
import { logger } from '@/utils/logger'

function isIgnorableResizeObserverMessage(message: unknown): boolean {
  const normalizedMessage = `${message || ''}`.trim()
  return (
    normalizedMessage.includes('ResizeObserver loop limit exceeded') ||
    normalizedMessage.includes('ResizeObserver loop completed with undelivered notifications')
  )
}

/**
 * Vue 运行时错误处理
 *
 * 所有 Vue 组件/生命周期异常最终会在这里汇总。logger.error 会同时：
 *  - 开发环境打到 console（体验）；
 *  - 生产环境批量上报到 /telemetry/logs（观测）。
 * 不再依赖散落的 console.error，统一由 logger 决定打不打。
 */
export function vueErrorHandler(err: unknown, instance: any, info: string) {
  logger.error('sys.vue_error', {
    err,
    info,
    componentName: (instance as { $options?: { name?: string; __name?: string } } | null)?.$options
      ?.name,
  })
}

/**
 * 全局脚本错误处理
 */
export function scriptErrorHandler(
  message: Event | string,
  source?: string,
  lineno?: number,
  colno?: number,
  error?: Error
): boolean {
  if (isIgnorableResizeObserverMessage(typeof message === 'string' ? message : error?.message)) {
    return true
  }
  logger.error('sys.script_error', {
    message: typeof message === 'string' ? message : String(message),
    source,
    lineno,
    colno,
    err: error,
  })
  return true // 阻止默认控制台报错，可根据需求改
}

/**
 * Promise 未捕获错误处理
 */
export function registerPromiseErrorHandler() {
  window.addEventListener('unhandledrejection', (event) => {
    if (isIgnorableResizeObserverMessage(event.reason?.message || event.reason)) {
      event.preventDefault()
      return
    }
    logger.error('sys.promise_rejection', { err: event.reason })
  })
}

/**
 * 资源加载错误处理 (img, script, css...)
 */
export function registerResourceErrorHandler() {
  window.addEventListener(
    'error',
    (event: Event) => {
      const target = event.target as HTMLElement
      if (
        target &&
        (target.tagName === 'IMG' || target.tagName === 'SCRIPT' || target.tagName === 'LINK')
      ) {
        logger.warn('sys.resource_error', {
          tagName: target.tagName,
          src:
            (target as HTMLImageElement).src ||
            (target as HTMLScriptElement).src ||
            (target as HTMLLinkElement).href,
        })
      }
    },
    true // 捕获阶段才能监听到资源错误
  )
}

/**
 * 安装统一错误处理
 */
export function setupErrorHandle(app: App) {
  app.config.errorHandler = vueErrorHandler
  window.onerror = scriptErrorHandler
  registerPromiseErrorHandler()
  registerResourceErrorHandler()
}
