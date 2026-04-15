<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Load Balancers</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/loadbalancers/new' }">Create Load Balancer</v-btn>
    </div>

    <v-row dense class="mb-6">
      <v-col cols="auto"><StatCard label="Public"   :value="publicCount"   color="#1a5fa8" /></v-col>
      <v-col cols="auto"><StatCard label="Internal" :value="internalCount" color="#475569" /></v-col>
      <v-col cols="auto"><StatCard label="Total"    :value="items.length" /></v-col>
    </v-row>

    <v-card v-if="items.length === 0 && !loading" variant="outlined" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-scale-balance</v-icon>
        <div class="text-h6 mb-2">No load balancers yet</div>
        <div class="text-medium-emphasis mb-4">Create your first load balancer to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/loadbalancers/new' }">Create Load Balancer</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-scale-balance</v-icon>
        Load Balancers
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table :headers="headers" :items="items" :loading="loading" density="compact" hover no-data-text="No load balancers">
        <template #item.name="{ item }">
          <router-link :to="`/loadbalancers/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
        </template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
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
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { useNamespaceStore } from '@/stores/namespace'
import StatCard from '@/components/StatCard.vue'

const nsStore  = useNamespaceStore()
const lbStore  = useLoadBalancersStore()
const items    = ref<LoadBalancer[]>([])
const loading  = ref(false)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const publicCount   = computed(() => items.value.filter(lb => lb.type === 'Public').length)
const internalCount = computed(() => items.value.filter(lb => lb.type === 'Internal').length)

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Type',    key: 'type'      },
  { title: 'IPs',     key: 'ips'       },
  { title: 'Created', key: 'createdAt' },
  { title: '',        key: 'actions', sortable: false },
]

async function load() {
  loading.value = true
  items.value   = await api.loadBalancers.list(nsStore.active)
  loading.value = false
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await lbStore.deleteLoadBalancer(deleteTarget.value)
  deleting.value = false; deleteDialog.value = false
  await load()
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
