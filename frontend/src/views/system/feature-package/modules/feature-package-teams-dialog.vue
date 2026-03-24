<template>
  <ElDialog
    v-model="visible"
    :title="`开通团队 - ${packageName}`"
    width="920px"
    destroy-on-close
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        平台为团队开通功能包后，团队功能边界会自动同步；这里支持团队包和平台/团队共享包。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>功能包 {{ packageName }}</ElTag>
        <ElTag type="success" effect="plain" round>已开通 {{ selectedTeamIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>团队总数 {{ teams.length }}</ElTag>
      </div>

      <ElInput
        v-model="keyword"
        clearable
        placeholder="搜索团队名称或备注"
        class="toolbar-search"
      />

      <ElTable :data="filteredTeams" border max-height="420">
        <ElTableColumn width="60">
          <template #default="{ row }">
            <ElCheckbox
              :model-value="selectedTeamIds.includes(row.id)"
              @change="toggleSelection(row.id, $event)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="name" label="团队名称" min-width="180" show-overflow-tooltip />
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
    </div>

    <template #footer>
      <ElButton @click="visible = false">取消</ElButton>
      <ElButton type="primary" :loading="saving" @click="handleSave">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { fetchGetTeamList } from '@/api/team'
  import { fetchGetFeaturePackageTeams, fetchSetFeaturePackageTeams } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    contextType?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    packageId: '',
    packageName: '',
    contextType: 'team'
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
  const teams = ref<Api.SystemManage.TeamListItem[]>([])
  const selectedTeamIds = ref<string[]>([])

  const filteredTeams = computed(() => {
    const currentKeyword = keyword.value.trim().toLowerCase()
    if (!currentKeyword) return teams.value
    return teams.value.filter((item) =>
      [item.name, item.remark].filter(Boolean).join(' ').toLowerCase().includes(currentKeyword)
    )
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
      const [teamRes, bindingRes] = await Promise.all([
        fetchGetTeamList({ current: 1, size: 1000 }),
        fetchGetFeaturePackageTeams(props.packageId)
      ])
      teams.value = teamRes?.records || []
      selectedTeamIds.value = [...(bindingRes?.team_ids || [])]
    } catch (error: any) {
      ElMessage.error(error?.message || '加载功能包团队失败')
    } finally {
      loading.value = false
    }
  }

  function toggleSelection(teamId: string, checked: boolean | string | number) {
    if (checked) {
      if (!selectedTeamIds.value.includes(teamId)) {
        selectedTeamIds.value = [...selectedTeamIds.value, teamId]
      }
      return
    }
    selectedTeamIds.value = selectedTeamIds.value.filter((item) => item !== teamId)
  }

  async function handleSave() {
    if (!props.packageId) return
    if (!supportsTeamContext(props.contextType)) {
      ElMessage.warning('仅团队侧可生效的功能包支持开通团队')
      return
    }
    saving.value = true
    try {
      await fetchSetFeaturePackageTeams(props.packageId, selectedTeamIds.value)
      ElMessage.success('功能包团队已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存功能包团队失败')
    } finally {
      saving.value = false
    }
  }

  function supportsTeamContext(contextType?: string) {
    return contextType === 'team' || contextType === 'common'
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
