<template>
  <ElDialog
    v-model="visible"
    :title="`团队补充权限 - ${teamName}`"
    width="960px"
    destroy-on-close
    class="team-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里配置的是团队边界之外的额外补充权限。正式开通能力请通过功能包完成；这里保存的内容只会作为团队补充能力叠加到功能包展开结果之上。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>团队 {{ teamName }}</ElTag>
        <ElTag type="success" effect="plain" round>边界结果 {{ selectedIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>候选 {{ permissionActions.length }}</ElTag>
        <ElTag effect="plain" round>功能包 {{ featurePackages.length }}</ElTag>
        <ElTag type="warning" effect="plain" round>功能包展开 {{ derivedActionIds.length }}</ElTag>
        <ElTag type="primary" effect="plain" round>额外补充 {{ manualActionIds.length }}</ElTag>
      </div>

      <div v-if="featurePackages.length" class="package-card">
        <div class="package-title">已开通功能包</div>
        <div class="package-help">功能包决定团队正式开通的基础能力；当前弹窗只负责补充功能包之外的少量例外权限。</div>
        <div class="package-tags">
          <ElTag
            v-for="item in featurePackages"
            :key="item.id"
            type="success"
            effect="plain"
            round
          >
            {{ item.name }} · {{ item.actionCount ?? 0 }}
          </ElTag>
        </div>
      </div>

      <div class="source-note">
        <span>黄色统计表示来自功能包展开的基础能力，蓝色统计表示团队额外补充能力。</span>
      </div>

      <ElAlert
        v-if="fromCache"
        type="warning"
        :closable="false"
        title="当前团队边界来源暂时从缓存回退读取，建议后续刷新团队边界缓存。"
      />

      <div v-if="derivedActions.length || manualActions.length" class="source-detail-grid">
        <div v-if="derivedActions.length" class="source-card source-card--derived">
          <div class="source-header">
            <div class="source-title">功能包展开明细</div>
            <ElButton
              v-if="selectedDerivedPackage"
              type="warning"
              text
              @click="goToFeaturePackagePage(selectedDerivedPackage)"
            >
              前往功能包页
            </ElButton>
          </div>
          <div v-if="derivedSourcePackages.length" class="package-filter-row">
            <ElTag
              :type="selectedDerivedPackageId ? 'info' : 'warning'"
              effect="plain"
              round
              class="package-filter-tag"
              @click="selectedDerivedPackageId = ''"
            >
              全部功能包
            </ElTag>
            <ElTag
              v-for="item in derivedSourcePackages"
              :key="item.id"
              :type="selectedDerivedPackageId === item.id ? 'warning' : 'info'"
              effect="plain"
              round
              class="package-filter-tag"
              @click="selectedDerivedPackageId = selectedDerivedPackageId === item.id ? '' : item.id"
            >
              {{ item.name }}
            </ElTag>
          </div>
          <div class="source-tags">
            <ElTag
              v-for="item in filteredDerivedActions"
              :key="item.id"
              type="warning"
              effect="plain"
              round
              :title="buildDerivedSourceText(item.id)"
            >
              {{ item.name }}
            </ElTag>
          </div>
        </div>

        <div v-if="manualActions.length" class="source-card source-card--manual">
          <div class="source-title">团队补充明细</div>
          <div class="source-tags">
            <ElTag
              v-for="item in manualActions"
              :key="item.id"
              type="primary"
              effect="plain"
              round
            >
              {{ item.name }}
            </ElTag>
          </div>
        </div>
      </div>

      <PermissionActionCascaderPanel
        :actions="permissionActions"
        :selected-ids="selectedIds"
        footer-text="这里只保存团队补充权限，最终团队边界 = 功能包展开 + 团队补充。提交前会自动展开父级选择。"
        @update:selected-ids="selectedIds = $event"
      />
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
import { useRouter } from 'vue-router'
import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
import { fetchGetTeamActions, fetchGetTeamActionOrigins, fetchSetTeamActions } from '@/api/team'
import { fetchGetPermissionActionList, fetchGetTeamFeaturePackages } from '@/api/system-manage'

interface Props {
  modelValue: boolean
  teamId: string
  teamName: string
}

const props = defineProps<Props>()
const router = useRouter()

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
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const selectedIds = ref<string[]>([])
const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
const derivedActionIds = ref<string[]>([])
const manualActionIds = ref<string[]>([])
const derivedSourceMap = ref<Record<string, string[]>>({})
const selectedDerivedPackageId = ref('')
const fromCache = ref(false)
const derivedActions = computed(() => {
  const idSet = new Set(derivedActionIds.value)
  return permissionActions.value.filter((item) => idSet.has(item.id))
})
const manualActions = computed(() => {
  const idSet = new Set(manualActionIds.value)
  return permissionActions.value.filter((item) => idSet.has(item.id))
})
const derivedSourcePackages = computed(() => {
  const packageIdSet = new Set(Object.values(derivedSourceMap.value).flat())
  return featurePackages.value.filter((item) => packageIdSet.has(item.id))
})
const filteredDerivedActions = computed(() => {
  if (!selectedDerivedPackageId.value) return derivedActions.value
  return derivedActions.value.filter((item) => (derivedSourceMap.value[item.id] || []).includes(selectedDerivedPackageId.value))
})
const selectedDerivedPackage = computed(
  () => featurePackages.value.find((item) => item.id === selectedDerivedPackageId.value) || null
)

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      loadData()
    }
  }
)

