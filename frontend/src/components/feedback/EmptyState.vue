<template><div class="empty-state"><v-icon :icon="icon" size="42" color="medium-emphasis"/><h2>{{ title }}</h2><p>{{ message }}</p><slot /></div></template>
<script setup lang="ts">
withDefaults(defineProps<{ title: string; message: string; icon?: string }>(), { icon: 'mdi-tray' })
</script>
<style scoped>.empty-state{text-align:center;padding:48px 24px;border:1px dashed var(--ic-border);border-radius:var(--ic-radius-md)}h2{font-size:16px;margin:12px 0 4px}p{color:var(--ic-text-muted);font-size:14px;margin:0 0 16px}</style>
