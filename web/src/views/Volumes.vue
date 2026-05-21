<template>
  <div class="volumes-page">
    <PageHeader :title="t('volumes.title')">
      <button @click="refreshVolumes" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('volumes.refresh') }}
      </button>
      <button
        v-if="machines.length > 0"
        @click="showCreateModal = true"
        class="btn btn-primary"
      >
        <Plus :size="16" />
        {{ t('volumes.create') }}
      </button>
    </PageHeader>

    <CollapsibleFilters
      v-if="!loading && !error && volumes.length > 0"
      v-model="searchQuery"
      :search-placeholder="t('volumes.searchPlaceholder')"
      search-label="Search"
      filter-label="Filters"
      :has-filters="true"
    >
      <template #filters>
        <select v-model="driverFilter" class="input">
          <option value="">{{ t('volumes.allDrivers') }}</option>
          <option value="local">local</option>
          <option value="nfs">nfs</option>
          <option value="cifs">cifs</option>
          <option value="tmpfs">tmpfs</option>
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
      <span>{{ t('volumes.refresh') }}...</span>
    </div>

    <div v-else-if="error" class="error-state card text-center">
      <p class="mb-4">{{ error }}</p>
      <button @click="refreshVolumes" class="btn btn-secondary">{{ t('common.refresh') }}</button>
    </div>

    <div v-else-if="volumes.length === 0" class="empty-state card text-center">
      <p class="mb-4">{{ t('volumes.noVolumes') }}</p>
      <button
        v-if="machines.length > 0"
        @click="showCreateModal = true"
        class="btn btn-primary"
      >
        {{ t('volumes.create') }}
      </button>
    </div>

    <div v-else-if="filteredVolumes.length === 0" class="empty-state card text-center">
      <p>{{ t('volumes.noVolumes') }}</p>
    </div>

    <Transition name="view-fade" mode="out-in">
      <div v-if="viewMode === 'card'" key="card">
        <template v-for="group in groupedItems" :key="group.machineId">
          <div class="machine-section-header">
            <Server :size="16" class="icon" />
            {{ group.machineName }}
            <span class="count">{{ group.items.length }} {{ t('common.volumes') }}</span>
          </div>
          <div class="card-grid">
            <VolumeCard
              v-for="vol in group.items"
              :key="`${vol.machineId}:${vol.name}`"
              :volume="vol"
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
            <span class="count">{{ group.items.length }} {{ t('common.volumes') }}</span>
          </div>
          <div class="item-list">
            <div v-for="vol in group.items" :key="`${vol.machineId}:${vol.name}`" class="item-list-row">
              <div class="item-list-info">
                <div class="item-list-title">{{ vol.name }}</div>
                <div class="item-list-meta">
                  <span class="badge badge-info">{{ vol.driver }}</span>
                  <span class="badge badge-info">{{ vol.scope }}</span>
                  <span class="text-muted truncate" :title="vol.mountpoint">{{ vol.mountpoint }}</span>
                  <span>{{ vol.machine }}</span>
                </div>
              </div>
              <div class="item-list-actions">
                <button @click="handleInspect(vol.name)" class="btn btn-sm btn-ghost">
                  <Eye :size="14" /> {{ t('common.inspect') }}
                </button>
                <button @click="handleDelete(`${vol.machineId}:${vol.name}`)" class="btn btn-sm btn-ghost btn-danger-text">
                  <Trash2 :size="14" /> {{ t('common.delete') }}
                </button>
              </div>
            </div>
          </div>
        </template>
      </div></Transition>

    <InspectModal
      :visible="showInspect"
      title="Volume Details"
      :fields="inspectFields"
      @close="showInspect = false"
    />

    <VolumeCreateModal
      v-if="machines.length > 0"
      :machine-id="defaultCreateMachineId"
      :machines="machines"
      :visible="showCreateModal"
      @created="onVolumeCreated"
      @close="showCreateModal = false"
    />

    <ConfirmModal
      :visible="confirmState !== null"
      :title="confirmState?.title || ''"
      :message="confirmState?.message || ''"
      :confirm-text="confirmState?.confirmText"
      :danger="confirmState?.danger ?? false"
      :disabled="confirmBusy"
      @confirm="confirmAction"
      @cancel="cancelConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus, Eye, Trash2, Server } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { volumeService } from '@/services/volumeService'
