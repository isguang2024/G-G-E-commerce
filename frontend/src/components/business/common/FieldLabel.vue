<template>
  <span class="field-label">
    <span v-if="required" class="field-label__required">*</span>
    <span class="field-label__text">{{ label }}</span>
    <ElTooltip
      v-if="help"
      :content="help"
      placement="top"
      :show-after="120"
      popper-class="field-label-tooltip"
      :raw-content="rawContent"
    >
      <span class="field-label__icon">
        <ElIcon><QuestionFilled /></ElIcon>
      </span>
    </ElTooltip>
  </span>
</template>

<script setup lang="ts">
  import { ElIcon, ElTooltip } from 'element-plus'
  import { QuestionFilled } from '@element-plus/icons-vue'

  defineOptions({ name: 'FieldLabel' })

  withDefaults(
    defineProps<{
      label: string
      help?: string
      required?: boolean
      rawContent?: boolean
    }>(),
    {
      help: '',
      required: false,
      rawContent: false
    }
  )
</script>

<style scoped lang="scss">
  .field-label {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    line-height: 1;
  }

  .field-label__required {
    color: var(--el-color-danger);
    margin-right: 2px;
  }

  .field-label__icon {
    display: inline-flex;
    align-items: center;
    color: var(--el-text-color-secondary);
    cursor: help;
    font-size: 14px;

    &:hover {
      color: var(--el-color-primary);
    }
  }
</style>

<style lang="scss">
  .field-label-tooltip {
    max-width: 320px;
    line-height: 1.6;
  }
</style>
