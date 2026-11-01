<template>
  <v-container fluid class="pa-6">
    <PageHeader
      :title="machine ? machine.name : machineName"
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
          :loading="powerLoading"
          @click="togglePower"
        >
          {{ machine.power === 'On' ? 'Power Off' : 'Power On' }}
        </v-btn>
      </template>
    </PageHeader>

    <!-- Skeleton while loading -->
    <template v-if="loading">
      <v-row class="mb-4">
        <v-col v-for="i in 3" :key="i" cols="auto">
          <v-skeleton-loader type="chip" width="120" />
        </v-col>
      </v-row>
      <v-skeleton-loader type="table-row@6" />
    </template>

    <div v-else-if="error" class="pa-8 text-center">
      <v-icon size="48" color="error" class="mb-3">mdi-alert-circle-outline</v-icon>
      <div class="text-h6 mb-2">Failed to load machine</div>
      <div class="text-medium-emphasis mb-4">{{ error }}</div>
      <v-btn color="primary" @click="load">Retry</v-btn>
    </div>

    <template v-else-if="machine">
      <!-- Status chips -->
      <div class="d-flex flex-wrap ga-2 mb-4">
        <StatusBadge :state="machine.state" />
        <v-chip size="small" variant="tonal" color="info" prepend-icon="mdi-chip">{{ machine.machineClass }}</v-chip>
        <CopyChip v-for="ip in machine.ips" :key="ip" :text="ip" color="accent" />
        <v-chip v-if="machine.image" size="small" variant="outlined" prepend-icon="mdi-disc">
          <span style="max-width:240px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap">{{ machine.image }}</span>
        </v-chip>
      </div>

      <!-- Tabs -->
      <v-tabs v-model="tab" color="primary" class="mb-4">
        <v-tab value="overview">
          <v-icon size="16" class="mr-1">mdi-information-outline</v-icon>Overview
        </v-tab>
        <v-tab value="terminal">
          <v-icon size="16" class="mr-1">mdi-console</v-icon>Terminal
        </v-tab>
        <v-tab value="networking">
          <v-icon size="16" class="mr-1">mdi-lan</v-icon>Networking
        </v-tab>
        <v-tab value="storage">
          <v-icon size="16" class="mr-1">mdi-database-outline</v-icon>Storage
        </v-tab>
        <v-tab value="events">
          <v-icon size="16" class="mr-1">mdi-clock-outline</v-icon>Events
        </v-tab>
        <v-tab value="conditions">
          <v-icon size="16" class="mr-1">mdi-check-all</v-icon>Conditions
        </v-tab>
        <v-tab value="yaml">
          <v-icon size="16" class="mr-1">mdi-code-braces</v-icon>YAML
        </v-tab>
      </v-tabs>

      <v-window v-model="tab">
        <!-- ── Overview ──────────────────────────────────────── -->
        <v-window-item value="overview">
          <v-row>
            <v-col cols="12" md="6">
              <v-card :style="cardStyle" variant="flat" class="pa-4">
                <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
                  <v-icon size="16" color="primary">mdi-information-outline</v-icon>
                  Machine Info
                </div>
                <v-table density="compact">
                  <tbody>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1" width="140">Name</td>
                      <td><CopyChip :text="machine.name" color="default" variant="text" /></td>
                    </tr>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1">Namespace</td>
                      <td>{{ machine.namespace }}</td>
                    </tr>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1">Machine Class</td>
                      <td><code class="text-primary">{{ machine.machineClass }}</code></td>
                    </tr>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1">Image</td>
                      <td><code class="text-caption" style="word-break:break-all">{{ machine.image || '—' }}</code></td>
                    </tr>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1">Power</td>
                      <td>
                        <v-chip :color="machine.power === 'On' ? 'success' : 'default'" size="x-small" variant="tonal">
                          {{ machine.power }}
                        </v-chip>
                      </td>
                    </tr>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1">State</td>
                      <td><StatusBadge :state="machine.state" /></td>
                    </tr>
                    <tr>
                      <td class="text-medium-emphasis pr-4 py-1">Created</td>
                      <td :title="absoluteTime(machine.createdAt)">{{ relativeTime(machine.createdAt) }}</td>
                    </tr>
                  </tbody>
                </v-table>
              </v-card>
            </v-col>

            <v-col cols="12" md="6">
              <v-card :style="cardStyle" variant="flat" class="pa-4">
                <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
                  <v-icon size="16" color="secondary">mdi-database</v-icon>
                  Volume Attachments
                </div>
                <div v-if="machine.volumes.length === 0" class="text-medium-emphasis text-caption">No volumes attached</div>
                <div v-for="vol in machine.volumes" :key="vol" class="d-flex align-center ga-2 mb-2">
                  <v-icon size="14" color="secondary">mdi-database</v-icon>
                  <router-link :to="`/volumes/${vol}`" class="text-decoration-none text-primary">
                    <code>{{ vol }}</code>
                  </router-link>
                </div>
              </v-card>

              <!-- Annotations card -->
              <v-card v-if="machine.annotations && Object.keys(machine.annotations).length" :style="cardStyle" variant="flat" class="pa-4 mt-4">
                <div class="text-subtitle-2 font-weight-bold mb-3 d-flex align-center ga-2">
                  <v-icon size="16" color="accent">mdi-tag-outline</v-icon>
                  Annotations
                </div>
                <div v-for="(val, key) in machine.annotations" :key="key" class="d-flex ga-2 mb-1 text-caption">
                  <code class="text-medium-emphasis">{{ key }}</code>
                  <span class="text-primary">{{ val }}</span>
                </div>
              </v-card>
            </v-col>
          </v-row>
        </v-window-item>

        <!-- ── Terminal ──────────────────────────────────────── -->
        <v-window-item value="terminal">
          <v-card :style="cardStyle" variant="flat" class="overflow-hidden">
            <v-alert
              v-if="machine.state !== 'Running'"
              type="warning"
              variant="tonal"
              class="ma-4"
              density="compact"
            >
              Machine is not Running (state: {{ machine.state }}). Console may be unavailable.
            </v-alert>
            <div style="height: 520px; padding: 8px">
              <MachineTerminal :ws-url="consoleUrl" :active="tab === 'terminal'" />
            </div>
          </v-card>
        </v-window-item>

        <!-- ── Networking ────────────────────────────────────── -->
        <v-window-item value="networking">
          <v-card :style="cardStyle" variant="flat" class="pa-4">
            <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
              <v-icon size="16" color="info">mdi-lan</v-icon>
              Network Interfaces
            </div>
            <div v-if="nicsLoading" class="d-flex justify-center pa-8">
              <v-progress-circular indeterminate size="24" color="primary" />
            </div>
            <div v-else-if="nics.length === 0" class="text-medium-emphasis">No network interfaces found.</div>
            <v-table v-else density="compact">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Network</th>
                  <th>IPs</th>
                  <th>State</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="nic in nics" :key="nic.name">
                  <td><code>{{ nic.name }}</code></td>
                  <td>
                    <router-link v-if="nic.network" :to="`/networks/${nic.network}`" class="text-decoration-none text-primary">
                      {{ nic.network }}
                    </router-link>
                    <span v-else class="text-medium-emphasis">—</span>
                  </td>
                  <td>
                    <CopyChip v-for="ip in nic.ips" :key="ip" :text="ip" color="info" class="mr-1" />
                    <span v-if="!nic.ips.length" class="text-medium-emphasis">—</span>
                  </td>
                  <td><StatusBadge :state="nic.state" /></td>
                </tr>
              </tbody>
            </v-table>
          </v-card>
        </v-window-item>

        <!-- ── Storage ───────────────────────────────────────── -->
        <v-window-item value="storage">
          <v-card :style="cardStyle" variant="flat" class="pa-4">
            <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
              <v-icon size="16" color="secondary">mdi-database-outline</v-icon>
              Volume Attachments
            </div>
            <div v-if="machine.volumes.length === 0" class="text-medium-emphasis">No volumes attached.</div>
            <v-table v-else density="compact">
              <thead>
                <tr><th>Name</th><th>Actions</th></tr>
              </thead>
              <tbody>
                <tr v-for="vol in machine.volumes" :key="vol">
                  <td>
                    <router-link :to="`/volumes/${vol}`" class="text-primary text-decoration-none font-weight-medium">
                      <v-icon size="14" color="secondary" class="mr-1">mdi-database</v-icon>{{ vol }}
                    </router-link>
                  </td>
                  <td>
                    <v-btn size="x-small" variant="text" :to="`/volumes/${vol}`" icon="mdi-open-in-new" />
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-card>
        </v-window-item>

        <!-- ── Events ───────────────────────────────────────── -->
        <v-window-item value="events">
          <v-card :style="cardStyle" variant="flat" class="pa-4">
            <ResourceEvents :resource-name="machineName" resource-kind="Machine" />
          </v-card>
        </v-window-item>

        <!-- ── Conditions ──────────────────────────────────── -->
        <v-window-item value="conditions">
          <v-card :style="cardStyle" variant="flat" class="pa-4">
            <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
              <v-icon size="16" color="primary">mdi-check-all</v-icon>
              Status Conditions
            </div>
            <StatusConditions :conditions="machine?.conditions" />
          </v-card>
        </v-window-item>

        <!-- ── YAML ─────────────────────────────────────────── -->
        <v-window-item value="yaml">
          <ResourceYaml v-if="machine" :resource="machine as Record<string, unknown>" :filename="machineName" />
        </v-window-item>
      </v-window>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, defineAsyncComponent } from 'vue'
