import type { PropsWithChildren } from 'react'
import { FluentProvider } from '@fluentui/react-components'
import { useShellStore } from '@/features/shell/store/useShellStore'
import { appDarkTheme, appLightTheme } from '@/shared/config/theme'

export function FluentThemeProvider({ children }: PropsWithChildren) {
  const themeMode = useShellStore((state) => state.themeMode)

  return <FluentProvider theme={themeMode === 'dark' ? appDarkTheme : appLightTheme}>{children}</FluentProvider>
}
