export type StatusSeverity = 'healthy' | 'warning' | 'error' | 'pending' | 'unknown'

const STATUS_GROUPS: Readonly<Record<Exclude<StatusSeverity, 'unknown'>, readonly string[]>> = Object.freeze({
  error: Object.freeze(['error', 'failed', 'failure', 'invalid', 'unavailable']),
  warning: Object.freeze(['warning', 'degraded']),
  pending: Object.freeze(['pending', 'creating', 'provisioning', 'deleting', 'terminating']),
  healthy: Object.freeze(['running', 'available', 'bound', 'ready', 'active', 'succeeded', 'on']),
})

export function classifyStatus(value?: string | null): StatusSeverity {
  const normalized = value?.trim().toLowerCase()
  if (!normalized) return 'unknown'
  for (const [severity, values] of Object.entries(STATUS_GROUPS)) {
    if (values.includes(normalized)) return severity as StatusSeverity
  }
  return 'unknown'
}

const PRESENTATION = {
  healthy: { color: 'success', icon: 'mdi-check-circle-outline' },
  warning: { color: 'warning', icon: 'mdi-alert-outline' },
  error: { color: 'error', icon: 'mdi-alert-circle-outline' },
  pending: { color: 'info', icon: 'mdi-clock-outline' },
  unknown: { color: 'secondary', icon: 'mdi-help-circle-outline' },
} as const

export function statusPresentation(value?: string | null) {
  const severity = classifyStatus(value)
  return { label: value?.trim() || 'Unknown', severity, ...PRESENTATION[severity] }
}
