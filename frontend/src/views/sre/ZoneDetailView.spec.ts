import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import { sreApi } from '@/api/sre'
import ZoneDetailView from './ZoneDetailView.vue'
vi.mock('vue-router', () => ({ useRoute: () => ({ params: { zone: 'wdf4a' } }) }))
describe('zone drill-down', () => {
  it('loads only detail and bounded graph, tolerates absent gaps and filters locally', async () => {
    const zone = { key: 'zone-a', type: 'zone', provenance: { name: 'wdf4a', kind: 'Zone', sourceId: 'local' } }
    vi.spyOn(sreApi, 'zone').mockResolvedValue({ data: { zone, gaps: null } } as any)
    const list = vi.spyOn(sreApi, 'zones')
    const graph = vi.spyOn(sreApi, 'graph').mockResolvedValue({ data: { resources: [zone, { key: 'node-a', type: 'physical-node', provenance: { name: 'worker-a', kind: 'Node', sourceId: 'local' } }], relations: [], truncated: false } } as any)
    const passthrough = { template: '<div><slot /></div>' }
    const wrapper = mount(ZoneDetailView, { global: { stubs: {
      AppPage: passthrough, TopologyRelations: true, RouterLink: RouterLinkStub,
      VAlert: passthrough, VCard: passthrough, VCardText: passthrough, VCardTitle: passthrough, VSkeletonLoader: true,
      VTextField: { props: ['modelValue'], template: '<input @input="$emit(\'update:modelValue\', $event.target.value)" />' },
    } } })
    await flushPromises()
    expect(wrapper.text()).toContain('worker-a')
    expect(list).not.toHaveBeenCalled()
    expect(graph).toHaveBeenCalledTimes(1)
    await wrapper.get('input').setValue('missing')
    expect(wrapper.text()).toContain('No resources match')
    expect(graph).toHaveBeenCalledTimes(1)
    wrapper.unmount()
    vi.restoreAllMocks()
  })
})
