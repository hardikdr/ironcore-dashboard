export function relativeTime(iso?: string): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  if (diff < 60_000)          return 'just now'
  if (diff < 3_600_000)       return Math.floor(diff / 60_000) + 'm ago'
  if (diff < 86_400_000)      return Math.floor(diff / 3_600_000) + 'h ago'
  if (diff < 7 * 86_400_000)  return Math.floor(diff / 86_400_000) + 'd ago'
  return new Date(iso).toLocaleDateString('en', { month: 'short', day: 'numeric' })
}

export function absoluteTime(iso?: string): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}
