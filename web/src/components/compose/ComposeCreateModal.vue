<template>
  <div v-if="show" class="modal-overlay">
    <div class="card modal-card" @click.stop>
      <div class="modal-header">
        <h3 class="modal-title">Create Compose Project</h3>
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

      <div class="tabs flex gap-2 mb-3">
        <button
          class="btn"
          :class="tab === 'template' ? 'btn-primary' : 'btn-ghost'"
          @click="tab = 'template'"
        >
          From Template
        </button>
        <button
          class="btn"
          :class="tab === 'blank' ? 'btn-primary' : 'btn-ghost'"
          @click="tab = 'blank'"
        >
          Blank
        </button>
      </div>

      <TargetMachineSelect
        v-model="selectedMachineId"
        select-id="compose-target-machine"
        :options="machineOptions"
        :disabled="submitting"
        hint="Compose project will be created on the selected Docker host."
      />

      <div v-if="tab === 'template'">
        <div class="template-grid grid grid-cols-2 gap-3">
          <div
            v-for="t in templates"
            :key="t.name"
            class="template-card card card-hover"
            :class="{ 'template-selected': selectedTemplate?.name === t.name }"
            @click="selectTemplate(t)"
          >
            <div class="template-name font-medium">{{ t.name }}</div>
            <div class="template-desc text-xs text-muted">{{ t.description }}</div>
          </div>
        </div>
      </div>

      <div v-if="tab === 'blank' || selectedTemplate" class="mt-4">
        <div class="form-group">
          <label class="form-label">Project Name</label>
          <input
            v-model="projectName"
            class="input"
            placeholder="my-project"
            :disabled="submitting"
            required
          />
        </div>
        <div class="form-group">
          <label class="form-label">File Path <span class="text-xs">(optional)</span></label>
          <input v-model="filePath" class="input" placeholder="e.g. /home/user/myapp/docker-compose.yml" :disabled="submitting" />
        </div>
        <div class="form-group">
          <label class="form-label">YAML Content</label>
          <textarea
            v-model="yamlContent"
            class="input font-mono"
            rows="10"
            placeholder="version: '3'..."
            :disabled="submitting"
          ></textarea>
        </div>
      </div>

      <div v-if="error" class="form-error">{{ error }}</div>

      <div class="modal-actions">
        <button
          type="button"
          class="btn btn-ghost"
          :disabled="submitting"
          @click="$emit('close')"
        >
          Cancel
        </button>
        <button
          type="button"
          class="btn btn-primary"
          :disabled="!isValid || submitting"
          @click="create"
        >
          {{ submitting ? 'Creating...' : 'Create' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { X } from 'lucide-vue-next'
import { composeService } from '@/services/composeService'
import TargetMachineSelect from '@/components/ui/TargetMachineSelect.vue'
import type { ComposeProject, ComposeTemplate, RemoteMachine } from '@/types'

const props = defineProps<{
  show: boolean
  machineId: string
  machines: RemoteMachine[]
  templates: ComposeTemplate[]
}>()

const emit = defineEmits<{
  close: []
  created: [project: ComposeProject]
}>()

const tab = ref<'template' | 'blank'>('template')
const selectedTemplate = ref<ComposeTemplate | null>(null)
const projectName = ref('')
const yamlContent = ref('')
const filePath = ref('')
const submitting = ref(false)
const error = ref('')
const selectedMachineId = ref(props.machineId)

const machineOptions = computed(() => props.machines.map((machine) => ({
  value: machine.id,
  label: machine.name,
  status: machine.status,
})))

const isValid = computed(() => {
  return projectName.value.trim().length > 0 && yamlContent.value.trim().length > 0
})

function selectTemplate(t: ComposeTemplate) {
  selectedTemplate.value = t
  yamlContent.value = t.content
  // Auto-fill project name based on template name
  const autoName = t.name
    .toLowerCase()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
  projectName.value = autoName
}

async function create() {
  if (!isValid.value) return

  if (!selectedMachineId.value) {
    error.value = 'Target machine is required'
    return
  }

  error.value = ''
  submitting.value = true

  try {
    const project = await composeService.createProject(selectedMachineId.value, {
      name: projectName.value.trim(),
      content: yamlContent.value,
      file_path: filePath.value || undefined,
    })
    emit('created', project)
    emit('close')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Failed to create compose project'
    error.value = msg
  } finally {
    submitting.value = false
  }
}

watch(
  () => props.show,
  (val) => {
    if (val) {
      tab.value = 'template'
      selectedTemplate.value = null
      projectName.value = ''
      yamlContent.value = ''
      filePath.value = ''
      error.value = ''
      submitting.value = false
      selectedMachineId.value = props.machineId
    }
  }
)

watch(() => props.machineId, (machineId) => {
  if (!props.show) return
  selectedMachineId.value = machineId
})
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
  max-width: 580px;
  width: 100%;
  max-height: 90vh;
  overflow-y: auto;
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

.template-grid {
  max-height: 260px;
  overflow-y: auto;
}

.template-card {
  padding: var(--space-3);
  cursor: pointer;
  transition: border-color var(--transition-fast);
}

.template-card:hover {
  border-color: var(--color-accent);
}

.template-selected {
  border-color: var(--color-accent);
  background: var(--color-background-interactive-weaker);
}

.template-name {
  margin-bottom: var(--space-1);
}

.template-desc {
  line-height: var(--line-height-normal);
}

.mt-4 {
  margin-top: var(--space-4);
}

.mb-3 {
  margin-bottom: var(--space-3);
}

.form-group {
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
