import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import AccountUsageOverviewCard from '../AccountUsageOverviewCard.vue'

const { list, getBatchUsage, push } = vi.hoisted(() => ({
  list: vi.fn(),
  getBatchUsage: vi.fn(),
  push: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: { list, getBatchUsage }
  }
}))

vi.mock('vue-router', () => ({
  useRouter: () => ({ push })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) => params?.count == null ? key : `${key}:${params.count}`
    })
  }
})

const accounts = [
  {
    id: 1,
    name: 'primary@example.com',
    platform: 'openai',
    type: 'oauth',
    status: 'active',
    schedulable: true
  },
  {
    id: 2,
    name: 'backup@example.com',
    platform: 'openai',
    type: 'oauth',
    status: 'active',
    schedulable: false
  }
]

const usage = {
  '1': {
    updated_at: null,
    five_hour: { utilization: 42, resets_at: null, remaining_seconds: 0 },
    seven_day: { utilization: 85, resets_at: null, remaining_seconds: 0 },
    seven_day_sonnet: null
  },
  '2': {
    updated_at: null,
    five_hour: { utilization: 18, resets_at: null, remaining_seconds: 0 },
    seven_day: { utilization: 25, resets_at: null, remaining_seconds: 0 },
    seven_day_sonnet: null
  }
}

describe('AccountUsageOverviewCard', () => {
  beforeEach(() => {
    list.mockReset()
    getBatchUsage.mockReset()
    push.mockReset()
    list.mockResolvedValue({ items: accounts, total: 2, page: 1, page_size: 1000, pages: 1 })
    getBatchUsage.mockResolvedValue({ usage, errors: {} })
  })

  it('summarizes the highest window and opens reusable account usage rows', async () => {
    const wrapper = mount(AccountUsageOverviewCard, {
      global: {
        stubs: {
          BaseDialog: {
            props: ['show', 'title'],
            template: '<div v-if="show" data-testid="usage-dialog"><slot /><slot name="footer" /></div>'
          },
          AccountUsageCell: {
            props: ['account'],
            template: '<div class="usage-cell">{{ account.name }}</div>'
          },
          UsageProgressBar: {
            props: ['label', 'utilization'],
            template: '<div class="usage-progress">{{ label }} {{ utilization }}%</div>'
          },
          LoadingSpinner: true,
          Icon: true
        }
      }
    })

    await flushPromises()

    expect(list).toHaveBeenCalledWith(1, 1000, expect.objectContaining({ status: 'active' }))
    expect(getBatchUsage).toHaveBeenCalledWith([1, 2], false)
    expect(wrapper.get('[data-testid="account-usage-card"]').text()).toContain('primary@example.com')
    expect(wrapper.get('[data-testid="account-usage-card"]').text()).toContain('7d 85%')
    expect(wrapper.get('[data-testid="account-usage-card"]').text()).toContain('backup@example.com')
    expect(wrapper.get('[data-testid="account-usage-card"]').text()).toContain('7d 25%')

    await wrapper.get('[data-testid="account-usage-card"]').trigger('click')
    expect(wrapper.get('[data-testid="usage-dialog"]').text()).toContain('primary@example.com')
    expect(wrapper.get('[data-testid="usage-dialog"]').text()).toContain('backup@example.com')

    await wrapper.get('[data-testid="refresh-account-usage"]').trigger('click')
    await flushPromises()
    expect(getBatchUsage).toHaveBeenLastCalledWith([1, 2], true)
  })
})
