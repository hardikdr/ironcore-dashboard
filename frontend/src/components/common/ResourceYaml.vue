<template>
  <v-card :style="cardStyle" variant="flat" class="pa-0">
    <div class="d-flex align-center justify-space-between pa-3" style="border-bottom:1px solid rgba(128,128,128,0.12)">
      <span class="text-caption text-medium-emphasis font-weight-medium">
        <v-icon size="14" class="mr-1">mdi-code-braces</v-icon>
        Raw YAML
      </span>
      <div class="d-flex ga-2">
        <v-btn
          size="x-small"
          variant="text"
          prepend-icon="mdi-content-copy"
          :color="copied ? 'success' : undefined"
          @click="copyYaml"
        >
          {{ copied ? 'Copied!' : 'Copy' }}
        </v-btn>
        <v-btn
          size="x-small"
          variant="text"
          prepend-icon="mdi-download"
          @click="downloadYaml"
        >
          Download
        </v-btn>
      </div>
    </div>
    <pre
      class="text-caption pa-4"
      style="overflow-x:auto; overflow-y:auto; max-height:500px; white-space:pre; font-family:monospace; line-height:1.5;"
    >{{ yamlText }}</pre>
  </v-card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useCardStyle } from '@/composables/useCardStyle'

const props = defineProps<{
  resource: Record<string, unknown>
  filename?: string
}>()

const { cardStyle } = useCardStyle()
const copied = ref(false)

const yamlText = computed(() => toYaml(props.resource, 0))

function toYaml(obj: unknown, indent: number): string {
  if (obj === null || obj === undefined) return 'null'
  if (typeof obj === 'string') {
    if (obj.includes('\n') || obj.includes(':') || obj.includes('#') || obj === '') {
      return `"${obj.replace(/"/g, '\\"')}"`
    }
    return obj
  }
  if (typeof obj === 'number' || typeof obj === 'boolean') return String(obj)
  if (Array.isArray(obj)) {
    if (obj.length === 0) return '[]'
    const pad = ' '.repeat(indent)
    return '\n' + obj.map(item => `${pad}- ${toYaml(item, indent + 2)}`).join('\n')
  }
  if (typeof obj === 'object') {
    const record = obj as Record<string, unknown>
    const keys = Object.keys(record)
    if (keys.length === 0) return '{}'
    const pad = ' '.repeat(indent)
    return keys.map(k => {
      const val = record[k]
      if (typeof val === 'object' && val !== null && !Array.isArray(val)) {
        const nested = toYaml(val, indent + 2)
        if (nested === '{}') return `${pad}${k}: {}`
        return `${pad}${k}:\n${nested}`
      }
      const rendered = toYaml(val, indent + 2)
      if (rendered.startsWith('\n')) return `${pad}${k}:${rendered}`
      return `${pad}${k}: ${rendered}`
    }).join('\n')
  }
  return String(obj)
}

async function copyYaml() {
  try {
    await navigator.clipboard.writeText(yamlText.value)
    copied.value = true
    setTimeout(() => { copied.value = false }, 2000)
  } catch {
    // clipboard not available in all contexts
  }
}

function downloadYaml() {
  const blob = new Blob([yamlText.value], { type: 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = (props.filename || 'resource') + '.yaml'
  a.click()
  URL.revokeObjectURL(url)
}
</script>
