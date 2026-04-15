<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Networking</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/networks/new' }">Create Network</v-btn>
    </div>

    <v-row dense class="mb-6">
      <v-col cols="auto"><StatCard label="Available" :value="availableCount" color="#16a34a" /></v-col>
      <v-col cols="auto"><StatCard label="Pending"   :value="pendingCount"   color="#d97706" /></v-col>
      <v-col cols="auto"><StatCard label="Total"     :value="networks.length" /></v-col>
    </v-row>

    <v-card v-if="networks.length === 0 && !loading" variant="outlined" rounded="lg" class="mb-6">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-lan-disconnect</v-icon>
        <div class="text-h6 mb-2">No networks yet</div>
        <div class="text-medium-emphasis mb-4">Create your first network to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/networks/new' }">Create Network</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else variant="outlined" rounded="lg" class="mb-6">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-lan</v-icon>
        Networks
        <v-chip size="x-small" color="primary">{{ networks.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="netHeaders" :items="networks" :loading="loading" density="compact" hover no-data-text="No networks">
        <template #item.name="{ item }">
          <router-link :to="`/networks/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
        </template>
        <template #item.actions="{ item }">
          <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" @click="confirmDelete(item.name)" />
        </template>
      </v-data-table>
    </v-card>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-network-outline</v-icon>
        Network Interfaces
        <v-chip size="x-small" color="primary">{{ interfaces.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="nicHeaders" :items="interfaces" :loading="loading" density="compact" hover no-data-text="No network interfaces">
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ deleteTarget }}"?</v-card-title>
        <v-card-text>This action cannot be undone. The network will be permanently deleted from namespace <b>{{ nsStore.active }}</b>.</v-card-text>
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
import { ref, computed, onMounted, watch } from 'vue'
import { api, type Network, type NetworkInterface } from '@/api/client'
import { useNetworksStore } from '@/stores/networks'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'
import StatCard from '@/components/StatCard.vue'

const nsStore    = useNamespaceStore()
const netStore   = useNetworksStore()
const networks   = ref<Network[]>([])
const interfaces = ref<NetworkInterface[]>([])
const loading    = ref(false)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const availableCount = computed(() => networks.value.filter(n => (n as any).state === 'Available').length)
const pendingCount   = computed(() => networks.value.filter(n => (n as any).state === 'Pending').length)

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
  loading.value = true
  const ns = nsStore.active
  ;[networks.value, interfaces.value] = await Promise.all([
    api.networks.list(ns), api.networks.listInterfaces(ns)
  ])
  loading.value = false
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await netStore.deleteNetwork(deleteTarget.value)
  deleting.value = false; deleteDialog.value = false
  await load()
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
