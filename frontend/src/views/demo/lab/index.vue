<template>
  <div class="demo-app-lab">
    <ElCard shadow="never" class="demo-app-lab__card">
      <template #header>
        <div class="demo-app-lab__header">
          <span>Demo App 验证页</span>
          <ElTag type="success" effect="light">/demo/lab</ElTag>
        </div>
      </template>
      <p class="demo-app-lab__desc">
        该页面用于验证 path_prefix 多 APP 切换链路。
      </p>
      <ul class="demo-app-lab__list">
        <li>当前 APP：demo-app</li>
        <li>入口前缀：/demo</li>
        <li>目标：与 platform-admin、account-portal 切换互不污染</li>
      </ul>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  defineOptions({ name: 'DemoAppLab' })
</script>

<style scoped lang="scss">
  .demo-app-lab {
    padding: 20px;
  }

  .demo-app-lab__card {
    max-width: 760px;
  }

  .demo-app-lab__header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 600;
  }

  .demo-app-lab__desc {
    margin-bottom: 12px;
    color: var(--art-text-gray-700);
  }

  .demo-app-lab__list {
    margin: 0;
    padding-left: 18px;
    color: var(--art-text-gray-700);
  }
</style>
