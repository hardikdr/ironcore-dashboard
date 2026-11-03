const BASE = '/api/sre/v1'

export interface SreSourceMeta {
  id: string
  collectedAt?: string
  freshness: 'fresh' | 'stale' | 'unknown' | string
  access: 'ok' | 'forbidden' | 'unavailable' | string
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
  provenance: { environment?: string; source?: string; kind: string; namespace?: string; name: string; uid?: string }
  state?: string
  health?: string
  labels?: Record<string, string>
  relationships?: unknown[]
}

export interface SreObservation extends SreResource {
  reason: string
  message: string
  severity: string
  confidence: string
  firstObserved: string
  lastObserved: string
}

export interface SreResourceDTO { id: string; resource: SreResource }
export interface SreSearchResult extends SreResourceDTO {
  matchedField: string
  matchKind: string
  fragment: string
  ranges: Array<{ start: number; end: number }>
}

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
  observations: (signal?: AbortSignal) => get<SreObservation[]>('/observations', signal),
}
