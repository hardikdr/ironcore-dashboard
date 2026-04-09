# Task 06 — TopBar, Sidebar, NamespaceSwitcher, and MachinesView

## Prerequisite

Task 05 (frontend scaffold) must be complete. Verify:

```bash
cd frontend && npm run dev  # must start without errors
```

## Your job

Replace the placeholder layout components with the real ones, and implement the Machines list view with stats cards, searchable table, power toggle, and delete.

By the end:
- Blue topbar with IronCore Fe logo, top nav tabs (Machines/Volumes/Networking/IPAM), namespace switcher dropdown
- Sidebar grouped by Compute / Storage / Networking / IP Management
- Machines view with stat cards (Running/Pending/Shutdown/Total), searchable data-table, power toggle, delete with confirm dialog
- Switching namespace reloads the machines table

## IronCore logo

Copy the logo to the frontend public folder:

```bash
cp ../../../ironcore-logo.png frontend/public/ironcore-logo.png
```

If `ironcore-logo.png` doesn't exist in `dashboard-workspace/`, look for it in:
- `ironcore/docs/` or `ironcore/logo/`
- Download the IronCore Fe (iron hexagon) logo from the IronCore GitHub org if needed
- As a fallback, use `https://raw.githubusercontent.com/ironcore-dev/ironcore/main/docs/assets/logo.png`

The logo is referenced as `/ironcore-logo.png` in the `<v-img>` component.

## Files to create / modify

| Action | File |
|--------|------|
| Create | `frontend/src/components/StatusBadge.vue` |
| Create | `frontend/src/components/StatCard.vue` |
| Create | `frontend/src/components/NamespaceSwitcher.vue` |
| Create | `frontend/src/components/TopBar.vue` |
| Create | `frontend/src/components/Sidebar.vue` |
| Create | `frontend/src/stores/machines.ts` |
| Replace | `frontend/src/views/MachinesView.vue` (replace stub) |
| Modify | `frontend/src/layouts/DashboardLayout.vue` (use real TopBar + Sidebar) |

## Step-by-step

### Step 1 — Create `frontend/src/components/StatusBadge.vue`

```vue
<template>
  <v-chip :color="color" size="small" label>
    <v-icon start size="x-small">mdi-circle</v-icon>
    {{ label }}
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string }>()

const color = computed(() => ({
  Running:    'success',
  Available:  'success',
  Allocated:  'success',
  Pending:    'warning',
  Shutdown:   'default',
  Stopped:    'default',
  Error:      'error',
  Terminated: 'error',
  Terminating:'error',
}[props.state] ?? 'default'))

const label = computed(() => props.state || '—')
</script>
```

### Step 2 — Create `frontend/src/components/StatCard.vue`

```vue
<template>
  <v-card variant="outlined" rounded="lg">
    <v-card-text class="pa-4">
      <div class="text-h4 font-weight-bold" :style="{ color: valueColor }">{{ value }}</div>
      <div class="text-caption text-uppercase text-medium-emphasis mt-1">{{ label }}</div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{ value: string | number; label: string; valueColor?: string }>()
</script>
```

### Step 3 — Create `frontend/src/components/NamespaceSwitcher.vue`

```vue
<template>
  <v-menu>
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        variant="outlined"
        color="white"
        size="small"
        class="text-white"
        style="border-color: rgba(255,255,255,0.4)"
      >
        <v-icon start>mdi-package-variant</v-icon>
        {{ nsStore.active }}
        <v-icon end>mdi-chevron-down</v-icon>
      </v-btn>
    </template>
    <v-list density="compact" min-width="200">
      <v-list-subheader>Switch Project / Namespace</v-list-subheader>
      <v-list-item
        v-for="ns in nsStore.namespaces"
        :key="ns"
        :value="ns"
        :active="ns === nsStore.active"
        active-color="primary"
        @click="nsStore.setActive(ns)"
      >
        <v-list-item-title>{{ ns }}</v-list-item-title>
      </v-list-item>
      <v-list-item v-if="!nsStore.namespaces.length" disabled>
        <v-list-item-title class="text-medium-emphasis">No namespaces found</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<script setup lang="ts">
import { useNamespaceStore } from '@/stores/namespace'
const nsStore = useNamespaceStore()
</script>
```

### Step 4 — Create `frontend/src/components/TopBar.vue`

