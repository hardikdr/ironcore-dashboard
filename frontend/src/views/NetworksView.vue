<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">Networking</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <!-- Networks table -->
    <v-card variant="outlined" rounded="lg" class="mb-6">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-lan</v-icon>
        Networks
        <v-chip size="x-small" color="primary">{{ networks.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="netHeaders"
        :items="networks"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No networks"
      />
    </v-card>

    <!-- Network Interfaces table -->
    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-network-outline</v-icon>
        Network Interfaces
        <v-chip size="x-small" color="primary">{{ interfaces.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="nicHeaders"
        :items="interfaces"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No network interfaces"
      >
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.ips="{ item }">
          <v-chip
            v-for="ip in item.ips"
            :key="ip"
            size="x-small"
            color="info"
            class="mr-1"
          >{{ ip }}</v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Network, type NetworkInterface } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const nsStore    = useNamespaceStore()
const networks   = ref<Network[]>([])
const interfaces = ref<NetworkInterface[]>([])
const loading    = ref(false)

const netHeaders = [
  { title: 'Name',    key: 'name'      },
  { title: 'Created', key: 'createdAt' },
]

const nicHeaders = [
  { title: 'Name',    key: 'name'      },
  { title: 'State',   key: 'state'     },
  { title: 'IPs',     key: 'ips'       },
  { title: 'Network', key: 'network'   },
  { title: 'Machine', key: 'machine'   },
]

async function load() {
  loading.value = true
  const ns = nsStore.active
  ;[networks.value, interfaces.value] = await Promise.all([
    api.networks.list(ns),
    api.networks.listInterfaces(ns),
  ])
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
