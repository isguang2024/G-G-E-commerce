import { create } from 'zustand'
import type { ShellTab } from '@/shared/types/navigation'
import { appConfig } from '@/shared/config/app-config'
import { readJsonStorage, writeJsonStorage } from '@/shared/lib/storage'

export type ThemeMode = 'light' | 'dark'

const SHELL_STORAGE_KEY = 'frontend-fluentV2.shell.preferences'

type StoredShellState = {
  themeMode: ThemeMode
  navCollapsed: boolean
  currentSpaceKey: string
  navExpandedBySpace: Record<string, string[]>
  navCollapsedBySpace: Record<string, string[]>
  openTabs: ShellTab[]
  tabsEnabled: boolean
}

type ShellState = {
  themeMode: ThemeMode
  navCollapsed: boolean
  mobileNavOpen: boolean
  currentSpaceKey: string
  navExpandedBySpace: Record<string, string[]>
  navCollapsedBySpace: Record<string, string[]>
  activeTopContext: string
  openTabs: ShellTab[]
  tabsEnabled: boolean
  toggleTheme: () => void
  toggleNavCollapsed: () => void
  setMobileNavOpen: (open: boolean) => void
  setCurrentSpaceKey: (spaceKey: string) => void
  toggleNavItemExpanded: (spaceKey: string, itemId: string) => void
  setNavExpandedForSpace: (spaceKey: string, itemIds: string[]) => void
  pruneNavExpandedForSpace: (spaceKey: string, validItemIds: string[]) => void
  toggleNavItemCollapsed: (spaceKey: string, itemId: string) => void
  setNavCollapsedForSpace: (spaceKey: string, itemIds: string[]) => void
  pruneNavCollapsedForSpace: (spaceKey: string, validItemIds: string[]) => void
  setActiveTopContext: (label: string) => void
  registerTab: (tab: ShellTab) => void
  replaceTabs: (tabs: ShellTab[]) => void
  closeTab: (path: string) => void
  closeTabs: (paths: string[]) => void
  clearTabs: () => void
  toggleTabPinned: (path: string) => void
  closeOtherTabs: (path: string) => void
  closeTabsToLeft: (path: string) => void
  closeTabsToRight: (path: string) => void
  reorderTabs: (sourcePath: string, targetPath: string) => void
  setTabsEnabled: (enabled: boolean) => void
}

function sortTabsByPin(tabs: ShellTab[]) {
  const pinnedTabs = tabs.filter((item) => item.pinned)
  const normalTabs = tabs.filter((item) => !item.pinned)
  return [...pinnedTabs, ...normalTabs]
}

function normalizeExpandedIds(itemIds: string[]) {
  return [...new Set(itemIds.map((itemId) => `${itemId}`.trim()).filter(Boolean))]
}

function normalizeExpandedBySpace(navExpandedBySpace: Record<string, string[]>) {
  return Object.fromEntries(
    Object.entries(navExpandedBySpace)
      .map(([spaceKey, itemIds]) => [spaceKey, normalizeExpandedIds(Array.isArray(itemIds) ? itemIds : [])])
      .filter(([, itemIds]) => itemIds.length > 0),
  )
}

function normalizeCollapsedBySpace(navCollapsedBySpace: Record<string, string[]>) {
  return Object.fromEntries(
    Object.entries(navCollapsedBySpace)
      .map(([spaceKey, itemIds]) => [spaceKey, normalizeExpandedIds(Array.isArray(itemIds) ? itemIds : [])])
      .filter(([, itemIds]) => itemIds.length > 0),
  )
}

function readShellState(): StoredShellState {
  if (typeof window === 'undefined') {
    return {
      themeMode: 'light',
      navCollapsed: false,
      currentSpaceKey: appConfig.defaultSpaceKey,
      navExpandedBySpace: {},
      navCollapsedBySpace: {},
      openTabs: [],
      tabsEnabled: true,
    }
  }

  const stored = readJsonStorage<Partial<StoredShellState>>(window.localStorage, SHELL_STORAGE_KEY)
  return {
    themeMode: stored?.themeMode === 'dark' ? 'dark' : 'light',
    navCollapsed: Boolean(stored?.navCollapsed),
    currentSpaceKey: `${stored?.currentSpaceKey || appConfig.defaultSpaceKey}`.trim() || appConfig.defaultSpaceKey,
    navExpandedBySpace: normalizeExpandedBySpace(
      stored?.navExpandedBySpace && typeof stored.navExpandedBySpace === 'object' ? stored.navExpandedBySpace : {},
    ),
    navCollapsedBySpace: normalizeCollapsedBySpace(
      stored?.navCollapsedBySpace && typeof stored.navCollapsedBySpace === 'object' ? stored.navCollapsedBySpace : {},
    ),
    openTabs: Array.isArray(stored?.openTabs) ? stored.openTabs : [],
    tabsEnabled: stored?.tabsEnabled !== false,
  }
}

