# Task 05 — Vue 3 Frontend Scaffold

## Prerequisite

None — this task runs in parallel with Tasks 03 and 04. You only need `ironcore-dashboard/` to exist.

## Your job

Create the complete Vue 3 frontend scaffold: package.json, Vite config, TypeScript config, app entry point, Vuetify plugin, router, API client, namespace store, and the DashboardLayout. No views yet — those come in Tasks 06–08.

By the end:
- `cd frontend && npm install && npm run dev` starts Vite at http://localhost:5173 without errors
- The layout shell (blue topbar, sidebar, router outlet) renders
- All type-safe API client functions are defined (even though no views call them yet)

## Working directory

All work in `ironcore-dashboard/frontend/`.

## Theme

Primary color is IronCore blue: `#1a5fa8`. Light theme only for v1.

## Files to create

```
frontend/
├── package.json
├── vite.config.ts
├── tsconfig.json
├── index.html
└── src/
    ├── main.ts
    ├── App.vue
    ├── plugins/
    │   └── vuetify.ts
    ├── router/
    │   └── index.ts
    ├── api/
    │   └── client.ts
    ├── stores/
    │   └── namespace.ts
    └── layouts/
        └── DashboardLayout.vue
```

## Step-by-step

### Step 1 — Create `frontend/package.json`

```json
{
  "name": "ironcore-dashboard-frontend",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc && vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "vue": "^3.4.0",
    "vue-router": "^4.3.0",
    "pinia": "^2.1.0",
    "vuetify": "^3.5.0",
    "@mdi/font": "^7.4.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "@vitejs/plugin-vue": "^5.0.0",
    "typescript": "^5.3.0",
    "vite": "^5.1.0",
    "vue-tsc": "^2.0.0"
  }
}
```

### Step 2 — Create `frontend/vite.config.ts`

```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://localhost:8080'
    }
  },
  build: {
    outDir: '../dist/frontend'
  }
})
```

### Step 3 — Create `frontend/tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "module": "ESNext",
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "skipLibCheck": true,
    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "preserve",
    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true,
    "paths": {
      "@/*": ["./src/*"]
    }
  },
  "include": ["src/**/*.ts", "src/**/*.d.ts", "src/**/*.tsx", "src/**/*.vue"]
}
```

### Step 4 — Create `frontend/index.html`

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>IronCore Dashboard</title>
  </head>
  <body>
    <div id="app"></div>
    <script type="module" src="/src/main.ts"></script>
  </body>
</html>
```

### Step 5 — Create `frontend/src/plugins/vuetify.ts`

```typescript
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary:    '#1a5fa8',
          secondary:  '#475569',
          success:    '#16a34a',
          warning:    '#d97706',
          error:      '#b91c1c',
          info:       '#0369a1',
          background: '#f4f6fa',
          surface:    '#ffffff',
        }
      }
    }
  }
})
```

### Step 6 — Create `frontend/src/api/client.ts`

```typescript
const BASE = '/api/v1'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...options
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  if (res.status === 204) return undefined as T
  return res.json()
}

export const api = {
  namespaces: {
    list: () => request<string[]>('/namespaces')
  },
  machines: {
    list:   (ns: string) => request<Machine[]>(`/namespaces/${ns}/machines`),
    get:    (ns: string, name: string) => request<Machine>(`/namespaces/${ns}/machines/${name}`),
    create: (ns: string, body: CreateMachineRequest) =>
      request<Machine>(`/namespaces/${ns}/machines`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/machines/${name}`, { method: 'DELETE' }),
    power:  (ns: string, name: string, power: 'On'|'Off') =>
      request<Machine>(`/namespaces/${ns}/machines/${name}/power`, {
        method: 'PATCH', body: JSON.stringify({ power })
      })
  },
  machineClasses: {
    list: () => request<MachineClass[]>('/machineclasses')
  },
  volumes: {
    list:   (ns: string) => request<Volume[]>(`/namespaces/${ns}/volumes`),
    create: (ns: string, body: CreateVolumeRequest) =>
      request<Volume>(`/namespaces/${ns}/volumes`, { method: 'POST', body: JSON.stringify(body) }),
    delete: (ns: string, name: string) =>
      request<void>(`/namespaces/${ns}/volumes/${name}`, { method: 'DELETE' })
  },
  networks: {
    list:           (ns: string) => request<Network[]>(`/namespaces/${ns}/networks`),
    listInterfaces: (ns: string) => request<NetworkInterface[]>(`/namespaces/${ns}/networkinterfaces`)
  },
  virtualIPs: {
    list: (ns: string) => request<VirtualIP[]>(`/namespaces/${ns}/virtualips`)
  },
  loadBalancers: {
    list: (ns: string) => request<LoadBalancer[]>(`/namespaces/${ns}/loadbalancers`)
  },
  prefixes: {
    list: (ns: string) => request<Prefix[]>(`/namespaces/${ns}/prefixes`)
  }
}

