<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Virtual IPs</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/virtualips/new' }">Create Virtual IP</v-btn>
    </div>

    <v-row dense class="mb-6">
      <v-col cols="auto"><StatCard label="Allocated" :value="allocatedCount" color="#16a34a" /></v-col>
      <v-col cols="auto"><StatCard label="Pending"   :value="pendingCount"   color="#d97706" /></v-col>
      <v-col cols="auto"><StatCard label="Total"     :value="items.length" /></v-col>
    </v-row>

    <v-card v-if="items.length === 0 && !loading" variant="outlined" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-earth-off</v-icon>
        <div class="text-h6 mb-2">No virtual IPs yet</div>
        <div class="text-medium-emphasis mb-4">Create your first Virtual IP to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/virtualips/new' }">Create Virtual IP</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-earth</v-icon>
        Virtual IPs
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="headers" :items="items" :loading="loading" density="compact" hover :no-data-text="loadError ?? 'No virtual IPs'">
        <template #item.name="{ item }">
          <router-link :to="`/virtualips/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
        </template>
        <template #item.ip="{ item }">
          <v-chip v-if="item.ip" size="x-small" color="deep-purple" class="font-weight-bold">{{ item.ip }}</v-chip>
          <span v-else class="text-medium-emphasis">Pending…</span>
        </template>
        <template #item.actions="{ item }">
          <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" @click="confirmDelete(item.name)" />
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ deleteTarget }}"?</v-card-title>
        <v-card-text>This action cannot be undone. The Virtual IP will be permanently deleted from namespace <b>{{ nsStore.active }}</b>.</v-card-text>
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
import { api, type VirtualIP } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatCard from '@/components/StatCard.vue'

const nsStore  = useNamespaceStore()
const items    = ref<VirtualIP[]>([])
const loading  = ref(false)
const loadError = ref<string | null>(null)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const allocatedCount = computed(() => items.value.filter(v => v.ip).length)
const pendingCount   = computed(() => items.value.filter(v => !v.ip).length)

const headers = [
  { title: 'Name',      key: 'name'      },
  { title: 'IP',        key: 'ip'        },
  { title: 'Type',      key: 'type'      },
  { title: 'IP Family', key: 'ipFamily'  },
  { title: 'Created',   key: 'createdAt' },
  { title: '',          key: 'actions', sortable: false },
]

async function load() {
  loading.value = true
  loadError.value = null
  try {
    items.value = await api.virtualIPs.list(nsStore.active)
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  try {
    await api.virtualIPs.delete(nsStore.active, deleteTarget.value)
    deleteDialog.value = false
    await load()
  } catch (e: unknown) {
    console.error('Delete failed:', e)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
