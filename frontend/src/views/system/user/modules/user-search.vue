<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :rules="rules"
    @reset="handleReset"
    @search="handleSearch"
  >
  </ArtSearchBar>
</template>

<script setup lang="ts">
  import { fetchGetRoleList } from '@/api/system-manage'

  interface Props {
    modelValue: Record<string, any>
  }
  interface Emits {
    (e: 'update:modelValue', value: Record<string, any>): void
    (e: 'search', params: Record<string, any>): void
    (e: 'reset'): void
  }
  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  // 表单数据双向绑定
  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  // 校验规则
  const rules = {
    // userName: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
  }

  // 状态选项（与后端 status 一致：active / inactive）
  const statusOptions = [
    { label: '正常', value: 'active' },
    { label: '禁用', value: 'inactive' }
  ]

  // 角色列表选项
  const roleOptions = ref<Array<{ label: string; value: string }>>([])
  const roleLoading = ref(false)

  // 加载角色列表
  const loadRoles = async () => {
    try {
      roleLoading.value = true
      const res = await fetchGetRoleList({ current: 1, size: 1000 })
      const roles = res?.records || []
      roleOptions.value = roles.map((role: Api.SystemManage.RoleListItem) => ({
        label: `${role.roleName} (${role.roleCode})`,
        value: role.roleId // 使用角色ID而不是角色编码
      }))
    } catch (error) {
      console.error('加载角色列表失败:', error)
      roleOptions.value = []
    } finally {
      roleLoading.value = false
    }
  }

  // 组件挂载时加载角色列表
  onMounted(() => {
    loadRoles()
  })

  // 表单配置（后端列表接口支持 userName、status、roleCode 筛选）
  const formItems = computed(() => [
    {
      label: '用户名',
      key: 'userName',
      type: 'input',
      placeholder: '请输入用户名',
      clearable: true
    },
    {
      label: '手机号',
      key: 'userPhone',
      type: 'input',
      props: { placeholder: '请输入手机号', maxlength: 11 },
      clearable: true
    },
    {
      label: '邮箱',
      key: 'userEmail',
      type: 'input',
      props: { placeholder: '请输入邮箱' },
      clearable: true
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '请选择状态',
        options: statusOptions,
        clearable: true
      }
    },
    {
      label: '角色',
      key: 'roleId',
      type: 'select',
      props: {
        placeholder: '请选择角色',
        options: roleOptions.value,
        loading: roleLoading.value,
        clearable: true,
        filterable: true
      }
    }
  ])

  function handleReset() {
    emit('reset')
  }

  async function handleSearch() {
    await searchBarRef.value.validate()
    emit('search', formData.value)
  }
</script>