// ── Type definitions (mirror backend JSON) ──────────────────────────────
export interface Machine {
  name: string; namespace: string; state: string; power: string
  machineClass: string; image: string; ips: string[]; volumes: string[]; createdAt: string
}
export interface CreateMachineRequest {
  name: string; machineClass: string; image: string
  networkName: string; volumes: VolumeAttachment[]; power: string
}
export interface VolumeAttachment { name: string; sizeBytes: number; volumeClass: string }
export interface Volume {
  name: string; namespace: string; state: string; sizeBytes: number; volumeClass: string; createdAt: string
}
export interface CreateVolumeRequest { name: string; volumeClass: string; sizeBytes: number }
export interface MachineClass { name: string; cpu: string; ram: string }
export interface Network { name: string; namespace: string; createdAt: string }
export interface NetworkInterface {
  name: string; namespace: string; state: string; ips: string[]
  network: string; machine: string; createdAt: string
}
export interface VirtualIP {
  name: string; namespace: string; ip: string; type: string; ipFamily: string; createdAt: string
}
export interface LoadBalancer {
  name: string; namespace: string; type: string; ips: string[]; createdAt: string
}
export interface Prefix {
  name: string; namespace: string; prefix: string; phase: string; createdAt: string
}
```

### Step 7 — Create `frontend/src/stores/namespace.ts`

```typescript
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { api } from '@/api/client'

export const useNamespaceStore = defineStore('namespace', () => {
  const namespaces = ref<string[]>([])
  const active     = ref<string>('default')

  async function load() {
    namespaces.value = await api.namespaces.list()
    if (namespaces.value.length && !namespaces.value.includes(active.value)) {
      active.value = namespaces.value[0]
    }
  }

  function setActive(ns: string) { active.value = ns }

  return { namespaces, active, load, setActive }
})
```

### Step 8 — Create `frontend/src/router/index.ts`

```typescript
import { createRouter, createWebHistory } from 'vue-router'
import DashboardLayout from '@/layouts/DashboardLayout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: DashboardLayout,
      redirect: '/machines',
      children: [
        { path: 'machines',          component: () => import('@/views/MachinesView.vue') },
        { path: 'machines/new',      component: () => import('@/views/MachineCreateView.vue') },
        { path: 'machines/:name',    component: () => import('@/views/MachineDetailView.vue') },
        { path: 'volumes',           component: () => import('@/views/VolumesView.vue') },
        { path: 'networks',          component: () => import('@/views/NetworksView.vue') },
        { path: 'virtualips',        component: () => import('@/views/VirtualIPsView.vue') },
        { path: 'loadbalancers',     component: () => import('@/views/LoadBalancersView.vue') },
        { path: 'prefixes',          component: () => import('@/views/IPPrefixesView.vue') },
      ]
    }
  ]
})

export default router
```

### Step 9 — Create `frontend/src/layouts/DashboardLayout.vue`

The layout wraps the full page: TopBar at top, sidebar on left, content area on right. TopBar and Sidebar components will be created in Task 06.

For now, use placeholder components:

```vue
<template>
  <v-app theme="light">
    <!-- TopBar placeholder — Task 06 will create the real component -->
    <v-app-bar color="primary" elevation="2" height="52">
      <v-app-bar-title class="text-white font-weight-bold">IronCore Dashboard</v-app-bar-title>
    </v-app-bar>

    <!-- Sidebar placeholder -->
    <v-navigation-drawer permanent width="220">
      <v-list density="compact" nav class="pt-2">
        <v-list-item to="/machines"      title="Machines"       prepend-icon="mdi-monitor" rounded="lg" />
        <v-list-item to="/volumes"       title="Volumes"        prepend-icon="mdi-database" rounded="lg" />
        <v-list-item to="/networks"      title="Networking"     prepend-icon="mdi-lan" rounded="lg" />
        <v-list-item to="/prefixes"      title="IPAM"           prepend-icon="mdi-ip-network" rounded="lg" />
      </v-list>
    </v-navigation-drawer>

    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { useNamespaceStore } from '@/stores/namespace'
import { onMounted } from 'vue'

const nsStore = useNamespaceStore()
onMounted(() => nsStore.load())
</script>
```

### Step 10 — Create stub views so the router doesn't 404

The router imports views lazily. Create minimal stubs for each:

```bash
mkdir -p frontend/src/views
```

For each of the 8 view files, create a minimal stub:

`frontend/src/views/MachinesView.vue`:
```vue
<template><v-container fluid class="pa-6"><h1 class="text-h5">Machines</h1></v-container></template>
```

Create identical stubs (change the title) for:
- `MachineCreateView.vue` → "Create Machine"
- `MachineDetailView.vue` → "Machine Detail"
- `VolumesView.vue` → "Volumes"
- `NetworksView.vue` → "Networking"
- `VirtualIPsView.vue` → "Virtual IPs"
- `LoadBalancersView.vue` → "Load Balancers"
- `IPPrefixesView.vue` → "IP Prefixes"

### Step 11 — Create `frontend/src/main.ts` and `App.vue`

`frontend/src/main.ts`:
```typescript
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import vuetify from './plugins/vuetify'

createApp(App).use(createPinia()).use(router).use(vuetify).mount('#app')
```

`frontend/src/App.vue`:
```vue
<template><router-view /></template>
```

### Step 12 — Install and verify

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:5173 — should show the blue topbar, sidebar, and "Machines" placeholder. No console errors.

### Step 13 — Commit

```bash
git add frontend/
git commit -m "feat: Vue 3 + Vuetify 3 + Pinia + router scaffold with stub views"
```

## Done criteria

- `npm run dev` starts without errors at http://localhost:5173
- Navigating to `/machines`, `/volumes`, `/networks` shows the stub views
- No TypeScript or console errors
- Namespace store is wired (will fail gracefully if backend isn't running)

## Next tasks (can run in parallel after this)

- Task 06: Real TopBar, Sidebar, NamespaceSwitcher, MachinesView
- Task 07: Create Machine wizard
- Task 08: Volumes, Networking, IPAM views