<template><AppPage title="Dependency Explorer" subtitle="Directed topology rendered only from server-side graph evidence"><v-alert v-if="error" type="error" variant="tonal">{{ error }}</v-alert><v-skeleton-loader v-else-if="loading" type="heading,card"/><EvidenceGraph v-else-if="graph&&rootKey" :graph="graph" :root-key="rootKey"/><v-alert v-else type="info" variant="tonal">No dependency evidence is available for this resource.</v-alert></AppPage></template>
<script setup lang="ts">
import { computed,onMounted,ref } from 'vue';import { useRoute } from 'vue-router';import { sreApi,type SreGraph } from '@/api/sre';import AppPage from '@/components/app/AppPage.vue';import EvidenceGraph from '@/components/sre/EvidenceGraph.vue'
const route=useRoute(),graph=ref<SreGraph|null>(null),loading=ref(true),error=ref<string|null>(null),rootKey=computed(()=>graph.value?.resources[0]?.key||'')
onMounted(async()=>{try{graph.value=(await sreApi.graph(String(route.params.id))).data}catch(cause){error.value=cause instanceof Error?cause.message:'Dependencies unavailable'}finally{loading.value=false}})
</script>
