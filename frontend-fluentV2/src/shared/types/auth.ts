export interface AuthSession {
  accessToken: string
  refreshToken: string
  expiresIn: number
  issuedAt: number
}

export interface CurrentUserRole {
  id: string
  code: string
  name: string
  description?: string
}

export interface CurrentUser {
  id: string
  username: string
  displayName: string
  email: string
  avatarUrl: string
  phone: string
  status: string
  isSuperAdmin: boolean
  currentTenantId: string | null
  roles: CurrentUserRole[]
  actions: string[]
  badges: string[]
}

export interface TenantContext {
  currentTenantId: string | null
}
