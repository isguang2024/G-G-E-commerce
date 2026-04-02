# 树 + 详情面板代码模式

用于菜单树、资源树、分类树、组织结构树这类后台页面。

以下代码骨架强调布局与分工，具体 API 以当前 Fluent UI React v9 Storybook 为准。

## 适用场景

- 菜单管理
- 文件或资源层级管理
- 组织结构和分类维护

## 结构要点

- 左侧 `Tree`：层级浏览、定位、切换节点
- 顶部工具栏：创建、刷新、筛选
- 右侧详情区：属性、说明、操作、子项摘要
- 复杂编辑可升级为 `Drawer`

## 代码骨架

```tsx
import * as React from 'react';
import {
  Button,
  Field,
  Input,
  MessageBar,
  Toolbar,
  ToolbarButton,
  Tree,
  TreeItem,
  TreeItemLayout,
} from '@fluentui/react-components';

type TreeNode = {
  id: string;
  label: string;
  children?: TreeNode[];
};

const data: TreeNode[] = [
  {
    id: 'system',
    label: '系统管理',
    children: [
      { id: 'menu', label: '菜单管理' },
      { id: 'page', label: '页面管理' },
    ],
  },
  {
    id: 'team',
    label: '团队空间',
    children: [{ id: 'members', label: '成员管理' }],
  },
];

function renderNodes(nodes: TreeNode[], onSelect: (node: TreeNode) => void): React.ReactNode {
  return nodes.map(node => (
    <TreeItem key={node.id} itemType={node.children?.length ? 'branch' : 'leaf'}>
      <TreeItemLayout onClick={() => onSelect(node)}>{node.label}</TreeItemLayout>
      {node.children?.length ? renderNodes(node.children, onSelect) : null}
    </TreeItem>
  ));
}

export function TreeDetailPanePage() {
  const [selected, setSelected] = React.useState<TreeNode | null>(null);
  const [keyword, setKeyword] = React.useState('');

  return (
    <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: 16 }}>
      <aside>
        <Field label="筛选节点">
          <Input value={keyword} onChange={(_, data) => setKeyword(data.value)} />
        </Field>

        <Toolbar aria-label="树工具栏">
          <ToolbarButton>新建</ToolbarButton>
          <ToolbarButton>刷新</ToolbarButton>
        </Toolbar>

        <Tree aria-label="菜单树">{renderNodes(data, setSelected)}</Tree>
      </aside>

      <main>
        {selected ? (
          <section>
            <h2>{selected.label}</h2>
            <p>这里放节点属性、说明、权限、关联对象和操作。</p>
            <Button appearance="primary">编辑节点</Button>
          </section>
        ) : (
          <MessageBar>请选择左侧节点查看详情。</MessageBar>
        )}
      </main>
    </div>
  );
}
```

## 代码评审重点

- 左侧树是否只负责定位，不承担复杂编辑
- 右侧详情区是否承载主编辑任务
- 树节点点击、展开、选择行为是否清晰
- 是否错误地把应用导航和对象树混在一起
