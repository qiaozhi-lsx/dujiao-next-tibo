<template>
  <div class="tibo-scope" :class="{ 'is-product': isProductPage }">
    <!-- 顶栏 -->
    <header class="j8-header">
      <RouterLink class="j8-brand" to="/" :title="brandName">
        <span class="j8-brand-mark">
          <img v-if="brandLogo" :src="brandLogo" :alt="brandName" />
          <b v-else>{{ brandInitial }}</b>
        </span>
        <span class="j8-brand-text">
          <strong>{{ brandName }}</strong>
          <small>{{ brandDomain }}</small>
        </span>
      </RouterLink>

      <nav class="j8-nav" aria-label="主导航">
        <RouterLink to="/" :class="{ 'is-active': route.path === '/' }">
          <ShoppingBag />{{ t('tibo.nav.shop') }}
        </RouterLink>
        <RouterLink :to="ordersPath" :class="{ 'is-active': route.path.startsWith('/guest/orders') || route.path.startsWith('/me/orders') }">
          <ClipboardList />{{ t('tibo.nav.orders') }}
        </RouterLink>
        <RouterLink v-if="blogEnabled" to="/blog" :class="{ 'is-active': route.path.startsWith('/blog') }">
          <BookOpen />{{ t('tibo.nav.tutorial') }}
        </RouterLink>
        <RouterLink to="/2fa" :class="{ 'is-active': route.path === '/2fa' }">
          <ShieldCheck />{{ t('tibo.nav.tfa') }}
        </RouterLink>
        <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer">
          <Ticket />{{ t('tibo.nav.support') }}
        </a>
        <RouterLink v-if="noticeEnabled" to="/notice" :class="{ 'is-active': route.path === '/notice' }">
          <Megaphone />{{ t('tibo.nav.notice') }}
        </RouterLink>
        <RouterLink v-if="aboutEnabled" to="/about" :class="{ 'is-active': route.path === '/about' }">
          <Info />{{ t('tibo.nav.about') }}
        </RouterLink>
        <RouterLink :to="accountPath" :class="{ 'is-active': route.path.startsWith('/me') || route.path.startsWith('/auth') }">
          <User />{{ t('tibo.nav.account') }}
        </RouterLink>
      </nav>

      <div class="j8-actions">
        <RouterLink v-if="noticeEnabled" to="/notice" class="j8-icon-btn j8-only-mobile" :aria-label="t('tibo.nav.notice')">
          <Megaphone />
        </RouterLink>
        <RouterLink to="/2fa" class="j8-icon-btn j8-only-mobile" :aria-label="t('tibo.nav.tfa')" :title="t('tibo.nav.tfa')">
          <KeyRound />
        </RouterLink>
        <div class="j8-lang" ref="langEl">
          <button type="button" class="j8-icon-btn" :aria-label="t('navbar.selectLanguage')" @click="langOpen = !langOpen">
            <Languages />
          </button>
          <div v-if="langOpen" class="j8-popover">
            <button v-for="lang in languages" :key="lang.code" type="button" :class="{ 'is-active': appStore.locale === lang.code }" @click="changeLanguage(lang.code)">
              {{ lang.name }}
              <span v-if="appStore.locale === lang.code" class="lang-dot"></span>
            </button>
          </div>
        </div>
        <button type="button" class="j8-icon-btn" :aria-label="t('resellerConsole.common.toggleTheme')" @click="toggleTheme">
          <Sun v-if="theme === 'dark'" />
          <Moon v-else />
        </button>
        <RouterLink to="/cart" class="j8-icon-btn" :aria-label="t('navbar.cart')">
          <ShoppingCart />
          <span v-if="cartCount > 0" class="j8-cart-count">{{ cartCount }}</span>
        </RouterLink>
        <RouterLink :to="accountPath" class="j8-account-btn j8-only-desktop">
          <User />
          <span>{{ userAuthStore.isAuthenticated ? t('navbar.personalCenter') : t('navbar.login') }}</span>
        </RouterLink>
      </div>
    </header>

    <!-- 页面内容 -->
    <main class="j8-main">
      <slot />
    </main>

    <!-- 页脚 -->
    <footer class="j8-footer">
      <div>
        <span class="j8-brand-mark small">
          <img v-if="brandLogo" :src="brandLogo" :alt="brandName" />
          <b v-else>{{ brandInitial }}</b>
        </span>
        <p>
          <b>{{ brandName }}</b>
          <small>© {{ year }} {{ brandName }} · {{ t('tibo.footer.rights') }}</small>
        </p>
      </div>
      <nav aria-label="页脚导航">
        <RouterLink :to="ordersPath">{{ t('tibo.nav.orders') }}</RouterLink>
        <RouterLink v-if="blogEnabled" to="/blog">{{ t('tibo.nav.tutorial') }}</RouterLink>
        <RouterLink to="/2fa">{{ t('tibo.nav.tfa') }}</RouterLink>
        <RouterLink v-if="noticeEnabled" to="/notice">{{ t('tibo.nav.notice') }}</RouterLink>
        <RouterLink v-if="aboutEnabled" to="/about">{{ t('tibo.nav.about') }}</RouterLink>
        <RouterLink to="/terms">{{ t('footer.terms') }}</RouterLink>
        <RouterLink to="/privacy">{{ t('footer.privacy') }}</RouterLink>
        <a v-if="supportHref" :href="supportHref" target="_blank" rel="noopener noreferrer">{{ t('tibo.nav.support') }}</a>
      </nav>
    </footer>

    <!-- 悬浮客服 -->
    <button
      v-if="contacts.length"
      type="button"
      class="j8-support j8-support--contact"
      :aria-expanded="contactOpen"
      :aria-label="t('tibo.contact.title')"
      @click="contactOpen = !contactOpen"
    >
      <Headset />
      <span>{{ t('tibo.contact.short') }}</span>
    </button>
    <a v-else-if="supportHref" class="j8-support" :href="supportHref" target="_blank" rel="noopener noreferrer" :aria-label="t('tibo.contact.title')">
      <Headset />
      <span>{{ t('tibo.contact.short') }}</span>
    </a>

    <div v-if="contactOpen" class="j8-contact-panel">
      <header>
        <span class="j8-contact-badge"><Headset /></span>
        <strong>{{ t('tibo.contact.title') }}</strong>
        <button type="button" class="j8-contact-close" :aria-label="t('common.cancel')" @click="contactOpen = false"><X /></button>
      </header>
      <ul class="j8-contact-list">
        <li v-for="c in contacts" :key="c.key">
          <span class="j8-contact-icon" :class="`is-${c.key}`" v-html="c.icon"></span>
          <span class="j8-contact-text">
            <span>{{ c.label }}</span>
            <code>{{ c.value }}</code>
          </span>
          <a v-if="c.href" :href="c.href" target="_blank" rel="noopener noreferrer">{{ t('tibo.contact.open') }}</a>
        </li>
      </ul>
      <p>{{ t('tibo.contact.hint') }}</p>
    </div>

    <!-- 移动端底部导航 -->
    <nav class="j8-bottom-nav" style="grid-template-columns: repeat(5, 1fr)">
      <RouterLink to="/" :class="{ 'is-active': route.path === '/' }"><ShoppingBag /><span>{{ t('tibo.nav.shop') }}</span></RouterLink>
      <RouterLink :to="ordersPath" :class="{ 'is-active': route.path.includes('/orders') && !route.path.includes('/guest/orders') }"><ClipboardList /><span>{{ t('tibo.nav.orders') }}</span></RouterLink>
      <RouterLink to="/cart"><ShoppingCart /><span>{{ t('navbar.cart') }}</span><span v-if="cartCount > 0" class="j8-cart-count">{{ cartCount }}</span></RouterLink>
      <RouterLink v-if="noticeEnabled" to="/notice" :class="{ 'is-active': route.path === '/notice' }"><Megaphone /><span>{{ t('tibo.nav.notice') }}</span></RouterLink>
      <RouterLink v-else to="/2fa" :class="{ 'is-active': route.path === '/2fa' }"><ShieldCheck /><span>{{ t('tibo.nav.tfa') }}</span></RouterLink>
      <RouterLink :to="accountPath" :class="{ 'is-active': route.path.startsWith('/me') || route.path.startsWith('/auth') }"><User /><span>{{ t('tibo.nav.account') }}</span></RouterLink>
    </nav>

    <!-- 回到顶部 -->
    <Transition name="page-fade">
      <button v-if="showBackTop" type="button" class="back-to-top back-to-top--j8" @click="scrollTop">
        <ArrowUp />
      </button>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  ArrowUp, BookOpen, ClipboardList, Headset, Info, KeyRound, Languages, Megaphone,
  Moon, ShieldCheck, ShoppingBag, ShoppingCart, Sun, Ticket, User, X,
} from 'lucide-vue-next'
import { useAppStore } from '../../../stores/app'
import { useCartStore } from '../../../stores/cart'
import { useUserAuthStore } from '../../../stores/userAuth'
import { useNavConfig } from '../../../composables/useNavConfig'
import { useTheme } from '../../../utils/theme'
import { getImageUrl } from '../../../utils/image'
import '@fontsource-variable/inter'
import '../styles/ji8.css'

