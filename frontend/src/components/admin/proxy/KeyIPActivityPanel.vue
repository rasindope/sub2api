<template>
  <BaseDialog :show="show" :title="t('admin.proxies.ipActivity.title')" width="extra-wide" @close="emit('close')">
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
      <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.description') }}</p>
      <div class="flex items-center gap-2">
        <DateRangePicker v-model:start-date="startDate" v-model:end-date="endDate" @change="onRangeChange" />
        <button type="button" class="btn btn-secondary" :class="onlyIssues ? 'ring-2 ring-amber-400/40' : ''" @click="onlyIssues = !onlyIssues">
          {{ t('admin.proxies.ipActivity.onlyIssues') }}
        </button>
        <button type="button" class="btn btn-secondary" :class="thresholdsOpen ? 'ring-2 ring-primary-400/40' : ''" @click="toggleThresholds">
          {{ t('admin.proxies.ipActivity.thresholds') }}
        </button>
        <button type="button" class="btn btn-secondary" :disabled="loading" @click="load">
          {{ loading ? t('common.loading') : t('common.refresh') }}
        </button>
      </div>
    </div>

    <div v-if="thresholdsOpen" class="mb-4 rounded-xl border border-primary-200 bg-primary-50/40 p-4 dark:border-primary-500/20 dark:bg-primary-500/5">
      <div v-if="thresholdsLoading" class="py-4 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
      <form v-else @submit.prevent="saveThresholds">
        <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
          <label class="block">
            <span class="input-label">{{ t('admin.proxies.ipActivity.minimumOverlap') }}</span>
            <input v-model.number="thresholds.minimum_overlap_seconds" type="number" min="1" max="3600" required class="input" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.proxies.ipActivity.highSingleOverlap') }}</span>
            <input v-model.number="thresholds.high_single_overlap_seconds" type="number" :min="thresholds.minimum_overlap_seconds" max="86400" required class="input" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.proxies.ipActivity.highOverlapCount') }}</span>
            <input v-model.number="thresholds.high_overlap_count" type="number" min="1" max="100000" required class="input" />
          </label>
          <label class="block">
            <span class="input-label">{{ t('admin.proxies.ipActivity.highTotalOverlap') }}</span>
            <input v-model.number="thresholds.high_total_overlap_seconds" type="number" :min="thresholds.minimum_overlap_seconds" max="10000000" required class="input" />
          </label>
        </div>
        <div class="mt-3 flex flex-wrap items-center justify-between gap-3">
          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.thresholdDescription') }}</p>
          <button type="submit" class="btn btn-primary btn-sm" :disabled="thresholdsSaving">
            {{ thresholdsSaving ? t('common.saving') : t('common.save') }}
          </button>
        </div>
      </form>
    </div>

    <div class="mb-4 grid grid-cols-3 gap-2 sm:gap-4">
      <div class="rounded-xl border border-gray-200 bg-white p-3 dark:border-dark-700 dark:bg-dark-900">
        <div class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.activeKeys') }}</div>
        <div class="mt-1 text-2xl font-semibold tabular-nums text-gray-900 dark:text-white">{{ data.active_keys }}</div>
      </div>
      <div class="rounded-xl border border-amber-200 bg-amber-50/50 p-3 dark:border-amber-500/20 dark:bg-amber-500/5">
        <div class="text-xs text-amber-700 dark:text-amber-300">{{ t('admin.proxies.ipActivity.watch') }}</div>
        <div class="mt-1 text-2xl font-semibold tabular-nums text-amber-700 dark:text-amber-300">{{ data.watch_keys }}</div>
      </div>
      <div class="rounded-xl border border-red-200 bg-red-50/50 p-3 dark:border-red-500/20 dark:bg-red-500/5">
        <div class="text-xs text-red-700 dark:text-red-300">{{ t('admin.proxies.ipActivity.high') }}</div>
        <div class="mt-1 text-2xl font-semibold tabular-nums text-red-700 dark:text-red-300">{{ data.high_risk_keys }}</div>
      </div>
    </div>

    <div class="hidden overflow-hidden rounded-xl border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-900 sm:block">
      <table class="w-full text-left text-sm">
        <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-800 dark:text-gray-400">
          <tr>
            <th class="px-4 py-3 font-medium">Key</th><th class="px-4 py-3 font-medium">{{ t('admin.proxies.ipActivity.status') }}</th>
            <th class="px-4 py-3 text-right font-medium">{{ t('admin.proxies.ipActivity.rangeIPs') }}</th>
            <th class="px-4 py-3 text-right font-medium">{{ t('admin.proxies.ipActivity.overlaps') }}</th>
            <th class="px-4 py-3 text-right font-medium">{{ t('admin.proxies.ipActivity.maxOverlap') }}</th>
            <th class="px-4 py-3 text-right font-medium">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
          <tr v-for="item in visibleItems" :key="item.api_key_id" class="hover:bg-gray-50 dark:hover:bg-dark-800/60">
            <td class="px-4 py-3 font-medium text-gray-900 dark:text-white">{{ label(item) }}</td>
            <td class="px-4 py-3"><span class="rounded-full px-2 py-1 text-xs font-medium" :class="riskClass(item.risk_level)">{{ riskLabel(item.risk_level) }}</span></td>
            <td class="px-4 py-3 text-right tabular-nums">{{ item.distinct_ip_count }}</td>
            <td class="px-4 py-3 text-right tabular-nums">{{ item.overlap_count_15m }}</td>
            <td class="px-4 py-3 text-right tabular-nums">{{ seconds(item.max_overlap_seconds_15m) }}</td>
            <td class="px-4 py-3 text-right"><button class="text-primary-600 hover:text-primary-700 dark:text-primary-400" @click="selected = item">{{ t('admin.proxies.ipActivity.details') }}</button></td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="space-y-3 sm:hidden">
      <button v-for="item in visibleItems" :key="item.api_key_id" type="button" class="w-full rounded-xl border border-gray-200 bg-white p-4 text-left dark:border-dark-700 dark:bg-dark-900" @click="selected = item">
        <div class="flex items-center justify-between gap-2"><span class="truncate font-medium text-gray-900 dark:text-white">{{ label(item) }}</span><span class="rounded-full px-2 py-1 text-xs" :class="riskClass(item.risk_level)">{{ riskLabel(item.risk_level) }}</span></div>
        <div class="mt-3 grid grid-cols-2 gap-2 text-center text-xs text-gray-500 dark:text-gray-400">
          <div><strong class="block text-base text-gray-900 dark:text-white">{{ item.distinct_ip_count }}</strong>{{ t('admin.proxies.ipActivity.rangeIPs') }}</div>
          <div><strong class="block text-base text-gray-900 dark:text-white">{{ item.overlap_count_15m }}</strong>{{ t('admin.proxies.ipActivity.overlaps') }}</div>
        </div>
      </button>
    </div>

    <div v-if="!loading && !visibleItems.length" class="py-12 text-center text-sm text-gray-400">{{ t('admin.proxies.ipActivity.empty') }}</div>
    <ApiKeyIpDetailsDialog :show="Boolean(selected)" :item="selected" :start-date="startDate" :end-date="endDate" @close="selected = null" @geo-failed="appStore.showError(t('admin.dashboard.ipGeoFailed'))" />
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import BaseDialog from '@/components/common/BaseDialog.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import ApiKeyIpDetailsDialog from '@/components/charts/ApiKeyIpDetailsDialog.vue'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { ApiKeyIPActivityItem, ApiKeyIPActivityResponse, ApiKeyIPRiskLevel } from '@/types'
import type { APIKeyIPRiskSettings } from '@/api/admin/settings'

