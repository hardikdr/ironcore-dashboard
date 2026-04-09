<template>
  <v-container fluid class="pa-6">
    <!-- Page header -->
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Machines</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <div class="d-flex ga-2">
        <v-btn variant="outlined" prepend-icon="mdi-refresh" @click="store.load()">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/machines/new' }">Create Machine</v-btn>
      </div>
    </div>

    <!-- Stat cards -->
    <v-row class="mb-6" dense>
      <v-col v-for="s in stats" :key="s.label" cols="auto">
        <StatCard :value="s.value" :label="s.label" :value-color="s.color" />
      </v-col>
    </v-row>

    <!-- Search -->
    <v-text-field
      v-model="search"
      prepend-inner-icon="mdi-magnify"
      placeholder="Search machines…"
      variant="outlined"
      density="compact"
      hide-details
      class="mb-4"
      style="max-width: 320px"
    />

    <!-- Table -->
    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-monitor</v-icon>
        All Machines
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
        <v-spacer />
      </v-card-title>

      <v-data-table
        :headers="headers"
        :items="filtered"
        :loading="store.loading"
        density="compact"
        hover
        item-key="name"
        :no-data-text="store.error ?? 'No machines found'"
      >
        <template #item.state="{ item }">
          <StatusBadge :state="item.state" />
        </template>
        <template #item.machineClass="{ item }">
          <span class="font-weight-bold">{{ item.machineClass }}</span>
        </template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
        </template>
        <template #item.volumes="{ item }">
          <v-chip v-for="v in item.volumes" :key="v" size="x-small" variant="outlined" class="mr-1">{{ v }}</v-chip>
        </template>
        <template #item.actions="{ item }">
          <div class="d-flex ga-1">
            <v-btn
              size="x-small" variant="outlined"
              :icon="item.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
              :color="item.power === 'On' ? 'warning' : 'success'"
              :title="item.power === 'On' ? 'Power Off' : 'Power On'"
              @click.stop="store.setPower(item.name, item.power === 'On' ? 'Off' : 'On')"
            />
            <v-btn
              size="x-small" variant="outlined" icon="mdi-delete" color="error"
              title="Delete"
              @click.stop="confirmDelete(item.name)"
            />
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Delete confirm dialog -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title>Delete machine?</v-card-title>
        <v-card-text>
          Machine <strong>{{ pendingDelete }}</strong> will be permanently deleted.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMachinesStore } from '@/stores/machines'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'
import StatCard from '@/components/StatCard.vue'

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
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
  { title: '',             key: 'actions', sortable: false },
]

const filtered = computed(() =>
  store.items.filter(m => m.name.toLowerCase().includes(search.value.toLowerCase()))
)

const stats = computed(() => [
  { label: 'Running',  value: store.items.filter(m => m.state === 'Running').length,  color: '#16a34a' },
  { label: 'Pending',  value: store.items.filter(m => m.state === 'Pending').length,  color: '#d97706' },
  { label: 'Shutdown', value: store.items.filter(m => m.state === 'Shutdown').length, color: '#94a3b8' },
  { label: 'Total',    value: store.items.length,                                     color: '#1e293b' },
])

function confirmDelete(name: string) {
  pendingDelete.value = name
  deleteDialog.value  = true
}

async function doDelete() {
  await store.deleteMachine(pendingDelete.value)
  deleteDialog.value = false
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
