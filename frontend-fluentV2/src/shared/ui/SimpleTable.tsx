import type { ReactNode } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableCellLayout,
  TableHeader,
  TableHeaderCell,
  TableRow,
  makeStyles,
  mergeClasses,
  tokens,
} from '@fluentui/react-components'

const useStyles = makeStyles({
  tableWrap: {
    overflowX: 'auto',
    borderRadius: tokens.borderRadiusXLarge,
    border: `1px solid ${tokens.colorNeutralStroke2}`,
    backgroundColor: tokens.colorNeutralBackground1,
  },
  cell: {
    minWidth: '120px',
  },
  row: {
    cursor: 'default',
  },
  interactiveRow: {
    cursor: 'pointer',
  },
  selectedRow: {
    backgroundColor: tokens.colorNeutralBackground3,
  },
})

export interface SimpleTableColumn<T> {
  key: string
  header: string
  render: (record: T) => ReactNode
}

export function SimpleTable<T>({
  columns,
  items,
  rowKey,
  onRowClick,
  selectedRowKey,
}: {
  columns: SimpleTableColumn<T>[]
  items: T[]
  rowKey: (record: T) => string
  onRowClick?: (record: T) => void
  selectedRowKey?: string | null
}) {
  const styles = useStyles()
  return (
    <div className={styles.tableWrap}>
      <Table aria-label="data-table">
        <TableHeader>
          <TableRow>
            {columns.map((column) => (
              <TableHeaderCell key={column.key}>{column.header}</TableHeaderCell>
            ))}
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((record) => (
            <TableRow
              key={rowKey(record)}
              className={mergeClasses(
                styles.row,
                onRowClick ? styles.interactiveRow : undefined,
                selectedRowKey && selectedRowKey === rowKey(record) ? styles.selectedRow : undefined,
              )}
              onClick={onRowClick ? () => onRowClick(record) : undefined}
            >
              {columns.map((column) => (
                <TableCell key={column.key} className={styles.cell}>
                  <TableCellLayout>{column.render(record)}</TableCellLayout>
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
