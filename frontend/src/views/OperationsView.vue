<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Operations"
      subtitle="Kubernetes events across all resource types"
      icon="mdi-clipboard-list-outline"
    />

    <!-- Filter bar -->
    <v-row class="mb-4" align="center" dense>
      <v-col cols="12" sm="5" md="4">
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          placeholder="Search reason, message, resource…"
          density="compact"
          variant="outlined"
          hide-details
          clearable
        />
      </v-col>
      <v-col cols="12" sm="auto">
        <div class="d-flex flex-wrap ga-2">
          <v-btn-toggle v-model="typeFilter" density="compact" divided variant="outlined" rounded="lg">
            <v-btn value="">All</v-btn>
            <v-btn value="Warning" color="warning">
              <v-icon start size="14">mdi-alert-outline</v-icon>Warning
            </v-btn>
            <v-btn value="Normal" color="success">
              <v-icon start size="14">mdi-information-outline</v-icon>Normal
            </v-btn>
          </v-btn-toggle>

          <v-btn-toggle v-model="kindFilter" density="compact" divided variant="outlined" rounded="lg">
            <v-btn value="">All kinds</v-btn>
            <v-btn v-for="k in availableKinds" :key="k" :value="k">{{ k }}</v-btn>
          </v-btn-toggle>
        </div>
      </v-col>
      <v-spacer />
      <v-col cols="auto">
        <v-btn
          size="small"
          variant="tonal"
          prepend-icon="mdi-refresh"
          :loading="loading"
          @click="load"
        >Refresh</v-btn>
      </v-col>
    </v-row>

    <!-- Summary chips -->
    <div v-if="!loading && events.length > 0" class="d-flex flex-wrap ga-2 mb-4">
      <v-chip size="small" color="error" variant="tonal" prepend-icon="mdi-alert-outline">
        {{ warningCount }} warning{{ warningCount !== 1 ? 's' : '' }}
      </v-chip>
      <v-chip size="small" color="success" variant="tonal" prepend-icon="mdi-check-circle-outline">
        {{ normalCount }} normal
      </v-chip>
      <v-chip size="small" variant="tonal" prepend-icon="mdi-clock-outline">
        Last refreshed {{ lastRefreshedLabel }}
      </v-chip>
    </div>

    <!-- Error state -->
    <v-alert v-if="error" type="error" variant="tonal" class="mb-4">
      {{ error }}
      <template #append>
        <v-btn variant="text" size="small" @click="load">Retry</v-btn>
      </template>
    </v-alert>

    <!-- Empty state -->
    <div v-else-if="!loading && filteredEvents.length === 0" class="pa-12 text-center">
      <v-icon size="56" color="success" class="mb-4">mdi-check-circle-outline</v-icon>
      <div class="text-h6 mb-2">No events</div>
      <div class="text-medium-emphasis">
        {{ events.length > 0 ? 'No events match your filter.' : 'No events recorded in this namespace.' }}
      </div>
    </div>

    <!-- Events table -->
    <v-card v-else :style="cardStyle" variant="flat">
      <v-data-table
        :headers="headers"
        :items="filteredEvents"
        :loading="loading"
        :items-per-page="50"
        density="compact"
        :sort-by="[{ key: 'lastTime', order: 'desc' }]"
      >
        <template #item.type="{ item }">
          <v-chip
            :color="item.type === 'Warning' ? 'warning' : 'success'"
            size="x-small"
            variant="tonal"
            :prepend-icon="item.type === 'Warning' ? 'mdi-alert-outline' : 'mdi-information-outline'"
          >{{ item.type }}</v-chip>
        </template>

        <template #item.regarding="{ item }">
          <router-link
            v-if="resourceLink(item)"
            :to="resourceLink(item)!"
            class="text-decoration-none text-primary font-weight-medium"
          >{{ item.regarding }}</router-link>
          <code v-else class="text-medium-emphasis">{{ item.regarding }}</code>
          <v-chip
            size="x-small"
            variant="text"
            class="ml-1 text-medium-emphasis"
          >{{ item.regardsKind }}</v-chip>
        </template>

        <template #item.reason="{ item }">
          <code class="text-caption">{{ item.reason }}</code>
        </template>

        <template #item.message="{ item }">
          <span class="text-caption" style="white-space:normal; word-break:break-word;">{{ item.message }}</span>
        </template>

        <template #item.count="{ item }">
          <v-chip v-if="item.count > 1" size="x-small" variant="tonal" color="info">×{{ item.count }}</v-chip>
          <span v-else class="text-caption text-medium-emphasis">—</span>
        </template>

        <template #item.lastTime="{ item }">
          <span class="text-caption text-medium-emphasis" :title="item.lastTime">
            {{ relativeTime(item.lastTime) }}
          </span>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type ResourceEvent } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle } from '@/composables/useCardStyle'
import { relativeTime } from '@/composables/useTime'
import PageHeader from '@/components/common/PageHeader.vue'

const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const events      = ref<ResourceEvent[]>([])
const loading     = ref(false)
const error       = ref<string | null>(null)
const search      = ref('')
const typeFilter  = ref('')
const kindFilter  = ref('')
const lastRefreshed = ref<Date | null>(null)

const lastRefreshedLabel = computed(() => {
  if (!lastRefreshed.value) return '—'
  return relativeTime(lastRefreshed.value.toISOString())
})

async function load() {
  loading.value = true
  error.value = null
  try {
    events.value = await api.events.list(nsStore.active)
    lastRefreshed.value = new Date()
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

const warningCount = computed(() => events.value.filter(e => e.type === 'Warning').length)
const normalCount  = computed(() => events.value.filter(e => e.type === 'Normal').length)

const availableKinds = computed(() => {
  const s = new Set(events.value.map(e => e.regardsKind).filter(Boolean))
  return [...s].sort()
})

const filteredEvents = computed(() => {
  let list = events.value
  if (typeFilter.value) list = list.filter(e => e.type === typeFilter.value)
  if (kindFilter.value) list = list.filter(e => e.regardsKind === kindFilter.value)
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    list = list.filter(e =>
      e.reason.toLowerCase().includes(q) ||
      e.message.toLowerCase().includes(q) ||
      e.regarding.toLowerCase().includes(q)
    )
  }
  return list
})

const KIND_ROUTES: Record<string, string> = {
  Machine:     '/machines',
  Volume:      '/volumes',
  Network:     '/networks',
  VirtualIP:   '/virtualips',
  LoadBalancer:'/loadbalancers',
  Prefix:      '/prefixes',
}

function resourceLink(e: ResourceEvent): string | null {
  const base = KIND_ROUTES[e.regardsKind]
  if (!base || !e.regarding) return null
  return `${base}/${e.regarding}`
}

const headers = [
  { title: 'Type',     key: 'type',      width: '100px',  sortable: true },
  { title: 'Resource', key: 'regarding', width: '200px',  sortable: true },
  { title: 'Reason',   key: 'reason',    width: '160px',  sortable: true },
  { title: 'Message',  key: 'message',   sortable: false },
  { title: 'Count',    key: 'count',     width: '70px',   sortable: true },
  { title: 'Last seen',key: 'lastTime',  width: '130px',  sortable: true },
]

onMounted(load)
</script>
