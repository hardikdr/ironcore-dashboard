<template>
  <v-card variant="outlined" rounded="lg">
    <v-card-text class="pa-4">
      <div class="text-h4 font-weight-bold" :style="{ color: color ?? valueColor }">{{ value }}</div>
      <div class="text-caption text-uppercase text-medium-emphasis mt-1">{{ label }}</div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ value: string | number; label: string; valueColor?: string; color?: string }>()
</script>
