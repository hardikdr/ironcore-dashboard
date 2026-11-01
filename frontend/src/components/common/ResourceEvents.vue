<template>
  <div>
    <div class="text-subtitle-2 font-weight-bold mb-3 d-flex align-center ga-2">
      <v-icon size="16" color="primary">mdi-clock-outline</v-icon>
      Events
      <v-chip v-if="events.length" size="x-small" variant="tonal">{{ events.length }}</v-chip>
    </div>

    <div v-if="loading" class="py-4">
      <v-skeleton-loader type="list-item-two-line@3" />
    </div>
    <div v-else-if="error" class="text-caption text-error py-2">
      <v-icon size="14" class="mr-1">mdi-alert-circle-outline</v-icon>{{ error }}
    </div>
    <div v-else-if="events.length === 0" class="text-caption text-medium-emphasis py-2">
      No events found.
    </div>
    <v-timeline v-else density="compact" side="end" truncate-line="both" class="event-timeline">
      <v-timeline-item
        v-for="ev in events"
        :key="ev.name"
        :dot-color="ev.type === 'Warning' ? 'warning' : 'success'"
        size="x-small"
      >
        <div class="d-flex align-start justify-space-between">
          <div>
            <span class="text-caption font-weight-medium">{{ ev.reason }}</span>
            <span v-if="ev.count > 1" class="text-caption text-medium-emphasis ml-1">(×{{ ev.count }})</span>
            <div class="text-caption text-medium-emphasis" style="max-width:420px; white-space:pre-wrap; word-break:break-word;">{{ ev.message }}</div>
          </div>
          <span class="text-caption text-medium-emphasis ml-4 flex-shrink-0">{{ relativeTime(ev.lastTime || ev.firstTime) }}</span>
        </div>
      </v-timeline-item>
    </v-timeline>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type ResourceEvent } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { relativeTime } from '@/composables/useTime'

const props = defineProps<{
  resourceName: string
  resourceKind: string
}>()

const nsStore = useNamespaceStore()
const events  = ref<ResourceEvent[]>([])
const loading = ref(false)
const error   = ref<string | null>(null)

async function load() {
  loading.value = true; error.value = null
  try {
    events.value = await api.events.list(nsStore.active, props.resourceName, props.resourceKind)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => [props.resourceName, props.resourceKind, nsStore.active], load)
</script>

<style scoped>
.event-timeline :deep(.v-timeline-item__body) {
  padding-bottom: 8px;
}
</style>
