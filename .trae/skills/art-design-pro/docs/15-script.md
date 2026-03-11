# 脚本说明 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/in-depth/script.html

## 可用脚本

Art Design Pro 提供了丰富的 npm 脚本，用于开发、构建、代码检查等。

### 开发相关

#### `dev`
- **命令**: `vite --open`
- **描述**: 启动 Vite 开发服务器，并在默认浏览器中自动打开应用
- **用途**: 日常开发和调试

#### `serve`
- **命令**: `vite preview`
- **描述**: 预览构建后的应用，模拟生产环境
- **用途**: 本地查看构建结果

### 构建相关

#### `build`
- **命令**: `vue-tsc --noEmit && vite build`
- **描述**: 
  1. 运行 TypeScript 类型检查（不生成输出文件）
  2. 使用 Vite 构建生产版本
- **用途**: 打包生产环境代码

### 代码质量

#### `lint`
- **命令**: `eslint`
- **描述**: 运行 ESLint 检查代码质量和代码风格问题
- **用途**: 检查代码规范

#### `fix`
- **命令**: `eslint --fix`
- **描述**: 运行 ESLint 并自动修复可修复的问题
- **用途**: 自动修复代码格式问题

#### `lint:prettier`
- **命令**: `prettier --write "**/*.{js,cjs,ts,json,tsx,css,less,scss,vue,html,md}"`
- **描述**: 使用 Prettier 格式化所有指定类型的文件
- **用途**: 统一代码风格

#### `lint:stylelint`
- **命令**: `stylelint "**/*.{css,scss,vue}" --fix`
- **描述**: 使用 Stylelint 检查并自动修复样式问题
- **用途**: 检查 CSS/SCSS 代码规范

#### `lint:lint-staged`
- **命令**: `lint-staged`
- **描述**: 仅检查和格式化暂存的文件
- **用途**: 提交前代码质量检查

### Git 相关

#### `prepare`
- **命令**: `husky`
- **描述**: 设置 Husky Git 钩子
- **用途**: 在 Git 操作前运行预定义的脚本

#### `commit`
- **命令**: `git-cz`
- **描述**: 使用 Commitizen 规范化提交消息
- **用途**: 确保提交信息格式一致

## 💡 使用建议

### 日常开发流程

```bash
# 1. 启动开发服务器
pnpm dev

# 2. 代码提交前检查
pnpm lint:lint-staged

# 3. 规范化提交
pnpm commit
```

### 代码质量保证

```bash
# 定期运行代码检查
pnpm lint

# 自动修复问题
pnpm fix

# 格式化所有代码
pnpm lint:prettier
pnpm lint:stylelint
```

### 生产构建

```bash
# 类型检查 + 构建
pnpm build

# 预览构建结果
pnpm serve
```

## 📝 注意事项

1. **提交前检查**: 使用 `pnpm commit` 而不是 `git commit`，确保提交信息规范
2. **自动修复**: 优先使用 `pnpm fix` 自动修复可修复的问题
3. **类型安全**: `build` 命令会先进行类型检查，确保代码类型安全
4. **暂存检查**: `lint:lint-staged` 只检查暂存文件，提高提交速度

## 📚 相关文档

- 代码规范：`docs/18-standard.md`
- 构建部署：`docs/12-build.md`
- Git 提交规范：`docs/18-standard.md`（提交规范部分）
