<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- Row 1: Core Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Total API Keys -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
                <Icon name="key" size="md" class="text-blue-600 dark:text-blue-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.apiKeys') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_api_keys }}
                </p>
                <p class="text-xs text-green-600 dark:text-green-400">
                  {{ stats.active_api_keys }} {{ t('common.active') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Service Accounts -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
                <Icon name="server" size="md" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.accounts') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_accounts }}
                </p>
                <p class="text-xs">
                  <span class="text-green-600 dark:text-green-400"
                    >{{ stats.normal_accounts }} {{ t('common.active') }}</span
                  >
                  <span v-if="stats.error_accounts > 0" class="ml-1 text-red-500"
                    >{{ stats.error_accounts }} {{ t('common.error') }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Today Requests -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
                <Icon name="chart" size="md" class="text-green-600 dark:text-green-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayRequests') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.today_requests }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
                </p>
              </div>
            </div>
          </div>

          <!-- Performance (RPM/TPM) -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-violet-100 p-2 dark:bg-violet-900/30">
                <Icon name="bolt" size="md" class="text-violet-600 dark:text-violet-400" :stroke-width="2" />
              </div>
              <div class="flex-1">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.performance') }}
                </p>
                <div class="flex items-baseline gap-2">
                  <p class="text-xl font-bold text-gray-900 dark:text-white">
                    {{ formatTokens(stats.rpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
                </div>
                <div class="flex items-baseline gap-2">
                  <p class="text-sm font-semibold text-violet-600 dark:text-violet-400">
                    {{ formatTokens(stats.tpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Token Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <AccountUsageOverviewCard />

          <!-- Today Tokens -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
                <Icon name="cube" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.today_tokens) }}
                </p>
                <p class="text-xs">
                  <span
                    class="text-green-600 dark:text-green-400"
                    :title="t('admin.dashboard.actual')"
                    >${{ formatCost(stats.today_actual_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-orange-500 dark:text-orange-400"
                    :title="t('admin.dashboard.accountCost')"
                    >${{ formatCost(stats.today_account_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-gray-400 dark:text-gray-500"
                    :title="t('admin.dashboard.standard')"
                    >${{ formatCost(stats.today_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Total Tokens -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-indigo-100 p-2 dark:bg-indigo-900/30">
                <Icon name="database" size="md" class="text-indigo-600 dark:text-indigo-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.totalTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.total_tokens) }}
                </p>
                <p class="text-xs">
                  <span
                    class="text-green-600 dark:text-green-400"
                    :title="t('admin.dashboard.actual')"
                    >${{ formatCost(stats.total_actual_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-orange-500 dark:text-orange-400"
                    :title="t('admin.dashboard.accountCost')"
                    >${{ formatCost(stats.total_account_cost) }}</span
                  >
                  <span class="text-gray-400 dark:text-gray-500"> / </span>
                  <span
                    class="text-gray-400 dark:text-gray-500"
                    :title="t('admin.dashboard.standard')"
                    >${{ formatCost(stats.total_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Avg Response Time -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
                <Icon name="clock" size="md" class="text-rose-600 dark:text-rose-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.avgResponse') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatDuration(stats.average_duration_ms) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ stats.active_users }} {{ t('admin.dashboard.activeUsers') }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="space-y-6">
          <!-- Date Range Filter -->
          <div class="card p-4">
            <div class="flex flex-wrap items-center gap-4">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('admin.dashboard.timeRange') }}:</span
                >
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>
              <button @click="loadDashboardStats" :disabled="chartsLoading" class="btn btn-secondary">
                {{ t('common.refresh') }}
              </button>
              <div class="ml-auto flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('admin.dashboard.granularity') }}:</span
                >
                <div class="w-28">
                  <Select
                    v-model="granularity"
                    :options="granularityOptions"
                    @change="loadChartData"
                  />
                </div>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
            <ModelDistributionChart
              class="min-w-0 lg:col-span-2"
              :model-stats="modelStats"
              :enable-ranking-view="true"
              :ranking-items="rankingItems"
              :ranking-total-cost="rankingTotalCost"
              :ranking-total-requests="rankingTotalRequests"
              :ranking-total-tokens="rankingTotalTokens"
              :api-key-ranking-items="apiKeyRankingItems"
              :api-key-ranking-total-actual-cost="apiKeyRankingTotalActualCost"
              :api-key-ranking-total-requests="apiKeyRankingTotalRequests"
              :api-key-ranking-total-tokens="apiKeyRankingTotalTokens"
              :loading="chartsLoading"
              :ranking-loading="rankingLoading"
              :ranking-error="rankingError"
              :api-key-ranking-loading="apiKeyRankingLoading"
              :api-key-ranking-error="apiKeyRankingError"
              default-ranking-view="api_key_spending_ranking"
              :start-date="startDate"
              :end-date="endDate"
              @ranking-click="goToAccountUsage"
            />

            <div class="card min-w-0 p-4" data-testid="active-key-concurrency-card">
              <div class="flex items-center justify-between gap-3">
                <div>
                  <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.dashboard.activeKeyConcurrency') }}</h3>
                  <p class="mt-1 text-xs text-gray-400 dark:text-gray-500">{{ t('admin.dashboard.activeKeyConcurrencyHint') }}</p>
                </div>
                <span class="h-2.5 w-2.5 shrink-0 rounded-full" :class="totalKeyConcurrency > 0 ? 'bg-emerald-500' : 'bg-gray-300 dark:bg-dark-600'" />
              </div>
              <div class="mt-4 flex items-end gap-2 border-b border-gray-100 pb-3 dark:border-dark-700">
                <span class="text-3xl font-bold tabular-nums text-gray-900 dark:text-white">{{ totalKeyConcurrency }}</span>
                <span class="pb-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.totalConcurrency') }}</span>
              </div>
              <div v-if="keyConcurrencyLoading && activeConcurrencyKeys.length === 0" class="flex h-28 items-center justify-center">
                <LoadingSpinner size="sm" />
              </div>
              <div v-else-if="activeConcurrencyKeys.length === 0" class="flex h-28 items-center justify-center text-xs text-gray-400">
                {{ t('admin.dashboard.noActiveKeyConcurrency') }}
              </div>
              <div v-else class="mt-2 max-h-32 space-y-1.5 overflow-auto pr-1">
                <div v-for="key in activeConcurrencyKeys" :key="key.id" class="flex min-w-0 items-center gap-2 rounded-md bg-gray-50 px-2.5 py-1.5 dark:bg-dark-800/60">
                  <span class="min-w-0 flex-1 truncate text-xs font-medium text-gray-700 dark:text-gray-300" :title="key.name || `Key #${key.id}`">{{ key.name || `Key #${key.id}` }}</span>
                  <span class="shrink-0 rounded bg-emerald-100 px-2 py-0.5 text-xs font-semibold tabular-nums text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400">{{ key.current_concurrency }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- API Key Usage Trend (Full Width) -->
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dashboard.apiKeyUsageTrend') }} (Top 12)
            </h3>
            <div class="h-64">
              <div v-if="apiKeyTrendLoading" class="flex h-full items-center justify-center">
                <LoadingSpinner size="md" />
              </div>
              <Line v-else-if="apiKeyTrendChartData" :data="apiKeyTrendChartData" :options="lineOptions" />
              <div
                v-else
                class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.dashboard.noDataAvailable') }}
              </div>
            </div>
          </div>

          <!-- Token Usage Trend (Full Width) -->
          <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type {
  DashboardStats,
  TrendDataPoint,
  ModelStat,
  ApiKeyUsageTrendPoint,
  AccountSpendingRankingItem,
  ApiKeySpendingRankingItem
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'
import AccountUsageOverviewCard from '@/components/admin/dashboard/AccountUsageOverviewCard.vue'
import type { SimpleApiKey } from '@/api/admin/usage'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const router = useRouter()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const apiKeyTrendLoading = ref(false)
const rankingLoading = ref(false)
const rankingError = ref(false)
const apiKeyRankingLoading = ref(false)
const apiKeyRankingError = ref(false)

// Chart data
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const apiKeyTrend = ref<ApiKeyUsageTrendPoint[]>([])
const rankingItems = ref<AccountSpendingRankingItem[]>([])
const rankingTotalCost = ref(0)
const rankingTotalRequests = ref(0)
const rankingTotalTokens = ref(0)
const apiKeyRankingItems = ref<ApiKeySpendingRankingItem[]>([])
const apiKeyRankingTotalActualCost = ref(0)
const apiKeyRankingTotalRequests = ref(0)
const apiKeyRankingTotalTokens = ref(0)
const concurrencyKeys = ref<SimpleApiKey[]>([])
const keyConcurrencyLoading = ref(false)
let concurrencyTimer: ReturnType<typeof setInterval> | null = null
let chartLoadSeq = 0
let apiKeyTrendLoadSeq = 0
let rankingLoadSeq = 0
let apiKeyRankingLoadSeq = 0
const rankingLimit = 12
const activeConcurrencyKeys = computed(() => concurrencyKeys.value
  .filter(key => key.current_concurrency > 0)
  .sort((left, right) => right.current_concurrency - left.current_concurrency))
const totalKeyConcurrency = computed(() => activeConcurrencyKeys.value.reduce(
  (total, key) => total + key.current_concurrency,
  0
))

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const getTodayRangeDates = (): { start: string; end: string } => {
  const today = formatLocalDate(new Date())
  return { start: today, end: today }
}

// Date range
const granularity = ref<'day' | 'hour'>('hour')
const defaultRange = getTodayRangeDates()
const startDate = ref(defaultRange.start)
const endDate = ref(defaultRange.end)

// Granularity options for Select component
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

// Dark mode detection
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

// Chart colors
const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb'
}))

// Line chart options (for API key trend chart)
const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => {
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    }
  }
}))

// API key trend chart data
const apiKeyTrendChartData = computed(() => {
  if (!apiKeyTrend.value?.length) return null

  const getDisplayName = (point: ApiKeyUsageTrendPoint): string => {
    const keyName = point.key_name?.trim()
    if (keyName) return keyName
    return `Key #${point.api_key_id}`
  }

  // Group by api_key_id to avoid merging different keys with the same display name
  const apiKeyGroups = new Map<number, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  apiKeyTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = point.api_key_id
    if (!apiKeyGroups.has(key)) {
      apiKeyGroups.set(key, { name: getDisplayName(point), data: new Map() })
    }
    apiKeyGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const colors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#ec4899',
    '#14b8a6',
    '#f97316',
    '#6366f1',
    '#84cc16',
    '#06b6d4',
    '#a855f7'
  ]

  const datasets = Array.from(apiKeyGroups.values()).map((group, idx) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: colors[idx % colors.length],
    backgroundColor: `${colors[idx % colors.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return {
    labels: sortedDates,
    datasets
  }
})

// Format helpers
const formatTokens = (value: number | undefined): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const toFiniteNumber = (value: unknown): number => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
}

const formatNumber = (value: number | null | undefined): string => {
  return toFiniteNumber(value).toLocaleString()
}

const formatCost = (value: number | null | undefined): string => {
  const safeValue = toFiniteNumber(value)
  if (safeValue >= 1000) {
    return (safeValue / 1000).toFixed(2) + 'K'
  } else if (safeValue >= 1) {
    return safeValue.toFixed(2)
  } else if (safeValue >= 0.01) {
    return safeValue.toFixed(3)
  }
  return safeValue.toFixed(4)
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const goToAccountUsage = (item: AccountSpendingRankingItem) => {
  void router.push({
    path: '/admin/usage',
    query: {
      account_id: String(item.account_id),
      start_date: startDate.value,
      end_date: endDate.value
    }
  })
}

// Date range change handler
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  // Auto-select granularity based on date range
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))

  // If range is 1 day, use hourly granularity
  if (daysDiff <= 1) {
    granularity.value = 'hour'
  } else {
    granularity.value = 'day'
  }

  loadChartData()
}

// Load data
const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadAPIKeyTrend = async () => {
  const currentSeq = ++apiKeyTrendLoadSeq
  apiKeyTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getApiKeyUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: rankingLimit
    })
    if (currentSeq !== apiKeyTrendLoadSeq) return
    apiKeyTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== apiKeyTrendLoadSeq) return
    console.error('Error loading API key trend:', error)
    apiKeyTrend.value = []
  } finally {
    if (currentSeq === apiKeyTrendLoadSeq) {
      apiKeyTrendLoading.value = false
    }
  }
}

const loadAccountSpendingRanking = async () => {
  const currentSeq = ++rankingLoadSeq
  rankingLoading.value = true
  rankingError.value = false
  try {
    const response = await adminAPI.dashboard.getAccountSpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== rankingLoadSeq) return
    rankingItems.value = response.ranking || []
    rankingTotalCost.value = response.total_account_cost || 0
    rankingTotalRequests.value = response.total_requests || 0
    rankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== rankingLoadSeq) return
    console.error('Error loading account spending ranking:', error)
    rankingItems.value = []
    rankingTotalCost.value = 0
    rankingTotalRequests.value = 0
    rankingTotalTokens.value = 0
    rankingError.value = true
  } finally {
    if (currentSeq === rankingLoadSeq) {
      rankingLoading.value = false
    }
  }
}

const loadAPIKeySpendingRanking = async () => {
  const currentSeq = ++apiKeyRankingLoadSeq
  apiKeyRankingLoading.value = true
  apiKeyRankingError.value = false
  try {
    const response = await adminAPI.dashboard.getApiKeySpendingRanking({
      start_date: startDate.value,
      end_date: endDate.value,
      limit: rankingLimit
    })
    if (currentSeq !== apiKeyRankingLoadSeq) return
    apiKeyRankingItems.value = response.ranking || []
    apiKeyRankingTotalActualCost.value = response.total_actual_cost || 0
    apiKeyRankingTotalRequests.value = response.total_requests || 0
    apiKeyRankingTotalTokens.value = response.total_tokens || 0
  } catch (error) {
    if (currentSeq !== apiKeyRankingLoadSeq) return
    console.error('Error loading API key spending ranking:', error)
    apiKeyRankingItems.value = []
    apiKeyRankingTotalActualCost.value = 0
    apiKeyRankingTotalRequests.value = 0
    apiKeyRankingTotalTokens.value = 0
    apiKeyRankingError.value = true
  } finally {
    if (currentSeq === apiKeyRankingLoadSeq) {
      apiKeyRankingLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await Promise.all([
    loadDashboardSnapshot(true),
    loadAPIKeyTrend(),
    loadAccountSpendingRanking(),
    loadAPIKeySpendingRanking()
  ])
}

const loadKeyConcurrency = async () => {
  keyConcurrencyLoading.value = true
  try {
    concurrencyKeys.value = await adminAPI.usage.searchApiKeys(undefined, undefined, 100)
  } catch (error) {
    console.error('Error loading API key concurrency:', error)
  } finally {
    keyConcurrencyLoading.value = false
  }
}

const loadChartData = async () => {
  await Promise.all([
    loadDashboardSnapshot(false),
    loadAPIKeyTrend(),
    loadAccountSpendingRanking(),
    loadAPIKeySpendingRanking()
  ])
}

onMounted(() => {
  loadDashboardStats()
  void loadKeyConcurrency()
  concurrencyTimer = setInterval(loadKeyConcurrency, 5000)
})

onUnmounted(() => {
  if (concurrencyTimer) clearInterval(concurrencyTimer)
})
</script>

<style scoped>
</style>
