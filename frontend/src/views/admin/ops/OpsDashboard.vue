<template>
  <component :is="isFullscreen ? 'div' : AppLayout" :class="isFullscreen ? 'flex min-h-screen flex-col justify-center bg-gray-50 dark:bg-dark-950' : ''">
    <div :class="[isFullscreen ? 'p-4 md:p-6' : '', 'space-y-6 pb-12']">
      <div
        v-if="errorMessage"
        class="rounded-2xl bg-red-50 p-4 text-sm text-red-600 dark:bg-red-900/20 dark:text-red-400"
      >
        {{ errorMessage }}
      </div>

      <OpsDashboardSkeleton v-if="loading && !hasLoadedOnce" :fullscreen="isFullscreen" />

      <OpsDashboardHeader
        v-else-if="opsEnabled"
        :overview="overview"
        :platform="platform"
        :group-id="groupId"
        :api-key-ids="selectedApiKeyIds"
        :time-range="timeRange"
        :query-mode="queryMode"
        :loading="loading"
        :last-updated="lastUpdated"
        :thresholds="metricThresholds"
        :auto-refresh-enabled="autoRefreshEnabled"
        :auto-refresh-countdown="autoRefreshCountdown"
        :fullscreen="isFullscreen"
        :custom-start-time="customStartTime"
        :custom-end-time="customEndTime"
        @update:time-range="onTimeRangeChange"
        @update:platform="onPlatformChange"
        @update:group="onGroupChange"
        @open-api-key-filter="openApiKeyFilter"
        @update:query-mode="onQueryModeChange"
        @update:custom-time-range="onCustomTimeRangeChange"
        @refresh="fetchData"
        @open-request-details="handleOpenRequestDetails"
        @open-error-details="openErrorDetails"
        @open-settings="showSettingsDialog = true"
        @open-alert-rules="showAlertRulesCard = true"
        @enter-fullscreen="enterFullscreen"
        @exit-fullscreen="exitFullscreen"
      />

      <!-- Row: Concurrency + Throughput -->
      <div v-if="opsEnabled && !(loading && !hasLoadedOnce)" class="grid grid-cols-1 gap-6 lg:grid-cols-4">
        <div class="lg:col-span-1 min-h-[360px]">
          <OpsConcurrencyCard :platform-filter="platform" :group-id-filter="groupId" :refresh-token="dashboardRefreshToken" />
        </div>
        <div class="lg:col-span-1 h-[360px]">
          <OpsSwitchRateTrendChart
            :points="switchTrend?.points ?? []"
            :loading="loadingSwitchTrend"
            :time-range="switchTrendTimeRange"
            :fullscreen="isFullscreen"
          />
        </div>
        <div class="lg:col-span-2 h-[360px]">
          <OpsThroughputTrendChart
            :points="throughputTrend?.points ?? []"
            :by-platform="throughputTrend?.by_platform ?? []"
            :top-groups="throughputTrend?.top_groups ?? []"
            :loading="loadingTrend"
            :time-range="timeRange"
            :fullscreen="isFullscreen"
            @select-platform="handleThroughputSelectPlatform"
            @select-group="handleThroughputSelectGroup"
            @open-details="handleOpenRequestDetails"
          />
        </div>
      </div>

      <!-- Row: Visual Analysis (baseline 3-up grid) -->
      <div v-if="opsEnabled && !(loading && !hasLoadedOnce)" class="grid grid-cols-1 gap-6 md:grid-cols-3">
        <OpsLatencyChart :latency-data="latencyHistogram" :loading="loadingLatency" />
        <OpsErrorDistributionChart
          :data="errorDistribution"
          :loading="loadingErrorDistribution"
          @open-details="openErrorDetails('request')"
        />
        <OpsErrorTrendChart
          :points="errorTrend?.points ?? []"
          :loading="loadingErrorTrend"
          :time-range="timeRange"
          @open-request-errors="openErrorDetails('request')"
          @open-upstream-errors="openErrorDetails('upstream')"
        />
      </div>

      <!-- Row: OpenAI Token Stats -->
      <div v-if="opsEnabled && showOpenAITokenStats && !(loading && !hasLoadedOnce)" class="grid grid-cols-1 gap-6">
        <OpsOpenAITokenStatsCard
          :platform-filter="platform"
          :group-id-filter="groupId"
          :api-key-ids="selectedApiKeyIds"
          :refresh-token="dashboardRefreshToken"
        />
      </div>

      <!-- Alert Events -->
      <OpsAlertEventsCard v-if="opsEnabled && showAlertEvents && !(loading && !hasLoadedOnce)" />

      <!-- System Logs -->
      <OpsSystemLogTable
        v-if="opsEnabled && !(loading && !hasLoadedOnce)"
        :platform-filter="platform"
        :refresh-token="dashboardRefreshToken"
      />

      <!-- Settings Dialog (hidden in fullscreen mode) -->
      <template v-if="!isFullscreen">
        <BaseDialog :show="showApiKeyFilter" :title="t('admin.ops.sourceKeyFilter.title')" width="wide" @close="closeApiKeyFilter">
          <div class="space-y-4">
            <input
              v-model="apiKeySearch"
              type="search"
              :placeholder="t('admin.ops.sourceKeyFilter.searchPlaceholder')"
              class="w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 dark:border-dark-600 dark:bg-dark-800 dark:text-white"
            />

            <div class="flex items-center justify-between gap-3 text-xs">
              <span class="text-gray-500 dark:text-gray-400">{{ t('admin.ops.sourceKeyFilter.selectedCount', { count: pendingApiKeyIds.length }) }}</span>
              <button type="button" class="font-semibold text-blue-600 hover:text-blue-700 dark:text-blue-400" @click="selectVisibleApiKeys">
                {{ t('admin.ops.sourceKeyFilter.selectVisible') }}
              </button>
            </div>

            <div class="max-h-[420px] overflow-y-auto rounded-lg border border-gray-200 dark:border-dark-700">
              <div v-if="loadingApiKeys" class="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.loading') }}</div>
              <div v-else-if="apiKeyOptions.length === 0" class="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('common.noData') }}</div>
              <label
                v-for="key in apiKeyOptions"
                :key="key.id"
                class="flex cursor-pointer items-center gap-3 border-b border-gray-100 px-4 py-3 last:border-b-0 hover:bg-gray-50 dark:border-dark-700 dark:hover:bg-dark-800"
              >
                <input
                  type="checkbox"
                  class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  :checked="pendingApiKeyIds.includes(key.id)"
                  @change="toggleApiKey(key.id)"
                />
                <span class="min-w-0 flex-1 truncate text-sm font-medium text-gray-800 dark:text-gray-100">{{ key.name || `Key #${key.id}` }}</span>
                <span class="font-mono text-xs text-gray-400">#{{ key.id }}</span>
              </label>
            </div>
          </div>
          <template #footer>
            <div class="flex items-center justify-between gap-3">
              <button type="button" class="btn btn-secondary" @click="clearApiKeyFilter">{{ t('admin.ops.sourceKeyFilter.clearAndApply') }}</button>
              <div class="flex gap-3">
                <button type="button" class="btn btn-secondary" @click="closeApiKeyFilter">{{ t('common.cancel') }}</button>
                <button type="button" class="btn btn-primary" @click="applyApiKeyFilter">{{ t('admin.ops.sourceKeyFilter.apply') }}</button>
              </div>
            </div>
          </template>
        </BaseDialog>

        <OpsSettingsDialog :show="showSettingsDialog" @close="showSettingsDialog = false" @saved="onSettingsSaved" />

        <BaseDialog :show="showAlertRulesCard" :title="t('admin.ops.alertRules.title')" width="extra-wide" @close="showAlertRulesCard = false">
          <OpsAlertRulesCard />
        </BaseDialog>

        <OpsErrorDetailsModal
          :show="showErrorDetails"
          :time-range="timeRange"
          :custom-start-time="customStartTime"
          :custom-end-time="customEndTime"
          :platform="platform"
          :group-id="groupId"
          :api-key-ids="selectedApiKeyIds"
          :error-type="errorDetailsType"
          @update:show="showErrorDetails = $event"
          @openErrorDetail="openError"
        />

        <OpsErrorDetailModal v-model:show="showErrorModal" :error-id="selectedErrorId" :error-type="errorDetailsType" />

        <OpsRequestDetailsModal
          v-model="showRequestDetails"
          :time-range="timeRange"
          :preset="requestDetailsPreset"
          :platform="platform"
          :group-id="groupId"
          :api-key-ids="selectedApiKeyIds"
          @openErrorDetail="openError"
        />
      </template>
    </div>
  </component>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useDebounceFn, useIntervalFn } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api'
