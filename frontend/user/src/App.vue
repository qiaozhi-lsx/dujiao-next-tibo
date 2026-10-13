<template>
  <div id="app" class="min-h-screen bg-background text-foreground flex flex-col">
    <!-- qiaozhi 模板：自带顶栏/页脚的外壳包裹页面（控制台仍走下方分支） -->
    <QiaozhiLayout v-if="isQiaozhi && !isResellerConsole">
      <ErrorBoundary>
        <RouterView v-slot="{ Component }">
          <Transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </ErrorBoundary>
    </QiaozhiLayout>

    <!-- vault 模板：自带顶栏/页脚的外壳包裹页面（控制台仍走下方分支） -->
    <VaultLayout v-else-if="isVault && !isResellerConsole">
      <ErrorBoundary>
        <RouterView v-slot="{ Component }">
          <Transition name="page-fade" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </ErrorBoundary>
    </VaultLayout>

    <!-- classic 模板 / 分销控制台（保持原有结构不变） -->
    <template v-else>
      <Navbar v-if="!isResellerConsole" />
      <main class="flex-1" :class="isResellerConsole ? '' : 'pb-14 lg:pb-0'">
        <ErrorBoundary>
          <RouterView v-slot="{ Component }">
            <Transition name="page-fade" mode="out-in">
              <component :is="Component" />
            </Transition>
          </RouterView>
        </ErrorBoundary>
      </main>
      <Footer v-if="!isResellerConsole" />
      <BackToTop v-if="!isResellerConsole" />
      <MobileBottomNav v-if="!isResellerConsole" />
    </template>

    <Loading :loading="appStore.loading" />
    <Toast />
    <ConfirmDialog />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from './stores/app'
import { getActiveTemplate } from './templates/registry'
import Navbar from './components/Navbar.vue'
import Footer from './components/Footer.vue'
import Loading from './components/Loading.vue'
import Toast from './components/Toast.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import ErrorBoundary from './components/ErrorBoundary.vue'
import BackToTop from './components/BackToTop.vue'
import MobileBottomNav from './components/MobileBottomNav.vue'
// 两套外壳都 eager 打进主 bundle：虽然每个 user 只用一套，但省掉一次 chunk RTT
// （整站只有 2 套 Layout，总 bundle 增量 < 30KB gzip，比首屏白屏期 500-1000ms 划算）
import QiaozhiLayout from './templates/qiaozhi/layout/QiaozhiLayout.vue'
import VaultLayout from './templates/vault/layout/VaultLayout.vue'

// config 由 router.beforeEach 统一加载，无需在此重复调用
const appStore = useAppStore()
const route = useRoute()
const isResellerConsole = computed(() => route.meta.resellerConsole === true)
// getActiveTemplate 读取 appStore.config（响应式），config 加载后会重新计算
const isQiaozhi = computed(() => getActiveTemplate() === 'qiaozhi')
const isVault = computed(() => getActiveTemplate() === 'vault')
</script>

<style>
.page-fade-enter-active,
.page-fade-leave-active {
  transition: opacity 200ms ease;
}

.page-fade-enter-from,
.page-fade-leave-to {
  opacity: 0;
}
</style>
