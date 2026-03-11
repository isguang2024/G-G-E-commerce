# 同步代码 | Art Design Pro

来源：https://www.artd.pro/docs/zh/guide/update.html

## 同步开源项目更新

当开源项目有更新时，您可以通过以下步骤同步代码到您的项目：

### 步骤 1：添加开源仓库地址

在自己的仓库里面新增开源仓库地址：

```bash
git remote add upstream https://github.com/Daymychen/art-design-pro
```

### 步骤 2：拉取开源仓库的更新

```bash
git fetch upstream
```

### 步骤 3：合并更新

拉取开源项目更新代码：

```bash
# 切换到本地 main 分支
git checkout main

# 合并更新
git merge upstream/main
```

### 步骤 4：解决冲突

如果代码有冲突：

1. 手动解决冲突文件
2. 测试解决冲突后的代码
3. 提交代码

## 💡 最佳实践

1. **定期同步**：建议每月同步一次，获取最新功能和修复
2. **分支保护**：在非 main 分支上进行合并测试
3. **备份重要修改**：同步前确保重要代码已提交
4. **查看更新日志**：了解新版本的变化和Breaking Changes
