<template>
  <ElDialog
    v-model="visible"
    :title="`角色权限配置 - ${roleTitle}`"
    width="1040px"
    destroy-on-close
    class="role-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        统一配置角色菜单权限、功能权限和数据权限。菜单影响可见入口，功能权限决定动作能力，数据权限控制资源范围。
      </div>

      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="菜单权限" name="menus">
          <div class="menu-panel">
            <div class="menu-toolbar">
              <ElInput
                v-model="menuKeyword"
                placeholder="搜索菜单名称或路由"
                clearable
                class="menu-search"
              >
                <template #prefix>
                  <ElIcon><Search /></ElIcon>
                </template>
              </ElInput>
              <div class="menu-toolbar__actions">
                <ElButton text @click="toggleMenuExpand(true)">全部展开</ElButton>
                <ElButton text @click="toggleMenuExpand(false)">全部收起</ElButton>
                <ElButton text @click="toggleAllMenus(true)">全选当前菜单</ElButton>
                <ElButton text @click="toggleAllMenus(false)">清空当前菜单</ElButton>
              </div>
            </div>

            <div class="menu-summary">
              <ElTag effect="plain" round>已选菜单 {{ selectedMenuIds.length }}</ElTag>
              <ElTag type="info" effect="plain" round>总计 {{ menuNodeCount }}</ElTag>
            </div>

            <div class="menu-tree-shell">
              <ElScrollbar max-height="520px">
                <ElTree
                  ref="menuTreeRef"
                  :data="menuTreeData"
                  show-checkbox
                  node-key="id"
                  :props="{ label: 'label', children: 'children' }"
                  :filter-node-method="filterMenuNode"
                  @check="handleMenuCheck"
                >
                  <template #default="{ data }">
                    <div class="menu-node">
                      <span class="menu-node__title">{{ data.label }}</span>
                      <span class="menu-node__path">{{ data.path || '/' }}</span>
                    </div>
                  </template>
                </ElTree>
              </ElScrollbar>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="功能权限" name="actions">
          <PermissionActionWorkbench
            mode="role"
            :actions="permissionActions"
            :decision-map="actionDecisionMap"
            :loading="loading"
            search-placeholder="搜索功能权限名称、权限 ID、模块归属"
            @update:decision-map="actionDecisionMap = $event"
          />
        </ElTabPane>

        <ElTabPane label="数据权限" name="data">
          <div class="data-panel">
            <div class="data-summary">
              <ElTag effect="plain" round>资源 {{ dataRows.length }}</ElTag>
              <ElTag type="warning" effect="plain" round>已配置 {{ configuredDataCount }}</ElTag>
            </div>

            <ElTable :data="dataRows" border class="data-table">
              <ElTableColumn prop="resourceName" label="资源" min-width="220" />
              <ElTableColumn label="数据范围" min-width="260">
                <template #default="{ row }">
                  <ElSelect
                    v-model="row.selectedScope"
                    placeholder="选择作用域"
                    clearable
                    style="width: 100%"
                  >
                    <ElOption
                      v-for="scope in scopeOptions"
                      :key="scope.scopeCode"
                      :label="scope.scopeName"
                      :value="scope.scopeCode"
                    />
                  </ElSelect>
                </template>
              </ElTableColumn>
            </ElTable>
          </div>
        </ElTabPane>
      </ElTabs>
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import PermissionActionWorkbench from '@/components/business/permission/PermissionActionWorkbench.vue'
import {
  fetchGetMenuTreeAll,
  fetchGetPermissionActionList,
  fetchGetRoleActions,
  fetchGetRoleDataPermissions,
  fetchGetRoleMenus,
  fetchSetRoleActions,
  fetchSetRoleDataPermissions,
  fetchSetRoleMenus
} from '@/api/system-manage'

interface RoleMenuNode {
  id: string
  label: string
  path?: string
  children?: RoleMenuNode[]
}

interface DataRow {
  resourceCode: string
  resourceName: string
  selectedScope: string
}

interface Props {
  modelValue: boolean
  roleData?: Api.SystemManage.RoleListItem
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'success'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const saving = ref(false)
const activeTab = ref('menus')
const menuTreeRef = ref()
const menuKeyword = ref('')
const menuTreeData = ref<RoleMenuNode[]>([])
const selectedMenuIds = ref<string[]>([])
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const actionDecisionMap = ref<Record<string, '' | 'allow' | 'deny'>>({})
const dataRows = ref<DataRow[]>([])
const scopeOptions = ref<Api.SystemManage.RoleDataPermissionScopeOption[]>([])

const roleTitle = computed(() => props.roleData?.roleName || '')
const menuNodeCount = computed(() => flattenMenuIds(menuTreeData.value).length)
const configuredDataCount = computed(() => dataRows.value.filter((item) => item.selectedScope).length)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      loadData()
    }
  }
)

watch(menuKeyword, (value) => {
  if (menuTreeRef.value) {
    menuTreeRef.value.filter(value)
  }
})

