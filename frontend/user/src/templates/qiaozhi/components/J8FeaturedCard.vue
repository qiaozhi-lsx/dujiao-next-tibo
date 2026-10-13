<template>
  <RouterLink
    :to="`/products/${product.slug}`"
    class="j8-fcard j8-reveal is-pending"
    :class="[`is-${variant}`, wideClass]"
    :data-reveal="1"
  >
    <div class="j8-fcard-head">
      <div>
        <h3>{{ title }}<em v-if="em">{{ em }}</em></h3>
        <p>{{ desc }}</p>
      </div>
      <span v-if="tag" class="j8-fcard-tag">{{ tag }}</span>
    </div>

    <!-- ========== 艺术区 ========== -->
    <div class="j8-fart" :class="`is-${variant}`" aria-hidden="true">
      <!-- ChatGPT 卡堆 -->
      <div v-if="variant === 'chatgpt'" class="fa-gpt">
        <div class="fa-gpt-card c3"><small>{{ shortTitle }}</small><b>{{ tier3.name }}</b><em>{{ tier3.price }}</em></div>
        <div class="fa-gpt-card c2"><small>{{ shortTitle }}</small><b>{{ tier2.name }}</b><em>{{ tier2.price }}</em></div>
        <div class="fa-gpt-card c1"><small>{{ shortTitle }}</small><b>{{ tier1.name }}</b><em>{{ tier1.price }}</em></div>
      </div>

      <!-- Gemini 星轨 -->
      <div v-else-if="variant === 'gemini'" class="fa-gem">
        <span class="fa-glow"></span>
        <span class="fa-gem-orbit"><i></i></span>
        <div class="fa-gem-star">
          <svg viewBox="0 0 160 160">
            <defs>
              <linearGradient id="j8-gem-grad" x1="0" y1="1" x2="1" y2="0">
                <stop stop-color="#4f8bff" /><stop offset=".55" stop-color="#8d6bff" /><stop offset="1" stop-color="#d65fb4" />
              </linearGradient>
            </defs>
            <path fill="url(#j8-gem-grad)" d="M80 5C91 48 112 69 155 80C112 91 91 112 80 155C69 112 48 91 5 80C48 69 69 48 80 5Z" />
          </svg>
        </div>
        <svg class="fa-twinkle t1" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <svg class="fa-twinkle t2" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <svg class="fa-twinkle t3" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <div v-if="artNumber" class="fa-num">
          <b><span class="fa-count" :style="{ '--to': artNumber }"></span><small>{{ artNumberUnit }}</small></b>
          <span>{{ artNumberLabel }}</span>
        </div>
      </div>

      <!-- X Premium 3D 卡 -->
      <div v-else-if="variant === 'x'" class="fa-x">
        <div class="fa-x-card">
          <div class="fa-x-logo">
            <svg viewBox="0 0 24 24" fill="currentColor">
              <path d="M18.9 2H22l-6.77 7.74L23.2 22h-6.24l-4.89-7.4L5.6 22H2.47l7.15-8.17L1.2 2h6.4l4.43 6.77L18.9 2ZM17.8 20h1.73L6.65 3.9H4.8L17.8 20Z" />
            </svg>
            <b>Premium</b>
          </div>
          <div class="fa-x-meta">
            <small>MEMBERSHIP</small>
            <em>{{ xDuration }}</em>
          </div>
          <i class="fa-x-sheen"></i>
        </div>
      </div>

      <!-- 多邻国 -->
      <div v-else-if="variant === 'duo'" class="fa-duo">
        <span class="fa-glow"></span>
        <span class="fa-word w1">Hello!</span>
        <span class="fa-word w2">Bonjour</span>
        <span class="fa-word w3">¡Hola!</span>
        <span class="fa-word w4">你好</span>
        <div class="fa-icon">
          <span class="j8-brand-icon"><img v-if="icon" :src="icon" alt="" loading="lazy" /><Bird v-else class="fa-fallback-icon" /></span>
        </div>
        <div v-if="artNumber" class="fa-num">
          <b><span class="fa-count" :style="{ '--to': artNumber }"></span><small>{{ artNumberUnit }}</small></b>
          <span>{{ artNumberLabel }}</span>
        </div>
      </div>

      <!-- Muse -->
      <div v-else-if="variant === 'muse'" class="fa-muse">
        <span class="fa-glow"></span>
        <span class="fa-muse-ring">
          <i class="a1"><Sparkles /></i>
          <i class="a2"><Heart /></i>
          <i class="a3"><Star /></i>
          <i class="a4"><Sun /></i>
        </span>
        <div class="fa-icon">
          <span class="j8-brand-icon"><img v-if="icon" :src="icon" alt="" loading="lazy" /><Smile v-else class="fa-fallback-icon" /></span>
        </div>
        <svg class="fa-twinkle t1" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <svg class="fa-twinkle t2" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <span v-if="chips[0]" class="fa-chip c1">{{ chips[0] }}</span>
        <span v-if="chips[1]" class="fa-chip c2">{{ chips[1] }}</span>
        <span v-if="chips[2]" class="fa-chip c3">{{ chips[2] }}</span>
        <div v-if="artNumber" class="fa-num">
          <b><span class="fa-count" :style="{ '--to': artNumber }"></span><small>{{ artNumberUnit }}</small></b>
        </div>
      </div>

      <!-- 工具合集（通栏） -->
      <div v-else-if="variant === 'tools'" class="fa-tools">
        <span class="fa-glow"></span>
        <svg class="fa-twinkle t1" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <svg class="fa-twinkle t2" viewBox="0 0 24 24"><path d="M12 2l2.4 7.6L22 12l-7.6 2.4L12 22l-2.4-7.6L2 12l7.6-2.4Z" /></svg>
        <div class="fa-tools-row">
          <span v-for="(tool, i) in tools" :key="tool.id" class="fa-tool" :class="{ 'is-out': tool.soldOut }" :style="{ animationDelay: `${i * -0.55}s` }">
            <span class="j8-brand-icon tool-icon"><img v-if="tool.icon" :src="tool.icon" alt="" loading="lazy" /><span v-else class="tool-letter">{{ tool.letter }}</span></span>
            <em>{{ tool.price }}</em>
          </span>
        </div>
      </div>

      <!-- 默认 -->
      <div v-else class="fa-default">
        <span class="fa-glow"></span>
        <div class="fa-icon">
          <span class="j8-brand-icon"><img v-if="icon" :src="icon" alt="" loading="lazy" /><Package v-else class="fa-fallback-icon" /></span>
        </div>
      </div>
    </div>

    <div class="j8-fcard-foot">
      <div class="j8-fcard-price">
        <strong>{{ priceParts.symbol }}{{ priceParts.major }}</strong>
        <small>{{ priceParts.minor }} {{ t('qiaozhi.card.priceFrom') }}</small>
      </div>
      <span class="j8-fcard-cta">
        {{ ctaText }}
        <ChevronRight />
      </span>
    </div>
  </RouterLink>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Bird, ChevronRight, Heart, Package, Smile, Sparkles, Star, Sun } from 'lucide-vue-next'
