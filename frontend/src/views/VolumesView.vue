<template>
  <v-container fluid class="pa-6">
    <PageHeader
      title="Volumes"
      subtitle="Block storage volumes"
      icon="mdi-database"
    >
      <template #actions>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/volumes/new' }">Create Volume</v-btn>
      </template>
    </PageHeader>

    <v-row dense class="mb-6">
      <v-col cols="auto">
        <StatCard label="Available" :value="availableCount" icon="mdi-check-circle-outline" color="success" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Pending" :value="pendingCount" icon="mdi-clock-outline" color="warning" />
      </v-col>
      <v-col cols="auto">
        <StatCard label="Total" :value="store.items.length" icon="mdi-database" color="primary" />
      </v-col>
    </v-row>

    <v-card v-if="store.items.length === 0 && !store.loading" :style="cardStyle" variant="flat" rounded="lg">
      <v-card-text class="text-center pa-12">
        <v-icon size="48" color="medium-emphasis" class="mb-4">mdi-database-off-outline</v-icon>
        <div class="text-h6 mb-2">No volumes yet</div>
        <div class="text-medium-emphasis mb-4">Create your first volume to get started.</div>
        <v-btn color="primary" prepend-icon="mdi-plus" :to="{ path: '/volumes/new' }">Create Volume</v-btn>
      </v-card-text>
    </v-card>

    <v-card v-else :style="cardStyle" variant="flat" rounded="lg">
      <div class="d-flex align-center ga-2 pa-4" :style="headerStyle">
        <v-icon color="primary">mdi-database</v-icon>
        <span class="font-weight-bold">Volumes</span>
        <v-chip size="x-small" color="primary">{{ filtered.length }}</v-chip>
        <v-spacer />
        <v-text-field
          v-model="search"
          prepend-inner-icon="mdi-magnify"
          placeholder="Search…"
          variant="outlined"
          density="compact"
          hide-details
          clearable
          style="max-width:240px"
        />
      </div>
      <v-data-table
        :headers="headers"
        :items="filtered"
        :loading="store.loading"
        density="compact"
        hover
        :no-data-text="store.error ?? 'No volumes'"
      >
        <template #item.name="{ item }">
          <router-link :to="`/volumes/${item.name}`" class="text-primary text-decoration-none font-weight-medium">
            {{ item.name }}
          </router-link>
        </template>
        <template #item.state="{ item }"><StatusBadge :state="item.state" /></template>
        <template #item.sizeBytes="{ item }">{{ formatBytes(item.sizeBytes) }}</template>
        <template #item.createdAt="{ item }">
          <span class="text-caption text-medium-emphasis">{{ relativeTime(item.createdAt) }}</span>
        </template>
        <template #item.actions="{ item }">
          <v-btn
            size="x-small" variant="outlined" icon="mdi-delete" color="error"
            :aria-label="`Delete volume ${item.name}`"
            title="Delete volume"
            @click="confirmDelete(item.name)"
          />
        </template>
      </v-data-table>
    </v-card>

    <v-dialog v-model="deleteDialog" max-width="400">
      <v-card>
        <v-card-title class="d-flex align-center ga-2">
          <v-icon color="error">mdi-alert-circle-outline</v-icon>
          Delete volume?
        </v-card-title>
        <v-card-text>Volume <strong>{{ deleteTarget }}</strong> will be permanently deleted from namespace <b>{{ nsStore.active }}</b>. This action cannot be undone.</v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn variant="outlined" :disabled="deleting" @click="deleteDialog = false">Cancel</v-btn>
          <v-btn color="error" :loading="deleting" @click="doDelete">Delete</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useVolumesStore }   from '@/stores/volumes'
import { useNamespaceStore } from '@/stores/namespace'
import { useCardStyle }      from '@/composables/useCardStyle'
import { relativeTime }      from '@/composables/useTime'
import PageHeader   from '@/components/common/PageHeader.vue'
import StatusBadge  from '@/components/common/StatusBadge.vue'
import StatCard     from '@/components/common/StatCard.vue'

const store   = useVolumesStore()
const nsStore = useNamespaceStore()
const { cardStyle, headerStyle } = useCardStyle()
const search       = ref('')
const deleteDialog = ref(false)
const deleteTarget = ref('')
const deleting     = ref(false)

const availableCount = computed(() => store.items.filter(v => v.state === 'Available').length)
const pendingCount   = computed(() => store.items.filter(v => v.state === 'Pending').length)

const filtered = computed(() => {
  const q = search.value.toLowerCase()
  return q ? store.items.filter(v => v.name.toLowerCase().includes(q) || v.volumeClass.toLowerCase().includes(q)) : store.items
})

const headers = [
  { title: 'Name',         key: 'name'        },
  { title: 'State',        key: 'state'       },
  { title: 'Size',         key: 'sizeBytes'   },
  { title: 'Volume Class', key: 'volumeClass' },
  { title: 'Created',      key: 'createdAt'   },
  { title: '',             key: 'actions', sortable: false },
]

function formatBytes(b: number) {
  if (!b) return '—'
  const gb = b / 1024 / 1024 / 1024
  return gb >= 1000 ? `${(gb / 1024).toFixed(1)} TB` : `${Math.round(gb)} GB`
}

function confirmDelete(name: string) { deleteTarget.value = name; deleteDialog.value = true }
async function doDelete() {
  deleting.value = true
  await store.deleteVolume(deleteTarget.value)
  deleting.value = false; deleteDialog.value = false
}

onMounted(() => store.load())
watch(() => nsStore.active, () => store.load())
</script>
