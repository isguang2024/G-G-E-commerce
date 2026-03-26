<template>
  <ElDrawer
    v-model="visible"
    size="1120px"
    destroy-on-close
    class="menu-permission-dialog config-drawer"
    direction="rtl">
    <template #header>
      <div class="dialog-title">
        <span class="dialog-title-text">{{ dialogTitle }}</span>
        <span v-if="dialogPath" class="dialog-title-path">{{ dialogPath }}</span>
      </div>
    </template>

    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        配置访问当前菜单所需的功能权限。父级和子级都可以直接选择，保存时会自动换算成具体权限项。
        菜单功能权限只影响入口访问条件。用于控制个别页面不让某些权限不足或者含有权限的人进入。
      </div>

      <PermissionActionCascaderPanel
        :actions="permissionActions"
        :selected-ids="selectedPermissions"
        footer-text="用于控制个别页面不让某些权限不足或者含有权限的人进入。"
        @update:selected-ids="selectedPermissions = $event"
      />

      <div class="advanced-card">
        <div class="advanced-title">高级配置</div>
        <div class="advanced-grid">
          <div class="advanced-item">
            <span class="advanced-label">权限满足方式</span>
            <ElRadioGroup v-model="matchMode" size="small" class="compact-radio-group">
              <ElRadio value="any" border>任意满足</ElRadio>
              <ElRadio value="all" border>全部满足</ElRadio>
            </ElRadioGroup>
          </div>
          <div class="advanced-item">
            <span class="advanced-label">功能权限不满足时</span>
            <ElRadioGroup v-model="visibilityMode" size="small" class="compact-radio-group">
              <ElRadio value="hide" border>隐藏菜单</ElRadio>
              <ElRadio value="show" border>显示菜单</ElRadio>
            </ElRadioGroup>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
import { fetchGetPermissionActionList } from '@/api/system-manage'
import { buildScopedActionKey, resolveActionKey } from '@/utils/permission/action'

let cachedPermissionActions: Api.SystemManage.PermissionActionItem[] | null = null
let permissionActionsPromise: Promise<Api.SystemManage.PermissionActionItem[]> | null = null

interface Props {
  modelValue: boolean
  menuData?: any
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'submit', data: {
    requiredActions: string[]
    actionMatchMode: 'any' | 'all'
    actionVisibilityMode: 'hide' | 'show'
  }): void
  (e: 'cancel'): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const saving = ref(false)
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const selectedPermissions = ref<string[]>([])
const matchMode = ref<'any' | 'all'>('any')
const visibilityMode = ref<'hide' | 'show'>('hide')

const menuTitle = computed(() => props.menuData?.meta?.title || props.menuData?.name || '当前菜单')
const dialogPath = computed(() => `${props.menuData?.path || ''}`.trim())
const dialogTitle = computed(() => `功能权限 - ${menuTitle.value}`)

watch(
  () => props.modelValue,
  (open) => {
    if (open) loadData()
  }
)

async function loadData() {
  loading.value = true
  try {
    const records = await loadPermissionActions()
    permissionActions.value = records

    const meta = props.menuData?.meta || {}
    const initialActions = Array.from(
      new Set(
        [meta.requiredAction, ...(Array.isArray(meta.requiredActions) ? meta.requiredActions : [])]
          .map((item: string) => `${item || ''}`.trim())
          .filter(Boolean)
      )
    )

    selectedPermissions.value = normalizeRequiredActions(initialActions, records)
    matchMode.value = meta.actionMatchMode === 'all' ? 'all' : 'any'
    visibilityMode.value = meta.actionVisibilityMode === 'show' ? 'show' : 'hide'
  } catch (error: any) {
    ElMessage.error(error?.message || '加载菜单功能权限失败')
  } finally {
    loading.value = false
  }
}

