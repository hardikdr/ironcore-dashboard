<template>
  <v-menu>
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        variant="outlined"
        color="white"
        size="small"
        class="text-white"
        style="border-color: rgba(255,255,255,0.4)"
      >
        <v-icon start>mdi-package-variant</v-icon>
        {{ nsStore.active }}
        <v-icon end>mdi-chevron-down</v-icon>
      </v-btn>
    </template>
    <v-list density="compact" min-width="200">
      <v-list-subheader>Switch Project / Namespace</v-list-subheader>
      <v-list-item
        v-for="ns in nsStore.namespaces"
        :key="ns"
        :value="ns"
        :active="ns === nsStore.active"
        active-color="primary"
        @click="nsStore.setActive(ns)"
      >
        <v-list-item-title>{{ ns }}</v-list-item-title>
      </v-list-item>
      <v-list-item v-if="!nsStore.namespaces.length" disabled>
        <v-list-item-title class="text-medium-emphasis">No namespaces found</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { useNamespaceStore } from '@/stores/namespace'
const nsStore = useNamespaceStore()
</script>
