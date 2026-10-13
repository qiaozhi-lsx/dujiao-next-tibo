<template>
  <div class="j8-me-panel">
    <div class="j8-me-card">
      <div class="j8-me-card-head">
        <div>
          <h2>{{ t('personalCenter.profile.title') }}</h2>
          <small>{{ t('personalCenter.profile.subtitle') }}</small>
        </div>
        <div class="j8-me-card-acts">
          <span class="j8-me-pill is-ok">{{ t('personalCenter.tabs.profile') }}</span>
        </div>
      </div>

      <p
        v-if="profileAlert"
        class="j8-ic-alert"
        :class="alertClass(profileAlert.level)"
        :role="profileAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ profileAlert.message }}
      </p>

      <form class="j8-ic" novalidate @submit.prevent="handleSaveProfile">
        <div class="j8-ic-field">
          <label for="qz-me-profile-email">{{ t('personalCenter.profile.emailLabel') }}</label>
          <input id="qz-me-profile-email" :value="userProfileStore.profile?.email || ''" type="email" disabled />
        </div>

        <div class="j8-ic-row">
          <div class="j8-ic-field">
            <label for="qz-me-profile-nickname">{{ t('personalCenter.profile.nicknameLabel') }}</label>
            <input
              id="qz-me-profile-nickname"
              v-model="profileForm.nickname"
              type="text"
              autocomplete="nickname"
              :placeholder="t('personalCenter.profile.nicknamePlaceholder')"
            />
          </div>

          <div class="j8-ic-field">
            <label for="qz-me-profile-locale">{{ t('personalCenter.profile.localeLabel') }}</label>
            <J8Select
              id="qz-me-profile-locale"
              :model-value="profileForm.locale"
              :items="[{ value: 'zh-CN', label: '简体中文' }, { value: 'zh-TW', label: '繁體中文' }, { value: 'en-US', label: 'English' }]"
              @update:model-value="profileForm.locale = $event as string"
            />
          </div>
        </div>

        <div class="j8-me-card-acts">
          <J8Button type="submit" :loading="userProfileStore.savingProfile">
            {{ userProfileStore.savingProfile ? t('personalCenter.profile.saving') : t('personalCenter.profile.save') }}
          </J8Button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PageAlert, PageAlertLevel } from '../../../utils/alerts'
import { useUserProfileStore } from '../../../stores/userProfile'
import J8Button from '../components/J8Button.vue'
import J8Select from '../components/J8Select.vue'

const { t } = useI18n()
const userProfileStore = useUserProfileStore()

const profileForm = reactive({
  nickname: '',
  locale: 'zh-CN',
})

const profileAlert = ref<PageAlert | null>(null)

/** 警示条 tone：success → 品牌色底、warning → 琥珀底、error → 红底 */
const alertClass = (level: PageAlertLevel) => {
  if (level === 'error') return 'is-error'
  if (level === 'success') return 'j8-ok-alert'
  return ''
}

const handleSaveProfile = async () => {
  profileAlert.value = null
  const payload = {
    nickname: profileForm.nickname.trim(),
    locale: profileForm.locale,
  }
  const ok = await userProfileStore.saveProfile(payload)
  if (!ok) {
    profileAlert.value = {
      level: 'error',
      message: userProfileStore.profileError || t('personalCenter.common.saveFailed'),
    }
    return
  }
  profileAlert.value = {
    level: 'success',
    message: t('personalCenter.profile.saveSuccess'),
  }
}

watch(
  () => userProfileStore.profile,
  (profile) => {
    if (!profile) return
    profileForm.nickname = profile.nickname || ''
    profileForm.locale = profile.locale || 'zh-CN'
  },
  { immediate: true }
)
</script>

<style scoped>
/* 成功态警示条（复用 j8-ic-alert 底色体系，绿/品牌色 tone 与登录页一致） */
.j8-ic-alert.j8-ok-alert {
  background: var(--accent-soft);
  color: var(--accent-strong);
}

/* 只读邮箱：全局输入样式未覆盖 :disabled，这里补齐可辨识的置灰态 */
.j8-ic-field input:not([type="radio"]):not([type="checkbox"]):disabled {
  background: var(--surface-2);
  color: var(--muted);
  cursor: not-allowed;
}
</style>
