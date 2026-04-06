<template>
  <div class="preview-shell" v-loading="loading">
    <template v-if="previewMode === 'single'">
      <div class="single-header">
        <div class="single-title">
          <span>{{ activePackage?.name || '功能包详情' }}</span>
          <ElTag
            v-if="activePackage"
            :type="getPackageTypeTagType(activePackage.packageType)"
            effect="light"
            round
          >
            {{ formatPackageType(activePackage.packageType) }}
          </ElTag>
        </div>
        <div v-if="activePackage?.description" class="single-description">
          {{ activePackage.description }}
        </div>
      </div>

      <div class="summary-card">
        <ElTag type="info" effect="light" round>
          {{ activePackage?.packageKey || '-' }}
        </ElTag>
        <ElTag type="info" effect="light" round>
          展开基础包 {{ expandedBasePackages.length }}
        </ElTag>
        <ElTag type="info" effect="light" round> 权限 {{ actionPreviewItems.length }} </ElTag>
        <ElTag type="info" effect="light" round> 菜单 {{ menuPreviewItems.length }} </ElTag>
      </div>

      <div class="preview-grid">
        <div class="preview-card">
          <div class="preview-title">当前功能包</div>
          <div class="preview-tags">
            <ElTag
              v-if="activePackage"
              :type="getPackageTypeTagType(activePackage.packageType)"
              effect="light"
              round
            >
              {{ activePackage.name }}
            </ElTag>
          </div>
        </div>

        <div class="preview-card">
          <div class="preview-title">组合展开后的基础包</div>
          <div class="preview-tags">
            <ElTag
              v-for="item in expandedBasePackages"
              :key="item.id"
              type="info"
              effect="light"
              round
              :title="buildBasePackageTitle(item.id)"
            >
              {{ item.name }}
            </ElTag>
          </div>
          <div v-if="!expandedBasePackages.length" class="preview-empty"
            >当前没有展开出的基础包。</div
          >
        </div>
      </div>
    </template>

    <template v-else>
      <div class="summary-card">
        <ElTag type="info" effect="light" round> 直绑 {{ selectedPackages.length }} </ElTag>
        <ElTag type="info" effect="light" round>
          展开基础包 {{ expandedBasePackages.length }}
        </ElTag>
        <ElTag type="info" effect="light" round> 展开权限 {{ actionPreviewItems.length }} </ElTag>
        <ElTag type="info" effect="light" round> 展开菜单 {{ menuPreviewItems.length }} </ElTag>
      </div>

      <div class="preview-grid">
        <div class="preview-card">
          <div class="preview-title">当前授予包</div>
          <div class="preview-tags">
            <ElTag
              v-for="item in selectedPackages"
              :key="item.id"
              :type="getPackageTypeTagType(item.packageType)"
              effect="light"
              round
              :title="`${formatPackageType(item.packageType)} · ${item.packageKey}`"
            >
              {{ item.name }}
            </ElTag>
          </div>
          <div v-if="!selectedPackages.length" class="preview-empty">当前尚未选择功能包。</div>
        </div>

        <div class="preview-card">
          <div class="preview-title">组合展开后的基础包</div>
          <div class="preview-tags">
            <ElTag
              v-for="item in expandedBasePackages"
              :key="item.id"
              type="info"
              effect="light"
              round
              :title="buildBasePackageTitle(item.id)"
            >
              {{ item.name }}
            </ElTag>
          </div>
          <div v-if="!expandedBasePackages.length" class="preview-empty"
            >当前没有展开出的基础包。</div
          >
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import {
    fetchGetFeaturePackageActions,
    fetchGetFeaturePackageChildren,
    fetchGetFeaturePackageMenus
  } from '@/api/system-manage'

  interface PackageDetail {
    package: Api.SystemManage.FeaturePackageItem
    childPackages: Api.SystemManage.FeaturePackageItem[]
    actionItems: Array<{ id: string; label: string }>
    menuItems: Array<{ id: string; label: string }>
  }

  interface Props {
    packageId?: string
    packageItem?: Api.SystemManage.FeaturePackageItem
    selectedPackageIds?: string[]
    packages?: Api.SystemManage.FeaturePackageItem[]
  }

  const props = withDefaults(defineProps<Props>(), {
    packageId: '',
    packageItem: undefined,
    selectedPackageIds: () => [],
    packages: () => []
  })

  const loading = ref(false)
  const detailMap = ref<Record<string, PackageDetail>>({})
  const knownPackages = ref<Record<string, Api.SystemManage.FeaturePackageItem>>({})

  const previewMode = computed(() => (props.packageId || props.packageItem ? 'single' : 'multiple'))

  const activePackage = computed(() => {
    if (props.packageItem) return props.packageItem
    if (props.packageId) return knownPackages.value[props.packageId] || null
    return null
  })

  const activePackageIds = computed(() => {
    if (previewMode.value === 'single') {
      return activePackage.value?.id ? [activePackage.value.id] : []
    }
    return props.selectedPackageIds
  })

  const selectedPackages = computed(() => {
    if (previewMode.value === 'single') {
      return activePackage.value ? [activePackage.value] : []
    }
    return activePackageIds.value
      .map((id) => knownPackages.value[id] || props.packages.find((item) => item.id === id))
      .filter(Boolean) as Api.SystemManage.FeaturePackageItem[]
  })

  const expandedBasePackageEntries = computed(() => {
    const map = new Map<
      string,
      { item: Api.SystemManage.FeaturePackageItem; sourceNames: string[] }
    >()
    selectedPackages.value.forEach((item) => {
      if (item.packageType === 'bundle') {
        const detail = detailMap.value[item.id]
        ;(detail?.childPackages || []).forEach((child) => {
          const current = map.get(child.id) || { item: child, sourceNames: [] }
          if (!current.sourceNames.includes(item.name)) current.sourceNames.push(item.name)
          map.set(child.id, current)
        })
        return
      }
      const current = map.get(item.id) || { item, sourceNames: [] }
      if (!current.sourceNames.includes(item.name)) current.sourceNames.push(item.name)
      map.set(item.id, current)
    })
    return Array.from(map.values())
  })

  const expandedBasePackages = computed(() =>
    expandedBasePackageEntries.value.map((item) => item.item)
  )

  const actionPreviewEntries = computed(() => {
    const map = new Map<string, { id: string; label: string; sourceNames: string[] }>()
    expandedBasePackageEntries.value.forEach((entry) => {
      const detail = detailMap.value[entry.item.id]
      ;(detail?.actionItems || []).forEach((action) => {
        const current = map.get(action.id) || {
          id: action.id,
          label: action.label,
          sourceNames: []
        }
        entry.sourceNames.forEach((name) => {
          if (!current.sourceNames.includes(name)) current.sourceNames.push(name)
        })
        map.set(action.id, current)
      })
    })
    return Array.from(map.values())
  })

  const menuPreviewEntries = computed(() => {
    const map = new Map<string, { id: string; label: string; sourceNames: string[] }>()
    expandedBasePackageEntries.value.forEach((entry) => {
      const detail = detailMap.value[entry.item.id]
      ;(detail?.menuItems || []).forEach((menu) => {
        const current = map.get(menu.id) || { id: menu.id, label: menu.label, sourceNames: [] }
        entry.sourceNames.forEach((name) => {
          if (!current.sourceNames.includes(name)) current.sourceNames.push(name)
        })
        map.set(menu.id, current)
      })
    })
    return Array.from(map.values())
  })

  const actionPreviewItems = computed(() =>
    actionPreviewEntries.value.map(({ id, label }) => ({ id, label }))
  )
  const menuPreviewItems = computed(() =>
    menuPreviewEntries.value.map(({ id, label }) => ({ id, label }))
  )

  watch(
    () => [
      props.packageId,
      props.packageItem?.id,
      props.selectedPackageIds.join(','),
      props.packages.map((item) => item.id).join(',')
    ],
    () => {
      syncKnownPackages()
      void loadDetails()
    },
    { immediate: true }
  )

  function syncKnownPackages() {
    const next = { ...knownPackages.value }
    props.packages.forEach((item) => {
      next[item.id] = item
    })
    if (props.packageItem?.id) {
      next[props.packageItem.id] = props.packageItem
    }
    knownPackages.value = next
  }

  async function loadDetails() {
    if (!activePackageIds.value.length) return
    loading.value = true
    try {
      for (const targetPackageId of activePackageIds.value) {
        await ensurePackageDetail(targetPackageId)
      }
    } finally {
      loading.value = false
    }
  }

  async function ensurePackageDetail(packageId: string) {
    if (detailMap.value[packageId]) return detailMap.value[packageId]
    const target = knownPackages.value[packageId]
    if (!target) return null

    if (target.packageType === 'bundle') {
      const childRes = await fetchGetFeaturePackageChildren(packageId)
      const childPackages = childRes?.packages || []
      if (childPackages.length) {
        knownPackages.value = {
          ...knownPackages.value,
          ...Object.fromEntries(childPackages.map((item) => [item.id, item]))
        }
      }
      const detail: PackageDetail = {
        package: target,
        childPackages,
        actionItems: [],
        menuItems: []
      }
      detailMap.value = {
        ...detailMap.value,
        [packageId]: detail
      }
      for (const child of childPackages) {
        await ensurePackageDetail(child.id)
      }
      return detail
    }

    const [actionRes, menuRes] = await Promise.all([
      fetchGetFeaturePackageActions(packageId),
      fetchGetFeaturePackageMenus(packageId)
    ])
    const detail: PackageDetail = {
      package: target,
      childPackages: [],
      actionItems: (actionRes?.actions || []).map((item) => ({
        id: item.id,
        label: item.name || item.permissionKey || item.id
      })),
      menuItems: (menuRes?.menus || []).map((item: any) => ({
        id: `${item?.id || ''}`,
        label: item?.meta?.title || item?.title || item?.name || item?.path || '未命名菜单'
      }))
    }
    detailMap.value = {
      ...detailMap.value,
      [packageId]: detail
    }
    return detail
  }

  function buildBasePackageTitle(packageId: string) {
    const target = expandedBasePackageEntries.value.find((item) => item.item.id === packageId)
    if (!target) return ''
    return `来源授予包：${target.sourceNames.join('、')}`
  }

  function formatPackageType(packageType?: string) {
    return packageType === 'bundle' ? '组合包' : '基础包'
  }

  function getPackageTypeTagType(packageType?: string) {
    return packageType === 'bundle' ? 'warning' : 'primary'
  }
</script>

<style scoped lang="scss">
  .preview-shell {
    display: flex;
    flex-direction: column;
    gap: 14px;
  }

  .single-header {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .single-title {
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--art-text-strong);
    font-size: 14px;
    font-weight: 700;
  }

  .single-description {
    color: var(--art-text-muted);
    font-size: 13px;
    line-height: 1.7;
  }

  .summary-card,
  .preview-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 12px;
  }

  .preview-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
  }

  .preview-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 16px 18px;
    border-radius: 16px;
    border: 1px solid color-mix(in srgb, var(--default-border) 82%, white);
    background: linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.94));
    box-shadow: var(--art-shadow-sm);
  }

  .preview-title {
    color: var(--art-text-muted);
    font-size: 13px;
    font-weight: 600;
  }

  .preview-empty {
    color: var(--art-text-soft);
    font-size: 12px;
    line-height: 1.6;
  }

  .summary-card :deep(.el-tag),
  .preview-tags :deep(.el-tag) {
    border-radius: 9999px;
  }

  @media (max-width: 960px) {
    .preview-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