import { useRoute } from 'vue-router'
import { api, wsConsoleUrl, type Machine, type NetworkInterface } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useMachinesStore } from '@/stores/machines'
import { useCardStyle } from '@/composables/useCardStyle'
import { relativeTime, absoluteTime } from '@/composables/useTime'
import PageHeader      from '@/components/common/PageHeader.vue'
import StatusBadge     from '@/components/common/StatusBadge.vue'
import CopyChip        from '@/components/common/CopyChip.vue'
import ResourceEvents    from '@/components/common/ResourceEvents.vue'
import ResourceYaml      from '@/components/common/ResourceYaml.vue'
import StatusConditions  from '@/components/common/StatusConditions.vue'

// Dynamic import — xterm.js only loads when this component renders
const MachineTerminal = defineAsyncComponent(
  () => import('@/components/terminal/MachineTerminal.vue')
)

const route   = useRoute()
const nsStore = useNamespaceStore()
const machinesStore = useMachinesStore()
const { cardStyle } = useCardStyle()

const machineName = route.params.name as string
const tab         = ref('overview')
const machine     = ref<Machine | null>(null)
const loading     = ref(true)
const error       = ref<string | null>(null)
const powerLoading = ref(false)
const nics        = ref<NetworkInterface[]>([])
const nicsLoading = ref(false)

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

async function loadNics() {
  nicsLoading.value = true
  try {
    const all = await api.networks.listInterfaces(nsStore.active)
    nics.value = all.filter(n => n.machine === machineName)
  } catch { /* NICs are best-effort */ }
  finally { nicsLoading.value = false }
}

// Load NICs when networking tab is first opened
watch(tab, (t) => { if (t === 'networking' && nics.value.length === 0) loadNics() })

async function togglePower() {
  if (!machine.value) return
  powerLoading.value = true
  try {
    const next = machine.value.power === 'On' ? 'Off' : 'On'
    await machinesStore.setPower(machineName, next)
    await load()
  } finally {
    powerLoading.value = false
  }
}

onMounted(load)
</script>
