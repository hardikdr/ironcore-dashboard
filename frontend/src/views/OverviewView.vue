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
          </v-data-table>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card :style="cardStyle" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
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
import { useTheme } from 'vuetify'
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

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() =>
  isDark.value
    ? 'background: rgba(20,30,51,0.8); border: 1px solid rgba(129,140,248,0.1);'
    : 'background: rgba(255,255,255,0.9); border: 1px solid rgba(99,102,241,0.1);'
)

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
  machinesStore.items.slice(0,4).forEach(m => rows.push({
    name: m.name, type: 'Machine', status: m.state || m.power || '—', created: shortDate(m.createdAt), link: `/machines/${m.name}`
  }))
  volumesStore.items.slice(0,3).forEach(v => rows.push({
    name: v.name, type: 'Volume', status: v.state || '—', created: shortDate(v.createdAt), link: `/volumes/${v.name}`
  }))
  networksStore.items.slice(0,2).forEach((n: { name: string; createdAt?: string }) => rows.push({
    name: n.name, type: 'Network', status: 'Active', created: shortDate(n.createdAt), link: `/networks/${n.name}`
  }))
  return rows.sort((a,b) => b.created.localeCompare(a.created)).slice(0, 8)
})

function shortDate(iso?: string) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('en', { month: 'short', day: 'numeric' })
}

function typeColor(type: string) {
  const m: Record<string, string> = { Machine: 'primary', Volume: 'info', Network: 'accent' }
  return m[type] ?? 'secondary'
}

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
