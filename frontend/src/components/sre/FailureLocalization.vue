<template><v-card variant="outlined"><v-card-title>Failure localization</v-card-title><v-card-subtitle>{{ summary.localization }}</v-card-subtitle><v-card-text><div class="fact-grid"><div v-for="fact in summary.facts" :key="fact.label" class="fact"><span>{{ fact.label }}</span><strong>{{ fact.value || 'unknown' }}</strong><EvidenceBadge :fact="fact" @explain="emit('explain',$event)"/></div></div><v-alert v-for="gap in summary.gaps" :key="gap.area" type="warning" variant="tonal" density="compact" class="mt-3"><strong>{{ gap.area }} · {{ gap.state }}</strong><br>{{ gap.message }}</v-alert></v-card-text></v-card></template>
<script setup lang="ts">
import type { SreFact,SreVMSummary } from '@/api/sre'
import EvidenceBadge from './EvidenceBadge.vue'
defineProps<{ summary:SreVMSummary }>()
const emit=defineEmits<{ explain:[fact:SreFact] }>()
</script>
<style scoped>.fact-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}.fact{display:grid;grid-template-columns:1fr auto;gap:5px;padding:12px;border:1px solid var(--ic-border);border-radius:8px}.fact>span{grid-column:1/-1;color:var(--ic-text-muted);font-size:11px;text-transform:uppercase}.fact strong{overflow-wrap:anywhere}@media(max-width:800px){.fact-grid{grid-template-columns:1fr 1fr}}@media(max-width:520px){.fact-grid{grid-template-columns:1fr}}</style>
