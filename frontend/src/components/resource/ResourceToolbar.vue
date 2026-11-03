<template><div class="resource-toolbar"><v-text-field :model-value="search" :label="searchLabel" prepend-inner-icon="mdi-magnify" clearable hide-details @update:model-value="$emit('update:search', $event ?? '')"/><slot name="filters"/><v-spacer/><slot name="view"/><v-btn icon="mdi-refresh" variant="text" :loading="loading" aria-label="Refresh resources" @click="$emit('refresh')"/></div></template>
<script setup lang="ts">
withDefaults(defineProps<{ search: string; searchLabel?: string; loading?: boolean }>(), { searchLabel: 'Search resources', loading: false })
defineEmits<{ 'update:search': [value: string]; refresh: [] }>()
</script>
<style scoped>.resource-toolbar{display:flex;align-items:center;gap:12px;margin-bottom:16px}.resource-toolbar :deep(.v-text-field){max-width:420px}@media(max-width:600px){.resource-toolbar :deep(.v-text-field){max-width:none;flex:1}}</style>