async function loadData() {
  if (!props.roleData?.roleId) return
  loading.value = true
  activeTab.value = 'menus'
  try {
    const [menuTree, roleMenus, actionList, roleActions, dataPermissionRes] = await Promise.all([
      fetchGetMenuTreeAll(),
      fetchGetRoleMenus(props.roleData.roleId),
      fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal' }),
      fetchGetRoleActions(props.roleData.roleId),
      fetchGetRoleDataPermissions(props.roleData.roleId)
    ])

    menuTreeData.value = normalizeMenus(Array.isArray(menuTree) ? menuTree : [])
    selectedMenuIds.value = (roleMenus?.menu_ids || []).map((item) => `${item}`)
    permissionActions.value = actionList?.records || []

    const nextDecisionMap: Record<string, '' | 'allow' | 'deny'> = {}
    ;(roleActions?.actions || []).forEach((item) => {
      if (item.action_id && (item.effect === 'allow' || item.effect === 'deny')) {
        nextDecisionMap[item.action_id] = item.effect
      }
    })
    actionDecisionMap.value = nextDecisionMap

    scopeOptions.value = (dataPermissionRes?.available_scopes || []).map((item) => ({
      scopeCode: item.scope_code,
      scopeName: item.scope_name
    }))
    const selectedScopeMap = new Map<string, string>()
    ;(dataPermissionRes?.permissions || []).forEach((item) => {
      selectedScopeMap.set(item.resource_code, item.scope_code)
    })
    dataRows.value = (dataPermissionRes?.resources || []).map((item) => ({
      resourceCode: item.resource_code,
      resourceName: item.resource_name,
      selectedScope: selectedScopeMap.get(item.resource_code) || ''
    }))

    await nextTick()
    menuTreeRef.value?.setCheckedKeys(selectedMenuIds.value)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载角色权限失败')
  } finally {
    loading.value = false
  }
}

function normalizeMenus(items: any[]): RoleMenuNode[] {
  return items.map((item) => ({
    id: `${item.id}`,
    label: item.meta?.title || item.name || item.path || '未命名菜单',
    path: item.path || '',
    children: Array.isArray(item.children) ? normalizeMenus(item.children) : []
  }))
}

function flattenMenuIds(items: RoleMenuNode[]): string[] {
  return items.flatMap((item) => [item.id, ...flattenMenuIds(item.children || [])])
}

function filterMenuNode(value: string, data: any) {
  if (!value) return true
  const keyword = value.trim().toLowerCase()
  return `${data?.label || ''} ${data?.path || ''}`.toLowerCase().includes(keyword)
}

function handleMenuCheck(_: any, payload: any) {
  selectedMenuIds.value = (payload.checkedKeys || []).map((item: string | number) => `${item}`)
}

function toggleMenuExpand(expanded: boolean) {
  const nodes = menuTreeRef.value?.store?.nodesMap || {}
  Object.values(nodes).forEach((node: any) => {
    node.expanded = expanded
  })
}

function toggleAllMenus(selected: boolean) {
  const ids = flattenMenuIds(menuTreeData.value)
  const next = new Set(selectedMenuIds.value)
  ids.forEach((id) => {
    if (selected) {
      next.add(id)
    } else {
      next.delete(id)
    }
  })
  selectedMenuIds.value = Array.from(next)
  menuTreeRef.value?.setCheckedKeys(selectedMenuIds.value)
}

function handleCancel() {
  visible.value = false
}

async function handleSave() {
  if (!props.roleData?.roleId) return
  saving.value = true
  try {
    const actionPayload = Object.entries(actionDecisionMap.value)
      .filter(([, effect]) => effect === 'allow' || effect === 'deny')
      .map(([actionId, effect]) => ({
        action_id: actionId,
        effect: effect as 'allow' | 'deny'
      }))

    const dataPayload = dataRows.value
      .filter((item) => item.selectedScope)
      .map((item) => ({
        resource_code: item.resourceCode,
        scope_code: item.selectedScope
      }))

    await Promise.all([
      fetchSetRoleMenus(props.roleData.roleId, selectedMenuIds.value),
      fetchSetRoleActions(props.roleData.roleId, actionPayload),
      fetchSetRoleDataPermissions(props.roleData.roleId, dataPayload)
    ])

    ElMessage.success('角色权限已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存角色权限失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped lang="scss">
.dialog-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dialog-note {
  color: #6b7280;
  line-height: 1.6;
}

.permission-tabs :deep(.el-tabs__content) {
  padding-top: 12px;
}

.menu-panel,
.data-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.menu-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
}

.menu-search {
  max-width: 360px;
}

.menu-toolbar__actions,
.menu-summary,
.data-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.menu-tree-shell {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  padding: 12px;
  background: #fff;
}

.menu-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  min-width: 0;
}

.menu-node__title {
  color: #111827;
}

.menu-node__path {
  color: #6b7280;
  font-size: 13px;
}

.data-table {
  border-radius: 16px;
  overflow: hidden;
}

@media (max-width: 960px) {
  .menu-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .menu-search {
    max-width: none;
  }
}
</style>
