<template>
  <v-chip
    :color="chipColor"
    size="x-small"
    variant="tonal"
    :prepend-icon="chipIcon"
    class="font-weight-medium"
    style="font-size: 11px"
  >
    <template #prepend>
      <v-icon :icon="chipIcon" size="12" aria-hidden="true" />
    </template>
    {{ state }}
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string }>()

const chipColor = computed(() => {
  const s = (props.state || '').toLowerCase()
  if (s === 'running' || s === 'ready' || s === 'available' || s === 'bound' || s === 'allocated' || s === 'on' || s === 'active') return 'success'
  if (s === 'pending' || s === 'powering on' || s === 'poweringon') return 'warning'
  if (s === 'error' || s === 'failed' || s === 'unreachable') return 'error'
  if (s === 'off' || s === 'terminated' || s === 'poweringoff' || s === 'powering off') return 'default'
  if (s === 'paused') return 'info'
  return 'secondary'
})

const chipIcon = computed(() => {
  const s = (props.state || '').toLowerCase()
  if (s === 'running' || s === 'ready' || s === 'available' || s === 'on' || s === 'active') return 'mdi-check-circle-outline'
  if (s === 'pending' || s === 'poweringon') return 'mdi-clock-outline'
  if (s === 'error' || s === 'failed') return 'mdi-alert-circle-outline'
  if (s === 'off' || s === 'terminated') return 'mdi-circle-off-outline'
  if (s === 'paused') return 'mdi-pause-circle-outline'
  return 'mdi-circle-outline'
})
</script>
