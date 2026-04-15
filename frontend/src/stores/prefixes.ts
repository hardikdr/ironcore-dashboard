import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Prefix, type PrefixDetail, type CreatePrefixRequest } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const usePrefixesStore = defineStore('prefixes', () => {
  const items   = ref<Prefix[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true; error.value = null
    try { items.value = await api.prefixes.list(ns) }
    catch (e: any) { error.value = e.message }
    finally { loading.value = false }
  }

  async function get(name: string): Promise<PrefixDetail> {
    return api.prefixes.get(useNamespaceStore().active, name)
  }

  async function create(body: CreatePrefixRequest) {
    await api.prefixes.create(useNamespaceStore().active, body)
    await load()
  }

  async function deletePrefix(name: string) {
    await api.prefixes.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, get, create, deletePrefix }
})
