import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

import type { DashboardStats } from '@/types'
import DashboardView from '../DashboardView.vue'

const { getSnapshotV2, getApiKeyUsageTrend, getAccountSpendingRanking, getApiKeySpendingRanking } = vi.hoisted(() => ({
  getSnapshotV2: vi.fn(),
  getApiKeyUsageTrend: vi.fn(),
  getAccountSpendingRanking: vi.fn(),
  getApiKeySpendingRanking: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    dashboard: {
      getSnapshotV2,
      getApiKeyUsageTrend,
      getAccountSpendingRanking,
      getApiKeySpendingRanking
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn()
  })
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({
    push: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

const formatLocalDate = (date: Date): string => {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const createDashboardStats = (): DashboardStats => ({
  total_users: 0,
  today_new_users: 0,
  active_users: 0,
  hourly_active_users: 0,
  stats_updated_at: '',
  stats_stale: false,
  total_api_keys: 0,
  active_api_keys: 0,
  total_accounts: 0,
  normal_accounts: 0,
  error_accounts: 0,
  ratelimit_accounts: 0,
  overload_accounts: 0,
  total_requests: 0,
  total_input_tokens: 0,
  total_output_tokens: 0,
  total_cache_creation_tokens: 0,
  total_cache_read_tokens: 0,
  total_tokens: 0,
  total_cost: 0,
  total_actual_cost: 0,
  total_account_cost: 0,
  today_requests: 0,
  today_input_tokens: 0,
  today_output_tokens: 0,
  today_cache_creation_tokens: 0,
  today_cache_read_tokens: 0,
  today_tokens: 0,
  today_cost: 0,
  today_actual_cost: 0,
  today_account_cost: 0,
  average_duration_ms: 0,
  uptime: 0,
  rpm: 0,
  tpm: 0
})

describe('admin DashboardView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())

    getSnapshotV2.mockReset()
    getApiKeyUsageTrend.mockReset()
    getAccountSpendingRanking.mockReset()
    getApiKeySpendingRanking.mockReset()

    getSnapshotV2.mockResolvedValue({
      stats: createDashboardStats(),
      trend: [],
      models: []
    })
    getApiKeyUsageTrend.mockResolvedValue({
      trend: [],
      start_date: '',
      end_date: '',
      granularity: 'hour'
    })
    getAccountSpendingRanking.mockResolvedValue({
      ranking: [],
      total_account_cost: 0,
      total_requests: 0,
      total_tokens: 0,
      start_date: '',
      end_date: ''
    })
    getApiKeySpendingRanking.mockResolvedValue({
      ranking: [],
      total_actual_cost: 0,
      total_requests: 0,
      total_tokens: 0,
      start_date: '',
      end_date: ''
    })
  })

  it('uses today as default dashboard range', async () => {
    mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          AccountUsageOverviewCard: true,
          Line: true
        }
      }
    })

    await flushPromises()

    const now = new Date()

    expect(getSnapshotV2).toHaveBeenCalledTimes(1)
    expect(getSnapshotV2).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(now),
      end_date: formatLocalDate(now),
      granularity: 'hour'
    }))
    expect(getApiKeyUsageTrend).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(now),
      end_date: formatLocalDate(now),
      granularity: 'hour',
      limit: 12
    }))
    expect(getAccountSpendingRanking).toHaveBeenCalledWith(expect.objectContaining({
      start_date: formatLocalDate(now),
      end_date: formatLocalDate(now),
      limit: 12
    }))
  })

  it('keeps the token usage trend below the full-width key usage trend', async () => {
    const wrapper = mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          Icon: true,
          DateRangePicker: true,
          Select: true,
          ModelDistributionChart: true,
          TokenUsageTrend: true,
          AccountUsageOverviewCard: true,
          Line: true
        }
      }
    })

    await flushPromises()

    expect(wrapper.find('[data-testid="toggle-token-usage-trend"]').exists()).toBe(false)
    expect(wrapper.find('token-usage-trend-stub').exists()).toBe(true)
    const distributionCharts = wrapper.findAllComponents({ name: 'ModelDistributionChart' })
    expect(distributionCharts).toHaveLength(1)
    expect(distributionCharts[0].props('wideRankingLayout')).toBe(false)
    expect(wrapper.html().indexOf('admin.dashboard.apiKeyUsageTrend')).toBeLessThan(
      wrapper.html().indexOf('token-usage-trend-stub')
    )
  })
})
