<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Networking" subtitle="Networks and network interfaces" icon="mdi-lan">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="store.loading" @click="load()">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/networks/new' }">Create Network</v-btn>
      </template>
    </PageHeader>

    <template v-if="store.loading && store.items.length === 0">
      <v-skeleton-loader type="table-row@6" />
    </template>

    <template v-else-if="store.error && store.items.length === 0">
      <EmptyState variant="error" title="Failed to load networks" :subtitle="store.error">
        <v-btn color="primary" prepend-icon="mdi-refresh" @click="load()">Retry</v-btn>
      </EmptyState>
    </template>

    <template v-else>
      <v-row dense class="mb-4">
        <v-col cols="auto"><StatCard label="Total" :value="store.items.length" icon="mdi-lan" color="primary" /></v-col>
        <v-col cols="auto"><StatCard label="Interfaces" :value="interfaces.length" icon="mdi-network-outline" color="blue" /></v-col>
      </v-row>

      <!-- Networks table -->
      <div class="d-flex align-center ga-3 flex-wrap mb-4">
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          placeholder="Filter networks…"
          variant="outlined"
          density="compact"
          hide-details
          clearable
          style="max-width:300px"
        />
        <v-spacer />
        <span class="text-caption text-medium-emphasis">{{ filteredNetworks.length }} of {{ store.items.length }}</span>
      </div>

      <v-card :style="cardStyle" variant="flat" rounded="lg" class="mb-6">
        <v-data-table
          :headers="netHeaders"
          :items="filteredNetworks"
          :loading="store.loading"
          density="compact"
          hover
          item-key="name"
          :items-per-page="pageSize"
        >
          <template #bottom>
            <div class="d-flex align-center justify-end ga-4 pa-3">
              <span class="text-caption text-medium-emphasis">Rows per page</span>
              <v-select
                v-model="pageSize"
                :items="[20, 50, 100]"
                density="compact"
                variant="outlined"
                hide-details
                style="max-width:80px"
              />
              <v-pagination
                v-model="page"
                :length="Math.ceil(filteredNetworks.length / pageSize)"
                density="compact"
                size="small"
                total-visible="5"
              />
            </div>
          </template>

          <template #item.name="{ item }">
            <router-link :to="`/networks/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
          </template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #item.actions="{ item }">
            <v-btn
              size="x-small" variant="outlined" icon="mdi-delete" color="error"
              :aria-label="`Delete network ${item.name}`"
              title="Delete network"
              @click="confirmDelete(item.name)"
            />
          </template>

          <template #no-data>
            <EmptyState
              :variant="search ? 'no-results' : 'empty'"
              :title="search ? 'No matching networks' : 'No networks'"
              :subtitle="search ? 'Try clearing the search.' : 'Create your first network to get started.'"
            >
              <v-btn v-if="!search" color="primary" prepend-icon="mdi-plus" to="/networks/new">Create Network</v-btn>
              <v-btn v-else variant="outlined" @click="search = ''">Clear search</v-btn>
            </EmptyState>
          </template>
        </v-data-table>
      </v-card>

      <!-- Network Interfaces table -->
      <div class="text-subtitle-2 font-weight-bold mb-3">Network Interfaces</div>
      <v-card :style="cardStyle" variant="flat" rounded="lg">
        <v-data-table
          :headers="nicHeaders"
          :items="interfaces"
          :loading="nicLoading"
          density="compact"
          hover
          item-key="name"
          :items-per-page="20"
        >
          <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
          <template #item.ips="{ item }">
            <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
            <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
          </template>
          <template #no-data>
            <EmptyState variant="empty" title="No network interfaces" subtitle="Network interfaces are created when machines attach to networks." />
          </template>
        </v-data-table>
      </v-card>
    </template>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete network?
        </v-card-title>
        <v-card-text>Network <strong>{{ deleteTarget }}</strong> will be permanently deleted. This action cannot be undone.</v-card-text>
        <v-alert v-if="deleteError" type="error" variant="tonal" class="mx-4 mb-2" density="compact">{{ deleteError }}</v-alert>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" :disabled="deleting" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deleting" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { api, type NetworkInterface } from '@/api/client'
import { useNetworksStore }  from '@/stores/networks'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader  from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import StatCard    from '@/components/common/StatCard.vue'
import EmptyState  from '@/components/common/EmptyState.vue'

const store   = useNetworksStore()
const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const search     = ref('')
const page       = ref(1)
const pageSize   = ref(20)
const interfaces = ref<NetworkInterface[]>([])
const nicLoading = ref(false)

const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)
const deleteError  = ref<string | null>(null)

const filteredNetworks = computed(() => {
  const q = search.value.toLowerCase()
  return q ? store.items.filter(n => n.name.toLowerCase().includes(q)) : store.items
})

watch(search, () => { page.value = 1 })

const netHeaders = [
  { title: 'Name',    key: 'name'      },
  { title: 'Created', key: 'createdAt' },
  { title: '',        key: 'actions', sortable: false },
]
const nicHeaders = [
  { title: 'Name',    key: 'name'    },
  { title: 'State',   key: 'state'   },
  { title: 'IPs',     key: 'ips'     },
  { title: 'Network', key: 'network' },
  { title: 'Machine', key: 'machine' },
]

async function load() {
  store.load()
  nicLoading.value = true
  try {
    interfaces.value = await api.networks.listInterfaces(nsStore.active)
  } finally {
    nicLoading.value = false
  }
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true; deleteError.value = null }
async function doDelete() {
  deleting.value = true
  deleteError.value = null
  try {
    await api.networks.delete(nsStore.active, deleteTarget.value)
    deleteDialog.value = false
    load()
  } catch (e: unknown) {
    deleteError.value = e instanceof Error ? e.message : String(e)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
