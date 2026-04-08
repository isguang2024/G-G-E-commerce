import { computed, ref, watch } from 'vue'
import {
  buildBatchCommands,
  buildDecisionOptions,
  buildFeatureOptions,
  buildStateOptions,
  getDecisionLabel as formatDecisionLabel,
  getDecisionTagType as formatDecisionTagType,
  getSelectedLabel,
  groupActions,
  moduleKeywords,
  normalizeDecision,
  type DecisionValue,
  type WorkbenchActionItem,
  type WorkbenchMode
} from './permission-action-workbench.helpers'

interface UsePermissionActionWorkbenchProps {
  mode: WorkbenchMode
  actions: WorkbenchActionItem[]
  loading: boolean
  selectedIds: string[]
  decisionMap: Record<string, DecisionValue>
  searchPlaceholder: string
}

type Emit = {
  (e: 'update:selectedIds', value: string[]): void
  (e: 'update:decisionMap', value: Record<string, DecisionValue>): void
}

export function usePermissionActionWorkbench(props: UsePermissionActionWorkbenchProps, emit: Emit) {
  const searchKeyword = ref('')
  const featureFilter = ref('')
  const stateFilter = ref('all')
  const showMeta = ref(false)
  const compactMode = ref(false)
  const activeFeatures = ref<string[]>([])

  function isSelected(actionId: string) {
    return props.selectedIds.includes(actionId)
  }

  function getDecision(actionId: string): DecisionValue {
    return props.decisionMap[actionId] || ''
  }

  function isActive(actionId: string) {
    if (props.mode === 'menu' || props.mode === 'collaboration') {
      return isSelected(actionId)
    }
    return Boolean(getDecision(actionId))
  }

  function matchState(actionId: string) {
    if (props.mode === 'menu' || props.mode === 'collaboration') {
      if (stateFilter.value === 'selected') return isSelected(actionId)
      if (stateFilter.value === 'unselected') return !isSelected(actionId)
      return true
    }

    const decision = getDecision(actionId)
    if (props.mode === 'user') {
      if (stateFilter.value === 'inherit') return !decision
      return stateFilter.value === 'all' ? true : decision === stateFilter.value
    }

    if (stateFilter.value === 'unset') return !decision
    return stateFilter.value === 'all' ? true : decision === stateFilter.value
  }

  const featureOptions = computed(() => buildFeatureOptions(props.actions))

  const filteredActions = computed(() => {
    const keyword = searchKeyword.value.trim().toLowerCase()
    return props.actions.filter((item) => {
      if (
        featureFilter.value &&
        `${item.featureGroupId || item.featureKind || ''}` !== featureFilter.value
      ) {
        return false
      }

      if (keyword) {
        const text = [
          item.name,
          item.permissionKey,
          item.code,
          item.description,
          moduleKeywords(item)
        ]
          .filter(Boolean)
          .join(' ')
          .toLowerCase()
        if (!text.includes(keyword)) {
          return false
        }
      }

      if (!matchState(item.id)) {
        return false
      }

      return true
    })
  })

  const groupedActions = computed(() => groupActions(filteredActions.value, isActive))

  const totalCount = computed(() => props.actions.length)
  const visibleCount = computed(() => filteredActions.value.length)
  const selectedCount = computed(() => props.actions.filter((item) => isActive(item.id)).length)

  const selectedLabel = computed(() => getSelectedLabel(props.mode))
  const positiveText = computed(() => (props.mode === 'collaboration' ? '已开通' : '已选'))
  const neutralText = computed(() => (props.mode === 'collaboration' ? '未开通' : '未选'))

  const decisionOptions = computed(() => buildDecisionOptions(props.mode))
  const stateOptions = computed(() => buildStateOptions(props.mode))
  const stateFilterPlaceholder = computed(() => stateOptions.value[0]?.label || '状态')
  const batchCommands = computed(() => buildBatchCommands(props.mode))

  watch(
    groupedActions,
    (value) => {
      const keys = value.map((item) => item.key)
      if (!activeFeatures.value.length) {
        activeFeatures.value = keys
        return
      }
      activeFeatures.value = activeFeatures.value.filter((item) => keys.includes(item))
      if (!activeFeatures.value.length) {
        activeFeatures.value = keys
      }
    },
    { immediate: true }
  )

  function toggleSelection(actionId: string, value: boolean) {
    const next = new Set(props.selectedIds)
    if (value) {
      next.add(actionId)
    } else {
      next.delete(actionId)
    }
    emit('update:selectedIds', Array.from(next))
  }

  function setDecision(actionId: string, value: string) {
    const next = { ...props.decisionMap }
    const normalized = normalizeDecision(value)
    if (!normalized) {
      delete next[actionId]
    } else {
      next[actionId] = normalized
    }
    emit('update:decisionMap', next)
  }

  function getDecisionLabel(actionId: string) {
    return formatDecisionLabel(props.mode, getDecision(actionId))
  }

  function getDecisionTagType(actionId: string) {
    return formatDecisionTagType(getDecision(actionId))
  }

  function expandAll() {
    activeFeatures.value = groupedActions.value.map((item) => item.key)
  }

  function collapseAll() {
    activeFeatures.value = []
  }

  function handleBatchCommand(command: string) {
    const visibleIds = filteredActions.value.map((item) => item.id)
    if (!visibleIds.length) return

    if (props.mode === 'menu' || props.mode === 'collaboration') {
      const next = new Set(props.selectedIds)
      if (command === 'select-visible') {
        visibleIds.forEach((id) => next.add(id))
      } else if (command === 'clear-visible') {
        visibleIds.forEach((id) => next.delete(id))
      } else if (command === 'clear-all') {
        next.clear()
      }
      emit('update:selectedIds', Array.from(next))
      return
    }

    const next = { ...props.decisionMap }
    if (command === 'clear-all') {
      emit('update:decisionMap', {})
      return
    }

    let decision: DecisionValue = ''
    if (command === 'allow-visible') decision = 'allow'
    if (command === 'deny-visible') decision = 'deny'
    if (command === 'unset-visible' || command === 'inherit-visible') decision = ''

    visibleIds.forEach((id) => {
      if (!decision) {
        delete next[id]
      } else {
        next[id] = decision
      }
    })
    emit('update:decisionMap', next)
  }

  return {
    searchKeyword,
    featureFilter,
    stateFilter,
    showMeta,
    compactMode,
    activeFeatures,
    featureOptions,
    groupedActions,
    totalCount,
    visibleCount,
    selectedCount,
    selectedLabel,
    positiveText,
    neutralText,
    decisionOptions,
    stateOptions,
    stateFilterPlaceholder,
    batchCommands,
    isSelected,
    toggleSelection,
    setDecision,
    getDecisionLabel,
    getDecisionTagType,
    expandAll,
    collapseAll,
    handleBatchCommand
  }
}
