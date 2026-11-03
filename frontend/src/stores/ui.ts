import { defineStore } from 'pinia'
import { ref } from 'vue'

export type ConnectionState = 'checking' | 'connected' | 'disconnected'
const THEME_KEY = 'ironcore-theme'
const RAIL_KEY = 'ironcore-navigation-rail'

export const useUIStore = defineStore('ui', () => {
  const themeName = ref(localStorage.getItem(THEME_KEY) === 'dark' ? 'dark' : 'light')
  const navigationRail = ref(localStorage.getItem(RAIL_KEY) === 'true')
  const connectionState = ref<ConnectionState>('checking')
  const lastRefreshAt = ref<Date | null>(null)
  function setTheme(value: 'light' | 'dark') { themeName.value = value; localStorage.setItem(THEME_KEY, value) }
  function setNavigationRail(value: boolean) { navigationRail.value = value; localStorage.setItem(RAIL_KEY, String(value)) }
  async function checkHealth() {
    connectionState.value = 'checking'
    try { const response = await fetch('/healthz'); connectionState.value = response.ok ? 'connected' : 'disconnected' }
    catch { connectionState.value = 'disconnected' }
    lastRefreshAt.value = new Date()
  }
  return { themeName, navigationRail, connectionState, lastRefreshAt, setTheme, setNavigationRail, checkHealth }
})
