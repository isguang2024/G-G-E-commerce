<template>
  <div class="preview-shell" v-loading="loading">
    <div class="summary-card">
      <ElTag effect="plain" round>直绑 {{ selectedPackages.length }}</ElTag>
      <ElTag type="warning" effect="plain" round>展开基础包 {{ expandedBasePackages.length }}</ElTag>
      <ElTag type="success" effect="plain" round>展开权限 {{ actionPreviewItems.length }}</ElTag>
      <ElTag type="info" effect="plain" round>展开菜单 {{ menuPreviewItems.length }}</ElTag>
    </div>

    <div class="preview-grid">
      <div class="preview-card">
        <div class="preview-title">当前授予包</div>
        <div class="preview-tags">
          <ElTag
            v-for="item in selectedPackages"
            :key="item.id"
            :type="item.packageType === 'bundle' ? 'warning' : 'success'"
            effect="plain"
            round
            :title="`${formatPackageType(item.packageType)} · ${item.packageKey}`"
          >
            {{ item.name }}
          </ElTag>
        </div>
        <div v-if="!selectedPackages.length" class="preview-empty">当前尚未选择功能包。</div>
      </div>

      <div class="preview-card">
        <div class="preview-title">组合包展开后的基础包</div>
        <div class="preview-tags">
          <ElTag
            v-for="item in expandedBasePackages"
            :key="item.id"
            type="warning"
            effect="plain"
            round
            :title="buildBasePackageTitle(item.id)"
          >
            {{ item.name }}
          </ElTag>
        </div>
        <div v-if="!expandedBasePackages.length" class="preview-empty">当前没有展开出的基础包。</div>
      </div>
    </div>

    <div class="preview-grid">
      <div class="preview-card">
        <div class="preview-title">展开权限预览</div>
        <div class="preview-tags">
          <ElTag
            v-for="item in visibleActionItems"
            :key="item.id"
            type="success"
            effect="plain"
            round
            :title="buildActionTitle(item.id)"
          >
            {{ item.label }}
          </ElTag>
        </div>
        <div v-if="actionPreviewItems.length > previewLimit" class="preview-more">
          还有 {{ actionPreviewItems.length - previewLimit }} 项权限未展开显示
        </div>
        <div v-if="!actionPreviewItems.length" class="preview-empty">当前没有展开出的权限。</div>
      </div>

      <div class="preview-card">
        <div class="preview-title">展开菜单预览</div>
        <div class="preview-tags">
          <ElTag
            v-for="item in visibleMenuItems"
            :key="item.id"
            type="info"
            effect="plain"
            round
            :title="buildMenuTitle(item.id)"
          >
            {{ item.label }}
          </ElTag>
        </div>
        <div v-if="menuPreviewItems.length > previewLimit" class="preview-more">
          还有 {{ menuPreviewItems.length - previewLimit }} 项菜单未展开显示
        </div>
        <div v-if="!menuPreviewItems.length" class="preview-empty">当前没有展开出的菜单。</div>
      </div>
    </div>
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
    selectedPackageIds: string[]
    packages: Api.SystemManage.FeaturePackageItem[]
  }

  const props = defineProps<Props>()

  const loading = ref(false)
  const previewLimit = 24
  const detailMap = ref<Record<string, PackageDetail>>({})
  const knownPackages = ref<Record<string, Api.SystemManage.FeaturePackageItem>>({})

  const selectedPackages = computed(() =>
    props.selectedPackageIds
      .map((id) => knownPackages.value[id] || props.packages.find((item) => item.id === id))
      .filter(Boolean) as Api.SystemManage.FeaturePackageItem[]
  )

  const expandedBasePackageEntries = computed(() => {
    const map = new Map<string, { item: Api.SystemManage.FeaturePackageItem; sourceNames: string[] }>()
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

  const expandedBasePackages = computed(() => expandedBasePackageEntries.value.map((item) => item.item))

  const actionPreviewEntries = computed(() => {
    const map = new Map<string, { id: string; label: string; sourceNames: string[] }>()
    expandedBasePackageEntries.value.forEach((entry) => {
      const detail = detailMap.value[entry.item.id]
      ;(detail?.actionItems || []).forEach((action) => {
        const current = map.get(action.id) || { id: action.id, label: action.label, sourceNames: [] }
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

  const actionPreviewItems = computed(() => actionPreviewEntries.value.map(({ id, label }) => ({ id, label })))
  const menuPreviewItems = computed(() => menuPreviewEntries.value.map(({ id, label }) => ({ id, label })))
  const visibleActionItems = computed(() => actionPreviewItems.value.slice(0, previewLimit))
  const visibleMenuItems = computed(() => menuPreviewItems.value.slice(0, previewLimit))

  watch(
    () => [props.selectedPackageIds.join(','), props.packages.map((item) => item.id).join(',')],
    () => {
      syncKnownPackages()
      loadDetails()
    },
    { immediate: true }
  )

  function syncKnownPackages() {
    const next = { ...knownPackages.value }
    props.packages.forEach((item) => {
      next[item.id] = item
    })
    knownPackages.value = next
  }

  async function loadDetails() {
    if (!props.selectedPackageIds.length) {
      return
    }
    loading.value = true
    try {
      for (const packageId of props.selectedPackageIds) {
        await ensurePackageDetail(packageId)
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

  function buildActionTitle(actionId: string) {
    const target = actionPreviewEntries.value.find((item) => item.id === actionId)
    if (!target) return ''
    return `来源基础包：${target.sourceNames.join('、')}`
  }

  function buildMenuTitle(menuId: string) {
    const target = menuPreviewEntries.value.find((item) => item.id === menuId)
    if (!target) return ''
    return `来源基础包：${target.sourceNames.join('、')}`
  }

  function formatPackageType(packageType?: string) {
    return packageType === 'bundle' ? '组合包' : '基础包'
  }
</script>

<style scoped lang="scss">
  .preview-shell {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .summary-card,
  .preview-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .preview-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }

  .preview-card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    padding: 12px 14px;
    border-radius: 12px;
    border: 1px solid #e5e7eb;
    background: #fff;
  }

  .preview-title {
    color: #475569;
    font-size: 13px;
  }

  .preview-empty,
  .preview-more {
    color: #94a3b8;
    font-size: 12px;
    line-height: 1.6;
  }

  @media (max-width: 960px) {
    .preview-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
