<template>
  <div class="p-4 register-log-page art-full-height">
    <ElCard class="art-table-card register-log-main" shadow="never">
      <ElForm :inline="true" :model="filter" class="register-log-filters">
        <ElFormItem label="来源">
          <ElInput v-model="filter.source" clearable placeholder="self / invite" />
        </ElFormItem>
        <ElFormItem label="入口 Code">
          <ElInput v-model="filter.entry_code" clearable />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="load(1)">查询</ElButton>
          <ElButton @click="reset">重置</ElButton>
        </ElFormItem>
      </ElForm>

      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load()">
        <template #left>
          <div class="register-log-header">
            <div class="register-log-title">注册记录</div>
            <div class="register-log-tip">记录会冻结注册策略快照，便于回溯当时的入口、来源和落地规则。</div>
            <div class="table-toolbar-tip">保留服务端分页，筛选条件变化会从第一页重新拉取。</div>
          </div>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="list"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import { ElMessage, ElTag } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import { fetchListRegisterLogs } from '@/domains/governance/api/register'

  defineOptions({ name: 'SystemRegisterLog' })

  const list = ref<any[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const filter = reactive<any>({ source: '', entry_code: '' })
  const pagination = computed(() => ({
    current: page.value,
    size: pageSize.value,
    total: total.value
  }))
  const columns = computed<ColumnOption[]>(() => [
    { type: 'index', label: '序号', width: 70 },
    { prop: 'username', label: '用户名', width: 160, showOverflowTooltip: true },
    { prop: 'email', label: '邮箱', minWidth: 200, showOverflowTooltip: true },
    { prop: 'register_source', label: '来源', width: 100 },
    { prop: 'register_entry_code', label: '入口 Code', minWidth: 160, showOverflowTooltip: true },
    { prop: 'register_app_key', label: '注册 App', width: 140 },
    {
      prop: 'policy_snapshot',
      label: '策略快照',
      minWidth: 320,
      formatter: (row) =>
        h('div', {}, [
          h(
            'div',
            { class: 'snapshot-tags' },
            buildSnapshotTags(row.policy_snapshot).map((tag) => h(ElTag, { effect: 'plain' }, () => tag))
          ),
          h('div', { class: 'text-xs text-gray-500 mt-1' }, buildSnapshotLanding(row.policy_snapshot))
        ])
    },
    { prop: 'register_ip', label: 'IP', width: 140 },
    { prop: 'agreement_version', label: '协议版本', width: 110 },
    { prop: 'created_at', label: '注册时间', minWidth: 180 }
  ])

  const load = async (p?: number) => {
    if (p) page.value = p
    loading.value = true
    try {
      const data: any = await fetchListRegisterLogs({
        source: filter.source || undefined,
        entry_code: filter.entry_code || undefined,
        page: page.value,
        page_size: pageSize.value
      })
      list.value = data?.records || []
      total.value = data?.total || 0
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  const reset = () => {
    filter.source = ''
    filter.entry_code = ''
    load(1)
  }

  const handleSizeChange = (size: number) => {
    pageSize.value = size
    load(1)
  }

  const handleCurrentChange = (current: number) => {
    load(current)
  }

  const buildSnapshotTags = (snapshot?: Record<string, any> | null) => {
    if (!snapshot) return ['未冻结快照']
    const tags = [snapshot.allow_public_register ? '公开注册' : '关闭公开注册']
    if (snapshot.require_invite) tags.push('邀请码')
    if (snapshot.require_email_verify) tags.push('邮箱验证')
    if (snapshot.require_captcha) tags.push('人机验证')
    if (snapshot.auto_login) tags.push('自动登录')
    if (Array.isArray(snapshot.role_codes) && snapshot.role_codes.length) {
      tags.push(`角色:${snapshot.role_codes.join(',')}`)
    }
    if (Array.isArray(snapshot.feature_package_keys) && snapshot.feature_package_keys.length) {
      tags.push(`功能包:${snapshot.feature_package_keys.join(',')}`)
    }
    return tags
  }

  const buildSnapshotLanding = (snapshot?: Record<string, any> | null) => {
    if (!snapshot) return '未记录 landing 快照'
    return `${snapshot.target_app_key || '-'} / ${snapshot.target_navigation_space_key || '-'} / ${
      snapshot.target_home_path || '-'
    }`
  }

  onMounted(() => load(1))
</script>

<style scoped>
  .page-card {
    padding: 24px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 20px;
    background: #fff;
    box-shadow:
      0 12px 30px rgb(15 23 42 / 5%),
      0 2px 8px rgb(15 23 42 / 4%);
  }

  .register-log-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .register-log-main {
    flex: 1;
    min-height: 0;
  }

  .register-log-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .register-log-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .register-log-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .register-log-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .register-log-filters {
    margin-bottom: 4px;
  }

  .table-toolbar-tip {
    font-size: 13px;
    color: var(--el-text-color-secondary);
  }

  .snapshot-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
</style>
