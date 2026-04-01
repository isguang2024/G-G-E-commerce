import type { SpaceKey } from '@/shared/types/navigation'

export interface SessionUser {
  id: string
  displayName: string
  title: string
  email: string
  primarySpaceKey: SpaceKey
  badges: string[]
}
