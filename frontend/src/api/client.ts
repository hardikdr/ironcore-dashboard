const BASE = '/api/v1'

const TOKEN_STORAGE_KEY = 'ironcore-dashboard-token'

/** Returns the bearer token stored in localStorage, or null if not set. */
export function getAuthToken(): string | null {
  return localStorage.getItem(TOKEN_STORAGE_KEY)
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  const token = getAuthToken()
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  const res = await fetch(`${BASE}${path}`, {
    headers,
    ...options
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  namespaces: {
    list: () => request<string[]>('/namespaces')
  },
  machines: {
    list:   (ns: string) => request<Machine[]>(`/namespaces/${ns}/machines`),
    get:    (ns: string, name: string) => request<Machine>(`/namespaces/${ns}/machines/${name}`),
    create: (ns: string, body: CreateMachineRequest) =>
      request<Machine>(`/namespaces/${ns}/machines`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/machines/${name}`, { method: 'DELETE' }),
    power:  (ns: string, name: string, power: 'On'|'Off') =>
      request<Machine>(`/namespaces/${ns}/machines/${name}/power`, {
        method: 'PATCH', body: JSON.stringify({ power })
      })
  },
  machineClasses: {
    list: () => request<MachineClass[]>('/machineclasses')
  },
  volumeClasses: {
    list: () => request<VolumeClass[]>('/volumeclasses')
  },
  volumes: {
    list:   (ns: string) => request<Volume[]>(`/namespaces/${ns}/volumes`),
    get:    (ns: string, name: string) => request<VolumeDetail>(`/namespaces/${ns}/volumes/${name}`),
    create: (ns: string, body: CreateVolumeRequest) =>
      request<Volume>(`/namespaces/${ns}/volumes`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/volumes/${name}`, { method: 'DELETE' })
  },
  networks: {
    list:           (ns: string) => request<Network[]>(`/namespaces/${ns}/networks`),
    get:            (ns: string, name: string) => request<NetworkDetail>(`/namespaces/${ns}/networks/${name}`),
    create:         (ns: string, body: CreateNetworkRequest) =>
      request<Network>(`/namespaces/${ns}/networks`, { method: 'POST', body: JSON.stringify(body) }),
    delete:         (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/networks/${name}`, { method: 'DELETE' }),
    listInterfaces: (ns: string, filters?: { machine?: string; network?: string }) => {
      const params = new URLSearchParams()
      if (filters?.machine) params.set('machine', filters.machine)
      if (filters?.network) params.set('network', filters.network)
      const qs = params.toString() ? `?${params.toString()}` : ''
      return request<NetworkInterface[]>(`/namespaces/${ns}/networkinterfaces${qs}`)
    }
  },
  virtualIPs: {
    list:   (ns: string) => request<VirtualIP[]>(`/namespaces/${ns}/virtualips`),
    get:    (ns: string, name: string) => request<VirtualIPDetail>(`/namespaces/${ns}/virtualips/${name}`),
    create: (ns: string, body: CreateVirtualIPRequest) =>
      request<VirtualIP>(`/namespaces/${ns}/virtualips`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/virtualips/${name}`, { method: 'DELETE' })
  },
  loadBalancers: {
    list:   (ns: string) => request<LoadBalancer[]>(`/namespaces/${ns}/loadbalancers`),
    get:    (ns: string, name: string) => request<LoadBalancerDetail>(`/namespaces/${ns}/loadbalancers/${name}`),
    create: (ns: string, body: CreateLoadBalancerRequest) =>
      request<LoadBalancer>(`/namespaces/${ns}/loadbalancers`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/loadbalancers/${name}`, { method: 'DELETE' })
  },
  prefixes: {
    list:   (ns: string) => request<Prefix[]>(`/namespaces/${ns}/prefixes`),
    get:    (ns: string, name: string) => request<PrefixDetail>(`/namespaces/${ns}/prefixes/${name}`),
    create: (ns: string, body: CreatePrefixRequest) =>
      request<Prefix>(`/namespaces/${ns}/prefixes`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/prefixes/${name}`, { method: 'DELETE' }),
    listAllocations: (ns: string) => request<PrefixAllocation[]>(`/namespaces/${ns}/prefixallocations`)
  },
  servers: {
    list: () => request<Server[]>('/servers')
  },
  // Phase 3 endpoints
  machinePools: {
    list: () => request<MachinePool[]>('/machinepools')
  },
  volumePools: {
    list: () => request<VolumePool[]>('/volumepools')
  },
  volumeSnapshots: {
    list:   (ns: string) => request<VolumeSnapshot[]>(`/namespaces/${ns}/volumesnapshots`),
    delete: (ns: string, name: string) => request<void>(`/namespaces/${ns}/volumesnapshots/${name}`, { method: 'DELETE' })
  },
  natGateways: {
    list:   (ns: string) => request<NATGateway[]>(`/namespaces/${ns}/natgateways`),
    delete: (ns: string, name: string) => request<void>(`/namespaces/${ns}/natgateways/${name}`, { method: 'DELETE' })
  },
  resourceQuotas: {
    list: (ns: string) => request<ResourceQuota[]>(`/namespaces/${ns}/resourcequotas`)
  },
  events: {
    list: (ns: string, regarding?: string, kind?: string) => {
      const params = new URLSearchParams()
      if (regarding) params.set('regarding', regarding)
      if (kind) params.set('kind', kind)
      const qs = params.toString() ? `?${params.toString()}` : ''
      return request<ResourceEvent[]>(`/namespaces/${ns}/events${qs}`)
    }
  },
  // Phase 4 endpoints
  networkPolicies: {
    list: (ns: string) => request<NetworkPolicy[]>(`/namespaces/${ns}/networkpolicies`),
  },
  buckets: {
    list:   (ns: string) => request<Bucket[]>(`/namespaces/${ns}/buckets`),
    delete: (ns: string, name: string) => request<void>(`/namespaces/${ns}/buckets/${name}`, { method: 'DELETE' }),
  },
  bucketClasses: {
    list: () => request<BucketClass[]>('/bucketclasses'),
  },
  bucketPools: {
    list: () => request<BucketPool[]>('/bucketpools'),
  },
  mode: {
    get: () => request<{ demo: boolean }>('/mode'),
  },
  auth: {
    checkNamespaceAccess: (ns: string) =>
      request<{ allowed: boolean; namespace: string }>(`/auth/namespace-access?namespace=${encodeURIComponent(ns)}`),
  },
}

// ── Type definitions ──────────────────────────────────────────────────────

export interface Condition {
  type: string; status: 'True' | 'False' | 'Unknown' | string
  reason?: string; message?: string; lastTransitionTime: string
}

export interface Machine {
  name: string; namespace: string; state: string; power: string
  machineClass: string; image: string; ips: string[]; volumes: string[]
  annotations?: Record<string, string>; createdAt: string
  conditions?: Condition[]
}
export interface CreateMachineRequest {
  name: string; machineClass: string; image: string
  networkName: string; ip: string; volumes: VolumeAttachment[]; power: string
}
export interface VolumeAttachment { name: string; sizeBytes: number; volumeClass: string }
export interface MachineClass { name: string; cpu: string; ram: string }

export interface VolumeClass { name: string; storage: string }

export interface Volume {
  name: string; namespace: string; state: string; sizeBytes: number; volumeClass: string; createdAt: string
}
export interface VolumeDetail {
  name: string; namespace: string; volumeClass: string; sizeGiB: number
  state: string; volumeID: string; accessDriver: string; machineRef?: string; createdAt: string
  conditions?: Condition[]
}
export interface CreateVolumeRequest {
  name: string; volumeClass: string; sizeGiB: number; encryptionSecret?: string
}

export interface Network { name: string; namespace: string; state?: string; createdAt: string }
export interface NetworkDetail {
  name: string; namespace: string; state: string; providerID: string
  peerings: { name: string; state: string }[]; createdAt: string
}
export interface CreateNetworkRequest { name: string }
export interface NetworkInterface {
  name: string; namespace: string; state: string; ips: string[]
  network: string; machine: string; createdAt: string
}

export interface VirtualIP {
  name: string; namespace: string; ip: string; type: string; ipFamily: string; createdAt: string
}
export interface VirtualIPDetail {
  name: string; namespace: string; type: string; ipFamily: string
  ip: string; targetRef: string; createdAt: string
}
export interface CreateVirtualIPRequest { name: string; type: string; ipFamily: string }

export interface LBPort { protocol: string; port: number; endPort?: number }
export interface LoadBalancer {
  name: string; namespace: string; type: string; ips: string[]; createdAt: string
}
export interface LoadBalancerDetail {
  name: string; namespace: string; type: string; ipFamilies: string[]
  networkRef: string; ports: LBPort[]; ips: string[]; createdAt: string
}
export interface CreateLoadBalancerRequest {
  name: string; type: string; ipFamilies: string[]; networkRef: string; ports: LBPort[]
}

export interface Prefix {
  name: string; namespace: string; prefix: string; phase: string; createdAt: string
}
export interface PrefixDetail {
  name: string; namespace: string; ipFamily: string; prefix: string
  phase: string; parentRef: string; used: string[]; createdAt: string
}
export interface CreatePrefixRequest {
  name: string; ipFamily: string; prefix?: string; prefixLength?: number; parentRef?: string
}

export interface Server {
  name:        string
  bmcAddress:  string
  powerState:  string
  location:    string
  claimStatus: string
  machineRef?: string
}

export function wsConsoleUrl(ns: string, name: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/api/v1/ws/machines/${ns}/${name}/console`
}

// Phase 3 types

export interface MachinePool {
  name: string; state: string; availableMachineClasses: string[]; createdAt: string
}

export interface VolumePool {
  name: string; state: string; availableVolumeClasses: string[]; createdAt: string
}

export interface VolumeSnapshot {
  name: string; namespace: string; volumeRef: string; state: string; createdAt: string
}

export interface NATGateway {
  name: string; namespace: string; networkRef: string; ips: string[]; createdAt: string
}

export interface PrefixAllocation {
  name: string; namespace: string; prefixRef: string; prefix: string; phase: string; createdAt: string
}

export interface ResourceQuota {
  name: string; namespace: string
  hard: Record<string, string>; used: Record<string, string>; createdAt: string
}

export interface ResourceEvent {
  namespace: string; name: string; reason: string; message: string
  type: string; count: number; firstTime: string; lastTime: string
  regarding: string; regardsKind: string
}

// Phase 4 types

export interface NetworkPolicy {
  name: string; namespace: string; networkRef: string
  policyTypes: string[]; ingressRules: number; egressRules: number; createdAt: string
}

export interface Bucket {
  name: string; namespace: string; bucketClassRef: string
  bucketPoolRef: string; state: string; endpoint: string; createdAt: string
}

export interface BucketClass {
  name: string; capabilities: Record<string, string>; createdAt: string
}

export interface BucketPool {
  name: string; state: string; availableBucketClasses: string[]; createdAt: string
}
