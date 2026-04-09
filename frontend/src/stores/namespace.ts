import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export const useNamespaceStore = defineStore('namespace', () => {
  const namespaces = ref<string[]>([])
  const active     = ref<string>('default')

  async function load() {
    namespaces.value = await api.namespaces.list()
    if (namespaces.value.length && !namespaces.value.includes(active.value)) {
      active.value = namespaces.value[0]
    }
  }

  function setActive(ns: string) { active.value = ns }

  return { namespaces, active, load, setActive }
})
