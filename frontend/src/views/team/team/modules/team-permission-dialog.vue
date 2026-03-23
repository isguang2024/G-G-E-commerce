<template>
  <ElDialog
    v-model="visible"
    :title="`团队边界 - ${teamName}`"
    width="960px"
    destroy-on-close
    class="team-permission-dialog"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        这里配置的是团队边界减法。正式开通能力请通过功能包完成；这里保存的内容只会从功能包展开结果中屏蔽个别权限，不会额外开通新能力。
      </div>

      <PermissionSummaryTags :items="summaryItems" />

      <div v-if="featurePackages.length" class="package-card">
        <div class="package-title">已开通功能包</div>
        <div class="package-help">功能包决定团队正式开通的基础能力；当前弹窗只负责从这些能力中屏蔽个别权限。</div>
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
        <span>黄色统计表示来自功能包展开的基础能力，蓝色统计表示当前团队边界已屏蔽的权限。</span>
      </div>

      <ElAlert
        v-if="fromCache"
        type="warning"
        :closable="false"
        title="当前团队边界来源暂时从缓存回退读取，建议后续刷新团队边界快照。"
      />

      <PermissionSourcePanels
        v-model="selectedDerivedPackageId"
        :packages="featurePackages"
        :source-map="derivedSourceMap"
        :derived-items="derivedActionItems"
        :blocked-items="blockedActionItems"
        derived-title="功能包展开明细"
        blocked-title="团队边界屏蔽明细"
        open="actions"
        blocked-tag-type="primary"
        filtered-blocked-empty-text="当前筛选下暂无团队边界屏蔽项"
        empty-title="当前暂无团队边界来源"
        empty-text="请先检查团队功能包、团队上下文或团队边界快照是否已经生成。"
      />

      <PermissionActionCascaderPanel
        :actions="permissionActions"
        :selected-ids="selectedIds"
        footer-text="这里只保存团队边界结果；提交时会自动反推出需要屏蔽的权限，最终团队权限 = 功能包展开 - 团队屏蔽。提交前会自动展开父级选择。"
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
import PermissionSourcePanels from '@/components/business/permission/PermissionSourcePanels.vue'
import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
import PermissionSummaryTags from '@/components/business/permission/PermissionSummaryTags.vue'
import { fetchGetTeamActions, fetchGetTeamActionOrigins, fetchSetTeamActions } from '@/api/team'
import { fetchGetPermissionActionList, fetchGetTeamFeaturePackages } from '@/api/system-manage'

interface Props {
  modelValue: boolean
  teamId: string
  teamName: string
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
const permissionActions = ref<Api.SystemManage.PermissionActionItem[]>([])
const selectedIds = ref<string[]>([])
const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
const derivedActionIds = ref<string[]>([])
const blockedActionIds = ref<string[]>([])
const derivedSourceMap = ref<Record<string, string[]>>({})
const selectedDerivedPackageId = ref('')
const fromCache = ref(false)
const derivedActions = computed(() => {
  const idSet = new Set(derivedActionIds.value)
  return permissionActions.value.filter((item) => idSet.has(item.id))
})
const blockedActions = computed(() => {
  const idSet = new Set(blockedActionIds.value)
  return permissionActions.value.filter((item) => idSet.has(item.id))
})
const blockedActionCount = computed(() => blockedActions.value.length)
const summaryItems = computed(() => [
  { label: '团队', value: props.teamName || '-' },
  { label: '边界结果', value: selectedIds.value.length, type: 'success' as const },
  { label: '候选', value: permissionActions.value.length, type: 'info' as const },
  { label: '功能包', value: featurePackages.value.length },
  { label: '功能包展开', value: derivedActionIds.value.length, type: 'warning' as const },
      { label: '边界已屏蔽', value: blockedActionCount.value, type: 'primary' as const }
])
const derivedActionItems = computed(() => derivedActions.value.map((item) => ({ id: item.id, label: item.name })))
const blockedActionItems = computed(() => blockedActions.value.map((item) => ({ id: item.id, label: item.name })))

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
    blockedActionIds.value = [...(originsRes?.blockedActionIds || originsRes?.manualActionIds || [])]
    derivedSourceMap.value = Object.fromEntries(
      (originsRes?.derivedSources || []).map((item) => [item.actionId, item.packageIds])
    )
    selectedDerivedPackageId.value = ''
    fromCache.value = Boolean(originsRes?.fromCache)
  } catch (error: any) {
    ElMessage.error(error?.message || '加载团队边界失败')
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
    ElMessage.success('团队边界已保存')
    emit('success')
    visible.value = false
  } catch (error: any) {
    ElMessage.error(error?.message || '保存团队边界失败')
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

</style>
