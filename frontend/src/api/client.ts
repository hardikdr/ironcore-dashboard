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
  volumes: {
    list:   (ns: string) => request<Volume[]>(`/namespaces/${ns}/volumes`),
    create: (ns: string, body: CreateVolumeRequest) =>
      request<Volume>(`/namespaces/${ns}/volumes`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/volumes/${name}`, { method: 'DELETE' })
  },
  networks: {
    list:           (ns: string) => request<Network[]>(`/namespaces/${ns}/networks`),
    listInterfaces: (ns: string) => request<NetworkInterface[]>(`/namespaces/${ns}/networkinterfaces`)
  },
  virtualIPs: {
    list: (ns: string) => request<VirtualIP[]>(`/namespaces/${ns}/virtualips`)
  },
  loadBalancers: {
    list: (ns: string) => request<LoadBalancer[]>(`/namespaces/${ns}/loadbalancers`)
  },
  prefixes: {
    list: (ns: string) => request<Prefix[]>(`/namespaces/${ns}/prefixes`)
  }
}

// ── Type definitions (mirror backend JSON) ──────────────────────────────
export interface Machine {
  name: string; namespace: string; state: string; power: string
  machineClass: string; image: string; ips: string[]; volumes: string[]; createdAt: string
}
export interface CreateMachineRequest {
  name: string; machineClass: string; image: string
  networkName: string; ip: string; volumes: VolumeAttachment[]; power: string
}
export interface VolumeAttachment { name: string; sizeBytes: number; volumeClass: string }
export interface Volume {
  name: string; namespace: string; state: string; sizeBytes: number; volumeClass: string; createdAt: string
}
export interface CreateVolumeRequest { name: string; volumeClass: string; sizeBytes: number }
export interface MachineClass { name: string; cpu: string; ram: string }
export interface Network { name: string; namespace: string; createdAt: string }
export interface NetworkInterface {
  name: string; namespace: string; state: string; ips: string[]
  network: string; machine: string; createdAt: string
}
export interface VirtualIP {
  name: string; namespace: string; ip: string; type: string; ipFamily: string; createdAt: string
}
export interface LoadBalancer {
  name: string; namespace: string; type: string; ips: string[]; createdAt: string
}
export interface Prefix {
  name: string; namespace: string; prefix: string; phase: string; createdAt: string
}
