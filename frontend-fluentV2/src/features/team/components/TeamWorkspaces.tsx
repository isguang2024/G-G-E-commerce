import { useEffect, useMemo, useState } from 'react'
import { Badge, Button, Checkbox, Field, Input, Select, Spinner, Textarea, makeStyles } from '@fluentui/react-components'
import { Add20Regular, Delete20Regular, Save20Regular } from '@fluentui/react-icons'
import { useSearchParams } from 'react-router-dom'
import { useUserListQuery } from '@/features/access/access.service'
import { useAuthStore } from '@/features/auth/auth.store'
import { PageContainer } from '@/features/shell/components/PageContainer'
import {
  useAddMyTeamMemberMutation,
  useAddTeamMemberMutation,
  useCreateTeamMutation,
  useDeleteTeamMutation,
  useMyTeamBoundaryRolesQuery,
  useMyTeamMemberRolesQuery,
  useMyTeamMembersQuery,
  useRemoveMyTeamMemberMutation,
  useRemoveTeamMemberMutation,
  useSetMyTeamMemberRolesMutation,
  useTeamDetailQuery,
  useTeamListQuery,
  useTeamMembersQuery,
  useUpdateMyTeamMemberRoleMutation,
  useUpdateTeamMemberRoleMutation,
  useUpdateTeamMutation,
} from '@/features/team/team.service'
import { EmptyState, ErrorState, LoadingState } from '@/shared/ui/AsyncState'
import { LinkCardGrid } from '@/shared/ui/LinkCardGrid'
import { PageStatusBanner } from '@/shared/ui/PageStatusBanner'
import { PropertyGrid } from '@/shared/ui/PropertyGrid'
import { SectionCard } from '@/shared/ui/SectionCard'
import { SimpleTable } from '@/shared/ui/SimpleTable'
import { TwoPaneWorkbench, WorkbenchStack } from '@/shared/ui/WorkbenchLayouts'
import type { TeamSavePayload } from '@/shared/types/admin'

const useStyles = makeStyles({
  form: {
    display: 'grid',
    gap: '12px',
  },
  formGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, minmax(0, 1fr))',
    gap: '12px',
    '@media (max-width: 960px)': {
      gridTemplateColumns: '1fr',
    },
  },
  checkboxList: {
    display: 'grid',
    gap: '8px',
    maxHeight: '240px',
    overflowY: 'auto',
  },
  actionRow: {
    display: 'flex',
    gap: '10px',
    flexWrap: 'wrap',
  },
  badgeRow: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
  },
})

function updateSearchParams(
  searchParams: URLSearchParams,
  setSearchParams: ReturnType<typeof useSearchParams>[1],
  patch: Record<string, string | null | undefined>,
) {
  const next = new URLSearchParams(searchParams)
  Object.entries(patch).forEach(([key, value]) => {
    if (!value) next.delete(key)
    else next.set(key, value)
  })
  setSearchParams(next, { replace: true })
}

function createTeamDraft(): TeamSavePayload {
  return {
    name: '',
    remark: '',
    plan: 'free',
    maxMembers: 10,
    status: 'active',
  }
}

function TeamPageFeedback({
  feedback,
}: {
  feedback: { intent: 'success' | 'error'; title: string; description: string } | null
}) {
  return feedback ? <PageStatusBanner intent={feedback.intent} title={feedback.title} description={feedback.description} /> : null
}