import {
  opsAPI,
  type OpsDashboardOverview,
  type OpsErrorDistributionResponse,
  type OpsErrorTrendResponse,
  type OpsLatencyHistogramResponse,
  type OpsThroughputTrendResponse,
  type OpsMetricThresholds
} from '@/api/admin/ops'
import type { SimpleApiKey } from '@/api/admin/usage'
import { useAdminSettingsStore, useAppStore, useAuthStore } from '@/stores'
import OpsDashboardHeader from './components/OpsDashboardHeader.vue'
import OpsDashboardSkeleton from './components/OpsDashboardSkeleton.vue'
import OpsConcurrencyCard from './components/OpsConcurrencyCard.vue'
import OpsErrorDetailModal from './components/OpsErrorDetailModal.vue'
import OpsErrorDistributionChart from './components/OpsErrorDistributionChart.vue'
import OpsErrorDetailsModal from './components/OpsErrorDetailsModal.vue'
import OpsErrorTrendChart from './components/OpsErrorTrendChart.vue'
import OpsLatencyChart from './components/OpsLatencyChart.vue'
import OpsThroughputTrendChart from './components/OpsThroughputTrendChart.vue'
import OpsSwitchRateTrendChart from './components/OpsSwitchRateTrendChart.vue'
import OpsAlertEventsCard from './components/OpsAlertEventsCard.vue'
import OpsOpenAITokenStatsCard from './components/OpsOpenAITokenStatsCard.vue'
import OpsSystemLogTable from './components/OpsSystemLogTable.vue'
import OpsRequestDetailsModal, { type OpsRequestDetailsPreset } from './components/OpsRequestDetailsModal.vue'
import OpsSettingsDialog from './components/OpsSettingsDialog.vue'
import OpsAlertRulesCard from './components/OpsAlertRulesCard.vue'

