<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { opsAPI, type OpsNginxTimingKeyDetails, type OpsNginxTimingKeyMetric, type OpsNginxTimingMetric, type OpsPercentiles } from '@/api/admin/ops'
import { formatNumber } from '@/utils/format'

type PercentileField = 'p99_ms' | 'p95_ms' | 'p90_ms' | 'p50_ms' | 'avg_ms' | 'max_ms'

interface Props {
  modelValue: boolean
  metric: OpsNginxTimingMetric
  timeRange: string
  customStartTime?: string | null
  customEndTime?: string | null
  apiKeyIds?: number[]
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
const rows = computed(() => {
  const items = [...(details.value?.items ?? [])]
  return items.sort((a, b) => metricScore(b) - metricScore(a) || b.http_request_count - a.http_request_count || a.api_key_id - b.api_key_id)
})

function close() {
  emit('update:modelValue', false)
}

function metricScore(row: OpsNginxTimingKeyMetric): number {
  switch (props.metric) {
    case 'requests':
      return row.http_request_count
    case 'ingress_errors':
      return row.client_timeout_408_count + row.client_closed_499_count + row.server_error_5xx_count
    default:
      return metricPercentiles(row).p99_ms ?? -1
  }
}

function metricPercentiles(row: OpsNginxTimingKeyMetric): OpsPercentiles {
  switch (props.metric) {
    case 'gateway_connect':
      return row.upstream_connect_time
    case 'upstream_response':
      return row.upstream_response_time
    case 'client_overhead':
      return row.client_overhead_time
    default:
      return row.request_time
  }
}

function formatMs(value?: number | null): string {
  return typeof value === 'number' && Number.isFinite(value) ? formatNumber(value) : '-'
}

function formatMetricPercentile(row: OpsNginxTimingKeyMetric, field: PercentileField): string {
  return formatMs(metricPercentiles(row)[field])
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
</script>

<template>
  <BaseDialog :show="modelValue" :title="t('admin.ops.nginxTiming.details.title', { metric: modalTitle })" width="extra-wide" @close="close">
    <div class="flex min-h-0 flex-col">
      <div class="mb-4 flex flex-wrap items-start justify-between gap-3 border-b border-gray-100 pb-4 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
        <div class="space-y-1">
          <div>{{ t('admin.ops.nginxTiming.details.description') }}</div>
          <div v-if="details?.available">{{ t('admin.ops.nginxTiming.details.matchedRequests', { count: formatNumber(details.matched_request_count) }) }}</div>
          <div v-if="details?.unattributed_error_count" class="text-amber-600 dark:text-amber-400">
            {{ t('admin.ops.nginxTiming.details.unattributedErrors', { count: formatNumber(details.unattributed_error_count) }) }}
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
        {{ t('admin.ops.nginxTiming.details.empty') }}
      </div>

      <div v-else class="overflow-auto rounded-xl border border-gray-200 dark:border-dark-700">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-900">
            <tr>
              <th class="sticky left-0 bg-gray-50 px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:bg-dark-900 dark:text-gray-400">Key</th>
              <template v-if="props.metric === 'requests'">
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">HTTP</th>
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">{{ t('admin.ops.nginxTiming.details.success') }}</th>
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">WS</th>
              </template>
              <template v-else-if="props.metric === 'ingress_errors'">
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">HTTP</th>
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">408</th>
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">499</th>
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">5xx</th>
              </template>
              <template v-else>
                <th class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">HTTP</th>
                <th v-for="field in percentileFields" :key="field.key" class="px-4 py-3 text-right text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                  {{ field.label }} (ms)
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
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatNumber(row.http_request_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatNumber(row.success_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatNumber(row.websocket_session_count) }}</td>
              </template>
              <template v-else-if="props.metric === 'ingress_errors'">
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatNumber(row.http_request_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-rose-600 dark:text-rose-400">{{ formatNumber(row.client_timeout_408_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-rose-600 dark:text-rose-400">{{ formatNumber(row.client_closed_499_count) }}</td>
                <td class="px-4 py-3 text-right text-sm text-rose-600 dark:text-rose-400">{{ formatNumber(row.server_error_5xx_count) }}</td>
              </template>
              <template v-else>
                <td class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">{{ formatNumber(row.http_request_count) }}</td>
                <td v-for="field in percentileFields" :key="field.key" class="px-4 py-3 text-right text-sm text-gray-700 dark:text-gray-200">
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
