<template>
  <div class="card p-4">
    <div class="mb-4 flex flex-col items-stretch gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">
        {{ !enableRankingView || activeView === 'model_distribution'
          ? t('admin.dashboard.modelDistribution')
          : activeView === 'spending_ranking'
            ? t('admin.dashboard.spendingRankingTitle')
            : t('admin.dashboard.apiKeySpendingRankingTitle') }}
      </h3>
      <div class="flex flex-wrap items-center gap-2 sm:justify-end">
        <div
          v-if="showSourceToggle"
          class="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800"
        >
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="source === 'requested'
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="emit('update:source', 'requested')"
          >
            {{ t('usage.requestedModel') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="source === 'upstream'
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="emit('update:source', 'upstream')"
          >
            {{ t('usage.upstreamModel') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="source === 'mapping'
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="emit('update:source', 'mapping')"
          >
            {{ t('usage.mapping') }}
          </button>
        </div>
        <div
          v-if="showMetricToggle"
          class="inline-flex rounded-lg border border-gray-200 bg-gray-50 p-0.5 dark:border-dark-700 dark:bg-dark-800"
        >
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="metric === 'tokens'
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="emit('update:metric', 'tokens')"
          >
            {{ t('admin.dashboard.metricTokens') }}
          </button>
          <button
            type="button"
            class="rounded-md px-2.5 py-1 text-xs font-medium transition-colors"
            :class="metric === 'actual_cost'
              ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
              : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="emit('update:metric', 'actual_cost')"
          >
            {{ t('admin.dashboard.metricActualCost') }}
          </button>
        </div>
        <div v-if="enableRankingView" class="grid w-full grid-cols-3 rounded-lg bg-gray-100 p-1 dark:bg-dark-800 sm:inline-flex sm:w-auto">
          <button
            type="button"
            class="rounded-md px-1.5 py-1 text-[11px] font-medium transition-colors sm:px-2.5 sm:text-xs"
            :class="
              activeView === 'model_distribution'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
            "
            @click="activeView = 'model_distribution'"
          >
            {{ t('admin.dashboard.viewModelDistribution') }}
          </button>
          <button
            type="button"
            class="rounded-md px-1.5 py-1 text-[11px] font-medium transition-colors sm:px-2.5 sm:text-xs"
            :class="
              activeView === 'spending_ranking'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
            "
            @click="activeView = 'spending_ranking'"
          >
            {{ t('admin.dashboard.viewSpendingRanking') }}
          </button>
          <button
            type="button"
            class="rounded-md px-1.5 py-1 text-[11px] font-medium transition-colors sm:px-2.5 sm:text-xs"
            :class="
              activeView === 'api_key_spending_ranking'
                ? 'bg-white text-gray-900 shadow-sm dark:bg-dark-700 dark:text-white'
                : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'
            "
            @click="activeView = 'api_key_spending_ranking'"
          >
            {{ t('admin.dashboard.viewApiKeySpendingRanking') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="activeView === 'model_distribution' && loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="activeView === 'model_distribution' && displayModelStats.length > 0 && chartData"
      class="flex flex-col items-center gap-4 sm:flex-row sm:gap-6"
    >
      <div class="h-48 w-48 shrink-0">
        <Doughnut :data="chartData" :options="doughnutOptions" />
      </div>
      <div class="max-h-48 w-full min-w-0 flex-1 overflow-auto">
        <table class="w-full text-xs">
          <thead>
            <tr class="text-gray-500 dark:text-gray-400">
              <th class="pb-2 text-left">{{ t('admin.dashboard.model') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.requests') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.tokens') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.actual') }}</th>
              <th v-if="showAccountCost" class="pb-2 text-right">{{ t('admin.dashboard.accountCost') }}</th>
              <th class="pb-2 text-right">{{ t('admin.dashboard.standard') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="model in displayModelStats" :key="model.model">
              <tr
                class="border-t border-gray-100 transition-colors dark:border-dark-700"
                :class="enableBreakdown ? 'cursor-pointer hover:bg-gray-50 dark:hover:bg-dark-700/40' : ''"
                @click="enableBreakdown && toggleBreakdown('model', model.model)"
              >
                <td
                  class="max-w-[100px] truncate py-1.5 font-medium"
                  :class="enableBreakdown ? 'text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300' : 'text-gray-900 dark:text-white'"
                  :title="model.model"
                >
                  <span class="inline-flex items-center gap-1">
                    <svg v-if="enableBreakdown && expandedKey === `model-${model.model}`" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
                    <svg v-else-if="enableBreakdown" class="h-3 w-3 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
                    {{ model.model }}
                  </span>
                </td>
                <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                  {{ formatNumber(model.requests) }}
                </td>
                <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                  {{ formatTokens(model.total_tokens) }}
                </td>
                <td class="py-1.5 text-right text-green-600 dark:text-green-400">
                  ${{ formatCost(model.actual_cost) }}
                </td>
                <td v-if="showAccountCost" class="py-1.5 text-right text-orange-500 dark:text-orange-400">
                  ${{ formatCost(model.account_cost) }}
                </td>
                <td class="py-1.5 text-right text-gray-400 dark:text-gray-500">
                  ${{ formatCost(model.cost) }}
                </td>
              </tr>
              <tr v-if="expandedKey === `model-${model.model}`">
                <td :colspan="distributionColspan" class="p-0">
                  <UserBreakdownSubTable
                    :items="breakdownItems"
                    :loading="breakdownLoading"
                    :show-account-cost="showAccountCost"
                  />
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>
    <div
      v-else-if="activeView === 'model_distribution'"
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>

    <div v-else-if="activeRankingState.loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div
      v-else-if="activeRankingState.error"
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.failedToLoad') }}
    </div>
    <div
      v-else-if="activeRankingDisplayItems.length > 0 && rankingChartData"
      class="flex flex-col items-center gap-4 sm:flex-row sm:gap-6"
    >
      <div class="contents">
        <div class="h-48 w-48 shrink-0">
          <Doughnut :data="rankingChartData" :options="rankingDoughnutOptions" />
        </div>
        <div class="max-h-48 w-full min-w-0 flex-1 overflow-auto">
          <table class="w-full table-fixed text-[11px] sm:text-xs" :class="showApiKeyExtendedColumns ? 'lg:min-w-[640px]' : ''">
          <thead>
            <tr class="text-gray-500 dark:text-gray-400">
              <th :class="showApiKeyExtendedColumns ? 'w-[28%]' : isApiKeyRankingView ? 'w-[32%]' : 'w-[40%]'" class="pb-2 text-left">{{ activeRankingNameHeader }}</th>
              <th :class="showApiKeyExtendedColumns ? 'w-[14%]' : isApiKeyRankingView ? 'w-[16%]' : 'w-[20%]'" class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th v-if="isApiKeyRankingView" :class="showApiKeyExtendedColumns ? 'w-[16%]' : 'w-[20%]'" class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingAverageDuration') }}</th>
              <th :class="showApiKeyExtendedColumns ? 'w-[14%]' : isApiKeyRankingView ? 'w-[16%]' : 'w-[20%]'" class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingTokens') }}</th>
              <th :class="showApiKeyExtendedColumns ? 'w-[14%]' : isApiKeyRankingView ? 'w-[16%]' : 'w-[20%]'" class="pb-2 text-right">{{ t('admin.dashboard.spendingRankingSpend') }}</th>
              <th v-if="showApiKeyExtendedColumns" class="w-[14%] pb-2 text-right">{{ t('admin.dashboard.spendingRankingShare') }}</th>
            </tr>
          </thead>
          <tbody>
            <template v-for="(item, index) in activeRankingDisplayItems" :key="getRankingRowKey(item, index)">
              <tr
                class="border-t border-gray-100 transition-colors dark:border-dark-700"
                :class="item.isOther
                  ? 'bg-gray-50/70 dark:bg-dark-700/20'
                  : 'cursor-pointer hover:bg-gray-50 dark:hover:bg-dark-700/40'"
                @click="handleRankingClick(item)"
              >
                <td class="py-1.5">
                  <div class="flex min-w-0 items-center gap-2">
                    <span class="shrink-0 text-[11px] font-semibold text-gray-500 dark:text-gray-400">
                      {{ item.isOther ? 'Σ' : `#${index + 1}` }}
                    </span>
                    <button
                      v-if="isApiKeyRankingItem(item) && !item.isOther"
                      type="button"
                      class="flex min-w-0 items-center gap-1 text-left font-medium text-blue-600 hover:text-blue-800 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-blue-500 dark:text-blue-400 dark:hover:text-blue-300"
                      :aria-expanded="expandedApiKeyID === item.api_key_id"
                      :aria-controls="`api-key-models-${item.api_key_id}`"
                      @click.stop="selectApiKeyModels(item)"
                    >
                      <Icon
                        :name="expandedApiKeyID === item.api_key_id ? 'chevronDown' : 'chevronRight'"
                        size="xs"
                        class="shrink-0 text-gray-400"
                      />
                      <span class="block max-w-[180px] truncate" :title="getRankingRowLabel(item)">
                        {{ getRankingRowLabel(item) }}
                      </span>
                      <span v-if="item.distinct_ip_count" class="shrink-0 text-[10px] text-gray-400 dark:text-gray-500">
                        {{ t('admin.dashboard.ipCountShort', { count: item.distinct_ip_count }) }}
                      </span>
                    </button>
                    <span
                      v-else
                      class="block max-w-[180px] truncate font-medium text-gray-900 dark:text-white"
                      :title="getRankingRowLabel(item)"
                    >
                      {{ getRankingRowLabel(item) }}
                    </span>
                  </div>
                </td>
                <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                  {{ formatNumber(item.requests) }}
                </td>
                <td v-if="isApiKeyRankingView" class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                  {{ formatAverageDuration(getRankingAverageDuration(item)) }}
                </td>
                <td class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                  {{ formatTokens(item.tokens) }}
                </td>
                <td class="py-1.5 text-right text-green-600 dark:text-green-400">
                  ${{ formatCost(getRankingCost(item)) }}
                </td>
                <td v-if="showApiKeyExtendedColumns" class="py-1.5 text-right text-gray-600 dark:text-gray-400">
                  {{ formatRankingShare(getRankingCost(item)) }}
                </td>
              </tr>
              <tr
                v-if="isApiKeyRankingItem(item) && expandedApiKeyID === item.api_key_id"
                :id="`api-key-models-${item.api_key_id}`"
              >
                <td :colspan="apiKeyRankingColspan" class="p-0">
                  <div class="bg-gray-50/70 px-2 py-3 dark:bg-dark-700/30 sm:px-6">
                    <div v-if="apiKeyModelsLoading" class="flex items-center justify-center py-3">
                      <LoadingSpinner />
                    </div>
                    <div v-else-if="apiKeyModelsError" class="py-2 text-center text-xs text-gray-400">
                      {{ t('admin.dashboard.failedToLoad') }}
                    </div>
                    <div v-else-if="apiKeyModelStats.length === 0" class="py-2 text-center text-xs text-gray-400">
                      {{ t('admin.dashboard.noDataAvailable') }}
                    </div>
                    <table v-else class="w-full table-fixed text-[10px] sm:text-xs">
                      <thead>
                        <tr class="text-gray-400 dark:text-gray-500">
                          <th class="w-[28%] pb-1.5 text-left">{{ t('admin.dashboard.model') }}</th>
                          <th class="w-[14%] pb-1.5 text-right">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
                          <th class="w-[18%] pb-1.5 text-right">{{ t('admin.dashboard.spendingRankingAverageDuration') }}</th>
                          <th class="w-[14%] pb-1.5 text-right">{{ t('admin.dashboard.spendingRankingTokens') }}</th>
                          <th class="w-[14%] pb-1.5 text-right">{{ t('admin.dashboard.spendingRankingSpend') }}</th>
                          <th class="w-[12%] pb-1.5 text-right">{{ t('admin.dashboard.spendingRankingShare') }}</th>
                        </tr>
                      </thead>
                      <tbody>
                        <tr v-for="model in apiKeyModelStats" :key="model.model" class="border-t border-gray-200/70 dark:border-dark-600/70">
                          <td class="truncate py-1.5 font-medium text-gray-700 dark:text-gray-200" :title="model.model">{{ model.model }}</td>
                          <td class="py-1.5 text-right text-gray-500 dark:text-gray-400">{{ formatNumber(model.requests) }}</td>
                          <td class="py-1.5 text-right text-gray-500 dark:text-gray-400">{{ formatAverageDuration(model.average_duration_ms) }}</td>
                          <td class="py-1.5 text-right text-gray-500 dark:text-gray-400">{{ formatTokens(model.total_tokens) }}</td>
                          <td class="py-1.5 text-right text-green-600 dark:text-green-400">${{ formatCost(model.actual_cost) }}</td>
                          <td class="py-1.5 text-right text-gray-500 dark:text-gray-400">{{ formatApiKeyModelShare(model.actual_cost) }}</td>
                        </tr>
                      </tbody>
                    </table>
                    <div v-if="item.ip_usages?.length" class="mt-3 border-t border-gray-200 pt-2 dark:border-dark-600">
                      <div class="mb-1 flex items-center justify-between text-[10px] text-gray-400 dark:text-gray-500 sm:text-xs">
                        <span>{{ t('admin.dashboard.ipLocation') }}</span>
                        <span>{{ t('admin.dashboard.ipTopOnly') }}</span>
                      </div>
                      <table class="w-full table-fixed text-[10px] sm:text-xs">
                        <thead>
                          <tr class="text-gray-400 dark:text-gray-500">
                            <th class="w-[48%] pb-1 text-left">IP</th>
                            <th class="w-[20%] pb-1 text-right">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
                            <th class="w-[32%] pb-1 text-right">{{ t('admin.dashboard.ipLastSeen') }}</th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr v-for="ipUsage in item.ip_usages" :key="ipUsage.ip_address" class="border-t border-gray-200/70 dark:border-dark-600/70">
                            <td class="truncate py-1 text-gray-600 dark:text-gray-300" :title="ipUsage.ip_address">{{ ipUsage.ip_address }}</td>
                            <td class="py-1 text-right text-gray-500 dark:text-gray-400">{{ formatNumber(ipUsage.requests) }}</td>
                            <td class="truncate py-1 text-right text-gray-500 dark:text-gray-400" :title="formatIPLastSeen(ipUsage.last_seen_at)">{{ formatIPLastSeen(ipUsage.last_seen_at) }}</td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                </td>
              </tr>
            </template>
          </tbody>
          </table>
        </div>
      </div>
    </div>
    <div
      v-else
      class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400"
    >
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import UserBreakdownSubTable from './UserBreakdownSubTable.vue'
import type {
  AccountSpendingRankingItem,
  ApiKeySpendingRankingItem,
  ModelStat,
  UserBreakdownItem
} from '@/types'
import { getModelStats, getUserBreakdown } from '@/api/admin/dashboard'

ChartJS.register(ArcElement, Tooltip, Legend)

const { t } = useI18n()

type DistributionMetric = 'tokens' | 'actual_cost'
type ModelSource = 'requested' | 'upstream' | 'mapping'
type RankingView = 'model_distribution' | 'spending_ranking' | 'api_key_spending_ranking'
type AccountRankingDisplayItem = AccountSpendingRankingItem & { isOther?: boolean }
type ApiKeyRankingDisplayItem = ApiKeySpendingRankingItem & { isOther?: boolean }
type RankingDisplayItem = AccountRankingDisplayItem | ApiKeyRankingDisplayItem
const props = withDefaults(defineProps<{
  modelStats: ModelStat[]
  upstreamModelStats?: ModelStat[]
  mappingModelStats?: ModelStat[]
  source?: ModelSource
  enableRankingView?: boolean
  rankingItems?: AccountSpendingRankingItem[]
  rankingTotalCost?: number
  rankingTotalRequests?: number
  rankingTotalTokens?: number
  apiKeyRankingItems?: ApiKeySpendingRankingItem[]
  apiKeyRankingTotalActualCost?: number
  apiKeyRankingTotalRequests?: number
  apiKeyRankingTotalTokens?: number
  loading?: boolean
  metric?: DistributionMetric
  showSourceToggle?: boolean
  showMetricToggle?: boolean
  enableBreakdown?: boolean
  showAccountCost?: boolean
  rankingLoading?: boolean
  rankingError?: boolean
  apiKeyRankingLoading?: boolean
  apiKeyRankingError?: boolean
  defaultRankingView?: RankingView
  wideRankingLayout?: boolean
  startDate?: string
  endDate?: string
  filters?: Record<string, any>
}>(), {
  upstreamModelStats: () => [],
  mappingModelStats: () => [],
  source: 'requested',
  enableRankingView: false,
  rankingItems: () => [],
  rankingTotalCost: 0,
  rankingTotalRequests: 0,
  rankingTotalTokens: 0,
  apiKeyRankingItems: () => [],
  apiKeyRankingTotalActualCost: 0,
  apiKeyRankingTotalRequests: 0,
  apiKeyRankingTotalTokens: 0,
  loading: false,
  metric: 'tokens',
  showSourceToggle: false,
  showMetricToggle: false,
  enableBreakdown: true,
  showAccountCost: true,
  rankingLoading: false,
  rankingError: false,
  apiKeyRankingLoading: false,
  apiKeyRankingError: false,
  defaultRankingView: 'model_distribution',
  wideRankingLayout: false
})

const expandedKey = ref<string | null>(null)
const breakdownItems = ref<UserBreakdownItem[]>([])
const breakdownLoading = ref(false)
const expandedApiKeyID = ref<number | null>(null)
const apiKeyModelStats = ref<ModelStat[]>([])
const apiKeyModelsLoading = ref(false)
const apiKeyModelsError = ref(false)
let apiKeyModelsLoadSeq = 0

const toggleBreakdown = async (type: string, id: string) => {
  const key = `${type}-${id}`
  if (expandedKey.value === key) {
    expandedKey.value = null
    return
  }
  expandedKey.value = key
  breakdownLoading.value = true
  breakdownItems.value = []
  try {
    const res = await getUserBreakdown({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      model: id,
      model_source: props.source,
    })
    breakdownItems.value = res.users || []
  } catch {
    breakdownItems.value = []
  } finally {
    breakdownLoading.value = false
  }
}

const emit = defineEmits<{
  'update:metric': [value: DistributionMetric]
  'update:source': [value: ModelSource]
  'ranking-click': [item: AccountSpendingRankingItem]
}>()

const enableRankingView = computed(() => props.enableRankingView)
const activeView = ref<RankingView>(props.enableRankingView ? props.defaultRankingView : 'model_distribution')
const showAccountCost = computed(() => props.showAccountCost)
const distributionColspan = computed(() => showAccountCost.value ? 6 : 5)
const isApiKeyRankingView = computed(() => activeView.value === 'api_key_spending_ranking')
const showApiKeyExtendedColumns = computed(() => isApiKeyRankingView.value && props.wideRankingLayout)
const apiKeyRankingColspan = computed(() => showApiKeyExtendedColumns.value ? 6 : 5)

const chartColors = [
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

const displayModelStats = computed(() => {
  const sourceStats = props.source === 'upstream'
    ? props.upstreamModelStats
    : props.source === 'mapping'
      ? props.mappingModelStats
      : props.modelStats
  if (!sourceStats?.length) return []

  const metricKey = props.metric === 'actual_cost' ? 'actual_cost' : 'total_tokens'
  return [...sourceStats].sort((a, b) => toFiniteNumber(b[metricKey]) - toFiniteNumber(a[metricKey]))
})

const chartData = computed(() => {
  if (!displayModelStats.value.length) return null

  return {
    labels: displayModelStats.value.map((m) => m.model),
    datasets: [
      {
        data: displayModelStats.value.map((m) => toFiniteNumber(props.metric === 'actual_cost' ? m.actual_cost : m.total_tokens)),
        backgroundColor: chartColors.slice(0, displayModelStats.value.length),
        borderWidth: 0
      }
    ]
  }
})

const rankingChartData = computed(() => {
  const items = activeRankingItems.value
  if (!items.length) return null

  const labels = items.map((item, index) => `#${index + 1} ${getRankingEntityLabel(item)}`)
  const data = items.map(getRankingCost)
  const backgroundColor = chartColors.slice(0, items.length)

  if (otherRankingItem.value) {
    labels.push(t('admin.dashboard.spendingRankingOther'))
    data.push(getRankingCost(otherRankingItem.value))
    backgroundColor.push('#94a3b8')
  }

  return {
    labels,
    datasets: [
      {
        data,
        backgroundColor,
        borderWidth: 0
      }
    ]
  }
})

const activeRankingItems = computed<RankingDisplayItem[]>(() => activeView.value === 'api_key_spending_ranking'
  ? props.apiKeyRankingItems
  : props.rankingItems)

const activeRankingTotals = computed(() => activeView.value === 'api_key_spending_ranking'
  ? {
      totalCost: props.apiKeyRankingTotalActualCost,
      totalRequests: props.apiKeyRankingTotalRequests,
      totalTokens: props.apiKeyRankingTotalTokens
    }
  : {
      totalCost: props.rankingTotalCost,
      totalRequests: props.rankingTotalRequests,
      totalTokens: props.rankingTotalTokens
    })

const activeRankingState = computed(() => activeView.value === 'api_key_spending_ranking'
  ? { loading: props.apiKeyRankingLoading, error: props.apiKeyRankingError }
  : { loading: props.rankingLoading, error: props.rankingError })

const activeRankingNameHeader = computed(() => activeView.value === 'api_key_spending_ranking'
  ? t('admin.dashboard.spendingRankingApiKey')
  : t('admin.dashboard.spendingRankingAccount'))

const apiKeyModelTotalActualCost = computed(() => apiKeyModelStats.value.reduce(
  (sum, model) => sum + toFiniteNumber(model.actual_cost),
  0
))

const otherRankingItem = computed<RankingDisplayItem | null>(() => {
  const items = activeRankingItems.value
  if (!items.length) return null

  const rankedCost = items.reduce((sum, item) => sum + getRankingCost(item), 0)
  const rankedRequests = items.reduce((sum, item) => sum + toFiniteNumber(item.requests), 0)
  const rankedTokens = items.reduce((sum, item) => sum + toFiniteNumber(item.tokens), 0)

  const otherCost = Math.max((activeRankingTotals.value.totalCost || 0) - rankedCost, 0)
  const otherRequests = Math.max((activeRankingTotals.value.totalRequests || 0) - rankedRequests, 0)
  const otherTokens = Math.max((activeRankingTotals.value.totalTokens || 0) - rankedTokens, 0)

  if (otherCost <= 0.000001 && otherRequests <= 0 && otherTokens <= 0) return null

  if (activeView.value === 'api_key_spending_ranking') {
    return {
      api_key_id: 0,
      key_name: '',
      user_id: 0,
      email: '',
      actual_cost: otherCost,
      requests: otherRequests,
      tokens: otherTokens,
      isOther: true
    }
  }

  return {
    account_id: 0,
    account_name: '',
    platform: '',
    account_cost: otherCost,
    requests: otherRequests,
    tokens: otherTokens,
    isOther: true
  }
})

const activeRankingDisplayItems = computed<RankingDisplayItem[]>(() => {
  const items = activeRankingItems.value
  if (!items.length) return []
  return otherRankingItem.value
    ? [...items, otherRankingItem.value]
    : [...items]
})

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          const formattedValue = props.metric === 'actual_cost'
            ? `$${formatCost(value)}`
            : formatTokens(value)
          return `${context.label}: ${formattedValue} (${percentage}%)`
        }
      }
    }
  }
}))

const rankingDoughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      callbacks: {
        label: (context: any) => {
          const value = context.raw as number
          const total = context.dataset.data.reduce((a: number, b: number) => a + b, 0)
          const percentage = total > 0 ? ((value / total) * 100).toFixed(1) : '0.0'
          return `${context.label}: $${formatCost(value)} (${percentage}%)`
        }
      }
    }
  }
}))

