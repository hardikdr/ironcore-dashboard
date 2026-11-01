<template>
  <div>
    <div v-if="!conditions || conditions.length === 0" class="text-caption text-medium-emphasis pa-4">
      No conditions reported.
    </div>
    <v-table v-else density="compact">
      <thead>
        <tr>
          <th style="width:120px">Type</th>
          <th style="width:80px">Status</th>
          <th>Reason</th>
          <th>Message</th>
          <th style="width:120px">Last Transition</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="c in conditions" :key="c.type">
          <td><code class="text-caption">{{ c.type }}</code></td>
          <td>
            <v-chip
              :color="conditionColor(c.status)"
              size="x-small"
              variant="tonal"
              :prepend-icon="conditionIcon(c.status)"
            >{{ c.status }}</v-chip>
          </td>
          <td class="text-caption">{{ c.reason || '—' }}</td>
          <td class="text-caption text-medium-emphasis" style="max-width:320px; white-space:normal; word-break:break-word">
            {{ c.message || '—' }}
          </td>
          <td class="text-caption text-medium-emphasis" :title="c.lastTransitionTime">
            {{ relativeTime(c.lastTransitionTime) }}
          </td>
        </tr>
      </tbody>
    </v-table>
  </div>
</template>

<script setup lang="ts">
import { relativeTime } from '@/composables/useTime'

export interface Condition {
  type: string
  status: 'True' | 'False' | 'Unknown' | string
  reason?: string
  message?: string
  lastTransitionTime: string
}

defineProps<{
  conditions?: Condition[]
}>()

function conditionColor(status: string): string {
  if (status === 'True') return 'success'
  if (status === 'False') return 'error'
  return 'warning'
}

function conditionIcon(status: string): string {
  if (status === 'True') return 'mdi-check-circle-outline'
  if (status === 'False') return 'mdi-alert-circle-outline'
  return 'mdi-help-circle-outline'
}
</script>
