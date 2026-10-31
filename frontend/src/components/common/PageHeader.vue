<template>
  <div class="d-flex align-center justify-space-between mb-6 flex-wrap ga-3">
    <div>
      <!-- Breadcrumb -->
      <nav v-if="breadcrumbs?.length" aria-label="Breadcrumb" class="d-flex align-center ga-1 mb-1">
        <template v-for="(crumb, i) in breadcrumbs" :key="crumb.text">
          <router-link v-if="crumb.to" :to="crumb.to" class="text-caption text-medium-emphasis text-decoration-none breadcrumb-link">
            {{ crumb.text }}
          </router-link>
          <span v-else class="text-caption text-medium-emphasis" aria-current="page">{{ crumb.text }}</span>
          <v-icon v-if="i < breadcrumbs.length - 1" size="12" class="text-disabled" aria-hidden="true">mdi-chevron-right</v-icon>
        </template>
      </nav>
      <!-- Title row -->
      <div class="d-flex align-center ga-3">
        <div
          v-if="icon"
          class="d-flex align-center justify-center rounded-lg"
          :style="iconStyle"
        >
          <v-icon :color="iconColor ?? 'primary'" size="20">{{ icon }}</v-icon>
        </div>
        <div>
          <h1 class="text-h5 font-weight-bold">{{ title }}</h1>
          <div v-if="subtitle" class="text-caption text-medium-emphasis mt-0">{{ subtitle }}</div>
        </div>
        <slot name="title-append" />
      </div>
    </div>

    <div class="d-flex ga-2 align-center flex-wrap">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCardStyle } from '@/composables/useCardStyle'

defineProps<{
  title:    string
  subtitle?: string
  icon?:    string
  iconColor?: string
  breadcrumbs?: { text: string; to?: string }[]
}>()

const { isDark } = useCardStyle()

const iconStyle = computed(() => ({
  width: '42px',
  height: '42px',
  background: isDark.value ? 'rgba(129,140,248,0.12)' : 'rgba(99,102,241,0.1)',
  flexShrink: '0',
}))
</script>

<style scoped>
.breadcrumb-link:hover { text-decoration: underline; }
</style>
