<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Bucket Classes" subtitle="Object storage class definitions and capabilities" icon="mdi-bucket">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading && items.length === 0">
      <v-skeleton-loader type="table-row@3" />
    </template>
    <template v-else-if="error && items.length === 0">
      <EmptyState variant="error" title="Failed to load bucket classes" :subtitle="error">
        <v-btn color="primary" prepend-icon="mdi-refresh" @click="load">Retry</v-btn>
      </EmptyState>
    </template>
    <template v-else>
      <v-card :style="cardStyle" variant="flat" rounded="lg">
        <v-data-table
          :headers="headers"
          :items="items"
          :loading="loading"
          density="compact"
          hover
          item-key="name"
          :items-per-page="50"
        >
          <template #item.capabilities="{ item }">
            <div class="d-flex flex-wrap ga-1">
              <v-chip
                v-for="(val, key) in item.capabilities"
                :key="key"
                size="x-small"
                color="secondary"
                variant="tonal"
              >{{ key }}={{ val }}</v-chip>
              <span v-if="Object.keys(item.capabilities).length === 0" class="text-medium-emphasis">—</span>
            </div>
          </template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #no-data>
            <EmptyState variant="empty" title="No bucket classes" subtitle="Bucket classes are provisioned by the infrastructure administrator." />
          </template>
        </v-data-table>
      </v-card>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type BucketClass } from '@/api/client'
import { useCardStyle } from '@/composables/useCardStyle'
import { relativeTime } from '@/composables/useTime'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const { cardStyle } = useCardStyle()

const items   = ref<BucketClass[]>([])
const loading = ref(false)
const error   = ref<string | null>(null)

const headers = [
  { title: 'Name',         key: 'name'         },
  { title: 'Capabilities', key: 'capabilities', sortable: false },
  { title: 'Created',      key: 'createdAt'    },
]

async function load() {
  loading.value = true; error.value = null
  try { items.value = await api.bucketClasses.list() }
  catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

onMounted(load)
</script>
