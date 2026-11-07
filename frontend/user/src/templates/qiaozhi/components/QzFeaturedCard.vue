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
              <!-- 三个色标的 stop-color 在 <style> 里走 --qz-art-gemini-2 / -1 / -3 -->
              <linearGradient id="qz-gem-grad" x1="0" y1="1" x2="1" y2="0">
                <stop class="g1" /><stop class="g2" offset=".55" /><stop class="g3" offset="1" />
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
import { artVariant, displayTags, type ArtVariant } from '../utils/featured'
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
// 展示用标签：art:<variant> 这类控制标签不上卡
const visibleTags = computed(() => displayTags(props.product?.tags))
const tag = computed(() => (visibleTags.value[0] || '').slice(0, 12))
const chips = computed(() => visibleTags.value.slice(0, 3).map((x) => x.slice(0, 10)))

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

// 变体判定：工具行优先；其余先看 tags 里的 art:<variant>，没有再按 slug 关键词映射（utils/featured.ts）
const variant = computed<ArtVariant>(() => {
  if (props.tools && props.tools.length) return 'tools'
  const tags = props.product?.tags
  return artVariant(slug.value, Array.isArray(tags) ? tags : [])
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
   这里放卡内布局、艺术区容器（.qz-fart）的尺寸、艺术层本身（.qz-fart.is-<variant> 底色与 .fa-* 插画）以及
   hover / 按压联动。品牌色全部走 tokens.css 的 --qz-art-* token（亮暗差异在 token 里切换，本块没有 .dark 规则，
   也没有颜色字面量；字号只用 --qz-text-*）。艺术层的 fa-* 关键帧在本块底部。 */
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

/* ---------- 艺术区容器 ---------- */
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

/* ---------- 艺术层：变体底色（亮暗在 --qz-art-*-wash 里切换） ---------- */
.qz-fart.is-chatgpt { background: radial-gradient(ellipse at 50% 62%, var(--qz-accent-subtle), transparent 68%); }
.qz-fart.is-gemini { background: var(--qz-art-gemini-wash); }
.qz-fart.is-x { background: radial-gradient(at 50% 75%, var(--qz-art-x-wash), transparent 62%); }
.qz-fart.is-duo { background: var(--qz-art-duo-wash); }
.qz-fart.is-muse { background: var(--qz-art-muse-wash); }
.qz-fart.is-tools { background: var(--qz-art-tools-wash); }

/* ---------- 艺术层：通用件（光晕 / 大数字 / 浮动图标 / 星光） ---------- */
.fa-glow {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 230px;
  height: 230px;
  border-radius: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
}
.fa-num {
  position: absolute;
  right: var(--qz-space-4);
  bottom: var(--qz-space-3);
  z-index: 3;
  text-align: right;
}
.fa-num b {
  display: block;
  font-size: var(--qz-text-3xl);
  font-weight: 800;
  line-height: 1;
  letter-spacing: -0.05em;
  font-variant-numeric: tabular-nums;
}
/* 数字由 --to（内联）经 counter 渲染；is-in 时 fa-count 关键帧从 0 起跳 */
.fa-count { --fa-n: var(--to); counter-reset: fa-n var(--fa-n); }
.fa-count::before { content: counter(fa-n); }
.fa-num b small {
  margin-left: 3px;
  font-size: var(--qz-text-sm);
  font-weight: 700;
  letter-spacing: 0;
}
.fa-num > span {
  display: block;
  margin-top: 3px;
  color: var(--qz-text-muted);
  font-size: var(--qz-text-2xs);
}
.fa-icon {
  position: absolute;
  left: 50%;
  top: 50%;
  z-index: 1;
  width: 104px;
  height: 104px;
  transform: translate(-50%, -50%);
  animation: fa-bob 5.2s ease-in-out infinite;
}
.fa-icon .qz-brand-icon,
.fa-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
}
.fa-icon::after {
  content: "";
  position: absolute;
  left: 14%;
  right: 14%;
  bottom: -16px;
  height: 14px;
  border-radius: 50%;
  background: radial-gradient(closest-side, var(--qz-art-icon-shadow), transparent);
}
/* 基础星光取 Gemini 的淡紫；Muse / 工具卡各自覆盖 */
.fa-twinkle {
  position: absolute;
  width: 16px;
  height: 16px;
  fill: var(--qz-art-gemini-twinkle);
  animation: fa-twinkle 3.2s ease-in-out infinite;
}
.fa-twinkle.t1 { left: 18%; top: 30%; }
.fa-twinkle.t2 { right: 22%; top: 18%; width: 10px; height: 10px; animation-delay: -1.1s; }
.fa-twinkle.t3 { left: 30%; bottom: 16%; width: 12px; height: 12px; animation-delay: -2.1s; }

/* ---------- 艺术层：ChatGPT 卡堆 ---------- */
.fa-gpt {
  display: grid;
  place-items: center;
  animation: fa-float 6s ease-in-out infinite;
}
.fa-gpt-card {
  position: absolute;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 196px;
  height: 122px;
  padding: 14px var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  color: var(--qz-art-gpt-ink);
  box-shadow: var(--qz-art-gpt-shadow);
  transition: transform var(--qz-duration-spring) var(--qz-ease-spring);
}
.fa-gpt-card small {
  font-size: var(--qz-text-2xs);
  font-weight: 600;
  opacity: 0.62;
}
.fa-gpt-card b {
  font-size: var(--qz-text-xl);
  font-weight: 700;
  line-height: var(--qz-leading-display);
  letter-spacing: -0.02em;
}
.fa-gpt-card em {
  align-self: flex-end;
  font-size: var(--qz-text-lg);
  font-style: normal;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.fa-gpt-card.c1 { z-index: 3; background: var(--qz-art-gpt-1); transform: translate(-38px, 16px) rotate(-7deg); }
.fa-gpt-card.c2 { z-index: 2; background: var(--qz-art-gpt-2); transform: translate(4px, -4px) rotate(2deg); }
.fa-gpt-card.c3 { z-index: 1; background: var(--qz-art-gpt-3); transform: translate(46px, -22px) rotate(10deg); }

/* ---------- 艺术层：Gemini 星轨 ---------- */
.fa-gem .fa-glow { background: radial-gradient(circle, color-mix(in srgb, var(--qz-art-gemini-1) 30%, transparent), transparent 66%); }
.fa-gem-orbit {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 260px;
  height: 112px;
  border: 1px solid color-mix(in srgb, var(--qz-art-gemini-1) 28%, transparent);
  border-radius: 50%;
  transform: translate(-50%, -50%) rotate(-16deg);
  animation: fa-orbit 16s linear infinite;
}
.fa-gem-orbit i {
  position: absolute;
  left: 50%;
  top: -4px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--qz-art-gemini-1);
  box-shadow: 0 0 10px color-mix(in srgb, var(--qz-art-gemini-1) 80%, transparent);
}
.fa-gem-star {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 112px;
  height: 112px;
  filter: drop-shadow(var(--qz-art-gemini-shadow));
  transform: translate(-50%, -50%);
  animation: fa-gem-float 6s ease-in-out infinite;
}
/* 星形渐变的三个色标（模板里的 <stop class="g1/g2/g3">）：蓝 → 紫 → 粉 */
.fa-gem-star .g1 { stop-color: var(--qz-art-gemini-2); }
.fa-gem-star .g2 { stop-color: var(--qz-art-gemini-1); }
.fa-gem-star .g3 { stop-color: var(--qz-art-gemini-3); }
.fa-gem .fa-num b { color: var(--qz-art-gemini-num); }

/* ---------- 艺术层：X Premium 3D 卡 ---------- */
.fa-x {
  display: grid;
  place-items: center;
}
.fa-x-card {
  position: relative;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  width: 238px;
  height: 146px;
  padding: 18px var(--qz-space-5);
  overflow: hidden;
  border-radius: var(--qz-radius-control);
  color: var(--qz-art-white);
  background: linear-gradient(150deg, var(--qz-art-x-1), var(--qz-art-x-2) 58%, var(--qz-art-x-3));
  box-shadow: var(--qz-art-x-shadow);
  transform: rotate(-8deg);
  animation: fa-x-float 5.4s ease-in-out infinite;
}
.fa-x-logo {
  display: flex;
  align-items: center;
  gap: 10px;
}
.fa-x-logo svg { width: 25px; height: 25px; }
.fa-x-logo b {
  font-size: var(--qz-text-xl);
  font-weight: 700;
  letter-spacing: -0.01em;
}
.fa-x-meta {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: var(--qz-space-2);
}
.fa-x-meta small {
  color: color-mix(in srgb, var(--qz-art-white) 50%, transparent);
  font-size: var(--qz-text-2xs);
  letter-spacing: 0.24em;
}
.fa-x-meta em {
  color: color-mix(in srgb, var(--qz-art-white) 85%, transparent);
  font-size: var(--qz-text-2xs);
  font-style: normal;
  letter-spacing: 0.1em;
  white-space: nowrap;
}
.fa-x-sheen {
  position: absolute;
  inset: -40% -30%;
  background: linear-gradient(105deg, transparent 42%, color-mix(in srgb, var(--qz-art-white) 20%, transparent) 50%, transparent 58%);
  animation: fa-sheen 5s ease-in-out infinite;
  pointer-events: none;
}

/* ---------- 艺术层：多邻国 ---------- */
.fa-duo .fa-glow { background: radial-gradient(circle, color-mix(in srgb, var(--qz-art-duo-1) 22%, transparent), transparent 66%); }
.fa-word {
  position: absolute;
  z-index: 2;
  padding: 6px var(--qz-space-3);
  border-radius: var(--qz-radius-sm);
  background: var(--qz-surface);
  color: var(--qz-art-duo-ink);
  font-size: var(--qz-text-sm);
  font-weight: 800;
  white-space: nowrap;
  box-shadow: var(--qz-art-duo-shadow);
  animation: fa-word 4.6s ease-in-out infinite;
}
.fa-word.w1 { left: 9%; top: 18%; rotate: -8deg; }
.fa-word.w2 { right: 9%; top: 24%; rotate: 7deg; animation-delay: -1.2s; }
.fa-word.w3 { left: 14%; bottom: 16%; rotate: 5deg; font-size: var(--qz-text-xs); animation-delay: -2.4s; }
.fa-word.w4 { left: 40%; top: 5%; rotate: -3deg; font-size: var(--qz-text-2xs); animation-delay: -3.3s; }
.fa-duo .fa-num b { color: var(--qz-art-duo-num); }

/* ---------- 艺术层：Muse ---------- */
.fa-muse .fa-glow {
  background: radial-gradient(
    circle,
    color-mix(in srgb, var(--qz-art-muse-3) 26%, transparent),
    color-mix(in srgb, var(--qz-art-muse-2) 8%, transparent) 45%,
    transparent 68%
  );
}
.fa-muse-ring {
  position: absolute;
  left: 50%;
  top: 50%;
  width: 164px;
  height: 164px;
  border: 1.5px dashed color-mix(in srgb, var(--qz-art-muse-1) 35%, transparent);
  border-radius: 50%;
  transform: translate(-50%, -50%);
  animation: fa-spin 18s linear infinite;
}
/* 徽章描边取表面色：亮色下就是白，暗色下随表面变深（原 .dark 覆盖） */
.fa-muse-ring i {
  position: absolute;
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border: 2px solid var(--qz-surface);
  border-radius: 50%;
  color: var(--qz-art-white);
  box-shadow: var(--qz-art-muse-badge-shadow);
  animation: fa-unspin 18s linear infinite;
}
.fa-muse-ring .a1 { top: -13px; left: calc(50% - 13px); background: var(--qz-art-muse-badge-1); }
.fa-muse-ring .a2 { right: -13px; top: calc(50% - 13px); background: var(--qz-art-muse-badge-2); }
.fa-muse-ring .a3 { bottom: -13px; left: calc(50% - 13px); background: var(--qz-art-muse-badge-3); }
.fa-muse-ring .a4 { left: -13px; top: calc(50% - 13px); background: var(--qz-art-muse-badge-4); }
.fa-muse .fa-twinkle { fill: var(--qz-art-muse-twinkle); }
.fa-muse .fa-twinkle.t1 { left: 14%; top: 16%; }
.fa-muse .fa-twinkle.t2 { left: 70%; top: 12%; right: auto; }
.fa-chip {
  position: absolute;
  bottom: 30%;
  z-index: 2;
  padding: 5px 10px;
  border-radius: var(--qz-radius-pill);
  opacity: 0;
  background: var(--qz-art-muse-chip);
  color: var(--qz-art-muse-chip-ink);
  font-size: var(--qz-text-xs);
  font-weight: 800;
  white-space: nowrap;
  box-shadow: var(--qz-art-muse-chip-shadow);
  animation: fa-rise 4.8s ease-out infinite;
}
.fa-chip.c1 { left: 8%; }
.fa-chip.c2 { left: 60%; bottom: 46%; animation-delay: -1.6s; }
.fa-chip.c3 { left: 28%; animation-delay: -3.2s; }
.fa-muse .fa-num b {
  background-image: var(--qz-art-muse-num);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

/* ---------- 艺术层：工具合集（通栏） ---------- */
.fa-tools .fa-glow {
  width: 320px;
  background: radial-gradient(color-mix(in srgb, var(--qz-art-tools-1) 16%, transparent), transparent 66%);
}
.fa-tools .fa-twinkle { fill: var(--qz-art-tools-twinkle); }
.fa-tools .fa-twinkle.t1 { left: 5%; top: 16%; }
.fa-tools .fa-twinkle.t2 { right: 6%; top: auto; bottom: 18%; animation-delay: -1.6s; }
.fa-tools-row {
  position: absolute;
  inset: 0;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: clamp(14px, 4vw, 46px);
}
.fa-tool {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--qz-space-2);
  animation: fa-float 3.6s ease-in-out infinite;
}
.fa-tool em {
  padding: 2px 9px;
  border-radius: var(--qz-radius-pill);
  background: var(--qz-surface);
  color: var(--qz-text);
  font-size: var(--qz-text-xs);
  font-style: normal;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  box-shadow: var(--qz-art-tools-shadow);
}
.fa-tool.is-out {
  filter: grayscale(1);
  opacity: 0.5;
}

/* ---------- 底部：起价 + CTA ----------
   价格与「起」后缀钉在一行；一行放不下（en-US @1024 的「Choose a plan」）时整颗 CTA 换到下一行靠右，而不是把后缀挤到价格下面 */
.qz-fcard-foot {
  display: flex;
  flex-wrap: wrap;
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
  white-space: nowrap;
}
.qz-fcard-cta { flex: none; margin-left: auto; }
.qz-fcard-cta svg { transition: transform var(--qz-duration-standard) var(--qz-ease-standard); }
.qz-fcard:hover .qz-fcard-cta svg { transform: translateX(3px); }

/* ---------- 艺术区 hover 联动 ---------- */
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
/* 图标底始终是白瓷砖（品牌 logo 多为透明底，暗色下也要可辨），hover 位移走 transition */
.tool-icon img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: var(--qz-art-white);
  transition: transform var(--qz-duration-standard) var(--qz-ease-standard);
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

  /* 艺术层：大数字降一档，工具行收紧 */
  .fa-num b { font-size: var(--qz-text-2xl); }
  .fa-num b small { font-size: var(--qz-text-xs); }
  .fa-tools-row { gap: 10px; padding: 0 var(--qz-space-2); }
}

