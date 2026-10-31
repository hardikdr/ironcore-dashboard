<template>
  <v-card
    :style="cardStyle"
    class="stat-card pa-4 position-relative overflow-hidden"
    height="110"
  >
    <!-- Background glow -->
    <div class="stat-glow" :style="glowStyle" />

    <div class="d-flex align-start justify-space-between h-100">
      <div class="d-flex flex-column justify-space-between h-100">
        <div class="text-caption text-medium-emphasis font-weight-medium text-uppercase" style="letter-spacing:0.06em">
          {{ label }}
        </div>
        <div>
          <div class="text-h4 font-weight-bold" :style="{ color: valueColor }">
            <span v-if="loading">—</span>
            <span v-else>{{ value }}</span>
          </div>
          <div v-if="sub" class="text-caption text-medium-emphasis mt-1">{{ sub }}</div>
        </div>
      </div>
      <div
        class="stat-icon d-flex align-center justify-center rounded-lg"
        :style="iconBgStyle"
      >
        <v-icon :color="color" size="22">{{ icon }}</v-icon>
      </div>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from 'vuetify'

const props = defineProps<{
  label:   string
  value:   number | string
  icon:    string
  color?:  string
  sub?:    string
  loading?: boolean
}>()

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)

const valueColor = computed(() => `rgb(var(--v-theme-${props.color ?? 'primary'}))`)

const cardStyle = computed(() => isDark.value
  ? 'background: rgba(20,30,51,0.8); border: 1px solid rgba(129,140,248,0.12);'
  : 'background: rgba(255,255,255,0.9); border: 1px solid rgba(99,102,241,0.12);'
)

const glowStyle = computed(() => ({
  position: 'absolute',
  top: '-20px',
  right: '-20px',
  width: '80px',
  height: '80px',
  borderRadius: '50%',
  background: `rgba(var(--v-theme-${props.color ?? 'primary'}), 0.08)`,
  filter: 'blur(20px)',
}))

const iconBgStyle = computed(() => ({
  width: '44px',
  height: '44px',
  background: `rgba(var(--v-theme-${props.color ?? 'primary'}), 0.12)`,
  flexShrink: 0,
}))
</script>

<style scoped>
.stat-card { transition: transform 0.2s, box-shadow 0.2s; }
.stat-card:hover { transform: translateY(-2px); box-shadow: 0 8px 30px rgba(0,0,0,0.2) !important; }
.stat-glow { pointer-events: none; }
.stat-icon { flex-shrink: 0; }
</style>
