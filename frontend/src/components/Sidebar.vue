<template>
  <v-list density="compact" nav class="pt-2">
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Compute
    </div>
    <v-list-item to="/machines"      prepend-icon="mdi-monitor"        title="Machines"       rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Storage
    </div>
    <v-list-item to="/volumes"       prepend-icon="mdi-database"       title="Volumes"        rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      Networking
    </div>
    <v-list-item to="/networks"      prepend-icon="mdi-lan"            title="Networks"       rounded="lg" />
    <v-list-item to="/virtualips"    prepend-icon="mdi-earth"          title="Virtual IPs"    rounded="lg" />
    <v-list-item to="/loadbalancers" prepend-icon="mdi-scale-balance"  title="Load Balancers" rounded="lg" />

    <v-divider class="my-2" />
    <div class="text-caption text-uppercase text-medium-emphasis px-3 pt-2 pb-1 font-weight-bold">
      IP Management
    </div>
    <v-list-item to="/prefixes"      prepend-icon="mdi-ip-network"     title="IP Prefixes"    rounded="lg" />
  </v-list>
</template>
