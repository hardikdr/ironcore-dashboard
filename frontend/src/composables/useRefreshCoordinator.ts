export interface RefreshTarget { loaded(): boolean; refresh(silent: boolean): Promise<void> }
export interface RefreshCoordinatorOptions { healthCheck(): Promise<void>; targets: RefreshTarget[]; document?: Document; intervalMs?: number }

export function createRefreshCoordinator(options: RefreshCoordinatorOptions) {
  const doc = options.document ?? document
  let timer: ReturnType<typeof setInterval> | undefined
  let refreshing = false
  async function run(targets: RefreshTarget[], silent: boolean) {
    if (refreshing) return
    refreshing = true
    try { await Promise.allSettled([options.healthCheck(), ...targets.map(target => target.refresh(silent))]) }
    finally { refreshing = false }
  }
  const refreshGlobal = () => run(options.targets, false)
  const refreshAutomatic = () => doc.hidden ? Promise.resolve() : run(options.targets.filter(target => target.loaded()), true)
  const handleVisibilityChange = () => { if (!doc.hidden) void refreshAutomatic() }
  function start() { if (!timer) { timer = setInterval(() => void refreshAutomatic(), options.intervalMs ?? 30_000); doc.addEventListener('visibilitychange', handleVisibilityChange) } }
  function stop() { if (timer) clearInterval(timer); timer = undefined; doc.removeEventListener('visibilitychange', handleVisibilityChange) }
  return { refreshGlobal, start, stop, handleVisibilityChange }
}
