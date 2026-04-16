<template>
  <v-container fluid class="pa-6">
    <!-- Breadcrumb -->
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/machines" class="text-primary text-decoration-none">Machines</router-link>
      <span>›</span>
      <span>Launch a Machine</span>
    </div>

    <h1 class="text-h5 font-weight-bold mb-1">Launch a Machine</h1>
    <p class="text-medium-emphasis mb-6">Configure and launch a new IronCore virtual machine.</p>

    <div class="d-flex ga-6 align-start">

      <!-- ── Left: accordion form ── -->
      <div style="flex:1;min-width:0">

        <!-- Name -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-label-outline</v-icon>
            Name
          </v-card-title>
          <v-card-text>
            <v-text-field
              v-model="form.name"
              label="Machine name"
              variant="outlined"
              density="compact"
              placeholder="e.g. web-server-01"
              :rules="[v => !!v || 'Required']"
              :error-messages="errors.name"
            />
          </v-card-text>
        </v-card>

        <!-- Machine type -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-cpu-64-bit</v-icon>
            Machine Type
          </v-card-title>
          <v-card-text class="pa-0">
            <v-table density="compact">
              <thead>
                <tr>
                  <th style="width:40px"></th>
                  <th>Type</th>
                  <th>vCPU</th>
                  <th>Memory</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="mc in machineClasses"
                  :key="mc.name"
                  style="cursor:pointer"
                  :style="form.machineClass === mc.name ? 'background-color: rgba(26,95,168,0.08)' : ''"
                  @click="form.machineClass = mc.name"
                >
                  <td>
                    <v-radio
                      :model-value="form.machineClass"
                      :value="mc.name"
                      hide-details
                      color="primary"
                    />
                  </td>
                  <td><strong>{{ mc.name }}</strong></td>
                  <td>{{ mc.cpu }}</td>
                  <td>{{ mc.ram }}</td>
                </tr>
                <tr v-if="!machineClasses.length">
                  <td colspan="4" class="text-center text-medium-emphasis pa-4">
                    No machine classes found in cluster
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>

        <!-- OS Image -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-disc</v-icon>
            OS Image
          </v-card-title>
          <v-card-text>
            <v-select
              v-model="form.image"
              :items="imageOptions"
              label="OS Image"
              variant="outlined"
              density="compact"
            />
          </v-card-text>
        </v-card>

        <!-- Network -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-lan</v-icon>
            Network
          </v-card-title>
          <v-card-text>
            <v-select
              v-model="form.networkName"
              :items="networkNames"
              label="Network"
              variant="outlined"
              density="compact"
              :no-data-text="'No networks found — create one first'"
              class="mb-3"
            />
            <v-text-field
              v-model="form.ip"
              label="IP Address"
              variant="outlined"
              density="compact"
              placeholder="e.g. 10.0.0.2"
              :rules="[v => !!v || 'Required']"
              :error-messages="errors.ip"
            />
          </v-card-text>
        </v-card>

        <!-- Storage -->
        <v-card variant="outlined" rounded="lg" class="mb-4">
          <v-card-title class="text-subtitle-1 font-weight-bold pa-4 pb-2">
            <v-icon start color="primary">mdi-database</v-icon>
            Storage
          </v-card-title>
          <v-card-text>
            <div
              v-for="(vol, i) in form.volumes"
              :key="i"
              class="d-flex ga-3 align-center mb-3"
            >
              <v-text-field
                v-model="vol.name"
                label="Volume name"
                variant="outlined"
                density="compact"
                hide-details
              />
              <v-text-field
                v-model.number="vol.sizeGiB"
                label="Size (GiB)"
                type="number"
                variant="outlined"
                density="compact"
                hide-details
                style="max-width:110px"
              />
              <v-select
                v-model="vol.volumeClass"
                :items="['standard', 'fast-ssd']"
                label="Class"
                variant="outlined"
                density="compact"
                hide-details
                style="max-width:130px"
              />
              <v-btn
                icon="mdi-delete"
                size="small"
                variant="text"
                color="error"
                :disabled="false"
                @click="form.volumes.splice(i, 1)"
              />
            </div>
            <v-btn
              variant="outlined"
              color="primary"
              prepend-icon="mdi-plus"
              size="small"
              @click="addVolume"
            >
              Add volume
            </v-btn>
          </v-card-text>
        </v-card>

        <!-- Error -->
        <v-alert v-if="submitError" type="error" class="mb-4" closable @click:close="submitError = ''">
          {{ submitError }}
        </v-alert>

        <!-- Actions -->
        <div class="d-flex ga-3">
          <v-btn color="primary" :loading="submitting" size="large" @click="submit">
            Launch Machine
          </v-btn>
          <v-btn variant="outlined" size="large" :to="{ path: '/machines' }">
            Cancel
          </v-btn>
        </div>
      </div>

      <!-- ── Right: Summary panel ── -->
      <v-card
        variant="outlined"
        rounded="lg"
        style="width:260px;position:sticky;top:80px;flex-shrink:0"
      >
        <v-card-title class="text-subtitle-2 font-weight-bold pa-4 pb-2">Summary</v-card-title>
        <v-divider />
        <v-list density="compact" class="pa-2">
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">NAME</span></template>
            <template #subtitle>{{ form.name || '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">MACHINE TYPE</span></template>
            <template #subtitle>{{ form.machineClass || '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">OS IMAGE</span></template>
            <template #subtitle>{{ form.image || '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">NETWORK</span></template>
            <template #subtitle>{{ form.networkName || '—' }}</template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">VOLUMES</span></template>
            <template #subtitle>
              <span v-if="form.volumes.length">
                {{ form.volumes.map(v => `${v.name} (${v.sizeGiB} GiB)`).join(', ') }}
              </span>
              <span v-else class="text-medium-emphasis">—</span>
            </template>
          </v-list-item>
          <v-list-item>
            <template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template>
            <template #subtitle>{{ nsStore.active }}</template>
          </v-list-item>
        </v-list>
        <v-card-actions class="pa-4 pt-2">
          <v-btn color="primary" block :loading="submitting" @click="submit">Launch Machine</v-btn>
        </v-card-actions>
      </v-card>

    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { api, type MachineClass, type Network } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const router  = useRouter()
