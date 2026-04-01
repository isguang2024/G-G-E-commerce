import type { NavigationSpace } from '@/shared/types/navigation'

export const spacesMock: NavigationSpace[] = [
  {
    key: 'default',
    label: '默认菜单空间',
    description: '用于壳层验证与跨域迁移预览的公共入口。',
    kind: 'default',
    defaultLandingRoute: '/welcome',
  },
  {
    key: 'platform-governance',
    label: '平台治理空间',
    description: '偏平台治理与系统配置的管理上下文。',
    kind: 'platform',
    defaultLandingRoute: '/system',
  },
  {
    key: 'team-collaboration',
    label: '团队协作空间',
    description: '偏团队工作台、成员协作与消息收件链路的上下文。',
    kind: 'team',
    defaultLandingRoute: '/workspace',
  },
]
