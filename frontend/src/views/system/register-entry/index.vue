<template>
  <div class="p-4">
    <div class="mb-4 flex items-center justify-between">
      <h3 class="text-lg font-semibold">注册入口</h3>
      <ElButton type="primary" @click="openCreate">新建入口</ElButton>
    </div>
    <ElTable :data="list" border stripe>
      <ElTableColumn prop="entry_code" label="入口 Code" width="160" />
      <ElTableColumn prop="name" label="名称" width="180" />
      <ElTableColumn prop="app_key" label="App" width="140" />
      <ElTableColumn prop="host" label="Host" width="160" />
      <ElTableColumn prop="path_prefix" label="Path 前缀" width="200" />
      <ElTableColumn prop="policy_code" label="策略 Code" width="160" />
      <ElTableColumn prop="status" label="状态" width="100" />
      <ElTableColumn label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <ElButton link type="primary" @click="openEdit(row)">编辑</ElButton>
          <ElPopconfirm title="确认删除该入口？" @confirm="remove(row)">
            <template #reference>
              <ElButton link type="danger">删除</ElButton>
            </template>
          </ElPopconfirm>
        </template>
      </ElTableColumn>
    </ElTable>

    <ElDialog v-model="dialogVisible" :title="editing ? '编辑入口' : '新建入口'" width="620px">
      <ElForm :model="form" label-width="140px">
        <ElFormItem label="入口 Code" required>
          <ElInput v-model="form.entry_code" :disabled="!!editing" />
        </ElFormItem>
        <ElFormItem label="名称" required>
          <ElInput v-model="form.name" />
        </ElFormItem>
        <ElFormItem label="App Key" required>
          <ElInput v-model="form.app_key" placeholder="如 account-portal" />
        </ElFormItem>
        <ElFormItem label="Host">
          <ElInput v-model="form.host" placeholder="留空匹配任意 host" />
        </ElFormItem>
        <ElFormItem label="Path 前缀">
          <ElInput v-model="form.path_prefix" placeholder="如 /account/auth/register" />
        </ElFormItem>
        <ElFormItem label="Register Source">
          <ElInput v-model="form.register_source" placeholder="self / invite / ..." />
        </ElFormItem>
        <ElFormItem label="策略 Code" required>
          <ElInput v-model="form.policy_code" placeholder="如 default.self" />
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
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput v-model="form.remark" type="textarea" :rows="2" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton type="primary" @click="submit">保存</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { ElMessage } from 'element-plus'
  import {
    fetchCreateRegisterEntry,
    fetchDeleteRegisterEntry,
    fetchListRegisterEntries,
    fetchUpdateRegisterEntry
  } from '@/api/system-manage/register'

  defineOptions({ name: 'SystemRegisterEntry' })

  const list = ref<any[]>([])
  const dialogVisible = ref(false)
  const editing = ref<any>(null)

  const emptyForm = () => ({
    entry_code: '',
    name: '',
    app_key: 'account-portal',
    host: '',
    path_prefix: '',
    register_source: 'self',
    policy_code: '',
    status: 'enabled',
    sort_order: 0,
    allow_public_register: null,
    remark: ''
  })
  const form = reactive<any>(emptyForm())

  const load = async () => {
    try {
      const data: any = await fetchListRegisterEntries()
      list.value = data?.records || []
    } catch (e: any) {
      ElMessage.error(e?.message || '加载失败')
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

  const submit = async () => {
    try {
      const payload = { ...form }
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

  const remove = async (row: any) => {
    try {
      await fetchDeleteRegisterEntry(row.id)
      ElMessage.success('已删除')
      await load()
    } catch (e: any) {
      ElMessage.error(e?.message || '删除失败')
    }
  }

  onMounted(load)
</script>
