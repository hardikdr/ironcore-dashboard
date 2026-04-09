# Task 08 — Volumes, Networking, and IPAM Views

## Prerequisite

Task 05 (frontend scaffold) must be complete. Task 04 (backend APIs) should also be complete so the views can fetch real data.

Verify:
```bash
cd frontend && npm run dev  # starts without errors
```

## Your job

Implement all remaining resource list views: Volumes, Networks + NetworkInterfaces, VirtualIPs, LoadBalancers, and IP Prefixes.

Each view follows the same pattern as MachinesView:
1. Page header (title + namespace)
2. Stat cards (optional, only where meaningful)
3. `v-data-table` with `StatusBadge` for state fields
4. Reload when namespace changes (`watch(() => nsStore.active, load)`)

## Files to create

| File | Purpose |
|------|---------|
| `frontend/src/stores/volumes.ts` | Volumes Pinia store |
| `frontend/src/views/VolumesView.vue` | Volumes list (replace stub) |
| `frontend/src/views/NetworksView.vue` | Networks + NetworkInterfaces (replace stub) |
| `frontend/src/views/VirtualIPsView.vue` | Virtual IPs list (replace stub) |
| `frontend/src/views/LoadBalancersView.vue` | Load Balancers list (replace stub) |
| `frontend/src/views/IPPrefixesView.vue` | IP Prefixes list (replace stub) |

## Step-by-step

