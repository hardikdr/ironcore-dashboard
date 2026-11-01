<template>
  <v-container fluid class="pa-6">
    <PageHeader
      :title="name"
      icon="mdi-database"
      :breadcrumbs="[{ text: 'Volumes', to: '/volumes' }, { text: name }]"
    >
      <template #actions>
        <v-btn variant="outlined" prepend-icon="mdi-arrow-left" to="/volumes" aria-label="Back to volumes">Back</v-btn>
        <v-btn
          color="error"
          variant="outlined"
          prepend-icon="mdi-delete"
          aria-label="Delete this volume"
          @click="deleteDialog = true"
        >Delete</v-btn>
      </template>
    </PageHeader>

    <!-- Skeleton while loading -->
    <template v-if="loading">
      <v-skeleton-loader type="table-row@4" class="mb-4" />
    </template>

    <!-- Error state with retry -->
    <div v-else-if="loadError" class="pa-8 text-center">
      <v-icon size="48" color="error" class="mb-3">mdi-alert-circle-outline</v-icon>
      <div class="text-h6 mb-2">Failed to load volume</div>
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
                <td class="text-medium-emphasis pr-4 py-1" width="160">Volume Class</td>
                <td><code class="text-primary">{{ detail.volumeClass || '—' }}</code></td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Size</td>
                <td class="font-weight-medium">{{ detail.sizeGiB }} GiB</td>
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
                <td class="text-medium-emphasis pr-4 py-1" width="160">State</td>
                <td><StatusBadge :state="detail.state" /></td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Volume ID</td>
                <td>
                  <CopyChip v-if="detail.volumeID" :text="detail.volumeID" color="info" />
                  <span v-else class="text-medium-emphasis">—</span>
                </td>
              </tr>
              <tr>
                <td class="text-medium-emphasis pr-4 py-1">Access Driver</td>
                <td>{{ detail.accessDriver || '—' }}</td>
              </tr>
              <tr v-if="detail.machineRef">
                <td class="text-medium-emphasis pr-4 py-1">Attached to</td>
                <td>
                  <router-link :to="`/machines/${detail.machineRef}`" class="text-decoration-none text-primary d-flex align-center ga-1">
                    <v-icon size="14" color="primary">mdi-server</v-icon>
                    {{ detail.machineRef }}
                  </router-link>
                </td>
              </tr>
            </tbody>
          </v-table>
        </v-card>
      </v-col>
    </v-row>

    <!-- Events section -->
    <v-card v-if="detail" :style="cardStyle" variant="flat" rounded="lg" class="pa-4 mt-4">
      <ResourceEvents :resource-name="name" resource-kind="Volume" />
    </v-card>

    <!-- YAML section -->
    <div v-if="detail" class="mt-4">
      <ResourceYaml :resource="detail as Record<string, unknown>" :filename="name" />
    </div>

    <v-dialog v-model="deleteDialog" max-width="420" :scrim="true">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete volume?
        </v-card-title>
        <v-card-text>
          Volume <strong>{{ name }}</strong> will be permanently deleted. This action cannot be undone.
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
import { api, type VolumeDetail } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle } from '@/composables/useCardStyle'
import { relativeTime, absoluteTime } from '@/composables/useTime'
import PageHeader      from '@/components/common/PageHeader.vue'
import StatusBadge     from '@/components/common/StatusBadge.vue'
import CopyChip        from '@/components/common/CopyChip.vue'
import ResourceEvents  from '@/components/common/ResourceEvents.vue'
import ResourceYaml    from '@/components/common/ResourceYaml.vue'

const route   = useRoute()
const router  = useRouter()
const nsStore = useNamespaceStore()
const { cardStyle } = useCardStyle()

const name    = route.params.name as string
const detail  = ref<VolumeDetail | null>(null)
const loading = ref(true)
const deleteDialog = ref(false)
const deleting     = ref(false)
const loadError    = ref<string | null>(null)
const deleteError  = ref<string | null>(null)

async function load() {
  loading.value = true
  loadError.value = null
  try {
    detail.value = await api.volumes.get(nsStore.active, name)
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
    await api.volumes.delete(nsStore.active, name)
    router.push('/volumes')
  } catch (e: unknown) {
    deleteError.value = e instanceof Error ? e.message : String(e)
  } finally {
    deleting.value = false
  }
}

onMounted(load)
</script>