const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const adminSettingsStore = useAdminSettingsStore()
const authStore = useAuthStore()
const { t } = useI18n()

const opsEnabled = computed(() => adminSettingsStore.opsMonitoringEnabled)

type TimeRange = '5m' | '30m' | '1h' | '6h' | '24h' | 'custom'
const allowedTimeRanges = new Set<TimeRange>(['5m', '30m', '1h', '6h', '24h', 'custom'])

type QueryMode = 'auto' | 'raw' | 'preagg'
const allowedQueryModes = new Set<QueryMode>(['auto', 'raw', 'preagg'])

const loading = ref(true)
const hasLoadedOnce = ref(false)
const errorMessage = ref('')
const lastUpdated = ref<Date | null>(new Date())

const timeRange = ref<TimeRange>('1h')
const platform = ref<string>('')
const groupId = ref<number | null>(null)
const selectedApiKeyIds = ref<number[]>([])
const pendingApiKeyIds = ref<number[]>([])
const showApiKeyFilter = ref(false)
const apiKeySearch = ref('')
const apiKeyOptions = ref<SimpleApiKey[]>([])
const loadingApiKeys = ref(false)
const queryMode = ref<QueryMode>('auto')
const customStartTime = ref<string | null>(null)
const customEndTime = ref<string | null>(null)
const switchTrendWindowHours = 5
const switchTrendTimeRange = `${switchTrendWindowHours}h`
const switchTrendWindowMs = switchTrendWindowHours * 60 * 60 * 1000

const QUERY_KEYS = {
  timeRange: 'tr',
  platform: 'platform',
  groupId: 'group_id',
  apiKeyIds: 'api_key_ids',
  queryMode: 'mode',
  fullscreen: 'fullscreen',

  // Deep links
  openErrorDetails: 'open_error_details',
  errorType: 'error_type',
  alertRuleId: 'alert_rule_id',
  openAlertRules: 'open_alert_rules'
} as const

const isApplyingRouteQuery = ref(false)
const isSyncingRouteQuery = ref(false)

// Fullscreen mode
const isFullscreen = computed(() => {
  const val = route.query[QUERY_KEYS.fullscreen]
  return val === '1' || val === 'true'
})

