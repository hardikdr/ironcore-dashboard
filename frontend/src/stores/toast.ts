import { defineStore } from 'pinia'
import { ref } from 'vue'

export interface Toast {
  id: number
  message: string
  color: 'error' | 'warning' | 'success' | 'info'
  timeout: number
}

let nextId = 1

export const useToastStore = defineStore('toast', () => {
  const toasts = ref<Toast[]>([])

  function show(message: string, color: Toast['color'] = 'info', timeout = 4000) {
    const id = nextId++
    toasts.value.push({ id, message, color, timeout })
    setTimeout(() => dismiss(id), timeout + 200)
  }

  function dismiss(id: number) {
    const idx = toasts.value.findIndex(t => t.id === id)
    if (idx !== -1) toasts.value.splice(idx, 1)
  }

  const error   = (msg: string) => show(msg, 'error', 6000)
  const success = (msg: string) => show(msg, 'success', 3000)
  const warn    = (msg: string) => show(msg, 'warning', 4000)
  const info    = (msg: string) => show(msg, 'info', 3000)

  return { toasts, dismiss, error, success, warn, info }
})
