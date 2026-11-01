<template>
  <v-app :theme="vuetifyTheme.global.name.value">
    <!-- ── Top Bar ─────────────────────────────────────────────────── -->
    <v-app-bar :elevation="0" height="56" :style="topBarStyle">
      <template #prepend>
        <v-app-bar-nav-icon aria-label="Toggle navigation drawer" @click="toggleDrawer" />
        <div class="d-flex align-center ga-2 ml-1">
          <v-avatar size="28" rounded="sm" color="primary">
            <v-icon size="18" color="white">mdi-hexagon-multiple</v-icon>
          </v-avatar>
          <span class="text-subtitle-1 font-weight-bold">IronCore</span>
          <v-chip v-if="!xs" size="x-small" color="primary" variant="tonal" class="ml-1">Dashboard</v-chip>
        </div>
      </template>

      <template #append>
        <div class="d-flex align-center ga-1" style="min-width:0; padding-right:12px;">
          <!-- Search: full button on ≥sm, icon-only on xs -->
          <v-btn
            v-if="!xs"
            variant="outlined"
            size="small"
            prepend-icon="mdi-magnify"
            class="search-trigger text-medium-emphasis"
            aria-label="Search resources (Ctrl+K)"
            title="Search resources (Ctrl+K)"
            style="min-width:140px; max-width:160px; justify-content:flex-start; font-weight:400;"
            @click="searchOpen = true"
          >
            Search…
            <template #append>
              <v-kbd class="ml-2 text-caption" style="opacity:0.5">⌘K</v-kbd>
            </template>
          </v-btn>
          <v-btn
            v-else
            icon="mdi-magnify"
            variant="text"
            size="small"
            aria-label="Search resources"
            @click="searchOpen = true"
          />

          <!-- DEMO MODE indicator (visible when running against fixture data) -->
          <v-chip
            v-if="isDemoMode"
            color="warning"
            variant="tonal"
            size="small"
            prepend-icon="mdi-test-tube"
            class="font-weight-bold"
            title="Running with fixture data — not connected to a real cluster"
          >
            <span v-if="!xs">DEMO</span>
          </v-chip>

          <!-- Cluster status: chip on ≥md, dot-icon on smaller -->
          <v-chip
            v-if="mdAndUp && !isDemoMode"
            :color="clusterOk ? 'success' : 'error'"
            variant="tonal"
            size="small"
            :prepend-icon="clusterOk ? 'mdi-circle-small' : 'mdi-alert-circle-outline'"
          >
            {{ clusterOk ? 'Connected' : 'Disconnected' }}
          </v-chip>
          <v-icon
            v-else-if="!isDemoMode"
            :color="clusterOk ? 'success' : 'error'"
            size="18"
            :title="clusterOk ? 'Cluster connected' : 'Cluster disconnected'"
          >
            {{ clusterOk ? 'mdi-circle' : 'mdi-alert-circle' }}
          </v-icon>

          <!-- Namespace switcher: select on ≥sm, compact icon+label on xs -->
          <v-select
            v-if="!xs"
            v-model="nsStore.active"
            :items="nsStore.namespaces"
            hide-details
            density="compact"
            variant="outlined"
            style="min-width:130px; max-width:180px; flex-shrink:1;"
            prepend-inner-icon="mdi-folder-outline"
            rounded="lg"
            aria-label="Select namespace"
            @update:model-value="nsStore.setActive($event)"
          />
          <v-menu v-else>
            <template #activator="{ props: menuProps }">
              <v-btn v-bind="menuProps" variant="text" size="small" :title="`Namespace: ${nsStore.active}`">
                <v-icon size="16" class="mr-1">mdi-folder-outline</v-icon>
                <span class="text-caption" style="max-width:60px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;">{{ nsStore.active }}</span>
                <v-icon size="14">mdi-chevron-down</v-icon>
              </v-btn>
            </template>
            <v-list density="compact" min-width="160">
              <v-list-item
                v-for="ns in nsStore.namespaces"
                :key="ns"
                :title="ns"
                :active="ns === nsStore.active"
                active-color="primary"
                @click="nsStore.setActive(ns)"
              />
            </v-list>
          </v-menu>

          <!-- Theme toggle -->
          <v-btn
            :icon="isDark ? 'mdi-weather-sunny' : 'mdi-weather-night'"
            variant="text"
            size="small"
            :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
            @click="toggleTheme"
          />

          <!-- Refresh indicator — hidden on xs to save space -->
          <v-btn
            v-if="!xs"
            icon="mdi-refresh"
            variant="text"
            size="small"
            :loading="anyLoading"
            aria-label="Refresh all data"
            @click="refreshAll"
          />

          <!-- User avatar — hidden on xs to preserve space for namespace+demo chips -->
          <v-avatar v-if="!xs" color="primary" size="30" style="cursor:pointer; flex-shrink:0;">
            <span class="text-caption font-weight-bold">IC</span>
          </v-avatar>
        </div>
      </template>
    </v-app-bar>

    <!-- ── Navigation Drawer ──────────────────────────────────────── -->
    <v-navigation-drawer
      v-model="drawer"
      :rail="rail && !mobile"
      :style="drawerStyle"
      width="240"
      rail-width="60"
      :temporary="mobile"
      @click:outside="mobile && (drawer = false)"
    >
      <v-list density="compact" nav class="pt-2 pb-0">
        <NavItem to="/overview" icon="mdi-view-dashboard-outline" label="Overview" :rail="rail && !mobile" @click="onNavClick" />

        <NavGroup label="Compute" :rail="rail && !mobile" storage-key="compute" :default-open="true">
          <NavItem to="/machines"     icon="mdi-server"              label="Machines"        :rail="rail && !mobile" :badge="machineCount" @click="onNavClick" />
          <NavItem to="/machinepools" icon="mdi-server-network"      label="Machine Pools"   :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/topology"     icon="mdi-sitemap-outline"     label="Topology"        :rail="rail && !mobile" @click="onNavClick" />
        </NavGroup>

        <NavGroup label="Storage" :rail="rail && !mobile" storage-key="storage" :default-open="true">
          <NavItem to="/volumes"         icon="mdi-database-outline"  label="Volumes"         :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/volumesnapshots" icon="mdi-camera"            label="Snapshots"       :rail="rail && !mobile" @click="onNavClick" />
        </NavGroup>

        <NavGroup label="Networking" :rail="rail && !mobile" storage-key="networking" :default-open="true">
          <NavItem to="/networks"          icon="mdi-lan"                label="Networks"          :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/virtualips"        icon="mdi-earth"              label="Virtual IPs"       :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/loadbalancers"     icon="mdi-scale-balance"      label="Load Balancers"    :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/natgateways"       icon="mdi-transit-transfer"   label="NAT Gateways"      :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/networkinterfaces" icon="mdi-ethernet"           label="Interfaces"        :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/networkpolicies"   icon="mdi-shield-lock-outline" label="Network Policies" :rail="rail && !mobile" @click="onNavClick" />
        </NavGroup>

        <NavGroup label="Object Storage" :rail="rail && !mobile" storage-key="objstorage" :default-open="false">
          <NavItem to="/buckets"       icon="mdi-bucket-outline"     label="Buckets"         :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/bucketclasses" icon="mdi-bucket"             label="Bucket Classes"  :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/bucketpools"   icon="mdi-bucket-plus"        label="Bucket Pools"    :rail="rail && !mobile" @click="onNavClick" />
        </NavGroup>

        <NavGroup label="IPAM" :rail="rail && !mobile" storage-key="ipam" :default-open="true">
          <NavItem to="/prefixes"   icon="mdi-ip-network-outline"    label="IP Prefixes"     :rail="rail && !mobile" @click="onNavClick" />
        </NavGroup>

        <NavGroup label="Operations" :rail="rail && !mobile" storage-key="operations" :default-open="true">
          <NavItem to="/operations"    icon="mdi-clipboard-list-outline" label="Events"          :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/servers"        icon="mdi-server-security"   label="Servers"         :rail="rail && !mobile" @click="onNavClick" />
          <NavItem to="/resourcequotas" icon="mdi-gauge"             label="Resource Quotas" :rail="rail && !mobile" @click="onNavClick" />
        </NavGroup>
      </v-list>

      <template #append>
        <v-divider />
        <v-list density="compact" nav class="py-2">
          <NavItem to="/settings" icon="mdi-cog-outline" label="Settings" :rail="rail && !mobile" @click="onNavClick" />
          <div v-if="!(rail && !mobile)" class="px-3 pb-2 pt-1">
            <div class="text-caption" style="opacity:0.35">IronCore · {{ nsStore.active }}</div>
          </div>
        </v-list>
      </template>
    </v-navigation-drawer>

    <!-- ── Main Content ───────────────────────────────────────────── -->
    <v-main :style="mainStyle" style="min-width:0; overflow-x:hidden;">
      <router-view v-slot="{ Component }">
        <v-fade-transition mode="out-in" :duration="120">
          <component :is="Component" :key="$route.path" />
        </v-fade-transition>
      </router-view>
    </v-main>

    <!-- ── Global Toasts ─────────────────────────────────────────── -->
    <ToastNotifications />

    <!-- ── Global Search ─────────────────────────────────────────── -->
    <GlobalSearch v-model="searchOpen" />
  </v-app>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useTheme, useDisplay } from 'vuetify'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import { useNamespaceStore }     from '@/stores/namespace'