async function loadPermissionActions() {
  if (cachedPermissionActions) {
    return cachedPermissionActions
  }

  if (!permissionActionsPromise) {
    permissionActionsPromise = fetchGetPermissionActionList({
      current: 1,
      size: 1000,
      status: 'normal'
    })
      .then((res) => res?.records || [])
      .then((records) => {
        cachedPermissionActions = records
        return records
      })
      .finally(() => {
        permissionActionsPromise = null
      })
  }

  return permissionActionsPromise
}

function normalizeRequiredActions(
  actions: string[],
  availableActions: Api.SystemManage.PermissionActionItem[]
) {
  const actionKeyMap = new Map<string, string>()

  availableActions.forEach((item) => {
    const key = buildScopedActionKey(item.permissionKey || `${item.resourceCode}:${item.actionCode}`)
    actionKeyMap.set(key, key)
  })

  return Array.from(
    new Set(
      actions.map((item) => {
        const raw = resolveActionKey(item).key
        return actionKeyMap.get(raw) || item
      })
    )
  )
}

function handleCancel() {
  emit('cancel')
  visible.value = false
}

async function handleSave() {
  saving.value = true
  try {
    emit('submit', {
      requiredActions: expandSelectedValues(selectedPermissions.value, permissionActions.value),
      actionMatchMode: matchMode.value,
      actionVisibilityMode: visibilityMode.value
    })
    visible.value = false
  } finally {
    saving.value = false
  }
}

function expandSelectedValues(
  values: string[],
  actions: Api.SystemManage.PermissionActionItem[]
) {
  const result = new Set<string>()
  const featureMap = new Map<string, Api.SystemManage.PermissionActionItem[]>()
  const moduleMap = new Map<string, Api.SystemManage.PermissionActionItem[]>()

  actions.forEach((action) => {
    const featureKey = `${action.featureGroupId || action.featureKind || 'business'}`
    const moduleKey = `${action.moduleGroupId || action.moduleCode || action.resourceCode || 'default'}`
    const featureValue = `feature:${featureKey}`
    const moduleValue = `module:${featureKey}:${moduleKey}`

    const featureItems = featureMap.get(featureValue) || []
    featureItems.push(action)
    featureMap.set(featureValue, featureItems)

    const moduleItems = moduleMap.get(moduleValue) || []
    moduleItems.push(action)
    moduleMap.set(moduleValue, moduleItems)
  })

  values.forEach((value) => {
    if (featureMap.has(value)) {
      featureMap.get(value)!.forEach((item) => result.add(item.id))
      return
    }
    if (moduleMap.has(value)) {
      moduleMap.get(value)!.forEach((item) => result.add(item.id))
      return
    }
    result.add(value)
  })

  return Array.from(result)
}
</script>

<style scoped lang="scss">
.dialog-title {
  display: flex;
  align-items: baseline;
  gap: 8px;
}

.dialog-title-text {
  color: var(--el-text-color-primary);
  font-size: var(--el-dialog-title-font-size);
  font-weight: 400;
  line-height: var(--el-dialog-font-line-height);
}

.dialog-title-path {
  color: #9ca3af;
  font-size: 12px;
  font-weight: 400;
  line-height: var(--el-dialog-font-line-height);
}

.dialog-shell {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dialog-note,
.summary-label,
.advanced-label {
  color: #6b7280;
}

.dialog-note {
  line-height: 1.6;
}

.advanced-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 14px;
  background: #fbfcfe;
}
.advanced-item {
  display: flex;
  flex-direction: column;
}

.advanced-card {
  padding: 12px 16px;
}

.advanced-title {
  margin-bottom: 8px;
  color: #111827;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
}

.advanced-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px 16px;
}

.advanced-item {
  gap: 6px;
}

.advanced-label {
  font-size: 12px;
  line-height: 1.2;
}

.compact-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.compact-radio-group :deep(.el-radio.is-bordered) {
  height: 30px;
  margin-right: 0;
  padding: 0 12px;
}

@media (max-width: 900px) {
  .advanced-grid {
    grid-template-columns: 1fr;
  }
}
</style>
