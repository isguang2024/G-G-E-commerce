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
  openTabs: ShellTab[]
  tabsEnabled: boolean
}

type ShellState = {
  themeMode: ThemeMode
  navCollapsed: boolean
  mobileNavOpen: boolean
  currentSpaceKey: string
  activeTopContext: string
  openTabs: ShellTab[]
  tabsEnabled: boolean
  toggleTheme: () => void
  toggleNavCollapsed: () => void
  setMobileNavOpen: (open: boolean) => void
  setCurrentSpaceKey: (spaceKey: string) => void
  setActiveTopContext: (label: string) => void
  registerTab: (tab: ShellTab) => void
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

function readShellState(): StoredShellState {
  if (typeof window === 'undefined') {
    return {
      themeMode: 'light',
      navCollapsed: false,
      currentSpaceKey: appConfig.defaultSpaceKey,
      openTabs: [],
      tabsEnabled: true,
    }
  }

  const stored = readJsonStorage<Partial<StoredShellState>>(window.localStorage, SHELL_STORAGE_KEY)
  return {
    themeMode: stored?.themeMode === 'dark' ? 'dark' : 'light',
    navCollapsed: Boolean(stored?.navCollapsed),
    currentSpaceKey: `${stored?.currentSpaceKey || appConfig.defaultSpaceKey}`.trim() || appConfig.defaultSpaceKey,
    openTabs: Array.isArray(stored?.openTabs) ? stored.openTabs : [],
    tabsEnabled: stored?.tabsEnabled !== false,
  }
}

function persistShellState(state: Pick<ShellState, 'themeMode' | 'navCollapsed' | 'currentSpaceKey' | 'openTabs' | 'tabsEnabled'>) {
  if (typeof window === 'undefined') {
    return
  }

  writeJsonStorage(window.localStorage, SHELL_STORAGE_KEY, {
    themeMode: state.themeMode,
    navCollapsed: state.navCollapsed,
    currentSpaceKey: state.currentSpaceKey,
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
        openTabs: state.openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { currentSpaceKey }
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
        openTabs,
        tabsEnabled: state.tabsEnabled,
      })
      return { openTabs }
    }),
  closeTab: (path) =>
    set((state) => {
      const openTabs = state.openTabs.filter((item) => item.path !== path)
      persistShellState({
        themeMode: state.themeMode,
        navCollapsed: state.navCollapsed,
        currentSpaceKey: state.currentSpaceKey,
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
        openTabs: state.openTabs,
        tabsEnabled,
      })
      return { tabsEnabled }
    }),
}))
