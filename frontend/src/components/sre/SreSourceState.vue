<template><v-chip size="small" variant="tonal" :color="state.color" :prepend-icon="state.icon">{{ state.label }}</v-chip></template>
<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ access: string; freshness: string }>()
const state = computed(() => {
  if (props.access === 'not-configured') return { label: 'Not configured', color: 'default', icon: 'mdi-minus-circle-outline' }
  if (props.access === 'denied') return { label: 'RBAC denied', color: 'warning', icon: 'mdi-shield-lock-outline' }
  if (props.access === 'unreachable') return { label: 'Unreachable', color: 'error', icon: 'mdi-lan-disconnect' }
  if (props.access === 'unsupported') return { label: 'API unsupported', color: 'warning', icon: 'mdi-api-off' }
  if (props.access !== 'ok') return { label: 'Unknown', color: 'default', icon: 'mdi-help-circle-outline' }
  if (props.freshness === 'stale') return { label: 'Stale', color: 'warning', icon: 'mdi-clock-alert-outline' }
  if (props.freshness !== 'fresh') return { label: 'Freshness unknown', color: 'default', icon: 'mdi-clock-question-outline' }
  return { label: 'Healthy', color: 'success', icon: 'mdi-check-circle-outline' }
})
</script>
