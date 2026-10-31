import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type VirtualIP } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useVirtualIPsStore = defineStore('virtualips', () => {
  const items   = ref<VirtualIP[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true; error.value = null
    try { items.value = await api.virtualIPs.list(ns) }
    catch (e: any) { error.value = e.message }
    finally { loading.value = false }
  }

  async function deleteItem(name: string) {
    await api.virtualIPs.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, deleteItem }
})
