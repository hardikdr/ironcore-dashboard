import { defineStore } from 'pinia'
import { ref } from 'vue'
import { sreApi, type SreMeta, type SreObservation, type SreResourceDTO } from '@/api/sre'

export const useSreLandscapeStore = defineStore('sre-landscape', () => {
  const summary = ref<Record<string, number>>({})
  const zones = ref<SreResourceDTO[]>([])
  const observations = ref<SreObservation[]>([])
  const meta = ref<SreMeta | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  async function load(environment = 'all', signal?: AbortSignal) {
    loading.value = true
    error.value = null
    try {
      const filter = environment === 'all' ? undefined : environment
      const [summaryResult, zoneResult, observationResult] = await Promise.all([
        sreApi.summary(filter, signal), sreApi.zones(filter, signal), sreApi.observations(signal),
      ])
      summary.value = summaryResult.data
      zones.value = zoneResult.data
      observations.value = observationResult.data
      meta.value = summaryResult.meta
    } catch (cause) {
      if ((cause as Error).name !== 'AbortError') error.value = cause instanceof Error ? cause.message : 'Landscape unavailable'
    } finally { loading.value = false }
  }
  return { summary, zones, observations, meta, loading, error, load }
})