import { useLocalized } from '../../../composables/useProduct'
import { getFirstImageUrl, getImageUrl } from '../../../utils/image'
import { useQiaozhiPrice } from '../utils/price'

const props = defineProps<{
  product: any
  tools?: { id: number; icon: string; letter: string; price: string; soldOut: boolean }[]
}>()

const { t } = useI18n()
const { getLocalizedText } = useLocalized()
const { split: splitPrice } = useQiaozhiPrice()

const slug = computed(() => String(props.product?.slug || '').toLowerCase())
const title = computed(() => getLocalizedText(props.product?.title))
const shortTitle = computed(() => (title.value.split(/[\s｜|·—-]/)[0] || '').slice(0, 12))
const desc = computed(() => {
  const raw = getLocalizedText(props.product?.description)
  const text = raw ? raw.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim() : getLocalizedText(props.product?.category?.name)
  return text.length > 40 ? `${text.slice(0, 40)}…` : text
})
const tag = computed(() => {
  const tags = props.product?.tags
  return Array.isArray(tags) && tags.length ? String(tags[0]).slice(0, 12) : ''
})
const chips = computed(() => (Array.isArray(props.product?.tags) ? props.product.tags.map((x: any) => String(x).slice(0, 10)).slice(0, 3) : []))

const icon = computed(() => {
  const primary = getFirstImageUrl(props.product?.images)
  if (primary) return primary
  const catIcon = props.product?.category?.icon
  return catIcon ? getImageUrl(catIcon) : ''
})

