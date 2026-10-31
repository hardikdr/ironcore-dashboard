<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Machines"
      subtitle="Virtual machine instances"
      icon="mdi-server"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" @click="store.load()">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/machines/new' }">Create Machine</v-btn>
      </template>
    </PageHeader>

    <v-row class="mb-6" dense>
      <v-col v-for="s in stats" :key="s.label" cols="auto">
        <StatCard :value="s.value" :label="s.label" :icon="s.icon" :color="s.color" />
      </v-col>
    </v-row>

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

    <v-card :style="cardStyle" variant="flat" rounded="lg">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-server</v-icon>
        <span class="font-weight-bold">All Machines</span>
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
        <v-spacer />
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
              @click.stop="store.setPower(item.name, item.power === 'On' ? 'Off' : 'On')"
            />
            <v-btn size="x-small" variant="outlined" icon="mdi-delete" color="error" @click.stop="confirmDelete(item.name)" />
          </div>
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title>Delete machine?</v-card-title>
        <v-card-text>Machine <strong>{{ pendingDelete }}</strong> will be permanently deleted.</v-card-text>
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
import { useTheme } from 'vuetify'
import { useMachinesStore } from '@/stores/machines'
import { useNamespaceStore } from '@/stores/namespace'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import StatCard from '@/components/common/StatCard.vue'

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
const theme   = useTheme()
const search  = ref('')
const deleteDialog  = ref(false)
const pendingDelete = ref('')

const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() => ({
  background: isDark.value ? 'rgba(20,30,51,0.8)' : '#ffffff',
  border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.12)'}`,
}))
const headerStyle = computed(() => ({
  background: isDark.value ? 'rgba(14,21,40,0.9)' : '#f8fafc',
  borderBottom: `1px solid ${isDark.value ? 'rgba(129,140,248,0.08)' : 'rgba(0,0,0,0.08)'}`,
}))

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
  { label: 'Running',  value: store.items.filter(m => m.state === 'Running').length,  icon: 'mdi-check-circle-outline', color: 'success' },
  { label: 'Pending',  value: store.items.filter(m => m.state === 'Pending').length,  icon: 'mdi-clock-outline',        color: 'warning' },
  { label: 'Shutdown', value: store.items.filter(m => m.state === 'Shutdown').length, icon: 'mdi-power-off',            color: 'secondary' },
  { label: 'Total',    value: store.items.length,                                     icon: 'mdi-server',               color: 'primary' },
])

function confirmDelete(name: string) { pendingDelete.value = name; deleteDialog.value = true }
async function doDelete() { await store.deleteMachine(pendingDelete.value); deleteDialog.value = false }

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
