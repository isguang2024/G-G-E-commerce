<!-- 用户菜单 -->
<template>
  <ElPopover
    ref="userMenuPopover"
    placement="bottom-end"
    :width="240"
    :hide-after="0"
    :offset="10"
    trigger="click"
    :show-arrow="false"
    popper-class="user-menu-popover"
    popper-style="padding: 5px 16px;"
  >
    <template #reference>
      <img
        class="size-8.5 mr-5 c-p rounded-full max-sm:w-6.5 max-sm:h-6.5 max-sm:mr-[16px]"
        src="@imgs/user/avatar.webp"
        alt="avatar"
      />
    </template>
    <template #default>
      <div class="pt-3">
        <div class="flex-c pb-1 px-0">
          <img
            class="w-10 h-10 mr-3 ml-0 overflow-hidden rounded-full float-left"
            src="@imgs/user/avatar.webp"
          />
          <div class="w-[calc(100%-60px)] h-full">
            <span class="block text-sm font-medium text-g-800 truncate">{{
              userInfo.userName
            }}</span>
            <span class="block mt-0.5 text-xs text-g-500 truncate">{{ userInfo.email }}</span>
          </div>
        </div>
        <div v-if="teamList.length" class="team-switcher-wrap">
          <ArtTenantSwitcher compact />
        </div>
        <ul class="py-4 mt-3 border-t border-g-300/80">
          <li v-if="hasPlatformAccess" class="btn-item" @click="enterPlatformManagement">
            <ArtSvgIcon icon="ri:building-line" />
            <span>进入平台管理</span>
          </li>
          <li class="btn-item" @click="goPage('/user-center')">
            <ArtSvgIcon icon="ri:user-3-line" />
            <span>{{ $t('topBar.user.userCenter') }}</span>
          </li>
          <li class="btn-item" @click="refreshPermissionsAndMenus">
            <ArtSvgIcon icon="ri:refresh-line" />
            <span>刷新状态</span>
          </li>
          <div class="w-full h-px my-2 bg-g-300/80"></div>
          <div class="log-out c-p" @click="loginOut">
            {{ $t('topBar.user.logout') }}
          </div>
        </ul>
      </div>
    </template>
  </ElPopover>
</template>

<script setup lang="ts">
  import { useI18n } from 'vue-i18n'
  import { useRouter } from 'vue-router'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { useUserStore } from '@/store/modules/user'
  import { useTenantStore } from '@/store/modules/tenant'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import { refreshCurrentUserInfoContext, refreshUserMenus, refreshUserAccessAndMenus } from '@/router'
  import ArtTenantSwitcher from './ArtTenantSwitcher.vue'

  defineOptions({ name: 'ArtUserMenu' })

  const router = useRouter()
  const { t } = useI18n()
  const userStore = useUserStore()
  const tenantStore = useTenantStore()
  const menuSpaceStore = useMenuSpaceStore()

  const { getUserInfo: userInfo } = storeToRefs(userStore)
  const { teamList, hasPlatformAccess, currentContextMode } = storeToRefs(tenantStore)
  const userMenuPopover = ref()

  const goPage = (path: string): void => {
    closeUserMenu()
    const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(path)
    if (nextTarget.mode === 'router') {
      router.push(nextTarget.target)
      return
    }
    window.location.assign(nextTarget.target)
  }

  const enterPlatformManagement = async (): Promise<void> => {
    if (currentContextMode.value === 'platform') {
      closeUserMenu()
      const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
        menuSpaceStore.resolveSpaceLandingPath()
      )
      if (nextTarget.mode === 'router') {
        router.push(nextTarget.target)
        return
      }
      window.location.assign(nextTarget.target)
      return
    }
    closeUserMenu()
    tenantStore.enterPlatformContext()
    await refreshCurrentUserInfoContext()
    await refreshUserMenus()
    const nextTarget = menuSpaceStore.resolveSpaceNavigationTarget(
      menuSpaceStore.resolveSpaceLandingPath()
    )
    if (nextTarget.mode === 'router') {
      router.push(nextTarget.target)
      return
    }
    window.location.assign(nextTarget.target)
  }

  const refreshPermissionsAndMenus = async (): Promise<void> => {
    closeUserMenu()
    try {
      await refreshUserAccessAndMenus()
      await router.replace({
        path: router.currentRoute.value.path,
        query: router.currentRoute.value.query,
        hash: router.currentRoute.value.hash
      })
      ElMessage.success('状态已刷新')
    } catch (error) {
      ElMessage.error('刷新状态失败')
    }
  }

  /**
   * 用户登出确认
   */
  const loginOut = (): void => {
    closeUserMenu()
    setTimeout(() => {
      ElMessageBox.confirm(t('common.logOutTips'), t('common.tips'), {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        customClass: 'login-out-dialog'
      }).then(() => {
        userStore.logOut()
      })
    }, 200)
  }

  /**
   * 关闭用户菜单弹出层
   */
  const closeUserMenu = (): void => {
    setTimeout(() => {
      userMenuPopover.value.hide()
    }, 100)
  }
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  @layer components {
    .btn-item {
      @apply flex items-center p-2 mb-3 select-none rounded-md cursor-pointer last:mb-0;

      span {
        @apply text-sm;
      }

      .art-svg-icon {
        @apply mr-2 text-base;
      }

      &:hover {
        background-color: var(--art-gray-200);
      }
    }
  }

  .log-out {
    @apply py-1.5
    mt-5
    text-xs
    text-center
    border
    border-g-400
    rounded-md
    transition-all
    duration-200
    hover:shadow-xl;
  }

  .team-switcher-wrap {
    padding-top: 14px;
    margin-top: 12px;
    border-top: 1px solid rgb(209 213 219 / 0.8);
  }
</style>
