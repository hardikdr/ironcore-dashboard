import { flushPromises, shallowMount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import DashboardLayout from './DashboardLayout.vue'
import SreDashboardLayout from './SreDashboardLayout.vue'
import { useNamespaceStore } from '@/stores/namespace'
import { useUIStore } from '@/stores/ui'
import { useSreStatusStore } from '@/stores/sre/status'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
import { useMachinesStore } from '@/stores/machines'
import { useVolumesStore } from '@/stores/volumes'
import { useNetworksStore } from '@/stores/networks'
import { useVirtualIPsStore } from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore } from '@/stores/prefixes'

vi.mock('vuetify', () => ({ useDisplay: () => ({ mobile: ref(false) }) }))
afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks(); vi.unstubAllGlobals() })

describe('console navigation lifecycle', () => {
  for (const [path, component] of [['/overview', DashboardLayout], ['/sre/landscape', SreDashboardLayout]] as const) {
    it(`does not leave polling running when leaving ${path} during initial loading`, async () => {
      vi.useFakeTimers()
      vi.stubGlobal('localStorage', { getItem: vi.fn(), setItem: vi.fn() })
      const pinia = createPinia()
      setActivePinia(pinia)
      let finish!: () => void
      const pending = new Promise<void>(resolve => { finish = resolve })
      vi.spyOn(useNamespaceStore(), 'load').mockImplementation(() => pending)
      vi.spyOn(useSreStatusStore(), 'load').mockImplementation(() => pending)
      vi.spyOn(useSreLandscapeStore(), 'load').mockResolvedValue()
      vi.spyOn(useUIStore(), 'checkHealth').mockResolvedValue()
      for (const store of [useMachinesStore(), useVolumesStore(), useNetworksStore(), useVirtualIPsStore(), useLoadBalancersStore(), usePrefixesStore()]) {
        vi.spyOn(store, 'load').mockResolvedValue()
      }
      const router = createRouter({ history: createMemoryHistory(), routes: [{ path, component: { template: '<div />' } }] })
      await router.push(path)
      const stubs = Object.fromEntries(['VApp', 'VAppBar', 'VAppBarNavIcon', 'VBtn', 'VCard', 'VCardTitle', 'VCardText', 'VChip', 'VDialog', 'VIcon', 'VList', 'VListItem', 'VMain', 'VNavigationDrawer', 'VSelect', 'VTextField', 'RouterLink', 'RouterView'].map(name => [name, true]))
      const wrapper = shallowMount(component, { global: { plugins: [pinia, router], stubs } })
      wrapper.unmount()
      finish()
      await flushPromises()
      expect(vi.getTimerCount()).toBe(0)
    })
  }
})
