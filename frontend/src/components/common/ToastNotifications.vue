<template>
  <div class="toast-stack">
    <transition-group name="toast">
      <v-snackbar
        v-for="t in toasts"
        :key="t.id"
        :model-value="true"
        :color="t.color"
        :timeout="t.timeout"
        location="bottom right"
        variant="tonal"
        rounded="lg"
        class="mb-1"
        @update:model-value="store.dismiss(t.id)"
      >
        <v-icon class="mr-2" size="18">{{ iconFor(t.color) }}</v-icon>
        {{ t.message }}
        <template #actions>
          <v-btn icon="mdi-close" size="x-small" variant="text" @click="store.dismiss(t.id)" />
        </template>
      </v-snackbar>
    </transition-group>
  </div>
</template>

<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useToastStore } from '@/stores/toast'

const store = useToastStore()
const { toasts } = storeToRefs(store)

function iconFor(color: string) {
  if (color === 'error')   return 'mdi-alert-circle-outline'
  if (color === 'success') return 'mdi-check-circle-outline'
  if (color === 'warning') return 'mdi-alert-outline'
  return 'mdi-information-outline'
}
</script>