/* ---------- 艺术层关键帧（只有本组件用；scoped 编译会给名字加作用域后缀，与上面的 animation 一并改写） ----------
   浮动 / 轨道 / 闪烁 / 扫光 / 字词起伏 / 芯片上升 / 数字计数。 */
@keyframes fa-float { 0%, 100% { transform: translateY(4px); } 50% { transform: translateY(-5px); } }
@keyframes fa-bob { 0%, 100% { transform: translate(-50%, -50%); } 50% { transform: translate(-50%, -57%); } }
@keyframes fa-orbit { 0% { transform: translate(-50%, -50%) rotate(-16deg); } 100% { transform: translate(-50%, -50%) rotate(344deg); } }
@keyframes fa-unspin { 100% { transform: rotate(-360deg); } }
@keyframes fa-count { 0% { --fa-n: 0; } }
@keyframes fa-spin { 0% { transform: translate(-50%, -50%) rotate(0deg); } 100% { transform: translate(-50%, -50%) rotate(360deg); } }
@keyframes fa-gem-float { 0%, 100% { transform: translate(-50%, -50%) rotate(-6deg) scale(0.97); } 50% { transform: translate(-50%, -57%) rotate(6deg) scale(1.04); } }
@keyframes fa-twinkle { 0%, 100% { opacity: 0.3; transform: scale(0.7); } 50% { opacity: 1; transform: scale(1.1); } }
@keyframes fa-x-float { 0%, 100% { transform: rotate(-8deg) translateY(4px); } 50% { transform: rotate(-5deg) translateY(-6px); } }
@keyframes fa-sheen { 0%, 22% { opacity: 0; transform: translate(-60%); } 28% { opacity: 1; } 56% { opacity: 1; transform: translate(60%); } 62%, 100% { opacity: 0; transform: translate(60%); } }
@keyframes fa-word { 0%, 100% { translate: 0; } 50% { translate: 0 -8px; } }
@keyframes fa-rise { 0% { opacity: 0; transform: translateY(24px) scale(0.9); } 15%, 65% { opacity: 1; } 100% { opacity: 0; transform: translateY(-112px) scale(1); } }
</style>