export function TeamTeamWorkspace({ routeId }: { routeId: string }) {
  const styles = useStyles()
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedTeamId = searchParams.get('selectedTeamId') || ''
  const isCreating = searchParams.get('mode') === 'new'
  const listQuery = useTeamListQuery({ current: 1, size: 40 })
  const detailQuery = useTeamDetailQuery(isCreating ? null : selectedTeamId)
  const membersQuery = useTeamMembersQuery(isCreating ? null : selectedTeamId)
  const userListQuery = useUserListQuery({ current: 1, size: 80 })
  const createMutation = useCreateTeamMutation()
  const deleteMutation = useDeleteTeamMutation()
  const updateMutation = useUpdateTeamMutation(selectedTeamId || '')
  const addMemberMutation = useAddTeamMemberMutation(selectedTeamId || '')
  const removeMemberMutation = useRemoveTeamMemberMutation(selectedTeamId || '')
  const updateMemberRoleMutation = useUpdateTeamMemberRoleMutation(selectedTeamId || '')
  const [draft, setDraft] = useState<TeamSavePayload>(createTeamDraft())
  const [newMemberUserId, setNewMemberUserId] = useState('')
  const [newMemberRoleCode, setNewMemberRoleCode] = useState('team_member')
  const [feedback, setFeedback] = useState<{ intent: 'success' | 'error'; title: string; description: string } | null>(null)

  useEffect(() => {
    if (isCreating) {
      setDraft(createTeamDraft())
      return
    }
    if (detailQuery.data) {
      setDraft({
        name: detailQuery.data.name,
        remark: detailQuery.data.remark,
        plan: detailQuery.data.plan,
        maxMembers: detailQuery.data.maxMembers,
        status: detailQuery.data.status,
      })
    }
  }, [detailQuery.data, isCreating])

  useEffect(() => {
    if (!selectedTeamId || isCreating || !listQuery.data) return
    const stillExists = listQuery.data.records.some((item) => item.id === selectedTeamId)
    if (!stillExists) {
      updateSearchParams(searchParams, setSearchParams, { selectedTeamId: '' })
    }
  }, [isCreating, listQuery.data, searchParams, selectedTeamId, setSearchParams])

  async function handleSave() {
    try {
      if (isCreating) {
        const created = await createMutation.mutateAsync(draft)
        updateSearchParams(searchParams, setSearchParams, { selectedTeamId: created.id, mode: '' })
        setFeedback({
          intent: 'success',
          title: '团队已创建',
          description: `团队「${created.name}」已创建并切换为当前上下文。`,
        })
        return
      }
      await updateMutation.mutateAsync(draft)
      setFeedback({
        intent: 'success',
        title: '团队已保存',
        description: `团队「${draft.name || '未命名团队'}」的基本信息已更新。`,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '团队保存失败'
      setFeedback({ intent: 'error', title: '团队保存失败', description: message })
    }
  }

  async function handleDelete() {
    if (!selectedTeamId || !window.confirm('确认删除当前团队吗？')) return

    try {
      const teamName = detailQuery.data?.name || draft.name || '当前团队'
      await deleteMutation.mutateAsync(selectedTeamId)
      updateSearchParams(searchParams, setSearchParams, { selectedTeamId: '', mode: '' })
      setFeedback({
        intent: 'success',
        title: '团队已删除',
        description: `团队「${teamName}」已删除，列表和详情区已同步重置。`,
      })
    } catch (error) {
      const message = error instanceof Error ? error.message : '团队删除失败'
      setFeedback({ intent: 'error', title: '团队删除失败', description: message })
    }
  }

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <>
          <Button appearance="secondary" icon={<Add20Regular />} onClick={() => updateSearchParams(searchParams, setSearchParams, { mode: 'new', selectedTeamId: '' })}>
            新建团队
          </Button>
          {selectedTeamId && !isCreating ? (
            <Button appearance="secondary" icon={<Delete20Regular />} disabled={deleteMutation.isPending} onClick={() => void handleDelete()}>
              删除团队
            </Button>
          ) : null}
          <Button appearance="primary" icon={<Save20Regular />} disabled={createMutation.isPending || updateMutation.isPending} onClick={() => void handleSave()}>
            {createMutation.isPending || updateMutation.isPending ? <Spinner size="tiny" /> : '保存团队'}
          </Button>
        </>
      }
    >
      <TeamPageFeedback feedback={feedback} />
      <TwoPaneWorkbench
        primary={
          <SectionCard title="团队列表" description="团队治理页承接团队本身、管理员摘要和成员清单。">
            {listQuery.isLoading ? <LoadingState label="正在加载团队列表" /> : null}
            {listQuery.isError ? <ErrorState description={listQuery.error.message} /> : null}
            {listQuery.data ? (
              <SimpleTable
                columns={[
                  { key: 'name', header: '团队', render: (item) => item.name },
                  { key: 'plan', header: '套餐', render: (item) => item.plan },
                  { key: 'maxMembers', header: '成员上限', render: (item) => `${item.maxMembers}` },
                ]}
                items={listQuery.data.records}
                onRowClick={(item) => updateSearchParams(searchParams, setSearchParams, { selectedTeamId: item.id, mode: '' })}
                rowKey={(item) => item.id}
                selectedRowKey={isCreating ? '' : selectedTeamId}
              />
            ) : null}
          </SectionCard>
        }
        secondary={
          <WorkbenchStack>
            <SectionCard title={isCreating ? '新建团队' : detailQuery.data ? `编辑团队 · ${detailQuery.data.name}` : '团队详情'} description="团队基本信息、管理员摘要和成员工作区会在右侧持续联动。">
              {!isCreating && !selectedTeamId ? <EmptyState title="选择左侧团队或新建一条记录" /> : null}
              {selectedTeamId && !isCreating && detailQuery.isLoading ? <LoadingState label="正在读取团队详情" /> : null}
              {selectedTeamId && !isCreating && detailQuery.isError ? <ErrorState description={detailQuery.error.message} /> : null}
              {isCreating || detailQuery.data ? (
                <div className={styles.form}>
                  {detailQuery.data ? (
                    <PropertyGrid
                      items={[
                        { label: '套餐', value: detailQuery.data.plan || '-' },
                        { label: '成员上限', value: `${detailQuery.data.maxMembers}` },
                        { label: '状态', value: detailQuery.data.status || '-' },
                        { label: '更新时间', value: detailQuery.data.updatedAt || '-' },
                      ]}
                    />
                  ) : null}
                  <div className={styles.formGrid}>
                    <Field label="团队名称">
                      <Input value={draft.name} onChange={(_, data) => setDraft((prev) => ({ ...prev, name: data.value }))} />
                    </Field>
                    <Field label="套餐">
                      <Input value={draft.plan} onChange={(_, data) => setDraft((prev) => ({ ...prev, plan: data.value }))} />
                    </Field>
                    <Field label="成员上限">
                      <Input type="number" value={`${draft.maxMembers}`} onChange={(_, data) => setDraft((prev) => ({ ...prev, maxMembers: Number(data.value || 0) }))} />
                    </Field>
                    <Field label="状态">
                      <Input value={draft.status} onChange={(_, data) => setDraft((prev) => ({ ...prev, status: data.value }))} />
                    </Field>
                  </div>
                  <Field label="备注">
                    <Textarea value={draft.remark} onChange={(_, data) => setDraft((prev) => ({ ...prev, remark: data.value }))} />
                  </Field>
                </div>
              ) : null}
            </SectionCard>

            {detailQuery.data?.adminUsers?.length ? (
              <SectionCard title="团队管理员摘要" description="管理员身份和团队基础上下文保持在同一工作面。">
                <div className={styles.badgeRow}>
                  {detailQuery.data.adminUsers.map((item) => (
                    <Badge key={item.userId} appearance="tint">
                      {item.nickName || item.userName}
                    </Badge>
                  ))}
                </div>
              </SectionCard>
            ) : null}

            {selectedTeamId && !isCreating ? (
              <SectionCard title="成员管理" description="支持直接添加成员并调整基础团队角色。">
                <div className={styles.actionRow}>
                  <Select value={newMemberUserId} onChange={(event) => setNewMemberUserId(event.target.value)}>
                    <option value="">选择用户</option>
                    {userListQuery.data?.records.map((item) => <option key={item.id} value={item.id}>{item.userName}</option>)}
                  </Select>
                  <Select value={newMemberRoleCode} onChange={(event) => setNewMemberRoleCode(event.target.value)}>
                    <option value="team_member">team_member</option>
                    <option value="team_admin">team_admin</option>
                  </Select>
                  <Button
                    appearance="primary"
                    disabled={!newMemberUserId}
                    onClick={() => {
                      void (async () => {
                        try {
                          await addMemberMutation.mutateAsync({ userId: newMemberUserId, roleCode: newMemberRoleCode })
                          setNewMemberUserId('')
                          setFeedback({
                            intent: 'success',
                            title: '成员已加入团队',
                            description: '团队成员列表已刷新，可以继续调整角色。',
                          })
                        } catch (error) {
                          const message = error instanceof Error ? error.message : '添加成员失败'
                          setFeedback({ intent: 'error', title: '添加成员失败', description: message })
                        }
                      })()
                    }}
                  >
                    添加成员
                  </Button>
                </div>
                {membersQuery.isLoading ? <LoadingState label="正在加载团队成员" /> : null}
                {membersQuery.data?.length ? (
                  <SimpleTable
                    columns={[
                      { key: 'user', header: '成员', render: (item) => item.displayName },
                      { key: 'email', header: '邮箱', render: (item) => item.userEmail || '-' },
                      {
                        key: 'role',
                        header: '角色',
                        render: (item) => (
                          <Select
                            value={item.roleCode}
                            onChange={(event) => {
                              void (async () => {
                                try {
                                  await updateMemberRoleMutation.mutateAsync({ userId: item.userId, roleCode: event.target.value })
                                  setFeedback({
                                    intent: 'success',
                                    title: '成员角色已更新',
                                    description: `成员「${item.displayName}」的团队角色已更新。`,
                                  })
                                } catch (error) {
                                  const message = error instanceof Error ? error.message : '更新成员角色失败'
                                  setFeedback({ intent: 'error', title: '更新成员角色失败', description: message })
                                }
                              })()
                            }}
                          >
                            <option value="team_member">team_member</option>
                            <option value="team_admin">team_admin</option>
                          </Select>
                        ),
                      },
                      {
                        key: 'action',
                        header: '操作',
                        render: (item) => (
                          <Button
                            appearance="subtle"
                            onClick={() => {
                              void (async () => {
                                try {
                                  await removeMemberMutation.mutateAsync(item.userId)
                                  setFeedback({
                                    intent: 'success',
                                    title: '成员已移除',
                                    description: `成员「${item.displayName}」已从当前团队移除。`,
                                  })
                                } catch (error) {
                                  const message = error instanceof Error ? error.message : '移除成员失败'
                                  setFeedback({ intent: 'error', title: '移除成员失败', description: message })
                                }
                              })()
                            }}
                          >
                            移除
                          </Button>
                        ),
                      },
                    ]}
                    items={membersQuery.data}
                    rowKey={(item) => item.id || item.userId}
                  />
                ) : (
                  <EmptyState title="当前团队还没有成员" />
                )}
              </SectionCard>
            ) : null}
          </WorkbenchStack>
        }
      />
    </PageContainer>
  )
}

