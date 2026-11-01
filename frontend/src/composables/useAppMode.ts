import { ref } from 'vue'
import { api } from '@/api/client'

const _demo = ref<boolean | null>(null)

export function useAppMode() {
  async function load() {
    if (_demo.value !== null) return
    try {
      const mode = await api.mode.get()
      _demo.value = mode.demo
    } catch {
      _demo.value = false
    }
  }

  return { isDemoMode: _demo, load }
}
