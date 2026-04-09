<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">Load Balancers</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-scale-balance</v-icon>
        Load Balancers
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No load balancers"
      >
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">
            {{ ip }}
          </v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type LoadBalancer } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const nsStore = useNamespaceStore()
const items   = ref<LoadBalancer[]>([])
const loading = ref(false)

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Type',    key: 'type'      },
  { title: 'IPs',     key: 'ips'       },
  { title: 'Created', key: 'createdAt' },
]

async function load() {
  loading.value = true
  items.value   = await api.loadBalancers.list(nsStore.active)
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
