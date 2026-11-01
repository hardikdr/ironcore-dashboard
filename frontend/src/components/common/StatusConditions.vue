<template>
  <div>
    <!-- Summary chip row -->
    <div class="d-flex align-center ga-2 mb-2">
      <v-chip
        v-if="!loading && conditions.length > 0"
        :color="summaryColor"
        size="small"
        variant="tonal"
        :prepend-icon="summaryIcon"
      >
        {{ summaryLabel }}
      </v-chip>
      <v-skeleton-loader v-else-if="loading" type="chip" width="100" />
      <span v-else class="text-caption text-medium-emphasis">No conditions reported</span>

      <v-btn
        v-if="!loading && conditions.length > 0"
        variant="text"
        size="x-small"
        :icon="expanded ? 'mdi-chevron-up' : 'mdi-chevron-down'"
        :aria-expanded="expanded.toString()"
        aria-label="Toggle conditions"
        @click="expanded = !expanded"
      />
    </div>

    <!-- Expanded table -->
    <v-expand-transition>
      <div v-if="expanded && conditions.length > 0">
        <v-table density="compact" class="rounded">
          <thead>
            <tr>
              <th class="text-caption">Type</th>
              <th class="text-caption">Status</th>
              <th class="text-caption">Reason</th>
              <th class="text-caption">Message</th>
              <th class="text-caption">Since</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="c in conditions" :key="c.type">
              <td class="text-caption font-weight-medium">{{ c.type }}</td>
              <td>
                <v-chip :color="statusColor(c.status)" size="x-small" variant="tonal">
                  {{ c.status }}
                </v-chip>
              </td>
              <td class="text-caption">{{ c.reason || '—' }}</td>
              <td class="text-caption" style="max-width: 300px; word-break: break-word;">
                {{ c.message || '—' }}
              </td>
              <td class="text-caption text-no-wrap">
                <span v-if="c.lastTransitionTime" :title="absoluteTime(c.lastTransitionTime)">
                  {{ relativeTime(c.lastTransitionTime) }}
                </span>
                <span v-else class="text-medium-emphasis">—</span>
              </td>
            </tr>
          </tbody>
        </v-table>
      </div>
    </v-expand-transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { relativeTime, absoluteTime } from '@/composables/useTime'

export interface Condition {
  type: string
  status: 'True' | 'False' | 'Unknown' | string
  reason?: string
  message?: string
  lastTransitionTime: string
}

const props = defineProps<{
  conditions?: Condition[]
  loading?: boolean
}>()

const expanded = ref(false)
const conditions = computed(() => props.conditions ?? [])

function statusColor(status: string): string {
  if (status === 'True') return 'success'
  if (status === 'False') return 'error'
  return 'warning'
}

const issueCount = computed(() =>
  conditions.value.filter(c => c.status === 'False').length
)
const unknownCount = computed(() =>
  conditions.value.filter(c => c.status === 'Unknown').length
)

const summaryColor = computed(() => {
  if (issueCount.value > 0) return 'error'
  if (unknownCount.value > 0) return 'warning'
  return 'success'
})

const summaryIcon = computed(() => {
  if (issueCount.value > 0) return 'mdi-alert-circle'
  if (unknownCount.value > 0) return 'mdi-help-circle'
  return 'mdi-check-circle'
})

const summaryLabel = computed(() => {
  if (issueCount.value > 0) return `${issueCount.value} Issue${issueCount.value > 1 ? 's' : ''}`
  if (unknownCount.value > 0) return 'Unknown'
  return 'All Healthy'
})
</script>