function exitFullscreen() {
  const nextQuery = { ...route.query }
  delete nextQuery[QUERY_KEYS.fullscreen]
  router.replace({ query: nextQuery })
}

function enterFullscreen() {
  const nextQuery = { ...route.query, [QUERY_KEYS.fullscreen]: '1' }
  router.replace({ query: nextQuery })
}

function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && isFullscreen.value) {
    exitFullscreen()
  }
}

let dashboardFetchController: AbortController | null = null
let dashboardFetchSeq = 0

function isCanceledRequest(err: unknown): boolean {
  return (
    !!err &&
    typeof err === 'object' &&
    'code' in err &&
    (err as Record<string, unknown>).code === 'ERR_CANCELED'
  )
}

function abortDashboardFetch() {
  if (dashboardFetchController) {
    dashboardFetchController.abort()
    dashboardFetchController = null
  }
}

const readQueryString = (key: string): string => {
  const value = route.query[key]
  if (typeof value === 'string') return value
  if (Array.isArray(value) && typeof value[0] === 'string') return value[0]
  return ''
}

const readQueryNumber = (key: string): number | null => {
  const raw = readQueryString(key)
  if (!raw) return null
  const n = Number.parseInt(raw, 10)
  return Number.isFinite(n) ? n : null
}

function normalizeApiKeyIds(ids: number[]): number[] {
  return [...new Set(ids.filter((id) => Number.isInteger(id) && id > 0))]
    .sort((a, b) => a - b)
    .slice(0, 100)
}

function parseApiKeyIds(raw: string): number[] {
  if (!raw) return []
  return normalizeApiKeyIds(raw.split(',').map((value) => Number.parseInt(value, 10)))
}

const apiKeyFilterStorageKey = computed(() => `ops-dashboard:key-filter:${authStore.user?.id ?? 'anonymous'}`)

function readPersistedApiKeyFilter(): number[] {
  try {
    const raw = window.localStorage.getItem(apiKeyFilterStorageKey.value)
    const ids = raw ? JSON.parse(raw) : []
    return Array.isArray(ids) ? normalizeApiKeyIds(ids.map(Number)) : []
  } catch {
    return []
  }
}

function persistApiKeyFilter(): void {
  window.localStorage.setItem(apiKeyFilterStorageKey.value, JSON.stringify(selectedApiKeyIds.value))
}

const applyRouteQueryToState = () => {
  const nextTimeRange = readQueryString(QUERY_KEYS.timeRange)
  if (nextTimeRange && allowedTimeRanges.has(nextTimeRange as TimeRange)) {
    timeRange.value = nextTimeRange as TimeRange
  }

  platform.value = readQueryString(QUERY_KEYS.platform) || ''

  const groupIdRaw = readQueryNumber(QUERY_KEYS.groupId)
  groupId.value = typeof groupIdRaw === 'number' && groupIdRaw > 0 ? groupIdRaw : null

  if (Object.prototype.hasOwnProperty.call(route.query, QUERY_KEYS.apiKeyIds)) {
    selectedApiKeyIds.value = parseApiKeyIds(readQueryString(QUERY_KEYS.apiKeyIds))
  }

  const nextMode = readQueryString(QUERY_KEYS.queryMode)
  if (nextMode && allowedQueryModes.has(nextMode as QueryMode)) {
    queryMode.value = nextMode as QueryMode
  } else {
    const fallback = adminSettingsStore.opsQueryModeDefault || 'auto'
    queryMode.value = allowedQueryModes.has(fallback as QueryMode) ? (fallback as QueryMode) : 'auto'
  }

  // Deep links
  const openRules = readQueryString(QUERY_KEYS.openAlertRules)
  if (openRules === '1' || openRules === 'true') {
    showAlertRulesCard.value = true
  }

  const ruleID = readQueryNumber(QUERY_KEYS.alertRuleId)
  if (typeof ruleID === 'number' && ruleID > 0) {
    showAlertRulesCard.value = true
  }

  const openErr = readQueryString(QUERY_KEYS.openErrorDetails)
  if (openErr === '1' || openErr === 'true') {
    const typ = readQueryString(QUERY_KEYS.errorType)
    errorDetailsType.value = typ === 'upstream' ? 'upstream' : 'request'
    showErrorDetails.value = true
  }
}

