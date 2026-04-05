<template>
  <div class="art-full-height">
    <AdminWorkspaceHero
      title="团队角色与权限"
      description="统一查看当前团队角色、功能包、菜单裁剪和权限裁剪，先把团队边界收口。"
      :metrics="heroMetrics"
    >
      <div class="team-role-hero-actions">
        <ElSelect
          v-model="selectedAppKey"
          clearable
          filterable
          placeholder="选择 App"
          class="team-role-app-select"
          @change="handleManagedAppChange"
        >
          <ElOption
            v-for="item in appOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
        <ElButton v-if="hasAction('team.member.manage')" type="primary" @click="openAddDialog">新增团队角色</ElButton>
      </div>
    </AdminWorkspaceHero>

    <ElCard class="art-table-card" shadow="never">
      <template #header>
        <div class="header-row">
          <div>
            <div class="font-medium">当前团队角色管理</div>
            <div class="mt-2 text-sm text-gray-500">
              基础团队角色默认继承当前团队已开通功能包；当前团队自定义角色先绑定功能包，再在功能包范围内维护菜单权限和角色权限。
            </div>
          </div>
          <ElButton v-if="hasAction('team.member.manage')" type="primary" @click="openAddDialog">新增团队角色</ElButton>
        </div>
      </template>

      <ArtTableHeader
        v-model:columns="columnChecks"
        v-model:showSearchBar="showSearchBar"
        :loading="loading"
        @refresh="refreshData"
      />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <TeamRoleDialog v-model="roleDialog" :dialog-type="dialogType" :role-data="currentRoleData" @success="onSuccess" />
    <TeamRolePackageDialog
      v-model="packageDialog"
      :role-data="currentRoleData"
      :app-key="targetAppKey"
      @success="onSuccess"
    />
    <TeamRoleMenuDialog
      v-model="menuDialog"
      :role-data="currentRoleData"
      :app-key="targetAppKey"
      @success="onSuccess"
    />
    <TeamRoleActionDialog
      v-model="actionDialog"
      :role-data="currentRoleData"
      :app-key="targetAppKey"
      @success="onSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, ref, watch } from 'vue'
  import { ElMessageBox, ElTag } from 'element-plus'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { ButtonMoreItem } from '@/components/core/forms/art-button-more/index.vue'
  import { useAuth } from '@/hooks/core/useAuth'
  import { useTable } from '@/hooks/core/useTable'
  import { fetchDeleteMyTeamRole, fetchGetMyTeamBoundaryRoles } from '@/api/team'
  import { useManagedAppScope } from '@/hooks/business/useManagedAppScope'
  import { fetchGetApps } from '@/api/system-manage'
  import ArtButtonMore from '@/components/core/forms/art-button-more/index.vue'
  import TeamRoleDialog from './modules/team-role-dialog.vue'
  import TeamRolePackageDialog from './modules/team-role-package-dialog.vue'
  import TeamRoleMenuDialog from './modules/team-role-menu-dialog.vue'
  import TeamRoleActionDialog from './modules/team-role-action-dialog.vue'

  defineOptions({ name: 'TeamRolesAndPermissions' })

  type RoleListItem = Api.SystemManage.RoleListItem
  const { hasAction } = useAuth()
  const { targetAppKey, setManagedAppKey } = useManagedAppScope()

  const showSearchBar = ref(false)
  const appList = ref<Api.SystemManage.AppItem[]>([])
  const selectedAppKey = ref('')
  const roleDialog = ref(false)
  const packageDialog = ref(false)
  const menuDialog = ref(false)
  const actionDialog = ref(false)
  const dialogType = ref<'add' | 'edit'>('add')
  const currentRoleData = ref<RoleListItem | undefined>(undefined)
  const heroMetrics = computed(() => [
    { label: '当前 App', value: targetAppKey.value },
    { label: '角色总数', value: data.value.length || 0 },
    { label: '基础角色', value: baseRoleCount.value },
    { label: '团队自定义', value: customRoleCount.value }
  ])
  const baseRoleCount = computed(() => data.value.filter((item) => item.isGlobal).length)
  const customRoleCount = computed(() => data.value.filter((item) => !item.isGlobal).length)
  const appOptions = computed(() =>
    appList.value.map((item) => ({
      label: item.name ? `${item.name}（${item.appKey}）` : item.appKey,
      value: item.appKey
    }))
  )

  const { columns, columnChecks, data, loading, pagination, handleSizeChange, handleCurrentChange, refreshData } =
    useTable({
      core: {
        apiFn: async () => {
          if (!targetAppKey.value) {
            return {
              records: [],
              total: 0,
              current: 1,
              size: 20
            }
          }
          const list = await fetchGetMyTeamBoundaryRoles(targetAppKey.value)
          return {
            records: list,
            total: list.length,
            current: 1,
            size: list.length || 20
          }
        },
        apiParams: {},
        excludeParams: [],
        columnsFactory: () => [
          { prop: 'roleName', label: '角色名称', minWidth: 140 },
          { prop: 'roleCode', label: '角色编码', minWidth: 140 },
          {
            prop: 'scope',
            label: '来源',
            width: 110,
            formatter: (row: RoleListItem) =>
              h(
                ElTag,
                { type: row.isGlobal ? 'info' : 'success', effect: 'plain' },
                () => (row.isGlobal ? '基础角色' : '团队自定义')
              )
          },
          { prop: 'description', label: '描述', minWidth: 200, showOverflowTooltip: true },
          { prop: 'createTime', label: '创建时间', width: 170 },
          {
            prop: 'operation',
            label: '操作',
            width: 220,
            fixed: 'right',
            formatter: (row: RoleListItem) => {
              const list = [
                { key: 'packages', label: row.isGlobal ? '查看功能包' : '功能包', icon: 'ri:apps-2-line' },
                { key: 'menus', label: row.isGlobal ? '查看菜单裁剪' : '菜单裁剪', icon: 'ri:menu-line' },
                { key: 'actions', label: row.isGlobal ? '查看权限裁剪' : '权限裁剪', icon: 'ri:shield-keyhole-line' }
              ]
              if (!row.isGlobal && hasAction('team.member.manage')) {
                list.unshift({ key: 'edit', label: '编辑角色', icon: 'ri:edit-line' })
                list.push({ key: 'delete', label: '删除角色', icon: 'ri:delete-bin-line' })
              }
              return h('div', [
                h(ArtButtonMore, {
                  list,
                  onClick: (item: ButtonMoreItem) => buttonMoreClick(item, row)
                })
              ])
            }
          }
        ]
      }
    })

  function openAddDialog() {
    dialogType.value = 'add'
    currentRoleData.value = undefined
    roleDialog.value = true
  }

  async function buttonMoreClick(item: ButtonMoreItem, row: RoleListItem) {
    currentRoleData.value = row
    if (item.key === 'edit') {
      dialogType.value = 'edit'
      roleDialog.value = true
      return
    }
    if (item.key === 'menus') {
      menuDialog.value = true
      return
    }
    if (item.key === 'packages') {
      packageDialog.value = true
      return
    }
    if (item.key === 'actions') {
      actionDialog.value = true
      return
    }
    if (item.key === 'delete') {
      await ElMessageBox.confirm(`确认删除团队角色“${row.roleName}”吗？`, '删除确认', { type: 'warning' })
      await fetchDeleteMyTeamRole(row.roleId)
      onSuccess()
    }
  }

  function onSuccess() {
    refreshData()
  }

  async function loadAppOptions() {
    const res = await fetchGetApps()
    appList.value = res.records || []
  }

  async function handleManagedAppChange(value?: string) {
    await setManagedAppKey(`${value || ''}`.trim())
  }

  onMounted(() => {
    selectedAppKey.value = targetAppKey.value
    loadAppOptions().catch(() => {
      appList.value = []
    })
  })

  watch(
    () => targetAppKey.value,
    (value) => {
      selectedAppKey.value = value || ''
      refreshData()
    }
  )
</script>

<style scoped lang="scss">
  .team-role-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .team-role-app-select {
    width: 240px;
  }

  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
  }
</style>

