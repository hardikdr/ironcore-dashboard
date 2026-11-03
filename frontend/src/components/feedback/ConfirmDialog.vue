<template><v-dialog :model-value="modelValue" max-width="480" :persistent="loading" @update:model-value="!loading && $emit('update:modelValue', $event)"><v-card><v-card-title>{{ title }}</v-card-title><v-card-text><p>{{ message }}</p><p class="font-weight-medium mt-3">{{ resourceName }}</p><v-text-field v-if="requireText" v-model="confirmation" :label="`Type ${requireText} to confirm`" class="mt-4"/><v-alert v-if="error" type="error" variant="tonal" class="mt-4">{{ error }}</v-alert></v-card-text><v-card-actions><v-spacer/><v-btn :disabled="loading" @click="$emit('update:modelValue', false)">Cancel</v-btn><v-btn color="error" variant="flat" :loading="loading" :disabled="!!requireText && confirmation !== requireText" @click="$emit('confirm')">Confirm</v-btn></v-card-actions></v-card></v-dialog></template>
<script setup lang="ts">
import { ref, watch } from 'vue'
const props = withDefaults(defineProps<{ modelValue:boolean; title:string; resourceName:string; message:string; loading?:boolean; error?:string|null; requireText?:string }>(), { loading:false })
defineEmits<{ 'update:modelValue':[value:boolean]; confirm:[] }>()
const confirmation = ref('')
watch(() => props.modelValue, value => { if (!value) confirmation.value = '' })
</script>
