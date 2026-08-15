<template>
  <button
    type="button"
    class="card group w-full p-4 text-left transition-colors hover:border-emerald-200 hover:bg-emerald-50/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500/40 dark:hover:border-emerald-800 dark:hover:bg-emerald-950/20"
    data-testid="account-usage-card"
    :aria-label="t('admin.dashboard.accountUsageOpen')"
    @click="openDetails"
  >
    <div class="flex items-center justify-between gap-2">
      <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
        {{ t('admin.dashboard.accountUsage') }} · 7d
      </p>
      <Icon name="chevronRight" size="sm" class="shrink-0 text-gray-300 transition-transform group-hover:translate-x-0.5 group-hover:text-emerald-500 dark:text-dark-500" />
    </div>

    <div v-if="overviewLoading && !usageAccounts.length" class="mt-2 space-y-2">
      <div v-for="index in 3" :key="index" class="flex items-center gap-2">
        <div class="h-3 min-w-0 flex-1 animate-pulse rounded bg-gray-200 dark:bg-dark-600" />
        <div class="h-3 w-24 animate-pulse rounded bg-gray-200 dark:bg-dark-600" />
      </div>
    </div>
    <p v-else-if="overviewError && !usageAccounts.length" class="mt-2 text-xs text-red-500">
      {{ t('admin.dashboard.accountUsageFailed') }}
    </p>
    <p v-else-if="usageAccounts.length === 0" class="mt-2 text-xs text-gray-500 dark:text-gray-400">
      {{ t('admin.dashboard.accountUsageEmpty') }}
    </p>
    <div v-else class="mt-2 space-y-1.5">
      <div v-for="account in usageAccounts" :key="account.id" class="flex min-w-0 items-center gap-2">
        <span class="min-w-0 flex-1 truncate text-[11px] font-medium text-gray-700 dark:text-gray-300" :title="account.name">
          {{ account.name }}
        </span>
        <span v-if="usageLoadingByAccountId[String(account.id)]" class="h-3 w-24 animate-pulse rounded bg-gray-200 dark:bg-dark-600" />
        <span
          v-else-if="usageErrorByAccountId[String(account.id)] || usageByAccountId[String(account.id)]?.error"
          class="w-24 text-right text-[10px] text-red-500"
        >
          {{ t('common.error') }}
        </span>
        <UsageProgressBar
          v-else-if="usageByAccountId[String(account.id)]?.seven_day"
          label="7d"
          :utilization="usageByAccountId[String(account.id)]!.seven_day!.utilization"
          color="emerald"
        />
        <span v-else class="w-24 text-right text-[10px] text-gray-400">7d -</span>
      </div>
    </div>
  </button>

  <BaseDialog
    :show="showDetails"
    :title="t('admin.dashboard.accountUsage')"
    width="wide"
    :close-on-click-outside="true"
    @close="showDetails = false"
  >
    <div class="mb-4 flex flex-wrap items-center justify-between gap-3 border-b border-gray-100 pb-4 dark:border-dark-700">
      <div class="text-sm text-gray-500 dark:text-gray-400">
        <span v-if="attentionCount > 0" class="font-medium text-amber-600 dark:text-amber-400">
          {{ t('admin.dashboard.accountUsageAttention', { count: attentionCount }) }}
        </span>
        <span v-else class="font-medium text-green-600 dark:text-green-400">
          {{ t('admin.dashboard.accountUsageHealthy') }}
        </span>
        <span> · {{ t('admin.dashboard.accountUsageAccounts', { count: usageAccounts.length }) }}</span>
      </div>
      <button
        type="button"
        class="btn btn-secondary"
        data-testid="refresh-account-usage"
        :disabled="overviewLoading || usageAccounts.length === 0"
        @click="loadOverview(true)"
      >
        <Icon name="refresh" size="sm" :class="{ 'animate-spin': overviewLoading }" />
        {{ t('admin.dashboard.accountUsageRefresh') }}
      </button>
    </div>

    <div v-if="overviewLoading && !usageAccounts.length" class="flex min-h-40 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="overviewError && !usageAccounts.length" class="flex min-h-40 flex-col items-center justify-center gap-3 text-sm text-gray-500 dark:text-gray-400">
      <span>{{ t('admin.dashboard.accountUsageFailed') }}</span>
      <button type="button" class="btn btn-secondary" @click="loadOverview(false)">
        <Icon name="refresh" size="sm" />
        {{ t('admin.dashboard.accountUsageRefresh') }}
      </button>
    </div>
    <div v-else-if="usageAccounts.length === 0" class="flex min-h-40 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('admin.dashboard.accountUsageEmpty') }}
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="account in sortedAccounts"
        :key="account.id"
        class="grid gap-3 rounded-lg border border-gray-200 px-4 py-3 dark:border-dark-700 sm:grid-cols-[minmax(0,220px)_1fr] sm:items-center"
      >
        <div class="min-w-0">
          <div class="flex min-w-0 items-center gap-2">
            <span class="truncate text-sm font-medium text-gray-900 dark:text-white" :title="account.name">
              {{ account.name }}
            </span>
            <span class="shrink-0 rounded bg-gray-100 px-1.5 py-0.5 text-[10px] uppercase text-gray-500 dark:bg-dark-700 dark:text-gray-400">
              {{ account.platform }}
            </span>
          </div>
          <p v-if="!account.schedulable" class="mt-1 text-[11px] text-gray-400 dark:text-gray-500">
            {{ t('admin.dashboard.accountUsageNotScheduled') }}
          </p>
        </div>
        <AccountUsageCell
          :account="account"
          :batched-usage="usageByAccountId[String(account.id)] ?? null"
          :batched-usage-error="usageErrorByAccountId[String(account.id)] ?? null"
          :batched-usage-loading="usageLoadingByAccountId[String(account.id)] === true"
          :request-batched-usage="requestAccountUsage"
          @usage-loaded="handleUsageLoaded(account.id, $event)"
        />
      </div>
    </div>

    <template #footer>
      <button type="button" class="btn btn-secondary" @click="goToAccounts">
        {{ t('admin.dashboard.manageAccounts') }}
        <Icon name="arrowRight" size="sm" />
      </button>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { adminAPI } from '@/api/admin'
