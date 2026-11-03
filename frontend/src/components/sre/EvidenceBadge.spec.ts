import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import EvidenceBadge from './EvidenceBadge.vue'

describe('EvidenceBadge', () => {
  it('opens a fact-specific explanation instead of presenting a relation as unexplained truth', async () => {
    const fact = { label: 'Physical host', value: 'node-a', sourceId: 'zone-a', evidence: 'provider-id', confidence: 'proven' }
    const wrapper = mount(EvidenceBadge, { props: { fact }, global: { stubs: { VChip: { template: '<button type="button"><slot /></button>' }, VIcon: true } } })
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('explain')?.[0]).toEqual([fact])
    expect(wrapper.text()).toContain('proven')
  })

  it('labels missing evidence as unknown', () => {
    const wrapper = mount(EvidenceBadge, { props: { fact: { label: 'Shoot', value: 'unknown', confidence: 'unknown' } }, global: { stubs: { VChip: { template: '<span><slot /></span>' }, VIcon: true } } })
    expect(wrapper.text()).toContain('unknown')
  })
})
