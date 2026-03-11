# 规范 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/project/standard.html

## 代码规范工具

- **Eslint**: JS 代码检查
- **Prettier**: 代码格式化
- **Stylelint**: CSS 代码检查
- **Commitlint**: Git 提交信息检查
- **Husky**: Git 钩子工具
- **Lint-staged**: Git 提交前运行代码校验
- **cz-git**: 可视化提交工具

## 自动化

代码提交会自动执行配置好的文件，自动完成代码校验和格式化。

`package.json` 配置：

```json
{
  "lint-staged": {
    "*.{js,ts}": [
      "eslint --fix",
      "prettier --write"
    ],
    "*.{cjs,json}": [
      "prettier --write"
    ],
    "*.{vue,html}": [
      "eslint --fix",
      "prettier --write",
      "stylelint --fix"
    ],
    "*.{scss,css}": [
      "stylelint --fix",
      "prettier --write"
    ],
    "*.md": [
      "prettier --write"
    ]
  }
}
```

## 常用命令

```bash
# 检查项目中的 js 语法
pnpm lint

# 修复项目中 js 语法错误
pnpm fix

# 使用 Prettier 格式化所有文件
pnpm lint:prettier

# 使用 Stylelint 检查和修复样式
pnpm lint:stylelint

# 运行 lint-staged 检查暂存文件
pnpm lint:lint-staged

# 设置 Husky Git 钩子
pnpm prepare

# 使用 Commitizen 规范化提交消息
pnpm commit
```

## 提交规范

### 提交类型

```bash
feat     // 新增功能
fix      // 修复缺陷
docs     // 文档变更
style    // 代码格式（不影响功能）
refactor // 代码重构（不包括 bug 修复、功能新增）
perf     // 性能优化
test     // 添加疏漏测试或已有测试改动
build    // 构建流程、外部依赖变更
ci       // 修改 CI 配置、脚本
revert   // 回滚 commit
chore    // 对构建过程或辅助工具的更改
wip      // 对构建过程或辅助工具的更改
```

### 提交代码流程

```bash
git add .
pnpm commit
# 选择提交类型，填写提交信息
git push
```

## 注意事项

1. **不要跳过 Husky 钩子**：使用 `pnpm commit` 而不是 `git commit`
2. **提交前自动格式化**：Lint-staged 会自动格式化暂存的文件
3. **遵循提交规范**：使用 Commitizen 选择正确的提交类型
