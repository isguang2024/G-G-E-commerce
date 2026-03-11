# 国际化 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/in-depth/locale.html

## 概述

项目使用 vue-i18n 插件，目前集成了中文和英文两种语言包。

目前对菜单、顶栏、设置中心等组件进行了国际化，其他地方根据需求自行配置。

## 目录结构

```bash
├── language
│   ├── index.ts      // 配置文件
│   └── locales       // 语言包目录
│       ├── zh.json   // 中文包
│       └── en.json   // 英文包
```

## 在模版中使用

```html
<p>{{ $t('setting.color.title') }}</p>
```

## 如何获取当前语言

```typescript
import { useI18n } from "vue-i18n";
const { locale } = useI18n();
```

## 如何配置多语言

修改 `src/locales/index.ts` 在 messages 中增加你要的配置的语言，然后在 langs 目录新建一个文件，如 `en.ts`：

```typescript
import { createI18n } from "vue-i18n";
import en from "./en";
import zh from "./zh";
import { LanguageEnum } from "@/enums/appEnum";

const lang = createI18n({
  locale: LanguageEnum.ZH,        // 设置语言类型
  legacy: false,                  // 如果要支持compositionAPI，此项必须设置为false
  globalInjection: true,          // 全局注册$t方法
  fallbackLocale: LanguageEnum.ZH, // 设置备用语言
  messages: {
    en,
    zh,
  },
});

export default lang;
```
