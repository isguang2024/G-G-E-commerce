/**
 * 快速入口配置
 * 包含：应用列表、快速链接等配置
 */
import { WEB_LINKS } from '@/utils/constants'
import type { FastEnterConfig } from '@/types/config'

const fastEnterConfig: FastEnterConfig = {
  // 显示条件（屏幕宽度）
  minWidth: 1450,
  // 应用列表
  applications: [
    {
      id: 'console',
      name: '工作台',
      description: '系统概览与数据统计',
      icon: 'ri:pie-chart-line',
      iconColor: '#377dff',
      enabled: true,
      order: 1,
      routeName: 'Console'
    },
    {
      id: 'role',
      name: '角色管理',
      description: '维护个人空间角色与空间权限',
      icon: 'ri:shield-user-line',
      iconColor: '#0f766e',
      enabled: true,
      order: 2,
      routeName: 'Role'
    },
    {
      id: 'user',
      name: '用户管理',
      description: '查看平台账号、角色归属和权限诊断',
      icon: 'ri:user-settings-line',
      iconColor: '#2563eb',
      enabled: true,
      order: 3,
      routeName: 'User'
    },
    {
      id: 'menu',
      name: '菜单管理',
      description: '维护菜单树、菜单分组和备份',
      icon: 'ri:menu-line',
      iconColor: '#f97316',
      enabled: true,
      order: 4,
      routeName: 'Menus'
    },
    {
      id: 'page',
      name: '页面管理',
      description: '维护页面注册表和运行时页面',
      icon: 'ri:layout-4-line',
      iconColor: '#7c3aed',
      enabled: true,
      order: 5,
      routeName: 'PageManagement'
    },
    {
      id: 'api-endpoint',
      name: 'API 管理',
      description: '同步 API 注册表与诊断未注册接口',
      icon: 'ri:route-line',
      iconColor: '#dc2626',
      enabled: true,
      order: 6,
      routeName: 'ApiEndpoint'
    },
    {
      id: 'docs',
      name: '项目文档',
      description: '查看项目说明与规范文档',
      icon: 'ri:file-text-line',
      iconColor: '#0ea5e9',
      enabled: true,
      order: 7,
      link: WEB_LINKS.DOCS
    }
  ],
  // 快速链接
  quickLinks: [
    {
      id: 'user-center',
      name: '个人中心',
      enabled: true,
      order: 1,
      routeName: 'UserCenter'
    },
    {
      id: 'collaboration-workspace-members',
      name: '协作空间成员',
      enabled: true,
      order: 2,
      routeName: 'CollaborationWorkspaceMembers'
    },
    {
      id: 'feature-package',
      name: '功能包管理',
      enabled: true,
      order: 3,
      routeName: 'FeaturePackage'
    },
    {
      id: 'support',
      name: '技术支持',
      enabled: true,
      order: 4,
      link: WEB_LINKS.COMMUNITY
    }
  ]
}

export default Object.freeze(fastEnterConfig)
