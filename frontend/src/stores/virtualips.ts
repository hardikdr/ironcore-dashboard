import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type VirtualIP } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useVirtualIPsStore = defineStore('virtualips', () => {
  const items   = ref<VirtualIP[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load(silent = false) {
    const ns = useNamespaceStore().active
    if (!silent) loading.value = true
    error.value = null
    try { items.value = await api.virtualIPs.list(ns) }
    catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
    finally { loading.value = false }
  }

  async function deleteItem(name: string) {
    await api.virtualIPs.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, deleteItem }
})
