<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { opsAPI, type OpsNginxTimingKeyDetails, type OpsNginxTimingKeyMetric, type OpsNginxTimingMetric, type OpsPercentiles } from '@/api/admin/ops'
import { formatNumberLocaleString } from '@/utils/format'

type PercentileField = 'p99_ms' | 'p95_ms' | 'p90_ms' | 'p50_ms' | 'avg_ms' | 'max_ms'
type SortOrder = 'asc' | 'desc'
type ClientTimingDetail = 'total' | 'upload' | 'response_receive'
type SortKey =
  | 'key_name'
  | 'http_request_count'
  | 'sample_count'
  | 'success_count'
  | 'websocket_session_count'
  | PercentileField

interface Props {
  modelValue: boolean
  metric: OpsNginxTimingMetric
  timeRange: string
  customStartTime?: string | null
  customEndTime?: string | null
  apiKeyIds?: number[]
  summary?: OpsPercentiles
  thresholdMs?: number | null
}

const props = defineProps<Props>()
const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()
const { t } = useI18n()

const loading = ref(false)
const details = ref<OpsNginxTimingKeyDetails | null>(null)

const percentileFields: Array<{ key: PercentileField, label: string }> = [
  { key: 'p99_ms', label: 'P99' },
  { key: 'p95_ms', label: 'P95' },
  { key: 'p90_ms', label: 'P90' },
  { key: 'p50_ms', label: 'P50' },
  { key: 'avg_ms', label: 'Avg' },
  { key: 'max_ms', label: 'Max' }
]

const modalTitle = computed(() => t(`admin.ops.nginxTiming.metricTitles.${props.metric}`))
const hasDurationMetric = computed(() => props.metric !== 'requests')
const isClientOverhead = computed(() => props.metric === 'client_overhead')
const clientTimingDetail = ref<ClientTimingDetail>('total')
const metricDescriptionKeys: Partial<Record<OpsNginxTimingMetric, string>> = {
  request_time: 'requestTime',
  client_overhead: 'clientOverhead'
}
const metricDescription = computed(() => {
  if (isClientOverhead.value) {
    switch (clientTimingDetail.value) {
      case 'upload':
        return t('admin.ops.nginxTiming.tooltips.clientUpload')
      case 'response_receive':
        return t('admin.ops.nginxTiming.tooltips.clientResponseReceive')
      default:
        return t('admin.ops.nginxTiming.tooltips.clientOverhead')
    }
  }
  const key = metricDescriptionKeys[props.metric]
  return key ? t(`admin.ops.nginxTiming.tooltips.${key}`) : ''
})
const sortKey = ref<SortKey>(defaultSortKey(props.metric))
const sortOrder = ref<SortOrder>('desc')
const durationCountSortKey = computed<'sample_count' | 'http_request_count'>(() => {
  return isClientOverhead.value ? 'sample_count' : 'http_request_count'
})
const durationCountLabel = computed(() => {
  return isClientOverhead.value ? t('admin.ops.nginxTiming.details.samples') : 'HTTP'
})
const clientTimingSampleCount = computed(() => {
  return (details.value?.items ?? []).reduce((count, row) => count + timingSampleCount(row), 0)
})
const rows = computed(() => {
  const items = [...(details.value?.items ?? [])].filter((row) => !isClientOverhead.value || timingSampleCount(row) > 0)
  return items.sort((a, b) => {
    const comparison = compareSortValues(sortValue(a, sortKey.value), sortValue(b, sortKey.value), sortOrder.value)
    if (comparison !== 0) return comparison
    return a.api_key_id - b.api_key_id
  })
})

function close() {
  emit('update:modelValue', false)
}

function defaultSortKey(metric: OpsNginxTimingMetric): SortKey {
  switch (metric) {
    case 'requests':
      return 'http_request_count'
    default:
      return 'p99_ms'
  }
}

function metricPercentiles(row: OpsNginxTimingKeyMetric): OpsPercentiles {
  if (isClientOverhead.value) {
    switch (clientTimingDetail.value) {
      case 'upload':
        return row.client_upload_time
      case 'response_receive':
        return row.client_response_receive_time
      default:
        return row.client_overhead_time
    }
  }
  return row.request_time
}

