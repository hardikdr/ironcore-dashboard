<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/virtualips" class="text-primary text-decoration-none">Virtual IPs</router-link>
      <span>›</span><span>Create Virtual IP</span>
    </div>
    <h1 class="text-h5 font-weight-bold mb-1">Create a Virtual IP</h1>
    <p class="text-medium-emphasis mb-6">Allocate a new public IP address.</p>

    <div class="d-flex ga-6 align-start">
      <div style="flex:1;min-width:0">
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-label-outline</v-icon>Name
          </v-card-title>
          <v-card-text>
            <v-text-field v-model="form.name" label="Virtual IP name" variant="outlined" density="compact"
              placeholder="e.g. my-public-ip" :error-messages="errors.name" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-earth</v-icon>Type
          </v-card-title>
          <v-card-text>
            <v-radio-group v-model="form.type" inline>
              <v-radio label="Public" value="Public" color="primary" />
            </v-radio-group>
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-ip-network</v-icon>IP Family
          </v-card-title>
          <v-card-text>
            <v-radio-group v-model="form.ipFamily" inline>
              <v-radio label="IPv4" value="IPv4" color="primary" />
              <v-radio label="IPv6" value="IPv6" color="primary" />
            </v-radio-group>
          </v-card-text>
        </v-card>

        <v-alert v-if="submitError" type="error" class="mb-4" closable @click:close="submitError = ''">{{ submitError }}</v-alert>
        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" :disabled="submitting" size="large" @click="submit">Create Virtual IP</v-btn>
          <v-btn variant="outlined" size="large" :to="{ path: '/virtualips' }">Cancel</v-btn>
        </div>
      </div>

      <v-card variant="outlined" rounded="lg" style="width:260px;position:sticky;top:80px;flex-shrink:0">
        <v-card-title class="text-subtitle-2 font-weight-bold pa-4 pb-2">Summary</v-card-title>
        <v-divider />
        <v-list density="compact" class="pa-2">
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAME</span></template><template #subtitle>{{ form.name || '—' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">TYPE</span></template><template #subtitle>{{ form.type }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">IP FAMILY</span></template><template #subtitle>{{ form.ipFamily }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template><template #subtitle>{{ nsStore.active }}</template></v-list-item>
        </v-list>
        <v-card-actions class="pa-4 pt-2">
          <v-btn color="primary" block :loading="submitting" :disabled="submitting" @click="submit">Create Virtual IP</v-btn>
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
const form = ref({ name: '', type: 'Public', ipFamily: 'IPv4' })

async function submit() {
  errors.value = {}
  if (!form.value.name) { errors.value.name = 'Required'; return }
  submitting.value = true; submitError.value = ''
  try {
    await api.virtualIPs.create(nsStore.active, {
      name: form.value.name, type: form.value.type, ipFamily: form.value.ipFamily,
    })
    router.push('/virtualips')
  } catch (e: unknown) {
    submitError.value = e instanceof Error ? e.message : String(e)
  } finally {
    submitting.value = false
  }
}
</script>
