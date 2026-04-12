<template>
  <div class="dictionary-page">
    <div class="dictionary-layout">
      <!-- Left: Dict Type Panel -->
      <div class="dict-type-side">
        <DictTypePanel
          :selected-id="selectedTypeId"
          @select="handleSelectType"
          @refresh="handleTypeRefresh"
        />
      </div>

      <!-- Right: Dict Item Panel -->
      <div class="dict-item-main">
        <DictItemPanel
          v-if="selectedType"
          :dict-type="selectedType"
          :key="selectedType.id"
          @type-updated="handleTypeRefresh"
        />
        <div v-else class="dict-item-empty">
          <ElEmpty description="请从左侧选择一个字典类型" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref } from 'vue'
  import type { DictTypeSummary } from '@/api/system-manage/dictionary'
  import DictTypePanel from './modules/dict-type-panel.vue'
  import DictItemPanel from './modules/dict-item-panel.vue'

  const selectedTypeId = ref<string>('')
  const selectedType = ref<DictTypeSummary | null>(null)

  function handleSelectType(type: DictTypeSummary) {
    selectedTypeId.value = type.id
    selectedType.value = type
  }

  function handleTypeRefresh() {
    // Reset selection when types list refreshes
    selectedTypeId.value = ''
    selectedType.value = null
  }
</script>

<style scoped lang="scss">
  .dictionary-page {
    height: 100%;
    padding: 16px;
    box-sizing: border-box;
  }

  .dictionary-layout {
    display: flex;
    height: 100%;
    gap: 16px;
  }

  .dict-type-side {
    width: 360px;
    flex-shrink: 0;
  }

  .dict-item-main {
    flex: 1;
    min-width: 0;
  }

  .dict-item-empty {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--el-bg-color);
    border-radius: 4px;
    border: 1px solid var(--el-border-color-lighter);
  }
</style>
