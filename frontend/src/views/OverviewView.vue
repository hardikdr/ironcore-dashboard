<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Overview"
      subtitle="IronCore infrastructure at a glance"
      icon="mdi-view-dashboard-outline"
    />

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

    <!-- ── Recent Activity + Quick Actions ───────────────────────── -->
    <v-row>
      <v-col cols="12" md="8">
        <v-card :style="cardStyle" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-history</v-icon>
            Recent Resources
          </div>

          <!-- Attention items -->
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
              <StatusBadge :state="item.status" />
            </template>
            <template #item.name="{ item }">
              <router-link :to="item.link" class="text-decoration-none font-weight-medium">{{ item.name }}</router-link>
            </template>
            <template #item.created="{ item }">
              <span class="text-caption text-medium-emphasis">{{ relativeTime(item.created) }}</span>
            </template>
          </v-data-table>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <!-- Health Summary — actionable operator-relevant alerts -->
        <v-card :style="cardStyle" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-3 d-flex align-center ga-2">
            <v-icon size="16" :color="healthColor">mdi-shield-check-outline</v-icon>
            Fleet Health
          </div>
          <div v-if="healthAlerts.length === 0" class="d-flex align-center ga-2 text-caption" style="color:#34d399">
            <v-icon size="16" color="success">mdi-check-circle-outline</v-icon>
            All systems operational
          </div>
          <div v-else class="d-flex flex-column ga-2">
            <div
              v-for="alert in healthAlerts"
              :key="alert.label"
              class="d-flex align-center justify-space-between"
            >
              <div class="d-flex align-center ga-2">
                <v-icon size="14" :color="alert.color">{{ alert.icon }}</v-icon>
                <span class="text-caption">{{ alert.label }}</span>
              </div>
              <router-link :to="alert.link" class="text-caption text-decoration-none" style="color:inherit; opacity:0.7;">
                View →
              </router-link>
            </div>
          </div>
        </v-card>

        <!-- Quick create -->
        <v-card :style="cardStyle" class="pa-4 mt-4">
          <div class="text-subtitle-2 font-weight-bold mb-3 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-lightning-bolt</v-icon>
            Quick Actions
          </div>
          <div class="d-flex flex-column ga-2">
            <v-btn
              v-for="action in quickActions"
              :key="action.label"
              :to="action.to"
              :prepend-icon="action.icon"
              :color="action.color"
              variant="tonal"
              class="justify-start"
              size="small"
              rounded="lg"
            >
              {{ action.label }}
            </v-btn>
          </div>
        </v-card>

        <!-- Namespace info -->
        <v-card :style="cardStyle" class="pa-4 mt-4">
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
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useCardStyle }  from '@/composables/useCardStyle'
import { relativeTime }  from '@/composables/useTime'
import PageHeader   from '@/components/common/PageHeader.vue'
import StatCard     from '@/components/common/StatCard.vue'
import StatusBadge  from '@/components/common/StatusBadge.vue'
import DonutChart   from '@/components/charts/DonutChart.vue'
import OverviewRow  from '@/components/common/OverviewRow.vue'
import { useMachinesStore }   from '@/stores/machines'
import { useVolumesStore }    from '@/stores/volumes'
import { useNetworksStore }   from '@/stores/networks'
import { useVirtualIPsStore } from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore }   from '@/stores/prefixes'
import { useNamespaceStore }  from '@/stores/namespace'

const theme = useCardStyle()
const { cardStyle } = theme

const machinesStore  = useMachinesStore()
const volumesStore   = useVolumesStore()
const networksStore  = useNetworksStore()
const vipsStore      = useVirtualIPsStore()
const lbsStore       = useLoadBalancersStore()
const prefixesStore  = usePrefixesStore()
const nsStore        = useNamespaceStore()

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

// ── Recent resources table ─────────────────────────────────────────
const recentHeaders = [
  { title: 'Name',    key: 'name',    width: '40%' },
  { title: 'Type',    key: 'type',    width: '20%' },
  { title: 'Status',  key: 'status',  width: '25%' },
  { title: 'Created', key: 'created', width: '15%' },
]

