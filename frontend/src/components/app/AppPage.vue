<template>
  <section class="app-page ic-page-gutter">
    <AppBreadcrumbs :items="breadcrumbs" />
    <header class="app-page__header"><div><div class="d-flex align-center ga-3"><h1>{{ title }}</h1><slot name="status"><StatusBadge v-if="status" :state="status" /></slot></div><p v-if="subtitle">{{ subtitle }}</p></div><div class="app-page__actions"><slot name="actions" /></div></header>
    <div v-if="$slots.tabs" class="app-page__tabs"><slot name="tabs" /></div><div class="app-page__content"><slot /></div>
  </section>
</template>
<script setup lang="ts">
import AppBreadcrumbs, { type BreadcrumbItem } from './AppBreadcrumbs.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
defineProps<{ title: string; subtitle?: string; breadcrumbs?: BreadcrumbItem[]; status?: string }>()
</script>
<style scoped>
.app-page{padding-block:24px 48px;max-width:1680px;margin:0 auto}.app-page__header{display:flex;justify-content:space-between;align-items:flex-start;gap:24px;margin-bottom:24px}.app-page h1{font-size:24px;line-height:1.25;font-weight:650;letter-spacing:-.02em;margin:0}.app-page p{color:var(--ic-text-muted);font-size:14px;margin:6px 0 0}.app-page__actions{display:flex;gap:8px;flex-wrap:wrap;justify-content:flex-end}.app-page__tabs{border-bottom:1px solid var(--ic-border);margin-bottom:24px}@media(max-width:600px){.app-page{padding-top:16px}.app-page__header{display:block}.app-page h1{font-size:20px}.app-page__actions{justify-content:flex-start;margin-top:16px}}
</style>