const route = useRoute()
const { t } = useI18n()
const appStore = useAppStore()
const cartStore = useCartStore()
const userAuthStore = useUserAuthStore()
const { theme, toggleTheme } = useTheme()
const { aboutEnabled, blogEnabled, noticeEnabled } = useNavConfig()

const langOpen = ref(false)
const contactOpen = ref(false)
const showBackTop = ref(false)
const langEl = ref<HTMLElement | null>(null)
const year = new Date().getFullYear()

const brandName = computed(() => String(appStore.config?.brand?.site_name || '').trim() || 'D&J Studio')
const brandDomain = computed(() => {
  const domain = String(appStore.config?.brand?.site_domain || '').trim()
  if (domain) return domain.replace(/^https?:\/\//, '')
  try {
    return window.location.host ? window.location.host.toUpperCase() : ''
  } catch {
    return ''
  }
})
const brandInitial = computed(() => brandName.value.slice(0, 1))
const brandLogo = computed(() => {
  const raw = String(appStore.config?.brand?.site_logo || '').trim()
  return raw ? getImageUrl(raw) : ''
})

const isProductPage = computed(() => route.name === 'product-detail')
const ordersPath = computed(() => (userAuthStore.isAuthenticated ? '/me/orders' : '/guest/orders'))
const accountPath = computed(() => (userAuthStore.isAuthenticated ? '/me' : '/auth/login'))
const cartCount = computed(() => cartStore.totalItems)

const contact = computed(() => appStore.config?.contact as { telegram?: string; whatsapp?: string; email?: string; support_url?: string } | undefined)
const supportHref = computed(() => {
  const c = contact.value
  if (c?.support_url) return c.support_url
  if (c?.telegram) return c.telegram
  if (c?.whatsapp) return `https://wa.me/${c.whatsapp.replace(/[^\d]/g, '')}`
  if (c?.email) return `mailto:${c.email}`
  return ''
})

const ICONS = {
  telegram: '<svg viewBox="0 0 24 24"><path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/></svg>',
  whatsapp: '<svg viewBox="0 0 24 24"><path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 0 1-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 0 1-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 0 1 2.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0 0 12.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 0 0 5.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 0 0-3.48-8.413Z"/></svg>',
  email: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="4" width="20" height="16" rx="2"/><path d="m22 7-8.97 5.7a1.94 1.94 0 0 1-2.06 0L2 7"/></svg>',
}

const contacts = computed(() => {
  const c = contact.value
  if (!c) return []
  const out: { key: string; label: string; value: string; href?: string; icon: string }[] = []
  if (c.telegram) {
    const href = c.telegram.startsWith('http') ? c.telegram : `https://t.me/${c.telegram.replace(/^@/, '')}`
    out.push({ key: 'telegram', label: 'Telegram', value: c.telegram.replace(/^https?:\/\/t\.me\//, '@'), href, icon: ICONS.telegram })
  }
  if (c.whatsapp) {
    out.push({ key: 'whatsapp', label: 'WhatsApp', value: c.whatsapp, href: `https://wa.me/${c.whatsapp.replace(/[^\d]/g, '')}`, icon: ICONS.whatsapp })
  }
  if (c.email) {
    out.push({ key: 'email', label: t('footer.email'), value: c.email, href: `mailto:${c.email}`, icon: ICONS.email })
  }
  return out
})

const languages = [
  { code: 'zh-CN', name: '简体中文' },
  { code: 'zh-TW', name: '繁體中文' },
  { code: 'en-US', name: 'English' },
]

const changeLanguage = (code: string) => {
  appStore.setLocale(code)
  langOpen.value = false
}

const scrollTop = () => window.scrollTo({ top: 0, behavior: 'smooth' })

const onDocClick = (e: MouseEvent) => {
  const target = e.target as Node
  if (langOpen.value && langEl.value && !langEl.value.contains(target)) langOpen.value = false
}
const onScroll = () => {
  showBackTop.value = window.scrollY > 400
}

onMounted(() => {
  document.addEventListener('click', onDocClick)
  window.addEventListener('scroll', onScroll, { passive: true })
  // 浮层（Toast / ConfirmDialog / 公告弹窗）Teleport 到 body，靠这个 class 拿到 ji8 配色
  document.body.classList.add('tibo-tokens')
})
onUnmounted(() => {
  document.removeEventListener('click', onDocClick)
  window.removeEventListener('scroll', onScroll)
  document.body.classList.remove('tibo-tokens')
})
</script>

<style scoped>
.j8-header :deep(svg),
.j8-nav :deep(svg),
.j8-actions :deep(svg) {
  width: 16px;
  height: 16px;
}

.j8-actions :deep(svg) {
  width: 18px;
  height: 18px;
}

.j8-support :deep(svg) {
  width: 22px;
  height: 22px;
}

.j8-bottom-nav :deep(svg) {
  width: 18px;
  height: 18px;
}

.j8-contact-icon :deep(svg) {
  width: 22px;
  height: 22px;
  fill: currentColor;
}

.lang-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--accent);
}

.back-to-top--j8 {
  position: fixed;
  right: 18px;
  bottom: 158px;
  z-index: 46;
  width: 52px;
  height: 52px;
  display: grid;
  place-items: center;
  border-radius: 50%;
  border: 1px solid var(--line);
  background: var(--surface);
  color: var(--accent-strong);
  box-shadow: var(--shadow);
  cursor: pointer;
  transition: border-color 0.15s;
}

.back-to-top--j8:hover {
  border-color: var(--accent-line);
}

.back-to-top--j8 :deep(svg) {
  width: 20px;
  height: 20px;
}

@media (max-width: 760px) {
  .back-to-top--j8 {
    right: 10px;
    bottom: calc(148px + env(safe-area-inset-bottom));
    width: 46px;
    height: 46px;
  }
}
</style>
