import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import ModelDistributionChart from '../ModelDistributionChart.vue'

const { getModelStats, getUserBreakdown } = vi.hoisted(() => ({
  getModelStats: vi.fn(),
  getUserBreakdown: vi.fn()
}))

vi.mock('@/api/admin/dashboard', () => ({ getModelStats, getUserBreakdown }))

const messages: Record<string, string> = {
  'admin.dashboard.modelDistribution': 'Model Distribution',
  'admin.dashboard.spendingRankingTitle': 'Account Spending Ranking',
  'admin.dashboard.apiKeySpendingRankingTitle': 'Key Spending Ranking',
  'admin.dashboard.viewModelDistribution': 'Model Distribution',
  'admin.dashboard.viewSpendingRanking': 'Account Spending Ranking',
  'admin.dashboard.viewApiKeySpendingRanking': 'Key Spending Ranking',
  'admin.dashboard.spendingRankingAccount': 'Account',
  'admin.dashboard.spendingRankingAccountFallback': 'Account #{id}',
  'admin.dashboard.spendingRankingApiKey': 'Key',
  'admin.dashboard.spendingRankingRequests': 'Requests',
  'admin.dashboard.spendingRankingAverageDuration': 'Avg Response',
  'admin.dashboard.spendingRankingTokens': 'Tokens',
  'admin.dashboard.spendingRankingSpend': 'Spend',
  'admin.dashboard.spendingRankingShare': 'Spend Share',
  'admin.dashboard.spendingRankingOther': 'Others',
  'admin.dashboard.model': 'Model',
  'admin.dashboard.requests': 'Requests',
  'admin.dashboard.tokens': 'Tokens',
  'admin.dashboard.actual': 'Actual',
  'admin.dashboard.accountCost': 'Account Cost',
  'admin.dashboard.standard': 'Standard',
  'admin.dashboard.metricTokens': 'By Tokens',
  'admin.dashboard.metricActualCost': 'By Actual Cost',
  'admin.dashboard.noDataAvailable': 'No data available',
  'admin.redeem.userPrefix': 'User #{id}',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => {
        const message = messages[key] ?? key
        return params
          ? message.replace(/\{(\w+)\}/g, (_, name: string) => String(params[name]))
          : message
      },
    }),
  }
})

vi.mock('vue-chartjs', () => ({
  Doughnut: {
    props: ['data'],
    template: '<div class="chart-data">{{ JSON.stringify(data) }}</div>',
  },
}))

