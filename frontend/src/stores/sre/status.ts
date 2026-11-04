import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { sreApi, type SreMeta, type SreSourceMeta } from '@/api/sre'

export const useSreStatusStore = defineStore('sre-status', () => {
  const sources = ref<SreSourceMeta[]>([])
  const meta = ref<SreMeta | null>(null)
  const generation = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const environment = ref('all')
  const environments = computed(() => ['all', ...new Set(sources.value.map(source => source.environment).filter((value): value is string => Boolean(value)))])
  const freshness = computed(() => meta.value?.partial ? 'Partial' : sources.value.some(source => source.freshness !== 'fresh') ? 'Stale' : 'Fresh')

  async function load(signal?: AbortSignal) {
    loading.value = true
    error.value = null
    try {
      const status = await sreApi.status(signal)
      generation.value = status.data.generation
      sources.value = status.meta.sources
      meta.value = status.meta
    } catch (cause) {
      if ((cause as Error).name !== 'AbortError') error.value = cause instanceof Error ? cause.message : 'SRE status unavailable'
    } finally { loading.value = false }
  }

  function setEnvironment(value: string) { environment.value = value }
  return { sources, meta, generation, loading, error, environment, environments, freshness, load, setEnvironment }
})
