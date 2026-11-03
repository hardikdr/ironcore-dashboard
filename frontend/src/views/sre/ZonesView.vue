<template><AppPage title="Zones" subtitle="Physical capacity and dependent workloads by zone"><v-card variant="outlined"><v-list v-if="landscape.zones.length" lines="two"><v-list-item v-for="zone in landscape.zones" :key="zone.id" :to="`/sre/zones/${encodeURIComponent(zone.resource.provenance.name)}`" :title="zone.resource.provenance.name" :subtitle="`${zone.resource.provenance.environment || 'unknown environment'} · ${zone.resource.health || 'unknown health'}`"><template #prepend><v-icon>mdi-map-marker-radius-outline</v-icon></template></v-list-item></v-list><v-card-text v-else>No zone evidence is available.</v-card-text></v-card></AppPage></template>
<script setup lang="ts">
import AppPage from '@/components/app/AppPage.vue'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
const landscape=useSreLandscapeStore()
</script>
