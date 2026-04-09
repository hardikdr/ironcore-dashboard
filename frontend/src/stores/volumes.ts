import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Volume } from '@/api/client'
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
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function deleteVolume(name: string) {
    await api.volumes.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, deleteVolume }
})
