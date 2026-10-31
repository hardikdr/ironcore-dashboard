<template>
  <v-app :theme="theme">
    <!-- ── Top Bar ─────────────────────────────────────────────────── -->
    <v-app-bar
      :elevation="0"
      height="56"
      :style="topBarStyle"
    >
      <template #prepend>
        <v-app-bar-nav-icon @click="rail = !rail" />
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
          <!-- Cluster status -->
          <v-chip size="small" color="success" variant="tonal" prepend-icon="mdi-circle-small">
            Connected
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
            @update:model-value="nsStore.setActive($event)"
          />

          <!-- Theme toggle -->
          <v-btn
            :icon="theme === 'dark' ? 'mdi-weather-sunny' : 'mdi-weather-night'"
            variant="text"
            size="small"
            @click="toggleTheme"
          />

          <!-- Notifications -->
          <v-btn icon="mdi-bell-outline" variant="text" size="small">
            <v-icon>mdi-bell-outline</v-icon>
            <v-badge color="error" content="3" floating />
          </v-btn>

          <!-- User avatar -->
          <v-avatar
            color="primary"
            size="34"
            style="cursor:pointer"
          >
            <span class="text-caption font-weight-bold">IC</span>
          </v-avatar>
        </div>
      </template>
    </v-app-bar>

    <!-- ── Navigation Drawer ──────────────────────────────────────── -->
    <v-navigation-drawer
      v-model="drawer"
      :rail="rail"
      :style="drawerStyle"
      width="240"
      rail-width="60"
    >
      <v-list density="compact" nav class="pt-2 pb-0">
        <!-- Overview -->
        <NavItem to="/" icon="mdi-view-dashboard-outline" label="Overview" :rail="rail" />

        <v-divider class="my-2" />
        <NavSection label="Compute" :rail="rail" />

        <NavItem to="/machines"   icon="mdi-server"              label="Machines"        :rail="rail" :badge="machineCount" />
        <NavItem to="/topology"   icon="mdi-sitemap-outline"     label="Topology"        :rail="rail" />
        <NavItem to="/servers"    icon="mdi-server-security"     label="Dedicated Hosts" :rail="rail" />

        <v-divider class="my-2" />
        <NavSection label="Storage" :rail="rail" />

        <NavItem to="/volumes"    icon="mdi-database-outline"    label="Volumes"         :rail="rail" />

        <v-divider class="my-2" />
        <NavSection label="Networking" :rail="rail" />

        <NavItem to="/networks"      icon="mdi-lan"              label="Networks"        :rail="rail" />
        <NavItem to="/virtualips"    icon="mdi-earth"            label="Virtual IPs"     :rail="rail" />
        <NavItem to="/loadbalancers" icon="mdi-scale-balance"    label="Load Balancers"  :rail="rail" />

        <v-divider class="my-2" />
        <NavSection label="IPAM" :rail="rail" />

        <NavItem to="/prefixes"   icon="mdi-ip-network-outline"  label="IP Prefixes"     :rail="rail" />
      </v-list>

      <template #append>
        <v-divider />
        <v-list density="compact" nav class="py-2">
          <NavItem to="/settings" icon="mdi-cog-outline"  label="Settings" :rail="rail" />
          <div v-if="!rail" class="px-3 pb-2 pt-1">
            <div class="text-caption" style="color: rgba(255,255,255,0.3)">IronCore v0.3 · k8s v1.30</div>
          </div>
        </v-list>
      </template>
    </v-navigation-drawer>

    <!-- ── Main Content ───────────────────────────────────────────── -->
    <v-main :style="mainStyle">
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTheme } from 'vuetify'
import { useNamespaceStore } from '@/stores/namespace'
import { useMachinesStore } from '@/stores/machines'
import NavItem from '@/components/nav/NavItem.vue'
import NavSection from '@/components/nav/NavSection.vue'

const vuetifyTheme = useTheme()
const theme = ref<'dark' | 'light'>('dark')
const rail   = ref(false)
const drawer = ref(true)

const nsStore = useNamespaceStore()
const machinesStore = useMachinesStore()

const machineCount = computed(() => machinesStore.items.length || undefined)

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  vuetifyTheme.global.name.value = theme.value
}

const isDark = computed(() => theme.value === 'dark')

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

onMounted(async () => {
  await nsStore.load()
  await machinesStore.load()
})
</script>
