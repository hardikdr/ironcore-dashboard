<template>
  <v-app :theme="vuetifyTheme.global.name.value">
    <!-- ── Top Bar ─────────────────────────────────────────────────── -->
    <v-app-bar :elevation="0" height="56" :style="topBarStyle">
      <template #prepend>
        <v-app-bar-nav-icon aria-label="Toggle navigation drawer" @click="mobile ? (drawer = !drawer) : (rail = !rail)" />
        <div class="d-flex align-center ga-2 ml-1">
          <v-avatar size="28" rounded="sm" color="primary">
            <v-icon size="18" color="white">mdi-hexagon-multiple</v-icon>
          </v-avatar>
          <span class="text-subtitle-1 font-weight-bold">IronCore</span>
          <v-chip size="x-small" color="primary" variant="tonal" class="ml-1">Dashboard</v-chip>
        </div>
      </template>

      <template #append>
        <div class="d-flex align-center ga-2 pr-3">
          <!-- Global search trigger -->
          <v-btn
            variant="outlined"
            size="small"
            prepend-icon="mdi-magnify"
            class="search-trigger text-medium-emphasis"
            aria-label="Search resources (Ctrl+K)"
            title="Search resources (Ctrl+K)"
            style="min-width:160px; justify-content:flex-start; font-weight:400;"
            @click="searchOpen = true"
          >
            Search…
            <template #append>
              <v-kbd class="ml-2 text-caption" style="opacity:0.5">⌘K</v-kbd>
            </template>
          </v-btn>

          <!-- Cluster status -->
          <v-chip
            :color="clusterOk ? 'success' : 'error'"
            variant="tonal"
            size="small"
            :prepend-icon="clusterOk ? 'mdi-circle-small' : 'mdi-alert-circle-outline'"
          >
            {{ clusterOk ? 'Connected' : 'Disconnected' }}
          </v-chip>

          <!-- Namespace switcher -->
          <v-select
            v-model="nsStore.active"
            :items="nsStore.namespaces"
            hide-details
            density="compact"
            variant="outlined"
            style="min-width:160px; max-width:200px"
            prepend-inner-icon="mdi-folder-outline"
            rounded="lg"
            aria-label="Select namespace"
            @update:model-value="nsStore.setActive($event)"
          />

          <!-- Theme toggle -->
          <v-btn
            :icon="isDark ? 'mdi-weather-sunny' : 'mdi-weather-night'"
            variant="text"
            size="small"
            :aria-label="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
            :title="isDark ? 'Switch to light mode' : 'Switch to dark mode'"
            @click="toggleTheme"
          />

          <!-- Refresh indicator -->
          <v-btn
            icon="mdi-refresh"
            variant="text"
            size="small"
            :loading="anyLoading"
            aria-label="Refresh all data"
            title="Refresh all data"
            @click="refreshAll"
          />

          <!-- User avatar -->
          <v-avatar color="primary" size="34" style="cursor:pointer">
            <span class="text-caption font-weight-bold">IC</span>
          </v-avatar>
        </div>
      </template>
    </v-app-bar>

    <!-- ── Navigation Drawer ──────────────────────────────────────── -->
    <v-navigation-drawer v-model="drawer" :rail="rail && !mobile" :style="drawerStyle" width="240" rail-width="60" :temporary="mobile">
      <v-list density="compact" nav class="pt-2 pb-0">
        <NavItem to="/overview" icon="mdi-view-dashboard-outline" label="Overview" :rail="rail" />

        <NavGroup label="Compute" :rail="rail" storage-key="compute" :default-open="true">
          <NavItem to="/machines"   icon="mdi-server"              label="Machines"        :rail="rail" :badge="machineCount" />
          <NavItem to="/topology"   icon="mdi-sitemap-outline"     label="Topology"        :rail="rail" />
        </NavGroup>

        <NavGroup label="Storage" :rail="rail" storage-key="storage" :default-open="true">
          <NavItem to="/volumes"    icon="mdi-database-outline"    label="Volumes"         :rail="rail" />
        </NavGroup>

        <NavGroup label="Networking" :rail="rail" storage-key="networking" :default-open="true">
          <NavItem to="/networks"      icon="mdi-lan"              label="Networks"        :rail="rail" />
          <NavItem to="/virtualips"    icon="mdi-earth"            label="Virtual IPs"     :rail="rail" />
          <NavItem to="/loadbalancers" icon="mdi-scale-balance"    label="Load Balancers"  :rail="rail" />
        </NavGroup>

        <NavGroup label="IPAM" :rail="rail" storage-key="ipam" :default-open="true">
          <NavItem to="/prefixes"   icon="mdi-ip-network-outline"  label="IP Prefixes"     :rail="rail" />
        </NavGroup>

        <NavGroup label="Operations" :rail="rail" storage-key="operations" :default-open="true">
          <NavItem to="/servers"    icon="mdi-server-security"     label="Servers"         :rail="rail" />
        </NavGroup>
      </v-list>

      <template #append>
        <v-divider />
        <v-list density="compact" nav class="py-2">
          <NavItem to="/settings" icon="mdi-cog-outline" label="Settings" :rail="rail" />
          <div v-if="!rail" class="px-3 pb-2 pt-1">
            <div class="text-caption" style="opacity:0.35">IronCore · {{ nsStore.active }}</div>
          </div>
        </v-list>
      </template>
    </v-navigation-drawer>

    <!-- ── Main Content ───────────────────────────────────────────── -->
    <v-main :style="mainStyle">
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
const { mobile } = useDisplay()
const rail   = ref(false)
const drawer = ref(true)
const searchOpen = ref(false)

const nsStore       = useNamespaceStore()
const machinesStore = useMachinesStore()
const volumesStore  = useVolumesStore()
const networksStore = useNetworksStore()
const vipsStore     = useVirtualIPsStore()
const lbsStore      = useLoadBalancersStore()
const prefixesStore = usePrefixesStore()

// Theme derived purely from Vuetify — no duplicate local ref
const isDark = computed(() => vuetifyTheme.global.current.value.dark)

function toggleTheme() {
  vuetifyTheme.global.name.value = isDark.value ? 'light' : 'dark'
}

const clusterOk    = ref(true)
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

// Auto-refresh every 30 seconds
let refreshTimer: ReturnType<typeof setInterval> | null = null

// Re-load everything when namespace changes
watch(() => nsStore.active, () => refreshAll(false))

onMounted(async () => {
  await nsStore.load()
  await refreshAll(false)
  // Ping for cluster health
  try {
    await fetch('/healthz')
    clusterOk.value = true
  } catch {
    clusterOk.value = false
  }
  refreshTimer = setInterval(() => refreshAll(true), 30_000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>