function timingSampleCount(row: OpsNginxTimingKeyMetric): number {
  const count = clientTimingDetail.value === 'total'
    ? row.client_overhead_sample_count
    : clientTimingDetail.value === 'upload'
      ? row.client_upload_sample_count
      : row.client_response_receive_sample_count
  return Number.isFinite(count) ? count : 0
}

function sortValue(row: OpsNginxTimingKeyMetric, key: SortKey): string | number | null {
  switch (key) {
    case 'key_name':
      return row.key_name || `Key #${row.api_key_id}`
    case 'http_request_count':
      return row.http_request_count
    case 'sample_count':
      return timingSampleCount(row)
    case 'success_count':
      return row.success_count
    case 'websocket_session_count':
      return row.websocket_session_count
    default:
      return metricPercentiles(row)[key] ?? null
  }
}

function compareSortValues(left: string | number | null, right: string | number | null, order: SortOrder): number {
  if (left === null && right === null) return 0
  if (left === null) return 1
  if (right === null) return -1
  let comparison: number
  if (typeof left === 'string' && typeof right === 'string') {
    comparison = left.localeCompare(right, undefined, { numeric: true, sensitivity: 'base' })
  } else {
    comparison = Number(left) - Number(right)
  }
  return order === 'asc' ? comparison : -comparison
}

function toggleSort(key: SortKey) {
  if (sortKey.value === key) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
    return
  }
  sortKey.value = key
  sortOrder.value = key === 'key_name' ? 'asc' : 'desc'
}

function sortIconName(key: SortKey): 'arrowUp' | 'arrowDown' | 'arrowsUpDown' {
  if (sortKey.value !== key) return 'arrowsUpDown'
  return sortOrder.value === 'asc' ? 'arrowUp' : 'arrowDown'
}

function sortHeaderClass(key: SortKey, align: 'left' | 'right'): string {
  const alignment = align === 'left' ? 'justify-start' : 'justify-end'
  const color = sortKey.value === key ? 'text-primary-600 dark:text-primary-400' : 'text-gray-500 hover:text-gray-800 dark:text-gray-400 dark:hover:text-gray-100'
  return `inline-flex w-full items-center gap-1 ${alignment} ${color}`
}

function ariaSort(key: SortKey): 'ascending' | 'descending' | 'none' {
  if (sortKey.value !== key) return 'none'
  return sortOrder.value === 'asc' ? 'ascending' : 'descending'
}

function formatMs(value?: number | null): string {
  return typeof value === 'number' && Number.isFinite(value) ? formatNumberLocaleString(value) : '-'
}

function formatCount(value?: number | null): string {
  return typeof value === 'number' && Number.isFinite(value) ? formatNumberLocaleString(value) : '-'
}

function isThresholdExceeded(value?: number | null): boolean {
  return (!isClientOverhead.value || clientTimingDetail.value === 'total') && props.thresholdMs != null && value != null && value > props.thresholdMs
}

function durationValueClass(value?: number | null): string {
  return isThresholdExceeded(value) ? 'text-rose-600 dark:text-rose-400' : 'text-gray-700 dark:text-gray-200'
}

function formatMetricPercentile(row: OpsNginxTimingKeyMetric, field: PercentileField): string {
  return formatMs(metricPercentiles(row)[field])
}

function metricPercentileClass(row: OpsNginxTimingKeyMetric, field: PercentileField): string {
  return durationValueClass(metricPercentiles(row)[field])
}

function buildParams() {
  const params: {
    time_range?: '5m' | '30m' | '1h' | '6h' | '24h'
    start_time?: string
    end_time?: string
    api_key_ids?: string
  } = {}
  if (props.timeRange === 'custom' && props.customStartTime && props.customEndTime) {
    params.start_time = props.customStartTime
    params.end_time = props.customEndTime
  } else if (props.timeRange !== 'custom') {
    params.time_range = props.timeRange as '5m' | '30m' | '1h' | '6h' | '24h'
  } else {
    params.time_range = '1h'
  }
  if (props.apiKeyIds?.length) params.api_key_ids = props.apiKeyIds.join(',')
  return params
}

