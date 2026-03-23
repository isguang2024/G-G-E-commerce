<template>
  <ElDialog
    v-model="visible"
    :title="`团队角色功能权限 - ${roleTitle}`"
    width="980px"
    destroy-on-close
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        {{
          props.roleData?.isGlobal
            ? '基础团队角色默认继承当前团队已开通功能包，这里只读展示最终可用角色权限。'
            : '请先为角色绑定功能包，再在角色功能包展开范围内配置角色权限。'
        }}
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>角色 {{ roleTitle }}</ElTag>
        <ElTag :type="props.roleData?.isGlobal ? 'info' : 'success'" effect="plain" round>
          {{ props.roleData?.isGlobal ? '基础角色' : '团队自定义' }}
        </ElTag>
        <ElTag type="warning" effect="plain" round>角色功能包 {{ featurePackages.length }}</ElTag>
        <ElTag type="primary" effect="plain" round>{{ inherited ? '继承团队功能包' : '角色独立功能包' }}</ElTag>
        <ElTag type="warning" effect="plain" round>功能包展开 {{ derivedActionCount }}</ElTag>
        <ElTag type="success" effect="plain" round>可配 {{ actions.length }}</ElTag>
        <ElTag effect="plain" round>已选 {{ selectedIds.length }}</ElTag>
      </div>

      <div v-if="featurePackages.length" class="package-summary-card">
        <div class="source-title">当前角色功能包</div>
        <div class="source-tags">
          <ElTag v-for="item in featurePackages" :key="item.id" type="success" effect="plain" round>
            {{ item.name }}
          </ElTag>
        </div>
      </div>

      <div v-if="derivedActions.length" class="source-detail-grid">
        <div v-if="derivedActions.length" class="source-card source-card--derived">
          <div class="source-header">
            <div class="source-title">功能包展开能力</div>
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
      </div>

      <PermissionActionCascaderPanel
        :actions="actions"
        :selected-ids="selectedIds"
        footer-text="保存后该团队角色只会在当前角色功能包展开范围内生效。"
        @update:selected-ids="selectedIds = $event"
      />
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton v-if="!props.roleData?.isGlobal" type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import PermissionActionCascaderPanel from '@/components/business/permission/PermissionActionCascaderPanel.vue'
  import {
    fetchGetMyTeamRoleActions,
    fetchGetMyTeamRolePackages,
    fetchSetMyTeamRoleActions
  } from '@/api/team'

  interface Props {
    modelValue: boolean
    roleData?: Api.SystemManage.RoleListItem
  }

  const props = defineProps<Props>()
  const router = useRouter()
  const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void; (e: 'success'): void }>()

  const loading = ref(false)
  const saving = ref(false)
  const actions = ref<Api.SystemManage.PermissionActionItem[]>([])
  const selectedIds = ref<string[]>([])
  const derivedActionIds = ref<string[]>([])
  const derivedSourceMap = ref<Record<string, string[]>>({})
  const featurePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedDerivedPackageId = ref('')
  const derivedActionCount = ref(0)
  const inherited = ref(false)
  const derivedActions = computed(() => {
    const idSet = new Set(derivedActionIds.value)
    return actions.value.filter((item) => idSet.has(item.id))
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

  const visible = computed({
    get: () => props.modelValue,
    set: (value) => emit('update:modelValue', value)
  })
  const roleTitle = computed(() => props.roleData?.roleName || '')

  watch(
    () => props.modelValue,
    async (open) => {
      if (!open || !props.roleData?.roleId) return
      loading.value = true
      try {
        const [packagesRes, selected] = await Promise.all([
          fetchGetMyTeamRolePackages(props.roleData.roleId),
          fetchGetMyTeamRoleActions(props.roleData.roleId)
        ])
        actions.value = selected?.actions || []
        selectedIds.value = [...(selected?.action_ids || [])]
        derivedActionIds.value = [...(selected?.available_action_ids || [])]
        derivedSourceMap.value = Object.fromEntries(
          (selected?.derived_sources || []).map((item) => [item.action_id, item.package_ids])
        )
        featurePackages.value = packagesRes?.packages || []
        selectedDerivedPackageId.value = ''
        inherited.value = Boolean(packagesRes?.inherited)
        derivedActionCount.value = selected?.available_action_ids?.length || 0
      } catch (error: any) {
        ElMessage.error(error?.message || '加载团队角色功能权限失败')
      } finally {
        loading.value = false
      }
    }
  )

  async function handleSave() {
    if (!props.roleData?.roleId) return
    saving.value = true
    try {
      await fetchSetMyTeamRoleActions(props.roleData.roleId, selectedIds.value)
      ElMessage.success('团队角色功能权限已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队角色功能权限失败')
    } finally {
      saving.value = false
    }
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

  .source-detail-grid {
    display: grid;
    gap: 12px;
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

  .package-summary-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 12px;
    border: 1px solid #d6e8d8;
    background: #f7fcf7;
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
