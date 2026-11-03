<template><v-chip size="x-small" variant="tonal" :color="color" :clickable="explainable" :aria-label="`Explain evidence for ${fact.label}`" @click="explainable && emit('explain', fact)"><v-icon start size="13">mdi-link-variant</v-icon>{{ fact.confidence || 'unknown' }}</v-chip></template>
<script setup lang="ts">
import { computed } from 'vue'
import type { SreFact } from '@/api/sre'
const props=defineProps<{ fact:SreFact }>(),emit=defineEmits<{ explain:[fact:SreFact] }>()
const explainable=computed(()=>Boolean(props.fact.evidence||props.fact.sourceId||props.fact.resource))
const color=computed(()=>props.fact.confidence==='proven'?'success':props.fact.confidence==='corroborated'?'info':props.fact.confidence==='derived'?'warning':'default')
</script>
