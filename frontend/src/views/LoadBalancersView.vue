<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Load Balancers"
      subtitle="Traffic distribution endpoints"
      icon="mdi-scale-balance"
    >
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/loadbalancers/new' }">Create Load Balancer</v-btn>
      </template>
    </PageHeader>

    <v-row dense class="mb-6">
      <v-col cols="auto"><StatCard label="Public"   :value="publicCount"   icon="mdi-earth"        color="info"      /></v-col>
      <v-col cols="auto"><StatCard label="Internal" :value="internalCount" icon="mdi-lock-outline"  color="secondary" /></v-col>
      <v-col cols="auto"><StatCard label="Total"    :value="items.length"  icon="mdi-scale-balance" color="primary"   /></v-col>
    </v-row>

    <v-card v-if="items.length === 0 && !loading" :style="cardStyle" variant="flat" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-scale-balance</v-icon>
        <div class="text-h6 mb-2">No load balancers yet</div>
        <div class="text-medium-emphasis mb-4">Create your first load balancer to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/loadbalancers/new' }">Create Load Balancer</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else :style="cardStyle" variant="flat" rounded="lg">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-scale-balance</v-icon>
        <span class="font-weight-bold">Load Balancers</span>
        <v-chip size="x-small" color="primary">{{ filtered.length }}</v-chip>
        <v-spacer />
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          placeholder="Search…"
          variant="outlined"
          density="compact"
          hide-details
          clearable
          style="max-width:240px"
        />
      </div>
      <v-data-table :headers="headers" :items="filtered" :loading="loading" density="compact" hover :no-data-text="loadError ?? 'No load balancers'">
        <template #item.name="{ item }">
          <router-link :to="`/loadbalancers/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
        </template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
        </template>
        <template #item.createdAt="{ item }">
          <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
        </template>
        <template #item.actions="{ item }">
          <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" @click="confirmDelete(item.name)" />
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ deleteTarget }}"?</v-card-title>
        <v-card-text>This action cannot be undone. The load balancer will be permanently deleted from namespace <b>{{ nsStore.active }}</b>.</v-card-text>
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
import { api, type LoadBalancer } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader from '@/components/common/PageHeader.vue'
import StatCard   from '@/components/common/StatCard.vue'

const nsStore  = useNamespaceStore()
const { cardStyle, headerStyle } = useCardStyle()
const items    = ref<LoadBalancer[]>([])
const loading  = ref(false)
const loadError = ref<string | null>(null)
const search       = ref('')
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const publicCount   = computed(() => items.value.filter(lb => lb.type === 'Public').length)
const internalCount = computed(() => items.value.filter(lb => lb.type === 'Internal').length)
const filtered      = computed(() => {
  const q = search.value.toLowerCase()
  return q ? items.value.filter(i => i.name.toLowerCase().includes(q)) : items.value
})

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Type',    key: 'type'      },
  { title: 'IPs',     key: 'ips'       },
  { title: 'Created', key: 'createdAt' },
  { title: '',        key: 'actions', sortable: false },
]

async function load() {
  loading.value = true
  loadError.value = null
  try {
    items.value = await api.loadBalancers.list(nsStore.active)
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
    await api.loadBalancers.delete(nsStore.active, deleteTarget.value)
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
