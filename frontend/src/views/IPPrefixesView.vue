<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">IP Prefixes</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-ip-network</v-icon>
        IP Prefixes
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No IP prefixes"
      >
        <template #item.prefix="{ item }">
          <code class="text-primary">{{ item.prefix || '—' }}</code>
        </template>
        <template #item.phase="{ item }">
          <StatusBadge :state="item.phase" />
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Prefix } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const nsStore = useNamespaceStore()
const items   = ref<Prefix[]>([])
const loading = ref(false)

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Prefix',  key: 'prefix'    },
  { title: 'Phase',   key: 'phase'     },
  { title: 'Created', key: 'createdAt' },
]

async function load() {
  loading.value = true
  items.value   = await api.prefixes.list(nsStore.active)
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
