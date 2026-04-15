import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type LoadBalancer, type LoadBalancerDetail, type CreateLoadBalancerRequest } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useLoadBalancersStore = defineStore('loadbalancers', () => {
  const items   = ref<LoadBalancer[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true
    error.value   = null
    try {
      items.value = await api.loadBalancers.list(ns)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : String(e)
    } finally {
      loading.value = false
    }
  }

  async function get(name: string): Promise<LoadBalancerDetail> {
    return api.loadBalancers.get(useNamespaceStore().active, name)
  }

  async function create(body: CreateLoadBalancerRequest) {
    await api.loadBalancers.create(useNamespaceStore().active, body)
    await load()
  }

  async function deleteLoadBalancer(name: string) {
    await api.loadBalancers.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, get, create, deleteLoadBalancer }
})
