# Task 10 — Machine Detail View + Final Polish

## Prerequisites

All Tasks 01–09 must be complete. The single binary must work:

```bash
make build
./bin/ironcore-dashboard --addr :8080 --kubeconfig ~/.kube/config
# Open http://localhost:8080 — full dashboard loads
```

## Your job

1. Implement the Machine Detail view (`/machines/:name`)
2. Run through the full smoke test checklist
3. Fix any issues found during testing

## Files to create / modify

| Action | File |
|--------|------|
| Replace stub | `frontend/src/views/MachineDetailView.vue` |
| Fix as needed | Any file with issues found during smoke test |

## Step-by-step

### Step 1 — Replace `frontend/src/views/MachineDetailView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <!-- Breadcrumb -->
    <div class="d-flex align-center ga-2 mb-4 text-caption text-medium-emphasis">
      <router-link to="/machines" class="text-primary text-decoration-none">Machines</router-link>
      <span>›</span>
      <span>{{ route.params.name }}</span>
    </div>

    <!-- Loading state -->
    <v-progress-circular
      v-if="!machine && !loadError"
      indeterminate
      color="primary"
      class="mt-10 d-block mx-auto"
    />

    <!-- Error state -->
    <v-alert v-if="loadError" type="error" class="mb-4">
      {{ loadError }}
    </v-alert>

    <!-- Loaded -->
    <template v-if="machine">
      <!-- Header -->
      <div class="d-flex align-center justify-space-between mb-6">
        <div>
          <h1 class="text-h5 font-weight-bold">{{ machine.name }}</h1>
          <StatusBadge :state="machine.state" class="mt-1" />
        </div>
        <div class="d-flex ga-2">
          <v-btn
            variant="outlined"
            :prepend-icon="machine.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
            :color="machine.power === 'On' ? 'warning' : 'success'"
            :loading="powerLoading"
            @click="togglePower"
          >
            {{ machine.power === 'On' ? 'Power Off' : 'Power On' }}
          </v-btn>
          <v-btn
            variant="outlined"
            color="error"
            prepend-icon="mdi-delete"
            @click="deleteDialog = true"
          >
            Delete
          </v-btn>
        </div>
      </div>

      <!-- Detail cards -->
      <v-row>
        <!-- Details -->
        <v-col cols="12" md="6">
          <v-card variant="outlined" rounded="lg">
            <v-card-title class="pa-4 text-subtitle-1 font-weight-bold">
              <v-icon start color="primary">mdi-information-outline</v-icon>
              Details
            </v-card-title>
            <v-divider />
            <v-list density="compact">
              <v-list-item title="Machine type">
                <template #subtitle>
                  <strong>{{ machine.machineClass }}</strong>
                </template>
              </v-list-item>
              <v-list-item title="OS Image" :subtitle="machine.image" />
              <v-list-item title="Power state" :subtitle="machine.power" />
              <v-list-item title="Namespace" :subtitle="machine.namespace" />
              <v-list-item title="Created" :subtitle="machine.createdAt" />
            </v-list>
          </v-card>
        </v-col>

        <!-- Network + Volumes -->
        <v-col cols="12" md="6">
          <v-card variant="outlined" rounded="lg" class="mb-4">
            <v-card-title class="pa-4 text-subtitle-1 font-weight-bold">
              <v-icon start color="primary">mdi-lan</v-icon>
              Network
            </v-card-title>
            <v-divider />
            <v-card-text>
              <div v-if="machine.ips.length" class="d-flex flex-wrap ga-1">
                <v-chip
                  v-for="ip in machine.ips"
                  :key="ip"
                  color="info"
                  size="small"
                >{{ ip }}</v-chip>
              </div>
              <span v-else class="text-medium-emphasis">No IPs assigned yet</span>
            </v-card-text>
          </v-card>

          <v-card variant="outlined" rounded="lg">
            <v-card-title class="pa-4 text-subtitle-1 font-weight-bold">
              <v-icon start color="primary">mdi-database</v-icon>
              Volumes
            </v-card-title>
            <v-divider />
            <v-card-text>
              <div v-if="machine.volumes.length" class="d-flex flex-wrap ga-1">
                <v-chip
                  v-for="v in machine.volumes"
                  :key="v"
                  variant="outlined"
                  size="small"
                >{{ v }}</v-chip>
              </div>
              <span v-else class="text-medium-emphasis">No volumes attached</span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </template>

    <!-- Delete confirm dialog -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title>Delete machine?</v-card-title>
        <v-card-text>
          Machine <strong>{{ machine?.name }}</strong> will be permanently deleted.
          This cannot be undone.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deleteLoading" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, type Machine } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const route   = useRoute()
