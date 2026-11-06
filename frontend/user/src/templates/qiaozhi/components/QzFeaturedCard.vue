<template>
  <RouterLink
    :to="`/products/${product.slug}`"
    class="qz-panel qz-fcard qz-reveal is-pending"
    :class="[`is-${variant}`, wideClass]"
    :data-reveal="1"
  >
    <div class="qz-fcard-head">
      <div>
        <h3>{{ title }}<em v-if="em">{{ em }}</em></h3>
        <p>{{ desc }}</p>
      </div>
      <span v-if="tag" class="qz-chip qz-fcard-tag">{{ tag }}</span>
    </div>

    <!-- ========== 艺术区 ========== -->
    <div class="qz-fart" :class="`is-${variant}`" aria-hidden="true">
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
              <linearGradient id="qz-gem-grad" x1="0" y1="1" x2="1" y2="0">
                <stop stop-color="#4f8bff" /><stop offset=".55" stop-color="#8d6bff" /><stop offset="1" stop-color="#d65fb4" />
              </linearGradient>
            </defs>
            <path fill="url(#qz-gem-grad)" d="M80 5C91 48 112 69 155 80C112 91 91 112 80 155C69 112 48 91 5 80C48 69 69 48 80 5Z" />
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
            <small>{{ t('qiaozhi.featured.membershipEyebrow') }}</small>
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
          <span class="qz-brand-icon"><img v-if="icon" :src="icon" alt="" loading="lazy" /><Bird v-else class="fa-fallback-icon" /></span>
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
          <span class="qz-brand-icon"><img v-if="icon" :src="icon" alt="" loading="lazy" /><Smile v-else class="fa-fallback-icon" /></span>
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
            <span class="qz-brand-icon tool-icon"><img v-if="tool.icon" :src="tool.icon" alt="" loading="lazy" /><span v-else class="tool-letter">{{ tool.letter }}</span></span>
            <em>{{ tool.price }}</em>
          </span>
        </div>
      </div>

      <!-- 默认 -->
      <div v-else class="fa-default">
        <span class="fa-glow"></span>
        <div class="fa-icon">
          <span class="qz-brand-icon"><img v-if="icon" :src="icon" alt="" loading="lazy" /><Package v-else class="fa-fallback-icon" /></span>
        </div>
      </div>
    </div>

    <div class="qz-fcard-foot">
      <div class="qz-fcard-price">
        <strong>{{ priceParts.symbol }}{{ priceParts.major }}</strong>
        <small>{{ priceParts.minor }} {{ t('qiaozhi.card.priceFrom') }}</small>
      </div>
      <span class="qz-btn qz-btn--primary qz-btn--sm qz-fcard-cta">
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
import { startingPrice, useQiaozhiPrice } from '../utils/price'

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

// 「起」价 = 启用 SKU 的最低价；无 SKU 回退商品价（C10：以前直接读 price_amount，多 SKU 时显示的不是最低档）
const priceParts = computed(() => splitPrice(startingPrice(props.product)))

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
/* 精选卡：表面 / 边框 / 圆角 / 内边距来自 .qz-panel，标签来自 .qz-chip，CTA 来自 .qz-btn--primary.qz-btn--sm。
   这里放卡内布局、艺术区容器（.qz-fart）的尺寸与 hover / 按压联动。
   艺术层本身（.fa-* 与 .qz-fart.is-<variant> 的品牌渐变）仍在 legacy.css：品牌色要等 --qz-brand-* token
   （Task 10），本块不含颜色字面量。关键帧 fa-* 同样来自 legacy.css。 */
.qz-fcard {
  display: flex;
  flex-direction: column;
  min-height: 300px;
  overflow: hidden;
  text-decoration: none;
  cursor: pointer;
  transition:
    transform var(--qz-duration-considered) var(--qz-ease-standard),
    border-color var(--qz-duration-standard) var(--qz-ease-standard),
    box-shadow var(--qz-duration-standard) var(--qz-ease-standard);
}
@media (hover: hover) and (pointer: fine) {
  .qz-fcard:hover {
    transform: translateY(-5px);
    border-color: var(--qz-border-strong);
    box-shadow: var(--qz-shadow-raised);
  }
}
.qz-fcard:active {
  transform: translateY(-1px) scale(0.985);
  transition-duration: var(--qz-duration-instant);
}

/* ---------- 头部：标题 / 说明 / 标签 ---------- */
.qz-fcard-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--qz-space-3);
}
.qz-fcard-head > div { min-width: 0; }
.qz-fcard h3 {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0 0.32em;
  margin: 0;
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: var(--qz-tracking-display);
}
/* 标题里拆出的副题（如「3 / 6 个月」） */
.qz-fcard h3 em {
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-base);
  font-style: normal;
  font-weight: 600;
}
.qz-fcard-head p {
  margin: var(--qz-space-2) 0 0;
  color: var(--qz-text-secondary);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}
.qz-fcard-tag { flex: none; margin-top: var(--qz-space-1); }