import { useMachinesStore }      from '@/stores/machines'
import { useVolumesStore }       from '@/stores/volumes'
import { useNetworksStore }      from '@/stores/networks'
import { useVirtualIPsStore }    from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore }      from '@/stores/prefixes'
import NavItem            from '@/components/nav/NavItem.vue'
import NavGroup           from '@/components/nav/NavGroup.vue'
import ToastNotifications from '@/components/common/ToastNotifications.vue'
import GlobalSearch       from '@/components/common/GlobalSearch.vue'

const vuetifyTheme = useTheme()
const { mobile, xs, mdAndUp } = useDisplay()
const route = useRoute()

// Drawer is closed by default on mobile; open on desktop
const drawer = ref(!mobile.value)
const rail   = ref(false)
const searchOpen = ref(false)

// Keep drawer closed on mobile when breakpoint changes to mobile
watch(mobile, isMobile => { if (isMobile) drawer.value = false })

function toggleDrawer() {
  if (mobile.value) drawer.value = !drawer.value
  else rail.value = !rail.value
}

// Close drawer on mobile after navigating
function onNavClick() {
  if (mobile.value) drawer.value = false
}

// Also close on route change (handles programmatic navigation)
watch(() => route.path, () => { if (mobile.value) drawer.value = false })

const nsStore       = useNamespaceStore()
const machinesStore = useMachinesStore()
const volumesStore  = useVolumesStore()
const networksStore = useNetworksStore()
const vipsStore     = useVirtualIPsStore()
const lbsStore      = useLoadBalancersStore()
const prefixesStore = usePrefixesStore()

