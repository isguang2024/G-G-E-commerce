<template>
  <div class="p-4 register-entry-page art-full-height">
    <ElCard class="art-table-card register-entry-main" shadow="never">
      <ArtTableHeader layout="refresh,fullscreen" :loading="loading" @refresh="load">
        <template #left>
          <div class="register-entry-header">
            <div class="register-entry-title">注册入口</div>
            <div class="register-entry-tip">
              入口决定“哪个 Host + Path 会命中哪套注册策略”。推荐顺序是：先准备注册策略，再为不同域名或路径补注册入口，最后访问 URL 验证命中结果。
            </div>
          </div>
        </template>
        <template #right>
          <ElButton @click="applyDefaultEntry">填入默认入口</ElButton>
          <ElButton type="primary" @click="openCreate">新建入口</ElButton>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="pagedList"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>

    <ElDialog v-model="dialogVisible" :title="editing ? '编辑入口' : '新建入口'" width="860px">
      <div class="dialog-layout">
        <ElForm :model="form" label-width="140px" class="dialog-form">
          <ElFormItem label="入口 Code" required>
            <ElInput v-model="form.entry_code" :disabled="!!editing" placeholder="如 default / invite-only" />
          </ElFormItem>
          <ElFormItem label="名称" required>
            <ElInput v-model="form.name" placeholder="给运营可读的入口名称" />
          </ElFormItem>
          <ElFormItem label="App Key" required>
            <ElInput v-model="form.app_key" placeholder="如 account-portal" />
            <div class="field-tip">公开注册页建议归属到 account-portal，便于和登录/找回密码统一治理。</div>
          </ElFormItem>
          <ElFormItem label="Host">
            <ElInput v-model="form.host" placeholder="留空匹配任意 host；也可填 account.example.com" />
          </ElFormItem>
          <ElFormItem label="Path 前缀">
            <ElInput v-model="form.path_prefix" placeholder="如 /account/auth/register" />
            <div class="field-tip">按前缀命中，越具体越优先。一个策略可挂多个入口。</div>
          </ElFormItem>
          <ElFormItem label="Register Source">
            <ElInput v-model="form.register_source" placeholder="self / invite / ..." />
          </ElFormItem>
          <ElFormItem label="策略 Code" required>
            <ElInput v-model="form.policy_code" placeholder="如 default.self" />
            <div class="field-tip">策略负责“注册后去哪、是否需要邀请码/邮箱验证/验证码、绑定哪些角色与功能包”。</div>
          </ElFormItem>
          <ElFormItem label="登录页模板 Key">
            <ElSelect v-model="form.login_page_key" filterable allow-create default-first-option>
              <ElOption
                v-for="item in templateList"
                :key="item.template_key"
                :label="`${item.template_key} · ${item.name}`"
                :value="item.template_key"
              />
            </ElSelect>
            <div class="field-tip">优先级低于 URL query，未填写时回退到 APP auth.login_page_key 或 default。</div>
          </ElFormItem>
          <ElFormItem label="状态">
            <ElSelect v-model="form.status">
              <ElOption label="enabled" value="enabled" />
              <ElOption label="disabled" value="disabled" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="排序">
            <ElInputNumber v-model="form.sort_order" :min="0" />
          </ElFormItem>
          <ElFormItem label="允许公开注册">
            <ElSelect v-model="form.allow_public_register" clearable placeholder="继承策略">
              <ElOption :value="true" label="是" />
              <ElOption :value="false" label="否" />
            </ElSelect>
            <div class="field-tip">留空表示继承策略；显式设置后，优先级高于策略同名开关。</div>
          </ElFormItem>
          <ElFormItem label="备注">
            <ElInput v-model="form.remark" type="textarea" :rows="2" />
          </ElFormItem>
        </ElForm>

        <div class="dialog-preview">
          <div class="preview-card">
            <div class="preview-title">配置预览</div>
            <ElDescriptions :column="1" border size="small">
              <ElDescriptionsItem label="命中规则">{{ formMatchRule }}</ElDescriptionsItem>
              <ElDescriptionsItem label="验证地址">{{ previewVerifyUrl }}</ElDescriptionsItem>
              <ElDescriptionsItem label="策略来源">{{ form.policy_code || '待填写策略 Code' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="模板 Key">{{ form.login_page_key || 'default' }}</ElDescriptionsItem>
              <ElDescriptionsItem label="公开注册">
                {{ resolvePublicRegisterTag(form.allow_public_register).label }}
              </ElDescriptionsItem>
              <ElDescriptionsItem label="注册来源">{{ form.register_source || 'self' }}</ElDescriptionsItem>
            </ElDescriptions>
          </div>

          <ElAlert
            class="mt-4"
            type="success"
            :closable="false"
            title="保存后怎么验"
            :description="`1. 打开 ${previewVerifyUrl}；2. 检查页面顶部是否显示命中入口和策略；3. 如果未命中，优先检查 host/path_prefix 与排序。`"
          />
        </div>
      </div>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="submit">保存</ElButton>
      </template>
    </ElDialog>

  </div>
</template>

<script setup lang="ts">
  import { computed, h, onMounted, reactive, ref } from 'vue'
  import { ElButton, ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import type { ColumnOption } from '@/types/component'
  import {
    fetchCreateRegisterEntry,
    fetchDeleteRegisterEntry,
    fetchListLoginPageTemplates,
    fetchListRegisterEntries,
    fetchUpdateRegisterEntry
  } from '@/domains/governance/api/register'

  defineOptions({ name: 'SystemRegisterEntry' })

  const DEFAULT_ENTRY = {
    entry_code: 'default',
    name: '默认公开注册入口',
    app_key: 'account-portal',
    host: '',
    path_prefix: '/account/auth/register',
    register_source: 'self',
    policy_code: 'default.self',
    login_page_key: 'default',
    status: 'enabled',
    sort_order: 0,
    allow_public_register: null,
    remark: '默认本地入口：命中 account-portal 注册页'
  }

  const list = ref<any[]>([])
  const templateList = ref<any[]>([])
  const dialogVisible = ref(false)
  const editing = ref<any>(null)
  const loading = ref(false)
  const pagination = reactive({
    current: 1,
    size: 10,
    total: 0
  })

  const emptyForm = () => ({ ...DEFAULT_ENTRY })
  const form = reactive<any>(emptyForm())

  const formMatchRule = computed(() => buildMatchRule(form))
  const previewVerifyUrl = computed(() => buildVerifyUrl(form))
  const pagedList = computed(() => {
    const start = (pagination.current - 1) * pagination.size
    return list.value.slice(start, start + pagination.size)
  })
  const columns = computed<ColumnOption[]>(() => [
    { type: 'index', label: '序号', width: 70 },
    { prop: 'entry_code', label: '入口 Code', minWidth: 160, showOverflowTooltip: true },
    { prop: 'name', label: '名称', minWidth: 180, showOverflowTooltip: true },
    { prop: 'app_key', label: 'App', width: 140 },
    { prop: 'login_page_key', label: '登录页模板', width: 140 },
    {
      prop: 'match_rule',
      label: '命中规则',
      minWidth: 280,
      formatter: (row) =>
        h('div', {}, [
          h('div', { class: 'font-medium' }, buildMatchRule(row)),
          h('div', { class: 'text-xs text-gray-500' }, buildVerifyUrl(row))
        ])
    },
    { prop: 'policy_code', label: '策略 Code', minWidth: 160, showOverflowTooltip: true },
    {
      prop: 'allow_public_register',
      label: '公开注册',
      width: 140,
      formatter: (row) => {
        const state = resolvePublicRegisterTag(row.allow_public_register)
        return h(ElTag, { type: state.type, effect: 'plain' }, () => state.label)
      }
    },
    { prop: 'status', label: '状态', width: 110 },
    {
      prop: 'actions',
      label: '操作',
      width: 150,
      fixed: 'right',
      formatter: (row) =>
        h('div', { class: 'table-actions' }, [
          h(
            ElButton,
            {
              link: true,
              type: 'primary',
              onClick: () => openEdit(row)
            },
            () => '编辑'
          ),
          h(
            ElButton,
            {
              link: true,
              type: 'danger',
              onClick: () => confirmRemove(row)
            },
            () => '删除'
          )
        ])
    }
  ])

  const load = async () => {
    loading.value = true
    try {
      const data: any = await fetchListRegisterEntries()
      list.value = data?.records || []
      pagination.total = list.value.length
      syncCurrentPage()
      const templates: any = await fetchListLoginPageTemplates()
      templateList.value = templates?.records || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
    } finally {
      loading.value = false
    }
  }

  const openCreate = () => {
    editing.value = null
    Object.assign(form, emptyForm())
    dialogVisible.value = true
  }

  const openEdit = (row: any) => {
    editing.value = row
    Object.assign(form, emptyForm(), row)
    dialogVisible.value = true
  }

  const applyDefaultEntry = () => {
    editing.value = null
    Object.assign(form, emptyForm())
    dialogVisible.value = true
  }

  const submit = async () => {
    try {
      const payload = { ...form }
      payload.login_page_key = `${payload.login_page_key || ''}`.trim() || 'default'
      if (payload.allow_public_register === '' || payload.allow_public_register === undefined) {
        payload.allow_public_register = null
      }
      if (editing.value) {
        await fetchUpdateRegisterEntry(editing.value.id, payload)
        ElMessage.success('更新成功')
      } else {
        await fetchCreateRegisterEntry(payload)
        ElMessage.success('创建成功')
      }
      dialogVisible.value = false
      await load()
    } catch (e: any) {
      ElMessage.error(e?.message || '保存失败')
    }
  }

  const confirmRemove = async (row: any) => {
    try {
      await ElMessageBox.confirm(`确认删除入口“${row.name || row.entry_code}”吗？`, '删除确认', {
        type: 'warning'
      })
      await remove(row)
    } catch {}
  }

  const remove = async (row: any) => {
    try {
      await fetchDeleteRegisterEntry(row.id)
      ElMessage.success('已删除')
      await load()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }

  function buildMatchRule(row: {
    host?: string
    path_prefix?: string
    status?: string
  }) {
    const host = `${row.host || ''}`.trim() || '任意 Host'
    const pathPrefix = `${row.path_prefix || ''}`.trim() || '任意路径'
    return `${host} + ${pathPrefix}`
  }

  function buildVerifyUrl(row: { host?: string; path_prefix?: string }) {
    const host = `${row.host || ''}`.trim() || 'localhost'
    const pathPrefix = `${row.path_prefix || ''}`.trim() || '/account/auth/register'
    return `https://${host}${pathPrefix}`
  }

  function resolvePublicRegisterTag(value: boolean | null | undefined) {
    if (value === true) return { label: '入口强制开启', type: 'success' as const }
    if (value === false) return { label: '入口强制关闭', type: 'warning' as const }
    return { label: '继承策略', type: 'info' as const }
  }

  function syncCurrentPage() {
    const totalPages = Math.max(1, Math.ceil((pagination.total || 0) / pagination.size))
    if (pagination.current > totalPages) {
      pagination.current = totalPages
    }
  }

  function handleSizeChange(size: number) {
    pagination.size = size
    pagination.current = 1
    syncCurrentPage()
  }

  function handleCurrentChange(current: number) {
    pagination.current = current
    syncCurrentPage()
  }

  onMounted(load)
</script>

<style scoped>
  .register-entry-page {
    display: flex;
    flex-direction: column;
    min-height: 0;
  }

  .register-entry-main {
    flex: 1;
    min-height: 0;
  }

  .register-entry-main :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .register-entry-header {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 4px 0;
  }

  .register-entry-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .register-entry-tip {
    color: var(--el-text-color-secondary);
    line-height: 1.7;
  }

  .table-actions {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .dialog-layout {
    display: grid;
    grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.8fr);
    gap: 20px;
  }

  .dialog-form {
    min-width: 0;
  }

  .dialog-preview {
    min-width: 0;
  }

  .preview-card {
    padding: 16px;
    border-radius: 16px;
    border: 1px solid var(--el-border-color-light);
    background: linear-gradient(180deg, rgb(248 250 252 / 95%), #fff);
  }

  .preview-title {
    margin-bottom: 12px;
    font-weight: 600;
    color: var(--el-text-color-primary);
  }

  .field-tip {
    margin-top: 6px;
    font-size: 12px;
    line-height: 1.6;
    color: var(--el-text-color-secondary);
  }
</style>