/* ---------- 艺术区容器（内容由 legacy.css 的 .fa-* 绘制） ---------- */
.qz-fart {
  position: relative;
  flex: 1 1 0%;
  min-height: 170px;
  margin: var(--qz-space-3) calc(-1 * var(--qz-space-2));
  overflow: hidden;
  border-radius: var(--qz-radius-control);
  transition: transform var(--qz-duration-considered) var(--qz-ease-standard);
}
.qz-fart > div { position: absolute; inset: 0; }
.qz-fcard:hover .qz-fart,
.qz-fcard:focus-visible .qz-fart { transform: scale(1.04); }
.qz-fcard:active .qz-fart {
  transform: scale(0.985);
  transition-duration: var(--qz-duration-instant);
}
@media (hover: none) {
  .qz-fcard:active .qz-fart { transform: scale(1.02); }
}

/* ---------- 底部：起价 + CTA ---------- */
.qz-fcard-foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--qz-space-3);
  margin-top: auto;
}
.qz-fcard-price {
  display: flex;
  align-items: baseline;
  gap: var(--qz-space-1);
  min-width: 0;
}
.qz-fcard-price strong {
  color: var(--qz-text);
  font-size: var(--qz-text-xl);
  font-weight: 700;
  letter-spacing: var(--qz-tracking-display);
  font-variant-numeric: tabular-nums;
}
.qz-fcard-price small {
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
}
.qz-fcard-cta { flex: none; }
.qz-fcard-cta svg { transition: transform var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-fcard:hover .qz-fcard-cta svg { transform: translateX(3px); }

/* ---------- 艺术区 hover 联动（元素由 legacy.css 的 .fa-* 定位） ---------- */
.qz-fcard:hover .fa-gpt-card.c1 { transform: translate(-70px, 18px) rotate(-11deg); }
.qz-fcard:hover .fa-gpt-card.c2 { transform: translateY(-10px) rotate(0deg); }
.qz-fcard:hover .fa-gpt-card.c3 { transform: translate(72px, -10px) rotate(12deg); }
.qz-fcard.is-in .fa-count { animation: fa-count 1.6s var(--qz-ease-enter) 0.35s 1 backwards; }
.qz-fcard:hover .fa-tool img { transform: translateY(-4px) scale(1.06); }
/* 悬停时暂停浮动 / 摆动，松手恢复 */
.qz-fcard.is-x:hover .fa-x-card,
.qz-fcard.is-gemini:hover .fa-gem-star,
.qz-fcard.is-duo:hover .fa-icon,
.qz-fcard.is-muse:hover .fa-icon { animation-play-state: paused; }
.qz-fcard:hover .fa-x-sheen { animation-duration: 2.4s; }
.qz-fcard:hover .fa-muse-ring,
.qz-fcard:hover .fa-gem-orbit { animation-duration: 7s; }
.qz-fcard:hover .fa-chip { animation-duration: 3s; }

/* ---------- 工具合集通栏卡 ---------- */
.qz-fcard.is-wide-full {
  flex-direction: row;
  align-items: center;
  gap: var(--qz-space-6);
  min-height: 0;
}
.qz-fcard.is-wide-full .qz-fcard-head {
  flex: 0 0 270px;
  flex-direction: column;
  gap: var(--qz-space-2);
}
.qz-fcard.is-wide-full .qz-fcard-tag { margin-top: 0; }
.qz-fcard.is-wide-full .qz-fart {
  flex: 1 1 0%;
  align-self: stretch;
  min-height: 150px;
  margin: 0;
}
.qz-fcard.is-wide-full .qz-fcard-foot {
  flex: none;
  flex-direction: column;
  align-items: flex-end;
  gap: var(--qz-space-3);
  margin-top: 0;
}

/* 组件内自绘的小件 */
.fa-fallback-icon {
  width: 46px;
  height: 46px;
  color: var(--qz-accent-strong);
}
.fa-muse-ring :deep(svg) {
  width: 13px;
  height: 13px;
}
.tool-icon {
  width: 72px;
  height: 72px;
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
  object-fit: contain;
  box-shadow: var(--qz-shadow-raised);
}
.tool-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.tool-letter {
  color: var(--qz-accent-strong);
  font-size: var(--qz-text-2xl);
  font-weight: 800;
}

@media (max-width: 1023.98px) {
  .qz-fcard.is-wide-full .qz-fcard-head { flex-basis: 200px; }
  .tool-icon { width: 60px; height: 60px; }
}

@media (max-width: 760px) {
  /* 带上基类名，靠特指度覆盖 .qz-panel 的 padding */
  .qz-panel.qz-fcard { padding: var(--qz-space-4); }
  .qz-fcard:hover { transform: none; }
  .qz-fart {
    min-height: 160px;
    margin: var(--qz-space-3) calc(-1 * var(--qz-space-1));
  }
  .qz-fart > div { scale: 0.86; }
  .qz-fcard.is-wide-full {
    flex-direction: column;
    align-items: stretch;
    gap: 0;
    min-height: 300px;
  }
  .qz-fcard.is-wide-full .qz-fcard-head { flex: none; flex-direction: row; }
  .qz-fcard.is-wide-full .qz-fcard-tag { margin-top: var(--qz-space-1); }
  .qz-fcard.is-wide-full .qz-fart {
    flex: 1 1 0%;
    min-height: 160px;
    margin: var(--qz-space-3) calc(-1 * var(--qz-space-1));
  }
  .qz-fcard.is-wide-full .qz-fcard-foot {
    flex-direction: row;
    align-items: center;
    margin-top: auto;
  }
  .tool-icon { width: 52px; height: 52px; }
}
</style>
