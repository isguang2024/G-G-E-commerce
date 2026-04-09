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

  onMounted(() => load(1))
</script>
