<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/prefixes" class="text-primary text-decoration-none">IP Prefixes</router-link>
      <span>›</span><span>Create IP Prefix</span>
    </div>
    <h1 class="text-h5 font-weight-bold mb-1">Create an IP Prefix</h1>
    <p class="text-medium-emphasis mb-6">Allocate a new IP prefix from IPAM.</p>

    <div class="d-flex ga-6 align-start">
      <div style="flex:1;min-width:0">

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-label-outline</v-icon>Name</v-card-title>
          <v-card-text>
            <v-text-field v-model="form.name" label="Prefix name" variant="outlined" density="compact"
              placeholder="e.g. my-subnet" :error-messages="errors.name" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-ip-network</v-icon>IP Family</v-card-title>
          <v-card-text>
            <v-radio-group v-model="form.ipFamily" inline>
              <v-radio label="IPv4" value="IPv4" color="primary" />
              <v-radio label="IPv6" value="IPv6" color="primary" />
            </v-radio-group>
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-format-list-bulleted</v-icon>Allocation</v-card-title>
          <v-card-text>
            <v-radio-group v-model="form.allocationMode" inline class="mb-3">
              <v-radio label="Specific CIDR" value="cidr" color="primary" />
              <v-radio label="Request prefix length" value="length" color="primary" />
            </v-radio-group>
            <v-text-field v-if="form.allocationMode === 'cidr'" v-model="form.prefix"
              label="CIDR" variant="outlined" density="compact" placeholder="e.g. 10.0.0.0/24"
              :error-messages="errors.prefix" />
            <v-text-field v-else v-model.number="form.prefixLength"
              label="Prefix length" type="number" variant="outlined" density="compact"
              placeholder="e.g. 24" :error-messages="errors.prefixLength" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-sitemap</v-icon>Parent Prefix (optional)</v-card-title>
          <v-card-text>
            <v-select v-model="form.parentRef" :items="prefixNames" label="Parent prefix" variant="outlined"
              density="compact" clearable no-data-text="No prefixes found" />
          </v-card-text>
        </v-card>

        <v-alert v-if="submitError" type="error" class="mb-4" closable @click:close="submitError = ''">{{ submitError }}</v-alert>
        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" size="large" @click="submit">Create IP Prefix</v-btn>
          <v-btn variant="outlined" size="large" :to="{ path: '/prefixes' }">Cancel</v-btn>
        </div>
      </div>

      <v-card variant="outlined" rounded="lg" style="width:260px;position:sticky;top:80px;flex-shrink:0">
        <v-card-title class="text-subtitle-2 font-weight-bold pa-4 pb-2">Summary</v-card-title>
        <v-divider />
        <v-list density="compact" class="pa-2">
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAME</span></template><template #subtitle>{{ form.name || '—' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">IP FAMILY</span></template><template #subtitle>{{ form.ipFamily }}</template></v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">ALLOCATION</span></template>
            <template #subtitle>{{ form.allocationMode === 'cidr' ? (form.prefix || '—') : (form.prefixLength ? `/${form.prefixLength}` : '—') }}</template>
          </v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">PARENT</span></template><template #subtitle>{{ form.parentRef || 'None' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template><template #subtitle>{{ nsStore.active }}</template></v-list-item>
        </v-list>
        <v-card-actions class="pa-4 pt-2">
          <v-btn color="primary" block :loading="submitting" @click="submit">Create IP Prefix</v-btn>
        </v-card-actions>
      </v-card>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Prefix } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const router  = useRouter()
const nsStore = useNamespaceStore()
const existingPrefixes = ref<Prefix[]>([])
const submitting  = ref(false)
const submitError = ref('')
const errors      = ref<{ name?: string; prefix?: string; prefixLength?: string }>({})

const form = ref({
  name: '', ipFamily: 'IPv4', allocationMode: 'cidr',
  prefix: '', prefixLength: 24, parentRef: '',
})

const prefixNames = computed(() => existingPrefixes.value.map(p => p.name))

async function submit() {
  errors.value = {}
  if (!form.value.name) { errors.value.name = 'Required'; return }
  if (form.value.allocationMode === 'cidr' && !form.value.prefix) {
    errors.value.prefix = 'Required'; return
  }
  if (form.value.allocationMode === 'length' && (!form.value.prefixLength || form.value.prefixLength < 1)) {
    errors.value.prefixLength = 'Must be at least 1'; return
  }

  submitting.value = true; submitError.value = ''
  try {
    const body: { name: string; ipFamily: string; prefix?: string; prefixLength?: number; parentRef?: string } = {
      name: form.value.name,
      ipFamily: form.value.ipFamily,
    }
    if (form.value.allocationMode === 'cidr') body.prefix = form.value.prefix
    else body.prefixLength = form.value.prefixLength
    if (form.value.parentRef) body.parentRef = form.value.parentRef
    await api.prefixes.create(nsStore.active, body)
    router.push('/prefixes')
  } catch (e: unknown) {
    submitError.value = e instanceof Error ? e.message : String(e)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  existingPrefixes.value = await api.prefixes.list(nsStore.active)
})
</script>
