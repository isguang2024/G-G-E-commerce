<!-- 登录页面 -->
<template>
  <div class="flex w-full h-screen">
    <LoginLeftView />

    <div class="relative flex-1">
      <AuthTopBar />

      <div class="auth-right-wrap">
        <div class="form">
          <h3 class="title">{{ $t('login.title') }}</h3>
          <p class="sub-title">{{ $t('login.subTitle') }}</p>
          <ElForm
            ref="formRef"
            :model="formData"
            :rules="rules"
            :key="formKey"
            @keyup.enter="handleSubmit"
            style="margin-top: 25px"
          >
            <ElFormItem prop="username">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.username') || '用户名'"
                v-model.trim="formData.username"
                type="text"
              />
            </ElFormItem>
            <ElFormItem prop="password">
              <ElInput
                class="custom-height"
                :placeholder="$t('login.placeholder.password')"
                v-model.trim="formData.password"
                type="password"
                autocomplete="off"
                show-password
              />
            </ElFormItem>

            <div class="flex-cb mt-2 text-sm">
              <ElCheckbox v-model="formData.rememberPassword">{{
                $t('login.rememberPwd')
              }}</ElCheckbox>
              <RouterLink class="text-theme" :to="{ name: 'ForgetPassword' }">{{
                $t('login.forgetPwd')
              }}</RouterLink>
            </div>

            <div style="margin-top: 30px">
              <ElButton
                class="w-full custom-height"
                type="primary"
                @click="handleSubmit"
                :loading="loading"
                v-ripple
              >
                {{ $t('login.btnText') }}
              </ElButton>
            </div>

            <div class="mt-5 text-sm text-gray-600">
              <span>{{ $t('login.noAccount') }}</span>
              <RouterLink class="text-theme" :to="{ name: 'Register' }">{{
                $t('login.register')
              }}</RouterLink>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import AppConfig from '@/config'
  import { useUserStore } from '@/store/modules/user'
  import {
    hasPersonalWorkspaceAccessByUserInfo,
    useCollaborationWorkspaceStore
  } from '@/store/modules/collaboration-workspace'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { useI18n } from 'vue-i18n'
  import { HttpError } from '@/utils/http/error'
  import { fetchLogin } from '@/api/auth'
  import { ElNotification, type FormInstance, type FormRules } from 'element-plus'
  import { resetRouterState } from '@/router/guards/beforeEach'
  import { RoutesAlias } from '@/router/routesAlias'

  defineOptions({ name: 'Login' })

  const { t, locale } = useI18n()
  const formKey = ref(0)

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const userStore = useUserStore()
  const collaborationWorkspaceStore = useCollaborationWorkspaceStore()
  const menuSpaceStore = useMenuSpaceStore()
  const router = useRouter()
  const route = useRoute()

  const systemName = AppConfig.systemInfo.name
  const formRef = ref<FormInstance>()
  const LOGIN_REMEMBER_KEY = 'gg-login-remember'

  // 登录表单默认值（不再预置系统账号密码）
  const formData = reactive({
    username: '',
    password: '',
    rememberPassword: false
  })

  const loadRememberedCredentials = () => {
    try {
      const raw = localStorage.getItem(LOGIN_REMEMBER_KEY)
      if (!raw) return
      const parsed = JSON.parse(raw) as {
        username?: string
        password?: string
        rememberPassword?: boolean
      }
      formData.username = parsed.username || ''
      formData.password = parsed.password || ''
      formData.rememberPassword = !!parsed.rememberPassword
    } catch (error) {
      console.warn('[Login] 读取记住密码失败，已忽略:', error)
      localStorage.removeItem(LOGIN_REMEMBER_KEY)
    }
  }

  const persistRememberedCredentials = () => {
    if (formData.rememberPassword) {
      localStorage.setItem(
        LOGIN_REMEMBER_KEY,
        JSON.stringify({
          username: formData.username,
          password: formData.password,
          rememberPassword: true
        })
      )
    } else {
      localStorage.removeItem(LOGIN_REMEMBER_KEY)
    }
  }

  const rules = computed<FormRules>(() => ({
    username: [{ required: true, message: t('login.placeholder.username'), trigger: 'blur' }],
    password: [{ required: true, message: t('login.placeholder.password'), trigger: 'blur' }]
  }))

  const loading = ref(false)

  const normalizeRedirect = (raw?: string) => {
    const cleanPath = `${raw || ''}`.trim()
    if (!cleanPath) return '/'

    let current = cleanPath
    const decodeOnce = (value: string) => {
      try {
        return decodeURIComponent(value)
      } catch {
        return value
      }
    }

    let safeIterations = 0
    while (current.includes('redirect=') && safeIterations < 5) {
      const redirectIndex = current.indexOf('redirect=')
      current = decodeOnce(current.slice(redirectIndex + 'redirect='.length))
      safeIterations += 1
    }

    const normalized = decodeOnce(current).trim()
    if (normalized.startsWith('#/')) {
      return normalized.slice(1)
    }
    if (normalized.startsWith('/#/')) {
      return normalized.slice(2)
    }
    if (!normalized || !normalized.startsWith('/')) return '/'
    if (normalized.startsWith('/auth/login')) return '/'

    return normalized
  }

  const gotoAfterLogin = async (landingPath: string) => {
    const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(landingPath)
    if (nextTarget.mode === 'location') {
      window.location.assign(nextTarget.target)
      return
    }

    const fallbackUrl = () => {
      return new URL(router.resolve(landingPath).href, window.location.origin).toString()
    }
    const fallbackFallbackUrl = fallbackUrl()

    const hasJumpedOut = () => {
      const currentPath = router.currentRoute.value.path
      return currentPath !== RoutesAlias.Login
    }

    try {
      await router.replace(landingPath)
      await nextTick()
      if (!hasJumpedOut()) {
        setTimeout(() => {
          if (!hasJumpedOut()) {
            window.location.assign(fallbackFallbackUrl)
          }
        }, 900)
      }
    } catch (error) {
      console.warn('[Login] 登录导航失败，尝试兜底跳转:', error)
      window.location.assign(fallbackFallbackUrl)
    }
  }

  const safeInitLoginContext = async (
    preferredCollaborationWorkspaceId: string,
    preferredLegacyCollaborationWorkspaceId: string
  ) => {
    try {
      await collaborationWorkspaceStore.loadMyCollaborationWorkspaces({
        preferredCollaborationWorkspaceId,
        preferredLegacyCollaborationWorkspaceId,
        preferredWorkspaceId: `${userStore.getUserInfo.current_auth_workspace_id || ''}`,
        preferredWorkspaceType: `${userStore.getUserInfo.current_auth_workspace_type || ''}`,
        preferPersonalWorkspace: collaborationWorkspaceStore.hasPersonalWorkspaceAccess
      })
      menuSpaceStore.syncRuntimeHost()
      await menuSpaceStore.refreshRuntimeConfig(true)
      await menuSpaceStore.syncResolvedCurrentSpace()
    } catch (error) {
      console.warn('[Login] 登录初始化上下文失败，仍允许进入应用:', error)
    }
  }

  // 登录
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      const valid = await formRef.value.validate()
      if (!valid) return

      loading.value = true

      const { username, password } = formData
      const response = await fetchLogin({
        username,
        password
      })

      if (!response.access_token) {
        throw new Error('Login failed - no token received')
      }

      resetRouterState(0)
      userStore.setToken(response.access_token, response.refresh_token)
      userStore.setLoginStatus(true)

      if (response.user) {
        userStore.syncLoginUserIdentity(response.user.id)
        const userInfo: Api.Auth.UserInfo = {
          ...response.user,
          userId: response.user.id,
          userName: response.user.username || response.user.email,
          avatar: response.user.avatar_url,
          roles: response.user.is_super_admin ? ['R_SUPER'] : ['R_USER'],
          buttons: [],
          actions: response.user.actions || []
        }
        userStore.setUserInfo(userInfo)
        collaborationWorkspaceStore.setPersonalWorkspaceAccess(
          hasPersonalWorkspaceAccessByUserInfo(userInfo)
        )
      }

      persistRememberedCredentials()
      const displayName =
        response.user?.nickname || response.user?.username || response.user?.email || systemName
      showLoginSuccessNotice(displayName)

      await safeInitLoginContext(
        response.user?.current_collaboration_workspace_id || '',
        response.user?.collaboration_workspace_id ||
          response.user?.current_collaboration_workspace_id ||
          ''
      )
      const landingPath = normalizeRedirect(route.query.redirect as string)
      await gotoAfterLogin(landingPath)
    } catch (error) {
      if (error instanceof HttpError) {
        // handle silently
      } else {
        console.error('[Login] Unexpected error:', error)
      }
    } finally {
      loading.value = false
    }
  }

  // 登录成功提示
  const showLoginSuccessNotice = (displayName: string) => {
    setTimeout(() => {
      ElNotification({
        title: t('login.success.title'),
        type: 'success',
        duration: 2500,
        zIndex: 10000,
        message: `${t('login.success.message')}, ${displayName}!`
      })
    }, 1000)
  }

  onMounted(() => {
    loadRememberedCredentials()
  })
</script>

<style scoped>
  @import './style.css';
</style>

<style lang="scss" scoped>
  :deep(.el-select__wrapper) {
    height: 40px !important;
  }
</style>
