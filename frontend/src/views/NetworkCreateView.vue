<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/networks" class="text-primary text-decoration-none">Networking</router-link>
      <span>›</span><span>Create Network</span>
    </div>
    <h1 class="text-h5 font-weight-bold mb-1">Create a Network</h1>
    <p class="text-medium-emphasis mb-6">Provision a new IronCore network.</p>

    <div class="d-flex ga-6 align-start">
      <div style="flex:1;min-width:0">
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-label-outline</v-icon>Name
          </v-card-title>
          <v-card-text>
            <v-text-field v-model="form.name" label="Network name" variant="outlined" density="compact"
              placeholder="e.g. my-network" :error-messages="errors.name" />
          </v-card-text>
        </v-card>

        <v-alert v-if="submitError" type="error" class="mb-4" closable @click:close="submitError = ''">{{ submitError }}</v-alert>
        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" size="large" @click="submit">Create Network</v-btn>
          <v-btn variant="outlined" size="large" :to="{ path: '/networks' }">Cancel</v-btn>
        </div>
      </div>

      <v-card variant="outlined" rounded="lg" style="width:260px;position:sticky;top:80px;flex-shrink:0">
        <v-card-title class="text-subtitle-2 font-weight-bold pa-4 pb-2">Summary</v-card-title>
        <v-divider />
        <v-list density="compact" class="pa-2">
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">NAME</span></template>
            <template #subtitle>{{ form.name || '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template>
            <template #subtitle>{{ nsStore.active }}</template>
          </v-list-item>
        </v-list>
        <v-card-actions class="pa-4 pt-2">
          <v-btn color="primary" block :loading="submitting" @click="submit">Create Network</v-btn>
        </v-card-actions>
      </v-card>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const router  = useRouter()
const nsStore = useNamespaceStore()
const submitting  = ref(false)
const submitError = ref('')
const errors      = ref<{ name?: string }>({})
const form        = ref({ name: '' })

async function submit() {
  errors.value = {}
  if (!form.value.name) { errors.value.name = 'Required'; return }
  submitting.value = true; submitError.value = ''
  try {
    await api.networks.create(nsStore.active, { name: form.value.name })
    router.push('/networks')
  } catch (e: unknown) {
    submitError.value = e instanceof Error ? e.message : String(e)
  } finally {
    submitting.value = false
  }
}
</script>
