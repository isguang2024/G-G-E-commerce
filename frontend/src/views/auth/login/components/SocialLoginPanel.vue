<template>
  <div v-if="visible" class="social-login-panel">
    <el-alert
      v-if="capabilityBlocked"
      class="mb-3"
      type="warning"
      :closable="false"
      title="社交登录暂不可用"
      :description="capabilityReasonText"
    />
    <div class="social-divider">
      <span>其他登录方式</span>
    </div>

    <div v-if="safeItems.length > 0" class="social-items">
      <a
        v-for="item in safeItems"
        :key="item.key"
        class="social-item"
        :href="item.url"
        :target="item.target"
        rel="noopener noreferrer"
        :aria-disabled="capabilityBlocked"
        @click.prevent="capabilityBlocked ? undefined : void 0"
      >
        <img v-if="isImageUrl(item.icon)" class="social-item-icon-img" :src="item.icon" :alt="item.name" />
        <span v-else class="social-item-icon">{{ item.icon || '•' }}</span>
        <span class="social-item-name">{{ item.name }}</span>
      </a>
    </div>

    <div
      v-if="safeCustomHtml"
      class="social-custom-html"
      v-html="safeCustomHtml"
    />
  </div>
</template>

<script setup lang="ts">
  import { computed } from 'vue'
  import type { AuthTemplateSocial, AuthTemplateSocialItem } from '@/domains/auth/useAuthPageTemplate'

  interface SocialItemView {
    key: string
    name: string
    icon: string
    url: string
    target: '_self' | '_blank'
  }

  const props = withDefaults(
    defineProps<{
      enabled?: boolean
      social?: AuthTemplateSocial
      loginPageKey?: string
      pageScene?: 'login' | 'register' | 'forget_password'
    }>(),
    {
      enabled: false,
      social: () => ({}),
      loginPageKey: '',
      pageScene: 'login'
    }
  )

  function isImageUrl(icon: string): boolean {
    return /^https?:\/\//i.test(icon) || icon.startsWith('data:image/')
  }

  function toSafeItem(item: AuthTemplateSocialItem, index: number): SocialItemView | null {
    const key = `${item.key || `social-${index}`}`.trim()
    const name = `${item.name || item.key || '社交登录'}`.trim()
    const icon = `${item.icon || ''}`.trim()
    const url = `${item.url || ''}`.trim()
    if (!url) return null
    if (!/^https?:\/\//i.test(url) && !url.startsWith('/')) return null
    let finalURL = url
    if (/^\/auth\/oauth\/[^/]+\/authorize/i.test(url)) {
      const base = new URL(url, window.location.origin)
      base.searchParams.set('page_scene', props.pageScene || 'login')
      if (`${props.loginPageKey || ''}`.trim()) {
        base.searchParams.set('login_page_key', `${props.loginPageKey}`.trim())
      }
      base.searchParams.set('redirect_path', window.location.pathname)
      finalURL = `${base.pathname}${base.search}`
    }
    return {
      key,
      name,
      icon,
      url: finalURL,
      target: /^https?:\/\//i.test(url) ? '_blank' : '_self'
    }
  }

  function sanitizeHtml(input: string): string {
    if (!input) return ''
    let html = `${input}`
    html = html.replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, '')
    html = html.replace(/\son\w+\s*=\s*(['"]).*?\1/gi, '')
    html = html.replace(/\s(href|src)\s*=\s*(['"])\s*javascript:[^'"]*\2/gi, '')
    return html.trim()
  }

  const safeItems = computed<SocialItemView[]>(() => {
    const items = Array.isArray(props.social?.items) ? props.social?.items : []
    return items
      .map((item, idx) => toSafeItem(item || {}, idx))
      .filter((item): item is SocialItemView => Boolean(item))
  })

  const safeCustomHtml = computed(() => sanitizeHtml(`${props.social?.customHtml || ''}`))
  const capabilityBlocked = computed(() => props.social?.capability?.allow === false)
  const capabilityReason = computed(() => `${props.social?.capability?.reason || ''}`.trim())
  const hideWhenNoEnabledProvider = computed(() => capabilityReason.value === 'no_enabled_provider')
  const capabilityReasonText = computed(() => {
    const reason = capabilityReason.value
    if (reason === 'public_register_disabled') return '当前注册策略未开启公开注册，请联系管理员。'
    if (reason === 'no_enabled_provider') return '系统未启用任何社交登录提供方。'
    if (reason === 'provider_query_failed') return '暂时无法读取社交登录配置，请稍后重试。'
    return reason || '请联系管理员检查社交登录配置。'
  })

  const visible = computed(() => {
    if (hideWhenNoEnabledProvider.value) return false
    if (!props.enabled && !capabilityBlocked.value) return false
    return safeItems.value.length > 0 || Boolean(safeCustomHtml.value)
  })
</script>

<style scoped>
  .social-login-panel {
    margin-top: 18px;
  }

  .social-divider {
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    margin-bottom: 12px;
  }

  .social-divider::before,
  .social-divider::after {
    content: '';
    height: 1px;
    flex: 1;
    background: var(--el-border-color);
  }

  .social-items {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .social-item {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border: 1px solid var(--el-border-color);
    border-radius: 999px;
    padding: 6px 12px;
    text-decoration: none;
    color: var(--el-text-color-primary);
    background: #fff;
    transition: all 0.2s ease;
  }

  .social-item:hover {
    border-color: var(--auth-primary-color, var(--el-color-primary));
    color: var(--auth-primary-color, var(--el-color-primary));
  }

  .social-item[aria-disabled='true'] {
    opacity: 0.5;
    pointer-events: none;
    cursor: not-allowed;
  }

  .social-item-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    font-size: 12px;
    line-height: 1;
  }

  .social-item-icon-img {
    width: 18px;
    height: 18px;
    border-radius: 50%;
    object-fit: cover;
  }

  .social-item-name {
    font-size: 12px;
    line-height: 1;
  }

  .social-custom-html {
    margin-top: 12px;
  }
</style>
