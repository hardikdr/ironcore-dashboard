<template>
  <v-app :theme="ui.themeName">
    <a class="skip-link" href="#main-content">Skip to main content</a>
    <v-app-bar elevation="0" height="56" class="app-toolbar">
      <template #prepend><v-app-bar-nav-icon aria-label="Toggle navigation" @click="toggleNavigation"/><router-link to="/overview" class="brand"><v-icon color="primary">mdi-hexagon-multiple</v-icon><span>IronCore</span></router-link></template>
      <v-btn v-if="!mobile" class="search-trigger ml-8" prepend-icon="mdi-magnify" variant="tonal" @click="searchOpen=true">Search resources and actions… <kbd>⌘ K</kbd></v-btn>
      <template #append><div class="toolbar-actions"><ConnectionStatus :state="ui.connectionState"/><v-select v-if="!mobile" :model-value="ns.active" :items="ns.namespaces" hide-details density="compact" aria-label="Select namespace" prepend-inner-icon="mdi-folder-outline" @update:model-value="ns.setActive($event)"/><v-btn icon="mdi-refresh" variant="text" :loading="refreshing" aria-label="Refresh all data" @click="refresh"/><v-btn :icon="ui.themeName==='dark'?'mdi-weather-sunny':'mdi-weather-night'" variant="text" :aria-label="ui.themeName==='dark'?'Switch to light mode':'Switch to dark mode'" @click="toggleTheme"/></div></template>
    </v-app-bar>
    <v-navigation-drawer v-model="drawer" :rail="ui.navigationRail && !mobile" :temporary="mobile" width="244" rail-width="60" class="app-navigation">
      <v-list nav density="compact"><NavItem to="/overview" icon="mdi-view-dashboard-outline" label="Overview" :rail="ui.navigationRail"/><NavSection label="Compute" :rail="ui.navigationRail"/><NavItem to="/machines" icon="mdi-server-outline" label="Machines" :rail="ui.navigationRail" :badge="machines.items.length||undefined"/><NavItem to="/topology" icon="mdi-sitemap-outline" label="Topology" :rail="ui.navigationRail"/><NavItem to="/servers" icon="mdi-server-security" label="Dedicated Hosts" :rail="ui.navigationRail"/><NavSection label="Storage" :rail="ui.navigationRail"/><NavItem to="/volumes" icon="mdi-database-outline" label="Volumes" :rail="ui.navigationRail"/><NavSection label="Networking" :rail="ui.navigationRail"/><NavItem to="/networks" icon="mdi-lan" label="Networks" :rail="ui.navigationRail"/><NavItem to="/virtualips" icon="mdi-earth" label="Virtual IPs" :rail="ui.navigationRail"/><NavItem to="/loadbalancers" icon="mdi-scale-balance" label="Load Balancers" :rail="ui.navigationRail"/><NavSection label="IPAM" :rail="ui.navigationRail"/><NavItem to="/prefixes" icon="mdi-ip-network-outline" label="IP Prefixes" :rail="ui.navigationRail"/></v-list>
      <template #append><v-list nav density="compact"><NavItem to="/settings" icon="mdi-cog-outline" label="Settings" :rail="ui.navigationRail"/></v-list></template>
    </v-navigation-drawer>
    <v-main id="main-content" tabindex="-1"><router-view v-slot="{Component}"><component :is="Component" :key="$route.path"/></router-view></v-main>
    <v-dialog v-model="searchOpen" max-width="640"><v-card><v-card-title class="d-flex align-center ga-2"><v-icon>mdi-magnify</v-icon>Search</v-card-title><v-card-text><v-text-field v-model="query" autofocus label="Search resources and actions" hide-details/><v-list class="mt-3"><v-list-item v-for="item in searchResults" :key="item.to" :to="item.to" :prepend-icon="item.icon" :title="item.label" @click="searchOpen=false"/><v-list-item v-if="!searchResults.length" title="No loaded resource or action matched" subtitle="Try a resource name or navigation destination."/></v-list></v-card-text></v-card></v-dialog>
    <ToastNotifications/>
  </v-app>
