<template>
  <div class="networks-page">
    <PageHeader :title="t('networks.title')">
      <button @click="refreshNetworks" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('networks.refresh') }}
      </button>
      <button
        v-if="machines.length > 0"
        @click="showCreateModal = true"
        class="btn btn-primary"
      >
        <Plus :size="16" />
        {{ t('networks.create') }}
      </button>
    </PageHeader>

    <CollapsibleFilters
      v-if="!loading && !error && networks.length > 0"
      v-model="searchQuery"
      :search-placeholder="t('networks.searchPlaceholder')"
      search-label="Search"
      filter-label="Filters"
      :has-filters="true"
    >
      <template #filters>
        <select v-model="driverFilter" class="input">
          <option value="">{{ t('networks.allDrivers') }}</option>
          <option value="bridge">Bridge</option>
          <option value="host">Host</option>
          <option value="overlay">Overlay</option>
          <option value="macvlan">Macvlan</option>
          <option value="ipvlan">Ipvlan</option>
        </select>
        <select v-model="scopeFilter" class="input">
          <option value="">{{ t('networks.allScopes') }}</option>
          <option value="local">Local</option>
          <option value="global">Global</option>
          <option value="swarm">Swarm</option>
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
      <span>{{ t('networks.refresh') }}...</span>
    </div>

    <div v-else-if="error" class="error-state card text-center">
      <p class="mb-4">{{ error }}</p>
      <button @click="refreshNetworks" class="btn btn-secondary">{{ t('common.refresh') }}</button>
    </div>

    <div v-else-if="networks.length === 0" class="empty-state card text-center">
      <p class="mb-4">{{ t('networks.noNetworks') }}</p>
      <button
        v-if="machines.length > 0"
        @click="showCreateModal = true"
        class="btn btn-primary"
      >
        {{ t('networks.create') }}
      </button>
    </div>

    <div v-else-if="filteredNetworks.length === 0" class="empty-state card text-center">
      <p>{{ t('networks.noNetworks') }}</p>
    </div>

    <template v-else>
      <Transition name="view-fade" mode="out-in">
        <div v-if="viewMode === 'card'" key="card">
          <template v-for="group in groupedItems" :key="group.machineId">
            <div class="machine-section-header">
              <Server :size="16" class="icon" />
              {{ group.machineName }}
              <span class="count">{{ group.items.length }} {{ t('common.networks') }}</span>
            </div>
            <div class="card-grid">
              <NetworkCard
                v-for="network in group.items"
                :key="network.id"
                :network="network"
                @delete="handleDelete"
                @inspect="handleInspect"
              />
            </div>
          </template>
        </div>
        <div v-else key="list">
          <template v-for="group in groupedItems" :key="group.machineId">
            <div class="machine-section-header">
              <Server :size="16" class="icon" />
              {{ group.machineName }}
              <span class="count">{{ group.items.length }} {{ t('common.networks') }}</span>
            </div>
            <div class="item-list">
              <div
                v-for="network in group.items"
                :key="network.id"
                class="item-list-row"
              >
                <div class="item-list-info">
                  <div class="item-list-title">{{ network.name }}</div>
                  <div class="item-list-meta">
                    <span>{{ network.driver }}</span>
                    <span class="badge badge-info badge-sm">{{ network.scope }}</span>
                    <span>{{ network.containers?.length ?? 0 }} containers</span>
                  </div>
                </div>
                <div class="item-list-actions">
                  <button
                    class="btn btn-ghost btn-sm"
                    @click="handleInspect(network.id)"
                  >
                    <Eye :size="14" /> {{ t('common.inspect') }}
                  </button>
                  <button
                    class="btn btn-sm btn-ghost btn-danger-text"
                    @click="handleDelete(network.id)"
                  >
                    <Trash2 :size="14" /> {{ t('common.delete') }}
                  </button>
                </div>
              </div>
            </div>
          </template>
        </div>
      </Transition>
    </template>

    <InspectModal
      :visible="showInspect"
      title="Network Details"
      :fields="inspectFields"
      @close="showInspect = false"
    />

    <NetworkCreateModal
      v-if="machines.length > 0"
      :machine-id="defaultCreateMachineId"
      :machines="machines"
      :visible="showCreateModal"
      @created="onNetworkCreated"
      @close="showCreateModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus, Eye, Trash2, Server } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { networkService } from '@/services/networkService'
