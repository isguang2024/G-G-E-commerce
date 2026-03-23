<template>
  <ElDialog
    v-model="visible"
    :title="`团队功能包 - ${teamName}`"
    width="920px"
    destroy-on-close
  >
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        团队功能包由平台统一开通。保存后会同步刷新该团队的功能权限边界。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>团队 {{ teamName }}</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ selectedPackageIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>总计 {{ packages.length }}</ElTag>
      </div>

      <ElInput
        v-model="keyword"
        clearable
        placeholder="搜索功能包名称或编码"
        class="toolbar-search"
      />

      <ElTable :data="filteredPackages" border max-height="420">
        <ElTableColumn width="60">
          <template #default="{ row }">
            <ElCheckbox
              :model-value="selectedPackageIds.includes(row.id)"
              @change="toggleSelection(row.id, $event)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="packageKey" label="功能包编码" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="name" label="功能包名称" min-width="160" show-overflow-tooltip />
        <ElTableColumn label="上下文" width="100">
          <template #default="{ row }">
            <ElTag
              :type="
                row.contextType === 'platform'
                  ? 'warning'
                  : row.contextType === 'platform,team'
                    ? 'info'
                    : 'success'
              "
            >
              {{ formatContext(row.contextType) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="描述" min-width="220" show-overflow-tooltip />
      </ElTable>

      <FeaturePackageGrantPreview
        :selected-package-ids="selectedPackageIds"
        :packages="packages"
      />
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
  import FeaturePackageGrantPreview from '@/components/business/permission/FeaturePackageGrantPreview.vue'
  import {
    fetchGetFeaturePackageList,
    fetchGetTeamFeaturePackages,
    fetchSetTeamFeaturePackages
  } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    teamId: string
    teamName: string
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: false,
    teamId: '',
    teamName: ''
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
  const packages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedPackageIds = ref<string[]>([])

  const filteredPackages = computed(() => {
    const currentKeyword = keyword.value.trim().toLowerCase()
    if (!currentKeyword) return packages.value
    return packages.value.filter((item) =>
      [item.packageKey, item.name, item.description]
        .filter(Boolean)
        .join(' ')
        .toLowerCase()
        .includes(currentKeyword)
    )
  })

  watch(
    () => props.modelValue,
    (open) => {
      if (open) loadData()
    }
  )

  async function loadData() {
    if (!props.teamId) return
    loading.value = true
    try {
      const [listRes, teamRes] = await Promise.all([
        fetchGetFeaturePackageList({ current: 1, size: 1000, contextType: 'team', status: 'normal' }),
        fetchGetTeamFeaturePackages(props.teamId)
      ])
      packages.value = listRes?.records || []
      selectedPackageIds.value = [...(teamRes?.package_ids || [])]
    } catch (error: any) {
      ElMessage.error(error?.message || '加载团队功能包失败')
    } finally {
      loading.value = false
    }
  }

  function toggleSelection(packageId: string, checked: boolean | string | number) {
    if (checked) {
      if (!selectedPackageIds.value.includes(packageId)) {
        selectedPackageIds.value = [...selectedPackageIds.value, packageId]
      }
      return
    }
    selectedPackageIds.value = selectedPackageIds.value.filter((item) => item !== packageId)
  }

  function formatContext(contextType?: string) {
    if (contextType === 'platform,team') return '平台/团队'
    if (contextType === 'platform') return '平台'
    return '团队'
  }

  async function handleSave() {
    if (!props.teamId) return
    saving.value = true
    try {
      await fetchSetTeamFeaturePackages(props.teamId, selectedPackageIds.value)
      ElMessage.success('团队功能包已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存团队功能包失败')
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

  .summary-card {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .toolbar-search {
    width: 320px;
  }
</style>
