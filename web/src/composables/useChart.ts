import { ref } from 'vue'

interface ChartDataPoint {
  cpu: number
  memory: number
  disk: number
  time: string
}

const DEFAULT_TOTAL_POINTS = 60
const DEFAULT_UPDATE_INTERVAL = 2000

export function useChart(
  totalPoints = DEFAULT_TOTAL_POINTS,
  updateInterval = DEFAULT_UPDATE_INTERVAL
) {
  const labels = ref<string[]>([])
  const cpu = ref<number[]>(new Array(totalPoints).fill(0))
  const memory = ref<number[]>(new Array(totalPoints).fill(0))
  const disk = ref<number[]>(new Array(totalPoints).fill(0))

  const initLabels = () => {
    const result: string[] = []
    const now = Date.now()
    for (let i = totalPoints - 1; i >= 0; i--) {
      const t = new Date(now - i * updateInterval)
      result.push(
        t.toLocaleString('zh-CN', {
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
          hour12: false
        })
      )
    }
    labels.value = result
  }

  const loadHistory = async (apiGet: (url: string) => Promise<{ data: { success: boolean; data: ChartDataPoint[] } }>): Promise<boolean> => {
    try {
      const res = await apiGet('/dashboard/resources/history?minutes=5')
      if (res.data.success && res.data.data.length > 0) {
        const history = res.data.data.slice(-totalPoints)
        const cpuArr: number[] = []
        const memoryArr: number[] = []
        const diskArr: number[] = []
        const labelArr: string[] = []

        for (const m of history) {
          cpuArr.push(m.cpu ?? 0)
          memoryArr.push(m.memory ?? 0)
          diskArr.push(m.disk ?? 0)
          labelArr.push(m.time ?? '')
        }

        const fillCount = totalPoints - cpuArr.length
        cpu.value = [...new Array(fillCount).fill(0), ...cpuArr]
        memory.value = [...new Array(fillCount).fill(0), ...memoryArr]
        disk.value = [...new Array(fillCount).fill(0), ...diskArr]
        labels.value = [...new Array(fillCount).fill(''), ...labelArr]
        return true
      }
      return false
    } catch (e) {
      console.error('Failed to load history:', e)
      return false
    }
  }

  const addDataPoint = (point: ChartDataPoint) => {
    const shift = <T>(arr: T[], val: T) => [...arr.slice(1), val]
    
    cpu.value = shift(cpu.value, point.cpu ?? 0)
    memory.value = shift(memory.value, point.memory ?? 0)
    disk.value = shift(disk.value, point.disk ?? 0)
    labels.value = shift(labels.value, point.time ?? '')
  }

  return {
    labels,
    cpu,
    memory,
    disk,
    initLabels,
    loadHistory,
    addDataPoint
  }
}
