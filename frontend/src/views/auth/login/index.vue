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

            <!-- 推拽验证 -->
            <div class="relative pb-5 mt-6">
              <div
                class="relative z-[2] overflow-hidden select-none rounded-lg border border-transparent tad-300"
                :class="{ '!border-[#FF4E4F]': !isPassing && isClickPass }"
              >
                <ArtDragVerify
                  ref="dragVerify"
                  v-model:value="isPassing"
                  :text="$t('login.sliderText')"
                  textColor="var(--art-gray-700)"
                  :successText="$t('login.sliderSuccessText')"
                  progressBarBg="var(--main-color)"
                  :background="isDark ? '#26272F' : '#F1F1F4'"
                  handlerBg="var(--default-box-color)"
                />
              </div>
              <p
                class="absolute top-0 z-[1] px-px mt-2 text-xs text-[#f56c6c] tad-300"
                :class="{ 'translate-y-10': !isPassing && isClickPass }"
              >
                {{ $t('login.placeholder.slider') }}
              </p>
            </div>

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
  import { useTenantStore } from '@/store/modules/tenant'
  import { useI18n } from 'vue-i18n'
  import { HttpError } from '@/utils/http/error'
  import { fetchLogin } from '@/api/auth'
  import { ElNotification, type FormInstance, type FormRules } from 'element-plus'
  import { useSettingStore } from '@/store/modules/setting'

  defineOptions({ name: 'Login' })

  const settingStore = useSettingStore()
  const { isDark } = storeToRefs(settingStore)
  const { t, locale } = useI18n()
  const formKey = ref(0)

  // 监听语言切换，重置表单
  watch(locale, () => {
    formKey.value++
  })

  const dragVerify = ref()

  const userStore = useUserStore()
  const tenantStore = useTenantStore()
  const router = useRouter()
  const route = useRoute()
  const isPassing = ref(false)
  const isClickPass = ref(false)

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

  // 登录
  const handleSubmit = async () => {
    if (!formRef.value) return

    try {
      // 表单验证
      const valid = await formRef.value.validate()
      if (!valid) return

      // 拖拽验证
      if (!isPassing.value) {
        isClickPass.value = true
        return
      }

      loading.value = true

      // 登录请求（使用 username 作为登录凭证）
      const { username, password } = formData

      const response = await fetchLogin({
        username: username, // 使用 username 字段
        password
      })

      // 验证token
      if (!response.access_token) {
        throw new Error('Login failed - no token received')
      }

      // 存储 token 和登录状态
      userStore.setToken(response.access_token, response.refresh_token)
      userStore.setLoginStatus(true)

      // 存储用户信息（映射后端返回的数据）
      if (response.user) {
        const userInfo: Api.Auth.UserInfo = {
          ...response.user,
          // 兼容字段映射
          userId: response.user.id,
          userName: response.user.username || response.user.email,
          avatar: response.user.avatar_url,
          roles: response.user.is_super_admin ? ['R_SUPER'] : ['R_USER'],
          buttons: [],
          actions: response.user.actions || [],
          scoped_actions: response.user.scoped_actions || response.user.scopedActions || [],
          scopedActions: response.user.scoped_actions || response.user.scopedActions || []
        }
        userStore.setUserInfo(userInfo)
      }

      await tenantStore.loadMyTeams()

      // 登录成功处理
      persistRememberedCredentials()
      const displayName =
        response.user?.nickname || response.user?.username || response.user?.email || systemName
      showLoginSuccessNotice(displayName)

      // 获取 redirect 参数，如果存在则跳转到指定页面，否则跳转到首页
      const redirect = route.query.redirect as string
      router.push(redirect || '/')
    } catch (error) {
      // 处理 HttpError
      if (error instanceof HttpError) {
        // console.log(error.code)
      } else {
        // 处理非 HttpError
        // ElMessage.error('登录失败，请稍后重试')
        console.error('[Login] Unexpected error:', error)
      }
    } finally {
      loading.value = false
      resetDragVerify()
    }
  }

  // 重置拖拽验证
  const resetDragVerify = () => {
    dragVerify.value.reset()
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
