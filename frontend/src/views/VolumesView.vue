<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Volumes</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/volumes/new' }">Create Volume</v-btn>
    </div>

    <v-row dense class="mb-6">
      <v-col cols="auto">
        <StatCard label="Available" :value="availableCount" color="#16a34a" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Pending" :value="pendingCount" color="#d97706" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Total" :value="store.items.length" />
      </v-col>
    </v-row>

    <v-card v-if="store.items.length === 0 && !store.loading" variant="outlined" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-database-off-outline</v-icon>
        <div class="text-h6 mb-2">No volumes yet</div>
        <div class="text-medium-emphasis mb-4">Create your first volume to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/volumes/new' }">Create Volume</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else variant="outlined" rounded="lg">
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
        <template #item.name="{ item }">
          <router-link :to="`/volumes/${item.name}`" class="text-primary text-decoration-none font-weight-medium">
            {{ item.name }}
          </router-link>
        </template>
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.sizeBytes="{ item }">{{ formatBytes(item.sizeBytes) }}</template>
        <template #item.actions="{ item }">
          <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error"
            @click="confirmDelete(item.name)" />
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ deleteTarget }}"?</v-card-title>
        <v-card-text>This action cannot be undone. The volume will be permanently deleted from namespace <b>{{ nsStore.active }}</b>.</v-card-text>
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
import { useVolumesStore } from '@/stores/volumes'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'
import StatCard from '@/components/StatCard.vue'

const store   = useVolumesStore()
const nsStore = useNamespaceStore()

const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

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

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await store.deleteVolume(deleteTarget.value)
  deleting.value = false; deleteDialog.value = false
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
