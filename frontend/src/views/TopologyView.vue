<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Infrastructure Topology"
      subtitle="Region → AZ → Rack view of all machines"
      icon="mdi-sitemap-outline"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" @click="store.load()">Refresh</v-btn>
      </template>
    </PageHeader>

    <div v-if="store.loading" class="d-flex justify-center pa-12">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <div v-else-if="grouped.size === 0" class="text-center pa-12 text-medium-emphasis">
      <v-icon size="48" class="mb-4">mdi-server-off</v-icon>
      <div class="text-h6">No machines found</div>
    </div>

    <template v-else>
      <div v-for="[region, azMap] in grouped" :key="region" class="mb-8">
        <!-- Region header -->
        <div class="d-flex align-center ga-2 mb-4">
          <v-icon color="primary" size="20">mdi-earth</v-icon>
          <span class="text-h6 font-weight-bold">{{ region }}</span>
          <v-chip size="x-small" color="primary" variant="tonal">Region</v-chip>
        </div>

        <div v-for="[az, rackMap] in azMap" :key="az" class="mb-6 pl-4">
          <!-- AZ header -->
          <div class="d-flex align-center ga-2 mb-3">
            <v-icon color="secondary" size="16">mdi-layers-outline</v-icon>
            <span class="text-subtitle-1 font-weight-medium">{{ az }}</span>
            <v-chip size="x-small" color="secondary" variant="tonal">AZ</v-chip>
          </div>

          <!-- Racks row -->
          <div class="d-flex flex-wrap ga-4 pl-4">
            <div v-for="[rack, machines] in rackMap" :key="rack">
              <v-card :style="cardStyle" variant="flat" class="pa-3" min-width="140">
                <div class="d-flex align-center ga-1 mb-2">
                  <v-icon size="14" color="accent">mdi-server-network</v-icon>
                  <span class="text-caption font-weight-bold">{{ rack }}</span>
                </div>
                <div class="d-flex flex-column ga-1">
                  <v-tooltip
                    v-for="m in machines"
                    :key="m.name"
                    :text="`${m.name} · ${m.machineClass} · ${m.state}`"
                    location="top"
                  >
                    <template #activator="{ props: tp }">
                      <router-link :to="`/machines/${m.name}`" class="text-decoration-none">
                        <div
                          v-bind="tp"
                          class="machine-tile rounded d-flex align-center px-2"
                          :style="tileStyle(m.state)"
                        >
                          <v-icon size="12" class="mr-1">mdi-server</v-icon>
                          <span style="font-size:11px; max-width:90px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap">
                            {{ m.name }}
                          </span>
                        </div>
                      </router-link>
                    </template>
                  </v-tooltip>
                </div>
              </v-card>
            </div>
          </div>
        </div>
      </div>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useTheme } from 'vuetify'
import { useMachinesStore } from '@/stores/machines'
import { useNamespaceStore } from '@/stores/namespace'
import PageHeader from '@/components/common/PageHeader.vue'
import type { Machine } from '@/api/client'

const store   = useMachinesStore()
const nsStore = useNamespaceStore()
const theme   = useTheme()

const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() => ({
  background: isDark.value ? 'rgba(20,30,51,0.8)' : '#ffffff',
  border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.12)'}`,
}))

function parseLocation(m: Machine): { region: string; az: string; rack: string } {
  const annotations = (m as unknown as Record<string, unknown>)['annotations'] as Record<string, string> | undefined
  const loc = annotations?.['metal.ironcore.dev/location'] ?? ''
  const parts = loc.split('/')
  return {
    region: parts[0] || 'default',
    az:     parts[1] || 'default',
    rack:   parts[2] || 'unassigned',
  }
}

const grouped = computed(() => {
  const map = new Map<string, Map<string, Map<string, Machine[]>>>()
  for (const m of store.items) {
    const { region, az, rack } = parseLocation(m)
    if (!map.has(region)) map.set(region, new Map())
    const azMap = map.get(region)!
    if (!azMap.has(az)) azMap.set(az, new Map())
    const rackMap = azMap.get(az)!
    if (!rackMap.has(rack)) rackMap.set(rack, [])
    rackMap.get(rack)!.push(m)
  }
  return map
})

function tileStyle(state: string) {
  const dark = isDark.value
  const colors: Record<string, string> = {
    Running:  dark ? 'rgba(52,211,153,0.15)' : 'rgba(16,185,129,0.1)',
    Pending:  dark ? 'rgba(251,191,36,0.15)'  : 'rgba(245,158,11,0.1)',
    Shutdown: dark ? 'rgba(148,163,184,0.12)' : 'rgba(100,116,139,0.08)',
    Error:    dark ? 'rgba(248,113,113,0.15)' : 'rgba(239,68,68,0.1)',
  }
  const borders: Record<string, string> = {
    Running:  '#34d399',
    Pending:  '#fbbf24',
    Shutdown: '#64748b',
    Error:    '#f87171',
  }
  return {
    height: '28px',
    background: colors[state] ?? colors['Shutdown'],
    border: `1px solid ${borders[state] ?? borders['Shutdown']}`,
    color: 'inherit',
  }
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>

<style scoped>
.machine-tile { transition: opacity 0.15s; }
.machine-tile:hover { opacity: 0.8; }
</style>