### Step 1 — Create `frontend/src/stores/volumes.ts`

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Volume } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useVolumesStore = defineStore('volumes', () => {
  const items   = ref<Volume[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true
    error.value   = null
    try {
      items.value = await api.volumes.list(ns)
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function deleteVolume(name: string) {
    await api.volumes.delete(useNamespaceStore().active, name)
    await load()
  }

  return { items, loading, error, load, deleteVolume }
})
```

### Step 2 — Replace `frontend/src/views/VolumesView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Volumes</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <v-btn color="primary" prepend-icon="mdi-plus" disabled>Create Volume</v-btn>
    </div>

    <!-- Stat cards -->
    <v-row dense class="mb-6">
      <v-col cols="auto">
        <v-card variant="outlined" rounded="lg">
          <v-card-text class="pa-4">
            <div class="text-h4 font-weight-bold" style="color:#16a34a">{{ availableCount }}</div>
            <div class="text-caption text-uppercase text-medium-emphasis mt-1">Available</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="auto">
        <v-card variant="outlined" rounded="lg">
          <v-card-text class="pa-4">
            <div class="text-h4 font-weight-bold" style="color:#d97706">{{ pendingCount }}</div>
            <div class="text-caption text-uppercase text-medium-emphasis mt-1">Pending</div>
          </v-card-text>
        </v-card>
      </v-col>
      <v-col cols="auto">
        <v-card variant="outlined" rounded="lg">
          <v-card-text class="pa-4">
            <div class="text-h4 font-weight-bold">{{ store.items.length }}</div>
            <div class="text-caption text-uppercase text-medium-emphasis mt-1">Total</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-database</v-icon>
        Volumes
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="store.items"
        :loading="store.loading"
        density="compact"
        hover
        :no-data-text="store.error ?? 'No volumes'"
      >
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.sizeBytes="{ item }">{{ formatBytes(item.sizeBytes) }}</template>
        <template #item.actions="{ item }">
          <v-btn
            size="x-small" variant="outlined" icon="mdi-delete" color="error"
            @click="store.deleteVolume(item.name)"
          />
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useVolumesStore } from '@/stores/volumes'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const store   = useVolumesStore()
const nsStore = useNamespaceStore()

const availableCount = computed(() => store.items.filter(v => v.state === 'Available').length)
const pendingCount   = computed(() => store.items.filter(v => v.state === 'Pending').length)

const headers = [
  { title: 'Name',         key: 'name'        },
  { title: 'State',        key: 'state'       },
  { title: 'Size',         key: 'sizeBytes'   },
  { title: 'Volume Class', key: 'volumeClass' },
  { title: 'Created',      key: 'createdAt'   },
  { title: '',             key: 'actions', sortable: false },
]

function formatBytes(b: number) {
  if (!b) return '—'
  const gb = b / 1024 / 1024 / 1024
  return gb >= 1000 ? `${(gb / 1024).toFixed(1)} TB` : `${Math.round(gb)} GB`
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
```

### Step 3 — Replace `frontend/src/views/NetworksView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">Networking</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <!-- Networks table -->
    <v-card variant="outlined" rounded="lg" class="mb-6">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-lan</v-icon>
        Networks
        <v-chip size="x-small" color="primary">{{ networks.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="netHeaders"
        :items="networks"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No networks"
      />
    </v-card>

    <!-- Network Interfaces table -->
    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-network-outline</v-icon>
        Network Interfaces
        <v-chip size="x-small" color="primary">{{ interfaces.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="nicHeaders"
        :items="interfaces"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No network interfaces"
      >
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.ips="{ item }">
          <v-chip
            v-for="ip in item.ips"
            :key="ip"
            size="x-small"
            color="info"
            class="mr-1"
          >{{ ip }}</v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Network, type NetworkInterface } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const nsStore    = useNamespaceStore()
const networks   = ref<Network[]>([])
const interfaces = ref<NetworkInterface[]>([])
const loading    = ref(false)

const netHeaders = [
  { title: 'Name',    key: 'name'      },
  { title: 'Created', key: 'createdAt' },
]

const nicHeaders = [
  { title: 'Name',    key: 'name'      },
  { title: 'State',   key: 'state'     },
  { title: 'IPs',     key: 'ips'       },
  { title: 'Network', key: 'network'   },
  { title: 'Machine', key: 'machine'   },
]

async function load() {
  loading.value = true
  const ns = nsStore.active
  ;[networks.value, interfaces.value] = await Promise.all([
    api.networks.list(ns),
    api.networks.listInterfaces(ns),
  ])
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
```

### Step 4 — Replace `frontend/src/views/VirtualIPsView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">Virtual IPs</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-earth</v-icon>
        Virtual IPs
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No virtual IPs"
      >
        <template #item.ip="{ item }">
          <v-chip v-if="item.ip" size="x-small" color="deep-purple" class="font-weight-bold">
            {{ item.ip }}
          </v-chip>
          <span v-else class="text-medium-emphasis">Pending…</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type VirtualIP } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const nsStore = useNamespaceStore()
const items   = ref<VirtualIP[]>([])
const loading = ref(false)

const headers = [
  { title: 'Name',      key: 'name'      },
  { title: 'IP',        key: 'ip'        },
  { title: 'Type',      key: 'type'      },
  { title: 'IP Family', key: 'ipFamily'  },
  { title: 'Created',   key: 'createdAt' },
]

async function load() {
  loading.value = true
  items.value   = await api.virtualIPs.list(nsStore.active)
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
```

### Step 5 — Replace `frontend/src/views/LoadBalancersView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">Load Balancers</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-scale-balance</v-icon>
        Load Balancers
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No load balancers"
      >
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">
            {{ ip }}
          </v-chip>
          <span v-if="!item.ips.length" class="text-medium-emphasis">—</span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type LoadBalancer } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const nsStore = useNamespaceStore()
const items   = ref<LoadBalancer[]>([])
const loading = ref(false)

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Type',    key: 'type'      },
  { title: 'IPs',     key: 'ips'       },
  { title: 'Created', key: 'createdAt' },
]

async function load() {
  loading.value = true
  items.value   = await api.loadBalancers.list(nsStore.active)
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
```

### Step 6 — Replace `frontend/src/views/IPPrefixesView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <h1 class="text-h5 font-weight-bold mb-2">IP Prefixes</h1>
    <div class="text-caption text-medium-emphasis mb-6">{{ nsStore.active }}</div>

    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-ip-network</v-icon>
        IP Prefixes
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </v-card-title>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
        no-data-text="No IP prefixes"
      >
        <template #item.prefix="{ item }">
          <code class="text-primary">{{ item.prefix || '—' }}</code>
        </template>
        <template #item.phase="{ item }">
          <StatusBadge :state="item.phase" />
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { api, type Prefix } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'

const nsStore = useNamespaceStore()
const items   = ref<Prefix[]>([])
const loading = ref(false)

const headers = [
  { title: 'Name',    key: 'name'      },
  { title: 'Prefix',  key: 'prefix'    },
  { title: 'Phase',   key: 'phase'     },
  { title: 'Created', key: 'createdAt' },
]

async function load() {
  loading.value = true
  items.value   = await api.prefixes.list(nsStore.active)
  loading.value = false
}

onMounted(load)
watch(() => nsStore.active, load)
</script>
```

### Step 7 — Verify in browser

With the backend running:

```bash
cd frontend && npm run dev
```

Navigate to each tab and verify:
- http://localhost:5173/volumes — Volumes table loads
- http://localhost:5173/networks — Networks + Network Interfaces tables load
- http://localhost:5173/virtualips — Virtual IPs table loads
- http://localhost:5173/loadbalancers — Load Balancers table loads
- http://localhost:5173/prefixes — IP Prefixes table loads

All should show empty tables with "No X found" text (not error messages) when cluster is empty.

### Step 8 — Commit

```bash
git add frontend/src/stores/volumes.ts frontend/src/views/
git commit -m "feat: Volumes, Networks, VIPs, Load Balancers, IPAM views"
```

## Done criteria

- All 5 new views render without errors
- Each reloads data when namespace changes
- StatusBadge shows for state/phase fields
- `npm run build` passes TypeScript type checking

## Note

The "Create Volume" button is disabled in this task — standalone volume creation UI is v2. Volumes are created inline when launching a machine (Task 07).