<template>
  <v-list nav density="compact">
    <NavItem to="/sre/landscape" icon="mdi-earth" label="Landscape" :rail="rail"/>
    <NavItem to="/sre/triage" icon="mdi-alert-decagram-outline" label="Triage" :rail="rail"/>
    <NavSection label="Infrastructure" :rail="rail"/>
    <NavItem to="/sre/zones" icon="mdi-map-marker-radius-outline" label="Zones" :rail="rail"/>
    <NavItem to="/sre/vms" icon="mdi-server-outline" label="Virtual Machines" :rail="rail"/>
    <NavItem to="/sre/shoots" icon="mdi-sprout-outline" label="Shoots" :rail="rail"/>
    <NavSection label="Dependencies" :rail="rail"/>
    <NavItem to="/sre/networking" icon="mdi-lan" label="Networking" :rail="rail"/>
    <NavItem to="/sre/storage" icon="mdi-database-outline" label="Storage" :rail="rail"/>
    <NavItem to="/sre/sources" icon="mdi-database-sync-outline" label="Sources" :rail="rail"/>
  </v-list>
</template>
<script setup lang="ts">
import NavItem from '@/components/nav/NavItem.vue'
import NavSection from '@/components/nav/NavSection.vue'
defineProps<{ rail: boolean }>()
</script>
