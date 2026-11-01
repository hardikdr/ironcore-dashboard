<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Network Policies" subtitle="Network traffic rules for interface selectors" icon="mdi-shield-lock-outline">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading && items.length === 0">
      <v-skeleton-loader type="table-row@3" />
    </template>
    <template v-else-if="error && items.length === 0">
      <EmptyState variant="error" title="Failed to load network policies" :subtitle="error">
        <v-btn color="primary" prepend-icon="mdi-refresh" @click="load">Retry</v-btn>
      </EmptyState>
    </template>
    <template v-else>
      <v-card :style="cardStyle" variant="flat" rounded="lg">
        <v-data-table
          :headers="headers"
          :items="items"
          :loading="loading"
          density="compact"
          hover
          item-key="name"
          :items-per-page="20"
        >
          <template #item.networkRef="{ item }">
            <router-link v-if="item.networkRef" :to="`/networks/${item.networkRef}`" class="text-primary text-decoration-none">{{ item.networkRef }}</router-link>
            <span v-else class="text-medium-emphasis">—</span>
          </template>
          <template #item.policyTypes="{ item }">
            <div class="d-flex flex-wrap ga-1">
              <v-chip
                v-for="pt in item.policyTypes"
                :key="pt"
                size="x-small"
                :color="pt === 'Ingress' ? 'blue' : 'orange'"
                variant="tonal"
              >{{ pt }}</v-chip>
              <span v-if="!item.policyTypes.length" class="text-medium-emphasis">—</span>
            </div>
          </template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #no-data>
            <EmptyState variant="empty" title="No network policies" subtitle="Network policies are configured by infrastructure operators to regulate traffic." />
          </template>
        </v-data-table>
      </v-card>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type NetworkPolicy } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader from '@/components/common/PageHeader.vue'
import EmptyState from '@/components/common/EmptyState.vue'

const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const items   = ref<NetworkPolicy[]>([])
const loading = ref(false)
const error   = ref<string | null>(null)

const headers = [
  { title: 'Name',         key: 'name'        },
  { title: 'Network',      key: 'networkRef'  },
  { title: 'Policy Types', key: 'policyTypes', sortable: false },
  { title: 'Ingress Rules', key: 'ingressRules' },
  { title: 'Egress Rules',  key: 'egressRules'  },
  { title: 'Created',      key: 'createdAt'   },
]

async function load() {
  loading.value = true; error.value = null
  try { items.value = await api.networkPolicies.list(nsStore.active) }
  catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { loading.value = false }
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