```vue
<template>
  <v-app-bar color="primary" elevation="2" height="52">
    <template #prepend>
      <div
        class="d-flex align-center ga-2 px-4"
        style="min-width:220px;border-right:1px solid rgba(255,255,255,0.15)"
      >
        <v-avatar size="32" rounded="sm">
          <v-img src="/ironcore-logo.png" />
        </v-avatar>
        <div>
          <div class="text-white font-weight-bold" style="font-size:15px;line-height:1.1">IronCore</div>
          <div class="text-white" style="font-size:10px;opacity:0.6">Cloud Dashboard</div>
        </div>
      </div>
    </template>

    <v-tabs color="white" class="ml-2">
      <v-tab :to="{ path: '/machines' }"     prepend-icon="mdi-monitor">Machines</v-tab>
      <v-tab :to="{ path: '/volumes' }"      prepend-icon="mdi-database">Volumes</v-tab>
      <v-tab :to="{ path: '/networks' }"     prepend-icon="mdi-lan">Networking</v-tab>
      <v-tab :to="{ path: '/prefixes' }"     prepend-icon="mdi-ip-network">IPAM</v-tab>
    </v-tabs>

    <template #append>
      <div class="d-flex align-center ga-3 pr-4">
        <NamespaceSwitcher />
        <div class="d-flex align-center ga-1 text-caption" style="color:rgba(255,255,255,0.7)">
          <v-icon size="10" color="success">mdi-circle</v-icon>
          ironcore-in-a-box
        </div>
        <v-btn icon="mdi-help-circle-outline" color="white" variant="text" size="small" />
        <v-avatar color="secondary" size="30" style="cursor:pointer">
          <span class="text-caption font-weight-bold text-white">DV</span>
        </v-avatar>
      </div>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
import NamespaceSwitcher from './NamespaceSwitcher.vue'
</script>
```

### Step 5 — Create `frontend/src/components/Sidebar.vue`

```vue
<template>
  <v-list density="compact" nav class="pt-2">
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Compute
    </div>
    <v-list-item to="/machines"      prepend-icon="mdi-monitor"        title="Machines"       rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Storage
    </div>
    <v-list-item to="/volumes"       prepend-icon="mdi-database"       title="Volumes"        rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Networking
    </div>
    <v-list-item to="/networks"      prepend-icon="mdi-lan"            title="Networks"       rounded="lg" />
    <v-list-item to="/virtualips"    prepend-icon="mdi-earth"          title="Virtual IPs"    rounded="lg" />
    <v-list-item to="/loadbalancers" prepend-icon="mdi-scale-balance"  title="Load Balancers" rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      IP Management
    </div>
    <v-list-item to="/prefixes"      prepend-icon="mdi-ip-network"     title="IP Prefixes"    rounded="lg" />
  </v-list>
</template>
```

### Step 6 — Update `frontend/src/layouts/DashboardLayout.vue`

Replace the placeholder layout with the real components:

```vue
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
```

### Step 7 — Create `frontend/src/stores/machines.ts`

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api, type Machine } from '@/api/client'
import { useNamespaceStore } from './namespace'

