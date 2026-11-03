<template>
  <v-chip :color="presentation.color" variant="tonal" size="small" :prepend-icon="presentation.icon" :aria-label="`API connection: ${presentation.label}`">{{ presentation.label }}</v-chip>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import type { ConnectionState } from '@/stores/ui'
const props = defineProps<{ state: ConnectionState }>()
const presentation = computed(() => ({ checking: { label: 'Checking', color: 'info', icon: 'mdi-progress-clock' }, connected: { label: 'Connected', color: 'success', icon: 'mdi-check-circle-outline' }, disconnected: { label: 'Disconnected', color: 'error', icon: 'mdi-alert-circle-outline' } }[props.state]))
</script>
