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

    <CollapsibleFilters
      v-if="!loading && !error && containers.length > 0"
      v-model="searchQuery"
      :search-placeholder="t('containers.searchPlaceholder')"
      search-label="Search"
      filter-label="Filters"
      :has-filters="true"
    >
      <template #filters>
        <select v-model="statusFilter" class="input">
          <option value="">{{ t('containers.allStatuses') }}</option>
          <option value="running">{{ t('containers.running') }}</option>
          <option value="stopped">{{ t('containers.stopped') }}</option>
          <option value="paused">{{ t('containers.paused') }}</option>
          <option value="restarting">{{ t('containers.restarting') }}</option>
          <option value="exited">{{ t('containers.exited') }}</option>
          <option value="created">{{ t('containers.created') }}</option>
        </select>
        <select v-model="machineFilter" class="input">
          <option value="">{{ t('common.allMachines') }}</option>
          <option v-for="opt in machineOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
        </select>
      </template>
      <template #right>
        <ViewToggle v-model="viewMode" />
      </template>
    </CollapsibleFilters>

    <div v-if="loading" class="loading-state">
      <RefreshCw :size="24" class="spinning" />
      <span>{{ t('containers.refresh') }}...</span>
    </div>

    <div v-else-if="error" class="error-state card text-center">
      <p class="mb-4">{{ error }}</p>
      <button @click="refreshContainers" class="btn btn-secondary">{{ t('common.refresh') }}</button>
    </div>

    <div v-else-if="containers.length === 0" class="empty-state card text-center">
      <p class="mb-4">{{ t('containers.noContainers') }}</p>
    </div>

    <div v-else-if="filteredContainers.length === 0" class="empty-state card text-center">
      <p>{{ t('containers.noContainers') }}</p>
    </div>

    <Transition name="view-fade" mode="out-in">
      <div v-if="viewMode === 'card'" key="card">
        <template v-for="group in groupedItems" :key="group.machineId">
          <div class="machine-section-header">
            <Server :size="16" class="icon" />
            {{ group.machineName }}
            <span class="count">{{ group.items.length }} {{ t('common.containers') }}</span>
          </div>
          <div class="card-grid">
            <ContainerCard
              v-for="container in group.items"
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
        </template>
      </div>

      <div v-else key="list">
        <template v-for="group in groupedItems" :key="group.machineId">
          <div class="machine-section-header">
            <Server :size="16" class="icon" />
            {{ group.machineName }}
            <span class="count">{{ group.items.length }} {{ t('common.containers') }}</span>
          </div>
          <div class="item-list">
            <div v-for="container in group.items" :key="container.id" class="item-list-row">
              <div class="item-list-info">
                <div class="item-list-title">{{ container.name }}</div>
                <div class="item-list-meta">
                  <span class="text-muted">Image: {{ container.image }}</span>
                  <span class="badge" :class="container.status === 'running' ? 'badge-success' : container.status === 'stopped' || container.status === 'exited' ? 'badge-error' : 'badge-warning'">{{ container.status }}</span>
                  <span>{{ container.machine }}</span>
                </div>
              </div>
              <div class="item-list-actions">
                <button @click="handleInspect(container.id)" class="btn btn-sm btn-ghost">
                  <Info :size="14" /> {{ t('common.details') }}
                </button>
                <button v-if="container.status === 'stopped'" @click="startContainer(container.id)" class="btn btn-sm btn-ghost">
                  <Play :size="14" /> {{ t('common.start') }}
                </button>
                <button v-if="container.status === 'running'" @click="stopContainer(container.id)" class="btn btn-sm btn-ghost">
                  <Square :size="14" /> {{ t('common.stop') }}
                </button>
                <button @click="showLogs(container.id)" class="btn btn-sm btn-ghost">
                  <FileText :size="14" /> {{ t('common.logs') }}
                </button>
                <button @click="deleteContainer(container.id)" class="btn btn-sm btn-ghost btn-danger-text">
                  <Trash2 :size="14" /> {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div></Transition>

    <InspectModal
      :visible="showInspect"
      title="Container Details"
      :fields="inspectFields"
      @close="showInspect = false"
    />

    <ContainerCreateModal
      :visible="showCreateModal"
      :machine-id="defaultCreateMachineId"
      :machines="createMachineOptions"
      @close="showCreateModal = false"
      @created="onContainerCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus, Server, Info, Play, Square, FileText, Trash2 } from 'lucide-vue-next'
import { t } from '@/i18n'
import api from '@/utils/api'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import CollapsibleFilters from '@/components/ui/CollapsibleFilters.vue'
import ContainerCard from '@/components/container/ContainerCard.vue'
import ContainerCreateModal from '@/components/container/ContainerCreateModal.vue'
import InspectModal from '@/components/ui/InspectModal.vue'
import { useViewMode } from '@/composables/useViewMode'
import { useMachineFilter } from '@/composables/useMachineFilter'

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

interface ContainerApiItem {
  id: string
  name: string
  image: string
  status: Container['status']
  ports?: string[] | null
  createdAt?: string
  created_at?: string
  cached_at?: string
  machineId?: string
  machine_id?: string
}

interface ContainerListResponse {
  containers?: ContainerApiItem[]
}

const LOCAL_MACHINE_ID = 'local'
const LOCAL_MACHINE_NAME = 'Local'

const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
const statusFilter = ref('')
const viewMode = useViewMode('containers')
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

const getMachineName = (machineId: string, machineNames: Map<string, string>) => {
  const name = machineNames.get(machineId)
  if (name) return name
  return machineId === LOCAL_MACHINE_ID ? LOCAL_MACHINE_NAME : machineId
}

const normalizeContainer = (
  container: ContainerApiItem,
  machineId: string,
  machineName: string,
): Container => ({
  id: container.id,
  name: container.name,
  image: container.image,
  status: container.status,
  ports: Array.isArray(container.ports) ? container.ports : [],
  createdAt: container.createdAt || container.created_at || container.cached_at || '',
  machine: machineName,
  machineId,
})

const refreshContainers = async () => {
  loading.value = true
  error.value = ''
  try {
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const machineNames = new Map(allMachines.map((m) => [m.id, m.name]))
    machineNames.set(LOCAL_MACHINE_ID, machineNames.get(LOCAL_MACHINE_ID) || LOCAL_MACHINE_NAME)

    const containersByMachine = new Map<string, Container>()
    const cachedData = await api.get<ContainerListResponse>('/containers')

    for (const c of cachedData?.containers || []) {
      const machineId = c.machineId || c.machine_id || LOCAL_MACHINE_ID
      const machineName = getMachineName(machineId, machineNames)
      const container = normalizeContainer(c, machineId, machineName)
      containersByMachine.set(`${container.machineId}:${container.id}`, container)
    }

    const results = await Promise.all(
      allMachines.map(async (m) => {
        try {
          const conts = await remoteMachineService.listContainers(m.id)
          return conts.map((c) => normalizeContainer(c, m.id, m.name))
        } catch {
          return [] as Container[]
        }
      })
    )

    for (const conts of results) {
      for (const c of conts) {
        containersByMachine.set(`${c.machineId}:${c.id}`, c)
      }
    }

    const allContainers = Array.from(containersByMachine.values())
    allContainers.sort((a, b) => {
      if (a.machine !== b.machine) return a.machine.localeCompare(b.machine)
      return a.name.localeCompare(b.name)
    })

    containers.value = allContainers
  } catch (e) {
    error.value = e instanceof Error ? e.message : t('errors.loginFailed')
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

const { machineFilter, machineOptions, groupedItems } = useMachineFilter(filteredContainers, machines, (c) => c.machineId, (c) => c.machine)

const createMachineOptions = computed<RemoteMachine[]>(() => {
  const options = [...machines.value]
  if (!options.some((machine) => machine.id === LOCAL_MACHINE_ID)) {
    options.unshift({
      id: LOCAL_MACHINE_ID,
      name: LOCAL_MACHINE_NAME,
      host: 'localhost',
      port: 0,
      username: '',
      auth_method: 'password',
      docker_host: '',
      status: 'unknown',
      created_at: '',
      updated_at: '',
    })
  }
  return options
})

const defaultCreateMachineId = computed(() => {
  if (machineFilter.value) return machineFilter.value
  return createMachineOptions.value[0]?.id ?? LOCAL_MACHINE_ID
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

const onContainerCreated = (_container: { id: string; name: string; image: string; machineId: string }) => {
  showCreateModal.value = false
  refreshContainers()
}

onMounted(() => refreshContainers())
</script>

<style scoped>
.containers-page {
  max-width: 1400px;
  margin: 0 auto;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-3);
  padding: var(--space-16) 0;
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
}

.error-state {
  padding: var(--space-10) var(--space-6);
}

.empty-state {
  padding: var(--space-10) var(--space-6);
}

@media (max-width: 768px) {
  .item-list-row {
    grid-template-columns: 1fr;
    gap: var(--space-2);
  }
  .item-list-actions {
    justify-content: flex-end;
  }
}
</style>
