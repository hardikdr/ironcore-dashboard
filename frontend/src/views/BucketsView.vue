<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Buckets" subtitle="Object storage buckets" icon="mdi-bucket-outline">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading && items.length === 0">
      <v-skeleton-loader type="table-row@3" />
    </template>
    <template v-else-if="error && items.length === 0">
      <EmptyState variant="error" title="Failed to load buckets" :subtitle="error">
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
          :items-per-page="20"
        >
          <template #item.state="{ item }">
            <StatusBadge :state="item.state" />
          </template>
          <template #item.endpoint="{ item }">
            <a v-if="item.endpoint" :href="item.endpoint" target="_blank" rel="noopener" class="text-primary text-decoration-none" style="font-family:monospace; font-size:12px; max-width:220px; display:inline-block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;" :title="item.endpoint">{{ item.endpoint }}</a>
            <span v-else class="text-medium-emphasis">—</span>
          </template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #item.actions="{ item }">
            <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" :aria-label="`Delete bucket ${item.name}`" @click="confirmDelete(item.name)" />
          </template>
          <template #no-data>
            <EmptyState variant="empty" title="No buckets" subtitle="Buckets provide S3-compatible object storage." />
          </template>
        </v-data-table>
      </v-card>
    </template>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="d-flex align-center ga-2"><v-icon color="error">mdi-alert-circle-outline</v-icon> Delete bucket?</v-card-title>
        <v-card-text>Bucket <strong>{{ deleteTarget }}</strong> and all its contents will be permanently deleted.</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deleting" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Bucket } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader  from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import EmptyState  from '@/components/common/EmptyState.vue'

const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const items        = ref<Bucket[]>([])
const loading      = ref(false)
const error        = ref<string | null>(null)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const headers = [
  { title: 'Name',     key: 'name'           },
  { title: 'Class',    key: 'bucketClassRef' },
  { title: 'Pool',     key: 'bucketPoolRef'  },
  { title: 'State',    key: 'state'          },
  { title: 'Endpoint', key: 'endpoint', sortable: false },
  { title: 'Created',  key: 'createdAt'      },
  { title: '',         key: 'actions', sortable: false },
]

async function load() {
  loading.value = true; error.value = null
  try { items.value = await api.buckets.list(nsStore.active) }
  catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await api.buckets.delete(nsStore.active, deleteTarget.value)
  deleting.value = false; deleteDialog.value = false; load()
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
