<template><v-tabs :model-value="value" show-arrows @update:model-value="$emit('update:value', String($event))"><v-tab v-for="tab in tabs" :key="tab.value" :value="tab.value" :prepend-icon="tab.icon">{{ tab.label }}</v-tab></v-tabs></template>
<script setup lang="ts">
defineProps<{ value: string; tabs: { value: string; label: string; icon?: string }[] }>()
defineEmits<{ 'update:value': [value: string] }>()
</script>
