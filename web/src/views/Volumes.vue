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

    <div v-if="!loading && !error && volumes.length > 0" class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('volumes.searchPlaceholder')"
          type="text"
          class="input"
        />
      </div>
      <div class="filter-options">
        <select v-model="driverFilter" class="input">
          <option value="">{{ t('volumes.allDrivers') }}</option>
          <option value="local">local</option>
          <option value="nfs">nfs</option>
          <option value="cifs">cifs</option>
          <option value="tmpfs">tmpfs</option>
        </select>
      </div>
    </div>

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

    <div v-else class="card-grid">
      <VolumeCard
        v-for="vol in filteredVolumes"
        :key="`${vol.machineId}:${vol.name}`"
        :volume="vol"
        @delete="handleDelete"
      />
    </div>

    <VolumeCreateModal
      v-if="machines.length > 0"
      :machine-id="machines[0].id"
      :visible="showCreateModal"
      @created="onVolumeCreated"
      @close="showCreateModal = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { RefreshCw, Plus } from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import { volumeService } from '@/services/volumeService'
import type { Volume, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import VolumeCard from '@/components/volume/VolumeCard.vue'
import VolumeCreateModal from '@/components/volume/VolumeCreateModal.vue'

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

const handleDelete = async (volumeKey: string) => {
  const volume = volumes.value.find(v => `${v.machineId}:${v.name}` === volumeKey)
  if (!volume) return

  if (!confirm(t('volumes.confirmDelete'))) return

  try {
    await volumeService.deleteVolume(volume.machineId, volume.name)
    await refreshVolumes()
  } catch (e) {
    const msg = e instanceof Error ? e.message : t('errors.loginFailed')
    error.value = msg
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
