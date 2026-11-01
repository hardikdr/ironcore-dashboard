import { ref, computed } from 'vue'

export function useSortable<T extends Record<string, unknown>>(items: () => T[]) {
  const sortKey = ref<string | null>(null)
  const sortDir = ref<'asc' | 'desc'>('asc')

  function toggleSort(key: string) {
    if (sortKey.value === key) {
      sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
    } else {
      sortKey.value = key
      sortDir.value = 'asc'
    }
  }

  const sorted = computed(() => {
    const list = [...items()]
    if (!sortKey.value) return list
    const key = sortKey.value
    const dir = sortDir.value === 'asc' ? 1 : -1
    return list.sort((a, b) => {
      const av = a[key] ?? ''
      const bv = b[key] ?? ''
      if (av < bv) return -dir
      if (av > bv) return dir
      return 0
    })
  })

  return { sortKey, sortDir, sorted, toggleSort }
}
