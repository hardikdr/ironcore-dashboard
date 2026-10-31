<template>
  <v-container fluid class="pa-6">
    <PageHeader title="Machines" subtitle="Virtual machine instances" icon="mdi-server">
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="store.loading" @click="store.load()">Refresh</v-btn>
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

    <template v-else>
      <v-row class="mb-6" dense>
        <v-col v-for="s in stats" :key="s.label" cols="auto">
          <StatCard :value="s.value" :label="s.label" :icon="s.icon" :color="s.color" />
        </v-col>
      </v-row>

      <v-text-field
        v-model="search"
        prepend-inner-icon="mdi-magnify"
        placeholder="Search by name, class, image…"
        variant="outlined"
        density="compact"
        hide-details
        class="mb-4"
        clearable
        style="max-width: 360px"
      />

      <v-card :style="cardStyle" variant="flat" rounded="lg">
        <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
          <v-icon color="primary">mdi-server</v-icon>
          <span class="font-weight-bold">All Machines</span>
          <v-chip size="x-small" color="primary">{{ filtered.length }}</v-chip>
          <v-spacer />
          <span v-if="filtered.length !== store.items.length" class="text-caption text-medium-emphasis">
            {{ filtered.length }} of {{ store.items.length }}
          </span>
        </div>

        <v-data-table
          :headers="headers"
          :items="filtered"
          :loading="store.loading"
          density="compact"
          hover
          item-key="name"
          :no-data-text="store.error ?? 'No machines found'"
        >
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
            <div class="d-flex ga-1">
              <v-btn
                size="x-small" variant="outlined"
                :icon="item.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
                :color="item.power === 'On' ? 'warning' : 'success'"
                :aria-label="item.power === 'On' ? `Power off ${item.name}` : `Power on ${item.name}`"
                :title="item.power === 'On' ? 'Power off' : 'Power on'"
                @click.stop="store.setPower(item.name, item.power === 'On' ? 'Off' : 'On')"
              />
              <v-btn
                size="x-small" variant="outlined" icon="mdi-delete" color="error"
                :aria-label="`Delete machine ${item.name}`"
                title="Delete machine"
                @click.stop="confirmDelete(item.name)"
              />
            </div>
          </template>
        </v-data-table>
      </v-card>
    </template>

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

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
const { cardStyle, headerStyle } = useCardStyle()
const search  = ref('')
const deleteDialog  = ref(false)
const pendingDelete = ref('')

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
  const q = search.value.toLowerCase()
  if (!q) return store.items
  return store.items.filter(m =>
    m.name.toLowerCase().includes(q) ||
    m.machineClass.toLowerCase().includes(q) ||
    m.image.toLowerCase().includes(q)
  )
})

const stats = computed(() => {
  const counts = store.items.reduce((acc, m) => {
    const s = (m.state ?? '').toLowerCase()
    if (s === 'running')  acc.running++
    else if (s === 'pending')  acc.pending++
    else if (s === 'shutdown') acc.shutdown++
    return acc
  }, { running: 0, pending: 0, shutdown: 0 })
  return [
    { label: 'Running',  value: counts.running,  icon: 'mdi-check-circle-outline', color: 'success' },
    { label: 'Pending',  value: counts.pending,  icon: 'mdi-clock-outline',        color: 'warning' },
    { label: 'Shutdown', value: counts.shutdown, icon: 'mdi-power-off',            color: 'secondary' },
    { label: 'Total',    value: store.items.length, icon: 'mdi-server',            color: 'primary' },
  ]
})

function confirmDelete(name: string) { pendingDelete.value = name; deleteDialog.value = true }
async function doDelete() { await store.deleteMachine(pendingDelete.value); deleteDialog.value = false }

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
