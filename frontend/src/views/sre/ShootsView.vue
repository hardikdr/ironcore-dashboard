<template><AppPage title="Shoots" subtitle="Gardener installations and their observed worker topology"><v-alert v-if="error" type="error" variant="tonal">{{ error }}</v-alert><v-card variant="outlined"><v-list v-if="items.length" lines="two"><v-list-item v-for="item in items" :key="item.id" :to="`/sre/shoots/${item.id}`" :title="item.resource.provenance.name" :subtitle="`${item.resource.provenance.namespace || 'unknown project'} · ${item.resource.provenance.installationRole || 'unknown role'} · ${item.resource.state || 'unknown state'}`"><template #prepend><v-icon>mdi-sprout-outline</v-icon></template></v-list-item></v-list><v-card-text v-else-if="!loading">No Shoot evidence is available.</v-card-text><v-skeleton-loader v-else type="list-item-three-line@3"/></v-card></AppPage></template>
<script setup lang="ts">
import { onMounted,ref } from 'vue'
import { sreApi,type SreResourceDTO } from '@/api/sre'
import AppPage from '@/components/app/AppPage.vue'
const items=ref<SreResourceDTO[]>([]),loading=ref(true),error=ref<string|null>(null)
onMounted(async()=>{try{items.value=(await sreApi.shoots()).data}catch(cause){error.value=cause instanceof Error?cause.message:'Shoots unavailable'}finally{loading.value=false}})
</script>
