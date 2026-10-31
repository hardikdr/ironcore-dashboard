<template>
  <v-list-item
    :to="to"
    :prepend-icon="icon"
    :title="rail ? undefined : label"
    rounded="lg"
    active-color="primary"
    class="nav-item mb-1"
  >
    <template v-if="badge && !rail" #append>
      <v-chip size="x-small" color="primary" variant="tonal">{{ badge }}</v-chip>
    </template>
    <v-tooltip v-if="rail" :text="label" location="end" />
  </v-list-item>
</template>

<script setup lang="ts">
defineProps<{
  to: string
  icon: string
  label: string
  rail: boolean
  badge?: number
}>()
</script>

<style scoped>
.nav-item {
  min-height: 36px !important;
  font-size: 13px !important;
}
</style>
