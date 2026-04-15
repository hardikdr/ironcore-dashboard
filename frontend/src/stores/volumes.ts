import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Volume, type VolumeDetail, type CreateVolumeRequest } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useVolumesStore = defineStore('volumes', () => {
  const items   = ref<Volume[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true
    error.value   = null
    try {
      items.value = await api.volumes.list(ns)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function get(name: string): Promise<VolumeDetail> {
    return api.volumes.get(useNamespaceStore().active, name)
  }

  async function create(body: CreateVolumeRequest) {
    await api.volumes.create(useNamespaceStore().active, body)
    await load()
  }

  async function deleteVolume(name: string) {
    await api.volumes.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, get, create, deleteVolume }
})
