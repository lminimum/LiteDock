<template>
  <div v-if="visible" class="modal-overlay" @click.self="$emit('close')">
    <div class="card modal-card" @click.stop>
      <h3 class="modal-title">Create Container</h3>

      <form @submit.prevent="handleSubmit">
        <TargetMachineSelect
          v-model="selectedMachineId"
          select-id="container-target-machine"
          :options="machineOptions"
          :disabled="submitting"
          hint="Container will be created on the selected Docker host."
        />

        <div class="form-field">
          <label class="form-label">Name</label>
          <input
            v-model="name"
            class="input"
            placeholder="container name"
            :disabled="submitting"
          />
        </div>

        <div class="form-field">
          <label class="form-label">Image <span class="required">*</span></label>
          <input
            v-model="image"
            class="input"
            placeholder="e.g. nginx:latest"
            :disabled="submitting"
          />
        </div>

        <div class="form-field">
          <label class="form-label">Command</label>
          <input
            v-model="cmdInput"
            class="input"
            :placeholder="t('containers.createModal.cmdPlaceholder')"
            :disabled="submitting"
          />
        </div>

        <div class="form-field">
          <label class="form-label">Environment (key=value)</label>
          <textarea
            v-model="envInput"
            class="input font-mono"
            rows="3"
            placeholder="NODE_ENV=production&#10;PORT=8080"
            :disabled="submitting"
          ></textarea>
        </div>

        <div class="form-field">
          <label class="form-label">Ports (host:container)</label>
          <textarea
            v-model="portsInput"
            class="input font-mono"
            rows="2"
            placeholder="8080:80&#10;443:443"
            :disabled="submitting"
          ></textarea>
        </div>

        <div class="form-field">
          <label class="form-label">Volumes (host:container)</label>
          <textarea
            v-model="volumesInput"
            class="input font-mono"
            rows="2"
            placeholder="/host/path:/container/path"
            :disabled="submitting"
          ></textarea>
        </div>

        <div class="form-field">
          <label class="form-label">Network</label>
          <select v-model="network" class="input" :disabled="submitting">
            <option value="">bridge (default)</option>
            <option value="host">host</option>
            <option value="none">none</option>
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
            :disabled="submitting || !image.trim()"
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
import { remoteMachineService } from '@/services/remoteMachineService'
import { t } from '@/i18n'
import TargetMachineSelect from '@/components/ui/TargetMachineSelect.vue'
import type { RemoteMachine } from '@/types'

const props = defineProps<{
  machineId: string
  machines: RemoteMachine[]
  visible: boolean
}>()

const emit = defineEmits<{
  created: [container: { id: string; name: string; image: string; machineId: string }]
  close: []
}>()

const name = ref('')
const image = ref('')
const cmdInput = ref('')
const envInput = ref('')
const portsInput = ref('')
const volumesInput = ref('')
const network = ref('')
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
    image.value = ''
    cmdInput.value = ''
    envInput.value = ''
    portsInput.value = ''
    volumesInput.value = ''
    network.value = ''
    error.value = ''
    submitting.value = false
    selectedMachineId.value = props.machineId
  }
})

watch(() => props.machineId, (machineId) => {
  if (!props.visible) return
  selectedMachineId.value = machineId
})

function parseLines(input: string): string[] {
  return input
    .split('\n')
    .map(l => l.trim())
    .filter(l => l.length > 0)
}

const handleSubmit = async () => {
  if (!image.value.trim()) {
    error.value = 'Image is required'
    return
  }

  if (!selectedMachineId.value) {
    error.value = 'Target machine is required'
    return
  }

  error.value = ''
  submitting.value = true

  try {
    const cmd = cmdInput.value.trim() ? [cmdInput.value.trim()] : undefined
    const env = parseLines(envInput.value)
    const ports = parseLines(portsInput.value)
    const volumes = parseLines(volumesInput.value)

    const result = await remoteMachineService.createContainer(selectedMachineId.value, {
      name: name.value.trim() || undefined,
      image: image.value.trim(),
      env: env.length > 0 ? env : undefined,
      ports: ports.length > 0 ? ports : undefined,
      volumes: volumes.length > 0 ? volumes : undefined,
      network: network.value || undefined,
      cmd,
    })

    emit('created', {
      id: result.id,
      name: name.value.trim() || result.id.substring(0, 12),
      image: image.value.trim(),
      machineId: selectedMachineId.value,
    })
    emit('close')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Failed to create container'
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
  max-width: 520px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
  padding: var(--space-6);
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

.required {
  color: var(--color-error);
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
