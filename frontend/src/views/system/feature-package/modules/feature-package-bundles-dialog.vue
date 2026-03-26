<template>
  <ElDrawer
    v-model="visible"
    :title="`组合包基础包配置 - ${packageName}`"
    size="920px"
    destroy-on-close
    direction="rtl"
    class="config-drawer">
    <div class="dialog-shell" v-loading="loading">
      <div class="dialog-note">
        组合包只聚合基础包，不直接维护功能范围和菜单。这里保存的是组合包展开时要包含的基础包集合。
      </div>

      <div class="summary-card">
        <ElTag effect="plain" round>组合包 {{ packageName }}</ElTag>
        <ElTag type="warning" effect="plain" round>上下文 {{ contextLabel }}</ElTag>
        <ElTag type="success" effect="plain" round>已选 {{ selectedPackageIds.length }}</ElTag>
        <ElTag type="info" effect="plain" round>可选 {{ filteredPackages.length }}</ElTag>
      </div>

      <ElInput
        v-model="keyword"
        clearable
        placeholder="搜索基础包名称、编码或描述"
        class="toolbar-search"
      />

      <ElTable :data="filteredPackages" border max-height="440">
        <ElTableColumn width="60">
          <template #default="{ row }">
            <ElCheckbox
              :model-value="selectedPackageIds.includes(row.id)"
              @change="toggleSelection(row.id, $event)"
            />
          </template>
        </ElTableColumn>
        <ElTableColumn prop="packageKey" label="基础包编码" min-width="220" show-overflow-tooltip />
        <ElTableColumn prop="name" label="基础包名称" min-width="180" show-overflow-tooltip />
        <ElTableColumn label="上下文" width="120">
          <template #default="{ row }">
            <ElTag
              effect="plain"
              :type="row.contextType === 'platform' ? 'success' : row.contextType === 'team' ? 'info' : 'warning'"
            >
              {{ formatContextType(row.contextType) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="100">
          <template #default="{ row }">
            <ElTag :type="row.status === 'normal' ? 'success' : 'warning'">
              {{ row.status === 'normal' ? '正常' : '停用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="description" label="描述" min-width="220" show-overflow-tooltip />
      </ElTable>
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
  import {
    fetchGetFeaturePackageChildren,
    fetchGetFeaturePackageList,
    fetchSetFeaturePackageChildren
  } from '@/api/system-manage'

  interface Props {
    modelValue: boolean
    packageId: string
    packageName: string
    contextType?: 'platform' | 'team' | 'common' | string
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
  const basePackages = ref<Api.SystemManage.FeaturePackageItem[]>([])
  const selectedPackageIds = ref<string[]>([])

  const filteredPackages = computed(() => {
    const currentKeyword = keyword.value.trim().toLowerCase()
    return basePackages.value
      .filter((item) => item.id !== props.packageId)
      .filter((item) => supportsChildPackage(props.contextType || 'team', item.contextType || 'team'))
      .filter((item) => {
        if (!currentKeyword) return true
        return [item.packageKey, item.name, item.description]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
          .includes(currentKeyword)
      })
  })

  const contextLabel = computed(() => formatContextType(props.contextType))

  watch(
    () => props.modelValue,
    (open) => {
      if (open) loadData()
    }
  )

  async function loadData() {
    if (!props.packageId) return
    loading.value = true
    keyword.value = ''
    try {
      const [packageRes, bundleRes] = await Promise.all([
        fetchGetFeaturePackageList({
          current: 1,
          size: 1000,
          packageType: 'base'
        }),
        fetchGetFeaturePackageChildren(props.packageId)
      ])
      basePackages.value = packageRes?.records || []
      selectedPackageIds.value = [...(bundleRes?.child_package_ids || [])]
    } catch (error: any) {
      ElMessage.error(error?.message || '加载组合包基础包失败')
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

  async function handleSave() {
    if (!props.packageId) return
    saving.value = true
    try {
      await fetchSetFeaturePackageChildren(props.packageId, selectedPackageIds.value)
      ElMessage.success('组合包基础包已保存')
      emit('success')
      visible.value = false
    } catch (error: any) {
      ElMessage.error(error?.message || '保存组合包基础包失败')
    } finally {
      saving.value = false
    }
  }

  function supportsChildPackage(bundleContextType: string, childContextType: string) {
    if (bundleContextType === 'platform') {
      return childContextType === 'platform' || childContextType === 'common'
    }
    if (bundleContextType === 'team') {
      return childContextType === 'team' || childContextType === 'common'
    }
    if (bundleContextType === 'common') {
      return ['platform', 'team', 'common'].includes(childContextType)
    }
    return false
  }

  function formatContextType(contextType?: string) {
    if (contextType === 'platform') return '平台'
    if (contextType === 'team') return '团队'
    if (contextType === 'common') return '通用'
    return contextType || '-'
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
