<template><AppPage title="Networking" subtitle="Search networks and load balancers; selector matches are resolved by the server evidence graph"><v-row><v-col cols="12" lg="7"><v-card variant="outlined"><v-card-title>Find a network resource</v-card-title><v-card-text><GlobalSreSearch/></v-card-text></v-card></v-col><v-col cols="12" lg="5"><v-card variant="outlined"><v-card-title>Selector evidence</v-card-title><v-card-text>Load-balancer targets are never inferred in the browser. Open a result to see matched NIC, VM, and zone counts with selector-match evidence.</v-card-text></v-card></v-col></v-row></AppPage></template>
<script setup lang="ts">
import AppPage from '@/components/app/AppPage.vue'
import GlobalSreSearch from '@/components/sre/GlobalSreSearch.vue'
</script>
