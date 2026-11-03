<template><AppPage :title="node?.resource.provenance.name || 'Physical node'" subtitle="Observed physical placement and reverse dependencies"><v-alert v-if="error" type="error" variant="tonal">{{ error }}</v-alert><v-skeleton-loader v-else-if="loading" type="heading,card"/><template v-else-if="node"><v-card variant="outlined"><v-card-text><div class="facts"><div><span>State</span><strong>{{ node.resource.state || 'unknown' }}</strong></div><div><span>Health</span><strong>{{ node.resource.health || 'unknown' }}</strong></div><div><span>Source</span><strong>{{ node.resource.provenance.sourceId }}</strong></div><div><span>Addresses</span><strong>{{ node.resource.addresses?.join(', ') || 'unknown' }}</strong></div></div></v-card-text></v-card><v-card variant="outlined" class="mt-5"><v-card-title>Incoming and outgoing relationships</v-card-title><v-card-text><TopologyRelations v-if="graph" :graph="graph" :root-key="node.resource.key"/></v-card-text></v-card></template></AppPage></template>
<script setup lang="ts">
import { onMounted,ref } from 'vue'
import { useRoute } from 'vue-router'
import { sreApi,type SreGraph,type SreResourceDTO } from '@/api/sre'
import AppPage from '@/components/app/AppPage.vue'
import TopologyRelations from '@/components/sre/TopologyRelations.vue'
const route=useRoute(),node=ref<SreResourceDTO|null>(null),graph=ref<SreGraph|null>(null),loading=ref(true),error=ref<string|null>(null)
onMounted(async()=>{try{const id=String(route.params.id),[detail,topology]=await Promise.all([sreApi.node(id),sreApi.graph(id)]);node.value=detail.data;graph.value=topology.data}catch(cause){error.value=cause instanceof Error?cause.message:'Node unavailable'}finally{loading.value=false}})
</script>
<style scoped>.facts{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px}.facts span{display:block;color:var(--ic-text-muted);font-size:11px;text-transform:uppercase}.facts strong{overflow-wrap:anywhere}@media(max-width:700px){.facts{grid-template-columns:1fr 1fr}}</style>
