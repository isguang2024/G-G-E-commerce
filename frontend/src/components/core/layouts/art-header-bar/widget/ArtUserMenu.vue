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
        <div v-if="workspaceList.length" class="team-switcher-wrap">
          <ArtTenantSwitcher compact />
        </div>
        <ul class="py-4 mt-3 border-t border-g-300/80">
          <li
            v-if="
              hasPlatformAccess && currentAuthWorkspaceType !== 'personal' && personalWorkspace?.id
            "
            class="btn-item"
            @click="enterPlatformManagement"
          >
            <ArtSvgIcon icon="ri:building-line" />
            <span>切换到个人工作空间</span>
          </li>
          <li class="btn-item" @click="goPage('/user-center')">
            <ArtSvgIcon icon="ri:user-3-line" />
            <span>{{ $t('topBar.user.userCenter') }}</span>
          </li>
          <li class="btn-item" @click="refreshPermissionsAndMenus">
            <ArtSvgIcon icon="ri:refresh-line" />
            <span>刷新状态</span>
          </li>
          <li class="btn-item btn-item--context">
            <ArtSvgIcon icon="ri:briefcase-4-line" />
            <span>
              {{
                currentAuthWorkspaceType === 'collaboration'
                  ? '当前授权工作空间：协作空间'
                  : '当前授权工作空间：个人工作空间'
              }}
            </span>
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
  import { useWorkspaceStore } from '@/store/modules/workspace'
  import { useMenuStore } from '@/store/modules/menu'
  import { useMenuSpaceStore } from '@/store/modules/menu-space'
  import {
    refreshCurrentUserInfoContext,
    refreshUserMenus,
    refreshUserAccessAndMenus
  } from '@/router'
  import ArtTenantSwitcher from './ArtTenantSwitcher.vue'
  import { findRegisteredRouteByPath } from '@/utils/router'

  defineOptions({ name: 'ArtUserMenu' })

  const router = useRouter()
  const { t } = useI18n()
  const userStore = useUserStore()
  const collaborationWorkspaceStore = useTenantStore()
  const workspaceStore = useWorkspaceStore()
  const menuStore = useMenuStore()
  const menuSpaceStore = useMenuSpaceStore()

  const { getUserInfo: userInfo } = storeToRefs(userStore)
  const { hasPlatformAccess } = storeToRefs(collaborationWorkspaceStore)
  const { workspaceList, personalWorkspace, currentAuthWorkspaceType } = storeToRefs(workspaceStore)
  const userMenuPopover = ref()

  const resolveNavigationTarget = (
    path: string,
    routeCandidates: string[] = []
  ): { mode: 'router' | 'location'; target: string } => {
    const targetCandidates = Array.from(
      new Set([path, ...routeCandidates].filter((item) => `${item || ''}`.trim()))
    )
    for (const candidate of targetCandidates) {
      const routeRecord = findRegisteredRouteByPath(router, candidate)
      if (routeRecord) {
        return menuSpaceStore.resolveSpaceNavigationTarget(
          candidate,
          `${routeRecord.meta?.spaceKey || ''}`.trim() || undefined
        )
      }
    }
    const routeName = 'UserCenter'
    if (router.hasRoute(routeName) && path === '/user-center') {
      const resolvedByName = router.resolve({ name: routeName })
      if (resolvedByName?.path) {
        const routeRecord = findRegisteredRouteByPath(router, resolvedByName.path)
        return menuSpaceStore.resolveSpaceNavigationTarget(
          resolvedByName.path,
          `${routeRecord?.meta?.spaceKey || ''}`.trim() || undefined
        )
      }
    }
    const fallbackTarget = menuSpaceStore.resolveSpaceNavigationTarget(path)
    return fallbackTarget
  }

  const resolveUserCenterNavigationTarget = (): { mode: 'router' | 'location'; target: string } => {
    const candidatePath = '/dashboard/console/user-center'
    const routeRecord = findRegisteredRouteByPath(router, candidatePath)
    if (routeRecord) {
      return menuSpaceStore.resolveSpaceNavigationTarget(
        candidatePath,
        `${routeRecord.meta?.spaceKey || ''}`.trim() || undefined
      )
    }
    if (router.hasRoute('UserCenter')) {
      const resolvedByName = router.resolve({ name: 'UserCenter' })
      if (resolvedByName?.path) {
        const resolvedRecord = findRegisteredRouteByPath(router, resolvedByName.path)
        return menuSpaceStore.resolveSpaceNavigationTarget(
          resolvedByName.path,
          `${resolvedRecord?.meta?.spaceKey || ''}`.trim() || undefined
        )
      }
    }
    return menuSpaceStore.resolveSpaceNavigationTarget(candidatePath)
  }

  const navigateByTarget = (target: { mode: 'router' | 'location'; target: string }): void => {
    if (target.mode === 'router') {
      router.push(target.target)
      return
    }
    window.location.assign(target.target)
  }

  const goPage = (path: string): void => {
    closeUserMenu()
    if (path === '/user-center') {
      const nextTarget = resolveUserCenterNavigationTarget()
      navigateByTarget(nextTarget)
      return
    }
    const nextTarget = resolveNavigationTarget(path)
    navigateByTarget(nextTarget)
  }

  const enterPlatformManagement = async (): Promise<void> => {
    const resolveLandingTarget = () => {
      const landingPath = menuStore.getHomePath() || '/'
      const target = resolveNavigationTarget(landingPath, [
        '/dashboard/console',
        '/workspace/inbox',
        '/'
      ])
      if (target) {
        return target
      }
      return menuSpaceStore.resolveSpaceNavigationTarget('/')
    }

    if (currentAuthWorkspaceType.value === 'personal') {
      closeUserMenu()
      const nextTarget = resolveLandingTarget()
      if (nextTarget.mode === 'router') {
        router.push(nextTarget.target)
        return
      }
      window.location.assign(nextTarget.target)
      return
    }
    closeUserMenu()
    if (!personalWorkspace.value?.id) return
    await workspaceStore.switchWorkspace(personalWorkspace.value.id)
    await refreshCurrentUserInfoContext()
    await refreshUserMenus()
    const nextTarget = resolveLandingTarget()
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
    } catch {
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

  .btn-item--context {
    cursor: default;
    color: var(--art-text-muted);
  }

  .btn-item--context:hover {
    background-color: transparent;
  }
</style>

