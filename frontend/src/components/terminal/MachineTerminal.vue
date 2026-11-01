<template>
  <div ref="container" class="terminal-container">
    <!-- Console unavailable banner (real mode) -->
    <div v-if="consoleError" class="console-unavailable d-flex flex-column align-center justify-center pa-8 text-center">
      <v-icon size="48" color="warning" class="mb-4">mdi-console-line</v-icon>
      <div class="text-subtitle-1 font-weight-bold mb-2">Console Unavailable</div>
      <div class="text-body-2 text-medium-emphasis" style="max-width:420px">{{ consoleError }}</div>
    </div>

    <template v-else>
      <!-- Connection status bar -->
      <div class="term-statusbar d-flex align-center ga-2 px-3 py-1">
        <v-icon :color="currentColor" size="10">mdi-circle</v-icon>
        <span class="text-caption" :style="{ color: currentColor }">{{ currentLabel }}</span>
        <v-spacer />
        <v-btn
          v-if="showReconnect"
          size="x-small"
          variant="tonal"
          color="primary"
          prepend-icon="mdi-refresh"
          @click="manualReconnect"
        >Reconnect</v-btn>
      </div>
      <div ref="termEl" class="term-body" />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useTheme } from 'vuetify'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{ wsUrl: string; active?: boolean }>()

const container = ref<HTMLElement | null>(null)
const termEl    = ref<HTMLElement | null>(null)
const theme     = useTheme()

type ConnStatus = 'connecting' | 'connected' | 'disconnected' | 'error'
const status = ref<ConnStatus>('disconnected')
const STATUS_COLOR: Record<ConnStatus, string> = { connecting: '#fbbf24', connected: '#34d399', disconnected: '#94a3b8', error: '#f87171' }
const STATUS_LABEL: Record<ConnStatus, string> = { connecting: 'Connecting…', connected: 'Connected', disconnected: 'Disconnected', error: 'Connection failed' }
const currentColor = computed(() => STATUS_COLOR[status.value])
const currentLabel = computed(() => STATUS_LABEL[status.value])

const showReconnect = ref(false)
const consoleError  = ref<string | null>(null)

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null
let resizeObserver: ResizeObserver | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let reconnectAttempts = 0
const MAX_RECONNECT = 3

function buildTheme() {
  const dark = theme.global.current.value.dark
  return dark
    ? { background: '#0d1117', foreground: '#c9d1d9', cursor: '#58a6ff', selectionBackground: '#264f78' }
    : { background: '#ffffff', foreground: '#1f2328', cursor: '#0969da', selectionBackground: '#b6d4f8' }
}

function initTerminal() {
  if (!termEl.value || terminal) return
  terminal = new Terminal({
    fontFamily: 'JetBrains Mono, Fira Code, Cascadia Code, monospace',
    fontSize: 13,
    lineHeight: 1.4,
    cursorBlink: true,
    theme: buildTheme(),
    allowTransparency: true,
  })
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())
  terminal.open(termEl.value)
  fitAddon.fit()
  terminal.onData((data: string) => {
    if (ws?.readyState === WebSocket.OPEN) ws.send(data)
  })
  resizeObserver = new ResizeObserver(() => { fitAddon?.fit(); sendResize() })
  resizeObserver.observe(termEl.value)
}

function connect() {
  if (!terminal) initTerminal()
  status.value = 'connecting'
  showReconnect.value = false
  ws = new WebSocket(props.wsUrl)

  ws.onopen = () => {
    status.value = 'connected'
    reconnectAttempts = 0
    sendResize()
  }

  ws.onmessage = (ev) => {
    if (typeof ev.data === 'string') {
      try {
        const frame = JSON.parse(ev.data) as { type?: string; message?: string }
        if (frame.type === 'error') {
          consoleError.value = frame.message ?? 'Console unavailable.'
          disconnect()
          return
        }
      } catch { /* not JSON — pass through to terminal */ }
    }
    terminal?.write(typeof ev.data === 'string' ? ev.data : new Uint8Array(ev.data))
  }

  ws.onerror = () => {
    terminal?.write('\r\n\x1b[31mWebSocket error — could not connect to console.\x1b[0m\r\n')
    status.value = 'error'
  }

  ws.onclose = () => {
    status.value = 'disconnected'
    terminal?.write('\r\n\x1b[90m[Connection closed]\x1b[0m\r\n')
    scheduleReconnect()
  }
}

function scheduleReconnect() {
  if (reconnectAttempts >= MAX_RECONNECT) {
    showReconnect.value = true
    return
  }
  const delay = [1000, 2000, 4000][reconnectAttempts] ?? 4000
  reconnectAttempts++
  terminal?.write(`\x1b[90m[Reconnecting in ${delay / 1000}s… attempt ${reconnectAttempts}/${MAX_RECONNECT}]\x1b[0m\r\n`)
  reconnectTimer = setTimeout(() => connect(), delay)
}

function manualReconnect() {
  reconnectAttempts = 0
  showReconnect.value = false
  connect()
}

function sendResize() {
  if (ws?.readyState === WebSocket.OPEN && terminal) {
    ws.send(JSON.stringify({ type: 'resize', cols: terminal.cols, rows: terminal.rows }))
  }
}

function disconnect() {
  if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null }
  resizeObserver?.disconnect()
  ws?.close()
  ws = null
}

function dispose() {
  disconnect()
  terminal?.dispose()
  terminal = null
  fitAddon = null
}

// Only connect when active prop turns true (lazy activation)
watch(() => props.active, (active) => {
  if (active) connect()
  else disconnect()
}, { immediate: false })

watch(() => props.wsUrl, () => {
  disconnect()
  if (props.active !== false) connect()
})

onMounted(() => {
  if (props.active !== false) connect()
})

onUnmounted(dispose)
</script>

<style scoped>
.terminal-container {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  min-height: 440px;
  background: #0d1117;
  border-radius: 8px;
  overflow: hidden;
}
.console-unavailable {
  flex: 1;
  background: #0d1117;
  color: #c9d1d9;
}
.term-statusbar {
  background: rgba(0,0,0,0.4);
  border-bottom: 1px solid rgba(255,255,255,0.06);
  min-height: 28px;
  flex-shrink: 0;
}
.term-body {
  flex: 1;
  padding: 8px;
  overflow: hidden;
}
</style>
