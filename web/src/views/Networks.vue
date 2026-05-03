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

    <div v-if="!loading && !error && networks.length > 0" class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('networks.searchPlaceholder')"
          type="text"
          class="input"
        />
      </div>
      <div class="filter-options">
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
      </div>
    </div>

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

    <div v-else class="card-grid">
      <NetworkCard
        v-for="network in filteredNetworks"
        :key="network.id"
        :network="network"
        @delete="handleDelete"
        @inspect="handleInspect"
      />
    </div>

    <InspectModal
      :visible="showInspect"
      title="Network Details"
      :fields="inspectFields"
      @close="showInspect = false"
    />

    <NetworkCreateModal
      v-if="machines.length > 0"
      :machine-id="machines[0].id"
      :visible="showCreateModal"
      @created="onNetworkCreated"
      @close="showCreateModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { networkService } from '@/services/networkService'
import type { Network, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import NetworkCard from '@/components/network/NetworkCard.vue'
import NetworkCreateModal from '@/components/network/NetworkCreateModal.vue'
import InspectModal from '@/components/ui/InspectModal.vue'

interface NetworkWithMachine extends Network {
  machineId: string
  machine: string
}

const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
const driverFilter = ref('')
const scopeFilter = ref('')
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

.filters {
  display: flex;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
  padding: var(--space-4);
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
}

.search-box {
  flex: 1;
}

.filter-options {
  display: flex;
  gap: var(--space-3);
}

.filter-options select {
  min-width: 130px;
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

  .filters {
    flex-direction: column;
  }

  .filter-options {
    flex-direction: column;
  }

  .filter-options select {
    min-width: 100%;
  }
}
</style>
