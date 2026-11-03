<template><AppPage title="Sources" subtitle="Collection access, freshness, and object coverage by configured read-only source">
  <template #actions><SnapshotStamp :generated-at="status.meta?.generatedAt" :partial="status.meta?.partial"/></template>
  <v-alert v-if="status.error" type="error" variant="tonal" class="mb-5">{{ status.error }}</v-alert>
  <div class="source-grid"><v-card v-for="source in status.sources" :key="source.id" variant="outlined"><v-card-title class="source-title"><span>{{ source.id }}</span><SreSourceState :access="source.access" :freshness="source.freshness"/></v-card-title><v-card-text><dl><div><dt>Collected</dt><dd>{{ time(source.collectedAt) }}</dd></div><div><dt>Duration</dt><dd>{{ duration(source.duration) }}</dd></div><div><dt>Objects</dt><dd>{{ total(source.objectCount) }}</dd></div><div><dt>Error class</dt><dd>{{ source.errorClass || 'none' }}</dd></div></dl></v-card-text></v-card>
    <v-card v-if="!status.sources.length" variant="outlined"><v-card-text>No SRE sources are configured or no snapshot has been collected yet.</v-card-text></v-card></div>
</AppPage></template>
<script setup lang="ts">
import AppPage from '@/components/app/AppPage.vue'
import SnapshotStamp from '@/components/sre/SnapshotStamp.vue'
import SreSourceState from '@/components/sre/SreSourceState.vue'
import { useSreStatusStore } from '@/stores/sre/status'
const status=useSreStatusStore()
const time=(value?:string)=>value?new Intl.DateTimeFormat(undefined,{dateStyle:'medium',timeStyle:'medium'}).format(new Date(value)):'never'
const duration=(value:number)=>value?`${Math.round(value/1_000_000)} ms`:'—'
const total=(counts?:Record<string,number>)=>Object.values(counts??{}).reduce((sum,value)=>sum+value,0)
</script>
<style scoped>.source-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:16px}.source-title{display:flex;align-items:center;justify-content:space-between;gap:12px;flex-wrap:wrap}dl{display:grid;grid-template-columns:1fr 1fr;gap:12px;margin:0}dl div{min-width:0}dt{color:var(--ic-text-muted);font-size:11px;text-transform:uppercase}dd{margin:3px 0 0;overflow-wrap:anywhere}@media(max-width:720px){.source-grid,dl{grid-template-columns:1fr}}</style>
