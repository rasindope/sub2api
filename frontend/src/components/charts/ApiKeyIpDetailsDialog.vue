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
      </div>

      <IpGeoBatchToolbar :ips="item.ip_usages?.map((usage) => usage.ip_address) ?? []" @failed="emit('geoFailed')" />

      <div v-if="item.ip_usages?.length" class="overflow-x-auto">
        <table class="w-full min-w-[680px] text-left text-xs">
          <thead>
            <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-gray-400">
              <th class="px-3 py-2 font-medium">IP</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.dashboard.ipLocation') }}</th>
              <th class="px-3 py-2 text-right font-medium">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.dashboard.ipFirstSeen') }}</th>
              <th class="px-3 py-2 font-medium">{{ t('admin.dashboard.ipLastSeen') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="usage in item.ip_usages" :key="usage.ip_address" class="border-b border-gray-100 dark:border-dark-700">
              <td class="whitespace-nowrap px-3 py-2 font-mono text-gray-900 dark:text-white">{{ usage.ip_address }}</td>
              <td class="px-3 py-2"><IpGeoCell :ip="usage.ip_address" /></td>
              <td class="px-3 py-2 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatNumber(usage.requests) }}</td>
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
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import IpGeoBatchToolbar from '@/components/common/IpGeoBatchToolbar.vue'
import IpGeoCell from '@/components/common/IpGeoCell.vue'
import type { ApiKeySpendingRankingItem } from '@/types'
import { formatDateTime } from '@/utils/format'

defineProps<{
  show: boolean
  item: ApiKeySpendingRankingItem | null
}>()

const emit = defineEmits<{
  close: []
  geoFailed: []
}>()

const { t } = useI18n()

const getKeyLabel = (item: ApiKeySpendingRankingItem) => item.key_name || `Key #${item.api_key_id}`
const formatNumber = (value: number) => Number(value || 0).toLocaleString()
</script>