const recentResources = computed(() => {
  const rows: { name: string; type: string; status: string; created: string; link: string }[] = []
  machinesStore.items.forEach(m => rows.push({
    name: m.name, type: 'Machine', status: m.state || m.power || '—', created: m.createdAt, link: `/machines/${m.name}`
  }))
  volumesStore.items.forEach(v => rows.push({
    name: v.name, type: 'Volume', status: v.state || '—', created: v.createdAt, link: `/volumes/${v.name}`
  }))
  networksStore.items.forEach((n: { name: string; createdAt?: string }) => rows.push({
    name: n.name, type: 'Network', status: 'Active', created: n.createdAt ?? '', link: `/networks/${n.name}`
  }))
  return rows.sort((a,b) => (b.created > a.created ? 1 : -1)).slice(0, 8)
})

// ── Attention items ────────────────────────────────────────────────
const pendingMachines = computed(() =>
  machinesStore.items.filter(m => m.state?.toLowerCase() === 'pending').length
)
const errorVolumes = computed(() =>
  volumesStore.items.filter(v => v.state?.toLowerCase() !== 'available' && v.state?.toLowerCase() !== 'bound').length
)

function typeColor(type: string) {
  const m: Record<string, string> = { Machine: 'primary', Volume: 'info', Network: 'accent' }
  return m[type] ?? 'secondary'
}

// ── Fleet health alerts ────────────────────────────────────────────
const healthAlerts = computed(() => {
  const alerts: { label: string; icon: string; color: string; link: string }[] = []
  const pendingMachines = machinesStore.items.filter(m => m.state?.toLowerCase() === 'pending').length
  const errorMachines   = machinesStore.items.filter(m => m.state?.toLowerCase() === 'error').length
  const errorVolumes    = volumesStore.items.filter(v => v.state?.toLowerCase() === 'error').length
  const pendingVolumes  = volumesStore.items.filter(v => v.state?.toLowerCase() === 'pending').length

  if (errorMachines > 0) alerts.push({
    label: `${errorMachines} machine${errorMachines > 1 ? 's' : ''} in error`,
    icon: 'mdi-alert-circle', color: 'error', link: '/machines'
  })
  if (pendingMachines > 0) alerts.push({
    label: `${pendingMachines} machine${pendingMachines > 1 ? 's' : ''} pending`,
    icon: 'mdi-clock-outline', color: 'warning', link: '/machines'
  })
  if (errorVolumes > 0) alerts.push({
    label: `${errorVolumes} volume${errorVolumes > 1 ? 's' : ''} in error`,
    icon: 'mdi-database-alert', color: 'error', link: '/volumes'
  })
  if (pendingVolumes > 0) alerts.push({
    label: `${pendingVolumes} volume${pendingVolumes > 1 ? 's' : ''} pending`,
    icon: 'mdi-clock-outline', color: 'warning', link: '/volumes'
  })
  return alerts
})

const healthColor = computed(() => {
  const hasError = healthAlerts.value.some(a => a.color === 'error')
  const hasWarning = healthAlerts.value.some(a => a.color === 'warning')
  return hasError ? 'error' : hasWarning ? 'warning' : 'success'
})

// ── Quick actions ──────────────────────────────────────────────────
const quickActions = [
  { label: 'Create Machine',    to: '/machines/new',      icon: 'mdi-plus', color: 'primary' },
  { label: 'Create Volume',     to: '/volumes/new',       icon: 'mdi-plus', color: 'info' },
  { label: 'Create Network',    to: '/networks/new',      icon: 'mdi-plus', color: 'accent' },
  { label: 'View Topology',     to: '/topology',          icon: 'mdi-sitemap-outline', color: 'secondary' },
  { label: 'Dedicated Hosts',   to: '/servers',           icon: 'mdi-server-security', color: 'warning' },
]

onMounted(async () => {
  await Promise.all([
    machinesStore.load(),
    volumesStore.load(),
    networksStore.load(),
    vipsStore.load(),
    lbsStore.load(),
    prefixesStore.load(),
  ])
})
</script>
