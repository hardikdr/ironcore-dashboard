<template>
  <v-container fluid class="pa-6">
    <div class="d-flex align-center ga-2 mb-2 text-caption text-medium-emphasis">
      <router-link to="/virtualips" class="text-primary text-decoration-none">Virtual IPs</router-link>
      <span>›</span><span>{{ name }}</span>
    </div>

    <div class="d-flex align-center ga-3 mb-6">
      <h1 class="text-h5 font-weight-bold">{{ name }}</h1>
      <v-spacer />
      <v-btn color="error" variant="outlined" prepend-icon="mdi-delete" @click="deleteDialog = true">Delete</v-btn>
    </div>

    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

    <v-row v-if="detail">
      <v-col cols="12" md="6">
        <v-card variant="outlined" rounded="lg">
          <v-card-title class="pa-4 pb-2 text-subtitle-1 font-weight-bold">Details</v-card-title>
          <v-divider />
          <v-list density="compact" class="pa-2">
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">TYPE</span></template><template #subtitle>{{ detail.type }}</template></v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">IP FAMILY</span></template><template #subtitle>{{ detail.ipFamily }}</template></v-list-item>
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
            <v-list-item>
              <template #title><span class="text-caption text-medium-emphasis">ALLOCATED IP</span></template>
              <template #subtitle>
                <v-chip v-if="detail.ip" size="small" color="deep-purple">{{ detail.ip }}</v-chip>
                <span v-else class="text-medium-emphasis">Pending…</span>
              </template>
            </v-list-item>
            <v-list-item><template #title><span class="text-caption text-medium-emphasis">ATTACHED NIC</span></template><template #subtitle>{{ detail.targetRef || 'Unattached' }}</template></v-list-item>
          </v-list>
        </v-card>
      </v-col>
    </v-row>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title>Delete "{{ name }}"?</v-card-title>
        <v-card-text>This action cannot be undone. The Virtual IP will be permanently deleted.</v-card-text>
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
import { api, type VirtualIPDetail } from '@/api/client'
import { useNamespaceStore } from '@/stores/namespace'

const route   = useRoute()
const router  = useRouter()
const nsStore = useNamespaceStore()
const name    = route.params.name as string
const detail  = ref<VirtualIPDetail | null>(null)
const loading = ref(true)
const deleteDialog = ref(false)
const deleting     = ref(false)

onMounted(async () => {
  try { detail.value = await api.virtualIPs.get(nsStore.active, name) }
  finally { loading.value = false }
})

async function doDelete() {
  deleting.value = true
  try {
    await api.virtualIPs.delete(nsStore.active, name)
    router.push('/virtualips')
  } finally {
    deleting.value = false
  }
}
</script>
