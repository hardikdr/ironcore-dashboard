<template>
  <v-app-bar color="primary" elevation="2" height="52">
    <template #prepend>
      <div
        class="d-flex align-center ga-2 px-4"
        style="min-width:220px;border-right:1px solid rgba(255,255,255,0.15)"
      >
        <v-avatar size="32" rounded="sm">
          <v-img src="/ironcore-logo.png" />
        </v-avatar>
        <div>
          <div class="text-white font-weight-bold" style="font-size:15px;line-height:1.1">IronCore</div>
          <div class="text-white" style="font-size:10px;opacity:0.6">Cloud Dashboard</div>
        </div>
      </div>
    </template>

    <v-tabs color="white" class="ml-2">
      <v-tab :to="{ path: '/machines' }"     prepend-icon="mdi-monitor">Machines</v-tab>
      <v-tab :to="{ path: '/volumes' }"      prepend-icon="mdi-database">Volumes</v-tab>
      <v-tab :to="{ path: '/networks' }"     prepend-icon="mdi-lan">Networking</v-tab>
      <v-tab :to="{ path: '/prefixes' }"     prepend-icon="mdi-ip-network">IPAM</v-tab>
    </v-tabs>

    <template #append>
      <div class="d-flex align-center ga-3 pr-4">
        <NamespaceSwitcher />
        <div class="d-flex align-center ga-1 text-caption" style="color:rgba(255,255,255,0.7)">
          <v-icon size="10" color="success">mdi-circle</v-icon>
          ironcore-in-a-box
        </div>
        <v-btn icon="mdi-help-circle-outline" color="white" variant="text" size="small" />
        <v-avatar color="secondary" size="30" style="cursor:pointer">
          <span class="text-caption font-weight-bold text-white">DV</span>
        </v-avatar>
      </div>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
import NamespaceSwitcher from './NamespaceSwitcher.vue'
</script>