const nsStore = useNamespaceStore()

const machineClasses = ref<MachineClass[]>([])
const networks       = ref<Network[]>([])
const submitting     = ref(false)
const submitError    = ref('')
const errors         = ref<{ name?: string; ip?: string }>({})

const imageOptions = [
  'ubuntu-22.04',
  'ubuntu-24.04',
  'debian-12',
  'almalinux-9',
]

const networkNames = computed(() => networks.value.map(n => n.name))

const form = ref({
  name:         '',
  machineClass: '',
  image:        'ubuntu-22.04',
  networkName:  '',
  ip:           '',
  volumes: [] as { name: string; sizeGiB: number; volumeClass: string }[],
})

function addVolume() {
  form.value.volumes.push({
    name:        `data-${form.value.volumes.length}`,
    sizeGiB:     100,
    volumeClass: 'standard',
  })
}

async function submit() {
  errors.value = {}
  if (!form.value.name) {
    errors.value.name = 'Machine name is required'
    return
  }
  if (!form.value.ip) {
    errors.value.ip = 'IP address is required'
    return
  }
  if (!form.value.machineClass || !form.value.networkName) {
    submitError.value = 'Machine type and network are required.'
    return
  }

  submitting.value = true
  submitError.value = ''

  try {
    await api.machines.create(nsStore.active, {
      name:         form.value.name,
      machineClass: form.value.machineClass,
      image:        form.value.image,
      networkName:  form.value.networkName,
      ip:           form.value.ip,
      power:        'On',
      volumes:      form.value.volumes.map(v => ({
        name:        v.name,
        sizeBytes:   v.sizeGiB * 1024 * 1024 * 1024,
        volumeClass: v.volumeClass,
      })),
    })
    router.push('/machines')
  } catch (e: any) {
    submitError.value = e.message
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  const [classes, nets] = await Promise.all([
    api.machineClasses.list(),
    api.networks.list(nsStore.active),
  ])
  machineClasses.value = classes
  networks.value       = nets

  if (classes.length)  form.value.machineClass = classes[0].name
  if (nets.length)     form.value.networkName  = nets[0].name
})
</script>
