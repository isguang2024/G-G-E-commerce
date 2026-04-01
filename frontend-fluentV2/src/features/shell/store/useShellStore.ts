import { create } from 'zustand'
import type { ShellTab, SpaceKey } from '@/shared/types/navigation'
import type { SessionUser } from '@/shared/types/session'

export type ThemeMode = 'light' | 'dark'

const SHELL_TABS_STORAGE_KEY = 'frontend-fluentV2.shell.tabs'

type StoredTabsState = {
  openTabs: ShellTab[]
  tabsEnabled: boolean
}

type ShellState = {
  themeMode: ThemeMode
  navCollapsed: boolean
  mobileNavOpen: boolean
  currentSpaceKey: SpaceKey
  currentUser: SessionUser | null
  activeTopContext: string
  openTabs: ShellTab[]
  tabsEnabled: boolean
  toggleTheme: () => void
  toggleNavCollapsed: () => void
  setMobileNavOpen: (open: boolean) => void
  setCurrentSpaceKey: (spaceKey: SpaceKey) => void
  setCurrentUser: (user: SessionUser | null) => void
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
  toggleTabsEnabled: () => void
  setTabsEnabled: (enabled: boolean) => void
}

function sortTabsByPin(tabs: ShellTab[]) {
  const pinnedTabs = tabs.filter((item) => item.pinned)
  const normalTabs = tabs.filter((item) => !item.pinned)
  return [...pinnedTabs, ...normalTabs]
}

function readStoredTabsState(): StoredTabsState {
  if (typeof window === 'undefined') {
    return {
      openTabs: [],
      tabsEnabled: true,
    }
  }

  try {
    const rawValue = window.localStorage.getItem(SHELL_TABS_STORAGE_KEY)
    if (!rawValue) {
      return {
        openTabs: [],
        tabsEnabled: true,
      }
    }

    const parsed = JSON.parse(rawValue) as Partial<StoredTabsState>
    return {
      openTabs: Array.isArray(parsed.openTabs) ? parsed.openTabs : [],
      tabsEnabled: parsed.tabsEnabled !== false,
    }
  } catch {
    return {
      openTabs: [],
      tabsEnabled: true,
    }
  }
}

function writeStoredTabsState(state: StoredTabsState) {
  if (typeof window === 'undefined') {
    return
  }

  window.localStorage.setItem(SHELL_TABS_STORAGE_KEY, JSON.stringify(state))
}

function syncTabState(openTabs: ShellTab[], tabsEnabled: boolean) {
  const nextState = {
    openTabs: sortTabsByPin(openTabs),
    tabsEnabled,
  }

  writeStoredTabsState(nextState)
  return nextState
}

const initialStoredTabsState = readStoredTabsState()
const initialTabState = syncTabState(initialStoredTabsState.openTabs, initialStoredTabsState.tabsEnabled)

export const useShellStore = create<ShellState>((set) => ({
  themeMode: 'light',
  navCollapsed: false,
  mobileNavOpen: false,
  currentSpaceKey: 'default',
  currentUser: null,
  activeTopContext: '首页',
  openTabs: initialTabState.openTabs,
  tabsEnabled: initialTabState.tabsEnabled,
  toggleTheme: () => set((state) => ({ themeMode: state.themeMode === 'light' ? 'dark' : 'light' })),
  toggleNavCollapsed: () => set((state) => ({ navCollapsed: !state.navCollapsed })),
  setMobileNavOpen: (mobileNavOpen) => set({ mobileNavOpen }),
  setCurrentSpaceKey: (currentSpaceKey) => set({ currentSpaceKey }),
  setCurrentUser: (currentUser) => set({ currentUser }),
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

      return syncTabState(nextTabs, state.tabsEnabled)
    }),
  closeTab: (path) =>
    set((state) => syncTabState(state.openTabs.filter((item) => item.path !== path), state.tabsEnabled)),
  closeTabs: (paths) =>
    set((state) => {
      if (!paths.length) {
        return {}
      }

      const pathSet = new Set(paths)
      return syncTabState(
        state.openTabs.filter((item) => !pathSet.has(item.path)),
        state.tabsEnabled,
      )
    }),
  clearTabs: () =>
    set((state) => {
      writeStoredTabsState({
        openTabs: [],
        tabsEnabled: state.tabsEnabled,
      })

      return {
        openTabs: [],
      }
    }),
  toggleTabPinned: (path) =>
    set((state) =>
      syncTabState(
        state.openTabs.map((item) =>
          item.path === path
            ? {
                ...item,
                pinned: !item.pinned,
              }
            : item,
        ),
        state.tabsEnabled,
      ),
    ),
  closeOtherTabs: (path) =>
    set((state) =>
      syncTabState(
        state.openTabs.filter((item) => item.path === path || item.pinned),
        state.tabsEnabled,
      ),
    ),
  closeTabsToLeft: (path) =>
    set((state) => {
      const targetIndex = state.openTabs.findIndex((item) => item.path === path)
      if (targetIndex < 0) {
        return {}
      }

      return syncTabState(
        state.openTabs.filter((item, index) => index >= targetIndex || item.pinned),
        state.tabsEnabled,
      )
    }),
  closeTabsToRight: (path) =>
    set((state) => {
      const targetIndex = state.openTabs.findIndex((item) => item.path === path)
      if (targetIndex < 0) {
        return {}
      }

      return syncTabState(
        state.openTabs.filter((item, index) => index <= targetIndex || item.pinned),
        state.tabsEnabled,
      )
    }),
  reorderTabs: (sourcePath, targetPath) =>
    set((state) => {
      if (sourcePath === targetPath) {
        return {}
      }

      const nextTabs = state.openTabs.slice()
      const sourceIndex = nextTabs.findIndex((item) => item.path === sourcePath)
      const targetIndex = nextTabs.findIndex((item) => item.path === targetPath)

      if (sourceIndex < 0 || targetIndex < 0) {
        return {}
      }

      const [sourceTab] = nextTabs.splice(sourceIndex, 1)
      nextTabs.splice(targetIndex, 0, sourceTab)

      return syncTabState(nextTabs, state.tabsEnabled)
    }),
  toggleTabsEnabled: () =>
    set((state) => {
      const tabsEnabled = !state.tabsEnabled
      writeStoredTabsState({
        openTabs: state.openTabs,
        tabsEnabled,
      })

      return { tabsEnabled }
    }),
  setTabsEnabled: (tabsEnabled) =>
    set((state) => {
      writeStoredTabsState({
        openTabs: state.openTabs,
        tabsEnabled,
      })

      return { tabsEnabled }
    }),
}))
