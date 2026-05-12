<template>
  <div class="machines-page">
    <PageHeader :title="t('machines.title')">
      <button @click="refreshMachines" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('machines.refresh') }}
      </button>
      <button @click="openAddModal" class="btn btn-primary">
        <Plus :size="16" />
        {{ t('machines.addMachine') }}
      </button>
    </PageHeader>

    <div class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('machines.searchPlaceholder')"
          type="text"
          class="input"
        />
      </div>
      <ViewToggle v-model="viewMode" />
    </div>

    <div v-if="machines.length === 0 && !loading" class="empty-state">
      <Server :size="48" class="empty-icon" />
      <p>{{ t('machines.empty') }}</p>
      <button @click="openAddModal" class="btn btn-primary">
        <Plus :size="16" />
        {{ t('machines.addFirst') }}
      </button>
    </div>

    <Transition name="view-fade" mode="out-in">
      <div v-if="viewMode === 'card'" class="machines-grid" key="card">
        <div
          v-for="machine in filteredMachines"
          :key="machine.id"
          class="machine-card"
          :class="{ 'status-online': machine.status === 'online', 'machine-local': machine.id === 'local' }"
        >
          <div class="machine-header">
            <div class="machine-name">
              {{ machine.name }}
              <span v-if="machine.id === 'local'" class="badge badge-primary badge-sm" style="margin-left: 8px;">本地</span>
            </div>
            <div class="badge" :class="getStatusClass(machine.status)">
              {{ t(`machines.status.${machine.status}`) }}
            </div>
          </div>

          <div class="machine-info">
            <div v-if="machine.id !== 'local'" class="info-item">
              <span class="label">{{ t('machines.host') }}</span>
              <span class="value">{{ machine.host }}:{{ machine.port }}</span>
            </div>
            <div v-if="machine.id !== 'local'" class="info-item">
              <span class="label">{{ t('machines.username') }}</span>
              <span class="value">{{ machine.username }}</span>
            </div>
            <div v-if="machine.id !== 'local'" class="info-item">
              <span class="label">{{ t('machines.authMethod') }}</span>
              <span class="badge" :class="machine.auth_method === 'password' ? 'badge-info' : 'badge-success'">
                {{ machine.auth_method === 'password' ? 'Password' : 'SSH Key' }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">{{ t('machines.dockerHost') }}</span>
              <span class="value mono">{{ machine.docker_host }}</span>
            </div>
            <div v-if="machine.id === 'local'" class="info-item">
              <span class="label">连接方式</span>
              <span class="badge badge-primary">Unix Socket（本机）</span>
            </div>
          </div>

          <div class="machine-actions">
            <button @click="testConnection(machine.id)" class="btn btn-sm btn-secondary" :disabled="testingId === machine.id">
              <Wifi :size="14" :class="{ 'spinning': testingId === machine.id }" />
              {{ t('machines.test') }}
            </button>
            <button @click="openEditModal(machine)" class="btn btn-sm btn-secondary" :disabled="machine.id === 'local'">
              <Pencil :size="14" />
              {{ t('machines.edit') }}
            </button>
            <button @click="viewContainers(machine.id)" class="btn btn-sm btn-secondary">
              <Box :size="14" />
              {{ t('machines.containers') }}
            </button>
            <button @click="deleteMachine(machine.id)" class="btn btn-sm btn-ghost btn-danger-text" :disabled="machine.id === 'local'">
              <Trash2 :size="14" />
              {{ t('machines.delete') }}
            </button>
          </div>
        </div>
      </div>

      <div v-else class="item-list" key="list">
        <div v-for="machine in filteredMachines" :key="machine.id" class="item-list-row">
          <div class="item-list-info">
            <div class="item-list-title">{{ machine.name }}</div>
            <div class="item-list-meta">
              <span class="badge" :class="machine.status === 'online' ? 'badge-success' : 'badge-error'">{{ machine.status }}</span>
              <span v-if="machine.id !== 'local'" class="text-muted">{{ machine.host }}:{{ machine.port }}</span>
              <span class="text-muted">{{ machine.docker_host }}</span>
              <span v-if="machine.id === 'local'" class="badge badge-primary">本机</span>
              <span v-else class="badge" :class="machine.auth_method === 'password' ? 'badge-info' : 'badge-success'">{{ machine.auth_method === 'password' ? 'Password' : 'SSH Key' }}</span>
            </div>
          </div>
          <div class="item-list-actions">
            <button @click="testConnection(machine.id)" class="btn btn-sm btn-secondary" :disabled="testingId === machine.id">
              <Wifi :size="14" :class="{ 'spinning': testingId === machine.id }" /> Test
            </button>
            <button @click="openEditModal(machine)" class="btn btn-sm btn-secondary" :disabled="machine.id === 'local'">
              <Pencil :size="14" /> Edit
            </button>
            <button @click="viewContainers(machine.id)" class="btn btn-sm btn-secondary">
              <Box :size="14" /> Containers
            </button>
            <button @click="deleteMachine(machine.id)" class="btn btn-sm btn-ghost btn-danger-text" :disabled="machine.id === 'local'">
              <Trash2 :size="14" /> Delete
            </button>
          </div>
        </div>
      </div>
    </Transition>

    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ editingMachine ? t('machines.editMachine') : t('machines.addMachine') }}</h2>
          <button @click="closeModal" class="btn btn-ghost btn-sm">
            <X :size="16" />
          </button>
        </div>

        <div class="modal-body">
          <div v-if="formError" class="error-message">{{ formError }}</div>

          <div class="form-group">
            <label>Name</label>
            <input v-model="form.name" type="text" class="input" required />
          </div>

          <div class="form-row">
            <div class="form-group" style="flex: 2">
              <label>{{ t('machines.host') }}</label>
              <input v-model="form.host" type="text" class="input" required />
            </div>
            <div class="form-group" style="flex: 1">
              <label>{{ t('machines.port') }}</label>
              <input v-model.number="form.port" type="number" class="input" min="1" max="65535" />
            </div>
          </div>

          <div class="form-group">
            <label>{{ t('machines.username') }}</label>
            <input v-model="form.username" type="text" class="input" required />
          </div>

          <div class="form-group">
            <label>{{ t('machines.authMethod') }}</label>
            <select v-model="form.auth_method" class="input">
              <option value="password">{{ t('machines.password') }}</option>
              <option value="key">{{ t('machines.sshKey') }}</option>
            </select>
          </div>

          <div v-if="form.auth_method === 'password'" class="form-group">
            <label>{{ t('machines.password') }}</label>
            <input
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              class="input"
              :placeholder="editingMachine ? t('machines.leaveBlank') : ''"
            />
          </div>

          <div v-if="form.auth_method === 'key'" class="form-group">
            <label>{{ t('machines.sshKey') }}</label>
            <textarea
              v-model="form.ssh_key"
              class="input textarea"
              :placeholder="t('machines.sshKeyPlaceholder')"
              rows="6"
            ></textarea>
          </div>

          <div class="form-group">
            <label>{{ t('machines.dockerHost') }}</label>
            <input v-model="form.docker_host" type="text" class="input" />
          </div>
        </div>

        <div class="modal-actions">
          <button @click="closeModal" class="btn btn-secondary">{{ t('common.cancel') }}</button>
          <button @click="saveMachine" class="btn btn-primary" :disabled="saving">
            {{ t('machines.saving') }}
          </button>
        </div>
      </div>
    </div>

    <div v-if="toast.show" class="toast" :class="toast.type">
      {{ toast.message }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  RefreshCw,
  Plus,
  Wifi,
  Pencil,
  Box,
  Trash2,
  X,
  Server
} from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { RemoteMachine, CreateMachineRequest, UpdateMachineRequest } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import { useViewMode } from '@/composables/useViewMode'

