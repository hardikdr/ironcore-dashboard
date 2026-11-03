<template>
  <v-chip
    :color="presentation.color"
    size="x-small"
    variant="tonal"
    :prepend-icon="presentation.icon"
    :aria-label="`Status: ${presentation.label}, ${presentation.severity}`"
    class="font-weight-medium"
    style="font-size: 11px"
  >
    <template #prepend>
      <v-icon :icon="presentation.icon" size="12" aria-hidden="true" />
    </template>
    {{ presentation.label }}
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { statusPresentation } from '@/lib/status'

const props = defineProps<{ state?: string | null }>()
const presentation = computed(() => statusPresentation(props.state))
</script>
