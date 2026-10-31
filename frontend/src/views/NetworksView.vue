<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Networking"
      subtitle="Networks and network interfaces"
      icon="mdi-lan"
    >
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/networks/new' }">Create Network</v-btn>
      </template>
    </PageHeader>

    <v-row dense class="mb-6">
      <v-col cols="auto"><StatCard label="Available" :value="availableCount" icon="mdi-check-circle-outline" color="success" /></v-col>
      <v-col cols="auto"><StatCard label="Pending"   :value="pendingCount"   icon="mdi-clock-outline"        color="warning" /></v-col>
      <v-col cols="auto"><StatCard label="Total"     :value="networks.length" icon="mdi-lan"                 color="primary" /></v-col>
    </v-row>

    <v-card v-if="networks.length === 0 && !loading" :style="cardStyle" variant="flat" rounded="lg" class="mb-6">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-lan-disconnect</v-icon>
        <div class="text-h6 mb-2">No networks yet</div>
        <div class="text-medium-emphasis mb-4">Create your first network to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/networks/new' }">Create Network</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else :style="cardStyle" variant="flat" rounded="lg" class="mb-6">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-lan</v-icon>
        <span class="font-weight-bold">Networks</span>
        <v-chip size="x-small" color="primary">{{ networks.length }}</v-chip>
      </div>
      <v-data-table :headers="netHeaders" :items="networks" :loading="loading" density="compact" hover no-data-text="No networks">
        <template #item.name="{ item }">
          <router-link :to="`/networks/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
        </template>
        <template #item.actions="{ item }">
          <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" @click="confirmDelete(item.name)" />
        </template>
      </v-data-table>
    </v-card>

    <v-card :style="cardStyle" variant="flat" rounded="lg">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-network-outline</v-icon>
        <span class="font-weight-bold">Network Interfaces</span>
        <v-chip size="x-small" color="primary">{{ interfaces.length }}</v-chip>
      </div>
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
import { useTheme } from 'vuetify'
import { api, type Network, type NetworkInterface } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import StatCard from '@/components/common/StatCard.vue'

const nsStore    = useNamespaceStore()
const theme      = useTheme()
const networks   = ref<Network[]>([])
const interfaces = ref<NetworkInterface[]>([])
const loading    = ref(false)
const loadError  = ref<string | null>(null)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() => ({
  background: isDark.value ? 'rgba(20,30,51,0.8)' : '#ffffff',
  border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.12)'}`,
}))
const headerStyle = computed(() => ({
  background: isDark.value ? 'rgba(14,21,40,0.9)' : '#f8fafc',
  borderBottom: `1px solid ${isDark.value ? 'rgba(129,140,248,0.08)' : 'rgba(0,0,0,0.08)'}`,
}))

const availableCount = computed(() => networks.value.filter(n => n.state === 'Available').length)
const pendingCount   = computed(() => networks.value.filter(n => n.state === 'Pending').length)

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
  loadError.value = null
  const ns = nsStore.active
  try {
    ;[networks.value, interfaces.value] = await Promise.all([
      api.networks.list(ns), api.networks.listInterfaces(ns)
    ])
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
    await api.networks.delete(nsStore.active, deleteTarget.value)
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