const router = useRouter()

const loading = ref(false)
const saving = ref(false)
const testingId = ref<string | null>(null)
const searchQuery = ref('')
const machines = ref<RemoteMachine[]>([])
const showModal = ref(false)
const editingMachine = ref<RemoteMachine | null>(null)
const showPassword = ref(false)
const formError = ref('')
const viewMode = useViewMode('machines')

const form = ref({
  name: '',
  host: '',
  port: 22,
  username: '',
  auth_method: 'password' as 'password' | 'key',
  password: '',
  ssh_key: '',
  docker_host: '/var/run/docker.sock'
})

const toast = ref({
  show: false,
  message: '',
  type: 'toast-success'
})

const filteredMachines = computed(() => {
  let filtered = machines.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(m =>
      m.name.toLowerCase().includes(q) ||
      m.host.toLowerCase().includes(q)
    )
  }
  return filtered
})

const getStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    online: 'badge-success',
    offline: 'badge-error',
    unknown: 'badge-warning'
  }
  return classMap[status] || ''
}

const refreshMachines = async () => {
  loading.value = true
  try {
    machines.value = await remoteMachineService.list()
  } catch (e) {
    console.error('Failed to load machines:', e)
  } finally {
    loading.value = false
  }
}

const openAddModal = () => {
  editingMachine.value = null
  form.value = {
    name: '',
    host: '',
    port: 22,
    username: '',
    auth_method: 'password',
    password: '',
    ssh_key: '',
    docker_host: '/var/run/docker.sock'
  }
  formError.value = ''
  showModal.value = true
}

const openEditModal = (machine: RemoteMachine) => {
  editingMachine.value = machine
  form.value = {
    name: machine.name,
    host: machine.host,
    port: machine.port,
    username: machine.username,
    auth_method: machine.auth_method,
    password: '',
    ssh_key: machine.ssh_key_path || '',
    docker_host: machine.docker_host
  }
  formError.value = ''
  showModal.value = true
}