const router  = useRouter()
const nsStore = useNamespaceStore()

const machine      = ref<Machine | null>(null)
const loadError    = ref<string>('')
const powerLoading = ref(false)
const deleteLoading = ref(false)
const deleteDialog  = ref(false)

onMounted(async () => {
  try {
    machine.value = await api.machines.get(nsStore.active, route.params.name as string)
  } catch (e: any) {
    loadError.value = e.message
  }
})

async function togglePower() {
  if (!machine.value) return
  const next = machine.value.power === 'On' ? 'Off' : 'On'
  powerLoading.value = true
  try {
    machine.value = await api.machines.power(nsStore.active, machine.value.name, next)
  } finally {
    powerLoading.value = false
  }
}

async function doDelete() {
  if (!machine.value) return
  deleteLoading.value = true
  try {
    await api.machines.delete(nsStore.active, machine.value.name)
    router.push('/machines')
  } finally {
    deleteLoading.value = false
  }
}
</script>
```

### Step 2 — Build and run full smoke test

```bash
make build
./bin/ironcore-dashboard --addr :8080 --kubeconfig ~/.kube/config
```

Go through this checklist — fix any issues you find before committing:

**Navigation**
- [ ] http://localhost:8080 loads the dashboard (blue topbar, sidebar, machines view)
- [ ] Refreshing the browser at http://localhost:8080/machines works (not 404)
- [ ] Clicking top nav tabs switches sections

**Namespace switcher**
- [ ] Namespace dropdown in topbar shows namespaces from the cluster
- [ ] Switching namespace reloads data in the current view

**Machines**
- [ ] Machines list loads (empty `[]` if no machines, not an error)
- [ ] Click "Create Machine" → wizard at `/machines/new` loads
- [ ] Wizard shows machine classes from API (or "No machine classes found" gracefully)
- [ ] Wizard shows networks from API
- [ ] Create a machine → redirected to `/machines` → new machine appears with Pending badge
- [ ] Power on/off button works (machine power state updates)
- [ ] Delete opens confirm dialog → deletes and removes from list
- [ ] Click machine name → detail view loads at `/machines/:name`
- [ ] Power toggle works from detail view
- [ ] Delete from detail view redirects back to `/machines`

**Other views**
- [ ] `/volumes` loads without errors
- [ ] `/networks` loads without errors (two tables: Networks + Network Interfaces)
- [ ] `/virtualips` loads without errors
- [ ] `/loadbalancers` loads without errors
- [ ] `/prefixes` loads without errors

**API**
- [ ] `curl http://localhost:8080/api/v1/namespaces` returns namespace list
- [ ] `curl http://localhost:8080/api/v1/namespaces/default/machines` returns `[]`
- [ ] `curl http://localhost:8080/healthz` returns `ok`

### Step 3 — Fix any issues found

Common issues to check:
- **SPA refresh 404**: The `/*` route in `server.go` must return `index.html` for unknown paths
- **Namespace switcher empty**: Verify `GET /api/v1/namespaces` works and returns data
- **TypeScript errors**: Run `cd frontend && npm run build` to catch all type errors

### Step 4 — Final commit

```bash
git add .
git commit -m "feat: machine detail view, smoke test passing, v1 complete"
```

## Done criteria

All checklist items above pass. The project delivers:

- Single Go binary serving full dashboard + API
- Machines: list, create, delete, power on/off, detail view
- Volumes: list
- Networks + NetworkInterfaces: list
- VirtualIPs: list
- LoadBalancers: list
- IP Prefixes: list
- Namespace switcher synced with real cluster namespaces
- Light theme, IronCore blue, IronCore logo in topbar