const buildQueryFromState = () => {
  const next: Record<string, any> = { ...route.query }

  Object.values(QUERY_KEYS).forEach((k) => {
    delete next[k]
  })

  if (timeRange.value !== '1h') next[QUERY_KEYS.timeRange] = timeRange.value
  if (platform.value) next[QUERY_KEYS.platform] = platform.value
  if (typeof groupId.value === 'number' && groupId.value > 0) next[QUERY_KEYS.groupId] = String(groupId.value)
  if (selectedApiKeyIds.value.length > 0) next[QUERY_KEYS.apiKeyIds] = selectedApiKeyIds.value.join(',')
  if (queryMode.value !== 'auto') next[QUERY_KEYS.queryMode] = queryMode.value

  return next
}

const syncQueryToRoute = useDebounceFn(async () => {
  if (isApplyingRouteQuery.value) return
  const nextQuery = buildQueryFromState()

  const curr = route.query as Record<string, any>
  const nextKeys = Object.keys(nextQuery)
  const currKeys = Object.keys(curr)
  const sameLength = nextKeys.length === currKeys.length
  const sameValues = sameLength && nextKeys.every((k) => String(curr[k] ?? '') === String(nextQuery[k] ?? ''))
  if (sameValues) return

  try {
    isSyncingRouteQuery.value = true
    await router.replace({ query: nextQuery })
  } finally {
    isSyncingRouteQuery.value = false
  }
}, 250)

const overview = ref<OpsDashboardOverview | null>(null)
const metricThresholds = ref<OpsMetricThresholds | null>(null)

const throughputTrend = ref<OpsThroughputTrendResponse | null>(null)
const loadingTrend = ref(false)

const switchTrend = ref<OpsThroughputTrendResponse | null>(null)
const loadingSwitchTrend = ref(false)

const latencyHistogram = ref<OpsLatencyHistogramResponse | null>(null)
const loadingLatency = ref(false)

const errorTrend = ref<OpsErrorTrendResponse | null>(null)
const loadingErrorTrend = ref(false)

const errorDistribution = ref<OpsErrorDistributionResponse | null>(null)
const loadingErrorDistribution = ref(false)

const selectedErrorId = ref<number | null>(null)
const showErrorModal = ref(false)

const showErrorDetails = ref(false)
const errorDetailsType = ref<'request' | 'upstream'>('request')

const showRequestDetails = ref(false)
const requestDetailsPreset = ref<OpsRequestDetailsPreset>({
  title: '',
  kind: 'all',
  sort: 'created_at_desc'
})

const showSettingsDialog = ref(false)
const showAlertRulesCard = ref(false)

applyRouteQueryToState()

// Auto refresh settings
const showAlertEvents = ref(true)
const showOpenAITokenStats = ref(false)
const autoRefreshEnabled = ref(false)
const autoRefreshIntervalMs = ref(30000) // default 30 seconds
const autoRefreshCountdown = ref(0)

// Used to trigger child component refreshes in a single shared cadence.
const dashboardRefreshToken = ref(0)

// Countdown timer (drives auto refresh; updates every second)
const { pause: pauseCountdown, resume: resumeCountdown } = useIntervalFn(
  () => {
    if (!autoRefreshEnabled.value) return
    if (!opsEnabled.value) return
    if (loading.value) return

    if (autoRefreshCountdown.value <= 0) {
      // Fetch immediately when the countdown reaches 0.
      // fetchData() will reset the countdown to the full interval.
      fetchData()
      return
    }

    autoRefreshCountdown.value -= 1
  },
  1000,
  { immediate: false }
)

// Load ops dashboard presentation settings from backend.
async function loadDashboardAdvancedSettings() {
  try {
    const settings = await opsAPI.getAdvancedSettings()
    showAlertEvents.value = settings.display_alert_events
    showOpenAITokenStats.value = settings.display_openai_token_stats
    autoRefreshEnabled.value = settings.auto_refresh_enabled
    autoRefreshIntervalMs.value = settings.auto_refresh_interval_seconds * 1000
    autoRefreshCountdown.value = settings.auto_refresh_interval_seconds
  } catch (err) {
    console.error('[OpsDashboard] Failed to load dashboard advanced settings', err)
    showAlertEvents.value = true
    showOpenAITokenStats.value = false
    autoRefreshEnabled.value = false
    autoRefreshIntervalMs.value = 30000
    autoRefreshCountdown.value = 0
  }
}

