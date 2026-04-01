import type { PageMeta } from '@/shared/types/navigation'

export const pageMetaMock: Record<string, PageMeta> = {
  welcome: {
    routeId: 'welcome',
    title: '首页',
    subtitle: '这一页只负责说明新迁移线的目标、当前边界和后续迁移方式。',
    groupLabel: 'Welcome',
    breadcrumbs: [{ label: '首页', path: '/welcome' }],
  },
  'workspace-home': {
    routeId: 'workspace-home',
    title: '工作台',
    subtitle: '保留 Vue 主线里的工作区壳层语义，但先用新的容器、标题区和模块引导表达。',
    groupLabel: 'Workspace',
    breadcrumbs: [{ label: '工作台', path: '/workspace' }],
  },
  'workspace-inbox': {
    routeId: 'workspace-inbox',
    title: '收件中心',
    subtitle: '后续将接入消息、通知和待办链路；当前只保留迁移占位。',
    groupLabel: 'Workspace',
    breadcrumbs: [
      { label: '工作台', path: '/workspace' },
      { label: '收件中心', path: '/workspace/inbox' },
    ],
  },
  'team-home': {
    routeId: 'team-home',
    title: '团队总览',
    subtitle: '团队资料、成员边界和协作视角会在后续迁移中逐步替换为 React 页面。',
    groupLabel: 'Team',
    breadcrumbs: [
      { label: '团队协作', path: '/team' },
      { label: '团队总览', path: '/team' },
    ],
  },
  'team-members': {
    routeId: 'team-members',
    title: '团队成员',
    subtitle: '成员列表与角色边界将在壳层稳定后迁入 React。',
    groupLabel: 'Team',
    breadcrumbs: [
      { label: '团队协作', path: '/team' },
      { label: '团队成员', path: '/team/members' },
    ],
  },
  'team-roles': {
    routeId: 'team-roles',
    title: '团队角色与权限',
    subtitle: '先保留角色治理的入口位置，后续再对接真实权限和数据源。',
    groupLabel: 'Team',
    breadcrumbs: [
      { label: '团队协作', path: '/team' },
      { label: '团队角色与权限', path: '/team/roles' },
    ],
  },
  'message-home': {
    routeId: 'message-home',
    title: '消息中心',
    subtitle: '通知、模板、收件对象与记录链路将在后续阶段逐步替换为 React 模块。',
    groupLabel: 'Message',
    breadcrumbs: [
      { label: '消息中心', path: '/message' },
      { label: '概览', path: '/message' },
    ],
  },
  'message-template': {
    routeId: 'message-template',
    title: '消息模板',
    subtitle: '沿用当前信息架构的导航位置，但不复刻旧页面实现。',
    groupLabel: 'Message',
    breadcrumbs: [
      { label: '消息中心', path: '/message' },
      { label: '消息模板', path: '/message/template' },
    ],
  },
  'system-home': {
    routeId: 'system-home',
    title: '系统管理',
    subtitle: '平台治理空间的主入口，用于承接页面、菜单、接口、角色和功能包等治理型模块。',
    groupLabel: 'System',
    breadcrumbs: [{ label: '系统管理', path: '/system' }],
  },
  'system-menu': {
    routeId: 'system-menu',
    title: '菜单管理',
    subtitle: '首期只验证标题区、操作区、筛选区和工作区结构，不接真实 API。',
    groupLabel: 'System',
    breadcrumbs: [
      { label: '系统管理', path: '/system' },
      { label: '菜单管理', path: '/system/menu' },
    ],
  },
  'system-page': {
    routeId: 'system-page',
    title: '页面管理',
    subtitle: '保留治理入口位置，后续迁移页面、路由与页面元数据维护能力。',
    groupLabel: 'System',
    breadcrumbs: [
      { label: '系统管理', path: '/system' },
      { label: '页面管理', path: '/system/page' },
    ],
  },
  'system-role': {
    routeId: 'system-role',
    title: '角色管理',
    subtitle: '平台角色与菜单/功能权限的治理入口将在后续阶段迁入。',
    groupLabel: 'System',
    breadcrumbs: [
      { label: '系统管理', path: '/system' },
      { label: '角色管理', path: '/system/role' },
    ],
  },
  'system-user': {
    routeId: 'system-user',
    title: '用户管理',
    subtitle: '首期仅保留入口和标题层，等待真实 API 与列表模式接入。',
    groupLabel: 'System',
    breadcrumbs: [
      { label: '系统管理', path: '/system' },
      { label: '用户管理', path: '/system/user' },
    ],
  },
  'system-api': {
    routeId: 'system-api',
    title: 'API 端点管理',
    subtitle: '后续用于承接接口注册、分类与权限归属的治理页面。',
    groupLabel: 'System',
    breadcrumbs: [
      { label: '系统管理', path: '/system' },
      { label: 'API 端点管理', path: '/system/api-endpoint' },
    ],
  },
  'system-package': {
    routeId: 'system-package',
    title: '功能包管理',
    subtitle: '保留功能包治理的导航位置，后续接入真实数据。',
    groupLabel: 'System',
    breadcrumbs: [
      { label: '系统管理', path: '/system' },
      { label: '功能包管理', path: '/system/feature-package' },
    ],
  },
}
