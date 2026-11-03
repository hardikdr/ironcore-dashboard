<template><AppPage title="Landscape" subtitle="Cross-environment operational inventory from immutable read-only snapshots">
  <template #actions><SnapshotStamp :generated-at="landscape.meta?.generatedAt" :partial="landscape.meta?.partial"/></template>
  <v-alert v-if="landscape.meta?.partial" type="warning" variant="tonal" class="mb-5" title="Partial landscape">Some sources are stale, inaccessible, or incomplete. Available evidence is shown without filling gaps.</v-alert>
  <div class="metric-grid mb-6"><router-link v-for="item in metrics" :key="item.label" :to="item.to" class="metric"><v-icon :icon="item.icon" color="primary"/><span>{{ item.label }}</span><strong>{{ item.value }}</strong></router-link></div>
  <v-row><v-col cols="12" lg="8"><v-card variant="outlined"><v-card-title>Zone map</v-card-title><v-card-subtitle>Operational footprint; each tile opens the zone evidence workspace.</v-card-subtitle>
    <div v-if="landscape.zones.length" class="zone-grid pa-4"><router-link v-for="zone in landscape.zones" :key="zone.id" :to="`/sre/zones/${encodeURIComponent(zone.resource.provenance.name)}`" class="zone"><v-icon>mdi-map-marker-radius-outline</v-icon><strong>{{ zone.resource.provenance.name }}</strong><span>{{ zone.resource.health || 'unknown' }}</span></router-link></div>
    <div v-else class="empty pa-6">No zone evidence is present in this snapshot.</div>
  </v-card></v-col><v-col cols="12" lg="4"><v-card variant="outlined"><v-card-title>On-call signal</v-card-title><v-card-text><strong class="signal">{{ landscape.observations.length }}</strong><p>derived observations require review</p><v-btn to="/sre/triage" variant="tonal" block>Open triage</v-btn></v-card-text></v-card></v-col></v-row>
</AppPage></template>
<script setup lang="ts">
import { computed } from 'vue'
import AppPage from '@/components/app/AppPage.vue'
import SnapshotStamp from '@/components/sre/SnapshotStamp.vue'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
const landscape=useSreLandscapeStore()
const count=(...keys:string[])=>keys.reduce((total,key)=>total+(landscape.summary[key]??0),0)
const metrics=computed(()=>[
  {label:'Virtual machines',value:count('ironcore-vm','iron-machine','IronCoreVM','IronMachine'),to:'/sre/vms',icon:'mdi-server-outline'},
  {label:'Zones',value:count('zone','Zone'),to:'/sre/zones',icon:'mdi-map-marker-radius-outline'},
  {label:'Shoots',value:count('shoot','Shoot'),to:'/sre/shoots',icon:'mdi-sprout-outline'},
  {label:'Volumes',value:count('volume','Volume'),to:'/sre/storage',icon:'mdi-database-outline'},
  {label:'Load balancers',value:count('load-balancer','LoadBalancer'),to:'/sre/networking',icon:'mdi-scale-balance'},
  {label:'Physical nodes',value:count('physical-node','PhysicalNode'),to:'/sre/zones',icon:'mdi-server-security'},
])
</script>
<style scoped>.metric-grid{display:grid;grid-template-columns:repeat(6,minmax(0,1fr));gap:12px}.metric{display:grid;grid-template-columns:auto 1fr;gap:4px 9px;padding:16px;border:1px solid var(--ic-border);border-radius:var(--ic-radius-md);background:var(--ic-surface);color:var(--ic-text);text-decoration:none}.metric strong{grid-column:2;font-size:24px}.zone-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}.zone{display:grid;grid-template-columns:auto 1fr;gap:3px 9px;padding:14px;border:1px solid var(--ic-border);border-radius:8px;color:var(--ic-text);text-decoration:none}.zone span{grid-column:2;color:var(--ic-text-muted);font-size:12px}.empty{color:var(--ic-text-muted)}.signal{font-size:36px}@media(max-width:1100px){.metric-grid{grid-template-columns:repeat(3,1fr)}}@media(max-width:600px){.metric-grid,.zone-grid{grid-template-columns:1fr 1fr}.metric{padding:12px}}</style>
