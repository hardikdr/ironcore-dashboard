<template>
  <v-app theme="light">
    <TopBar />
    <v-navigation-drawer permanent width="220">
      <Sidebar />
    </v-navigation-drawer>
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import TopBar from '@/components/TopBar.vue'
import Sidebar from '@/components/Sidebar.vue'
import { useNamespaceStore } from '@/stores/namespace'
import { onMounted } from 'vue'

const nsStore = useNamespaceStore()
onMounted(() => nsStore.load())
</script>