const { t } = useI18n()
const props = defineProps<{ show: boolean }>()
const emit = defineEmits<{ close: []; updated: [value: ApiKeyIPActivityResponse] }>()
const appStore = useAppStore()
const loading = ref(false)
const onlyIssues = ref(false)
const thresholdsOpen = ref(false)
const thresholdsLoading = ref(false)
const thresholdsSaving = ref(false)
const thresholdsLoaded = ref(false)
const thresholds = reactive<APIKeyIPRiskSettings>({ minimum_overlap_seconds: 10, high_single_overlap_seconds: 120, high_overlap_count: 5, high_total_overlap_seconds: 180 })
const selected = ref<ApiKeyIPActivityItem | null>(null)
const data = ref<ApiKeyIPActivityResponse>({ items: [], active_keys: 0, watch_keys: 0, high_risk_keys: 0, generated_at: '' })
const formatDate = (date: Date) => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
const today = formatDate(new Date())
const startDate = ref(today)
const endDate = ref(today)
const liveAutoRefresh = ref(true)
const visibleItems = computed(() => onlyIssues.value ? data.value.items.filter((item) => item.risk_level !== 'normal') : data.value.items)
let timer: ReturnType<typeof setInterval> | undefined

const load = async () => {
  loading.value = true
  try {
    data.value = await adminAPI.dashboard.getApiKeyIPActivity({ limit: 50, start_date: startDate.value, end_date: endDate.value })
    emit('updated', data.value)
  }
  catch { appStore.showError(t('admin.proxies.ipActivity.loadFailed')) }
  finally { loading.value = false }
}
const label = (item: ApiKeyIPActivityItem) => item.key_name || `Key #${item.api_key_id}`
const seconds = (value: number) => `${Number(value || 0).toFixed(1)}s`
const riskLabel = (risk: ApiKeyIPRiskLevel) => t(`admin.proxies.ipActivity.${risk}`)
const riskClass = (risk: ApiKeyIPRiskLevel) => ({
  normal: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300',
  watch: 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-300',
  high: 'bg-red-50 text-red-700 dark:bg-red-500/10 dark:text-red-300'
})[risk]