import type { Network, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import NetworkCard from '@/components/network/NetworkCard.vue'
import NetworkCreateModal from '@/components/network/NetworkCreateModal.vue'
import InspectModal from '@/components/ui/InspectModal.vue'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import CollapsibleFilters from '@/components/ui/CollapsibleFilters.vue'
import { useViewMode } from '@/composables/useViewMode'
import { useMachineFilter } from '@/composables/useMachineFilter'

interface NetworkWithMachine extends Network {
  machineId: string
  machine: string
}

const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
const driverFilter = ref('')
const scopeFilter = ref('')
const viewMode = useViewMode('networks')
const showCreateModal = ref(false)
const machines = ref<RemoteMachine[]>([])
const networks = ref<NetworkWithMachine[]>([])

const showInspect = ref(false)
const selectedNetwork = ref<NetworkWithMachine | null>(null)

const inspectFields = computed(() => {
  const n = selectedNetwork.value
  if (!n) return []
  return [
    { label: 'ID', value: n.id },
    { label: 'Name', value: n.name },
    { label: 'Driver', value: n.driver },
    { label: 'Scope', value: n.scope },
    { label: 'Internal', value: n.internal ? 'Yes' : 'No' },
    { label: 'Attachable', value: n.attachable !== undefined ? (n.attachable ? 'Yes' : 'No') : '-' },
    { label: 'Containers', value: String(n.containers?.length ?? 0) },
    { label: 'Machine', value: n.machine },
  ]
})

const handleInspect = (networkId: string) => {
  const n = networks.value.find(net => net.id === networkId)
  if (n) {
    selectedNetwork.value = n
    showInspect.value = true
  }
}

const filteredNetworks = computed(() => {
  let filtered = networks.value

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(n => n.name.toLowerCase().includes(q))
  }

  if (driverFilter.value) {
    filtered = filtered.filter(n => n.driver === driverFilter.value)
  }

  if (scopeFilter.value) {
    filtered = filtered.filter(n => n.scope === scopeFilter.value)
  }

  return filtered
})

const { machineFilter, machineOptions, groupedItems } = useMachineFilter(
  filteredNetworks,
  machines,
  (n) => n.machineId,
  (n) => n.machine,
)

const defaultCreateMachineId = computed(() => {
  if (machineFilter.value) return machineFilter.value
  return machines.value[0]?.id ?? ''
})

const refreshNetworks = async () => {
  loading.value = true
  error.value = ''
  try {
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const allNetworks: NetworkWithMachine[] = []
    await Promise.all(
      allMachines.map(async (m) => {
        try {
          const nets = await networkService.listNetworks(m.id)
          for (const n of nets) {
            allNetworks.push({ ...n, machineId: m.id, machine: m.name })
          }
        } catch {
          // machine offline or unreachable — skip silently
        }
      })
    )

    allNetworks.sort((a, b) => {
      if (a.machine !== b.machine) return a.machine.localeCompare(b.machine)
      return a.name.localeCompare(b.name)
    })

    networks.value = allNetworks
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  } finally {
    loading.value = false
  }
}

const handleDelete = async (networkId: string) => {
  const network = networks.value.find(n => n.id === networkId)
  if (!network) return

  if (!confirm(t('networks.confirmDelete'))) return

  try {
    await networkService.deleteNetwork(network.machineId, networkId)
    await refreshNetworks()
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  }
}

const onNetworkCreated = () => {
  refreshNetworks()
}

onMounted(() => refreshNetworks())
</script>

<style scoped>
.networks-page {
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
  .card-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .item-list-row {
    grid-template-columns: 1fr;
    gap: var(--space-2);
  }

  .item-list-row .item-list-actions {
    justify-content: flex-end;
  }
}
</style>
