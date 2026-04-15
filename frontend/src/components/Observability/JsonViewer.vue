<!--
  JsonViewer.vue — 观察性页面专用的轻量 JSON 树形查看器。

  设计取舍：
   - 不引第三方（vue-json-pretty / vue-json-viewer 各自都有 100KB+ 体积），手写
     一个递归 SFC 即可覆盖审计/遥测抽屉里的 Before / After / Metadata / Payload。
   - 默认展开 1 层；超过 1 层的对象/数组初始折叠，避免一打开就铺一屏。
   - 复制按钮固定在根节点右上角，复制的是格式化（2-space）后的整段 JSON，方便
     运维粘到工单 / 飞书。
   - 不依赖 ElIcon，用 ▸/▾ unicode；保持 0 额外依赖、bundle 体积最小。
-->
<template>
  <div v-if="isRoot" class="json-viewer">
    <div class="json-viewer-actions">
      <ElButton link type="primary" size="small" @click="copyAll">{{ copied ? '已复制' : '复制' }}</ElButton>
    </div>
    <div class="json-viewer-body">
      <JsonViewer :data="data" :is-root="false" :depth="0" />
    </div>
  </div>
  <template v-else>
    <!-- primitives -->
    <span v-if="data === null" class="jv-null">null</span>
    <span v-else-if="data === undefined" class="jv-null">undefined</span>
    <span v-else-if="kind === 'boolean'" class="jv-bool">{{ String(data) }}</span>
    <span v-else-if="kind === 'number'" class="jv-num">{{ data }}</span>
    <span v-else-if="kind === 'string'" class="jv-str">"{{ data }}"</span>

    <!-- container: array / object -->
    <span v-else-if="kind === 'array' || kind === 'object'" class="jv-block">
      <span class="jv-toggle" @click="collapsed = !collapsed">{{ collapsed ? '▸' : '▾' }}</span>
      <span class="jv-summary" @click="collapsed = !collapsed">
        {{ openBracket }}<span v-if="collapsed" class="jv-count"> {{ entryCount }} </span>{{ collapsed ? closeBracket : '' }}
      </span>
      <div v-if="!collapsed" class="jv-children">
        <div v-for="(child, key) in entries" :key="String(key)" class="jv-row">
          <span v-if="kind === 'object'" class="jv-key">"{{ key }}"</span>
          <span v-else class="jv-key jv-index">{{ key }}</span>
          <span class="jv-colon">:</span>
          <JsonViewer :data="child" :is-root="false" :depth="depth + 1" />
        </div>
        <div class="jv-close">{{ closeBracket }}</div>
      </div>
    </span>

    <!-- fallback: function / symbol / etc. -->
    <span v-else class="jv-unknown">{{ String(data) }}</span>
  </template>
</template>

<script setup lang="ts">
  import { computed, ref } from 'vue'
  import { ElButton, ElMessage } from 'element-plus'

  // 自递归引用：必须 defineOptions({name}) 才能在 <template> 里写 <JsonViewer>。
  defineOptions({ name: 'JsonViewer' })

  const props = withDefaults(
    defineProps<{
      data: unknown
      /** 仅根节点为 true，渲染外层框 + 复制按钮；自递归时一律 false。 */
      isRoot?: boolean
      /** 当前嵌套深度。≤1 时默认展开，避免一打开就刷屏。 */
      depth?: number
    }>(),
    { isRoot: true, depth: 0 }
  )

  /** "array" | "object" | "string" | "number" | "boolean" | null/undefined => 走 v-if */
  const kind = computed<'array' | 'object' | 'string' | 'number' | 'boolean' | 'other'>(() => {
    const v = props.data
    if (Array.isArray(v)) return 'array'
    if (v && typeof v === 'object') return 'object'
    if (typeof v === 'string') return 'string'
    if (typeof v === 'number') return 'number'
    if (typeof v === 'boolean') return 'boolean'
    return 'other'
  })

  const entries = computed<Record<string, unknown> | unknown[]>(() => {
    if (kind.value === 'array') return props.data as unknown[]
    if (kind.value === 'object') return props.data as Record<string, unknown>
    return {}
  })

  const entryCount = computed(() => {
    if (kind.value === 'array') return (props.data as unknown[]).length
    if (kind.value === 'object') return Object.keys(props.data as Record<string, unknown>).length
    return 0
  })

  const openBracket = computed(() => (kind.value === 'array' ? '[' : '{'))
  const closeBracket = computed(() => (kind.value === 'array' ? ']' : '}'))

  // 默认展开规则：根节点 + 一层子节点；再深就折叠，等用户按需展开。
  const collapsed = ref(props.depth >= 2)

  // ─── 复制 ────────────────────────────────────────────────────────────────────
  const copied = ref(false)
  let copyTimer: ReturnType<typeof setTimeout> | null = null
  async function copyAll() {
    try {
      const text = JSON.stringify(props.data, null, 2)
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(text)
      } else {
        // fallback：execCommand 路径，兼容老浏览器/非 https 场景
        const ta = document.createElement('textarea')
        ta.value = text
        ta.style.position = 'fixed'
        ta.style.opacity = '0'
        document.body.appendChild(ta)
        ta.select()
        document.execCommand('copy')
        document.body.removeChild(ta)
      }
      copied.value = true
      if (copyTimer) clearTimeout(copyTimer)
      copyTimer = setTimeout(() => (copied.value = false), 1500)
    } catch (e: any) {
      ElMessage.error('复制失败：' + (e?.message || String(e)))
    }
  }
</script>

<style scoped>
  .json-viewer {
    position: relative;
    background: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    padding: 10px 12px 12px;
    max-height: 360px;
    overflow: auto;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono',
      'Courier New', monospace;
    font-size: 12px;
    line-height: 1.55;
    color: var(--el-text-color-primary);
  }

  .json-viewer-actions {
    position: sticky;
    top: -10px;
    margin-top: -2px;
    display: flex;
    justify-content: flex-end;
    background: var(--el-fill-color-light);
    z-index: 1;
  }

  .json-viewer-body {
    margin-top: 4px;
    word-break: break-all;
    white-space: pre-wrap;
  }

  .jv-block {
    display: inline-block;
    vertical-align: top;
  }

  .jv-toggle {
    cursor: pointer;
    user-select: none;
    color: var(--el-text-color-secondary);
    margin-right: 2px;
    width: 12px;
    display: inline-block;
    text-align: center;
  }

  .jv-summary {
    cursor: pointer;
    user-select: none;
    color: var(--el-text-color-secondary);
  }

  .jv-count {
    color: var(--el-text-color-placeholder);
    font-style: italic;
    margin: 0 2px;
  }

  .jv-children {
    margin-left: 14px;
    border-left: 1px dashed var(--el-border-color-lighter);
    padding-left: 8px;
  }

  .jv-row {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .jv-key {
    color: var(--el-color-primary);
  }
  .jv-index {
    color: var(--el-text-color-placeholder);
  }
  .jv-colon {
    color: var(--el-text-color-secondary);
  }

  .jv-close {
    color: var(--el-text-color-secondary);
    margin-left: -6px;
  }

  .jv-null {
    color: var(--el-color-danger);
    font-style: italic;
  }
  .jv-bool {
    color: var(--el-color-warning);
  }
  .jv-num {
    color: var(--el-color-info);
  }
  .jv-str {
    color: var(--el-color-success);
  }
  .jv-unknown {
    color: var(--el-text-color-secondary);
  }
</style>
