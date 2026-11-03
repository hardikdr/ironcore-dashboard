import { beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

beforeEach(() => {
  setActivePinia(createPinia())
})

class ResizeObserverStub {
  observe() {}
  unobserve() {}
  disconnect() {}
}

Object.defineProperty(globalThis, 'ResizeObserver', { value: ResizeObserverStub, configurable: true })
Object.defineProperty(window, 'matchMedia', {
  configurable: true,
  value: (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener() {},
    removeEventListener() {},
    addListener() {},
    removeListener() {},
    dispatchEvent: () => false,
  }),
})
Object.defineProperty(window, 'visualViewport', {
  configurable: true,
  value: { width: 1024, height: 768, scale: 1, addEventListener() {}, removeEventListener() {} },
})
