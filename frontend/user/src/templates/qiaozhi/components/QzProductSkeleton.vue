<template>
  <div class="qz-grid">
    <div v-for="i in count" :key="i" class="qz-skeleton-card" aria-hidden="true">
      <span class="qz-skeleton-art"><span class="qz-skeleton"></span></span>
      <span class="qz-skeleton qz-skeleton-title"></span>
      <span class="qz-skeleton qz-skeleton-spec"></span>
      <div class="qz-skeleton-bottom">
        <span class="qz-skeleton qz-skeleton-price"></span>
        <span class="qz-skeleton qz-skeleton-action"></span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ count: number }>()
</script>

<style scoped>
/* ---------- 商品网格：与 Home.vue / Products.vue 对齐，保证骨架容器布局独立成立 ---------- */
.qz-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

/* ---------- 加载骨架：商品卡外形 + base .qz-skeleton 占位块（艺术区 / 标题 / 规格 / 价格 / 按钮） ---------- */
.qz-skeleton-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 280px;
  padding: var(--qz-space-5);
  border: 1px solid var(--qz-border);
  border-radius: var(--qz-radius-control);
  background: var(--qz-surface);
}
.qz-skeleton-card .qz-skeleton { border-radius: var(--qz-radius-xs); }
.qz-skeleton-art { display: grid; place-items: center; height: 160px; margin-top: 6px; }
.qz-skeleton-art > .qz-skeleton { width: 54%; height: 92px; opacity: .6; border-radius: var(--qz-radius-sm); }
.qz-skeleton-title { width: 48%; height: 24px; }
.qz-skeleton-spec { width: 62%; height: 12px; margin-top: 2px; }
.qz-skeleton-bottom { display: flex; align-items: center; justify-content: space-between; margin-top: auto; }
.qz-skeleton-price { width: 28%; height: 22px; }
.qz-skeleton-card .qz-skeleton-action { width: 86px; height: 36px; border-radius: var(--qz-radius-pill); }

@media (max-width: 760px) {
  .qz-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--qz-space-2);
  }
  .qz-skeleton-card { min-height: 220px; padding: var(--qz-space-4); }
  .qz-skeleton-art { height: 110px; }
  .qz-skeleton-art > .qz-skeleton { height: 70px; }
}
</style>
