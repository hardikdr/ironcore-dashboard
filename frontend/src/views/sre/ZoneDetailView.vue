<template>
  <AppPage :title="summary?.zone.provenance.name || String(route.params.zone)" subtitle="Observed capacity, workloads and relationships" :breadcrumbs="[{ title: 'Zones', to: '/sre/zones' }]">
    <template #actions><v-btn variant="outlined" prepend-icon="mdi-refresh" :loading="loading" @click="load">Refresh snapshot</v-btn></template>
    <v-alert v-if="error" type="error" variant="tonal" class="mb-4">{{ error }}</v-alert>
    <v-skeleton-loader v-if="loading && !summary" type="heading,card" />
    <template v-if="summary">
      <v-alert v-if="summary.gaps?.length" type="warning" variant="tonal" class="mb-4" title="Partial zone evidence">{{ summary.gaps.map(g => `${g.area}: ${g.state}`).join(' · ') }}</v-alert>
      <p class="evidence-note">Counts reflect the collected snapshot. Zero means not observed; it does not prove absence.</p>
      <div class="metrics"><div v-for="item in metrics" :key="item[0]"><span>{{ item[0] }}</span><strong>{{ item[1] }}</strong></div></div>
      <v-card variant="outlined" class="mt-5">
        <v-card-title>Related inventory</v-card-title>
        <v-card-text>
          <p class="evidence-note">Resources connected to this zone in a bounded topology traversal. Relationships do not necessarily imply physical placement.</p>
          <v-text-field v-model="query" label="Filter by name, kind or namespace" prepend-inner-icon="mdi-magnify" clearable hide-details class="mb-4" />
          <v-alert v-if="graphError" type="warning" variant="tonal">{{ graphError }}</v-alert>
          <v-alert v-else-if="graph?.truncated" type="info" variant="tonal" class="mb-3">This is a partial inventory because the traversal limit was reached.</v-alert>
          <div class="inventory-wrap"><table v-if="resources.length"><thead><tr><th>Resource</th><th>Kind</th><th>Namespace</th><th>Observed state</th></tr></thead><tbody><tr v-for="resource in resources" :key="resource.key"><td><router-link :to="resourceRoute(resource)">{{ resource.provenance.name }}</router-link></td><td>{{ resource.provenance.kind }}</td><td>{{ resource.provenance.namespace || '—' }}</td><td>{{ resource.health || resource.state || 'Unknown' }}</td></tr></tbody></table>
          <p v-else-if="!graphError" class="evidence-note">{{ query ? 'No resources match this filter.' : 'No related resources were observed in the available snapshot.' }}</p></div>
        </v-card-text>
      </v-card>
      <v-card variant="outlined" class="mt-5"><v-card-title>Direct relationships</v-card-title><v-card-text><TopologyRelations v-if="graph" :graph="graph" :root-key="summary.zone.key" /><p v-else>No relationship graph is available.</p></v-card-text></v-card>
    </template>
  </AppPage>
</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { sreApi, type SreGraph, type SreZoneSummary } from '@/api/sre'
import { resourceId, resourceRoute } from '@/lib/sreNavigation'
import AppPage from '@/components/app/AppPage.vue'
import TopologyRelations from '@/components/sre/TopologyRelations.vue'
const route = useRoute(), summary = ref<SreZoneSummary | null>(null), graph = ref<SreGraph | null>(null)
const loading = ref(false), error = ref<string | null>(null), graphError = ref<string | null>(null), query = ref('')
const controller = new AbortController()
const metrics = computed(() => summary.value ? [['Management nodes', summary.value.managementNodes], ['Compute nodes', summary.value.computeNodes], ['Storage nodes', summary.value.storageNodes], ['Machine pools', summary.value.machinePools], ['Volume pools', summary.value.volumePools], ['Hosted VMs', summary.value.hostedVms], ['Dependent Shoots', summary.value.dependentShoots]] : [])
const resources = computed(() => {
  const term = (query.value || '').trim().toLowerCase()
  return (graph.value?.resources || []).filter(resource => resource.key !== summary.value?.zone.key && [resource.provenance.name, resource.provenance.kind, resource.provenance.namespace].join(' ').toLowerCase().includes(term))
})
async function load() {
  if (loading.value) return
  loading.value = true; error.value = null; graphError.value = null
  try {
    summary.value = (await sreApi.zone(String(route.params.zone), controller.signal)).data
    graph.value = null
    try { graph.value = (await sreApi.graph(resourceId(summary.value.zone.key), controller.signal)).data }
    catch (cause) { if (!controller.signal.aborted) graphError.value = cause instanceof Error ? cause.message : 'Relationships unavailable' }
  } catch (cause) { if (!controller.signal.aborted) error.value = cause instanceof Error ? cause.message : 'Zone unavailable' }
  finally { loading.value = false }
}
onMounted(load)
onUnmounted(() => controller.abort())
</script>
<style scoped>
.evidence-note{color:var(--ic-text-muted);font-size:13px;margin:0 0 16px;line-height:1.6}.metrics{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:12px}.metrics>div{padding:16px;border:1px solid var(--ic-border);border-radius:var(--ic-radius-md);background:var(--ic-surface)}.metrics span{display:block;color:var(--ic-text-muted);font-size:12px}.metrics strong{font-size:24px}.inventory-wrap{overflow-x:auto}table{width:100%;border-collapse:collapse;font-size:13px}th,td{text-align:left;padding:12px;border-bottom:1px solid var(--ic-border)}th{color:var(--ic-text-muted);font-weight:500}a{color:rgb(var(--v-theme-primary));overflow-wrap:anywhere}@media(max-width:700px){.metrics{grid-template-columns:1fr 1fr}th,td{padding:10px 8px}}
</style>
