<template>
  <!-- 资料面板没有专属样式：区块 / 表单 / 动作组来自 PersonalCenter.vue 的 .qz-me-* 共享块，
       .qz-panel / .qz-section-head / .qz-field 三件套 / .qz-alert / QzButton 来自 styles/base.css。 -->
  <div class="qz-me-panel">
    <section class="qz-panel qz-me-sec">
      <div class="qz-section-head">
        <div>
          <h2>{{ t('personalCenter.profile.title') }}</h2>
          <small>{{ t('personalCenter.profile.subtitle') }}</small>
        </div>
        <div class="qz-me-acts">
          <span class="qz-chip">{{ t('personalCenter.tabs.profile') }}</span>
        </div>
      </div>

      <p
        v-if="profileAlert"
        class="qz-alert"
        :class="alertClass(profileAlert.level)"
        :role="profileAlert.level === 'error' ? 'alert' : 'status'"
      >
        {{ profileAlert.message }}
      </p>

      <form class="qz-me-form" novalidate @submit.prevent="handleSaveProfile">
        <div class="qz-field">
          <label class="qz-field-label" for="qz-me-profile-email">{{ t('personalCenter.profile.emailLabel') }}</label>
          <input id="qz-me-profile-email" class="qz-field-control" :value="userProfileStore.profile?.email || ''" type="email" disabled />
        </div>

        <div class="qz-me-form-row">
          <div class="qz-field">
            <label class="qz-field-label" for="qz-me-profile-nickname">{{ t('personalCenter.profile.nicknameLabel') }}</label>
            <input
              id="qz-me-profile-nickname"
              v-model="profileForm.nickname"
              class="qz-field-control"
              type="text"
              autocomplete="nickname"
              :placeholder="t('personalCenter.profile.nicknamePlaceholder')"
            />
          </div>

          <div class="qz-field">
            <label class="qz-field-label" for="qz-me-profile-locale">{{ t('personalCenter.profile.localeLabel') }}</label>
            <QzSelect
              id="qz-me-profile-locale"
              :model-value="profileForm.locale"
              :items="[{ value: 'zh-CN', label: '简体中文' }, { value: 'zh-TW', label: '繁體中文' }, { value: 'en-US', label: 'English' }]"
              @update:model-value="profileForm.locale = $event as string"
            />
          </div>
        </div>

        <div class="qz-me-acts qz-me-form-acts">
          <QzButton type="submit" :loading="userProfileStore.savingProfile">
            {{ userProfileStore.savingProfile ? t('personalCenter.profile.saving') : t('personalCenter.profile.save') }}
          </QzButton>
        </div>
      </form>
    </section>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { PageAlert, PageAlertLevel } from '../../../utils/alerts'
import { useUserProfileStore } from '../../../stores/userProfile'
import QzButton from '../components/QzButton.vue'
import QzSelect from '../components/QzSelect.vue'

const { t } = useI18n()
const userProfileStore = useUserProfileStore()

const profileForm = reactive({
  nickname: '',
  locale: 'zh-CN',
})

const profileAlert = ref<PageAlert | null>(null)

/** 提示条 tone（base .qz-alert）：success → is-success、warning → is-warning、error → is-danger */
const alertClass = (level: PageAlertLevel) => {
  if (level === 'error') return 'is-danger'
  if (level === 'success') return 'is-success'
  return 'is-warning'
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
