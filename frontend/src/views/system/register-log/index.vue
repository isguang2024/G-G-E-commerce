<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-lg font-semibold">注册记录</h3>
    </div>
    <ElForm :inline="true" :model="filter" class="mb-3">
      <ElFormItem label="来源">
        <ElInput v-model="filter.source" clearable placeholder="self / invite" />
      </ElFormItem>
      <ElFormItem label="入口 Code">
        <ElInput v-model="filter.entry_code" clearable />
      </ElFormItem>
      <ElFormItem label="策略 Code">
        <ElInput v-model="filter.policy_code" clearable />
      </ElFormItem>
      <ElFormItem>
        <ElButton type="primary" @click="load(1)">查询</ElButton>
        <ElButton @click="reset">重置</ElButton>
      </ElFormItem>
    </ElForm>
    <ElTable :data="list" border stripe>
      <ElTableColumn prop="username" label="用户名" width="160" />
      <ElTableColumn prop="email" label="邮箱" width="200" />
      <ElTableColumn prop="register_source" label="来源" width="100" />
      <ElTableColumn prop="register_entry_code" label="入口 Code" width="160" />
      <ElTableColumn prop="register_policy_code" label="策略 Code" width="160" />
      <ElTableColumn prop="register_app_key" label="注册 App" width="140" />
      <ElTableColumn label="策略快照" min-width="260">
        <template #default="{ row }">
          <div class="snapshot-tags">
            <ElTag v-for="tag in buildSnapshotTags(row.policy_snapshot)" :key="tag" effect="plain">
              {{ tag }}
            </ElTag>
          </div>
          <div class="text-xs text-gray-500 mt-1">{{ buildSnapshotLanding(row.policy_snapshot) }}</div>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="register_ip" label="IP" width="140" />
      <ElTableColumn prop="agreement_version" label="协议版本" width="110" />
      <ElTableColumn prop="created_at" label="注册时间" width="200" />
    </ElTable>
    <div class="mt-3 flex justify-end">
      <ElPagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[20, 50, 100]"
        layout="total, sizes, prev, pager, next"
        @current-change="load()"
        @size-change="load(1)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import { fetchListRegisterLogs } from '@/api/system-manage/register'

  defineOptions({ name: 'SystemRegisterLog' })

  const list = ref<any[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const filter = reactive<any>({ source: '', entry_code: '', policy_code: '' })

  const load = async (p?: number) => {
    if (p) page.value = p
    try {
      const data: any = await fetchListRegisterLogs({
        source: filter.source || undefined,
        entry_code: filter.entry_code || undefined,
        policy_code: filter.policy_code || undefined,
        page: page.value,
        page_size: pageSize.value
      })
      list.value = data?.records || []
      total.value = data?.total || 0
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    }
  }

  const reset = () => {
    filter.source = ''
    filter.entry_code = ''
    filter.policy_code = ''
    load(1)
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
  .snapshot-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
</style>
