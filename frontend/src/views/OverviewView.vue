<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Overview"
      subtitle="IronCore infrastructure at a glance"
      icon="mdi-view-dashboard-outline"
    />

    <!-- ── Needs Attention ────────────────────────────────────────── -->
    <NeedsAttention :items="attentionItems" class="mb-6" />

    <!-- ── Stats Row ────────────────────────────────────────────── -->
    <v-row class="mb-6" dense>
      <v-col v-for="s in stats" :key="s.label" cols="6" sm="4" md="2">
        <StatCard v-bind="s" />
      </v-col>
    </v-row>

    <!-- ── Charts + Activity ─────────────────────────────────────── -->
    <v-row class="mb-6">
      <!-- Machine state donut -->
      <v-col cols="12" md="4">
        <v-card :style="cardStyle" class="pa-4" height="280">
          <div class="text-subtitle-2 font-weight-bold mb-4">Machine States</div>
          <div style="height:190px">
            <DonutChart
              :segments="machineSegments"
              center-label="Total"
              :center-value="machinesStore.items.length"
            />
          </div>
          <!-- Legend -->
          <div class="d-flex flex-wrap ga-3 mt-3">
            <div v-for="s in machineSegments" :key="s.label" class="d-flex align-center ga-1">
              <div :style="{ width:'8px', height:'8px', borderRadius:'2px', background: s.color }" />
              <span class="text-caption">{{ s.label }} ({{ s.value }})</span>
            </div>
          </div>
        </v-card>
      </v-col>

      <!-- Volume usage -->
      <v-col cols="12" md="4">
        <v-card :style="cardStyle" class="pa-4" height="280">
          <div class="text-subtitle-2 font-weight-bold mb-4">Storage Overview</div>
          <div style="height:190px">
            <DonutChart
              :segments="volumeSegments"
              center-label="Volumes"
              :center-value="volumesStore.items.length"
            />
          </div>
          <div class="d-flex flex-wrap ga-3 mt-3">
            <div v-for="s in volumeSegments" :key="s.label" class="d-flex align-center ga-1">
              <div :style="{ width:'8px', height:'8px', borderRadius:'2px', background: s.color }" />
              <span class="text-caption">{{ s.label }} ({{ s.value }})</span>
            </div>
          </div>
        </v-card>
      </v-col>

      <!-- Network overview -->
      <v-col cols="12" md="4">
        <v-card :style="cardStyle" class="pa-4" height="280">
          <div class="text-subtitle-2 font-weight-bold mb-4">Networking</div>
          <div class="d-flex flex-column ga-3 mt-2">
            <OverviewRow icon="mdi-lan" label="Networks"      :value="networksStore.items.length"      color="info" />
            <OverviewRow icon="mdi-earth" label="Virtual IPs"  :value="vipsStore.items.length"             color="accent" />
            <OverviewRow icon="mdi-scale-balance" label="Load Balancers" :value="lbsStore.items.length"   color="secondary" />
            <OverviewRow icon="mdi-ip-network" label="IP Prefixes" :value="prefixesStore.items.length"    color="primary" />
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- ── Recent Resources + Sidebar ───────────────────────────── -->
    <v-row class="mb-6">
      <v-col cols="12" md="8">
        <v-card :style="cardStyle" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-history</v-icon>
            Recent Resources
          </div>

          <!-- Attention chips inline -->
          <div v-if="pendingMachines || errorVolumes" class="d-flex flex-wrap ga-2 mb-3">
            <v-chip
              v-if="pendingMachines"
              color="warning"
              size="small"
              variant="tonal"
              prepend-icon="mdi-clock-outline"
            >
              {{ pendingMachines }} machine{{ pendingMachines > 1 ? 's' : '' }} pending
            </v-chip>
            <v-chip
              v-if="errorVolumes"
              color="error"
              size="small"
              variant="tonal"
              prepend-icon="mdi-alert-outline"
            >
              {{ errorVolumes }} volume{{ errorVolumes > 1 ? 's' : '' }} not available
            </v-chip>
          </div>

          <v-data-table
            :headers="recentHeaders"
            :items="recentResources"
            :items-per-page="8"
            density="compact"
            hide-default-footer
          >
            <template #item.type="{ item }">
              <v-chip :color="typeColor(item.type)" size="x-small" variant="tonal">{{ item.type }}</v-chip>
            </template>
            <template #item.status="{ item }">
              <StatusBadge :state="item.status" size="small" />
            </template>
            <template #item.name="{ item }">
              <router-link :to="item.link" class="text-decoration-none font-weight-medium">{{ item.name }}</router-link>
            </template>
            <template #item.created="{ item }">
              <span class="text-caption text-medium-emphasis">{{ relativeTime(item.created) }}</span>
            </template>
            <template #item.lastSeen="{ item }">
              <span class="text-caption text-medium-emphasis">{{ item.lastSeen ? relativeTime(item.lastSeen) : '—' }}</span>
            </template>
          </v-data-table>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <!-- Health Summary -->
        <v-card :style="cardStyle" class="pa-4 mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" :color="healthColor">mdi-heart-pulse</v-icon>
            Health Summary
          </div>

          <!-- All OK banner -->
          <div v-if="healthItems.length === 0" class="d-flex align-center ga-2">
            <v-icon color="success" size="20">mdi-check-circle-outline</v-icon>
            <span class="text-body-2 text-success">All systems OK</span>
          </div>

          <!-- Issues list -->
          <div v-else class="d-flex flex-column ga-2">
            <div
              v-for="item in healthItems"
              :key="item.label"
              class="d-flex align-center ga-2 text-body-2"
            >
              <v-icon :color="item.color" size="18">{{ item.icon }}</v-icon>
              <router-link :to="item.link" class="text-decoration-none" :class="`text-${item.color}`">
                {{ item.count }} {{ item.label }}
              </router-link>
            </div>
          </div>
        </v-card>

        <!-- Namespace info -->
        <v-card :style="cardStyle" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-3 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-folder-outline</v-icon>
            Active Namespace
          </div>
          <v-chip color="primary" variant="tonal" size="small" prepend-icon="mdi-folder">
            {{ nsStore.active }}
          </v-chip>
          <div class="text-caption text-medium-emphasis mt-2">
            {{ nsStore.namespaces.length }} namespace(s) available
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- ── Recent Events ──────────────────────────────────────────── -->
    <v-row>
      <v-col cols="12">
        <v-card :style="cardStyle" class="pa-4">
          <div class="d-flex align-center justify-space-between mb-4">
            <div class="text-subtitle-2 font-weight-bold d-flex align-center ga-2">
              <v-icon size="16" color="warning">mdi-bell-outline</v-icon>
              Recent Warning Events
            </div>
            <v-btn
              v-if="recentEvents.length > 0"
              variant="text"
              size="small"
              to="/operations"
              append-icon="mdi-chevron-right"
            >
              View all
            </v-btn>
          </div>

          <!-- Loading -->
          <div v-if="eventsLoading" class="d-flex justify-center py-4">
            <v-progress-circular indeterminate size="24" />
          </div>

          <!-- Empty state -->
          <div v-else-if="recentEvents.length === 0" class="d-flex align-center ga-2 text-medium-emphasis py-2">
            <v-icon size="18" color="success">mdi-check-circle-outline</v-icon>
            <span class="text-body-2">No warning events in this namespace</span>
          </div>

          <!-- Events table -->
          <div v-else>
            <div
              v-for="ev in recentEvents"
              :key="`${ev.name}-${ev.lastTime}`"
              class="d-flex align-start py-2 ga-3"
              style="border-bottom: 1px solid rgba(128,128,128,0.08);"
            >
              <v-icon
                :color="ev.type === 'Warning' ? 'warning' : 'info'"
                size="16"
                class="mt-1 flex-shrink-0"
              >
                {{ ev.type === 'Warning' ? 'mdi-alert' : 'mdi-information' }}
              </v-icon>
              <div class="flex-grow-1 min-width-0">
                <div class="text-body-2">
                  <span class="font-weight-medium">{{ ev.regarding }}</span>
                  <v-chip size="x-small" variant="tonal" class="ml-1" :color="kindColor(ev.regardsKind)">{{ ev.regardsKind }}</v-chip>
                </div>
                <div class="text-caption text-medium-emphasis text-truncate">{{ ev.reason }}: {{ ev.message }}</div>
              </div>
              <div class="text-caption text-medium-emphasis flex-shrink-0">{{ relativeTime(ev.lastTime) }}</div>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useCardStyle }  from '@/composables/useCardStyle'
