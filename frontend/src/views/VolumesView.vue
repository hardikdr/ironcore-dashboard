<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Volumes</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus" disabled>Create Volume</v-btn>
    </div>

    <!-- Stat cards -->
    <v-row dense class="mb-6">
      <v-col cols="auto">
        <v-card variant="outlined" rounded="lg">
          <v-card-text class="pa-4">
            <div class="text-h4 font-weight-bold" style="color:#16a34a">{{ availableCount }}</div>
            <div class="text-caption text-uppercase text-medium-emphasis mt-1">Available</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="auto">
        <v-card variant="outlined" rounded="lg">
          <v-card-text class="pa-4">
            <div class="text-h4 font-weight-bold" style="color:#d97706">{{ pendingCount }}</div>
            <div class="text-caption text-uppercase text-medium-emphasis mt-1">Pending</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="auto">
        <v-card variant="outlined" rounded="lg">
          <v-card-text class="pa-4">
            <div class="text-h4 font-weight-bold">{{ store.items.length }}</div>
            <div class="text-caption text-uppercase text-medium-emphasis mt-1">Total</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-database</v-icon>
        Volumes
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="store.items"
        :loading="store.loading"
        density="compact"
        hover
        :no-data-text="store.error ?? 'No volumes'"
      >
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.sizeBytes="{ item }">{{ formatBytes(item.sizeBytes) }}</template>
        <template #item.actions="{ item }">
          <v-btn
            size="x-small" variant="outlined" icon="mdi-delete" color="error"
            @click="store.deleteVolume(item.name)"
          />
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useVolumesStore } from '@/stores/volumes'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const store   = useVolumesStore()
const nsStore = useNamespaceStore()

const availableCount = computed(() => store.items.filter(v => v.state === 'Available').length)
const pendingCount   = computed(() => store.items.filter(v => v.state === 'Pending').length)

const headers = [
  { title: 'Name',         key: 'name'        },
  { title: 'State',        key: 'state'       },
  { title: 'Size',         key: 'sizeBytes'   },
  { title: 'Volume Class', key: 'volumeClass' },
  { title: 'Created',      key: 'createdAt'   },
  { title: '',             key: 'actions', sortable: false },
]

function formatBytes(b: number) {
  if (!b) return '—'
  const gb = b / 1024 / 1024 / 1024
  return gb >= 1000 ? `${(gb / 1024).toFixed(1)} TB` : `${Math.round(gb)} GB`
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
