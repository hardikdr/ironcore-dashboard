<template><div class="role-summary"><div><span>Installation role</span><strong>{{ installationRole || 'unknown' }}</strong></div><div><span>Provider type</span><strong>{{ providerType || 'unknown' }}</strong></div><div v-if="isBareMetal" class="bare"><span>Compute path</span><strong>Bare metal — no IronCore VM path</strong></div><div v-else><span>Compute path</span><strong>{{ workerVms }} worker VMs</strong></div></div></template>
<script setup lang="ts">
import { computed } from 'vue'
const props=defineProps<{ installationRole?:string; providerType?:string; workerVms:number }>()
const isBareMetal=computed(()=>props.installationRole==='bare-metal'||props.providerType==='bare-metal')
</script>
<style scoped>.role-summary{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:12px}.role-summary>div{padding:12px;border:1px solid var(--ic-border);border-radius:8px;display:flex;flex-direction:column;gap:4px}.role-summary span{color:var(--ic-text-muted);font-size:11px;text-transform:uppercase}.bare{border-color:rgb(var(--v-theme-info))!important}@media(max-width:650px){.role-summary{grid-template-columns:1fr}}</style>
