<template><AppPage title="On-call Triage" subtitle="Evidence-derived observations ordered by operational severity">
  <template #actions><SnapshotStamp :generated-at="landscape.meta?.generatedAt" :partial="landscape.meta?.partial"/></template>
  <v-alert v-if="landscape.meta?.partial" type="warning" variant="tonal" class="mb-5" title="Triage is based on partial evidence">Missing source data can hide impact. Unknowns are not treated as healthy.</v-alert>
  <v-card variant="outlined"><v-list v-if="ordered.length" lines="three"><v-list-item v-for="item in ordered" :key="item.key" :title="item.reason || 'Observation'" :subtitle="item.message"><template #prepend><v-icon :color="color(item.severity)">mdi-alert-circle-outline</v-icon></template><template #append><div class="observation-meta"><v-chip size="x-small" :color="color(item.severity)" variant="tonal">{{ item.severity || 'unknown' }}</v-chip><span>{{ item.confidence }} confidence</span></div></template></v-list-item></v-list><v-card-text v-else class="empty">No observations were derived from the current snapshot.</v-card-text></v-card>
</AppPage></template>
<script setup lang="ts">
import { computed } from 'vue'
import AppPage from '@/components/app/AppPage.vue'
import SnapshotStamp from '@/components/sre/SnapshotStamp.vue'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
const landscape=useSreLandscapeStore(),rank:Record<string,number>={critical:0,error:1,warning:2,info:3}
const ordered=computed(()=>[...landscape.observations].sort((a,b)=>(rank[a.severity]??9)-(rank[b.severity]??9)))
const color=(severity:string)=>severity==='critical'||severity==='error'?'error':severity==='warning'?'warning':'info'
</script>
<style scoped>.observation-meta{display:flex;flex-direction:column;align-items:flex-end;gap:4px;font-size:11px;color:var(--ic-text-muted)}.empty{color:var(--ic-text-muted)}@media(max-width:600px){.observation-meta span{display:none}}</style>
