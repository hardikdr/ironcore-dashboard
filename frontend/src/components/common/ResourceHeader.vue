<!-- Dense status strip shown between PageHeader and tabs on resource detail views.
     Displays key facts (state, class, age, IDs) without adding vertical scroll. -->
<template>
  <div class="resource-header d-flex align-center flex-wrap ga-2 py-3 mb-4" :style="stripStyle">
    <!-- State badge -->
    <StatusBadge v-if="state" :state="state" />

    <!-- Named chips -->
    <template v-for="chip in chips" :key="chip.label">
      <v-chip
        :color="chip.color ?? 'default'"
        :prepend-icon="chip.icon"
        size="small"
        variant="tonal"
        :title="chip.label"
      >
        <span class="text-caption text-medium-emphasis mr-1">{{ chip.label }}</span>
        <span class="font-weight-medium">{{ chip.value }}</span>
      </v-chip>
    </template>

    <v-spacer />

    <!-- Created timestamp -->
    <span
      v-if="createdAt"
      class="text-caption text-medium-emphasis"
      :title="absoluteTime(createdAt)"
    >
      Created {{ relativeTime(createdAt) }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCardStyle } from '@/composables/useCardStyle'
import { relativeTime, absoluteTime } from '@/composables/useTime'
import StatusBadge from '@/components/common/StatusBadge.vue'

export interface ResourceChip {
  label: string
  value: string
  icon?: string
  color?: string
}

const props = defineProps<{
  state?:     string
  chips?:     ResourceChip[]
  createdAt?: string
}>()

const { cardStyle } = useCardStyle()

const stripStyle = computed(() => ({
  ...cardStyle.value,
  borderRadius: '8px',
  padding: '8px 16px',
  borderLeft: props.state
    ? `3px solid rgb(var(--v-theme-${stateColor(props.state)}))`
    : undefined,
}))

function stateColor(s: string): string {
  const l = s.toLowerCase()
  if (l === 'running' || l === 'available' || l === 'bound' || l === 'active') return 'success'
  if (l === 'pending' || l === 'provisioning' || l === 'terminating') return 'warning'
  if (l === 'error' || l === 'failed') return 'error'
  return 'primary'
}
</script>
