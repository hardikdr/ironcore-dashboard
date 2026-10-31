<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Server Detail"
      icon="mdi-server-network"
      :breadcrumbs="[{ text: 'Servers', to: '/servers' }, { text: serverName }]"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-arrow-left" to="/servers">Back</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading">
      <v-skeleton-loader type="table-row@3" class="mb-4" />
    </template>

    <div v-else-if="loadError" class="pa-8 text-center">
      <v-icon size="48" color="error" class="mb-3">mdi-alert-circle-outline</v-icon>
      <div class="text-h6 mb-2">Failed to load server</div>
      <div class="text-medium-emphasis mb-4">{{ loadError }}</div>
      <v-btn color="primary" @click="load">Retry</v-btn>
    </div>

    <div v-else-if="!server" class="pa-8 text-center text-medium-emphasis">
      <v-icon size="48" class="mb-3">mdi-server-off</v-icon>
      <div class="text-h6">Server "{{ serverName }}" not found</div>
    </div>

    <v-row v-else>
      <!-- BMC Info -->
      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" class="pa-4 mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-chip</v-icon>
            BMC Information
          </div>
          <v-table density="compact">
            <tbody>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Address</td>
                <td><code>{{ server.bmcAddress || '—' }}</code></td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Power State</td>
                <td><StatusBadge :state="server.powerState" /></td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Claim Status</td>
                <td>
                  <v-chip :color="server.claimStatus === 'Claimed' ? 'primary' : 'default'" size="x-small" variant="tonal">
                    {{ server.claimStatus }}
                  </v-chip>
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-card>
      </v-col>

      <!-- Location -->
      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" class="pa-4 mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="secondary">mdi-map-marker-outline</v-icon>
            Location
          </div>
          <div v-if="server.location">
            <v-chip
              v-for="(part, i) in server.location.split('/')"
              :key="i"
              :color="(['primary', 'secondary', 'accent'] as const)[i] ?? 'default'"
              size="small"
              variant="tonal"
              class="mr-2 mb-2"
            >
              {{ (['Region', 'AZ', 'Rack'] as const)[i] ?? 'Part' }}: {{ part }}
            </v-chip>
          </div>
          <div v-else class="text-medium-emphasis text-caption">No location labels configured.</div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { api, type Server } from '@/api/client'
import { useCardStyle } from '@/composables/useCardStyle'
import PageHeader  from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const route      = useRoute()
const serverName = route.params.name as string
const server     = ref<Server | null>(null)
const loading    = ref(true)
const loadError  = ref<string | null>(null)

const { cardStyle } = useCardStyle()

async function load() {
  loading.value = true
  loadError.value = null
  try {
    const all = await api.servers.list()
    server.value = all.find(s => s.name === serverName) ?? null
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
