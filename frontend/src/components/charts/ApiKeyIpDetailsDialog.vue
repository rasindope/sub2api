<template>
  <BaseDialog
    :show="show"
    :title="t('admin.dashboard.ipDetailsTitle', { key: keyLabel })"
    width="wide"
    @close="emit('close')"
  >
    <div class="flex flex-wrap items-center gap-x-6 gap-y-2 border-b border-gray-200 pb-3 text-sm dark:border-dark-700">
      <div>
        <span class="text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.ipDistinctCount') }}</span>
        <span class="ml-2 font-semibold tabular-nums text-gray-900 dark:text-white">{{ item?.distinct_ip_count ?? 0 }}</span>
      </div>
      <div>
        <span class="text-gray-500 dark:text-gray-400">{{ t('admin.dashboard.spendingRankingRequests') }}</span>
        <span class="ml-2 font-semibold tabular-nums text-gray-900 dark:text-white">{{ formatNumber(item?.requests ?? 0) }}</span>
      </div>
    </div>

    <IpGeoBatchToolbar :ips="ipAddresses" @failed="emit('geoFailed')" />

    <div v-if="ipUsages.length > 0" class="overflow-x-auto">
      <table class="w-full min-w-[680px] text-left text-xs">
        <thead class="text-gray-500 dark:text-gray-400">
          <tr>
            <th class="py-2 pr-4 font-medium">IP</th>
            <th class="py-2 pr-4 font-medium">{{ t('admin.dashboard.ipLocation') }}</th>
            <th class="py-2 pr-4 text-right font-medium">{{ t('admin.dashboard.spendingRankingRequests') }}</th>
            <th class="py-2 pr-4 font-medium">{{ t('admin.dashboard.ipFirstSeen') }}</th>
            <th class="py-2 font-medium">{{ t('admin.dashboard.ipLastSeen') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="usage in ipUsages"
            :key="usage.ip_address"
            class="border-t border-gray-100 dark:border-dark-700"
          >
            <td class="py-2 pr-4 font-mono text-gray-800 dark:text-gray-200">{{ usage.ip_address }}</td>
            <td class="py-2 pr-4"><IpGeoCell :ip="usage.ip_address" /></td>
            <td class="py-2 pr-4 text-right tabular-nums text-gray-700 dark:text-gray-300">{{ formatNumber(usage.requests) }}</td>
            <td class="py-2 pr-4 whitespace-nowrap text-gray-500 dark:text-gray-400">{{ formatDateTimeToMinute(usage.first_seen_at) }}</td>
            <td class="py-2 whitespace-nowrap text-gray-500 dark:text-gray-400">{{ formatDateTimeToMinute(usage.last_seen_at) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-else class="py-8 text-center text-sm text-gray-400 dark:text-gray-500">
      {{ t('admin.dashboard.ipNoData') }}
    </div>

    <p v-if="hasMore" class="mt-3 text-xs text-gray-400 dark:text-gray-500">
      {{ t('admin.dashboard.ipTopOnly', { shown: ipUsages.length, total: item?.distinct_ip_count ?? 0 }) }}
    </p>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import IpGeoBatchToolbar from '@/components/common/IpGeoBatchToolbar.vue'
import IpGeoCell from '@/components/common/IpGeoCell.vue'
import { formatDateTimeToMinute, formatNumber } from '@/utils/format'
import type { ApiKeySpendingRankingItem } from '@/types'

const props = defineProps<{
  show: boolean
  item: ApiKeySpendingRankingItem | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'geoFailed'): void
}>()

const { t } = useI18n()
const ipUsages = computed(() => props.item?.ip_usages ?? [])
const ipAddresses = computed(() => ipUsages.value.map((usage) => usage.ip_address))
const keyLabel = computed(() => props.item?.key_name || `Key #${props.item?.api_key_id ?? ''}`)
const hasMore = computed(() => (props.item?.distinct_ip_count ?? 0) > ipUsages.value.length)
</script>
