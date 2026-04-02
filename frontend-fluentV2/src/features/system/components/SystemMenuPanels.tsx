import {
  Badge,
  Body1,
  Body1Strong,
  Button,
  Caption1,
  Dialog,
  DialogActions,
  DialogBody,
  DialogContent,
  DialogSurface,
  DialogTitle,
  Divider,
  Field,
  Input,
  Select,
  Spinner,
  Switch,
  Textarea,
  makeStyles,
  tokens,
} from '@fluentui/react-components'
import { SectionCard } from '@/shared/ui/SectionCard'
import { PageStatusBanner } from '@/shared/ui/PageStatusBanner'
import { MenuReadonlyField } from '@/features/system/components/MenuTreeNode'
import type { MenuDeletePreview, MenuManageGroup, MenuMutationDraft, MenuNode, MenuNodeDetail, MenuNodeKind, MenuPageBinding } from '@/shared/types/menu'

type EditorMode = 'view' | 'create-root' | 'create-sibling' | 'create-child' | 'edit'

const useStyles = makeStyles({
  stack: {
    display: 'grid',
    gap: '16px',
  },
  summaryHeader: {
    display: 'grid',
    gap: '10px',
  },
  summaryLine: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
    alignItems: 'center',
  },
  fieldGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 760px)': {
      gridTemplateColumns: '1fr',
    },
  },
  listBlock: {
    display: 'grid',
    gap: '8px',
  },
  listRow: {
    display: 'grid',
    gap: '2px',
    padding: '10px 12px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
  },
  emptyBlock: {
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorNeutralBackground2,
    color: tokens.colorNeutralForeground3,
  },
  formGrid: {
    display: 'grid',
    gap: '12px',
  },
  formActions: {
    display: 'flex',
    gap: '12px',
    flexWrap: 'wrap',
    justifyContent: 'flex-end',
  },
  dangerBox: {
    display: 'grid',
    gap: '12px',
    padding: '14px',
    borderRadius: tokens.borderRadiusLarge,
    backgroundColor: tokens.colorPaletteRedBackground1,
    border: `1px solid ${tokens.colorPaletteRedBorder2}`,
  },
  dialogSummary: {
    display: 'grid',
    gap: '8px',
  },
  dialogGrid: {
    display: 'grid',
    gap: '8px',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    '@media (max-width: 720px)': {
      gridTemplateColumns: '1fr',
    },
  },
})

