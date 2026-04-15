<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/volumes" class="text-primary text-decoration-none">Volumes</router-link>
      <span>›</span><span>Create Volume</span>
    </div>
    <h1 class="text-h5 font-weight-bold mb-1">Create a Volume</h1>
    <p class="text-medium-emphasis mb-6">Provision a new IronCore storage volume.</p>

    <div class="d-flex ga-6 align-start">
      <div style="flex:1;min-width:0">

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-label-outline</v-icon>Name
          </v-card-title>
          <v-card-text>
            <v-text-field v-model="form.name" label="Volume name" variant="outlined" density="compact"
              placeholder="e.g. my-data-volume" :error-messages="errors.name" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-harddisk</v-icon>Volume Class
          </v-card-title>
          <v-card-text class="pa-0">
            <v-table density="compact">
              <thead><tr><th style="width:40px"></th><th>Class</th><th>Capacity</th></tr></thead>
              <tbody>
                <tr v-for="vc in volumeClasses" :key="vc.name" style="cursor:pointer"
                  :style="form.volumeClass === vc.name ? 'background-color:rgba(26,95,168,0.08)' : ''"
                  @click="form.volumeClass = vc.name">
                  <td><v-radio :model-value="form.volumeClass" :value="vc.name" hide-details color="primary" /></td>
                  <td><strong>{{ vc.name }}</strong></td>
                  <td>{{ vc.storage }}</td>
                </tr>
                <tr v-if="!volumeClasses.length">
                  <td colspan="3" class="text-center text-medium-emphasis pa-4">No volume classes found</td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-database-outline</v-icon>Size
          </v-card-title>
          <v-card-text>
            <v-text-field v-model.number="form.sizeGiB" label="Size (GiB)" type="number" variant="outlined"
              density="compact" min="1" :error-messages="errors.sizeGiB" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-lock-outline</v-icon>Encryption
          </v-card-title>
          <v-card-text>
            <v-switch v-model="form.encryptionEnabled" label="Enable encryption" color="primary" hide-details class="mb-3" />
            <v-text-field v-if="form.encryptionEnabled" v-model="form.encryptionSecret"
              label="Secret name" variant="outlined" density="compact"
              placeholder="e.g. my-encryption-secret" />
          </v-card-text>
        </v-card>

        <v-alert v-if="submitError" type="error" class="mb-4" closable @click:close="submitError = ''">{{ submitError }}</v-alert>
        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" size="large" @click="submit">Create Volume</v-btn>
          <v-btn variant="outlined" size="large" :to="{ path: '/volumes' }">Cancel</v-btn>
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
            <template #title><span class="text-caption text-medium-emphasis">CLASS</span></template>
            <template #subtitle>{{ form.volumeClass || '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">SIZE</span></template>
            <template #subtitle>{{ form.sizeGiB ? `${form.sizeGiB} GiB` : '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">ENCRYPTION</span></template>
            <template #subtitle>{{ form.encryptionEnabled ? (form.encryptionSecret || 'enabled') : 'None' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template>
            <template #subtitle>{{ nsStore.active }}</template>
          </v-list-item>
        </v-list>
        <v-card-actions class="pa-4 pt-2">
          <v-btn color="primary" block :loading="submitting" @click="submit">Create Volume</v-btn>
        </v-card-actions>
      </v-card>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type VolumeClass } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const router  = useRouter()
const nsStore = useNamespaceStore()
const volumeClasses = ref<VolumeClass[]>([])
const submitting    = ref(false)
const submitError   = ref('')
const errors        = ref<{ name?: string; sizeGiB?: string }>({})

const form = ref({
  name: '', volumeClass: '', sizeGiB: 100,
  encryptionEnabled: false, encryptionSecret: '',
})

async function submit() {
  errors.value = {}
  if (!form.value.name)        { errors.value.name    = 'Required'; return }
  if (!form.value.sizeGiB || form.value.sizeGiB < 1) { errors.value.sizeGiB = 'Must be at least 1 GiB'; return }
  if (!form.value.volumeClass) { submitError.value = 'Select a volume class'; return }

  submitting.value = true; submitError.value = ''
  try {
    await api.volumes.create(nsStore.active, {
      name: form.value.name,
      volumeClass: form.value.volumeClass,
      sizeGiB: form.value.sizeGiB,
      encryptionSecret: form.value.encryptionEnabled ? form.value.encryptionSecret : undefined,
    })
    router.push('/volumes')
  } catch (e: unknown) {
    submitError.value = e instanceof Error ? e.message : String(e)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  const classes = await api.volumeClasses.list()
  volumeClasses.value = classes
  if (classes.length) form.value.volumeClass = classes[0].name
})
</script>
