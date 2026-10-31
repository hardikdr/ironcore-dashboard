<template>
  <v-container fluid class="pa-6">
    <PageHeader
      :title="name"
      icon="mdi-ip-network"
      :breadcrumbs="[{ text: 'IP Prefixes', to: '/prefixes' }, { text: name }]"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-arrow-left" to="/prefixes" aria-label="Back to IP prefixes">Back</v-btn>
        <v-btn
          color="error"
          variant="outlined"
          prepend-icon="mdi-delete"
          aria-label="Delete this IP prefix"
          @click="deleteDialog = true"
        >Delete</v-btn>
      </template>
    </PageHeader>

    <template v-if="loading">
      <v-skeleton-loader type="table-row@4" class="mb-4" />
    </template>

    <div v-else-if="loadError" class="pa-8 text-center">
      <v-icon size="48" color="error" class="mb-3">mdi-alert-circle-outline</v-icon>
      <div class="text-h6 mb-2">Failed to load IP prefix</div>
      <div class="text-medium-emphasis mb-4">{{ loadError }}</div>
      <v-btn color="primary" @click="load">Retry</v-btn>
    </div>

    <v-row v-else-if="detail">
      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" rounded="lg" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="primary">mdi-information-outline</v-icon>
            Details
          </div>
          <v-table density="compact">
            <tbody>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1" width="160">CIDR</td>
                <td>
                  <CopyChip v-if="detail.prefix" :text="detail.prefix" color="primary" />
                  <span v-else class="text-medium-emphasis">—</span>
                </td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">IP Family</td>
                <td>
                  <v-chip size="x-small" variant="tonal" :color="detail.ipFamily === 'IPv6' ? 'secondary' : 'primary'">
                    {{ detail.ipFamily }}
                  </v-chip>
                </td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Parent</td>
                <td>{{ detail.parentRef || 'None' }}</td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Namespace</td>
                <td>{{ detail.namespace }}</td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Created</td>
                <td :title="absoluteTime(detail.createdAt)">{{ relativeTime(detail.createdAt) }}</td>
              </tr>
            </tbody>
          </v-table>
        </v-card>
      </v-col>

      <v-col cols="12" md="6">
        <v-card :style="cardStyle" variant="flat" rounded="lg" class="pa-4">
          <div class="text-subtitle-2 font-weight-bold mb-4 d-flex align-center ga-2">
            <v-icon size="16" color="secondary">mdi-check-circle-outline</v-icon>
            Status
          </div>
          <v-table density="compact">
            <tbody>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1" width="160">Phase</td>
                <td><StatusBadge :state="detail.phase" /></td>
              </tr>
            </tbody>
          </v-table>

          <template v-if="detail.used.length">
            <div class="text-caption text-medium-emphasis font-weight-medium mt-4 mb-2">Allocated sub-prefixes</div>
            <div class="d-flex flex-wrap ga-1">
              <v-chip v-for="u in detail.used" :key="u" size="small" color="info" variant="tonal">
                <code>{{ u }}</code>
              </v-chip>
            </div>
          </template>
          <div v-else class="text-caption text-medium-emphasis mt-4">No sub-prefixes allocated.</div>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="deleteDialog" max-width="420">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete IP prefix?
        </v-card-title>
        <v-card-text>
          IP prefix <strong>{{ name }}</strong> will be permanently deleted. This action cannot be undone.
        </v-card-text>
        <v-alert v-if="deleteError" type="error" variant="tonal" class="mx-4 mb-2" density="compact">{{ deleteError }}</v-alert>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deleting" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, type PrefixDetail } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle } from '@/composables/useCardStyle'
import { relativeTime, absoluteTime } from '@/composables/useTime'
import PageHeader  from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import CopyChip    from '@/components/common/CopyChip.vue'

const route   = useRoute()
const router  = useRouter()
const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const name    = route.params.name as string
const detail  = ref<PrefixDetail | null>(null)
const loading = ref(true)
const deleteDialog = ref(false)
const deleting     = ref(false)
const loadError    = ref<string | null>(null)
const deleteError  = ref<string | null>(null)

async function load() {
  loading.value = true
  loadError.value = null
  try {
    detail.value = await api.prefixes.get(nsStore.active, name)
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function doDelete() {
  deleting.value = true
  deleteError.value = null
  try {
    await api.prefixes.delete(nsStore.active, name)
    router.push('/prefixes')
  } catch (e: unknown) {
    deleteError.value = e instanceof Error ? e.message : String(e)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>
