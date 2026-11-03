<template><v-navigation-drawer :model-value="modelValue" location="right" temporary width="560" @update:model-value="emit('update:modelValue',$event)"><div class="drawer-header"><div><strong>Sanitized source</strong><p>Read-only fields returned by the SRE redaction boundary.</p></div><v-btn icon="mdi-close" variant="text" aria-label="Close source drawer" @click="emit('update:modelValue',false)"/></div><pre tabindex="0">{{ formatted }}</pre></v-navigation-drawer></template>
<script setup lang="ts">
import { computed } from 'vue'
const props=defineProps<{ modelValue:boolean; details:unknown }>(),emit=defineEmits<{ 'update:modelValue':[value:boolean] }>()
const formatted=computed(()=>JSON.stringify(props.details??{},null,2))
</script>
<style scoped>.drawer-header{display:flex;justify-content:space-between;gap:12px;padding:20px;border-bottom:1px solid var(--ic-border)}.drawer-header p{margin:4px 0 0;color:var(--ic-text-muted);font-size:12px}pre{margin:0;padding:20px;white-space:pre-wrap;overflow-wrap:anywhere;font:12px/1.55 ui-monospace;color:var(--ic-text);user-select:text}</style>
