<template>
  <v-dialog v-model="open" max-width="600" scrollable>
    <v-card rounded="xl" :style="cardStyle">
      <v-card-text class="pa-0">
        <div class="d-flex align-center px-4 pt-4 pb-2 ga-2">
          <v-icon color="medium-emphasis">mdi-magnify</v-icon>
          <v-text-field
            ref="inputRef"
            v-model="query"
            placeholder="Search resources…"
            variant="plain"
            density="compact"
            hide-details
            autofocus
            class="flex-grow-1"
            @keydown.escape="open = false"
          />
          <v-kbd class="text-caption">Esc</v-kbd>
        </div>

        <v-divider />

        <div v-if="query.length < 2" class="px-4 py-6 text-center text-medium-emphasis text-body-2">
          Type at least 2 characters to search
        </div>

        <v-list v-else-if="results.length > 0" density="compact" class="py-2">
          <v-list-item
            v-for="r in results"
            :key="r.kind + r.name"
            :to="r.path"
            rounded="lg"
            class="mx-2 mb-1"
            @click="open = false"
          >
            <template #prepend>
              <v-icon :color="r.color" size="20">{{ r.icon }}</v-icon>
            </template>
            <v-list-item-title class="font-weight-medium text-body-2">{{ r.name }}</v-list-item-title>
            <v-list-item-subtitle class="text-caption">{{ r.kind }}</v-list-item-subtitle>
            <template #append>
              <StatusBadge v-if="r.state" :state="r.state" />
            </template>
          </v-list-item>
        </v-list>

        <div v-else class="px-4 py-6 text-center text-medium-emphasis text-body-2">
          No results for "{{ query }}"
        </div>

        <v-divider />
        <div class="d-flex align-center ga-3 px-4 py-2 text-caption text-medium-emphasis">
          <span><v-kbd>↑↓</v-kbd> navigate</span>
          <span><v-kbd>Enter</v-kbd> open</span>
          <span><v-kbd>Esc</v-kbd> close</span>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useTheme } from 'vuetify'
import { useMachinesStore }      from '@/stores/machines'
import { useVolumesStore }       from '@/stores/volumes'
import { useNetworksStore }      from '@/stores/networks'
import { useVirtualIPsStore }    from '@/stores/virtualips'
import { useLoadBalancersStore } from '@/stores/loadbalancers'
import { usePrefixesStore }      from '@/stores/prefixes'
import StatusBadge from '@/components/common/StatusBadge.vue'

const open  = defineModel<boolean>({ default: false })
const query = ref('')
const vuetifyTheme = useTheme()

const isDark = computed(() => vuetifyTheme.global.current.value.dark)
const cardStyle = computed(() => isDark.value
  ? 'background: #0d1526; border: 1px solid rgba(129,140,248,0.15);'
  : 'background: #fff; border: 1px solid rgba(99,102,241,0.15);'
)

const machinesStore  = useMachinesStore()
const volumesStore   = useVolumesStore()
const networksStore  = useNetworksStore()
const vipsStore      = useVirtualIPsStore()
const lbsStore       = useLoadBalancersStore()
const prefixesStore  = usePrefixesStore()

interface SearchResult {
  kind: string
  name: string
  path: string
  icon: string
  color: string
  state?: string
}

const results = computed<SearchResult[]>(() => {
  const q = query.value.trim().toLowerCase()
  if (q.length < 2) return []

  const out: SearchResult[] = []

  machinesStore.items.forEach(m => {
    if (m.name.toLowerCase().includes(q) || m.machineClass.toLowerCase().includes(q)) {
      out.push({ kind: 'Machine', name: m.name, path: `/machines/${m.name}`, icon: 'mdi-server', color: 'primary', state: m.state })
    }
  })

  volumesStore.items.forEach(v => {
    if (v.name.toLowerCase().includes(q)) {
      out.push({ kind: 'Volume', name: v.name, path: `/volumes/${v.name}`, icon: 'mdi-database-outline', color: 'blue', state: v.state })
    }
  })

  networksStore.items.forEach((n: { name: string }) => {
    if (n.name.toLowerCase().includes(q)) {
      out.push({ kind: 'Network', name: n.name, path: `/networks/${n.name}`, icon: 'mdi-lan', color: 'teal' })
    }
  })

  vipsStore.items.forEach(v => {
    if (v.name.toLowerCase().includes(q) || (v.ip ?? '').includes(q)) {
      out.push({ kind: 'Virtual IP', name: v.name, path: `/virtualips/${v.name}`, icon: 'mdi-earth', color: 'orange' })
    }
  })

  lbsStore.items.forEach(lb => {
    if (lb.name.toLowerCase().includes(q)) {
      out.push({ kind: 'Load Balancer', name: lb.name, path: `/loadbalancers/${lb.name}`, icon: 'mdi-scale-balance', color: 'purple' })
    }
  })

  prefixesStore.items.forEach(p => {
    if (p.name.toLowerCase().includes(q) || (p.prefix ?? '').includes(q)) {
      out.push({ kind: 'IP Prefix', name: p.name, path: `/prefixes/${p.name}`, icon: 'mdi-ip-network-outline', color: 'cyan', state: p.phase })
    }
  })

  return out.slice(0, 20)
})

watch(open, v => { if (!v) query.value = '' })

function onKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    open.value = true
  }
  if (e.key === '/' && (e.target as HTMLElement).tagName !== 'INPUT' && (e.target as HTMLElement).tagName !== 'TEXTAREA') {
    e.preventDefault()
    open.value = true
  }
}

onMounted(() => document.addEventListener('keydown', onKeydown))
onUnmounted(() => document.removeEventListener('keydown', onKeydown))
</script>
