const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
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
    listInterfaces: (ns: string) => request<NetworkInterface[]>(`/namespaces/${ns}/networkinterfaces`)
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
      request<void>(`/namespaces/${ns}/prefixes/${name}`, { method: 'DELETE' })
  }
}

// ── Type definitions ──────────────────────────────────────────────────────

export interface Machine {
  name: string; namespace: string; state: string; power: string
  machineClass: string; image: string; ips: string[]; volumes: string[]; createdAt: string
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
  state: string; volumeID: string; accessDriver: string; createdAt: string
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
