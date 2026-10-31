<template>
  <v-chip
    v-bind="$attrs"
    :color="copied ? 'success' : (color ?? 'default')"
    :variant="(variant ?? 'tonal') as any"
    size="x-small"
    class="cursor-pointer"
    :title="copied ? 'Copied!' : `Click to copy: ${text}`"
    @click.stop="doCopy"
  >
    <slot>{{ text }}</slot>
    <v-icon v-if="copied" size="11" class="ml-1">mdi-check</v-icon>
    <v-icon v-else size="11" class="ml-1">mdi-content-copy</v-icon>
  </v-chip>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = defineProps<{ text: string; color?: string; variant?: string }>()
const copied = ref(false)

async function doCopy() {
  try {
    await navigator.clipboard.writeText(props.text)
    copied.value = true
    setTimeout(() => { copied.value = false }, 1500)
  } catch { /* non-https: silently skip */ }
}
</script>
