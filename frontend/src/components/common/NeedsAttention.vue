<template>
  <v-card v-if="items.length > 0" :style="urgentStyle" class="pa-0 overflow-hidden">
    <!-- Header bar -->
    <div class="d-flex align-center px-4 py-3" :style="headerStyle">
      <v-icon :color="severityColor" size="18" class="mr-2">{{ severityIcon }}</v-icon>
      <span class="text-subtitle-2 font-weight-bold">{{ severityLabel }}</span>
      <v-spacer />
      <v-chip :color="severityColor" size="x-small" variant="tonal">
        {{ items.length }} item{{ items.length > 1 ? 's' : '' }}
      </v-chip>
    </div>

    <!-- Items -->
    <div class="px-4 py-2">
      <div
        v-for="item in items"
        :key="item.key"
        class="d-flex align-center py-2 attention-item"
        style="border-bottom: 1px solid rgba(128,128,128,0.08); cursor:pointer;"
        @click="item.link ? router.push(item.link) : undefined"
      >
        <v-icon :color="item.color" size="16" class="mr-2 flex-shrink-0">{{ item.icon }}</v-icon>
        <div class="flex-grow-1 min-width-0">
          <span class="text-body-2 font-weight-medium">{{ item.count > 1 ? `${item.count} ` : '' }}{{ item.label }}</span>
          <span v-if="item.detail" class="text-caption text-medium-emphasis ml-1">{{ item.detail }}</span>
        </div>
        <v-icon v-if="item.link" size="14" color="medium-emphasis" class="ml-2 flex-shrink-0">mdi-chevron-right</v-icon>
      </div>
      <!-- Remove bottom border from last item -->
      <style scoped>.attention-item:last-child { border-bottom: none !important; }</style>
    </div>
  </v-card>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCardStyle } from '@/composables/useCardStyle'

export interface AttentionItem {
  key: string
  label: string
  count: number
  icon: string
  color: 'error' | 'warning' | 'info'
  link?: string
  detail?: string
}

const props = defineProps<{
  items: AttentionItem[]
}>()

const router = useRouter()
const { isDark, cardStyle } = useCardStyle()

const hasErrors   = computed(() => props.items.some(i => i.color === 'error'))
const hasWarnings = computed(() => props.items.some(i => i.color === 'warning'))

const severityColor = computed(() => {
  if (hasErrors.value)   return 'error'
  if (hasWarnings.value) return 'warning'
  return 'info'
})

const severityIcon = computed(() => {
  if (hasErrors.value)   return 'mdi-alert-circle'
  if (hasWarnings.value) return 'mdi-alert'
  return 'mdi-information'
})

const severityLabel = computed(() => {
  if (hasErrors.value)   return 'Needs Immediate Attention'
  if (hasWarnings.value) return 'Attention Required'
  return 'Notice'
})

const urgentStyle = computed(() => ({
  ...cardStyle.value,
  border: `1px solid rgba(var(--v-theme-${severityColor.value}), 0.35)`,
}))

const headerStyle = computed(() => ({
  background: isDark.value
    ? `rgba(var(--v-theme-${severityColor.value}), 0.08)`
    : `rgba(var(--v-theme-${severityColor.value}), 0.06)`,
  borderBottom: `1px solid rgba(var(--v-theme-${severityColor.value}), 0.15)`,
}))
</script>
