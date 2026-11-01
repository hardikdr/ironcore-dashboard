<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Volumes" subtitle="Block storage volumes" icon="mdi-database">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="store.loading" @click="store.load()">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/volumes/new' }">Create Volume</v-btn>
      </template>
    </PageHeader>

    <template v-if="store.loading && store.items.length === 0">
      <v-skeleton-loader type="table-row@6" />
    </template>

    <template v-else-if="store.error && store.items.length === 0">
      <EmptyState variant="error" title="Failed to load volumes" :subtitle="store.error">
        <v-btn color="primary" prepend-icon="mdi-refresh" @click="store.load()">Retry</v-btn>
      </EmptyState>
    </template>

    <template v-else>
      <v-row dense class="mb-4">
        <v-col cols="auto">
          <StatCard label="Available" :value="availableCount" icon="mdi-check-circle-outline" color="success" />
        </v-col>
        <v-col cols="auto">
          <StatCard label="Pending" :value="pendingCount" icon="mdi-clock-outline" color="warning" />
        </v-col>
        <v-col cols="auto">
          <StatCard label="Error" :value="errorCount" icon="mdi-alert-circle-outline" color="error" />
        </v-col>
        <v-col cols="auto">
          <StatCard label="Total" :value="store.items.length" icon="mdi-database" color="primary" />
        </v-col>
      </v-row>

      <!-- Toolbar -->
      <div class="d-flex align-center ga-3 flex-wrap mb-4">
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          placeholder="Filter by name or class…"
          variant="outlined"
          density="compact"
          hide-details
          clearable
          style="max-width:300px"
        />
        <div class="d-flex ga-2">
          <v-chip
            v-for="s in stateFilters"
            :key="s.value"
            :color="stateFilter === s.value ? s.color : undefined"
            :variant="stateFilter === s.value ? 'tonal' : 'outlined'"
            size="small"
            clickable
            @click="stateFilter = stateFilter === s.value ? null : s.value"
          >{{ s.label }}</v-chip>
        </div>
        <v-spacer />
        <span class="text-caption text-medium-emphasis">{{ filtered.length }} of {{ store.items.length }}</span>
      </div>

      <v-card :style="cardStyle" variant="flat" rounded="lg">
        <v-data-table
          :headers="headers"
          :items="filtered"
          :loading="store.loading"
          density="compact"
          hover
          item-key="name"
          :items-per-page="pageSize"
          :no-data-text="store.error ?? 'No volumes'"
        >
          <template #bottom>
            <div class="d-flex align-center justify-end ga-4 pa-3">
              <span class="text-caption text-medium-emphasis">Rows per page</span>
              <v-select
                v-model="pageSize"
                :items="[20, 50, 100]"
                density="compact"
                variant="outlined"
                hide-details
                style="max-width:80px"
              />
              <v-pagination
                v-model="page"
                :length="Math.ceil(filtered.length / pageSize)"
                density="compact"
                size="small"
                total-visible="5"
              />
            </div>
          </template>

          <template #item.name="{ item }">
            <router-link :to="`/volumes/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
          </template>
          <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
          <template #item.sizeBytes="{ item }">{{ formatBytes(item.sizeBytes) }}</template>
          <template #item.createdAt="{ item }">
            <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
          </template>
          <template #item.actions="{ item }">
            <v-btn
              size="x-small" variant="outlined" icon="mdi-delete" color="error"
              :aria-label="`Delete volume ${item.name}`"
              title="Delete volume"
              @click="confirmDelete(item.name)"
            />
          </template>

          <template #no-data>
            <EmptyState
              :variant="search || stateFilter ? 'no-results' : 'empty'"
              :title="search || stateFilter ? 'No matching volumes' : 'No volumes'"
              :subtitle="search || stateFilter ? 'Try clearing the search or filter.' : 'Create your first volume to get started.'"
            >
              <v-btn v-if="!search && !stateFilter" color="primary" prepend-icon="mdi-plus" to="/volumes/new">Create Volume</v-btn>
              <v-btn v-else variant="outlined" @click="search = ''; stateFilter = null">Clear filters</v-btn>
            </EmptyState>
          </template>
        </v-data-table>
      </v-card>
    </template>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete volume?
        </v-card-title>
        <v-card-text>Volume <strong>{{ deleteTarget }}</strong> will be permanently deleted. This action cannot be undone.</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" :disabled="deleting" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deleting" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useVolumesStore }   from '@/stores/volumes'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader   from '@/components/common/PageHeader.vue'
import StatusBadge  from '@/components/common/StatusBadge.vue'
import StatCard     from '@/components/common/StatCard.vue'
import EmptyState   from '@/components/common/EmptyState.vue'

const store   = useVolumesStore()
const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const search      = ref('')
const stateFilter = ref<string | null>(null)
const page        = ref(1)
const pageSize    = ref(20)
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const stateFilters = [
  { label: 'Available', value: 'Available', color: 'success' },
  { label: 'Pending',   value: 'Pending',   color: 'warning' },
  { label: 'Error',     value: 'Error',     color: 'error' },
]

const availableCount = computed(() => store.items.filter(v => v.state === 'Available').length)
const pendingCount   = computed(() => store.items.filter(v => v.state === 'Pending').length)
const errorCount     = computed(() => store.items.filter(v => v.state === 'Error').length)

const filtered = computed(() => {
  let items = store.items
  const q = search.value.toLowerCase()
  if (q) items = items.filter(v => v.name.toLowerCase().includes(q) || v.volumeClass.toLowerCase().includes(q))
  if (stateFilter.value) items = items.filter(v => v.state === stateFilter.value)
  return items
})

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

watch([search, stateFilter], () => { page.value = 1 })

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await store.deleteVolume(deleteTarget.value)
  deleting.value = false; deleteDialog.value = false
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
