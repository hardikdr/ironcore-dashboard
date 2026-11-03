<template><AppPage :title="summary?.vm.provenance.name || 'Virtual machine'" subtitle="Read-only troubleshooting workspace">
  <template #actions><v-btn variant="tonal" prepend-icon="mdi-code-json" :loading="rawLoading" @click="openRaw">Sanitized source</v-btn></template>
  <v-alert v-if="error" type="error" variant="tonal">{{ error }}</v-alert><v-skeleton-loader v-else-if="loading" type="heading, paragraph, card"/>
  <template v-else-if="summary"><FailureLocalization :summary="summary" @explain="explain"/><v-tabs v-model="tab" class="mt-5" show-arrows><v-tab v-for="item in tabs" :key="item" :value="item">{{ item }}</v-tab></v-tabs><v-window v-model="tab"><v-window-item v-for="item in tabs" :key="item" :value="item"><v-card variant="outlined" class="tab-panel"><v-card-title>{{ item }}</v-card-title><v-card-text><template v-if="item==='Overview'"><dl><div><dt>State</dt><dd>{{ summary.vm.state || 'unknown' }}</dd></div><div><dt>Health</dt><dd>{{ summary.vm.health || 'unknown' }}</dd></div><div><dt>Source</dt><dd>{{ summary.vm.provenance.sourceId || 'unknown' }}</dd></div></dl></template><template v-else-if="item==='Events/source'">Open “Sanitized source” to inspect the redacted source object. Collection gaps are shown above.</template><template v-else>Related {{ item.toLowerCase() }} facts are shown in failure localization. Dependency evidence is available on each badge.</template></v-card-text></v-card></v-window-item></v-window></template>
  <v-dialog v-model="evidenceOpen" max-width="560"><v-card><v-card-title>Why this relation is shown</v-card-title><v-card-text v-if="selectedFact"><dl><div><dt>Fact</dt><dd>{{ selectedFact.label }}: {{ selectedFact.value }}</dd></div><div><dt>Evidence</dt><dd>{{ selectedFact.evidence || 'unknown' }}</dd></div><div><dt>Confidence</dt><dd>{{ selectedFact.confidence || 'unknown' }}</dd></div><div><dt>Source</dt><dd>{{ selectedFact.sourceId || 'unknown' }}</dd></div><div><dt>Resource key</dt><dd>{{ selectedFact.resource || 'unknown' }}</dd></div></dl></v-card-text><v-card-actions><v-spacer/><v-btn @click="evidenceOpen=false">Close</v-btn></v-card-actions></v-card></v-dialog>
  <SourceDrawer v-model="rawOpen" :details="rawDetails"/>
</AppPage></template>
<script setup lang="ts">
import { onMounted,ref,watch } from 'vue'
import { useRoute } from 'vue-router'
import { sreApi,type SreFact,type SreVMSummary } from '@/api/sre'
import AppPage from '@/components/app/AppPage.vue'
import FailureLocalization from '@/components/sre/FailureLocalization.vue'
import SourceDrawer from '@/components/sre/SourceDrawer.vue'
const route=useRoute(),summary=ref<SreVMSummary|null>(null),loading=ref(true),error=ref<string|null>(null),tab=ref('Overview'),tabs=['Overview','Compute path','Networking','Storage','Ownership','Events/source'],evidenceOpen=ref(false),selectedFact=ref<SreFact|null>(null),rawOpen=ref(false),rawLoading=ref(false),rawDetails=ref<unknown>(null)
async function load(){loading.value=true;error.value=null;try{summary.value=(await sreApi.vm(String(route.params.id))).data}catch(cause){error.value=cause instanceof Error?cause.message:'VM diagnosis unavailable'}finally{loading.value=false}}
function explain(fact:SreFact){selectedFact.value=fact;evidenceOpen.value=true}
async function openRaw(){rawLoading.value=true;try{rawDetails.value=(await sreApi.sanitized(String(route.params.id))).data.details;rawOpen.value=true}catch(cause){error.value=cause instanceof Error?cause.message:'Sanitized source unavailable'}finally{rawLoading.value=false}}
onMounted(load);watch(()=>route.params.id,load)
</script>
<style scoped>.tab-panel{border-top-left-radius:0;border-top-right-radius:0;min-height:180px}dl{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px;margin:0}dl div{min-width:0}dt{color:var(--ic-text-muted);font-size:11px;text-transform:uppercase}dd{margin:3px 0 0;overflow-wrap:anywhere}@media(max-width:600px){dl{grid-template-columns:1fr}}</style>
