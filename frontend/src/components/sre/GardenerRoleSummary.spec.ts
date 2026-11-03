import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import GardenerRoleSummary from './GardenerRoleSummary.vue'

describe('GardenerRoleSummary', () => {
  it('does not fabricate an IronCore VM path for the wdf4ai bare-metal installation', () => {
    const wrapper = mount(GardenerRoleSummary, { props: { installationRole: 'bare-metal', providerType: 'bare-metal', workerVms: 0 } })
    expect(wrapper.text()).toContain('Bare metal')
    expect(wrapper.text()).toContain('no IronCore VM path')
    expect(wrapper.find('a[href*="/sre/vms"]').exists()).toBe(false)
  })

  it('preserves installation role and provider type for virtual-machine installations', () => {
    const wrapper = mount(GardenerRoleSummary, { props: { installationRole: 'runtime', providerType: 'ironcore', workerVms: 3 } })
    expect(wrapper.text()).toContain('runtime')
    expect(wrapper.text()).toContain('ironcore')
    expect(wrapper.text()).toContain('3 worker VMs')
  })
})
