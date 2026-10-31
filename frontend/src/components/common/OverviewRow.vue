<template>
  <div class="d-flex align-center justify-space-between pa-2 rounded-lg" :style="rowStyle">
    <div class="d-flex align-center ga-2">
      <v-icon :color="color" size="16">{{ icon }}</v-icon>
      <span class="text-body-2">{{ label }}</span>
    </div>
    <span class="text-h6 font-weight-bold" :style="{ color: `rgb(var(--v-theme-${color}))` }">{{ value }}</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useTheme } from 'vuetify'

defineProps<{ icon: string; label: string; value: number; color: string }>()

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)
const rowStyle = computed(() =>
  isDark.value ? 'background: rgba(255,255,255,0.03);' : 'background: rgba(0,0,0,0.02);'
)
</script>
