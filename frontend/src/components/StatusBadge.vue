<template>
  <v-chip :color="color" size="small" label>
    <v-icon start size="x-small">mdi-circle</v-icon>
    {{ label }}
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string }>()

const color = computed(() => ({
  Running:    'success',
  Available:  'success',
  Allocated:  'success',
  Pending:    'warning',
  Shutdown:   'default',
  Stopped:    'default',
  Error:      'error',
  Terminated: 'error',
  Terminating:'error',
}[props.state] ?? 'default'))

const label = computed(() => props.state || '—')
</script>
