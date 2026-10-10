<template>
  <div class="tibo-scope store-layout">
    <!-- 顶栏 -->
    <header class="site-header">
      <div class="header-inner">
        <RouterLink class="brand" to="/" :aria-label="brandName">
          <img v-if="brandLogo" :src="brandLogo" :alt="brandName" class="shop-brand-logo" />
          <svg v-else viewBox="0 0 32 32" fill="none" aria-hidden="true">
            <path d="M9 24V13c0-4 3-7 7-7M16 26V13c0-4 3-7 7-7M23 24V15" stroke="currentColor" stroke-width="3.5" stroke-linecap="round" />
          </svg>
          <span class="shop-brand-name">{{ brandName }}</span>
        </RouterLink>

        <nav aria-label="主导航">
          <RouterLink to="/" class="nav-link" :class="{ selected: route.path === '/' }">{{ t('products.allCategories') }}</RouterLink>
          <RouterLink :to="ordersPath" class="nav-link" active-class="selected">{{ t('tibo.nav.myOrders') }}</RouterLink>
          <RouterLink v-if="aboutEnabled" to="/about" class="nav-link" active-class="selected">{{ t('tibo.nav.help') }}</RouterLink>
        </nav>

        <div class="header-actions">
          <!-- 语言切换 -->
          <div class="lang-wrap" ref="langEl">
            <button type="button" class="lang-button" :aria-label="t('navbar.selectLanguage')" @click="langOpen = !langOpen">
              {{ currentLangLabel }}
            </button>
            <div v-if="langOpen" class="lang-menu">
              <button v-for="lang in languages" :key="lang.code" type="button" class="lang-menu-item" :class="{ active: appStore.locale === lang.code }" @click="changeLanguage(lang.code)">
                {{ lang.name }}
                <span v-if="appStore.locale === lang.code" class="lang-dot"></span>
              </button>
            </div>
          </div>

          <!-- 主题 -->
          <button type="button" class="theme-button" :aria-label="t('resellerConsole.common.toggleTheme')" @click="toggleTheme">
            <Sun v-if="theme === 'dark'" class="ha-icon" />
            <Moon v-else class="ha-icon" />
          </button>

          <!-- 联系支持 -->
          <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer" class="support-button">
            <LifeBuoy class="ha-icon" /><span>{{ t('tibo.nav.support') }}</span>
          </a>

          <!-- 登录 / 个人中心 -->
          <RouterLink v-if="!userAuthStore.isAuthenticated" to="/auth/login" class="login-link">{{ t('navbar.login') }}</RouterLink>
          <RouterLink v-else to="/me" class="me-link"><User class="ha-icon" /><span>{{ t('navbar.personalCenter') }}</span></RouterLink>
        </div>
      </div>
    </header>

    <!-- 页面内容 -->
    <main>
      <slot />
    </main>

    <!-- 页脚 -->
    <footer class="site-footer">
      <div class="page-width">
        <div class="footer-top">
          <RouterLink class="footer-brand" to="/">
            <img v-if="brandLogo" :src="brandLogo" :alt="brandName" />
            <svg v-else viewBox="0 0 32 32" fill="none" aria-hidden="true">
              <path d="M9 24V13c0-4 3-7 7-7M16 26V13c0-4 3-7 7-7M23 24V15" stroke="currentColor" stroke-width="3.5" stroke-linecap="round" />
            </svg>
            <span>{{ brandName }}</span>
          </RouterLink>
          <span class="footer-tagline">{{ brandDescription || t('tibo.footer.tagline') }}</span>
          <button type="button" class="text-button" @click="toggleTheme">
            <Sun v-if="theme === 'light'" class="ft-icon" />
            <Moon v-else class="ft-icon" />
            {{ theme === 'light' ? t('tibo.footer.darkAppearance') : t('tibo.footer.lightAppearance') }}
          </button>
        </div>
        <div class="footer-bottom">
          <div>
            <span>© {{ year }} {{ brandName }}</span>
          </div>
          <div>
            <RouterLink to="/terms">{{ t('footer.terms') }}</RouterLink>
            <RouterLink to="/privacy">{{ t('footer.privacy') }}</RouterLink>
            <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer">{{ t('tibo.nav.support') }}</a>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { LifeBuoy, Moon, Sun, User } from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useUserAuthStore } from '../../../stores/userAuth'