async function fetchDetails() {
  if (!props.modelValue) return
  loading.value = true
  try {
    details.value = await opsAPI.getNginxTimingKeyDetails(buildParams())
  } catch (err) {
    console.error('[OpsNginxTimingDetailsModal] Failed to load Key metrics', err)
    details.value = null
  } finally {
    loading.value = false
  }
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) void fetchDetails()
  }
)

watch(
  () => [props.metric, props.timeRange, props.customStartTime, props.customEndTime, props.apiKeyIds?.join(',')] as const,
  () => {
    if (props.modelValue) void fetchDetails()
  }
)

watch(
  () => props.metric,
  (metric) => {
    sortKey.value = defaultSortKey(metric)
    sortOrder.value = 'desc'
  }
)
</script>

<template>
  <BaseDialog :show="modelValue" :title="t('admin.ops.nginxTiming.details.title', { metric: modalTitle })" width="extra-wide" @close="close">
    <div class="flex min-h-0 flex-col">
      <div class="mb-4 flex flex-wrap items-start justify-between gap-3 border-b border-gray-100 pb-4 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
        <div class="space-y-1">
          <div>{{ t('admin.ops.nginxTiming.details.description') }}</div>
          <div v-if="details?.available">{{ t('admin.ops.nginxTiming.details.matchedRequests', { count: formatCount(details.matched_request_count) }) }}</div>
          <div v-if="hasDurationMetric" class="mt-3 flex flex-wrap items-center gap-x-5 gap-y-1 rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-900 dark:text-gray-300">
            <span>{{ t('admin.ops.nginxTiming.details.meaning', { value: metricDescription }) }}</span>
            <span v-if="isClientOverhead">
              {{ t('admin.ops.nginxTiming.details.validSamples', { count: formatCount(clientTimingSampleCount) }) }}
            </span>
            <span v-else>
              {{ t('admin.ops.nginxTiming.details.currentP99') }}:
              <strong :class="durationValueClass(summary?.p99_ms)">{{ formatMs(summary?.p99_ms) }} ms</strong>
            </span>
            <span v-if="!isClientOverhead && thresholdMs != null">
              {{ t('admin.ops.nginxTiming.details.redThreshold') }}:
              <strong>{{ formatMs(thresholdMs) }} ms</strong>
            </span>
          </div>
          <div v-if="isClientOverhead" class="mt-3 inline-flex rounded-lg bg-gray-100 p-1 dark:bg-dark-900" role="tablist">
            <button
              type="button"
              role="tab"
              :aria-selected="clientTimingDetail === 'total'"
              class="rounded-md px-3 py-1.5 text-xs font-semibold"
              :class="clientTimingDetail === 'total' ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-400' : 'text-gray-500 dark:text-gray-400'"
              @click="clientTimingDetail = 'total'"
            >
              {{ t('admin.ops.nginxTiming.details.clientTotal') }}
            </button>
            <button
              type="button"
              role="tab"
              :aria-selected="clientTimingDetail === 'upload'"
              class="rounded-md px-3 py-1.5 text-xs font-semibold"
              :class="clientTimingDetail === 'upload' ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-400' : 'text-gray-500 dark:text-gray-400'"
              @click="clientTimingDetail = 'upload'"
            >
              {{ t('admin.ops.nginxTiming.details.clientUpload') }}
            </button>
            <button
              type="button"
              role="tab"
              :aria-selected="clientTimingDetail === 'response_receive'"
              class="rounded-md px-3 py-1.5 text-xs font-semibold"
              :class="clientTimingDetail === 'response_receive' ? 'bg-white text-primary-600 shadow-sm dark:bg-dark-800 dark:text-primary-400' : 'text-gray-500 dark:text-gray-400'"
              @click="clientTimingDetail = 'response_receive'"
            >
              {{ t('admin.ops.nginxTiming.details.clientResponseReceive') }}
            </button>
          </div>
        </div>
        <button type="button" class="btn btn-secondary btn-sm" @click="fetchDetails">
          {{ t('common.refresh') }}
        </button>
      </div>

      <div v-if="loading" class="py-12 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ t('common.loading') }}
      </div>

      <div v-else-if="!details?.available" class="rounded-xl border border-dashed border-gray-200 p-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
        {{ t('admin.ops.nginxTiming.logUnavailable') }}
      </div>

      <div v-else-if="rows.length === 0" class="rounded-xl border border-dashed border-gray-200 p-10 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
        {{ isClientOverhead ? t('admin.ops.nginxTiming.details.clientTimingEmpty') : t('admin.ops.nginxTiming.details.empty') }}
      </div>

      <div v-else class="overflow-auto rounded-xl border border-gray-200 dark:border-dark-700">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-900">
            <tr>
              <th :aria-sort="ariaSort('key_name')" class="sticky left-0 bg-gray-50 px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider dark:bg-dark-900">
                <button type="button" :class="sortHeaderClass('key_name', 'left')" @click="toggleSort('key_name')">
                  Key
                  <Icon :name="sortIconName('key_name')" size="xs" aria-hidden="true" />
                </button>
              </th>
              <template v-if="props.metric === 'requests'">
                <th :aria-sort="ariaSort('http_request_count')" class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider dark:text-gray-400">
                  <button type="button" :class="sortHeaderClass('http_request_count', 'right')" @click="toggleSort('http_request_count')">HTTP<Icon :name="sortIconName('http_request_count')" size="xs" aria-hidden="true" /></button>
                </th>
                <th :aria-sort="ariaSort('success_count')" class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider dark:text-gray-400">
                  <button type="button" :class="sortHeaderClass('success_count', 'right')" @click="toggleSort('success_count')">{{ t('admin.ops.nginxTiming.details.success') }}<Icon :name="sortIconName('success_count')" size="xs" aria-hidden="true" /></button>
                </th>
                <th :aria-sort="ariaSort('websocket_session_count')" class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider dark:text-gray-400">
                  <button type="button" :class="sortHeaderClass('websocket_session_count', 'right')" @click="toggleSort('websocket_session_count')">WS<Icon :name="sortIconName('websocket_session_count')" size="xs" aria-hidden="true" /></button>
                </th>
              </template>
              <template v-else>
                <th :aria-sort="ariaSort(durationCountSortKey)" class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider dark:text-gray-400">
                  <button type="button" :class="sortHeaderClass(durationCountSortKey, 'right')" @click="toggleSort(durationCountSortKey)">{{ durationCountLabel }}<Icon :name="sortIconName(durationCountSortKey)" size="xs" aria-hidden="true" /></button>
                </th>
                <th v-for="field in percentileFields" :key="field.key" :aria-sort="ariaSort(field.key)" class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider dark:text-gray-400">
                  <button type="button" :class="sortHeaderClass(field.key, 'right')" @click="toggleSort(field.key)">
                    {{ field.label }} (ms)
                    <Icon :name="sortIconName(field.key)" size="xs" aria-hidden="true" />
                  </button>
                </th>
              </template>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-800">
            <tr v-for="row in rows" :key="row.api_key_id" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
              <td class="sticky left-0 max-w-[280px] truncate bg-white px-4 py-3 text-sm font-semibold text-gray-800 dark:bg-dark-800 dark:text-gray-100" :title="row.key_name">
                {{ row.key_name || `Key #${row.api_key_id}` }}
              </td>
              <template v-if="props.metric === 'requests'">
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatCount(row.http_request_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatCount(row.success_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatCount(row.websocket_session_count) }}</td>
              </template>
              <template v-else>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatCount(isClientOverhead ? timingSampleCount(row) : row.http_request_count) }}</td>
                <td v-for="field in percentileFields" :key="field.key" class="px-4 py-3 text-right text-sm" :class="metricPercentileClass(row, field.key)">
                  {{ formatMetricPercentile(row, field.key) }}
                </td>
              </template>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </BaseDialog>
</template>
