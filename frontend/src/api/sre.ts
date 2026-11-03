const BASE = '/api/sre/v1'

export interface SreSourceMeta {
  id: string
  collectedAt?: string
  freshness: 'fresh' | 'stale' | 'unknown' | string
  access: 'ok' | 'denied' | 'unreachable' | 'unsupported' | 'not-configured' | 'unknown' | string
  duration: number
  errorClass?: string
  objectCount?: Record<string, number>
}

export interface SreMeta {
  generatedAt: string
  sources: SreSourceMeta[]
  partial: boolean
  warnings: string[]
  continue?: string
}

export interface SreEnvelope<T> { data: T; meta: SreMeta }

export interface SreResource {
  type: string
  key: string
  provenance: { sourceId: string; environment?: string; clusterRole?: string; installationRole?: string; group?: string; version?: string; kind: string; namespace?: string; name: string; uid?: string; collectedAt?: string; freshness?: string; access?: string; evidence?: string }
  state?: string
  health?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
  providerId?: string
  addresses?: string[]
  conditions?: Array<{ type: string; status: string; reason?: string; message?: string }>
  sanitized?: Record<string, unknown>
  relationships?: unknown[]
}

export interface SreObservation extends SreResource {
  reason: string
  message: string
  severity: string
  confidence: string
  firstObserved: string
  lastObserved: string
  related?: string[]
}

export interface SreResourceDTO { id: string; resource: SreResource }
export interface SreSearchResult extends SreResourceDTO {
  matchedField: string
  matchKind: string
  fragment: string
  ranges: Array<{ start: number; end: number }>
}

export interface SreFact { label: string; value: string; sourceId?: string; resource?: string; evidence?: string; confidence?: string }
export interface SreGap { area: string; state: string; message: string }
export interface SreVMSummary { vm: SreResource; localization: string; facts: SreFact[]; gaps: SreGap[] }
export interface SreRelation { from: string; to: string; type: string; evidence: string; fieldPath: string; confidence: string }
export interface SreGraph { resources: SreResource[]; relations: SreRelation[]; truncated: boolean }
export interface SreZoneSummary { zone: SreResource; state: string; managementOnly: boolean; managementNodes: number; computeNodes: number; storageNodes: number; machinePools: number; volumePools: number; hostedVms: number; dependentShoots: number; facts: SreFact[]; gaps: SreGap[] }
export interface SreShootSummary { shoot: SreResource; workerVms: number; workerGroups: number; relationConfidence: string; facts: SreFact[]; gaps: SreGap[] }

async function get<T>(path: string, signal?: AbortSignal): Promise<SreEnvelope<T>> {
  const response = await fetch(`${BASE}${path}`, signal ? { signal } : {})
  if (!response.ok) {
    const payload = await response.json().catch(() => ({ error: response.statusText }))
    throw new Error(payload.error || response.statusText)
  }
  return response.json()
}

function query(params: Record<string, string | number | undefined>) {
  const values = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) if (value !== undefined && value !== '') values.set(key, String(value))
  const encoded = values.toString()
  return encoded ? `?${encoded}` : ''
}

export const sreApi = {
  status: (signal?: AbortSignal) => get<{ status: string; generation: number }>('/status', signal),
  sources: (signal?: AbortSignal) => get<SreSourceMeta[]>('/sources', signal),
  summary: (environment?: string, signal?: AbortSignal) => get<Record<string, number>>(`/summary${query({ environment })}`, signal),
  search: (value: string, signal?: AbortSignal) => get<SreSearchResult[]>(`/search${query({ q: value, limit: 25 })}`, signal),
  zones: (environment?: string, signal?: AbortSignal) => get<SreResourceDTO[]>(`/zones${query({ environment })}`, signal),
  zone: (name: string, signal?: AbortSignal) => get<SreZoneSummary>(`/zones/${encodeURIComponent(name)}`, signal),
  node: (id: string, signal?: AbortSignal) => get<SreResourceDTO>(`/nodes/${encodeURIComponent(id)}`, signal),
  network: (id: string, signal?: AbortSignal) => get<SreResourceDTO>(`/networks/${encodeURIComponent(id)}`, signal),
  loadBalancer: (id: string, signal?: AbortSignal) => get<SreResourceDTO>(`/loadbalancers/${encodeURIComponent(id)}`, signal),
  volume: (id: string, signal?: AbortSignal) => get<SreResourceDTO>(`/volumes/${encodeURIComponent(id)}`, signal),
  vms: (filters: { environment?: string; state?: string; continue?: string } = {}, signal?: AbortSignal) => get<SreResourceDTO[]>(`/vms${query({ ...filters, limit: 100 })}`, signal),
  vm: (id: string, signal?: AbortSignal) => get<SreVMSummary>(`/vms/${encodeURIComponent(id)}`, signal),
  shoots: (environment?: string, signal?: AbortSignal) => get<SreResourceDTO[]>(`/shoots${query({ environment, limit: 100 })}`, signal),
  shoot: (id: string, signal?: AbortSignal) => get<SreShootSummary>(`/shoots/${encodeURIComponent(id)}`, signal),
  graph: (id: string, signal?: AbortSignal) => get<SreGraph>(`/graph${query({ root: id, depth: 4, limit: 300 })}`, signal),
  sanitized: (id: string, signal?: AbortSignal) => get<{ id: string; details: unknown }>(`/resources/${encodeURIComponent(id)}/sanitized`, signal),
  observations: (signal?: AbortSignal) => get<SreObservation[]>('/observations', signal),
}
