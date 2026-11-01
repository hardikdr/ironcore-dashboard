<template>
  <v-container fluid class="pa-6">
    <PageHeader title="NAT Gateways" subtitle="Network address translation gateways" icon="mdi-transit-transfer">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading && items.length === 0">
      <v-skeleton-loader type="table-row@3" />
    </template>
    <template v-else-if="error && items.length === 0">
      <EmptyState variant="error" title="Failed to load NAT gateways" :subtitle="error">
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
          <template #item.networkRef="{ item }">
            <router-link v-if="item.networkRef" :to="`/networks/${item.networkRef}`" class="text-primary text-decoration-none">{{ item.networkRef }}</router-link>
          </template>
          <template #item.ips="{ item }">
            <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
            <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
          </template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #item.actions="{ item }">
            <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" :aria-label="`Delete NAT gateway ${item.name}`" @click="confirmDelete(item.name)" />
          </template>
          <template #no-data>
            <EmptyState variant="empty" title="No NAT gateways" subtitle="NAT gateways provide outbound internet access for network resources." />
          </template>
        </v-data-table>
      </v-card>
    </template>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="d-flex align-center ga-2"><v-icon color="error">mdi-alert-circle-outline</v-icon> Delete NAT gateway?</v-card-title>
        <v-card-text>NAT gateway <strong>{{ deleteTarget }}</strong> will be permanently deleted.</v-card-text>
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
import { api, type NATGateway } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader  from '@/components/common/PageHeader.vue'
import EmptyState  from '@/components/common/EmptyState.vue'

const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const items   = ref<NATGateway[]>([])
const loading = ref(false)
const error   = ref<string | null>(null)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const headers = [
  { title: 'Name',       key: 'name'       },
  { title: 'Network',    key: 'networkRef'  },
  { title: 'IPs',        key: 'ips'         },
  { title: 'Created',    key: 'createdAt'   },
  { title: '',           key: 'actions', sortable: false },
]

async function load() {
  loading.value = true; error.value = null
  try { items.value = await api.natGateways.list(nsStore.active) }
  catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await api.natGateways.delete(nsStore.active, deleteTarget.value)
  deleting.value = false; deleteDialog.value = false; load()
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
