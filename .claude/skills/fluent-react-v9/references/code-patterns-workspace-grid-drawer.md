# 工作区列表 + 抽屉代码模式

用于后台最常见的“列表治理页”。

以下代码骨架强调结构与职责划分，具体 props、hooks 和列定义以当前 Fluent UI React v9 Storybook 为准。

## 适用场景

- 用户、角色、菜单、工单、订单、资源记录
- 主体是列表，详情或编辑放在右侧抽屉
- 希望保留主列表上下文，不强跳整页

## 结构要点

- 页面头部：标题、摘要、主操作
- 筛选区：高频筛选与搜索
- 工具栏：刷新、批量操作、视图切换
- 主区：`DataGrid`
- 次级区：`InlineDrawer` 或响应式 `Drawer`

## 代码骨架

```tsx
import * as React from 'react';
import {
  Button,
  Drawer,
  DrawerBody,
  DrawerHeader,
  DrawerHeaderTitle,
  Field,
  Input,
  MessageBar,
  Toolbar,
  ToolbarButton,
  createTableColumn,
  DataGrid,
  DataGridBody,
  DataGridCell,
  DataGridHeader,
  DataGridHeaderCell,
  DataGridRow,
} from '@fluentui/react-components';

type RowItem = {
  id: string;
  name: string;
  status: string;
  owner: string;
};

const columns = [
  createTableColumn<RowItem>({
    columnId: 'name',
    renderHeaderCell: () => '名称',
    renderCell: item => item.name,
  }),
  createTableColumn<RowItem>({
    columnId: 'status',
    renderHeaderCell: () => '状态',
    renderCell: item => item.status,
  }),
  createTableColumn<RowItem>({
    columnId: 'owner',
    renderHeaderCell: () => '负责人',
    renderCell: item => item.owner,
  }),
];

export function WorkspaceGridDrawerPage() {
  const [query, setQuery] = React.useState('');
  const [selected, setSelected] = React.useState<RowItem | null>(null);
  const [open, setOpen] = React.useState(false);

  const items = React.useMemo<RowItem[]>(
    () => [
      { id: '1', name: '用户同步任务', status: '运行中', owner: '系统管理员' },
      { id: '2', name: '菜单审计记录', status: '已完成', owner: '运营团队' },
    ],
    [],
  );

  const filtered = items.filter(item => item.name.includes(query));

  return (
    <div>
      <header>
        <h1>任务管理</h1>
        <p>查看状态、处理记录，并在右侧抽屉完成详情与编辑。</p>
        <Button appearance="primary">新建任务</Button>
      </header>

      <section>
        <Field label="搜索">
          <Input value={query} onChange={(_, data) => setQuery(data.value)} />
        </Field>
      </section>

      <Toolbar aria-label="任务工具栏">
        <ToolbarButton>刷新</ToolbarButton>
        <ToolbarButton>导出</ToolbarButton>
      </Toolbar>

      <DataGrid items={filtered} columns={columns}>
        <DataGridHeader>
          <DataGridRow>
            {({ renderHeaderCell }) => (
              <DataGridHeaderCell>{renderHeaderCell()}</DataGridHeaderCell>
            )}
          </DataGridRow>
        </DataGridHeader>
        <DataGridBody<RowItem>>
          {({ item, rowId }) => (
            <DataGridRow
              key={rowId}
              onClick={() => {
                setSelected(item);
                setOpen(true);
              }}
            >
              {({ renderCell }) => <DataGridCell>{renderCell(item)}</DataGridCell>}
            </DataGridRow>
          )}
        </DataGridBody>
      </DataGrid>

      <Drawer open={open} onOpenChange={(_, data) => setOpen(data.open)}>
        <DrawerHeader>
          <DrawerHeaderTitle>{selected?.name ?? '详情'}</DrawerHeaderTitle>
        </DrawerHeader>
        <DrawerBody>
          {selected ? (
            <div>
              <MessageBar>这里放详情、编辑表单或操作记录。</MessageBar>
            </div>
          ) : null}
        </DrawerBody>
      </Drawer>
    </div>
  );
}
```

## 代码评审重点

- 列表和详情职责是否分开
- `DataGrid` 是否真的是最佳选择
- 抽屉是否真的保留了主列表上下文
- 行点击、勾选、行内按钮是否有冲突
