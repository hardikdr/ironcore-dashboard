<template>
  <div class="console-mode-switch" role="group" aria-label="Console mode">
    <button type="button" :class="{ active: !isSre }" aria-label="Open User Console" @click="router.push('/overview')">User</button>
    <button type="button" :class="{ active: isSre }" aria-label="Open SRE Console" @click="router.push('/sre/landscape')">SRE</button>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
const route = useRoute()
const router = useRouter()
const isSre = computed(() => route.path.startsWith('/sre/'))
</script>
<style scoped>
.console-mode-switch{display:flex;padding:2px;border:1px solid var(--ic-border);border-radius:8px;background:var(--ic-canvas)}
button{border:0;border-radius:6px;background:transparent;color:var(--ic-text-muted);font:inherit;font-size:12px;font-weight:600;padding:5px 9px;cursor:pointer}
button.active{background:var(--ic-surface);color:var(--ic-text);box-shadow:0 1px 3px rgb(0 0 0/.16)}
</style>
