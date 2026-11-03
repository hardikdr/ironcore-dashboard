import { createPinia, setActivePinia } from 'pinia'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useSreSearchStore } from './search'

describe('SRE search store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  it('uses server-backed GET search and keeps matched-field metadata', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(JSON.stringify({
      data: [{ id: 'opaque', resource: { type: 'ironcore-vm', provenance: { name: 'vm-a' } }, matchedField: 'uid', matchKind: 'exact', fragment: 'vm-uid', ranges: [{ start: 0, end: 6 }] }],
      meta: { generatedAt: '2026-11-03T12:00:00Z', sources: [], partial: false, warnings: [] },
    }), { status: 200, headers: { 'Content-Type': 'application/json' } }))
    const store = useSreSearchStore()

    await store.run('vm-uid')

    expect(fetchMock).toHaveBeenCalledOnce()
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe('/api/sre/v1/search?q=vm-uid&limit=25')
    expect((init as RequestInit).method).toBeUndefined()
    expect(store.results[0]).toMatchObject({ id: 'opaque', matchedField: 'uid', matchKind: 'exact' })
  })
})