function persistShellState(
  state: Pick<
    ShellState,
    'themeMode' | 'navCollapsed' | 'currentSpaceKey' | 'navExpandedBySpace' | 'openTabs' | 'tabsEnabled'
  > &
    Partial<Pick<ShellState, 'navCollapsedBySpace'>>,
) {
  if (typeof window === 'undefined') {
    return
  }

  writeJsonStorage(window.localStorage, SHELL_STORAGE_KEY, {
    themeMode: state.themeMode,
    navCollapsed: state.navCollapsed,
    currentSpaceKey: state.currentSpaceKey,
    navExpandedBySpace: normalizeExpandedBySpace(state.navExpandedBySpace),
    navCollapsedBySpace: normalizeCollapsedBySpace(state.navCollapsedBySpace || {}),
    openTabs: sortTabsByPin(state.openTabs),
    tabsEnabled: state.tabsEnabled,
  })
}

const initialState = readShellState()

export const useShellStore = create<ShellState>((set) => ({
  themeMode: initialState.themeMode,
  navCollapsed: initialState.navCollapsed,
  mobileNavOpen: false,
  currentSpaceKey: initialState.currentSpaceKey,
  navExpandedBySpace: initialState.navExpandedBySpace,
  navCollapsedBySpace: initialState.navCollapsedBySpace,
  activeTopContext: '首页',
  openTabs: sortTabsByPin(initialState.openTabs),
  tabsEnabled: initialState.tabsEnabled,
  toggleTheme: () =>
    set((state) => {
      const themeMode: ThemeMode = state.themeMode === 'light' ? 'dark' : 'light'
      persistShellState({
        themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        navCollapsedBySpace: state.navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { themeMode }
    }),
  toggleNavCollapsed: () =>
    set((state) => {
      const navCollapsed = !state.navCollapsed
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        navCollapsedBySpace: state.navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navCollapsed }
    }),
  setMobileNavOpen: (mobileNavOpen) => set({ mobileNavOpen }),
  setCurrentSpaceKey: (currentSpaceKey) =>
    set((state) => {
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        navCollapsedBySpace: state.navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { currentSpaceKey }
    }),
  toggleNavItemExpanded: (spaceKey, itemId) =>
    set((state) => {
      const expandedIds = new Set(state.navExpandedBySpace[spaceKey] || [])
      if (expandedIds.has(itemId)) {
        expandedIds.delete(itemId)
      } else {
        expandedIds.add(itemId)
      }

      const collapsedIds = new Set(state.navCollapsedBySpace[spaceKey] || [])
      collapsedIds.delete(itemId)

      const navExpandedBySpace = {
        ...state.navExpandedBySpace,
        [spaceKey]: normalizeExpandedIds([...expandedIds]),
      }
      const navCollapsedBySpace = {
        ...state.navCollapsedBySpace,
        [spaceKey]: normalizeExpandedIds([...collapsedIds]),
      }
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace,
        navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navExpandedBySpace, navCollapsedBySpace }
    }),
  setNavExpandedForSpace: (spaceKey, itemIds) =>
    set((state) => {
      const navExpandedBySpace = {
        ...state.navExpandedBySpace,
        [spaceKey]: normalizeExpandedIds(itemIds),
      }
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace,
        navCollapsedBySpace: state.navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navExpandedBySpace }
    }),
  pruneNavExpandedForSpace: (spaceKey, validItemIds) =>
    set((state) => {
      const validIdSet = new Set(normalizeExpandedIds(validItemIds))
      const nextExpandedIds = normalizeExpandedIds(state.navExpandedBySpace[spaceKey] || []).filter((itemId) =>
        validIdSet.has(itemId),
      )

      const currentExpandedIds = normalizeExpandedIds(state.navExpandedBySpace[spaceKey] || [])
      if (
        currentExpandedIds.length === nextExpandedIds.length &&
        currentExpandedIds.every((itemId, index) => itemId === nextExpandedIds[index])
      ) {
        return state
      }

      const navExpandedBySpace = {
        ...state.navExpandedBySpace,
        [spaceKey]: nextExpandedIds,
      }
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace,
        navCollapsedBySpace: state.navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navExpandedBySpace }
    }),
  toggleNavItemCollapsed: (spaceKey, itemId) =>
    set((state) => {
      const collapsedIds = new Set(state.navCollapsedBySpace[spaceKey] || [])
      if (collapsedIds.has(itemId)) {
        collapsedIds.delete(itemId)
      } else {
        collapsedIds.add(itemId)
      }

      const navCollapsedBySpace = {
        ...state.navCollapsedBySpace,
        [spaceKey]: normalizeExpandedIds([...collapsedIds]),
      }
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navCollapsedBySpace }
    }),
  setNavCollapsedForSpace: (spaceKey, itemIds) =>
    set((state) => {
      const navCollapsedBySpace = {
        ...state.navCollapsedBySpace,
        [spaceKey]: normalizeExpandedIds(itemIds),
      }
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navCollapsedBySpace }
    }),
  pruneNavCollapsedForSpace: (spaceKey, validItemIds) =>
    set((state) => {
      const validIdSet = new Set(normalizeExpandedIds(validItemIds))
      const nextCollapsedIds = normalizeExpandedIds(state.navCollapsedBySpace[spaceKey] || []).filter((itemId) =>
        validIdSet.has(itemId),
      )

      const currentCollapsedIds = normalizeExpandedIds(state.navCollapsedBySpace[spaceKey] || [])
      if (
        currentCollapsedIds.length === nextCollapsedIds.length &&
        currentCollapsedIds.every((itemId, index) => itemId === nextCollapsedIds[index])
      ) {
        return state
      }

      const navCollapsedBySpace = {
        ...state.navCollapsedBySpace,
        [spaceKey]: nextCollapsedIds,
      }
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        navCollapsedBySpace,
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { navCollapsedBySpace }
    }),
  setActiveTopContext: (activeTopContext) => set({ activeTopContext }),
  registerTab: (tab) =>
    set((state) => {
      const nextTabs = state.openTabs.slice()
      const matchedIndex = nextTabs.findIndex((item) => item.path === tab.path)

      if (matchedIndex >= 0) {
        nextTabs[matchedIndex] = {
          ...nextTabs[matchedIndex],
          ...tab,
          pinned: nextTabs[matchedIndex].pinned || tab.pinned,
        }
      } else {
        nextTabs.push(tab)
      }

      const openTabs = sortTabsByPin(nextTabs)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  replaceTabs: (openTabs) =>
    set((state) => {
      const normalizedTabs = sortTabsByPin(openTabs)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs: normalizedTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs: normalizedTabs }
    }),
  closeTab: (path) =>
    set((state) => {
      const openTabs = state.openTabs.filter((item) => item.path !== path)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  closeTabs: (paths) =>
    set((state) => {
      if (!paths.length) {
        return state
      }

      const pathSet = new Set(paths)
      const openTabs = state.openTabs.filter((item) => !pathSet.has(item.path))
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  clearTabs: () =>
    set((state) => {
      const openTabs: ShellTab[] = []
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  toggleTabPinned: (path) =>
    set((state) => {
      const openTabs = state.openTabs.map((item) =>
          item.path === path
            ? {
                ...item,
                pinned: !item.pinned,
              }
            : item,
        )
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  closeOtherTabs: (path) =>
    set((state) => {
      const openTabs = state.openTabs.filter((item) => item.path === path || item.pinned)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  closeTabsToLeft: (path) =>
    set((state) => {
      const targetIndex = state.openTabs.findIndex((item) => item.path === path)
      if (targetIndex < 0) {
        return state
      }

      const openTabs = state.openTabs.filter((item, index) => index >= targetIndex || item.pinned)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  closeTabsToRight: (path) =>
    set((state) => {
      const targetIndex = state.openTabs.findIndex((item) => item.path === path)
      if (targetIndex < 0) {
        return state
      }

      const openTabs = state.openTabs.filter((item, index) => index <= targetIndex || item.pinned)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  reorderTabs: (sourcePath, targetPath) =>
    set((state) => {
      if (sourcePath === targetPath) {
        return state
      }

      const nextTabs = state.openTabs.slice()
      const sourceIndex = nextTabs.findIndex((item) => item.path === sourcePath)
      const targetIndex = nextTabs.findIndex((item) => item.path === targetPath)

      if (sourceIndex < 0 || targetIndex < 0) {
        return state
      }

      const [sourceTab] = nextTabs.splice(sourceIndex, 1)
      nextTabs.splice(targetIndex, 0, sourceTab)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs: nextTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs: nextTabs }
    }),
  setTabsEnabled: (tabsEnabled) =>
    set((state) => {
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
        navExpandedBySpace: state.navExpandedBySpace,
        openTabs: state.openTabs,
        tabsEnabled,
      })
      return { tabsEnabled }
    }),
}))
