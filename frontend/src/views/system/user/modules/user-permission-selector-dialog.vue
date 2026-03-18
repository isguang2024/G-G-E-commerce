<template>
  <ElDialog
    v-model="visible"
    :title="`用户功能权限 - ${userTitle}`"
    width="980px"
    destroy-on-close
    class="user-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        配置个人功能权限。默认继承角色，仅在例外场景下使用单独配置。
      </div>

      <ElTabs v-model="activeTab" class="permission-tabs">
        <ElTabPane label="例外配置" name="custom">
          <PermissionActionWorkbench
            mode="user"
            :actions="permissionActions"
            :decision-map="decisionMap"
            :loading="loading"
            search-placeholder="搜索权限名称、权限 ID、模块归属"
            @update:decision-map="decisionMap = $event"
          />
        </ElTabPane>

        <ElTabPane label="角色继承" name="roles">
          <div class="roles-panel">
            <div class="roles-summary">
              <ElTag effect="plain" round>角色 {{ roleTags.length }}</ElTag>
              <ElTag type="warning" effect="plain" round>例外 {{ overrideCount }}</ElTag>
            </div>

            <div class="roles-list">
              <ElEmpty v-if="roleTags.length === 0" description="当前用户未绑定角色" />
              <div v-else class="role-tag-list">
                <ElTag
                  v-for="role in roleTags"
                  :key="role"
                  effect="plain"
                  size="large"
                  round
                >
                  {{ role }}
                </ElTag>
              </div>
            </div>
          </div>
        </ElTabPane>

        <ElTabPane label="结果预览" name="preview">
          <div class="preview-panel">
            <div class="preview-summary">
              <ElTag effect="plain" round>有效权限 {{ effectivePermissions.length }}</ElTag>
              <ElTag type="warning" effect="plain" round>例外配置 {{ overrideCount }}</ElTag>
            </div>

            <ElScrollbar max-height="520px">
              <section
                v-for="group in previewGroups"
                :key="group.key"
                class="preview-group"
              >
                <header class="preview-group__header">
                  <span>{{ group.label }}</span>
                  <ElTag effect="plain" size="small" round>{{ group.items.length }}</ElTag>
                </header>
                <div class="preview-group__body">
                  <ElTag
                    v-for="item in group.items"
                    :key="item.id"
                    :type="item.effect === 'allow' ? 'success' : 'warning'"
                    effect="plain"
                    round
                  >
                    {{ item.name }}
                  </ElTag>
                </div>
              </section>
            </ElScrollbar>
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
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import PermissionActionWorkbench from '@/components/business/permission/PermissionActionWorkbench.vue'
import { fetchGetPermissionActionList, fetchGetUserActions, fetchSetUserActions } from '@/api/system-manage'

interface Props {
  modelValue: boolean
  userData?: Api.SystemManage.UserListItem
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
const activeTab = ref('custom')
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const decisionMap = ref<Record<string, '' | 'allow' | 'deny'>>({})

const userTitle = computed(() => props.userData?.nickName || props.userData?.userName || '')
const overrideCount = computed(() => Object.keys(decisionMap.value).length)
const roleTags = computed(() => {
  const detailRoles = ((props.userData as any)?.roleDetails || []) as Array<{ code?: string; name?: string }>
  if (detailRoles.length > 0) {
    return detailRoles.map((item) => item.name || item.code || '').filter(Boolean)
  }
  const roleCodes = ((props.userData as any)?.roles || []) as string[]
  return roleCodes.filter(Boolean)
})

const effectivePermissions = computed(() => {
  return permissionActions.value.filter((item) => decisionMap.value[item.id] !== 'deny')
})

const previewGroups = computed(() => {
  const map = new Map<string, { key: string; label: string; items: Array<{ id: string; name: string; effect: string }> }>()
  effectivePermissions.value.forEach((item) => {
    const key = `${item.moduleCode || item.resourceCode || 'default'}`
    if (!map.has(key)) {
      map.set(key, {
        key,
        label: item.moduleCode || item.resourceCode || '未分类模块',
        items: []
      })
    }
    map.get(key)!.items.push({
      id: item.id,
      name: item.name,
      effect: decisionMap.value[item.id] || 'inherit'
    })
  })
  return Array.from(map.values())
})

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      loadData()
    }
  }
)

async function loadData() {
  if (!props.userData?.id) return
  loading.value = true
  activeTab.value = 'custom'
  try {
    const [actionsRes, currentRes] = await Promise.all([
      fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal', scopeCode: 'global' }),
      fetchGetUserActions(props.userData.id)
    ])

    permissionActions.value = (actionsRes?.records || []).filter((item) => !item.requiresTenantContext)

    const nextMap: Record<string, '' | 'allow' | 'deny'> = {}
    currentRes.forEach((item) => {
      if (item.actionId && (item.effect === 'allow' || item.effect === 'deny')) {
        nextMap[item.actionId] = item.effect
      }
    })
    decisionMap.value = nextMap
  } catch (error: any) {
    ElMessage.error(error?.message || '加载用户权限失败')
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  visible.value = false
}

async function handleSave() {
  if (!props.userData?.id) return
  saving.value = true
  try {
    const payload = Object.entries(decisionMap.value)
      .filter(([, effect]) => effect === 'allow' || effect === 'deny')
      .map(([actionId, effect]) => ({
        action_id: actionId,
        effect: effect as 'allow' | 'deny'
      }))
    await fetchSetUserActions(props.userData.id, payload)
    ElMessage.success('用户功能权限已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存用户权限失败')
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

.roles-panel,
.preview-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.roles-summary,
.preview-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.roles-list,
.preview-group {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 16px;
  background: #fff;
}

.roles-list {
  min-height: 220px;
  padding: 20px;
}

.role-tag-list,
.preview-group__body {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.preview-group {
  margin-bottom: 12px;
  overflow: hidden;
}

.preview-group__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fbfcfe;
  border-bottom: 1px solid var(--el-border-color-lighter);
  color: #111827;
  font-weight: 600;
}

.preview-group__body {
  padding: 16px;
}
</style>
