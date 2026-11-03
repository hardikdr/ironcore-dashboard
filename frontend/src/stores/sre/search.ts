import { defineStore } from 'pinia'
import { ref } from 'vue'
import { sreApi, type SreMeta, type SreSearchResult } from '@/api/sre'

export const useSreSearchStore = defineStore('sre-search', () => {
  const results = ref<SreSearchResult[]>([])
  const meta = ref<SreMeta | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function run(value: string, signal?: AbortSignal) {
    const term = value.trim()
    if (term.length < 2) { results.value = []; return }
    loading.value = true
    error.value = null
    try {
      const response = await sreApi.search(term, signal)
      results.value = response.data
      meta.value = response.meta
    } catch (cause) {
      if ((cause as Error).name !== 'AbortError') error.value = cause instanceof Error ? cause.message : 'Search failed'
    } finally { loading.value = false }
  }

  function clear() { results.value = []; error.value = null }
  return { results, meta, loading, error, run, clear }
})