const formatTokens = (value: number): string => {
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatNumber = (value: number): string => {
  return toFiniteNumber(value).toLocaleString()
}

const isApiKeyRankingItem = (item: RankingDisplayItem): item is ApiKeyRankingDisplayItem => {
  return 'api_key_id' in item
}

function getRankingCost(item: RankingDisplayItem): number {
  return toFiniteNumber(isApiKeyRankingItem(item) ? item.actual_cost : item.account_cost)
}

const getRankingAccountLabel = (item: AccountSpendingRankingItem): string => {
  const name = item.account_name?.trim() || t('admin.dashboard.spendingRankingAccountFallback', { id: item.account_id })
  const platform = item.platform?.trim()
  return platform ? `${name} (${platform})` : name
}

const getRankingApiKeyLabel = (item: ApiKeySpendingRankingItem): string => {
  if (item.key_name) return item.key_name
  return `Key #${item.api_key_id}`
}

const getRankingEntityLabel = (item: RankingDisplayItem): string => {
  return isApiKeyRankingItem(item) ? getRankingApiKeyLabel(item) : getRankingAccountLabel(item)
}

const getRankingRowLabel = (item: RankingDisplayItem): string => {
  if (item.isOther) return t('admin.dashboard.spendingRankingOther')
  return getRankingEntityLabel(item)
}

const getRankingRowKey = (item: RankingDisplayItem, index: number): string => {
  if (item.isOther) return `others-${activeView.value}`
  return isApiKeyRankingItem(item) ? `api-key-${item.api_key_id}-${index}` : `account-${item.account_id}-${index}`
}

const selectApiKeyModels = async (item: ApiKeySpendingRankingItem) => {
  if (expandedApiKeyID.value === item.api_key_id) {
    expandedApiKeyID.value = null
    apiKeyModelStats.value = []
    apiKeyModelsLoading.value = false
    apiKeyModelsError.value = false
    apiKeyModelsLoadSeq++
    return
  }
  const currentSeq = ++apiKeyModelsLoadSeq
  expandedApiKeyID.value = item.api_key_id
  apiKeyModelStats.value = []
  apiKeyModelsLoading.value = true
  apiKeyModelsError.value = false
  try {
    const response = await getModelStats({
      ...props.filters,
      start_date: props.startDate,
      end_date: props.endDate,
      api_key_id: item.api_key_id,
      model_source: props.source
    })
    if (currentSeq !== apiKeyModelsLoadSeq) return
    apiKeyModelStats.value = [...(response.models || [])].sort(
      (a, b) => toFiniteNumber(b.actual_cost) - toFiniteNumber(a.actual_cost)
    )
  } catch {
    if (currentSeq !== apiKeyModelsLoadSeq) return
    apiKeyModelsError.value = true
  } finally {
    if (currentSeq === apiKeyModelsLoadSeq) apiKeyModelsLoading.value = false
  }
}

const handleRankingClick = (item: RankingDisplayItem) => {
  if (item.isOther) return
  if (isApiKeyRankingItem(item)) {
    void selectApiKeyModels(item)
    return
  }
  emit('ranking-click', item)
}

const toFiniteNumber = (value: unknown): number => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : 0
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

const formatAverageDuration = (value: number | null | undefined): string => {
  const milliseconds = toFiniteNumber(value)
  if (milliseconds <= 0) return '-'
  return milliseconds >= 1000
    ? `${(milliseconds / 1000).toFixed(2)}s`
    : `${Math.round(milliseconds)}ms`
}

const formatIPLastSeen = (value: string): string => {
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '-' : date.toLocaleString()
}

const formatPercentage = (value: number, total: number): string => {
  return total > 0 ? `${((value / total) * 100).toFixed(1)}%` : '0.0%'
}

const formatRankingShare = (actualCost: number): string => {
  return formatPercentage(toFiniteNumber(actualCost), activeRankingTotals.value.totalCost)
}

const formatApiKeyModelShare = (actualCost: number): string => {
  return formatPercentage(toFiniteNumber(actualCost), apiKeyModelTotalActualCost.value)
}

const getRankingAverageDuration = (item: RankingDisplayItem): number | undefined => {
  return isApiKeyRankingItem(item) ? item.average_duration_ms : undefined
}

watch(() => props.apiKeyRankingItems, (items) => {
  if (!items.some((item) => item.api_key_id === expandedApiKeyID.value)) {
    expandedApiKeyID.value = null
    apiKeyModelStats.value = []
    apiKeyModelsLoading.value = false
    apiKeyModelsError.value = false
    apiKeyModelsLoadSeq++
  }
})

watch(() => [props.startDate, props.endDate], () => {
  expandedApiKeyID.value = null
  apiKeyModelStats.value = []
  apiKeyModelsLoading.value = false
  apiKeyModelsError.value = false
  apiKeyModelsLoadSeq++
})
</script>
