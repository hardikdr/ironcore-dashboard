<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="IP Prefixes"
      subtitle="IP address allocation ranges"
      icon="mdi-ip-network"
    >
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/prefixes/new' }">Create IP Prefix</v-btn>
      </template>
    </PageHeader>

    <v-row dense class="mb-6">
      <v-col cols="auto"><StatCard label="Allocated" :value="allocatedCount" icon="mdi-check-circle-outline" color="success" /></v-col>
      <v-col cols="auto"><StatCard label="Pending"   :value="pendingCount"   icon="mdi-clock-outline"        color="warning" /></v-col>
      <v-col cols="auto"><StatCard label="Total"     :value="items.length"   icon="mdi-ip-network"           color="primary" /></v-col>
    </v-row>

    <v-card v-if="items.length === 0 && !loading" :style="cardStyle" variant="flat" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-ip-network-outline</v-icon>
        <div class="text-h6 mb-2">No IP prefixes yet</div>
        <div class="text-medium-emphasis mb-4">Create your first IP prefix to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/prefixes/new' }">Create IP Prefix</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else :style="cardStyle" variant="flat" rounded="lg">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-ip-network</v-icon>
        <span class="font-weight-bold">IP Prefixes</span>
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
      <v-data-table :headers="headers" :items="filtered" :loading="loading" density="compact" hover :no-data-text="loadError ?? 'No IP prefixes'">
        <template #item.name="{ item }">
          <router-link :to="`/prefixes/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
        </template>
        <template #item.prefix="{ item }">
          <code class="text-primary">{{ item.prefix || '—' }}</code>
        </template>
        <template #item.phase="{ item }"><StatusBadge :state="item.phase" /></template>
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
        <v-card-text>This action cannot be undone. The IP prefix will be permanently deleted from namespace <b>{{ nsStore.active }}</b>.</v-card-text>
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
import { api, type Prefix } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader  from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import StatCard    from '@/components/common/StatCard.vue'

const nsStore      = useNamespaceStore()
const { cardStyle, headerStyle } = useCardStyle()
const items        = ref<Prefix[]>([])
const loading      = ref(false)
const loadError    = ref<string | null>(null)
const search       = ref('')
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const allocatedCount = computed(() => items.value.filter(p => p.phase === 'Allocated').length)
const pendingCount   = computed(() => items.value.filter(p => p.phase === 'Pending').length)
const filtered       = computed(() => {
  const q = search.value.toLowerCase()
  return q ? items.value.filter(i => i.name.toLowerCase().includes(q) || (i.prefix ?? '').toLowerCase().includes(q)) : items.value
})

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Prefix',  key: 'prefix'    },
  { title: 'Phase',   key: 'phase'     },
  { title: 'Created', key: 'createdAt' },
  { title: '',        key: 'actions', sortable: false },
]

async function load() {
  loading.value = true
  loadError.value = null
  try {
    items.value = await api.prefixes.list(nsStore.active)
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
    await api.prefixes.delete(nsStore.active, deleteTarget.value)
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
