import { flushPromises, mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import ConsoleModeSwitch from './ConsoleModeSwitch.vue'

async function mountedAt(path: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/overview', component: { template: '<div />' } },
      { path: '/machines', component: { template: '<div />' } },
      { path: '/sre/landscape', component: { template: '<div />' } },
      { path: '/sre/vms', component: { template: '<div />' } },
    ],
  })
  await router.push(path)
  await router.isReady()
  return { router, wrapper: mount(ConsoleModeSwitch, { global: { plugins: [router] } }) }
}

describe('ConsoleModeSwitch', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', { getItem: vi.fn(), setItem: vi.fn(), removeItem: vi.fn(), clear: vi.fn() })
  })

  it('opens SRE Console without deriving mode from local storage', async () => {
    localStorage.setItem('ironcore-console-mode', 'user')
    const { router, wrapper } = await mountedAt('/machines')
    await wrapper.get('[aria-label="Open SRE Console"]').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/sre/landscape')
  })

  it('returns to User Console from an SRE deep link', async () => {
    const { router, wrapper } = await mountedAt('/sre/vms')
    await wrapper.get('[aria-label="Open User Console"]').trigger('click')
    await flushPromises()
    expect(router.currentRoute.value.fullPath).toBe('/overview')
  })
})
