import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Network, type NetworkDetail, type CreateNetworkRequest } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useNetworksStore = defineStore('networks', () => {
  const items   = ref<Network[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true
    error.value   = null
    try {
      items.value = await api.networks.list(ns)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function get(name: string): Promise<NetworkDetail> {
    return api.networks.get(useNamespaceStore().active, name)
  }

  async function create(body: CreateNetworkRequest) {
    await api.networks.create(useNamespaceStore().active, body)
    await load()
  }

  async function deleteNetwork(name: string) {
    await api.networks.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, get, create, deleteNetwork }
})
