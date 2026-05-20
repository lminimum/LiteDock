<template>
  <div v-if="visible" class="modal-overlay">
    <div class="card modal-card" @click.stop>
      <div class="modal-header">
        <h3 class="modal-title">Create Network</h3>
        <button
          type="button"
          class="btn btn-ghost btn-sm"
          @click="$emit('close')"
          :disabled="submitting"
          aria-label="Close"
        >
          <X :size="16" />
        </button>
      </div>

      <form @submit.prevent="handleSubmit">
        <TargetMachineSelect
          v-model="selectedMachineId"
          select-id="network-target-machine"
          :options="machineOptions"
          :disabled="submitting"
          hint="Network will be created on the selected Docker host."
        />

        <div class="form-field">
          <label class="form-label">Name</label>
          <input
            v-model="name"
            class="input"
            placeholder="network name"
            :disabled="submitting"
          />
        </div>

        <div class="form-field">
          <label class="form-label">Driver</label>
          <select v-model="driver" class="input" :disabled="submitting">
            <option v-for="d in drivers" :key="d" :value="d">{{ d }}</option>
          </select>
        </div>

        <div v-if="error" class="form-error">{{ error }}</div>

        <div class="modal-actions">
          <button
            type="button"
            class="btn btn-ghost"
            @click="$emit('close')"
            :disabled="submitting"
          >
            Cancel
          </button>
          <button
            type="submit"
            class="btn btn-primary"
            :disabled="submitting || !name.trim()"
          >
            {{ submitting ? 'Creating...' : 'Create' }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { networkService } from '@/services/networkService'
import TargetMachineSelect from '@/components/ui/TargetMachineSelect.vue'
import type { Network, RemoteMachine } from '@/types'

const props = defineProps<{
  machineId: string
  machines: RemoteMachine[]
  visible: boolean
}>()

const emit = defineEmits<{
  created: [network: Network]
  close: []
}>()

const drivers = ['bridge', 'host', 'overlay', 'macvlan', 'ipvlan']

const name = ref('')
const driver = ref('bridge')
const error = ref('')
const submitting = ref(false)
const selectedMachineId = ref(props.machineId)

const machineOptions = computed(() => props.machines.map((machine) => ({
  value: machine.id,
  label: machine.name,
  status: machine.status,
})))

watch(() => props.visible, (val) => {
  if (val) {
    name.value = ''
    driver.value = 'bridge'
    error.value = ''
    submitting.value = false
    selectedMachineId.value = props.machineId
  }
})

watch(() => props.machineId, (machineId) => {
  if (!props.visible) return
  selectedMachineId.value = machineId
})

const handleSubmit = async () => {
  if (!name.value.trim()) {
    error.value = 'Network name is required'
    return
  }

  if (!selectedMachineId.value) {
    error.value = 'Target machine is required'
    return
  }

  error.value = ''
  submitting.value = true

  try {
    const network = await networkService.createNetwork(selectedMachineId.value, {
      name: name.value.trim(),
      driver: driver.value
    })
    emit('created', network)
    emit('close')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Failed to create network'
    error.value = msg
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--color-background-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: var(--space-4);
}

.modal-card {
  max-width: 420px;
  width: 100%;
  padding: var(--space-6);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-6);
}

.modal-header .modal-title {
  margin-bottom: 0;
}

.modal-title {
  margin-bottom: var(--space-6);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-6);
}

.modal-header .modal-title {
  margin-bottom: 0;
}

.modal-title {
  margin-bottom: var(--space-6);
}

.form-field {
  margin-bottom: var(--space-4);
}

.form-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
}

.form-error {
  margin-top: var(--space-3);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  color: var(--color-error);
  background: var(--color-error-bg);
  border-radius: var(--radius-sm);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  margin-top: var(--space-6);
  padding-top: var(--space-4);
  border-top: 1px solid var(--color-border-weak);
}
</style>