</template>
<script setup lang="ts">
import { computed,onMounted,onUnmounted,ref,watch } from 'vue'
import { useDisplay } from 'vuetify'
import { useNamespaceStore } from '@/stores/namespace'
import { useUIStore } from '@/stores/ui'
import { useMachinesStore } from '@/stores/machines'
import { useVolumesStore } from '@/stores/volumes'
import { useNetworksStore } from '@/stores/networks'
import { useVirtualIPsStore } from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore } from '@/stores/prefixes'
import ConnectionStatus from '@/components/status/ConnectionStatus.vue'
import NavItem from '@/components/nav/NavItem.vue'
import NavSection from '@/components/nav/NavSection.vue'
import ToastNotifications from '@/components/common/ToastNotifications.vue'
const {mobile}=useDisplay(),ui=useUIStore(),ns=useNamespaceStore(),machines=useMachinesStore(),volumes=useVolumesStore(),networks=useNetworksStore(),vips=useVirtualIPsStore(),lbs=useLoadBalancersStore(),prefixes=usePrefixesStore()
const drawer=ref(true),refreshing=ref(false),searchOpen=ref(false),query=ref(''),collections=[machines,volumes,networks,vips,lbs,prefixes]
const navigation=[['Overview','/overview','mdi-view-dashboard-outline'],['Machines','/machines','mdi-server-outline'],['Volumes','/volumes','mdi-database-outline'],['Networks','/networks','mdi-lan'],['Virtual IPs','/virtualips','mdi-earth'],['Load Balancers','/loadbalancers','mdi-scale-balance'],['IP Prefixes','/prefixes','mdi-ip-network-outline'],['Settings','/settings','mdi-cog-outline']]
const searchResults=computed(()=>{const q=query.value.trim().toLowerCase(),entries=[...navigation.map(([label,to,icon])=>({label,to,icon})),...machines.items.map(x=>({label:x.name,to:`/machines/${x.name}`,icon:'mdi-server-outline'})),...volumes.items.map(x=>({label:x.name,to:`/volumes/${x.name}`,icon:'mdi-database-outline'})),...networks.items.map(x=>({label:x.name,to:`/networks/${x.name}`,icon:'mdi-lan'}))];return(q?entries.filter(x=>x.label.toLowerCase().includes(q)):entries).slice(0,12)})
function toggleNavigation(){if(mobile.value)drawer.value=!drawer.value;else ui.setNavigationRail(!ui.navigationRail)}
function toggleTheme(){ui.setTheme(ui.themeName==='dark'?'light':'dark')}
async function refresh(silent=false){if(refreshing.value)return;refreshing.value=true;await Promise.allSettled([ui.checkHealth(),...collections.map(store=>store.load(silent))]);refreshing.value=false}
function keyboard(event:KeyboardEvent){if((event.metaKey||event.ctrlKey)&&event.key.toLowerCase()==='k'){event.preventDefault();searchOpen.value=true}}
let timer:ReturnType<typeof setInterval>|undefined
watch(()=>ns.active,()=>void refresh(false))
onMounted(async()=>{window.addEventListener('keydown',keyboard);await ns.load();await refresh(false);timer=setInterval(()=>{if(!document.hidden)void refresh(true)},30_000)})
onUnmounted(()=>{window.removeEventListener('keydown',keyboard);if(timer)clearInterval(timer)})
</script>
<style scoped>
.app-toolbar{border-bottom:1px solid var(--ic-border)!important;background:var(--ic-surface)!important}.brand{display:flex;align-items:center;gap:8px;color:var(--ic-text);font-weight:700;text-decoration:none;font-size:16px}.toolbar-actions{display:flex;align-items:center;gap:8px;padding-right:12px}.toolbar-actions :deep(.v-select){width:190px}.search-trigger{min-width:300px;justify-content:flex-start;color:var(--ic-text-muted)}kbd{margin-left:auto;font:11px ui-monospace;padding:2px 5px;border:1px solid var(--ic-border);border-radius:4px}.app-navigation{border-right:1px solid var(--ic-border)!important;background:var(--ic-surface)!important}:global(.v-main){background:var(--ic-canvas)}
</style>
