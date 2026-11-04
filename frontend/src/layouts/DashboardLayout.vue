<template>
  <v-app :theme="isSre ? 'dark' : ui.themeName" :class="{ 'sre-console': isSre }">
    <a class="skip-link" href="#main-content">Skip to main content</a>
    <v-app-bar elevation="0" :height="isSre ? 64 : 56" class="app-toolbar">
      <template #prepend>
        <v-app-bar-nav-icon v-if="mobile || !isSre" aria-label="Toggle navigation" @click="toggleNavigation"/>
        <router-link :to="isSre ? '/sre/landscape' : '/overview'" class="brand"><span v-if="isSre" class="brand-mark" aria-hidden="true"></span><v-icon v-else color="primary">mdi-hexagon-multiple</v-icon><span>IronCore</span></router-link>
        <span v-if="isSre" class="sre-product">{{ route.path === '/sre/triage' ? 'On-call' : 'SRE Console' }}</span>
        <ConsoleModeSwitch v-else class="ml-4"/>
      </template>
      <v-btn v-if="!mobile && !isSre" class="search-trigger ml-8" prepend-icon="mdi-magnify" variant="tonal" @click="searchOpen=true">Search resources and actions… <kbd>⌘ K</kbd></v-btn>
      <template #append><div class="toolbar-actions">
        <div v-if="isSre" class="sre-live"><span class="sre-live__dot"></span>{{ ui.connectionState === 'connected' ? 'Live' : 'Offline' }} · {{ currentTime }}</div>
        <ConnectionStatus v-else :state="ui.connectionState"/>
        <template v-if="isSre">
          <v-btn v-if="!mobile" class="sre-search" prepend-icon="mdi-magnify" variant="outlined" @click="searchOpen=true">{{ route.path === '/sre/triage' ? 'Search incidents or resources' : 'Find any resource' }}</v-btn>
          <div class="sre-avatar" aria-label="Operator profile">OD</div>
        </template>
        <v-select v-else-if="!mobile" :model-value="ns.active" :items="ns.namespaces" hide-details density="compact" aria-label="Select namespace" prepend-inner-icon="mdi-folder-outline" @update:model-value="ns.setActive($event)"/>
        <v-btn v-if="!isSre" icon="mdi-refresh" variant="text" :loading="refreshing" aria-label="Refresh all data" @click="refresh(false)"/>
        <v-btn v-if="!isSre" :icon="ui.themeName==='dark'?'mdi-weather-sunny':'mdi-weather-night'" variant="text" :aria-label="ui.themeName==='dark'?'Switch to light mode':'Switch to dark mode'" @click="toggleTheme"/>
      </div></template>
    </v-app-bar>
    <v-navigation-drawer v-model="drawer" :rail="!isSre && ui.navigationRail && !mobile" :temporary="mobile" :width="isSre ? 252 : 244" rail-width="60" class="app-navigation">
      <SreNavigation v-if="isSre" :rail="ui.navigationRail"/>
      <v-list v-else nav density="compact">
        <NavItem to="/overview" icon="mdi-view-dashboard-outline" label="Overview" :rail="ui.navigationRail"/><NavSection label="Compute" :rail="ui.navigationRail"/>
        <NavItem to="/machines" icon="mdi-server-outline" label="Machines" :rail="ui.navigationRail" :badge="machines.items.length||undefined"/><NavItem to="/topology" icon="mdi-sitemap-outline" label="Topology" :rail="ui.navigationRail"/><NavItem to="/servers" icon="mdi-server-security" label="Dedicated Hosts" :rail="ui.navigationRail"/>
        <NavSection label="Storage" :rail="ui.navigationRail"/><NavItem to="/volumes" icon="mdi-database-outline" label="Volumes" :rail="ui.navigationRail"/>
        <NavSection label="Networking" :rail="ui.navigationRail"/><NavItem to="/networks" icon="mdi-lan" label="Networks" :rail="ui.navigationRail"/><NavItem to="/virtualips" icon="mdi-earth" label="Virtual IPs" :rail="ui.navigationRail"/><NavItem to="/loadbalancers" icon="mdi-scale-balance" label="Load Balancers" :rail="ui.navigationRail"/>
        <NavSection label="IPAM" :rail="ui.navigationRail"/><NavItem to="/prefixes" icon="mdi-ip-network-outline" label="IP Prefixes" :rail="ui.navigationRail"/>
      </v-list>
      <template v-if="!isSre" #append><v-list nav density="compact"><NavItem to="/settings" icon="mdi-cog-outline" label="Settings" :rail="ui.navigationRail"/></v-list></template>
    </v-navigation-drawer>
    <v-main id="main-content" tabindex="-1"><router-view v-slot="{Component}"><component :is="Component" :key="$route.path"/></router-view></v-main>
    <v-dialog v-model="searchOpen" max-width="640"><v-card>
      <v-card-title class="d-flex align-center ga-2"><v-icon>mdi-magnify</v-icon>{{ isSre ? 'Search operational resources' : 'Search' }}</v-card-title>
      <v-card-text><v-text-field v-model="query" autofocus :label="isSre ? 'Name, UID, provider ID, node or shoot' : 'Search resources and actions'" hide-details :loading="isSre && sreSearch.loading"/><v-list class="mt-3">
        <v-list-item v-for="item in searchResults" :key="item.key" :to="item.to" :prepend-icon="item.icon" :title="item.label" :subtitle="item.subtitle" @click="searchOpen=false"/>
        <v-list-item v-if="!searchResults.length" :title="isSre && query.trim().length < 2 ? 'Enter at least two characters' : 'No resource or action matched'" subtitle="Try a name, UID, provider ID, node, shoot, or navigation destination."/>
      </v-list></v-card-text>
    </v-card></v-dialog>
    <ToastNotifications/>
  </v-app>
