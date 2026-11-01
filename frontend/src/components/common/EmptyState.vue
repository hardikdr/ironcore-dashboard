<template>
  <div class="empty-state d-flex flex-column align-center justify-center text-center py-10 px-6">
    <v-icon :color="iconColor" size="56" class="mb-4 empty-icon">{{ icon }}</v-icon>
    <div class="text-h6 font-weight-semibold mb-1">{{ title }}</div>
    <div v-if="subtitle" class="text-body-2 text-medium-emphasis mb-4" style="max-width:320px">{{ subtitle }}</div>
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  icon?: string
  title: string
  subtitle?: string
  variant?: 'empty' | 'error' | 'no-results'
}>(), {
  icon: undefined,
  subtitle: undefined,
  variant: 'empty',
})

const resolvedIcon = computed(() => {
  if (props.icon) return props.icon
  if (props.variant === 'error') return 'mdi-alert-circle-outline'
  if (props.variant === 'no-results') return 'mdi-magnify-close'
  return 'mdi-tray-remove'
})

const icon = computed(() => resolvedIcon.value)

const iconColor = computed(() => {
  if (props.variant === 'error') return 'error'
  return 'medium-emphasis'
})
</script>

<style scoped>
.empty-icon { opacity: 0.4; }
</style>