import type { Volume, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import VolumeCard from '@/components/volume/VolumeCard.vue'
import VolumeCreateModal from '@/components/volume/VolumeCreateModal.vue'
import InspectModal from '@/components/ui/InspectModal.vue'
import ConfirmModal from '@/components/ui/ConfirmModal.vue'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import CollapsibleFilters from '@/components/ui/CollapsibleFilters.vue'
import { useViewMode } from '@/composables/useViewMode'
import { useMachineFilter } from '@/composables/useMachineFilter'

interface VolumeWithMachine extends Volume {
  machineId: string
  machine: string
}

const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
const driverFilter = ref('')
const showCreateModal = ref(false)
const machines = ref<RemoteMachine[]>([])
const volumes = ref<VolumeWithMachine[]>([])

const viewMode = useViewMode('volumes')
const showInspect = ref(false)
const selectedVolume = ref<VolumeWithMachine | null>(null)

const confirmState = ref<{
  title: string
  message: string
  confirmText?: string
  danger?: boolean
  action: 'delete'
  id: string
} | null>(null)
const confirmBusy = ref(false)

const inspectFields = computed(() => {
  const v = selectedVolume.value
  if (!v) return []
  return [
    { label: 'Name', value: v.name },
    { label: 'Driver', value: v.driver },
    { label: 'Scope', value: v.scope },
    { label: 'Mountpoint', value: v.mountpoint },
    { label: 'Size', value: v.size !== undefined ? `${v.size} bytes` : '-' },
    { label: 'Created At', value: v.createdAt || '-' },
    { label: 'Machine', value: v.machine },
  ]
})

const handleInspect = (volumeName: string) => {
  const v = volumes.value.find(vol => vol.name === volumeName)
  if (v) {
    selectedVolume.value = v
    showInspect.value = true
  }
}

const cancelConfirm = () => {
  if (confirmBusy.value) return
  confirmState.value = null
}

const openDeleteConfirm = (volumeKey: string) => {
  confirmState.value = {
    title: t('volumes.delete'),
    message: t('volumes.confirmDelete'),
    confirmText: t('volumes.delete'),
    danger: true,
    action: 'delete',
    id: volumeKey,
  }
}

const confirmAction = async () => {
  const state = confirmState.value
  if (!state || confirmBusy.value) return
  confirmBusy.value = true
  confirmState.value = null

  try {
    if (state.action === 'delete') {
      await performDeleteVolume(state.id)
    }
  } finally {
    confirmBusy.value = false
  }
}

const performDeleteVolume = async (volumeKey: string) => {
  try {
    const volume = volumes.value.find(v => `${v.machineId}:${v.name}` === volumeKey)
    if (!volume) return
    await volumeService.deleteVolume(volume.machineId, volume.name)
    await refreshVolumes()
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  }
}

const handleDelete = async (volumeKey: string) => {
  openDeleteConfirm(volumeKey)
}

const filteredVolumes = computed(() => {
  let filtered = volumes.value

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(v => v.name.toLowerCase().includes(q))
  }

  if (driverFilter.value) {
    filtered = filtered.filter(v => v.driver === driverFilter.value)
  }

  return filtered
})

const { machineFilter, machineOptions, groupedItems } = useMachineFilter(filteredVolumes, machines, (v) => v.machineId, (v) => v.machine)

const defaultCreateMachineId = computed(() => {
  if (machineFilter.value) return machineFilter.value
  return machines.value[0]?.id ?? ''
})

const refreshVolumes = async () => {
  loading.value = true
  error.value = ''
  try {
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const allVolumes: VolumeWithMachine[] = []
    await Promise.all(
      allMachines.map(async (m) => {
        try {
          const vols = await volumeService.listVolumes(m.id)
          for (const v of vols) {
            allVolumes.push({ ...v, machineId: m.id, machine: m.name })
          }
        } catch {
          // machine offline or unreachable — skip silently
        }
      })
    )

    allVolumes.sort((a, b) => {
      if (a.machine !== b.machine) return a.machine.localeCompare(b.machine)
      return a.name.localeCompare(b.name)
    })

    volumes.value = allVolumes
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
  } finally {
    loading.value = false
  }
}

const onVolumeCreated = () => {
  refreshVolumes()
}

onMounted(() => refreshVolumes())
</script>

<style scoped>
.volumes-page {
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

  .item-list-actions {
    justify-content: flex-end;
  }
}
</style>
