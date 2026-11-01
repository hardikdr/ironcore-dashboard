<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Network Interfaces" subtitle="Virtual network interface cards attached to machines" icon="mdi-ethernet">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading && items.length === 0">
      <v-skeleton-loader type="table-row@4" />
    </template>
    <template v-else-if="error && items.length === 0">
      <EmptyState variant="error" title="Failed to load network interfaces" :subtitle="error">
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
          <template #item.network="{ item }">
            <router-link v-if="item.network" :to="`/networks/${item.network}`" class="text-primary text-decoration-none">{{ item.network }}</router-link>
            <span v-else class="text-medium-emphasis">—</span>
          </template>
          <template #item.machine="{ item }">
            <router-link v-if="item.machine" :to="`/machines/${item.machine}`" class="text-primary text-decoration-none">{{ item.machine }}</router-link>
            <span v-else class="text-medium-emphasis">—</span>
          </template>
          <template #item.ips="{ item }">
            <div class="d-flex flex-wrap ga-1">
              <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" variant="tonal">{{ ip }}</v-chip>
              <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
            </div>
          </template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #no-data>
            <EmptyState variant="empty" title="No network interfaces" subtitle="Interfaces are created automatically when machines are provisioned." />
          </template>
        </v-data-table>
      </v-card>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type NetworkInterface } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader  from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import EmptyState  from '@/components/common/EmptyState.vue'

const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const items   = ref<NetworkInterface[]>([])
const loading = ref(false)
const error   = ref<string | null>(null)

const headers = [
  { title: 'Name',    key: 'name'    },
  { title: 'State',   key: 'state'   },
  { title: 'Network', key: 'network' },
  { title: 'Machine', key: 'machine' },
  { title: 'IPs',     key: 'ips', sortable: false },
  { title: 'Created', key: 'createdAt' },
]

async function load() {
  loading.value = true; error.value = null
  try { items.value = await api.networks.listInterfaces(nsStore.active) }
  catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
