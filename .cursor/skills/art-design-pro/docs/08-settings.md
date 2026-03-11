# 系统配置 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/essentials/settings.html

## 系统 Logo 配置

配置文件：`src/components/core/base/ArtLogo.vue`

```vue
<template>
  <div class="art-logo">
    <img :style="logoStyle" src="@imgs/common/logo.png" alt="logo" />
  </div>
</template>
```

如需更换 Logo，只需修改图片资源路径即可。

## 系统名称配置

配置文件：`src/config/index.ts`

```typescript
const appConfig: SystemConfig = {
  systemInfo: {
    name: "Art Design Pro", // 系统名称
  },
};
```

## 全局配置

配置文件路径：`src/config/setting.ts`

```typescript
const appConfig: SystemConfig = {
  // 系统信息
  systemInfo: {
    name: "Art Design Pro",
  },

  // 系统主题列表
  settingThemeList: [
    {
      name: "Light",
      theme: SystemThemeEnum.LIGHT,
      color: ["#fff", "#fff"],
    },
    {
      name: "Dark",
      theme: SystemThemeEnum.DARK,
      color: ["#22252A"],
    },
    {
      name: "System",
      theme: SystemThemeEnum.AUTO,
      color: ["#fff", "#22252A"],
    },
  ],

  // 菜单布局列表
  menuLayoutList: [
    { name: "Left", value: MenuTypeEnum.LEFT },
    { name: "Top", value: MenuTypeEnum.TOP },
    { name: "Mixed", value: MenuTypeEnum.TOP_LEFT },
    { name: "Dual Column", value: MenuTypeEnum.DUAL_MENU },
  ],

  // 系统主色
  systemMainColor: [
    "#5D87FF",  // 主色
    "#B48DF3",  // 紫色
    "#1D84FF",  // 蓝色
    "#60C041",  // 绿色
    "#38C0FC",  // 青色
    "#F9901F",  // 橙色
    "#FF80C8",  // 粉色
  ] as const,

  // 系统其他项默认配置
  systemSetting: {
    defaultMenuWidth: 240,        // 菜单宽度
    defaultCustomRadius: "0.75",  // 自定义圆角
    defaultTabStyle: "tab-default", // 标签样式
  },
};
```
