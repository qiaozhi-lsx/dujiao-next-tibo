<template>
  <div class="x-premium-scene" aria-hidden="true">
    <div class="x-card-float">
      <div class="x-member-card">
        <div class="x-scene-wordmark">
          <svg viewBox="0 0 24 24" aria-hidden="true">
            <path
              fill="currentColor"
              d="M18.9 2H22l-6.77 7.74L23.2 22h-6.24l-4.89-7.4L5.6 22H2.47l7.15-8.17L1.2 2h6.4l4.43 6.77L18.9 2ZM17.8 20h1.73L6.65 3.9H4.8L17.8 20Z"
            />
          </svg>
          <span>{{ wordmark }}</span>
        </div>
        <div class="x-card-caption">
          <span>MEMBERSHIP</span>
          <span>{{ durationLabel }} <small>MONTHS</small></span>
        </div>
        <div class="x-card-sheen"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{
    wordmark?: string
    durationLabel?: string
  }>(),
  {
    wordmark: 'Premium',
    durationLabel: '3 / 6',
  },
)
</script>
