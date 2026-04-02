import * as React from 'react';
import {
  Body1Strong,
  Button,
  Caption1,
  Field,
  Input,
  MessageBar,
  Title3,
  Toolbar,
  ToolbarButton,
  Tree,
  TreeItem,
  TreeItemLayout,
  makeStyles,
} from '@fluentui/react-components';

type TreeNode = {
  id: string;
  label: string;
  children?: TreeNode[];
};

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '16px',
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: '320px minmax(0, 1fr)',
    gap: '16px',
    alignItems: 'start',
  },
  side: {
    display: 'grid',
    gap: '12px',
  },
  detail: {
    display: 'grid',
    gap: '12px',
  },
});

const treeData: TreeNode[] = [
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
      {node.children?.length ? <Tree>{renderNodes(node.children, onSelect)}</Tree> : null}
    </TreeItem>
  ));
}

export function TreeDetailPanePage() {
  const styles = useStyles();
  const [keyword, setKeyword] = React.useState('');
  const [selected, setSelected] = React.useState<TreeNode | null>(treeData[0]);

  return (
    <div className={styles.page}>
      <header>
        <Title3>树形治理页</Title3>
        <Caption1>左侧 Tree 承担定位，右侧详情区承担属性和主编辑任务。</Caption1>
      </header>

      <div className={styles.layout}>
        <aside className={styles.side}>
          <Field label="筛选节点">
            <Input value={keyword} onChange={(_, data) => setKeyword(data.value)} />
          </Field>

          <Toolbar aria-label="树工具栏">
            <ToolbarButton>新建节点</ToolbarButton>
            <ToolbarButton>刷新</ToolbarButton>
          </Toolbar>

          <Tree aria-label="菜单树">{renderNodes(treeData, setSelected)}</Tree>
        </aside>

        <section className={styles.detail}>
          {selected ? (
            <>
              <Body1Strong>{selected.label}</Body1Strong>
              <Caption1>这里承接节点属性、说明、关联对象和操作区。</Caption1>
              <Button appearance="primary">编辑节点</Button>
            </>
          ) : (
            <MessageBar>请选择左侧节点查看详情。</MessageBar>
          )}
        </section>
      </div>
    </div>
  );
}
