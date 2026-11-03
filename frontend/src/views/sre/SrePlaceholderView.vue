<template>
  <section class="sre-placeholder">
    <p class="eyebrow">SRE Console · Read-only</p>
    <h1>{{ title }}</h1>
    <p>This operational view is backed by immutable, sanitized snapshots. Detailed visualization is added in the next implementation slice.</p>
  </section>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
const route = useRoute()
const title = computed(() => String(route.meta.title ?? 'Operations'))
</script>
<style scoped>
.sre-placeholder{max-width:1100px;margin:0 auto;padding:32px}.eyebrow{text-transform:uppercase;letter-spacing:.08em;color:var(--ic-text-muted);font-size:12px;font-weight:700}h1{margin:8px 0 12px}p{max-width:680px;color:var(--ic-text-muted)}
</style>
