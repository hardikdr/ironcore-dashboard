<template><div class="snapshot-stamp"><v-icon size="16">mdi-camera-outline</v-icon><span>Snapshot {{ formatted }}</span><v-chip v-if="partial" size="x-small" color="warning" variant="tonal">Partial</v-chip></div></template>
<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ generatedAt?: string; partial?: boolean }>()
const formatted = computed(() => props.generatedAt ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'medium' }).format(new Date(props.generatedAt)) : 'not available')
</script>
<style scoped>.snapshot-stamp{display:flex;align-items:center;gap:7px;color:var(--ic-text-muted);font-size:12px;flex-wrap:wrap}</style>