async function loadApiKeyOptions() {
  loadingApiKeys.value = true
  try {
    apiKeyOptions.value = await adminAPI.usage.searchApiKeys(undefined, apiKeySearch.value.trim(), 100)
  } catch (err) {
    console.error('[OpsDashboard] Failed to load API key filter options', err)
    apiKeyOptions.value = []
  } finally {
    loadingApiKeys.value = false
  }
}

const reloadApiKeyOptions = useDebounceFn(() => {
  void loadApiKeyOptions()
}, 250)

function openApiKeyFilter() {
  pendingApiKeyIds.value = [...selectedApiKeyIds.value]
  apiKeySearch.value = ''
  showApiKeyFilter.value = true
  void loadApiKeyOptions()
}

function closeApiKeyFilter() {
  pendingApiKeyIds.value = [...selectedApiKeyIds.value]
  showApiKeyFilter.value = false
}

function toggleApiKey(id: number) {
  const next = new Set(pendingApiKeyIds.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  pendingApiKeyIds.value = normalizeApiKeyIds([...next])
}

function selectVisibleApiKeys() {
  pendingApiKeyIds.value = normalizeApiKeyIds([
    ...pendingApiKeyIds.value,
    ...apiKeyOptions.value.map((key) => key.id)
  ])
}

function applyApiKeyFilter() {
  selectedApiKeyIds.value = normalizeApiKeyIds(pendingApiKeyIds.value)
  persistApiKeyFilter()
  showApiKeyFilter.value = false
}

function clearApiKeyFilter() {
  selectedApiKeyIds.value = []
  pendingApiKeyIds.value = []
  persistApiKeyFilter()
  showApiKeyFilter.value = false
}

function handleThroughputSelectPlatform(nextPlatform: string) {
  platform.value = nextPlatform || ''
  groupId.value = null
}

function handleThroughputSelectGroup(nextGroupId: number) {
  const id = Number.isFinite(nextGroupId) && nextGroupId > 0 ? nextGroupId : null
  groupId.value = id
}

function handleOpenRequestDetails(preset?: OpsRequestDetailsPreset) {
  const basePreset: OpsRequestDetailsPreset = {
    title: t('admin.ops.requestDetails.title'),
    kind: 'all',
    sort: 'created_at_desc'
  }

  requestDetailsPreset.value = { ...basePreset, ...(preset ?? {}) }
  if (!requestDetailsPreset.value.title) requestDetailsPreset.value.title = basePreset.title
  // Ensure only one modal visible at a time.
  showErrorDetails.value = false
  showErrorModal.value = false
  showRequestDetails.value = true
}

function openErrorDetails(kind: 'request' | 'upstream') {
  errorDetailsType.value = kind
  // Ensure only one modal visible at a time.
  showRequestDetails.value = false
  showErrorModal.value = false
  showErrorDetails.value = true
}

function onTimeRangeChange(v: string | number | boolean | null) {
  if (typeof v !== 'string') return
  if (!allowedTimeRanges.has(v as TimeRange)) return
  timeRange.value = v as TimeRange
}

function onCustomTimeRangeChange(startTime: string, endTime: string) {
  customStartTime.value = startTime
  customEndTime.value = endTime
}

async function onSettingsSaved() {
  await loadDashboardAdvancedSettings()
  loadThresholds()
  fetchData()
}

function onPlatformChange(v: string | number | boolean | null) {
  platform.value = typeof v === 'string' ? v : ''
}

function onGroupChange(v: string | number | boolean | null) {
  if (v === null) {
    groupId.value = null
    return
  }
  if (typeof v === 'number') {
    groupId.value = v > 0 ? v : null
    return
  }
  if (typeof v === 'string') {
    const n = Number.parseInt(v, 10)
    groupId.value = Number.isFinite(n) && n > 0 ? n : null
  }
}

function onQueryModeChange(v: string | number | boolean | null) {
  if (typeof v !== 'string') return
  if (!allowedQueryModes.has(v as QueryMode)) return
  queryMode.value = v as QueryMode
}

function openError(id: number) {
  selectedErrorId.value = id
  // Ensure only one modal visible at a time.
  showErrorDetails.value = false
  showRequestDetails.value = false
  showErrorModal.value = true
}

function buildApiParams() {
  const params: any = {
    platform: platform.value || undefined,
    group_id: groupId.value ?? undefined,
    api_key_ids: selectedApiKeyIds.value.length ? selectedApiKeyIds.value.join(',') : undefined,
    mode: queryMode.value
  }

  if (timeRange.value === 'custom') {
    if (customStartTime.value && customEndTime.value) {
      params.start_time = customStartTime.value
      params.end_time = customEndTime.value
    } else {
      // Safety fallback: avoid sending time_range=custom (backend may not support it)
      params.time_range = '1h'
    }
  } else {
    params.time_range = timeRange.value
  }

  return params
}

function buildSwitchTrendParams() {
  const params: any = {
    platform: platform.value || undefined,
    group_id: groupId.value ?? undefined,
    api_key_ids: selectedApiKeyIds.value.length ? selectedApiKeyIds.value.join(',') : undefined,
    mode: queryMode.value
  }
  const endTime = new Date()
  const startTime = new Date(endTime.getTime() - switchTrendWindowMs)
  params.start_time = startTime.toISOString()
  params.end_time = endTime.toISOString()
  return params
}

async function refreshOverviewWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  try {
    const data = await opsAPI.getDashboardOverview(buildApiParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    overview.value = data
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    overview.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadOverview'))
  }
}

async function refreshSwitchTrendWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  loadingSwitchTrend.value = true
  try {
    const data = await opsAPI.getThroughputTrend(buildSwitchTrendParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    switchTrend.value = data
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    switchTrend.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadSwitchTrend'))
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loadingSwitchTrend.value = false
    }
  }
}

