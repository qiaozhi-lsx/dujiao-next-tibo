<template>
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
        class="qz-me-alert"
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

/** 警示条 tone：success → 成功色、warning → 警示色（默认）、error → 错误色 */
const alertClass = (level: PageAlertLevel) => {
  if (level === 'error') return 'is-error'
  if (level === 'success') return 'is-success'
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
/* 资料面板：区块 .qz-panel + .qz-section-head、表单 .qz-field 三件套、提交 QzButton（styles/base.css）；只读邮箱走 .qz-field-control:disabled。 */
.qz-me-panel {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

.qz-panel.qz-me-sec {
  display: flex;
  flex-direction: column;
  gap: var(--qz-space-4);
}

.qz-me-sec > .qz-section-head {
  align-items: center;
  margin-bottom: 0;
}

.qz-section-head small {
  display: block;
  margin-top: var(--qz-space-1);
  color: var(--qz-text-muted);
  font-size: var(--qz-text-xs);
  line-height: var(--qz-leading-body);
}

.qz-me-acts {
  display: flex;
  flex: none;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: var(--qz-space-2);
}

/* ---------- 提示条：默认警示色，is-error 错误色，is-success 成功色 ---------- */
.qz-me-alert {
  margin: 0;
  padding: var(--qz-space-3) var(--qz-space-4);
  border-radius: var(--qz-radius-control);
  background: color-mix(in oklab, var(--qz-warning) 12%, transparent);
  color: var(--qz-warning);
  font-size: var(--qz-text-sm);
  font-weight: 500;
  line-height: var(--qz-leading-body);
  overflow-wrap: anywhere;
}

.qz-me-alert.is-error {
  background: color-mix(in oklab, var(--qz-danger) 12%, transparent);
  color: var(--qz-danger);
}

.qz-me-alert.is-success {
  background: color-mix(in oklab, var(--qz-success) 12%, transparent);
  color: var(--qz-success);
}

/* ---------- 表单 ---------- */
.qz-me-form {
  display: grid;
  gap: var(--qz-space-4);
}

.qz-me-form-row {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--qz-space-4);
}

.qz-me-form-acts {
  justify-content: flex-start;
}

@media (max-width: 760px) {
  .qz-panel.qz-me-sec {
    padding: var(--qz-space-4);
  }

  .qz-me-form-row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