const isDark = computed(() => vuetifyTheme.global.current.value.dark)

function toggleTheme() {
  vuetifyTheme.global.name.value = isDark.value ? 'light' : 'dark'
}

const clusterOk    = ref(true)
const isDemoMode   = ref(false)
const machineCount = computed(() => machinesStore.items.length || undefined)
const anyLoading   = computed(() =>
  machinesStore.loading || volumesStore.loading || networksStore.loading
)

const topBarStyle = computed(() => isDark.value
  ? 'background: rgba(13,21,38,0.95); border-bottom: 1px solid rgba(129,140,248,0.1); backdrop-filter: blur(12px);'
  : 'background: rgba(255,255,255,0.95); border-bottom: 1px solid rgba(99,102,241,0.1); backdrop-filter: blur(12px);'
)
const drawerStyle = computed(() => isDark.value
  ? 'background: #080e1c; border-right: 1px solid rgba(129,140,248,0.08);'
  : 'background: #f8fafc; border-right: 1px solid rgba(99,102,241,0.08);'
)
const mainStyle = computed(() => isDark.value
  ? 'background: #070b14;'
  : 'background: #f1f5f9;'
)

async function refreshAll(silent = true) {
  await Promise.allSettled([
    machinesStore.load(silent),
    volumesStore.load(silent),
    networksStore.load(silent),
    vipsStore.load(silent),
    lbsStore.load(silent),
    prefixesStore.load(silent),
  ])
}

let refreshTimer: ReturnType<typeof setInterval> | null = null

watch(() => nsStore.active, () => refreshAll(false))

onMounted(async () => {
  await nsStore.load()
  await refreshAll(false)
  try {
    await fetch('/healthz')
    clusterOk.value = true
  } catch {
    clusterOk.value = false
  }
  try {
    const mode = await api.mode.get()
    isDemoMode.value = mode.demo
  } catch {
    isDemoMode.value = false
  }
  refreshTimer = setInterval(() => refreshAll(true), 30_000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>

<style>
/* Prevent any child from causing horizontal overflow */
.v-main { overflow-x: hidden !important; }
/* Ensure tables scroll horizontally rather than stretch the viewport */
.v-data-table { overflow-x: auto; }
</style>