import { relativeTime }  from '@/composables/useTime'
import PageHeader    from '@/components/common/PageHeader.vue'
import StatCard      from '@/components/common/StatCard.vue'
import StatusBadge   from '@/components/common/StatusBadge.vue'
import DonutChart    from '@/components/charts/DonutChart.vue'
import OverviewRow   from '@/components/common/OverviewRow.vue'
import NeedsAttention, { type AttentionItem } from '@/components/common/NeedsAttention.vue'
import { useMachinesStore }   from '@/stores/machines'
import { useVolumesStore }    from '@/stores/volumes'
import { useNetworksStore }   from '@/stores/networks'
import { useVirtualIPsStore } from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore }   from '@/stores/prefixes'
import { useNamespaceStore }  from '@/stores/namespace'
import { api, type ResourceEvent } from '@/api/client'

const { cardStyle } = useCardStyle()

const machinesStore  = useMachinesStore()
const volumesStore   = useVolumesStore()
const networksStore  = useNetworksStore()
const vipsStore      = useVirtualIPsStore()
const lbsStore       = useLoadBalancersStore()
const prefixesStore  = usePrefixesStore()
const nsStore        = useNamespaceStore()

// ── Events ─────────────────────────────────────────────────────────
const eventsLoading = ref(false)
const events = ref<ResourceEvent[]>([])

