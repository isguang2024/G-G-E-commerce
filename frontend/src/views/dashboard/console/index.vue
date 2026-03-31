<template>
  <div class="console-page art-full-height">
    <AdminWorkspaceHero
      title="后台工作台"
      description="把当前待处理模块、常用入口和本轮重点集中在一屏内，进入后台后先看这里。"
      :metrics="heroMetrics"
    >
      <div class="console-hero-actions">
        <ElButton type="primary" @click="go('/system/page')" v-ripple>进入页面管理</ElButton>
        <ElButton @click="go('/system/api-endpoint')" v-ripple>查看 API 管理</ElButton>
      </div>
    </AdminWorkspaceHero>

    <div class="console-grid">
      <section class="console-panel console-panel--wide">
        <header class="console-panel__header">
          <div>
            <div class="console-panel__title">当前工作面</div>
            <p class="console-panel__desc">根据当前团队或管理入口，优先进入对应的治理链路。</p>
          </div>
        </header>
        <div class="console-focus-list">
          <button
            v-for="item in focusItems"
            :key="item.title"
            type="button"
            class="console-focus-item"
            @click="go(item.path)"
          >
            <div class="console-focus-item__title">{{ item.title }}</div>
            <div class="console-focus-item__text">{{ item.text }}</div>
            <div class="console-focus-item__path">{{ item.path }}</div>
          </button>
        </div>
      </section>

      <section class="console-panel">
        <header class="console-panel__header">
          <div>
            <div class="console-panel__title">当前概览</div>
            <p class="console-panel__desc">把当前团队和当前入口状态收在一张卡片里，不额外强调空间模型。</p>
          </div>
        </header>
        <div class="console-scope-card">
          <div class="console-scope-card__label">当前入口</div>
          <div class="console-scope-card__name">{{ currentScopeName }}</div>
          <p class="console-scope-card__text">{{ currentScopeDescription }}</p>
        </div>
      </section>

      <section class="console-panel">
        <header class="console-panel__header">
          <div>
            <div class="console-panel__title">本轮优先级</div>
            <p class="console-panel__desc">界面收口先做统一表达，再做深层配置能力。</p>
          </div>
        </header>
        <ul class="console-task-list">
          <li v-for="item in nextTasks" :key="item">{{ item }}</li>
        </ul>
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import { useRouter } from 'vue-router'
  import AdminWorkspaceHero from '@/components/business/layout/AdminWorkspaceHero.vue'
  import { useTenantStore } from '@/store/modules/tenant'

  defineOptions({ name: 'Console' })

  const router = useRouter()
  const tenantStore = useTenantStore()
  const { currentContextMode, currentTeam, teamList, hasPlatformAccess } = storeToRefs(tenantStore)

  const heroMetrics = computed(() => [
    { label: '团队数量', value: teamList.value.length || 0 },
    { label: '当前团队', value: currentTeam.value?.name || '未选择团队' },
    { label: '管理入口', value: currentContextMode.value === 'platform' ? '平台管理' : '团队工作台' }
  ])

  const currentScopeName = computed(() =>
    currentContextMode.value === 'platform' ? '平台管理' : currentTeam.value?.name || '未选择团队'
  )
  const currentScopeDescription = computed(() =>
    currentContextMode.value === 'platform'
      ? '这里用于处理菜单、页面、权限键、API 元数据和平台角色等平台治理能力。'
      : '这里用于处理团队成员、团队角色、边界和已开通功能包等日常团队能力。'
  )

  const focusItems = computed(() =>
    currentContextMode.value === 'platform'
      ? [
          {
            title: '页面与菜单主链',
            text: '先确认导航入口、页面挂载和访问模式是否一致。',
            path: '/system/page'
          },
          {
            title: 'API 注册与权限归属',
            text: '查看分类、失效接口和未注册路由，保持注册表闭环。',
            path: '/system/api-endpoint'
          },
          {
            title: '用户与角色治理',
            text: '处理平台身份、功能包与权限测试链路。',
            path: '/system/user'
          }
        ]
      : [
          {
            title: '团队成员',
            text: '先确认当前团队成员、角色和菜单边界是否生效。',
            path: '/team/team-members'
          },
          {
            title: '团队角色边界',
            text: '统一查看团队角色、功能包、菜单和功能权限。',
            path: '/system/team-roles-permissions'
          },
          {
            title: '团队总览',
            text: '检查团队资料、功能包和成员入口是否可用。',
            path: '/team/team'
          }
        ]
  )

  const nextTasks = [
    '继续收口系统页头部和操作区，减少不同页面之间的结构漂移。',
    '菜单、用户、API 管理继续对齐页面管理的分段式抽屉和顶部摘要。',
    '后续再补页面树轻量接口，收掉最后的全量树数据拉取。'
  ]

  const go = (path: string) => {
    router.push(path)
  }
