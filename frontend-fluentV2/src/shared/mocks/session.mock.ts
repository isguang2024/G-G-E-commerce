import type { SessionUser } from '@/shared/types/session'

export const currentUserMock: SessionUser = {
  id: 'user-shell-admin',
  displayName: '张启航',
  title: '平台治理负责人',
  email: 'qihang.zhang@gg-e.local',
  primarySpaceKey: 'platform-governance',
  badges: ['平台管理员', '迁移试点'],
}

export function buildMockSessionUser(account: string): SessionUser {
  const normalizedAccount = account.trim()
  const derivedName = normalizedAccount.includes('@')
    ? normalizedAccount.split('@')[0] || currentUserMock.displayName
    : normalizedAccount || currentUserMock.displayName
  const email = normalizedAccount.includes('@') ? normalizedAccount : `${normalizedAccount || 'user'}@gg-e.local`

  return {
    ...currentUserMock,
    id: `user-${derivedName.toLowerCase().replace(/[^a-z0-9]+/g, '-') || 'shell-admin'}`,
    displayName: derivedName,
    email,
  }
}