</template>

<script setup lang="ts">
import { computed,onMounted,onUnmounted,ref,watch } from 'vue'
import { useRoute } from 'vue-router'
import { useDisplay } from 'vuetify'
import { useNamespaceStore } from '@/stores/namespace'
import { useUIStore } from '@/stores/ui'
import { useMachinesStore } from '@/stores/machines'
import { useVolumesStore } from '@/stores/volumes'
import { useNetworksStore } from '@/stores/networks'
import { useVirtualIPsStore } from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore } from '@/stores/prefixes'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
import { useSreSearchStore } from '@/stores/sre/search'
import { useSreStatusStore } from '@/stores/sre/status'
import ConnectionStatus from '@/components/status/ConnectionStatus.vue'
import ConsoleModeSwitch from '@/components/app/ConsoleModeSwitch.vue'
import NavItem from '@/components/nav/NavItem.vue'
import NavSection from '@/components/nav/NavSection.vue'
import SreNavigation from '@/components/sre/SreNavigation.vue'
import ToastNotifications from '@/components/common/ToastNotifications.vue'

const {mobile}=useDisplay(),route=useRoute(),ui=useUIStore(),ns=useNamespaceStore(),machines=useMachinesStore(),volumes=useVolumesStore(),networks=useNetworksStore(),vips=useVirtualIPsStore(),lbs=useLoadBalancersStore(),prefixes=usePrefixesStore(),sreStatus=useSreStatusStore(),sreLandscape=useSreLandscapeStore(),sreSearch=useSreSearchStore()
const drawer=ref(true),refreshing=ref(false),searchOpen=ref(false),query=ref(''),isSre=computed(()=>route.path.startsWith('/sre/')),collections=[machines,volumes,networks,vips,lbs,prefixes]
const currentTime=ref('')
function updateClock(){currentTime.value=new Intl.DateTimeFormat('en-GB',{hour:'2-digit',minute:'2-digit',timeZoneName:'short'}).format(new Date())}
const navigation=[['Overview','/overview','mdi-view-dashboard-outline'],['Machines','/machines','mdi-server-outline'],['Volumes','/volumes','mdi-database-outline'],['Networks','/networks','mdi-lan'],['Virtual IPs','/virtualips','mdi-earth'],['Load Balancers','/loadbalancers','mdi-scale-balance'],['IP Prefixes','/prefixes','mdi-ip-network-outline'],['Settings','/settings','mdi-cog-outline']]
function sreRoute(kind:string,id:string){const value=kind.toLowerCase();if(value.includes('vm'))return`/sre/vms/${id}`;if(value==='shoot')return`/sre/shoots/${id}`;if(value.includes('volume'))return`/sre/volumes/${id}`;if(value.includes('loadbalancer'))return`/sre/loadbalancers/${id}`;if(value.includes('node')||value.includes('server'))return`/sre/nodes/${id}`;return`/sre/dependencies/${id}`}
const searchResults=computed(()=>{if(isSre.value)return sreSearch.results.slice(0,25).map(item=>({key:item.id,label:item.resource.provenance.name,to:sreRoute(item.resource.provenance.kind,item.id),icon:'mdi-magnify',subtitle:`${item.resource.provenance.kind} · matched ${item.matchedField} (${item.matchKind})`}));const q=query.value.trim().toLowerCase(),entries=[...navigation.map(([label,to,icon])=>({key:to,label,to,icon,subtitle:''})),...machines.items.map(x=>({key:`machine-${x.name}`,label:x.name,to:`/machines/${x.name}`,icon:'mdi-server-outline',subtitle:'Machine'})),...volumes.items.map(x=>({key:`volume-${x.name}`,label:x.name,to:`/volumes/${x.name}`,icon:'mdi-database-outline',subtitle:'Volume'})),...networks.items.map(x=>({key:`network-${x.name}`,label:x.name,to:`/networks/${x.name}`,icon:'mdi-lan',subtitle:'Network'}))];return(q?entries.filter(x=>x.label.toLowerCase().includes(q)):entries).slice(0,12)})
function toggleNavigation(){if(mobile.value)drawer.value=!drawer.value;else ui.setNavigationRail(!ui.navigationRail)}
function toggleTheme(){ui.setTheme(ui.themeName==='dark'?'light':'dark')}
async function refresh(silent=false){if(refreshing.value)return;refreshing.value=true;if(isSre.value)await Promise.allSettled([ui.checkHealth(),sreStatus.load(),sreLandscape.load(sreStatus.environment)]);else await Promise.allSettled([ui.checkHealth(),...collections.map(store=>store.load(silent))]);refreshing.value=false}
function keyboard(event:KeyboardEvent){if((event.metaKey||event.ctrlKey)&&event.key.toLowerCase()==='k'){event.preventDefault();searchOpen.value=true}}
let refreshTimer:ReturnType<typeof setInterval>|undefined,searchTimer:ReturnType<typeof setTimeout>|undefined,searchController:AbortController|undefined
watch(query,value=>{if(!isSre.value)return;if(searchTimer)clearTimeout(searchTimer);searchController?.abort();if(value.trim().length<2){sreSearch.clear();return}searchTimer=setTimeout(()=>{searchController=new AbortController();void sreSearch.run(value,searchController.signal)},180)})
watch(()=>sreStatus.environment,()=>{if(isSre.value)void sreLandscape.load(sreStatus.environment)})
watch(isSre,async value=>{query.value='';if(value)await refresh(false);else{await ns.load();await refresh(false)}})
onMounted(async()=>{window.addEventListener('keydown',keyboard);updateClock();if(!isSre.value)await ns.load();await refresh(false);refreshTimer=setInterval(()=>{updateClock();if(!document.hidden)void refresh(true)},30_000)})
onUnmounted(()=>{window.removeEventListener('keydown',keyboard);if(refreshTimer)clearInterval(refreshTimer);if(searchTimer)clearTimeout(searchTimer);searchController?.abort()})
</script>
<style scoped>
.app-toolbar{border-bottom:1px solid var(--ic-border)!important;background:var(--ic-surface)!important}.brand{display:flex;align-items:center;gap:12px;color:var(--ic-text);font-weight:700;text-decoration:none;font-size:18px}.toolbar-actions{display:flex;align-items:center;gap:14px;padding-right:18px}.toolbar-actions :deep(.v-select){width:190px}.search-trigger{min-width:300px;justify-content:flex-start;color:var(--ic-text-muted)}kbd{margin-left:auto;font:11px ui-monospace;padding:2px 5px;border:1px solid var(--ic-border);border-radius:4px}.app-navigation{border-right:1px solid var(--ic-border)!important;background:var(--ic-surface)!important}:global(.v-main){background:var(--ic-canvas)}
.brand-mark{width:16px;height:16px;background:#9cabff;transform:rotate(45deg);display:inline-block}.sre-product{margin-left:20px;padding-left:20px;border-left:1px solid var(--ic-border);font-size:17px;color:var(--ic-text-muted)}.sre-live{font-size:15px;color:#54dca0;white-space:nowrap}.sre-live__dot{display:inline-block;width:11px;height:11px;border-radius:50%;background:#54dca0;margin-right:8px;box-shadow:0 0 12px rgba(84,220,160,.2)}.sre-search{height:42px!important;border-color:var(--ic-border)!important;color:var(--ic-text)!important;text-transform:none!important;font-size:15px!important}.sre-avatar{display:grid;place-items:center;width:42px;height:42px;border-radius:50%;background:#9cabff;color:#10131b;font-weight:700}
:global(.sre-console .v-application__wrap){background:#0f1117}:global(.sre-console .v-main){background:#0f1117}:global(.sre-console .app-navigation){padding-top:18px}:global(.sre-console .v-navigation-drawer__content){overflow-y:auto}
@media(max-width:960px){.console-mode-switch{margin-left:4px!important}.brand>span:not(.brand-mark){display:none}.sre-product{margin-left:10px;padding-left:10px}.sre-live{display:none}}
</style>
<style scoped>
.brand-mark{width:13px;height:13px}.sre-product{margin-left:16px;padding-left:16px;font-size:14px}.sre-live{font-size:13px}.sre-live__dot{width:8px;height:8px}.sre-search{height:36px!important;font-size:13px!important}.sre-avatar{width:34px;height:34px;font-size:12px}.toolbar-actions{gap:10px;padding-right:14px}
</style>
