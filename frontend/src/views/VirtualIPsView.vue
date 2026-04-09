<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">Virtual IPs</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-earth</v-icon>
        Virtual IPs
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No virtual IPs"
      >
        <template #item.ip="{ item }">
          <v-chip v-if="item.ip" size="x-small" color="deep-purple" class="font-weight-bold">
            {{ item.ip }}
          </v-chip>
          <span v-else class="text-medium-emphasis">Pending…</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type VirtualIP } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const nsStore = useNamespaceStore()
const items   = ref<VirtualIP[]>([])
const loading = ref(false)

const headers = [
  { title: 'Name',      key: 'name'      },
  { title: 'IP',        key: 'ip'        },
  { title: 'Type',      key: 'type'      },
  { title: 'IP Family', key: 'ipFamily'  },
  { title: 'Created',   key: 'createdAt' },
]

async function load() {
  loading.value = true
  items.value   = await api.virtualIPs.list(nsStore.active)
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
