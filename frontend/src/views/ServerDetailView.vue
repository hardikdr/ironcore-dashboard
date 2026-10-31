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

    <div v-if="loading" class="d-flex justify-center pa-12">
      <v-progress-circular indeterminate color="primary" />
    </div>

    <div v-else-if="!server" class="text-center pa-12 text-medium-emphasis">
      Server "{{ serverName }}" not found.
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
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useTheme } from 'vuetify'
import { api, type Server } from '@/api/client'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const route      = useRoute()
const theme      = useTheme()
const serverName = route.params.name as string
const server     = ref<Server | null>(null)
const loading    = ref(true)

const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() => ({
  background: isDark.value ? 'rgba(20,30,51,0.8)' : '#ffffff',
  border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.12)'}`,
}))

async function load() {
  loading.value = true
  try {
    const all = await api.servers.list()
    server.value = all.find(s => s.name === serverName) ?? null
  } catch {
    server.value = null
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