async function loadEvents() {
  eventsLoading.value = true
  try {
    events.value = await api.events.list(nsStore.active)
  } catch {
    events.value = []
  } finally {
    eventsLoading.value = false
  }
}

const recentEvents = computed(() =>
  events.value
    .filter(e => e.type === 'Warning')
    .sort((a, b) => (b.lastTime > a.lastTime ? 1 : -1))
    .slice(0, 6)
)

function kindColor(kind: string) {
  const m: Record<string, string> = { Machine: 'primary', Volume: 'info', Network: 'accent' }
  return m[kind] ?? 'secondary'
}

// ── Stats ──────────────────────────────────────────────────────────
const stats = computed(() => [
  { label: 'Machines',      value: machinesStore.items.length,  icon: 'mdi-server',             color: 'primary',   sub: runningMachines.value + ' running' },
  { label: 'Running',       value: runningMachines.value,       icon: 'mdi-check-circle-outline', color: 'success' },
  { label: 'Volumes',       value: volumesStore.items.length,   icon: 'mdi-database-outline',   color: 'info',      sub: formatStorage() },
  { label: 'Networks',      value: networksStore.items.length, icon: 'mdi-lan',              color: 'accent' },
  { label: 'Virtual IPs',   value: vipsStore.items.length,      icon: 'mdi-earth',              color: 'secondary' },
  { label: 'Load Balancers',value: lbsStore.items.length,       icon: 'mdi-scale-balance',      color: 'warning' },
])

const runningMachines = computed(() =>
  machinesStore.items.filter(m => m.state?.toLowerCase() === 'running').length
)

function formatStorage() {
  const total = volumesStore.items.reduce((acc, v) => acc + (v.sizeBytes || 0), 0)
  if (total === 0) return '—'
  return (total / (1024 ** 3)).toFixed(0) + ' GiB total'
}

// ── Charts ─────────────────────────────────────────────────────────
const machineSegments = computed(() => {
  const running  = machinesStore.items.filter(m => m.state?.toLowerCase() === 'running').length
  const pending  = machinesStore.items.filter(m => m.state?.toLowerCase() === 'pending').length
  const error    = machinesStore.items.filter(m => m.state?.toLowerCase() === 'error').length
  const off      = machinesStore.items.filter(m => !['running','pending','error'].includes(m.state?.toLowerCase() ?? '')).length
  return [
    { label: 'Running', value: running, color: '#34d399' },
    { label: 'Pending', value: pending, color: '#fbbf24' },
    { label: 'Error',   value: error,   color: '#f87171' },
    { label: 'Other',   value: off,     color: '#475569'  },
  ].filter(s => s.value > 0)
})

const volumeSegments = computed(() => {
  const available = volumesStore.items.filter(v => v.state?.toLowerCase() === 'available').length
  const bound     = volumesStore.items.filter(v => v.state?.toLowerCase() === 'bound').length
  const other     = volumesStore.items.filter(v => !['available','bound'].includes(v.state?.toLowerCase() ?? '')).length
  return [
    { label: 'Available', value: available, color: '#34d399' },
    { label: 'Bound',     value: bound,     color: '#818cf8' },
    { label: 'Other',     value: other,     color: '#475569' },
  ].filter(s => s.value > 0)
})

