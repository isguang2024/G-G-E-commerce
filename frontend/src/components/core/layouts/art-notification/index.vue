<template>
  <div
    class="art-notification-panel art-card-sm !shadow-xl"
    :style="{
      transform: show ? 'scaleY(1)' : 'scaleY(0.9)',
      opacity: show ? 1 : 0
    }"
    v-show="visible"
    @click.stop
  >
    <div class="flex-cb px-3.5 mt-3.5">
      <span class="text-base font-medium text-g-800">消息中心</span>
      <span
        class="text-xs text-g-800 px-1.5 py-1 c-p select-none rounded hover:bg-g-200"
        :class="{ 'opacity-40 pointer-events-none': !allowMarkAllRead || currentUnreadCount === 0 }"
        @click="handleMarkAllRead"
      >
        {{ activeBoxType === 'todo' ? '去处理' : '全部已读' }}
      </span>
    </div>

    <ul class="box-border flex items-end w-full h-12.5 px-3.5 border-b-d">
      <li
        v-for="(item, index) in barList"
        :key="item.key"
        class="h-12 leading-12 mr-5 overflow-hidden text-[13px] text-g-700 c-p select-none"
        :class="{ 'bar-active': barActiveIndex === index }"
        @click="changeBar(index)"
      >
        {{ item.name }} ({{ item.num }})
      </li>
    </ul>

    <div class="w-full h-[calc(100%-95px)]">
      <div class="h-[calc(100%-60px)] overflow-y-scroll scrollbar-thin">
        <ul v-if="currentList.length">
          <li
            v-for="item in currentList"
            :key="item.id"
            class="message-row"
            :class="{ 'is-unread': item.delivery_status === 'unread' }"
            @click="handleItemClick(item)"
          >
            <div class="message-row__avatar" :class="resolveAvatarClass(item)">
              <img
                v-if="item.sender_avatar_snapshot"
                :src="item.sender_avatar_snapshot"
                class="size-full rounded-xl object-cover"
              />
              <ArtSvgIcon v-else :icon="resolveItemIcon(item)" class="text-lg !bg-transparent" />
            </div>

            <div class="message-row__body">
              <div class="message-row__head">
                <h4 class="message-row__title">{{ item.title }}</h4>
                <span v-if="item.delivery_status === 'unread'" class="message-row__dot"></span>
              </div>
              <p class="message-row__summary">{{ resolveSummary(item) }}</p>
              <div class="message-row__meta">
                <div class="message-row__meta-tags">
                  <span class="message-chip">{{ resolveSourceLabel(item) }}</span>
                  <span v-if="resolveTeamTag(item)" class="message-chip is-team">{{ resolveTeamTag(item) }}</span>
                </div>
                <span class="message-row__time">{{ formatTime(item.last_action_at || item.published_at || item.created_at) }}</span>
              </div>

              <div v-if="item.box_type === 'todo'" class="message-row__todo-actions">
                <ElButton size="small" type="primary" plain @click.stop="handleTodo(item, 'done')">
                  完成
                </ElButton>
                <ElButton size="small" @click.stop="handleTodo(item, 'ignored')">忽略</ElButton>
              </div>
            </div>
          </li>
        </ul>

        <div v-else class="message-empty">
          <ArtSvgIcon icon="system-uicons:inbox" class="text-5xl" />
          <p class="mt-3.5 text-xs !bg-transparent">当前没有{{ currentBarName }}</p>
        </div>
      </div>

      <div class="relative box-border w-full px-3.5">
        <ElButton class="w-full mt-3" @click="handleViewAll" v-ripple>查看全部</ElButton>
      </div>
    </div>

    <div class="h-25"></div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, watch } from 'vue'
  import { ElMessage } from 'element-plus'
  import { useRouter } from 'vue-router'
  import { storeToRefs } from 'pinia'
  import { useMessageStore } from '@/store/modules/message'
  import { useTenantStore } from '@/store/modules/tenant'

  defineOptions({ name: 'ArtNotification' })

  const props = defineProps<{
    value: boolean
  }>()

  const emit = defineEmits<{
    'update:value': [value: boolean]
  }>()

  const router = useRouter()
  const messageStore = useMessageStore()
  const tenantStore = useTenantStore()
  const { previewMap } = storeToRefs(messageStore)

  const show = ref(false)
  const visible = ref(false)
  const barActiveIndex = ref(0)

  const barList = computed(() => [
    { key: 'notice', name: '通知', num: previewMap.value.notice.length },
    { key: 'message', name: '消息', num: previewMap.value.message.length },
    { key: 'todo', name: '待办', num: previewMap.value.todo.length }
  ])

  const activeBoxType = computed<Api.Message.BoxType>(() => {
    return (barList.value[barActiveIndex.value]?.key || 'notice') as Api.Message.BoxType
  })

  const currentBarName = computed(() => barList.value[barActiveIndex.value]?.name || '消息')
  const currentList = computed(() => previewMap.value[activeBoxType.value] || [])
  const currentUnreadCount = computed(
    () => currentList.value.filter((item) => item.delivery_status === 'unread').length
  )
  const allowMarkAllRead = computed(() => activeBoxType.value !== 'todo')

  const animatePanel = (open: boolean) => {
    if (open) {
      visible.value = true
      setTimeout(() => {
        show.value = true
      }, 5)
      messageStore.loadPanelData(true).catch(() => {
        ElMessage.error('获取消息失败')
      })
      return
    }
    show.value = false
    setTimeout(() => {
      visible.value = false
    }, 300)
  }

  const changeBar = (index: number) => {
    barActiveIndex.value = index
    messageStore.loadPreview(activeBoxType.value, true).catch(() => {
      ElMessage.error('获取消息失败')
    })
  }

  const formatTime = (value?: string) => {
    if (!value) return '刚刚'
    const date = new Date(value)
    if (Number.isNaN(date.getTime())) return value
    return new Intl.DateTimeFormat('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date)
  }

  const resolveItemIcon = (item: Api.Message.InboxItem) => {
    if (item.box_type === 'todo') return 'ri:task-line'
    if (item.sender_type === 'service' || item.sender_type === 'automation') return 'ri:robot-2-line'
    if (item.box_type === 'message') return 'ri:message-2-line'
    return 'ri:notification-3-line'
  }

  const resolveAvatarClass = (item: Api.Message.InboxItem) => {
    if (item.box_type === 'todo') return 'is-todo'
    if (item.box_type === 'message') return 'is-message'
    return 'is-notice'
  }

  const resolveSourceLabel = (item: Api.Message.InboxItem) => {
    if (item.sender_name_snapshot) return item.sender_name_snapshot
    if (item.sender_type === 'system') return '系统'
    if (item.sender_type === 'service') return item.sender_service_key || '服务'
    if (item.sender_type === 'automation') return '自动任务'
    return item.box_type === 'todo' ? '待处理' : '站内消息'
  }

  const plainTextFromHtml = (value?: string) => {
    const target = `${value || ''}`.trim()
    if (!target) return ''
    if (typeof window === 'undefined') {
      return target.replace(/<[^>]+>/g, ' ').replace(/&nbsp;/g, ' ').replace(/\s+/g, ' ').trim()
    }
    const parser = new DOMParser()
    const doc = parser.parseFromString(target, 'text/html')
    return (doc.body.textContent || '').replace(/\s+/g, ' ').trim()
  }

  const resolveSummary = (item: Api.Message.InboxItem) => {
    return plainTextFromHtml(item.summary) || plainTextFromHtml(item.content) || '点击查看消息详情'
  }

  const resolveTeamTag = (item: Api.Message.InboxItem) => {
    const teamId = item.scope_type === 'team'
      ? (item.scope_id || item.recipient_team_id || item.target_tenant_id)
      : (item.recipient_team_id || item.target_tenant_id || item.scope_id)
    if (!teamId) return ''
    const teamName = tenantStore.teamList.find((team) => team.id === teamId)?.name || ''
    return teamName ? `团队 · ${teamName}` : ''
  }

  const navigateByItem = async (item: Api.Message.InboxItem) => {
    await router.push({
      path: '/workspace/inbox',
      query: {
        deliveryId: item.id,
        boxType: item.box_type
      }
    })
  }

  const handleItemClick = async (item: Api.Message.InboxItem) => {
    try {
      if (item.delivery_status === 'unread') {
        await messageStore.markRead(item.id)
      }
      emit('update:value', false)
      await navigateByItem(item)
    } catch (error) {
      ElMessage.error('打开消息失败')
    }
  }

  const handleMarkAllRead = async () => {
    if (activeBoxType.value === 'todo') {
      handleViewAll()
      return
    }
    if (currentUnreadCount.value === 0) return
    try {
      await messageStore.markReadAll(activeBoxType.value)
      ElMessage.success('已标记当前分类为已读')
    } catch (error) {
      ElMessage.error('批量已读失败')
    }
  }

  const handleTodo = async (item: Api.Message.InboxItem, action: 'done' | 'ignored') => {
    try {
      await messageStore.handleTodo(item.id, action)
      ElMessage.success(action === 'done' ? '待办已完成' : '待办已忽略')
    } catch (error) {
      ElMessage.error('处理待办失败')
    }
  }

  const handleViewAll = () => {
    emit('update:value', false)
    router.push({
      path: '/workspace/inbox',
      query: {
        boxType: activeBoxType.value
      }
    })
  }

  watch(
    () => props.value,
    (newValue) => {
      animatePanel(newValue)
    },
    { immediate: true }
  )
</script>

<style scoped lang="scss">
  @reference '@styles/core/tailwind.css';

  .art-notification-panel {
    position: absolute;
    top: 58px;
    right: 20px;
    z-index: 60;
    width: min(360px, calc(100vw - 32px));
    height: 500px;
    overflow: hidden;
    transform-origin: top right;
    transition:
      transform 0.3s ease,
      opacity 0.3s ease;
    will-change: transform, opacity;
  }

  @media (max-width: 640px) {
    .art-notification-panel {
      top: 64px;
      right: 12px;
      width: calc(100vw - 24px);
      height: min(80vh, 560px);
    }
  }

  .bar-active {
    color: var(--theme-color) !important;
    border-bottom: 2px solid var(--theme-color);
  }

  .message-row {
    display: flex;
    gap: 12px;
    padding: 12px 13px;
    cursor: pointer;
    border-bottom: 1px solid rgb(226 232 240 / 0.7);
    transition:
      background-color 0.2s ease,
      transform 0.2s ease;
  }

  .message-row:hover {
    background: rgb(248 250 252 / 0.9);
  }

  .message-row.is-unread {
    background: linear-gradient(90deg, rgb(239 246 255 / 0.85), transparent 72%);
  }

  .message-row__avatar {
    display: flex;
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    border-radius: 14px;
  }

  .message-row__avatar.is-notice {
    background: rgb(219 234 254 / 0.75);
    color: #1d4ed8;
  }

  .message-row__avatar.is-message {
    background: rgb(220 252 231 / 0.85);
    color: #047857;
  }

  .message-row__avatar.is-todo {
    background: rgb(254 242 242 / 0.95);
    color: #dc2626;
  }

  .message-row__body {
    min-width: 0;
    flex: 1;
  }

  .message-row__head {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .message-row__title {
    margin: 0;
    overflow: hidden;
    font-size: 13px;
    font-weight: 600;
    line-height: 1.5;
    color: #0f172a;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .message-row__dot {
    width: 7px;
    height: 7px;
    flex-shrink: 0;
    border-radius: 999px;
    background: #ef4444;
  }

  .message-row__summary {
    margin: 4px 0 0;
    display: -webkit-box;
    overflow: hidden;
    font-size: 12px;
    line-height: 1.55;
    color: #64748b;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .message-row__meta {
    margin-top: 8px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .message-row__meta-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    min-width: 0;
  }

  .message-chip {
    display: inline-flex;
    align-items: center;
    padding: 2px 8px;
    border-radius: 999px;
    background: rgb(241 245 249 / 1);
    font-size: 11px;
    color: #475569;
  }

  .message-chip.is-team {
    background: rgb(219 234 254 / 0.9);
    color: #1d4ed8;
  }

  .message-row__time {
    font-size: 11px;
    color: #94a3b8;
  }

  .message-row__todo-actions {
    display: flex;
    gap: 8px;
    margin-top: 10px;
  }

  .message-empty {
    position: relative;
    top: 100px;
    text-align: center;
    color: #64748b;
    background: transparent;
  }

  .scrollbar-thin::-webkit-scrollbar {
    width: 5px !important;
  }

  .dark .scrollbar-thin::-webkit-scrollbar-track {
    background-color: var(--default-box-color);
  }

  .dark .scrollbar-thin::-webkit-scrollbar-thumb {
    background-color: #222 !important;
  }
</style>
