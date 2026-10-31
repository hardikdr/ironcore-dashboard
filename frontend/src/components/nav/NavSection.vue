<template>
  <div v-if="!rail" class="nav-section-label px-3 mb-1">{{ label }}</div>
  <v-divider v-else class="my-1" />
</template>

<script setup lang="ts">
defineProps<{ label: string; rail: boolean }>()
</script>

<style scoped>
.nav-section-label {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 0.6);
  padding-top: 4px;
}
</style>