// ── Needs Attention ────────────────────────────────────────────────
const pendingMachines = computed(() =>
  machinesStore.items.filter(m => m.state?.toLowerCase() === 'pending').length
)
const errorVolumes = computed(() =>
  volumesStore.items.filter(v => v.state && !['available', 'bound'].includes(v.state.toLowerCase())).length
)
const warningEvents = computed(() =>
  events.value.filter(e => e.type === 'Warning').length
)

const attentionItems = computed(() => {
  const items: AttentionItem[] = []
  if (pendingMachines.value > 0) {
    items.push({
      key: 'pending-machines',
      label: `machine${pendingMachines.value > 1 ? 's' : ''} pending`,
      count: pendingMachines.value,
      icon: 'mdi-timer-sand',
      color: 'warning',
      link: '/machines',
      detail: 'Waiting for scheduling or hardware',
    })
  }
  if (errorVolumes.value > 0) {
    items.push({
      key: 'error-volumes',
      label: `volume${errorVolumes.value > 1 ? 's' : ''} with issues`,
      count: errorVolumes.value,
      icon: 'mdi-alert-circle-outline',
      color: 'error',
      link: '/volumes',
      detail: 'Not available or in error state',
    })
  }
  if (warningEvents.value > 0) {
    items.push({
      key: 'warning-events',
      label: `warning event${warningEvents.value > 1 ? 's' : ''}`,
      count: warningEvents.value,
      icon: 'mdi-bell-alert-outline',
      color: 'warning',
      link: '/operations',
      detail: 'Recent warnings from Kubernetes events',
    })
  }
  return items
})

// ── Health Summary ─────────────────────────────────────────────────
const healthItems = computed(() => {
  const items: { label: string; count: number; icon: string; color: string; link: string }[] = []
  if (pendingMachines.value > 0) {
    items.push({
      label: `Pending machine${pendingMachines.value > 1 ? 's' : ''}`,
      count: pendingMachines.value,
      icon: 'mdi-timer-sand',
      color: 'warning',
      link: '/machines',
    })
  }
  if (errorVolumes.value > 0) {
    items.push({
      label: `volume${errorVolumes.value > 1 ? 's' : ''} with issues`,
      count: errorVolumes.value,
      icon: 'mdi-alert-circle-outline',
      color: 'error',
      link: '/volumes',
    })
  }
  return items
})

const healthColor = computed(() => {
  if (healthItems.value.some(h => h.color === 'error')) return 'error'
  if (healthItems.value.some(h => h.color === 'warning')) return 'warning'
  return 'success'
})

// ── Recent resources table ─────────────────────────────────────────
const recentHeaders = [
  { title: 'Name',      key: 'name',     width: '35%' },
  { title: 'Type',      key: 'type',     width: '15%' },
  { title: 'Status',    key: 'status',   width: '20%' },
  { title: 'Created',   key: 'created',  width: '15%' },
  { title: 'Last seen', key: 'lastSeen', width: '15%' },
]

const recentResources = computed(() => {
  const rows: { name: string; type: string; status: string; created: string; lastSeen: string; link: string }[] = []
  machinesStore.items.slice(0,4).forEach(m => rows.push({
    name: m.name, type: 'Machine', status: m.state || m.power || '—', created: m.createdAt, lastSeen: m.createdAt, link: `/machines/${m.name}`
  }))
  volumesStore.items.slice(0,3).forEach(v => rows.push({
    name: v.name, type: 'Volume', status: v.state || '—', created: v.createdAt, lastSeen: v.createdAt, link: `/volumes/${v.name}`
  }))
  networksStore.items.slice(0,2).forEach((n: { name: string; createdAt?: string }) => rows.push({
    name: n.name, type: 'Network', status: 'Active', created: n.createdAt ?? '', lastSeen: n.createdAt ?? '', link: `/networks/${n.name}`
  }))
  return rows.sort((a,b) => (b.created > a.created ? 1 : -1)).slice(0, 8)
})

function typeColor(type: string) {
  const m: Record<string, string> = { Machine: 'primary', Volume: 'info', Network: 'accent' }
  return m[type] ?? 'secondary'
}

onMounted(async () => {
  await Promise.all([
    machinesStore.load(),
    volumesStore.load(),
    networksStore.load(),
    vipsStore.load(),
    lbsStore.load(),
    prefixesStore.load(),
    loadEvents(),
  ])
})
</script>
