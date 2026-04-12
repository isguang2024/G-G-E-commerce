<template>
  <ElDrawer
    v-model="visible"
    :title="`开通协作空间 - ${packageName}`"
    size="920px"
    destroy-on-close
    direction="rtl"
    class="config-drawer"
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        功能包可绑定多个 App；未绑定 App 时，对所有 App 生效。这里仅维护该功能包在协作空间侧的开通关系。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>功能包 {{ packageName }}</ElTag>
        <ElTag type="success" effect="plain" round
          >已开通 {{ selectedCollaborationWorkspaceIds.length }}</ElTag
        >
        <ElTag type="info" effect="plain" round
          >协作空间总数 {{ collaborationWorkspaces.length }}</ElTag
        >
      </div>

      <ElInput
        v-model="keyword"
        clearable
        placeholder="搜索协作空间名称或备注"
        class="toolbar-search"
      />

      <ElTable :data="pagedCollaborationWorkspaces" border max-height="420">
        <ElTableColumn width="60">
          <template #default="{ row }">
            <ElCheckbox
              :model-value="selectedCollaborationWorkspaceIds.includes(row.id)"
              @change="toggleSelection(row.id, $event)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="name" label="协作空间名称" min-width="180" show-overflow-tooltip />
        <ElTableColumn prop="remark" label="备注" min-width="180" show-overflow-tooltip />
        <ElTableColumn prop="plan" label="套餐" width="100" />
        <ElTableColumn label="状态" width="100">
          <template #default="{ row }">
            <ElTag :type="row.status === 'active' ? 'success' : 'warning'">
              {{ row.status === 'active' ? '正常' : '停用' }}
            </ElTag>
          </template>
        </ElTableColumn>
      </ElTable>
      <WorkspacePagination
        v-model:current-page="pagination.current"
        v-model:page-size="pagination.size"
        :total="filteredCollaborationWorkspaces.length"
        compact
      />
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import WorkspacePagination from '@/components/business/tables/WorkspacePagination.vue'
  import {
    fetchGetFeaturePackageCollaborationWorkspaces,
    fetchGetCollaborationWorkspaceOptions,
    fetchSetFeaturePackageCollaborationWorkspaces
  } from '@/domains/governance/api'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    workspaceScope?: 'all' | 'personal' | 'collaboration' | string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    packageId: '',
    packageName: '',
    workspaceScope: 'all'
  })

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
  const keyword = ref('')
  const collaborationWorkspaces = ref<Api.SystemManage.CollaborationWorkspaceListItem[]>([])
  const selectedCollaborationWorkspaceIds = ref<string[]>([])
  const pagination = ref({
    current: 1,
    size: 10
  })

  const filteredCollaborationWorkspaces = computed(() => {
    const currentKeyword = keyword.value.trim().toLowerCase()
    if (!currentKeyword) return collaborationWorkspaces.value
    return collaborationWorkspaces.value.filter((item) =>
      [item.name, item.remark].filter(Boolean).join(' ').toLowerCase().includes(currentKeyword)
    )
  })

  const pagedCollaborationWorkspaces = computed(() => {
    const start = (pagination.value.current - 1) * pagination.value.size
    return filteredCollaborationWorkspaces.value.slice(start, start + pagination.value.size)
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) loadData()
    }
  )

  async function loadData() {
    if (!props.packageId) return
    loading.value = true
    try {
      const [collaborationWorkspaceRes, bindingRes] = await Promise.all([
        fetchGetCollaborationWorkspaceOptions(),
        fetchGetFeaturePackageCollaborationWorkspaces(props.packageId)
      ])
      collaborationWorkspaces.value = collaborationWorkspaceRes?.records || []
      selectedCollaborationWorkspaceIds.value = [...(bindingRes?.collaboration_workspace_ids || [])]
      pagination.value.current = 1
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包协作空间失败')
    } finally {
      loading.value = false
    }
  }

  function toggleSelection(collaborationWorkspaceId: string, checked: boolean | string | number) {
    if (checked) {
      if (!selectedCollaborationWorkspaceIds.value.includes(collaborationWorkspaceId)) {
        selectedCollaborationWorkspaceIds.value = [
          ...selectedCollaborationWorkspaceIds.value,
          collaborationWorkspaceId
        ]
      }
      return
    }
    selectedCollaborationWorkspaceIds.value = selectedCollaborationWorkspaceIds.value.filter(
      (item) => item !== collaborationWorkspaceId
    )
  }

  async function handleSave() {
    if (!props.packageId) return
    saving.value = true
    try {
      const stats = await fetchSetFeaturePackageCollaborationWorkspaces(
        props.packageId,
        selectedCollaborationWorkspaceIds.value
      )
      ElMessage.success(formatRefreshMessage(stats))
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存功能包协作空间失败')
    } finally {
      saving.value = false
    }
  }

  watch(keyword, () => {
    pagination.value.current = 1
  })

  function formatRefreshMessage(stats?: Api.SystemManage.RefreshStats) {
    return `本次增量刷新：角色 ${stats?.roleCount || 0}、协作空间 ${stats?.collaborationWorkspaceCount || 0}、用户 ${stats?.userCount || 0}、耗时 ${stats?.elapsedMilliseconds || 0} ms`
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

  .toolbar-search {
    width: 320px;
  }
</style>
