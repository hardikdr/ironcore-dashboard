import type { SreResource } from '@/api/sre'

export function resourceId(key: string): string {
  const bytes = new TextEncoder().encode(key)
  return btoa(Array.from(bytes, byte => String.fromCharCode(byte)).join(''))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function resourceRoute(resource: SreResource): string {
  const id = resourceId(resource.key)
  const kind = resource.provenance.kind.toLowerCase()
  if (kind === 'zone') return `/sre/zones/${encodeURIComponent(resource.provenance.name)}`
  if (kind.includes('vm')) return `/sre/vms/${id}`
  if (kind === 'shoot') return `/sre/shoots/${id}`
  if (kind.includes('node') || kind === 'server') return `/sre/nodes/${id}`
  if (kind === 'volume') return `/sre/volumes/${id}`
  if (kind === 'loadbalancer') return `/sre/loadbalancers/${id}`
  return `/sre/dependencies/${id}`
}
