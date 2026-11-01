<template>
  <div class="nav-group">
    <!-- Collapsed (rail) mode: just a divider -->
    <template v-if="rail">
      <v-divider class="my-2" />
    </template>

    <!-- Expanded mode: collapsible section -->
    <template v-else>
      <button
        class="nav-group-header d-flex align-center justify-space-between px-3 py-1 w-100"
        :aria-expanded="expanded"
        :aria-label="`${label} section`"
        @click="toggle"
      >
        <span class="nav-section-label">{{ label }}</span>
        <v-icon size="14" :class="{ 'rotated': !expanded }" class="nav-chevron">mdi-chevron-down</v-icon>
      </button>

      <v-expand-transition>
        <div v-show="expanded">
          <slot />
        </div>
      </v-expand-transition>
    </template>

    <!-- Rail mode: always show slot items -->
    <template v-if="rail">
      <slot />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const props = defineProps<{
  label: string
  rail: boolean
  storageKey?: string
  defaultOpen?: boolean
}>()

const expanded = ref(props.defaultOpen !== false)

const key = () => `nav-group-${props.storageKey ?? props.label.toLowerCase()}`

onMounted(() => {
  const saved = localStorage.getItem(key())
  if (saved !== null) expanded.value = saved === 'true'
})

function toggle() {
  expanded.value = !expanded.value
  localStorage.setItem(key(), String(expanded.value))
}
</script>

<style scoped>
.nav-group-header {
  background: none;
  border: none;
  cursor: pointer;
  outline: none;
  border-radius: 6px;
  transition: background 0.15s;
}
.nav-group-header:hover { background: rgba(99, 102, 241, 0.06); }
.nav-group-header:focus-visible { outline: 2px solid rgba(99,102,241,0.5); }
.nav-section-label {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(148, 163, 184, 0.6);
}
.nav-chevron { transition: transform 0.2s ease; opacity: 0.5; }
.nav-chevron.rotated { transform: rotate(-90deg); }
</style>
