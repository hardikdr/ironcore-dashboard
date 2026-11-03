<template><AppPage :title="summary?.shoot.provenance.name || 'Shoot'" subtitle="Gardener ownership and worker placement"><v-alert v-if="error" type="error" variant="tonal">{{ error }}</v-alert><v-skeleton-loader v-else-if="loading" type="heading,card"/><template v-else-if="summary"><v-alert v-if="summary.gaps.length" type="warning" variant="tonal" class="mb-4" title="Partial Gardener topology">Some source inventories are denied, stale, or unavailable; observed relations remain visible.</v-alert><GardenerRoleSummary :installation-role="summary.shoot.provenance.installationRole" :provider-type="String(summary.shoot.sanitized?.providerType || 'unknown')" :worker-vms="summary.workerVms"/><v-card variant="outlined" class="mt-5"><v-card-title>Worker topology</v-card-title><v-card-subtitle>{{ summary.workerGroups }} worker groups · {{ summary.relationConfidence }} relation confidence</v-card-subtitle><v-card-text><TopologyRelations v-if="graph" :graph="graph" :root-key="summary.shoot.key"/></v-card-text></v-card></template></AppPage></template>
<script setup lang="ts">
import { onMounted,ref } from 'vue'
import { useRoute } from 'vue-router'
import { sreApi,type SreGraph,type SreShootSummary } from '@/api/sre'
import AppPage from '@/components/app/AppPage.vue'
import GardenerRoleSummary from '@/components/sre/GardenerRoleSummary.vue'
import TopologyRelations from '@/components/sre/TopologyRelations.vue'
const route=useRoute(),summary=ref<SreShootSummary|null>(null),graph=ref<SreGraph|null>(null),loading=ref(true),error=ref<string|null>(null)
onMounted(async()=>{try{const id=String(route.params.id),[detail,topology]=await Promise.all([sreApi.shoot(id),sreApi.graph(id)]);summary.value=detail.data;graph.value=topology.data}catch(cause){error.value=cause instanceof Error?cause.message:'Shoot unavailable'}finally{loading.value=false}})
</script>
