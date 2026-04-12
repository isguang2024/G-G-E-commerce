# 4.1 文档导航模板结构设计

生成时间：2026-04-12  
任务节点：`V5-FOLDER-OPTIMIZE/STAGE-DOCS-NAVIGATION/1`

## 设计结论

文档导航采用统一五段式结构：

1. 文件定位
2. 快速入口
3. 主题索引
4. 阅读顺序
5. 维护约定

该结构覆盖“找入口、看职责、按场景阅读、持续维护”四类核心需求。

## 模板落地

- 新增模板文件：`docs/guides/doc-navigation-template.md`
- 模板适配对象：
  - 根 `README.md`
  - `docs/README.md`
  - `docs/INDEX.md`
  - 核心子目录 `README.md`

## 后续动作

- 4.2 基于该模板校准 `docs/INDEX.md`。
- 4.3 补齐核心目录 README 覆盖率。
