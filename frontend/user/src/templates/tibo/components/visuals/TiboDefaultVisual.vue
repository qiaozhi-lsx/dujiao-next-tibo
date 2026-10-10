<template>
  <div class="store-mail-art" aria-hidden="true">
    <div class="store-envelope"></div>
    <span class="store-envelope-label">
      <Package class="mail-icon" />
      {{ label }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { Package } from 'lucide-vue-next'

defineProps<{
  label?: string
}>()
</script>

<style scoped>
.mail-icon {
  width: 16px;
  height: 16px;
}
</style>
