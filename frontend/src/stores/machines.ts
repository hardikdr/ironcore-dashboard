import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Machine } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useMachinesStore = defineStore('machines', () => {
  const items   = ref<Machine[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load(silent = false) {
    const ns = useNamespaceStore().active
    if (!silent) loading.value = true
    error.value = null
    try {
      items.value = await api.machines.list(ns)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function deleteMachine(name: string) {
    await api.machines.delete(useNamespaceStore().active, name)
    await load()
  }

  async function setPower(name: string, power: 'On' | 'Off') {
    await api.machines.power(useNamespaceStore().active, name, power)
    await load()
  }

  return { items, loading, error, load, deleteMachine, setPower }
})
