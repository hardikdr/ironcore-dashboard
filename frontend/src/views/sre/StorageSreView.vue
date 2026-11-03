<template><AppPage title="Storage" subtitle="Volume placement and cross-zone evidence"><v-alert v-if="crossZone.length" type="warning" variant="tonal" class="mb-5" title="Cross-zone observations"><ul><li v-for="item in crossZone" :key="item.key">{{ item.message }}</li></ul></v-alert><v-row><v-col cols="12" lg="7"><v-card variant="outlined"><v-card-title>Find a volume or pool</v-card-title><v-card-text><GlobalSreSearch/></v-card-text></v-card></v-col><v-col cols="12" lg="5"><v-card variant="outlined"><v-card-title>Placement policy</v-card-title><v-card-text>VolumePool placement is shown only where the backend graph provides object-reference or label evidence. Unknown placement remains explicit.</v-card-text></v-card></v-col></v-row></AppPage></template>
<script setup lang="ts">
import { computed } from 'vue'
import AppPage from '@/components/app/AppPage.vue'
import GlobalSreSearch from '@/components/sre/GlobalSreSearch.vue'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
const landscape=useSreLandscapeStore(),crossZone=computed(()=>landscape.observations.filter(item=>`${item.reason} ${item.message}`.toLowerCase().includes('zone')))
</script>
