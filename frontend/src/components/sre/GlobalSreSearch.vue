<template><div><v-text-field v-model="query" label="Find a VM by name, UID, IP, provider ID, node or Shoot" prepend-inner-icon="mdi-magnify" clearable :loading="search.loading" :error-messages="search.error || undefined"/><v-list v-if="search.results.length" variant="outlined" rounded="lg"><v-list-item v-for="item in search.results" :key="item.id" :to="routeFor(item.resource.provenance.kind,item.id)" :title="item.resource.provenance.name" :subtitle="`${item.resource.provenance.kind} · ${item.matchedField} ${item.matchKind}`"><template #append><v-chip size="x-small" variant="tonal">{{ item.fragment }}</v-chip></template></v-list-item></v-list><p v-else-if="query.trim().length>=2&&!search.loading" class="empty">No matching operational resource was found.</p></div></template>
<script setup lang="ts">
import { onUnmounted,ref,watch } from 'vue'
import { useSreSearchStore } from '@/stores/sre/search'
const search=useSreSearchStore(),query=ref('')
let timer:ReturnType<typeof setTimeout>|undefined,controller:AbortController|undefined
watch(query,value=>{if(timer)clearTimeout(timer);controller?.abort();if(value.trim().length<2){search.clear();return}timer=setTimeout(()=>{controller=new AbortController();void search.run(value,controller.signal)},180)})
onUnmounted(()=>{if(timer)clearTimeout(timer);controller?.abort()})
function routeFor(kind:string,id:string){const value=kind.toLowerCase();if(value.includes('vm'))return`/sre/vms/${id}`;if(value==='shoot')return`/sre/shoots/${id}`;if(value.includes('node'))return`/sre/nodes/${id}`;return`/sre/dependencies/${id}`}
</script>
<style scoped>.empty{color:var(--ic-text-muted);font-size:13px;padding:12px}</style>
