import { AppRouteRecord } from '@/types/router'

export const teamRoutes: AppRouteRecord = {
  path: '/team',
  name: 'TeamRoot',
  component: '/index/index',
  meta: {
    title: 'menus.system.team',
    icon: 'ri:team-line',
    roles: ['R_SUPER', 'R_ADMIN']
  },
  children: [
    {
      path: '',
      name: 'Team',
      component: '/team/team',
      meta: {
        title: 'menus.system.team',
        keepAlive: true,
        roles: ['R_SUPER', 'R_ADMIN']
      }
    },
    {
      path: 'team-roles-permissions',
      name: 'TeamRolesAndPermissions',
      component: '/system/team-roles-permissions',
      meta: {
        title: 'menus.system.teamRolesAndPermissions',
        keepAlive: true,
        roles: ['R_SUPER']
      }
    }
  ]
}
