import { matchPath } from 'react-router-dom'
import type { RouteDefinition } from '@/shared/types/navigation'
import type { ShellTab } from '@/shared/types/navigation'
import { pageMetaMock } from '@/shared/mocks/page-meta.mock'
import { MigrationPlaceholderPage } from '@/pages/placeholder/MigrationPlaceholderPage'
import { SystemHomePage } from '@/pages/system/SystemHomePage'
import { SystemMenuPage } from '@/pages/system/SystemMenuPage'
import { WelcomePage } from '@/pages/welcome/WelcomePage'
import { WorkspaceHomePage } from '@/pages/workspace/WorkspaceHomePage'

export const routeDefinitions: RouteDefinition[] = [
  { id: 'welcome', path: '/welcome', group: 'welcome', status: 'implemented', shellTitle: '首页' },
  { id: 'workspace-home', path: '/workspace', group: 'workspace', status: 'implemented', shellTitle: '工作台' },
  { id: 'workspace-inbox', path: '/workspace/inbox', group: 'workspace', status: 'placeholder', shellTitle: '收件中心' },
  { id: 'team-home', path: '/team', group: 'team', status: 'placeholder', shellTitle: '团队总览' },
  { id: 'team-members', path: '/team/members', group: 'team', status: 'placeholder', shellTitle: '团队成员' },
  { id: 'team-roles', path: '/team/roles', group: 'team', status: 'placeholder', shellTitle: '团队角色与权限' },
  { id: 'message-home', path: '/message', group: 'message', status: 'placeholder', shellTitle: '消息中心' },
  { id: 'message-template', path: '/message/template', group: 'message', status: 'placeholder', shellTitle: '消息模板' },
  { id: 'system-home', path: '/system', group: 'system', status: 'implemented', shellTitle: '系统管理' },
  { id: 'system-menu', path: '/system/menu', group: 'system', status: 'implemented', shellTitle: '菜单管理' },
  { id: 'system-page', path: '/system/page', group: 'system', status: 'placeholder', shellTitle: '页面管理' },
  { id: 'system-role', path: '/system/role', group: 'system', status: 'placeholder', shellTitle: '角色管理' },
  { id: 'system-user', path: '/system/user', group: 'system', status: 'placeholder', shellTitle: '用户管理' },
  { id: 'system-api', path: '/system/api-endpoint', group: 'system', status: 'placeholder', shellTitle: 'API 端点管理' },
  { id: 'system-package', path: '/system/feature-package', group: 'system', status: 'placeholder', shellTitle: '功能包管理' },
]

const groupLabelMap: Record<RouteDefinition['group'], string> = {
  welcome: '首页',
  workspace: '工作台',
  team: '团队协作',
  message: '消息中心',
  system: '系统管理',
}

const routeMap = new Map(routeDefinitions.map((definition) => [definition.id, definition]))

export function getRouteDefinition(routeId: string) {
  return routeMap.get(routeId)
}

export function resolveRouteByPath(pathname: string) {
  return [...routeDefinitions]
    .sort((left, right) => right.path.length - left.path.length)
    .find((definition) => matchPath({ path: definition.path, end: true }, pathname))
}

export function resolveShellTabByPath(pathname: string): ShellTab | null {
  const routeDefinition = resolveRouteByPath(pathname)
  if (!routeDefinition) {
    return null
  }

  const pageMeta = pageMetaMock[routeDefinition.id]

  return {
    routeId: routeDefinition.id,
    path: routeDefinition.path,
    label: pageMeta?.title || routeDefinition.shellTitle,
    group: routeDefinition.group,
    groupLabel: groupLabelMap[routeDefinition.group],
  }
}

export function renderRouteElement(routeId: string) {
  switch (routeId) {
    case 'welcome':
      return <WelcomePage routeId={routeId} />
    case 'workspace-home':
      return <WorkspaceHomePage routeId={routeId} />
    case 'system-home':
      return <SystemHomePage routeId={routeId} />
    case 'system-menu':
      return <SystemMenuPage routeId={routeId} />
    default:
      return <MigrationPlaceholderPage routeId={routeId} />
  }
}
