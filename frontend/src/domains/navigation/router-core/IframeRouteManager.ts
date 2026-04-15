/**
 * Iframe 路由管理器
 *
 * 负责管理 iframe 类型的路由
 *
 * @module router/core/IframeRouteManager
 * @author Art Design Pro Team
 */

import type { AppRouteRecord } from '@/types/router'
import {
  APP_SCOPE_GLOBAL,
  normalizeAppScopeKey,
  readActiveAppScopeKey
} from '@/domains/app-runtime/app-scope'
import { logger } from '@/utils/logger'

export class IframeRouteManager {
  private static instance: IframeRouteManager
  private iframeRouteBuckets: Record<string, AppRouteRecord[]> = {}
  private static readonly LEGACY_STORAGE_KEY = 'iframeRoutes'
  private static readonly STORAGE_PREFIX = 'iframeRoutes:app:'

  private constructor() {}

  static getInstance(): IframeRouteManager {
    if (!IframeRouteManager.instance) {
      IframeRouteManager.instance = new IframeRouteManager()
    }
    return IframeRouteManager.instance
  }

  /**
   * 添加 iframe 路由
   */
  add(route: AppRouteRecord, appKey?: string): void {
    const scopeKey = this.resolveScopeKey(appKey)
    const routes = this.ensureBucket(scopeKey)
    if (!routes.find((r) => r.path === route.path)) {
      routes.push(route)
    }
  }

  /**
   * 获取所有 iframe 路由
   */
  getAll(appKey?: string): AppRouteRecord[] {
    return this.ensureBucket(this.resolveScopeKey(appKey))
  }

  /**
   * 根据路径查找 iframe 路由
   */
  findByPath(path: string, appKey?: string): AppRouteRecord | undefined {
    return this.getAll(appKey).find((route) => route.path === path)
  }

  /**
   * 清空所有 iframe 路由
   */
  clear(appKey?: string): void {
    const scopeKey = this.resolveScopeKey(appKey)
    this.iframeRouteBuckets[scopeKey] = []
  }

  /**
   * 保存到 sessionStorage
   */
  save(): void {
    Object.entries(this.iframeRouteBuckets).forEach(([scopeKey, routes]) => {
      if (routes.length > 0) {
        sessionStorage.setItem(this.buildStorageKey(scopeKey), JSON.stringify(routes))
      } else {
        sessionStorage.removeItem(this.buildStorageKey(scopeKey))
      }
    })
    sessionStorage.removeItem(IframeRouteManager.LEGACY_STORAGE_KEY)
  }

  /**
   * 保存当前 APP scope 的 iframe 路由
   */
  saveCurrentScope(appKey?: string): void {
    const scopeKey = this.resolveScopeKey(appKey)
    const routes = this.ensureBucket(scopeKey)
    if (routes.length > 0) {
      sessionStorage.setItem(this.buildStorageKey(scopeKey), JSON.stringify(routes))
      return
    }
    sessionStorage.removeItem(this.buildStorageKey(scopeKey))
  }

  /**
   * 从 sessionStorage 加载
   */
  load(appKey?: string): void {
    const scopeKey = this.resolveScopeKey(appKey)
    try {
      const data = sessionStorage.getItem(this.buildStorageKey(scopeKey))
      if (data) {
        this.iframeRouteBuckets[scopeKey] = JSON.parse(data)
        return
      }
      this.tryMigrateLegacy(scopeKey)
    } catch (error) {
      logger.error('navigation.iframe_route_manager.load_failed', { err: error, scopeKey })
      this.iframeRouteBuckets[scopeKey] = []
    }
  }

  private buildStorageKey(scopeKey: string): string {
    return `${IframeRouteManager.STORAGE_PREFIX}${scopeKey}`
  }

  private resolveScopeKey(appKey?: string): string {
    const normalized = normalizeAppScopeKey(appKey || readActiveAppScopeKey())
    return normalized || APP_SCOPE_GLOBAL
  }

  private ensureBucket(scopeKey: string): AppRouteRecord[] {
    if (!this.iframeRouteBuckets[scopeKey]) {
      this.iframeRouteBuckets[scopeKey] = []
    }
    return this.iframeRouteBuckets[scopeKey]
  }

  private tryMigrateLegacy(scopeKey: string): void {
    const legacyData = sessionStorage.getItem(IframeRouteManager.LEGACY_STORAGE_KEY)
    if (!legacyData) {
      this.iframeRouteBuckets[scopeKey] = []
      return
    }
    try {
      this.iframeRouteBuckets[scopeKey] = JSON.parse(legacyData)
      sessionStorage.setItem(this.buildStorageKey(scopeKey), legacyData)
      sessionStorage.removeItem(IframeRouteManager.LEGACY_STORAGE_KEY)
    } catch {
      this.iframeRouteBuckets[scopeKey] = []
    }
  }
}
