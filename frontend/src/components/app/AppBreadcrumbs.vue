<template><nav v-if="items?.length" aria-label="Breadcrumb"><v-breadcrumbs :items="items" density="compact" class="pa-0 mb-2 text-caption" /></nav></template>
<script setup lang="ts">
export interface BreadcrumbItem { title: string; to?: string; disabled?: boolean }
defineProps<{ items?: BreadcrumbItem[] }>()
</script>
