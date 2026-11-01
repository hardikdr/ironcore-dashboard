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

      <!-- Authentication -->
      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" class="pa-4 mb-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="warning">mdi-key-outline</v-icon>
            Authentication
          </div>

          <v-alert
            v-if="!tokenValue"
            type="info"
            variant="tonal"
            density="compact"
            class="mb-4"
            icon="mdi-information-outline"
          >
            No token configured — using server default credentials
          </v-alert>

          <v-text-field
            v-model="tokenInput"
            label="Bearer token"
            :type="showToken ? 'text' : 'password'"
            variant="outlined"
            density="compact"
            clearable
            :append-inner-icon="showToken ? 'mdi-eye-off' : 'mdi-eye'"
            @click:append-inner="showToken = !showToken"
            class="mb-3"
          />

          <v-btn
            color="primary"
            variant="tonal"
            prepend-icon="mdi-content-save"
            :disabled="tokenInput === tokenValue"
            @click="saveToken"
          >
            Save token
          </v-btn>
          <v-btn
            v-if="tokenValue"
            color="error"
            variant="text"
            prepend-icon="mdi-delete-outline"
            class="ml-2"
            @click="clearToken"
          >
            Clear
          </v-btn>

          <v-alert
            v-if="tokenSaved"
            type="success"
            variant="tonal"
            density="compact"
            class="mt-3"
          >
            Token saved successfully.
          </v-alert>

          <div class="text-caption text-medium-emphasis mt-4">
            Token is stored in browser localStorage and sent with every API request. Clear it when you're done.
          </div>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useTheme } from 'vuetify'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import PageHeader from '@/components/common/PageHeader.vue'
import { getAuthToken } from '@/api/client'

const TOKEN_STORAGE_KEY = 'ironcore-dashboard-token'

const theme   = useTheme()
const nsStore = useNamespaceStore()
const { isDark, cardStyle } = useCardStyle()

const apiServer  = window.location.origin
const tokenValue = ref<string | null>(getAuthToken())
const tokenInput = ref<string>(getAuthToken() ?? '')
const showToken  = ref(false)
const tokenSaved = ref(false)

function toggleTheme() {
  theme.global.name.value = isDark.value ? 'light' : 'dark'
}

function saveToken() {
  const t = tokenInput.value?.trim() ?? ''
  if (t) {
    localStorage.setItem(TOKEN_STORAGE_KEY, t)
  } else {
    localStorage.removeItem(TOKEN_STORAGE_KEY)
  }
  tokenValue.value = t || null
  tokenSaved.value = true
  setTimeout(() => { tokenSaved.value = false }, 3000)
}

function clearToken() {
  localStorage.removeItem(TOKEN_STORAGE_KEY)
  tokenValue.value = null
  tokenInput.value = ''
  tokenSaved.value = false
}
</script>