import type { Account, AccountUsageInfo, UsageProgress } from '@/types'
import AccountUsageCell from '@/components/account/AccountUsageCell.vue'
import UsageProgressBar from '@/components/account/UsageProgressBar.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()
const router = useRouter()

const showDetails = ref(false)
const overviewLoading = ref(false)
const overviewError = ref(false)
const usageAccounts = ref<Account[]>([])
const usageByAccountId = ref<Record<string, AccountUsageInfo | null>>({})
const usageErrorByAccountId = ref<Record<string, string | null>>({})
const usageLoadingByAccountId = ref<Record<string, boolean>>({})
let loadSequence = 0

const supportsUsageWindows = (account: Account): boolean => {
  if (account.platform === 'gemini') return true
  return account.type === 'oauth' || account.type === 'setup-token'
}

const progressUtilization = (progress: UsageProgress | null | undefined): number | null => {
  const value = Number(progress?.utilization)
  return Number.isFinite(value) ? value : null
}

const getMaxUtilization = (usage: AccountUsageInfo | null | undefined): number | null => {
  if (!usage) return null
  const values = [
    usage.five_hour,
    usage.seven_day,
    usage.seven_day_sonnet,
    usage.seven_day_fable,
    usage.thirty_day,
    usage.gemini_shared_daily,
    usage.gemini_pro_daily,
    usage.gemini_flash_daily,
    usage.gemini_shared_minute,
    usage.gemini_pro_minute,
    usage.gemini_flash_minute
  ].map(progressUtilization).filter((value): value is number => value !== null)

  for (const quota of Object.values(usage.antigravity_quota ?? {})) {
    const value = Number(quota.utilization)
    if (Number.isFinite(value)) values.push(value)
  }

  return values.length > 0 ? Math.max(...values) : null
}

