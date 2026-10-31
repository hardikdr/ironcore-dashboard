<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Machine Detail"
      icon="mdi-server"
      :breadcrumbs="[{ text: 'Machines', to: '/machines' }, { text: machineName }]"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-arrow-left" to="/machines">Back</v-btn>
        <v-btn
          v-if="machine"
          variant="outlined"
          :prepend-icon="machine.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
          :color="machine.power === 'On' ? 'warning' : 'success'"
          @click="togglePower"
        >
          {{ machine.power === 'On' ? 'Power Off' : 'Power On' }}
        </v-btn>
      </template>
    </PageHeader>

    <div v-if="loading" class="d-flex justify-center pa-12">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <div v-else-if="error" class="pa-8 text-center text-error">{{ error }}</div>

    <template v-else-if="machine">
      <!-- Status chips -->
      <div class="d-flex flex-wrap ga-2 mb-4">
        <StatusBadge :state="machine.state" />
        <v-chip size="small" variant="tonal" color="info" prepend-icon="mdi-chip">{{ machine.machineClass }}</v-chip>
        <v-chip v-for="ip in machine.ips" :key="ip" size="small" variant="tonal" color="accent" prepend-icon="mdi-ip">{{ ip }}</v-chip>
      </div>

      <!-- Tabs -->
      <v-tabs v-model="tab" color="primary" class="mb-4">
        <v-tab value="overview">Overview</v-tab>
        <v-tab value="terminal">Terminal</v-tab>
        <v-tab value="networking">Networking</v-tab>
        <v-tab value="storage">Storage</v-tab>
      </v-tabs>

      <v-window v-model="tab">
        <!-- Overview -->
        <v-window-item value="overview">
          <v-row>
            <v-col cols="12" md="6">
              <v-card :style="cardStyle" variant="flat" class="pa-4">
                <div class="text-subtitle-2 font-weight-bold mb-4">Machine Info</div>
                <v-table density="compact">
                  <tbody>
                    <tr><td class="text-medium-emphasis pr-4 py-1">Name</td><td>{{ machine.name }}</td></tr>
                    <tr><td class="text-medium-emphasis pr-4 py-1">Namespace</td><td>{{ machine.namespace }}</td></tr>
                    <tr><td class="text-medium-emphasis pr-4 py-1">Machine Class</td><td><code>{{ machine.machineClass }}</code></td></tr>
                    <tr><td class="text-medium-emphasis pr-4 py-1">Image</td><td><code class="text-caption">{{ machine.image }}</code></td></tr>
                    <tr><td class="text-medium-emphasis pr-4 py-1">Power</td><td>{{ machine.power }}</td></tr>
                    <tr><td class="text-medium-emphasis pr-4 py-1">Created</td><td>{{ formatDate(machine.createdAt) }}</td></tr>
                  </tbody>
                </v-table>
              </v-card>
            </v-col>
            <v-col cols="12" md="6">
              <v-card :style="cardStyle" variant="flat" class="pa-4">
                <div class="text-subtitle-2 font-weight-bold mb-4">Volume Attachments</div>
                <div v-if="machine.volumes.length === 0" class="text-medium-emphasis text-caption">No volumes attached</div>
                <div v-for="vol in machine.volumes" :key="vol" class="d-flex align-center ga-2 mb-2">
                  <v-icon size="14" color="secondary">mdi-database</v-icon>
                  <code>{{ vol }}</code>
                </div>
              </v-card>
            </v-col>
          </v-row>
        </v-window-item>

        <!-- Terminal -->
        <v-window-item value="terminal">
          <v-card :style="cardStyle" variant="flat" class="overflow-hidden">
            <v-alert
              v-if="machine.state !== 'Running'"
              type="warning"
              variant="tonal"
              class="ma-4"
              density="compact"
            >
              Machine is not Running (state: {{ machine.state }}). Console may not be available.
            </v-alert>
            <div style="height: 480px; padding: 8px">
              <MachineTerminal :ws-url="consoleUrl" />
            </div>
          </v-card>
        </v-window-item>

        <!-- Networking -->
        <v-window-item value="networking">
          <v-card :style="cardStyle" variant="flat" class="pa-4">
            <div class="text-subtitle-2 font-weight-bold mb-4">IP Addresses</div>
            <div v-if="machine.ips.length === 0" class="text-medium-emphasis">No IP addresses assigned.</div>
            <v-chip
              v-for="ip in machine.ips"
              :key="ip"
              color="info"
              variant="tonal"
              class="mr-2 mb-2"
              prepend-icon="mdi-ip"
            >
              {{ ip }}
            </v-chip>
          </v-card>
        </v-window-item>

        <!-- Storage -->
        <v-window-item value="storage">
          <v-card :style="cardStyle" variant="flat" class="pa-4">
            <div class="text-subtitle-2 font-weight-bold mb-4">Volume Attachments</div>
            <div v-if="machine.volumes.length === 0" class="text-medium-emphasis">No volumes attached.</div>
            <v-chip
              v-for="vol in machine.volumes"
              :key="vol"
              variant="outlined"
              class="mr-2 mb-2"
              prepend-icon="mdi-database"
            >
              {{ vol }}
            </v-chip>
          </v-card>
        </v-window-item>
      </v-window>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from 'vuetify'
import { api, wsConsoleUrl, type Machine } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useMachinesStore } from '@/stores/machines'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import MachineTerminal from '@/components/terminal/MachineTerminal.vue'

const route   = useRoute()
const nsStore = useNamespaceStore()
const machinesStore = useMachinesStore()
const theme   = useTheme()

const machineName = route.params.name as string
const tab     = ref('overview')
const machine = ref<Machine | null>(null)
const loading = ref(true)
const error   = ref<string | null>(null)

const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() => ({
  background: isDark.value ? 'rgba(20,30,51,0.8)' : '#ffffff',
  border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.12)'}`,
}))

const consoleUrl = computed(() => wsConsoleUrl(nsStore.active, machineName))

async function load() {
  loading.value = true
  error.value = null
  try {
    machine.value = await api.machines.get(nsStore.active, machineName)
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function togglePower() {
  if (!machine.value) return
  const next = machine.value.power === 'On' ? 'Off' : 'On'
  await machinesStore.setPower(machineName, next)
  await load()
}

function formatDate(s: string) {
  try { return new Date(s).toLocaleString() } catch { return s }
}

onMounted(load)
</script>
