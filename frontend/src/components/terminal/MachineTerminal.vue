<template>
  <div ref="container" class="terminal-container" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useTheme } from 'vuetify'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{ wsUrl: string }>()

const container = ref<HTMLElement | null>(null)
const theme = useTheme()

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null
let resizeObserver: ResizeObserver | null = null

function buildTheme() {
  const dark = theme.global.current.value.dark
  return dark
    ? { background: '#0d1117', foreground: '#c9d1d9', cursor: '#58a6ff', selectionBackground: '#264f78' }
    : { background: '#ffffff', foreground: '#1f2328', cursor: '#0969da', selectionBackground: '#b6d4f8' }
}

function connect() {
  if (!container.value) return

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
  terminal.open(container.value)
  fitAddon.fit()

  ws = new WebSocket(props.wsUrl)

  ws.onopen = () => { sendResize() }

  ws.onmessage = (ev) => {
    terminal?.write(typeof ev.data === 'string' ? ev.data : new Uint8Array(ev.data))
  }

  ws.onerror = () => {
    terminal?.write('\r\n\x1b[31mWebSocket error — could not connect to console.\x1b[0m\r\n')
  }

  ws.onclose = () => {
    terminal?.write('\r\n\x1b[90m[Connection closed]\x1b[0m\r\n')
  }

  terminal.onData((data: string) => {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(data)
    }
  })

  resizeObserver = new ResizeObserver(() => {
    fitAddon?.fit()
    sendResize()
  })
  resizeObserver.observe(container.value)
}

function sendResize() {
  if (ws?.readyState === WebSocket.OPEN && terminal) {
    ws.send(JSON.stringify({ type: 'resize', cols: terminal.cols, rows: terminal.rows }))
  }
}

function disconnect() {
  resizeObserver?.disconnect()
  ws?.close()
  terminal?.dispose()
  terminal = null
  fitAddon = null
  ws = null
}

watch(() => props.wsUrl, () => {
  disconnect()
  connect()
})

onMounted(connect)
onUnmounted(disconnect)
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100%;
  min-height: 400px;
  background: #0d1117;
  border-radius: 8px;
  overflow: hidden;
}
</style>
