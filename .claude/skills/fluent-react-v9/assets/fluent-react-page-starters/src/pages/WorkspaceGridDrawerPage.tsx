import * as React from 'react';
import {
  Badge,
  Body1Strong,
  Button,
  Caption1,
  DataGrid,
  DataGridBody,
  DataGridCell,
  DataGridHeader,
  DataGridHeaderCell,
  DataGridRow,
  Drawer,
  DrawerBody,
  DrawerHeader,
  DrawerHeaderTitle,
  Field,
  Input,
  MessageBar,
  TableCellLayout,
  Title3,
  Toolbar,
  ToolbarButton,
  createTableColumn,
  makeStyles,
  tokens,
} from '@fluentui/react-components';

type RowItem = {
  id: string;
  name: string;
  status: '运行中' | '已完成' | '待处理';
  owner: string;
};

const useStyles = makeStyles({
  page: {
    display: 'grid',
    gap: '16px',
  },
  header: {
    display: 'grid',
    gap: '8px',
  },
  layout: {
    display: 'grid',
    gridTemplateColumns: 'minmax(0, 1fr) 360px',
    gap: '16px',
    alignItems: 'start',
  },
  filters: {
    display: 'grid',
    gridTemplateColumns: 'minmax(220px, 320px)',
    gap: '12px',
  },
  drawerSurface: {
    minHeight: '420px',
    backgroundColor: tokens.colorNeutralBackground1,
  },
});

const rows: RowItem[] = [
  { id: '1', name: '用户同步任务', status: '运行中', owner: '系统管理员' },
  { id: '2', name: '菜单审计记录', status: '已完成', owner: '运营团队' },
  { id: '3', name: '权限清理工单', status: '待处理', owner: '安全小组' },
];

const columns = [
  createTableColumn<RowItem>({
    columnId: 'name',
    renderHeaderCell: () => '名称',
    renderCell: item => (
      <TableCellLayout description={item.id}>
        <Body1Strong>{item.name}</Body1Strong>
      </TableCellLayout>
    ),
  }),
  createTableColumn<RowItem>({
    columnId: 'status',
    renderHeaderCell: () => '状态',
    renderCell: item => (
      <Badge appearance="outline" color={item.status === '运行中' ? 'brand' : item.status === '已完成' ? 'success' : 'warning'}>
        {item.status}
      </Badge>
    ),
  }),
  createTableColumn<RowItem>({
    columnId: 'owner',
    renderHeaderCell: () => '负责人',
    renderCell: item => item.owner,
  }),
];

export function WorkspaceGridDrawerPage() {
  const styles = useStyles();
  const [query, setQuery] = React.useState('');
  const [selected, setSelected] = React.useState<RowItem | null>(rows[0]);
  const filtered = React.useMemo(
    () => rows.filter(item => item.name.includes(query) || item.owner.includes(query)),
    [query],
  );

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Title3>任务治理页</Title3>
        <Caption1>主区使用 DataGrid，右侧使用 inline Drawer 承接详情与编辑。</Caption1>
        <Button appearance="primary">新建任务</Button>
      </header>

      <section className={styles.filters}>
        <Field label="搜索任务" hint="按名称或负责人筛选">
          <Input value={query} onChange={(_, data) => setQuery(data.value)} />
        </Field>
      </section>

      <Toolbar aria-label="任务工具栏">
        <ToolbarButton>刷新</ToolbarButton>
        <ToolbarButton>导出</ToolbarButton>
        <ToolbarButton>批量处理</ToolbarButton>
      </Toolbar>

      <div className={styles.layout}>
        <DataGrid items={filtered} columns={columns} getRowId={item => item.id}>
          <DataGridHeader>
            <DataGridRow>
              {({ renderHeaderCell }) => (
                <DataGridHeaderCell>{renderHeaderCell()}</DataGridHeaderCell>
              )}
            </DataGridRow>
          </DataGridHeader>
          <DataGridBody<RowItem>>
            {({ item, rowId }) => (
              <DataGridRow<RowItem> key={rowId} onClick={() => setSelected(item)}>
                {({ renderCell }) => <DataGridCell>{renderCell(item)}</DataGridCell>}
              </DataGridRow>
            )}
          </DataGridBody>
        </DataGrid>

        <Drawer className={styles.drawerSurface} open={Boolean(selected)} position="end" type="inline" separator>
          <DrawerHeader>
            <DrawerHeaderTitle>{selected?.name ?? '详情'}</DrawerHeaderTitle>
          </DrawerHeader>
          <DrawerBody>
            {selected ? (
              <div style={{ display: 'grid', gap: 12 }}>
                <MessageBar>这里放详情摘要、编辑区和操作记录。</MessageBar>
                <Body1Strong>负责人：{selected.owner}</Body1Strong>
                <Caption1>当前状态：{selected.status}</Caption1>
                <Button appearance="primary">编辑任务</Button>
                <Button onClick={() => setSelected(null)}>关闭详情</Button>
              </div>
            ) : null}
          </DrawerBody>
        </Drawer>
      </div>
    </div>
  );
}