const accountNeedsAttention = (account: Account): boolean => {
  const key = String(account.id)
  const usage = usageByAccountId.value[key]
  return Boolean(
    usageErrorByAccountId.value[key]
      || usage?.error
      || usage?.is_forbidden
      || usage?.is_banned
      || usage?.needs_reauth
      || usage?.needs_verify
      || (getMaxUtilization(usage) ?? 0) >= 80
  )
}

const attentionCount = computed(() => usageAccounts.value.filter(accountNeedsAttention).length)
const sortedAccounts = computed(() => [...usageAccounts.value].sort((left, right) => {
  const attentionDiff = Number(accountNeedsAttention(right)) - Number(accountNeedsAttention(left))
  if (attentionDiff !== 0) return attentionDiff
  const usageDiff = (getMaxUtilization(usageByAccountId.value[String(right.id)]) ?? -1)
    - (getMaxUtilization(usageByAccountId.value[String(left.id)]) ?? -1)
  return usageDiff !== 0 ? usageDiff : left.name.localeCompare(right.name)
}))

const setAccountsLoading = (accounts: Account[], loading: boolean) => {
  usageLoadingByAccountId.value = Object.fromEntries(accounts.map(account => [String(account.id), loading]))
}

const loadOverview = async (force: boolean) => {
  const sequence = ++loadSequence
  overviewLoading.value = true
  overviewError.value = false
  try {
    const response = await adminAPI.accounts.list(1, 1000, {
      status: 'active',
      include_scheduler_score: '0',
      sort_by: 'name',
      sort_order: 'asc'
    })
    if (sequence !== loadSequence) return

    const accounts = response.items.filter(supportsUsageWindows)
    usageAccounts.value = accounts
    setAccountsLoading(accounts, true)
    if (accounts.length === 0) return

    const result = await adminAPI.accounts.getBatchUsage(accounts.map(account => account.id), force)
    if (sequence !== loadSequence) return
    usageByAccountId.value = result.usage ?? {}
    usageErrorByAccountId.value = result.errors ?? {}
  } catch (error) {
    if (sequence !== loadSequence) return
    overviewError.value = true
    if (usageAccounts.value.length > 0) {
      usageErrorByAccountId.value = Object.fromEntries(
        usageAccounts.value.map(account => [String(account.id), t('admin.dashboard.accountUsageFailed')])
      )
    }
    console.error('Failed to load dashboard account usage:', error)
  } finally {
    if (sequence === loadSequence) {
      setAccountsLoading(usageAccounts.value, false)
      overviewLoading.value = false
    }
  }
}

const requestAccountUsage = (account: Account, options?: { force?: boolean }) => {
  const key = String(account.id)
  if (!options?.force && (usageByAccountId.value[key] || usageLoadingByAccountId.value[key])) return
  usageLoadingByAccountId.value = { ...usageLoadingByAccountId.value, [key]: true }
  void adminAPI.accounts.getBatchUsage([account.id], options?.force === true)
    .then((result) => {
      usageByAccountId.value = { ...usageByAccountId.value, [key]: result.usage?.[key] ?? null }
      usageErrorByAccountId.value = { ...usageErrorByAccountId.value, [key]: result.errors?.[key] ?? null }
    })
    .catch((error) => {
      usageErrorByAccountId.value = { ...usageErrorByAccountId.value, [key]: t('admin.dashboard.accountUsageFailed') }
      console.error('Failed to load account usage:', error)
    })
    .finally(() => {
      usageLoadingByAccountId.value = { ...usageLoadingByAccountId.value, [key]: false }
    })
}

const handleUsageLoaded = (accountId: number, usage: AccountUsageInfo) => {
  const key = String(accountId)
  usageByAccountId.value = { ...usageByAccountId.value, [key]: usage }
  usageErrorByAccountId.value = { ...usageErrorByAccountId.value, [key]: null }
}

const openDetails = () => {
  showDetails.value = true
  if (overviewError.value && !overviewLoading.value) void loadOverview(false)
}

const goToAccounts = () => {
  showDetails.value = false
  void router.push('/admin/accounts')
}

onMounted(() => {
  void loadOverview(false)
})
</script>