</script>

<style scoped lang="scss">
  .console-page {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .console-hero-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
  }

  .console-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.5fr) minmax(300px, 0.8fr);
    gap: 16px;
  }

  .console-panel {
    display: flex;
    min-width: 0;
    flex-direction: column;
    gap: 14px;
    padding: 18px;
    border: 1px solid color-mix(in srgb, var(--el-border-color) 78%, white);
    border-radius: 20px;
    background:
      radial-gradient(circle at top left, rgb(255 255 255 / 0.92), transparent 48%),
      linear-gradient(180deg, rgb(255 255 255 / 0.98), rgb(248 250 252 / 0.95));
  }

  .console-panel--wide {
    grid-row: span 2;
  }

  .console-panel__header {
    display: flex;
    justify-content: space-between;
    gap: 12px;
  }

  .console-panel__title {
    font-size: 16px;
    font-weight: 700;
    color: #0f172a;
  }

  .console-panel__desc {
    margin: 4px 0 0;
    font-size: 13px;
    line-height: 1.6;
    color: #64748b;
  }

  .console-focus-list {
    display: grid;
    gap: 12px;
  }

  .console-focus-item {
    display: flex;
    width: 100%;
    flex-direction: column;
    gap: 6px;
    padding: 16px;
    text-align: left;
    cursor: pointer;
    border: 1px solid rgb(226 232 240 / 0.92);
    border-radius: 16px;
    background: rgb(255 255 255 / 0.76);
    transition:
      transform 0.18s ease,
      border-color 0.18s ease,
      box-shadow 0.18s ease;
  }

  .console-focus-item:hover {
    transform: translateY(-2px);
    border-color: rgb(125 211 252 / 0.9);
    box-shadow: 0 14px 30px rgb(15 23 42 / 0.08);
  }

  .console-focus-item__title {
    font-size: 15px;
    font-weight: 700;
    color: #0f172a;
  }

  .console-focus-item__text {
    font-size: 13px;
    line-height: 1.6;
    color: #475569;
  }

  .console-focus-item__path {
    font-family: 'JetBrains Mono', Consolas, monospace;
    font-size: 12px;
    color: #0f766e;
  }

  .console-scope-card {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 16px;
    border-radius: 16px;
    background: linear-gradient(135deg, rgb(240 253 244 / 0.94), rgb(239 246 255 / 0.92));
  }

  .console-scope-card__label {
    font-size: 12px;
    letter-spacing: 0.06em;
    color: #64748b;
    text-transform: uppercase;
  }

  .console-scope-card__name {
    font-size: 20px;
    font-weight: 700;
    color: #0f172a;
  }

  .console-scope-card__text {
    margin: 0;
    font-size: 13px;
    line-height: 1.6;
    color: #475569;
  }

  .console-task-list {
    margin: 0;
    padding-left: 18px;
    color: #334155;
    font-size: 13px;
    line-height: 1.8;
  }

  @media (max-width: 1080px) {
    .console-grid {
      grid-template-columns: 1fr;
    }

    .console-panel--wide {
      grid-row: auto;
    }
  }
</style>

