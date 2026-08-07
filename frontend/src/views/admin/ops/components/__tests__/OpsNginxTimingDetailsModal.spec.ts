import { describe, expect, it, vi } from 'vitest'
import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import OpsNginxTimingDetailsModal from '../OpsNginxTimingDetailsModal.vue'
import type { OpsNginxTimingKeyDetails, OpsNginxTimingKeyMetric } from '@/api/admin/ops'

const mockGetNginxTimingKeyDetails = vi.fn()

vi.mock('@/api/admin/ops', () => ({
  opsAPI: {
    getNginxTimingKeyDetails: (...args: unknown[]) => mockGetNginxTimingKeyDetails(...args),
  },
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const BaseDialogStub = defineComponent({
  props: { show: Boolean },
  template: '<div v-if="show"><slot /></div>',
})

function item(id: number, keyName: string, p99: number | null, p50: number | null): OpsNginxTimingKeyMetric {
  return {
    api_key_id: id,
    key_name: keyName,
    http_request_count: id * 10,
    websocket_session_count: 0,
    success_count: id * 10,
    client_timeout_408_count: 0,
    client_closed_499_count: 0,
    server_error_5xx_count: 0,
    upstream_unreached_count: 0,
    request_time: { p99_ms: p99, p50_ms: p50 },
    upstream_connect_time: {},
    upstream_header_time: {},
    upstream_response_time: {},
    client_overhead_time: {},
  }
}

const response: OpsNginxTimingKeyDetails = {
  available: true,
  start_time: '2026-08-07T00:00:00Z',
  end_time: '2026-08-07T01:00:00Z',
  window_clamped: false,
  key_filter_applied: false,
  matched_request_count: 3,
  unattributed_error_count: 0,
  items: [
    item(1, 'Beta', 300, 100),
    item(2, 'Alpha', 100, 900),
    item(3, 'Gamma', null, null),
  ],
}

describe('OpsNginxTimingDetailsModal', () => {
  it('sorts duration headers and keeps missing values last', async () => {
    mockGetNginxTimingKeyDetails.mockResolvedValue(response)
    const wrapper = mount(OpsNginxTimingDetailsModal, {
      props: { modelValue: false, metric: 'request_time', timeRange: '1h' },
      global: { stubs: { BaseDialog: BaseDialogStub, Icon: true } },
    })

    await wrapper.setProps({ modelValue: true })
    await flushPromises()

    const rowKeys = () => wrapper.findAll('tbody tr').map((row) => row.find('td').text())
    const p50Header = wrapper.findAll('thead button').find((button) => button.text().includes('P50'))

    expect(wrapper.findAll('thead button')).toHaveLength(8)
    expect(rowKeys()).toEqual(['Beta', 'Alpha', 'Gamma'])
    expect(p50Header).toBeDefined()

    await p50Header!.trigger('click')
    expect(rowKeys()).toEqual(['Alpha', 'Beta', 'Gamma'])

    await p50Header!.trigger('click')
    expect(rowKeys()).toEqual(['Beta', 'Alpha', 'Gamma'])
  })
})
