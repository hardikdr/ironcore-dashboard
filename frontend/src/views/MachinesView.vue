<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Machines" subtitle="Virtual machine instances" icon="mdi-server">
      <template #actions>
        <span v-if="lastFetch" class="text-caption text-medium-emphasis align-self-center">
          Updated {{ relativeTime(lastFetch) }}
        </span>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="store.loading" @click="refresh">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/machines/new' }">Create Machine</v-btn>
      </template>
    </PageHeader>

    <!-- Skeleton during initial load -->
    <template v-if="store.loading && store.items.length === 0">
      <v-row class="mb-6" dense>
        <v-col v-for="i in 4" :key="i" cols="auto">
          <v-skeleton-loader type="card" width="140" height="110" />
        </v-col>
      </v-row>
      <v-skeleton-loader type="table-row@8" />
    </template>

    <!-- Error state -->
    <template v-else-if="store.error && store.items.length === 0">
      <EmptyState variant="error" title="Failed to load machines" :subtitle="store.error">
        <v-btn color="primary" prepend-icon="mdi-refresh" @click="store.load()">Retry</v-btn>
      </EmptyState>
    </template>

    <template v-else>
      <!-- Stat cards -->
      <v-row class="mb-4" dense>
        <v-col v-for="s in stats" :key="s.label" cols="auto">
          <StatCard :value="s.value" :label="s.label" :icon="s.icon" :color="s.color" />
        </v-col>
      </v-row>

      <!-- Toolbar: search + filters + bulk actions -->
      <div class="d-flex align-center ga-3 flex-wrap mb-4">
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          placeholder="Filter by name, class, image…"
          variant="outlined"
          density="compact"
          hide-details
          clearable
          style="max-width: 300px"
        />

        <!-- State filter chips -->
        <div class="d-flex ga-2 flex-wrap" role="group" aria-label="Filter by state">
          <v-chip
            v-for="state in stateFilters"
            :key="state.value"
            :color="stateFilter === state.value ? state.color : undefined"
            :variant="stateFilter === state.value ? 'tonal' : 'outlined'"
            size="small"
            clickable
            :aria-pressed="stateFilter === state.value ? 'true' : 'false'"
            @click="stateFilter = stateFilter === state.value ? null : state.value"
          >
            {{ state.label }}
          </v-chip>
        </div>

        <!-- Bulk action bar (shown when items are selected) -->
        <template v-if="selected.length > 0">
          <v-divider vertical class="mx-2" />
          <span class="text-caption text-medium-emphasis">{{ selected.length }} selected</span>
          <v-btn
            size="small" variant="outlined" color="warning" prepend-icon="mdi-power"
            @click="bulkPower('Off')"
          >Power Off</v-btn>
          <v-btn
            size="small" variant="outlined" color="success" prepend-icon="mdi-power"
            @click="bulkPower('On')"
          >Power On</v-btn>
          <v-btn
            size="small" variant="outlined" color="error" prepend-icon="mdi-delete"
            @click="bulkDeleteDialog = true"
          >Delete</v-btn>
        </template>

        <v-spacer />
        <span class="text-caption text-medium-emphasis">{{ filtered.length }} of {{ store.items.length }}</span>
      </div>

      <v-card :style="cardStyle" variant="flat" rounded="lg">
        <v-data-table
          v-model="selected"
          :headers="headers"
          :items="filtered"
          :loading="store.loading"
          density="compact"
          hover
          item-key="name"
          show-select
          :items-per-page="pageSize"
          :no-data-text="store.error ?? 'No machines found'"
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
            <router-link :to="`/machines/${item.name}`" class="text-primary text-decoration-none font-weight-medium">{{ item.name }}</router-link>
          </template>
          <template #item.state="{ item }">
            <StatusBadge :state="item.state" />
          </template>
          <template #item.machineClass="{ item }">
            <span class="font-weight-bold">{{ item.machineClass }}</span>
          </template>
          <template #item.ips="{ item }">
            <CopyChip v-for="ip in item.ips" :key="ip" :text="ip" color="info" class="mr-1" />
          </template>
          <template #item.volumes="{ item }">
            <v-chip v-for="v in item.volumes" :key="v" size="x-small" variant="outlined" class="mr-1">{{ v }}</v-chip>
          </template>
          <template #item.createdAt="{ item }">
            <span :title="absoluteTime(item.createdAt)" class="text-caption text-medium-emphasis">
              {{ relativeTime(item.createdAt) }}
            </span>
          </template>
          <template #item.actions="{ item }">
            <v-menu location="bottom end">
              <template #activator="{ props: menuProps }">
                <v-btn
                  v-bind="menuProps"
                  size="x-small"
                  variant="text"
                  icon="mdi-dots-vertical"
                  :aria-label="`Actions for ${item.name}`"
                  @click.stop
                />
              </template>
              <v-list density="compact">
                <v-list-item
                  :prepend-icon="item.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
                  :title="item.power === 'On' ? 'Power off' : 'Power on'"
                  @click="store.setPower(item.name, item.power === 'On' ? 'Off' : 'On')"
                />
                <v-list-item
                  prepend-icon="mdi-delete"
                  title="Delete"
                  class="text-error"
                  @click="confirmDelete(item.name)"
                />
              </v-list>
            </v-menu>
          </template>

          <!-- Empty state for filtered results -->
          <template #no-data>
            <EmptyState
              :variant="search || stateFilter ? 'no-results' : 'empty'"
              :title="search || stateFilter ? 'No matching machines' : 'No machines'"
              :subtitle="search || stateFilter
                ? 'Try clearing the search or filter.'
                : 'Create your first machine to get started.'"
            >
              <v-btn v-if="!search && !stateFilter" color="primary" prepend-icon="mdi-plus" to="/machines/new">Create Machine</v-btn>
              <v-btn v-else variant="outlined" @click="search = ''; stateFilter = null">Clear filters</v-btn>
            </EmptyState>
          </template>
        </v-data-table>
      </v-card>
    </template>

    <!-- Single delete dialog -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete machine?
        </v-card-title>
        <v-card-text>Machine <strong>{{ pendingDelete }}</strong> will be permanently deleted.</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Bulk delete dialog -->
    <v-dialog v-model="bulkDeleteDialog" max-width="480">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete {{ selected.length }} machines?
        </v-card-title>
        <v-card-text>
          This will permanently delete all {{ selected.length }} selected machines. This action cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" @click="bulkDeleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="doBulkDelete">Delete All</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMachinesStore }  from '@/stores/machines'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime, absoluteTime } from '@/composables/useTime'
