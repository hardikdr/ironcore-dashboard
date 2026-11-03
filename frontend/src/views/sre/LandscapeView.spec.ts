import { mount, RouterLinkStub } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it } from 'vitest'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
import LandscapeView from './LandscapeView.vue'

describe('LandscapeView', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('links aggregate counts to the corresponding operational inventory', () => {
    const store = useSreLandscapeStore()
    store.summary = { IronCoreVM: 42, Zone: 3, Shoot: 7, Volume: 9 }
    store.meta = { generatedAt: '2026-11-03T12:00:00Z', sources: [], partial: true, warnings: ['one source denied'], }
    const passthrough = { template: '<div><slot /></div>' }
    const wrapper = mount(LandscapeView, { global: { stubs: {
      RouterLink: RouterLinkStub,
      AppPage: { template: '<main><slot /><slot name="actions" /></main>' },
      VAlert: { props: ['title'], template: '<aside>{{ title }}<slot /></aside>' },
      SnapshotStamp: { template: '<span>snapshot</span>' },
      VIcon: passthrough, VCard: passthrough, VCardTitle: passthrough, VCardSubtitle: passthrough,
      VCardText: passthrough, VRow: passthrough, VCol: passthrough, VBtn: passthrough,
    } } })
    expect(wrapper.text()).toContain('42')
    expect(wrapper.text()).toContain('Partial landscape')
    expect(wrapper.findAllComponents(RouterLinkStub).map(link => link.props('to'))).toContain('/sre/vms')
  })
})
