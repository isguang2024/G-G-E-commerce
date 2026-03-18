<template>
  <ElDialog
    v-model="visible"
    :title="`功能门槛 - ${menuTitle}`"
    width="960px"
    destroy-on-close
    class="menu-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        配置访问当前菜单所需的功能权限。默认只在入口控制层生效，用于菜单展示与访问前置判断。
      </div>

      <div class="menu-summary">
        <div class="summary-block">
          <span class="summary-label">菜单</span>
          <strong>{{ menuTitle }}</strong>
        </div>
        <div class="summary-block">
          <span class="summary-label">路由</span>
          <strong>{{ menuData?.path || '-' }}</strong>
        </div>
      </div>

      <PermissionActionWorkbench
        mode="menu"
        :actions="permissionActions"
        :selected-ids="selectedPermissions"
        :loading="loading"
        search-placeholder="搜索权限名称、权限 ID、模块归属"
        @update:selected-ids="selectedPermissions = $event"
      />

      <div class="advanced-card">
        <div class="advanced-title">高级配置</div>
        <div class="advanced-grid">
          <div class="advanced-item">
            <span class="item-label">匹配方式</span>
            <ElRadioGroup v-model="matchMode">
              <ElRadioButton label="any">任意满足</ElRadioButton>
              <ElRadioButton label="all">全部满足</ElRadioButton>
            </ElRadioGroup>
          </div>
          <div class="advanced-item">
            <span class="item-label">权限不足时</span>
            <ElRadioGroup v-model="visibilityMode">
              <ElRadioButton label="hide">隐藏菜单</ElRadioButton>
              <ElRadioButton label="show">显示菜单</ElRadioButton>
            </ElRadioGroup>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <ElButton @click="handleCancel">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import PermissionActionWorkbench from '@/components/business/permission/PermissionActionWorkbench.vue'
import { fetchGetPermissionActionList } from '@/api/system-manage'
import { buildScopedActionKey, resolveActionKey } from '@/utils/permission/action'

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

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      loadData()
    }
  }
)

async function loadData() {
  loading.value = true
  try {
    const res = await fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal' })
    const records = res?.records || []
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
    ElMessage.error(error?.message || '加载菜单功能门槛失败')
  } finally {
    loading.value = false
  }
}

function normalizeRequiredActions(
  actions: string[],
  availableActions: Api.SystemManage.PermissionActionItem[]
) {
  const scopedKeyMap = new Map<string, string>()
  const unscopedKeyMap = new Map<string, string[]>()

  availableActions.forEach((item) => {
    const scopedKey = buildScopedActionKey(
      `${item.resourceCode}:${item.actionCode}`,
      item.scopeCode || item.scope
    )
    const unscopedKey = resolveActionKey(scopedKey).key
    scopedKeyMap.set(scopedKey, scopedKey)
    const current = unscopedKeyMap.get(unscopedKey) || []
    current.push(scopedKey)
    unscopedKeyMap.set(unscopedKey, current)
  })

  return Array.from(
    new Set(
      actions.map((item) => {
        if (scopedKeyMap.has(item)) {
          return item
        }

        const raw = resolveActionKey(item)
        const candidates = unscopedKeyMap.get(raw.key) || []
        if (!candidates.length) {
          return item
        }

        if (raw.scope) {
          const exactKey = buildScopedActionKey(raw.key, raw.scope)
          const exact = candidates.find((candidate) => candidate === exactKey)
          if (exact) {
            return exact
          }
        }

        return candidates.find((candidate) => candidate.endsWith('@global')) || candidates[0]
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
      requiredActions: [...selectedPermissions.value],
      actionMatchMode: matchMode.value,
      actionVisibilityMode: visibilityMode.value
    })
    visible.value = false
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

.menu-summary,
.advanced-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  background: #fbfcfe;
}

.menu-summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  padding: 16px;
}

.summary-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.summary-label,
.item-label {
  color: #6b7280;
  font-size: 13px;
}

.advanced-card {
  padding: 16px;
}

.advanced-title {
  margin-bottom: 12px;
  color: #111827;
  font-weight: 600;
}

.advanced-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.advanced-item {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

@media (max-width: 900px) {
  .menu-summary,
  .advanced-grid {
    grid-template-columns: 1fr;
  }
}
</style>
