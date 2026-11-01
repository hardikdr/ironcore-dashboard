import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'
import { useToastStore } from '@/stores/toast'

const STORAGE_KEY = 'ironcore-active-ns'

export const useNamespaceStore = defineStore('namespace', () => {
  const namespaces    = ref<string[]>([])
  const active        = ref<string>(localStorage.getItem(STORAGE_KEY) ?? 'default')
  const accessChecking = ref<boolean>(false)

  async function load() {
    namespaces.value = await api.namespaces.list()
    if (namespaces.value.length && !namespaces.value.includes(active.value)) {
      active.value = namespaces.value[0]
      localStorage.setItem(STORAGE_KEY, active.value)
    }
  }

  async function setActive(ns: string) {
    accessChecking.value = true
    try {
      const result = await api.auth.checkNamespaceAccess(ns)
      if (!result.allowed) {
        const toast = useToastStore()
        toast.error(`You don't have access to namespace '${ns}'`)
        return
      }
    } catch {
      // Fail open: if the check errors (demo mode, network issue), proceed with the switch.
    } finally {
      accessChecking.value = false
    }
    active.value = ns
    localStorage.setItem(STORAGE_KEY, ns)
  }

  return { namespaces, active, accessChecking, load, setActive }
})
