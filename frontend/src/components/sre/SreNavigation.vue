<template><nav class="sre-nav" aria-label="SRE console">
  <router-link v-for="item in items" :key="item.to" :to="item.to" class="sre-nav__item"><v-icon :icon="item.icon" size="18"/><span>{{ item.label }}</span><span v-if="item.badge" class="sre-nav__badge">{{ item.badge }}</span></router-link>
</nav></template>
<script setup lang="ts">
import { computed } from 'vue'
import { useSreLandscapeStore } from '@/stores/sre/landscape'
defineProps<{ rail: boolean }>()
const landscape=useSreLandscapeStore()
const items=computed(()=>[
  {to:'/sre/landscape',icon:'mdi-view-grid-outline',label:'Landscape'},
  {to:'/sre/zones',icon:'mdi-rhombus-outline',label:'Zones'},
  {to:'/sre/shoots',icon:'mdi-circle-double',label:'Shoots'},
  {to:'/sre/vms',icon:'mdi-rhombus-medium-outline',label:'IronCore VMs'},
  {to:'/sre/networking',icon:'mdi-swap-horizontal',label:'Networking'},
  {to:'/sre/storage',icon:'mdi-tray-full',label:'Storage'},
  {to:'/sre/triage',icon:'mdi-alert-outline',label:'Alerts',badge:landscape.observations.length||undefined},
])
</script>
<style scoped>.sre-nav{gap:4px!important;padding:0 12px!important}.sre-nav__item{gap:13px!important;min-height:42px!important;padding:0 14px!important;border-radius:7px!important;font-size:13px!important}.sre-nav__badge{width:23px!important;height:23px!important;font-size:11px}</style>
<style scoped>.sre-nav{display:flex;flex-direction:column;gap:7px;padding:0 18px}.sre-nav__item{display:flex;align-items:center;gap:16px;min-height:54px;padding:0 18px;border-radius:10px;color:var(--ic-text-muted);text-decoration:none;font-size:16px;transition:background .15s,color .15s}.sre-nav__item:hover{background:rgba(156,171,255,.06);color:var(--ic-text)}.sre-nav__item.router-link-active{background:#2d294f;color:#acb8ff;font-weight:650}.sre-nav__badge{display:grid;place-items:center;margin-left:auto;width:30px;height:30px;border-radius:50%;background:#ff7b8b;color:#1b1115;font-weight:700}</style>
