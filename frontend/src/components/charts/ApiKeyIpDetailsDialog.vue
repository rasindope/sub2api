<template>
  <BaseDialog
    :show="show"
    :title="t('admin.dashboard.ipDetailsTitle', { key: item ? getKeyLabel(item) : '' })"
    width="wide"
    @close="emit('close')"
  >
    <template v-if="item">
      <div class="flex flex-wrap items-center gap-x-6 gap-y-2 border-b border-gray-200 pb-3 text-sm dark:border-dark-700">
        <div>
          <span class="text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.ipDistinctCount') }}</span>
          <span class="ml-2 font-semibold tabular-nums text-gray-900 dark:text-white">{{ item.distinct_ip_count ?? 0 }}</span>
        </div>
        <div>
          <span class="text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.spendingRankingRequests') }}</span>
          <span class="ml-2 font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatNumber(item.requests) }}</span>
        </div>
        <template v-if="getActivity(item)">
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.active15m') }}</span>
            <span class="ml-2 font-semibold tabular-nums text-gray-900 dark:text-white">{{ getActivity(item)?.active_ip_count_15m }}</span>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.overlaps') }}</span>
            <span class="ml-2 font-semibold tabular-nums text-amber-600 dark:text-amber-400">{{ getActivity(item)?.overlap_count_15m }}</span>
          </div>
          <div>
            <span class="text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.maxOverlap') }}</span>
            <span class="ml-2 font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatSeconds(getActivity(item)?.max_overlap_seconds_15m) }}</span>
          </div>
        </template>
      </div>

      <IpGeoBatchToolbar :ips="item.ip_usages?.map((usage) => usage.ip_address) ?? []" @failed="emit('geoFailed')" />

      <div v-if="item.ip_usages?.length" class="space-y-2 sm:hidden">
        <div v-for="usage in item.ip_usages" :key="usage.ip_address" class="rounded-xl border border-gray-200 p-3 dark:border-dark-700">
          <div class="flex items-start justify-between gap-3">
            <div class="font-mono text-sm text-gray-900 dark:text-white">{{ usage.ip_address }}</div>
            <span v-if="usage.active_15m" class="rounded-full bg-emerald-50 px-2 py-0.5 text-xs text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300">{{ t('admin.proxies.ipActivity.active') }}</span>
          </div>
          <div class="mt-2"><IpGeoCell :ip="usage.ip_address" /></div>
          <div class="mt-3 grid grid-cols-2 gap-2 text-xs text-gray-500 dark:text-gray-400">
            <span>{{ t('admin.dashboard.spendingRankingRequests') }} {{ formatNumber(usage.requests) }}</span>
            <span>{{ t('admin.proxies.ipActivity.overlaps') }} {{ usage.overlap_count_15m ?? 0 }}</span>
            <span class="col-span-2">{{ formatDateTime(usage.last_seen_at) }}</span>
          </div>
        </div>
      </div>
      <div v-if="item.ip_usages?.length" class="hidden overflow-x-auto sm:block">
        <table class="w-full text-left text-xs">
          <thead>
            <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-gray-400">
              <th class="px-3 py-2 font-medium">IP</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.dashboard.ipLocation') }}</th>
              <th class="px-3 py-2 text-right font-medium">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th class="px-3 py-2 text-right font-medium">{{ t('admin.proxies.ipActivity.overlaps') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.dashboard.ipFirstSeen') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.dashboard.ipLastSeen') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="usage in item.ip_usages" :key="usage.ip_address" class="border-b border-gray-100 dark:border-dark-700">
              <td class="whitespace-nowrap px-3 py-2 font-mono text-gray-900 dark:text-white">{{ usage.ip_address }}</td>
              <td class="px-3 py-2"><IpGeoCell :ip="usage.ip_address" /></td>
              <td class="px-3 py-2 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatNumber(usage.requests) }}</td>
              <td class="px-3 py-2 text-right tabular-nums" :class="usage.overlap_count_15m ? 'text-amber-600 dark:text-amber-400' : 'text-gray-400'">{{ usage.overlap_count_15m ?? 0 }}</td>
              <td class="whitespace-nowrap px-3 py-2 text-gray-600 dark:text-gray-400">{{ formatDateTime(usage.first_seen_at) }}</td>
              <td class="whitespace-nowrap px-3 py-2 text-gray-600 dark:text-gray-400">{{ formatDateTime(usage.last_seen_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div v-else class="py-8 text-center text-sm text-gray-400 dark:text-gray-500">
        {{ t('admin.dashboard.ipNoData') }}
      </div>
      <p v-if="(item.distinct_ip_count ?? 0) > (item.ip_usages?.length ?? 0)" class="mt-3 text-xs text-gray-400 dark:text-gray-500">
        {{ t('admin.dashboard.ipTopOnly', { shown: item.ip_usages?.length ?? 0, total: item.distinct_ip_count }) }}
      </p>

      <section v-if="getActivity(item)" class="mt-6 border-t border-gray-200 pt-5 dark:border-dark-700">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div>
            <h3 class="font-medium text-gray-900 dark:text-white">{{ t('admin.proxies.ipActivity.overlapDetails') }}</h3>
            <p class="mt-0.5 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.proxies.ipActivity.overlapDetailsHint') }}</p>
          </div>
          <span class="text-xs tabular-nums text-gray-400">{{ overlaps.length }}</span>
        </div>
        <div v-if="overlapsLoading" class="py-6 text-center text-sm text-gray-400">{{ t('common.loading') }}</div>
        <div v-else-if="overlapsError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10 dark:text-red-300">{{ t('admin.proxies.ipActivity.overlapLoadFailed') }}</div>
        <div v-else-if="overlaps.length">
          <div class="space-y-2 sm:hidden">
            <div v-for="overlap in overlaps" :key="`${overlap.overlap_end_at}-${overlap.ip_a}-${overlap.ip_b}`" class="rounded-xl border border-amber-200/70 bg-amber-50/30 p-3 dark:border-amber-500/20 dark:bg-amber-500/5">
              <div class="text-xs text-gray-500 dark:text-gray-400">{{ formatDateTime(overlap.overlap_start_at) }} → {{ formatDateTime(overlap.overlap_end_at) }}</div>
              <div class="mt-2 flex items-center justify-between gap-3"><span class="font-mono text-xs text-gray-900 dark:text-white">{{ overlap.ip_a }}</span><span class="text-gray-400">↔</span><span class="font-mono text-xs text-gray-900 dark:text-white">{{ overlap.ip_b }}</span></div>
              <div class="mt-2 text-right text-sm font-medium tabular-nums text-amber-600 dark:text-amber-400">{{ formatSeconds(overlap.overlap_seconds) }}</div>
            </div>
          </div>
          <div class="hidden overflow-x-auto rounded-xl border border-gray-200 dark:border-dark-700 sm:block">
            <table class="w-full text-left text-xs">
              <thead class="bg-gray-50 text-gray-500 dark:bg-dark-800 dark:text-gray-400"><tr><th class="px-3 py-2 font-medium">{{ t('admin.proxies.ipActivity.overlapTime') }}</th><th class="px-3 py-2 font-medium">IP A</th><th class="px-3 py-2 font-medium">IP B</th><th class="px-3 py-2 text-right font-medium">{{ t('admin.proxies.ipActivity.duration') }}</th></tr></thead>
              <tbody><tr v-for="overlap in overlaps" :key="`${overlap.overlap_end_at}-${overlap.ip_a}-${overlap.ip_b}`" class="border-t border-gray-100 dark:border-dark-700"><td class="whitespace-nowrap px-3 py-2 text-gray-600 dark:text-gray-400">{{ formatDateTime(overlap.overlap_start_at) }} → {{ formatDateTime(overlap.overlap_end_at) }}</td><td class="whitespace-nowrap px-3 py-2 font-mono text-gray-900 dark:text-white">{{ overlap.ip_a }}</td><td class="whitespace-nowrap px-3 py-2 font-mono text-gray-900 dark:text-white">{{ overlap.ip_b }}</td><td class="px-3 py-2 text-right font-medium tabular-nums text-amber-600 dark:text-amber-400">{{ formatSeconds(overlap.overlap_seconds) }}</td></tr></tbody>
            </table>
          </div>
        </div>
        <div v-else class="py-6 text-center text-sm text-gray-400">{{ t('admin.proxies.ipActivity.noOverlapDetails') }}</div>
      </section>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { getApiKeyIPOverlaps } from '@/api/admin/dashboard'
import BaseDialog from '@/components/common/BaseDialog.vue'
import IpGeoBatchToolbar from '@/components/common/IpGeoBatchToolbar.vue'
import IpGeoCell from '@/components/common/IpGeoCell.vue'
import type { ApiKeyIPActivityItem, ApiKeyIPOverlap, ApiKeySpendingRankingItem } from '@/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{
  show: boolean
  item: ApiKeySpendingRankingItem | ApiKeyIPActivityItem | null
  startDate?: string
  endDate?: string
}>()

const emit = defineEmits<{
  close: []
  geoFailed: []
}>()

const { t } = useI18n()

type IPDetailsItem = ApiKeySpendingRankingItem | ApiKeyIPActivityItem
const getKeyLabel = (item: IPDetailsItem) => item.key_name || `Key #${item.api_key_id}`
const getActivity = (item: IPDetailsItem) => 'risk_level' in item ? item : null
const formatNumber = (value: number) => Number(value || 0).toLocaleString()
const formatSeconds = (value?: number) => `${Number(value || 0).toFixed(1)}s`
const overlaps = ref<ApiKeyIPOverlap[]>([])
const overlapsLoading = ref(false)
const overlapsError = ref(false)
let overlapLoadSeq = 0

watch(
  () => [props.show, props.item?.api_key_id, props.startDate, props.endDate] as const,
  async ([show, apiKeyId]) => {
    const seq = ++overlapLoadSeq
    overlaps.value = []
    overlapsError.value = false
    overlapsLoading.value = false
    if (!show || !apiKeyId || !props.item || !getActivity(props.item)) return
    overlapsLoading.value = true
    try {
      const response = await getApiKeyIPOverlaps(apiKeyId, { limit: 50, start_date: props.startDate, end_date: props.endDate })
      if (seq === overlapLoadSeq) overlaps.value = response.items
    }
    catch {
      if (seq === overlapLoadSeq) overlapsError.value = true
    }
    finally {
      if (seq === overlapLoadSeq) overlapsLoading.value = false
    }
  },
  { immediate: true }
)
</script>