import PageHeader   from '@/components/common/PageHeader.vue'
import StatusBadge  from '@/components/common/StatusBadge.vue'
import StatCard     from '@/components/common/StatCard.vue'
import CopyChip     from '@/components/common/CopyChip.vue'
import EmptyState   from '@/components/common/EmptyState.vue'

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const search      = ref('')
const stateFilter = ref<string | null>(null)
const page        = ref(1)
const pageSize    = ref(20)
const selected    = ref<string[]>([])
const lastFetch   = ref<string | null>(null)

const deleteDialog     = ref(false)
const bulkDeleteDialog = ref(false)
const pendingDelete    = ref('')

const stateFilters = [
  { label: 'Running',     value: 'Running',     color: 'success' },
  { label: 'Pending',     value: 'Pending',     color: 'warning' },
  { label: 'Terminating', value: 'Terminating', color: 'error' },
]

const headers = [
  { title: 'Name',         key: 'name'         },
  { title: 'Status',       key: 'state'        },
  { title: 'Size',         key: 'machineClass' },
  { title: 'Image',        key: 'image'        },
  { title: 'Volumes',      key: 'volumes'      },
  { title: 'IP Addresses', key: 'ips'          },
  { title: 'Created',      key: 'createdAt'    },
  { title: '',             key: 'actions', sortable: false },
]

const filtered = computed(() => {
  let items = store.items
  const q = search.value.toLowerCase()
  if (q) {
    items = items.filter(m =>
      m.name.toLowerCase().includes(q) ||
      m.machineClass.toLowerCase().includes(q) ||
      m.image.toLowerCase().includes(q)
    )
  }
  if (stateFilter.value) {
    items = items.filter(m => m.state === stateFilter.value)
  }
  return items
})

const stats = computed(() => {
  const counts = store.items.reduce((acc, m) => {
    const s = (m.state ?? '').toLowerCase()
    if (s === 'running')      acc.running++
    else if (s === 'pending') acc.pending++
    else                      acc.other++
    return acc
  }, { running: 0, pending: 0, other: 0 })
  return [
    { label: 'Running',  value: counts.running, icon: 'mdi-check-circle-outline', color: 'success' },
    { label: 'Pending',  value: counts.pending, icon: 'mdi-clock-outline',        color: 'warning' },
    { label: 'Other',    value: counts.other,   icon: 'mdi-power-off',            color: 'secondary' },
    { label: 'Total',    value: store.items.length, icon: 'mdi-server',           color: 'primary' },
  ]
})

// Reset to page 1 when filter changes
watch([search, stateFilter], () => { page.value = 1 })

async function refresh() {
  await store.load()
  lastFetch.value = new Date().toISOString()
}

function confirmDelete(name: string) { pendingDelete.value = name; deleteDialog.value = true }
async function doDelete() {
  await store.deleteMachine(pendingDelete.value)
  deleteDialog.value = false
}

async function bulkPower(state: 'On' | 'Off') {
  await Promise.allSettled(selected.value.map(name => store.setPower(name, state)))
  selected.value = []
}

async function doBulkDelete() {
  await Promise.allSettled(selected.value.map(name => store.deleteMachine(name)))
  selected.value = []
  bulkDeleteDialog.value = false
}

onMounted(async () => { await store.load(); lastFetch.value = new Date().toISOString() })
watch(() => nsStore.active, async () => { await store.load(); lastFetch.value = new Date().toISOString() })
</script>
