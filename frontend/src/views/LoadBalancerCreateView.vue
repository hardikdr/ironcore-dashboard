<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/loadbalancers" class="text-primary text-decoration-none">Load Balancers</router-link>
      <span>›</span><span>Create Load Balancer</span>
    </div>
    <h1 class="text-h5 font-weight-bold mb-1">Create a Load Balancer</h1>
    <p class="text-medium-emphasis mb-6">Configure a new IronCore load balancer.</p>

    <div class="d-flex ga-6 align-start">
      <div style="flex:1;min-width:0">

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-label-outline</v-icon>Name</v-card-title>
          <v-card-text>
            <v-text-field v-model="form.name" label="Load balancer name" variant="outlined" density="compact" placeholder="e.g. my-lb" :error-messages="errors.name" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-swap-horizontal</v-icon>Type & IP Family</v-card-title>
          <v-card-text>
            <v-label class="text-caption mb-1">Type</v-label>
            <v-radio-group v-model="form.type" inline class="mb-3">
              <v-radio label="Public" value="Public" color="primary" />
              <v-radio label="Internal" value="Internal" color="primary" />
            </v-radio-group>
            <v-label class="text-caption mb-1">IP Families</v-label>
            <div class="d-flex ga-4">
              <v-checkbox v-model="form.ipv4" label="IPv4" color="primary" hide-details />
              <v-checkbox v-model="form.ipv6" label="IPv6" color="primary" hide-details />
            </div>
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-lan</v-icon>Network</v-card-title>
          <v-card-text>
            <v-select v-model="form.networkRef" :items="networkNames" label="Network" variant="outlined"
              density="compact" no-data-text="No networks found — create one first" :error-messages="errors.networkRef" />
          </v-card-text>
        </v-card>

        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2"><v-icon start color="primary">mdi-connection</v-icon>Ports</v-card-title>
          <v-card-text>
            <div v-for="(port, i) in form.ports" :key="i" class="d-flex ga-3 align-center mb-3">
              <v-select v-model="port.protocol" :items="['TCP','UDP']" label="Protocol" variant="outlined" density="compact" hide-details style="max-width:110px" />
              <v-text-field v-model.number="port.port" label="Port" type="number" variant="outlined" density="compact" hide-details style="max-width:100px" />
              <v-text-field v-model.number="port.endPort" label="End Port" type="number" variant="outlined" density="compact" hide-details style="max-width:110px" placeholder="optional" />
              <v-btn icon="mdi-delete" size="small" variant="text" color="error" @click="form.ports.splice(i,1)" />
            </div>
            <v-btn variant="outlined" color="primary" prepend-icon="mdi-plus" size="small" @click="addPort">Add Port</v-btn>
          </v-card-text>
        </v-card>

        <v-alert v-if="submitError" type="error" class="mb-4" closable @click:close="submitError = ''">{{ submitError }}</v-alert>
        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" size="large" @click="submit">Create Load Balancer</v-btn>
          <v-btn variant="outlined" size="large" :to="{ path: '/loadbalancers' }">Cancel</v-btn>
        </div>
      </div>

      <v-card variant="outlined" rounded="lg" style="width:260px;position:sticky;top:80px;flex-shrink:0">
        <v-card-title class="text-subtitle-2 font-weight-bold pa-4 pb-2">Summary</v-card-title>
        <v-divider />
        <v-list density="compact" class="pa-2">
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAME</span></template><template #subtitle>{{ form.name || '—' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">TYPE</span></template><template #subtitle>{{ form.type }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">IP FAMILIES</span></template><template #subtitle>{{ ipFamilies.join(', ') || '—' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NETWORK</span></template><template #subtitle>{{ form.networkRef || '—' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">PORTS</span></template><template #subtitle>{{ form.ports.length ? `${form.ports.length} port(s)` : '—' }}</template></v-list-item>
          <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template><template #subtitle>{{ nsStore.active }}</template></v-list-item>
        </v-list>
        <v-card-actions class="pa-4 pt-2">
          <v-btn color="primary" block :loading="submitting" @click="submit">Create Load Balancer</v-btn>
        </v-card-actions>
      </v-card>
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type Network } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const router  = useRouter()
const nsStore = useNamespaceStore()
const networks    = ref<Network[]>([])
const submitting  = ref(false)
const submitError = ref('')
const errors      = ref<{ name?: string; networkRef?: string }>({})

const form = ref({
  name: '', type: 'Public', ipv4: true, ipv6: false, networkRef: '',
  ports: [] as { protocol: string; port: number; endPort: number }[],
})

const networkNames  = computed(() => networks.value.map(n => n.name))
const ipFamilies    = computed(() => [
  ...(form.value.ipv4 ? ['IPv4'] : []),
  ...(form.value.ipv6 ? ['IPv6'] : []),
])

function addPort() { form.value.ports.push({ protocol: 'TCP', port: 80, endPort: 0 }) }

async function submit() {
  errors.value = {}
  if (!form.value.name)       { errors.value.name = 'Required'; return }
  if (!form.value.networkRef) { errors.value.networkRef = 'Required'; return }
  if (!ipFamilies.value.length) { submitError.value = 'Select at least one IP family'; return }

  submitting.value = true; submitError.value = ''
  try {
    await api.loadBalancers.create(nsStore.active, {
      name: form.value.name, type: form.value.type,
      ipFamilies: ipFamilies.value, networkRef: form.value.networkRef,
      ports: form.value.ports.map(p => ({
        protocol: p.protocol, port: p.port,
        ...(p.endPort > 0 ? { endPort: p.endPort } : {}),
      })),
    })
    router.push('/loadbalancers')
  } catch (e: unknown) {
    submitError.value = e instanceof Error ? e.message : String(e)
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  networks.value = await api.networks.list(nsStore.active)
  if (networks.value.length) form.value.networkRef = networks.value[0].name
})
</script>
