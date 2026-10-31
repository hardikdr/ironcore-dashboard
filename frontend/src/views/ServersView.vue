<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Dedicated Servers"
      subtitle="metal-operator bare-metal hosts"
      icon="mdi-server-network"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-refresh" @click="load">Refresh</v-btn>
      </template>
    </PageHeader>

    <v-row class="mb-6" dense>
      <v-col cols="auto">
        <StatCard label="On" :value="onCount" icon="mdi-power" color="success" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Off" :value="offCount" icon="mdi-power-off" color="error" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Claimed" :value="claimedCount" icon="mdi-check-circle-outline" color="primary" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Total" :value="items.length" icon="mdi-server-network" color="secondary" />
      </v-col>
    </v-row>

    <v-card v-if="items.length === 0 && !loading" :style="cardStyle" variant="flat" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-server-off</v-icon>
        <div class="text-h6 mb-2">No servers found</div>
        <div class="text-medium-emphasis">metal-operator Server CRDs are not present in this cluster.</div>
      </v-card-text>
    </v-card>

    <v-card v-else :style="cardStyle" variant="flat" rounded="lg">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-server-network</v-icon>
        <span class="font-weight-bold">All Servers</span>
        <v-chip size="x-small" color="primary">{{ items.length }}</v-chip>
      </div>
      <v-data-table
        :headers="headers"
        :items="items"
        :loading="loading"
        density="compact"
        hover
      >
        <template #item.name="{ item }">
          <router-link :to="`/servers/${item.name}`" class="text-primary text-decoration-none font-weight-medium">
            {{ item.name }}
          </router-link>
        </template>
        <template #item.powerState="{ item }">
          <StatusBadge :state="item.powerState" />
        </template>
        <template #item.claimStatus="{ item }">
          <v-chip
            :color="item.claimStatus === 'Claimed' ? 'primary' : 'default'"
            size="x-small"
            variant="tonal"
          >
            {{ item.claimStatus }}
          </v-chip>
        </template>
      </v-data-table>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTheme } from 'vuetify'
import { api, type Server } from '@/api/client'
import PageHeader from '@/components/common/PageHeader.vue'
import StatCard from '@/components/common/StatCard.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'

const theme   = useTheme()
const items   = ref<Server[]>([])
const loading = ref(false)

const isDark = computed(() => theme.global.current.value.dark)
const cardStyle = computed(() => ({
  background: isDark.value ? 'rgba(20,30,51,0.8)' : '#ffffff',
  border: `1px solid ${isDark.value ? 'rgba(129,140,248,0.1)' : 'rgba(0,0,0,0.12)'}`,
}))
const headerStyle = computed(() => ({
  background: isDark.value ? 'rgba(14,21,40,0.9)' : '#f8fafc',
  borderBottom: `1px solid ${isDark.value ? 'rgba(129,140,248,0.08)' : 'rgba(0,0,0,0.08)'}`,
}))

const onCount      = computed(() => items.value.filter(s => s.powerState === 'On').length)
const offCount     = computed(() => items.value.filter(s => s.powerState === 'Off').length)
const claimedCount = computed(() => items.value.filter(s => s.claimStatus === 'Claimed').length)

const headers = [
  { title: 'Name',         key: 'name'        },
  { title: 'BMC Address',  key: 'bmcAddress'  },
  { title: 'Power State',  key: 'powerState'  },
  { title: 'Location',     key: 'location'    },
  { title: 'Claim Status', key: 'claimStatus' },
]

async function load() {
  loading.value = true
  try { items.value = await api.servers.list() }
  catch { items.value = [] }
  finally { loading.value = false }
}

onMounted(load)
</script>
