import { ref, computed, type Ref } from 'vue'
import type { RemoteMachine } from '@/types'

export interface MachineGroup<T> {
  machineId: string
  machineName: string
  items: T[]
}

export function useMachineFilter<T extends Record<string, unknown>>(
  items: Ref<T[]>,
  machines: Ref<RemoteMachine[]>,
  getMachineId: (item: T) => string,
  getMachineName: (item: T) => string,
) {
  const machineFilter = ref('')

  const machineOptions = computed(() => {
    const unique = new Map<string, string>()
    for (const m of machines.value) {
      unique.set(m.id, m.name)
    }
    for (const item of items.value) {
      const id = getMachineId(item)
      if (id && !unique.has(id)) {
        unique.set(id, getMachineName(item))
      }
    }
    return Array.from(unique.entries()).map(([id, name]) => ({
      value: id,
      label: name,
    }))
  })

  const groupedItems = computed(() => {
    const filtered = machineFilter.value
      ? items.value.filter((item) => getMachineId(item) === machineFilter.value)
      : items.value

    const groupMap = new Map<string, MachineGroup<T>>()
    for (const item of filtered) {
      const id = getMachineId(item)
      const name = getMachineName(item)
      if (!groupMap.has(id)) {
        groupMap.set(id, { machineId: id, machineName: name, items: [] })
      }
      groupMap.get(id)!.items.push(item)
    }
    return Array.from(groupMap.values()).sort((a, b) =>
      a.machineName.localeCompare(b.machineName),
    )
  })

  return {
    machineFilter,
    machineOptions,
    groupedItems,
  }
}
