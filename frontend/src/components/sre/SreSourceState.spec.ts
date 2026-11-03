import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import SreSourceState from './SreSourceState.vue'

describe('SreSourceState', () => {
  it.each([
    ['not-configured', 'fresh', 'Not configured'],
    ['denied', 'fresh', 'RBAC denied'],
    ['unreachable', 'fresh', 'Unreachable'],
    ['unsupported', 'fresh', 'API unsupported'],
    ['ok', 'stale', 'Stale'],
    ['ok', 'fresh', 'Healthy'],
  ])('distinguishes %s/%s as %s', (access, freshness, label) => {
    const wrapper = mount(SreSourceState, { props: { access, freshness }, global: { stubs: { VChip: { template: '<span><slot /></span>' } } } })
    expect(wrapper.text()).toContain(label)
  })
})
