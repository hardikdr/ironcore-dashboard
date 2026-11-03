<template><dl class="property-list"><template v-for="item in items" :key="item.label"><dt>{{ item.label }}</dt><dd :class="{ mono:item.mono }"><slot :name="item.label" :item="item">{{ item.value || '—' }}<v-btn v-if="item.copyable && item.value" icon="mdi-content-copy" size="x-small" variant="text" :aria-label="`Copy ${item.label}`" @click="copy(item.value)"/></slot></dd></template></dl></template>
<script setup lang="ts">
defineProps<{ items: { label: string; value: string; copyable?: boolean; mono?: boolean }[] }>()
function copy(value: string) { void window.navigator.clipboard.writeText(value) }
</script>
<style scoped>.property-list{display:grid;grid-template-columns:minmax(140px,220px) 1fr;margin:0;border-top:1px solid var(--ic-border)}dt,dd{margin:0;padding:12px;border-bottom:1px solid var(--ic-border);font-size:14px}dt{color:var(--ic-text-muted)}.mono{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}@media(max-width:600px){.property-list{grid-template-columns:1fr}dt{padding-bottom:2px;border-bottom:0}dd{padding-top:2px}}</style>