async function loadData() {
  if (!props.teamId) return
  loading.value = true
  try {
    const [actionsRes, currentRes, packageRes, originsRes] = await Promise.all([
      fetchGetPermissionActionList({ current: 1, size: 1000, status: 'normal', contextType: 'team' }),
      fetchGetTeamActions(props.teamId),
      fetchGetTeamFeaturePackages(props.teamId),
      fetchGetTeamActionOrigins(props.teamId)
    ])

    permissionActions.value = actionsRes?.records || []
    selectedIds.value = [...(currentRes?.actionIds || [])]
    featurePackages.value = packageRes?.packages || []
    derivedActionIds.value = [...(originsRes?.derivedActionIds || [])]
    manualActionIds.value = [...(originsRes?.manualActionIds || [])]
    derivedSourceMap.value = Object.fromEntries(
      (originsRes?.derivedSources || []).map((item) => [item.actionId, item.packageIds])
    )
    selectedDerivedPackageId.value = ''
    fromCache.value = Boolean(originsRes?.fromCache)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载团队补充权限失败')
  } finally {
    loading.value = false
  }
}

function handleCancel() {
  visible.value = false
}

async function handleSave() {
  if (!props.teamId) return
  saving.value = true
  try {
    await fetchSetTeamActions(props.teamId, expandSelectedValues(selectedIds.value, permissionActions.value))
    ElMessage.success('团队补充权限已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存团队补充权限失败')
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
    const featureKey = `${action.featureKind || 'business'}`
    const moduleKey = `${action.moduleCode || action.resourceCode || 'default'}`
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

function buildDerivedSourceText(actionId: string) {
  const packageIdSet = new Set(derivedSourceMap.value[actionId] || [])
  const names = featurePackages.value.filter((item) => packageIdSet.has(item.id)).map((item) => item.name)
  return names.length ? `来源功能包：${names.join('、')}` : '来源功能包未命名'
}

function goToFeaturePackagePage(item: Api.SystemManage.FeaturePackageItem) {
  router.push({
    name: 'FeaturePackage',
    query: {
      packageKey: item.packageKey,
      contextType: item.contextType || 'team',
      open: 'actions'
    }
  })
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

.summary-card {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
}

.package-title {
  color: #475569;
  font-size: 13px;
}

.package-help {
  color: #64748b;
  font-size: 12px;
  line-height: 1.5;
}

.package-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.source-note {
  font-size: 12px;
  color: #64748b;
}

.source-detail-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
}

.source-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
  background: #fff;
}

.source-card--derived {
  border-color: #f3d38a;
  background: #fffaf0;
}

.source-card--manual {
  border-color: #bfd3ff;
  background: #f5f9ff;
}

.source-title {
  font-size: 13px;
  color: #475569;
}

.source-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.source-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.package-filter-tag {
  cursor: pointer;
}
</style>