describe('ModelDistributionChart', () => {
  beforeEach(() => {
    getModelStats.mockReset()
    getUserBreakdown.mockReset()
  })

  const modelStats = [
    {
      model: 'model-a',
      requests: 8,
      input_tokens: 100,
      output_tokens: 50,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
      total_tokens: 1000,
      cost: 1.5,
      actual_cost: 0.2,
      account_cost: 0.8,
    },
    {
      model: 'model-b',
      requests: 3,
      input_tokens: 40,
      output_tokens: 20,
      cache_creation_tokens: 0,
      cache_read_tokens: 0,
      total_tokens: 500,
      cost: 0.5,
      actual_cost: 1.4,
      account_cost: 0.3,
    },
  ]

  it('uses total_tokens and token ordering by default', () => {
    const wrapper = mount(ModelDistributionChart, {
      props: {
        modelStats,
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    expect(chartData.labels).toEqual(['model-a', 'model-b'])
    expect(chartData.datasets[0].data).toEqual([1000, 500])

    const rows = wrapper.findAll('tbody tr')
    expect(rows[0].text()).toContain('model-a')
    expect(rows[1].text()).toContain('model-b')

    const options = (wrapper.vm as any).$?.setupState.doughnutOptions
    const label = options.plugins.tooltip.callbacks.label({
      label: 'model-a',
      raw: 1000,
      dataset: { data: [1000, 500] },
    })
    expect(label).toBe('model-a: 1.00K (66.7%)')
  })

  it('uses actual_cost and reorders rows in actual cost mode', () => {
    const wrapper = mount(ModelDistributionChart, {
      props: {
        modelStats,
        metric: 'actual_cost',
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    expect(chartData.labels).toEqual(['model-b', 'model-a'])
    expect(chartData.datasets[0].data).toEqual([1.4, 0.2])

    const rows = wrapper.findAll('tbody tr')
    expect(rows[0].text()).toContain('model-b')
    expect(rows[1].text()).toContain('model-a')

    const options = (wrapper.vm as any).$?.setupState.doughnutOptions
    const label = options.plugins.tooltip.callbacks.label({
      label: 'model-b',
      raw: 1.4,
      dataset: { data: [1.4, 0.2] },
    })
    expect(label).toBe('model-b: $1.40 (87.5%)')
  })

  it('can hide account cost for user usage stats without account_cost', () => {
    const wrapper = mount(ModelDistributionChart, {
      props: {
        modelStats,
        showAccountCost: false,
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    expect(wrapper.text()).not.toContain('Account Cost')
    expect(wrapper.findAll('thead th')).toHaveLength(5)
    expect(wrapper.findAll('tbody tr')[0].findAll('td')).toHaveLength(5)
  })

  it('renders upstream account spend and Others with a dedicated chart color', async () => {
    const wrapper = mount(ModelDistributionChart, {
      props: {
        modelStats: [],
        enableRankingView: true,
        rankingItems: [
          { account_id: 1, account_name: 'primary', platform: 'openai', account_cost: 12, requests: 10, tokens: 1000 },
          { account_id: 2, account_name: 'backup', platform: 'anthropic', account_cost: 8, requests: 6, tokens: 600 },
          { account_id: 3, account_name: '', platform: '', account_cost: 0, requests: 0, tokens: 0 },
        ],
        rankingTotalCost: 30,
        rankingTotalRequests: 20,
        rankingTotalTokens: 2000,
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    const rankingButton = wrapper.findAll('button').find((button) => button.text() === 'Account Spending Ranking')
    expect(rankingButton).toBeTruthy()
    await rankingButton!.trigger('click')

    const chartData = JSON.parse(wrapper.find('.chart-data').text())
    expect(chartData.labels).toEqual([
      '#1 primary (openai)',
      '#2 backup (anthropic)',
      '#3 Account #3',
      'Others',
    ])
    expect(chartData.datasets[0].data).toEqual([12, 8, 0, 10])
    expect(chartData.datasets[0].backgroundColor[0]).toBe('#3b82f6')
    expect(chartData.datasets[0].backgroundColor[3]).toBe('#94a3b8')
    expect(chartData.datasets[0].backgroundColor[3]).not.toBe(chartData.datasets[0].backgroundColor[0])

    const rows = wrapper.findAll('tbody tr')
    expect(rows).toHaveLength(4)
    expect(rows[0].text()).toContain('primary (openai)')
    expect(rows[1].text()).toContain('backup (anthropic)')
    expect(rows[2].text()).toContain('Account #3')
    expect(rows[3].text()).toContain('Others')
    expect(rows[3].text()).toContain('4')
    expect(rows[3].text()).toContain('400')
    expect(rows[3].text()).toContain('$10.00')

    await rows[0].trigger('click')
    expect(wrapper.emitted('ranking-click')?.[0]?.[0]).toEqual(expect.objectContaining({ account_id: 1 }))
  })

  it('selects the top key and shows its model distribution beside the ranking', async () => {
    getModelStats.mockImplementation(async ({ api_key_id }: { api_key_id: number }) => ({
      models: api_key_id === 9
        ? [
            { model: 'gpt-5', requests: 3, input_tokens: 100, output_tokens: 50, cache_creation_tokens: 0, cache_read_tokens: 0, total_tokens: 450, cost: 2, actual_cost: 4, average_duration_ms: 2500 },
            { model: 'gpt-4.1', requests: 1, input_tokens: 20, output_tokens: 30, cache_creation_tokens: 0, cache_read_tokens: 0, total_tokens: 50, cost: 1, actual_cost: 1, average_duration_ms: 800 }
          ]
        : [
            { model: 'claude-sonnet', requests: 2, input_tokens: 40, output_tokens: 60, cache_creation_tokens: 0, cache_read_tokens: 0, total_tokens: 100, cost: 1, actual_cost: 2, average_duration_ms: 1200 }
          ]
    }))
    const wrapper = mount(ModelDistributionChart, {
      props: {
        modelStats: [],
        enableRankingView: true,
        defaultRankingView: 'api_key_spending_ranking',
        wideRankingLayout: true,
        startDate: '2026-08-13',
        endDate: '2026-08-13',
        apiKeyRankingItems: [
          {
            api_key_id: 9,
            key_name: 'sales-key',
            user_id: 1,
            email: 'owner@example.com',
            actual_cost: 5,
            requests: 4,
            tokens: 500,
            average_duration_ms: 1500,
          },
          {
            api_key_id: 10,
            key_name: 'support-key',
            user_id: 1,
            email: 'owner@example.com',
            actual_cost: 2,
            requests: 2,
            tokens: 100,
            average_duration_ms: 1200,
          },
        ],
        apiKeyRankingTotalActualCost: 7,
        apiKeyRankingTotalRequests: 6,
        apiKeyRankingTotalTokens: 600,
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    expect(wrapper.find('h3').text()).toBe('Key Spending Ranking')
    expect(wrapper.text()).toContain('sales-key')
    expect(wrapper.text()).not.toContain('owner@example.com')
    expect(wrapper.text()).toContain('1.50s')
    const rankingChartData = JSON.parse(wrapper.find('.chart-data').text())
    expect(rankingChartData.labels).toEqual(['#1 sales-key', '#2 support-key'])
    expect(rankingChartData.datasets[0].data).toEqual([5, 2])

    await flushPromises()

    expect(getModelStats).toHaveBeenCalledWith(expect.objectContaining({
      api_key_id: 9,
      start_date: '2026-08-13',
      end_date: '2026-08-13',
      model_source: 'requested'
    }))
    expect(wrapper.text()).toContain('gpt-5')
    expect(wrapper.text()).toContain('gpt-4.1')
    expect(wrapper.text()).toContain('2.50s')
    expect(wrapper.text()).toContain('800ms')
    expect(wrapper.text()).toContain('80.0%')
    expect(wrapper.get('tbody button').attributes('aria-expanded')).toBe('true')
    expect(wrapper.find('tr.lg\\:hidden').exists()).toBe(true)
    expect(wrapper.find('tr.lg\\:hidden table').exists()).toBe(false)

    await wrapper.findAll('tbody button')[1].trigger('click')
    await flushPromises()

    expect(getModelStats).toHaveBeenLastCalledWith(expect.objectContaining({ api_key_id: 10 }))
    expect(wrapper.text()).toContain('claude-sonnet')
  })

  it('keeps the key ranking at four columns until the chart is wide', () => {
    const wrapper = mount(ModelDistributionChart, {
      props: {
        modelStats: [],
        enableRankingView: true,
        defaultRankingView: 'api_key_spending_ranking',
        apiKeyRankingItems: [
          { api_key_id: 9, key_name: 'sales-key', user_id: 1, email: 'owner@example.com', actual_cost: 5, requests: 4, tokens: 500 }
        ],
        apiKeyRankingTotalActualCost: 5,
        apiKeyRankingTotalRequests: 4,
        apiKeyRankingTotalTokens: 500
      },
      global: { stubs: { LoadingSpinner: true } }
    })

    expect(wrapper.findAll('thead').at(0)!.findAll('th')).toHaveLength(4)
    expect(wrapper.findAll('thead').at(0)!.text()).not.toContain('Owner')
    expect(wrapper.findAll('thead').at(0)!.text()).not.toContain('Avg Response')
    expect(wrapper.findAll('thead').at(0)!.text()).not.toContain('Spend Share')
  })

})
