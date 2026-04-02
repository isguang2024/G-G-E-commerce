export const queryKeys = {
  auth: {
    currentUser: ['auth', 'current-user'] as const,
  },
  navigation: {
    spaces: ['navigation', 'spaces'] as const,
    manifest: (spaceKey: string) => ['navigation', 'manifest', spaceKey] as const,
  },
  menu: {
    tree: (spaceKey: string) => ['menu', 'tree', spaceKey] as const,
    groups: ['menu', 'groups'] as const,
    runtimePages: (spaceKey: string) => ['menu', 'runtime-pages', spaceKey] as const,
  },
} as const
