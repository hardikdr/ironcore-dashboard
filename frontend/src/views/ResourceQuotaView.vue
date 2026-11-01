<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Resource Quotas" subtitle="Namespace resource limits and usage" icon="mdi-gauge">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading && items.length === 0">
      <v-skeleton-loader type="card@2" />
    </template>
    <template v-else-if="error && items.length === 0">
      <EmptyState variant="error" title="Failed to load resource quotas" :subtitle="error">
        <v-btn color="primary" prepend-icon="mdi-refresh" @click="load">Retry</v-btn>
      </EmptyState>
    </template>
    <template v-else-if="items.length === 0">
      <EmptyState variant="empty" title="No resource quotas" subtitle="Resource quotas limit the total amount of resources per namespace." />
    </template>
    <template v-else>
      <div v-for="quota in items" :key="quota.name" class="mb-6">
        <div class="text-caption text-medium-emphasis mb-1">
          <v-icon size="12" class="mr-1">mdi-folder-outline</v-icon>
          Namespace: <strong>{{ quota.namespace }}</strong>
        </div>
        <div class="text-subtitle-1 font-weight-bold mb-3">{{ quota.name }}</div>
        <v-card :style="cardStyle" variant="flat" rounded="lg">
          <v-data-table
            :headers="quotaHeaders"
            :items="quotaRows(quota)"
            density="compact"
            :items-per-page="-1"
            hide-default-footer
          >
            <template #item.used="{ item }">
              <div class="d-flex align-center ga-2">
                <span>{{ item.used }}</span>
                <v-progress-linear
                  v-if="item.pct !== null"
                  :model-value="item.pct"
                  :color="item.pct > 90 ? 'error' : item.pct > 70 ? 'warning' : 'success'"
                  rounded
                  height="6"
                  style="min-width:80px; max-width:120px"
                />
                <span v-if="item.pct !== null" class="text-caption text-medium-emphasis">{{ item.pct }}%</span>
              </div>
            </template>
          </v-data-table>
        </v-card>
      </div>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type ResourceQuota } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const items   = ref<ResourceQuota[]>([])
const loading = ref(false)
const error   = ref<string | null>(null)

const quotaHeaders = [
  { title: 'Resource', key: 'resource' },
  { title: 'Hard Limit', key: 'hard' },
  { title: 'Used', key: 'used' },
]

function parseNum(s: string): number | null {
  const n = parseFloat(s)
  return isNaN(n) ? null : n
}

function quotaRows(q: ResourceQuota) {
  return Object.keys(q.hard).map(k => {
    const hard = parseNum(q.hard[k])
    const used = parseNum(q.used[k] ?? '0')
    const pct = (hard !== null && used !== null && hard > 0) ? Math.round((used / hard) * 100) : null
    return { resource: k, hard: q.hard[k], used: q.used[k] ?? '0', pct }
  })
}

async function load() {
  loading.value = true; error.value = null
  try { items.value = await api.resourceQuotas.list(nsStore.active) }
  catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