const toggleThresholds = async () => {
  thresholdsOpen.value = !thresholdsOpen.value
  if (!thresholdsOpen.value || thresholdsLoaded.value) return
  thresholdsLoading.value = true
  try {
    Object.assign(thresholds, await adminAPI.settings.getAPIKeyIPRiskSettings())
    thresholdsLoaded.value = true
  }
  catch (error: unknown) {
    thresholdsOpen.value = false
    appStore.showError(extractApiErrorMessage(error, t('admin.proxies.ipActivity.thresholdLoadFailed')))
  }
  finally { thresholdsLoading.value = false }
}

const saveThresholds = async () => {
  thresholdsSaving.value = true
  try {
    Object.assign(thresholds, await adminAPI.settings.updateAPIKeyIPRiskSettings({ ...thresholds }))
    appStore.showSuccess(t('admin.proxies.ipActivity.thresholdSaved'))
    load()
  }
  catch (error: unknown) { appStore.showError(extractApiErrorMessage(error, t('admin.proxies.ipActivity.thresholdSaveFailed'))) }
  finally { thresholdsSaving.value = false }
}

const stopPolling = () => { if (timer) clearInterval(timer); timer = undefined }
const syncPolling = () => {
  stopPolling()
  if (props.show && liveAutoRefresh.value) timer = setInterval(load, 30_000)
}
const onRangeChange = (range: { preset: string | null }) => {
  liveAutoRefresh.value = range.preset === 'today' || range.preset === 'last24Hours'
  selected.value = null
  load()
  syncPolling()
}

watch(() => props.show, (show) => {
  if (show) load()
  else selected.value = null
  syncPolling()
})
onMounted(() => { load(); syncPolling() })
onUnmounted(stopPolling)
</script>
