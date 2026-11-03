<template><AppPage title="Virtual Machines" subtitle="Server-backed operational search; inventories are not preloaded into the browser"><v-card variant="outlined"><v-card-title>Find the affected VM</v-card-title><v-card-subtitle>Exact identifiers rank above prefixes and tokens.</v-card-subtitle><v-card-text><GlobalSreSearch/></v-card-text></v-card></AppPage></template>
<script setup lang="ts">
import AppPage from '@/components/app/AppPage.vue'
import GlobalSreSearch from '@/components/sre/GlobalSreSearch.vue'
</script>
