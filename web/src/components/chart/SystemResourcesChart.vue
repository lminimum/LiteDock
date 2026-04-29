<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

import type { TooltipItem } from 'chart.js'

interface Props {
  labels: string[]
  cpu: number[]
  memory: number[]
  disk: number[]
}

const props = defineProps<Props>()

const getCSSColor = (varName: string): string => {
  return getComputedStyle(document.documentElement).getPropertyValue(varName).trim()
}

const hexToRgba = (hex: string, alpha: number): string => {
  if (!hex.startsWith('#')) return hex
  const r = parseInt(hex.slice(1, 3), 16)
  const g = parseInt(hex.slice(3, 5), 16)
  const b = parseInt(hex.slice(5, 7), 16)
  return `rgba(${r}, ${g}, ${b}, ${alpha})`
}

const getColor = (varName: string, alpha = 1) => hexToRgba(getCSSColor(varName), alpha)

const isMobile = ref(window.innerWidth <= 768)
const handleResize = () => { isMobile.value = window.innerWidth <= 768 }
onMounted(() => window.addEventListener('resize', handleResize))
onUnmounted(() => window.removeEventListener('resize', handleResize))

interface DatasetConfig {
  label: string
  data: number[]
  colorVar: string
}

const datasets: DatasetConfig[] = [
  { label: 'CPU %', data: props.cpu, colorVar: '--color-info' },
  { label: 'Memory %', data: props.memory, colorVar: '--color-warning' },
  { label: 'Disk %', data: props.disk, colorVar: '--color-success' }
]

const data = computed(() => ({
  labels: props.labels,
  datasets: datasets.map(d => ({
    label: d.label,
    data: d.data,
    borderColor: getColor(d.colorVar),
    backgroundColor: getColor(d.colorVar, 0.1),
    borderWidth: 2,
    tension: 0.4,
    pointRadius: 0,
    pointHoverRadius: 4,
    fill: true
  }))
}))

const options = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  scales: {
    x: {
      grid: { display: false },
      ticks: {
        maxTicksLimit: isMobile.value ? 4 : 6,
        color: getCSSColor('--color-text-weak')
      }
    },
    y: {
      display: false,
      min: 0,
      max: 100
    }
  },
  plugins: {
    legend: {
      position: (isMobile.value ? 'bottom' : 'top') as 'bottom' | 'top',
      align: (isMobile.value ? 'center' : 'end') as 'center' | 'end',
      labels: {
        boxWidth: 12,
        padding: 16,
        color: getCSSColor('--color-text-weak'),
        usePointStyle: true,
        pointStyle: 'circle'
      }
    },
    tooltip: {
      callbacks: {
        label: (context: TooltipItem<'line'>) => `${context.dataset.label}: ${context.parsed.y}%`
      }
    }
  }
}))
</script>

<template>
  <div class="chart-container">
    <Line :data="data" :options="options" />
  </div>
</template>

<style scoped>
.chart-container {
  height: 250px;
  width: 100%;
}

@media (max-width: 1024px) {
  .chart-container {
    height: 220px;
  }
}

@media (max-width: 768px) {
  .chart-container {
    height: 180px;
  }
}
</style>