export function SystemMenuPanels({
  activeDetail,
  detailLoading,
  detailErrorMessage,
  currentPages,
  pagesLoading,
  pagesErrorMessage,
  draft,
  submitError,
  createErrorMessage,
  updateErrorMessage,
  isDirty,
  editorMode,
  manageGroups,
  parentOptions,
  deleteDialogOpen,
  deleteMode,
  deleteTargetParentId,
  deletePreview,
  deletePreviewLoading,
  deletePreviewErrorMessage,
  deletePending,
  deleteErrorMessage,
  onOpenCreateSibling,
  onOpenCreateChild,
  onOpenEdit,
  onResetEditor,
  onKindChange,
  onPatchDraft,
  onSubmit,
  onDeleteDialogOpenChange,
  onDeleteModeChange,
  onDeleteTargetParentIdChange,
  onDeleteDialogOpen,
  onDeleteConfirm,
}: {
  activeDetail: MenuNodeDetail | null
  detailLoading: boolean
  detailErrorMessage?: string
  currentPages: MenuPageBinding[]
  pagesLoading: boolean
  pagesErrorMessage?: string
  draft: MenuMutationDraft | null
  submitError: string
  createErrorMessage?: string
  updateErrorMessage?: string
  isDirty: boolean
  editorMode: EditorMode
  manageGroups: MenuManageGroup[]
  parentOptions: MenuNode[]
  deleteDialogOpen: boolean
  deleteMode: MenuDeletePreview['mode']
  deleteTargetParentId: string
  deletePreview: MenuDeletePreview | null
  deletePreviewLoading: boolean
  deletePreviewErrorMessage?: string
  deletePending: boolean
  deleteErrorMessage?: string
  onOpenCreateSibling: () => void
  onOpenCreateChild: () => void
  onOpenEdit: () => void
  onResetEditor: () => void
  onKindChange: (nextKind: MenuNodeKind) => void
  onPatchDraft: (patch: Partial<MenuMutationDraft>) => void
  onSubmit: () => void
  onDeleteDialogOpenChange: (open: boolean) => void
  onDeleteModeChange: (mode: MenuDeletePreview['mode']) => void
  onDeleteTargetParentIdChange: (id: string) => void
  onDeleteDialogOpen: () => void
  onDeleteConfirm: () => void
}) {
  const styles = useStyles()

  return (
    <div className={styles.stack}>
      <SectionCard
        title="当前节点摘要"
        description="右侧保持治理工作台形态：摘要、详情、编辑表单、页面绑定和危险操作分区。"
        actions={
          activeDetail ? (
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              <Button appearance="secondary" onClick={onOpenCreateSibling}>
                新建同级
              </Button>
              <Button appearance="secondary" onClick={onOpenCreateChild}>
                新建子级
              </Button>
              <Button appearance="primary" onClick={onOpenEdit}>
                编辑当前节点
              </Button>
            </div>
          ) : null
        }
      >
        {detailLoading ? <Spinner label="正在加载节点详情" /> : null}
        {detailErrorMessage ? <PageStatusBanner intent="error" title="节点详情加载失败" description={detailErrorMessage} /> : null}
        {!activeDetail ? (
          <div className={styles.emptyBlock}>请选择左侧节点，或直接创建顶级菜单。</div>
        ) : (
          <div className={styles.summaryHeader}>
            <Body1Strong>{activeDetail.title || activeDetail.name}</Body1Strong>
            <div className={styles.summaryLine}>
              <Badge appearance="tint">{activeDetail.kind}</Badge>
              <Badge appearance="outline">{activeDetail.spaceKey}</Badge>
              {activeDetail.hidden ? (
                <Badge color="important" appearance="filled">
                  已隐藏
                </Badge>
              ) : null}
              {activeDetail.manageGroup?.name ? <Badge appearance="outline">{activeDetail.manageGroup.name}</Badge> : null}
            </div>
            <div className={styles.fieldGrid}>
              <MenuReadonlyField label="路径" value={activeDetail.path || '-'} />
              <MenuReadonlyField label="组件" value={activeDetail.component || '-'} />
              <MenuReadonlyField label="父级菜单" value={activeDetail.parent ? `${activeDetail.parent.title}` : '顶级菜单'} />
              <MenuReadonlyField label="子节点数量" value={`${activeDetail.childCount}`} />
            </div>
          </div>
        )}
      </SectionCard>

      <SectionCard title="基本信息与结构" description="只读区继续保留稳定详情，保存后树与详情保持一致。">
        {!activeDetail ? (
          <div className={styles.emptyBlock}>当前没有可展示的节点详情。</div>
        ) : (
          <>
            <div className={styles.fieldGrid}>
              <MenuReadonlyField label="菜单名称" value={activeDetail.name} />
              <MenuReadonlyField label="标题" value={activeDetail.title} />
              <MenuReadonlyField label="图标" value={activeDetail.icon || '-'} />
              <MenuReadonlyField label="排序" value={`${activeDetail.sortOrder}`} />
              <MenuReadonlyField label="访问模式" value={`${activeDetail.meta.accessMode || 'permission'}`} />
              <MenuReadonlyField label="外链地址" value={`${activeDetail.meta.link || '-'}`} />
              <MenuReadonlyField label="激活路径" value={`${activeDetail.meta.activePath || '-'}`} />
              <MenuReadonlyField label="权限键" value={activeDetail.permissionKeys.join('，') || '-'} />
            </div>
            <div className={styles.listBlock}>
              <Body1>Meta 核心字段</Body1>
              {Object.entries(activeDetail.meta).length ? (
                Object.entries(activeDetail.meta).map(([key, value]) => (
                  <div key={key} className={styles.listRow}>
                    <Body1>{key}</Body1>
                    <Caption1>{typeof value === 'string' ? value : JSON.stringify(value)}</Caption1>
                  </div>
                ))
              ) : (
                <div className={styles.emptyBlock}>当前节点没有额外 meta 字段。</div>
              )}
            </div>
          </>
        )}
      </SectionCard>

      <SectionCard title="编辑表单" description="第 8 版继续沿用受控编辑版，并把类型感知字段收进统一治理表单。">
        {submitError ? <PageStatusBanner intent="error" title="保存菜单失败" description={submitError} /> : null}
        {createErrorMessage ? <PageStatusBanner intent="error" title="创建菜单失败" description={createErrorMessage} /> : null}
        {updateErrorMessage ? <PageStatusBanner intent="error" title="保存菜单失败" description={updateErrorMessage} /> : null}
        {!draft ? (
          <div className={styles.emptyBlock}>请选择“编辑当前节点”，或创建顶级/同级/子级菜单后再填写表单。</div>
        ) : (
          <div className={styles.formGrid}>
            {isDirty ? <PageStatusBanner intent="warning" title="存在未保存修改" description="当前表单存在未保存修改，切换节点前会提示确认。" /> : null}
            <div className={styles.fieldGrid}>
              <Field label="菜单类型" required>
                <Select value={draft.kind} onChange={(event) => onKindChange(event.target.value as MenuNodeKind)}>
                  <option value="directory">directory</option>
                  <option value="entry">entry</option>
                  <option value="external">external</option>
                </Select>
              </Field>
              <Field label="所属空间">
                <Input value={draft.spaceKey} onChange={(_, data) => onPatchDraft({ spaceKey: data.value })} />
              </Field>
              <Field label="名称" required>
                <Input value={draft.name} onChange={(_, data) => onPatchDraft({ name: data.value })} />
              </Field>
              <Field label="标题" required>
                <Input value={draft.title} onChange={(_, data) => onPatchDraft({ title: data.value })} />
              </Field>
              <Field label="路径" required={draft.kind !== 'directory'}>
                <Input value={draft.path} onChange={(_, data) => onPatchDraft({ path: data.value })} />
              </Field>
              <Field label="组件" required={draft.kind === 'entry'}>
                <Input disabled={draft.kind !== 'entry'} value={draft.component} onChange={(_, data) => onPatchDraft({ component: data.value })} />
              </Field>
              <Field label="图标">
                <Input value={draft.icon} onChange={(_, data) => onPatchDraft({ icon: data.value })} />
              </Field>
              <Field label="排序">
                <Input type="number" value={`${draft.sortOrder}`} onChange={(_, data) => onPatchDraft({ sortOrder: Number(data.value || 0) })} />
              </Field>
              <Field label="父级节点">
                <Select value={draft.parentId || ''} onChange={(event) => onPatchDraft({ parentId: event.target.value || null })}>
                  <option value="">顶级菜单</option>
                  {parentOptions.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.title || item.name}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="管理分组">
                <Select value={draft.manageGroupId || ''} onChange={(event) => onPatchDraft({ manageGroupId: event.target.value || null })}>
                  <option value="">未分组</option>
                  {manageGroups.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="访问模式">
                <Select value={draft.accessMode} onChange={(event) => onPatchDraft({ accessMode: event.target.value })}>
                  <option value="permission">permission</option>
                  <option value="jwt">jwt</option>
                  <option value="public">public</option>
                </Select>
              </Field>
              <Field label="隐藏状态">
                <Switch checked={draft.hidden} label={draft.hidden ? '隐藏' : '显示'} onChange={(_, data) => onPatchDraft({ hidden: Boolean(data.checked) })} />
              </Field>
              <Field label="激活路径">
                <Input disabled={draft.kind !== 'entry'} value={draft.activePath} onChange={(_, data) => onPatchDraft({ activePath: data.value })} />
              </Field>
              <Field label="外链地址">
                <Input disabled={draft.kind !== 'external'} value={draft.externalLink} onChange={(_, data) => onPatchDraft({ externalLink: data.value })} />
              </Field>
            </div>
            <div className={styles.fieldGrid}>
              <Field label="页面缓存">
                <Switch checked={draft.keepAlive} disabled={draft.kind !== 'entry'} label={draft.keepAlive ? '开启' : '关闭'} onChange={(_, data) => onPatchDraft({ keepAlive: Boolean(data.checked) })} />
              </Field>
              <Field label="固定标签">
                <Switch checked={draft.fixedTab} disabled={draft.kind !== 'entry'} label={draft.fixedTab ? '开启' : '关闭'} onChange={(_, data) => onPatchDraft({ fixedTab: Boolean(data.checked) })} />
              </Field>
              <Field label="全屏页面">
                <Switch checked={draft.isFullPage} disabled={draft.kind !== 'entry'} label={draft.isFullPage ? '开启' : '关闭'} onChange={(_, data) => onPatchDraft({ isFullPage: Boolean(data.checked) })} />
              </Field>
            </div>
            <Field label="表单说明">
              <Textarea
                resize="vertical"
                value={
                  draft.kind === 'directory'
                    ? '目录类型用于导航分组，不强制绑定业务组件。'
                    : draft.kind === 'entry'
                      ? '入口类型要求稳定路径和组件字段，并可维护 activePath / keepAlive / fixedTab。'
                      : '外链类型会将跳转地址写入 meta.link，组件字段不会提交。'
                }
                readOnly
              />
            </Field>
            <div className={styles.formActions}>
              <Button appearance="secondary" onClick={onResetEditor}>
                取消
              </Button>
              <Button appearance="primary" onClick={onSubmit}>
                {editorMode === 'edit' ? '保存修改' : '创建菜单'}
              </Button>
            </div>
          </div>
        )}
      </SectionCard>

      <SectionCard title="关联页面信息" description="继续保持只读，不把页面治理与菜单治理强行混成一个大表单。">
        {pagesLoading ? <Spinner label="正在加载关联页面" /> : null}
        {pagesErrorMessage ? <PageStatusBanner intent="error" title="关联页面加载失败" description={pagesErrorMessage} /> : null}
        {currentPages.length ? (
          currentPages.map((item) => (
            <div key={item.pageKey} className={styles.listRow}>
              <Body1>{item.name || item.pageKey}</Body1>
              <Caption1>Page Key：{item.pageKey}</Caption1>
              <Caption1>路由：{item.routePath}</Caption1>
              <Caption1>组件：{item.component || '-'}</Caption1>
              <Caption1>访问模式：{item.accessMode || '-'}</Caption1>
              <Caption1>权限键：{item.permissionKey || '-'}</Caption1>
            </div>
          ))
        ) : (
          <div className={styles.emptyBlock}>当前节点没有关联的受管页面。</div>
        )}
      </SectionCard>

      <SectionCard title="危险操作" description="删除预检与删除确认独立隔离，不与常规编辑表单混在一起。">
        {!activeDetail ? (
          <div className={styles.emptyBlock}>选择节点后可查看删除影响并执行删除。</div>
        ) : (
          <div className={styles.dangerBox}>
            <Body1Strong>删除当前节点</Body1Strong>
            <Body1>删除前会获取真实删除预览；若后端返回的仅是最小影响范围，界面会按接口结果如实展示。</Body1>
            <div>
              <Button appearance="primary" onClick={onDeleteDialogOpen}>
                查看删除预览
              </Button>
            </div>
          </div>
        )}
      </SectionCard>

      <Dialog open={deleteDialogOpen} onOpenChange={(_, data) => onDeleteDialogOpenChange(data.open)}>
        <DialogSurface>
          <DialogBody>
            <DialogTitle>删除菜单确认</DialogTitle>
            <DialogContent>
              {!activeDetail ? (
                <div className={styles.emptyBlock}>当前没有可删除的菜单节点。</div>
              ) : (
                <div className={styles.dialogSummary}>
                  <Body1Strong>{activeDetail.title || activeDetail.name}</Body1Strong>
                  <Caption1>删除后会刷新菜单树、详情区和 URL 选中状态。</Caption1>
                  <div className={styles.dialogGrid}>
                    <Field label="删除方式">
                      <Select value={deleteMode} onChange={(event) => onDeleteModeChange(event.target.value as MenuDeletePreview['mode'])}>
                        <option value="single" disabled={activeDetail.childCount > 0}>single</option>
                        <option value="cascade">cascade</option>
                        <option value="promote_children" disabled={activeDetail.childCount === 0}>promote_children</option>
                      </Select>
                    </Field>
                    <Field label="子节点提升目标">
                      <Select disabled={deleteMode !== 'promote_children'} value={deleteTargetParentId} onChange={(event) => onDeleteTargetParentIdChange(event.target.value)}>
                        <option value="">顶级菜单</option>
                        {parentOptions.map((item) => (
                          <option key={item.id} value={item.id}>
                            {item.title || item.name}
                          </option>
                        ))}
                      </Select>
                    </Field>
                  </div>
                  <Divider />
                  {deletePreviewLoading ? <Spinner label="正在获取删除预览" /> : null}
                  {deletePreviewErrorMessage ? <PageStatusBanner intent="error" title="删除预览加载失败" description={deletePreviewErrorMessage} /> : null}
                  {deletePreview ? (
                    <div className={styles.dialogGrid}>
                      <MenuReadonlyField label="删除方式" value={deletePreview.mode} />
                      <MenuReadonlyField label="直接子节点" value={`${deletePreview.childCount}`} />
                      <MenuReadonlyField label="影响菜单数" value={`${deletePreview.menuCount}`} />
                      <MenuReadonlyField label="影响页面数" value={`${deletePreview.affectedPageCount}`} />
                      <MenuReadonlyField label="影响关系数" value={`${deletePreview.affectedRelationCount}`} />
                      <MenuReadonlyField label="预检说明" value={deletePreview.mode === 'single' ? '当前预检只删除当前节点' : '按后端删除模式返回影响范围'} />
                    </div>
                  ) : null}
                  {deleteErrorMessage ? <PageStatusBanner intent="error" title="删除失败" description={deleteErrorMessage} /> : null}
                </div>
              )}
            </DialogContent>
            <DialogActions>
              <Button appearance="secondary" onClick={() => onDeleteDialogOpenChange(false)}>
                取消
              </Button>
              <Button appearance="primary" disabled={!activeDetail || deletePreviewLoading || deletePending} onClick={onDeleteConfirm}>
                确认删除
              </Button>
            </DialogActions>
          </DialogBody>
        </DialogSurface>
      </Dialog>
    </div>
  )
}
