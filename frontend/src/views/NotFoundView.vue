<template>
  <v-container fluid class="pa-6 d-flex align-center justify-center" style="min-height: 70vh">
    <div class="text-center">
      <div class="text-h1 font-weight-bold mb-2" style="opacity:0.12">404</div>
      <v-icon size="64" color="medium-emphasis" class="mb-4">mdi-map-marker-question-outline</v-icon>
      <div class="text-h5 font-weight-bold mb-2">Page not found</div>
      <div class="text-medium-emphasis mb-6">The page you're looking for doesn't exist.</div>
      <v-btn color="primary" prepend-icon="mdi-arrow-left" to="/overview">Back to Overview</v-btn>
    </div>
  </v-container>
</template>
