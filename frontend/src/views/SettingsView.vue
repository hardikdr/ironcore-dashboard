<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Settings"
      subtitle="Dashboard preferences and cluster connection"
      icon="mdi-cog-outline"
    />

    <v-row>
      <!-- Appearance -->
      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" class="pa-4 mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-palette-outline</v-icon>
            Appearance
          </div>
          <div class="d-flex align-center justify-space-between">
            <div>
              <div class="text-body-2 font-weight-medium">Dark Mode</div>
              <div class="text-caption text-medium-emphasis">Switch between dark and light theme</div>
            </div>
            <v-switch
              :model-value="isDark"
              color="primary"
              hide-details
              @update:model-value="toggleTheme"
            />
          </div>
        </v-card>
      </v-col>

      <!-- Cluster Connection -->
      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" class="pa-4 mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="success">mdi-lan-connect</v-icon>
            Cluster Connection
          </div>
          <v-table density="compact">
            <tbody>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">API Server</td>
                <td><code class="text-caption">{{ apiServer }}</code></td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Namespaces</td>
                <td>{{ nsStore.namespaces.length }}</td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Active Namespace</td>
                <td><v-chip size="x-small" color="primary">{{ nsStore.active }}</v-chip></td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Status</td>
                <td><v-chip size="x-small" color="success" prepend-icon="mdi-circle-small">Connected</v-chip></td>
              </tr>
            </tbody>
          </v-table>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { useTheme } from 'vuetify'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import PageHeader from '@/components/common/PageHeader.vue'

const theme   = useTheme()
const nsStore = useNamespaceStore()
const { isDark, cardStyle } = useCardStyle()

const apiServer = window.location.origin

function toggleTheme() {
  theme.global.name.value = isDark.value ? 'light' : 'dark'
}
</script>