export const useMachinesStore = defineStore('machines', () => {
  const items   = ref<Machine[]>([])
  const loading = ref(false)
  const error   = ref<string | null>(null)

  async function load() {
    const ns = useNamespaceStore().active
    loading.value = true
    error.value   = null
    try {
      items.value = await api.machines.list(ns)
    } catch (e: any) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }

  async function deleteMachine(name: string) {
    await api.machines.delete(useNamespaceStore().active, name)
    await load()
  }

  async function setPower(name: string, power: 'On' | 'Off') {
    await api.machines.power(useNamespaceStore().active, name, power)
    await load()
  }

  return { items, loading, error, load, deleteMachine, setPower }
})
```

### Step 8 — Replace `frontend/src/views/MachinesView.vue`

```vue
<template>
  <v-container fluid class="pa-6">
    <!-- Page header -->
    <div class="d-flex align-center justify-space-between mb-6">
      <div>
        <h1 class="text-h5 font-weight-bold">Machines</h1>
        <div class="text-caption text-medium-emphasis">{{ nsStore.active }}</div>
      </div>
      <div class="d-flex ga-2">
        <v-btn variant="outlined" prepend-icon="mdi-refresh" @click="store.load()">Refresh</v-btn>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/machines/new' }">Create Machine</v-btn>
      </div>
    </div>

    <!-- Stat cards -->
    <v-row class="mb-6" dense>
      <v-col v-for="s in stats" :key="s.label" cols="auto">
        <StatCard :value="s.value" :label="s.label" :value-color="s.color" />
      </v-col>
    </v-row>

    <!-- Search -->
    <v-text-field
      v-model="search"
      prepend-inner-icon="mdi-magnify"
      placeholder="Search machines…"
      variant="outlined"
      density="compact"
      hide-details
      class="mb-4"
      style="max-width: 320px"
    />

    <!-- Table -->
    <v-card variant="outlined" rounded="lg">
      <v-card-title class="d-flex align-center ga-2 pa-4 bg-grey-lighten-5">
        <v-icon color="primary">mdi-monitor</v-icon>
        All Machines
        <v-chip size="x-small" color="primary">{{ store.items.length }}</v-chip>
        <v-spacer />
      </v-card-title>

      <v-data-table
        :headers="headers"
        :items="filtered"
        :loading="store.loading"
        density="compact"
        hover
        item-key="name"
        :no-data-text="store.error ?? 'No machines found'"
      >
        <template #item.state="{ item }">
          <StatusBadge :state="item.state" />
        </template>
        <template #item.machineClass="{ item }">
          <span class="font-weight-bold">{{ item.machineClass }}</span>
        </template>
        <template #item.ips="{ item }">
          <v-chip v-for="ip in item.ips" :key="ip" size="x-small" color="info" class="mr-1">{{ ip }}</v-chip>
        </template>
        <template #item.volumes="{ item }">
          <v-chip v-for="v in item.volumes" :key="v" size="x-small" variant="outlined" class="mr-1">{{ v }}</v-chip>
        </template>
        <template #item.actions="{ item }">
          <div class="d-flex ga-1">
            <v-btn
              size="x-small" variant="outlined"
              :icon="item.power === 'On' ? 'mdi-power-off' : 'mdi-power'"
              :color="item.power === 'On' ? 'warning' : 'success'"
              :title="item.power === 'On' ? 'Power Off' : 'Power On'"
              @click.stop="store.setPower(item.name, item.power === 'On' ? 'Off' : 'On')"
            />
            <v-btn
              size="x-small" variant="outlined" icon="mdi-delete" color="error"
              title="Delete"
              @click.stop="confirmDelete(item.name)"
            />
          </div>
        </template>
      </v-data-table>
    </v-card>

    <!-- Delete confirm dialog -->
    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title>Delete machine?</v-card-title>
        <v-card-text>
          Machine <strong>{{ pendingDelete }}</strong> will be permanently deleted.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMachinesStore } from '@/stores/machines'
import { useNamespaceStore } from '@/stores/namespace'
import StatusBadge from '@/components/StatusBadge.vue'
import StatCard from '@/components/StatCard.vue'

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
const search  = ref('')
const deleteDialog  = ref(false)
const pendingDelete = ref('')

const headers = [
  { title: 'Name',         key: 'name'         },
  { title: 'Status',       key: 'state'        },
  { title: 'Size',         key: 'machineClass' },
  { title: 'Image',        key: 'image'        },
  { title: 'Volumes',      key: 'volumes'      },
  { title: 'IP Addresses', key: 'ips'          },
  { title: '',             key: 'actions', sortable: false },
]

const filtered = computed(() =>
  store.items.filter(m => m.name.toLowerCase().includes(search.value.toLowerCase()))
)

const stats = computed(() => [
  { label: 'Running',  value: store.items.filter(m => m.state === 'Running').length,  color: '#16a34a' },
  { label: 'Pending',  value: store.items.filter(m => m.state === 'Pending').length,  color: '#d97706' },
  { label: 'Shutdown', value: store.items.filter(m => m.state === 'Shutdown').length, color: '#94a3b8' },
  { label: 'Total',    value: store.items.length,                                     color: '#1e293b' },
])

function confirmDelete(name: string) {
  pendingDelete.value = name
  deleteDialog.value  = true
}

async function doDelete() {
  await store.deleteMachine(pendingDelete.value)
  deleteDialog.value = false
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
```

### Step 9 — Verify in browser

With backend running (`make run --kubeconfig ~/.kube/config`):

```bash
cd frontend && npm run dev
```

Open http://localhost:5173

Expected:
- Blue topbar with IronCore logo, tab navigation, namespace switcher showing "default"
- Left sidebar grouped by Compute / Storage / Networking / IP Management
- Machines view with 4 stat cards (Running: 0, Pending: 0, Shutdown: 0, Total: 0)
- Empty table with "No machines found" message
- "Create Machine" button navigates to stub create view

### Step 10 — Commit

```bash
git add frontend/src/
git commit -m "feat: TopBar, Sidebar, NamespaceSwitcher, MachinesView with stats and table"
```

## Done criteria

- Blue topbar with logo and namespace switcher renders at http://localhost:5173
- Sidebar shows all resource groups
- Machines view loads data (empty is fine), switching namespace triggers reload
- No TypeScript or console errors