<template>
  <nav class="message-workspace-nav" aria-label="消息工作台导航">
    <button
      v-for="item in navItems"
      :key="item.key"
      type="button"
      class="message-workspace-nav__item"
      :class="{ 'is-active': item.key === current }"
      @click="go(item.path)"
    >
      <span>{{ item.label }}</span>
      <small>{{ item.description }}</small>
    </button>
  </nav>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'

  defineOptions({ name: 'MessageWorkspaceNav' })

  const props = defineProps<{
    scope: 'platform' | 'team'
    current: 'dispatch' | 'template' | 'sender' | 'group' | 'record'
  }>()

  const router = useRouter()

  const navItems = computed(() => {
    const prefix = props.scope === 'team' ? '/team' : '/system'
    return [
      { key: 'dispatch', label: '消息发送', description: '编辑并发出消息', path: `${prefix}/message` },
      { key: 'template', label: '消息模板', description: '维护摘要与正文模板', path: `${prefix}/message-template` },
      { key: 'sender', label: '发送人', description: '管理发信身份', path: `${prefix}/message-sender` },
      { key: 'group', label: '接收组', description: '管理固定接收规则', path: `${prefix}/message-recipient-group` },
      { key: 'record', label: '发送记录', description: '查看投递与审计', path: `${prefix}/message-record` }
    ]
  })

  const go = (path: string) => {
    router.push(path)
  }
</script>

<style scoped lang="scss">
  .message-workspace-nav {
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    gap: 10px;
  }

  .message-workspace-nav__item {
    display: grid;
    gap: 4px;
    padding: 14px 16px;
    border: 1px solid rgb(226 232 240 / 0.9);
    border-radius: 18px;
    background: rgb(255 255 255 / 0.96);
    text-align: left;
    transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease;
  }

  .message-workspace-nav__item span {
    font-size: 13px;
    font-weight: 700;
    color: #0f172a;
  }

  .message-workspace-nav__item small {
    font-size: 11px;
    line-height: 1.5;
    color: #64748b;
  }

  .message-workspace-nav__item.is-active {
    border-color: rgb(59 130 246 / 0.38);
    background: rgb(239 246 255 / 0.88);
  }

  .message-workspace-nav__item:hover {
    border-color: rgb(148 163 184 / 0.72);
    transform: translateY(-1px);
  }

  @media (max-width: 1180px) {
    .message-workspace-nav {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 640px) {
    .message-workspace-nav {
      grid-template-columns: 1fr;
    }
  }
</style>
