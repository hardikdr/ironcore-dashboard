<template>
  <div style="position:relative; width:100%; height:100%">
    <Doughnut :data="chartData" :options="chartOptions" />
    <div
      v-if="centerLabel"
      style="position:absolute;top:50%;left:50%;transform:translate(-50%,-50%);text-align:center;pointer-events:none"
    >
      <div class="text-h5 font-weight-bold">{{ centerValue }}</div>
      <div class="text-caption text-medium-emphasis">{{ centerLabel }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Doughnut } from 'vue-chartjs'
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from 'chart.js'
import { useTheme } from 'vuetify'

ChartJS.register(ArcElement, Tooltip, Legend)

const props = defineProps<{
  segments: { label: string; value: number; color: string }[]
  centerLabel?: string
  centerValue?: number | string
}>()

const theme = useTheme()
const isDark = computed(() => theme.global.current.value.dark)

const chartData = computed(() => ({
  labels:   props.segments.map(s => s.label),
  datasets: [{
    data:            props.segments.map(s => s.value),
    backgroundColor: props.segments.map(s => s.color),
    borderWidth: 0,
    hoverOffset: 4,
  }]
}))

const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '72%',
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      backgroundColor: isDark.value ? '#1e293b' : '#fff',
      titleColor:      isDark.value ? '#e2e8f0' : '#1e293b',
      bodyColor:       isDark.value ? '#94a3b8' : '#64748b',
      borderColor:     isDark.value ? 'rgba(129,140,248,0.2)' : 'rgba(99,102,241,0.2)',
      borderWidth: 1,
    }
  }
}))
</script>
