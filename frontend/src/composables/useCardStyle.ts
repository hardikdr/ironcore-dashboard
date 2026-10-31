import { computed } from 'vue'
import { useTheme } from 'vuetify'

export function useCardStyle() {
  const theme  = useTheme()
  const isDark = computed(() => theme.global.current.value.dark)

  const cardStyle = computed(() => ({
    background: isDark.value ? 'rgba(20,30,51,0.8)' : 'rgba(255,255,255,0.95)',
    border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.10)'}`,
  }))

  const headerStyle = computed(() => ({
    background: isDark.value ? 'rgba(14,21,40,0.9)' : '#f8fafc',
    borderBottom: `1px solid ${isDark.value ? 'rgba(129,140,248,0.08)' : 'rgba(0,0,0,0.07)'}`,
  }))

  const glassCard = computed(() =>
    isDark.value
      ? 'background: rgba(20,30,51,0.8); border: 1px solid rgba(129,140,248,0.1); backdrop-filter: blur(8px);'
      : 'background: rgba(255,255,255,0.95); border: 1px solid rgba(0,0,0,0.10);'
  )

  return { isDark, cardStyle, headerStyle, glassCard }
}
