<template>
  <div class="extra-schema-editor">
    <div class="extra-schema-toolbar">
      <div>
        <div class="extra-schema-title">{{ title }}</div>
        <div class="extra-schema-tip">
          版本固定为 <code>v1</code>。字段键名需唯一，且不能和运行时显式字段冲突。
        </div>
      </div>
      <ElButton type="primary" link @click="addField()">新增字段</ElButton>
    </div>

    <div v-if="draft.fields.length === 0" class="extra-schema-empty">
      暂未配置自定义参数。需要驱动扩展参数、回调上下文或业务附加字段时再添加。
    </div>

    <div v-else class="extra-schema-list">
      <div v-for="(field, index) in draft.fields" :key="field.uid" class="extra-schema-card">
        <div class="extra-schema-card-head">
          <div class="extra-schema-card-title">字段 {{ index + 1 }}</div>
          <ElButton type="danger" link @click="removeField(index)">删除</ElButton>
        </div>

        <ElForm label-width="96px">
          <ElFormItem label="字段键">
            <ElInput v-model="field.key" placeholder="如 callback_scene" />
          </ElFormItem>
          <ElFormItem label="显示名">
            <ElInput v-model="field.label" placeholder="留空将自动回退为字段键" />
          </ElFormItem>
          <ElFormItem label="字段类型">
            <ElSelect v-model="field.type" style="width: 220px" @change="onFieldTypeChange(field)">
              <ElOption label="字符串" value="string" />
              <ElOption label="数字" value="number" />
              <ElOption label="布尔" value="boolean" />
              <ElOption label="对象(JSON)" value="object" />
              <ElOption label="下拉选择" value="select" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="必填">
            <ElSwitch v-model="field.required" />
          </ElFormItem>
          <ElFormItem label="占位提示">
            <ElInput v-model="field.placeholder" placeholder="如 editor / 3600 / true" />
          </ElFormItem>
          <ElFormItem label="说明">
            <ElInput
              v-model="field.description"
              type="textarea"
              :autosize="{ minRows: 2, maxRows: 4 }"
              placeholder="告诉配置人员这个参数的用途和取值建议"
            />
          </ElFormItem>

          <ElFormItem label="默认值">
            <ElInput
              v-if="field.type === 'string' || field.type === 'select'"
              v-model="field.defaultString"
              placeholder="留空表示不设置默认值"
            />
            <ElInputNumber
              v-else-if="field.type === 'number'"
              v-model="field.defaultNumber"
              controls-position="right"
              style="width: 240px"
            />
            <ElSelect
              v-else-if="field.type === 'boolean'"
              v-model="field.defaultBoolean"
              style="width: 220px"
            >
              <ElOption label="不设置" value="" />
              <ElOption label="true" value="true" />
              <ElOption label="false" value="false" />
            </ElSelect>
            <ElInput
              v-else
              v-model="field.defaultObjectText"
              type="textarea"
              :autosize="{ minRows: 3, maxRows: 6 }"
              placeholder='{"scene":"editor"}'
            />
          </ElFormItem>

          <template v-if="field.type === 'select'">
            <div class="extra-schema-options-head">
              <div class="extra-schema-options-title">可选项</div>
              <ElButton type="primary" link @click="addOption(field)">新增选项</ElButton>
            </div>
            <div v-if="field.options.length === 0" class="extra-schema-options-empty">
              下拉字段至少配置一个选项。
            </div>
            <div
              v-for="(option, optionIndex) in field.options"
              :key="option.uid"
              class="extra-schema-option-row"
            >
              <ElInput v-model="option.value" placeholder="值，如 image" />
              <ElInput v-model="option.label" placeholder="显示名，如 图片" />
              <ElButton type="danger" link @click="removeOption(field, optionIndex)">删除</ElButton>
            </div>
          </template>
        </ElForm>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { reactive, watch } from 'vue'
  import {
    ElButton,
    ElForm,
    ElFormItem,
    ElInput,
    ElInputNumber,
    ElOption,
    ElSelect,
    ElSwitch
  } from 'element-plus'

  type ExtraSchemaFieldType = 'string' | 'number' | 'boolean' | 'object' | 'select'

  interface SchemaOptionDraft {
    uid: string
    label: string
    value: string
  }

  interface SchemaFieldDraft {
    uid: string
    key: string
    label: string
    type: ExtraSchemaFieldType
    required: boolean
    placeholder: string
    description: string
    defaultString: string
    defaultNumber?: number
    defaultBoolean: '' | 'true' | 'false'
    defaultObjectText: string
    options: SchemaOptionDraft[]
  }

  const props = withDefaults(
    defineProps<{
      modelValue?: Record<string, unknown>
      title?: string
    }>(),
    {
      title: '自定义参数'
    }
  )

  const draft = reactive({
    version: 'v1',
    fields: [] as SchemaFieldDraft[]
  })

  let uidSeed = 0
  function nextUid(prefix: string) {
    uidSeed += 1
    return `${prefix}-${uidSeed}`
  }

  function createEmptyOption(): SchemaOptionDraft {
    return {
      uid: nextUid('option'),
      label: '',
      value: ''
    }
  }

  function createEmptyField(type: ExtraSchemaFieldType = 'string'): SchemaFieldDraft {
    return {
      uid: nextUid('field'),
      key: '',
      label: '',
      type,
      required: false,
      placeholder: '',
      description: '',
      defaultString: '',
      defaultNumber: undefined,
      defaultBoolean: '',
      defaultObjectText: '',
      options: type === 'select' ? [createEmptyOption()] : []
    }
  }

  function parseDraftField(source: Record<string, unknown>): SchemaFieldDraft {
    const type = normalizeFieldType(source.type)
    return {
      uid: nextUid('field'),
      key: typeof source.key === 'string' ? source.key : '',
      label: typeof source.label === 'string' ? source.label : '',
      type,
      required: source.required === true,
      placeholder: typeof source.placeholder === 'string' ? source.placeholder : '',
      description: typeof source.description === 'string' ? source.description : '',
      defaultString:
        type === 'string' || type === 'select'
          ? typeof source.default_value === 'string'
            ? source.default_value
            : ''
          : '',
      defaultNumber:
        type === 'number' && typeof source.default_value === 'number'
          ? source.default_value
          : undefined,
      defaultBoolean:
        type === 'boolean'
          ? source.default_value === true
            ? 'true'
            : source.default_value === false
              ? 'false'
              : ''
          : '',
      defaultObjectText:
        type === 'object' && source.default_value && typeof source.default_value === 'object'
          ? JSON.stringify(source.default_value, null, 2)
          : '',
      options:
        type === 'select' && Array.isArray(source.options)
          ? source.options.map((item) => ({
              uid: nextUid('option'),
              label:
                typeof item === 'object' && item && 'label' in item ? String(item.label ?? '') : '',
              value:
                typeof item === 'object' && item && 'value' in item ? String(item.value ?? '') : ''
            }))
          : [createEmptyOption()]
    }
  }

  function normalizeFieldType(value: unknown): ExtraSchemaFieldType {
    switch (value) {
      case 'number':
      case 'boolean':
      case 'object':
      case 'select':
        return value
      default:
        return 'string'
    }
  }

  function hydrateFromValue(value?: Record<string, unknown>) {
    draft.version =
      typeof value?.version === 'string' && value.version.trim() ? value.version : 'v1'
    draft.fields = Array.isArray(value?.fields)
      ? value.fields
          .filter((item): item is Record<string, unknown> => !!item && typeof item === 'object')
          .map((item) => parseDraftField(item))
      : []
  }

  function addField(type: ExtraSchemaFieldType = 'string') {
    draft.fields.push(createEmptyField(type))
  }

  function removeField(index: number) {
    draft.fields.splice(index, 1)
  }

  function addOption(field: SchemaFieldDraft) {
    field.options.push(createEmptyOption())
  }

  function removeOption(field: SchemaFieldDraft, index: number) {
    field.options.splice(index, 1)
    if (field.options.length === 0) {
      field.options.push(createEmptyOption())
    }
  }

  function onFieldTypeChange(field: SchemaFieldDraft) {
    field.defaultString = ''
    field.defaultNumber = undefined
    field.defaultBoolean = ''
    field.defaultObjectText = ''
    field.options = field.type === 'select' ? [createEmptyOption()] : []
  }

  function buildSchema() {
    if (draft.fields.length === 0) {
      return { value: undefined as Record<string, unknown> | undefined }
    }

    const seenKeys = new Set<string>()
    const fields: Record<string, unknown>[] = []

    for (const field of draft.fields) {
      const key = field.key.trim()
      if (!key) {
        return { error: '自定义参数字段键不能为空' }
      }
      if (seenKeys.has(key)) {
        return { error: `自定义参数字段键 ${key} 重复` }
      }
      seenKeys.add(key)

      const item: Record<string, unknown> = {
        key,
        label: field.label.trim() || key,
        type: field.type,
        required: field.required
      }
      if (field.placeholder.trim()) item.placeholder = field.placeholder.trim()
      if (field.description.trim()) item.description = field.description.trim()

      if (field.type === 'string' || field.type === 'select') {
        const defaultString = field.defaultString.trim()
        if (defaultString) item.default_value = defaultString
      } else if (field.type === 'number') {
        if (field.defaultNumber !== undefined && Number.isFinite(Number(field.defaultNumber))) {
          item.default_value = Number(field.defaultNumber)
        }
      } else if (field.type === 'boolean') {
        if (field.defaultBoolean === 'true') item.default_value = true
        if (field.defaultBoolean === 'false') item.default_value = false
      } else if (field.type === 'object') {
        const text = field.defaultObjectText.trim()
        if (text) {
          let parsed: unknown
          try {
            parsed = JSON.parse(text)
          } catch {
            return { error: `字段 ${key} 的对象默认值不是合法 JSON` }
          }
          if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
            return { error: `字段 ${key} 的对象默认值必须是 JSON 对象` }
          }
          item.default_value = parsed
        }
      }

      if (field.type === 'select') {
        const options = field.options
          .map((option) => ({
            label: option.label.trim(),
            value: option.value.trim()
          }))
          .filter((option) => option.label || option.value)
        if (options.length === 0) {
          return { error: `字段 ${key} 需要至少一个下拉选项` }
        }
        if (options.some((option) => !option.value)) {
          return { error: `字段 ${key} 的下拉选项值不能为空` }
        }
        item.options = options.map((option) => ({
          label: option.label || option.value,
          value: option.value
        }))
      }

      fields.push(item)
    }

    return {
      value: {
        version: draft.version,
        fields
      }
    }
  }

  defineExpose({
    buildSchema,
    hydrateFromValue
  })

  watch(
    () => props.modelValue,
    (value) => {
      hydrateFromValue(value)
    },
    { immediate: true, deep: true }
  )
</script>

<style scoped>
  .extra-schema-editor {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .extra-schema-toolbar {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 12px;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
  }

  .extra-schema-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .extra-schema-tip {
    margin-top: 4px;
    font-size: 12px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
  }

  .extra-schema-empty {
    padding: 14px 16px;
    font-size: 13px;
    line-height: 1.7;
    color: var(--el-text-color-secondary);
    background: var(--el-fill-color-blank);
    border: 1px dashed var(--el-border-color);
    border-radius: 10px;
  }

  .extra-schema-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .extra-schema-card {
    padding: 14px 14px 6px;
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
  }

  .extra-schema-card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
  }

  .extra-schema-card-title {
    font-size: 13px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .extra-schema-options-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin: 4px 0 8px;
  }

  .extra-schema-options-title {
    font-size: 12px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .extra-schema-options-empty {
    margin-bottom: 8px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
  }

  .extra-schema-option-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
    gap: 8px;
    margin-bottom: 8px;
  }
</style>