import { useNavConfig } from '../../../composables/useNavConfig'
import { useTheme } from '../../../utils/theme'
import { getImageUrl } from '../../../utils/image'
import '../styles/tibo.css'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const userAuthStore = useUserAuthStore()
const { theme, toggleTheme } = useTheme()

const langOpen = ref(false)
const langEl = ref<HTMLElement | null>(null)
const year = new Date().getFullYear()

const brandName = computed(() => String(appStore.config?.brand?.site_name || '').trim() || 'D&J Studio')
const brandLogo = computed(() => {
  const raw = String(appStore.config?.brand?.site_logo || '').trim()
  return raw ? getImageUrl(raw) : ''
})
const brandDescription = computed(() => {
  const desc = appStore.config?.brand?.site_description
  if (desc && typeof desc === 'object') {
    const val = (desc as Record<string, string>)[appStore.locale] || (desc as Record<string, string>)['zh-CN'] || ''
    return typeof val === 'string' ? val.trim() : ''
  }
  return ''
})

const { aboutEnabled } = useNavConfig()

const ordersPath = computed(() => (userAuthStore.isAuthenticated ? '/me/orders' : '/guest/orders'))

const contact = computed(() => appStore.config?.contact as { telegram?: string; email?: string; support_url?: string } | undefined)
const supportHref = computed(() => {
  const c = contact.value
  if (c?.support_url) return c.support_url
  if (c?.telegram) return c.telegram
  if (c?.email) return `mailto:${c.email}`
  return ''
})

const languages = [
  { code: 'zh-CN', name: '简体中文', short: '中' },
  { code: 'zh-TW', name: '繁體中文', short: '繁' },
  { code: 'en-US', name: 'English', short: 'EN' },
]
const currentLangLabel = computed(() => languages.find((l) => l.code === appStore.locale)?.short || '中')

const changeLanguage = (code: string) => {
  appStore.setLocale(code)
  langOpen.value = false
}

const onDocClick = (e: MouseEvent) => {
  const target = e.target as Node
  if (langOpen.value && langEl.value && !langEl.value.contains(target)) langOpen.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
  // 浮层（Toast / ConfirmDialog / 公告弹窗）Teleport 到 body，靠这个 class 拿到 tibo 配色
  document.body.classList.add('tibo-tokens')
})
onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  document.body.classList.remove('tibo-tokens')
})
</script>

<style scoped>
.ha-icon {
  width: 17px;
  height: 17px;
}

.ft-icon {
  width: 13px;
  height: 13px;
}

.support-button,
.me-link {
  white-space: nowrap;
}

.login-link {
  display: inline-flex;
  align-items: center;
  height: 34px;
  padding: 0 14px;
  border-radius: 22px;
  background: var(--primary);
  color: var(--primary-foreground) !important;
  font-size: 13px;
}

.lang-wrap {
  position: relative;
}

.lang-menu {
  position: absolute;
  right: 0;
  top: calc(100% + 8px);
  z-index: 60;
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 148px;
  padding: 8px;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--card);
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.1);
}

.lang-menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 9px 12px;
  border: 0;
  border-radius: 9px;
  background: transparent;
  color: var(--muted-foreground);
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
}

.lang-menu-item:hover {
  background: var(--muted);
  color: var(--foreground);
}

.lang-menu-item.active {
  color: var(--primary);
}

.lang-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--primary);
}

.footer-tagline {
  flex: 1;
  min-width: 0;
}

@media (width <= 760px) {
  .support-button span,
  .me-link span {
    display: none;
  }

  .support-button,
  .me-link {
    width: 38px;
    padding: 0;
  }

  .footer-tagline {
    order: 3;
    width: 100%;
  }
}
</style>
