import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import OpsRequestDetailsModal from '../OpsRequestDetailsModal.vue'

const { listRequestDetails } = vi.hoisted(() => ({
  listRequestDetails: vi.fn()
}))

vi.mock('@/api/admin/ops', () => ({
  opsAPI: { listRequestDetails }
}))

vi.mock('@/stores', () => ({
  useAppStore: () => ({ showError: vi.fn(), showWarning: vi.fn() })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({ copyToClipboard: vi.fn() })
}))

vi.mock('@vueuse/core', () => ({
  useMediaQuery: () => ({ value: true })
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return { ...actual, useI18n: () => ({ t: (key: string) => key }) }
})

const BaseDialogStub = defineComponent({
  props: { show: Boolean },
  template: '<div v-if="show"><slot /></div>'
})

describe('OpsRequestDetailsModal', () => {
  it('shows the Key name with an ID fallback instead of platform', async () => {
    listRequestDetails.mockResolvedValue({
      items: [
        { kind: 'success', created_at: '2026-08-13T00:00:00Z', request_id: 'req-1', platform: 'openai', api_key_id: 1, api_key_name: '王唯迪' },
        { kind: 'success', created_at: '2026-08-13T00:00:01Z', request_id: 'req-2', platform: 'openai', api_key_id: 2 }
      ],
      total: 2,
      page: 1,
      page_size: 10
    })

    const wrapper = mount(OpsRequestDetailsModal, {
      props: { modelValue: false, timeRange: '1h', preset: { title: '请求明细' } },
      global: { stubs: { BaseDialog: BaseDialogStub, Pagination: true } }
    })
    await wrapper.setProps({ modelValue: true })
    await flushPromises()

    expect(wrapper.text()).toContain('admin.ops.requestDetails.table.apiKey')
    expect(wrapper.text()).toContain('王唯迪')
    expect(wrapper.text()).toContain('Key #2')
    expect(wrapper.text()).not.toContain('OPENAI')
  })
})
