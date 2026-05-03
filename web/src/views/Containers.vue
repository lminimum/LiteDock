<template>
  <div class="containers-page">
    <PageHeader :title="t('containers.title')">
      <button @click="refreshContainers" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('containers.refresh') }}
      </button>
      <button @click="showCreateModal = true" class="btn btn-primary">
        <Plus :size="16" />
        {{ t('containers.create') }}
      </button>
    </PageHeader>

    <div class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('containers.searchPlaceholder')"
          type="text"
          class="input"
        />
      </div>
      <div class="filter-options">
        <select v-model="statusFilter" class="input">
          <option value="">{{ t('containers.allStatuses') }}</option>
          <option value="running">{{ t('containers.running') }}</option>
          <option value="stopped">{{ t('containers.stopped') }}</option>
          <option value="paused">{{ t('containers.paused') }}</option>
          <option value="restarting">{{ t('containers.restarting') }}</option>
          <option value="exited">{{ t('containers.exited') }}</option>
          <option value="created">{{ t('containers.created') }}</option>
        </select>
      </div>
    </div>

    <div class="card-grid">
      <ContainerCard
        v-for="container in filteredContainers"
        :key="container.id"
        :container="container"
        @inspect="handleInspect"
        @delete="deleteContainer"
        @start="startContainer"
        @stop="stopContainer"
        @restart="restartContainer"
        @logs="showLogs"
      />
    </div>

    <InspectModal
      :visible="showInspect"
      title="Container Details"
      :fields="inspectFields"
      @close="showInspect = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus } from 'lucide-vue-next'
import { t } from '@/i18n'
import api from '@/utils/api'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import ContainerCard from '@/components/container/ContainerCard.vue'
import InspectModal from '@/components/ui/InspectModal.vue'

interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'restarting' | 'exited' | 'created'
  ports: string[]
  createdAt: string
  machine: string
  machineId: string
}

const loading = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const showCreateModal = ref(false)
const machines = ref<RemoteMachine[]>([])

const containers = ref<Container[]>([])

const showInspect = ref(false)
const selectedContainer = ref<Container | null>(null)

const inspectFields = computed(() => {
  const c = selectedContainer.value
  if (!c) return []
  return [
    { label: 'ID', value: c.id },
    { label: 'Name', value: c.name },
    { label: 'Image', value: c.image },
    { label: 'Status', value: c.status },
    { label: 'Ports', value: (c.ports || []).join(', ') || '-' },
    { label: 'Created At', value: c.createdAt || '-' },
    { label: 'Machine', value: c.machine },
  ]
})

const handleInspect = (id: string) => {
  const c = containers.value.find(con => con.id === id)
  if (c) {
    selectedContainer.value = c
    showInspect.value = true
  }
}

const refreshContainers = async () => {
  loading.value = true
  try {
    // Fetch local containers from /v1/containers
    const localData: any = await api.get('/containers')
    const localContainers: Container[] = (localData?.containers || []).map((c: any) => ({
      id: c.id,
      name: c.name,
      image: c.image,
      status: c.status,
      ports: c.ports || [],
      createdAt: c.created_at || c.cached_at,
      machine: 'Local',
      machineId: 'local'
    }))

    // Fetch remote machine containers
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const remoteContainers: Container[] = []
    await Promise.all(
      allMachines.map(async (m) => {
        try {
          const containers = await remoteMachineService.listContainers(m.id)
          for (const c of containers) {
            remoteContainers.push({ ...c, machine: m.name, machineId: m.id })
          }
        } catch (e) {
          // machine offline or unreachable
        }
      })
    )

    // Merge and sort
    const allContainers = [...localContainers, ...remoteContainers]
    allContainers.sort((a, b) => {
      if (a.machine !== b.machine) return a.machine.localeCompare(b.machine)
      return a.name.localeCompare(b.name)
    })

    containers.value = allContainers
  } catch (e) {
    console.error('Failed to load containers:', e)
  } finally {
    loading.value = false
  }
}

const filteredContainers = computed(() => {
  let filtered = containers.value

  if (searchQuery.value) {
    filtered = filtered.filter(container =>
      container.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      container.image.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      container.machine.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }

  if (statusFilter.value) {
    filtered = filtered.filter(container => container.status === statusFilter.value)
  }

  return filtered
})


const startContainer = async (id: string) => {
  const container = containers.value.find(c => c.id === id)
  if (container) container.status = 'running'
}

const stopContainer = async (id: string) => {
  const container = containers.value.find(c => c.id === id)
  if (container) container.status = 'stopped'
}

const restartContainer = async (id: string) => {
  console.log('restart container:', id)
}

const showLogs = (id: string) => {
  console.log('show logs:', id)
}

const deleteContainer = async (id: string) => {
  if (confirm(t('containers.confirmDelete'))) {
    containers.value = containers.value.filter(c => c.id !== id)
  }
}

onMounted(() => refreshContainers())
</script>

<style scoped>
.containers-page {
  max-width: 1400px;
  margin: 0 auto;
}

.filters {
  display: flex;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
  padding: var(--space-4);
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
}

.search-box {
  flex: 1;
}

.filter-options select {
  min-width: 140px;
}

@media (max-width: 768px) {
  .filters {
    flex-direction: column;
  }

  .filter-options select {
    min-width: 100%;
  }
}
</style>
