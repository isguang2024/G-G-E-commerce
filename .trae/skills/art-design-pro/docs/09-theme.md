# 主题配置 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/essentials/theme.html

## CSS 主题变量

CSS 变量包括主题颜色、背景颜色、文字颜色、边框颜色、阴影等，能自适应 Light 和 Dark 模式。

路径：`src/assets/styles/variables.scss`

## 使用示例

```scss
// 文字
color: var(--art-gray-100);
color: var(--art-gray-900);

// 边框
border: 1px solid var(--art-border-color);
border: 1px solid var(--art-border-dashed-color);

// 背景颜色
background-color: var(--art-main-bg-color);

// 阴影
box-shadow: var(--art-box-shadow);
box-shadow: var(--art-box-shadow-xs);

// 使用带透明度的颜色
color: rgba(var(--art-gray-800-rgb), 0.6);

// 主题色
color: var(--main-color);
background-color: var(--el-color-primary-light-1); // 最深
background-color: var(--el-color-primary-light-9); // 最浅
```

## Light 主题变量

```scss
:root {
  // Theme color
  --art-primary: 93, 135, 255;
  --art-secondary: 73, 190, 255;
  --art-error: 250, 137, 107;
  --art-info: 83, 155, 255;
  --art-success: 19, 222, 185;
  --art-warning: 255, 174, 31;
  --art-danger: 255, 77, 79;

  // Background color
  --art-gray-100: #f9f9f9;
  --art-gray-200: #f1f1f4;
  --art-gray-300: #dbdfe9;
  --art-gray-400: #c4cada;
  --art-gray-500: #99a1b7;
  --art-gray-600: #78829d;
  --art-gray-700: #4b5675;
  --art-gray-800: #252f4a;
  --art-gray-900: #071437;

  // Border
  --art-border-color: #eaebf1;
  --art-border-dashed-color: #dbdfe9;

  // Shadow
  --art-box-shadow-xs: 0 0.1rem 0.75rem 0.25rem rgba(0, 0, 0, 0.05);
  --art-box-shadow-sm: 0 0.1rem 1rem 0.25rem rgba(0, 0, 0, 0.05);
  --art-box-shadow: 0 0.5rem 1.5rem 0.5rem rgba(0, 0, 0, 0.075);
  --art-box-shadow-lg: 0 1rem 2rem 1rem rgba(0, 0, 0, 0.1);

  // Background
  --art-bg-color: #fafbfc;
  --art-main-bg-color: #ffffff;
}
```

## Dark 主题变量

```scss
html.dark {
  // Theme color
  --art-primary: 93, 135, 255;

  // Background color
  --art-gray-100: #1b1c22;
  --art-gray-200: #26272f;
  --art-gray-300: #363843;
  --art-gray-400: #464852;
  --art-gray-500: #636674;
  --art-gray-600: #808290;
  --art-gray-700: #9a9cae;
  --art-gray-800: #b5b7c8;
  --art-gray-900: #f5f5f5;

  // Border
  --art-border-color: #26272f;
  --art-border-dashed-color: #363843;

  // Background
  --art-bg-color: #070707;
  --art-main-bg-color: #161618;
}
```

## 媒体查询（设备尺寸）

```scss
$device-notebook: 1600px;     // notebook
$device-ipad-pro: 1180px;     // ipad pro
$device-ipad: 800px;          // ipad
$device-ipad-vertical: 900px; // ipad-竖屏
$device-phone: 500px;         // mobile
```
