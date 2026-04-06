<template>
  <span class="page-field-label">
    <span>{{ label }}</span>
    <ElTooltip
      :content="help"
      placement="top"
      :show-after="120"
      popper-class="page-field-label-tooltip"
    >
      <span class="page-field-label-icon">
        <ElIcon><QuestionFilled /></ElIcon>
      </span>
    </ElTooltip>
  </span>
</template>

<script setup lang="ts">
  import { ElIcon, ElTooltip } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'

  defineProps<{
    label: string
    help: string
  }>()
</script>

<style scoped lang="scss">
  .page-field-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    line-height: 1;
  }

  .page-field-label-icon {
    display: inline-flex;
    align-items: center;
    color: var(--el-text-color-secondary);
    cursor: help;
    font-size: 14px;
  }
</style>
