import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import EvidenceGraph from './EvidenceGraph.vue'

describe('EvidenceGraph', () => {
  it('renders only API-provided directed edges and always includes a table fallback', () => {
    const graph = {
      resources: [
        { key: 'a', type: 'vm', provenance: { sourceId: 's', kind: 'IronCoreVM', name: 'vm-a' } },
        { key: 'b', type: 'nic', provenance: { sourceId: 's', kind: 'NIC', name: 'nic-a' } },
      ],
      relations: [{ from: 'a', to: 'b', type: 'has-nic', evidence: 'uid-reference', fieldPath: '.spec.machineRef', confidence: 'proven' }],
      truncated: false,
    }
    const wrapper = mount(EvidenceGraph, { props: { graph, rootKey: 'a' }, global: { stubs: { TopologyRelations: { template: '<table aria-label="Relationship table"><tbody><tr><td>fallback</td></tr></tbody></table>' }, VChip: { template: '<span><slot /></span>' }, VAlert: { template: '<aside><slot /></aside>' } } } })
    expect(wrapper.findAll('[data-edge]').length).toBe(1)
    expect(wrapper.text()).toContain('vm-a')
    expect(wrapper.text()).toContain('nic-a')
    expect(wrapper.find('table[aria-label="Relationship table"]').exists()).toBe(true)
  })
})
