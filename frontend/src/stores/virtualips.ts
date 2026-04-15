import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type VirtualIP, type VirtualIPDetail, type CreateVirtualIPRequest } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useVirtualIPsStore = defineStore('virtualips', () => {
  const items   = ref<VirtualIP[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true
    error.value   = null
    try {
      items.value = await api.virtualIPs.list(ns)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function get(name: string): Promise<VirtualIPDetail> {
    return api.virtualIPs.get(useNamespaceStore().active, name)
  }

  async function create(body: CreateVirtualIPRequest) {
    await api.virtualIPs.create(useNamespaceStore().active, body)
    await load()
  }

  async function deleteVirtualIP(name: string) {
    await api.virtualIPs.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, get, create, deleteVirtualIP }
})
