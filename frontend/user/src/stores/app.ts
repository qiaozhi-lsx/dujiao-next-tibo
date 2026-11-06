import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import { configAPI } from '../api'
import { applyCustomScripts } from '../utils/customScripts'
import { getImageUrl } from '../utils/image'
import { getLocalizedText } from '../utils/resellerSiteConfig'
import { estimateServerNow } from '../utils/serverTime'
import { detectLocale, setI18nLocale } from '../i18n'
import { useHead } from '@unhead/vue'

export const useAppStore = defineStore('app', () => {
    // 与 vue-i18n 复用同一套语言检测逻辑，避免首次访问时
    // UI 语言（vue-i18n）与商品多语言取值（appStore.locale）不一致
    const locale = ref(detectLocale())
    const config = ref<any>(null)
    // 完整配置是否已就位。服务端内联的 window.__PUBLIC_CONFIG__ 只是 /public/config 的子集
    // （brand / seo / 模板等首屏字段），nav_config、email_verification_enabled 这类可选开关键
    // 不在其中：消费内联配置后为 false，任一次 configAPI.get() 成功赋值后为 true。
    // 可选开关一律用 utils/configFlags.ts 的 optionalFlagOn(configComplete, value) 读取，
    // 避免 `value !== false` 在内联阶段把缺失的键当成"开启"而闪一下。
    const configComplete = ref(false)
    const loading = ref(false)
    // 服务器与客户端的时间偏移量（毫秒），serverTime = clientTime + offset
    const serverTimeOffset = ref(0)
    const siteIconHref = computed(() => {
        const siteIcon = String(config.value?.brand?.site_icon || '').trim()
        return siteIcon ? getImageUrl(siteIcon) : '/dj.svg'
    })
    const isResellerTenant = computed(() => {
        return String(config.value?.tenant?.mode || '').trim().toLowerCase() === 'reseller'
    })
    const canAccessResellerConsole = computed(() => !!config.value && !isResellerTenant.value)

    // 设置语言：同时驱动 vue-i18n（内部按需加载语言包）
    const setLocale = (newLocale: string) => {
        locale.value = newLocale
        localStorage.setItem('locale', newLocale)
        void setI18nLocale(newLocale)
    }

    // 全局 head 默认值（标题/描述/keywords/favicon/html lang）。
    // canonical / og / twitter 等页面级别字段交由各页面通过 usePageSeo 接管，
    // 避免与页面级 useHead 冲突或产生重复标签。
    useHead({
        htmlAttrs: { lang: computed(() => locale.value) },
        title: () => {
            const seo = config.value?.seo
            const lang = locale.value
            const localized = getLocalizedText(seo?.title, lang)
            if (localized) return String(localized).trim() || undefined
            const siteName = String(config.value?.brand?.site_name || '').trim()
            return siteName || undefined
        },
        link: () => [{ key: 'favicon', rel: 'icon', href: siteIconHref.value }],
        meta: () => {
            const seo = config.value?.seo
            if (!seo) return []
            const lang = locale.value
            const tags: Array<{ name?: string; property?: string; content: string }> = []
            const keywords = getLocalizedText(seo.keywords, lang)
            if (keywords) {
                tags.push({ name: 'keywords', content: keywords })
            }
            const description = getLocalizedText(seo.description, lang)
            if (description) {
                tags.push({ name: 'description', content: description })
            }
            return tags
        }
    })

    // 更新SEO信息 (向后兼容的方法)
    const applySEO = () => {
        // 由于 useHead 已经是响应式的，这里不再需要显式调用
        // 留空函数以防其他组件出错
    }

    // 根据 /public/config 响应校准服务器与客户端的时间偏移量。
    // 该接口由 CDN 边缘缓存最多 10s(浏览器不缓存),响应体里的 server_time
    // 可能已过去 Age 秒,estimateServerNow 会把 Age 头加回去;server_time 缺失时不更新偏移。
    const syncServerTimeOffset = (
        response: { data: any; headers?: Headers },
        requestTime: number,
        responseTime: number,
    ) => {
        const estimatedServerNow = estimateServerNow({
            serverTime: response.data?.data?.server_time,
            ageHeader: response.headers?.get('age'),
            requestTime,
            responseTime,
        })
        if (estimatedServerNow === null) return
        serverTimeOffset.value = estimatedServerNow - responseTime
    }

    // 加载全局配置
    const loadConfig = async (force = false) => {
        if (config.value && !force) {
            applySEO()
            applyCustomScripts(config.value?.scripts)
            return
        }
        // Fast path：服务端已经把 initial config 注入到 window.__PUBLIC_CONFIG__，
        // 首次直接同步使用，省掉首帧 /public/config RTT。
        // 后台 refetch 一次拿完整响应（payment_channels / server_time / app_version 等
        // 为了减小首屏 payload 没有被 inline 的字段）。
        if (!force && typeof window !== 'undefined' && (window as any).__PUBLIC_CONFIG__) {
            config.value = (window as any).__PUBLIC_CONFIG__
            configComplete.value = false
            applySEO()
            applyCustomScripts(config.value?.scripts)
            void (async () => {
                try {
                    const requestTime = Date.now()
                    const response = await configAPI.get()
                    const responseTime = Date.now()
                    config.value = response.data.data
                    configComplete.value = true
                    syncServerTimeOffset(response, requestTime, responseTime)
                    applySEO()
                    applyCustomScripts(config.value?.scripts)
                } catch (error) {
                    // 后台 refresh 失败不影响首帧；下次 force=true 的调用会重试
                    console.error('Background config refresh failed:', error)
                }
            })()
            return
        }
        if (!force) loading.value = true
        try {
            const requestTime = Date.now()
            const response = await configAPI.get()
            const responseTime = Date.now()
            config.value = response.data.data
            configComplete.value = true
            syncServerTimeOffset(response, requestTime, responseTime)
            applySEO()
            applyCustomScripts(config.value?.scripts)
            // Print version to console
            if (config.value?.app_version) {
                console.log(
                    '%c Version %c ' + config.value.app_version + ' %c',
                    'background:#34c759;color:#fff;padding:2px 6px;border-radius:4px 0 0 4px;font-weight:bold;',
                    'background:#1d1d1f;color:#f5f5f7;padding:2px 6px;border-radius:0 4px 4px 0;',
                    'background:transparent;',
                )
            }
        } catch (error) {
            console.error('Failed to load config:', error)
        } finally {
            if (!force) loading.value = false
        }
    }

    // 获取校正后的服务器当前时间（毫秒时间戳）
    const getServerTime = () => Date.now() + serverTimeOffset.value

    // 获取校正后的服务器当前 Date 对象
    const getServerDate = () => new Date(getServerTime())

    return {
        locale,
        config,
        configComplete,
        loading,
        serverTimeOffset,
        isResellerTenant,
        canAccessResellerConsole,
        setLocale,
        loadConfig,
        applySEO,
        getServerTime,
        getServerDate,
    }
})