// 副标题（h3 里的渐变 em）：从标题里拆出「3 / 6 个月」这类副题
const em = computed(() => {
  const m = title.value.match(/[（(]([^）)]+)[）)]/)
  if (m) return (m[1] || '').slice(0, 14)
  return ''
})

const priceParts = computed(() => splitPrice(props.product?.price_amount))

// 变体判定：工具行优先，其余按 slug 关键词映射艺术区
const variant = computed(() => {
  if (props.tools && props.tools.length) return 'tools'
  const s = slug.value
  if (/chatgpt|gpt/.test(s)) return 'chatgpt'
  if (/gemini|bard|google/.test(s)) return 'gemini'
  if (/^x-|x-premium|twitter|tweet/.test(s)) return 'x'
  if (/duolingo|duo|lingo/.test(s)) return 'duo'
  if (/muse/.test(s)) return 'muse'
  return 'default'
})

const wideClass = computed(() => (variant.value === 'tools' ? 'is-wide-full' : ''))

// 艺术区数字（Gemini 的 18 个月 / 多邻国的 365 天）
const artNumber = computed(() => {
  const m = title.value.match(/(\d+)\s*(个月|月|天|年|month|months|days)/i)
  if (m) return Number(m[1])
  const any = title.value.match(/(\d{2,4})/)
  return any ? Number(any[1]) : 0
})
const artNumberUnit = computed(() => {
  const m = title.value.match(/\d+\s*(个月|月|天|年|month|months|days)/i)
  return m ? m[1] || '' : ''
})
const artNumberLabel = computed(() => (variant.value === 'duo' ? t('qiaozhi.featured.days') : t('qiaozhi.featured.months')))

// ChatGPT 卡堆三档（用商品自身价格与标题泛化）
const tier1 = computed(() => ({ name: em.value || shortTitle.value, price: `${priceParts.value.symbol}${priceParts.value.major}` }))
const tier2 = computed(() => ({ name: 'Pro', price: `${priceParts.value.symbol}${Math.round(Number(priceParts.value.major || 0) * 4) || priceParts.value.major}` }))
const tier3 = computed(() => ({ name: 'Pro Max', price: `${priceParts.value.symbol}${Math.round(Number(priceParts.value.major || 0) * 8) || priceParts.value.major}` }))

// X 卡时长文案：「3 / 6 个月」→ 3 / 6 MONTHS
const xDuration = computed(() => {
  const m = title.value.match(/(\d+)\s*[/、]\s*(\d+)/)
  if (m) return `${m[1]} / ${m[2]} MONTHS`
  return '3 / 6 MONTHS'
})

const ctaText = computed(() => t('qiaozhi.card.choosePlan'))
</script>

<style scoped>
.fa-fallback-icon {
  width: 46px;
  height: 46px;
  color: var(--accent-strong);
}

.fa-muse-ring :deep(svg) {
  width: 13px;
  height: 13px;
}

.tool-icon {
  width: 72px;
  height: 72px;
  border-radius: 18px;
  background: #fff;
  object-fit: contain;
  box-shadow: rgba(18, 35, 63, 0.13) 0 10px 22px;
}

.tool-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}

.tool-letter {
  font-size: 26px;
  font-weight: 800;
  color: var(--accent-strong);
}
</style>
