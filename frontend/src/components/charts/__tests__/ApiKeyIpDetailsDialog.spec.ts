import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import ApiKeyIpDetailsDialog from '../ApiKeyIpDetailsDialog.vue'

const { getApiKeyIPOverlaps } = vi.hoisted(() => ({ getApiKeyIPOverlaps: vi.fn() }))

vi.mock('@/api/admin/dashboard', () => ({ getApiKeyIPOverlaps }))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

describe('ApiKeyIpDetailsDialog', () => {
  it('shows full IP addresses and exact overlap interval', async () => {
    getApiKeyIPOverlaps.mockResolvedValue({ items: [{
      ip_a: '203.0.113.8',
      ip_b: '198.51.100.2',
      overlap_start_at: '2025-01-01T00:00:10Z',
      overlap_end_at: '2025-01-01T00:00:20Z',
      overlap_seconds: 10
    }] })
    const wrapper = mount(ApiKeyIpDetailsDialog, {
      props: {
        show: true,
        startDate: '2025-01-01',
        endDate: '2025-01-01',
        item: {
          api_key_id: 9,
          key_name: 'client-a',
          requests: 2,
          distinct_ip_count: 2,
          active_ip_count_15m: 2,
          overlap_ip_count_15m: 2,
          overlap_count_15m: 1,
          total_overlap_seconds_15m: 10,
          max_overlap_seconds_15m: 10,
          last_overlap_at: '2025-01-01T00:00:20Z',
          risk_level: 'watch' as const,
          ip_usages: []
        }
      },
      global: {
        stubs: {
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
          IpGeoBatchToolbar: true,
          IpGeoCell: true
        }
      }
    })
    await flushPromises()

    expect(getApiKeyIPOverlaps).toHaveBeenCalledWith(9, { limit: 50, start_date: '2025-01-01', end_date: '2025-01-01' })
    expect(wrapper.text()).toContain('203.0.113.8')
    expect(wrapper.text()).toContain('198.51.100.2')
    expect(wrapper.text()).toContain('10.0s')
  })
})
