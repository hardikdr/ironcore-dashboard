import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

const STORAGE_KEY = 'ironcore-active-ns'

export const useNamespaceStore = defineStore('namespace', () => {
  const namespaces = ref<string[]>([])
  const active     = ref<string>(localStorage.getItem(STORAGE_KEY) ?? 'default')

  async function load() {
    namespaces.value = await api.namespaces.list()
    if (namespaces.value.length && !namespaces.value.includes(active.value)) {
      active.value = namespaces.value[0]
      localStorage.setItem(STORAGE_KEY, active.value)
    }
  }

  function setActive(ns: string) {
    active.value = ns
    localStorage.setItem(STORAGE_KEY, ns)
  }

  return { namespaces, active, load, setActive }
})
