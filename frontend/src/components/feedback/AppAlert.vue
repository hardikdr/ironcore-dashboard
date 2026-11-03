<template><v-alert :type="type" variant="tonal" border="start"><slot /></v-alert></template>
<script setup lang="ts">
withDefaults(defineProps<{ type?: 'error' | 'warning' | 'info' | 'success' }>(), { type: 'info' })
</script>