const closeModal = () => {
  showModal.value = false
  editingMachine.value = null
}

const saveMachine = async () => {
  formError.value = ''
  if (!form.value.name || !form.value.host || !form.value.username) {
    formError.value = 'Name, host, and username are required'
    return
  }
  if (form.value.auth_method === 'password' && !editingMachine.value && !form.value.password) {
    formError.value = 'Password is required'
    return
  }
  if (form.value.auth_method === 'key' && !form.value.ssh_key) {
    formError.value = 'SSH key is required'
    return
  }

  saving.value = true
  try {
    const data: CreateMachineRequest | UpdateMachineRequest = {
      name: form.value.name,
      host: form.value.host,
      port: form.value.port,
      username: form.value.username,
      auth_method: form.value.auth_method,
      docker_host: form.value.docker_host
    }
    if (form.value.auth_method === 'password' && form.value.password) {
      (data as CreateMachineRequest).password = form.value.password
    }
    if (form.value.auth_method === 'key' && form.value.ssh_key) {
      (data as CreateMachineRequest).ssh_key = form.value.ssh_key
    }

    if (editingMachine.value) {
      await remoteMachineService.update(editingMachine.value.id, data)
    } else {
      await remoteMachineService.create(data as CreateMachineRequest)
    }
    closeModal()
    await refreshMachines()
  } catch (e: any) {
    formError.value = e?.response?.data?.error || e?.message || 'Failed to save'
  } finally {
    saving.value = false
  }
}

const testConnection = async (id: string) => {
  testingId.value = id
  try {
    await remoteMachineService.testConnection(id)
    showToast(t('machines.testSuccess'), 'toast-success')
    await refreshMachines()
  } catch (e) {
    showToast(t('machines.testFailed'), 'toast-error')
  } finally {
    testingId.value = null
  }
}

const viewContainers = (id: string) => {
  router.push(`/machines/${id}`)
}

const deleteMachine = async (id: string) => {
  if (!confirm(t('machines.confirmDelete'))) return
  try {
    await remoteMachineService.delete(id)
    await refreshMachines()
  } catch (e) {
    console.error('Failed to delete machine:', e)
  }
}

const showToast = (message: string, type: string) => {
  toast.value = { show: true, message, type }
  setTimeout(() => { toast.value.show = false }, 4000)
}

onMounted(() => refreshMachines())
</script>

<style scoped>
.machines-page {
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

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-16);
  color: var(--color-text-weak);
  gap: var(--space-4);
}

.empty-icon {
  opacity: 0.4;
}

.machines-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: var(--space-4);
}

.machine-card {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-4);
  border-left: 3px solid var(--color-border);
  transition: border-color var(--transition-fast);
}

.machine-card.status-online {
  border-left-color: var(--color-success);
}

.machine-card.machine-local {
  border-left-color: var(--color-accent);
  background: var(--color-background-weak);
}

.machine-card.machine-local .machine-name {
  color: var(--color-accent);
}

.badge-sm {
  font-size: var(--font-size-xs);
  padding: 1px 6px;
}

.machine-card:hover {
  border-color: var(--color-text-weaker);
}

.machine-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.machine-name {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.machine-info {
  margin-bottom: var(--space-4);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
}

.info-item .label {
  color: var(--color-text);
}

.info-item .value {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
}

.info-item .value.mono {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
}

.machine-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: var(--color-background-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: var(--space-4);
}

.modal {
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  width: 100%;
  max-width: 520px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border-weak);
}

.modal-header h2 {
  margin: 0;
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
}

.modal-body {
  padding: var(--space-5);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: var(--space-4) var(--space-5);
  border-top: 1px solid var(--color-border-weak);
}

.form-group {
  margin-bottom: var(--space-4);
}

.form-group label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
}

.form-row {
  display: flex;
  gap: var(--space-4);
}

.textarea {
  resize: vertical;
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
}

.error-message {
  padding: var(--space-3);
  background: var(--color-error-bg);
  color: var(--color-error);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  margin-bottom: var(--space-4);
}

.toast {
  position: fixed;
  bottom: var(--space-6);
  right: var(--space-6);
  padding: var(--space-3) var(--space-5);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  z-index: 2000;
  animation: slideUp 0.2s ease-out;
}

.toast-success {
  background: var(--color-success);
  color: #fdfcfc;
}

.toast-error {
  background: var(--color-error);
  color: #fdfcfc;
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 768px) {
  .machines-grid {
    grid-template-columns: 1fr;
  }

  .filters {
    flex-direction: column;
  }

  .machine-actions {
    justify-content: center;
  }

  .item-list-row {
    grid-template-columns: 1fr;
    gap: var(--space-2);
  }
  .item-list-actions {
    justify-content: flex-end;
  }
  .item-list-actions .btn {
    padding: var(--space-1) var(--space-2);
    font-size: var(--font-size-xs);
  }
}
</style>