async function refreshThroughputTrendWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  loadingTrend.value = true
  try {
    const data = await opsAPI.getThroughputTrend(buildApiParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    throughputTrend.value = data
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    throughputTrend.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadThroughputTrend'))
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loadingTrend.value = false
    }
  }
}

async function refreshCoreSnapshotWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  loadingTrend.value = true
  loadingErrorTrend.value = true
  try {
    const data = await opsAPI.getDashboardSnapshotV2(buildApiParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    overview.value = data.overview
    throughputTrend.value = data.throughput_trend
    errorTrend.value = data.error_trend
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    // Fallback to legacy split endpoints when snapshot endpoint is unavailable.
    await Promise.all([
      refreshOverviewWithCancel(fetchSeq, signal),
      refreshThroughputTrendWithCancel(fetchSeq, signal),
      refreshErrorTrendWithCancel(fetchSeq, signal)
    ])
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loadingTrend.value = false
      loadingErrorTrend.value = false
    }
  }
}

async function refreshLatencyHistogramWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  loadingLatency.value = true
  try {
    const data = await opsAPI.getLatencyHistogram(buildApiParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    latencyHistogram.value = data
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    latencyHistogram.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadLatencyHistogram'))
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loadingLatency.value = false
    }
  }
}

async function refreshErrorTrendWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  loadingErrorTrend.value = true
  try {
    const data = await opsAPI.getErrorTrend(buildApiParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    errorTrend.value = data
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    errorTrend.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadErrorTrend'))
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loadingErrorTrend.value = false
    }
  }
}

async function refreshErrorDistributionWithCancel(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  loadingErrorDistribution.value = true
  try {
    const data = await opsAPI.getErrorDistribution(buildApiParams(), { signal })
    if (fetchSeq !== dashboardFetchSeq) return
    errorDistribution.value = data
  } catch (err: any) {
    if (fetchSeq !== dashboardFetchSeq || isCanceledRequest(err)) return
    errorDistribution.value = null
    appStore.showError(err?.message || t('admin.ops.failedToLoadErrorDistribution'))
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loadingErrorDistribution.value = false
    }
  }
}

async function refreshDeferredPanels(fetchSeq: number, signal: AbortSignal) {
  if (!opsEnabled.value) return
  await Promise.all([
    refreshLatencyHistogramWithCancel(fetchSeq, signal),
    refreshErrorDistributionWithCancel(fetchSeq, signal)
  ])
}

function isOpsDisabledError(err: unknown): boolean {
  return (
    !!err &&
    typeof err === 'object' &&
    'code' in err &&
    typeof (err as Record<string, unknown>).code === 'string' &&
    (err as Record<string, unknown>).code === 'OPS_DISABLED'
  )
}

