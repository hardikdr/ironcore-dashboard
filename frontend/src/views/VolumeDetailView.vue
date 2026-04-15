<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/volumes" class="text-primary text-decoration-none">Volumes</router-link>
      <span>›</span><span>{{ name }}</span>
    </div>

    <div class="d-flex align-center ga-3 mb-6">
      <h1 class="text-h5 font-weight-bold">{{ name }}</h1>
      <StatusBadge v-if="detail" :state="detail.state" />
      <v-spacer />
      <v-btn color="error" variant="outlined" prepend-icon="mdi-delete" @click="deleteDialog = true">Delete</v-btn>
    </div>

    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />
    <v-alert v-if="loadError" type="error" class="mb-4">{{ loadError }}</v-alert>

    <v-row v-if="detail">
      <v-col cols="12" md="6">
        <v-card variant="outlined" rounded="lg">
          <v-card-title class="pa-4 pb-2 text-subtitle-1 font-weight-bold">Details</v-card-title>
          <v-divider />
          <v-list density="compact" class="pa-2">
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">VOLUME CLASS</span></template><template #subtitle>{{ detail.volumeClass || '—' }}</template></v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">SIZE</span></template><template #subtitle>{{ detail.sizeGiB }} GiB</template></v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">NAMESPACE</span></template><template #subtitle>{{ detail.namespace }}</template></v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">CREATED</span></template><template #subtitle>{{ detail.createdAt }}</template></v-list-item>
          </v-list>
        </v-card>
      </v-col>
      <v-col cols="12" md="6">
        <v-card variant="outlined" rounded="lg">
          <v-card-title class="pa-4 pb-2 text-subtitle-1 font-weight-bold">Status</v-card-title>
          <v-divider />
          <v-list density="compact" class="pa-2">
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">STATE</span></template><template #subtitle><StatusBadge :state="detail.state" /></template></v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">VOLUME ID</span></template><template #subtitle>{{ detail.volumeID || '—' }}</template></v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">ACCESS DRIVER</span></template><template #subtitle>{{ detail.accessDriver || '—' }}</template></v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ name }}"?</v-card-title>
        <v-card-text>This action cannot be undone. The volume will be permanently deleted.</v-card-text>
        <v-alert v-if="deleteError" type="error" class="mx-4 mb-2">{{ deleteError }}</v-alert>
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
import StatusBadge from '@/components/StatusBadge.vue'

const route   = useRoute()
const router  = useRouter()
const nsStore = useNamespaceStore()
const name    = route.params.name as string
const detail  = ref<VolumeDetail | null>(null)
const loading = ref(true)
const deleteDialog = ref(false)
const deleting     = ref(false)
const loadError    = ref<string | null>(null)
const deleteError  = ref<string | null>(null)

onMounted(async () => {
  try {
    detail.value = await api.volumes.get(nsStore.active, name)
  } catch (e: unknown) {
    loadError.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
})

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
</script>
