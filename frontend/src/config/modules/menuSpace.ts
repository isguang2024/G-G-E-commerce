import type { MenuSpaceConfig } from '@/types/config'

const menuSpaceConfig: MenuSpaceConfig = {
  defaultSpaceKey: 'default',
  spaces: [
    {
      spaceKey: 'default',
      spaceName: '默认菜单空间',
      spaceType: 'default',
      description: '未配置 Host 时的默认菜单空间',
      enabled: true,
      isDefault: true,
      defaultLandingRoute: '/'
    }
  ],
  hostBindings: []
}

export default Object.freeze(menuSpaceConfig)