async function fetchData() {
  if (!opsEnabled.value) return

  abortDashboardFetch()
  dashboardFetchSeq += 1
  const fetchSeq = dashboardFetchSeq
  dashboardFetchController = new AbortController()

  loading.value = true
  errorMessage.value = ''
  try {
    await Promise.all([
      refreshCoreSnapshotWithCancel(fetchSeq, dashboardFetchController.signal),
      refreshSwitchTrendWithCancel(fetchSeq, dashboardFetchController.signal),
    ])
    if (fetchSeq !== dashboardFetchSeq) return

    lastUpdated.value = new Date()

    // Trigger child component refreshes using the same cadence as the header.
    dashboardRefreshToken.value += 1

    // Reset auto refresh countdown after successful fetch
    if (autoRefreshEnabled.value) {
      autoRefreshCountdown.value = Math.floor(autoRefreshIntervalMs.value / 1000)
    }

    // Defer non-core visual panels to reduce initial blocking.
    void refreshDeferredPanels(fetchSeq, dashboardFetchController.signal)
  } catch (err) {
    if (!isOpsDisabledError(err)) {
      console.error('[ops] failed to fetch dashboard data', err)
      errorMessage.value = t('admin.ops.failedToLoadData')
    }
  } finally {
    if (fetchSeq === dashboardFetchSeq) {
      loading.value = false
      hasLoadedOnce.value = true
    }
  }
}

watch(
  () => [timeRange.value, platform.value, groupId.value, selectedApiKeyIds.value.join(','), queryMode.value] as const,
  () => {
    if (isApplyingRouteQuery.value) return
    if (opsEnabled.value) {
      fetchData()
    }
    syncQueryToRoute()
  }
)

watch(
  () => route.query,
  () => {
    if (isSyncingRouteQuery.value) return

    const prevTimeRange = timeRange.value
    const prevPlatform = platform.value
    const prevGroupId = groupId.value
    const prevApiKeyIds = selectedApiKeyIds.value.join(',')

    isApplyingRouteQuery.value = true
    applyRouteQueryToState()
    isApplyingRouteQuery.value = false

    const changed =
      prevTimeRange !== timeRange.value ||
      prevPlatform !== platform.value ||
      prevGroupId !== groupId.value ||
      prevApiKeyIds !== selectedApiKeyIds.value.join(',')
    if (changed) {
      if (opsEnabled.value) {
        fetchData()
      }
    }
  }
)

watch(apiKeySearch, () => {
  if (showApiKeyFilter.value) reloadApiKeyOptions()
})

onMounted(async () => {
  // Fullscreen mode: listen for ESC key
  window.addEventListener('keydown', handleKeydown)

  if (!Object.prototype.hasOwnProperty.call(route.query, QUERY_KEYS.apiKeyIds)) {
    selectedApiKeyIds.value = readPersistedApiKeyFilter()
  }

  await adminSettingsStore.fetch()
  if (!adminSettingsStore.opsMonitoringEnabled) {
    await router.replace('/admin/settings')
    return
  }

  // Load thresholds configuration
  loadThresholds()

  // Load auto refresh settings
  await loadDashboardAdvancedSettings()

  if (opsEnabled.value) {
    await fetchData()
  }

  // Start auto refresh if enabled
  if (autoRefreshEnabled.value) {
    resumeCountdown()
  }
})

async function loadThresholds() {
  try {
    const thresholds = await opsAPI.getMetricThresholds()
    metricThresholds.value = thresholds || null
  } catch (err) {
    console.warn('[OpsDashboard] Failed to load thresholds', err)
    metricThresholds.value = null
  }
}

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
  abortDashboardFetch()
  pauseCountdown()
})

// Watch auto refresh settings changes
watch(autoRefreshEnabled, (enabled) => {
  if (enabled) {
    autoRefreshCountdown.value = Math.floor(autoRefreshIntervalMs.value / 1000)
    resumeCountdown()
  } else {
    pauseCountdown()
    autoRefreshCountdown.value = 0
  }
})

// Reload auto refresh settings after settings dialog is closed
watch(showSettingsDialog, async (show) => {
  if (!show) {
    await loadDashboardAdvancedSettings()
  }
})
</script>