export function TeamMembersWorkspace({ routeId }: { routeId: string }) {
  const styles = useStyles()
  const currentTenantId = useAuthStore((state) => state.tenantContext.currentTenantId)
  const [searchParams, setSearchParams] = useSearchParams()
  const selectedUserId = searchParams.get('selectedUserId') || ''
  const membersQuery = useMyTeamMembersQuery()
  const boundaryRolesQuery = useMyTeamBoundaryRolesQuery()
  const currentMemberRolesQuery = useMyTeamMemberRolesQuery(selectedUserId || undefined)
  const userListQuery = useUserListQuery({ current: 1, size: 80 })
  const addMemberMutation = useAddMyTeamMemberMutation()
  const removeMemberMutation = useRemoveMyTeamMemberMutation()
  const updateMemberRoleMutation = useUpdateMyTeamMemberRoleMutation()
  const setBoundaryRolesMutation = useSetMyTeamMemberRolesMutation(selectedUserId || '')
  const [newMemberUserId, setNewMemberUserId] = useState('')
  const [newMemberRoleCode, setNewMemberRoleCode] = useState('team_member')
  const [boundaryRoleIds, setBoundaryRoleIds] = useState<string[]>([])
  const [feedback, setFeedback] = useState<{ intent: 'success' | 'error'; title: string; description: string } | null>(null)

  const selectedMember = useMemo(
    () => membersQuery.data?.find((item) => item.userId === selectedUserId) || null,
    [membersQuery.data, selectedUserId],
  )

  useEffect(() => {
    setBoundaryRoleIds(currentMemberRolesQuery.data || [])
  }, [currentMemberRolesQuery.data])

  useEffect(() => {
    if (!selectedUserId || !membersQuery.data) return
    const stillExists = membersQuery.data.some((item) => item.userId === selectedUserId)
    if (!stillExists) {
      updateSearchParams(searchParams, setSearchParams, { selectedUserId: '' })
    }
  }, [membersQuery.data, searchParams, selectedUserId, setSearchParams])

  function toggleId(id: string, checked: boolean) {
    setBoundaryRoleIds((prev) => (checked ? [...new Set([...prev, id])] : prev.filter((item) => item !== id)))
  }

  if (membersQuery.isError) {
    return (
      <PageContainer routeId={routeId}>
        <EmptyState title="当前账号暂无团队上下文" description={membersQuery.error.message || '团队成员页需要先加入团队。'} />
      </PageContainer>
    )
  }

  if (!currentTenantId && !membersQuery.isLoading && !membersQuery.data?.length) {
    return (
      <PageContainer routeId={routeId}>
        <WorkbenchStack>
          <SectionCard title="当前账号尚未加入团队" description="团队成员治理已经接入真实链路，但需要先拥有团队上下文。">
            <EmptyState title="先创建团队或加入现有团队" description="创建团队后，这里会自动恢复成员列表、角色分配和边界角色配置。" />
          </SectionCard>
          <SectionCard title="下一步建议" description="优先补齐团队上下文，再回到成员治理页继续回归。">
            <LinkCardGrid
              items={[
                { id: 'team', title: '进入团队治理', description: '先创建团队或检查当前账号的团队归属。', to: '/team/team', actionLabel: '打开团队治理' },
                { id: 'team-more', title: '打开团队入口目录', description: '查看团队域中可用的其他治理入口。', to: '/team/more', actionLabel: '查看团队入口' },
              ]}
            />
          </SectionCard>
        </WorkbenchStack>
      </PageContainer>
    )
  }

  return (
    <PageContainer
      routeId={routeId}
      actions={
        <Button
          appearance="primary"
          disabled={!currentTenantId || !selectedUserId || setBoundaryRolesMutation.isPending}
          onClick={() => {
            void (async () => {
              try {
                await setBoundaryRolesMutation.mutateAsync(boundaryRoleIds)
                setFeedback({
                  intent: 'success',
                  title: '边界角色分配已保存',
                  description: selectedMember ? `成员「${selectedMember.displayName}」的边界角色已更新。` : '边界角色分配已更新。',
                })
              } catch (error) {
                const message = error instanceof Error ? error.message : '保存边界角色分配失败'
                setFeedback({ intent: 'error', title: '保存边界角色分配失败', description: message })
              }
            })()
          }}
        >
          保存边界角色分配
        </Button>
      }
    >
      <TeamPageFeedback feedback={feedback} />
      <TwoPaneWorkbench
        primary={
          <SectionCard title="我的团队成员" description="成员列表、基础角色和边界角色分配会在这一页统一处理。">
            <div className={styles.actionRow}>
              <Select value={newMemberUserId} onChange={(event) => setNewMemberUserId(event.target.value)}>
                <option value="">选择用户</option>
                {userListQuery.data?.records.map((item) => <option key={item.id} value={item.id}>{item.userName}</option>)}
              </Select>
              <Select value={newMemberRoleCode} onChange={(event) => setNewMemberRoleCode(event.target.value)}>
                <option value="team_member">team_member</option>
                <option value="team_admin">team_admin</option>
              </Select>
              <Button
                appearance="secondary"
                disabled={!currentTenantId || !newMemberUserId}
                onClick={() => {
                  void (async () => {
                    try {
                      await addMemberMutation.mutateAsync({ userId: newMemberUserId, roleCode: newMemberRoleCode })
                      setNewMemberUserId('')
                      setFeedback({
                        intent: 'success',
                        title: '成员已加入团队',
                        description: '团队成员列表已刷新，可以继续配置边界角色。',
                      })
                    } catch (error) {
                      const message = error instanceof Error ? error.message : '添加成员失败'
                      setFeedback({ intent: 'error', title: '添加成员失败', description: message })
                    }
                  })()
                }}
              >
                添加成员
              </Button>
            </div>
            {membersQuery.isLoading ? <LoadingState label="正在加载团队成员" /> : null}
            {membersQuery.data?.length ? (
              <SimpleTable
                columns={[
                  { key: 'user', header: '成员', render: (item) => item.displayName },
                  { key: 'email', header: '邮箱', render: (item) => item.userEmail || '-' },
                  {
                    key: 'role',
                    header: '团队角色',
                    render: (item) => (
                      <Select
                        value={item.roleCode}
                        onChange={(event) => {
                          void (async () => {
                            try {
                              await updateMemberRoleMutation.mutateAsync({ userId: item.userId, roleCode: event.target.value })
                              setFeedback({
                                intent: 'success',
                                title: '团队角色已更新',
                                description: `成员「${item.displayName}」的团队角色已更新。`,
                              })
                            } catch (error) {
                              const message = error instanceof Error ? error.message : '更新团队角色失败'
                              setFeedback({ intent: 'error', title: '更新团队角色失败', description: message })
                            }
                          })()
                        }}
                      >
                        <option value="team_member">team_member</option>
                        <option value="team_admin">team_admin</option>
                      </Select>
                    ),
                  },
                  {
                    key: 'action',
                    header: '操作',
                    render: (item) => (
                      <Button
                        appearance="subtle"
                        onClick={() => {
                          void (async () => {
                            try {
                              await removeMemberMutation.mutateAsync(item.userId)
                              if (item.userId === selectedUserId) {
                                updateSearchParams(searchParams, setSearchParams, { selectedUserId: '' })
                              }
                              setFeedback({
                                intent: 'success',
                                title: '成员已移除',
                                description: `成员「${item.displayName}」已从当前团队移除。`,
                              })
                            } catch (error) {
                              const message = error instanceof Error ? error.message : '移除成员失败'
                              setFeedback({ intent: 'error', title: '移除成员失败', description: message })
                            }
                          })()
                        }}
                      >
                        移除
                      </Button>
                    ),
                  },
                ]}
                items={membersQuery.data}
                onRowClick={(item) => updateSearchParams(searchParams, setSearchParams, { selectedUserId: item.userId })}
                rowKey={(item) => item.id || item.userId}
                selectedRowKey={selectedUserId}
              />
            ) : (
              <EmptyState title="当前团队还没有成员" />
            )}
          </SectionCard>
        }
        secondary={
          <WorkbenchStack>
            <SectionCard title={selectedMember ? `成员详情 · ${selectedMember.displayName}` : '成员详情'} description="成员属性、角色状态和边界角色分配会随着左侧选中项联动。">
              {!selectedMember ? <EmptyState title="选择左侧成员查看详情" /> : (
                <div className={styles.form}>
                  <PropertyGrid items={selectedMember.contactItems} />
                  <PropertyGrid items={selectedMember.roleItems} />
                </div>
              )}
            </SectionCard>

            {selectedMember ? (
              <SectionCard title="边界角色分配" description="勾选后保存即可把当前成员绑定到对应团队边界角色。">
                <div className={styles.checkboxList}>
                  {boundaryRolesQuery.data?.map((item) => (
                    <Checkbox key={item.id} checked={boundaryRoleIds.includes(item.id)} label={`${item.name} (${item.code})`} onChange={(_, data) => toggleId(item.id, Boolean(data.checked))} />
                  ))}
                </div>
              </SectionCard>
            ) : null}
          </WorkbenchStack>
        }
      />
    </PageContainer>
  )
}

export function TeamMoreWorkspace({ routeId }: { routeId: string }) {
  return (
    <PageContainer routeId={routeId}>
      <SectionCard title="团队更多入口" description="团队域的成员、消息和治理入口被整理成目录卡片，便于按任务进入。">
        <LinkCardGrid
          items={[
            { id: 'team', title: '团队治理', description: '维护团队本身、管理员摘要和成员上限。', to: '/team/team', actionLabel: '进入团队治理' },
            { id: 'members', title: '团队成员', description: '处理成员加入、移除、团队角色和边界角色分配。', to: '/team/team-members', actionLabel: '进入成员治理' },
            { id: 'dispatch', title: '团队消息调度', description: '在团队域发起消息调度并追踪投递记录。', to: '/team/message', actionLabel: '打开消息调度' },
            { id: 'templates', title: '团队消息模板', description: '维护团队域消息模板。', to: '/team/message-template', actionLabel: '进入模板治理' },
            { id: 'records', title: '团队消息记录', description: '查看团队域消息投递详情和 payload 摘要。', to: '/team/message-record', actionLabel: '查看消息记录' },
          ]}
        />
      </SectionCard>
    </PageContainer>
  )
}
